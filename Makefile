.PHONY: help setup db-up db-down db-logs migrate-up migrate-down migrate-create sqlc run watch build clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make setup          - Set up development environment (git hooks)"
	@echo "  make db-up          - Start local PostgreSQL container"
	@echo "  make db-down        - Stop local PostgreSQL container"
	@echo "  make db-logs        - View PostgreSQL logs"
	@echo "  make migrate-up     - Run all up migrations"
	@echo "  make migrate-down   - Rollback a migration (usage: make migrate-down file=migrations/000002_xxx.down.sql)"
	@echo "  make migrate-create - Create a new migration (usage: make migrate-create name=my_migration)"
	@echo "  make sqlc           - Generate Go code from SQL queries"
	@echo "  make run            - Run the bot locally"
	@echo "  make build          - Build the bot binary"
	@echo "  make clean          - Clean build artifacts"

# Development setup
setup:
	git config core.hooksPath .githooks
	@echo "Git hooks configured"

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

# Migration commands (using docker exec + psql for local development)
migrate-up:
	@echo "Running all up migrations..."
	@for f in migrations/*.up.sql; do \
		echo "Applying $$f..."; \
		docker exec -i leagueofren-db psql -U leagueofren -d leagueofren < "$$f" 2>&1 || true; \
	done
	@echo "Migrations complete!"

migrate-down:
	@if [ -z "$(file)" ]; then \
		echo "Usage: make migrate-down file=migrations/000002_add_region.down.sql"; \
		exit 1; \
	fi
	@echo "Running down migration: $(file)..."
	docker exec -i leagueofren-db psql -U leagueofren -d leagueofren < "$(file)"
	@echo "Rollback complete!"

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make migrate-create name=my_migration"; \
		exit 1; \
	fi
	@num=$$(ls -1 migrations/*.up.sql 2>/dev/null | wc -l | tr -d ' '); \
	num=$$((num + 1)); \
	padded=$$(printf "%06d" $$num); \
	touch "migrations/$${padded}_$(name).up.sql"; \
	touch "migrations/$${padded}_$(name).down.sql"; \
	echo "Created migrations/$${padded}_$(name).up.sql"; \
	echo "Created migrations/$${padded}_$(name).down.sql"

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

# Build the bot
build:
	go build -o bin/bot cmd/bot/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean
