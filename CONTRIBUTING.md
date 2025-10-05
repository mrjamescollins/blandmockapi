# Contributing to Bland Mock API

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)

## Code of Conduct

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community

**Unacceptable behavior includes:**
- Trolling, insulting/derogatory comments, and personal attacks
- Public or private harassment
- Publishing others' private information without permission

## Getting Started

### Prerequisites

- Go 1.25 or later
- Docker and Docker Compose
- AWS CLI (for deployments)
- Node.js 20+ (for CDK)
- Make (optional)

### Setting Up Development Environment

1. **Fork and Clone**

```bash
# Fork on GitHub first, then clone your fork
git clone https://github.com/YOUR_USERNAME/blandmockapi.git
cd blandmockapi
```

2. **Install Dependencies**

```bash
# Go dependencies
go mod download

# CDK dependencies
cd cdk && npm install && cd ..
```

3. **Run Tests**

```bash
# Unit tests
go test ./internal/... -v

# Integration tests
go test -tags=integration ./test/integration/... -v
```

4. **Run Locally**

```bash
go run ./cmd/server -config ./examples
```

## Development Workflow

### Branch Naming

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `test/` - Test additions/improvements
- `refactor/` - Code refactoring

Examples:
- `feature/add-websocket-support`
- `fix/graphql-mutation-error`
- `docs/update-deployment-guide`

### Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Examples:**

```
feat(graphql): add support for GraphQL subscriptions

Implement WebSocket-based subscriptions for real-time updates.
Includes TOML configuration for subscription endpoints.

Closes #42
```

```
fix(router): handle concurrent requests to same endpoint

Fixed race condition when multiple methods are registered
on the same path simultaneously.
```

## Testing Guidelines

### Writing Tests

1. **Unit Tests**
   - Test individual functions and methods
   - Use table-driven tests for multiple scenarios
   - Mock external dependencies
   - Place in `*_test.go` files alongside code

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"basic case", "input", "output"},
        {"edge case", "", "default"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

2. **Integration Tests**
   - Test complete workflows
   - Use build tags: `// +build integration`
   - Place in `test/integration/`

3. **Test Coverage**
   - Aim for >80% coverage
   - Critical paths should have >90%
   - Run: `go test ./internal/... -cover`

### Running Tests

```bash
# All tests
./scripts/test.sh

# Unit tests only
go test ./internal/... -v

# With coverage
go test ./internal/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Integration tests
go test -tags=integration ./test/integration/... -v

# Specific package
go test ./internal/router/... -v

# Specific test
go test -run TestRegisterEndpoint ./internal/router/... -v
```

## Pull Request Process

### Before Submitting

1. **Update Documentation**
   - Update README if adding features
   - Add/update code comments
   - Update TOML examples if needed

2. **Run Tests**
   ```bash
   ./scripts/test.sh
   ```

3. **Format Code**
   ```bash
   go fmt ./...
   gofmt -s -w .
   ```

4. **Lint Code**
   ```bash
   golangci-lint run
   ```

5. **Update CHANGELOG**
   - Add entry under "Unreleased" section
   - Follow Keep a Changelog format

### Submitting PR

1. **Push to Your Fork**
   ```bash
   git push origin feature/my-feature
   ```

2. **Create Pull Request**
   - Use a clear, descriptive title
   - Reference related issues
   - Describe changes in detail
   - Include testing steps

3. **PR Template**

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manually tested

## Checklist
- [ ] Code follows project style
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] CHANGELOG updated
```

### Review Process

1. **Automated Checks**
   - GitHub Actions CI must pass
   - Code coverage should not decrease

2. **Code Review**
   - At least one approval required
   - Address all review comments
   - Keep discussion professional

3. **Merge**
   - Squash commits for cleaner history
   - Ensure CI passes before merge

## Code Style

### Go Style Guide

Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key points:**

1. **Naming**
   ```go
   // Good
   var userCount int
   func GetUserByID(id int) User

   // Bad
   var usrcnt int
   func get_user(id int) User
   ```

2. **Error Handling**
   ```go
   // Good
   result, err := DoSomething()
   if err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }

   // Bad
   result, _ := DoSomething() // Don't ignore errors
   ```

3. **Comments**
   ```go
   // Good
   // ProcessRequest handles incoming HTTP requests and routes them
   // to the appropriate handler based on the configured endpoints.
   func ProcessRequest(w http.ResponseWriter, r *http.Request) {

   // Bad
   // process request
   func ProcessRequest(w http.ResponseWriter, r *http.Request) {
   ```

4. **Formatting**
   - Use `gofmt` and `goimports`
   - Line length: aim for <120 characters
   - One blank line between functions

### TypeScript/CDK Style

1. **Use TypeScript strict mode**
2. **Follow AWS CDK best practices**
3. **Use meaningful variable names**
4. **Add JSDoc comments for public APIs**

### TOML Configuration

1. **Use clear, descriptive keys**
2. **Group related settings**
3. **Add comments for complex configurations**

```toml
# Server Configuration
[server]
port = 8080              # HTTP port
host = "0.0.0.0"        # Listen on all interfaces

# REST API Endpoints
[[endpoints]]
path = "/api/users"     # User management endpoint
method = "GET"
status = 200
```

## Documentation

### Code Documentation

- Public functions must have comments
- Complex logic needs explanation
- Use examples where helpful

```go
// LoadFromPath loads configuration from a file or directory.
// If path is a directory, all .toml files are loaded and merged.
//
// Example:
//   loader := config.New()
//   err := loader.LoadFromPath("./configs")
func (l *Loader) LoadFromPath(path string) error {
```

### README Updates

When adding features:
1. Update main README.md
2. Add examples to examples/
3. Update relevant guides (TESTING.md, DEPLOYMENT.md)

## Issue Reporting

### Bug Reports

Include:
- Clear title and description
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, Go version, etc.)
- Relevant logs/errors

### Feature Requests

Include:
- Use case description
- Proposed solution
- Alternative approaches considered
- Willingness to contribute

## Release Process

### Version Numbers

Follow [Semantic Versioning](https://semver.org/):
- MAJOR.MINOR.PATCH
- Example: 1.2.3

### Creating a Release

1. Update CHANGELOG.md
2. Update version in code
3. Create git tag
4. Push tag to trigger release

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

## Getting Help

- **Documentation**: Check README, TESTING.md, DEPLOYMENT.md
- **Issues**: Search existing issues
- **Discussions**: Use GitHub Discussions for questions
- **Contact**: Open an issue for bugs/features

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Credited in commit history

Thank you for contributing! 🎉
