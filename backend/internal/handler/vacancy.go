package handler

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
)

type VacancyHandler struct {
	companies *repository.CompanyRepo
	vacancies *repository.VacancyRepo
	log       *slog.Logger
}

func NewVacancyHandler(companies *repository.CompanyRepo, vacancies *repository.VacancyRepo, log *slog.Logger) *VacancyHandler {
	return &VacancyHandler{companies: companies, vacancies: vacancies, log: log}
}

// @Summary     Обновить вакансию
// @Tags        vacancy
// @Accept      json
// @Produce     json
// @Param       id     path integer true "ID вакансии"
// @Param       request body dto.VacancyUpdate true "изменяемые поля"
// @Success     200 {object} dto.Vacancy
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/vacancies/{id} [patch]
func (h *VacancyHandler) Update(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	vacancyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid vacancy id")
	}

	vacancy, err := h.vacancies.GetByID(c.Request().Context(), vacancyID)
	if err != nil {
		h.log.Error("get vacancy failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusNotFound, "vacancy not found")
	}

	isMember, err := h.companies.IsMember(c.Request().Context(), vacancy.CompanyID, userID)
	if err != nil {
		h.log.Error("check membership failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "not a member of this company")
	}

	var in dto.VacancyUpdate
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if in.ResponseTTLHours != nil && (*in.ResponseTTLHours < 1 || *in.ResponseTTLHours > 336) {
		return echo.NewHTTPError(http.StatusBadRequest, "response_ttl_hours must be between 1 and 336")
	}

	if in.WorkFormat != nil && !slices.Contains(dto.WorkFormats, *in.WorkFormat) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid work_format")
	}
	if in.EmploymentType != nil && !slices.Contains(dto.EmploymentTypes, *in.EmploymentType) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid employment_type")
	}

	saved, err := h.vacancies.Update(c.Request().Context(), vacancyID, repository.VacancyUpdate{
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
		IsActive:            in.IsActive,
	})
	if err != nil {
		h.log.Error("update vacancy failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.FromVacancy(saved))
}

// @Summary     Вакансии на карте
// @Tags        vacancy
// @Produce     json
// @Success     200 {array} dto.MapVacancy
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/vacancies/map [get]
func (h *VacancyHandler) Map(c echo.Context) error {
	items, err := h.vacancies.MapVacancies(c.Request().Context())
	if err != nil {
		h.log.Error("map vacancies failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if items == nil {
		items = []repository.MapVacancy{}
	}

	out := make([]dto.MapVacancy, 0, len(items))
	for _, m := range items {
		out = append(out, dto.MapVacancy{
			ID: m.ID, Title: m.Title, CompanyName: m.CompanyName,
			Verified: m.Verified, City: m.City, Lat: m.Lat, Lng: m.Lng,
			SalaryMin: m.SalaryMin, SalaryMax: m.SalaryMax,
		})
	}
	return c.JSON(http.StatusOK, out)
}
