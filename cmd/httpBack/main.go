package main

import (
	"log/slog"
	"os"
	"ozon-test-task/internal/app"
	"ozon-test-task/internal/config"
)

func main() {
	cfg := config.MustLoadConfigApp()

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetDefault(logger)
	app := app.NewApp(cfg)
	app.Run()
}
