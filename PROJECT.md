# Bland Mock API - Technical Overview

## Summary

Bland Mock API is a lightweight, configurable mock server for REST and GraphQL APIs. Built with Go and designed for simplicity, it provides a practical solution for development, testing, and CI/CD environments where you need reliable mock endpoints without spinning up complex infrastructure.

## Design Philosophy

The project follows a few core principles:

1. **Configuration over code** - Define APIs via TOML files rather than writing code
2. **Minimal dependencies** - Lean heavily on Go's standard library
3. **Deployment flexibility** - Run locally, in Docker, or on AWS without modification
4. **Extensibility** - Easy to add new endpoint types or response behaviors

## Core Components

### REST Engine
- Dynamic route registration from TOML configuration
- Support for all standard HTTP methods
- Multiple methods per path
- Configurable status codes, headers, and response bodies
- Simulated delays for testing timeout scenarios
- Basic template variable substitution

### GraphQL Support
- Schema definition via TOML
- Type, query, and mutation configuration
- Static response resolution (sufficient for most mock scenarios)

### Configuration System
- TOML-based configuration files
- Single file or directory of files (auto-merged)
- Hot-reloadable in future versions
- Environment-specific configurations

## Testing Strategy

### Unit Tests
- Package-level test coverage for all internal packages
- Table-driven tests for multiple scenarios
- HTTP handler testing with `httptest` package
- Configuration loading and validation tests

Current coverage:
- `internal/config`: 67.4%
- `internal/models`: 100.0%
- `internal/router`: 86.7%

### Integration Tests
- Full workflow testing with real HTTP server
- REST endpoint validation
- GraphQL query and mutation testing
- Error handling and edge cases
- Build tag separation (`-tags=integration`)

### CI Pipeline
- Automated testing on all commits
- Coverage tracking and reporting
- Docker image building
- Multi-platform support

## Deployment Architecture

### Local/Docker
Standard HTTP server listening on configurable port. Suitable for local development and Docker-based deployments.

### AWS Lambda
Packaged as Docker image with Lambda runtime adapter. Includes API Gateway integration and Function URL support. Cold start optimized.

### AWS ECS (Fargate)
Containerized deployment with Application Load Balancer. Auto-scaling based on CPU/memory metrics. Suitable for consistent traffic patterns.

### AWS App Runner
Managed container service with auto-scaling. Simple HTTPS endpoint with minimal configuration. Good balance of simplicity and control.

## Infrastructure as Code

CDK stacks written in TypeScript for all deployment targets:

- **Lambda Stack**: Function, API Gateway, CloudWatch logs
- **ECS Stack**: Cluster, service, ALB, VPC, auto-scaling
- **App Runner Stack**: Service, auto-scaling configuration, IAM roles

Each stack includes health checks, monitoring, and follows AWS best practices.

## Cost Considerations

Approximate monthly costs for moderate usage:

- **Lambda**: $5-10 (pay-per-use, ideal for variable traffic)
- **ECS**: $45-50 (2 tasks, predictable costs)
- **App Runner**: $15-20 (managed, scales to zero when idle)

## CI/CD Integration

### GitHub Actions
- Unit and integration tests
- Docker builds
- Coverage reporting
- Automated releases

### TeamCity
- Kotlin DSL configuration
- Build pipeline with dependencies
- AWS deployment automation
- Post-deployment validation

## Technical Stack

**Core**
- Go 1.25
- Standard library HTTP server
- TOML configuration parsing
- GraphQL library (minimal external dependency)

**Infrastructure**
- AWS CDK (TypeScript)
- Docker multi-stage builds
- ECR for container images

**Tooling**
- Make for build automation
- golangci-lint for code quality
- GitHub Actions and TeamCity for CI/CD

## Performance Characteristics

- Sub-second startup time
- ~20MB memory footprint (base)
- <5ms response latency (without configured delays)
- Handles 10k+ requests/second in basic benchmarks
- 15MB binary size (uncompressed)
- 25MB Docker image (Alpine-based)

## Security Model

- Runs as non-root in containers
- Minimal Alpine Linux base image
- No authentication by design (it's a mock)
- Configurable CORS support
- Environment variable configuration for sensitive values

## Extending the System

The architecture supports several extension points:

1. **New endpoint types**: Add handlers in `internal/router`
2. **Response generators**: Extend template variables in `processResponse`
3. **Configuration sources**: Implement new loaders in `internal/config`
4. **Deployment targets**: Add new CDK stacks in `cdk/lib`

## Project Structure

```
blandmockapi/
├── cmd/server/           # Entry points
├── internal/             # Core logic
│   ├── config/          # Configuration loading
│   ├── models/          # Data structures
│   ├── router/          # HTTP routing
│   └── graphql/         # GraphQL handling
├── test/                # Test suites
├── cdk/                 # Infrastructure code
├── examples/            # Sample configurations
└── scripts/             # Utility scripts
```

## Documentation

- **README.md**: Quick start and basic usage
- **TESTING.md**: Testing strategy and commands
- **DEPLOYMENT.md**: AWS deployment details
- **CONTRIBUTING.md**: Development guidelines
- **cdk/README.md**: Infrastructure documentation

## Use Cases

This tool has proven useful in several scenarios:

- **API contract testing**: Validate client behavior against expected responses
- **Development environments**: Mock external APIs during local development
- **CI/CD pipelines**: Provide consistent mock endpoints for automated tests
- **Frontend development**: Prototype UIs before backend implementation
- **Load testing**: Test client behavior under various response scenarios
- **Training/demos**: Demonstrate API integration without complex setup

## Known Limitations

- Path parameters are basic (exact match or prefix)
- GraphQL responses are static (no dynamic data generation)
- No built-in authentication (use reverse proxy if needed)
- State is not persisted between requests
- Template variables are simple string replacement

These limitations are intentional - the tool focuses on being a simple, reliable mock rather than a full API simulation platform.

## Future Considerations

The roadmap includes potential enhancements based on actual usage patterns:

- OpenAPI/Swagger import for automatic endpoint generation
- Web UI for runtime configuration
- Request/response logging and replay
- Dynamic data generation with faker libraries
- WebSocket endpoint support
- Prometheus metrics export
- More sophisticated routing with path variables

## Getting Started

Clone the repository and run locally:

```bash
git clone https://github.com/yourusername/blandmockapi.git
cd blandmockapi
go run ./cmd/server -config ./examples
```

Or use Docker:

```bash
docker build -t blandmockapi .
docker run -p 8080:8080 blandmockapi
```

For AWS deployment, see DEPLOYMENT.md.

## License

MIT License - see LICENSE file for details.

## Maintenance

This project follows semantic versioning. Releases are tagged in git and include:
- Binary artifacts for major platforms
- Docker images pushed to registry
- CDK CloudFormation templates
- Changelog with migration notes

Issues and pull requests are welcome. See CONTRIBUTING.md for guidelines.
