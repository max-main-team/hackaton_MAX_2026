package dto

import "time"

type MapVacancy struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	CompanyName string  `json:"company_name"`
	Verified    bool    `json:"verified"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	SalaryMin   *int    `json:"salary_min"`
	SalaryMax   *int    `json:"salary_max"`
}

type ReferralsResponse struct {
	Count int            `json:"count"`
	Items []ReferralItem `json:"items"`
}

type ReferralItem struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	JoinedAt  time.Time `json:"joined_at"`
}

type InvitedCompany struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type VerifyRequest struct {
	BotToken string `json:"bot_token"`
}

type ParseResumeRequest struct {
	Text       string `json:"text" example:"Иван Петров, Go-разработчик, 4 года опыта..."`
	SourceName string `json:"source_name" example:"resume.pdf"`
}

type ParseResumeResponse struct {
	Resume    Resume `json:"resume"`
	AIComment string `json:"ai_comment" example:"Извлекли должность, навыки, опыт; зарплату не нашли"`
}
