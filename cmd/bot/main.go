package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/anthropic"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/google"
	"github.com/jusunglee/leagueofren/internal/llm"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/samber/lo"
)

type Bot struct {
	queries    *db.Queries
	riotClient *riot.CachedClient
	translator *translation.Translator
}

func buildRegionChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(riot.ValidRegions))
	for i, region := range riot.ValidRegions {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  region,
			Value: region,
		}
	}
	return choices
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "subscribe",
		Description: "Subscribe to League of Legends summoner translations",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "Riot ID (e.g., name#tag)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "region",
				Description: "Server region",
				Required:    true,
				Choices:     buildRegionChoices(),
			},
		},
	},
	{
		Name:        "unsubscribe",
		Description: "Unsubscribe from a summoner",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "Riot ID (e.g., name#tag)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "region",
				Description: "Server region",
				Required:    true,
				Choices:     buildRegionChoices(),
			},
		},
	},
	{
		Name:        "list",
		Description: "List all subscriptions in this channel",
	},
}

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		slog.Error("DISCORD_TOKEN environment variable is required")
		os.Exit(1)
	}

	riotAPIKey := os.Getenv("RIOT_API_KEY")
	if riotAPIKey == "" {
		slog.Error("RIOT_API_KEY environment variable is required")
		os.Exit(1)
	}

	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		slog.Error("LLM_PROVIDER environment variable is required")
		os.Exit(1)
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		slog.Error("LLM_MODEL environment variable is required")
		os.Exit(1)
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		slog.Error("DISCORD_GUILD_ID environment variable is required")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to database")

	var client llm.Client

	switch provider {
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			fmt.Println("ANTHROPIC_API_KEY not set")
			os.Exit(1)
		}
		client = anthropic.NewClient(apiKey, anthropic.Model(model))
	case "google":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			fmt.Println("GOOGLE_API_KEY not set")
			os.Exit(1)
		}
		client, err = google.NewClient(ctx, apiKey, google.Model(model))
		if err != nil {
			fmt.Printf("Error creating Google client: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown provider: %s (use anthropic or google)\n", model)
		os.Exit(1)
	}

	translator := translation.NewTranslator(client)

	queries := db.New(pool)
	bot := &Bot{
		queries:    queries,
		riotClient: riot.NewCachedClient(riotAPIKey, queries),
		translator: translator,
	}
	slog.Info("riot API client initialized with caching")

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		slog.Error("failed to create Discord session", "error", err)
		os.Exit(1)
	}

	dg.AddHandler(bot.handleInteraction)
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info("connected to Discord", "username", r.User.Username, "discriminator", r.User.Discriminator)
	})

	err = dg.Open()
	if err != nil {
		slog.Error("failed to open Discord connection", "error", err)
		os.Exit(1)
	}
	defer dg.Close()

	if guildID != "" {
		slog.Info("registering commands to guild (instant updates)", "guild_id", guildID)
	} else {
		slog.Info("registering commands globally (may take up to 1 hour to propagate)")
	}

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, cmd)
		if err != nil {
			slog.Error("failed to register command", "command", cmd.Name, "error", err)
		} else {
			slog.Info("registered command", "command", cmd.Name)
		}
	}

	slog.Info("bot is running, press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down")
}

type HandlerResult struct {
	Response string
	Err      error
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	var result HandlerResult
	cmd := i.ApplicationCommandData().Name

	switch cmd {
	case "subscribe":
		result = b.handleSubscribe(i)
	case "unsubscribe":
		result = b.handleUnsubscribe(i)
	case "list":
		result = b.handleListForChannel(i)
	}

	respond(s, i, result.Response)

	if result.Err == nil {
		return
	}

	if _, ok := errors.AsType[*UserError](result.Err); ok {
		if os.Getenv("DISCORD_GUILD_ID") != "" {
			slog.Warn("user error", "command", cmd, "error", result.Err, "channel_id", i.ChannelID)
		}
	} else {
		slog.Error("command failed", "command", cmd, "error", result.Err, "channel_id", i.ChannelID)
	}
}

type UserError struct {
	Err error
}

func (e *UserError) Error() string {
	return e.Err.Error()
}

func (e *UserError) Unwrap() error {
	return e.Err
}

func NewUserError(err error) *UserError {
	return &UserError{Err: err}
}

func getOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}
	return ""
}

func (b *Bot) handleSubscribe(i *discordgo.InteractionCreate) HandlerResult {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		return HandlerResult{
			Response: "❌ Invalid Riot ID format. Use `name#tag` (e.g., `Faker#KR1`)",
			Err:      NewUserError(err),
		}
	}

	if !riot.IsValidRegion(region) {
		return HandlerResult{
			Response: fmt.Sprintf("❌ Invalid region: %s", region),
			Err:      NewUserError(fmt.Errorf("invalid region: %s", region)),
		}
	}

	account, err := b.riotClient.GetAccountByRiotID(context.Background(), gameName, tagLine, region)
	if errors.Is(err, riot.ErrNotFound) {
		return HandlerResult{
			Response: fmt.Sprintf("❌ Summoner **%s#%s** not found in **%s**", gameName, tagLine, region),
			Err:      NewUserError(err),
		}
	}
	if err != nil {
		return HandlerResult{
			Response: "❌ Failed to verify summoner. Please try again later.",
			Err:      fmt.Errorf("verify summoner %s in %s: %w", username, region, err),
		}
	}

	canonicalName := fmt.Sprintf("%s#%s", account.GameName, account.TagLine)

	_, err = b.queries.CreateSubscription(context.Background(), db.CreateSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return HandlerResult{
			Response: fmt.Sprintf("⚠️ Already subscribed to **%s** (%s)", canonicalName, region),
			Err:      NewUserError(err),
		}
	}
	if err != nil {
		return HandlerResult{
			Response: "❌ Failed to subscribe. Please try again later.",
			Err:      fmt.Errorf("create subscription for %s in %s: %w", canonicalName, region, err),
		}
	}

	slog.Info("subscription created", "username", canonicalName, "region", region, "channel_id", channelID)
	return HandlerResult{Response: fmt.Sprintf("✅ Subscribed to **%s** (%s)!", canonicalName, region)}
}

func (b *Bot) handleUnsubscribe(i *discordgo.InteractionCreate) HandlerResult {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		return HandlerResult{
			Response: "❌ Invalid Riot ID format. Use `name#tag`",
			Err:      NewUserError(err),
		}
	}

	canonicalName := fmt.Sprintf("%s#%s", gameName, tagLine)

	rowsAffected, err := b.queries.DeleteSubscription(context.Background(), db.DeleteSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
	})
	if err != nil {
		return HandlerResult{
			Response: "❌ Failed to unsubscribe. Please try again later.",
			Err:      fmt.Errorf("delete subscription for %s in %s: %w", canonicalName, region, err),
		}
	}
	if rowsAffected == 0 {
		return HandlerResult{
			Response: fmt.Sprintf("⚠️ No subscription found for **%s** (%s)", canonicalName, region),
			Err:      NewUserError(fmt.Errorf("subscription not found: %s in %s", canonicalName, region)),
		}
	}

	slog.Info("subscription deleted", "username", canonicalName, "region", region, "channel_id", channelID)
	return HandlerResult{Response: fmt.Sprintf("✅ Unsubscribed from **%s** (%s)!", canonicalName, region)}
}

func (b *Bot) handleListForChannel(i *discordgo.InteractionCreate) HandlerResult {
	channelID := i.ChannelID

	subs, err := b.queries.GetSubscriptionsByChannel(context.Background(), channelID)
	if err != nil {
		return HandlerResult{
			Response: "❌ Failed to list subscriptions. Please try again later.",
			Err:      fmt.Errorf("list subscriptions: %w", err),
		}
	}

	if len(subs) == 0 {
		return HandlerResult{Response: "No subscriptions in this channel. Use `/subscribe name#tag region` to add one!"}
	}

	content := "**Subscriptions in this channel:**\n"
	for _, sub := range subs {
		content += fmt.Sprintf("• %s (%s)\n", sub.LolUsername, sub.Region)
	}
	return HandlerResult{Response: content}
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		slog.Error("failed to respond to interaction", "error", err)
	}
}

func (b *Bot) evaluateSubscriptions(ctx context.Context) error {
	// WIP, don't touch this.
	// TODO: support query 1 with query + index, migrate to denorm last_eval_at into subscription
	// 0. Migrate to store server id
	// 1. Grab 1000 oldest subscriptions
	subs, err := b.queries.GetAllSubscriptions(context.Background(), 1000)
	if err != nil {
		return fmt.Errorf("getting all subscriptions %w", err)
	}

	// 2. Group by server ID, concatenate to 20 subscriptions per server id for fairness
	servers := lo.GroupBy(subs, func(s db.Subscription) string {
		return s.ServerID.String
	})
	// 3. For each server grouping
	for _, subs := range servers {
		// 4. For each subscription
		for _, sub := range subs {
			// 5. Grab game info
			// Cached client so no need to denormalize
			username, tag, err := riot.ParseRiotID(sub.LolUsername)
			if err != nil {
				return fmt.Errorf("parsing riot id: %w", err)
			}

			acc, err := b.riotClient.GetAccountByRiotID(ctx, username, tag, sub.Region)
			if err != nil {
				return fmt.Errorf("getting account by riot id: %w", err)
			}

			game, err := b.riotClient.GetActiveGame(ctx, acc.PUUID, sub.Region)
			if errors.Is(err, riot.ErrNotInGame) {
				// TODO: log that user is not in game, or emit a metric. Maybe feature flag for log level?
				continue
			}
			if err != nil {
				return fmt.Errorf("getting active game %w", err)
			}

			_, err = b.queries.GetEvalByGameAndSubscription(ctx,
				db.GetEvalByGameAndSubscriptionParams{
					GameID:         pgtype.Int8{Int64: game.GameID, Valid: true},
					SubscriptionID: sub.ID,
				})
			// if game exists,
			if !errors.Is(err, pgx.ErrNoRows) {
				// TODO: Log that we've already evaluated the game for this subscription.
				continue
			}

			// 6. Filter out only-english names and ignored names
			names := lo.FilterMap(game.Participants, func(p riot.Participant, i int) (string, bool) {
				if !containsForeignCharacters(p.GameName) {
					return "", false
				}
				return p.GameName, true
			})
			// 7. If empty, return
			if len(names) == 0 {
				// TODO: log
				continue
			}

			// 8. Grab any existing translated names
			// TODO: fix the dollar one parameter
			translations, err := b.queries.GetTranslations(ctx, names)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("getting translations: %w", err)
			}
			translatedUsernames := lo.Map(translations, func(t db.Translation, i int) string {
				return t.Username
			})
			names = lo.Filter(names, func(name string, i int) bool {
				return !slices.Contains(translatedUsernames, name)
			})

			// 8.5 ask Translator for the rest

			// 9. combine + transform into nice message format
			// 10. Send message
		}
	}
	return nil
}

func containsForeignCharacters(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

// TODO: job to auto delete subscriptions not positively eval'd in 2 weeks
// Limit
