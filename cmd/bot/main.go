package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/db"
)

func main() {
	// Railway will provide environment variables directly
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	log.Println("✅ Successfully connected to database!")

	queries := db.New(pool)

	// Test
	subscriptions, err := queries.GetAllSubscriptions(ctx)
	if err != nil {
		log.Fatalf("Failed to query subscriptions: %v", err)
	}

	fmt.Println("\n🎮 Hello World from LeagueOfRen! 🎮")
	fmt.Printf("\n📊 Current subscriptions count: %d\n", len(subscriptions))

	if len(subscriptions) > 0 {
		fmt.Println("\nSubscriptions:")
		for _, sub := range subscriptions {
			createdAt := "unknown"
			if sub.CreatedAt.Valid {
				createdAt = sub.CreatedAt.Time.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("  - Channel: %s, User: %s (created: %s)\n",
				sub.DiscordChannelID,
				sub.LolUsername,
				createdAt)
		}
	} else {
		fmt.Println("\n💡 No subscriptions yet. Use the Discord bot to add some!")
	}

	fmt.Println("\n✨ Bot is ready! Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Shutting down gracefully...")
}
