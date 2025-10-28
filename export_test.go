package httplib

const ContextKeyRequestLog = contextKeyRequestLog

func NewHandlerInfoFromPC(pc uintptr, file string, line int) HandlerInfo {
	return newHandlerInfoFromPC(pc, file, line)
}
