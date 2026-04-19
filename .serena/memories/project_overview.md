Project Purpose: Encanto / Whatomate is a team-based messaging operations platform centered on live conversations, operational assignment, internal collaboration, account operations, reporting, and licensing.
Tech Stack: Go, chi, pgx, sqlc for Backend. WebSocket inside Go. Svelte 5 and SvelteKit for Frontend. PostgreSQL and Redis for Infra. WhatsApp provider: whatsmeow later (Phase 2).
Code Style & Conventions: Clear separation of Interface, Logic, Data, and Operation layers. camelCase in Svelte, snake_case in DB, PascalCase/camelCase for Go.
Testing: Playwright for E2E tests (`npx playwright test`).
The project is currently in the design/planning phase, with comprehensive markdown specs.