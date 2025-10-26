package httplib_test

import (
	"bytes"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestLog(t *testing.T) {
	tests := []struct {
		name       string
		timestamp  time.Time
		newRequest func() *http.Request
		want       httplib.RequestLog
	}{
		{
			name: "request is nil",
			newRequest: func() *http.Request {
				return nil
			},
			want: httplib.RequestLog{},
		},
		{
			name:      "URL is not nil",
			timestamp: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			newRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://example.com/path?x=1", bytes.NewBufferString("body"))
				r.Header.Set("User-Agent", "ua/1.0")
				r.Header.Set("Referer", "https://ref.example.com/page")
				r.RemoteAddr = "192.0.2.10:12345"
				return r
			},
			want: httplib.RequestLog{
				Timestamp:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Method:        http.MethodPost,
				URL:           "https://example.com/path?x=1",
				ContentLength: 4,
				Proto:         "HTTP/1.1",
				Host:          "example.com",
				RemoteAddr:    "192.0.2.10:12345",
				UserAgent:     "ua/1.0",
				RequestURI:    "https://example.com/path?x=1",
				Referer:       "https://ref.example.com/page",
			},
		},
		{
			name:      "URL is nil",
			timestamp: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			newRequest: func() *http.Request {
				r := &http.Request{
					Method:        http.MethodGet,
					Proto:         "HTTP/1.1",
					ContentLength: 0,
					Host:          "example.net",
					// URL:       nil,
					RemoteAddr: "198.51.100.5:8080",
					Header:     make(http.Header),
					RequestURI: "/no-url",
				}
				r.Header.Set("User-Agent", "ua/2.0")
				r.Header.Set("Referer", "")
				return r
			},
			want: httplib.RequestLog{
				Timestamp:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Method:        http.MethodGet,
				URL:           "",
				ContentLength: 0,
				Proto:         "HTTP/1.1",
				Host:          "example.net",
				RemoteAddr:    "198.51.100.5:8080",
				UserAgent:     "ua/2.0",
				RequestURI:    "/no-url",
				Referer:       "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.newRequest()
			got := httplib.NewRequestLog(r, tt.timestamp)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequestLog_ToAttr(t *testing.T) {
	tests := []struct {
		name string
		log  *httplib.RequestLog
		want slog.Attr
	}{
		{
			name: "not-nil",
			log: &httplib.RequestLog{
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
			},
			want: slog.GroupAttrs("http_request",
				slog.String("timestamp", time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)),
				slog.String("method", http.MethodGet),
				slog.String("url", "https://example.com/a?b=c"),
				slog.String("host", "example.com"),
				slog.String("request_uri", "/a?b=c"),
				slog.Int64("content_length", 123),
				slog.String("proto", "HTTP/2.0"),
				slog.String("remote_addr", "203.0.113.1:4444"),
				slog.String("user_agent", "ua/3.0"),
				slog.String("referer", "https://ref.example.com/"),
			),
		},
		{
			name: "nil",
			log:  nil,
			want: slog.Attr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.log.ToAttr())
		})
	}
}

func TestRequestLog_GetIP(t *testing.T) {
	tests := []struct {
		name string
		log  *httplib.RequestLog
		want net.IP
	}{
		{
			name: "nil",
			log:  nil,
			want: nil,
		},
		{
			name: "invalid remote addr (no port)",
			log:  &httplib.RequestLog{RemoteAddr: "invalid"},
			want: nil,
		},
		{
			name: "unparsable host",
			log:  &httplib.RequestLog{RemoteAddr: "notanip:80"},
			want: nil,
		},
		{
			name: "valid IPv4",
			log:  &httplib.RequestLog{RemoteAddr: "192.0.2.1:1234"},
			want: net.ParseIP("192.0.2.1").To4(),
		},
		{
			name: "valid IPv6 (with brackets)",
			log:  &httplib.RequestLog{RemoteAddr: "[2001:db8::1]:443"},
			want: net.ParseIP("2001:db8::1").To16(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.log.GetIP())
		})
	}
}
