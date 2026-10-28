package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	WidgetAPIKey    string
	Timezone        string
	RefreshInterval time.Duration
}

func Load() *Config {
	port := os.Getenv("PORT")
	widgetAPIKey := os.Getenv("APIKEY")
	timezone := "Europe/Moscow"
	refreshInterval := 60

	if port == "" || widgetAPIKey == "" {
		return nil
	}

	return &Config{Port: port, WidgetAPIKey: widgetAPIKey,
		Timezone: timezone, RefreshInterval: time.Duration(refreshInterval)}
}
