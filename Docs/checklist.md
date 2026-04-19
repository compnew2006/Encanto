# Phase Checklist

Use this file as the concise status board for the implementation batches.

## Completed

### Phase 1: Stabilize the Final Picture

- [x] `1-1` Final scope approved
- [x] `1-2` Screen and behavior reference unified
- [x] `1-3` First-release success defined

Notes:

- The shipped protected navigation is intentionally limited to `Chat`, `Settings`, and `Profile`.
- Later-phase modules stay hidden instead of appearing as broken placeholders.

### Phase 2: Establish the Project Foundation

- [x] `2-1` Project structure mapped
- [x] `2-2` Conventions fixed
- [x] `2-3` Build order defined

Notes:

- The repo now ships `backend/`, `frontend/`, `deploy/`, root Playwright config, root E2E tests, `.env.example`, and local Docker Compose for Postgres and Redis.
- The backend is structured around `api`, `core`, `data`, `cache`, `config`, and `audit`.
- The frontend is a SvelteKit 5 app with shared permission helpers and protected app-shell routing.

### Phase 3: Identity and Context

- [x] `3-1` Access flow completed
- [x] `3-2` Current-user context completed
- [x] `3-3` Context switching completed

Notes:

- Login, refresh rotation, logout, `/api/me`, `/api/me/organizations`, settings updates, availability updates, and organization switching are implemented.
- Access tokens use `httpOnly` JWT cookies and refresh rotation is backed by Redis.
- Switching organizations rotates session state and changes downstream visibility and management queries immediately.

### Phase 4: Permissions and Visibility

- [x] `4-1` Action-based permission model defined
- [x] `4-2` Visibility rules implemented
- [x] `4-3` UI aligned with enforcement

Notes:

- The shared permission catalog is implemented in `backend/shared/permission_catalog.json` and consumed by both backend and frontend code.
- Role defaults, per-user permission overrides, and per-user contact visibility overrides are all implemented.
- Chat list, chat detail, messages, notes, settings navigation, and composer availability all use the same effective-access result.
- Direct access to blocked chats returns a controlled denial.

## Remaining

- [ ] Phase 5: Core models beyond the Phase 1-4 baseline
- [ ] Phase 6: Conversation workspace actions, richer inbox controls, and operational tooling
- [ ] Phase 7: Sending and receiving pipeline
- [ ] Phase 8: Realtime and notifications
- [ ] Phase 9: Broader settings center behavior
- [ ] Phase 10: WhatsApp account operations and health
- [ ] Phase 11: Contacts CRUD and closed conversations
- [ ] Phase 12: Licensing and limits
- [ ] Phase 13: Analytics
- [ ] Phase 14: Campaigns
- [ ] Phase 15: Reliability and audit expansion
- [ ] Phase 16: Final cleanup and extended handoff
