package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/db"
)

var queries *db.Queries

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "subscribe",
		Description: "Subscribe to League of Legends summoner translations",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "League of Legends summoner name",
				Required:    true,
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
				Description: "League of Legends summoner name",
				Required:    true,
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
		log.Fatal("DATABASE_URL environment variable is required")
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Connected to database")

	queries = db.New(pool)

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	dg.AddHandler(handleInteraction)
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("✅ Connected to Discord as %s#%s", r.User.Username, r.User.Discriminator)
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord connection: %v", err)
	}
	defer dg.Close()

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Failed to register command %s: %v", cmd.Name, err)
		} else {
			log.Printf("✅ Registered command: /%s", cmd.Name)
		}
	}

	log.Println("Bot is running. Press Ctrl+C to stop.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "subscribe":
		handleSubscribe(s, i)
	case "unsubscribe":
		handleUnsubscribe(s, i)
	case "list":
		handleListForChannel(s, i)
	}
}

func handleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	username := i.ApplicationCommandData().Options[0].StringValue()
	channelID := i.ChannelID

	_, err := queries.CreateSubscription(context.Background(), db.CreateSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      username,
	})

	var content string
	if err != nil {
		content = fmt.Sprintf("❌ Failed to subscribe: %v", err)
	} else {
		content = fmt.Sprintf("✅ Subscribed to **%s**!", username)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func handleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	username := i.ApplicationCommandData().Options[0].StringValue()
	channelID := i.ChannelID

	err := queries.DeleteSubscription(context.Background(), db.DeleteSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      username,
	})

	var content string
	if err != nil {
		content = fmt.Sprintf("❌ Failed to unsubscribe: %v", err)
	} else {
		content = fmt.Sprintf("✅ Unsubscribed from **%s**!", username)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func handleList() ([]db.Subscription, error) {
	return queries.GetAllSubscriptions(context.Background())
}

func handleListForChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channelID := i.ChannelID

	subs, err := queries.GetSubscriptionsByChannel(context.Background(), channelID)

	var content string
	if err != nil {
		content = fmt.Sprintf("❌ Failed to list subscriptions: %v", err)
	} else if len(subs) == 0 {
		content = "No subscriptions in this channel. Use `/subscribe <username>` to add one!"
	} else {
		content = "**Subscriptions in this channel:**\n"
		for _, sub := range subs {
			content += fmt.Sprintf("• %s\n", sub.LolUsername)
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}
