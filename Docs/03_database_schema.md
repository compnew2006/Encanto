# Database Schema

## Implementation Status

This document now reflects the executable Phase 1-4 schema that ships in the repo as of April 19, 2026.

Canonical sources:

- `backend/db/migrations/000001_phase_1_4.sql`
- `backend/db/queries/*.sql`
- `backend/shared/permission_catalog.json`

If this file ever disagrees with the migration or the shared permission catalog, treat the executable files as the source of truth.

## Shipped Tables

### Identity and tenant context

- `organizations`
- `organization_configs`
- `users`
- `user_organizations`

### Permissions and visibility

- `permissions`
- `custom_roles`
- `role_permissions`
- `user_permission_overrides`
- `user_contact_visibility_rules`

### Chat workspace baseline

- `whatsapp_instances`
- `contacts`
- `messages`
- `conversation_notes`
- `contact_user_states`

## Important Schema Decisions

### Permission vocabulary

The shipped permission model uses `view`, `create`, `update`, and `delete` as the canonical CRUD actions in code. Domain actions remain explicit and separate, including `send`, `revoke`, `export`, `manage`, `contacts.scope.*`, and `chats.unclaimed.*`.

`read_only` is not stored as a permission row. Read-only behavior is derived from the absence of mutating permissions and from tenant state such as `organization_configs.tenant_status = 'read_only'`.

### Role defaults

`custom_roles` stores the default visibility scope for members of the role:

- `default_scope_mode`
- `default_allowed_instance_ids`
- `default_allowed_phone_numbers`
- `can_view_unmasked_phone`

Supported scope modes in the migration are:

- `all_contacts`
- `instances_only`
- `allowed_numbers_only`
- `instances_plus_allowed_numbers`

### User-level overrides

Two different override mechanisms ship in Phase 1-4:

- `user_permission_overrides` stores explicit per-user `allow` or `deny` decisions for a permission key inside one organization.
- `user_contact_visibility_rules` stores per-user visibility scope overrides, including whether the user still inherits the role default scope.

The missing `user_contact_visibility_rules` table from the earlier planning docs is now implemented in the migration and used by the backend visibility resolver.

### Tenant state

`organization_configs` is the Phase 1-4 tenant configuration anchor. It currently ships with:

- `tenant_status` in `active | read_only | suspended`
- storage quota and usage fields
- upload cleanup scheduling fields

### Chat records

The shipped chat baseline is intentionally narrow:

- `contacts.status` is constrained to `assigned | pending | closed`
- `messages.direction` is constrained to `inbound | outbound`
- `contact_user_states` stores per-user hidden, pinned, and read-position state

The data model is ready for later chat actions, but Phase 1-4 only exposes list, detail, messages, notes, and composer gating in the product surface.

## Seeded Fixture Shape

The development seed currently creates:

- two organizations
- four users
- system and custom roles
- role permissions and user overrides
- sample WhatsApp instances
- four deterministic chats with messages and notes

Seed identities and deterministic chat IDs live in `backend/data/seeds.go`.

## Not Yet Shipped

The planning docs still reference additional schema for later phases, including licensing, campaigns, analytics, queueing, realtime event history, and operational inventory. Those tables are not part of the Phase 1-4 migration and should be treated as deferred scope until a later migration introduces them.
