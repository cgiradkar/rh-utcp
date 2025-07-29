package providers

import (
	"fmt"
	"sync"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

// Provider is the interface that all tool providers must implement
type Provider interface {
	// GetTools returns all tools offered by this provider
	GetTools() []utcp.Tool

	// GetName returns the provider name
	GetName() string

	// GetType returns the provider type (e.g., "jira", "wiki", "gitlab")
	GetType() string

	// IsEnabled returns whether the provider is enabled
	IsEnabled() bool
}

// Factory is a function that creates a new provider instance
type Factory func(config map[string]interface{}) (Provider, error)

// Registry manages provider factories and instances
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
		providers: make(map[string]Provider),
	}
}

// RegisterFactory registers a provider factory
func (r *Registry) RegisterFactory(providerType string, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[providerType]; exists {
		return fmt.Errorf("provider type %s already registered", providerType)
	}

	r.factories[providerType] = factory
	return nil
}

// CreateProvider creates a provider instance using the registered factory
func (r *Registry) CreateProvider(name, providerType string, config map[string]interface{}) error {
	r.mu.RLock()
	factory, exists := r.factories[providerType]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown provider type: %s", providerType)
	}

	// Add name to config
	config["name"] = name

	provider, err := factory(config)
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", name, err)
	}

	r.mu.Lock()
	r.providers[name] = provider
	r.mu.Unlock()

	return nil
}

// GetProvider returns a provider by name
func (r *Registry) GetProvider(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	return provider, exists
}

// GetAllProviders returns all registered providers
func (r *Registry) GetAllProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// GetEnabledProviders returns only enabled providers
func (r *Registry) GetEnabledProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0)
	for _, provider := range r.providers {
		if provider.IsEnabled() {
			providers = append(providers, provider)
		}
	}

	return providers
}

// GetAllTools returns all tools from all enabled providers
func (r *Registry) GetAllTools() []utcp.Tool {
	providers := r.GetEnabledProviders()

	var tools []utcp.Tool
	for _, provider := range providers {
		tools = append(tools, provider.GetTools()...)
	}

	return tools
}

// Clear removes all providers from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = make(map[string]Provider)
}

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	Name    string
	Type    string
	Enabled bool
	BaseURL string
}

// GetName returns the provider name
func (b *BaseProvider) GetName() string {
	return b.Name
}

// GetType returns the provider type
func (b *BaseProvider) GetType() string {
	return b.Type
}

// IsEnabled returns whether the provider is enabled
func (b *BaseProvider) IsEnabled() bool {
	return b.Enabled
}
