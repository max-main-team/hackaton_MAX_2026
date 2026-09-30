package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrActionNotFound = errors.New("action not found")

type RecruiterAction struct {
	ID              int64     `json:"id"`
	VacancyID       int64     `json:"vacancy_id"`
	RecruiterUserID int64     `json:"recruiter_user_id"`
	CandidateUserID int64     `json:"candidate_user_id"`
	Action          string    `json:"action"`
	CreatedAt       time.Time `json:"created_at"`
}

type InvitationDetail struct {
	ActionID          int64     `json:"action_id"`
	RecruiterUserID   int64     `json:"recruiter_user_id"`
	CandidateUserID   int64     `json:"candidate_user_id"`
	RecruiterName     string    `json:"recruiter_name"`
	RecruiterLogin    string    `json:"recruiter_login"`
	CompanyID         int64     `json:"company_id"`
	CompanyName       string    `json:"company_name"`
	CompanyVerified   bool      `json:"company_verified"`
	VacancyID         int64     `json:"vacancy_id"`
	VacancyTitle      string    `json:"vacancy_title"`
	VacancyCity       string    `json:"vacancy_city"`
	VacancyWorkFormat string    `json:"vacancy_work_format"`
	VacancySalaryMin  *int32    `json:"vacancy_salary_min"`
	VacancySalaryMax  *int32    `json:"vacancy_salary_max"`
	TTLHours          int       `json:"ttl_hours"`
	CreatedAt         time.Time `json:"created_at"`
	Response          *string   `json:"response"`
}

type MatchingRepo struct {
	pool *pgxpool.Pool
}

func NewMatchingRepo(pool *pgxpool.Pool) *MatchingRepo {
	return &MatchingRepo{pool: pool}
}

func (r *MatchingRepo) GetOrCreateAction(ctx context.Context, vacancyID, recruiterUserID, candidateUserID int64, action string) (RecruiterAction, error) {
	var out RecruiterAction
	err := r.pool.QueryRow(ctx, `
		INSERT INTO recruiter_actions (vacancy_id, recruiter_user_id, candidate_user_id, action)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (vacancy_id, candidate_user_id) DO NOTHING
		RETURNING id, vacancy_id, recruiter_user_id, candidate_user_id, action, created_at
	`, vacancyID, recruiterUserID, candidateUserID, action).Scan(
		&out.ID, &out.VacancyID, &out.RecruiterUserID, &out.CandidateUserID, &out.Action, &out.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = r.pool.QueryRow(ctx, `
			SELECT id, vacancy_id, recruiter_user_id, candidate_user_id, action, created_at
			FROM recruiter_actions
			WHERE vacancy_id = $1 AND candidate_user_id = $2
		`, vacancyID, candidateUserID).Scan(
			&out.ID, &out.VacancyID, &out.RecruiterUserID, &out.CandidateUserID, &out.Action, &out.CreatedAt,
		)
	}
	if err != nil {
		return RecruiterAction{}, fmt.Errorf("get or create action: %w", err)
	}
	return out, nil
}

func (r *MatchingRepo) HasActionForVacancy(ctx context.Context, vacancyID, candidateUserID int64) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM recruiter_actions WHERE vacancy_id = $1 AND candidate_user_id = $2)
	`, vacancyID, candidateUserID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("check action: %w", err)
	}
	return ok, nil
}

func (r *MatchingRepo) GetInvitationDetail(ctx context.Context, actionID int64) (InvitationDetail, error) {
	var d InvitationDetail
	err := r.pool.QueryRow(ctx, `
		SELECT ra.id, ra.recruiter_user_id, ra.candidate_user_id, COALESCE(u.username, ''),
		       c.id, c.name, c.verified,
		       v.id, v.title, v.city, v.work_format, v.salary_min, v.salary_max, v.response_ttl_hours, ra.created_at,
		       cr.response
		FROM recruiter_actions ra
		JOIN vacancies v ON v.id = ra.vacancy_id
		JOIN companies c ON c.id = v.company_id
		JOIN users u ON u.id = ra.recruiter_user_id
		LEFT JOIN candidate_responses cr ON cr.action_id = ra.id
		WHERE ra.id = $1 AND ra.action = 'invite'
	`, actionID).Scan(
		&d.ActionID, &d.RecruiterUserID, &d.CandidateUserID, &d.RecruiterLogin,
		&d.CompanyID, &d.CompanyName, &d.CompanyVerified,
		&d.VacancyID, &d.VacancyTitle, &d.VacancyCity, &d.VacancyWorkFormat, &d.VacancySalaryMin, &d.VacancySalaryMax, &d.TTLHours, &d.CreatedAt,
		&d.Response,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return InvitationDetail{}, ErrActionNotFound
	}
	if err != nil {
		return InvitationDetail{}, fmt.Errorf("get invitation: %w", err)
	}
	return d, nil
}

func (r *MatchingRepo) ListInvitationsByCandidate(ctx context.Context, candidateUserID int64) ([]InvitationDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ra.id, ra.recruiter_user_id, ra.candidate_user_id, COALESCE(u.username, ''),
		       c.id, c.name, c.verified,
		       v.id, v.title, v.city, v.work_format, v.salary_min, v.salary_max, v.response_ttl_hours, ra.created_at,
		       cr.response
		FROM recruiter_actions ra
		JOIN vacancies v ON v.id = ra.vacancy_id
		JOIN companies c ON c.id = v.company_id
		JOIN users u ON u.id = ra.recruiter_user_id
		LEFT JOIN candidate_responses cr ON cr.action_id = ra.id
		WHERE ra.candidate_user_id = $1 AND ra.action = 'invite'
		ORDER BY ra.created_at DESC
	`, candidateUserID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	var out []InvitationDetail
	for rows.Next() {
		var d InvitationDetail
		if err := rows.Scan(
			&d.ActionID, &d.RecruiterUserID, &d.CandidateUserID, &d.RecruiterLogin,
			&d.CompanyID, &d.CompanyName, &d.CompanyVerified,
			&d.VacancyID, &d.VacancyTitle, &d.VacancyCity, &d.VacancyWorkFormat, &d.VacancySalaryMin, &d.VacancySalaryMax, &d.TTLHours, &d.CreatedAt,
			&d.Response,
		); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *MatchingRepo) CreateResponse(ctx context.Context, actionID int64, response string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO candidate_responses (action_id, response)
		VALUES ($1, $2)
	`, actionID, response)
	if err != nil {
		return fmt.Errorf("create response: %w", err)
	}
	return nil
}
