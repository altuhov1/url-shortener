package services

import (
	"context"
	"errors"
	"fmt"

	"ozon-test-task/internal/storage"
)

const defaultMaxAttempts = 5

var (
	ErrEmptyURL = errors.New("empty url")
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	Save(ctx context.Context, original, short string) error
	GetShortByOriginal(ctx context.Context, original string) (string, error)
	GetOriginalByShort(ctx context.Context, short string) (string, error)
}

type ShortGenerator interface {
	Generate() string
}

type URLService struct {
	storage     Storage
	generator   ShortGenerator
	maxAttempts int
}

func NewURLService(s Storage, g ShortGenerator) *URLService {
	return &URLService{
		storage:     s,
		generator:   g,
		maxAttempts: defaultMaxAttempts,
	}
}

func (u *URLService) Shorten(ctx context.Context, original string) (string, error) {
	if original == "" {
		return "", ErrEmptyURL
	}

	short, err := u.storage.GetShortByOriginal(ctx, original)
	if err == nil {
		return short, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return "", fmt.Errorf("lookup by original: %w", err)
	}

	var lastErr error
	for i := 0; i < u.maxAttempts; i++ {
		candidate := u.generator.Generate()
		saveErr := u.storage.Save(ctx, original, candidate)
		switch {
		case saveErr == nil:
			return candidate, nil
		case errors.Is(saveErr, storage.ErrShortURLExists):
			lastErr = saveErr
			continue
		case errors.Is(saveErr, storage.ErrOriginalURLExists):
			existing, getErr := u.storage.GetShortByOriginal(ctx, original)
			if getErr != nil {
				return "", fmt.Errorf("fetch existing after race: %w", getErr)
			}
			return existing, nil
		default:
			return "", fmt.Errorf("save: %w", saveErr)
		}
	}

	return "", fmt.Errorf("could not generate unique short url after %d attempts: %w", u.maxAttempts, lastErr)
}

func (u *URLService) Resolve(ctx context.Context, short string) (string, error) {
	if short == "" {
		return "", ErrEmptyURL
	}

	original, err := u.storage.GetOriginalByShort(ctx, short)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("resolve: %w", err)
	}
	return original, nil
}
