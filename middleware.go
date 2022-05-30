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
		actor wracha.Actor[SerializableResponse]
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
		WithResponder(RespondCached(0, "*")),
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

	m.actor = wracha.NewActor[SerializableResponse](ActorNamePrefix+m.name, actorOptions).
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

		// Do not attempt to cache with empty keys. Immediately do the next in chain.
		if key == "" {
			m.logger.Debug("request has no key")

			next.ServeHTTP(w, r)
			return
		}

		// Will be true if the action is executed.
		run := false

		resp, err := m.actor.Do(
			r.Context(),
			wracha.KeyableStr(key),

			// Will be executed only once for each key.
			m.makeAction(func(nw http.ResponseWriter) { next.ServeHTTP(nw, r); run = true }),
		)
		if err != nil {
			m.logger.Debug("error returned by actor", err)

			m.serverErrHandler(err, w, r)
			return
		}

		fromCache := !run

		m.logger.Debug("responding from cache", fromCache)
		m.responder(w, r, CacheResult{FromCache: fromCache, Response: resp})
	})
}

func (m idempotencyMiddleware) makeAction(nextWriteTo func(http.ResponseWriter)) wracha.ActionFunc[SerializableResponse] {
	return func(ctx context.Context) (wracha.ActionResult[SerializableResponse], error) {
		// Capture the response by substituting http.ResponseWriter with a custom one.
		bw := response.NewBufferedResponseWriter()
		nextWriteTo(bw)

		resp := SerializableResponse{
			StatusCode: bw.StatusCode(),
			Header:     bw.Header(),
			Body:       bw.Body(),
		}

		return wracha.ActionResult[SerializableResponse]{Cache: true, Value: resp}, nil
	}
}
