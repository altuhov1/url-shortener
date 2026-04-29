package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrShortURLExists    = errors.New("short url already exists")
	ErrOriginalURLExists = errors.New("original url already exists")
)

type Storage interface {
	Save(ctx context.Context, original, short string) error
	GetShortByOriginal(ctx context.Context, original string) (string, error)
	GetOriginalByShort(ctx context.Context, short string) (string, error)
}
