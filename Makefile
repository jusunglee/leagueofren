.PHONY: help db-up db-down db-logs migrate-up migrate-down migrate-create sqlc run build clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make db-up          - Start local PostgreSQL container"
	@echo "  make db-down        - Stop local PostgreSQL container"
	@echo "  make db-logs        - View PostgreSQL logs"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-create - Create a new migration (usage: make migrate-create name=my_migration)"
	@echo "  make sqlc           - Generate Go code from SQL queries"
	@echo "  make run            - Run the bot locally"
	@echo "  make build          - Build the bot binary"
	@echo "  make clean          - Clean build artifacts"

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

# Migration commands (using migrate CLI)
# Install migrate: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate-up:
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "Loading DATABASE_URL from .env for local development..."; \
		export $$(cat .env | grep -v '^#' | grep DATABASE_URL | xargs) && \
		migrate -path migrations -database "$$DATABASE_URL" up; \
	else \
		migrate -path migrations -database "$$DATABASE_URL" up; \
	fi

migrate-down:
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "Loading DATABASE_URL from .env for local development..."; \
		export $$(cat .env | grep -v '^#' | grep DATABASE_URL | xargs) && \
		migrate -path migrations -database "$$DATABASE_URL" down 1; \
	else \
		migrate -path migrations -database "$$DATABASE_URL" down 1; \
	fi

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make migrate-create name=my_migration"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)

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

# Build the bot
build:
	go build -o bin/bot cmd/bot/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean
