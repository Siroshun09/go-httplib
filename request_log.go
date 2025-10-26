package httplib

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

// RequestLog represents structured HTTP request information for logging purposes.
//
// It captures key details from an http.Request at a specific point in time.
type RequestLog struct {
	Timestamp     time.Time
	Method        string
	URL           string
	ContentLength int64
	Proto         string
	Host          string
	RemoteAddr    string
	UserAgent     string
	RequestURI    string
	Referer       string
}

// NewRequestLog creates a RequestLog from an http.Request and timestamp.
//
// If r.URL is nil, the RequestLog.URL will be an empty string.
func NewRequestLog(r *http.Request, timestamp time.Time) RequestLog {
	if r == nil {
		return RequestLog{}
	}

	url := ""
	if r.URL != nil {
		url = r.URL.String()
	}

	return RequestLog{
		Timestamp:     timestamp,
		Method:        r.Method,
		URL:           url,
		ContentLength: r.ContentLength,
		Proto:         r.Proto,
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
		UserAgent:     r.UserAgent(),
		RequestURI:    r.RequestURI,
		Referer:       r.Referer(),
	}
}

// ToAttr converts the RequestLog to a structured slog.Attr for logging.
//
// Returns an empty slog.Attr if the RequestLog is nil.
func (l *RequestLog) ToAttr() slog.Attr {
	if l == nil {
		return slog.Attr{}
	}

	return slog.GroupAttrs("http_request",
		slog.String("timestamp", l.Timestamp.Format(time.RFC3339)),
		slog.String("method", l.Method),
		slog.String("url", l.URL),
		slog.String("host", l.Host),
		slog.String("request_uri", l.RequestURI),
		slog.Int64("content_length", l.ContentLength),
		slog.String("proto", l.Proto),
		slog.String("remote_addr", l.RemoteAddr),
		slog.String("user_agent", l.UserAgent),
		slog.String("referer", l.Referer),
	)
}

// GetIP extracts and parses the IP address from RemoteAddr.
//
// Returns nil if the RequestLog is nil, the address cannot be parsed, or the host portion is not a valid IP address.
// IPv4 addresses are returned as 4-byte representation, IPv6 as 16-byte.
func (l *RequestLog) GetIP() net.IP {
	if l == nil {
		return nil
	}

	host, _, err := net.SplitHostPort(l.RemoteAddr)
	if err != nil {
		return nil
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}

	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4
	}

	return ip.To16()
}
