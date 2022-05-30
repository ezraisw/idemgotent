package idemgotent

import (
	"time"

	"github.com/pwnedgod/wracha/adapter"
	"github.com/pwnedgod/wracha/logger"
)

type MiddlewareOption func(*idempotencyMiddleware)

func WithAdapter(adapter adapter.Adapter) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.adapter = adapter
	}
}

func WithLogger(logger logger.Logger) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.logger = logger
	}
}

func WithTTL(ttl time.Duration) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.ttl = ttl
	}
}

func WithKeySource(keySource KeySource) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.keySource = keySource
	}
}

func WithClientErrHandler(clientErrHandler ErrHandler) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.clientErrHandler = clientErrHandler
	}
}

func WithServerErrHandler(serverErrHandler ErrHandler) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.serverErrHandler = serverErrHandler
	}
}

func WithResponder(responder Responder) MiddlewareOption {
	return func(m *idempotencyMiddleware) {
		m.responder = responder
	}
}
