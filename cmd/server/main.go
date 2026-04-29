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

	"github.com/example/job-crud-app/internal/config"
	"github.com/example/job-crud-app/internal/db"
	"github.com/example/job-crud-app/internal/handlers"
	"github.com/example/job-crud-app/internal/logging"
	"github.com/example/job-crud-app/internal/migrate"
	"github.com/joho/godotenv"
)

func main() {
	os.Exit(run())
}

func run() int {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration", slog.Any("err", err))
		return 1
	}

	logger, cleanupLog, err := logging.New(cfg.LogLevel, cfg.LogFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logging init: %v\n", err)
		return 1
	}
	defer cleanupLog()

	slog.SetDefault(logger)

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database", slog.Any("err", err))
		return 1
	}
	defer pool.Close()

	if err := migrate.RunUp(ctx, pool, cfg.MigrationsDir); err != nil {
		slog.Error("migrate", slog.Any("err", err))
		return 1
	}

	jh := &handlers.Jobs{DB: pool, Log: logger}

	addr := cfg.ListenAddress()
	handler := newRouter(cfg, logger, jh)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening",
			slog.String("addr", addr),
			slog.String("service", cfg.ServiceName),
			slog.String("log_file", cfg.LogFilePath),
		)
		errCh <- srv.ListenAndServe()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server", slog.Any("err", err))
			return 1
		}
		slog.Info("server stopped")
		return 0
	case <-sig:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("shutdown", slog.Any("err", err))
	}
	slog.Info("server stopped")
	return 0
}
