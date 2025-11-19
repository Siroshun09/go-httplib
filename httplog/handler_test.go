package httplog_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/go-httplib/httplog"
	"github.com/stretchr/testify/assert"
)

func Test_httpAttrHandler_Enabled(t *testing.T) {
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})
	tests := []struct {
		name     string
		delegate slog.Handler
		level    slog.Level
		want     bool
	}{
		{
			name:     "slog.LevelDebug",
			delegate: handler,
			level:    slog.LevelDebug,
			want:     false,
		},
		{
			name:     "slog.LevelInfo",
			delegate: handler,
			level:    slog.LevelInfo,
			want:     false,
		},
		{
			name:     "slog.LevelWarn",
			delegate: handler,
			level:    slog.LevelWarn,
			want:     true,
		},
		{
			name:     "slog.LevelError",
			delegate: handler,
			level:    slog.LevelError,
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			assert.Equal(t, tt.want, httplog.NewHttpAttrHandler(tt.delegate).Enabled(ctx, tt.level))
		})
	}
}

type slogHandlerMock struct {
	handleCallCount int
	handledCtx      context.Context
	handledRecord   slog.Record
}

func (s *slogHandlerMock) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (s *slogHandlerMock) Handle(ctx context.Context, record slog.Record) error {
	s.handleCallCount++
	s.handledCtx = ctx
	s.handledRecord = record
	return nil
}

func (s *slogHandlerMock) WithAttrs(_ []slog.Attr) slog.Handler {
	return s
}

func (s *slogHandlerMock) WithGroup(_ string) slog.Handler {
	return s
}

var (
	testRequestLog = httplib.RequestLog{
		Timestamp:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		Method:        http.MethodGet,
		URL:           "https://example.com/a?b=c",
		ContentLength: 123,
		Proto:         "HTTP/2.0",
		Host:          "example.com",
		RemoteAddr:    "203.0.113.1:4444",
		UserAgent:     "ua/3.0",
		RequestURI:    "/a?b=c",
		Referer:       "https://ref.example.com/",
	}

	testResponseLog = httplib.ResponseLog{
		StatusCode:   http.StatusInternalServerError,
		ResponseSize: 100,
		Error:        errors.New("internal server error"),
		HandlerInfo: httplib.HandlerInfo{
			FuncName: "func_status_internal_server_error",
			File:     "file_status_internal_server_error.go",
			Line:     2,
		},
	}
)

func Test_httpAttrHandler_Handle(t *testing.T) {
	originalRecord := slog.NewRecord(time.Date(2025, 11, 20, 1, 2, 3, 4, time.UTC), slog.LevelWarn, "test log", 0)

	tests := []struct {
		name       string
		ctxFunc    func(ctx context.Context) context.Context
		wantRecord slog.Record
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "nil context",
			ctxFunc: func(ctx context.Context) context.Context {
				return nil
			},
			wantRecord: originalRecord,
			wantErr:    assert.NoError,
		},
		{
			name: "no context value",
			ctxFunc: func(ctx context.Context) context.Context {
				return ctx
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				var requestLog httplib.RequestLog
				var responseLog *httplib.ResponseLog
				record.AddAttrs(
					requestLog.ToAttr(),
					responseLog.ToAttr(0),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "only request log",
			ctxFunc: func(ctx context.Context) context.Context {
				return httplib.WithRequestLog(ctx, testRequestLog)
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				var responseLog *httplib.ResponseLog
				record.AddAttrs(
					testRequestLog.ToAttr(),
					responseLog.ToAttr(0),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "only response log",
			ctxFunc: func(ctx context.Context) context.Context {
				copiedResponseLog := testResponseLog
				return httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				var requestLog httplib.RequestLog
				record.AddAttrs(
					requestLog.ToAttr(),
					testResponseLog.ToAttr(0),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "only latency",
			ctxFunc: func(ctx context.Context) context.Context {
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				var requestLog httplib.RequestLog
				var responseLog *httplib.ResponseLog
				record.AddAttrs(
					requestLog.ToAttr(),
					responseLog.ToAttr(10*time.Second),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "both request and response log without latency",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithRequestLog(ctx, testRequestLog)
				copiedResponseLog := testResponseLog
				return httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				record.AddAttrs(
					testRequestLog.ToAttr(),
					testResponseLog.ToAttr(0),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "both response log and latency",
			ctxFunc: func(ctx context.Context) context.Context {
				copiedResponseLog := testResponseLog
				ctx = httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				var requestLog httplib.RequestLog
				record.AddAttrs(
					requestLog.ToAttr(),
					testResponseLog.ToAttr(10*time.Second),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "request log and response log with latency",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithRequestLog(ctx, testRequestLog)
				copiedResponseLog := testResponseLog
				ctx = httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			wantRecord: func() slog.Record {
				record := originalRecord.Clone()
				record.AddAttrs(
					testRequestLog.ToAttr(),
					testResponseLog.ToAttr(10*time.Second),
				)
				return record
			}(),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx = tt.ctxFunc(ctx)

			gotRecord := originalRecord.Clone()
			mockHandler := &slogHandlerMock{}
			err := httplog.NewHttpAttrHandler(mockHandler).Handle(ctx, gotRecord)
			assert.Equal(t, 1, mockHandler.handleCallCount)
			assert.Equal(t, ctx, mockHandler.handledCtx)
			assert.Equal(t, tt.wantRecord, mockHandler.handledRecord)
			tt.wantErr(t, err)
		})
	}
}

func Test_httpAttrHandler_Handle_Log(t *testing.T) {
	tests := []struct {
		name    string
		ctxFunc func(ctx context.Context) context.Context
		call    func(ctx context.Context, logger *slog.Logger)
		want    []string
	}{
		{
			name: "no context value",
			ctxFunc: func(ctx context.Context) context.Context {
				return ctx
			},
			call: func(ctx context.Context, logger *slog.Logger) {
				logger.Info("test log")
			},
			want: []string{
				`{` +
					`"time":"2000-01-01T09:00:00+09:00","level":"INFO","msg":"test log",` +
					`"http_request":{"timestamp":"0001-01-01T00:00:00Z","method":"","url":"","host":"","request_uri":"","content_length":0,"proto":"","remote_addr":"","user_agent":"","referer":""}` +
					`}`,
			},
		},
		{
			name: "log without context",
			ctxFunc: func(ctx context.Context) context.Context {
				return ctx
			},
			call: func(ctx context.Context, logger *slog.Logger) {
				logger.Info("test log")
			},
			want: []string{
				`{` +
					`"time":"2000-01-01T09:00:00+09:00","level":"INFO","msg":"test log",` +
					`"http_request":{"timestamp":"0001-01-01T00:00:00Z","method":"","url":"","host":"","request_uri":"","content_length":0,"proto":"","remote_addr":"","user_agent":"","referer":""}` +
					`}`,
			},
		},
		{
			name: "has request and response logs with latency",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithRequestLog(ctx, testRequestLog)
				copiedResponseLog := testResponseLog
				ctx = httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			call: func(ctx context.Context, logger *slog.Logger) {
				logger.InfoContext(ctx, "test log")
			},
			want: []string{
				`{` +
					`"time":"2000-01-01T09:00:00+09:00","level":"INFO","msg":"test log",` +
					`"http_request":{"timestamp":"2024-12-31T23:59:59Z","method":"GET","url":"https://example.com/a?b=c","host":"example.com","request_uri":"/a?b=c","content_length":123,"proto":"HTTP/2.0","remote_addr":"203.0.113.1:4444","user_agent":"ua/3.0","referer":"https://ref.example.com/"},` +
					`"http_response":{"latency":10000,"status_code":500,"response_size":100,"error":"internal server error","handler":{"func_name":"func_status_internal_server_error","file":"file_status_internal_server_error.go","line":2}}` +
					`}`,
			},
		},
		{
			name: "has request and response logs with latency, and additional attrs",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithRequestLog(ctx, testRequestLog)
				copiedResponseLog := testResponseLog
				ctx = httplib.WithResponseLogPtr(ctx, &copiedResponseLog)
				return httplib.WithLatency(ctx, 10*time.Second)
			},
			call: func(ctx context.Context, logger *slog.Logger) {
				logger.InfoContext(ctx, "test log", slog.String("key1", "value"), slog.Int("key2", 1), slog.Float64("key3", 3.14), slog.Bool("key4", true))
			},
			want: []string{
				`{` +
					`"time":"2000-01-01T09:00:00+09:00","level":"INFO","msg":"test log",` +
					`"key1":"value","key2":1,"key3":3.14,"key4":true,` +
					`"http_request":{"timestamp":"2024-12-31T23:59:59Z","method":"GET","url":"https://example.com/a?b=c","host":"example.com","request_uri":"/a?b=c","content_length":123,"proto":"HTTP/2.0","remote_addr":"203.0.113.1:4444","user_agent":"ua/3.0","referer":"https://ref.example.com/"},` +
					`"http_response":{"latency":10000,"status_code":500,"response_size":100,"error":"internal server error","handler":{"func_name":"func_status_internal_server_error","file":"file_status_internal_server_error.go","line":2}}` +
					`}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := t.Context()
				ctx = tt.ctxFunc(ctx)

				buf := &strings.Builder{}
				jsonHandler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
				httpAttrHandler := httplog.NewHttpAttrHandler(jsonHandler)
				logger := slog.New(httpAttrHandler)

				tt.call(ctx, logger)
				assert.Equal(t, tt.want, strings.Split(strings.TrimRight(buf.String(), "\n"), "\n"))
			})
		})
	}
}

func Test_httpAttrHandler_WithAttrs(t *testing.T) {
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})

	tests := []struct {
		name  string
		attrs []slog.Attr
	}{
		{
			name:  "nil attrs",
			attrs: nil,
		},
		{
			name:  "empty attrs",
			attrs: []slog.Attr{},
		},
		{
			name: "single attr",
			attrs: []slog.Attr{
				slog.String("key", "value"),
			},
		},
		{
			name: "multiple attrs",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
				slog.Int("key2", 1),
				slog.Float64("key3", 3.14),
				slog.Bool("key4", true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := httplog.NewHttpAttrHandler(handler).WithAttrs(tt.attrs)
			assert.True(t, httplog.IsHttpAttrHandler(got))
		})
	}
}

func Test_httpAttrHandler_WithGroup(t *testing.T) {
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})

	tests := []struct {
		name      string
		groupName string
	}{
		{
			name:      "empty group name",
			groupName: "",
		},
		{
			name:      "non-empty group name",
			groupName: "group",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := httplog.NewHttpAttrHandler(handler).WithGroup(tt.groupName)
			assert.True(t, httplog.IsHttpAttrHandler(got))
		})
	}
}

func TestIsHttpAttrHandler(t *testing.T) {
	tests := []struct {
		name string
		h    slog.Handler
		want bool
	}{
		{
			name: "nil",
			h:    nil,
			want: false,
		},
		{
			name: "not an http attr handler",
			h:    slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}),
			want: false,
		},
		{
			name: "http attr handler",
			h:    httplog.NewHttpAttrHandler(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, httplog.IsHttpAttrHandler(tt.h))
		})
	}
}
