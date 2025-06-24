package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tt-demo/internal/config"
	"tt-demo/internal/http-server/handlers/get"
	"tt-demo/internal/http-server/handlers/remove"
	"tt-demo/internal/http-server/handlers/save"
	"tt-demo/internal/http-server/handlers/update"
	"tt-demo/internal/storage/tarantool"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("start settings", slog.String("env", cfg.Env))
	log.Debug("debug enabled")

	storage, err := tarantool.New(cfg.DB, log)
	if err != nil {
		log.Error("failed to init storage", err)
		os.Exit(1)
	}
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/keys/{key}", get.New(log, storage))
	router.Post("/keys", save.New(log, storage))
	router.Put("/keys/{key}", update.New(log, storage))
	router.Delete("/keys/{key}", remove.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", err)
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("failed to close storage", err)
		}
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", err)

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		panic(fmt.Errorf("unknown environment %s", env))
	}

	return logger
}
