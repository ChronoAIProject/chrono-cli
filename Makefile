.PHONY: build install clean release test help

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -ldflags "-X github.com/aelfdevops/chrono/cmd.Version=$(VERSION) -X github.com/aelfdevops/chrono/cmd.BuildTime=$(BUILD_TIME) -X github.com/aelfdevops/chrono/cmd.Commit=$(COMMIT) -s -w"

# Binary name
BINARY_NAME := chrono

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✓ Built $(BINARY_NAME)"

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "✓ Installed $(BINARY_NAME)"

# Install to ~/go/bin
install-local: build
	@echo "Installing $(BINARY_NAME) to ~/go/bin..."
	@mkdir -p ~/go/bin
	@cp $(BINARY_NAME) ~/go/bin/
	@mkdir -p ~/go/share/chrono
	@cp -r templates ~/go/share/chrono/
	@echo "✓ Installed $(BINARY_NAME) to ~/go/bin"
	@echo "✓ Installed templates to ~/go/share/chrono/"

# Build release binaries for all platforms
release:
	@echo "Building release binaries..."
	@echo "Version: $(VERSION)"
	@echo ""
	@for GOOS in darwin linux windows; do \
		for GOARCH in amd64 arm64; do \
			if [ "$$GOOS" = "windows" ]; then \
				OUTPUT="$(BINARY_NAME)-$$GOOS-$$GOARCH.exe"; \
			else \
				OUTPUT="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
			fi; \
			echo "Building $$OUTPUT..."; \
			GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o "$$OUTPUT" .; \
		done; \
	done
	@echo "✓ Built all binaries"
	@sha256sum $(BINARY_NAME)-* > checksums.txt
	@echo "✓ Generated checksums.txt"

# Run tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Run integration test script
test-integration: build
	@echo "Running integration tests..."
	@./test.sh

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-*
	@rm -f checksums.txt
	@echo "✓ Cleaned"

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT)"

# Help
help:
	@echo "Chrono CLI Build System"
	@echo ""
	@echo "Targets:"
	@echo "  make build         - Build for current platform"
	@echo "  make install       - Install to /usr/local/bin (requires sudo)"
	@echo "  make install-local - Install to ~/go/bin"
	@echo "  make release       - Build release binaries for all platforms"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make version       - Show version info"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build for current platform"
	@echo "  make VERSION=1.0.0      # Build with specific version"
	@echo "  make install            # Install system-wide"
