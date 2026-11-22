package repository

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

func New(ctx context.Context, dsn string, logger *slog.Logger) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("Error to connect from database", "error", err)
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error("Error to ping database", "error", err)
		return nil, err
	}
	return &Storage{Pool: pool, Logger: logger}, nil

}

func (s *Storage) Close() {
	s.Pool.Close()
}
