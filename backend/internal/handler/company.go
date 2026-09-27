package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
)

type CompanyHandler struct {
	companies *repository.CompanyRepo
	vacancies *repository.VacancyRepo
	log       *slog.Logger
}

func NewCompanyHandler(companies *repository.CompanyRepo, vacancies *repository.VacancyRepo, log *slog.Logger) *CompanyHandler {
	return &CompanyHandler{companies: companies, vacancies: vacancies, log: log}
}

// Create создаёт компанию и добавляет создателя участником с выбранной позицией.
//
//	@Summary     Создать компанию
//	@Tags        company
//	@Accept      json
//	@Produce     json
//	@Param       request body dto.CompanyInput true "компания + позиция создателя"
//	@Success     200 {object} dto.Company
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/companies [post]
func (h *CompanyHandler) Create(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var in dto.CompanyInput
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := dto.ValidateCompanyInput(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	saved, err := h.companies.Create(c.Request().Context(), repository.Company{
		Name:        in.Name,
		Description: in.Description,
		Website:     in.Website,
		LogoURL:     in.LogoURL,
		Address:     in.Address,
	}, userID, in.Position)
	if err != nil {
		h.log.Error("create company failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.FromCompany(saved))
}

// ListMine возвращает компании текущего пользователя.
//
//	@Summary     Мои компании
//	@Tags        company
//	@Produce     json
//	@Success     200 {array} dto.Company
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/my/companies [get]
func (h *CompanyHandler) ListMine(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	list, err := h.companies.ListByUser(c.Request().Context(), userID)
	if err != nil {
		h.log.Error("list companies failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if list == nil {
		list = []repository.Company{}
	}

	out := make([]dto.Company, 0, len(list))
	for _, c := range list {
		out = append(out, dto.FromCompany(c))
	}
	return c.JSON(http.StatusOK, out)
}

// VacancyCreate создаёт вакансию в компании (только участник компании).
//
//	@Summary     Создать вакансию
//	@Tags        vacancy
//	@Accept      json
//	@Produce     json
//	@Param       id     path integer true "ID компании"
//	@Param       request body dto.VacancyInput true "вакансия"
//	@Success     200 {object} dto.Vacancy
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     403 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/companies/{id}/vacancies [post]
func (h *CompanyHandler) VacancyCreate(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid company id")
	}

	isMember, err := h.companies.IsMember(c.Request().Context(), companyID, userID)
	if err != nil {
		h.log.Error("check membership failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "not a member of this company")
	}

	var in dto.VacancyInput
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := dto.ValidateVacancyInput(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	saved, err := h.vacancies.Create(c.Request().Context(), repository.Vacancy{
		CompanyID:           companyID,
		Title:               in.Title,
		Description:         in.Description,
		RequiredSkills:      in.RequiredSkills,
		MinExperienceMonths: in.MinExperienceMonths,
		City:                in.City,
		WorkFormat:          in.WorkFormat,
		EmploymentType:      in.EmploymentType,
		SalaryMin:           in.SalaryMin,
		SalaryMax:           in.SalaryMax,
		ResponseTTLHours:    in.ResponseTTLHours,
		IsActive:            true,
	})
	if err != nil {
		h.log.Error("create vacancy failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.FromVacancy(saved))
}

// VacancyList возвращает вакансии компании (только участник компании).
//
//	@Summary     Вакансии компании
//	@Tags        vacancy
//	@Produce     json
//	@Param       id path integer true "ID компании"
//	@Success     200 {array} dto.Vacancy
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     403 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/companies/{id}/vacancies [get]
func (h *CompanyHandler) VacancyList(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid company id")
	}

	isMember, err := h.companies.IsMember(c.Request().Context(), companyID, userID)
	if err != nil {
		h.log.Error("check membership failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "not a member of this company")
	}

	list, err := h.vacancies.ListByCompany(c.Request().Context(), companyID)
	if err != nil {
		h.log.Error("list vacancies failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if list == nil {
		list = []repository.Vacancy{}
	}

	out := make([]dto.Vacancy, 0, len(list))
	for _, v := range list {
		out = append(out, dto.FromVacancy(v))
	}
	return c.JSON(http.StatusOK, out)
}

// Verify верифицирует компанию по токену её бота в MAX.
//
//	@Summary     Верифицировать компанию
//	@Description Проверяет токен бота через GET /me платформы MAX и помечает компанию верифицированной.
//	@Tags        company
//	@Accept      json
//	@Produce     json
//	@Param       id     path integer true "ID компании"
//	@Param       request body dto.VerifyRequest true "токен бота"
//	@Success     200 {object} dto.Company
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     403 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/companies/{id}/verify [post]
func (h *CompanyHandler) Verify(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid company id")
	}

	isMember, err := h.companies.IsMember(c.Request().Context(), companyID, userID)
	if err != nil {
		h.log.Error("check membership failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "not a member of this company")
	}

	var in dto.VerifyRequest
	if err := c.Bind(&in); err != nil || in.BotToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bot_token is required")
	}

	req, err := http.NewRequest(http.MethodGet, "https://platform-api2.max.ru/me", nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	req.Header.Set("Authorization", in.BotToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.log.Error("max me request failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusBadRequest, "verification failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bot token")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "verification failed")
	}

	var botInfo struct {
		UserID    int64   `json:"user_id"`
		FirstName string  `json:"first_name"`
		Username  *string `json:"username"`
		IsBot     bool    `json:"is_bot"`
	}
	if err := json.Unmarshal(body, &botInfo); err != nil || !botInfo.IsBot {
		return echo.NewHTTPError(http.StatusBadRequest, "token is not a bot token")
	}

	username := ""
	if botInfo.Username != nil {
		username = *botInfo.Username
	}

	saved, err := h.companies.MarkVerified(c.Request().Context(), companyID, botInfo.UserID, username)
	if err != nil {
		h.log.Error("mark verified failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.FromCompany(saved))
}
