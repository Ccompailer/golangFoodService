package httpx

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"golangFoodService/internal/platform/apperr"
	"golangFoodService/internal/platform/constants"
)

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func HTTPErrorHandler(log *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status := http.StatusInternalServerError
		title := constants.ErrInternalServerErrorTitle
		detail := ""

		var he *echo.HTTPError
		var app *apperr.AppError
		switch {
		case errors.As(err, &app):
			status = apperr.HTTPStatus(err)
			title = apperr.TitleOf(err)
			detail = app.Detail
		case errors.As(err, &he):
			status = he.Code
			title = http.StatusText(he.Code)
			if msg, ok := he.Message.(string); ok {
				detail = msg
			}
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
			status = http.StatusGatewayTimeout
			title = constants.ErrRequestTimeoutTitle
		default:
			log.Error("unhandled error", zap.Error(err), zap.String("path", c.Path()))
		}

		if status >= 500 && detail != "" && status == http.StatusInternalServerError {
			detail = ""
		}

		_ = c.JSON(status, Problem{
			Type:     "about:blank",
			Title:    title,
			Status:   status,
			Detail:   detail,
			Instance: c.Path(),
		})
	}
}
