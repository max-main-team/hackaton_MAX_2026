package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumeScore struct {
	ResumeID  int64
	VacancyID int64
	AlgoScore int
	AIScore   *int
	AIComment string
	AIModel   string
	UpdatedAt time.Time
}

type ScoreRepo struct {
	pool *pgxpool.Pool
}

func NewScoreRepo(pool *pgxpool.Pool) *ScoreRepo {
	return &ScoreRepo{pool: pool}
}

func (r *ScoreRepo) GetScore(ctx context.Context, resumeID, vacancyID int64) (*ResumeScore, error) {
	var s ResumeScore
	err := r.pool.QueryRow(ctx, `
		SELECT resume_id, vacancy_id, algo_score, ai_score, ai_comment, ai_model, updated_at
		FROM resume_scores
		WHERE resume_id = $1 AND vacancy_id = $2
	`, resumeID, vacancyID).Scan(
		&s.ResumeID, &s.VacancyID, &s.AlgoScore, &s.AIScore,
		&s.AIComment, &s.AIModel, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get score: %w", err)
	}
	return &s, nil
}

func (r *ScoreRepo) UpsertScore(ctx context.Context, s ResumeScore) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO resume_scores (resume_id, vacancy_id, algo_score, ai_score, ai_comment, ai_model, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (resume_id, vacancy_id) DO UPDATE SET
			algo_score = EXCLUDED.algo_score,
			ai_score   = EXCLUDED.ai_score,
			ai_comment = EXCLUDED.ai_comment,
			ai_model   = EXCLUDED.ai_model,
			updated_at = now()
	`, s.ResumeID, s.VacancyID, s.AlgoScore, s.AIScore, s.AIComment, s.AIModel)
	if err != nil {
		return fmt.Errorf("upsert score: %w", err)
	}
	return nil
}
