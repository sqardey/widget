package config

import (
	"context"
	"log"
	"os"
	"strings"
)

type Config struct {
	Port     string
	Timezone string
	Dsn      string
}

func Load(ctx context.Context) *Config {
	port := strings.Trim(os.Getenv("PORT"), `"' `)
	timezone := "Europe/Moscow"
	dsn := os.Getenv("CONN_STRING")

	if port == "" {
		port = "8080"
	}

	if dsn == "" {
		log.Fatal("dsn must ve set in environment")
	}

	return &Config{Port: port, Timezone: timezone, Dsn: dsn}
}
