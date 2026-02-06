# LeagueOfRen

<p>
    <a href="https://github.com/jusunglee/leagueofren/releases"><img src="https://img.shields.io/github/release/jusunglee/leagueofren.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/jusunglee/leagueofren"><img src="https://godoc.org/github.com/jusunglee/leagueofren?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/jusunglee/leagueofren/actions"><img src="https://github.com/jusunglee/leagueofren/actions/workflows/release.yml/badge.svg?branch=main" alt="Build Status"></a>
</p>

A Discord bot that translates Korean and Chinese summoner names in League of Legends games for subscribed users.

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
- **Deployment**: Railway, Docker, or standalone binary

## Quick Start

### Option A: Windows Users (Just Run It)

If you're on Windows and don't want to compile anything:

1. **Download** the latest `leagueofren-windows-amd64.exe` from [Releases](https://github.com/jusunglee/leagueofren/releases)

2. **Double-click** `leagueofren-windows-amd64.exe` to run

3. **Follow the setup wizard** - On first launch, an interactive wizard walks you through:
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

I leaned on Claude quite heavily here, but I would stop short of calling this "vibe-coded". I came up with the initial plan in [initial_plan.md](initial_plan.md) and implemented this with prompts progressively by driving the technical direction myself. You can see my entire prompt history in [claude_history.txt](claude_history.txt). I have never enabled allow all edits, I review every suggestion which is hopefully evident by the response history. I also welcome meta commentary on my prompt usage, always open to feedback on this front.
