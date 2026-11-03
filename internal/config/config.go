package config

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	JournalLogin    string
	JournalPassword string
	Port            string
	WidgetAPIKey    string
	Timezone        string
	Dsn             string
	RefreshInterval time.Duration
}

func Load(ctx context.Context) *Config {
	port := strings.Trim(os.Getenv("PORT"), `"' `)
	widgetAPIKey := os.Getenv("APIKEY")
	login := os.Getenv("JOURNAL_LOGIN")
	password := os.Getenv("JOURNAL_PASSWORD")
	timezone := "Europe/Moscow"
	refreshInterval := 60
	dsn := os.Getenv("CONN_STRING")

	if login == "" || password == "" {
		log.Fatal("JOURNAL_LOGIN and JOURNAL_PASSWORD must be set in environment")
	}
	if widgetAPIKey == "" {
		log.Fatal("Widget APIKey must be set in environment")
	}
	if port == "" {
		port = "8080"
	}

	if dsn == "" {
		log.Fatal("dsn must ve set in environment")
	}

	return &Config{Port: port, WidgetAPIKey: widgetAPIKey,
		Timezone: timezone, RefreshInterval: time.Duration(refreshInterval) * time.Minute,
		JournalLogin: login, JournalPassword: password, Dsn: dsn}
}
