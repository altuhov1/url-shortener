package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ozon-test-task/internal/storage"
)

const (
	pgUniqueViolationCode      = "23505"
	urlsPrimaryKey             = "urls_pkey"
	urlsOriginalURLUniqueIndex = "urls_original_url_key"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) Save(ctx context.Context, original, short string) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`,
		short, original,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		switch pgErr.ConstraintName {
		case urlsPrimaryKey:
			return storage.ErrShortURLExists
		case urlsOriginalURLUniqueIndex:
			return storage.ErrOriginalURLExists
		}
	}
	return fmt.Errorf("insert url: %w", err)
}

func (s *Storage) GetShortByOriginal(ctx context.Context, original string) (string, error) {
	var short string
	err := s.pool.QueryRow(
		ctx,
		`SELECT short_url FROM urls WHERE original_url = $1`,
		original,
	).Scan(&short)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrNotFound
		}
		return "", fmt.Errorf("query short by original: %w", err)
	}
	return short, nil
}

func (s *Storage) GetOriginalByShort(ctx context.Context, short string) (string, error) {
	var original string
	err := s.pool.QueryRow(
		ctx,
		`SELECT original_url FROM urls WHERE short_url = $1`,
		short,
	).Scan(&original)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrNotFound
		}
		return "", fmt.Errorf("query original by short: %w", err)
	}
	return original, nil
}
