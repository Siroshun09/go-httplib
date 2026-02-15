package httplog

import (
	"context"
	"log/slog"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/logs/v2"
)

// NewHTTPAttrLogger creates a new logger that adds slog.Attr of httplib.RequestLog and httplib.ResponseLog.
func NewHTTPAttrLogger(delegate logs.Logger) logs.Logger {
	if delegate == nil {
		panic("delegate cannot be nil")
	}
	return &logger{delegate: delegate}
}

type logger struct {
	delegate logs.Logger
}

func (l *logger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.delegate.Debug(ctx, msg, l.appendAttrs(ctx, attrs)...)
}

func (l *logger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.delegate.Info(ctx, msg, l.appendAttrs(ctx, attrs)...)
}

func (l *logger) Warn(ctx context.Context, err error, attrs ...slog.Attr) {
	l.delegate.Warn(ctx, err, l.appendAttrs(ctx, attrs)...)
}

func (l *logger) Error(ctx context.Context, err error, attrs ...slog.Attr) {
	l.delegate.Error(ctx, err, l.appendAttrs(ctx, attrs)...)
}

func (l *logger) appendAttrs(ctx context.Context, attrs []slog.Attr) []slog.Attr {
	if ctx == nil {
		return attrs
	}

	requestLog := httplib.GetRequestLogFromContext(ctx)
	responseLog := httplib.GetResponseLogPtrFromContext(ctx)
	latency := httplib.GetLatencyFromContext(ctx)
	return append(attrs, requestLog.ToAttr(), responseLog.ToAttr(latency))
}
