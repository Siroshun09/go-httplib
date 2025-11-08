package httplib

import (
	"context"
	"errors"
	"net/http"
)

// RenderOK renders a response with status code http.StatusOK without body.
func RenderOK(ctx context.Context, w http.ResponseWriter) {
	renderStatusCode(ctx, w, http.StatusOK, nil)
}

// RenderOKWithBody renders a response with status code http.StatusOK and body.
//
// Both RenderHeader and RenderBody will always be called, even if RenderHeader returns an error.
// The errors will be joined by errors.Join.
//
// When the bodyRenderer returns errors, this function will:
//   - Keep the status code as http.StatusOK
//   - Set ResponseLog in the context (the renderer error is not stored in ResponseLog.Error)
//   - Return the renderer error
func RenderOKWithBody(ctx context.Context, w http.ResponseWriter, bodyRenderer ResponseBodyRenderer) error {
	return renderWithBody(ctx, w, http.StatusOK, bodyRenderer, nil)
}

// RenderCreated renders a response with status code http.StatusCreated without body.
func RenderCreated(ctx context.Context, w http.ResponseWriter) {
	renderStatusCode(ctx, w, http.StatusCreated, nil)
}

// RenderCreatedWithBody renders a response with status code http.StatusCreated and body.
//
// Both RenderHeader and RenderBody will always be called, even if RenderHeader returns an error.
// The errors will be joined by errors.Join.
//
// When the bodyRenderer returns errors, this function will:
//   - Keep the status code as http.StatusCreated
//   - Set ResponseLog in the context (the renderer error is not stored in ResponseLog.Error)
//   - Return the renderer error
func RenderCreatedWithBody(ctx context.Context, w http.ResponseWriter, bodyRenderer ResponseBodyRenderer) error {
	return renderWithBody(ctx, w, http.StatusCreated, bodyRenderer, nil)
}

// RenderNoContent renders a response with status code http.StatusNoContent without body.
func RenderNoContent(ctx context.Context, w http.ResponseWriter) {
	renderStatusCode(ctx, w, http.StatusNoContent, nil)
}

// RenderNoContentForUnauthorized renders a response with status code http.StatusNoContent without body.
//
// This function can be used to send the same success response to the client when an authentication error occurs.
//
// The cause error will be used for ResponseLog.Error.
func RenderNoContentForUnauthorized(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusNoContent, cause)
}

// RenderRedirect renders a response with status code http.StatusTemporaryRedirect and a redirect url.
func RenderRedirect(ctx context.Context, w http.ResponseWriter, r *http.Request, url string) {
	resPtr := GetResponseLogPtrFromContext(ctx)
	if resPtr != nil {
		*resPtr = ResponseLog{
			StatusCode:  http.StatusTemporaryRedirect,
			HandlerInfo: NewHandlerInfo(1), // RenderRedirect -> caller
		}
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// RenderBadRequest renders a response with status code http.StatusBadRequest without body.
//
// The cause error will be used for ResponseLog.Error.
func RenderBadRequest(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusBadRequest, cause)
}

// RenderBadRequestWithBody renders a response with status code http.StatusBadRequest and body.
//
// Both RenderHeader and RenderBody will always be called, even if RenderHeader returns an error.
// The errors will be joined by errors.Join.
//
// When the bodyRenderer returns errors, this function will:
//   - Keep the status code as http.StatusBadRequest
//   - Set ResponseLog in the context (the renderer error is not stored in ResponseLog.Error)
//   - Return the renderer error
func RenderBadRequestWithBody(ctx context.Context, w http.ResponseWriter, bodyRenderer ResponseBodyRenderer, cause error) error {
	return renderWithBody(ctx, w, http.StatusBadRequest, bodyRenderer, cause)
}

// RenderUnauthorized renders a response with status code http.StatusUnauthorized without body.
//
// The cause error will be used for ResponseLog.Error.
func RenderUnauthorized(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusUnauthorized, cause)
}

// RenderForbidden renders a response with status code http.StatusForbidden without body.
//
// The cause error will be used for ResponseLog.Error.
func RenderForbidden(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusForbidden, cause)
}

// RenderNotFound renders a response with status code http.StatusNotFound without body.
//
// The cause error will be used for ResponseLog.Error.
func RenderNotFound(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusNotFound, cause)
}

// RenderConflict renders a response with status code http.StatusConflict without body.
//
// The cause error will be used for ResponseLog.Error.
func RenderConflict(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusConflict, cause)
}

// RenderInternalServerError renders a response with status code http.StatusInternalServerError without body.
//
// The cause error will be used for ResponseLog.Error.
func RenderInternalServerError(ctx context.Context, w http.ResponseWriter, cause error) {
	renderStatusCode(ctx, w, http.StatusInternalServerError, cause)
}

func renderStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, cause error) {
	_ = renderResponse(ctx, w, statusCode, nil, cause) // no error will be occurred
}

func renderWithBody(ctx context.Context, w http.ResponseWriter, statusCode int, bodyRenderer ResponseBodyRenderer, cause error) error {
	return renderResponse(ctx, w, statusCode, bodyRenderer, cause)
}

func renderResponse(ctx context.Context, w http.ResponseWriter, statusCode int, bodyRenderer ResponseBodyRenderer, cause error) error {
	var err error

	if bodyRenderer != nil {
		if headerErr := bodyRenderer.RenderHeader(ctx, w.Header()); headerErr != nil {
			err = headerErr
		}
	}

	w.WriteHeader(statusCode)
	size := int64(0)

	if bodyRenderer != nil {
		wrapped := responseBodyWriter{w: w}
		bodyErr := bodyRenderer.RenderBody(ctx, &wrapped)
		if bodyErr != nil {
			err = errors.Join(err, bodyErr)
		}
		size = wrapped.responseSize
	}

	resPtr := GetResponseLogPtrFromContext(ctx)
	if resPtr != nil {
		*resPtr = ResponseLog{
			StatusCode:   statusCode,
			ResponseSize: size,
			Error:        cause,
			// skip=3: renderResponse(0) -> renderStatusCode/renderWithBody(1) -> RenderXX(2) -> caller(3)
			HandlerInfo: NewHandlerInfo(3),
		}
	}

	return err
}
