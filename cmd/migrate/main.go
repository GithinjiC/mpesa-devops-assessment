package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/example/job-crud-app/internal/config"
	"github.com/example/job-crud-app/internal/db"
	"github.com/example/job-crud-app/internal/logging"
	"github.com/example/job-crud-app/internal/migrate"
	"github.com/joho/godotenv"
)

func main() {
	os.Exit(run())
}

func run() int {
	_ = godotenv.Load()

	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

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

	switch *direction {
	case "up":
		if err := migrate.RunUp(ctx, pool, cfg.MigrationsDir); err != nil {
			slog.Error("migration up failed", slog.Any("err", err))
			return 1
		}
	case "down":
		if err := migrate.RunDown(ctx, pool, cfg.MigrationsDir); err != nil {
			slog.Error("migration down failed", slog.Any("err", err))
			return 1
		}
	default:
		slog.Error("invalid direction", slog.String("direction", *direction))
		return 1
	}

	return 0
}
