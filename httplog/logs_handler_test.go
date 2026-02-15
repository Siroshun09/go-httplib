package httplog_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/go-httplib/httplog"
	"github.com/Siroshun09/logs/logmock/v2"
	"github.com/Siroshun09/logs/v2"
	"github.com/Siroshun09/logs/v2/plain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewHTTPAttrLogger(t *testing.T) {
	tests := []struct {
		name           string
		delegate       logs.Logger
		panicAssertion assert.PanicAssertionFunc
	}{
		{
			name:           "success",
			delegate:       plain.NewPlainLogger(io.Discard, nil),
			panicAssertion: assert.NotPanics,
		},
		{
			name:           "panic: delegate is nil",
			delegate:       nil,
			panicAssertion: assert.Panics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.panicAssertion(t, func() {
				got := httplog.NewHTTPAttrLogger(tt.delegate)
				assert.True(t, httplog.IsHTTPAttrLogger(got))
			})
		})
	}
}

func Test_logger_Log(t *testing.T) {
	type ctxCase struct {
		name              string
		ctxFunc           func(ctx context.Context) context.Context
		expectedAttrsFunc func() []slog.Attr
	}

	ctxCases := []ctxCase{
		{
			name: "nil context",
			ctxFunc: func(ctx context.Context) context.Context {
				return nil
			},
			expectedAttrsFunc: func() []slog.Attr {
				return nil
			},
		},
		{
			name: "no context value",
			ctxFunc: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedAttrsFunc: func() []slog.Attr {
				var requestLog httplib.RequestLog
				var responseLog *httplib.ResponseLog
				return []slog.Attr{
					requestLog.ToAttr(),
					responseLog.ToAttr(0),
				}
			},
		},
		{
			name: "only request log",
			ctxFunc: func(ctx context.Context) context.Context {
				return httplib.WithRequestLog(ctx, testRequestLog)
			},
			expectedAttrsFunc: func() []slog.Attr {
				var responseLog *httplib.ResponseLog
				return []slog.Attr{
					testRequestLog.ToAttr(),
					responseLog.ToAttr(0),
				}
			},
		},
		{
			name: "only response log",
			ctxFunc: func(ctx context.Context) context.Context {
				copiedResponseLog := testResponseLog
				return httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
			},
			expectedAttrsFunc: func() []slog.Attr {
				var requestLog httplib.RequestLog
				return []slog.Attr{
					requestLog.ToAttr(),
					testResponseLog.ToAttr(0),
				}
			},
		},
		{
			name: "only latency",
			ctxFunc: func(ctx context.Context) context.Context {
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			expectedAttrsFunc: func() []slog.Attr {
				var requestLog httplib.RequestLog
				var responseLog *httplib.ResponseLog
				return []slog.Attr{
					requestLog.ToAttr(),
					responseLog.ToAttr(10 * time.Second),
				}
			},
		},
		{
			name: "both request and response log without latency",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithRequestLog(ctx, testRequestLog)
				copiedResponseLog := testResponseLog
				return httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
			},
			expectedAttrsFunc: func() []slog.Attr {
				return []slog.Attr{
					testRequestLog.ToAttr(),
					testResponseLog.ToAttr(0),
				}
			},
		},
		{
			name: "both response log and latency",
			ctxFunc: func(ctx context.Context) context.Context {
				copiedResponseLog := testResponseLog
				ctx = httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			expectedAttrsFunc: func() []slog.Attr {
				var requestLog httplib.RequestLog
				return []slog.Attr{
					requestLog.ToAttr(),
					testResponseLog.ToAttr(10 * time.Second),
				}
			},
		},
		{
			name: "request log and response log with latency",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithRequestLog(ctx, testRequestLog)
				copiedResponseLog := testResponseLog
				ctx = httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			expectedAttrsFunc: func() []slog.Attr {
				return []slog.Attr{
					testRequestLog.ToAttr(),
					testResponseLog.ToAttr(10 * time.Second),
				}
			},
		},
	}

	attrsList := [][]slog.Attr{
		nil,
		{},
		{
			slog.String("key", "value"),
		},
	}

	for _, c := range ctxCases {
		t.Run(c.name, func(t *testing.T) {
			for _, attrs := range attrsList {
				t.Run(fmt.Sprintf("%#v", attrs), func(t *testing.T) {
					ctx := t.Context()
					ctx = c.ctxFunc(ctx)

					attrsCond := gomock.Cond(func(x []slog.Attr) bool {
						var expectedAttrs []slog.Attr
						expectedAttrs = append(expectedAttrs, attrs...)
						expectedAttrs = append(expectedAttrs, c.expectedAttrsFunc()...)

						if len(expectedAttrs) == 0 {
							return assert.Empty(t, x)
						}
						return assert.Equal(t, expectedAttrs, x)
					})

					t.Run("Debug", func(t *testing.T) {
						ctrl := gomock.NewController(t)
						mock := logmock.NewMockLogger(ctrl)

						mock.EXPECT().Debug(ctx, "test", attrsCond)

						l := httplog.NewHTTPAttrLogger(mock)
						l.Debug(ctx, "test", attrs...)
					})

					t.Run("Info", func(t *testing.T) {
						ctrl := gomock.NewController(t)
						mock := logmock.NewMockLogger(ctrl)

						mock.EXPECT().Info(ctx, "test", attrsCond)

						l := httplog.NewHTTPAttrLogger(mock)
						l.Info(ctx, "test", attrs...)
					})

					t.Run("Warn", func(t *testing.T) {
						ctrl := gomock.NewController(t)
						mock := logmock.NewMockLogger(ctrl)
						err := errors.New("test error")

						mock.EXPECT().Warn(ctx, err, attrsCond)

						l := httplog.NewHTTPAttrLogger(mock)
						l.Warn(ctx, err, attrs...)
					})

					t.Run("Error", func(t *testing.T) {
						ctrl := gomock.NewController(t)
						mock := logmock.NewMockLogger(ctrl)
						err := errors.New("test error")

						mock.EXPECT().Error(ctx, err, attrsCond)

						l := httplog.NewHTTPAttrLogger(mock)
						l.Error(ctx, err, attrs...)
					})
				})
			}
		})
	}
}

func TestIsHttpAttrLogger(t *testing.T) {
	tests := []struct {
		name string
		l    logs.Logger
		want bool
	}{
		{
			name: "nil",
			l:    nil,
			want: false,
		},
		{
			name: "not an http attr handler",
			l:    plain.NewPlainLogger(io.Discard, nil),
			want: false,
		},
		{
			name: "http attr handler",
			l:    httplog.NewHTTPAttrLogger(plain.NewPlainLogger(io.Discard, nil)),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, httplog.IsHTTPAttrLogger(tt.l))
		})
	}
}
