package httplib

import (
	"net/http"
)

const ContextKeyRequestLog = contextKeyRequestLog

func NewHandlerInfoFromPC(pc uintptr, file string, line int) HandlerInfo {
	return newHandlerInfoFromPC(pc, file, line)
}

type ResponseBodyWriter struct {
	responseBodyWriter
}

func NewResponseBodyWriter(w http.ResponseWriter) *ResponseBodyWriter {
	return &ResponseBodyWriter{responseBodyWriter: responseBodyWriter{w: w}}
}

func (w *ResponseBodyWriter) ResponseSize() int64 {
	return w.responseSize
}
