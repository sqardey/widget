package repository

import "context"

func (s *Storage) GetUserByKey(ctx context.Context, key string) (*User, error) {
	sqlQuery := `SELECT * FROM widget_users WHERE key = $1`
	s.Pool.QueryRow(ctx, sqlQuery, key)

}
