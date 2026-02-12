package web

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/web/handlers"
	"github.com/jusunglee/leagueofren/internal/web/middleware"
	"github.com/riverqueue/river"
)

type RateLimitConfig struct {
	Max           int
	WindowSeconds int
	MaxVotesPerIP int
}

type Router struct {
	repo           db.Repository
	log            *slog.Logger
	riot           *riot.DirectClient
	riverClient    *river.Client[pgx.Tx]
	allowedOrigins []string
	rateLimit      RateLimitConfig
}

func NewRouter(repo db.Repository, log *slog.Logger, riotClient *riot.DirectClient, riverClient *river.Client[pgx.Tx], allowedOrigins []string, rateLimit RateLimitConfig) *Router {
	return &Router{
		repo:           repo,
		log:            log,
		riot:           riotClient,
		riverClient:    riverClient,
		allowedOrigins: allowedOrigins,
		rateLimit:      rateLimit,
	}
}

func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	translationHandler := handlers.NewTranslationHandler(r.repo, r.log, r.riot, r.riverClient)
	voteHandler := handlers.NewVoteHandler(r.repo, r.log, r.rateLimit.MaxVotesPerIP)
	feedbackHandler := handlers.NewFeedbackHandler(r.repo, r.log)

	rateLimiter := middleware.NewRateLimiter(r.rateLimit.Max, r.rateLimit.WindowSeconds)

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

	return middleware.CORS(r.allowedOrigins)(mux)
}
