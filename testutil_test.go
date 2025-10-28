package httplib_test

import (
	"io"
	"net/http"

	"github.com/stretchr/testify/assert"
)

type errorReader struct {
	err error
}

func (r errorReader) Read(_ []byte) (n int, err error) {
	return 0, r.err
}

func (r errorReader) Close() error {
	return nil
}

type closeErrorReader struct {
	r   io.Reader
	err error
}

func (r closeErrorReader) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r closeErrorReader) Close() error {
	return r.err
}

func assertMaxBytesErrorFunc(expectedLimit int64) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		var maxBytesErr *http.MaxBytesError
		if !assert.ErrorAs(t, err, &maxBytesErr) {
			return false
		}
		return assert.Equal(t, expectedLimit, maxBytesErr.Limit)
	}
}

func toPtr[T any](value T) *T {
	return &value
}
