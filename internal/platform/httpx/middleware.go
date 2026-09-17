package httpx

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"golangFoodService/internal/platform/constants"
	"golangFoodService/internal/platform/id"
)

func StandardMiddleware(e *echo.Echo, log *zap.Logger) {
	e.Use(middleware.Recover())
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: id.New,
		TargetHeader: constants.HeaderRequestID,
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(constants.HeaderRequestID)
			reqLog := log.With(
				zap.String("request_id", reqID),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Path()),
			)
			c.Set("logger", reqLog)
			return next(c)
		}
	})
	e.Use(middleware.BodyLimit(constants.BodyLimit))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: constants.GzipLevel}))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{Timeout: constants.ReadTimeout}))
}

func LoggerFrom(c echo.Context) *zap.Logger {
	if v, ok := c.Get("logger").(*zap.Logger); ok && v != nil {
		return v
	}
	return zap.NewNop()
}
