-- Whatomate Production Schema (Planning Baseline)
-- SQL-first schema aligned with pgx/sqlc after live UI audit.
-- Live audit retained statuses, notifications, chatbot-adjacent flow config,
-- dashboard, analytics, campaigns, and rich per-instance operations in planning scope.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. IDENTITY & ACCESS
-- ============================================================================

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    logo_url TEXT,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT true,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations (deleted_at);

CREATE TABLE IF NOT EXISTS organization_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    worker_count INTEGER NOT NULL DEFAULT 0,
    max_queue_size INTEGER NOT NULL DEFAULT 250 CHECK (max_queue_size >= 0),
    max_concurrent_jobs INTEGER NOT NULL DEFAULT 2 CHECK (max_concurrent_jobs >= 0),
    max_users INTEGER NOT NULL DEFAULT 5 CHECK (max_users > 0 AND max_users <= 5),
    max_whatsapp_instances INTEGER NOT NULL DEFAULT 0 CHECK (max_whatsapp_instances >= 0),
    storage_quota_bytes BIGINT NOT NULL DEFAULT 5368709120 CHECK (storage_quota_bytes > 0),
    storage_used_bytes BIGINT NOT NULL DEFAULT 0 CHECK (storage_used_bytes >= 0),
    tenant_status TEXT NOT NULL DEFAULT 'active',
    queue_backpressure_mode TEXT NOT NULL DEFAULT 'defer',
    uploads_cleanup_retention_days INTEGER NOT NULL DEFAULT 30,
    uploads_cleanup_hour SMALLINT NOT NULL DEFAULT 3,
    uploads_cleanup_timezone TEXT,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    CHECK (tenant_status IN ('active', 'read_only', 'suspended')),
    CHECK (queue_backpressure_mode IN ('defer', 'reject_new', 'pause'))
);

CREATE TABLE IF NOT EXISTS slot_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL,
    total_slots INTEGER NOT NULL CHECK (total_slots >= 0),
    reserved_slots INTEGER NOT NULL DEFAULT 0 CHECK (reserved_slots >= 0),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (resource_type)
);

CREATE TABLE IF NOT EXISTS organization_slot_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    slot_inventory_id UUID NOT NULL REFERENCES slot_inventory(id) ON DELETE CASCADE,
    allocated_slots INTEGER NOT NULL DEFAULT 0 CHECK (allocated_slots >= 0),
    used_slots INTEGER NOT NULL DEFAULT 0 CHECK (used_slots >= 0),
    last_reconciled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (organization_id, slot_inventory_id),
    CHECK (used_slots <= allocated_slots)
);
CREATE INDEX IF NOT EXISTS idx_org_slot_allocations_slot_inventory
    ON organization_slot_allocations (slot_inventory_id, organization_id);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT,
    UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    is_default BOOLEAN NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_roles_org_name
    ON custom_roles (organization_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS role_permissions (
    custom_role_id UUID NOT NULL REFERENCES custom_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (custom_role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT,
    role_id UUID REFERENCES custom_roles(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_available BOOLEAN NOT NULL DEFAULT true,
    is_super_admin BOOLEAN NOT NULL DEFAULT false,
    avatar_url TEXT,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_login TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_org_active
    ON users (organization_id, is_active) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id UUID REFERENCES custom_roles(id),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_org_unique
    ON user_organizations(user_id, organization_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_org_default_unique
    ON user_organizations(user_id) WHERE is_default = true AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    last_used TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT,
    allow_auto_create BOOLEAN NOT NULL DEFAULT false,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (organization_id, provider)
);

-- ============================================================================
-- 2. OPERATIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    description TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_org_name
    ON teams (organization_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_availability_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    previous_state BOOLEAN,
    new_state BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_notifications_user_read_created
    ON user_notifications (user_id, is_read, created_at DESC);

-- ============================================================================
-- 3. WHATSAPP INSTANCES & STATUSES
-- ============================================================================

CREATE TABLE IF NOT EXISTS whatsapp_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    jid TEXT DEFAULT '',
    phone_number TEXT,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'disconnected',
    pairing_state TEXT NOT NULL DEFAULT 'unpaired',
    qr_code TEXT,
    session_data TEXT,
    send_blocked_until TIMESTAMPTZ,
    send_block_reason TEXT NOT NULL DEFAULT '',
    last_connected_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_whatsapp_instances_jid
    ON whatsapp_instances (jid) WHERE jid <> '';

CREATE TABLE IF NOT EXISTS instance_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS whatsapp_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    media_asset_id UUID,
    kind TEXT NOT NULL DEFAULT 'image',
    body TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    published_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_whatsapp_statuses_org_published_at
    ON whatsapp_statuses (organization_id, published_at DESC)
    WHERE deleted_at IS NULL;

-- ============================================================================
-- 4. MESSAGING
-- ============================================================================

CREATE TABLE IF NOT EXISTS media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size BIGINT CHECK (file_size IS NULL OR file_size >= 0),
    file_hash TEXT,
    source TEXT NOT NULL DEFAULT 'upload',
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_media_assets_org_created_at
    ON media_assets (organization_id, created_at DESC);

CREATE TABLE IF NOT EXISTS whatsapp_instance_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL UNIQUE REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    auto_sync_history BOOLEAN NOT NULL DEFAULT true,
    auto_download_incoming_media BOOLEAN NOT NULL DEFAULT true,
    source_tag_label TEXT,
    source_tag_display_mode TEXT NOT NULL DEFAULT 'instance_name',
    source_tag_color TEXT NOT NULL DEFAULT 'emerald',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS whatsapp_instance_health_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'disconnected',
    uptime_seconds BIGINT NOT NULL DEFAULT 0,
    queue_depth INTEGER NOT NULL DEFAULT 0,
    sent_today INTEGER NOT NULL DEFAULT 0,
    received_today INTEGER NOT NULL DEFAULT 0,
    failed_today INTEGER NOT NULL DEFAULT 0,
    error_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    observed_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_instance_health_snapshots_instance_observed
    ON whatsapp_instance_health_snapshots (instance_id, observed_at DESC);

CREATE TABLE IF NOT EXISTS instance_call_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL UNIQUE REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    reject_individual_calls BOOLEAN NOT NULL DEFAULT true,
    reject_group_calls BOOLEAN NOT NULL DEFAULT true,
    reply_mode TEXT NOT NULL DEFAULT 'reject_without_message',
    schedule_mode TEXT NOT NULL DEFAULT 'always_on',
    emergency_bypass_contacts TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    reply_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS instance_auto_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL UNIQUE REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    campaign_name_prefix TEXT,
    schedule_every_days INTEGER NOT NULL DEFAULT 7,
    delay_from_minutes INTEGER NOT NULL DEFAULT 1,
    delay_to_minutes INTEGER NOT NULL DEFAULT 3,
    campaign_status TEXT NOT NULL DEFAULT 'draft',
    message_body TEXT,
    media_asset_id UUID REFERENCES media_assets(id) ON DELETE SET NULL,
    last_evaluated_at TIMESTAMPTZ,
    next_evaluation_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS instance_rating_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL UNIQUE REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    follow_up_window_minutes INTEGER NOT NULL DEFAULT 15,
    template_ar TEXT,
    template_en TEXT,
    template_es TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS instance_assignment_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL UNIQUE REFERENCES whatsapp_instances(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    schedule_mode TEXT NOT NULL DEFAULT 'midnight',
    timezone TEXT,
    last_reset_at TIMESTAMPTZ,
    next_reset_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_whatsapp_statuses_media_asset'
    ) THEN
        ALTER TABLE whatsapp_statuses
            ADD CONSTRAINT fk_whatsapp_statuses_media_asset
            FOREIGN KEY (media_asset_id) REFERENCES media_assets(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    phone_number TEXT NOT NULL,
    name TEXT,
    avatar_url TEXT,
    assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    instance_id UUID REFERENCES whatsapp_instances(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    is_public BOOLEAN NOT NULL DEFAULT false,
    is_read BOOLEAN NOT NULL DEFAULT true,
    instance_source_label TEXT,
    last_message_preview TEXT,
    last_message_at TIMESTAMPTZ,
    last_inbound_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    closed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_contacts_org_phone_instance
    ON contacts (organization_id, phone_number, instance_id)
    WHERE instance_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contacts_tags ON contacts USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_contacts_deleted_at ON contacts (deleted_at);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    whatsapp_message_id TEXT,
    type TEXT NOT NULL DEFAULT 'text',
    body TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    direction TEXT NOT NULL,
    media_asset_id UUID REFERENCES media_assets(id) ON DELETE SET NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_messages_contact_created_at
    ON messages (contact_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_org_whatsapp_id
    ON messages (organization_id, whatsapp_message_id)
    WHERE whatsapp_message_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    color TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (organization_id, name)
);

CREATE TABLE IF NOT EXISTS contact_user_deletions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (contact_id, user_id)
);

CREATE TABLE IF NOT EXISTS contact_collaborators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'invited',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (contact_id, user_id)
);

CREATE TABLE IF NOT EXISTS conversation_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS chat_closure_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    score INTEGER NOT NULL,
    comment TEXT,
    closing_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    rated_at TIMESTAMPTZ DEFAULT now(),
    rating_message_snapshot TEXT,
    context_messages_snapshot TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS canned_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    title TEXT NOT NULL,
    shortcut TEXT,
    body TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- ============================================================================
-- 5. CHATBOT-ADJACENT CONFIGURATION
-- ============================================================================

CREATE TABLE IF NOT EXISTS chatbot_flows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    panel_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_chatbot_flows_org_name
    ON chatbot_flows (organization_id, name) WHERE deleted_at IS NULL;

-- ============================================================================
-- 6. EXTENSIBILITY & LICENSE
-- ============================================================================

CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    secret TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS custom_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS license_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    activation_token_encrypted TEXT NOT NULL,
    license_family_id TEXT NOT NULL,
    license_id TEXT NOT NULL,
    revision BIGINT NOT NULL DEFAULT 0,
    key_id TEXT NOT NULL,
    issuer TEXT NOT NULL,
    audience TEXT NOT NULL,
    product TEXT NOT NULL DEFAULT 'whatomate',
    hwid_full TEXT NOT NULL,
    hwid_hash TEXT NOT NULL,
    tier TEXT NOT NULL DEFAULT '',
    license_kind TEXT NOT NULL,
    trial_days INTEGER NOT NULL DEFAULT 0,
    max_organizations INTEGER NOT NULL DEFAULT 0,
    max_users_per_org INTEGER NOT NULL DEFAULT 0,
    max_whatsapp_endpoints_per_org INTEGER NOT NULL DEFAULT 0,
    max_workers INTEGER NOT NULL DEFAULT 0,
    max_workers_per_org INTEGER NOT NULL DEFAULT 0,
    max_storage_bytes_per_org BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    overages JSONB NOT NULL DEFAULT '{}'::jsonb,
    issued_at TIMESTAMPTZ NOT NULL,
    not_before TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    grace_deadline TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ NOT NULL,
    activated_at TIMESTAMPTZ NOT NULL,
    integrity_hmac TEXT NOT NULL,
    UNIQUE (license_id),
    UNIQUE (license_family_id, revision),
    UNIQUE (hwid_hash)
);
CREATE INDEX IF NOT EXISTS idx_license_records_status ON license_records (status);
CREATE INDEX IF NOT EXISTS idx_license_records_hwid_hash ON license_records (hwid_hash);

CREATE TABLE IF NOT EXISTS license_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    reason TEXT,
    status TEXT,
    license_family_id TEXT,
    license_id TEXT,
    hwid_hash TEXT,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_license_events_created_at ON license_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_license_events_license_id ON license_events (license_id);
CREATE INDEX IF NOT EXISTS idx_license_events_family_id ON license_events (license_family_id);
