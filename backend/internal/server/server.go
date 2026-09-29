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
	"max-miniapp/backend/internal/scoring"
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
	matchingRepo := repository.NewMatchingRepo(pool)
	scoreRepo := repository.NewScoreRepo(pool)
	aiClient := scoring.NewAIClient(s.cfg.AIAPIKey, s.cfg.AIBaseURL, s.cfg.AIModel)
	if aiClient.Enabled() {
		s.log.Info("ai scoring enabled", slog.String("base_url", s.cfg.AIBaseURL), slog.String("model", s.cfg.AIModel))
	}

	authH := handler.NewAuthHandler(users, s.cfg, s.log)
	meH := handler.NewMeHandler(users, companies, s.log)
	debugH := handler.NewDebugHandler(s.log)
	resumeH := handler.NewResumeHandler(resumes, aiClient, s.log)
	companyH := handler.NewCompanyHandler(companies, vacancies, s.log)
	vacancyH := handler.NewVacancyHandler(companies, vacancies, s.log)
	matchingH := handler.NewMatchingHandler(companies, vacancies, users, resumes, matchingRepo, scoreRepo, aiClient, s.cfg.AIModel, s.log)

	s.echo.GET("/swagger/*any", echoSwagger.EchoWrapHandler())
	s.echo.GET("/api/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	api := s.echo.Group("/api/v1")
	api.GET("/health", health.Health)
	api.POST("/auth", authH.Auth)
	api.POST("/auth/demo", authH.DemoAuth)

	private := api.Group("")
	private.Use(middleware.RequireAuth(s.cfg.JWTSecret))
	private.GET("/me", meH.Me)
	private.POST("/me/role", meH.SetRole)
	private.GET("/my/resume", resumeH.Get)
	private.PUT("/my/resume", resumeH.Put)
	private.POST("/my/resume/parse", resumeH.Parse)
	private.POST("/my/resume/confirm-activity", resumeH.ConfirmActivity)
	private.POST("/companies", companyH.Create)
	private.GET("/my/companies", companyH.ListMine)
	private.POST("/companies/:id/verify", companyH.Verify)
	private.GET("/vacancies/map", vacancyH.Map)
	private.GET("/my/referrals", meH.Referrals)
	private.GET("/my/company-referrals", meH.CompanyReferrals)
	private.POST("/debug/log", debugH.Log)
	private.GET("/companies/:id/vacancies", companyH.VacancyList)
	private.POST("/companies/:id/vacancies", companyH.VacancyCreate)
	private.PATCH("/vacancies/:id", vacancyH.Update)
	private.GET("/vacancies/:id/candidates", matchingH.Candidates)
	private.POST("/vacancies/:id/candidates/:candidateUserId/action", matchingH.Action)
	private.GET("/my/invitations", matchingH.Invitations)
	private.POST("/invitations/:id/respond", matchingH.Respond)
}

func (s *Server) Start() error {
	return s.echo.Start(s.cfg.Addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
