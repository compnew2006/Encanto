---
inclusion: always
---
# Technology Stack: Encanto

## Backend
- **Language**: Go (Golang), latest stable
- **Router**: go-chi/chi v5
- **Database Driver**: pgx v5
- **Query Generation**: sqlc (queries in `backend/db/queries/`, generated in `backend/data/sqlc/`)
- **WhatsApp**: go.mau.fi/whatsmeow
- **Auth**: golang-jwt/jwt v5 (httpOnly cookies, access + refresh tokens)
- **WebSocket**: gorilla/websocket
- **Cache**: Redis via `backend/cache/redis.go`

## Frontend
- **Framework**: Svelte 5 + SvelteKit
- **Language**: TypeScript
- **Testing**: Playwright (E2E)

## Database
- **Engine**: PostgreSQL
- **Migrations**: SQL files in `backend/db/migrations/` — applied in numeric order
- **Schema source of truth**: `backend/db/migrations/001_schema.sql`

## Infrastructure
- **Containerization**: Docker Compose (`deploy/docker-compose.yml`)
- **Config**: Environment variables loaded via `backend/config/config.go`

## Conventions
- All business logic lives in `backend/core/` — never in `backend/api/` handlers
- HTTP handlers in `backend/api/` only parse requests, call core services, and write responses
- SQL queries must go through sqlc — no raw query strings in handlers
- Build tags must NOT be used on production files; `//go:build ignore` is for scratch/prototype files only
- All secrets must be set via environment variables — no hardcoded defaults in production
