package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/kleffio/kleff-auth/internal/core/service/auth"
)

type HTTPError struct {
	Err        error
	Message    string
	StatusCode int
	Details    map[string]any
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func BadRequest(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func BadRequestf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusBadRequest,
	}
}

func BadRequestWithDetails(message string, details map[string]any) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    details,
	}
}

func Unauthorized(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func Unauthorizedf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusUnauthorized,
	}
}

func Forbidden(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func Forbiddenf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusForbidden,
	}
}

func NotFound(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

func NotFoundf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusNotFound,
	}
}

func Conflict(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func Conflictf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusConflict,
	}
}

func UnprocessableEntity(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

func UnprocessableEntityf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusUnprocessableEntity,
	}
}

func InternalServerError(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

func InternalServerErrorf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusInternalServerError,
	}
}

func ServiceUnavailable(message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
	}
}

func ServiceUnavailablef(format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: http.StatusServiceUnavailable,
	}
}

func Custom(statusCode int, message string) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: statusCode,
	}
}

func Customf(statusCode int, format string, args ...any) *HTTPError {
	return &HTTPError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: statusCode,
	}
}

func Wrap(err error, statusCode int, message string) *HTTPError {
	return &HTTPError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

type HandlerWithError func(w http.ResponseWriter, r *http.Request) error

func httpStatusAndMsg(err error) (int, string) {
	switch {
	case errors.Is(err, auth.ErrUnknownTenant):
		return http.StatusBadRequest, "unknown tenant"
	case errors.Is(err, auth.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, auth.ErrInvalidRefresh):
		return http.StatusUnauthorized, "invalid refresh token"
	case errors.Is(err, auth.ErrReuseDetected):
		return http.StatusForbidden, "refresh token reuse detected"
	case errors.Is(err, auth.ErrInvalidClient):
		return http.StatusBadRequest, "invalid oauth client"
	case errors.Is(err, auth.ErrInvalidRedirectURI):
		return http.StatusBadRequest, "invalid redirect uri"
	case errors.Is(err, auth.ErrUnsupportedProvider):
		return http.StatusBadRequest, "unsupported oauth provider"
	case errors.Is(err, auth.ErrInvalidState):
		return http.StatusBadRequest, "invalid oauth state"

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func ErrorMiddleware(h HandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err == nil {
			return
		} else {
			log.Printf("handler error %s %s: %+v", r.Method, r.URL.String(), err)

			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				respondWithError(w, httpErr.StatusCode, httpErr.Error(), httpErr.Details)
				return
			}

			statusCode, message := httpStatusAndMsg(err)
			respondWithError(w, statusCode, message, nil)
		}
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]any{
		"error": message,
	}

	if len(details) > 0 {
		response["details"] = details
	}

	_ = json.NewEncoder(w).Encode(response)
}
