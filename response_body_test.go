package httplib_test

import (
	"errors"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		wantData string
	}{
		{
			name:     "nil",
			data:     nil,
			wantData: `null`,
		},
		{
			name:     "string",
			data:     "a",
			wantData: `"a"`,
		},
		{
			name:     "int",
			data:     1,
			wantData: `1`,
		},
		{
			name:     "float64",
			data:     1.5,
			wantData: `1.5`,
		},
		{
			name:     "bool",
			data:     true,
			wantData: `true`,
		},
		{
			name:     "array",
			data:     []any{1, 2, 3},
			wantData: `[1,2,3]`,
		},
		{
			name:     "object",
			data:     map[string]any{"a": 1, "b": 2},
			wantData: `{"a":1,"b":2}`,
		},
		{
			name: "struct",
			data: struct {
				String string         `json:"string"`
				Int    int            `json:"int"`
				Float  float64        `json:"float"`
				Bool   bool           `json:"bool"`
				Array  []any          `json:"array"`
				Object map[string]any `json:"object"`
			}{
				String: "a",
				Int:    1,
				Float:  1.5,
				Bool:   true,
				Array:  []any{1, 2, 3},
				Object: map[string]any{"a": 1, "b": 2},
			},
			wantData: `{"string":"a","int":1,"float":1.5,"bool":true,"array":[1,2,3],"object":{"a":1,"b":2}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			renderer, err := httplib.JSONResponse(tt.data)
			require.NoError(t, err)

			w := httptest.NewRecorder()

			assert.NoError(t, renderer.RenderHeader(ctx, w.Header()))
			assert.NoError(t, renderer.RenderBody(ctx, w))

			assert.Equal(t, httplib.ContentTypeJSONUTF8, w.Header().Get("Content-Type"))
			assert.Equal(t, tt.wantData, w.Body.String())
		})
	}
}

func TestJSONResponse_Error(t *testing.T) {
	ctx := t.Context()

	renderer, err := httplib.JSONResponse("test")
	require.NoError(t, err)

	err = errors.New("response writer error")
	w := &errorResponseWriter{ResponseWriter: httptest.NewRecorder(), err: err}

	assert.NoError(t, renderer.RenderHeader(ctx, w.Header()))
	assert.EqualError(t, renderer.RenderBody(ctx, w), err.Error())

	assert.Equal(t, httplib.ContentTypeJSONUTF8, w.Header().Get("Content-Type"))
}

func TestJSONResponse_CircularReference(t *testing.T) {
	type node struct {
		Next *node `json:"next"`
	}

	n := &node{}
	n.Next = n // circular reference

	renderer, err := httplib.JSONResponse(n)
	assert.Error(t, err)
	assert.Nil(t, renderer)
}

func TestRawResponse(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "nil",
			data: nil,
		},
		{
			name: "string",
			data: []byte("a"),
		},
		{
			name: "binary",
			data: make([]byte, 1024),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			renderer := httplib.RawResponse(tt.data)
			w := httptest.NewRecorder()

			assert.NoError(t, renderer.RenderHeader(ctx, w.Header()))
			assert.NoError(t, renderer.RenderBody(ctx, w))

			assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
			assert.Equal(t, strconv.Itoa(len(tt.data)), w.Header().Get("Content-Length"))
			assert.Equal(t, tt.data, w.Body.Bytes())
		})
	}
}

func TestRawResponse_Error(t *testing.T) {
	ctx := t.Context()
	renderer := httplib.RawResponse([]byte("test"))
	err := errors.New("response writer error")
	w := &errorResponseWriter{ResponseWriter: httptest.NewRecorder(), err: err}

	assert.NoError(t, renderer.RenderHeader(ctx, w.Header()))
	assert.EqualError(t, renderer.RenderBody(ctx, w), err.Error())

	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "4", w.Header().Get("Content-Length"))
}

func TestRawResponseWithContentType(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		contentType httplib.ContentType
	}{
		{
			name:        "nil",
			data:        nil,
			contentType: httplib.ContentTypeOctetStream,
		},
		{
			name:        "string",
			data:        []byte("a"),
			contentType: httplib.ContentTypeTextPlain,
		},
		{
			name:        "binary",
			data:        make([]byte, 1024),
			contentType: httplib.ContentTypeOctetStream,
		},
		{
			name:        "custom",
			data:        []byte("a"),
			contentType: "custom/type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			renderer := httplib.RawResponseWithContentType(tt.data, tt.contentType)
			w := httptest.NewRecorder()

			assert.NoError(t, renderer.RenderHeader(ctx, w.Header()))
			assert.NoError(t, renderer.RenderBody(ctx, w))

			assert.Equal(t, tt.contentType, w.Header().Get("Content-Type"))
			assert.Equal(t, strconv.Itoa(len(tt.data)), w.Header().Get("Content-Length"))
			assert.Equal(t, tt.data, w.Body.Bytes())
		})
	}
}
