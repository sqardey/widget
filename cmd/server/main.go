package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/config"
	"github.com/sqardey/widget/internal/parser"
	httptransport "github.com/sqardey/widget/internal/transport/http"
)

func main() {
	cfg := config.Load()
	tmz, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("Something went wrong, error: ", err)
	}
	parser := parser.New(cfg.JournalLogin, cfg.JournalPassword)
	cache := cache.New(tmz)
	handler := httptransport.NewHandler(cache, cfg.WidgetAPIKey)
	cache.StartWorker(context.Background(), parser, cfg.RefreshInterval)

	log.Printf("Server started on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler.Routes()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
