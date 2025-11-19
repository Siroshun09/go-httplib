package httplog

import (
	"context"
	"log/slog"

	"github.com/Siroshun09/go-httplib"
)

type httpAttrHandler struct {
	delegate slog.Handler
}

func NewHttpAttrHandler(delegate slog.Handler) slog.Handler {
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
	return NewHttpAttrHandler(h.delegate.WithAttrs(attrs))
}

func (h httpAttrHandler) WithGroup(name string) slog.Handler {
	return NewHttpAttrHandler(h.delegate.WithGroup(name))
}
