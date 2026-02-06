package web

import (
	"log/slog"
	"net/http"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/jusunglee/leagueofren/internal/web/handlers"
	"github.com/jusunglee/leagueofren/internal/web/middleware"
)

type Config struct {
	AdminPassword string
}

type Router struct {
	repo       db.Repository
	log        *slog.Logger
	config     Config
	riot       *riot.DirectClient
	translator *translation.Translator
}

func NewRouter(repo db.Repository, log *slog.Logger, config Config, riotClient *riot.DirectClient, translator *translation.Translator) *Router {
	return &Router{
		repo:       repo,
		log:        log,
		config:     config,
		riot:       riotClient,
		translator: translator,
	}
}

func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	translationHandler := handlers.NewTranslationHandler(r.repo, r.log, r.riot, r.translator)
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
			middleware.RateLimit(rateLimiter),
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
