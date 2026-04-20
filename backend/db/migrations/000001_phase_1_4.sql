CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    logo_url TEXT,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT true,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_name_active
    ON organizations (name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS organization_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    tenant_status TEXT NOT NULL DEFAULT 'active',
    storage_quota_bytes BIGINT NOT NULL DEFAULT 5368709120,
    storage_used_bytes BIGINT NOT NULL DEFAULT 0,
    uploads_cleanup_retention_days INTEGER NOT NULL DEFAULT 30,
    uploads_cleanup_hour SMALLINT NOT NULL DEFAULT 3,
    uploads_cleanup_timezone TEXT,
    CHECK (tenant_status IN ('active', 'read_only', 'suspended'))
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT NOT NULL UNIQUE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    label TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT false,
    is_default BOOLEAN NOT NULL DEFAULT false,
    default_scope_mode TEXT NOT NULL DEFAULT 'all_contacts',
    default_allowed_instance_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_allowed_phone_numbers JSONB NOT NULL DEFAULT '[]'::jsonb,
    can_view_unmasked_phone BOOLEAN NOT NULL DEFAULT true,
    CHECK (default_scope_mode IN ('all_contacts', 'instances_only', 'allowed_numbers_only', 'instances_plus_allowed_numbers'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_roles_org_name_active
    ON custom_roles (organization_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS role_permissions (
    custom_role_id UUID NOT NULL REFERENCES custom_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (custom_role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    avatar_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_super_admin BOOLEAN NOT NULL DEFAULT false,
    availability_status TEXT NOT NULL DEFAULT 'available',
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_login TIMESTAMPTZ,
    CHECK (availability_status IN ('available', 'unavailable', 'busy'))
);

CREATE TABLE IF NOT EXISTS user_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES custom_roles(id) ON DELETE RESTRICT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_org_unique_active
    ON user_organizations (user_id, organization_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_org_default_unique_active
    ON user_organizations (user_id) WHERE is_default = true AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user_permission_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_key TEXT NOT NULL,
    mode TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (mode IN ('allow', 'deny')),
    UNIQUE (organization_id, user_id, permission_key)
);

CREATE TABLE IF NOT EXISTS user_contact_visibility_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope_mode TEXT NOT NULL DEFAULT 'all_contacts',
    allowed_instance_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    allowed_phone_numbers JSONB NOT NULL DEFAULT '[]'::jsonb,
    inherit_role_scope BOOLEAN NOT NULL DEFAULT true,
    can_view_unmasked_phone BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (scope_mode IN ('all_contacts', 'instances_only', 'allowed_numbers_only', 'instances_plus_allowed_numbers')),
    UNIQUE (organization_id, user_id)
);

CREATE TABLE IF NOT EXISTS whatsapp_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'connected',
    source_label TEXT,
    UNIQUE (organization_id, phone_number)
);

CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    phone_number TEXT NOT NULL,
    name TEXT,
    avatar_url TEXT,
    assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    instance_id UUID REFERENCES whatsapp_instances(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    last_message_preview TEXT,
    last_message_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    CHECK (status IN ('assigned', 'pending', 'closed'))
);

CREATE INDEX IF NOT EXISTS idx_contacts_org_status_updated
    ON contacts (organization_id, status, updated_at DESC);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    direction TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'text',
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'sent',
    sent_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    CHECK (direction IN ('inbound', 'outbound'))
);

CREATE INDEX IF NOT EXISTS idx_messages_contact_created
    ON messages (contact_id, created_at ASC);

CREATE TABLE IF NOT EXISTS conversation_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    body TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS contact_user_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_hidden BOOLEAN NOT NULL DEFAULT false,
    is_pinned BOOLEAN NOT NULL DEFAULT false,
    last_read_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (contact_id, user_id)
);

