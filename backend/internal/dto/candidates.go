package dto

import (
	"max-miniapp/backend/internal/scoring"
)

type CandidateItem struct {
	Score      int               `json:"score" example:"87"`
	FinalScore int               `json:"final_score" example:"85"`
	AIScore    *int              `json:"ai_score"`
	AIComment  string            `json:"ai_comment"`
	Breakdown  scoring.Breakdown `json:"breakdown"`
	User       User              `json:"user"`
	Resume     Resume            `json:"resume"`
}

type CandidatesResponse struct {
	Total int             `json:"total"`
	Items []CandidateItem `json:"items"`
}

type CandidateCard struct {
	User   User   `json:"user"`
	Resume Resume `json:"resume"`
}

type AllCandidatesResponse struct {
	Total int             `json:"total"`
	Items []CandidateCard `json:"items"`
}
