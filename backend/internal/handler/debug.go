package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type DebugHandler struct {
	log *slog.Logger
}

func NewDebugHandler(log *slog.Logger) *DebugHandler {
	return &DebugHandler{log: log}
}

type DebugLog struct {
	Where  string `json:"where"`
	Detail string `json:"detail"`
}

// Log — временная точка телеметрии клиентских ошибок (отладка вебвью MAX).
//
//	@Summary     Клиентская ошибка в лог
//	@Tags        debug
//	@Accept      json
//	@Produce     json
//	@Param       request body DebugLog true "куда и что упало"
//	@Success     200 {object} map[string]string
//	@Security    BearerAuth
//	@Router      /api/v1/debug/log [post]
func (h *DebugHandler) Log(c echo.Context) error {
	var in DebugLog
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	h.log.Warn("client error report", slog.String("where", in.Where), slog.String("detail", in.Detail))
	return c.JSON(http.StatusOK, map[string]string{"ok": "1"})
}
