package config

import "github.com/jackc/pgx/v5/pgxpool"

type Config struct {
	pool *pgxpool.Pool
	
}
