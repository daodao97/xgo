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
- `go test -v ./package_name -count=1` - Run tests without cache

### Development
- `go mod tidy` - Clean up dependencies
- `go mod download` - Download dependencies
- `go fmt ./...` - Format all code
- `go vet ./...` - Run static analysis

### Admin UI (for xadmin development)
- `cd xadmin/adminui && pnpm install` - Install frontend dependencies
- `cd xadmin/adminui && pnpm dev` - Start development server
- `cd xadmin/adminui && pnpm build` - Build for production
- `cd xadmin && ./admin_ui.sh` - Copy built assets to ui/ directory for embedding

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
- **xnotify**: Multi-provider notification system (Lark/Feishu, WeWork)

### Utility Modules
- **xcode**: Structured error/code type for REST API responses
- **xjson**: Fluent JSON path navigation and manipulation (wraps gjson/sjson)
- **xtype**: Type system extensions, collection operations, and advanced JSON binding
- **xutil**: Safe goroutine management, retry logic, and stack trace utilities
- **xenv**: Environment detection helpers (dev/prod)
- **xctx**: Type-safe context key definitions
- **utils**: General purpose helpers (JSON serialization, URL validation, etc.)

### Database Layer (xdb)
- Model-based ORM with chainable query builder
- Multi-database support: MySQL, PostgreSQL, SQLite
- Automatic placeholder conversion (`?` → `$1, $2` for PostgreSQL)
- Built-in caching with automatic invalidation
- Hook system for data transformation
- Relationship support (hasOne, hasMany)
- Transaction support with rollback
- Validation system
- Soft delete support via fake delete keys
- UPSERT operations (InsertOrUpdate, InsertIgnore) with cross-database compatibility

### Application Framework (xapp)
- Server lifecycle management
- Graceful shutdown handling with signal processing
- Multi-server support with goroutine management
- Signal handling with force-exit protection (double SIGINT/SIGTERM)
- Startup hook system for initialization tasks
- Command-line argument parsing via go-flags with env var support

### Admin System (xadmin)
- Schema-driven CRUD operations
- Customizable model hooks (BeforeCreate, AfterList, etc.)
- Vue.js-based frontend with Monaco editor
- Dynamic form generation
- Filtering and pagination support

### Type System (xtype)
- Fluent collection operations: `ArrStr`, `ArrMap`, `ArrInt64`, `MapStr`
- Smart type conversions and JSON binding with fuzzy decoding
- JSON-with-comments support via `JsonStrRemoveComments()`
- Template substitution with `{{variable}}` syntax in JSON strings
- `OrderedArrMap` for preserving JSON field order during marshaling

### JSON Utilities (xjson)
- Chainable JSON navigation using path syntax (powered by gjson)
- Type-safe value extraction: `Get("path.to.field").String()`
- In-place modifications: `Set("path", value)`
- Special conversions: `TimeByFormat()`, `Decimal()`, `Array()`, `Map()`

### Notification System (xnotify)
- Provider-based architecture for multiple notification backends
- Built-in support for Lark/Feishu and WeWork webhooks
- URL format: `lark://bot_id@mention1,mention2` or `wework://bot_id`
- Extensible via `RegisterProvider()` for custom integrations

### Safe Async Operations (xutil)
- `Go()` - Goroutine with automatic panic recovery and logging
- `GoWithCancel()` - Cancellable goroutines with context support
- `Retry[T]()` - Generic retry logic with configurable attempts and delays
- `Stack()` - Formatted stack traces for debugging

## Common Patterns

### xrequest HTTP Client Usage
```go
// Basic request with retry and timeout
resp, err := xrequest.New().
    SetTimeout(10 * time.Second).
    SetRetry(3, time.Second*2).
    SetHeaders(map[string]string{
        "Authorization": "Bearer token",
    }).
    Get("https://api.example.com/data")

// SSE (Server-Sent Events) streaming
resp, err := xrequest.New().Get("https://api.example.com/stream")
eventChan, _ := resp.SSE()
for event := range eventChan {
    fmt.Println("Event:", event)
}

// Relay/proxy upstream response (streaming)
totalBytes, err := resp.ToHttpResponseWriteV2(ginCtx.Writer, func(data []byte) (bool, []byte) {
    // Optional hook for processing chunks
    return true, data
})
```

### Database Initialization
```go
xdb.Init(xdb.Config{
    "default": {
        Driver: "mysql", // or "postgres", "sqlite"
        DSN:    "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True",
    },
})
```

### Model Usage
```go
m := xdb.New("users", xdb.WithConnection("default"))
user, err := m.FindById("123")
```

### Query Conditions (xdb)
```go
// Basic conditions
xdb.WhereEq("field", value)      // field = ?
xdb.WhereNotEq("field", value)   // field != ?
xdb.WhereGt("field", value)      // field > ?
xdb.WhereGte("field", value)     // field >= ?
xdb.WhereLt("field", value)      // field < ?
xdb.WhereLte("field", value)     // field <= ?
xdb.WhereIn("field", []any{1,2}) // field IN (?, ?)
xdb.WhereLike("field", "%val%")  // field LIKE ?
xdb.WhereBetween("f", 1, 10)     // f BETWEEN ? AND ?
xdb.WhereIsNil("field")          // field IS NULL
xdb.WhereNotNil("field")         // field IS NOT NULL

// OR conditions
xdb.WhereOrEq("status", 1)
xdb.WhereGroup(xdb.WhereEq("a", 1), xdb.WhereOrEq("b", 2)) // (a = ? OR b = ?)

// Sorting and pagination
xdb.OrderByDesc("created_at")
xdb.OrderByAsc("id")
xdb.Limit(10)
xdb.Offset(20)
xdb.Pagination(page, pageSize)
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

### JSON Navigation
```go
j := xjson.New(data)
name := j.Get("user.profile.name").String()
age := j.Get("user.age").Int64()
j.Set("user.status", "active")
```

### Collection Operations
```go
names := xtype.ArrStr{"alice", "bob", "alice"}
unique := names.Unique() // ["alice", "bob"]
upper := names.Map(func(s string) string { return strings.ToUpper(s) })
```

### Safe Async Execution
```go
xutil.Go(ctx, func() {
    // This will recover from panics and log errors
    doWork()
})

result, err := xutil.Retry(ctx, func(ctx context.Context) (Data, error) {
    return fetchData()
}, xutil.WithMaxRetries(5), xutil.WithDelay(2*time.Second))
```

### Notifications
```go
err := xnotify.Notify(ctx, "lark://bot_abc123@user1,user2", "Hello World")
```

### Rate Limiting
```go
// Sliding window rate limiter
l := limiter.NewSlidingWindow(redisClient, "api_limit", 100, time.Minute)
allowed, err := l.Allow(ctx, userID)

// Concurrency limiter
cl := limiter.NewConcurrency(redisClient, "concurrent_jobs", 10)
release, err := cl.Acquire(ctx, jobID)
defer release()
```

## Dependencies

The project uses standard Go modules with key dependencies:
- Gin for HTTP routing
- GORM-style database operations (custom implementation)
- Redis for caching and queues
- JWT libraries for authentication (HMAC & RSA)
- Vue.js 3 + Element Plus for admin frontend
- @okiss/oms, @okiss/vbtf for admin UI components
- Testify for testing assertions
- Cron v3 for job scheduling
- gjson/sjson for JSON path operations
- Decimal for precise decimal arithmetic

## Key Architectural Patterns

### Server Interface Design
All servers in xapp must implement the `Server` interface:
```go
type Server interface {
    Start() error  // Blocking call that runs the server
    Stop()         // Graceful shutdown
}
```
The App orchestrates multiple servers concurrently with proper lifecycle management.

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
- Frontend assets can be embedded or served from filesystem

### Provider Pattern (xnotify)
The notification system uses a registry pattern:
- Providers are registered via `RegisterProvider(scheme, sender)`
- Built-in providers for "lark" and "wework" schemes
- URL-based routing: `scheme://botid@mentions`
- Extensible for custom notification backends