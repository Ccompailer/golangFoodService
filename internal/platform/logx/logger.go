package logx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"golangFoodService/internal/platform/config/environment"
)

func New(env environment.Environment) (*zap.Logger, error) {
	if environment.IsProduction(env) {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return cfg.Build()
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build()
}

var Module = fx.Module("logx",
	fx.Provide(New),
	fx.Invoke(func(log *zap.Logger, lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStop: func(_ context.Context) error {
				_ = log.Sync()
				return nil
			},
		})
	}),
)
