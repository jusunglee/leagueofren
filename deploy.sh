#!/usr/bin/env bash
set -euo pipefail

# LeagueOfRen deploy script
# Usage: curl -sSL https://raw.githubusercontent.com/jusunglee/leagueofren/main/deploy.sh | bash

echo ""
echo "  ╔══════════════════════════════════════╗"
echo "  ║     League of Ren — Deploy Setup     ║"
echo "  ╚══════════════════════════════════════╝"
echo ""

# Check for docker
if ! command -v docker &> /dev/null; then
    echo "Installing Docker via official script..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
    echo "✓ Docker installed"
else
    echo "✓ Docker already installed"
fi

# Clone or update repo
INSTALL_DIR="/opt/leagueofren"
BRANCH="feat/companion-website"
if [ -d "$INSTALL_DIR" ]; then
    echo "✓ Repo exists at $INSTALL_DIR, pulling latest..."
    git -C "$INSTALL_DIR" fetch origin
    git -C "$INSTALL_DIR" checkout "$BRANCH"
    git -C "$INSTALL_DIR" pull origin "$BRANCH"
else
    echo "Cloning repo..."
    git clone -b "$BRANCH" https://github.com/jusunglee/leagueofren.git "$INSTALL_DIR"
    echo "✓ Cloned to $INSTALL_DIR ($BRANCH)"
fi

cd "$INSTALL_DIR"

echo ""
echo "─── Configuration ───"
echo ""
echo "You'll need:"
echo "  • A Riot Games API key (https://developer.riotgames.com)"
echo "  • An LLM API key (Anthropic or Google)"
echo ""

# Prompt for each value
read -rp "Riot API Key: " RIOT_API_KEY
echo ""

echo "LLM Provider:"
echo "  1) Anthropic (Claude)"
echo "  2) Google (Gemini)"
read -rp "Choose [1/2]: " LLM_CHOICE

if [ "$LLM_CHOICE" = "2" ]; then
    LLM_PROVIDER="google"
    LLM_MODEL="gemini-2.0-flash"
    read -rp "Google API Key: " LLM_KEY
    LLM_KEY_NAME="GOOGLE_API_KEY"
else
    LLM_PROVIDER="anthropic"
    LLM_MODEL="claude-haiku-4-5-20251001"
    read -rp "Anthropic API Key: " LLM_KEY
    LLM_KEY_NAME="ANTHROPIC_API_KEY"
fi

echo ""

# Generate a random postgres password
POSTGRES_PASSWORD=$(openssl rand -hex 16)

# Write .env.prod
cat > "$INSTALL_DIR/.env.prod" << ENVEOF
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
RIOT_API_KEY=$RIOT_API_KEY
LLM_PROVIDER=$LLM_PROVIDER
LLM_MODEL=$LLM_MODEL
${LLM_KEY_NAME}=$LLM_KEY
ENVEOF

chmod 600 "$INSTALL_DIR/.env.prod"

echo "✓ Configuration saved to $INSTALL_DIR/.env.prod"
echo "  (Postgres password auto-generated: $POSTGRES_PASSWORD)"
echo ""

# Deploy
echo "─── Deploying ───"
echo ""

cd "$INSTALL_DIR"
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build

echo ""
echo "─── Done! ───"
echo ""
echo "  Web server: http://$(hostname -I | awk '{print $1}'):3000"
echo "  Worker:     running (hourly refresh)"
echo "  Postgres:   internal (not exposed)"
echo ""
echo "  Point your Cloudflare DNS A record to: $(hostname -I | awk '{print $1}')"
echo "  Enable proxy (orange cloud) for free TLS"
echo ""
echo "  Manage:"
echo "    cd $INSTALL_DIR"
echo "    docker compose --env-file .env.prod -f docker-compose.prod.yml logs -f"
echo "    docker compose --env-file .env.prod -f docker-compose.prod.yml restart"
echo "    docker compose --env-file .env.prod -f docker-compose.prod.yml down"
echo ""
echo "  Update:"
echo "    cd $INSTALL_DIR && git pull && docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build"
echo ""
