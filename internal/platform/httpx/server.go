package httpx

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"golangFoodService/internal/platform/constants"
	"golangFoodService/internal/platform/health"
)

type Config struct {
	Addr string
}

func NewEcho(log *zap.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = HTTPErrorHandler(log)
	StandardMiddleware(e, log)
	return e
}

func RegisterHealth(e *echo.Echo, live, ready http.HandlerFunc) {
	e.GET("/live", echo.WrapHandler(live))
	e.GET("/ready", echo.WrapHandler(ready))
}

func Run(lc fx.Lifecycle, e *echo.Echo, cfg Config, log *zap.Logger, ready *health.Registry) {
	if ready != nil {
		RegisterHealth(e, health.Live, ready.Ready)
	} else {
		RegisterHealth(e, health.Live, health.Live)
	}

	srv := &http.Server{
		Addr:           cfg.Addr,
		Handler:        e,
		ReadTimeout:    constants.ReadTimeout,
		WriteTimeout:   constants.WriteTimeout,
		MaxHeaderBytes: constants.MaxHeaderBytes,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info("http listening", zap.String("addr", cfg.Addr))
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal("http server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, constants.WaitShutdownDuration)
			defer cancel()
			log.Info("http shutting down")
			return srv.Shutdown(shutdownCtx)
		},
	})
}
