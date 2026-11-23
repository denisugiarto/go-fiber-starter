.PHONY: run test vet build up down
run:   ## jalankan lokal (butuh .env dan Postgres: make up db)
	set -a; . ./.env; set +a; go run ./cmd/api
test:
	go test ./...
vet:
	go vet ./...
build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
up:
	docker compose up --build -d
down:
	docker compose down
