package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"encanto/shared"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	orgAlphaID = "11111111-1111-1111-1111-111111111111"
	orgBetaID  = "22222222-2222-2222-2222-222222222222"

	roleAlphaAdminID    = "11111111-aaaa-1111-aaaa-111111111111"
	roleAlphaReadOnlyID = "11111111-bbbb-1111-bbbb-111111111111"
	roleAlphaScopedID   = "11111111-cccc-1111-cccc-111111111111"
	roleBetaReadOnlyID  = "22222222-bbbb-2222-bbbb-222222222222"

	adminUserID    = "aaaaaaaa-1111-1111-1111-111111111111"
	internalUserID = "bbbbbbbb-2222-2222-2222-222222222222"
	readOnlyUserID = "cccccccc-3333-3333-3333-333333333333"
	scopedUserID   = "dddddddd-4444-4444-4444-444444444444"

	alphaSupportInstanceID = "eeeeeeee-1111-1111-1111-111111111111"
	alphaVipInstanceID     = "eeeeeeee-2222-2222-2222-222222222222"
	betaRetailInstanceID   = "eeeeeeee-3333-3333-3333-333333333333"

	chatAliceID = "f1111111-1111-1111-1111-111111111111"
	chatBobID   = "f2222222-2222-2222-2222-222222222222"
	chatCoraID  = "f3333333-3333-3333-3333-333333333333"
	chatDanaID  = "f4444444-4444-4444-4444-444444444444"
)

func SeedDevData(ctx context.Context, pool *pgxpool.Pool, catalog shared.PermissionCatalog) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, statement := range []string{
		`INSERT INTO organizations (id, name, slug, timezone)
		 VALUES ($1, $2, $3, 'Africa/Cairo')
		 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, slug = EXCLUDED.slug`,
	} {
		_ = statement
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organizations (id, name, slug, timezone)
		VALUES
			($1, 'Encanto Alpha', 'encanto-alpha', 'Africa/Cairo'),
			($2, 'Encanto Beta', 'encanto-beta', 'Africa/Cairo')
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name, slug = EXCLUDED.slug, timezone = EXCLUDED.timezone
	`, orgAlphaID, orgBetaID); err != nil {
		return fmt.Errorf("seed organizations: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_configs (organization_id, tenant_status, uploads_cleanup_timezone)
		VALUES
			($1, 'active', 'Africa/Cairo'),
			($2, 'active', 'Africa/Cairo')
		ON CONFLICT (organization_id) DO UPDATE
		SET tenant_status = EXCLUDED.tenant_status,
		    uploads_cleanup_timezone = EXCLUDED.uploads_cleanup_timezone
	`, orgAlphaID, orgBetaID); err != nil {
		return fmt.Errorf("seed org configs: %w", err)
	}

	for _, permission := range catalog.Permissions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO permissions (key, resource, action, label, description)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (key) DO UPDATE
			SET resource = EXCLUDED.resource,
			    action = EXCLUDED.action,
			    label = EXCLUDED.label,
			    description = EXCLUDED.description
		`, permission.Key, permission.Resource, permission.Action, permission.Label, permission.Description); err != nil {
			return fmt.Errorf("seed permission %s: %w", permission.Key, err)
		}
	}

	if err := seedRole(ctx, tx, roleAlphaAdminID, orgAlphaID, "Admin", "Full organization administrator", true, true, "all_contacts", nil, nil, true); err != nil {
		return err
	}
	if err := seedRole(ctx, tx, roleAlphaReadOnlyID, orgAlphaID, "Read Only", "View-only operator", false, false, "all_contacts", nil, nil, true); err != nil {
		return err
	}
	if err := seedRole(ctx, tx, roleAlphaScopedID, orgAlphaID, "Scoped Agent", "Limited by explicit allowed phone numbers", false, false, "allowed_numbers_only", nil, []string{"+201000000001", "+201000000003"}, false); err != nil {
		return err
	}
	if err := seedRole(ctx, tx, roleBetaReadOnlyID, orgBetaID, "Read Only", "Beta organization read-only operator", false, true, "all_contacts", nil, nil, true); err != nil {
		return err
	}

	adminPermissions := make([]string, 0, len(catalog.Permissions))
	for _, permission := range catalog.Permissions {
		adminPermissions = append(adminPermissions, permission.Key)
	}
	readOnlyPermissions := []string{
		"chats.view", "messages.view", "notes.view", "contacts.view", "contacts.scope.all",
	}
	scopedPermissions := []string{
		"chats.view", "messages.view", "messages.send", "notes.view", "contacts.view", "contacts.scope.allowed_numbers",
	}

	rolePermissions := map[string][]string{
		roleAlphaAdminID:    adminPermissions,
		roleAlphaReadOnlyID: readOnlyPermissions,
		roleAlphaScopedID:   scopedPermissions,
		roleBetaReadOnlyID:  readOnlyPermissions,
	}
	for roleID, permissions := range rolePermissions {
		if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE custom_role_id = $1`, roleID); err != nil {
			return fmt.Errorf("clear role permissions: %w", err)
		}
		for _, key := range permissions {
			if _, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (custom_role_id, permission_id)
				SELECT $1, id FROM permissions WHERE key = $2
				ON CONFLICT DO NOTHING
			`, roleID, key); err != nil {
				return fmt.Errorf("seed role permission %s/%s: %w", roleID, key, err)
			}
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, avatar_url, is_active, is_super_admin, availability_status, settings)
		VALUES
			($1, 'admin@example.com', $5, 'Alpha Admin', 'https://api.dicebear.com/9.x/identicon/svg?seed=admin', true, true, 'available', '{"theme":"light","language":"en","sidebarPinned":true}'),
			($2, 'internal@example.com', $5, 'Internal Switcher', 'https://api.dicebear.com/9.x/identicon/svg?seed=internal', true, false, 'busy', '{"theme":"dark","language":"en","sidebarPinned":false}'),
			($3, 'readonly@example.com', $5, 'Read Only Agent', 'https://api.dicebear.com/9.x/identicon/svg?seed=readonly', true, false, 'available', '{"theme":"light","language":"ar","sidebarPinned":true}'),
			($4, 'scoped@example.com', $5, 'Scoped Agent', 'https://api.dicebear.com/9.x/identicon/svg?seed=scoped', true, false, 'available', '{"theme":"system","language":"en","sidebarPinned":true}')
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email,
		    password_hash = EXCLUDED.password_hash,
		    full_name = EXCLUDED.full_name,
		    avatar_url = EXCLUDED.avatar_url,
		    availability_status = EXCLUDED.availability_status,
		    settings = EXCLUDED.settings
	`, adminUserID, internalUserID, readOnlyUserID, scopedUserID, string(passwordHash)); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_organizations (id, user_id, organization_id, role_id, is_default)
		VALUES
			(gen_random_uuid(), $1, $5, $9, true),
			(gen_random_uuid(), $2, $5, $9, true),
			(gen_random_uuid(), $2, $6, $10, false),
			(gen_random_uuid(), $3, $5, $7, true),
			(gen_random_uuid(), $4, $5, $8, true)
		ON CONFLICT DO NOTHING
	`, adminUserID, internalUserID, readOnlyUserID, scopedUserID, orgAlphaID, orgBetaID, roleAlphaReadOnlyID, roleAlphaScopedID, roleAlphaAdminID, roleBetaReadOnlyID); err != nil {
		return fmt.Errorf("seed memberships: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM user_permission_overrides
		WHERE organization_id IN ($1, $2)
	`, orgAlphaID, orgBetaID); err != nil {
		return fmt.Errorf("reset user permission overrides: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM user_contact_visibility_rules
		WHERE organization_id IN ($1, $2)
	`, orgAlphaID, orgBetaID); err != nil {
		return fmt.Errorf("reset user visibility rules: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_contact_visibility_rules (
			organization_id,
			user_id,
			scope_mode,
			allowed_instance_ids,
			allowed_phone_numbers,
			inherit_role_scope,
			can_view_unmasked_phone
		)
		VALUES (
			$1,
			$2,
			'allowed_numbers_only',
			'[]'::jsonb,
			$3::jsonb,
			false,
			false
		)
		ON CONFLICT (organization_id, user_id) DO UPDATE
		SET scope_mode = EXCLUDED.scope_mode,
		    allowed_instance_ids = EXCLUDED.allowed_instance_ids,
		    allowed_phone_numbers = EXCLUDED.allowed_phone_numbers,
		    inherit_role_scope = EXCLUDED.inherit_role_scope,
		    can_view_unmasked_phone = EXCLUDED.can_view_unmasked_phone,
		    updated_at = now()
	`, orgAlphaID, scopedUserID, mustJSONString([]string{"+201000000001", "+201000000003"})); err != nil {
		return fmt.Errorf("seed scoped visibility rule: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO whatsapp_instances (id, organization_id, name, phone_number, status, source_label)
		VALUES
			($1, $4, 'Alpha Support', '+200000000001', 'connected', 'support'),
			($2, $4, 'Alpha VIP', '+200000000002', 'connected', 'vip'),
			($3, $5, 'Beta Retail', '+200000000003', 'connected', 'retail')
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    phone_number = EXCLUDED.phone_number,
		    status = EXCLUDED.status,
		    source_label = EXCLUDED.source_label
	`, alphaSupportInstanceID, alphaVipInstanceID, betaRetailInstanceID, orgAlphaID, orgBetaID); err != nil {
		return fmt.Errorf("seed instances: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO contacts (
			id, organization_id, phone_number, name, assigned_user_id, instance_id, status,
			last_message_preview, last_message_at, tags
		)
		VALUES
			($1, $5, '+201000000001', 'Alice Scope', $6, $9, 'assigned', 'Need invoice copy', now() - interval '30 minutes', '["billing"]'),
			($2, $5, '+201000000002', 'Bob Pending', NULL, $10, 'pending', 'Waiting for assignment', now() - interval '2 hours', '["vip"]'),
			($3, $5, '+201000000003', 'Cora Allowed', $8, $10, 'assigned', 'Requested delivery window', now() - interval '90 minutes', '["priority"]'),
			($4, $7, '+202000000001', 'Dana Beta', $6, $11, 'assigned', 'Store follow-up needed', now() - interval '1 hour', '["retail"]')
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    assigned_user_id = EXCLUDED.assigned_user_id,
		    instance_id = EXCLUDED.instance_id,
		    status = EXCLUDED.status,
		    last_message_preview = EXCLUDED.last_message_preview,
		    last_message_at = EXCLUDED.last_message_at,
		    tags = EXCLUDED.tags
	`, chatAliceID, chatBobID, chatCoraID, chatDanaID, orgAlphaID, adminUserID, orgBetaID, scopedUserID, alphaSupportInstanceID, alphaVipInstanceID, betaRetailInstanceID); err != nil {
		return fmt.Errorf("seed contacts: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, organization_id, contact_id, direction, type, body, status, sent_by_user_id)
		VALUES
			(gen_random_uuid(), $1, $2, 'inbound', 'text', 'Hello, I need my invoice.', 'sent', NULL),
			(gen_random_uuid(), $1, $2, 'outbound', 'text', 'Sure, I will send it shortly.', 'sent', $3),
			(gen_random_uuid(), $1, $4, 'inbound', 'text', 'Can I still reply before this gets claimed?', 'sent', NULL),
			(gen_random_uuid(), $1, $5, 'inbound', 'text', 'Please reserve a late delivery slot.', 'sent', NULL),
			(gen_random_uuid(), $6, $7, 'inbound', 'text', 'When does the retail order arrive?', 'sent', NULL)
		ON CONFLICT DO NOTHING
	`, orgAlphaID, chatAliceID, adminUserID, chatBobID, chatCoraID, orgBetaID, chatDanaID); err != nil {
		return fmt.Errorf("seed messages: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO conversation_notes (id, organization_id, contact_id, author_user_id, body)
		VALUES
			(gen_random_uuid(), $1, $2, $3, 'Customer asked for finance follow-up.'),
			(gen_random_uuid(), $1, $4, $5, 'Visible to scoped user because number is explicitly allowed.'),
			(gen_random_uuid(), $6, $7, $3, 'Beta org note for switch-org verification.')
		ON CONFLICT DO NOTHING
	`, orgAlphaID, chatAliceID, adminUserID, chatCoraID, scopedUserID, orgBetaID, chatDanaID); err != nil {
		return fmt.Errorf("seed notes: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO contact_user_states (organization_id, contact_id, user_id, is_hidden, is_pinned)
		VALUES
			($1, $2, $3, false, true),
			($1, $4, $5, false, true)
		ON CONFLICT (contact_id, user_id) DO UPDATE
		SET is_hidden = EXCLUDED.is_hidden,
		    is_pinned = EXCLUDED.is_pinned,
		    updated_at = now()
	`, orgAlphaID, chatAliceID, adminUserID, chatCoraID, scopedUserID); err != nil {
		return fmt.Errorf("seed contact user states: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	return nil
}

func ResetAndSeed(ctx context.Context, pool *pgxpool.Pool, catalog shared.PermissionCatalog) error {
	if _, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			contact_user_states,
			conversation_notes,
			messages,
			contacts,
			whatsapp_instances,
			user_contact_visibility_rules,
			user_permission_overrides,
			user_organizations,
			users,
			role_permissions,
			custom_roles,
			permissions,
			organization_configs,
			organizations
		RESTART IDENTITY CASCADE
	`); err != nil {
		return fmt.Errorf("reset seed data: %w", err)
	}
	return SeedDevData(ctx, pool, catalog)
}

func seedRole(ctx context.Context, tx pgx.Tx, roleID, organizationID, name, description string, isSystem, isDefault bool, scopeMode string, allowedInstanceIDs []string, allowedPhoneNumbers []string, canViewUnmaskedPhone bool) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO custom_roles (
			id,
			organization_id,
			name,
			description,
			is_system,
			is_default,
			default_scope_mode,
			default_allowed_instance_ids,
			default_allowed_phone_numbers,
			can_view_unmasked_phone
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    description = EXCLUDED.description,
		    is_system = EXCLUDED.is_system,
		    is_default = EXCLUDED.is_default,
		    default_scope_mode = EXCLUDED.default_scope_mode,
		    default_allowed_instance_ids = EXCLUDED.default_allowed_instance_ids,
		    default_allowed_phone_numbers = EXCLUDED.default_allowed_phone_numbers,
		    can_view_unmasked_phone = EXCLUDED.can_view_unmasked_phone,
		    updated_at = now()
	`, roleID, organizationID, name, description, isSystem, isDefault, scopeMode, mustJSONString(allowedInstanceIDs), mustJSONString(allowedPhoneNumbers), canViewUnmaskedPhone); err != nil {
		return fmt.Errorf("seed role %s: %w", name, err)
	}
	return nil
}

func mustJSONString(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			trimmed = append(trimmed, value)
		}
	}
	encoded, _ := json.Marshal(trimmed)
	return string(encoded)
}
