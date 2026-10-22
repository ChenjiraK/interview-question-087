.PHONY: run build test fmt vet db-up db-down db-shell migrate-up migrate-down migrate-version

run: .env
	go run .

build:
	go build -o bin/server .

test:
	go test ./...

fmt:
	gofmt -w main.go internal

vet:
	go vet ./...

.env:
	cp .env.example .env

db-up: .env
	docker compose up -d --wait db

db-down: .env
	docker compose down

db-shell: db-up
	docker compose exec db sh -c 'exec psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'

migrate-up: db-up
	docker compose run --rm migrate up

# Roll back one migration. This drops the user table and its data for version 1.
migrate-down: db-up
	docker compose run --rm migrate down 1

migrate-version: db-up
	docker compose run --rm migrate version
