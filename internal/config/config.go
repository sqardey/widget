package config

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port      string
	Timezone  string
	Dsn       string
	BotSecret string
}

func Load(ctx context.Context, logger *slog.Logger) (*Config, error) {
	port := strings.Trim(os.Getenv("PORT"), `"' `)
	timezone := "Europe/Moscow"
	dsn := os.Getenv("CONN_STRING")
	botSecret := os.Getenv("BOT_SECRET")

	if port == "" {
		port = "8080"
	}

	if dsn == "" {
		logger.Warn("dsn must be set in environment")
		return nil, errors.New("dsn not be in environment")
	}
	if botSecret == "" {
		logger.Warn("Bot secret must be set in environment")
		return nil, errors.New("Bot secret not be in environment")
	}

	return &Config{Port: port, Timezone: timezone, Dsn: dsn, BotSecret: botSecret}, nil
}
