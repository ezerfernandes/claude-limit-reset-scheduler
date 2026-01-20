.PHONY: build test lint run clean install fmt vet check help

# Binary name
BINARY_NAME=calgo
# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod

# Build the binary
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/calgo

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Run the application
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
install:
	$(GOCMD) install ./cmd/calgo

# Format code
fmt:
	$(GOFMT) ./...

# Run go vet
vet:
	$(GOVET) ./...

# Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  run           - Build and run the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  check         - Run fmt, vet, lint, and test"
	@echo "  tidy          - Tidy go.mod dependencies"
	@echo "  help          - Show this help message"
