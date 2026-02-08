package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Web server metrics.
var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lor_http_requests_total",
		Help: "Total HTTP requests by route, method, and status code",
	}, []string{"route", "method", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lor_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"route", "method"})

	RateLimitHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lor_rate_limit_hits_total",
		Help: "Total rate limit rejections",
	})

	TranslationSubmissions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lor_translation_submissions_total",
		Help: "Translation submissions by result",
	}, []string{"result"})

	VotesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lor_votes_total",
		Help: "Votes cast by direction",
	}, []string{"direction"})

	LLMTranslationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "lor_llm_translation_duration_seconds",
		Help:    "LLM translation call duration in seconds",
		Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30},
	})
)

// Worker metrics.
var (
	RefreshCycleDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "lor_worker_refresh_duration_seconds",
		Help:    "Duration of each worker refresh cycle",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
	})

	PlayersProcessed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lor_worker_players_processed",
		Help: "Number of players processed in last refresh cycle",
	})

	RiotAPICallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lor_riot_api_calls_total",
		Help: "Riot API calls by endpoint and result",
	}, []string{"endpoint", "result"})

	RiotAPILatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lor_riot_api_duration_seconds",
		Help:    "Riot API call duration in seconds",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	}, []string{"endpoint"})

	PuuidBackfillTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lor_worker_puuid_backfill_total",
		Help: "PUUID backfill attempts by result",
	}, []string{"result"})
)

// Database pool metrics (gauges updated periodically).
var (
	DBPoolTotalConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lor_db_pool_total_conns",
		Help: "Total number of connections in the pool",
	})

	DBPoolIdleConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lor_db_pool_idle_conns",
		Help: "Number of idle connections in the pool",
	})

	DBPoolAcquiredConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lor_db_pool_acquired_conns",
		Help: "Number of acquired connections in the pool",
	})

	DBPoolMaxConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lor_db_pool_max_conns",
		Help: "Max connections configured for the pool",
	})
)
