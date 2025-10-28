package httplib_test

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseLog_ToAttr(t *testing.T) {
	tests := []struct {
		name     string
		Response *httplib.ResponseLog
		latency  time.Duration
		want     slog.Attr
	}{
		{
			name: "status code is StatusOK",
			Response: &httplib.ResponseLog{
				StatusCode:   http.StatusOK,
				ResponseSize: 100,
				Error:        nil,
				HandlerInfo: httplib.HandlerInfo{
					FuncName: "func_status_ok",
					File:     "file_status_ok.go",
					Line:     1,
				},
			},
			latency: 123 * time.Millisecond,
			want: slog.GroupAttrs("http_response",
				slog.Int64("latency", 123),
				slog.Int("status_code", http.StatusOK),
				slog.Int64("response_size", 100),
				slog.GroupAttrs("handler",
					slog.String("func_name", "func_status_ok"),
					slog.String("file", "file_status_ok.go"),
					slog.Int("line", 1),
				),
			),
		},
		{
			name: "status code is StatusInternalServerError",
			Response: &httplib.ResponseLog{
				StatusCode:   http.StatusInternalServerError,
				ResponseSize: 100,
				Error:        errors.New("internal server error"),
				HandlerInfo: httplib.HandlerInfo{
					FuncName: "func_status_internal_server_error",
					File:     "file_status_internal_server_error.go",
					Line:     2,
				},
			},
			latency: 1 * time.Second,
			want: slog.GroupAttrs("http_response",
				slog.Int64("latency", 1000),
				slog.Int("status_code", http.StatusInternalServerError),
				slog.Int64("response_size", 100),
				slog.String("error", "internal server error"),
				slog.GroupAttrs("handler",
					slog.String("func_name", "func_status_internal_server_error"),
					slog.String("file", "file_status_internal_server_error.go"),
					slog.Int("line", 2),
				),
			),
		},
		{
			name: "response size is 0",
			Response: &httplib.ResponseLog{
				StatusCode:   http.StatusOK,
				ResponseSize: 0,
				Error:        nil,
				HandlerInfo: httplib.HandlerInfo{
					FuncName: "func_zero_response_size",
					File:     "func_zero_response_size.go",
					Line:     1,
				},
			},
			latency: 123 * time.Millisecond,
			want: slog.GroupAttrs("http_response",
				slog.Int64("latency", 123),
				slog.Int("status_code", http.StatusOK),
				slog.Int64("response_size", 0),
				slog.GroupAttrs("handler",
					slog.String("func_name", "func_zero_response_size"),
					slog.String("file", "func_zero_response_size.go"),
					slog.Int("line", 1),
				),
			),
		},
		{
			name: "response size is -1",
			Response: &httplib.ResponseLog{
				StatusCode:   http.StatusOK,
				ResponseSize: -1,
				Error:        nil,
				HandlerInfo: httplib.HandlerInfo{
					FuncName: "func_minus_response_size",
					File:     "func_minus_response_size.go",
					Line:     1,
				},
			},
			latency: 123 * time.Millisecond,
			want: slog.GroupAttrs("http_response",
				slog.Int64("latency", 123),
				slog.Int("status_code", http.StatusOK),
				slog.Int64("response_size", -1),
				slog.GroupAttrs("handler",
					slog.String("func_name", "func_minus_response_size"),
					slog.String("file", "func_minus_response_size.go"),
					slog.Int("line", 1),
				),
			),
		},
		{
			name: "Error and HandlerInfo is not initialized",
			Response: &httplib.ResponseLog{
				StatusCode:   http.StatusOK,
				ResponseSize: 100,
				Error:        nil,
			},
			latency: 123 * time.Millisecond,
			want: slog.GroupAttrs("http_response",
				slog.Int64("latency", 123),
				slog.Int("status_code", http.StatusOK),
				slog.Int64("response_size", 100),
			),
		},
		{
			name: "HandlerInfo is not initialized",
			Response: &httplib.ResponseLog{
				StatusCode:   http.StatusInternalServerError,
				ResponseSize: 100,
				Error:        errors.New("internal server error"),
			},
			latency: 123 * time.Millisecond,
			want: slog.GroupAttrs("http_response",
				slog.Int64("latency", 123),
				slog.Int("status_code", http.StatusInternalServerError),
				slog.Int64("response_size", 100),
				slog.String("error", "internal server error"),
			),
		},
		{
			name:    "nil",
			latency: 123,
			want:    slog.Attr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.Response.ToAttr(tt.latency))
		})
	}
}

func TestHandlerInfo_ToAttr(t *testing.T) {
	tests := []struct {
		name        string
		HandlerInfo *httplib.HandlerInfo
		want        slog.Attr
	}{
		{
			name: "not nil",
			HandlerInfo: &httplib.HandlerInfo{
				FuncName: "func_name",
				File:     "file",
				Line:     1,
			},
			want: slog.GroupAttrs("handler",
				slog.String("func_name", "func_name"),
				slog.String("file", "file"),
				slog.Int("line", 1),
			),
		},
		{
			name:        "unknown",
			HandlerInfo: toPtr(httplib.UnknownHandlerInfo()),
			want: slog.GroupAttrs("handler",
				slog.String("func_name", "unknown"),
				slog.String("file", "unknown"),
				slog.Int("line", 0),
			),
		},
		{
			name:        "empty",
			HandlerInfo: &httplib.HandlerInfo{},
			want: slog.GroupAttrs("handler",
				slog.String("func_name", ""),
				slog.String("file", ""),
				slog.Int("line", 0),
			),
		},
		{
			name:        "nil",
			HandlerInfo: nil,
			want:        slog.Attr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.HandlerInfo.ToAttr())
		})
	}
}

func TestNewHandlerInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pc, file, line, ok := runtime.Caller(0) // this function
		require.True(t, ok)

		fn := runtime.FuncForPC(pc)
		require.NotNil(t, fn)
		want := httplib.HandlerInfo{
			FuncName: fn.Name(),
			File:     file,
			Line:     line,
		}

		got := httplib.NewHandlerInfo(0)
		assert.NotZero(t, got.Line)
		got.Line = line // ignore line diff
		assert.Equal(t, want, got)
	})

	t.Run("failure", func(t *testing.T) {
		assert.Equal(t, httplib.UnknownHandlerInfo(), httplib.NewHandlerInfo(math.MaxInt16))
	})
}

func Test_newHandlerInfoFromPC(t *testing.T) {
	pc, file, line, ok := runtime.Caller(0) // this function
	require.True(t, ok)

	tests := []struct {
		name string
		pc   uintptr
		file string
		line int
		want httplib.HandlerInfo
	}{
		{
			name: "success",
			pc:   pc,
			file: file,
			line: line,
			want: httplib.HandlerInfo{
				FuncName: "github.com/Siroshun09/go-httplib_test.Test_newHandlerInfoFromPC",
				File:     file,
				Line:     line,
			},
		},
		{
			name: "failure",
			pc:   0,
			file: "filename",
			line: 1,
			want: httplib.UnknownHandlerInfo(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, httplib.NewHandlerInfoFromPC(tt.pc, tt.file, tt.line))
		})
	}
}
