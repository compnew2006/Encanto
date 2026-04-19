-- name: ListUsersByOrganization :many
SELECT
  u.id,
  u.email,
  u.full_name,
  u.avatar_url,
  u.is_active,
  u.availability_status,
  u.settings,
  uo.role_id,
  r.name AS role_name
FROM users u
JOIN user_organizations uo ON uo.user_id = u.id
JOIN custom_roles r ON r.id = uo.role_id
WHERE uo.organization_id = @organization_id::uuid
  AND u.deleted_at IS NULL
  AND uo.deleted_at IS NULL
  AND r.deleted_at IS NULL
ORDER BY u.full_name ASC;

-- name: ListUserPermissionOverrides :many
SELECT permission_key, mode
FROM user_permission_overrides
WHERE organization_id = @organization_id::uuid
  AND user_id = @user_id::uuid
ORDER BY permission_key ASC;

-- name: DeleteUserPermissionOverrides :exec
DELETE FROM user_permission_overrides
WHERE organization_id = @organization_id::uuid
  AND user_id = @user_id::uuid;

-- name: InsertUserPermissionOverride :exec
INSERT INTO user_permission_overrides (
  organization_id,
  user_id,
  permission_key,
  mode
) VALUES (
  @organization_id::uuid,
  @user_id::uuid,
  @permission_key::text,
  @mode::text
)
ON CONFLICT (organization_id, user_id, permission_key)
DO UPDATE SET mode = EXCLUDED.mode, updated_at = now();

-- name: GetUserVisibilityRule :one
SELECT id, organization_id, user_id, scope_mode, allowed_instance_ids, allowed_phone_numbers,
       inherit_role_scope, can_view_unmasked_phone, created_at, updated_at
FROM user_contact_visibility_rules
WHERE organization_id = @organization_id::uuid
  AND user_id = @user_id::uuid;

-- name: UpsertUserVisibilityRule :one
INSERT INTO user_contact_visibility_rules (
  organization_id,
  user_id,
  scope_mode,
  allowed_instance_ids,
  allowed_phone_numbers,
  inherit_role_scope,
  can_view_unmasked_phone
) VALUES (
  @organization_id::uuid,
  @user_id::uuid,
  @scope_mode::text,
  @allowed_instance_ids::jsonb,
  @allowed_phone_numbers::jsonb,
  @inherit_role_scope::bool,
  @can_view_unmasked_phone::bool
)
ON CONFLICT (organization_id, user_id)
DO UPDATE SET
  scope_mode = EXCLUDED.scope_mode,
  allowed_instance_ids = EXCLUDED.allowed_instance_ids,
  allowed_phone_numbers = EXCLUDED.allowed_phone_numbers,
  inherit_role_scope = EXCLUDED.inherit_role_scope,
  can_view_unmasked_phone = EXCLUDED.can_view_unmasked_phone,
  updated_at = now()
RETURNING id, organization_id, user_id, scope_mode, allowed_instance_ids, allowed_phone_numbers,
          inherit_role_scope, can_view_unmasked_phone, created_at, updated_at;
