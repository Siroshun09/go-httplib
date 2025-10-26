package httplib

import (
	"encoding/json"
	"net/http"
)

const DefaultMaxRequestBodySize = 1 << 20 // 1MB

// DecodeJSONRequestBody decodes request body to T using JSON decoder.
//
// This function reads the request body up to DefaultMaxRequestBodySize.
// If the request body exceeds this size, the function returns http.MaxBytesError.
//
// The request body will be closed after decoding.
// This function ignores any error returned by Close.
func DecodeJSONRequestBody[T any](r *http.Request) (T, error) {
	body := http.MaxBytesReader(nil, r.Body, DefaultMaxRequestBodySize)
	defer body.Close()

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	var t T
	if err := decoder.Decode(&t); err != nil {
		var zero T
		return zero, err
	}

	return t, nil
}
