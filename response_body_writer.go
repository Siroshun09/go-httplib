package httplib

import (
	"net/http"
)

// responseBodyWriter implements io.Writer using http.ResponseWriter and counts the number of bytes written.
type responseBodyWriter struct {
	w            http.ResponseWriter
	responseSize int64
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	w.responseSize += int64(n)
	return n, err
}
