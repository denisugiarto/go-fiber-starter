.PHONY: run test vet build up down rename
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
rename:  ## ganti nama module: make rename MODULE=github.com/akun/proyek-saya
	@test -n "$(MODULE)" || { echo "pakai: make rename MODULE=github.com/akun/proyek-saya"; exit 1; }
	@old=$$(go list -m); \
	grep -rlF "$$old" --include='*.go' --include='go.mod' --include='*.md' . | xargs sed -i.bak "s#$$old#$(MODULE)#g"; \
	find . -name '*.bak' -delete; go mod tidy; go build ./... && echo "module sekarang: $(MODULE)"
