package httplib

import (
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"time"
)

// RequestLog represents structured HTTP request information for logging purposes.
//
// It captures key details from an http.Request at a specific point in time.
type RequestLog struct {
	// Timestamp is the time when the request was observed/logged.
	//
	// It is typically the time right before the handler starts processing the request.
	Timestamp time.Time

	// Method is the HTTP method of the request (e.g., "GET", "POST").
	Method string

	// URL is the full request URL as a string.
	//
	// This will be an empty string if the original request's URL was nil.
	URL string

	// ContentLength is the declared size of the request body in bytes.
	//
	// A value of -1 indicates that the length is unknown (see http.Request.ContentLength).
	ContentLength int64

	// Proto is the HTTP protocol version used by the client (e.g., "HTTP/1.1", "HTTP/2").
	Proto string

	// Host is the value of the request host (usually from the Host header).
	Host string

	// RemoteAddr is the client address in the form "IP:port" as reported by the server.
	//
	// Use GetIP to extract and parse the IP component.
	RemoteAddr string

	// UserAgent is the client user agent string.
	UserAgent string

	// RequestURI is the unmodified request-target as sent by the client.
	//
	// It may include the path and query string.
	RequestURI string

	// Referer is the URL of the resource from which the request originated.
	//
	// It is taken from the "Referer" header and may be empty.
	Referer string
}

// NewRequestLog creates a RequestLog from an http.Request and timestamp.
//
// If r is nil, the returned RequestLog will be empty.
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
// Returns a grouped slog.Attr with key "http_request" containing:
//   - timestamp: request timestamp in RFC3339 format
//   - method: HTTP method
//   - url: full request URL
//   - host: request host
//   - request_uri: unmodified request-target
//   - content_length: request body size in bytes (-1 if unknown)
//   - proto: HTTP protocol version
//   - remote_addr: client address (IP:port)
//   - user_agent: client user agent string
//   - referer: referring URL
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

// GetAddr extracts and parses the IP address from RemoteAddr.
//
// Returns an empty netip.Addr if the RequestLog is nil, the address cannot be parsed, or the host portion is not a valid IP address.
func (l *RequestLog) GetAddr() netip.Addr {
	if l == nil {
		return netip.Addr{}
	}

	addrPort, err := netip.ParseAddrPort(l.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}

	return addrPort.Addr()
}
