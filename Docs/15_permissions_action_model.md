# Action-Based Permissions Model

## Implementation Status

This file reflects the shipped Phase 1-4 permission model as of April 19, 2026.

Canonical source of truth:

- `backend/shared/permission_catalog.json`
- `backend/core/access.go`
- `frontend/src/lib/permissions.ts`

## Permission Structure

Every permission key follows the shape:

`resource.action`

Examples:

- `chats.view`
- `messages.send`
- `contacts.update`
- `roles.manage`
- `contacts.scope.allowed_numbers`

## Canonical Actions

The shipped CRUD vocabulary is:

- `view`
- `create`
- `update`
- `delete`

The older `edit` wording from planning docs is no longer canonical in code. Use `update` everywhere for implemented permissions and APIs.

Additional domain actions remain explicit:

- `send`
- `revoke`
- `export`
- `manage`
- `scope.*`

## Read-Only Semantics

`read_only` is not a stored permission row.

Read-only behavior is derived from effective access:

- a user who has `chats.view` but not `messages.send` can read allowed chats but cannot send
- a tenant in `organization_configs.tenant_status = 'read_only'` should retain safe read paths while write paths are blocked

This is why the product enforces individual action checks instead of checking a role name or a synthetic `read_only` permission.

## Shipped Permission Families

The current catalog supports the Phase 1-4 surface:

- chat visibility and pending-chat access
- messages send and revoke actions
- notes read and create actions
- contacts CRUD plus visibility scope modifiers
- settings management
- role management
- user viewing and updating

The precise key list lives in `backend/shared/permission_catalog.json` and is consumed by both the backend and frontend.

## Visibility Model

Visibility is resolved in layers:

1. The role provides the default scope through `custom_roles`.
2. Per-user visibility overrides live in `user_contact_visibility_rules`.
3. Per-user permission exceptions live in `user_permission_overrides`.
4. The backend resolves one effective-access result and applies it to list and detail endpoints.
5. The frontend consumes the same effective-access result to hide, disable, or explain actions.

Supported scope modes in the shipped schema:

- `all_contacts`
- `instances_only`
- `allowed_numbers_only`
- `instances_plus_allowed_numbers`

## UI and API Parity

The frontend does not check role names such as `admin` or `agent`. It checks effective permissions.

Current helpers and components:

- `hasPermission(...)`
- `denialReason(...)`
- `PermissionGate`
- `PermissionButton`

Current backend enforcement points include:

- `/api/roles*`
- `/api/users*`
- `/api/chats`
- `/api/chats/{contactID}`
- `/api/contacts/{contactID}/messages`
- `/api/contacts/{contactID}/notes`

The intended behavior is:

- hide objects the user should not discover
- disable actions the user may see but may not perform
- explain the denial when the action is disabled
- reject manipulated requests on the backend

## Phase 1-4 Notes

- Settings navigation is shown only when the user has the relevant settings, roles, or users permissions.
- Pending chats are hidden without `chats.unclaimed.view`.
- The composer is disabled without `messages.send`.
- Direct navigation to a blocked chat returns a controlled denial instead of leaking data.
