package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/riot"
)

type Bot struct {
	queries    *db.Queries
	riotClient *riot.Client
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

	guildID := os.Getenv("DISCORD_GUILD_ID")

	ctx := context.Background()

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to database")

	bot := &Bot{
		queries:    db.New(pool),
		riotClient: riot.NewClient(riotAPIKey),
	}
	slog.Info("riot API client initialized")

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

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "subscribe":
		b.handleSubscribe(s, i)
	case "unsubscribe":
		b.handleUnsubscribe(s, i)
	case "list":
		b.handleListForChannel(s, i)
	}
}

func getOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}
	return ""
}

func (b *Bot) handleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		respond(s, i, "❌ Invalid Riot ID format. Use `name#tag` (e.g., `Faker#KR1`)")
		return
	}

	if !riot.IsValidRegion(region) {
		respond(s, i, fmt.Sprintf("❌ Invalid region: %s", region))
		return
	}

	account, err := b.riotClient.GetAccountByRiotID(gameName, tagLine, region)
	if errors.Is(err, riot.ErrNotFound) {
		respond(s, i, fmt.Sprintf("❌ Summoner **%s#%s** not found in **%s**", gameName, tagLine, region))
		return
	}
	if err != nil {
		slog.Error("failed to verify summoner", "error", err, "username", username, "region", region)
		respond(s, i, "❌ Failed to verify summoner. Please try again later.")
		return
	}

	canonicalName := fmt.Sprintf("%s#%s", account.GameName, account.TagLine)

	_, err = b.queries.CreateSubscription(context.Background(), db.CreateSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		respond(s, i, fmt.Sprintf("⚠️ Already subscribed to **%s** (%s)", canonicalName, region))
		return
	}
	if err != nil {
		slog.Error("failed to create subscription", "error", err, "username", canonicalName, "region", region, "channel_id", channelID)
		respond(s, i, "❌ Failed to subscribe. Please try again later.")
		return
	}

	slog.Info("subscription created", "username", canonicalName, "region", region, "channel_id", channelID)
	respond(s, i, fmt.Sprintf("✅ Subscribed to **%s** (%s)!", canonicalName, region))
}

func (b *Bot) handleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		respond(s, i, "❌ Invalid Riot ID format. Use `name#tag`")
		return
	}

	canonicalName := fmt.Sprintf("%s#%s", gameName, tagLine)

	rowsAffected, err := b.queries.DeleteSubscription(context.Background(), db.DeleteSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
	})
	if err != nil {
		slog.Error("failed to delete subscription", "error", err, "username", canonicalName, "region", region, "channel_id", channelID)
		respond(s, i, "❌ Failed to unsubscribe. Please try again later.")
		return
	}
	if rowsAffected == 0 {
		respond(s, i, fmt.Sprintf("⚠️ No subscription found for **%s** (%s)", canonicalName, region))
		return
	}

	slog.Info("subscription deleted", "username", canonicalName, "region", region, "channel_id", channelID)
	respond(s, i, fmt.Sprintf("✅ Unsubscribed from **%s** (%s)!", canonicalName, region))
}

func (b *Bot) handleList() ([]db.Subscription, error) {
	return b.queries.GetAllSubscriptions(context.Background())
}

func (b *Bot) handleListForChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channelID := i.ChannelID

	subs, err := b.queries.GetSubscriptionsByChannel(context.Background(), channelID)
	if err != nil {
		slog.Error("failed to list subscriptions", "error", err, "channel_id", channelID)
		respond(s, i, "❌ Failed to list subscriptions. Please try again later.")
		return
	}

	if len(subs) == 0 {
		respond(s, i, "No subscriptions in this channel. Use `/subscribe name#tag region` to add one!")
		return
	}

	content := "**Subscriptions in this channel:**\n"
	for _, sub := range subs {
		content += fmt.Sprintf("• %s (%s)\n", sub.LolUsername, sub.Region)
	}
	respond(s, i, content)
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
