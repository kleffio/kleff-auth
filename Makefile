.PHONY: dev run build test fmt lint

# GoLang operands

GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

dev:
	docker compose -f docker-compose.yml up --build

run:
	go run ./cmd/authd

build:
	go build -o bin/authd ./cmd/authd

test:
	go test ./...

fmt:
	gofmt -s -w .

# Database Operands

DB_DSN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS := db/migrations

migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DB_DSN)" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DB_DSN)" down 1

migrate-force:
	migrate -path $(MIGRATIONS) -database "$(DB_DSN)" force ${v}

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS) -seq ${name}
