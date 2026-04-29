package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"ozon-test-task/internal/services"
)

type URLService interface {
	Shorten(ctx context.Context, original string) (string, error)
	Resolve(ctx context.Context, short string) (string, error)
}

type Handler struct {
	svc URLService
}

func NewHandler(svc URLService) *Handler {
	return &Handler{svc: svc}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

type resolveResponse struct {
	OriginalURL string `json:"original_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	short, err := h.svc.Shorten(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, services.ErrEmptyURL) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("shorten failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, shortenResponse{ShortURL: short})
}

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]

	original, err := h.svc.Resolve(r.Context(), short)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			writeError(w, http.StatusNotFound, "not found")
		case errors.Is(err, services.ErrEmptyURL):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			slog.Error("resolve failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, resolveResponse{OriginalURL: original})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
