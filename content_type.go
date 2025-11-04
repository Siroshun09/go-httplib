package httplib

// ContentType represents an HTTP Content-Type header value.
//
// These constants provide commonly used media types for convenience when
// setting HTTP headers (e.g. "Content-Type") or comparing values.
type ContentType = string

const (
	// ContentTypeTextPlain is a content type "text/plain"
	ContentTypeTextPlain ContentType = "text/plain"

	// ContentTypeJSON is a content type "application/json"
	ContentTypeJSON ContentType = "application/json"
	// ContentTypeJSONUTF8 is a content type "application/json; charset=utf-8"
	ContentTypeJSONUTF8 ContentType = "application/json; charset=utf-8"

	// ContentTypeOctetStream is a content type "application/octet-stream"
	ContentTypeOctetStream ContentType = "application/octet-stream"
)
