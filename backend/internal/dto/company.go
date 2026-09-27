package dto

import (
	"fmt"
	"time"

	"max-miniapp/backend/internal/repository"
)

type CompanyInput struct {
	Name        string `json:"name" example:"Ромашка"`
	Description string `json:"description" example:"ИТ-компания"`
	Website     string `json:"website" example:"https://romashka.ru"`
	LogoURL     string `json:"logo_url"`
	Address     string `json:"address" example:"Москва, ул. Тверская, 1"`
	Position    string `json:"position" enums:"owner,hr,employee" example:"owner"`
}

type Company struct {
	CompanyInput
	ID        int64     `json:"id"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var Positions = []string{"owner", "hr", "employee"}

func ValidateCompanyInput(in *CompanyInput) error {
	if in.Name == "" {
		return fmt.Errorf("name is required")
	}
	if in.Position == "" {
		in.Position = "owner"
	}
	if !contains(Positions, in.Position) {
		return fmt.Errorf("position must be one of: %v", Positions)
	}
	return nil
}

func FromCompany(c repository.Company) Company {
	return Company{
		CompanyInput: CompanyInput{
			Name:        c.Name,
			Description: c.Description,
			Website:     c.Website,
			LogoURL:     c.LogoURL,
			Address:     c.Address,
		},
		ID:        c.ID,
		Verified:  c.Verified,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type VacancyInput struct {
	Title               string   `json:"title" example:"Go-разработчик"`
	Description         string   `json:"description" example:"Платёжный сервис"`
	RequiredSkills      string   `json:"required_skills" example:"go, postgres"`
	MinExperienceMonths int      `json:"min_experience_months" example:"12"`
	City                string   `json:"city" example:"Москва"`
	WorkFormat          string   `json:"work_format" enums:"onsite,hybrid,remote" example:"hybrid"`
	EmploymentType      string   `json:"employment_type" enums:"full_time,part_time,contract,internship" example:"full_time"`
	SalaryMin           *int     `json:"salary_min" example:"180000"`
	SalaryMax           *int     `json:"salary_max" example:"280000"`
	ResponseTTLHours    int      `json:"response_ttl_hours" example:"48"`
	Lat                 *float64 `json:"lat"`
	Lng                 *float64 `json:"lng"`
}

type Vacancy struct {
	VacancyInput
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VacancyUpdate struct {
	ResponseTTLHours    *int  `json:"response_ttl_hours" example:"72"`
	MinExperienceMonths *int  `json:"min_experience_months"`
	IsActive            *bool `json:"is_active"`
}

func ValidateVacancyInput(in *VacancyInput) error {
	if in.Title == "" {
		return fmt.Errorf("title is required")
	}
	if !contains(WorkFormats, in.WorkFormat) {
		return fmt.Errorf("work_format must be one of: %v", WorkFormats)
	}
	if !contains(EmploymentTypes, in.EmploymentType) {
		return fmt.Errorf("employment_type must be one of: %v", EmploymentTypes)
	}
	if in.ResponseTTLHours == 0 {
		in.ResponseTTLHours = 48
	}
	if in.ResponseTTLHours < 1 || in.ResponseTTLHours > 336 {
		return fmt.Errorf("response_ttl_hours must be between 1 and 336")
	}
	return nil
}

func FromVacancy(v repository.Vacancy) Vacancy {
	return Vacancy{
		VacancyInput: VacancyInput{
			Title:               v.Title,
			Description:         v.Description,
			RequiredSkills:      v.RequiredSkills,
			MinExperienceMonths: v.MinExperienceMonths,
			City:                v.City,
			WorkFormat:          v.WorkFormat,
			EmploymentType:      v.EmploymentType,
			SalaryMin:           v.SalaryMin,
			SalaryMax:           v.SalaryMax,
			ResponseTTLHours:    v.ResponseTTLHours,
			Lat:                 v.Lat,
			Lng:                 v.Lng,
		},
		ID:        v.ID,
		CompanyID: v.CompanyID,
		IsActive:  v.IsActive,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
