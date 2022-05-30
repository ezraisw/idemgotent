package response

import (
	"bytes"
	"net/http"
)

type (
	BufferedResponseWriter struct {
		statusCode int
		header     http.Header
		bodyBuf    *bytes.Buffer
	}
)

func NewBufferedResponseWriter() *BufferedResponseWriter {
	return &BufferedResponseWriter{
		header:  make(http.Header),
		bodyBuf: &bytes.Buffer{},
	}
}

func (w *BufferedResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.bodyBuf.Write(p)
}

func (w *BufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w BufferedResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w BufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w BufferedResponseWriter) Body() []byte {
	return w.bodyBuf.Bytes()
}
