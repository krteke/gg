package internal

import (
	"log/slog"
	"os"
)

func InitLogger(verbose int) {
	level := slog.LevelWarn

	switch {
	case verbose == 1:
		level = slog.LevelInfo
	case verbose == 2:
		level = slog.LevelDebug
	case verbose >= 3:
		level = slog.Level(-8)
	}

	logger := slog.New(slog.NewTextHandler(
		os.Stderr, &slog.HandlerOptions{
			Level: level,
		}),
	)
	slog.SetDefault(logger)
}
