package httplib_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_responseBodyWriter_Write(t *testing.T) {
	t.Run("success: 10 bytes", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := httplib.NewResponseBodyWriter(w)
		data := make([]byte, 10)

		size, err := wrapped.Write(data)
		require.NoError(t, err)
		assert.EqualValues(t, len(data), size)
		assert.Equal(t, data, w.Body.Bytes())
		assert.EqualValues(t, len(data), wrapped.ResponseSize())
	})

	t.Run("failure: response writer error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		err := errors.New("response writer error")
		w := &errorResponseWriter{ResponseWriter: recorder, err: err}
		wrapped := httplib.NewResponseBodyWriter(w)
		data := make([]byte, 10)

		size, err := wrapped.Write(data)
		assert.Zero(t, size)
		assert.EqualError(t, err, err.Error())
		assert.Zero(t, wrapped.ResponseSize())
	})
}
