package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrVacancyNotFound = errors.New("vacancy not found")

type Vacancy struct {
	ID                  int64     `json:"id"`
	CompanyID           int64     `json:"company_id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	RequiredSkills      string    `json:"required_skills"`
	MinExperienceMonths int       `json:"min_experience_months"`
	City                string    `json:"city"`
	WorkFormat          string    `json:"work_format"`
	EmploymentType      string    `json:"employment_type"`
	SalaryMin           *int      `json:"salary_min"`
	SalaryMax           *int      `json:"salary_max"`
	ResponseTTLHours    int       `json:"response_ttl_hours"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type VacancyRepo struct {
	pool *pgxpool.Pool
}

func NewVacancyRepo(pool *pgxpool.Pool) *VacancyRepo {
	return &VacancyRepo{pool: pool}
}

const vacancyColumns = `
	id, company_id, title, description, required_skills, min_experience_months,
	city, work_format, employment_type, salary_min, salary_max,
	response_ttl_hours, is_active, created_at, updated_at
`

func scanVacancy(row pgx.Row) (Vacancy, error) {
	var v Vacancy
	err := row.Scan(
		&v.ID, &v.CompanyID, &v.Title, &v.Description, &v.RequiredSkills,
		&v.MinExperienceMonths, &v.City, &v.WorkFormat, &v.EmploymentType,
		&v.SalaryMin, &v.SalaryMax, &v.ResponseTTLHours,
		&v.IsActive, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return Vacancy{}, err
	}
	return v, nil
}

func (r *VacancyRepo) Create(ctx context.Context, v Vacancy) (Vacancy, error) {
	saved, err := scanVacancy(r.pool.QueryRow(ctx, `
		INSERT INTO vacancies (company_id, title, description, required_skills,
		                       min_experience_months, city, work_format, employment_type,
		                       salary_min, salary_max, response_ttl_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING `+vacancyColumns,
		v.CompanyID, v.Title, v.Description, v.RequiredSkills,
		v.MinExperienceMonths, v.City, v.WorkFormat, v.EmploymentType,
		v.SalaryMin, v.SalaryMax, v.ResponseTTLHours,
	))
	if err != nil {
		return Vacancy{}, fmt.Errorf("insert vacancy: %w", err)
	}
	return saved, nil
}

func (r *VacancyRepo) GetByID(ctx context.Context, id int64) (Vacancy, error) {
	saved, err := scanVacancy(r.pool.QueryRow(ctx, `
		SELECT `+vacancyColumns+` FROM vacancies WHERE id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Vacancy{}, ErrVacancyNotFound
	}
	if err != nil {
		return Vacancy{}, fmt.Errorf("get vacancy: %w", err)
	}
	return saved, nil
}

// Update частично обновляет вакансию: nil-аргументы не меняются.
func (r *VacancyRepo) Update(ctx context.Context, id int64, ttlHours, minExp *int, active *bool) (Vacancy, error) {
	saved, err := scanVacancy(r.pool.QueryRow(ctx, `
		UPDATE vacancies SET
			response_ttl_hours    = COALESCE($2, response_ttl_hours),
			min_experience_months = COALESCE($3, min_experience_months),
			is_active             = COALESCE($4, is_active),
			updated_at            = now()
		WHERE id = $1
		RETURNING `+vacancyColumns,
		id, ttlHours, minExp, active,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Vacancy{}, ErrVacancyNotFound
	}
	if err != nil {
		return Vacancy{}, fmt.Errorf("update vacancy: %w", err)
	}
	return saved, nil
}

func (r *VacancyRepo) ListByCompany(ctx context.Context, companyID int64) ([]Vacancy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+vacancyColumns+`
		FROM vacancies
		WHERE company_id = $1
		ORDER BY created_at DESC
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list vacancies: %w", err)
	}
	defer rows.Close()

	var out []Vacancy
	for rows.Next() {
		saved, err := scanVacancy(rows)
		if err != nil {
			return nil, fmt.Errorf("scan vacancy: %w", err)
		}
		out = append(out, saved)
	}
	return out, rows.Err()
}
