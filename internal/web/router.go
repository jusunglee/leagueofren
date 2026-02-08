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

type Router struct {
	repo       db.Repository
	log        *slog.Logger
	riot       *riot.DirectClient
	translator *translation.Translator
}

func NewRouter(repo db.Repository, log *slog.Logger, riotClient *riot.DirectClient, translator *translation.Translator) *Router {
	return &Router{
		repo:       repo,
		log:        log,
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
			middleware.PrometheusMetrics(),
			middleware.RequestLogger(r.log),
			middleware.CacheControl("public, s-maxage=5, max-age=0"),
		),
	)

	mux.Handle("GET /api/v1/translations/{id}",
		middleware.Chain(
			http.HandlerFunc(translationHandler.Get),
			middleware.PrometheusMetrics(),
			middleware.RequestLogger(r.log),
			middleware.CacheControl("public, s-maxage=5, max-age=0"),
		),
	)

	mux.Handle("POST /api/v1/translations",
		middleware.Chain(
			http.HandlerFunc(translationHandler.Create),
			middleware.PrometheusMetrics(),
			middleware.RequestLogger(r.log),
			middleware.RateLimit(rateLimiter),
		),
	)

	mux.Handle("POST /api/v1/translations/{id}/vote",
		middleware.Chain(
			http.HandlerFunc(voteHandler.Vote),
			middleware.PrometheusMetrics(),
			middleware.RequestLogger(r.log),
			middleware.RateLimit(rateLimiter),
		),
	)

	mux.Handle("POST /api/v1/translations/{id}/feedback",
		middleware.Chain(
			http.HandlerFunc(feedbackHandler.Create),
			middleware.PrometheusMetrics(),
			middleware.RequestLogger(r.log),
			middleware.RateLimit(rateLimiter),
		),
	)

	return middleware.CORS(mux)
}
