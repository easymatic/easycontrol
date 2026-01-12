# Makefile for easycontrol
# Builds binaries for Linux and macOS

# Binary name
BINARY_NAME=easycontrol

# Version (can be set via make VERSION=1.0.0)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# LDFLAGS for versioning
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all clean linux darwin-amd64 darwin-arm64 help

# Default target
all: linux darwin-amd64 darwin-arm64

# Build for Linux (amd64)
linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/main.go
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Build for macOS (amd64 - Intel)
darwin-amd64:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/main.go
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

# Build for macOS (arm64 - Apple Silicon)
darwin-arm64:
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/main.go
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

# Build for current platform
local:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "✓ Cleaned"

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "✓ Dependencies updated"

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Build binaries for all platforms (Linux, macOS Intel, macOS ARM)"
	@echo "  linux         - Build binary for Linux (amd64)"
	@echo "  darwin-amd64  - Build binary for macOS Intel (amd64)"
	@echo "  darwin-arm64  - Build binary for macOS Apple Silicon (arm64)"
	@echo "  local         - Build binary for current platform"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  deps          - Download and update dependencies"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make all              # Build all platforms"
	@echo "  make linux            # Build only for Linux"
	@echo "  make VERSION=1.0.0    # Build with specific version"
