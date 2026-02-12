# LeagueOfRen

<p>
    <a href="https://github.com/jusunglee/leagueofren/releases"><img src="https://img.shields.io/github/release/jusunglee/leagueofren.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/jusunglee/leagueofren"><img src="https://godoc.org/github.com/jusunglee/leagueofren?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/jusunglee/leagueofren/actions/workflows/ci.yml"><img src="https://github.com/jusunglee/leagueofren/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build Status"></a>
</p>

A Discord bot that translates Korean and Chinese summoner names in League of Legends games for subscribed users.

View submitted translations here: https://leagueofren.com

<figure>
  <table>
    <tr>
      <td align="center">
        <img src="https://github.com/user-attachments/assets/8ac774dd-c9d5-480d-8607-8a0a0e2d3080" width="100%" />
      </td>
      <td align="center">
        <img src="https://github.com/user-attachments/assets/c6e8239c-58fa-494c-8716-ee44a6b39944" width="100%" />
      </td>
    </tr>
  </table>

  <figcaption align="center">
    <em>Shoutout to these 2 randos I found on porofessor's live games</em>
  </figcaption>
</figure>

## Overview

LeagueOfRen monitors League of Legends players and automatically translates non-English summoner names in their games. When a subscribed player starts a game, the bot detects Korean/Chinese character usernames and provides translations in the Discord channel using AI.

## Quick Start

### Option A: Windows Users (Just Run It)

If you're on Windows and don't want to compile anything:

1. **Download** the latest `leagueofren-windows-amd64.zip` from [Releases](https://github.com/jusunglee/leagueofren/releases)

2. **Unzip** the archive to a folder of your choice

3. **Double-click** `leagueofren.exe` to run

4. **Follow the setup wizard** - On first launch, an interactive wizard walks you through:
   - Discord bot token setup (with link to Developer Portal)
   - Riot API key setup (with link to Developer Portal)
   - LLM provider choice (Anthropic or Google)
   - LLM API key setup

The wizard saves your configuration to `.env` automatically. The bot creates a local SQLite database - no PostgreSQL or Docker needed.

---

### Option B: Developers (Build from Source)

#### 1. Setup Environment

<details>
<summary><b>Windows (WSL required)</b></summary>

Enable WSL and install Ubuntu:

```powershell
wsl.exe --install
```

Install [Ubuntu from Microsoft Store](https://apps.microsoft.com/detail/9pdxgncfsczv) and open it after restarting.

**All commands below should be run inside the Ubuntu terminal.**

</details>

<details>
<summary><b>macOS</b></summary>

Install Homebrew if you haven't:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

</details>

#### 2. Install Go

**macOS:**

```bash
brew install go
```

**Linux/WSL:**

```bash
curl -LO https://go.dev/dl/go1.26rc2.linux-amd64.tar.gz && \
  rm -rf ~/go && \
  tar -C ~ -xzf go1.26rc2.linux-amd64.tar.gz && \
  rm go1.26rc2.linux-amd64.tar.gz && \
  echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc && \
  source ~/.bashrc
```

#### 3. Clone and Setup

```bash
git clone https://github.com/jusunglee/leagueofren.git
cd leagueofren
make setup
```

#### 4. Run the Bot

```bash
make run
```

On first run, an interactive setup wizard walks you through configuring:

- Discord bot token
- Riot API key
- LLM provider (Anthropic or Google)
- LLM API key

The wizard saves your configuration to `.env` automatically.

**With PostgreSQL (optional, for development):**

```bash
make db-up        # Start PostgreSQL in Docker
make schema-apply # Apply database schema
# Edit .env to set DATABASE_URL to your PostgreSQL connection string
make run
```


### Name Origin

The name comes from seeing 인 (in) and 人 (ren) frequently in games - Korean and Chinese characters meaning "person". I could never figure out the full underlying meanings of the names without looking them up (despite me being Korean), which inspired this bot.

## Other languages

In theory, this is language-scalable but I started this project scoped down since 99% of foreign names I saw in NA were chinese and korean (I also live in NYC so I might just be region-scoped with a larger Asian population). In addition, the value of this feature rests entirely on the robustness and accuracy of the translations from the LLMs. Due to, what I think to be, a cultural phenomenon unique to Korean/Chinese communities where there's a gold mine of online content produced in their respective languages about league to provide enough context to LLM scrapers, I think these 2 languages specifically are well suited perhaps next to English to be potential language candidates for this project. I wonder if to support other languages, we'd have to use an intelligent model adapter based on the language.

## Why Discord
We use Discord instead of just `/msg` -ing you in-game because it's, for good reason, not supported by the official riot server API. There's future plans to do this anyways with the game client API if enough people just want to deploy this locally, since you'll just be whispering to yourself and has gutted potential for abuse.

## Features

- **Subscribe to Players**: Track specific League of Legends usernames by region
- **Automatic Detection**: Monitors when subscribed players enter games
- **Smart Translation**: Uses AI (Claude Sonnet or Google Gemma) to translate Korean/Chinese usernames with context
- **Translation Caching**: Stores translations in PostgreSQL to reduce API costs
- **Riot API Caching**: Caches account lookups (24h) and game status (2min) to respect rate limits
- **Status Tracking**: Records each check with status (OFFLINE, NEW_TRANSLATIONS, etc.)

## Tech Stack

- **Language**: Go 1.26
- **Database**: SQLite (standalone) or PostgreSQL 16 (development/production)
- **Schema Management**: [Atlas](https://atlasgo.io/) (declarative migrations)
- **Discord**: WebSocket Gateway ([discordgo](https://github.com/bwmarrin/discordgo))
- **APIs**: Riot Games API, Anthropic API, Google AI API
- **Code Generation**: [sqlc](https://sqlc.dev/) (type-safe SQL)
- **TUI**: [Bubbletea](https://github.com/charmbracelet/bubbletea) (first-run setup wizard)
- **Releases**: [GoReleaser](https://goreleaser.com/) + GitHub Actions
- **Deployment**: Docker, or standalone binary

## Development Commands

```bash
# Database (PostgreSQL)
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

# Build
make build              # Build binary for current platform
make build-all          # Build for Windows, Linux, macOS
make build-windows      # Build Windows exe only
make build-linux        # Build Linux binary only
make build-darwin       # Build macOS binaries (amd64 + arm64)

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

## Discord Bot Setup

### Required Bot Permissions

When inviting the bot to your server, use the OAuth2 URL Generator in the [Discord Developer Portal](https://discord.com/developers/applications) with these settings:

**Scopes:**
- `bot`
- `applications.commands`

**Bot Permissions:**
- Send Messages
- Embed Links
- Use Slash Commands

Users who run `/subscribe` or `/unsubscribe` must have the **Manage Channels** permission in the channel where they're issuing the command. The `/list` command is available to all users.

## Discord Commands

- `/subscribe username:<name#tag> region:<region>` - Subscribe to a player (requires Manage Channels)
- `/unsubscribe username:<name#tag> region:<region>` - Unsubscribe from a player (requires Manage Channels)
- `/list` - List all subscriptions in this channel

Supported regions: NA, EUW, EUNE, KR, JP, BR, LAN, LAS, OCE, TR, RU

## Deployment

I deployed everything on a digital ocean droplet because I hate usage based pricing. You can deploy this on any box, and more ux friendly dev platforms like railway might just read and handle the docker files for you but I haven't tested it out.

### Droplet instructions
1. Clone the repo
2. Run `gh auth`
3. Run `make deploy`, which will prompt you to set up some env vars before running.


## Observability

The production stack includes a full monitoring setup deployed alongside the application via `docker-compose.prod.yml`.

### Stack

- **Prometheus** — scrapes metrics from the web server, worker, and nginx every 15s. 15-day retention.
- **Loki** — aggregates logs from all Docker containers via Promtail. 15-day retention.
- **Grafana** — visualizes everything through a pre-provisioned "League of Ren" dashboard.
- **Dozzle** — lightweight Docker log viewer for quick container log browsing.

### Accessing Grafana

I access grafana on the droplet via tailscale, you can choose to host + auth it if you wish if you'd rather expose it to the internet.

```
http://GRAFANA_HOST:3001
```

Login with `admin` / the value of `GRAFANA_PASSWORD` in your `.env` (defaults to `admin`).

### Dashboard

The provisioned dashboard auto-refreshes every 30s and includes:

- **Service Health** — uptime, request rate, error rate
- **HTTP Traffic** — request rate by route, p95 latency, status code breakdown, rate limit hits
- **Translations & Votes** — submission counts, vote activity, LLM latency
- **Worker** — refresh cycle duration, Riot API calls/latency, PUUID backfill progress
- **Database Pool** — connection pool utilization
- **Nginx** — active connections, request throughput
- **Go Runtime** — goroutine count, memory usage
- **Logs** — web, worker, and nginx logs with error aggregation (via Loki)

### Application Metrics

Custom Prometheus metrics are defined in `internal/metrics/metrics.go`:

| Metric | Description |
|--------|-------------|
| `lor_http_requests_total` | HTTP requests by route/method/status |
| `lor_http_request_duration_seconds` | Request latency histogram |
| `lor_rate_limit_hits_total` | Rate limit rejections |
| `lor_translation_submissions_total` | Translation submissions |
| `lor_votes_total` | Votes by direction |
| `lor_llm_translation_duration_seconds` | LLM API latency |
| `lor_worker_refresh_duration_seconds` | Worker refresh cycle time |
| `lor_riot_api_calls_total` | Riot API calls by endpoint/result |
| `lor_riot_api_duration_seconds` | Riot API latency |
| `lor_db_pool_*` | Database connection pool stats |

### Config Files

All monitoring config lives in `monitoring/`:

```
monitoring/
├── prometheus.yml                          # Prometheus scrape config
├── loki.yml                                # Loki storage/retention config
├── promtail.yml                            # Promtail → Loki log pipeline
└── grafana/
    ├── provisioning/
    │   ├── datasources/prometheus.yml      # Prometheus + Loki data sources
    │   └── dashboards/dashboards.yml       # Dashboard provisioning
    └── dashboards/
        └── leagueofren.json               # The pre-built dashboard
```

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
│   ├── setup/                  # First-run setup wizard (bubbletea TUI)
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

## Website
The website was 100% vibecoded. I'm not trying to spend hours figuring our which value of strokelinejoin to use.

## Bot
I leaned on Claude quite heavily here, but I would stop short of calling this "vibe-coded". I came up with the initial plan in [initial_plan.md](initial_plan.md) and implemented this with prompts progressively by driving the technical direction myself. You can see my entire prompt history in [claude_history.txt](claude_history.txt). I implemented the main evaluation loop by hand without AI (for fun), you can check out the commit history for the details.
