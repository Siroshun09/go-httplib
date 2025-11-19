package httplog

import (
	"log/slog"
)

func IsHttpAttrHandler(h slog.Handler) bool {
	_, ok := h.(*httpAttrHandler)
	return ok
}
