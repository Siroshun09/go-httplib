package httplib_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Siroshun09/go-httplib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RenderSuccess(t *testing.T) {
	tests := []struct {
		name           string
		f              func(ctx context.Context, w http.ResponseWriter)
		wantStatusCode int
	}{
		{
			name:           "OK",
			f:              httplib.RenderOK,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Created",
			f:              httplib.RenderCreated,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "NoContent",
			f:              httplib.RenderNoContent,
			wantStatusCode: http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
			w := httptest.NewRecorder()

			tt.f(ctx, w)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			assert.Empty(t, w.Body.Bytes())
			assertResponseLog(ctx, t, tt.wantStatusCode, 0, nil)
		})
	}
}

func Test_RenderSuccessWithBody(t *testing.T) {
	tests := []struct {
		name           string
		f              func(ctx context.Context, w http.ResponseWriter, bodyRenderer httplib.ResponseBodyRenderer) error
		wantStatusCode int
	}{
		{
			name:           "OK",
			f:              httplib.RenderOKWithBody,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Created",
			f:              httplib.RenderCreatedWithBody,
			wantStatusCode: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
			w := httptest.NewRecorder()
			data := []byte("test")

			require.NoError(t, tt.f(ctx, w, httplib.RawResponse(data)))

			assert.Equal(t, tt.wantStatusCode, w.Code)
			assert.Equal(t, data, w.Body.Bytes())
			assertResponseLog(ctx, t, tt.wantStatusCode, int64(len(data)), nil)
		})
	}
}

func Test_RenderSuccessWithBody_Error(t *testing.T) {
	tests := []struct {
		name           string
		f              func(ctx context.Context, w http.ResponseWriter, bodyRenderer httplib.ResponseBodyRenderer) error
		headerErr      error
		bodyErr        error
		wantErr        error
		wantStatusCode int
	}{
		{
			name:           "OK - header error",
			f:              httplib.RenderOKWithBody,
			headerErr:      errors.New("header error"),
			bodyErr:        nil,
			wantErr:        errors.New("header error"),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "OK - body error",
			f:              httplib.RenderOKWithBody,
			headerErr:      nil,
			bodyErr:        errors.New("body error"),
			wantErr:        errors.New("body error"),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "OK - both error",
			f:              httplib.RenderOKWithBody,
			headerErr:      errors.New("header error"),
			bodyErr:        errors.New("body error"),
			wantErr:        errors.Join(errors.New("header error"), errors.New("body error")),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Created - header error",
			f:              httplib.RenderCreatedWithBody,
			headerErr:      errors.New("header error"),
			bodyErr:        nil,
			wantErr:        errors.New("header error"),
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Created - body error",
			f:              httplib.RenderCreatedWithBody,
			headerErr:      nil,
			bodyErr:        errors.New("body error"),
			wantErr:        errors.New("body error"),
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Created - both error",
			f:              httplib.RenderCreatedWithBody,
			headerErr:      errors.New("header error"),
			bodyErr:        errors.New("body error"),
			wantErr:        errors.Join(errors.New("header error"), errors.New("body error")),
			wantStatusCode: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})

			got := tt.f(ctx, httptest.NewRecorder(), errorResponseBodyRenderer{headerErr: tt.headerErr, bodyErr: tt.bodyErr})
			assert.EqualError(t, got, tt.wantErr.Error())

			assertResponseLog(ctx, t, tt.wantStatusCode, 0, nil)
		})
	}
}

func Test_RenderRedirect(t *testing.T) {
	ctx := t.Context()

	ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://example.com/redirect", nil)

	httplib.RenderRedirect(ctx, w, r, "https://example.com/redirected")

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, `<a href="https://example.com/redirected">Temporary Redirect</a>.`, strings.TrimRight(w.Body.String(), "\n")) // ignore newlines
	assert.Equal(t, "https://example.com/redirected", w.Header().Get("Location"))
	assertResponseLogWithFuncName(ctx, t, http.StatusTemporaryRedirect, 0, nil, "github.com/Siroshun09/go-httplib_test.Test_RenderRedirect")
}

func Test_RenderRedirect_InvalidURL(t *testing.T) {
	ctx := t.Context()
	ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://example.com/redirect", nil)

	httplib.RenderRedirect(ctx, w, r, "https://example.com/invalid url with spaces")

	// Redirect should still work, but the URL will be escaped by http.Redirect
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "invalid")
	assertResponseLogWithFuncName(ctx, t, http.StatusTemporaryRedirect, 0, nil, "github.com/Siroshun09/go-httplib_test.Test_RenderRedirect_InvalidURL")
}

func Test_RenderError(t *testing.T) {
	tests := []struct {
		name           string
		cause          error
		f              func(ctx context.Context, w http.ResponseWriter, cause error)
		wantStatusCode int
	}{
		{
			name:           "BadRequest",
			cause:          errors.New("bad request"),
			f:              httplib.RenderBadRequest,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			cause:          errors.New("unauthorized"),
			f:              httplib.RenderUnauthorized,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden",
			cause:          errors.New("forbidden"),
			f:              httplib.RenderForbidden,
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "NotFound",
			cause:          errors.New("not found"),
			f:              httplib.RenderNotFound,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "Conflict",
			cause:          errors.New("conflict"),
			f:              httplib.RenderConflict,
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "InternalServerError",
			cause:          errors.New("internal server error"),
			f:              httplib.RenderInternalServerError,
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
			w := httptest.NewRecorder()

			tt.f(ctx, w, tt.cause)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			assert.Empty(t, w.Body.Bytes())
			assertResponseLog(ctx, t, tt.wantStatusCode, 0, tt.cause)
		})
	}
}

func Test_RenderErrorWithBody(t *testing.T) {
	tests := []struct {
		name           string
		cause          error
		f              func(ctx context.Context, w http.ResponseWriter, bodyRenderer httplib.ResponseBodyRenderer, cause error) error
		wantStatusCode int
	}{
		{
			name:           "BadRequest",
			cause:          errors.New("bad request"),
			f:              httplib.RenderBadRequestWithBody,
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
			w := httptest.NewRecorder()
			data := []byte("test")

			require.NoError(t, tt.f(ctx, w, httplib.RawResponse(data), tt.cause))

			assert.Equal(t, tt.wantStatusCode, w.Code)
			assert.Equal(t, data, w.Body.Bytes())
			assertResponseLog(ctx, t, tt.wantStatusCode, int64(len(data)), tt.cause)
		})
	}
}

func Test_RenderErrorWithBody_Error(t *testing.T) {
	tests := []struct {
		name           string
		f              func(ctx context.Context, w http.ResponseWriter, bodyRenderer httplib.ResponseBodyRenderer, cause error) error
		headerErr      error
		bodyErr        error
		wantErr        error
		wantStatusCode int
	}{
		{
			name:           "BadRequest - header error",
			f:              httplib.RenderBadRequestWithBody,
			headerErr:      errors.New("header error"),
			bodyErr:        nil,
			wantErr:        errors.New("header error"),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "BadRequest - body error",
			f:              httplib.RenderBadRequestWithBody,
			headerErr:      nil,
			bodyErr:        errors.New("body error"),
			wantErr:        errors.New("body error"),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "BadRequest - both error",
			f:              httplib.RenderBadRequestWithBody,
			headerErr:      errors.New("header error"),
			bodyErr:        errors.New("body error"),
			wantErr:        errors.Join(errors.New("header error"), errors.New("body error")),
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})

			got := tt.f(ctx, httptest.NewRecorder(), errorResponseBodyRenderer{headerErr: tt.headerErr, bodyErr: tt.bodyErr}, nil)
			assert.EqualError(t, got, tt.wantErr.Error())

			assertResponseLog(ctx, t, tt.wantStatusCode, 0, nil)
		})
	}
}

func Test_RenderWithBody_ErrorFromResponseWriter(t *testing.T) {
	tests := []struct {
		name           string
		f              func(ctx context.Context, w http.ResponseWriter) error
		wantStatusCode int
	}{
		{
			name: "RenderOK with body",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				return httplib.RenderOKWithBody(ctx, w, httplib.RawResponse([]byte("test")))
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "RenderCreated with body",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				return httplib.RenderCreatedWithBody(ctx, w, httplib.RawResponse([]byte("test")))
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "RenderBadRequest with body",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				return httplib.RenderBadRequestWithBody(ctx, w, httplib.RawResponse([]byte("test")), nil)
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
			err := errors.New("response writer error")
			w := &errorResponseWriter{ResponseWriter: httptest.NewRecorder(), err: err}

			got := tt.f(ctx, w)
			assert.EqualError(t, got, err.Error())

			assertResponseLog(ctx, t, tt.wantStatusCode, 0, nil)
		})
	}
}

/* Tests for created HandlerInfo */

func Test_RenderOK_HandlerInfo_FuncName(t *testing.T) {
	ctx := t.Context()

	ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
	w := httptest.NewRecorder()
	httplib.RenderOK(ctx, w)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.Bytes())
	assertResponseLogWithFuncName(ctx, t, http.StatusOK, 0, nil, "github.com/Siroshun09/go-httplib_test.Test_RenderOK_HandlerInfo_FuncName")
}

func Test_RenderInternalServerError_HandlerInfo_FuncName(t *testing.T) {
	ctx := t.Context()

	ctx = httplib.WithResponseLogPtr(ctx, &httplib.ResponseLog{})
	w := httptest.NewRecorder()
	err := errors.New("internal server error")
	httplib.RenderInternalServerError(ctx, w, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Body.Bytes())
	assertResponseLogWithFuncName(ctx, t, http.StatusInternalServerError, 0, err, "github.com/Siroshun09/go-httplib_test.Test_RenderInternalServerError_HandlerInfo_FuncName")
}

func assertResponseLog(ctx context.Context, t *testing.T, expectedStatusCode int, expectedResponseSize int64, expectedError error) {
	t.Helper()
	assertResponseLogWithFuncName(ctx, t, expectedStatusCode, expectedResponseSize, expectedError, "")
}

func assertResponseLogWithFuncName(ctx context.Context, t *testing.T, expectedStatusCode int, expectedResponseSize int64, expectedError error, expectedFuncName string) {
	t.Helper()

	res := httplib.GetResponseLogPtrFromContext(ctx)
	require.NotNil(t, res)

	assert.Equal(t, expectedStatusCode, res.StatusCode)
	assert.Equal(t, expectedResponseSize, res.ResponseSize)
	assert.Equal(t, expectedError, res.Error)

	// Only check that HandlerInfo is not empty or UnknownHandlerInfo
	assert.NotEmpty(t, res.HandlerInfo)
	assert.NotEqual(t, httplib.UnknownHandlerInfo(), res.HandlerInfo)

	if expectedFuncName != "" {
		assert.Equal(t, expectedFuncName, res.HandlerInfo.FuncName)
		assert.NotEmpty(t, res.HandlerInfo.File)
		assert.NotZero(t, res.HandlerInfo.Line)
	}
}

func Test_Render_WithoutResponseLogHolder(t *testing.T) {
	tests := []struct {
		name string
		f    func(ctx context.Context, w http.ResponseWriter) error
	}{
		{
			name: "RenderOK",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				httplib.RenderOK(ctx, w)
				return nil
			},
		},
		{
			name: "RenderOKWithBody",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				return httplib.RenderOKWithBody(ctx, w, httplib.RawResponse([]byte("test")))
			},
		},
		{
			name: "RenderRedirect",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				r := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
				httplib.RenderRedirect(ctx, w, r, "https://example.com")
				return nil
			},
		},
		{
			name: "RenderBadRequest",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				httplib.RenderBadRequest(ctx, w, nil)
				return nil
			},
		},
		{
			name: "RenderBadRequestWithBody",
			f: func(ctx context.Context, w http.ResponseWriter) error {
				return httplib.RenderBadRequestWithBody(ctx, w, httplib.RawResponse([]byte("test")), nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			w := httptest.NewRecorder()
			require.NoError(t, tt.f(ctx, w))

			// if ctx doesn't have a ResponseLogHolder, it should be nil even when calling RenderXXX
			assert.Nil(t, httplib.GetResponseLogPtrFromContext(ctx))
		})
	}
}
