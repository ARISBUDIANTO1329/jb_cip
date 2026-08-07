.PHONY: build run dev test docker-up docker-down migrate-up migrate-down migrate-version migrate-drop seed clean

APP_NAME=cip-api
APP_PATH=./cmd/api
GO=/usr/local/go/bin/go

build:
	$(GO) build -o bin/$(APP_NAME) $(APP_PATH)
	$(GO) build -o bin/migrate ./cmd/migrate

run:
	$(GO) run $(APP_PATH)

dev:
	cp -n .env.example .env 2>/dev/null || true
	$(GO) run $(APP_PATH)

test:
	$(GO) test ./... -v

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

migrate-up:
	./bin/migrate up

migrate-down:
	./bin/migrate down

migrate-version:
	./bin/migrate version

migrate-drop:
	./bin/migrate drop

seed:
	$(GO) run ./cmd/seed

clean:
	rm -rf bin/

.DEFAULT_GOAL := dev
