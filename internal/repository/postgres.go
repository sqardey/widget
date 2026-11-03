package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) *Storage {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Error to connect from database! %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("Error to ping database! %v", err)
	}
	return &Storage{Pool: pool}

}

func (s *Storage) Close() {
	s.Pool.Close()
}
