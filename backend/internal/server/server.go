package server

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"max-miniapp/backend/internal/config"
	"max-miniapp/backend/internal/handler"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
)

type Server struct {
	echo *echo.Echo
	cfg  *config.Config
	log  *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *Server {
	s := &Server{cfg: cfg, log: log}
	s.echo = echo.New()
	s.echo.HideBanner = true
	s.echo.HidePort = true

	if cfg.MaxBotToken == "" {
		log.Warn("MAX_BOT_TOKEN is empty: initData signature verification is disabled (dev mode only)")
	}

	s.setupMiddleware()
	s.setupRoutes(pool)
	return s
}

func (s *Server) setupMiddleware() {
	s.echo.Use(middleware.SlogLogger(s.log))
	s.echo.Use(echomw.Recover())
	s.echo.Use(echomw.CORS())
}

func (s *Server) setupRoutes(pool *pgxpool.Pool) {
	health := handler.NewHealthHandler(pool, s.log)
	users := repository.NewUserRepo(pool)
	auth := handler.NewAuthHandler(users, s.cfg, s.log)

	api := s.echo.Group("/api/v1")
	api.GET("/health", health.Health)
	api.POST("/auth", auth.Auth)

	private := api.Group("")
	private.Use(middleware.RequireAuth(s.cfg.JWTSecret))
}

func (s *Server) Start() error {
	return s.echo.Start(s.cfg.Addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
