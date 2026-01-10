.PHONY: build run test clean lint fmt help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Binary name
BINARY_NAME=eocall
BINARY_DIR=bin

# Main package
MAIN_PACKAGE=./cmd/server

# Build the project
build:
	@echo "Building..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Run the project
run:
	$(GOCMD) run $(MAIN_PACKAGE)

# Run with config
run-dev:
	$(GOCMD) run $(MAIN_PACKAGE) -config configs/config.yaml

# Test
test:
	$(GOTEST) -v ./...

# Test with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build files
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html

# Format code
fmt:
	$(GOFMT) -s -w .

# Lint
lint:
	golangci-lint run ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the application"
	@echo "  run-dev       - Run with config file"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build files"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Download dependencies"
