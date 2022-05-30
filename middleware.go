package idemgotent

import (
	"context"
	"net/http"
	"time"

	"github.com/pwnedgod/idemgotent/internal/response"
	"github.com/pwnedgod/wracha"
	"github.com/pwnedgod/wracha/adapter"
	"github.com/pwnedgod/wracha/codec/msgpack"
	"github.com/pwnedgod/wracha/logger"
)

type (
	idempotencyMiddleware struct {
		name string

		// Required options.
		adapter adapter.Adapter
		logger  logger.Logger

		// Optional options.
		ttl              time.Duration
		keySource        KeySource
		clientErrHandler ErrHandler
		serverErrHandler ErrHandler
		responder        Responder

		// Initialized.
		actor wracha.Actor[serializableResponse]
	}
)

const (
	ActorNamePrefix = "idemgotent-"

	HeaderIdempotencyKey = "Idempotency-Key"
)

var (
	defaultOptions = []MiddlewareOption{
		WithTTL(wracha.TTLDefault),
		WithKeySource(HeaderKeySource(HeaderIdempotencyKey)),
		WithClientErrHandler(JSONErrHandler(http.StatusBadRequest)),
		WithServerErrHandler(JSONErrHandler(http.StatusInternalServerError)),
		WithResponder(CachedResponder(0, "*")),
	}
)

func Middleware(name string, options ...MiddlewareOption) func(http.Handler) http.Handler {
	m := idempotencyMiddleware{name: name}

	m.applyOptions(defaultOptions)
	m.applyOptions(options)

	m.initActor()

	return m.middleware
}

func (m *idempotencyMiddleware) applyOptions(options []MiddlewareOption) {
	for _, option := range options {
		option(m)
	}
}

func (m *idempotencyMiddleware) initActor() {
	actorOptions := wracha.ActorOptions{
		Adapter: m.adapter,
		Logger:  m.logger,

		// There should be no necessity to override this.
		Codec: msgpack.NewCodec(),
	}

	m.actor = wracha.NewActor[serializableResponse](ActorNamePrefix+m.name, actorOptions).
		SetTTL(m.ttl)
}

func (m idempotencyMiddleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtain the key from the request.
		m.logger.Debug("obtaining key")

		key, err := m.keySource(r)
		if err != nil {
			m.logger.Debug("error while obtaining key", err)

			m.clientErrHandler(err, w, r)
			return
		}

		// Do not attempt to cache with empty keys. Skip to the next in chain.
		if key == "" {
			m.logger.Debug("request has no key")

			next.ServeHTTP(w, r)
			return
		}

		// Will be false if the action is executed.
		fromCache := true

		resp, err := m.actor.Do(
			r.Context(),
			wracha.KeyableStr(key),

			// Will be executed only once for each key.
			m.makeAction(func(nw http.ResponseWriter) { next.ServeHTTP(nw, r); fromCache = false }),
		)
		if err != nil {
			m.logger.Debug("error returned by actor", err)

			m.serverErrHandler(err, w, r)
			return
		}

		m.logger.Debug("responding from cache", fromCache)
		m.responder.Respond(w, r, CacheResult{FromCache: fromCache, Response: resp})
	})
}

func (m idempotencyMiddleware) makeAction(nextWriteTo func(http.ResponseWriter)) wracha.ActionFunc[serializableResponse] {
	return func(ctx context.Context) (wracha.ActionResult[serializableResponse], error) {
		// Capture the response by substituting http.ResponseWriter with a custom one.
		bw := response.NewBufferedResponseWriter()
		nextWriteTo(bw)

		return wracha.ActionResult[serializableResponse]{Cache: true, Value: m.makeSerializableResponse(bw)}, nil
	}
}

func (m idempotencyMiddleware) makeSerializableResponse(bw *response.BufferedResponseWriter) serializableResponse {
	var resp serializableResponse

	// To reduce cache size.
	resp.setStatusCode(bw.StatusCode(), m.responder.CacheStatusCode())
	resp.setHeader(bw.Header(), m.responder.CacheHeader())
	resp.setBody(bw.Body(), m.responder.CacheBody())

	return resp
}
