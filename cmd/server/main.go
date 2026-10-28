package main

import (
	"log"
	"net/http"
	"time"

	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/config"
	"github.com/sqardey/widget/internal/parser"
)

func main() {
	cfg := config.Load()
	tmz, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("Something went wrong, error: ", err)
	}
	parser := parser.New()
	cache := cache.New(tmz)
	http_transport
	handler := httptransport.NewHandler(cache, cfg.WidgetAPIKey)

	log.Printf("Server started on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler.Routes()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
