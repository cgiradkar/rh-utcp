package providers

import (
	"fmt"
	"testing"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

// MockProvider is a mock implementation of the Provider interface
type MockProvider struct {
	BaseProvider
	ToolsFunc func() []utcp.Tool
}

func (m *MockProvider) GetTools() []utcp.Tool {
	if m.ToolsFunc != nil {
		return m.ToolsFunc()
	}
	return []utcp.Tool{}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if registry.factories == nil {
		t.Error("factories map is nil")
	}

	if registry.providers == nil {
		t.Error("providers map is nil")
	}
}

func TestRegisterFactory(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	err := registry.RegisterFactory("test", func(config map[string]interface{}) (Provider, error) {
		return &MockProvider{
			BaseProvider: BaseProvider{
				Name:    "test",
				Type:    "test",
				Enabled: true,
			},
		}, nil
	})

	if err != nil {
		t.Errorf("RegisterFactory failed: %v", err)
	}

	// Test duplicate registration
	err = registry.RegisterFactory("test", func(config map[string]interface{}) (Provider, error) {
		return nil, nil
	})

	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestCreateProvider(t *testing.T) {
	registry := NewRegistry()

	// Register a factory
	registry.RegisterFactory("mock", func(config map[string]interface{}) (Provider, error) {
		name, _ := config["name"].(string)
		enabled, _ := config["enabled"].(bool)

		return &MockProvider{
			BaseProvider: BaseProvider{
				Name:    name,
				Type:    "mock",
				Enabled: enabled,
			},
		}, nil
	})

	// Test successful provider creation
	err := registry.CreateProvider("test-provider", "mock", map[string]interface{}{
		"enabled": true,
	})

	if err != nil {
		t.Errorf("CreateProvider failed: %v", err)
	}

	// Verify provider was created
	provider, exists := registry.GetProvider("test-provider")
	if !exists {
		t.Error("Provider not found after creation")
	}

	if provider.GetName() != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got %s", provider.GetName())
	}

	// Test creating provider with unknown type
	err = registry.CreateProvider("unknown", "unknown-type", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for unknown provider type, got nil")
	}
}

func TestGetProvider(t *testing.T) {
	registry := NewRegistry()

	// Add a provider directly
	mockProvider := &MockProvider{
		BaseProvider: BaseProvider{
			Name:    "test",
			Type:    "mock",
			Enabled: true,
		},
	}

	registry.providers["test"] = mockProvider

	// Test getting existing provider
	provider, exists := registry.GetProvider("test")
	if !exists {
		t.Error("Expected provider to exist")
	}

	if provider != mockProvider {
		t.Error("Got different provider instance")
	}

	// Test getting non-existent provider
	_, exists = registry.GetProvider("non-existent")
	if exists {
		t.Error("Expected provider to not exist")
	}
}

func TestGetAllProviders(t *testing.T) {
	registry := NewRegistry()

	// Add multiple providers
	provider1 := &MockProvider{
		BaseProvider: BaseProvider{Name: "p1", Type: "mock", Enabled: true},
	}
	provider2 := &MockProvider{
		BaseProvider: BaseProvider{Name: "p2", Type: "mock", Enabled: false},
	}

	registry.providers["p1"] = provider1
	registry.providers["p2"] = provider2

	providers := registry.GetAllProviders()

	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}

	// Check that both providers are in the result
	foundP1, foundP2 := false, false
	for _, p := range providers {
		if p.GetName() == "p1" {
			foundP1 = true
		}
		if p.GetName() == "p2" {
			foundP2 = true
		}
	}

	if !foundP1 || !foundP2 {
		t.Error("Not all providers were returned")
	}
}

func TestGetEnabledProviders(t *testing.T) {
	registry := NewRegistry()

	// Add providers with different enabled states
	provider1 := &MockProvider{
		BaseProvider: BaseProvider{Name: "enabled1", Type: "mock", Enabled: true},
	}
	provider2 := &MockProvider{
		BaseProvider: BaseProvider{Name: "disabled", Type: "mock", Enabled: false},
	}
	provider3 := &MockProvider{
		BaseProvider: BaseProvider{Name: "enabled2", Type: "mock", Enabled: true},
	}

	registry.providers["enabled1"] = provider1
	registry.providers["disabled"] = provider2
	registry.providers["enabled2"] = provider3

	providers := registry.GetEnabledProviders()

	if len(providers) != 2 {
		t.Errorf("Expected 2 enabled providers, got %d", len(providers))
	}

	// Verify only enabled providers are returned
	for _, p := range providers {
		if !p.IsEnabled() {
			t.Errorf("Got disabled provider: %s", p.GetName())
		}

		if p.GetName() == "disabled" {
			t.Error("Disabled provider should not be in enabled list")
		}
	}
}

func TestGetAllTools(t *testing.T) {
	registry := NewRegistry()

	// Create providers with tools
	tools1 := []utcp.Tool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}

	tools2 := []utcp.Tool{
		{Name: "tool3", Description: "Tool 3"},
	}

	provider1 := &MockProvider{
		BaseProvider: BaseProvider{Name: "p1", Type: "mock", Enabled: true},
		ToolsFunc: func() []utcp.Tool {
			return tools1
		},
	}

	provider2 := &MockProvider{
		BaseProvider: BaseProvider{Name: "p2", Type: "mock", Enabled: true},
		ToolsFunc: func() []utcp.Tool {
			return tools2
		},
	}

	provider3 := &MockProvider{
		BaseProvider: BaseProvider{Name: "p3", Type: "mock", Enabled: false},
		ToolsFunc: func() []utcp.Tool {
			return []utcp.Tool{{Name: "disabled_tool"}}
		},
	}

	registry.providers["p1"] = provider1
	registry.providers["p2"] = provider2
	registry.providers["p3"] = provider3

	allTools := registry.GetAllTools()

	// Should have 3 tools from enabled providers only
	if len(allTools) != 3 {
		t.Errorf("Expected 3 tools from enabled providers, got %d", len(allTools))
	}

	// Verify the tools
	expectedTools := map[string]bool{
		"tool1": false,
		"tool2": false,
		"tool3": false,
	}

	for _, tool := range allTools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		} else if tool.Name == "disabled_tool" {
			t.Error("Tool from disabled provider should not be included")
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestClear(t *testing.T) {
	registry := NewRegistry()

	// Add some providers
	registry.providers["p1"] = &MockProvider{
		BaseProvider: BaseProvider{Name: "p1"},
	}
	registry.providers["p2"] = &MockProvider{
		BaseProvider: BaseProvider{Name: "p2"},
	}

	if len(registry.providers) != 2 {
		t.Errorf("Expected 2 providers before clear, got %d", len(registry.providers))
	}

	// Clear the registry
	registry.Clear()

	if len(registry.providers) != 0 {
		t.Errorf("Expected 0 providers after clear, got %d", len(registry.providers))
	}

	// Verify factories are not cleared
	registry.RegisterFactory("test", func(config map[string]interface{}) (Provider, error) {
		return nil, nil
	})

	registry.Clear()

	if len(registry.factories) == 0 {
		t.Error("Factories should not be cleared")
	}
}

func TestBaseProvider(t *testing.T) {
	base := BaseProvider{
		Name:    "test-provider",
		Type:    "test-type",
		Enabled: true,
		BaseURL: "https://api.example.com",
	}

	if base.GetName() != "test-provider" {
		t.Errorf("Expected name 'test-provider', got %s", base.GetName())
	}

	if base.GetType() != "test-type" {
		t.Errorf("Expected type 'test-type', got %s", base.GetType())
	}

	if !base.IsEnabled() {
		t.Error("Expected provider to be enabled")
	}

	// Test disabled provider
	base.Enabled = false
	if base.IsEnabled() {
		t.Error("Expected provider to be disabled")
	}
}

func TestConcurrency(t *testing.T) {
	registry := NewRegistry()

	// Register factory
	registry.RegisterFactory("concurrent", func(config map[string]interface{}) (Provider, error) {
		name, _ := config["name"].(string)
		return &MockProvider{
			BaseProvider: BaseProvider{
				Name:    name,
				Type:    "concurrent",
				Enabled: true,
			},
		}, nil
	})

	// Run concurrent operations
	done := make(chan bool)

	// Create providers concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			name := fmt.Sprintf("provider-%d", id)
			registry.CreateProvider(name, "concurrent", map[string]interface{}{})
			done <- true
		}(i)
	}

	// Get providers concurrently
	for i := 0; i < 10; i++ {
		go func() {
			registry.GetAllProviders()
			registry.GetEnabledProviders()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all providers were created
	providers := registry.GetAllProviders()
	if len(providers) != 10 {
		t.Errorf("Expected 10 providers, got %d", len(providers))
	}
}
