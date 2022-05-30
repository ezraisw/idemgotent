package idemgotent

import "net/http"

type (
	SerializableResponse struct {
		StatusCode int         `msgpack:"statusCode"`
		Header     http.Header `msgpack:"header"`
		Body       []byte      `msgpack:"body"`
	}
)
