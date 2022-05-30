package idemgotent

import (
	"encoding/json"
	"net/http"
)

type (
	ErrHandler func(error, http.ResponseWriter, *http.Request)
)

const (
	HeaderContentType = "Content-Type"

	ContentTypeJSON = "application/json"

	KeyMessage = "message"
)

func JSONErrHandler(statusCode int) ErrHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		// Should never return an error.
		body, _ := json.Marshal(map[string]any{
			KeyMessage: err.Error(),
		})

		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.WriteHeader(statusCode)
		w.Write(body)
	}
}
