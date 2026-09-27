package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	PhotoURL  string    `json:"photo_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) UpsertUser(ctx context.Context, u User) (User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, username, first_name, last_name, photo_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			username   = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name  = EXCLUDED.last_name,
			photo_url  = EXCLUDED.photo_url,
			updated_at = now()
		RETURNING id, username, first_name, last_name, photo_url, created_at, updated_at
	`, u.ID, u.Username, u.FirstName, u.LastName, u.PhotoURL)

	var saved User
	if err := row.Scan(
		&saved.ID, &saved.Username, &saved.FirstName,
		&saved.LastName, &saved.PhotoURL, &saved.CreatedAt, &saved.UpdatedAt,
	); err != nil {
		return User{}, fmt.Errorf("upsert user: %w", err)
	}
	return saved, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, first_name, last_name, photo_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)

	var u User
	err := row.Scan(
		&u.ID, &u.Username, &u.FirstName,
		&u.LastName, &u.PhotoURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}
