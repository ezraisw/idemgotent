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
		Response CapturedResponse
	}

	Responder interface {
		// Whether to cache status code.
		//
		// If false and CacheResult.FromCache is true, CacheResult.Response.GetStatusCode() returns 0.
		CacheStatusCode() bool

		// Whether to cache header.
		//
		// If false and CacheResult.FromCache is true, CacheResult.Response.GetHeader() returns nil.
		CacheHeader() bool

		// Whether to cache body.
		//
		// If false and CacheResult.FromCache is true, CacheResult.Response.GetBody() returns nil.
		CacheBody() bool

		// The behaviour when responding to client.
		Respond(http.ResponseWriter, *http.Request, CacheResult)
	}

	cachedResponder struct {
		overrideStatusCode int
		wildcard           bool
		allowedHeaderNames []string
	}
)

// Copy header to the given ResponseWriter.
//
// `filterFn` can be nil to allow all header.
func (cr CacheResult) CopyHeaderTo(w http.ResponseWriter, filterFn func(string, []string) bool) {
	util.MCopyFilter(cr.Response.GetHeader(), w.Header(), filterFn)
}

// Respond with status code, header, and body according to the cache value.
// Ideally, the response will be identical for the same key.
//
// Optionally on cached responses, status code can be forced to be a specific value
// and headers can be filtered.
//
// Set `overrideStatusCode` to use the cached status code value.
// Set `allowedHeaderNames` to be "*" to allow all cached header values.
func CachedResponder(overrideStatusCode int, allowedHeaderNames ...string) Responder {
	r := &cachedResponder{
		overrideStatusCode: overrideStatusCode,
		wildcard:           util.Contains(allowedHeaderNames, "*"),
	}

	if !r.wildcard {
		r.allowedHeaderNames = allowedHeaderNames
	}

	return r
}

func (rp cachedResponder) CacheStatusCode() bool {
	return rp.overrideStatusCode == 0
}

func (rp cachedResponder) CacheHeader() bool {
	return rp.wildcard || len(rp.allowedHeaderNames) > 0
}

func (cachedResponder) CacheBody() bool {
	return true
}

func (rp cachedResponder) Respond(w http.ResponseWriter, r *http.Request, cr CacheResult) {
	if !cr.FromCache || rp.wildcard {
		cr.CopyHeaderTo(w, nil)
	} else if len(rp.allowedHeaderNames) > 0 {
		cr.CopyHeaderTo(w, func(name string, _ []string) bool { return util.Contains(rp.allowedHeaderNames, name) })
	}

	statusCode := cr.Response.GetStatusCode()
	if rp.overrideStatusCode != 0 {
		statusCode = rp.overrideStatusCode
	}
	w.WriteHeader(statusCode)

	w.Write(cr.Response.GetBody())
}
