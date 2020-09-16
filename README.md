# siber

Helper tools for [go-chi](https://github.com/go-chi/chi)

## Features

- API Errors
- Logging Middleware
- Session Store
- Encryption Helpers
- JSON request and body parsing
- Extensions of [faker](github.com/bxcodec/faker/v3)

## Install

```sh
go get -u github.com/tsaron/ansi
```

## Example

```go

// embed basic env in your env struct
type AppEnv struct {
    siber.BasicEnv
}

type Book struct {
    Name `json:"book_nme"`
}

func main() {
    // BasicEnv is only compatible with envconfig ATM
    env := loadEnv()
    logs := NewLogger(env.Name)

    router := chi.NewRouter()
    chix.DefaultMiddleware(env, log, router)
    router.Get("/api/v1/books/:id", func(w http.ResponseWriter, r *http.Request) {

    })

}
```
