package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"ozon-test-task/internal/config"
	"ozon-test-task/internal/handlers"
	"ozon-test-task/internal/middleware"
	"ozon-test-task/internal/services"
	"ozon-test-task/internal/storage"
	"ozon-test-task/internal/storage/inmemory"
	"ozon-test-task/internal/storage/postgres"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 30 * time.Second
	shutdownTimeout = 5 * time.Second
)

type App struct {
	cfg     *config.ConfigApp
	server  *http.Server
	pgPool  *pgxpool.Pool
	rootCtx context.Context
	cancel  context.CancelFunc
}

func NewApp(cfg *config.ConfigApp) *App {
	rootCtx, cancel := context.WithCancel(context.Background())

	a := &App{
		cfg:     cfg,
		rootCtx: rootCtx,
		cancel:  cancel,
	}

	store := a.initStorage()
	svc := services.NewURLService(store, services.NewRandomGenerator())
	handler := handlers.NewHandler(svc)
	a.server = a.buildServer(handler)

	return a
}

func (a *App) initStorage() storage.Storage {
	if a.cfg.Is_bd_in_memory == 1 {
		slog.Info("Using in-memory storage")
		return inmemory.New()
	}

	pool, err := postgres.NewPoolPg(a.cfg)
	if err != nil {
		slog.Error("Failed to initialize PG pool", "error", err)
		os.Exit(1)
	}
	a.pgPool = pool
	slog.Info("Using PostgreSQL storage")
	return postgres.NewStorage(pool)
}

func (a *App) buildServer(h *handlers.Handler) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods(http.MethodPost)
	r.HandleFunc("/shorten/{short}", h.Resolve).Methods(http.MethodGet)

	return &http.Server{
		Addr:         ":" + a.cfg.App_port,
		Handler:      middleware.ContextMiddleware(a.rootCtx, r),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func (a *App) Run() {
	go a.startServer()
	a.waitForShutdown()
}

func (a *App) startServer() {
	slog.Info("Server starting", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func (a *App) waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down gracefully...")
	a.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	if a.pgPool != nil {
		a.pgPool.Close()
	}

	slog.Info("Application stopped")
}
