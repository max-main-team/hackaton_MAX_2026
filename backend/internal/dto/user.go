// Package dto — API-модели: то, что реально ходит по HTTP.
// Доменные структуры из internal/repository наружу не отдаются,
// мапперы в этом пакете — единственный переход между слоями.
package dto

import (
	"fmt"
	"time"

	"max-miniapp/backend/internal/repository"
)

type ErrorResponse struct {
	Message string `json:"message" example:"invalid request body"`
}

type User struct {
	ID           int64     `json:"id" example:"42"`
	Username     string    `json:"username" example:"ivan"`
	FirstName    string    `json:"first_name" example:"Ivan"`
	LastName     string    `json:"last_name" example:"Petrov"`
	PhotoURL     string    `json:"photo_url"`
	LanguageCode string    `json:"language_code" example:"ru"`
	Role         string    `json:"role" example:"candidate"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FromUser(u repository.User) User {
	return User{
		ID:           u.ID,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		PhotoURL:     u.PhotoURL,
		LanguageCode: u.LanguageCode,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

type AuthRequest struct {
	InitData string `json:"initData" example:"user=%7B%22id%22%3A42%7D&auth_date=1771409719&hash=..."`
}

type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  User   `json:"user"`
}

type MeResponse struct {
	User                   User       `json:"user"`
	PersonalDataAcceptedAt *time.Time `json:"personal_data_accepted_at"`
}

type RoleRequest struct {
	Role               string `json:"role" enums:"candidate,recruiter" example:"candidate"`
	AcceptPersonalData bool   `json:"acceptPersonalData" example:"true"`
}

var Roles = []string{"candidate", "recruiter"}

func ValidateRole(role string) error {
	for _, r := range Roles {
		if r == role {
			return nil
		}
	}
	return fmt.Errorf("role must be one of: %v", Roles)
}
