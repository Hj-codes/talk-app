#!/bin/bash

# Voice Chat App Test Runner
# This script runs all tests with coverage and race detection

set -e

echo "🧪 Running Voice Chat App Tests"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Please run this script from the backend directory"
    exit 1
fi

# Clean up any previous test artifacts
print_status "Cleaning up previous test artifacts..."
rm -f coverage.out coverage.html

# Download dependencies
print_status "Downloading dependencies..."
go mod download
go mod tidy

# Run unit tests with race detection and coverage
print_status "Running unit tests with race detection..."
go test -race -coverprofile=coverage.out -covermode=atomic ./...

if [ $? -eq 0 ]; then
    print_status "✅ All tests passed!"
else
    print_error "❌ Some tests failed!"
    exit 1
fi

# Generate coverage report
print_status "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Display coverage summary
print_status "Coverage Summary:"
go tool cover -func=coverage.out | tail -1

# Run specific test categories
echo ""
print_status "Running specific test categories..."

# Unit tests only
print_status "Running unit tests only..."
go test -race -v ./models ./utils ./handlers

# Integration tests only
print_status "Running integration tests only..."
go test -race -v ./tests

# Benchmark tests
print_status "Running benchmark tests..."
go test -bench=. -benchmem ./...

# Check for race conditions in concurrent scenarios
print_status "Running extended race condition tests..."
go test -race -count=10 ./models -run="Concurrent|Race"

# Test with different build tags if any
print_status "Running tests with different configurations..."
go test -tags=integration -race ./...

echo ""
print_status "🎉 Test suite completed successfully!"
print_status "📊 Coverage report generated: coverage.html"
print_status "📈 Open coverage.html in your browser to view detailed coverage"

# Optional: Open coverage report in browser (uncomment if desired)
# if command -v open &> /dev/null; then
#     open coverage.html
# elif command -v xdg-open &> /dev/null; then
#     xdg-open coverage.html
# fi 