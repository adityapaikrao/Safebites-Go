-include .env
export

BINARY   := bin/server
CMD      := ./cmd/server
DATABASE_URL ?= postgres://safebites:safebites@localhost:5432/safebites?sslmode=disable

.PHONY: build run test test-cover lint migrate-up migrate-down migrate-create \
        docker-build docker-up docker-down deps tidy

## ── Build ───────────────────────────────────────────────────────────────────

build:
	@mkdir -p bin
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

## ── Test ────────────────────────────────────────────────────────────────────

test:
	go test ./... -v -count=1 -race

test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## ── Lint ────────────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./...

## ── Database Migrations ─────────────────────────────────────────────────────

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

## ── Dependencies ────────────────────────────────────────────────────────────

deps:
	go get github.com/go-chi/chi/v5
	go get github.com/go-chi/cors
	go get github.com/jackc/pgx/v5
	go get github.com/jackc/pgx/v5/pgxpool
	go get github.com/golang-migrate/migrate/v4
	go get github.com/golang-migrate/migrate/v4/database/postgres
	go get github.com/golang-migrate/migrate/v4/source/file
	go get github.com/golang-jwt/jwt/v5
	go get github.com/google/generative-ai-go/genai
	go get google.golang.org/api/option
	go get github.com/joho/godotenv
	go get github.com/stretchr/testify
	go get github.com/pashagolub/pgxmock/v4

tidy:
	go mod tidy

## ── Docker ──────────────────────────────────────────────────────────────────

docker-build:
	docker build -t safebites-backend .

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-postgres:
	docker compose up -d postgres
