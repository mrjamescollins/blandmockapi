# Testing Guide

Comprehensive testing strategy for Bland Mock API, including unit tests, integration tests, and deployment validation.

## Table of Contents

- [Unit Tests](#unit-tests)
- [Integration Tests](#integration-tests)
- [Test Coverage](#test-coverage)
- [CI/CD Integration](#cicd-integration)
- [Deployment Testing](#deployment-testing)

## Unit Tests

Unit tests cover individual components and functions in isolation.

### Running Unit Tests

```bash
# Run all unit tests
go test ./internal/... -v

# Run with coverage
go test ./internal/... -v -cover

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Test Organization

```
internal/
├── config/
│   └── loader_test.go      # Configuration loading tests
├── models/
│   └── config_test.go      # Data model tests
├── router/
│   ├── router_test.go      # Router logic tests
│   └── handler_test.go     # HTTP handler tests
└── graphql/
    └── handler_test.go     # GraphQL handler tests
```

### Writing Unit Tests

Example unit test structure:

```go
func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test"

    // Act
    result := MyFunction(input)

    // Assert
    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }
}
```

### Coverage Goals

- **Overall**: > 80%
- **Critical paths**: > 90%
- **Handler functions**: > 85%
- **Configuration loading**: > 70%

### Current Coverage

```
Package                                    Coverage
github.com/jimbo/blandmockapi/internal/config    67.4%
github.com/jimbo/blandmockapi/internal/models   100.0%
github.com/jimbo/blandmockapi/internal/router    86.7%
```

## Integration Tests

Integration tests validate the entire system working together.

### Running Integration Tests

```bash
# Run integration tests
go test -tags=integration ./test/integration/... -v

# With timeout
go test -tags=integration ./test/integration/... -v -timeout 60s

# Using the test script
./scripts/test.sh
```

### Integration Test Structure

Integration tests are located in `test/integration/` and use build tags:

```go
// +build integration

package integration

func TestHealthEndpoint(t *testing.T) {
    resp, err := http.Get(baseURL + "/health")
    // ... test implementation
}
```

### Test Coverage

Integration tests cover:

1. **Health Check**: Validates `/health` endpoint
2. **REST Endpoints**:
   - GET requests
   - POST requests
   - Method validation (405 responses)
   - Not found handling (404 responses)
3. **GraphQL**:
   - Queries
   - Mutations
   - Error handling
4. **Performance**:
   - Delay functionality
   - Timeout handling
5. **Headers**:
   - Custom headers
   - Content-Type validation

### Test Fixtures

Test fixtures are in `test/fixtures/`:

```toml
# test-config.toml
[server]
port = 8888

[[endpoints]]
path = "/test/hello"
method = "GET"
status = 200
response = '{"message": "Hello from test config"}'
```

## Test Coverage

### Generating Coverage Reports

```bash
# Unit test coverage
go test ./internal/... -coverprofile=coverage-unit.out
go tool cover -func=coverage-unit.out

# HTML report
go tool cover -html=coverage-unit.out -o coverage.html

# View in browser
open coverage.html
```

### Coverage Badges

Add to your README:

```markdown
[![Coverage](https://img.shields.io/codecov/c/github/mrjamescollins/blandmockapi)](https://codecov.io/gh/mrjamescollins/blandmockapi)
```

## CI/CD Integration

### GitHub Actions

Automated testing runs on every push and PR:

```yaml
# .github/workflows/ci.yml
- Run unit tests
- Run integration tests
- Build binaries
- Build Docker images
- Upload coverage to Codecov
```

### TeamCity

Configured build steps:

1. **Unit Tests**: Run all unit tests with coverage
2. **Integration Tests**: Run integration test suite
3. **Build Docker**: Build and push Docker images
4. **Deploy**: Deploy to AWS (Lambda/ECS/App Runner)

### Running TeamCity Builds

```bash
# Trigger builds via TeamCity REST API
curl -X POST https://teamcity.example.com/app/rest/buildQueue \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/xml" \
  -d '<build><buildType id="BlandMockApi_UnitTests"/></build>'
```

## Deployment Testing

### Post-Deployment Validation

Each deployment includes automated validation:

#### Lambda Deployment Test

```bash
# Get function URL
FUNCTION_URL=$(aws lambda get-function-url-config \
  --function-name blandmockapi-dev \
  --query 'FunctionUrl' --output text)

# Test health endpoint
curl -f "$FUNCTION_URL/health" || exit 1

# Test API endpoint
curl -f "$FUNCTION_URL/api/users" || exit 1
```

#### ECS Deployment Test

```bash
# Get ALB URL
ALB_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiEcsStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerUrl`].OutputValue' \
  --output text)

# Wait for service stability
aws ecs wait services-stable \
  --cluster blandmockapi-dev \
  --services blandmockapi-dev

# Test endpoint
curl -f "$ALB_URL/health" || exit 1
```

#### App Runner Deployment Test

```bash
# Get service URL
SERVICE_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiAppRunnerStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ServiceUrl`].OutputValue' \
  --output text)

# Test endpoint
curl -f "$SERVICE_URL/health" || exit 1
```

### Load Testing

#### Using Apache Bench

```bash
# Basic load test
ab -n 1000 -c 10 http://localhost:8080/api/users

# With keep-alive
ab -n 5000 -c 50 -k http://localhost:8080/api/users
```

#### Using hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run load test
hey -n 10000 -c 100 -m GET http://localhost:8080/api/users
```

#### Using k6

```javascript
// load-test.js
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
};

export default function () {
  let res = http.get('http://localhost:8080/api/users');
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
```

```bash
# Run k6 test
k6 run load-test.js
```

### Performance Benchmarks

```bash
# Run Go benchmarks
go test -bench=. -benchmem ./internal/...

# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Profile memory usage
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Testing Best Practices

### 1. Test Naming

```go
// Good
func TestConfigLoader_LoadFile_Success(t *testing.T)
func TestRouter_RegisterEndpoint_EmptyPath_ReturnsError(t *testing.T)

// Bad
func TestLoader(t *testing.T)
func Test1(t *testing.T)
```

### 2. Table-Driven Tests

```go
func TestServerConfig_GetPort(t *testing.T) {
    tests := []struct {
        name     string
        port     int
        expected int
    }{
        {"default zero", 0, 8080},
        {"custom port", 3000, 3000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := ServerConfig{Port: tt.port}
            got := cfg.GetPort()
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### 3. Use Subtests

```go
func TestEndpoints(t *testing.T) {
    t.Run("GET", func(t *testing.T) {
        // GET tests
    })

    t.Run("POST", func(t *testing.T) {
        // POST tests
    })
}
```

### 4. Clean Up Resources

```go
func TestWithTempFile(t *testing.T) {
    tmpDir := t.TempDir()  // Automatically cleaned up

    file := filepath.Join(tmpDir, "test.toml")
    // ... use file
}
```

### 5. Mock External Dependencies

```go
type MockClient struct {
    Response string
    Err      error
}

func (m *MockClient) Get(url string) (string, error) {
    return m.Response, m.Err
}
```

## Continuous Testing

### Pre-commit Hooks

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running tests..."
go test ./internal/... || exit 1

echo "Running linter..."
golangci-lint run || exit 1
```

### Watch Mode

```bash
# Install file watcher
go install github.com/cosmtrek/air@latest

# Run in watch mode
air
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./internal/router/...
```

### Run Specific Test

```bash
go test -run TestRegisterEndpoint ./internal/router/
```

### Debug with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/router/ -- -test.run TestRegisterEndpoint
```

## Troubleshooting

### Common Issues

**Issue**: Integration tests fail with connection refused
```bash
# Solution: Ensure server is running
go run ./cmd/server -config ./examples &
sleep 2  # Wait for startup
go test -tags=integration ./test/integration/...
```

**Issue**: Coverage not generated
```bash
# Solution: Ensure coverage flag is used
go test -coverprofile=coverage.out ./internal/...
```

**Issue**: Race conditions detected
```bash
# Solution: Run with race detector
go test -race ./internal/...
```

## Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Framework](https://github.com/stretchr/testify)
- [HTTP Testing in Go](https://golang.org/pkg/net/http/httptest/)
