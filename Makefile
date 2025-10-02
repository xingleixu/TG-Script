# TG-Script Project Makefile

.PHONY: all build test clean fmt vet bench deps docs examples dev profile coverage

# Default target
all: build test

# Build project
build:
	@echo "Building TG-Script..."
	go build -o bin/tg ./cmd/tg
	@echo "Build complete: bin/tg"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Static analysis
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Generate documentation
docs:
	@echo "Generating documentation..."
	go doc -all ./...

# Run examples
examples: build
	@echo "Running examples..."
	./bin/tg run examples/hello.tg
	@echo "Examples complete"

# Development mode (watch file changes)
dev:
	@echo "Starting development mode..."
	# File watching and auto-rebuild logic can be added here

# Performance profiling
profile:
	@echo "Running performance analysis..."
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...

# Code coverage
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html