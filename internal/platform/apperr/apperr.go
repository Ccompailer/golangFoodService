package apperr

import (
	"errors"
	"net/http"

	"golangFoodService/internal/platform/constants"
)

var (
	ErrInvalid      = errors.New("invalid argument")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrTimeout      = errors.New("timeout")
	ErrInternal     = errors.New("internal")
)

type AppError struct {
	Kind    error
	Title   string
	Detail  string
	Status  int
	wrapped error
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	if e.wrapped != nil {
		return e.wrapped.Error()
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return constants.ErrInternalServerErrorTitle
}

func (e *AppError) Unwrap() error {
	if e.wrapped != nil {
		return e.wrapped
	}
	return e.Kind
}

func Invalid(detail string) *AppError {
	return &AppError{Kind: ErrInvalid, Title: constants.ErrBadRequestTitle, Detail: detail, Status: http.StatusBadRequest}
}

func NotFound(detail string) *AppError {
	return &AppError{Kind: ErrNotFound, Title: constants.ErrNotFoundTitle, Detail: detail, Status: http.StatusNotFound}
}

func Conflict(detail string) *AppError {
	return &AppError{Kind: ErrConflict, Title: constants.ErrConflictTitle, Detail: detail, Status: http.StatusConflict}
}

func Unauthorized(detail string) *AppError {
	return &AppError{Kind: ErrUnauthorized, Title: constants.ErrUnauthorizedTitle, Detail: detail, Status: http.StatusUnauthorized}
}

func Forbidden(detail string) *AppError {
	return &AppError{Kind: ErrForbidden, Title: constants.ErrForbiddenTitle, Detail: detail, Status: http.StatusForbidden}
}

func Timeout(detail string) *AppError {
	return &AppError{Kind: ErrTimeout, Title: constants.ErrRequestTimeoutTitle, Detail: detail, Status: http.StatusGatewayTimeout}
}

func Internal(err error) *AppError {
	return &AppError{Kind: ErrInternal, Title: constants.ErrInternalServerErrorTitle, Status: http.StatusInternalServerError, wrapped: err}
}

func HTTPStatus(err error) int {
	var app *AppError
	if errors.As(err, &app) && app.Status != 0 {
		return app.Status
	}
	switch {
	case errors.Is(err, ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrTimeout):
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func TitleOf(err error) string {
	var app *AppError
	if errors.As(err, &app) && app.Title != "" {
		return app.Title
	}
	switch HTTPStatus(err) {
	case http.StatusBadRequest:
		return constants.ErrBadRequestTitle
	case http.StatusNotFound:
		return constants.ErrNotFoundTitle
	case http.StatusConflict:
		return constants.ErrConflictTitle
	case http.StatusUnauthorized:
		return constants.ErrUnauthorizedTitle
	case http.StatusForbidden:
		return constants.ErrForbiddenTitle
	case http.StatusGatewayTimeout:
		return constants.ErrRequestTimeoutTitle
	default:
		return constants.ErrInternalServerErrorTitle
	}
}
