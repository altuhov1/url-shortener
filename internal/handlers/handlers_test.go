package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"ozon-test-task/internal/services"
)

type mockSvc struct {
	shortenFn func(ctx context.Context, orig string) (string, error)
	resolveFn func(ctx context.Context, short string) (string, error)
}

func (m *mockSvc) Shorten(ctx context.Context, orig string) (string, error) {
	return m.shortenFn(ctx, orig)
}

func (m *mockSvc) Resolve(ctx context.Context, short string) (string, error) {
	return m.resolveFn(ctx, short)
}

func newRouter(h *Handler) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods(http.MethodPost)
	r.HandleFunc("/shorten/{short}", h.Resolve).Methods(http.MethodGet)
	return r
}

func TestShortenHandler_Success(t *testing.T) {
	svc := &mockSvc{
		shortenFn: func(_ context.Context, orig string) (string, error) {
			if orig != "https://example.com" {
				t.Fatalf("unexpected orig: %q", orig)
			}
			return "abc1234567", nil
		},
	}
	srv := newRouter(NewHandler(svc))

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp shortenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ShortURL != "abc1234567" {
		t.Fatalf("got %q, want abc1234567", resp.ShortURL)
	}
}

func TestShortenHandler_InvalidJSON(t *testing.T) {
	srv := newRouter(NewHandler(&mockSvc{}))
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`not-json`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", rr.Code)
	}
}

func TestShortenHandler_EmptyURL(t *testing.T) {
	srv := newRouter(NewHandler(&mockSvc{}))
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":""}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", rr.Code)
	}
}

func TestShortenHandler_ServiceError(t *testing.T) {
	svc := &mockSvc{
		shortenFn: func(_ context.Context, _ string) (string, error) {
			return "", errors.New("boom")
		},
	}
	srv := newRouter(NewHandler(svc))
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500", rr.Code)
	}
}

func TestResolveHandler_Success(t *testing.T) {
	svc := &mockSvc{
		resolveFn: func(_ context.Context, short string) (string, error) {
			if short != "abc1234567" {
				t.Fatalf("unexpected short: %q", short)
			}
			return "https://example.com", nil
		},
	}
	srv := newRouter(NewHandler(svc))
	req := httptest.NewRequest(http.MethodGet, "/shorten/abc1234567", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rr.Code)
	}
	var resp resolveResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.OriginalURL != "https://example.com" {
		t.Fatalf("got %q, want https://example.com", resp.OriginalURL)
	}
}

func TestResolveHandler_NotFound(t *testing.T) {
	svc := &mockSvc{
		resolveFn: func(_ context.Context, _ string) (string, error) {
			return "", services.ErrNotFound
		},
	}
	srv := newRouter(NewHandler(svc))
	req := httptest.NewRequest(http.MethodGet, "/shorten/missing000", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404", rr.Code)
	}
}
