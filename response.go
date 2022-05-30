package idemgotent

import "net/http"

type (
	serializableResponse struct {
		_msgpack struct{} `msgpack:",omitempty"`

		StatusCode int
		Header     http.Header
		Body       []byte

		// Unserializable fields.
		//
		// Allows serialization only on the required fields
		// but still have access to the data on return.

		pStatusCode int
		pHeader     http.Header
		pBody       []byte
	}

	CapturedResponse interface {
		// Obtain the captured status code from the response.
		GetStatusCode() int

		// Obtain the captured header from the response.
		GetHeader() http.Header

		// Obtain the captured body from the response.
		GetBody() []byte
	}
)

func (sr *serializableResponse) setStatusCode(statusCode int, serialize bool) {
	if serialize {
		sr.StatusCode = statusCode
		return
	}
	sr.pStatusCode = statusCode
}

func (sr *serializableResponse) setHeader(header http.Header, serialize bool) {
	if serialize {
		sr.Header = header
		return
	}
	sr.pHeader = header
}

func (sr *serializableResponse) setBody(body []byte, serialize bool) {
	if serialize {
		sr.Body = body
		return
	}
	sr.pBody = body
}

func (sr serializableResponse) GetStatusCode() int {
	if sr.StatusCode != 0 {
		return sr.StatusCode
	}
	return sr.pStatusCode
}

func (sr serializableResponse) GetHeader() http.Header {
	if sr.Header != nil {
		return sr.Header
	}
	return sr.pHeader
}

func (sr serializableResponse) GetBody() []byte {
	if sr.Body != nil {
		return sr.Body
	}
	return sr.pBody
}
