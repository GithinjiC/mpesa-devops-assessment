package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func New(level slog.Level, logFilePath string) (*slog.Logger, func(), error) {
	if logFilePath == "" {
		return nil, nil, fmt.Errorf("log file path is empty")
	}
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, nil, fmt.Errorf("create log directory: %w", err)
	}
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	mw := io.MultiWriter(os.Stdout, f)
	h := slog.NewJSONHandler(mw, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	cleanup := func() {
		_ = f.Sync()
		_ = f.Close()
	}
	return logger, cleanup, nil
}
