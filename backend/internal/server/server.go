package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	swaggerDocs "max-miniapp/backend/docs"
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

	swaggerDocs.SwaggerInfo.Host = "eclipse-sim.ru"
	swaggerDocs.SwaggerInfo.Schemes = []string{"https"}

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
	resumes := repository.NewResumeRepo(pool)
	companies := repository.NewCompanyRepo(pool)
	vacancies := repository.NewVacancyRepo(pool)

	authH := handler.NewAuthHandler(users, s.cfg, s.log)
	meH := handler.NewMeHandler(users, s.log)
	resumeH := handler.NewResumeHandler(resumes, s.log)
	companyH := handler.NewCompanyHandler(companies, vacancies, s.log)
	vacancyH := handler.NewVacancyHandler(companies, vacancies, s.log)
	candidatesH := handler.NewCandidatesHandler(companies, vacancies, users, resumes, s.log)

	s.echo.GET("/swagger/*any", echoSwagger.EchoWrapHandler())
	s.echo.GET("/api/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	api := s.echo.Group("/api/v1")
	api.GET("/health", health.Health)
	api.POST("/auth", authH.Auth)

	private := api.Group("")
	private.Use(middleware.RequireAuth(s.cfg.JWTSecret))
	private.GET("/me", meH.Me)
	private.POST("/me/role", meH.SetRole)
	private.GET("/my/resume", resumeH.Get)
	private.PUT("/my/resume", resumeH.Put)
	private.POST("/my/resume/confirm-activity", resumeH.ConfirmActivity)
	private.POST("/companies", companyH.Create)
	private.GET("/my/companies", companyH.ListMine)
	private.GET("/companies/:id/vacancies", companyH.VacancyList)
	private.POST("/companies/:id/vacancies", companyH.VacancyCreate)
	private.PATCH("/vacancies/:id", vacancyH.Update)
	private.GET("/vacancies/:id/candidates", candidatesH.Candidates)
}

func (s *Server) Start() error {
	return s.echo.Start(s.cfg.Addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
