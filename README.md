# LeagueOfRen

A Discord bot that translates Korean and Chinese summoner names in League of Legends games for subscribed users.

## Overview

LeagueOfRen monitors League of Legends players and automatically translates non-English summoner names in their games. When a subscribed player starts a game, the bot detects Korean/Chinese character usernames and provides translations in the Discord channel using AI.

### Name Origin

The name comes from seeing 인 (in) and 人 (ren) frequently in games - Korean and Chinese characters meaning "person". I could never figure out the full names without looking them up, which inspired this bot.

## Features

- **Subscribe to Players**: Track specific League of Legends usernames by region
- **Automatic Detection**: Monitors when subscribed players enter games
- **Smart Translation**: Uses AI (Claude Sonnet or Google Gemma) to translate Korean/Chinese usernames with context
- **Translation Caching**: Stores translations in PostgreSQL to reduce API costs
- **Riot API Caching**: Caches account lookups (24h) and game status (2min) to respect rate limits
- **Status Tracking**: Records each check with status (OFFLINE, NEW_TRANSLATIONS, etc.)

## Tech Stack

- **Language**: Go 1.26
- **Database**: PostgreSQL 16
- **Schema Management**: [Atlas](https://atlasgo.io/) (declarative migrations)
- **Discord**: WebSocket Gateway ([discordgo](https://github.com/bwmarrin/discordgo))
- **APIs**: Riot Games API, Anthropic API, Google AI API
- **Code Generation**: [sqlc](https://sqlc.dev/) (type-safe SQL)
- **Deployment**: Railway

## Quick Start

### 1. Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Atlas CLI](https://atlasgo.io/getting-started#installation): `brew install ariga/tap/atlas`

### 2. Get API Keys

| Service | Purpose | Get it here |
|---------|---------|-------------|
| Discord Bot Token | Bot authentication | [Discord Developer Portal](https://discord.com/developers/applications) → New Application → Bot → Token |
| Riot API Key | Player/game data | [Riot Developer Portal](https://developer.riotgames.com/) → Register → Get API Key |
| Anthropic API Key | AI translation (recommended) | [Anthropic Console](https://console.anthropic.com/) → API Keys |
| Google AI API Key | AI translation (alternative) | [Google AI Studio](https://aistudio.google.com/app/apikey) → Get API Key |

### 3. Clone and Setup

```bash
git clone https://github.com/jusunglee/leagueofren.git
cd leagueofren

# Install tools (Atlas, air for hot reload)
make setup
```

### 4. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your API keys:
```bash
DATABASE_URL=postgres://leagueofren:localdev123@localhost:5432/leagueofren?sslmode=disable
DISCORD_TOKEN=your_discord_bot_token
RIOT_API_KEY=your_riot_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key
GOOGLE_API_KEY=your_google_api_key  # Optional, for Gemma models

# Optional: For faster command registration during development
DISCORD_GUILD_ID=your_test_server_id
```

### 5. Start Database and Apply Schema

```bash
# Start PostgreSQL
make db-up

# Apply database schema
make schema-apply
```

### 6. Run the Bot

```bash
make run

# Or with hot reload for development
make watch
```

## Development Commands

```bash
# Database
make db-up              # Start PostgreSQL container
make db-down            # Stop PostgreSQL container
make db-logs            # View PostgreSQL logs

# Schema (Atlas - declarative migrations)
make schema-apply       # Apply schema.sql to database
make schema-diff        # Preview pending changes (dry run)
make schema-inspect     # View current database schema

# Code generation
make sqlc               # Regenerate Go code from SQL queries

# Run
make run                # Run the bot
make watch              # Run with hot reload (air)
make build              # Build binary to bin/bot

# Testing
make translate-test names="托儿索,페이커"                    # Test with Anthropic (default)
make translate-test names="托儿索" provider=google          # Test with Google Gemma
make translate-test names="托儿索" model=claude-haiku-4-5   # Test with specific model
```

## Schema Changes

This project uses [Atlas](https://atlasgo.io/) for declarative schema management. Instead of writing migrations, you edit `schema.sql` directly:

```bash
# 1. Edit schema.sql with your changes

# 2. Preview what Atlas will do
make schema-diff

# 3. Apply changes
make schema-apply

# 4. Regenerate Go code
make sqlc
```

## Architecture

```
Discord Bot Process
├── WebSocket → Discord Gateway (slash commands, messages)
├── HTTP Client → Riot API (account lookup, spectator)
│   └── PostgreSQL Cache (riot_account_cache, riot_game_cache)
├── HTTP Client → Anthropic/Google API (translate usernames)
└── PostgreSQL Pool → Database (subscriptions, translations, evals)
```

### Database Schema

- `subscriptions`: Discord channel + LoL username + region mappings
- `evals`: Polling check results with game_id tracking
- `translations`: Cached username translations
- `translation_to_evals`: Links translations to specific evals
- `feedback`: User feedback on translations
- `riot_account_cache`: Cached Riot account lookups (24h TTL)
- `riot_game_cache`: Cached game status checks (2min TTL)

## Discord Commands

- `/subscribe username:<name#tag> region:<region>` - Subscribe to a player
- `/unsubscribe username:<name#tag> region:<region>` - Unsubscribe from a player
- `/list` - List all subscriptions in this channel

Supported regions: NA, EUW, EUNE, KR, JP, BR, LAN, LAS, OCE, TR, RU

## Deployment (Railway)

```bash
# 1. Create Railway project and add PostgreSQL
railway init
railway add postgresql

# 2. Set environment variables
railway variables set DISCORD_TOKEN=xxx
railway variables set RIOT_API_KEY=xxx
railway variables set ANTHROPIC_API_KEY=xxx

# 3. Deploy
railway up
```

Railway auto-detects the Dockerfile and injects `DATABASE_URL`.

## Project Structure

```
leagueofren/
├── cmd/
│   ├── bot/                    # Discord bot entry point
│   └── translate-test/         # Translation testing CLI
├── internal/
│   ├── anthropic/              # Anthropic API client
│   ├── google/                 # Google AI API client
│   ├── llm/                    # LLM interface + utilities
│   ├── riot/                   # Riot API client with caching
│   ├── translation/            # Translation service
│   └── db/                     # Database layer (sqlc generated)
├── schema.sql                  # Database schema (source of truth)
├── atlas.hcl                   # Atlas configuration
├── sqlc.yaml                   # sqlc configuration
├── docker-compose.yml          # Local PostgreSQL
├── Dockerfile                  # Production build
└── Makefile                    # Development commands
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

- Riot Games for the League of Legends API
- Discord for the Gateway API
- Anthropic for Claude translation capabilities
- Google for Gemma model access

## Responsible AI Disclosure

I leaned on Claude quite heavily here, but I would stop short of calling this "vibe-coded". I came up with the initial plan in [initial_plan.md](initial_plan.md) and implemented this with prompts progressively by driving the technical direction myself. You can see my entire prompt history in [claude_history.txt](claude_history.txt). I have never enabled allow all edits, I review every suggestion which is hopefully evident by the response history. I also welcome meta commentary on my prompt usage, always open to feedback on this front.
