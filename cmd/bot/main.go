package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/anthropic"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/google"
	"github.com/jusunglee/leagueofren/internal/llm"
	"github.com/jusunglee/leagueofren/internal/logger"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/samber/lo"
)

type Bot struct {
	log                          *slog.Logger
	session                      *discordgo.Session
	queries                      *db.Queries
	riotClient                   *riot.CachedClient
	translator                   *translation.Translator
	maxSubscriptionsPerServer    int64
	evaluateSubscriptionsTimeout time.Duration
	evalExpirationDuration       time.Duration
	offlineActivityThreshold     time.Duration
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
	log := logger.New()
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.ErrorContext(ctx, "DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.ErrorContext(ctx, "DISCORD_TOKEN environment variable is required")
		os.Exit(1)
	}

	riotAPIKey := os.Getenv("RIOT_API_KEY")
	if riotAPIKey == "" {
		log.ErrorContext(ctx, "RIOT_API_KEY environment variable is required")
		os.Exit(1)
	}

	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		log.ErrorContext(ctx, "LLM_PROVIDER environment variable is required")
		os.Exit(1)
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		log.ErrorContext(ctx, "LLM_MODEL environment variable is required")
		os.Exit(1)
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		log.ErrorContext(ctx, "DISCORD_GUILD_ID environment variable is required")
		os.Exit(1)
	}

	maxSubscriptionsPerServerStr := os.Getenv("MAX_SUBSCRIPTIONS_PER_SERVER")
	if maxSubscriptionsPerServerStr == "" {
		maxSubscriptionsPerServerStr = "10"
	}
	maxSubscriptionsPerServer, err := strconv.ParseInt(maxSubscriptionsPerServerStr, 10, 64)
	if err != nil {
		log.ErrorContext(ctx, "MAX_SUBSCRIPTIONS_PER_SERVER environment variable is not a valid integer", "error", err)
		os.Exit(1)
	}

	evaluateSubscriptionsTimeoutStr := os.Getenv("EVALUATE_SUBSCRIPTIONS_TIMEOUT")
	if evaluateSubscriptionsTimeoutStr == "" {
		evaluateSubscriptionsTimeoutStr = "1m"
	}
	evaluateSubscriptionsTimeout, err := time.ParseDuration(evaluateSubscriptionsTimeoutStr)
	if err != nil {
		log.ErrorContext(ctx, "EVALUATE_SUBSCRIPTIONS_TIMEOUT environment variable is not a valid duration", "error", err)
		os.Exit(1)
	}

	evalExpirationDurationStr := os.Getenv("EVAL_EXPIRATION_DURATION")
	if evalExpirationDurationStr == "" {
		evalExpirationDurationStr = "504h" // 3 weeks
	}
	evalExpirationDuration, err := time.ParseDuration(evalExpirationDurationStr)
	if err != nil {
		log.ErrorContext(ctx, "EVAL_EXPIRATION_DURATION environment variable is not a valid duration", "error", err)
		os.Exit(1)
	}

	offlineActivityThresholdStr := os.Getenv("OFFLINE_ACTIVITY_THRESHOLD")
	if offlineActivityThresholdStr == "" {
		offlineActivityThresholdStr = "168h" // 1 week
	}
	offlineActivityThreshold, err := time.ParseDuration(offlineActivityThresholdStr)
	if err != nil {
		log.ErrorContext(ctx, "OFFLINE_ACTIVITY_THRESHOLD environment variable is not a valid duration", "error", err)
		os.Exit(1)
	}

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

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.ErrorContext(ctx, "failed to create Discord session", "error", err)
		os.Exit(1)
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.ErrorContext(ctx, "failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.InfoContext(ctx, "connected to database")

	queries := db.New(pool)
	translator := translation.NewTranslator(client, queries, provider, model)

	bot := &Bot{
		log:                          log,
		session:                      dg,
		queries:                      queries,
		riotClient:                   riot.NewCachedClient(riotAPIKey, queries),
		translator:                   translator,
		maxSubscriptionsPerServer:    maxSubscriptionsPerServer,
		evaluateSubscriptionsTimeout: evaluateSubscriptionsTimeout,
		evalExpirationDuration:       evalExpirationDuration,
		offlineActivityThreshold:     offlineActivityThreshold,
	}
	log.InfoContext(ctx, "riot API client initialized with caching")

	dg.AddHandler(bot.handleInteraction)
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.InfoContext(ctx, "connected to Discord", "username", r.User.Username, "discriminator", r.User.Discriminator)
	})

	err = dg.Open()
	if err != nil {
		log.ErrorContext(ctx, "failed to open Discord connection", "error", err)
		os.Exit(1)
	}
	defer dg.Close()

	if guildID != "" {
		log.InfoContext(ctx, "registering commands to guild", "guild_id", guildID)
		_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, "", []*discordgo.ApplicationCommand{})
		if err != nil {
			log.WarnContext(ctx, "failed to clear global commands", "error", err)
		} else {
			log.InfoContext(ctx, "cleared global commands")
		}
	} else {
		log.InfoContext(ctx, "registering commands globally (may take up to 1 hour to propagate)")
	}

	_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, guildID, commands)
	if err != nil {
		log.ErrorContext(ctx, "failed to register commands", "error", err)
	} else {
		log.InfoContext(ctx, "registered commands", "count", len(commands))
	}

	go (func() {
		for {
			evalCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
			err := bot.evaluateSubscriptions(evalCtx)
			if err != nil {
				log.ErrorContext(evalCtx, "running eval", "error", err)
			}
			time.Sleep(time.Minute)
		}
	})()

	go (func() {
		for {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
			err := bot.cleanupOldData(ctx)
			if err != nil {
				log.Error("deleting old data", "error", err)
			}
			time.Sleep(time.Hour)
		}
	})()

	log.InfoContext(ctx, "bot is running, press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.InfoContext(ctx, "shutting down")
}

type HandlerResult struct {
	Response string
	Err      error
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponent(s, i)
	case discordgo.InteractionModalSubmit:
		b.handleModalSubmit(s, i)
	}
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
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

	b.respond(s, i, result.Response)

	if result.Err == nil {
		return
	}

	if _, ok := errors.AsType[*UserError](result.Err); ok {
		if os.Getenv("DISCORD_GUILD_ID") != "" {
			b.log.WarnContext(ctx, "user error", "command", cmd, "error", result.Err, "channel_id", i.ChannelID)
		}
	} else {
		b.log.ErrorContext(ctx, "command failed", "command", cmd, "error", result.Err, "channel_id", i.ChannelID)
	}
}

func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	customID := i.MessageComponentData().CustomID
	messageID := i.Message.ID

	switch customID {
	case "feedback_good":
		_, err := b.queries.CreateFeedback(ctx, db.CreateFeedbackParams{
			DiscordMessageID: messageID,
			FeedbackText:     "👍",
		})
		if err != nil {
			b.log.ErrorContext(ctx, "failed to store positive feedback", "error", err, "message_id", messageID)
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Thanks for the feedback!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	case "feedback_fix":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "feedback_modal:" + messageID,
				Title:    "Suggest a Correction",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "correction_text",
								Label:       "What should the translation be?",
								Style:       discordgo.TextInputParagraph,
								Placeholder: "e.g., 托儿索 should be 'Torso' not 'Yasuo wannabe'",
								Required:    true,
								MaxLength:   500,
							},
						},
					},
				},
			},
		})
	}
}

func (b *Bot) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	data := i.ModalSubmitData()

	parts := strings.Split(data.CustomID, ":")
	if len(parts) != 2 || parts[0] != "feedback_modal" {
		return
	}
	messageID := parts[1]

	var correctionText string
	for _, row := range data.Components {
		if actionsRow, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range actionsRow.Components {
				if input, ok := comp.(*discordgo.TextInput); ok && input.CustomID == "correction_text" {
					correctionText = input.Value
				}
			}
		}
	}

	_, err := b.queries.CreateFeedback(ctx, db.CreateFeedbackParams{
		DiscordMessageID: messageID,
		FeedbackText:     correctionText,
	})
	if err != nil {
		b.log.ErrorContext(ctx, "failed to store correction feedback", "error", err, "message_id", messageID)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Thanks! Your correction has been recorded.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
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
	serverID := i.GuildID
	// TODO: Probably need to handle this better, it's a shame that discordgo doesn't have context built into interactions
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	count, err := b.queries.CountSubscriptionsByServer(ctx, serverID)
	if count > b.maxSubscriptionsPerServer {
		return HandlerResult{
			Response: "❌ Already at maxium subscription count per server, please /unsubscribe to some before subscribing to more.",
		}
	}

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

	account, err := b.riotClient.GetAccountByRiotID(ctx, gameName, tagLine, region)
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

	_, err = b.queries.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
		ServerID:         serverID,
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

	b.log.InfoContext(ctx, "subscription created", "username", canonicalName, "region", region, "channel_id", channelID)
	return HandlerResult{Response: fmt.Sprintf("✅ Subscribed to **%s** (%s)! Will autounsubscribe after 3 weeks of no gameplay.", canonicalName, region)}
}

func (b *Bot) handleUnsubscribe(i *discordgo.InteractionCreate) HandlerResult {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		return HandlerResult{
			Response: "❌ Invalid Riot ID format. Use `name#tag`",
			Err:      NewUserError(err),
		}
	}

	canonicalName := fmt.Sprintf("%s#%s", gameName, tagLine)

	rowsAffected, err := b.queries.DeleteSubscription(ctx, db.DeleteSubscriptionParams{
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

	b.log.InfoContext(ctx, "subscription deleted", "username", canonicalName, "region", region, "channel_id", channelID)
	return HandlerResult{Response: fmt.Sprintf("✅ Unsubscribed from **%s** (%s)!", canonicalName, region)}
}

func (b *Bot) handleListForChannel(i *discordgo.InteractionCreate) HandlerResult {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	channelID := i.ChannelID

	subs, err := b.queries.GetSubscriptionsByChannel(ctx, channelID)
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

func (b *Bot) respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		b.log.ErrorContext(ctx, "failed to respond to interaction", "error", err)
	}
}

func (b *Bot) evaluateSubscriptions(ctx context.Context) error {
	b.log.InfoContext(ctx, "starting eval loop...")
	subs, err := b.queries.GetAllSubscriptions(ctx, 1000)
	if err != nil {
		return fmt.Errorf("getting all subscriptions %w", err)
	}
	b.log.InfoContext(ctx, "subscriptions", "subs", subs, "err", err)

	servers := lo.GroupBy(subs, func(s db.Subscription) string {
		return s.ServerID
	})
	for _, subs := range servers {
		for _, sub := range subs {
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
				b.log.InfoContext(ctx, "user not in game", "username", sub.LolUsername, "region", sub.Region)
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

			if !errors.Is(err, pgx.ErrNoRows) {
				b.log.InfoContext(ctx, "game already evaluated", "subscription_id", sub.ID, "game_id", game.GameID)
				continue
			}

			names := lo.FilterMap(game.Participants, func(p riot.Participant, i int) (string, bool) {
				if !containsForeignCharacters(p.GameName) {
					return "", false
				}
				// Not sure the best way to handle error gracefully but no way it fails at this point right? Ignore for now.
				name, _, _ := riot.ParseRiotID(p.GameName)
				return name, true
			})

			if len(names) == 0 {
				b.log.InfoContext(ctx, "no foreign character names in game", "subscription_id", sub.ID, "game_id", game.GameID, "names", game.Participants)
				continue
			}

			translations, err := b.translator.TranslateUsernames(ctx, names)
			if err != nil {
				return fmt.Errorf("translating usernames: %w", err)
			}

			embed := formatTranslationEmbed(sub.LolUsername, translations)
			msg, err := b.session.ChannelMessageSendComplex(sub.DiscordChannelID, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Good ✓",
								CustomID: "feedback_good",
								Style:    discordgo.SuccessButton,
							},
							discordgo.Button{
								Label:    "Suggest Fix",
								CustomID: "feedback_fix",
								Style:    discordgo.SecondaryButton,
							},
						},
					},
				},
			})
			if err != nil {
				return fmt.Errorf("sending discord message: %w", err)
			}

			// Record the eval
			_, err = b.queries.CreateEval(ctx, db.CreateEvalParams{
				SubscriptionID:   sub.ID,
				EvalStatus:       "NEW_TRANSLATIONS",
				DiscordMessageID: pgtype.Text{String: msg.ID, Valid: true},
				GameID:           pgtype.Int8{Int64: game.GameID, Valid: true},
			})
			if err != nil {
				return fmt.Errorf("creating eval record: %w", err)
			}

			err = b.queries.UpdateSubscriptionLastEvaluatedAt(ctx, sub.ID)
			if err != nil {
				return fmt.Errorf("updating subscription last evaluated at: %w", err)
			}

			b.log.InfoContext(ctx, "sent translation message",
				"subscription_id", sub.ID,
				"channel_id", sub.DiscordChannelID,
				"game_id", game.GameID,
				"translations", len(translations))
		}
	}

	b.log.InfoContext(ctx, "Done evaluating subscriptions", "num_subs", len(subs))
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

func formatTranslationEmbed(username string, translations []translation.Translation) *discordgo.MessageEmbed {
	const maxInlineEntries = 8
	fields := make([]*discordgo.MessageEmbedField, 0, 25)

	inlineCount := min(len(translations), maxInlineEntries)
	for i := 0; i < inlineCount; i++ {
		t := translations[i]
		fields = append(fields,
			&discordgo.MessageEmbedField{Name: "Original", Value: t.Original, Inline: true},
			&discordgo.MessageEmbedField{Name: "Translation", Value: t.Translated, Inline: true},
		)
		if i < inlineCount-1 {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "\u200b", Value: "\u200b", Inline: false})
		}
	}

	if len(translations) > maxInlineEntries {
		var sb strings.Builder
		for _, t := range translations[maxInlineEntries:] {
			sb.WriteString(fmt.Sprintf("**%s** → %s\n", t.Original, t.Translated))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "\u200b",
			Value: sb.String(),
		})
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s is in a game!", username),
		Color:       0x5865F2,
		Description: "Translations for players in this match:",
		Fields:      fields,
	}
}

func (b *Bot) cleanupOldData(ctx context.Context) error {
	log := b.log.With("subsystem", "cleanup_old_data")
	// 1. Delete all evals older than a month
	rows, err := b.queries.DeleteEvals(ctx, pgtype.Timestamptz{Valid: true, Time: time.Now().Add(-b.evalExpirationDuration)})
	if err != nil {
		return fmt.Errorf("deleting old evals: %w", err)
	}
	log.InfoContext(ctx, "Deleted rows", slog.Int64("rows", rows))

	// 2. Find subs where their newest non-offline eval is older than 2 weeks
	subs, err := b.queries.FindSubscriptionsWithExpiredNewestOnlineEval(ctx, pgtype.Timestamptz{Valid: true, Time: time.Now().Add(-b.offlineActivityThreshold)})
	if len(subs) == 0 {
		log.InfoContext(ctx, "No expired subs")
		return nil
	}
	if err != nil {
		return fmt.Errorf("retrieving expired subscriptions: %w", err)
	}

	// 3. Delete these subscriptions
	subIds := lo.Map(subs, func(s db.FindSubscriptionsWithExpiredNewestOnlineEvalRow, _ int) int64 {
		return s.SubscriptionID
	})
	count, err := b.queries.DeleteSubscriptions(ctx, subIds)
	if err != nil {
		return fmt.Errorf("deleting expired subs: %w", err)
	}
	log.InfoContext(ctx, "deleted expired subs", slog.Int64("deleted_subs_count", count))

	return nil
}

// TODO: Support dockerization because I know people would want to run this on their local windows while also playing league
// TODO: Support ignore lists
// TODO: metrics into grafana/loki
// TODO: Coalesce same party-channel-server results into the same message, and ignore them
