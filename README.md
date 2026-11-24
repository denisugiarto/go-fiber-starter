# Go Fiber Starter

Template REST API siap pakai: Go + [Fiber v3](https://github.com/gofiber/fiber), PostgreSQL, auth JWT, validasi, dan dokumentasi OpenAPI, dengan arsitektur hexagonal.

## Isi
- Auth: daftar, masuk (bcrypt + JWT HS256), `GET /api/v1/me`
- Contoh CRUD terlindungi `/api/v1/notes` (tiap catatan hanya bisa diakses pemiliknya). Ganti atau hapus sesuai kebutuhan.
- Validasi request, format error seragam `{"error":{"code","message","details"}}`
- Konfigurasi dari environment variable, graceful shutdown, log terstruktur, request ID, CORS, recover
- `/health`, `/ready` (cek database), `/docs` (Swagger UI, butuh internet), `/openapi.yaml`
- Migrasi SQL otomatis saat start, Dockerfile (distroless, non-root), `docker-compose.yml`

## Pakai sebagai template
1. Klik **Use this template** di GitHub, atau `gh repo create akun/proyek-saya --template denisugiarto/go-fiber-starter --clone`.
2. Ganti nama module (sekali saja):
   ```sh
   make rename MODULE=github.com/akun/proyek-saya
   ```
3. Siapkan konfigurasi:
   ```sh
   cp .env.example .env
   # isi JWT_SECRET (min. 32 karakter): openssl rand -hex 32
   # ganti DB_PASSWORD, dan samakan password di DATABASE_URL
   ```
4. Jalankan:
   ```sh
   docker compose up --build        # API di http://localhost:8080, dokumen di /docs
   ```
   Tanpa Docker untuk API: sediakan Postgres (mis. `docker compose up db -d`), lalu `make run`.

Butuh Go 1.24+ untuk menjalankan tanpa Docker.

## Coba API
```sh
curl -X POST localhost:8080/api/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","name":"Budi","password":"rahasia123"}'
curl -X POST localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","password":"rahasia123"}'      # ambil access_token
curl localhost:8080/api/v1/me -H "Authorization: Bearer <token>"
```

## Perintah
| Perintah | Fungsi |
|---|---|
| `make run` | jalankan lokal (baca `.env`) |
| `make test` | tes (tanpa database) |
| `make vet` | analisis statis |
| `make build` | binary di `bin/api` |
| `make up` / `make down` | compose naik / turun |
| `make rename MODULE=...` | ganti nama module |

## Konfigurasi (environment variable)
| Nama | Wajib | Default | Keterangan |
|---|---|---|---|
| `DATABASE_URL` | ya | | URL PostgreSQL |
| `JWT_SECRET` | ya | | minimal 32 karakter acak |
| `PORT` | tidak | `8080` | |
| `JWT_TTL` | tidak | `24h` | masa berlaku token |
| `CORS_ORIGINS` | tidak | `http://localhost:3000` | dipisah koma |
| `APP_ENV` | tidak | `development` | |

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

## Menambah fitur
1. Entitas dan port di `internal/domain`
2. Aturan bisnis di `internal/usecase`
3. Repository di `internal/adapter/postgres` dan berkas migrasi `migrations/00N_*.sql`
4. Handler dan route di `internal/adapter/http`, lalu tambahkan ke `docs/openapi.yaml`

## Catatan sebelum produksi
- Jangan commit `.env`. Ganti `JWT_SECRET` dan `DB_PASSWORD` per lingkungan.
- Token belum bisa dicabut sebelum kedaluwarsa (belum ada refresh token atau denylist), dan belum ada pembatasan percobaan login. Tambahkan sebelum produksi.
- Repository Postgres belum punya tes integrasi otomatis. `docker compose` belum diuji dalam pengembangan template ini.
