package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"ozon-test-task/internal/config"

	_ "ozon-test-task/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.MustLoadConfigMigrate()
	if cfg.Is_bd_in_memory == 1 {
		return
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn("Файл .env не найден, использую системные переменные окружения")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PG_DBUser, cfg.PG_DBPassword, cfg.PG_DBHost, cfg.PG_PORT, cfg.PG_DBName, cfg.PG_DBSSLMode)

	if err := run(connStr); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(connStr string) error {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("did not connect to pgx: %w", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			slog.Error("db close failed", "err", cerr)
		}
	}()

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("did not up goose: %w", err)
	}
	return nil
}
