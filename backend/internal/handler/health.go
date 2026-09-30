package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
)

type HealthHandler struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewHealthHandler(pool *pgxpool.Pool, log *slog.Logger) *HealthHandler {
	return &HealthHandler{pool: pool, log: log}
}

// Health — живость сервиса и БД.
//
//	@Summary     Здоровье сервиса
//	@Tags        system
//	@Produce     json
//	@Success     200 {object} dto.HealthResponse
//	@Failure     503 {object} dto.HealthResponse
//	@Router      /api/v1/health [get]
func (h *HealthHandler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.pool.Ping(ctx); err != nil {
		h.log.Error("health check: db ping failed", slog.Any("err", err))
		dbStatus = "unavailable"
	}

	status := http.StatusOK
	if dbStatus != "ok" {
		status = http.StatusServiceUnavailable
	}
	return c.JSON(status, dto.HealthResponse{
		Status: "ok",
		DB:     dbStatus,
	})
}
