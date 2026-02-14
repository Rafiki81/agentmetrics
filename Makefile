BINARY_NAME=agentmetrics
BUILD_DIR=bin
VERSION=0.1.0
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build clean run install test fmt lint help

## build: Build binary for current platform
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agentmetrics
	@echo "✅ Binary: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Build and run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

## scan: Run a quick scan
scan: build
	./$(BUILD_DIR)/$(BINARY_NAME) scan

## watch: Run in watch mode
watch: build
	./$(BUILD_DIR)/$(BINARY_NAME) watch

## install: Install binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/agentmetrics

## clean: Remove generated files
clean:
	@rm -rf $(BUILD_DIR)
	@echo "🧹 Clean"

## test: Run tests
test:
	go test -v ./...

## fmt: Format code
fmt:
	go fmt ./...

## lint: Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

## deps: Download dependencies
deps:
	go mod tidy
	go mod download

## help: Show this help
help:
	@echo "◈ AgentMetrics — Available commands:"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
