package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/config"
	"github.com/sqardey/widget/internal/parser"
	httptransport "github.com/sqardey/widget/internal/transport/http"
)

func main() {
	// Контекст для фонового воркера и остановки сервера
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := config.Load(ctx)
	tmz, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	parserClient := parser.New(cfg.JournalLogin, cfg.JournalPassword)
	scheduleCache := cache.New(tmz)
	handler := httptransport.NewHandler(scheduleCache, cfg.WidgetAPIKey)

	scheduleCache.StartWorker(ctx, parserClient, cfg.RefreshInterval)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler.Routes(),
	}

	// Запуск HTTP-сервера в фоновой горутине
	go func() {
		log.Printf("Server started on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Ожидание системного сигнала SIGINT / SIGTERM
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped cleanly")
}
