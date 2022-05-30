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

### Extra Configuration

By design, this library is meant to be configurable and modular at certain parts.

#### Determining Key

By default, the library uses the value of the header `Idempotency-Key` for determining idempotency.
You can configure this by passing

```go
idemgotent.WithKeySource(idemgotent.HeaderKeySource("Custom-Idempotent-Key"))
```

to the options argument.

Alternatively, you can also define your own way to obtain the key by satisfying the `KeySource` function type.
Be careful when obtaining idempotency keys from body as you might have to do some workarounds to allow multiple reads of the request body.

```go
type KeySource func(r *http.Request) (string, error)
```

```go
func JSONKeySource(name string) idemgotent.KeySource {
	return func(r *http.Request) (string, error) {
		// ... unmarshal and read JSON.
	}
}
```

#### Responding to Clients

By default, the library responds with the previously cached response along with its status code and headers.
This is override-able by passing

```go
idemgotent.WithResponder(idemgotent.CachedResponder(http.StatusNotModified, "Content-Type"))
```

to the options argument.

You can also implement your own responder.

```go
type Responder interface {
	CacheStatusCode() bool
	CacheHeader() bool
	CacheBody() bool
	Respond(http.ResponseWriter, *http.Request, CacheResult)
}
```

```go
type conflictResponder struct {
}

func (conflictResponder) CacheStatusCode() bool {
	return false
}

func (conflictResponder) CacheHeader() bool {
	return false
}

func (conflictResponder) CacheBody() bool {
	return false
}

func (rp conflictResponder) Respond(w http.ResponseWriter, r *http.Request, cr CacheResult) {
	if cr.FromCache {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("{\"message\": \"idempotency violation\"}"))
		return
	}

	w.WriteHeader(cr.Response.GetStatusCode())
	cr.CopyHeaderTo(w, nil)
	w.Write(cr.Response.GetBody())
}
```
