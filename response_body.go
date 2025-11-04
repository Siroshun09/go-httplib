package httplib

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type ResponseBodyRenderer interface {
	RenderHeader(ctx context.Context, header http.Header) error
	RenderBody(ctx context.Context, w io.Writer) error
}

func JSONResponse(v any) (ResponseBodyRenderer, error) {
	data, err := json.Marshal(v) // TODO: use json.MarshalWrite after encoding/json/v2 is stabilized.
	if err != nil {
		return nil, err
	}

	return &rawResponseBodyRenderer{
		b:           data,
		contentType: ContentTypeJSONUTF8,
	}, nil
}

type rawResponseBodyRenderer struct {
	b           []byte
	contentType ContentType
}

func RawResponse(b []byte) ResponseBodyRenderer {
	return RawResponseWithContentType(b, ContentTypeOctetStream)
}

func RawResponseWithContentType(b []byte, contentType ContentType) ResponseBodyRenderer {
	return &rawResponseBodyRenderer{b: b, contentType: contentType}
}

func (r *rawResponseBodyRenderer) RenderHeader(_ context.Context, header http.Header) error {
	header.Set("Content-Type", r.contentType)
	header.Set("Content-Length", strconv.Itoa(len(r.b)))
	return nil
}

func (r *rawResponseBodyRenderer) RenderBody(_ context.Context, w io.Writer) error {
	_, err := w.Write(r.b)
	return err
}
