package idemgotent

import (
	"net/http"
)

type (
	KeySource func(r *http.Request) (string, error)
)

func HeaderKeySource(name string) KeySource {
	return func(r *http.Request) (string, error) {
		return r.Header.Get(name), nil
	}
}
