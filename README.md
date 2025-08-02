# PyAirtable Go Shared Library

A comprehensive shared library for Go microservices in the PyAirtable ecosystem, providing common packages for configuration, database management, caching, middleware, error handling, logging, metrics, health checks, and utilities.

[![CI](https://github.com/Reg-Kris/pyairtable-go-shared/actions/workflows/ci.yml/badge.svg)](https://github.com/Reg-Kris/pyairtable-go-shared/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Reg-Kris/pyairtable-go-shared)](https://goreportcard.com/report/github.com/Reg-Kris/pyairtable-go-shared)
[![GoDoc](https://godoc.org/github.com/Reg-Kris/pyairtable-go-shared?status.svg)](https://godoc.org/github.com/Reg-Kris/pyairtable-go-shared)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🔧 **Configuration Management** - Viper-based configuration with environment variable support
- 🗄️ **Database Integration** - GORM-based database utilities with connection pooling
- 🚀 **Redis Caching** - Redis client with circuit breaker pattern
- 🛡️ **Middleware Suite** - Authentication, logging, metrics, and rate limiting
- ❌ **Error Handling** - Standardized error types and responses
- 📝 **Structured Logging** - Zap-based logging with context support
- 📊 **Metrics Collection** - Prometheus metrics helpers
- 🏥 **Health Checks** - Comprehensive health check handlers
- 🛠️ **Common Utilities** - String manipulation, crypto, JSON, and time utilities
- 📋 **Shared Models** - User, Tenant, Workspace, and common DTOs
- 🧪 **Testing Utilities** - Test database setup, fixtures, and mocks

## Installation

```bash
go get github.com/Reg-Kris/pyairtable-go-shared
```

## Quick Start

### Configuration

```go
package main

import (
    "log"
    
    "github.com/Reg-Kris/pyairtable-go-shared/config"
)

func main() {
    cfg, err := config.Load("")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    if cfg.IsDevelopment() {
        log.Println("Running in development mode")
    }
}
```

### Database

```go
package main

import (
    "log"
    
    "github.com/Reg-Kris/pyairtable-go-shared/config"
    "github.com/Reg-Kris/pyairtable-go-shared/database"
    "github.com/Reg-Kris/pyairtable-go-shared/models"
)

func main() {
    cfg, _ := config.Load("")
    
    db, err := database.New(&cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Auto-migrate models
    db.Migrate(&models.User{}, &models.Tenant{}, &models.Workspace{})
    
    // Use generic repository
    userRepo := database.NewRepository[models.User](db)
    user := &models.User{
        Email:     "user@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    if err := userRepo.Create(user); err != nil {
        log.Fatal("Failed to create user:", err)
    }
}
```

### Caching

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/Reg-Kris/pyairtable-go-shared/cache"
    "github.com/Reg-Kris/pyairtable-go-shared/config"
)

func main() {
    cfg, _ := config.Load("")
    
    client, err := cache.New(&cfg.Redis)
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Set a value
    err = client.Set(ctx, "key", "value", 5*time.Minute)
    if err != nil {
        log.Fatal("Failed to set cache:", err)
    }
    
    // Get a value
    var value string
    err = client.Get(ctx, "key", &value)
    if err != nil {
        log.Fatal("Failed to get cache:", err)
    }
    
    log.Println("Cached value:", value)
}
```

### HTTP Server with Middleware

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/Reg-Kris/pyairtable-go-shared/config"
    "github.com/Reg-Kris/pyairtable-go-shared/logger"
    "github.com/Reg-Kris/pyairtable-go-shared/middleware"
    "github.com/Reg-Kris/pyairtable-go-shared/metrics"
    "github.com/Reg-Kris/pyairtable-go-shared/health"
)

func main() {
    cfg, _ := config.Load("")
    log, _ := logger.New(&cfg.Logger)
    metricsRegistry := metrics.New("myservice")
    healthChecker := health.NewChecker()
    
    r := gin.New()
    
    // Add middleware
    r.Use(middleware.RequestLogging(log))
    r.Use(metricsRegistry.Middleware())
    r.Use(middleware.JWT(middleware.AuthConfig{
        JWTSecret: cfg.Auth.JWTSecret,
        SkipPaths: []string{"/health", "/metrics"},
    }))
    
    // Health checks
    r.GET("/health", healthChecker.Handler())
    r.GET("/metrics", gin.WrapH(metricsRegistry.Handler()))
    
    // Protected routes
    api := r.Group("/api/v1")
    api.Use(middleware.RequireRole("user"))
    {
        api.GET("/profile", getProfile)
    }
    
    r.Run(":8080")
}

func getProfile(c *gin.Context) {
    userID := middleware.GetUserIDFromContext(c)
    c.JSON(200, gin.H{"user_id": userID})
}
```

## Package Documentation

### Configuration (`config`)

Provides configuration management using Viper with support for:
- YAML/JSON configuration files
- Environment variables with `PYAIRTABLE_` prefix
- Validation and type safety
- Development/production environment detection

### Database (`database`)

GORM-based database utilities featuring:
- Connection pooling and health checks
- Generic repository pattern
- Transaction support
- Migration helpers
- Connection statistics

### Cache (`cache`)

Redis caching with circuit breaker:
- Automatic failover with circuit breaker pattern
- JSON serialization/deserialization
- Distributed operations support
- Connection health monitoring
- Metrics and statistics

### Middleware (`middleware`)

HTTP middleware collection:
- **JWT Authentication** - Token validation and user context
- **Request Logging** - Structured request/response logging  
- **Rate Limiting** - Token bucket and sliding window algorithms
- **Security Logging** - Security event tracking
- **CORS Support** - Cross-origin request handling

### Error Handling (`errors`)

Standardized error management:
- Structured error types with HTTP status codes
- Error wrapping and context
- Predefined error constructors
- API error response formatting
- Error code constants

### Logging (`logger`)

Zap-based structured logging:
- Context-aware logging with request IDs
- Multiple output formats (JSON, console)
- Log level configuration
- Performance and audit logging helpers
- Global and contextual loggers

### Metrics (`metrics`)

Prometheus metrics integration:
- HTTP request metrics (duration, size, status)
- Database connection and query metrics
- Cache operation metrics  
- Business metrics (users, workspaces, records)
- Custom metrics support
- Gin middleware for automatic collection

### Health Checks (`health`)

Comprehensive health monitoring:
- Database connectivity checks
- Cache connectivity checks
- HTTP endpoint checks
- Memory and goroutine monitoring
- Custom health check registration
- Readiness and liveness probes

### Utilities (`utils`)

Common utility functions:
- **Crypto** - Password hashing, random generation, SHA256
- **Strings** - Case conversion, validation, sanitization
- **JSON** - Marshaling, unmarshaling, path extraction
- **Time** - Formatting, business day calculations, timezone handling

### Models (`models`)

Shared data models:
- **Base Models** - Common fields and interfaces
- **User Management** - Users, roles, permissions, sessions
- **Multi-tenancy** - Tenants, quotas, invitations
- **Workspaces** - Workspaces, tables, fields, records
- **API Responses** - Pagination, filtering, bulk operations

### Testing (`testing`)

Testing utilities and helpers:
- In-memory test database setup
- Test data fixtures and factories
- Mock repositories and services
- Database assertion helpers
- Test suites and cleanup utilities

## Environment Variables

The library supports configuration via environment variables with the `PYAIRTABLE_` prefix:

```bash
# Server Configuration
PYAIRTABLE_SERVER_HOST=localhost
PYAIRTABLE_SERVER_PORT=8080
PYAIRTABLE_SERVER_ENVIRONMENT=development

# Database Configuration
PYAIRTABLE_DATABASE_HOST=localhost
PYAIRTABLE_DATABASE_PORT=5432
PYAIRTABLE_DATABASE_USER=postgres
PYAIRTABLE_DATABASE_PASSWORD=secretpassword
PYAIRTABLE_DATABASE_DATABASE=pyairtable
PYAIRTABLE_DATABASE_SSL_MODE=disable

# Redis Configuration
PYAIRTABLE_REDIS_HOST=localhost
PYAIRTABLE_REDIS_PORT=6379
PYAIRTABLE_REDIS_PASSWORD=
PYAIRTABLE_REDIS_DATABASE=0

# Authentication Configuration
PYAIRTABLE_AUTH_JWT_SECRET=your-secret-key
PYAIRTABLE_AUTH_JWT_EXPIRATION=3600

# Logger Configuration
PYAIRTABLE_LOGGER_LEVEL=info
PYAIRTABLE_LOGGER_FORMAT=json
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

Using the testing utilities:

```go
package myservice

import (
    "testing"
    
    "github.com/Reg-Kris/pyairtable-go-shared/testing"
    "github.com/Reg-Kris/pyairtable-go-shared/models"
)

func TestUserService(t *testing.T) {
    // Set up test database
    testDB := testing.NewTestDB(t)
    defer testDB.Cleanup()
    
    // Migrate models
    testDB.Migrate(&models.User{})
    
    // Create test fixtures
    fixtures := testing.NewTestFixtures()
    user := fixtures.CreateTestUser()
    
    // Seed database
    testDB.Seed(user)
    
    // Test your service
    service := NewUserService(testDB)
    result, err := service.GetUser(user.ID)
    
    // Assertions
    if err != nil {
        t.Fatal(err)
    }
    if result.Email != user.Email {
        t.Errorf("Expected email %s, got %s", user.Email, result.Email)
    }
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write comprehensive tests for new features
- Update documentation for public APIs
- Run `golangci-lint run` before committing
- Ensure all tests pass with `go test ./...`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0
- Initial release with all core packages
- Configuration management with Viper
- Database utilities with GORM
- Redis caching with circuit breaker
- Complete middleware suite
- Structured logging with Zap
- Prometheus metrics integration
- Health check system
- Comprehensive testing utilities

## Support

For questions, issues, or contributions, please:

1. Check the [documentation](https://godoc.org/github.com/Reg-Kris/pyairtable-go-shared)
2. Search existing [issues](https://github.com/Reg-Kris/pyairtable-go-shared/issues)
3. Create a new issue if needed

## Architecture

This shared library follows clean architecture principles:

```
pyairtable-go-shared/
├── config/          # Configuration management
├── database/        # Database utilities and repositories
├── cache/           # Redis caching with circuit breaker
├── middleware/      # HTTP middleware (auth, logging, metrics, etc.)
├── errors/          # Standardized error handling
├── logger/          # Structured logging with Zap
├── metrics/         # Prometheus metrics helpers
├── health/          # Health check handlers
├── utils/           # Common utilities
├── models/          # Shared data models
├── testing/         # Testing utilities and fixtures
└── .github/         # CI/CD workflows
```

Each package is designed to be:
- **Self-contained** - Minimal dependencies between packages
- **Testable** - Comprehensive test coverage with mocks
- **Configurable** - Flexible configuration options
- **Observable** - Built-in logging and metrics
- **Resilient** - Error handling and circuit breakers