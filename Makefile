.PHONY: help build run test clean docker-build docker-run setup

# Default target
help:
	@echo "Available targets:"
	@echo "  setup        - Install dependencies and set up the project"
	@echo "  build        - Build the server binary"
	@echo "  run          - Run the server locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"

# Setup the project
setup:
	@echo "Setting up RH-UTCP project..."
	go mod tidy
	@echo "Creating .env file from example..."
	@if [ ! -f .env ]; then cp env.example .env; echo "Please edit .env with your credentials"; fi

# Build the server
build:
	@echo "Building RH-UTCP server..."
	go build -o bin/rh-utcp-server cmd/server/main.go

# Run the server
run:
	@echo "Starting RH-UTCP server..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Build Docker image
docker: docker-build

docker-build:
	@echo "Building Docker image..."
	podman build -t rh-utcp:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	podman run -p 8080:8080 --env-file .env rh-utcp:latest

# Development server with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Generate mocks (if needed for testing)
mocks:
	@echo "Generating mocks..."
	go generate ./... 