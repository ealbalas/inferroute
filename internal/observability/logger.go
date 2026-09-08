package observability

import (
	"log/slog"
	"os"
)

// NewLogger returns a structured JSON logger suitable for production.
// Log level is read from the LOG_LEVEL environment variable (default: info).
func NewLogger() *slog.Logger {
	level := slog.LevelInfo
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		_ = level.UnmarshalText([]byte(v))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
