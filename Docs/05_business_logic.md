# Business Logic

## Implementation Status

This document now describes the business logic that is actually implemented for Phases 1-4 as of April 19, 2026.

Canonical executable sources:

- `backend/api/router.go`
- `backend/core/session.go`
- `backend/core/access.go`
- `backend/core/chat.go`
- `frontend/src/hooks.server.ts`
- `frontend/src/routes/(app)/**`

Anything beyond the scope listed below is still planned work, not shipped behavior.

## Shipped Scope

### Phase 3: identity and context

Implemented and verified:

- login via `POST /api/auth/login`
- refresh rotation via `POST /api/auth/refresh`
- logout via `POST /api/auth/logout`
- organization switching via `POST /api/auth/switch-org`
- current session bootstrap via `GET /api/me`
- organization membership listing via `GET /api/me/organizations`
- personal settings updates via `PUT /api/me/settings`
- personal availability updates via `PUT /api/me/availability`

Session behavior:

- access tokens are short-lived JWTs stored in `httpOnly` cookies
- refresh tokens are rotated and backed by Redis
- org switching rotates both access and refresh state
- the frontend attempts refresh on protected-route bootstrap when a refresh cookie exists

### Phase 4: permissions and visibility

Implemented and verified:

- shared permission catalog consumed by backend and frontend
- role CRUD with permission matrix editing
- user availability updates
- per-user send restriction overrides
- per-user contact visibility overrides
- chat list and chat detail visibility enforcement
- messages and notes read views
- composer gating based on effective access

The same effective-access logic is used to decide:

- which chats appear in the list
- whether direct chat detail access is allowed
- whether pending chats are visible
- whether the composer is enabled
- whether the user sees masked or unmasked phone data

Blocked direct access returns a controlled denial instead of leaking contact data.

## Frontend Surface That Ships

Only these protected product areas are exposed in the navigation:

- `Chat`
- `Settings`
- `Profile`

The settings shell includes:

- organization switcher
- sidebar pin preference
- theme preference
- language preference
- availability control
- logout

Later-phase product areas such as dashboard, campaigns, analytics, chatbot, licensing, contacts CRUD, and instance operations remain hidden rather than stubbed.

## Seeded Verification Flows

The seed data supports the main Phase 1-4 proof cases:

- `admin@example.com` can manage roles and users
- `internal@example.com` can switch organizations
- `readonly@example.com` can read chats but cannot send
- `scoped@example.com` only sees the contacts allowed by the visibility resolver

All seeded users use password `password123` in local development.

## Enforcement Rules That Matter

### Read-only behavior

Read-only is derived behavior, not a stored permission. In practice:

- users can still open allowed chats and read message history
- the composer is disabled when `messages.send` is absent
- the UI explains the denial instead of silently failing

### Pending chats

- `chats.unclaimed.view` controls whether pending chats appear at all
- `chats.unclaimed.send` controls whether a user may send before assignment

### Organization switching

- switching is only allowed to active memberships
- the current organization in the session changes immediately
- subsequent `/api/me`, chat queries, role management, and user management all resolve against the new organization context

### Tenant read-only mode

When `organization_configs.tenant_status` is `read_only`, read surfaces stay available while mutating operations are expected to be blocked. The data model and permission logic support this, but the broader quota-cleanup and licensing flows are still deferred to later phases.

## Deferred Logic

The following flows remain out of scope for the current baseline:

- inbound WhatsApp ingestion
- outbound sending pipeline
- realtime WebSocket updates
- notifications and statuses
- assignment workflows, pin/hide actions, and close/reopen actions in the UI
- contacts CRUD and import/export screens
- instance management screens and connection lifecycle
- dashboard, analytics, campaigns, chatbot, and licensing

Those items remain planning material until their corresponding implementation phases land.
