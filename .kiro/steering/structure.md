---
inclusion: always
---
# Project Structure: Encanto

```
/
├── frontend/                  # Svelte 5 / SvelteKit web application
│   └── src/routes/            # SvelteKit file-based routing
├── backend/
│   ├── api/                   # HTTP handlers & WebSocket (NO build tags on production files)
│   ├── core/                  # Business logic services (ChatService, AccessService, etc.)
│   ├── models/                # Domain models (if separate from core/types.go)
│   ├── data/                  # PostgreSQL repositories
│   │   ├── sqlc/              # sqlc-generated code (DO NOT edit manually)
│   │   └── store.go           # Store interface and implementation
│   ├── workers/               # Background job processors and CRON runners
│   ├── cache/                 # Redis client and helpers
│   ├── audit/                 # Audit log writer
│   ├── config/                # Environment config loader
│   └── db/
│       ├── migrations/        # SQL migration files (numeric order)
│       └── queries/           # sqlc source SQL queries
└── deploy/
    └── docker-compose.yml
```

## Key Rules
- `backend/api/phase11_16.go` and `backend/api/store.go` are legacy prototype files tagged `//go:build ignore` — do NOT reference them in production code
- The canonical router is initialized in `backend/api/server.go` via `server.Router()`
- `backend/main.go` is the production entry point — `backend/cmd/server/main.go` may exist as an alternative
- Migrations run in numeric order; never modify an already-applied migration — add a new one instead
