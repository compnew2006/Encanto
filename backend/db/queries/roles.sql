-- name: ListPermissions :many
SELECT id, key, resource, action, label, description
FROM permissions
ORDER BY resource, key;

-- name: GetPermissionsByKeys :many
SELECT id, key, resource, action, label, description
FROM permissions
WHERE key = ANY(@keys::text[])
ORDER BY key;

-- name: ListRoles :many
SELECT id, organization_id, created_at, updated_at, deleted_at, name, description, is_system, is_default,
       default_scope_mode, default_allowed_instance_ids, default_allowed_phone_numbers, can_view_unmasked_phone
FROM custom_roles
WHERE organization_id = @organization_id::uuid
  AND deleted_at IS NULL
ORDER BY is_system DESC, name ASC;

-- name: GetRoleByID :one
SELECT id, organization_id, created_at, updated_at, deleted_at, name, description, is_system, is_default,
       default_scope_mode, default_allowed_instance_ids, default_allowed_phone_numbers, can_view_unmasked_phone
FROM custom_roles
WHERE id = @id::uuid
  AND organization_id = @organization_id::uuid
  AND deleted_at IS NULL;

-- name: CreateRole :one
INSERT INTO custom_roles (
  organization_id,
  name,
  description,
  is_system,
  is_default,
  default_scope_mode,
  default_allowed_instance_ids,
  default_allowed_phone_numbers,
  can_view_unmasked_phone
) VALUES (
  @organization_id::uuid,
  @name::text,
  @description::text,
  @is_system::bool,
  @is_default::bool,
  @default_scope_mode::text,
  @default_allowed_instance_ids::jsonb,
  @default_allowed_phone_numbers::jsonb,
  @can_view_unmasked_phone::bool
)
RETURNING id, organization_id, created_at, updated_at, deleted_at, name, description, is_system, is_default,
          default_scope_mode, default_allowed_instance_ids, default_allowed_phone_numbers, can_view_unmasked_phone;

-- name: UpdateRole :one
UPDATE custom_roles
SET
  name = @name::text,
  description = @description::text,
  default_scope_mode = @default_scope_mode::text,
  default_allowed_instance_ids = @default_allowed_instance_ids::jsonb,
  default_allowed_phone_numbers = @default_allowed_phone_numbers::jsonb,
  can_view_unmasked_phone = @can_view_unmasked_phone::bool,
  updated_at = now()
WHERE id = @id::uuid
  AND organization_id = @organization_id::uuid
RETURNING id, organization_id, created_at, updated_at, deleted_at, name, description, is_system, is_default,
          default_scope_mode, default_allowed_instance_ids, default_allowed_phone_numbers, can_view_unmasked_phone;

-- name: SoftDeleteRole :exec
UPDATE custom_roles
SET deleted_at = now(), updated_at = now()
WHERE id = @id::uuid
  AND organization_id = @organization_id::uuid
  AND is_system = false;

-- name: DeleteRolePermissionsByRole :exec
DELETE FROM role_permissions
WHERE custom_role_id = @custom_role_id::uuid;

-- name: InsertRolePermission :exec
INSERT INTO role_permissions (custom_role_id, permission_id)
VALUES (@custom_role_id::uuid, @permission_id::uuid)
ON CONFLICT DO NOTHING;

-- name: ListRolePermissionKeys :many
SELECT p.key
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE rp.custom_role_id = @custom_role_id::uuid
ORDER BY p.key;

