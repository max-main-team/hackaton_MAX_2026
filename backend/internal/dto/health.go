package dto

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db" example:"ok"`
}
