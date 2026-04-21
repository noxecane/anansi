# anansi

Helper toolkit for building Go web services with [chi](https://github.com/go-chi/chi). Covers server lifecycle, session management, request parsing, response handling, structured logging, encryption, and inter-service communication.

## Install

```sh
go get github.com/noxecane/anansi
```

## Usage

- [Server setup](#1-server-setup)
- [Session management](#2-session-management)
- [Request handling](#3-request-handling)
- [HTML rendering](#4-html-rendering)
- [Inter-service requests](#5-inter-service-requests)

---

### 1. Server setup

The typical entrypoint wires together environment loading, logging, middleware, and graceful shutdown.

```go
type AppEnv struct {
    Name        string `envconfig:"APP_NAME"`
    Environment string `envconfig:"APP_ENV"`
    Port        string `envconfig:"PORT"`
}

func main() {
    env := &AppEnv{}
    if err := anansi.LoadEnv(env); err != nil {
        log.Fatal(err)
    }

    logger := anansi.NewLogger(env.Name)
    redis  := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    secret := []byte("at-least-32-bytes-long-secret!!!")

    tokenStore := tokens.NewStore(redis, secret)
    sessions   := sessions.NewManager(tokenStore, secret, sessions.Config{
        IsProduction:   env.Environment == "production",
        BearerDuration: 24 * time.Hour,
        CookieDuration: 7 * 24 * time.Hour,
    })

    r := chi.NewRouter()
    r.Use(
        middleware.RequestID,
        requests.AttachLogger(logger),
        requests.Log,
        requests.CORS(env.Environment, "https://yourdomain.com"),
        requests.Timeout(30 * time.Second),
        responses.ResponseTime,
        api.Recoverer(env.Environment),
    )

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        api.Success(r, w, map[string]string{"status": "ok"})
    })

    // Graceful shutdown: ctx is cancelled on SIGINT/SIGTERM
    ctx, _ := anansi.WithCancel(context.Background())

    srv := &http.Server{Addr: ":" + env.Port, Handler: r}
    go srv.ListenAndServe()

    <-ctx.Done()
    srv.Shutdown(context.Background())
}
```

---

### 2. Session management

`sessions.Manager` supports three modes:

| Mode | Transport | Backing |
|---|---|---|
| Cookie | `Set-Cookie` header | Redis |
| Bearer | `Authorization: Bearer <token>` | Redis |
| Headless | `Authorization: <Scheme> <token>` | Stateless JWE |

```go
type UserSession struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
}

// --- Login: create a session ---
func login(m *sessions.Manager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := authenticateUser(r) // your logic

        sess := UserSession{UserID: user.ID, Role: user.Role}

        // Cookie-based
        token, err := m.NewSession(r.Context(), user.Email, sess)
        m.ToCookie(w, token, "/")

        // Bearer-based (same call, different transport)
        // m.ToAuth(w, token, false)

        // Headless / API key (stateless JWT)
        // token, err = m.NewHeadlessSession(sess)
        // m.ToAuth(w, token, true)

        api.Success(r, w, nil)
    }
}

// --- Middleware: require authenticated session ---
func requireAuth(m *sessions.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var sess UserSession
            api.Load(m, r, &sess) // panics with 401 on failure
            next.ServeHTTP(w, r)
        })
    }
}

// --- In a handler: read the session ---
func getProfile(m *sessions.Manager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var sess UserSession
        api.Load(m, r, &sess)
        api.Success(r, w, sess)
    }
}

// --- Logout ---
func logout(m *sessions.Manager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m.LogoutCookie(r, w)   // or m.LogoutAuth(r)
        api.Success(r, w, nil)
    }
}
```

---

### 3. Request handling

Anansi uses a **panic-based error flow**: handlers panic with `api.Err`, and `api.Recoverer` catches it and writes the HTTP response. This keeps handler code flat.

```go
type CreateBookingRequest struct {
    ClientName  string    `json:"client_name"  mod:"trim"    validate:"required"`
    RoomID      uint      `json:"room_id"                    validate:"required"`
    StartsAt    string    `json:"starts_at"                  validate:"required"`
}

func (r CreateBookingRequest) Validate() error {
    return validation.ValidateStruct(&r,
        validation.Field(&r.ClientName, validation.Required),
        validation.Field(&r.RoomID,     validation.Required, validation.Min(uint(1))),
    )
}

func createBooking(w http.ResponseWriter, r *http.Request) {
    // Parses JSON body, applies mold transformations, runs ozzo validation.
    // Panics with 400/415 on any failure.
    var req CreateBookingRequest
    api.ReadJSON(r, &req)

    // Extract typed URL params
    studioID := api.IDParam(r, "studioID")    // uint
    slug      := api.StringParam(r, "slug")   // string

    // Parse query params into a struct (same mod/validation support)
    type Filters struct {
        State string `json:"state"`
        Page  int    `json:"page"`
    }
    var filters Filters
    api.QueryParam(r, &filters)

    // Your domain logic here
    booking, err := bookingService.Create(r.Context(), studioID, req)
    if err != nil {
        panic(api.Err{
            Code:    http.StatusUnprocessableEntity,
            Message: "Could not create booking",
            Err:     err,
        })
    }

    api.Success(r, w, booking)
}
```

---

### 4. HTML rendering

For server-rendered pages, `html.Parse` wraps Go's `html/template` with integrated zerolog logging.

```go
var bookingPage = html.Parse("booking.html",
    "templates/layout.html",
    "templates/booking.html",
)

func showBooking(w http.ResponseWriter, r *http.Request) {
    id := api.IDParam(r, "id")
    booking := bookingService.Get(r.Context(), id)

    if err := bookingPage.Render(r, w, booking); err != nil {
        panic(err)
    }
}
```

Serve static assets:

```go
// Registers GET /public/* pointing to ./public/
html.Server(router, "public")
```

Manage cookies directly:

```go
// HttpOnly + Secure in prod, lax SameSite
ck := html.SecureCookie(isProd, &http.Cookie{Name: "pref", Value: "dark"})
http.SetCookie(w, ck)

// Strict SameSite (e.g. for CSRF-sensitive state)
ck = html.LockCookie(isProd, &http.Cookie{Name: "csrf", Value: token})
http.SetCookie(w, ck)
```

---

### 5. Inter-service requests

`ajax.Client` propagates request IDs and authentication across internal service calls.

```go
client := ajax.NewClient(ajax.Config{
    Service:          "booking-service",
    Secret:           []byte("shared-inter-service-secret!!!"),
    HeadlessScheme:   "Service",
    HeadlessDuration: time.Minute,
})

// Proxy auth from the incoming request
req, err := client.NewRequest(r, http.MethodGet, "http://payments-svc/invoices/42", nil)

// Replace auth with a custom session for the downstream service
type ServiceSession struct{ ServiceName string }
req, err = client.NewHeadlessRequest(r, http.MethodPost, "http://notify-svc/send", ServiceSession{"booking"}, body)

// Originate a new request (no parent HTTP request)
req, err = client.NewBaseRequest(ctx, http.MethodGet, "http://studio-svc/rooms", ServiceSession{"booking"}, nil)

// Parse response
res, _ := http.DefaultClient.Do(req)
if err := ajax.GetErr(res); err != nil {
    return err
}
var invoice Invoice
ajax.GetResponse(res, &invoice)
```

---

## API Reference

### `anansi` (root package)

| Symbol | Signature | Description |
|---|---|---|
| `LoadEnv` | `(env any) error` | Loads `.env` file then maps env vars into `env` using `envconfig` |
| `NewLogger` | `(service string) zerolog.Logger` | Structured zerolog logger with `service` and `host` fields |
| `WithCancel` | `(parent context.Context) (context.Context, context.CancelFunc)` | Like `context.WithCancel` but also cancels on `SIGINT`/`SIGTERM` |
| `Encrypt` | `(secret, value []byte) (string, error)` | NaCl secretbox encryption, returns URL-safe base64. Secret truncated to 32 bytes |
| `Decrypt` | `(secret []byte, encrypted string) ([]byte, error)` | Inverse of `Encrypt` |
| `RandomBytes` | `(n int) ([]byte, error)` | Cryptographically random bytes |
| `RandomString` | `(s int) (string, error)` | Cryptographically random hex string of length `s` |
| `Slugify` | `(s string) string` | Converts CamelCase to snake_case |
| `ParseISO` | `(date string) (time.Time, error)` | Parses ISO 8601 dates: `2006-01-02T15:04:05.000Z`, `2006-01-02`, `2006-01-02T15:04:05` |
| `SimpleMap` | `(m map[string][]string) map[string]interface{}` | Flattens header/query maps; lowercases keys |

---

### `api`

Panic-based HTTP handlers. `api.Recoverer` must be in the middleware chain to handle panics.

#### Error type

```go
type Err struct {
    Code    int         // HTTP status code
    Message string      // Human-readable message (JSON: "message")
    Data    interface{} // Structured detail, e.g. validation errors (JSON: "data")
    Err     error       // Wrapped cause, not serialised
}
```

#### Middleware

| Function | Description |
|---|---|
| `Recoverer(env string)` | Recovers panics. Writes `api.Err` as JSON. Prints stack trace in `dev`/`test`. Returns 504 on deadline exceeded |
| `Headless(manager)` | Requires a valid headless (JWE) token on every request. Panics 401 otherwise |

#### Request helpers

| Function | Panics on |
|---|---|
| `ReadJSON(r, v)` | 415 if not JSON, 400 if validation fails |
| `QueryParam(r, v)` | 400 if parsing or validation fails |
| `IDParam(r, name) uint` | 400 if URL param is not a valid integer |
| `StringParam(r, name) string` | Panics (non-HTTP) if param not in route |

#### Session helpers (panic on 401)

| Function | Description |
|---|---|
| `Load(m, r, v)` | Tries `Authorization` header, falls back to cookie |
| `LoadBearer(m, r, v)` | Requires `Authorization: Bearer <token>` |
| `LoadCookie(m, r, v)` | Requires session cookie |
| `LoadHeadless(m, r, v)` | Requires headless JWE token |

#### Response helpers

| Function | Description |
|---|---|
| `Success(r, w, v)` | Writes 200 JSON response, logs it |
| `Error(r, w, err)` | Writes `err.Code` JSON response, logs it |

---

### `sessions`

```go
func NewManager(store tokens.Store, secret []byte, config Config) *Manager
```

**Config fields:**

| Field | Default | Description |
|---|---|---|
| `IsProduction` | `false` | Enables `Secure` + `SameSite` on cookies |
| `HeadlessScheme` | `"API"` | Authorization scheme for JWE tokens |
| `HeadlessDuration` | `BearerDuration` | TTL for headless tokens |
| `BearerDuration` | `1h` | TTL for bearer tokens |
| `CookieKey` | `"anansi_session"` | Cookie name |
| `CookieDuration` | `BearerDuration` | TTL for cookie sessions |

**Manager methods:**

| Method | Description |
|---|---|
| `NewSession(ctx, sessionID, v) (string, error)` | Creates a Redis-backed token keyed by `sessionID` |
| `NewHeadlessSession(v) (string, error)` | Creates a stateless JWE token |
| `ToCookie(w, token, path)` | Writes token to a secure cookie |
| `ToAuth(w, token, isHeadless)` | Writes token to `Authorization` header |
| `FromCookie(r, v) error` | Reads + extends cookie session |
| `FromAuth(r, v) error` | Reads bearer or headless token |
| `Load(r, v) error` | Tries auth header, falls back to cookie |
| `LogoutCookie(r, w) error` | Revokes cookie session, clears cookie |
| `LogoutAuth(r) error` | Revokes bearer session (headless is stateless, no-op) |

**Errors:** `ErrEmptyHeader`, `ErrHeaderFormat`, `ErrUnsupportedScheme`, `ErrEmptyAuthCookie`

---

### `tokens`

Redis-backed token store with HMAC-SHA256 signing. Token is derived from the key, so the same key always produces the same token.

```go
func NewStore(r *redis.Client, secret []byte) Store
```

**Store interface:**

| Method | Description |
|---|---|
| `Commission(ctx, ttl, key, v) (string, error)` | Creates a token for `key`, stores `v` with the given TTL |
| `Peek(ctx, token, v) error` | Reads token data without changing TTL |
| `Extend(ctx, token, ttl, v) error` | Reads data and resets TTL to `ttl` |
| `Reset(ctx, key, v) error` | Updates token data while preserving existing TTL |
| `Decommission(ctx, token, v) error` | Reads data and deletes the token |
| `Revoke(ctx, key) error` | Deletes token by key without reading data |

**Error:** `ErrTokenNotFound`

---

### `jwt`

JWE (encrypted JWT) encoding/decoding using AES-256-GCM. Secret must be at least 32 bytes.

| Function | Description |
|---|---|
| `Encode(secret, ttl, v) (string, error)` | Encrypts `v` as a JWE token with expiry |
| `Decode(secret, token, v) error` | Decrypts and validates token, maps claims into `v` |

**Errors:** `ErrJWTExpired`, `ErrInvalidToken`

---

### `requests`

| Function | Description |
|---|---|
| `ReadBody(r) ([]byte, error)` | Reads body without consuming it (safe to read again) |
| `ReadJSON(r, v) error` | Decodes JSON body, runs mold transforms + ozzo validation |
| `QueryParams(r, v) error` | Maps query string into struct via `json` tags |
| `FormData(r, v) error` | Maps `application/x-www-form-urlencoded` into struct |
| `MultipartFormData(r, maxSize, v) error` | Maps multipart form fields (non-file) into struct |
| `IDParam(r, name) (uint, error)` | Extracts integer URL param |
| `StringParam(r, name) string` | Extracts string URL param, panics if missing from route |
| `CORS(env, origins...) Middleware` | Permissive CORS in `dev`, restricted to `origins` otherwise |
| `Timeout(d) Middleware` | Cancels request context after duration `d` |
| `AttachLogger(log) Middleware` | Attaches zerolog logger to request context |
| `Log` | Middleware that enriches the logger with method, URL, headers, JSON body |
| `AddModifier(tag, fn)` | Registers a custom mold modifier for use in `mod` struct tags |

**Built-in `mod` tags:** `lowercase`, `uppercase`, `smalltext` (lowercase + trim), plus all [mold](https://github.com/go-playground/mold) built-ins.

---

### `responses`

| Function | Description |
|---|---|
| `Send(w, code, data []byte)` | Writes raw JSON bytes with `Content-Type: application/json` |
| `ResponseTime` | Middleware that adds `X-Response-Time: Nms` header |
| `RequestDuration(reg)` | Middleware that reports `http_request_duration_seconds` histogram to Prometheus |

---

### `html`

| Function | Description |
|---|---|
| `Parse(name, files...) Template` | Parses Go HTML templates; panics if files are invalid |
| `(Template).Render(r, w, data)` | Executes the template and logs the render |
| `Server(router, folder)` | Registers `GET /<folder>/*` to serve static files from `./<folder>/` |
| `SecureCookie(isProd, ck)` | Sets `HttpOnly`, `Secure` (prod), `SameSite=Lax` (prod) |
| `LockCookie(isProd, ck)` | Like `SecureCookie` but with `SameSite=Strict` |

---

### `ajax`

Inter-service HTTP client that propagates request IDs and authentication.

```go
func NewClient(conf Config) Client
```

**Config fields:** `Secret []byte`, `Service string`, `HeadlessScheme string`, `HeadlessDuration time.Duration` (default 1m)

**Client methods:**

| Method | Description |
|---|---|
| `NewRequest(r, method, url, body)` | Proxies `X-Request-Id` and `Authorization` from parent request |
| `NewHeadlessRequest(r, method, url, session, body)` | Like `NewRequest` but replaces auth with a fresh JWE token |
| `NewBaseRequest(ctx, method, url, session, body)` | Originates a new request with a generated request ID |
| `GetErr(res) error` | Returns `api.Err` if response status >= 400, else nil |
| `GetResponse(res, v) error` | Decodes response body into `v` |

---

### `postgres`

| Function | Description |
|---|---|
| `Migrate(dir, dbName, schema, url) error` | Runs `golang-migrate` SQL migrations from `dir` (must be absolute path) |

**Error regexps** (match against `err.Error()`):

| Var | PG code | Meaning |
|---|---|---|
| `ErrIntegrity` | `#23000` | Integrity constraint |
| `ErrRestrict` | `#23001` | Restrict violation |
| `ErrNotNull` | `#23502` | Not-null violation |
| `ErrForeignKey` | `#23503` | Foreign key violation |
| `ErrDuplicate` | `#23505` | Unique violation |

---

### `json`

Drop-in replacement for `encoding/json` using [jsoniter](https://github.com/json-iterator/go) for faster marshaling.

```go
import "github.com/noxecane/anansi/json"

json.Marshal(v)
json.Unmarshal(data, v)
json.NewDecoder(r).Decode(v)
json.NewEncoder(w).Encode(v)
```
