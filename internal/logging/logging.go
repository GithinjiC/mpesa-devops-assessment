package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func New(level slog.Level, logFilePath string) (*slog.Logger, func(), error) {
	var writer io.Writer = os.Stdout
	cleanup := func() {}

	if logFilePath != "" {
		dir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, nil, fmt.Errorf("create log directory: %w", err)
		}
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		writer = io.MultiWriter(os.Stdout, f)
		cleanup = func() {
			_ = f.Sync()
			_ = f.Close()
		}
	}

	h := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	return slog.New(h), cleanup, nil
}
