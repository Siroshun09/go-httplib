package httplib

import (
	"log/slog"
	"runtime"
	"time"
)

// ResponseLog represents structured HTTP response information for logging purposes.
//
// It captures key details about an HTTP response at a specific point in time.
type ResponseLog struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// ResponseSize is the size of the response body in bytes.
	// A value of -1 indicates that the size is unknown or not applicable.
	ResponseSize int64

	// Error is any error that occurred during request processing.
	Error error

	// HandlerInfo contains metadata about the handler that processed the request.
	HandlerInfo HandlerInfo
}

// ToAttr converts the ResponseLog to a structured slog.Attr for logging.
//
// Returns a grouped slog.Attr with a key "http_response" containing:
//   - latency: request processing time in milliseconds
//   - status_code: HTTP status code
//   - response_size: response body size in bytes
//   - error: error message (included only if Error is not nil)
//   - handler: handler information (included only if HandlerInfo.FuncName is not empty)
//
// Returns an empty slog.Attr if the ResponseLog is nil.
func (r *ResponseLog) ToAttr(latency time.Duration) slog.Attr {
	if r == nil {
		return slog.Attr{}
	}

	attrs := make([]slog.Attr, 0, 4)

	attrs = append(
		attrs,
		slog.Int64("latency", latency.Milliseconds()),
		slog.Int("status_code", r.StatusCode),
		slog.Int64("response_size", r.ResponseSize),
	)

	if r.Error != nil {
		attrs = append(attrs, slog.String("error", r.Error.Error()))
	}

	if r.HandlerInfo.FuncName != "" { // include HandlerInfo if it is initialized, even if it is UnknownHandlerInfo
		attrs = append(attrs, r.HandlerInfo.ToAttr())
	}

	return slog.GroupAttrs("http_response", attrs...)
}

// HandlerInfo holds metadata about the HTTP handler function that processed the request.
//
// It includes the fully qualified function name and source location (file and line).
// When unavailable, use UnknownHandlerInfo.
type HandlerInfo struct {
	// FuncName is the fully qualified name of the handler function.
	FuncName string

	// File is the source file path where the handler is defined.
	File string

	// Line is the line number in the source file where the handler is defined.
	Line int
}

// ToAttr converts the HandlerInfo to a structured slog.Attr for logging.
//
// Returns a grouped slog.Attr with key "handler" containing:
//   - func_name: fully qualified function name
//   - file: source file path
//   - line: line number in the source file
//
// Returns an empty slog.Attr if the HandlerInfo is nil.
func (h *HandlerInfo) ToAttr() slog.Attr {
	if h == nil {
		return slog.Attr{}
	}

	return slog.GroupAttrs("handler",
		slog.String("func_name", h.FuncName),
		slog.String("file", h.File),
		slog.Int("line", h.Line),
	)
}

var unknownHandlerInfo = HandlerInfo{
	FuncName: "unknown",
	File:     "unknown",
	Line:     0,
}

// UnknownHandlerInfo returns a HandlerInfo representing an unknown handler.
func UnknownHandlerInfo() HandlerInfo {
	return unknownHandlerInfo
}

// NewHandlerInfo returns handler information for the caller.
//
// The skip parameter specifies the number of stack frames to skip before recording,
// where 0 identifies the caller of NewHandlerInfo.
func NewHandlerInfo(skip int) HandlerInfo {
	pc, file, line, ok := runtime.Caller(skip + 1) // skip 1 for this function
	if !ok {
		return UnknownHandlerInfo()
	}

	return newHandlerInfoFromPC(pc, file, line)
}

// newHandlerInfoFromPC creates a HandlerInfo from the given program counter and source location.
//
// Returns UnknownHandlerInfo if the program counter is invalid or does not correspond to a known function.
func newHandlerInfoFromPC(pc uintptr, file string, line int) HandlerInfo {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return UnknownHandlerInfo()
	}

	return HandlerInfo{
		FuncName: fn.Name(),
		File:     file,
		Line:     line,
	}
}
