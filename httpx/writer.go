package httpx

import (
	"net/http"
)

type StatusAwareResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *StatusAwareResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *StatusAwareResponseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func (w *StatusAwareResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *StatusAwareResponseWriter) Status() int {
	return w.statusCode
}

type InMemResponseWriter struct {
	header     http.Header
	StatusCode int
	Body       []byte
}

func (i *InMemResponseWriter) Header() http.Header {
	return i.header
}

func (i *InMemResponseWriter) Write(bytes []byte) (int, error) {
	i.Body = append(i.Body, bytes...)
	return len(bytes), nil
}

func (i *InMemResponseWriter) WriteHeader(statusCode int) {
	i.StatusCode = statusCode
}

func NewInMemResponseWriter() *InMemResponseWriter {
	return &InMemResponseWriter{
		header: make(http.Header),
	}
}
