package inmemory

import (
	"context"
	"sync"

	"ozon-test-task/internal/storage"
)

type Storage struct {
	mu              sync.RWMutex
	shortToOriginal map[string]string
	originalToShort map[string]string
}

func New() *Storage {
	return &Storage{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

func (s *Storage) Save(_ context.Context, original, short string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.originalToShort[original]; ok {
		return storage.ErrOriginalURLExists
	}
	if _, ok := s.shortToOriginal[short]; ok {
		return storage.ErrShortURLExists
	}

	s.shortToOriginal[short] = original
	s.originalToShort[original] = short
	return nil
}

func (s *Storage) GetShortByOriginal(_ context.Context, original string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	short, ok := s.originalToShort[original]
	if !ok {
		return "", storage.ErrNotFound
	}
	return short, nil
}

func (s *Storage) GetOriginalByShort(_ context.Context, short string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	original, ok := s.shortToOriginal[short]
	if !ok {
		return "", storage.ErrNotFound
	}
	return original, nil
}
