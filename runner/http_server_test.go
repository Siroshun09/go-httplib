package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServerRunner_nil_server(t *testing.T) {
	assert.PanicsWithValue(t, "server is nil", func() {
		NewHTTPServerRunner(nil, nil, nil)
	})
}
