package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/db/postgres"
	"github.com/jusunglee/leagueofren/internal/logger"
	"github.com/jusunglee/leagueofren/internal/metrics"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := mainE(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func mainE() error {
	_ = godotenv.Load()

	fs := ff.NewFlagSet("leagueofren-worker")
	var (
		databaseURL = fs.StringLong("database-url", "", "PostgreSQL connection URL")
		riotAPIKey  = fs.StringLong("riot-api-key", "", "Riot API key")
		interval    = fs.DurationLong("interval", 1*time.Hour, "Polling interval")
	)

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVars()); err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs))
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *databaseURL == "" {
		return errors.New("database-url is required")
	}
	if *riotAPIKey == "" {
		return errors.New("riot-api-key is required")
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	log := logger.New()

	repo, err := postgres.New(ctx, *databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer repo.Close()

	riotClient := riot.NewDirectClient(*riotAPIKey)

	log.InfoContext(ctx, "fetching champion data from Data Dragon")
	champMap, err := fetchChampionMap()
	if err != nil {
		return fmt.Errorf("fetching champion map: %w", err)
	}
	log.InfoContext(ctx, "loaded champion map", "count", len(champMap))

	// Serve Prometheus metrics on :9090
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsServer := &http.Server{Addr: ":9090", Handler: metricsMux}
		log.InfoContext(ctx, "starting metrics server", "addr", ":9090")
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.ErrorContext(ctx, "metrics server error", "error", err)
		}
	}()

	// Periodically export pgxpool stats as Prometheus gauges
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s := repo.PoolStats()
				metrics.DBPoolTotalConns.Set(float64(s.TotalConns()))
				metrics.DBPoolIdleConns.Set(float64(s.IdleConns()))
				metrics.DBPoolAcquiredConns.Set(float64(s.AcquiredConns()))
				metrics.DBPoolMaxConns.Set(float64(s.MaxConns()))
			case <-ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info("received signal, shutting down", "signal", sig)
		cancel(errors.New("signal received"))
	}()

	log.InfoContext(ctx, "worker starting", "interval", *interval)
	runRefresh(ctx, repo, riotClient, champMap, log)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runRefresh(ctx, repo, riotClient, champMap, log)
		case <-ctx.Done():
			log.Info("worker stopped")
			return nil
		}
	}
}

func runRefresh(ctx context.Context, repo db.Repository, riotClient *riot.DirectClient, champMap map[int64]string, log *slog.Logger) {
	cycleStart := time.Now()
	defer func() {
		metrics.RefreshCycleDuration.Observe(time.Since(cycleStart).Seconds())
	}()

	players, err := repo.ListAllPlayers(ctx)
	if err != nil {
		log.ErrorContext(ctx, "listing players", "error", err)
		return
	}

	metrics.PlayersProcessed.Set(float64(len(players)))
	log.InfoContext(ctx, "starting player refresh", "count", len(players))

	for _, player := range players {
		if ctx.Err() != nil {
			return
		}

		// Backfill puuid if missing
		if !player.Puuid.Valid {
			gameName, tagLine, err := riot.ParseRiotID(player.Username)
			if err != nil || tagLine == "" {
				continue
			}
			apiStart := time.Now()
			account, err := riotClient.GetAccountByRiotID(gameName, tagLine, player.Region)
			metrics.RiotAPILatency.WithLabelValues("account").Observe(time.Since(apiStart).Seconds())
			if err != nil {
				metrics.RiotAPICallsTotal.WithLabelValues("account", "error").Inc()
				metrics.PuuidBackfillTotal.WithLabelValues("error").Inc()
				log.WarnContext(ctx, "puuid lookup failed", "username", player.Username, "error", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			metrics.RiotAPICallsTotal.WithLabelValues("account", "success").Inc()
			player.Puuid = sql.NullString{String: account.PUUID, Valid: true}
			_, err = repo.UpsertPlayer(ctx, db.UpsertPlayerParams{
				Username: player.Username,
				Region:   player.Region,
				Puuid:    player.Puuid,
			})
			if err != nil {
				log.ErrorContext(ctx, "saving puuid", "username", player.Username, "error", err)
				continue
			}
			metrics.PuuidBackfillTotal.WithLabelValues("success").Inc()
			log.InfoContext(ctx, "backfilled puuid", "username", player.Username)
			time.Sleep(100 * time.Millisecond)
		}

		apiStart := time.Now()
		entries, err := riotClient.GetRankedEntries(player.Puuid.String, player.Region)
		metrics.RiotAPILatency.WithLabelValues("ranked").Observe(time.Since(apiStart).Seconds())
		if err != nil {
			metrics.RiotAPICallsTotal.WithLabelValues("ranked", "error").Inc()
			log.WarnContext(ctx, "fetching ranked entries", "username", player.Username, "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		metrics.RiotAPICallsTotal.WithLabelValues("ranked", "success").Inc()

		rank := extractSoloQueueRank(entries)

		time.Sleep(100 * time.Millisecond)

		apiStart = time.Now()
		masteries, err := riotClient.GetTopChampionMastery(player.Puuid.String, player.Region, 3)
		metrics.RiotAPILatency.WithLabelValues("mastery").Observe(time.Since(apiStart).Seconds())
		if err != nil {
			metrics.RiotAPICallsTotal.WithLabelValues("mastery", "error").Inc()
			log.WarnContext(ctx, "fetching champion mastery", "username", player.Username, "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		metrics.RiotAPICallsTotal.WithLabelValues("mastery", "success").Inc()

		champNames := make([]string, 0, len(masteries))
		for _, m := range masteries {
			if name, ok := champMap[m.ChampionID]; ok {
				champNames = append(champNames, name)
			}
		}

		champJSON, _ := json.Marshal(champNames)

		err = repo.UpdatePlayerStats(ctx, db.UpdatePlayerStatsParams{
			Username:     player.Username,
			Rank:         sql.NullString{String: rank, Valid: rank != ""},
			TopChampions: sql.NullString{String: string(champJSON), Valid: len(champNames) > 0},
		})
		if err != nil {
			log.ErrorContext(ctx, "updating player stats", "username", player.Username, "error", err)
			continue
		}

		log.InfoContext(ctx, "updated player", "username", player.Username, "rank", rank, "champions", champNames)
		time.Sleep(100 * time.Millisecond)
	}

	log.InfoContext(ctx, "player refresh complete")
}

func extractSoloQueueRank(entries []riot.LeagueEntry) string {
	for _, e := range entries {
		if e.QueueType == "RANKED_SOLO_5x5" {
			return e.Tier
		}
	}
	return ""
}

type dataDragonResponse struct {
	Data map[string]struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"data"`
}

func fetchChampionMap() (map[int64]string, error) {
	versionsResp, err := http.Get("https://ddragon.leagueoflegends.com/api/versions.json")
	if err != nil {
		return nil, fmt.Errorf("fetching versions: %w", err)
	}
	defer versionsResp.Body.Close()

	var versions []string
	if err := json.NewDecoder(versionsResp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("decoding versions: %w", err)
	}
	if len(versions) == 0 {
		return nil, errors.New("no versions returned")
	}
	latestVersion := versions[0]

	champResp, err := http.Get(fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", latestVersion))
	if err != nil {
		return nil, fmt.Errorf("fetching champion data: %w", err)
	}
	defer champResp.Body.Close()

	var dd dataDragonResponse
	if err := json.NewDecoder(champResp.Body).Decode(&dd); err != nil {
		return nil, fmt.Errorf("decoding champion data: %w", err)
	}

	result := make(map[int64]string, len(dd.Data))
	for _, champ := range dd.Data {
		id, err := strconv.ParseInt(champ.Key, 10, 64)
		if err != nil {
			continue
		}
		result[id] = champ.Name
	}
	return result, nil
}
