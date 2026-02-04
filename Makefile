.PHONY: help setup db-up db-down db-logs schema-apply schema-diff schema-inspect sqlc run watch build clean translate-test

# Default target
help:
	@echo "Available commands:"
	@echo "  make setup          - Set up development environment (git hooks, atlas)"
	@echo "  make db-up          - Start local PostgreSQL container"
	@echo "  make db-down        - Stop local PostgreSQL container"
	@echo "  make db-logs        - View PostgreSQL logs"
	@echo "  make schema-apply   - Apply schema changes to local database"
	@echo "  make schema-diff    - Show pending schema changes (dry run)"
	@echo "  make schema-inspect - Inspect current database schema"
	@echo "  make sqlc           - Generate Go code from SQL queries"
	@echo "  make run            - Run the bot locally"
	@echo "  make watch          - Run the bot with live reload"
	@echo "  make translate-test - Test translation (usage: make translate-test names=\"托儿索,페이커\")"
	@echo "  make build          - Build the bot binary"
	@echo "  make clean          - Clean build artifacts"

# Development setup
setup:
	git config core.hooksPath .githooks
	@command -v atlas >/dev/null 2>&1 || { echo "Installing atlas..."; brew install ariga/tap/atlas; }
	@echo "Development environment configured"

# Docker Compose commands
db-up:
	docker compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "PostgreSQL is ready!"

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres

# Atlas schema commands
schema-apply:
	atlas schema apply --env local --auto-approve

schema-diff:
	atlas schema apply --env local --dry-run

schema-inspect:
	atlas schema inspect --env local

# Generate code from SQL
sqlc:
	sqlc generate

# Run the bot
run:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
	fi
	go run cmd/bot/main.go

# Run the bot with live reload
watch:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
	fi
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/air-verse/air@latest; }
	air

# Test translation client
# usage: make translate-test names="托儿索,페이커"
# usage: make translate-test names="托儿索" provider=google
# usage: make translate-test names="托儿索" model=claude-haiku-4-5
translate-test:
	@if [ -z "$(names)" ]; then \
		echo "Usage: make translate-test names=\"托儿索,페이커\" [provider=anthropic|google] [model=MODEL]"; \
		exit 1; \
	fi
	go run cmd/translate-test/main.go -names "$(names)" -provider "$(or $(provider),anthropic)" -model "$(model)"

# Build the bot
build:
	go build -o bin/bot cmd/bot/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean
