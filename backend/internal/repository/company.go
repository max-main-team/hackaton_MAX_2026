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
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	LogoURL     string    `json:"logo_url"`
	Address     string    `json:"address"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CompanyRepo struct {
	pool *pgxpool.Pool
}

func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo {
	return &CompanyRepo{pool: pool}
}

const companyColumns = `
	id, name, description, website, logo_url, address, verified, created_at, updated_at
`

func scanCompany(row pgx.Row) (Company, error) {
	var c Company
	err := row.Scan(
		&c.ID, &c.Name, &c.Description, &c.Website, &c.LogoURL,
		&c.Address, &c.Verified, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return Company{}, err
	}
	return c, nil
}

// Create создаёт компанию и сразу добавляет создателя участником
// с выбранной позицией (owner/hr/employee) — в одной транзакции.
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

// UpdateMemberPosition меняет позицию участника в компании (owner/hr/employee).
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
