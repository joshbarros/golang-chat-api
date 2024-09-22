# Application name
APP_NAME=golang-chat-api

# Go parameters
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_RUN=$(GO_CMD) run
GO_TEST=$(GO_CMD) test

# Docker parameters
DOCKER_COMPOSE=docker-compose

# Load .env file
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# Makefile commands
.PHONY: all build run test docker docker-build docker-up docker-down wait-for-db migrate-up migrate-down

# Run the application locally
run:
	$(GO_RUN) cmd/app/main.go

# Build the Go application
build:
	$(GO_BUILD) -o $(APP_NAME) ./cmd/app/

# Run unit tests
test:
	$(GO_TEST) ./...

# Clean the build output
clean:
	rm -f $(APP_NAME)

# Docker commands
docker-build:
	$(DOCKER_COMPOSE) build

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

# Run the application inside Docker
docker-run: docker-build docker-up

# Set up everything for dev (useful for scripts)
dev-setup: docker-up wait-for-db build run

# Wait for PostgreSQL to be ready
wait-for-db:
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10  # Adjust the sleep time as necessary
	@until docker exec golang-chat-api-postgres-1 pg_isready -U postgres; do \
		echo "Waiting for PostgreSQL..."; \
		sleep 2; \
	done
	@echo "PostgreSQL is ready!"

# Migrate database up
migrate-up:
	migrate -path ./db/migrations -database "postgres://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable" up

# Migrate database down
migrate-down:
	migrate -path ./db/migrations -database "postgres://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable" down
