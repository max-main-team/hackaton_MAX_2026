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
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	PhotoURL     string    `json:"photo_url"`
	LanguageCode string    `json:"language_code"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) UpsertUser(ctx context.Context, u User) (User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, username, first_name, last_name, photo_url, language_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			username      = EXCLUDED.username,
			first_name    = EXCLUDED.first_name,
			last_name     = EXCLUDED.last_name,
			photo_url     = EXCLUDED.photo_url,
			language_code = EXCLUDED.language_code,
			updated_at    = now()
		RETURNING id, username, first_name, last_name, photo_url, language_code, COALESCE(role, '') AS role, created_at, updated_at
	`, u.ID, u.Username, u.FirstName, u.LastName, u.PhotoURL, u.LanguageCode)

	var saved User
	if err := row.Scan(
		&saved.ID, &saved.Username, &saved.FirstName,
		&saved.LastName, &saved.PhotoURL, &saved.LanguageCode, &saved.Role,
		&saved.CreatedAt, &saved.UpdatedAt,
	); err != nil {
		return User{}, fmt.Errorf("upsert user: %w", err)
	}
	return saved, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, first_name, last_name, photo_url, language_code, COALESCE(role, '') AS role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)

	var u User
	err := row.Scan(
		&u.ID, &u.Username, &u.FirstName,
		&u.LastName, &u.PhotoURL, &u.LanguageCode, &u.Role,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// UpdateRole меняет роль пользователя (candidate/recruiter).
func (r *UserRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET role = $2, updated_at = now() WHERE id = $1
	`, id, role)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

const ConsentPersonalData = "personal_data"

// UpsertConsent фиксирует согласие пользователя (повторное — обновляет дату).
func (r *UserRepo) UpsertConsent(ctx context.Context, userID int64, consent string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_consents (user_id, consent)
		VALUES ($1, $2)
		ON CONFLICT (user_id, consent) DO UPDATE SET accepted_at = now()
	`, userID, consent)
	if err != nil {
		return fmt.Errorf("upsert consent: %w", err)
	}
	return nil
}

// GetConsentAcceptedAt возвращает дату согласия или nil, если его нет.
func (r *UserRepo) GetConsentAcceptedAt(ctx context.Context, userID int64, consent string) (*time.Time, error) {
	var at *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT accepted_at
		FROM user_consents
		WHERE user_id = $1 AND consent = $2
	`, userID, consent).Scan(&at)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get consent: %w", err)
	}
	return at, nil
}

// SetReferrerIfEmpty записывает реферера только если он ещё не установлен.
func (r *UserRepo) SetReferrerIfEmpty(ctx context.Context, userID, referrerID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET referrer_id = $2, referrer_at = now()
		WHERE id = $1 AND referrer_id IS NULL AND id <> $2
	`, userID, referrerID)
	if err != nil {
		return fmt.Errorf("set referrer: %w", err)
	}
	return nil
}

type Referral struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	JoinedAt  time.Time `json:"joined_at"`
}

func (r *UserRepo) ListReferrals(ctx context.Context, userID int64) ([]Referral, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, first_name, created_at FROM users WHERE referrer_id = $1 ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list referrals: %w", err)
	}
	defer rows.Close()

	var out []Referral
	for rows.Next() {
		var ref Referral
		if err := rows.Scan(&ref.ID, &ref.FirstName, &ref.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan referral: %w", err)
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}

// SetPendingRefCompany запоминает компанию-реферера из диплинка refc_,
// если она существует и пользователь ещё не состоит в ней.
func (r *UserRepo) SetPendingRefCompany(ctx context.Context, userID, companyID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET pending_ref_company_id = $2
		WHERE id = $1 AND pending_ref_company_id IS NULL
		  AND EXISTS (SELECT 1 FROM companies WHERE id = $2)
		  AND NOT EXISTS (
			SELECT 1 FROM company_members
			WHERE company_members.user_id = users.id AND company_members.company_id = $2
		  )
	`, userID, companyID)
	if err != nil {
		return fmt.Errorf("set pending ref company: %w", err)
	}
	return nil
}
