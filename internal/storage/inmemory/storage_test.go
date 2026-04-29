package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"ozon-test-task/internal/storage"
)

func TestStorage_Save_Success(t *testing.T) {
	s := New()
	if err := s.Save(context.Background(), "https://example.com", "abc1234567"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorage_Save_DuplicateOriginal(t *testing.T) {
	s := New()
	ctx := context.Background()

	if err := s.Save(ctx, "https://example.com", "abc1234567"); err != nil {
		t.Fatalf("first save: %v", err)
	}
	err := s.Save(ctx, "https://example.com", "xyz1234567")
	if !errors.Is(err, storage.ErrOriginalURLExists) {
		t.Fatalf("expected ErrOriginalURLExists, got %v", err)
	}
}

func TestStorage_Save_DuplicateShort(t *testing.T) {
	s := New()
	ctx := context.Background()

	if err := s.Save(ctx, "https://example.com", "abc1234567"); err != nil {
		t.Fatalf("first save: %v", err)
	}
	err := s.Save(ctx, "https://other.com", "abc1234567")
	if !errors.Is(err, storage.ErrShortURLExists) {
		t.Fatalf("expected ErrShortURLExists, got %v", err)
	}
}

func TestStorage_GetShortByOriginal(t *testing.T) {
	s := New()
	ctx := context.Background()
	_ = s.Save(ctx, "https://example.com", "abc1234567")

	short, err := s.GetShortByOriginal(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if short != "abc1234567" {
		t.Fatalf("got %q, want abc1234567", short)
	}

	_, err = s.GetShortByOriginal(ctx, "https://missing.com")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStorage_GetOriginalByShort(t *testing.T) {
	s := New()
	ctx := context.Background()
	_ = s.Save(ctx, "https://example.com", "abc1234567")

	orig, err := s.GetOriginalByShort(ctx, "abc1234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orig != "https://example.com" {
		t.Fatalf("got %q, want https://example.com", orig)
	}

	_, err = s.GetOriginalByShort(ctx, "missing000")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStorage_Concurrent(t *testing.T) {
	s := New()
	ctx := context.Background()

	const n = 200
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			orig := fmt.Sprintf("https://example.com/%d", i)
			short := fmt.Sprintf("s%09d", i)
			if err := s.Save(ctx, orig, short); err != nil {
				t.Errorf("save %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		short := fmt.Sprintf("s%09d", i)
		orig, err := s.GetOriginalByShort(ctx, short)
		if err != nil {
			t.Fatalf("get %d: %v", i, err)
		}
		if orig != fmt.Sprintf("https://example.com/%d", i) {
			t.Fatalf("mismatch for %d: got %q", i, orig)
		}
	}
}
