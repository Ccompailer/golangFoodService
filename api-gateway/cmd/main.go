package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/food-service/api-gateway/internal/config"
	"github.com/food-service/api-gateway/internal/middlewares/auth"
	"github.com/food-service/api-gateway/internal/middlewares/auth/interfaces"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
		}

		os.Exit(1)
	}

	var logLevel slog.Level
	switch cfg.Log.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var authInterface interfaces.IAuthenticator

	authInstance, err := auth.NewAuthenticator(ctx, cfg.Auth)
	if err != nil {
		logger.Error("failed to create authenticator: %v", err)
	}

	authInterface = authInstance
	_ = authInterface

	srv := &http.Server{
		Addr: ":" + cfg.Server.Port,
		//Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		logger.Info("Starting API Gateway",
			"port", cfg.Server.Port,
			"environment", cfg.Server.Environment,
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
