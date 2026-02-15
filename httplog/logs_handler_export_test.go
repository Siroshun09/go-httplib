package httplog

import (
	"github.com/Siroshun09/logs/v2"
)

func IsHTTPAttrLogger(h logs.Logger) bool {
	_, ok := h.(*logger)
	return ok
}
