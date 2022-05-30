# Idemgotent

Middleware for providing idempotency for APIs.

Uses [WraCha](https://github.com/pwnedgod/wracha) as its base.

## Installation

Simply run the following command to install:

```
go get github.com/pwnedgod/idemgotent
```

## Usage
### Initialization

```go
package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/pwnedgod/idemgotent"
	"github.com/pwnedgod/wracha/adapter/goredis"
	"github.com/pwnedgod/wracha/logger/std"
)

func main() {
	// ... your router (example with go-chi)
	r := chi.NewRouter()

	// ... your redis client
	client := redis.NewClient(&redis.Options{
		// ...
	})

	middleware := idemgotent.Middleware("/your/path/to/route",
		idemgotent.WithAdapter(goredis.NewAdapter(client)),
		idemgotent.WithLogger(std.NewLogger()),
	)

	// ... (example 1 with go-chi)
	r.Group(func(r chi.Router) {
		r.Use(middleware)
		r.Post("/your/path/to/route", myHandler)
	})

	// ... (example 2 with go-chi)
	r.With(middleware).Post("/your/path/to/route", myHandler)
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// ...
}
```
