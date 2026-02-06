.PHONY: help setup db-up db-down db-logs schema-apply schema-diff schema-inspect sqlc run watch build build-all build-windows build-linux build-darwin clean translate-test

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
	@echo "  make build          - Build the bot binary for current platform"
	@echo "  make build-all      - Build for all platforms (Windows, Linux, macOS)"
	@echo "  make build-windows  - Build Windows exe"
	@echo "  make build-linux    - Build Linux binary"
	@echo "  make build-darwin   - Build macOS binary"
	@echo "  make clean          - Clean build artifacts"

# Development setup
setup:
	@go mod tidy

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

# Build the bot for current platform
build:
	go build -o bin/leagueofren cmd/bot/main.go

# Build for all platforms
build-all: build-windows build-linux build-darwin
	@echo "Built all platforms in bin/"

# Build Windows exe
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/leagueofren-windows-amd64.exe cmd/bot/main.go
	@echo "Built bin/leagueofren-windows-amd64.exe"

# Build Linux binary
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/leagueofren-linux-amd64 cmd/bot/main.go
	@echo "Built bin/leagueofren-linux-amd64"

# Build macOS binary
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/leagueofren-darwin-amd64 cmd/bot/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/leagueofren-darwin-arm64 cmd/bot/main.go
	@echo "Built bin/leagueofren-darwin-amd64 and bin/leagueofren-darwin-arm64"

# Clean build artifacts
clean:
	rm -rf bin/ dist/
	go clean
