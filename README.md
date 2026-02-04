# LeagueOfRen

A Discord bot that translates Korean and Chinese summoner names in League of Legends games for subscribed users.

## Overview

LeagueOfRen monitors League of Legends players and automatically translates non-English summoner names in their games. When a subscribed player starts a game, the bot detects Korean/Chinese character usernames and provides translations in the Discord channel using AI.

### Name Origin

The name comes from seeing 인 (in) and 人 (ren) frequently in games - Korean and Chinese characters meaning "person". I could never figure out the full names without looking them up, which inspired this bot.

## Features

- **Subscribe to Players**: Track specific League of Legends usernames
- **Automatic Detection**: Monitors when subscribed players enter games
- **Smart Translation**: Uses AI (Claude Haiku) to translate Korean/Chinese usernames
- **Translation Caching**: Stores translations to reduce API costs
- **Status Tracking**: Records each check with status (OFFLINE, NEW_TRANSLATIONS, etc.)

## Tech Stack

- **Language**: Go 1.26rc2, will change this to non-rc later this month when it's released
- **Database**: PostgreSQL 16
- **Discord**: WebSocket Gateway (discordgo)
- **APIs**: Riot Games API, Anthropic API
- **Deployment**: Railway
- **Tools**: sqlc (type-safe queries), golang-migrate (migrations)

## Architecture

```
Discord Bot Process
├── WebSocket → Discord Gateway (receive commands, send translations)
├── HTTP Client → Riot API (check if users are in-game)
├── HTTP Client → Anthropic API (translate usernames)
└── PostgreSQL Pool → Database (subscriptions, translations, evals)
```

### Database Schema

- `subscriptions`: Discord channel + LoL username mappings
- `evals`: Polling check results (OFFLINE, NEW_TRANSLATIONS, REUSE_TRANSLATIONS, NO_TRANSLATIONS)
- `translations`: Cached username translations
- `translation_to_evals`: Links translations to specific evals
- `feedback`: User feedback on translations

## Setup

### Prerequisites

- Go 1.26rc2 or later
- Docker and Docker Compose
- Discord Bot Token
- Riot Games API Key
- Anthropic API Key

### Local Development

1. **Clone the repository**

```bash
git clone https://github.com/jusunglee/leagueofren.git
cd leagueofren
```

2. **Start PostgreSQL**

```bash
make db-up
```

3. **Set up environment variables**

```bash
cp .env.example .env
# Edit .env and add your API keys:
# - DISCORD_TOKEN
# - RIOT_API_KEY
# - ANTHROPIC_API_KEY
```

4. **Run database migrations**

```bash
# Apply migrations directly (migrate CLI has issues with our setup)
docker exec -i leagueofren-db psql -U leagueofren -d leagueofren < migrations/000001_initial_schema.up.sql
```

5. **Run the bot**

```bash
make run
```

### Development Commands

```bash
make db-up          # Start PostgreSQL container
make db-down        # Stop PostgreSQL container
make db-logs        # View PostgreSQL logs
make sqlc           # Regenerate Go code from SQL queries
make run            # Run the bot locally
make build          # Build the bot binary
make clean          # Clean build artifacts
```

## Deployment

### Railway Deployment

1. **Create Railway project**

```bash
railway init
```

2. **Add PostgreSQL**

```bash
railway add postgresql
```

3. **Set environment variables**

```bash
railway variables set DISCORD_TOKEN=your_token
railway variables set RIOT_API_KEY=your_key
railway variables set ANTHROPIC_API_KEY=your_key
```

4. **Deploy**

```bash
git push railway main
```

Railway will automatically:

- Detect the Dockerfile
- Build the multi-stage image
- Inject DATABASE_URL
- Deploy the bot

## Usage

### Discord Commands

(To be implemented)

- `/subscribe <lol_username>` - Subscribe to a League of Legends player
- `/unsubscribe <lol_username>` - Unsubscribe from a player
- `/list` - List all subscriptions in this channel

### How It Works

1. User subscribes to a LoL username in a Discord channel
2. Bot polls Riot API periodically (every 2 minutes by default)
3. When player is in-game, bot extracts Korean/Chinese usernames
4. For new usernames, bot checks cache for existing translations
5. Bot calls Anthropic API for translation if not cached
6. Bot sends translation message to Discord channel
7. Status is recorded in `evals` table

## Configuration

Environment variables:

```bash
# Required
DATABASE_URL=postgres://user:pass@host:port/db
DISCORD_TOKEN=your_discord_bot_token
RIOT_API_KEY=your_riot_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key

# Optional
POLL_INTERVAL_SECONDS=120              # Default: 120 (2 minutes)
MAX_SUBSCRIPTIONS_PER_POLL=100         # Default: 100
```

## Project Structure

```
leagueofren/
├── cmd/bot/                    # Main application entry point
│   └── main.go
├── internal/
│   ├── bot/                    # Discord bot handlers (TODO)
│   ├── riot/                   # Riot API client (TODO)
│   ├── translation/            # Translation service (TODO)
│   ├── poller/                 # Polling loop (TODO)
│   └── db/                     # Database layer
│       ├── conn.go             # Connection pooling
│       ├── queries.sql         # SQL queries
│       ├── models.go           # Generated types
│       └── queries.sql.go      # Generated query functions
├── migrations/                 # Database migrations
├── docker-compose.yml          # Local PostgreSQL
├── Dockerfile                  # Railway deployment
├── Makefile                    # Development commands
└── sqlc.yaml                   # sqlc configuration
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

- Riot Games for the League of Legends API
- Discord for the Gateway API
- Anthropic for Claude Haiku translation capabilities

## Responsible AI Disclosure

I leaned on Claude quite heavily here. You can see my entire prompt history in [claude_history.txt](claude_history.txt). I have never enabled allow all edits, I review every suggestion which is hopefully evident by the response history. I also welcome meta commentary on my prompt usage, always open to feedback on this front.
