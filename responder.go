package idemgotent

import (
	"net/http"

	"github.com/pwnedgod/idemgotent/internal/util"
)

type (
	CacheResult struct {
		// Whether the action (the next in chain of handlers) is executed.
		//
		// True means that it has obtained the response from the cache.
		FromCache bool

		// The response (in serializable format) obtained
		// through cache or directly from the action.
		Response SerializableResponse
	}

	Responder func(http.ResponseWriter, *http.Request, CacheResult)
)

// Respond with status code, header, and body according to the cache value.
// Ideally, the response will be identical for the same key.
//
// Optionally on cached responses, status code can be forced to be a specific value
// and headers can be filtered.
//
// Set `statusCode` to use the cached status code value.
// Set `headerNames` to be "*" to allow all cached header values.
func RespondCached(statusCode int, headerNames ...string) Responder {
	wildcard := util.Contains(headerNames, "*")

	return func(w http.ResponseWriter, r *http.Request, cr CacheResult) {
		if cr.FromCache && statusCode != 0 {
			w.WriteHeader(statusCode)
		} else {
			w.WriteHeader(cr.Response.StatusCode)
		}

		for name, values := range cr.Response.Header {
			if cr.FromCache && !wildcard && !util.Contains(headerNames, name) {
				continue
			}
			w.Header()[name] = values
		}

		w.Write(cr.Response.Body)
	}
}
