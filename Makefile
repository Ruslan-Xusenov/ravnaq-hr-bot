.PHONY: build run test migrate-up migrate-down setup

# Variables
APP_NAME=hrbot
DB_URL=postgres://hrbot:hrbot_secret@localhost:5432/hrbot?sslmode=disable
MIGRATIONS_DIR=./migrations

build:
	@echo "Building application..."
	@go build -o bin/$(APP_NAME) ./cmd/server

run: build
	@echo "Running application..."
	@./bin/$(APP_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

setup:
	@echo "Setting up dependencies..."
	@go mod tidy
	@go install github.com/pressly/goose/v3/cmd/goose@latest

migrate-up:
	@echo "Running migrations..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	@echo "Rolling back migrations..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

down:
	@echo "Stopping services..."
	@docker-compose down