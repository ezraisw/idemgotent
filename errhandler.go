package idemgotent

import (
	"encoding/json"
	"net/http"
)

type (
	ErrHandler func(error, http.ResponseWriter, *http.Request)
)

func JSONErrHandler(statusCode int) ErrHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		// Should never return an error.
		body, _ := json.Marshal(map[string]any{
			"message": err.Error(),
		})

		w.WriteHeader(statusCode)
		w.Write(body)
	}
}
