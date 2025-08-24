# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

XGO is a comprehensive Go framework providing modular components for web development, database operations, caching, logging, and admin interfaces. The project follows a modular architecture with each package (`x*`) providing specific functionality.

## Common Commands

### Build and Test
- `go build ./...` - Build all packages
- `go test ./...` - Run all tests
- `go test -v ./package_name` - Run tests for specific package with verbose output
- `go test ./package_name -run TestName` - Run specific test

### Development
- `go mod tidy` - Clean up dependencies
- `go mod download` - Download dependencies
- `go fmt ./...` - Format all code
- `go vet ./...` - Run static analysis

### Admin UI (for xadmin development)
- `cd xadmin/adminui && npm install` - Install frontend dependencies
- `cd xadmin/adminui && npm run dev` - Start development server
- `cd xadmin/adminui && npm run build` - Build for production

### Package Testing
- `go test ./xdb -v` - Test database package
- `go test ./xjwt -v` - Test JWT package
- `go test ./limiter -v` - Test limiter package
- `go test ./xqueue -v` - Test queue package

## Architecture

### Core Framework Components
- **xapp**: Application lifecycle management with graceful startup/shutdown
- **xhttp**: HTTP server utilities with context, middleware, and response helpers
- **xdb**: Database ORM with MySQL/PostgreSQL support, caching, hooks, and validation
- **xlog**: Structured logging with color/pretty output handlers
- **xjwt**: JWT authentication with HMAC and RSA support

### Specialized Modules
- **xadmin**: Auto-generated CRUD admin interface with Vue.js frontend
- **cache**: Multi-backend caching interfaces (memory, Redis)
- **xqueue**: Message queue implementation (Redis-based)
- **limiter**: Rate limiting with sliding window and concurrency controls
- **xproxy**: HTTP proxy and static file serving
- **xredis**: Redis utilities and connection management
- **xcron**: Cron job scheduling with logging support
- **xtrace**: Request tracing utilities for Gin
- **xrequest**: HTTP client with retry and proxy support

### Database Layer (xdb)
- Model-based ORM with chainable query builder
- Built-in caching with automatic invalidation
- Hook system for data transformation
- Relationship support (hasOne, hasMany)
- Transaction support with rollback
- Validation system
- Soft delete support via fake delete keys

### Application Framework (xapp)
- Server lifecycle management
- Graceful shutdown handling
- Multi-server support with goroutine management
- Signal handling with force-exit protection
- Startup hook system

### Admin System (xadmin)
- Schema-driven CRUD operations
- Customizable model hooks (BeforeCreate, AfterList, etc.)
- Vue.js-based frontend with Monaco editor
- Dynamic form generation
- Filtering and pagination support

## Development Guidelines

### File Organization
- Keep Go files under 250 lines
- Limit directories to 8 files maximum
- Use proper package naming (`x` prefix for framework components)

### Database Models (xdb)
- Use `xdb.New("table_name")` for basic models
- Implement custom model functions via `NewModel` option
- Utilize hooks for data transformation and validation
- Leverage caching for frequently accessed data

### HTTP Handlers (xhttp)
- Use context for request-scoped data
- Implement proper error handling with structured responses
- Utilize middleware for cross-cutting concerns

### Admin Interface (xadmin)
- Define schemas for automatic CRUD generation
- Use hooks for custom business logic
- Implement proper validation in BeforeCreate/BeforeUpdate hooks

## Testing Patterns

- Use `testify` for assertions (`github.com/stretchr/testify`)
- Place tests alongside implementation files (`*_test.go`)
- Use table-driven tests for multiple scenarios
- Mock external dependencies appropriately

## Common Patterns

### Model Usage
```go
m := xdb.New("users", xdb.WithConnection("default"))
user, err := m.FindById("123")
```

### Application Setup
```go
app := xapp.NewApp()
app.AddStartup(initDatabase)
app.AddServer(func() xapp.Server { return httpServer })
app.Run()
```

### Admin Schema Definition
```go
xadmin.Register("users", &xadmin.Crud{
    NewModel: func(r *http.Request) xdb.Model {
        return xdb.New("users")
    },
    BeforeCreate: func(r *http.Request, record xdb.Record) (xdb.Record, error) {
        // Custom validation/transformation
        return record, nil
    },
})
```

## Dependencies

The project uses standard Go modules with key dependencies:
- Gin for HTTP routing
- GORM-style database operations (custom implementation)
- Redis for caching and queues
- JWT libraries for authentication (HMAC & RSA)
- Vue.js 3 + Element Plus for admin frontend
- Testify for testing assertions
- Cron v3 for job scheduling

## Key Architectural Patterns

### Model Interface Design
The xdb package uses a unified Model interface that supports:
- Chainable query building with fluent API
- Context-aware operations for request tracing
- Transaction support with automatic rollback
- Built-in caching layer with Redis backing
- Hook system for data validation and transformation

### Admin System Architecture
- Schema-driven approach using JSON configuration
- Hook-based customization (BeforeCreate, AfterList, etc.)
- Vue.js SPA with Monaco editor integration
- RESTful API auto-generation from model definitions