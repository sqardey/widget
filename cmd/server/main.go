package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/config"
	"github.com/sqardey/widget/internal/repository"
	httptransport "github.com/sqardey/widget/internal/transport/http"
)

func main() {
	// Контекст для фонового воркера и остановки сервера
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := config.Load(ctx)
	tmz, err := time.LoadLocation(cfg.Timezone)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err != nil {
		logger.Error("failed to load timezone", "Error", err)
		stop()
	}
	db := repository.New(ctx, cfg.Dsn)
	defer db.Close()
	scheduleCache := cache.New(tmz)
	handler := httptransport.NewHandler(scheduleCache, db, logger)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      25 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Запуск HTTP-сервера в фоновой горутине
	go func() {
		logger.Info("Server started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Lister error", "error", err)
			stop()
		}
	}()

	// Ожидание системного сигнала SIGINT / SIGTERM
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("Server forced to shutdown", "error", err)
	}
	logger.Info("Server stopped cleanly")
}
