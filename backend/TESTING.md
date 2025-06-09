# Testing Documentation for Voice Chat App

This document provides a comprehensive overview of the testing strategy and implementation for the voice chat application backend.

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Test Structure](#test-structure)
3. [Running Tests](#running-tests)
4. [Test Coverage](#test-coverage)
5. [Test Categories](#test-categories)
6. [Continuous Integration](#continuous-integration)
7. [Load Testing](#load-testing)
8. [Troubleshooting](#troubleshooting)

## Testing Strategy

Our testing approach follows the testing pyramid with emphasis on:

- **Unit Tests (70%)**: Fast, isolated tests for individual functions and methods
- **Integration Tests (20%)**: Tests for component interactions and workflows
- **End-to-End Tests (10%)**: Full system tests simulating real user scenarios

### Key Testing Principles

1. **Fast Feedback**: Tests should run quickly to enable rapid development
2. **Reliability**: Tests should be deterministic and not flaky
3. **Maintainability**: Tests should be easy to understand and modify
4. **Coverage**: Critical paths and edge cases should be thoroughly tested
5. **Isolation**: Tests should not depend on external services or state

## Test Structure

```
backend/
├── models/
│   ├── models.go
│   └── models_test.go          # Unit tests for user pool, connections
├── handlers/
│   ├── signalings.go
│   └── signalings_test.go      # Unit tests for WebSocket handlers
├── utils/
│   ├── config.go
│   ├── config_test.go          # Unit tests for configuration
│   ├── jwt.go
│   └── jwt_test.go             # Unit tests for JWT utilities
├── tests/
│   └── integration_test.go     # Integration tests
└── scripts/
    ├── run_tests.sh            # Test runner script
    └── load_test.md            # Load testing documentation
```

## Running Tests

### Prerequisites

1. Go 1.22 or later
2. All dependencies installed (`go mod tidy`)

### Quick Test Run

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

### Using the Test Runner Script

```bash
# Make script executable (Unix/Linux/macOS)
chmod +x scripts/run_tests.sh

# Run comprehensive test suite
./scripts/run_tests.sh
```

### Specific Test Categories

```bash
# Unit tests only
go test ./models ./utils ./handlers

# Integration tests only
go test ./tests

# Benchmark tests
go test -bench=. ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Coverage

### Coverage Targets

- **Overall Coverage**: > 80%
- **Critical Components**: > 90%
  - User pool management
  - WebSocket handling
  - JWT authentication
  - Matchmaking logic

### Viewing Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

## Test Categories

### 1. Unit Tests

#### Models Package (`models/models_test.go`)

**Connection Tests**:
- `TestConnection_WriteJSON`: Tests WebSocket message writing with various states
- `TestConnection_UpdatePing`: Tests ping timestamp updates
- `TestConnection_Close`: Tests connection cleanup

**UserPool Tests**:
- `TestNewUserPool`: Tests user pool initialization
- `TestUserPool_AddWaitingUser`: Tests adding users to waiting queue
- `TestUserPool_GetRandomWaitingUser`: Tests random user selection
- `TestUserPool_CreateRoom`: Tests room creation and user pairing
- `TestUserPool_RemoveUser`: Tests user removal and cleanup
- `TestUserPool_Stats`: Tests statistics generation

**Concurrency Tests**:
- `TestUserPool_ConcurrentAccess`: Tests thread-safe operations
- `TestUserPool_MatchmakingRaceCondition`: Tests race conditions in matching

#### Handlers Package (`handlers/signalings_test.go`)

**SignalingServer Tests**:
- `TestSignalingServer_GetStats`: Tests statistics endpoint
- `TestSignalingServer_handleFindMatch_NoPartner`: Tests matchmaking with no available partners
- `TestSignalingServer_handleFindMatch_WithPartner`: Tests successful matchmaking
- `TestSignalingServer_relaySignaling_ValidRelay`: Tests message relaying between users
- `TestSignalingServer_relaySignaling_InvalidCases`: Tests error handling in message relay

**Concurrency Tests**:
- `TestSignalingServer_ConcurrentMatchmaking`: Tests concurrent user matching
- `BenchmarkSignalingServer_HandleFindMatch`: Benchmarks matchmaking performance

#### Utils Package

**Config Tests** (`utils/config_test.go`):
- `TestLoadConfig_Defaults`: Tests default configuration values
- `TestLoadConfig_EnvironmentVariables`: Tests environment variable override
- `TestGetEnv`: Tests environment variable retrieval
- `TestGetIntEnv`: Tests integer environment variable parsing
- `TestGetDurationEnv`: Tests duration environment variable parsing
- `TestLoadConfig_Concurrent`: Tests concurrent config loading

**JWT Tests** (`utils/jwt_test.go`):
- `TestGenerateUUID`: Tests UUID generation uniqueness
- `TestGenerateToken`: Tests JWT token generation
- `TestValidateJWT_ValidToken`: Tests valid token validation
- `TestValidateJWT_InvalidToken`: Tests invalid token handling
- `TestValidateJWT_ExpiredToken`: Tests expired token handling
- `TestJWT_RoundTrip`: Tests token generation and validation cycle
- `TestJWT_ConcurrentGeneration`: Tests concurrent token operations

### 2. Integration Tests (`tests/integration_test.go`)

**Connection Tests**:
- `TestIntegration_SingleUserConnection`: Tests single user WebSocket connection
- `TestIntegration_TwoUserMatchmaking`: Tests complete matchmaking workflow
- `TestIntegration_HealthAndStatsEndpoints`: Tests HTTP endpoints

**Workflow Tests**:
- Complete user journey from connection to matching
- WebRTC signaling message relay
- User disconnection handling
- Multi-user scenarios

### 3. Benchmark Tests

**Performance Benchmarks**:
- `BenchmarkSignalingServer_HandleFindMatch`: Matchmaking performance
- `BenchmarkGenerateToken`: JWT token generation speed
- `BenchmarkValidateJWT`: JWT validation speed
- `BenchmarkGenerateUUID`: UUID generation speed
- `BenchmarkLoadConfig`: Configuration loading speed

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.22'
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Pre-commit Hooks

```bash
# Install pre-commit hooks
go install github.com/pre-commit/pre-commit@latest

# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: go-test
        name: go test
        entry: go test ./...
        language: system
        pass_filenames: false
      
      - id: go-vet
        name: go vet
        entry: go vet ./...
        language: system
        pass_filenames: false
```

## Load Testing

See [scripts/load_test.md](scripts/load_test.md) for comprehensive load testing strategies including:

- WebSocket connection testing with Artillery.js
- HTTP endpoint testing with Vegeta
- Custom Go load testing scripts
- Performance monitoring and metrics
- Stress testing scenarios

## Troubleshooting

### Common Test Issues

#### 1. Race Condition Failures

```bash
# Run with race detector
go test -race ./...

# Run multiple times to catch intermittent issues
go test -count=10 ./...
```

#### 2. Timeout Issues

```bash
# Increase test timeout
go test -timeout=30s ./...
```

#### 3. WebSocket Connection Issues

```bash
# Check if server is running
curl http://localhost:8080/health

# Verify WebSocket upgrade
wscat -c ws://localhost:8080/ws
```

#### 4. Memory Issues

```bash
# Profile memory usage
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Test Environment Setup

#### Environment Variables for Testing

```bash
export JWT_SECRET="test-secret-key"
export PORT="8080"
export LOG_LEVEL="debug"
```

#### Docker Test Environment

```dockerfile
# Dockerfile.test
FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go test -c ./...

CMD ["./test.test"]
```

### Debugging Failed Tests

#### 1. Verbose Output

```bash
go test -v ./... | grep FAIL
```

#### 2. Specific Test Debugging

```bash
# Run specific test with debugging
go test -v -run TestSpecificFunction ./package

# Add debug prints in test code
t.Logf("Debug info: %+v", variable)
```

#### 3. Test Coverage Analysis

```bash
# Find uncovered code
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

## Best Practices

### Writing Good Tests

1. **Use Table-Driven Tests**:
```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {"valid input", "test", "TEST", false},
    {"empty input", "", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := Function(tt.input)
        if tt.wantErr {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        }
    })
}
```

2. **Use Proper Test Cleanup**:
```go
func TestFunction(t *testing.T) {
    // Setup
    resource := setupResource()
    defer resource.Cleanup() // Always cleanup
    
    // Test logic
}
```

3. **Test Edge Cases**:
- Empty inputs
- Nil pointers
- Boundary conditions
- Error conditions

4. **Use Meaningful Test Names**:
```go
func TestUserPool_CreateRoom_WithValidUsers_ShouldCreateRoomAndUpdateUsers(t *testing.T)
func TestJWT_ValidateToken_WithExpiredToken_ShouldReturnError(t *testing.T)
```

### Performance Testing Guidelines

1. **Benchmark Critical Paths**: Focus on functions called frequently
2. **Use Realistic Data**: Test with production-like data sizes
3. **Measure Memory Allocations**: Use `-benchmem` flag
4. **Compare Results**: Track performance over time

This comprehensive testing strategy ensures the voice chat application is robust, performant, and maintainable. 