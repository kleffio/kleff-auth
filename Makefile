.PHONY: dev run build test fmt lint

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