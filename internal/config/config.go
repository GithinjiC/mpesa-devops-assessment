package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL           string
	Port                  int
	MigrationsDir         string
	LogFilePath           string
	LogLevel              slog.Level
	ServiceName           string
	HTTPListenHost        string
	HTTPReadHeaderTimeout time.Duration
	HTTPShutdownTimeout   time.Duration
	HTTPRequestTimeout    time.Duration
}

func requiredEnv(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set or empty", key)
	}
	return v, nil
}

func parseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid LOG_LEVEL %q (use debug, info, warn, error)", s)
	}
}

func parseDurationEnv(key string) (time.Duration, error) {
	s, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration %q: %w", key, s, err)
	}
	return d, nil
}

func Load() (*Config, error) {
	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	portStr, err := requiredEnv("PORT")
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("PORT: %w", err)
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("PORT must be between 1 and 65535")
	}

	migrationsDir, err := requiredEnv("MIGRATIONS_DIR")
	if err != nil {
		return nil, err
	}

	logLevelStr, err := requiredEnv("LOG_LEVEL")
	if err != nil {
		return nil, err
	}
	logLevel, err := parseLogLevel(logLevelStr)
	if err != nil {
		return nil, err
	}

	logFilePath := strings.TrimSpace(os.Getenv("LOG_FILE_PATH"))

	serviceName, err := requiredEnv("SERVICE_NAME")
	if err != nil {
		return nil, err
	}

	httpListenHost, err := requiredEnv("HTTP_LISTEN_HOST")
	if err != nil {
		return nil, err
	}

	readHdr, err := parseDurationEnv("HTTP_READ_HEADER_TIMEOUT")
	if err != nil {
		return nil, err
	}
	shutdown, err := parseDurationEnv("HTTP_SHUTDOWN_TIMEOUT")
	if err != nil {
		return nil, err
	}
	request, err := parseDurationEnv("HTTP_REQUEST_TIMEOUT")
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:           databaseURL,
		Port:                  port,
		MigrationsDir:         migrationsDir,
		LogFilePath:           logFilePath,
		LogLevel:              logLevel,
		ServiceName:           serviceName,
		HTTPListenHost:        httpListenHost,
		HTTPReadHeaderTimeout: readHdr,
		HTTPShutdownTimeout:   shutdown,
		HTTPRequestTimeout:    request,
	}, nil
}

func (c *Config) ListenAddress() string {
	return net.JoinHostPort(c.HTTPListenHost, strconv.Itoa(c.Port))
}
