package services

import (
	"context"
	"errors"
	"testing"

	"ozon-test-task/internal/storage"
)

type mockStorage struct {
	saveFn      func(ctx context.Context, orig, short string) error
	getShortFn  func(ctx context.Context, orig string) (string, error)
	getOrigFn   func(ctx context.Context, short string) (string, error)
	saveCalls   int
	lookupCalls int
}

func (m *mockStorage) Save(ctx context.Context, orig, short string) error {
	m.saveCalls++
	return m.saveFn(ctx, orig, short)
}

func (m *mockStorage) GetShortByOriginal(ctx context.Context, orig string) (string, error) {
	m.lookupCalls++
	return m.getShortFn(ctx, orig)
}

func (m *mockStorage) GetOriginalByShort(ctx context.Context, short string) (string, error) {
	return m.getOrigFn(ctx, short)
}

type seqGenerator struct {
	values []string
	idx    int
}

func (g *seqGenerator) Generate() string {
	if g.idx >= len(g.values) {
		return "exhausted!"
	}
	v := g.values[g.idx]
	g.idx++
	return v
}

func TestShorten_EmptyURL(t *testing.T) {
	svc := NewURLService(&mockStorage{}, &seqGenerator{values: []string{"abc1234567"}})
	_, err := svc.Shorten(context.Background(), "")
	if !errors.Is(err, ErrEmptyURL) {
		t.Fatalf("expected ErrEmptyURL, got %v", err)
	}
}

func TestShorten_NewURL(t *testing.T) {
	store := &mockStorage{
		getShortFn: func(_ context.Context, _ string) (string, error) {
			return "", storage.ErrNotFound
		},
		saveFn: func(_ context.Context, _, _ string) error { return nil },
	}
	gen := &seqGenerator{values: []string{"abc1234567"}}
	svc := NewURLService(store, gen)

	short, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if short != "abc1234567" {
		t.Fatalf("got %q, want abc1234567", short)
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected 1 save call, got %d", store.saveCalls)
	}
}

func TestShorten_AlreadyExists(t *testing.T) {
	store := &mockStorage{
		getShortFn: func(_ context.Context, _ string) (string, error) {
			return "existing00", nil
		},
		saveFn: func(_ context.Context, _, _ string) error {
			t.Fatalf("save should not be called")
			return nil
		},
	}
	svc := NewURLService(store, &seqGenerator{values: []string{"new0000000"}})

	short, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if short != "existing00" {
		t.Fatalf("got %q, want existing00", short)
	}
}

func TestShorten_CollisionRetry(t *testing.T) {
	saveResults := []error{storage.ErrShortURLExists, storage.ErrShortURLExists, nil}
	saveIdx := 0

	store := &mockStorage{
		getShortFn: func(_ context.Context, _ string) (string, error) {
			return "", storage.ErrNotFound
		},
		saveFn: func(_ context.Context, _, _ string) error {
			err := saveResults[saveIdx]
			saveIdx++
			return err
		},
	}
	gen := &seqGenerator{values: []string{"aaaaaaaaaa", "bbbbbbbbbb", "cccccccccc"}}
	svc := NewURLService(store, gen)

	short, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if short != "cccccccccc" {
		t.Fatalf("got %q, want cccccccccc", short)
	}
	if store.saveCalls != 3 {
		t.Fatalf("expected 3 save calls, got %d", store.saveCalls)
	}
}

func TestShorten_RaceConflict(t *testing.T) {
	lookupResults := []struct {
		val string
		err error
	}{
		{"", storage.ErrNotFound},
		{"raced00000", nil},
	}
	lookupIdx := 0

	store := &mockStorage{
		getShortFn: func(_ context.Context, _ string) (string, error) {
			r := lookupResults[lookupIdx]
			lookupIdx++
			return r.val, r.err
		},
		saveFn: func(_ context.Context, _, _ string) error {
			return storage.ErrOriginalURLExists
		},
	}
	svc := NewURLService(store, &seqGenerator{values: []string{"aaaaaaaaaa"}})

	short, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if short != "raced00000" {
		t.Fatalf("got %q, want raced00000", short)
	}
}

func TestShorten_MaxAttemptsExhausted(t *testing.T) {
	store := &mockStorage{
		getShortFn: func(_ context.Context, _ string) (string, error) {
			return "", storage.ErrNotFound
		},
		saveFn: func(_ context.Context, _, _ string) error {
			return storage.ErrShortURLExists
		},
	}
	svc := NewURLService(store, &seqGenerator{values: []string{"a", "b", "c", "d", "e"}})

	_, err := svc.Shorten(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, storage.ErrShortURLExists) {
		t.Fatalf("expected wrapped ErrShortURLExists, got %v", err)
	}
}

func TestShorten_LookupReturnsUnexpectedError(t *testing.T) {
	boom := errors.New("boom")
	store := &mockStorage{
		getShortFn: func(_ context.Context, _ string) (string, error) {
			return "", boom
		},
	}
	svc := NewURLService(store, &seqGenerator{values: []string{"a"}})
	_, err := svc.Shorten(context.Background(), "https://example.com")
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom, got %v", err)
	}
}

func TestResolve_Success(t *testing.T) {
	store := &mockStorage{
		getOrigFn: func(_ context.Context, _ string) (string, error) {
			return "https://example.com", nil
		},
	}
	svc := NewURLService(store, &seqGenerator{})

	orig, err := svc.Resolve(context.Background(), "abc1234567")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if orig != "https://example.com" {
		t.Fatalf("got %q, want https://example.com", orig)
	}
}

func TestResolve_NotFound(t *testing.T) {
	store := &mockStorage{
		getOrigFn: func(_ context.Context, _ string) (string, error) {
			return "", storage.ErrNotFound
		},
	}
	svc := NewURLService(store, &seqGenerator{})
	_, err := svc.Resolve(context.Background(), "missing000")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolve_EmptyShort(t *testing.T) {
	svc := NewURLService(&mockStorage{}, &seqGenerator{})
	_, err := svc.Resolve(context.Background(), "")
	if !errors.Is(err, ErrEmptyURL) {
		t.Fatalf("expected ErrEmptyURL, got %v", err)
	}
}
