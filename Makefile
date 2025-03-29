# SecretVM Attest REST Server - Makefile

# Go command and flags
GO=go
BINARY_NAME=secret-vm-attest-rest-server
DEBUG_FLAGS=-gcflags="all=-N -l"
RELEASE_FLAGS=-ldflags="-s -w"

# Directories
BUILD_DIR=build

# Ensure build directory exists
$(shell mkdir -p $(BUILD_DIR))

.PHONY: all debug release clean test lint run get-deps

# Default target
all: debug

# Download dependencies
get-deps:
	@echo "===== Downloading dependencies ====="
	@echo "Go version: $$($(GO) version)"
	@echo "Go modules: $$($(GO) env GOMOD)"
	@echo "Go proxy: $$(GOPROXY=direct $(GO) env GOPROXY)"
	@echo "Dependencies from go.mod:"
	@cat go.mod | grep require -A 100
	@echo "Starting downloads with direct connection..."
	GOPROXY=direct $(GO) mod download -x
	@echo "===== Dependencies download complete ====="

# Debug build
debug: get-deps
	@echo "===== Building debug version ====="
	@echo "Using build flags: $(DEBUG_FLAGS)"
	@echo "Main file: main.go"
	@echo "Output binary: $(BUILD_DIR)/$(BINARY_NAME)-debug"
	@echo "Building..."
	$(GO) build -v $(DEBUG_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-debug main.go
	@echo "Build completed at: $$(date)"
	@echo "File details: $$(ls -lh $(BUILD_DIR)/$(BINARY_NAME)-debug)"
	@echo "===== Debug build complete ====="

# Release build
release: get-deps
	@echo "===== Building release version ====="
	@echo "Using build flags: $(RELEASE_FLAGS)"
	@echo "Main file: main.go"
	@echo "Output binary: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Building..."
	$(GO) build -v $(RELEASE_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build completed at: $$(date)"
	@echo "File details: $$(ls -lh $(BUILD_DIR)/$(BINARY_NAME))"
	@echo "===== Release build complete ====="

# Clean build artifacts
clean:
	@echo "===== Cleaning build directory ====="
	@echo "Removing: $(BUILD_DIR)"
	@if [ -d "$(BUILD_DIR)" ]; then \
		echo "Contents before removal:"; \
		ls -la $(BUILD_DIR); \
	else \
		echo "Build directory doesn't exist. Nothing to clean."; \
	fi
	rm -rf $(BUILD_DIR)
	@echo "Removal complete at: $$(date)"
	@echo "===== Clean complete ====="

# Run tests
test: get-deps
	@echo "===== Running tests ====="
	@echo "Go version: $$($(GO) version)"
	@echo "Test packages: $$($(GO) list ./...)"
	@echo "Starting tests at: $$(date)"
	$(GO) test -v ./...
	@echo "Tests completed at: $$(date)"
	@echo "===== Tests complete ====="

# Run tests with coverage
test-coverage: get-deps
	@echo "===== Running tests with coverage ====="
	@echo "Go version: $$($(GO) version)"
	@echo "Test packages: $$($(GO) list ./...)"
	@echo "Starting tests at: $$(date)"
	$(GO) test -v -coverprofile=coverage.out ./...
	@echo "Coverage details:"
	$(GO) tool cover -func=coverage.out
	@echo "Generating HTML report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at: coverage.html"
	@echo "Report size: $$(ls -lh coverage.html)"
	@echo "Tests completed at: $$(date)"
	@echo "===== Coverage tests complete ====="

# Lint code
lint:
	@echo "===== Linting code ====="
	@echo "Checking for linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Found golangci-lint: $$(golangci-lint --version)"; \
		echo "Starting lint at: $$(date)"; \
		golangci-lint run -v; \
		echo "Lint completed at: $$(date)"; \
	else \
		echo "golangci-lint not found!"; \
		echo "Please install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@echo "===== Lint complete ====="

# Run the server in debug mode
run-debug: debug
	@echo "===== Running server (debug mode) ====="
	@echo "Binary: $(BUILD_DIR)/$(BINARY_NAME)-debug"
	@echo "Started at: $$(date)"
	@echo "Process info:"
	@echo "---------------"
	$(BUILD_DIR)/$(BINARY_NAME)-debug
	@echo "Server exited at: $$(date)"
	@echo "===== Server stopped ====="

# Run the server in release mode
run: release
	@echo "===== Running server (release mode) ====="
	@echo "Binary: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Started at: $$(date)"
	@echo "Process info:"
	@echo "---------------"
	$(BUILD_DIR)/$(BINARY_NAME)
	@echo "Server exited at: $$(date)"
	@echo "===== Server stopped ====="

# Build both debug and release versions
build-all: debug release
	@echo "===== All builds completed ====="
	@echo "Debug binary: $(BUILD_DIR)/$(BINARY_NAME)-debug"
	@echo "Release binary: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Build directory contents:"
	@ls -lh $(BUILD_DIR)/
	@echo "===== Build summary complete ====="

# Help menu
help:
	@echo "SecretVM Attest REST Server Makefile"
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
	@echo "  get-deps    Download dependencies directly"
	@echo "  test        Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  lint        Run linter"
	@echo "  run-debug   Build and run the debug version"
	@echo "  run         Build and run the release version"
	@echo "  help        Show this help message"
