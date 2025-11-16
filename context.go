package httplib

import (
	"context"
	"time"
)

type contextKey uint8

const (
	contextKeyRequestLog contextKey = iota
	contextKeyResponseLog
	contextKeyLatency
)

// GetRequestLogFromContext returns the RequestLog stored in the context.
//
// If the context does not contain a request log, or the stored value is nil, it returns the zero-value RequestLog.
func GetRequestLogFromContext(ctx context.Context) RequestLog {
	requestLog, ok := ctx.Value(contextKeyRequestLog).(*RequestLog)
	if !ok || requestLog == nil {
		return RequestLog{}
	}
	return *requestLog
}

// WithRequestLog returns a new context that carries a copy of the given RequestLog.
//
// The RequestLog value is stored as a pointer internally, but this function
// takes the value by copy to avoid later mutations of the original affecting
// the stored value.
func WithRequestLog(ctx context.Context, requestLog RequestLog) context.Context {
	return context.WithValue(ctx, contextKeyRequestLog, &requestLog)
}

// GetResponseLogPtrFromContext returns the ResponseLog pointer stored in the context.
//
// If the context does not contain a response log, or the stored value is nil,
// it returns nil. The returned pointer is the same instance stored in the
// context, so modifying the pointed value will be reflected in later reads.
//
// Note: ResponseLog is intended for use within a single HTTP request handler chain
// and is not safe for concurrent access from multiple goroutines. Each HTTP request
// should maintain its own ResponseLog instance.
func GetResponseLogPtrFromContext(ctx context.Context) *ResponseLog {
	holder, ok := ctx.Value(contextKeyResponseLog).(*ResponseLog)
	if !ok {
		return nil
	}
	return holder
}

// WithResponseLogPtr returns a new context that carries the provided ResponseLog pointer.
//
// Storing a nil pointer is allowed and can be used to explicitly mark absence.
// The pointer is not copied; later modifications through the pointer are
// observable via future reads from the context.
//
// Note: The ResponseLog should only be accessed within a single HTTP request handler
// chain. Concurrent access from multiple goroutines is not supported and may result
// in data races.
func WithResponseLogPtr(ctx context.Context, resPtr *ResponseLog) context.Context {
	return context.WithValue(ctx, contextKeyResponseLog, resPtr)
}

// GetLatencyFromContext returns the latency stored in the context.
//
// If the context does not contain latency, it returns 0.
func GetLatencyFromContext(ctx context.Context) time.Duration {
	latency, ok := ctx.Value(contextKeyLatency).(time.Duration)
	if !ok {
		return 0
	}
	return latency
}

// WithLatency returns a new context that carries the provided latency.
func WithLatency(ctx context.Context, latency time.Duration) context.Context {
	return context.WithValue(ctx, contextKeyLatency, latency)
}
