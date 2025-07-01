# Anansi - Go Web Framework Helper Library

## Project Overview

**Anansi** is a Go library that provides helper tools and utilities for building web applications with [go-chi](https://github.com/go-chi/chi). It's designed to streamline common web development tasks like session management, API error handling, request/response processing, and more.

- **Language**: Go 1.23.0+ (toolchain go1.24.1)
- **License**: MIT
- **Architecture**: Modular library with focused packages
- **Main Dependencies**: go-chi, zerolog, Redis, JWT, PostgreSQL migration tools

## Directory Structure

```
/Users/arewasmac/code/anansi/
├── api/                    # API utilities and middleware
│   ├── error.go           # HTTP error handling structures
│   ├── middleware.go      # Panic recovery and session middleware
│   ├── request.go         # Request utilities
│   ├── response.go        # Response utilities
│   └── sessions.go        # Session management for API
├── ajax/                  # AJAX utilities
│   ├── ajax.go           # AJAX request handling
│   └── req_id.go         # Request ID management
├── html/                  # HTML/template utilities
│   ├── cookies.go        # Cookie management
│   ├── files.go          # File serving utilities
│   └── template.go       # Template rendering
├── json/                  # JSON processing
│   └── jsoniter.go       # Fast JSON marshaling/unmarshaling
├── jwt/                   # JWT token management
│   └── jwt.go            # JWE token encoding/decoding
├── postgres/              # PostgreSQL utilities
│   ├── errors.go         # Database error handling
│   └── migrate.go        # Database migrations
├── requests/              # HTTP request processing
│   ├── body.go           # Request body parsing
│   ├── cors.go           # CORS handling
│   ├── middleware.go     # Request middleware
│   └── urlencoded.go     # URL-encoded form processing
├── responses/             # HTTP response utilities
│   ├── middleware.go     # Response middleware
│   ├── send.go           # Response sending utilities
│   └── writer.go         # Response writer utilities
├── sessions/              # Session management
│   └── store.go          # Session store implementation
├── tokens/                # Token management
│   └── redis_store.go    # Redis-based token storage
├── webpack/               # Webpack integration
│   └── webpack.go        # Asset management
├── env.go                 # Environment configuration
├── http.go                # HTTP utilities
├── logging.go             # Logging configuration
├── rand.go                # Random number generation
├── sealbox.go             # Encryption utilities
├── server.go              # Server utilities with graceful shutdown
├── slugs.go               # URL slug generation
└── time.go                # Time utilities
```

## Key Features & Components

### 1. Session Management (`sessions/`)
- **Multi-mode sessions**: Cookie, Bearer token, and Headless (API key) authentication
- **Configurable timeouts**: Different timeout settings for different session types
- **Secure cookies**: Production-ready cookie security
- **Token-based sessions**: Redis-backed stateful sessions with HMAC signing

### 2. API Error Handling (`api/`)
- **Structured errors**: Custom `Err` type with HTTP status codes and structured data
- **Panic recovery**: Middleware that handles panics gracefully
- **Environment-aware**: Different error reporting for dev/test vs production
- **Context-aware logging**: Integrated with zerolog for structured logging

### 3. JWT/JWE Support (`jwt/`)
- **JWE encryption**: Uses go-jose for encrypted JWT tokens
- **Custom claims**: Support for custom claim structures
- **Token validation**: Built-in expiration and signature validation
- **Secure by default**: Requires minimum 32-byte secrets

### 4. Request/Response Processing (`requests/`, `responses/`)
- **Body parsing**: JSON and URL-encoded form parsing
- **CORS support**: Built-in CORS middleware
- **Response utilities**: Standardized response sending
- **Middleware chain**: Composable request/response middleware

### 5. Token Storage (`tokens/`)
- **Redis backend**: High-performance token storage
- **Token lifecycle**: Commission, extend, peek, reset, decommission operations
- **HMAC signing**: Secure token generation using HMAC-SHA256
- **TTL support**: Automatic token expiration

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./api
go test ./sessions
go test ./jwt
```

### Building
```bash
# Build and verify module
go build ./...

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Linting & Quality
```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run staticcheck (if installed)
staticcheck ./...
```

## Architecture Patterns

### 1. Middleware Pattern
The library heavily uses the HTTP middleware pattern compatible with go-chi:
```go
func Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Pre-processing
            next.ServeHTTP(w, r)
            // Post-processing
        })
    }
}
```

### 2. Context-Based Session Management
Sessions are loaded and managed through request context:
```go
// Loading sessions with automatic fallback
err := manager.Load(r, &sessionData)
```

### 3. Structured Error Handling
API errors follow a consistent structure:
```go
type Err struct {
    Code    int         `json:"-"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Err     error       `json:"-"`
}
```

### 4. Interface-Based Design
Key components use interfaces for flexibility:
- `tokens.Store` interface for different storage backends
- `sessions.Manager` for different session strategies

## Configuration

### Environment Loading
```go
// embed basic env in your env struct
type AppEnv struct {
    anansi.BasicEnv
    // your custom fields
}

env := &AppEnv{}
err := anansi.LoadEnv(env) // loads from .env file and environment
```

### Session Configuration
```go
config := sessions.Config{
    IsProduction:     true,
    HeadlessScheme:   "API",
    BearerDuration:   time.Hour,
    CookieDuration:   24 * time.Hour,
    HeadlessDuration: 30 * time.Minute,
    CookieKey:        "app_session",
}
manager := sessions.NewManager(store, secret, config)
```

### Logging Setup
```go
logger := anansi.NewLogger("service-name")
// Creates structured logger with timestamp, service, and host fields
```

## Testing Strategy

The codebase follows Go testing conventions:
- Unit tests in `*_test.go` files alongside source code
- Table-driven tests for multiple scenarios
- Mock interfaces for external dependencies
- Integration tests for middleware chains

## Security Considerations

1. **Secret Management**: JWT/session secrets must be at least 32 bytes
2. **Secure Cookies**: Automatic secure/httpOnly flags in production
3. **Token Signing**: HMAC-SHA256 for token integrity
4. **Context Timeouts**: Built-in timeout handling for requests
5. **Panic Recovery**: Graceful error handling prevents information leakage

## Integration Example

```go
func main() {
    env := &AppEnv{}
    anansi.LoadEnv(env)
    
    logger := anansi.NewLogger(env.ServiceName)
    
    router := chi.NewRouter()
    router.Use(api.Recoverer(env.Environment))
    
    // Session middleware
    sessionManager := sessions.NewManager(tokenStore, secret, sessionConfig)
    router.Use(api.Headless(sessionManager))
    
    // Routes
    router.Get("/api/v1/users", handleUsers)
    
    http.ListenAndServe(":8080", router)
}
```

This library is designed to be a comprehensive toolkit for Go web applications, providing production-ready components for session management, API development, and request/response handling while maintaining flexibility and performance.

## Debugging Instructions

When debugging issues in the codebase, follow this systematic approach:

1. **Identify 5-7 potential sources** of the issue:
   - List all possible causes based on error messages and symptoms
   - Consider parameter mismatches, data format differences, timing issues
   - Think about external dependencies (Redis, database, network)
   - Include context/configuration problems

2. **Select the top 1-2 most likely sources**:
   - Prioritize based on probability and impact
   - Focus on the most common failure patterns first

3. **Add targeted logging to validate assumptions**:
   - Add temporary `fmt.Printf` statements at key points
   - Log input/output values, state changes, and external calls
   - Validate your hypothesis with concrete data

4. **Clean up and implement fix**:
   - Remove all debugging logging before committing
   - Implement the validated solution
   - Add tests to prevent regression

## Git Commit Strategy

When the user asks to "save our work" or create commits, follow this structured approach:

### 1. Analyze Changes by Domain
Use `git status` and `git diff` to examine all staged and unstaged changes. Group changes by:

**Business/Product Features (`feat:`):**
- New functionality or enhancements to existing features
- User-facing improvements (UI, UX, form handling)
- Feature completions that add business value

**Bug Fixes (`fix:`):**
- Corrections to broken functionality
- Security patches and vulnerability fixes
- Performance improvements that resolve issues

**Refactoring (`refactor:`):**
- Code restructuring without changing external behavior
- Note: Small refactors should be included with related `feat:` commits

**Maintenance (`chore:`):**
- Dependency updates, build configuration changes
- Documentation updates, tooling improvements
- Infrastructure and deployment-related changes

### 2. Create Logical Commit Groups
Organize related changes into cohesive commits:
- **Backend changes** (Go handlers, repositories, middleware)
- **Frontend changes** (Templates, CSS, JavaScript)
- **Configuration changes** (Environment, build files)
- **Database changes** (Migrations, schema updates)

### 3. Commit Message Format
Use conventional commits following email subject best practices:

**Subject Line (≤50 characters):**
- Frame as "what problem was solved" (Pieter Hintjens approach)
- Use imperative mood ("fix user login issue" not "fixed user login issue")
- No period at the end
- Focus on the problem solved, not implementation details

**Body (optional, wrap at 72 characters):**
```
type: solve specific problem with clear solution

- Detailed explanation of the problem that was solved
- Why this solution was chosen
- Any breaking changes or important notes

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### 4. Commit Sequence
1. **Check status**: `git status`, `git diff --name-status`
2. **Examine diffs**: `git diff` for each file to understand changes
3. **Group logically**: Stage related files together
4. **Commit incrementally**: Create focused, single-purpose commits
5. **Verify clean state**: Ensure `git status` shows clean working tree

### 5. Example Problem-Focused Commit Messages
- `feat: solve checkout abandonment for authenticated users`
- `fix: resolve session loss after OAuth redirect`
- `feat: prevent invalid form submissions at checkout`
- `chore: solve build failures in production environment`
- `refactor: eliminate inconsistent API response formats`

This approach ensures clean git history, makes code reviews easier, and provides clear documentation of development progress.