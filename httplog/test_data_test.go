package httplog_test

import (
	"errors"
	"net/http"
	"time"

	"github.com/Siroshun09/go-httplib"
)

var (
	testRequestLog = httplib.RequestLog{
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
	}

	testResponseLog = httplib.ResponseLog{
		StatusCode:   http.StatusInternalServerError,
		ResponseSize: 100,
		Error:        errors.New("internal server error"),
		HandlerInfo: httplib.HandlerInfo{
			FuncName: "func_status_internal_server_error",
			File:     "file_status_internal_server_error.go",
			Line:     2,
		},
	}
)
