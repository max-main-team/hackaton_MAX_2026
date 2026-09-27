package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewHealthHandler(pool *pgxpool.Pool, log *slog.Logger) *HealthHandler {
	return &HealthHandler{pool: pool, log: log}
}

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
	return c.JSON(status, map[string]string{
		"status": "ok",
		"db":     dbStatus,
	})
}
