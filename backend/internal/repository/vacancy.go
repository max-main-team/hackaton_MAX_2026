package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"max-miniapp/backend/internal/geo"
)

type VacancyUpdate struct {
	Title               *string
	Description         *string
	RequiredSkills      *string
	MinExperienceMonths *int
	City                *string
	WorkFormat          *string
	EmploymentType      *string
	SalaryMin           *int
	SalaryMax           *int
	ResponseTTLHours    *int
	IsActive            *bool
}

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
	Lat                 *float64  `json:"lat"`
	Lng                 *float64  `json:"lng"`
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
	city, work_format, employment_type, salary_min, salary_max, lat, lng,
	response_ttl_hours, is_active, created_at, updated_at
`

func scanVacancy(row pgx.Row) (Vacancy, error) {
	var v Vacancy
	err := row.Scan(
		&v.ID, &v.CompanyID, &v.Title, &v.Description, &v.RequiredSkills,
		&v.MinExperienceMonths, &v.City, &v.WorkFormat, &v.EmploymentType,
		&v.SalaryMin, &v.SalaryMax, &v.Lat, &v.Lng, &v.ResponseTTLHours,
		&v.IsActive, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return Vacancy{}, err
	}
	return v, nil
}

func (r *VacancyRepo) Create(ctx context.Context, v Vacancy) (Vacancy, error) {
	if v.Lat == nil {
		if coords, ok := geo.Lookup(v.City); ok {
			lat, lng := coords.Lat, coords.Lng
			v.Lat, v.Lng = &lat, &lng
		}
	}
	saved, err := scanVacancy(r.pool.QueryRow(ctx, `
		INSERT INTO vacancies (company_id, title, description, required_skills,
		                       min_experience_months, city, work_format, employment_type,
		                       salary_min, salary_max, lat, lng, response_ttl_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING `+vacancyColumns,
		v.CompanyID, v.Title, v.Description, v.RequiredSkills,
		v.MinExperienceMonths, v.City, v.WorkFormat, v.EmploymentType,
		v.SalaryMin, v.SalaryMax, v.Lat, v.Lng, v.ResponseTTLHours,
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
func (r *VacancyRepo) Update(ctx context.Context, id int64, in VacancyUpdate) (Vacancy, error) {
	saved, err := scanVacancy(r.pool.QueryRow(ctx, `
		UPDATE vacancies SET
			title                 = COALESCE($2, title),
			description           = COALESCE($3, description),
			required_skills       = COALESCE($4, required_skills),
			min_experience_months = COALESCE($5, min_experience_months),
			city                  = COALESCE($6, city),
			work_format           = COALESCE($7, work_format),
			employment_type       = COALESCE($8, employment_type),
			salary_min            = COALESCE($9, salary_min),
			salary_max            = COALESCE($10, salary_max),
			response_ttl_hours    = COALESCE($11, response_ttl_hours),
			is_active             = COALESCE($12, is_active),
			updated_at            = now()
		WHERE id = $1
		RETURNING `+vacancyColumns,
		id, in.Title, in.Description, in.RequiredSkills, in.MinExperienceMonths,
		in.City, in.WorkFormat, in.EmploymentType, in.SalaryMin, in.SalaryMax,
		in.ResponseTTLHours, in.IsActive,
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

type MapVacancy struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	CompanyName string  `json:"company_name"`
	Verified    bool    `json:"verified"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	SalaryMin   *int    `json:"salary_min"`
	SalaryMax   *int    `json:"salary_max"`
}

// MapVacancies — активные вакансии с координатами для карты.
func (r *VacancyRepo) MapVacancies(ctx context.Context) ([]MapVacancy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT v.id, v.title, c.name, c.verified, v.city, v.lat, v.lng,
		       v.salary_min, v.salary_max
		FROM vacancies v
		JOIN companies c ON c.id = v.company_id
		WHERE v.is_active = TRUE AND v.lat IS NOT NULL AND v.lng IS NOT NULL
		ORDER BY v.id
	`)
	if err != nil {
		return nil, fmt.Errorf("map vacancies: %w", err)
	}
	defer rows.Close()

	var out []MapVacancy
	for rows.Next() {
		var m MapVacancy
		if err := rows.Scan(&m.ID, &m.Title, &m.CompanyName, &m.Verified,
			&m.City, &m.Lat, &m.Lng, &m.SalaryMin, &m.SalaryMax); err != nil {
			return nil, fmt.Errorf("scan map vacancy: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
