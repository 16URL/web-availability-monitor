package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webaviable/internal/checker"
	"webaviable/internal/config"
	"webaviable/internal/httpserver"
	"webaviable/internal/scheduler"
	"webaviable/internal/service"
	"webaviable/storage/sqlite"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	sqliter, err := sqlite.New(cfg.Storage_Path)
	if err != nil {
		log.Error("failed to initialize storage", slog.String("storage_path", cfg.Storage_Path), slog.Any("error", err))
		os.Exit(1)
	}
	defer sqliter.Close()

	client := &http.Client{
		Timeout: cfg.HTTPServer.Timeout,
	}
	urlChecker := checker.NewChecker(client)
	service := service.NewService(sqliter, urlChecker)

	httpServer := httpserver.NewHTTPServer(cfg.HTTPServer.Address, cfg.HTTPServer.Timeout, cfg.HTTPServer.IdleTimeout, service)
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	scheduler := scheduler.NewScheduler(service, 5*time.Minute, log, 10)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go scheduler.Start(ctx)

	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.Run(); err != nil && err != http.ErrServerClosed {
			serverErr <- err

		}
	}()
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErr:
		log.Error("HTTP server failed", slog.Any("error", err))
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer shutdownCancel()
	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.Error("failed to stop HTTP server", slog.Any("error", err))
	}

	log.Info("service stopped")
}

const (
	envLocal = "local"
	encDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case encDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
