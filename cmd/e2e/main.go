package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/anthropic"
	"github.com/jusunglee/leagueofren/internal/bot"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/db/sqlite"
	"github.com/jusunglee/leagueofren/internal/google"
	"github.com/jusunglee/leagueofren/internal/llm"
	"github.com/jusunglee/leagueofren/internal/logger"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
)

func main() {
	if err := run(); err != nil {
		slog.Error("E2E FAILED", "error", err)
		os.Exit(1)
	}
	slog.Info("E2E PASSED")
}

func run() error {
	_ = godotenv.Load()

	// Required env vars
	discordToken := requireEnv("DISCORD_TOKEN")
	riotAPIKey := requireEnv("RIOT_API_KEY")
	llmProvider := requireEnv("LLM_PROVIDER")
	llmModel := requireEnv("LLM_MODEL")
	channelID := requireEnv("E2E_DISCORD_CHANNEL_ID")
	guildID := requireEnv("E2E_DISCORD_GUILD_ID")

	log := logger.New()
	ctx := context.Background()

	// Phase 1: Set up DB + Bot first (so we can act immediately when we find a live game)
	log.Info("Phase 1: Setting up DB and Bot...")
	dbPath := fmt.Sprintf("/tmp/leagueofren-e2e-%d.db", time.Now().UnixNano())
	defer os.Remove(dbPath)

	repo, err := sqlite.New(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("creating temp SQLite: %w", err)
	}
	defer repo.Close()

	var llmClient llm.Client
	switch llmProvider {
	case "anthropic":
		apiKey := requireEnv("ANTHROPIC_API_KEY")
		llmClient = anthropic.NewClient(apiKey, anthropic.Model(llmModel))
	case "google":
		apiKey := requireEnv("GOOGLE_API_KEY")
		llmClient, err = google.NewClient(ctx, apiKey, google.Model(llmModel))
		if err != nil {
			return fmt.Errorf("creating Google client: %w", err)
		}
	default:
		return fmt.Errorf("unsupported LLM_PROVIDER: %s", llmProvider)
	}

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return fmt.Errorf("creating Discord session: %w", err)
	}

	translator := translation.NewTranslator(llmClient, repo, llmProvider, llmModel)
	riotClient := riot.NewCachedClient(riotAPIKey, repo)
	directClient := riot.NewDirectClient(riotAPIKey)

	discordSession := bot.NewDiscordSession(dg)
	b := bot.New(
		bot.NewLogger(log),
		discordSession,
		bot.NewMessageServer(discordSession),
		repo,
		bot.NewRiotClient(riotClient),
		bot.NewTranslator(translator),
		bot.Config{
			MaxSubscriptionsPerServer:    100,
			EvaluateSubscriptionsTimeout: 2 * time.Minute,
			EvalExpirationDuration:       24 * time.Hour,
			OfflineActivityThreshold:     24 * time.Hour,
			TranslationRetentionDuration: 24 * time.Hour,
			FeedbackRetentionDuration:    24 * time.Hour,
			NumConsumers:                 1,
			GuildID:                      guildID,
			JobBufferSize:                5,
		},
	)

	// Phase 2: Find a live KR challenger player and pre-seed the cache
	log.Info("Phase 2: Finding live KR challenger player...")

	league, err := directClient.GetChallengerLeague("KR")
	if err != nil {
		return fmt.Errorf("getting challenger league: %w", err)
	}
	log.Info("fetched challenger ladder", "entries", len(league.Entries))

	var targetRiotID string
	var targetPUUID string
	var foundGame riot.ActiveGame
	for i, entry := range league.Entries {
		if entry.Puuid == "" {
			continue
		}
		game, err := directClient.GetActiveGame(entry.Puuid, "KR")
		if errors.Is(err, riot.ErrNotInGame) {
			continue
		}
		if err != nil {
			log.Warn("error checking active game", "index", i, "error", err)
			continue
		}

		// Find a participant with foreign characters
		for _, p := range game.Participants {
			if containsForeignCharacters(p.GameName) {
				targetRiotID = p.GameName
				targetPUUID = p.PUUID
				break
			}
		}
		if targetRiotID == "" {
			log.Info("game has no foreign-character names, skipping", "index", i, "game_id", game.GameID)
			continue
		}

		foundGame = game
		log.Info("found live game", "index", i, "game_id", game.GameID, "target", targetRiotID, "target_puuid", targetPUUID, "participants", len(game.Participants))
		break
	}

	if targetRiotID == "" {
		return fmt.Errorf("no live KR challenger games with foreign-character names found")
	}

	// Pre-seed the game cache for the TARGET player's PUUID so RunOnce
	// doesn't re-fetch from the slow KR API
	participantsJSON, err := json.Marshal(foundGame.Participants)
	if err != nil {
		return fmt.Errorf("marshalling participants: %w", err)
	}
	if err := repo.CacheGameStatus(ctx, db.CacheGameStatusParams{
		Puuid:        targetPUUID,
		Region:       "KR",
		InGame:       true,
		GameID:       sql.NullInt64{Int64: foundGame.GameID, Valid: true},
		Participants: participantsJSON,
	}); err != nil {
		return fmt.Errorf("pre-seeding game cache: %w", err)
	}
	log.Info("pre-seeded game cache", "puuid", targetPUUID)

	if targetRiotID == "" {
		return fmt.Errorf("no live KR challenger games with foreign-character names found")
	}

	// Phase 3: Subscribe via Bot
	log.Info("Phase 3: Subscribing...", "riot_id", targetRiotID, "region", "KR")
	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 60*time.Second)
	defer subscribeCancel()

	sub, err := b.Subscribe(subscribeCtx, channelID, targetRiotID, "KR", guildID)
	if err != nil {
		return fmt.Errorf("subscribing to %s: %w", targetRiotID, err)
	}
	log.Info("subscription created", "subscription_id", sub.ID, "username", sub.LolUsername)

	// Phase 4: Run a single produce/consume cycle (no WebSocket needed)
	log.Info("Phase 4: Running single produce/consume cycle...")
	runCtx, runCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer runCancel()

	if err := b.RunOnce(runCtx); err != nil {
		return fmt.Errorf("RunOnce failed: %w", err)
	}

	// Check for eval
	eval, err := repo.GetLatestEvalForSubscription(ctx, sub.ID)
	if err != nil {
		return fmt.Errorf("no eval found after RunOnce: %w", err)
	}
	if eval.EvalStatus != "NEW_TRANSLATIONS" || !eval.DiscordMessageID.Valid {
		return fmt.Errorf("eval has unexpected state: status=%s, has_message_id=%v", eval.EvalStatus, eval.DiscordMessageID.Valid)
	}
	log.Info("eval found!", "eval_id", eval.ID, "status", eval.EvalStatus, "discord_message_id", eval.DiscordMessageID.String)

	// Phase 5: Verify + cleanup
	log.Info("Phase 5: Verifying and cleaning up...")

	// Verify the Discord message exists
	msg, err := dg.ChannelMessage(channelID, eval.DiscordMessageID.String)
	if err != nil {
		return fmt.Errorf("verifying Discord message %s: %w", eval.DiscordMessageID.String, err)
	}
	log.Info("Discord message verified", "message_id", msg.ID, "embeds", len(msg.Embeds))

	// Cleanup: unsubscribe
	cleanupCtx, cleanupCancel := context.WithTimeout(ctx, 30*time.Second)
	defer cleanupCancel()

	if err := b.Unsubscribe(cleanupCtx, channelID, sub.LolUsername, "KR"); err != nil {
		log.Warn("cleanup: failed to unsubscribe", "error", err)
	}

	// Cleanup: delete the test Discord message
	if err := dg.ChannelMessageDelete(channelID, eval.DiscordMessageID.String); err != nil {
		log.Warn("cleanup: failed to delete Discord message", "error", err)
	} else {
		log.Info("cleanup: deleted test Discord message", "message_id", eval.DiscordMessageID.String)
	}

	log.Info("all verifications passed",
		"eval_id", eval.ID,
		"eval_status", eval.EvalStatus,
		"discord_message_id", eval.DiscordMessageID.String,
		"subscription", sub.LolUsername,
	)

	return nil
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		slog.Error("required environment variable not set", "key", key)
		os.Exit(1)
	}
	return val
}

func containsForeignCharacters(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}
