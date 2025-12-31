.PHONY: build run clean test lint deps coverage coverage-html

# Build variables
BINARY_NAME := rootly-tui
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the binary
build: deps
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rootly-tui

# Build and run
run: build
	./bin/$(BINARY_NAME)

# Run without building
dev:
	go run $(LDFLAGS) ./cmd/rootly-tui

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Run tests
test:
	go test -v -count=1 ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "Total coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

# Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	@which goimports > /dev/null 2>&1 && goimports -w . || echo "goimports not installed, skipping"

# Check for issues
check: fmt lint test

# Install locally
install: build
	cp bin/$(BINARY_NAME) /usr/local/bin/

# Uninstall
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)
