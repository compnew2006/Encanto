-- Encanto Database Schema
-- Migration 001: Initial Schema

-- Organizations (multi-tenant root)
CREATE TABLE IF NOT EXISTS organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    date_format TEXT NOT NULL DEFAULT 'DD MMM YYYY',
    locale      TEXT NOT NULL DEFAULT 'en',
    mask_phone_numbers   BOOLEAN NOT NULL DEFAULT false,
    tenant_status        TEXT NOT NULL DEFAULT 'active',
    max_members          INT NOT NULL DEFAULT 10,
    max_instances        INT NOT NULL DEFAULT 5,
    used_instances       INT NOT NULL DEFAULT 0,
    storage_used_label   TEXT NOT NULL DEFAULT '0 MB',
    storage_limit_label  TEXT NOT NULL DEFAULT '5 GiB',
    appearance           JSONB NOT NULL DEFAULT '{}',
    chat_settings        JSONB NOT NULL DEFAULT '{}',
    notification_settings JSONB NOT NULL DEFAULT '{}',
    cleanup_settings     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    name            TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'agent',
    status          TEXT NOT NULL DEFAULT 'online',
    avatar          TEXT NOT NULL DEFAULT '',
    language        TEXT NOT NULL DEFAULT 'en',
    theme_preset    TEXT NOT NULL DEFAULT 'ocean-breeze',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, email)
);

-- WhatsApp Instances
CREATE TABLE IF NOT EXISTS whatsapp_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    phone_number    TEXT NOT NULL DEFAULT '',
    jid             TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'disconnected',
    pairing_state   TEXT NOT NULL DEFAULT 'needs_qr',
    qr_code         TEXT NOT NULL DEFAULT '',
    slot_blocked    BOOLEAN NOT NULL DEFAULT false,
    settings        JSONB NOT NULL DEFAULT '{}',
    call_policy     JSONB NOT NULL DEFAULT '{}',
    auto_campaign   JSONB NOT NULL DEFAULT '{}',
    rating_settings JSONB NOT NULL DEFAULT '{}',
    assignment_reset JSONB NOT NULL DEFAULT '{}',
    -- Health metrics
    health_status        TEXT NOT NULL DEFAULT 'disconnected',
    health_uptime_label  TEXT NOT NULL DEFAULT '0m',
    health_queue_depth   INT NOT NULL DEFAULT 0,
    health_sent_today    INT NOT NULL DEFAULT 0,
    health_received_today INT NOT NULL DEFAULT 0,
    health_failed_today  INT NOT NULL DEFAULT 0,
    health_error_rate    TEXT NOT NULL DEFAULT '0.0%',
    health_observed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contacts (conversations)
CREATE TABLE IF NOT EXISTS contacts (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    instance_id          UUID NOT NULL REFERENCES whatsapp_instances(id),
    name                 TEXT NOT NULL,
    phone_number         TEXT NOT NULL,
    avatar               TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'pending',
    assigned_user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_user_name   TEXT NOT NULL DEFAULT '',
    instance_name        TEXT NOT NULL DEFAULT '',
    instance_source_label TEXT NOT NULL DEFAULT '',
    last_message_preview TEXT NOT NULL DEFAULT '',
    last_message_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_inbound_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at            TIMESTAMPTZ,
    is_public            BOOLEAN NOT NULL DEFAULT true,
    is_read              BOOLEAN NOT NULL DEFAULT false,
    unread_count         INT NOT NULL DEFAULT 0,
    tags                 JSONB NOT NULL DEFAULT '[]',
    metadata             JSONB NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Per-user contact state (pinned, hidden, last-read)
CREATE TABLE IF NOT EXISTS contact_user_states (
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id          UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    is_pinned           BOOLEAN NOT NULL DEFAULT false,
    is_hidden           BOOLEAN NOT NULL DEFAULT false,
    last_read_message_id TEXT NOT NULL DEFAULT '',
    last_opened_at      TIMESTAMPTZ,
    last_seen_at        TIMESTAMPTZ,
    PRIMARY KEY(user_id, contact_id)
);

-- Messages
CREATE TABLE IF NOT EXISTS messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id      UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    direction       TEXT NOT NULL DEFAULT 'inbound',
    type            TEXT NOT NULL DEFAULT 'text',
    body            TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'sent',
    file_name       TEXT NOT NULL DEFAULT '',
    file_size_label TEXT NOT NULL DEFAULT '',
    media_url       TEXT NOT NULL DEFAULT '',
    failure_reason  TEXT NOT NULL DEFAULT '',
    retry_count     INT NOT NULL DEFAULT 0,
    typed_for_ms    INT NOT NULL DEFAULT 0,
    reaction        TEXT NOT NULL DEFAULT '',
    revoked_at      TIMESTAMPTZ,
    can_retry       BOOLEAN NOT NULL DEFAULT false,
    can_revoke      BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Conversation notes
CREATE TABLE IF NOT EXISTS conversation_notes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name   TEXT NOT NULL DEFAULT '',
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contact collaborators
CREATE TABLE IF NOT EXISTS contact_collaborators (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name   TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'invited',
    invited_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Timeline events
CREATE TABLE IF NOT EXISTS timeline_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id      UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type      TEXT NOT NULL,
    actor_user_id   TEXT NOT NULL DEFAULT '',
    actor_name      TEXT NOT NULL DEFAULT '',
    summary         TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}',
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Quick replies
CREATE TABLE IF NOT EXISTS quick_replies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shortcut        TEXT NOT NULL,
    title           TEXT NOT NULL,
    body            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User notifications
CREATE TABLE IF NOT EXISTS user_notifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title               TEXT NOT NULL,
    body                TEXT NOT NULL,
    severity            TEXT NOT NULL DEFAULT 'info',
    related_contact_id  TEXT NOT NULL DEFAULT '',
    related_path        TEXT NOT NULL DEFAULT '',
    is_read             BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Status posts (WhatsApp statuses seen)
CREATE TABLE IF NOT EXISTS status_posts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id      TEXT NOT NULL DEFAULT '',
    contact_name    TEXT NOT NULL DEFAULT '',
    instance_id     TEXT NOT NULL DEFAULT '',
    instance_name   TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL,
    kind            TEXT NOT NULL DEFAULT 'text',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Campaigns
CREATE TABLE IF NOT EXISTS campaigns (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'draft',
    source              TEXT NOT NULL DEFAULT 'manual',
    linked_instance_id  TEXT NOT NULL DEFAULT '',
    content             TEXT NOT NULL DEFAULT '',
    filters             JSONB NOT NULL DEFAULT '{}',
    schedule            JSONB NOT NULL DEFAULT '{}',
    last_run_summary    TEXT NOT NULL DEFAULT 'No runs yet',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Campaign runs
CREATE TABLE IF NOT EXISTS campaign_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id     UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trigger         TEXT NOT NULL DEFAULT 'manual',
    status          TEXT NOT NULL DEFAULT 'running',
    job_id          TEXT NOT NULL DEFAULT '',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ,
    recipient_total INT NOT NULL DEFAULT 0,
    delivered       INT NOT NULL DEFAULT 0,
    failed          INT NOT NULL DEFAULT 0
);

-- Campaign recipients
CREATE TABLE IF NOT EXISTS campaign_recipients (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID NOT NULL REFERENCES campaign_runs(id) ON DELETE CASCADE,
    contact_id      TEXT NOT NULL DEFAULT '',
    contact_name    TEXT NOT NULL DEFAULT '',
    phone_number    TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',
    failure_reason  TEXT NOT NULL DEFAULT '',
    message_preview TEXT NOT NULL DEFAULT '',
    delivered_at    TIMESTAMPTZ
);

-- Background jobs
CREATE TABLE IF NOT EXISTS background_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    kind            TEXT NOT NULL,
    entity_type     TEXT NOT NULL DEFAULT '',
    entity_id       TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'running',
    summary         TEXT NOT NULL DEFAULT '',
    failure_reason  TEXT NOT NULL DEFAULT '',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

-- Webhook endpoints
CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    target_url      TEXT NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Webhook deliveries
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id      UUID NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_id        TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',
    attempt         INT NOT NULL DEFAULT 1,
    last_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_retry_at   TIMESTAMPTZ,
    response_code   INT NOT NULL DEFAULT 0,
    response_body   TEXT NOT NULL DEFAULT ''
);

-- Outbox events
CREATE TABLE IF NOT EXISTS outbox_events (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type           TEXT NOT NULL,
    entity_type          TEXT NOT NULL DEFAULT '',
    entity_id            TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'delivered',
    payload              JSONB NOT NULL DEFAULT '{}',
    delivery_count       INT NOT NULL DEFAULT 0,
    last_delivery_status TEXT NOT NULL DEFAULT 'delivered',
    occurred_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id   TEXT NOT NULL DEFAULT '',
    actor_name      TEXT NOT NULL DEFAULT '',
    action          TEXT NOT NULL,
    entity_type     TEXT NOT NULL DEFAULT '',
    entity_id       TEXT NOT NULL DEFAULT '',
    summary         TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}',
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Customer ratings
CREATE TABLE IF NOT EXISTS customer_ratings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id      TEXT NOT NULL DEFAULT '',
    contact_name    TEXT NOT NULL DEFAULT '',
    phone_number    TEXT NOT NULL DEFAULT '',
    agent_user_id   TEXT NOT NULL DEFAULT '',
    agent_name      TEXT NOT NULL DEFAULT '',
    score           INT NOT NULL DEFAULT 0,
    message         TEXT NOT NULL DEFAULT '',
    rated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    chat_path       TEXT NOT NULL DEFAULT '',
    source_event_id TEXT NOT NULL DEFAULT ''
);

-- License records (one per org)
CREATE TABLE IF NOT EXISTS license_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'active',
    hwid            TEXT NOT NULL DEFAULT '',
    short_id        TEXT NOT NULL DEFAULT '',
    last_key_hint   TEXT NOT NULL DEFAULT '',
    message         TEXT NOT NULL DEFAULT '',
    activate_url    TEXT NOT NULL DEFAULT '/settings/license',
    cleanup_url     TEXT NOT NULL DEFAULT '/license-cleanup',
    activated_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    max_contacts    INT NOT NULL DEFAULT 100,
    max_campaigns   INT NOT NULL DEFAULT 10,
    max_instances   INT NOT NULL DEFAULT 5,
    tier            TEXT NOT NULL DEFAULT 'growth',
    kind            TEXT NOT NULL DEFAULT 'offline-signed',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_contacts_org_status ON contacts(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_contacts_org_instance ON contacts(organization_id, instance_id);
CREATE INDEX IF NOT EXISTS idx_messages_contact ON messages(contact_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_instances_org ON whatsapp_instances(organization_id);
CREATE INDEX IF NOT EXISTS idx_timeline_contact ON timeline_events(contact_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_org ON user_notifications(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_org ON audit_logs(organization_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_org ON background_jobs(organization_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_campaigns_org ON campaigns(organization_id, updated_at DESC);
