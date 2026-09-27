package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"max-miniapp/backend/internal/repository"
)

type ResumeInput struct {
	Title            string          `json:"title" example:"Backend-разработчик"`
	Skills           string          `json:"skills" example:"go, postgres, docker"`
	ExperienceMonths int             `json:"experience_months" example:"36"`
	About            string          `json:"about" example:"Пишу сервисы на Go"`
	Education        string          `json:"education" example:"МГТУ, ИТ-факультет"`
	Links            json.RawMessage `json:"links" swaggertype:"array,object"`
	City             string          `json:"city" example:"Москва"`
	WorkFormat       string          `json:"work_format" enums:"onsite,hybrid,remote" example:"remote"`
	EmploymentType   string          `json:"employment_type" enums:"full_time,part_time,contract,internship" example:"full_time"`
	SalaryMin        *int            `json:"salary_min" example:"150000"`
	SalaryMax        *int            `json:"salary_max" example:"250000"`
}

type Resume struct {
	ResumeInput
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Source          string    `json:"source" enums:"manual,file_parse"`
	SourceText      string    `json:"source_text"`
	IsActive        bool      `json:"is_active"`
	LastConfirmedAt time.Time `json:"last_confirmed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ConfirmActivityRequest struct {
	Active bool `json:"active" example:"true"`
}

var WorkFormats = []string{"onsite", "hybrid", "remote"}
var EmploymentTypes = []string{"full_time", "part_time", "contract", "internship"}

type ResumeLink struct {
	Type string `json:"type" example:"github"`
	URL  string `json:"url" example:"https://github.com/ivan"`
}

// ValidateResumeInput проверяет перечисления и формат links.
func ValidateResumeInput(in ResumeInput) error {
	if in.Title == "" {
		return fmt.Errorf("title is required")
	}
	if in.ExperienceMonths < 0 || in.ExperienceMonths > 600 {
		return fmt.Errorf("experience_months must be between 0 and 600")
	}
	if !contains(WorkFormats, in.WorkFormat) {
		return fmt.Errorf("work_format must be one of: %v", WorkFormats)
	}
	if !contains(EmploymentTypes, in.EmploymentType) {
		return fmt.Errorf("employment_type must be one of: %v", EmploymentTypes)
	}
	if in.SalaryMin != nil && *in.SalaryMin < 0 {
		return fmt.Errorf("salary_min must be positive")
	}
	if in.SalaryMax != nil && *in.SalaryMax < 0 {
		return fmt.Errorf("salary_max must be positive")
	}
	if len(in.Links) > 0 {
		var links []ResumeLink
		if err := json.Unmarshal(in.Links, &links); err != nil {
			return fmt.Errorf("links must be an array of {type, url}")
		}
		for _, l := range links {
			if l.Type == "" || l.URL == "" {
				return fmt.Errorf("each link requires type and url")
			}
		}
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

func FromResume(r repository.Resume) Resume {
	return Resume{
		ResumeInput: ResumeInput{
			Title:            r.Title,
			Skills:           r.Skills,
			ExperienceMonths: r.ExperienceMonths,
			About:            r.About,
			Education:        r.Education,
			Links:            json.RawMessage(r.Links),
			City:             r.City,
			WorkFormat:       r.WorkFormat,
			EmploymentType:   r.EmploymentType,
			SalaryMin:        r.SalaryMin,
			SalaryMax:        r.SalaryMax,
		},
		ID:              r.ID,
		UserID:          r.UserID,
		Source:          r.Source,
		SourceText:      r.SourceText,
		IsActive:        r.IsActive,
		LastConfirmedAt: r.LastConfirmedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}
