# Go Fiber Starter

Template starter Go + [Fiber v3](https://github.com/gofiber/fiber) dengan arsitektur hexagonal, PostgreSQL, auth JWT, validasi, dan dokumentasi OpenAPI.

## Fitur
- Auth: daftar, masuk (bcrypt + JWT HS256), `GET /api/v1/me`
- Contoh CRUD terlindungi: `/api/v1/notes` (tiap catatan hanya bisa diakses pemiliknya)
- Validasi request (validator), format error seragam `{"error":{"code","message","details"}}`
- Konfigurasi dari environment variable, graceful shutdown, log terstruktur (`slog`), request ID, CORS, recover
- `/health` (liveness), `/ready` (cek database), `/docs` (Swagger UI, butuh internet untuk CDN), `/openapi.yaml`
- Migrasi SQL otomatis saat start (`migrations/*.sql`, tercatat di `schema_migrations`)
- Dockerfile (distroless, non-root) dan `docker-compose.yml`

## Struktur
```
cmd/api/                  titik masuk, merangkai semuanya
internal/domain/          entitas, error, port (interface); tanpa dependensi lain
internal/usecase/         aturan bisnis; hanya bergantung ke domain
internal/adapter/http/    Fiber: router, handler, DTO, middleware
internal/adapter/postgres pgx: repository + runner migrasi
internal/adapter/jwt/     implementasi TokenService
internal/adapter/memory/  repository in-memory (khusus tes)
internal/config/          baca env
migrations/  docs/        SQL migrasi dan OpenAPI (di-embed ke binary)
```
Aturan dependensi: `adapter -> usecase -> domain`. Domain tidak mengimpor siapa pun.

## Mulai cepat
```sh
cp .env.example .env          # isi JWT_SECRET: openssl rand -hex 32, dan ganti DB_PASSWORD
docker compose up --build     # API di http://localhost:8080, dokumen di /docs
```
Tanpa Docker untuk API: `docker compose up db -d`, lalu `make run`.

Coba:
```sh
curl -X POST localhost:8080/api/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","name":"Budi","password":"rahasia123"}'
curl -X POST localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","password":"rahasia123"}'      # ambil access_token
curl localhost:8080/api/v1/me -H "Authorization: Bearer <token>"
```

## Tes
`make test`. Tes memakai repository in-memory, jadi tidak butuh database. Repository Postgres belum punya tes integrasi otomatis.

## Menambah fitur
1. Entitas dan port di `internal/domain`
2. Aturan bisnis di `internal/usecase`
3. Repository di `internal/adapter/postgres` (+ berkas migrasi `migrations/00N_*.sql`)
4. Handler/route di `internal/adapter/http`, dan tambahkan ke `docs/openapi.yaml`

## Catatan
- Jangan commit `.env`. Ganti `JWT_SECRET` dan `DB_PASSWORD` per lingkungan.
- Token belum bisa dicabut sebelum kedaluwarsa (tanpa refresh token/denylist) dan belum ada rate limiting pada login; tambahkan sebelum produksi.
