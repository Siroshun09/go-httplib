package httplog

import (
	"log/slog"
)

func IsHTTPAttrHandler(h slog.Handler) bool {
	_, ok := h.(*httpAttrHandler)
	return ok
}
