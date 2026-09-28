package dto

import "time"

type RecruiterActionResponse struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action" enums:"invite,skip"`
	CreatedAt time.Time `json:"created_at"`
}

type InvitationCompany struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

type InvitationVacancy struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	City       string `json:"city"`
	WorkFormat string `json:"work_format"`
	SalaryMin  *int32 `json:"salary_min"`
	SalaryMax  *int32 `json:"salary_max"`
}

type Invitation struct {
	ID         int64             `json:"id"`
	Status     string            `json:"status" enums:"pending,overdue,resulted"`
	DeadlineAt time.Time         `json:"deadline_at"`
	HoursLeft  float64           `json:"hours_left"`
	Company    InvitationCompany `json:"company"`
	Vacancy    InvitationVacancy `json:"vacancy"`
	Response   *string           `json:"response"`
}

type RespondRequest struct {
	Response string `json:"response" enums:"accept,decline"`
}

type MatchResult struct {
	ID               int64             `json:"id"`
	Response         string            `json:"response"`
	Company          InvitationCompany `json:"company"`
	Vacancy          InvitationVacancy `json:"vacancy"`
	RecruiterContact string            `json:"recruiter_contact"`
}
