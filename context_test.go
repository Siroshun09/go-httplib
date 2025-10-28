package httplib_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
)

func TestContext_RequestLog(t *testing.T) {
	tests := []struct {
		name    string
		ctxFunc func(ctx context.Context) context.Context
		want    httplib.RequestLog
	}{
		{
			name: "no request log in context",
			ctxFunc: func(ctx context.Context) context.Context {
				return ctx
			},
			want: httplib.RequestLog{},
		},
		{
			name: "set request log and ensure copy is returned",
			ctxFunc: func(ctx context.Context) context.Context {
				rl := httplib.RequestLog{Method: http.MethodGet, Host: "example.com"}
				ctx = httplib.WithRequestLog(ctx, rl)
				// mutate original after storing to context to ensure value was copied
				rl.Method = http.MethodPost
				return ctx
			},
			want: httplib.RequestLog{Method: http.MethodGet, Host: "example.com"},
		},
		{
			name: "returns zero value when only response log is set",
			ctxFunc: func(ctx context.Context) context.Context {
				ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
				return ctx
			},
			want: httplib.RequestLog{},
		},
		{
			name: "wrong type in context",
			ctxFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, httplib.ContextKeyRequestLog, "wrong value")
			},
			want: httplib.RequestLog{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.ctxFunc(t.Context())
			assert.Equal(t, tt.want, httplib.GetRequestLogFromContext(ctx))
		})
	}
}

func TestContext_ResponseLogPtr(t *testing.T) {
	t.Run("not nil pointer", func(t *testing.T) {
		ctx := t.Context()

		// expect nil at first
		assert.Nil(t, httplib.GetResponseLogPtrFromContext(ctx))

		log := httplib.ResponseLog{
			StatusCode:   http.StatusOK,
			ResponseSize: 10,
		}
		ctx = httplib.WithResponseLogPtr(ctx, &log)

		// expect not-nil after storing, and the pointer is the same as the original
		ptr := httplib.GetResponseLogPtrFromContext(ctx)
		assert.NotNil(t, ptr)
		assert.Same(t, &log, ptr)

		// mutate through the pointer and ensure the change is reflected to the log ptr
		ptr.StatusCode = http.StatusInternalServerError
		assert.Equal(t, http.StatusInternalServerError, log.StatusCode)
		assert.Equal(t, http.StatusInternalServerError, httplib.GetResponseLogPtrFromContext(ctx).StatusCode)
	})

	t.Run("nil pointer", func(t *testing.T) {
		ctx := t.Context()

		// not initialized
		assert.Nil(t, httplib.GetResponseLogPtrFromContext(ctx))

		// explicitly set to nil
		ctx = httplib.WithResponseLogPtr(ctx, nil)
		assert.Nil(t, httplib.GetResponseLogPtrFromContext(ctx))

		// set to not-nil and then set to nil
		ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
		ctx = httplib.WithResponseLogPtr(ctx, nil)
		assert.Nil(t, httplib.GetResponseLogPtrFromContext(ctx))
	})
}
