package repository

import (
	"context"
	"time"

	"github.com/sqardey/widget/internal/domain"
)

func (s *Storage) GetUserByKey(ctx context.Context, key string) (*domain.User, error) {
	sqlQuery := `SELECT key, login, password, updated_at FROM widget_users WHERE key = $1`
	var user domain.User
	err := s.Pool.QueryRow(ctx, sqlQuery, key).Scan(&user.Key, &user.Login, &user.Password, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) PutNewUser(ctx context.Context, tg_id int64, login, password string) (string, error) {
	sqlQuery := `INSERT INTO widget_users (tg_id, login, password)
	VALUES ($1, $2, $3)
	ON CONFLICT (tg_id) DO UPDATE
	SET login = EXCLUDED.login,
	password = EXCLUDED.password
	RETURNING key;`
	var key string
	err := s.Pool.QueryRow(ctx, sqlQuery, tg_id, login, password).Scan(&key)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (s *Storage) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	sqlQuery := `SELECT key, login, password, updated_at FROM widget_users`
	rows, err := s.Pool.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.Key, &user.Login, &user.Password, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Storage) TouchUserUpdatedAt(ctx context.Context, key string, t time.Time) error {
	sqlQuery := `UPDATE widget_users SET updated_at = $1 WHERE key = $2`
	_, err := s.Pool.Exec(ctx, sqlQuery, t, key)
	return err
}
