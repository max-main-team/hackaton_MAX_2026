package dto

import (
	"max-miniapp/backend/internal/scoring"
)

type CandidateItem struct {
	Score     int               `json:"score" example:"87"`
	Breakdown scoring.Breakdown `json:"breakdown"`
	User      User              `json:"user"`
	Resume    Resume            `json:"resume"`
}

type CandidatesResponse struct {
	Total int             `json:"total"`
	Items []CandidateItem `json:"items"`
}
