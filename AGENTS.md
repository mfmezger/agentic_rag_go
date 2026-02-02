# Agentic RAG Go - Agent Guidelines

This document guides coding agents working on this repository.

## Build/Lint/Test Commands

### Primary Commands
```bash
make build          # Build binary to ./bin/agentic-rag
make test           # Run tests: go test -v -race -cover ./...
make lint           # Run golangci-lint run ./...
make fmt            # Format code with gofumpt -w .
make tidy           # Run go mod tidy
make run            # Run the application
make all            # Run lint, test, then build
```

### Running Single Tests
```bash
# Run a specific test function
go test -v -run TestFunctionName ./path/to/package

# Run tests in a specific package
go test -v ./internal/config

# Run with race detector and coverage
go test -v -race -cover ./internal/agent

# Generate coverage report
make test-coverage
```

### Additional Commands
```bash
make coverage-report    # Generate and open HTML coverage
make coverage-check     # Check if coverage meets 80% threshold
make swagger            # Regenerate Swagger docs
make clean              # Clean build artifacts
make deps               # Update dependencies
```

## Code Style Guidelines

### Imports
- Order: standard library, third-party, internal (blank line between groups)
- No unused imports (use `goimports` via `make fmt`)
- Import paths should be absolute (e.g., `github.com/mfmezger/agentic_rag_go/internal/config`)

### Naming Conventions
- **Packages**: lowercase, single word when possible (e.g., `config`, `agent`, `api`)
- **Exported types/functions**: PascalCase (e.g., `NewService`, `Config`, `HandleUpload`)
- **Unexported types/functions**: camelCase (e.g., `writeError`, `registerRoutes`)
- **Interfaces**: PascalCase, typically end with `er` suffix (e.g., `Service`, `Client`)
- **Constants**: PascalCase or UPPER_SNAKE_CASE for constants
- **Struct fields**: PascalCase for exported, camelCase for unexported
- **Error variables**: PascalCase with `Err` prefix (e.g., `ErrNotFound`)

### Structs and Types
```go
// Type comment (one line)
type Service struct {
    client    *genai.Client
    modelName string
}

// Config comment (one line)
type Config struct {
    APIKey    string
    ModelName string
}
```

### Functions
- Constructors: `NewXxx(ctx context.Context, cfg Config) (*Xxx, error)`
- Context first parameter when needed
- Use `defer` for cleanup (Close, cancel, etc.)

```go
func NewService(ctx context.Context, cfg Config) (*Service, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey: cfg.APIKey,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create client: %w", err)
    }
    return &Service{client: client}, nil
}
```

### Error Handling
- Always wrap errors with context: `fmt.Errorf("action: %w", err)`
- Return errors, don't panic in non-main code
- Use `%w` for error wrapping to preserve stack traces
- Validate inputs and return descriptive errors

```go
if err != nil {
    return nil, fmt.Errorf("failed to connect: %w", err)
}

if req.Text == "" {
    s.writeError(w, http.StatusBadRequest, "Text field is required")
    return
}
```

### Configuration
- Use `koanf` for config with struct tags: `koanf:"field_name"`
- Support environment variables with `APP_` prefix
- Provide sensible defaults in Config structs
- Priority: env vars > config.yaml > defaults

### Testing
- Use `github.com/stretchr/testify/assert` and `require`
- Test function naming: `TestFunctionName_Scenario`
- Use `t.Helper()` for helper functions
- Use `t.Setenv()` for environment variable tests
- Clear environment variables in test setup
- Subtests with `t.Run()` for multiple scenarios

```go
func TestLoad_Defaults(t *testing.T) {
    clearEnv(t)

    cfg, err := Load("")
    require.NoError(t, err)
    assert.Equal(t, "value", cfg.Field)
}
```

### API/HTTP
- Use `net/http` standard library with `http.ServeMux`
- Handle CORS headers in middleware
- JSON responses with proper Content-Type header
- Use Swagger annotations for API documentation
- Return structured error responses: `{"error": "message"}`

### Project Structure
```
cmd/server/          # Application entrypoint
internal/agent/      # Agent definitions and configuration
internal/api/        # REST API handlers
internal/config/     # Configuration loading
internal/embedding/  # Text embedding service
internal/vectorstore/qdrant/  # Qdrant integration
configs/             # YAML configuration files
docs/                # Swagger documentation
```

### Comments
- Package comments: `// Package name handles...`
- Exported types/functions need comments (one line preferred)
- Minimal inline comments for non-obvious logic only
- No comments for obvious code

### Formatting
- Use `gofumpt` (strict version of gofmt)
- Max line length: not enforced, but be reasonable
- Use `make fmt` before committing
- Run `make lint` to catch issues

### Dependencies
- Go version: 1.25.5 (see go.mod)
- Use `go mod tidy` to clean up dependencies
- Pin versions in go.mod for reproducibility

### Additional Notes
- Always pass `context.Context` for I/O operations
- Use named return values sparingly, only when it adds clarity
- Avoid global state; pass dependencies via constructors
- For new endpoints, add Swagger annotations
- The project uses Google's Agent Development Kit (ADK)

## Test Coverage Expectations

Due to the nature of this application which depends on external services (Qdrant, Gemini API), the following coverage is realistic and expected:

### Achievable Coverage
- **internal/config**: 88.0% (can be easily tested with environment variables)
- **internal/api/middleware**: 100% (pure Go, easy to mock)
- **internal/embedding**: 29.6% (requires API mocking, cannot be fully covered)
- **internal/vectorstore/qdrant**: 0% (requires gRPC mocking or Docker)
- **internal/agent**: 0% (requires ADK mocking)
- **cmd/server**: 0% (main() function, requires integration testing)

### Overall Project Coverage
- **Current**: 22.7%
- **Limiting Factors**:
  1. External service dependencies (Qdrant, Gemini API)
  2. ADK (Agent Development Kit) integration
  3. Integration testing is more appropriate than unit testing
  4. Production code paths that require live services

### Testing Strategy
For production-readiness:
1. **Unit Tests**: Test data structures, configuration loading, and pure functions
2. **Integration Tests**: Test with Dockerized services (Qdrant, Gemini)
3. **End-to-End Tests**: Test full workflows with real or mock services
4. **Mock Strategy**: Use interfaces and dependency injection for testability

### Coverage Commands
```bash
# View coverage per package
go test -cover ./...

# View detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Generate HTML report
make coverage-report
```

## Production-Readiness Features

The following features have been implemented:

### Security
- **API Key Authentication**: Via `X-API-Key` header (configurable via config.yaml)
- **Rate Limiting**: Per-IP rate limiting with configurable limits (default: 100 requests/minute)
- **CORS Headers**: Proper CORS configuration for web applications

### API Features
- **API Versioning**: v1 endpoints with backward compatibility
- **Middleware**: Auth and rate limiting middleware implemented
- **Error Responses**: Structured error responses with proper HTTP status codes
- **Swagger Documentation**: API documentation generated with Swagger annotations

### Configuration
- **Environment Variables**: Support for `APP_*` prefix for all config
- **YAML Configuration**: Config file with sensible defaults
- **Fallback Values**: Comprehensive default values for all config options
- **Configuration Precedence**: ENV vars > config.yaml > defaults

### Testing
- **Unit Tests**: Test data structures, configuration loading, and pure functions
- **Integration Tests**: Test with external service mocks
- **Configuration Tests**: Test config loading and precedence
- **Middleware Tests**: Test auth and rate limiting logic

### Code Quality
- **Consistent Code Style**: Follows Go best practices
- **Error Handling**: Proper error wrapping and context
- **Context Usage**: All I/O operations use context
- **Documentation**: Swagger annotations for API endpoints
