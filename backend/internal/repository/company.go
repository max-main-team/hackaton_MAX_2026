package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCompanyNotFound = errors.New("company not found")

type Company struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Website           string     `json:"website"`
	LogoURL           string     `json:"logo_url"`
	Address           string     `json:"address"`
	Verified          bool       `json:"verified"`
	BotUserID         *int64     `json:"bot_user_id"`
	BotUsername       string     `json:"bot_username"`
	ReferrerCompanyID *int64     `json:"referrer_company_id"`
	ReferrerAt        *time.Time `json:"referrer_at"`
	PromoUntil        *time.Time `json:"promo_until"`
	InviteQuota       int        `json:"invite_quota"`
	InviteUsed        int        `json:"invite_used"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CompanyRepo struct {
	pool *pgxpool.Pool
}

func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo {
	return &CompanyRepo{pool: pool}
}

const companyColumns = `
	id, name, description, website, logo_url, address, verified,
	bot_user_id, bot_username, referrer_company_id, referrer_at, promo_until,
	invite_quota, invite_used, created_at, updated_at
`

func scanCompany(row pgx.Row) (Company, error) {
	var c Company
	err := row.Scan(
		&c.ID, &c.Name, &c.Description, &c.Website, &c.LogoURL,
		&c.Address, &c.Verified, &c.BotUserID, &c.BotUsername,
		&c.ReferrerCompanyID, &c.ReferrerAt, &c.PromoUntil,
		&c.InviteQuota, &c.InviteUsed, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return Company{}, err
	}
	return c, nil
}

func (r *CompanyRepo) Create(ctx context.Context, c Company, creatorUserID int64, position string) (Company, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Company{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	saved, err := scanCompany(tx.QueryRow(ctx, `
		INSERT INTO companies (name, description, website, logo_url, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+companyColumns,
		c.Name, c.Description, c.Website, c.LogoURL, c.Address,
	))
	if err != nil {
		return Company{}, fmt.Errorf("insert company: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO company_members (company_id, user_id, position)
		VALUES ($1, $2, $3)
	`, saved.ID, creatorUserID, position); err != nil {
		return Company{}, fmt.Errorf("insert company member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Company{}, fmt.Errorf("commit company: %w", err)
	}
	return saved, nil
}

func (r *CompanyRepo) GetByID(ctx context.Context, id int64) (Company, error) {
	saved, err := scanCompany(r.pool.QueryRow(ctx, `
		SELECT `+companyColumns+` FROM companies WHERE id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Company{}, ErrCompanyNotFound
	}
	if err != nil {
		return Company{}, fmt.Errorf("get company: %w", err)
	}
	return saved, nil
}

func (r *CompanyRepo) IsMember(ctx context.Context, companyID, userID int64) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2
		)
	`, companyID, userID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return ok, nil
}

func (r *CompanyRepo) ListByUser(ctx context.Context, userID int64) ([]Company, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+companyColumns+`
		FROM companies c
		JOIN company_members m ON m.company_id = c.id
		WHERE m.user_id = $1
		ORDER BY c.id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list companies by user: %w", err)
	}
	defer rows.Close()

	var out []Company
	for rows.Next() {
		saved, err := scanCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		out = append(out, saved)
	}
	return out, rows.Err()
}

func (r *CompanyRepo) UpdateMemberPosition(ctx context.Context, companyID, userID int64, position string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE company_members SET position = $3
		WHERE company_id = $1 AND user_id = $2
	`, companyID, userID, position)
	if err != nil {
		return fmt.Errorf("update member position: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCompanyNotFound
	}
	return nil
}

func (r *CompanyRepo) MarkVerified(ctx context.Context, id int64, botUserID int64, botUsername string) (Company, error) {
	saved, err := scanCompany(r.pool.QueryRow(ctx, `
		UPDATE companies SET verified = TRUE, bot_user_id = $2, bot_username = $3, updated_at = now()
		WHERE id = $1
		RETURNING `+companyColumns, id, botUserID, botUsername))
	if err != nil {
		return Company{}, fmt.Errorf("mark verified: %w", err)
	}
	return saved, nil
}

type InvitedCompany struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *CompanyRepo) ConsumeInvite(ctx context.Context, companyID int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE companies SET invite_used = invite_used + 1, updated_at = now()
		WHERE id = $1 AND invite_used < invite_quota
	`, companyID)
	if err != nil {
		return false, fmt.Errorf("consume invite: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func (r *CompanyRepo) AddInviteQuota(ctx context.Context, companyID int64, delta int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE companies SET invite_quota = invite_quota + $2, updated_at = now()
		WHERE id = $1
	`, companyID, delta)
	if err != nil {
		return fmt.Errorf("add invite quota: %w", err)
	}
	return nil
}

func (r *CompanyRepo) ListInvitedCompanies(ctx context.Context, companyID int64) ([]InvitedCompany, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, created_at FROM companies
		WHERE referrer_company_id = $1
		ORDER BY created_at
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list invited companies: %w", err)
	}
	defer rows.Close()

	var out []InvitedCompany
	for rows.Next() {
		var c InvitedCompany
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan invited company: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CompanyRepo) ApplyB2BAttribution(ctx context.Context, companyID, creatorUserID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin attribution tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var pending *int64
	err = tx.QueryRow(ctx, `
		SELECT pending_ref_company_id FROM users
		WHERE id = $1 AND pending_ref_company_id IS NOT NULL
	`, creatorUserID).Scan(&pending)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read pending ref: %w", err)
	}

	var members int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM company_members WHERE user_id = $1
	`, creatorUserID).Scan(&members); err != nil {
		return fmt.Errorf("count members: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users SET pending_ref_company_id = NULL WHERE id = $1
	`, creatorUserID); err != nil {
		return fmt.Errorf("clear pending ref: %w", err)
	}

	if members > 1 || *pending == companyID {
		return tx.Commit(ctx)
	}

	refTag, err := tx.Exec(ctx, `
		UPDATE companies
		SET referrer_company_id = $2, referrer_at = now(), promo_until = now() + interval '30 days',
		    invite_quota = 6, updated_at = now()
		WHERE id = $1 AND referrer_company_id IS NULL
	`, companyID, *pending)
	if err != nil {
		return fmt.Errorf("set referrer company: %w", err)
	}

	if refTag.RowsAffected() == 1 {
		if _, err := tx.Exec(ctx, `
			UPDATE companies SET invite_quota = invite_quota + 5, updated_at = now()
			WHERE id = $1
		`, *pending); err != nil {
			return fmt.Errorf("reward referrer: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit attribution: %w", err)
	}
	return nil
}
