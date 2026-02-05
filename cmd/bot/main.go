package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/anthropic"
	"github.com/jusunglee/leagueofren/internal/bot"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/db/postgres"
	"github.com/jusunglee/leagueofren/internal/db/sqlite"
	"github.com/jusunglee/leagueofren/internal/google"
	"github.com/jusunglee/leagueofren/internal/llm"
	"github.com/jusunglee/leagueofren/internal/logger"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/setup"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := mainE(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
	slog.Info("exiting without error")
}

func mainE() error {
	if setup.NeedsSetup() {
		fmt.Println("No .env file found. Starting setup wizard...")
		completed, err := setup.Run()
		if err != nil {
			return fmt.Errorf("setup wizard failed: %w", err)
		}
		if !completed {
			return errors.New("setup cancelled")
		}
		fmt.Println("\nConfiguration saved! Starting bot...")
	}

	_ = godotenv.Load()

	fs := ff.NewFlagSet("leagueofren")

	var (
		databaseURL                  = fs.StringLong("database-url", "", "PostgreSQL connection URL")
		discordToken                 = fs.StringLong("discord-token", "", "Discord bot token")
		riotAPIKey                   = fs.StringLong("riot-api-key", "", "Riot Games API key")
		llmProvider                  = fs.StringEnumLong("llm-provider", "LLM provider", "anthropic", "google")
		llmModel                     = fs.StringLong("llm-model", "", "LLM model name")
		guildID                      = fs.StringLong("guild-id", "", "Discord guild ID for command registration")
		anthropicAPIKey              = fs.StringLong("anthropic-api-key", "", "Anthropic API key")
		googleAPIKey                 = fs.StringLong("google-api-key", "", "Google API key")
		maxSubscriptionsPerServer    = fs.Int64Long("max-subscriptions-per-server", 10, "Maximum subscriptions per Discord server")
		evaluateSubscriptionsTimeout = fs.DurationLong("evaluate-subscriptions-timeout", 1*time.Minute, "Timeout for evaluating subscriptions")
		evalExpirationDuration       = fs.DurationLong("eval-expiration-duration", 504*time.Hour, "Duration before evals expire (default 3 weeks)")
		offlineActivityThreshold     = fs.DurationLong("offline-activity-threshold", 168*time.Hour, "Duration of inactivity before auto-unsubscribe (default 1 week)")
		numConsumers                 = fs.Int64Long("num-consumers", 2, "Number of consumer goroutines")
	)

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVars()); err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs))
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *databaseURL == "" {
		return errors.New("database-url is required")
	}
	if *discordToken == "" {
		return errors.New("discord-token is required")
	}
	if *riotAPIKey == "" {
		return errors.New("riot-api-key is required")
	}
	if *llmModel == "" {
		return errors.New("llm-model is required")
	}

	var client llm.Client
	switch *llmProvider {
	case "anthropic":
		if *anthropicAPIKey == "" {
			return errors.New("anthropic-api-key is required when using anthropic provider")
		}
		client = anthropic.NewClient(*anthropicAPIKey, anthropic.Model(*llmModel))
	case "google":
		if *googleAPIKey == "" {
			return errors.New("google-api-key is required when using google provider")
		}
		var err error
		client, err = google.NewClient(context.Background(), *googleAPIKey, google.Model(*llmModel))
		if err != nil {
			return fmt.Errorf("creating Google client: %w", err)
		}
	}

	dg, err := discordgo.New("Bot " + *discordToken)
	if err != nil {
		return fmt.Errorf("creating Discord session: %w", err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())

	log := logger.New()

	var repo db.Repository
	if isSQLite(*databaseURL) {
		sqliteRepo, err := sqlite.New(ctx, *databaseURL)
		if err != nil {
			return fmt.Errorf("creating SQLite database: %w", err)
		}
		repo = sqliteRepo
		log.InfoContext(ctx, "connected to SQLite database", "path", *databaseURL)
	} else {
		pgRepo, err := postgres.New(ctx, *databaseURL)
		if err != nil {
			return fmt.Errorf("creating PostgreSQL connection: %w", err)
		}
		repo = pgRepo
		log.InfoContext(ctx, "connected to PostgreSQL database")
	}
	defer repo.Close()

	translator := translation.NewTranslator(client, repo, *llmProvider, *llmModel)
	riotClient := riot.NewCachedClient(*riotAPIKey, repo)
	log.InfoContext(ctx, "riot API client initialized with caching")

	b := bot.New(log, dg, repo, riotClient, translator, bot.Config{
		MaxSubscriptionsPerServer:    *maxSubscriptionsPerServer,
		EvaluateSubscriptionsTimeout: *evaluateSubscriptionsTimeout,
		EvalExpirationDuration:       *evalExpirationDuration,
		OfflineActivityThreshold:     *offlineActivityThreshold,
		NumConsumers:                 *numConsumers,
		GuildID:                      *guildID,
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("received signal, shutting down gracefully", "signal", sig, slog.Duration("timeout", time.Minute))
		cancel(errors.New("signal received"))

		sig = <-sigChan
		slog.Warn("received second signal, forcing exit", "signal", sig)
		os.Exit(1)
	}()

	return b.Run(ctx, cancel)
}

func isSQLite(url string) bool {
	if strings.HasPrefix(url, "sqlite://") {
		return true
	}
	if strings.HasSuffix(url, ".db") || strings.HasSuffix(url, ".sqlite") || strings.HasSuffix(url, ".sqlite3") {
		return true
	}
	return false
}
