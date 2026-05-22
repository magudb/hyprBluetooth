# hyprBluetooth Makefile

# Build variables
BINARY_NAME=hyprBluetooth
BUILD_DIR=build
INSTALL_PATH=/usr/local/bin

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Git / build metadata (matches .goreleaser.yml ldflags)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build clean test lint fmt vet install uninstall deps tidy setup build-all run help

# Default target
all: build

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built $(BINARY_NAME) in $(BUILD_DIR)/"

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build directory"

# Run tests
test:
	$(GOTEST) -v ./...

# Run golangci-lint
lint:
	golangci-lint run

# Format Go source
fmt:
	$(GOFMT) ./...

# Run go vet
vet:
	$(GOVET) ./...

# Run the full set of checks: fmt, vet, lint, test
check: fmt vet lint test

# Install to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall from system
uninstall:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_PATH)"
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Download dependencies
deps:
	$(GOGET) ./...

# Tidy up go.mod
tidy:
	$(GOMOD) tidy

# Setup development environment
setup: deps
	@echo "Setting up development environment..."
	@mkdir -p $(BUILD_DIR)
	@echo "Development environment ready"

# Build for the platforms targeted by .goreleaser.yml (linux only — bluetoothctl is Linux-only)
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "Multi-platform build complete"

# Build and run
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Display help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary into $(BUILD_DIR)/"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  lint        - Run golangci-lint"
	@echo "  fmt         - Format Go source"
	@echo "  vet         - Run go vet"
	@echo "  check       - Run fmt, vet, lint, and test"
	@echo "  install     - Install to $(INSTALL_PATH) (requires sudo)"
	@echo "  uninstall   - Remove from $(INSTALL_PATH) (requires sudo)"
	@echo "  deps        - Download dependencies"
	@echo "  tidy        - Tidy go.mod file"
	@echo "  setup       - Setup development environment"
	@echo "  build-all   - Build for linux amd64 and arm64"
	@echo "  run         - Build and run the app"
	@echo "  help        - Show this help message"
