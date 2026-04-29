package postgres

import (
	"context"
	"fmt"
	"ozon-test-task/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPoolPg(cfg *config.ConfigApp) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PG_DBUser, cfg.PG_DBPassword, cfg.PG_DBHost, cfg.PG_PORT, cfg.PG_DBName, cfg.PG_DBSSLMode)
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка конфигурации: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("база не отвечает: %w", err)
	}

	return pool, nil
}
