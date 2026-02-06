// Seed script for the companion website. Populates public_translations with
// sample data so you can iterate on the frontend design.
//
// Usage:
//
//	go run scripts/seed.go
//	go run scripts/seed.go --database-url postgres://leagueofren:localdev123@localhost:5432/leagueofren
//	go run scripts/seed.go --clear  (wipe all website tables first)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type translation struct {
	Username    string
	Translation string
	Explanation string
	Language    string
	Region      string
}

var samples = []translation{
	{"不知火舞", "Mai Shiranui", "Fighting game character from Fatal Fury/KOF", "chinese", "NA"},
	{"人人人", "Person Person Person", "", "chinese", "NA"},
	{"大魔王", "Great Demon King", "Common gaming nickname meaning 'boss-level player'", "chinese", "KR"},
	{"토르소", "Torso", "Phonetic transliteration, likely a Yasuo main", "korean", "KR"},
	{"페이커", "Faker", "The GOAT of League of Legends", "korean", "KR"},
	{"小鱼人", "Little Fish Man", "Refers to Fizz champion", "chinese", "NA"},
	{"꿈을꾸다", "To Dream", "Poetic summoner name", "korean", "KR"},
	{"暗黑破坏神", "Diablo", "Literally 'Dark Destruction God', reference to the game", "chinese", "NA"},
	{"하늘바라기", "Gazing at the Sky", "Similar to 'sunflower' in Korean sentiment", "korean", "KR"},
	{"龙王归来", "Return of the Dragon King", "Aurelion Sol reference", "chinese", "EUW"},
	{"独孤求败", "Seeking Defeat in Solitude", "Wuxia novel character who was undefeated", "chinese", "NA"},
	{"빛나는별", "Shining Star", "", "korean", "KR"},
	{"狂暴之心", "Heart of Fury", "Could be a Jinx reference", "chinese", "NA"},
	{"무한도전", "Infinite Challenge", "Named after the famous Korean variety show", "korean", "KR"},
	{"千里之行", "Journey of a Thousand Miles", "From the proverb 'starts with a single step'", "chinese", "EUW"},
	{"검은장미", "Black Rose", "Could reference LeBlanc's organization", "korean", "KR"},
	{"风中追风", "Chasing Wind in the Wind", "Yasuo-themed name", "chinese", "NA"},
	{"달빛소나타", "Moonlight Sonata", "Beethoven reference, elegant name", "korean", "KR"},
	{"铁甲雄兵", "Iron Armored Warrior", "Tanky player energy", "chinese", "EUW"},
	{"새벽이슬", "Dawn Dew", "Poetic nature reference", "korean", "KR"},
	{"一剑封喉", "One Sword Seals the Throat", "Assassin player energy", "chinese", "NA"},
	{"푸른하늘", "Blue Sky", "", "korean", "KR"},
	{"烈焰红唇", "Blazing Red Lips", "Flashy aggressive player", "chinese", "NA"},
	{"겨울왕국", "Frozen Kingdom", "Disney's Frozen in Korean", "korean", "KR"},
	{"醉卧沙场", "Drunk on the Battlefield", "From a Tang dynasty war poem", "chinese", "EUW"},
	{"별빛정원", "Starlight Garden", "", "korean", "KR"},
	{"九天揽月", "Reaching for the Moon in Nine Heavens", "Mao Zedong poem reference", "chinese", "NA"},
	{"천둥번개", "Thunder Lightning", "Aggressive mid laner vibes", "korean", "KR"},
	{"落花流水", "Falling Flowers Flowing Water", "Idiom meaning 'utter defeat' or 'scattered'", "chinese", "EUW"},
	{"은하수", "Milky Way", "", "korean", "KR"},
	{"血染战旗", "Blood-Stained War Banner", "Hardcore PvP energy", "chinese", "NA"},
	{"봄날의곰", "Spring Day Bear", "Cozy vibes", "korean", "KR"},
	{"笑傲江湖", "Laughing Proudly Over the Rivers and Lakes", "Jin Yong novel title, means 'carefree wanderer'", "chinese", "NA"},
	{"달콤한독", "Sweet Poison", "Teemo main energy", "korean", "KR"},
	{"天下无双", "Unrivaled Under Heaven", "Claims to be the best", "chinese", "KR"},
	{"하이퍼캐리", "Hyper Carry", "Just the English term transliterated to Korean", "korean", "NA"},
	{"绝地求生", "Survival in a Desperate Situation", "PUBG reference (the Chinese title)", "chinese", "NA"},
	{"미드갱킹", "Mid Ganking", "Literal gameplay description as a name", "korean", "KR"},
	{"刀锋意志", "Will of the Blade", "Irelia's Chinese title", "chinese", "EUW"},
	{"솔로킬장인", "Solo Kill Artisan", "Claims mastery of 1v1s", "korean", "KR"},
}

func main() {
	dsn := flag.String("database-url", "postgres://leagueofren:localdev123@localhost:5432/leagueofren?sslmode=disable", "PostgreSQL connection URL")
	clear := flag.Bool("clear", false, "Clear all website tables before seeding")
	flag.Parse()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("pinging database: %v", err)
	}

	if *clear {
		log.Println("Clearing website tables...")
		pool.Exec(ctx, "TRUNCATE public_translations, votes, public_feedback CASCADE")
	}

	log.Printf("Seeding %d translations...", len(samples))
	for _, s := range samples {
		upvotes := rand.IntN(200)
		downvotes := rand.IntN(30)
		hoursAgo := rand.IntN(720) // up to 30 days
		createdAt := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

		_, err := pool.Exec(ctx, `
			INSERT INTO public_translations (username, translation, explanation, language, region, riot_verified, upvotes, downvotes, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (username) DO UPDATE SET
				translation = EXCLUDED.translation,
				explanation = EXCLUDED.explanation,
				upvotes = EXCLUDED.upvotes,
				downvotes = EXCLUDED.downvotes,
				created_at = EXCLUDED.created_at
		`, s.Username, s.Translation, s.Explanation, s.Language, s.Region, rand.IntN(2) == 1, upvotes, downvotes, createdAt)
		if err != nil {
			log.Printf("  WARN: %s: %v", s.Username, err)
			continue
		}
		fmt.Printf("  ✓ %s → %s (%+d, %s ago)\n", s.Username, s.Translation, upvotes-downvotes, time.Duration(hoursAgo)*time.Hour)
	}

	var count int64
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM public_translations").Scan(&count)
	log.Printf("Done! %d translations in database.", count)
	log.Println("")
	log.Println("To start the site:")
	log.Println("  Terminal 1: go run cmd/web/main.go --database-url 'postgres://leagueofren:localdev123@localhost:5432/leagueofren?sslmode=disable'")
	log.Println("  Terminal 2: cd web && npm run dev")
	log.Println("  Open: http://localhost:5173")
}
