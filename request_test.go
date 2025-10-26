package httplib_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
)

func TestDecodeJSONRequestBody_String(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		want         string
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: valid json",
			data:         `"a"`,
			want:         "a",
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: invalid json",
			data:         `a`,
			want:         "",
			errAssertion: assert.Error,
		},
		{
			name:         "success: empty string",
			data:         `""`,
			want:         "",
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: empty data",
			data:         ``,
			want:         "",
			errAssertion: assert.Error,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         "",
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: number json",
			data:         `1`,
			want:         "",
			errAssertion: assert.Error,
		},
		{
			name:         "failure: boolean json",
			data:         `true`,
			want:         "",
			errAssertion: assert.Error,
		},
		{
			name:         "failure: array json",
			data:         `["a","b","c"]`,
			want:         "",
			errAssertion: assert.Error,
		},
		{
			name:         "failure: object json",
			data:         `{"a": "a"}`,
			want:         "",
			errAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[string](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_Int(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		want         int
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: valid json",
			data:         `1`,
			want:         1,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: invalid json",
			data:         `a`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: empty data",
			data:         ``,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         0,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: string json",
			data:         `"1"`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: boolean json",
			data:         `true`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: array json",
			data:         `[1,2,3]`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: object json",
			data:         `{"a":1}`,
			want:         0,
			errAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[int](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_Float64(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		want         float64
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: valid float json",
			data:         `1.5`,
			want:         1.5,
			errAssertion: assert.NoError,
		},
		{
			name:         "success: valid int as float",
			data:         `2`,
			want:         2.0,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: invalid json",
			data:         `a`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: empty data",
			data:         ``,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         0,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: string json",
			data:         `"1.5"`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: boolean json",
			data:         `false`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: array json",
			data:         `[1.1,2.2]`,
			want:         0,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: object json",
			data:         `{"a":1.5}`,
			want:         0,
			errAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[float64](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_Bool(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		want         bool
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: valid json",
			data:         `true`,
			want:         true,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: invalid json",
			data:         `tru`,
			want:         false,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: empty data",
			data:         ``,
			want:         false,
			errAssertion: assert.Error,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         false,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: string json",
			data:         `"true"`,
			want:         false,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: number json",
			data:         `1`,
			want:         false,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: array json",
			data:         `[true,false]`,
			want:         false,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: object json",
			data:         `{"a":true}`,
			want:         false,
			errAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[bool](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_Array(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		want         []string
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: valid json",
			data:         `["a","b","c"]`,
			want:         []string{"a", "b", "c"},
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: invalid json",
			data:         `["a","b",]`,
			want:         nil,
			errAssertion: assert.Error,
		},
		{
			name:         "success: empty array",
			data:         `[]`,
			want:         []string{},
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: empty data",
			data:         ``,
			want:         nil,
			errAssertion: assert.Error,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         nil,
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: string json",
			data:         `"a"`,
			want:         nil,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: number json",
			data:         `1`,
			want:         nil,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: boolean json",
			data:         `true`,
			want:         nil,
			errAssertion: assert.Error,
		},
		{
			name:         "failure: object json",
			data:         `{}`,
			want:         nil,
			errAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[[]string](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_Object(t *testing.T) {
	type testObject struct {
		A    string            `json:"a"`
		B    int               `json:"b"`
		C    bool              `json:"c"`
		List []string          `json:"list"`
		Map  map[string]string `json:"map"`
	}

	tests := []struct {
		name         string
		data         string
		want         testObject
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name: "success: valid json",
			data: `{"a":"a","b":1,"c":true,"list":["a","b","c"],"map":{"a":"b","c":"d"}}`,
			want: testObject{
				A:    "a",
				B:    1,
				C:    true,
				List: []string{"a", "b", "c"},
				Map:  map[string]string{"a": "b", "c": "d"},
			},
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: invalid json",
			data:         `{"a":"a"`,
			want:         testObject{},
			errAssertion: assert.Error,
		},
		{
			name:         "success: empty map",
			data:         `{}`,
			want:         testObject{},
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: empty data",
			data:         ``,
			want:         testObject{},
			errAssertion: assert.Error,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         testObject{},
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: string json",
			data:         `"a"`,
			want:         testObject{},
			errAssertion: assert.Error,
		},
		{
			name:         "failure: number json",
			data:         `1`,
			want:         testObject{},
			errAssertion: assert.Error,
		},
		{
			name:         "failure: boolean json",
			data:         `true`,
			want:         testObject{},
			errAssertion: assert.Error,
		},
		{
			name:         "failure: array json",
			data:         `["a","b","c"]`,
			want:         testObject{},
			errAssertion: assert.Error,
		},
		{
			name:         "failure: has unknown field",
			data:         `{"a":"a","b":1,"c":true,"list":["a","b","c"],"map":{"a":"b","c":"d"},"unknown":"unknown"}`,
			want:         testObject{},
			errAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[testObject](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_Any(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		want         any
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: string",
			data:         `"a"`,
			want:         "a",
			errAssertion: assert.NoError,
		},
		{
			name:         "success: int",
			data:         `1`,
			want:         1.0,
			errAssertion: assert.NoError,
		},
		{
			name:         "success: float64",
			data:         `1.5`,
			want:         1.5,
			errAssertion: assert.NoError,
		},
		{
			name:         "success: boolean",
			data:         `true`,
			want:         true,
			errAssertion: assert.NoError,
		},
		{
			name:         "success: array",
			data:         `[1,2,3]`,
			want:         []any{1.0, 2.0, 3.0},
			errAssertion: assert.NoError,
		},
		{
			name:         "success: object",
			data:         `{"a":1, "b":"c"}`,
			want:         map[string]any{"a": 1.0, "b": "c"},
			errAssertion: assert.NoError,
		},
		{
			name:         "success: null",
			data:         `null`,
			want:         nil,
			errAssertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.data)))
			got, err := httplib.DecodeJSONRequestBody[any](r)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeJSONRequestBody_CloseError(t *testing.T) {
	t.Run("only close error", func(t *testing.T) {
		reader := closeErrorReader{
			r:   bytes.NewReader([]byte(`"test"`)),
			err: errors.New("error"),
		}
		r := httptest.NewRequest(http.MethodPost, "/", reader)
		got, err := httplib.DecodeJSONRequestBody[string](r)
		assert.Equal(t, "test", got)
		assert.NoError(t, err)
	})

	t.Run("read error and close error", func(t *testing.T) {
		reader := closeErrorReader{
			r:   errorReader{err: errors.New("read error")},
			err: errors.New("close error"),
		}
		r := httptest.NewRequest(http.MethodPost, "/", reader)
		got, err := httplib.DecodeJSONRequestBody[string](r)
		assert.Empty(t, got)
		assert.EqualError(t, err, "read error")
	})
}

func TestDecodeJSONRequestBody_LargeRequest(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name:         "success: normal request",
			data:         []byte(`"test string"`),
			errAssertion: assert.NoError,
		},
		{
			name:         "success: equal to default max request size",
			data:         []byte(`"` + strings.Repeat("a", httplib.DefaultMaxRequestBodySize-2) + `"`),
			errAssertion: assert.NoError,
		},
		{
			name:         "failure: too large request (default + 1)",
			data:         []byte(`"` + strings.Repeat("a", httplib.DefaultMaxRequestBodySize+1-2) + `"`),
			errAssertion: assertMaxBytesErrorFunc(httplib.DefaultMaxRequestBodySize),
		},
		{
			name:         "failure: too large request (default * 2)",
			data:         []byte(`"` + strings.Repeat("a", httplib.DefaultMaxRequestBodySize*2-2) + `"`),
			errAssertion: assertMaxBytesErrorFunc(httplib.DefaultMaxRequestBodySize),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.data))
			_, err := httplib.DecodeJSONRequestBody[string](r)
			tt.errAssertion(t, err)
		})
	}
}
