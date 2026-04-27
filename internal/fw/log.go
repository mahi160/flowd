package fw

import (
	"log/slog"
	"os"
)

var L *slog.Logger

func InitLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	L = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(L)
}
