package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "max-miniapp/backend/docs"

	"max-miniapp/backend/internal/config"
	"max-miniapp/backend/internal/database"
	"max-miniapp/backend/internal/server"
	"max-miniapp/backend/internal/worker"
)

// @title           MAX Mini App — Reverse Hiring API
// @version         1.0
// @description     API мини-приложения реверс-найма в MAX.
// @description     Авторизация: POST /api/v1/auth c initData из MAX Bridge,
// @description     далее заголовок Authorization: Bearer <token> (JWT, 7 дней).
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Значение: "Bearer <jwt>"
func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg := config.Load()
	log := newLogger(cfg.Env)

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()
	log.Info("postgres connected")

	if err := database.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	log.Info("migrations applied")

	srv := server.New(cfg, log, pool)

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	worker.NewSurveyWorker(pool, cfg.MaxBotToken, log).Start(workerCtx)

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server started", slog.String("addr", cfg.Addr))
		errCh <- srv.Start()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
	case <-stop:
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
	}

	log.Info("server stopped")
	return nil
}

func newLogger(env string) *slog.Logger {
	if env == "prod" {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
