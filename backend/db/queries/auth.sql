-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, avatar_url, is_active, is_super_admin, availability_status, settings, last_login
FROM users
WHERE email = @email::text
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, avatar_url, is_active, is_super_admin, availability_status, settings, last_login
FROM users
WHERE id = @id::uuid
  AND deleted_at IS NULL;

-- name: ListUserMemberships :many
SELECT
  uo.id,
  uo.user_id,
  uo.organization_id,
  uo.role_id,
  uo.is_default,
  o.name AS organization_name,
  r.name AS role_name
FROM user_organizations uo
JOIN organizations o ON o.id = uo.organization_id
JOIN custom_roles r ON r.id = uo.role_id
WHERE uo.user_id = @user_id::uuid
  AND uo.deleted_at IS NULL
  AND o.deleted_at IS NULL
  AND r.deleted_at IS NULL
ORDER BY uo.is_default DESC, o.name ASC;

-- name: GetMembershipByUserAndOrg :one
SELECT
  uo.id,
  uo.user_id,
  uo.organization_id,
  uo.role_id,
  uo.is_default,
  o.name AS organization_name,
  r.name AS role_name
FROM user_organizations uo
JOIN organizations o ON o.id = uo.organization_id
JOIN custom_roles r ON r.id = uo.role_id
WHERE uo.user_id = @user_id::uuid
  AND uo.organization_id = @organization_id::uuid
  AND uo.deleted_at IS NULL
  AND o.deleted_at IS NULL
  AND r.deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = now(), updated_at = now()
WHERE id = @id::uuid;

-- name: UpdateUserSettings :exec
UPDATE users
SET settings = @settings::jsonb, updated_at = now()
WHERE id = @id::uuid;

-- name: UpdateUserAvailability :exec
UPDATE users
SET availability_status = @availability_status::text, updated_at = now()
WHERE id = @id::uuid;

