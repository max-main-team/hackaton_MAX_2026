package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrResumeNotFound = errors.New("resume not found")

type Resume struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	Title            string    `json:"title"`
	Skills           string    `json:"skills"`
	ExperienceMonths int       `json:"experience_months"`
	About            string    `json:"about"`
	Education        string    `json:"education"`
	Links            string    `json:"links"`
	City             string    `json:"city"`
	WorkFormat       string    `json:"work_format"`
	EmploymentType   string    `json:"employment_type"`
	SalaryMin        *int      `json:"salary_min"`
	SalaryMax        *int      `json:"salary_max"`
	Source           string    `json:"source"`
	SourceText       string    `json:"source_text"`
	ParseStatus      string    `json:"parse_status"`
	IsActive         bool      `json:"is_active"`
	LastConfirmedAt  time.Time `json:"last_confirmed_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ResumeVersion struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Snapshot  string    `json:"snapshot"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

type ResumeRepo struct {
	pool *pgxpool.Pool
}

func NewResumeRepo(pool *pgxpool.Pool) *ResumeRepo {
	return &ResumeRepo{pool: pool}
}

const resumeColumns = `
	id, user_id, title, skills, experience_months, about, education, links,
	city, work_format, employment_type, salary_min, salary_max,
	source, source_text, parse_status, is_active, last_confirmed_at, created_at, updated_at
`

func scanResume(row pgx.Row) (Resume, error) {
	var r Resume
	var links []byte
	err := row.Scan(
		&r.ID, &r.UserID, &r.Title, &r.Skills, &r.ExperienceMonths,
		&r.About, &r.Education, &links, &r.City, &r.WorkFormat,
		&r.EmploymentType, &r.SalaryMin, &r.SalaryMax,
		&r.Source, &r.SourceText, &r.ParseStatus, &r.IsActive,
		&r.LastConfirmedAt, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return Resume{}, err
	}
	r.Links = string(links)
	return r, nil
}

// UpsertResume сохраняет резюме и пишет снапшот в resume_versions
// в одной транзакции. Матчится всегда только текущая версия.
func (r *ResumeRepo) UpsertResume(ctx context.Context, res Resume) (Resume, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Resume{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		INSERT INTO resumes (user_id, title, skills, experience_months, about, education,
		                     links, city, work_format, employment_type, salary_min, salary_max,
		                     source, source_text, parse_status, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (user_id) DO UPDATE SET
			title             = EXCLUDED.title,
			skills            = EXCLUDED.skills,
			experience_months = EXCLUDED.experience_months,
			about             = EXCLUDED.about,
			education         = EXCLUDED.education,
			links             = EXCLUDED.links,
			city              = EXCLUDED.city,
			work_format       = EXCLUDED.work_format,
			employment_type   = EXCLUDED.employment_type,
			salary_min        = EXCLUDED.salary_min,
			salary_max        = EXCLUDED.salary_max,
			source            = EXCLUDED.source,
			source_text       = EXCLUDED.source_text,
			parse_status      = EXCLUDED.parse_status,
			is_active         = EXCLUDED.is_active,
			updated_at        = now()
		RETURNING `+resumeColumns,
		res.UserID, res.Title, res.Skills, res.ExperienceMonths, res.About,
		res.Education, res.Links, res.City, res.WorkFormat, res.EmploymentType,
		res.SalaryMin, res.SalaryMax, res.Source, res.SourceText, res.ParseStatus, res.IsActive,
	)
	saved, err := scanResume(row)
	if err != nil {
		return Resume{}, fmt.Errorf("upsert resume: %w", err)
	}

	snapshot, err := json.Marshal(saved)
	if err != nil {
		return Resume{}, fmt.Errorf("marshal resume snapshot: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO resume_versions (user_id, snapshot, source)
		VALUES ($1, $2, $3)
	`, saved.UserID, snapshot, saved.Source); err != nil {
		return Resume{}, fmt.Errorf("insert resume version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Resume{}, fmt.Errorf("commit resume: %w", err)
	}
	return saved, nil
}

func (r *ResumeRepo) GetByUserID(ctx context.Context, userID int64) (Resume, error) {
	saved, err := scanResume(r.pool.QueryRow(ctx, `
		SELECT `+resumeColumns+` FROM resumes WHERE user_id = $1
	`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resume{}, ErrResumeNotFound
	}
	if err != nil {
		return Resume{}, fmt.Errorf("get resume: %w", err)
	}
	return saved, nil
}

func (r *ResumeRepo) ListActive(ctx context.Context) ([]Resume, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+resumeColumns+` FROM resumes WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, fmt.Errorf("list active resumes: %w", err)
	}
	defer rows.Close()

	var out []Resume
	for rows.Next() {
		saved, err := scanResume(rows)
		if err != nil {
			return nil, fmt.Errorf("scan resume: %w", err)
		}
		out = append(out, saved)
	}
	return out, rows.Err()
}

func (r *ResumeRepo) TouchConfirmedAt(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE resumes SET last_confirmed_at = now() WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("touch confirmed_at: %w", err)
	}
	return nil
}

// SetConfirmActivity обрабатывает ответ на еженедельный опрос:
// active=true — подтверждает актуальность, active=false — отключает резюме.
func (r *ResumeRepo) SetConfirmActivity(ctx context.Context, userID int64, active bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE resumes
		SET is_active         = $2,
		    last_confirmed_at = CASE WHEN $2 THEN now() ELSE last_confirmed_at END
		WHERE user_id = $1
	`, userID, active)
	if err != nil {
		return fmt.Errorf("set confirm activity: %w", err)
	}
	return nil
}

func (r *ResumeRepo) ListVersions(ctx context.Context, userID int64) ([]ResumeVersion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, snapshot::text, source, created_at
		FROM resume_versions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list resume versions: %w", err)
	}
	defer rows.Close()

	var out []ResumeVersion
	for rows.Next() {
		var v ResumeVersion
		if err := rows.Scan(&v.ID, &v.UserID, &v.Snapshot, &v.Source, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan resume version: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// SetParseStatus обновляет статус AI-парсинга резюме.
func (r *ResumeRepo) SetParseStatus(ctx context.Context, userID int64, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE resumes SET parse_status = $2, updated_at = now() WHERE user_id = $1
	`, userID, status)
	if err != nil {
		return fmt.Errorf("set parse status: %w", err)
	}
	return nil
}
