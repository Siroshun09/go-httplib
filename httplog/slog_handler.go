package httplog

import (
	"context"
	"log/slog"

	"github.com/Siroshun09/go-httplib"
)

type httpAttrHandler struct {
	delegate slog.Handler
}

// NewHTTPAttrHandler creates a new handler that adds slog.Attr of httplib.RequestLog and httplib.ResponseLog to the log record.
func NewHTTPAttrHandler(delegate slog.Handler) slog.Handler {
	if delegate == nil {
		panic("delegate cannot be nil")
	}
	return &httpAttrHandler{delegate: delegate}
}

func (h httpAttrHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.delegate.Enabled(ctx, level)
}

func (h httpAttrHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		return h.delegate.Handle(ctx, record)
	}

	requestLog := httplib.GetRequestLogFromContext(ctx)
	responseLog := httplib.GetResponseLogPtrFromContext(ctx)
	latency := httplib.GetLatencyFromContext(ctx)

	record.AddAttrs(
		requestLog.ToAttr(),
		responseLog.ToAttr(latency),
	)

	return h.delegate.Handle(ctx, record)
}

func (h httpAttrHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewHTTPAttrHandler(h.delegate.WithAttrs(attrs))
}

func (h httpAttrHandler) WithGroup(name string) slog.Handler {
	return NewHTTPAttrHandler(h.delegate.WithGroup(name))
}
