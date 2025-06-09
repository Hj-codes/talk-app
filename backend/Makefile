# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=voice-chat-app
BINARY_PATH=bin/$(BINARY_NAME)

# Build flags
BUILD_FLAGS=-ldflags="-w -s"
BUILD_FLAGS_RACE=-race $(BUILD_FLAGS)

.PHONY: all build clean test coverage deps run dev docker-build docker-run help

# Default target
all: clean deps test build

# Build the application
build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH) .

# Build with race detection (for development)
build-race:
	$(GOBUILD) $(BUILD_FLAGS_RACE) -o $(BINARY_PATH) .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -rf bin/

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application
run: build
	./$(BINARY_PATH)

# Run in development mode with race detection
dev: build-race
	./$(BINARY_PATH)

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Security check (requires gosec)
security:
	gosec ./...

# Build Docker image
docker-build:
	docker build -t voice-chat-backend:latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 voice-chat-backend:latest

# Run with docker-compose
docker-compose-up:
	docker-compose up --build

# Stop docker-compose
docker-compose-down:
	docker-compose down

# Install development tools
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Create release build
release: clean deps test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH)-linux-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH)-windows-amd64.exe .

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-race     - Build with race detection"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  coverage       - Run tests with coverage"
	@echo "  bench          - Run benchmarks"
	@echo "  deps           - Download dependencies"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Build with race detection and run"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  security       - Run security checks"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-compose-up   - Run with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  install-tools  - Install development tools"
	@echo "  release        - Create release builds for multiple platforms"
	@echo "  help           - Show this help message"
