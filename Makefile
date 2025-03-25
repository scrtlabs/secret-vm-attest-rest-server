# SecretAI Attest REST Server - Makefile

# Go command and flags
GO=go
BINARY_NAME=secretai-attest-rest
DEBUG_FLAGS=-gcflags="all=-N -l"
RELEASE_FLAGS=-ldflags="-s -w"

# Directories
CMD_DIR=cmd
BUILD_DIR=build

# Ensure build directory exists
$(shell mkdir -p $(BUILD_DIR))

.PHONY: all debug release clean test lint run

# Default target
all: debug

# Debug build
debug:
	@echo "Building debug version..."
	$(GO) build $(DEBUG_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(CMD_DIR)/main.go
	@echo "Debug binary built at: $(BUILD_DIR)/$(BINARY_NAME)-debug"

# Release build
release:
	@echo "Building release version..."
	$(GO) build $(RELEASE_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "Release binary built at: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at: coverage.html"

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Run the server in debug mode
run-debug: debug
	@echo "Running server (debug)..."
	$(BUILD_DIR)/$(BINARY_NAME)-debug

# Run the server in release mode
run: release
	@echo "Running server (release)..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Build both debug and release versions
build-all: debug release
	@echo "All builds completed"

# Help menu
help:
	@echo "SecretAI Attest REST Server Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all         Build debug version (default)"
	@echo "  debug       Build with debug symbols"
	@echo "  release     Build optimized for production"
	@echo "  build-all   Build both debug and release versions"
	@echo "  clean       Remove build artifacts"
	@echo "  test        Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  lint        Run linter"
	@echo "  run-debug   Build and run the debug version"
	@echo "  run         Build and run the release version"
	@echo "  help        Show this help message"