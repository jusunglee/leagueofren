package web

import (
	"log/slog"
	"net/http"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/web/handlers"
	"github.com/jusunglee/leagueofren/internal/web/middleware"
)

type Config struct {
	AdminPassword string
	APIKey        string
	RiotAPIKey    string
}

type Router struct {
	repo   db.Repository
	log    *slog.Logger
	config Config
}

func NewRouter(repo db.Repository, log *slog.Logger, config Config) *Router {
	return &Router{
		repo:   repo,
		log:    log,
		config: config,
	}
}

func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	translationHandler := handlers.NewTranslationHandler(r.repo, r.log)
	voteHandler := handlers.NewVoteHandler(r.repo, r.log)
	feedbackHandler := handlers.NewFeedbackHandler(r.repo, r.log)

	rateLimiter := middleware.NewRateLimiter(30, 60)

	mux.Handle("GET /api/v1/translations",
		middleware.Chain(
			http.HandlerFunc(translationHandler.List),
			middleware.RequestLogger(r.log),
		),
	)

	mux.Handle("GET /api/v1/translations/{id}",
		middleware.Chain(
			http.HandlerFunc(translationHandler.Get),
			middleware.RequestLogger(r.log),
		),
	)

	mux.Handle("POST /api/v1/translations",
		middleware.Chain(
			http.HandlerFunc(translationHandler.Create),
			middleware.RequestLogger(r.log),
			middleware.APIKeyAuth(r.config.APIKey),
		),
	)

	mux.Handle("POST /api/v1/translations/{id}/vote",
		middleware.Chain(
			http.HandlerFunc(voteHandler.Vote),
			middleware.RequestLogger(r.log),
			middleware.RateLimit(rateLimiter),
		),
	)

	mux.Handle("POST /api/v1/translations/{id}/feedback",
		middleware.Chain(
			http.HandlerFunc(feedbackHandler.Create),
			middleware.RequestLogger(r.log),
			middleware.RateLimit(rateLimiter),
		),
	)

	mux.Handle("GET /admin/feedback",
		middleware.Chain(
			http.HandlerFunc(feedbackHandler.List),
			middleware.RequestLogger(r.log),
			middleware.BasicAuth(r.config.AdminPassword),
		),
	)

	return middleware.CORS(mux)
}
