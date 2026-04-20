package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// PGStore is the PostgreSQL-backed implementation of all data access.
type PGStore struct {
	db          *pgxpool.Pool
	waContainer *sqlstore.Container
}

// NewPGStore creates a PGStore and seeds the default org+user if empty.
func NewPGStore(db *pgxpool.Pool) *PGStore {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres@localhost:5432/encanto"
	}
	container, err := sqlstore.New(context.Background(), "pgx", dsn, nil)
	if err != nil {
		log.Printf("ERROR: failed to initialize whatsmeow store: %v", err)
	}

	s := &PGStore{
		db:          db,
		waContainer: container,
	}
	if err := s.seed(); err != nil {
		log.Printf("WARNING: seed failed: %v", err)
	}
	return s
}

func (s *PGStore) ctx() context.Context { return context.Background() }

// ---------- SEED (first run) ----------

func (s *PGStore) seed() error {
	var count int
	if err := s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM organizations`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	// default org
	var orgID string
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO organizations (name, slug, appearance, chat_settings, notification_settings, cleanup_settings)
		VALUES ('My Organization', 'my-org',
			'{"color_mode":"light","theme_preset":"ocean-breeze"}',
			'{"media_grouping_window_minutes":5,"sidebar_contact_view":"comfortable","sidebar_hover_expand":true,"pin_sidebar":true,"chat_background":"paper-grid","show_print_buttons":true,"show_download_buttons":true}',
			'{"email_notifications":true,"new_message_alerts":true,"notification_sound":"soft-bell","campaign_updates":true}',
			'{"retention_days":30,"run_hour":3,"timezone":"UTC","last_job_status":"never-run"}'
		) RETURNING id`).Scan(&orgID)
	if err != nil {
		return fmt.Errorf("seed org: %w", err)
	}

	// default admin user
	_, err = s.db.Exec(s.ctx(), `
		INSERT INTO users (organization_id, email, password_hash, name, role, status)
		VALUES ($1, 'admin@encanto.io', $2, 'Admin', 'admin', 'online')`,
		orgID, string(hash))
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	// default license
	now := time.Now()
	expiresAt := now.Add(365 * 24 * time.Hour)
	_, err = s.db.Exec(s.ctx(), `
		INSERT INTO license_records (organization_id, status, hwid, short_id, last_key_hint, message,
			max_contacts, max_campaigns, max_instances, tier, kind, activated_at, expires_at)
		VALUES ($1,'active','ENCANTO-DEV','ENCDEV','DEMO42','License is active.',
			100, 10, 5, 'growth', 'offline-signed', $2, $3)`,
		orgID, now, expiresAt)
	if err != nil {
		return fmt.Errorf("seed license: %w", err)
	}

	log.Printf("✅ Database seeded: org=%s, login=admin@encanto.io / admin123", orgID)
	return nil
}

// ---------- AUTH ----------

func (s *PGStore) GetUserResponse(orgID string) UserResponse {
	if orgID == "" {
		// pick first org
		_ = s.db.QueryRow(s.ctx(), `SELECT id FROM organizations LIMIT 1`).Scan(&orgID)
	}
	var u UserResponse
	var role, lang, theme string
	err := s.db.QueryRow(s.ctx(), `
		SELECT u.id, u.email, u.name, u.avatar, u.status, u.role, u.language, u.theme_preset
		FROM users u WHERE u.organization_id = $1 LIMIT 1`, orgID).
		Scan(&u.ID, &u.Email, &u.Name, &u.Avatar, &u.Status, &role, &lang, &theme)
	if err != nil {
		return u
	}
	u.Role = role
	u.Settings = UserSettings{Language: lang, ThemePreset: theme}
	orgs := s.listOrgs(u.ID, orgID)
	u.Organizations = orgs
	for _, o := range orgs {
		if o.ID == orgID {
			u.CurrentOrganization = o
			break
		}
	}
	return u
}

func (s *PGStore) GetUserByEmail(email, password string) (UserResponse, error) {
	var userID, orgID, hash, role, name, avatar, status, lang, theme string
	err := s.db.QueryRow(s.ctx(), `
		SELECT u.id, u.organization_id, u.password_hash, u.role, u.name, u.avatar, u.status, u.language, u.theme_preset
		FROM users u WHERE u.email = $1 AND u.is_active = true LIMIT 1`, email).
		Scan(&userID, &orgID, &hash, &role, &name, &avatar, &status, &lang, &theme)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserResponse{}, errors.New("invalid email or password")
		}
		return UserResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return UserResponse{}, errors.New("invalid email or password")
	}

	u := UserResponse{
		ID:       userID,
		Email:    email,
		Name:     name,
		Avatar:   avatar,
		Status:   status,
		Role:     role,
		Settings: UserSettings{Language: lang, ThemePreset: theme},
	}
	orgs := s.listOrgs(userID, orgID)
	u.Organizations = orgs
	for _, o := range orgs {
		if o.ID == orgID {
			u.CurrentOrganization = o
			break
		}
	}
	return u, nil
}

func (s *PGStore) listOrgs(userID, currentOrgID string) []Organization {
	rows, err := s.db.Query(s.ctx(), `
		SELECT o.id, o.name, u.role
		FROM organizations o
		JOIN users u ON u.organization_id = o.id AND u.id = $1
		ORDER BY o.name`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var orgs []Organization
	for rows.Next() {
		var o Organization
		_ = rows.Scan(&o.ID, &o.Name, &o.Role)
		orgs = append(orgs, o)
	}
	if len(orgs) == 0 {
		// fallback: just return the org
		var o Organization
		_ = s.db.QueryRow(s.ctx(), `SELECT id, name FROM organizations WHERE id = $1`, currentOrgID).
			Scan(&o.ID, &o.Name)
		o.Role = "admin"
		orgs = []Organization{o}
	}
	return orgs
}

func (s *PGStore) OrgAccessible(orgID string) bool {
	var count int
	_ = s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM organizations WHERE id = $1`, orgID).Scan(&count)
	return count > 0
}

func (s *PGStore) IsUserAdmin(orgID, userID string) bool {
	var role string
	if userID == "" {
		// Check if any admin exists in this org
		err := s.db.QueryRow(s.ctx(), `SELECT role FROM users WHERE organization_id = $1 ORDER BY role ASC LIMIT 1`, orgID).Scan(&role)
		if err != nil {
			return false
		}
		return role == "admin"
	}
	err := s.db.QueryRow(s.ctx(), `
		SELECT role FROM users WHERE organization_id = $1 AND id = $2`, orgID, userID).Scan(&role)
	if err != nil {
		return false
	}
	return role == "admin"
}

// ---------- SETTINGS ----------

func (s *PGStore) SettingsForOrg(orgID string) (SettingsSummary, error) {
	var sum SettingsSummary
	var appJSON, chatJSON, notifJSON, cleanJSON []byte
	var org generalOrgRow
	err := s.db.QueryRow(s.ctx(), `
		SELECT name, slug, timezone, date_format, locale, mask_phone_numbers,
			tenant_status, max_members, max_instances, used_instances,
			storage_used_label, storage_limit_label,
			appearance, chat_settings, notification_settings, cleanup_settings
		FROM organizations WHERE id = $1`, orgID).
		Scan(&org.name, &org.slug, &org.timezone, &org.dateFormat, &org.locale, &org.maskPhone,
			&org.tenantStatus, &org.maxMembers, &org.maxInstances, &org.usedInstances,
			&org.storageUsed, &org.storageLimit,
			&appJSON, &chatJSON, &notifJSON, &cleanJSON)
	if err != nil {
		return sum, errors.New("organization not found")
	}

	sum.General = GeneralSettings{
		OrganizationName:   org.name,
		Timezone:           org.timezone,
		DateFormat:         org.dateFormat,
		Locale:             org.locale,
		MaskPhoneNumbers:   org.maskPhone,
		TenantStatus:       org.tenantStatus,
		MaxMembers:         org.maxMembers,
		MaxInstances:       org.maxInstances,
		UsedInstances:      org.usedInstances,
		StorageUsedLabel:   org.storageUsed,
		StorageLimitLabel:  org.storageLimit,
	}
	_ = json.Unmarshal(appJSON, &sum.Appearance)
	_ = json.Unmarshal(chatJSON, &sum.Chat)
	_ = json.Unmarshal(notifJSON, &sum.Notifications)
	_ = json.Unmarshal(cleanJSON, &sum.Cleanup)

	// team members from users table
	rows, _ := s.db.Query(s.ctx(), `SELECT id, name, role, email, status, avatar FROM users WHERE organization_id = $1 AND is_active = true ORDER BY name`, orgID)
	defer rows.Close()
	for rows.Next() {
		var m WorkspaceUser
		_ = rows.Scan(&m.ID, &m.Name, &m.Role, &m.Email, &m.Status, &m.Avatar)
		sum.Team = append(sum.Team, m)
	}

	var qr []QuickReply
	qrows, _ := s.db.Query(s.ctx(), `SELECT id, shortcut, title, body FROM quick_replies WHERE organization_id = $1 ORDER BY shortcut`, orgID)
	defer qrows.Close()
	for qrows.Next() {
		var q QuickReply
		_ = qrows.Scan(&q.ID, &q.Shortcut, &q.Title, &q.Body)
		qr = append(qr, q)
	}
	sum.QuickReplies = qr

	return sum, nil
}

type generalOrgRow struct {
	name, slug, timezone, dateFormat, locale string
	maskPhone                                bool
	tenantStatus                             string
	maxMembers, maxInstances, usedInstances  int
	storageUsed, storageLimit                string
}

func (s *PGStore) ProfileForOrg(orgID string) (SettingsSummary, error) {
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) UpdateProfile(orgID string, req UpdateProfileRequest) (SettingsSummary, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE users SET name = $1, status = $2, language = $3, theme_preset = $4, updated_at = NOW()
		WHERE organization_id = $5`,
		req.Name, req.Status, req.Language, req.ThemePreset, orgID)
	if err != nil {
		return SettingsSummary{}, err
	}
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) UpdateGeneral(orgID string, req GeneralSettings) (SettingsSummary, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE organizations
		SET name = $1, timezone = $2, date_format = $3, locale = $4,
			mask_phone_numbers = $5, updated_at = NOW()
		WHERE id = $6`,
		req.OrganizationName, req.Timezone, req.DateFormat, req.Locale, req.MaskPhoneNumbers, orgID)
	if err != nil {
		return SettingsSummary{}, err
	}
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) UpdateAppearance(orgID string, req AppearanceSettings) (SettingsSummary, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `UPDATE organizations SET appearance = $1, updated_at = NOW() WHERE id = $2`, b, orgID)
	if err != nil {
		return SettingsSummary{}, err
	}
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) UpdateChatSettings(orgID string, req ChatSettings) (SettingsSummary, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `UPDATE organizations SET chat_settings = $1, updated_at = NOW() WHERE id = $2`, b, orgID)
	if err != nil {
		return SettingsSummary{}, err
	}
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) UpdateNotificationsSettings(orgID string, req NotificationSettings) (SettingsSummary, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `UPDATE organizations SET notification_settings = $1, updated_at = NOW() WHERE id = $2`, b, orgID)
	if err != nil {
		return SettingsSummary{}, err
	}
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) UpdateCleanupSettings(orgID, actorID string, req CleanupSettings) (SettingsSummary, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `UPDATE organizations SET cleanup_settings = $1, updated_at = NOW() WHERE id = $2`, b, orgID)
	if err != nil {
		return SettingsSummary{}, err
	}
	return s.SettingsForOrg(orgID)
}

func (s *PGStore) RunCleanup(orgID, actorID string) (BackgroundJob, error) {
	retentionDays := 30
	var cleanJSON []byte
	if err := s.db.QueryRow(s.ctx(), `SELECT cleanup_settings FROM organizations WHERE id = $1`, orgID).Scan(&cleanJSON); err == nil {
		var cs CleanupSettings
		if json.Unmarshal(cleanJSON, &cs) == nil && cs.RetentionDays > 0 {
			retentionDays = cs.RetentionDays
		}
	}

	// Delete old closed contacts
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	res, _ := s.db.Exec(s.ctx(), `DELETE FROM contacts WHERE organization_id = $1 AND status = 'closed' AND closed_at < $2`, orgID, cutoff)

	deleted := int(res.RowsAffected())
	summary := fmt.Sprintf("Cleanup: removed %d closed conversations older than %d days.", deleted, retentionDays)
	job := s.recordJob(orgID, "cleanup", "contacts", orgID, summary)
	s.finishJob(job.ID, "completed", "")

	// refresh
	if err := s.db.QueryRow(s.ctx(), `SELECT id, kind, entity_type, entity_id, status, summary, started_at, finished_at FROM background_jobs WHERE id = $1`, job.ID).
		Scan(&job.ID, &job.Kind, &job.EntityType, &job.EntityID, &job.Status, &job.Summary, &job.StartedAt, &job.FinishedAt); err != nil {
		return job, nil
	}
	return job, nil
}

// ---------- INSTANCES ----------

func (s *PGStore) ListInstances(orgID string) ([]WhatsAppInstance, error) {
	query := `
		SELECT id, organization_id, name, phone_number, jid, status, pairing_state, qr_code, slot_blocked,
			settings, call_policy, auto_campaign, rating_settings, assignment_reset,
			health_status, health_uptime_label, health_queue_depth, health_sent_today,
			health_received_today, health_failed_today, health_error_rate, health_observed_at,
			created_at, updated_at
		FROM whatsapp_instances`
	
	var rows pgx.Rows
	var err error
	if orgID == "" {
		rows, err = s.db.Query(s.ctx(), query+" ORDER BY name")
	} else {
		rows, err = s.db.Query(s.ctx(), query+" WHERE organization_id = $1 ORDER BY name", orgID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInstances(rows)
}

func scanInstances(rows pgx.Rows) ([]WhatsAppInstance, error) {
	var instances []WhatsAppInstance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	if instances == nil {
		instances = []WhatsAppInstance{}
	}
	return instances, nil
}

func scanInstance(row interface {
	Scan(dest ...interface{}) error
}) (WhatsAppInstance, error) {
	var inst WhatsAppInstance
	var settingsB, callPolicyB, autoCampaignB, ratingB, assignResetB []byte
	err := row.Scan(
		&inst.ID, &inst.OrganizationID, &inst.Name, &inst.PhoneNumber, &inst.JID, &inst.Status,
		&inst.PairingState, &inst.QRCode, &inst.SlotBlocked,
		&settingsB, &callPolicyB, &autoCampaignB, &ratingB, &assignResetB,
		&inst.Health.Status, &inst.Health.UptimeLabel, &inst.Health.QueueDepth,
		&inst.Health.SentToday, &inst.Health.ReceivedToday, &inst.Health.FailedToday,
		&inst.Health.ErrorRate, &inst.Health.ObservedAt,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return inst, err
	}
	_ = json.Unmarshal(settingsB, &inst.Settings)
	_ = json.Unmarshal(callPolicyB, &inst.CallPolicy)
	_ = json.Unmarshal(autoCampaignB, &inst.AutoCampaign)
	_ = json.Unmarshal(ratingB, &inst.RatingSettings)
	_ = json.Unmarshal(assignResetB, &inst.AssignmentReset)
	return inst, nil
}

func (s *PGStore) ListInstanceHealth(orgID string) ([]InstanceHealthSummary, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, name, health_status, health_uptime_label, health_queue_depth,
			health_sent_today, health_received_today, health_failed_today,
			health_error_rate, health_observed_at
		FROM whatsapp_instances WHERE organization_id = $1 ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []InstanceHealthSummary
	for rows.Next() {
		var h InstanceHealthSummary
		_ = rows.Scan(&h.ID, &h.Name, &h.Status, &h.UptimeLabel,
			&h.QueueDepth, &h.SentToday, &h.ReceivedToday, &h.FailedToday,
			&h.ErrorRate, &h.ObservedAt)
		result = append(result, h)
	}
	if result == nil {
		result = []InstanceHealthSummary{}
	}
	return result, nil
}

func (s *PGStore) CreateInstance(orgID, actorID string, req CreateInstanceRequest) (WhatsAppInstance, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return WhatsAppInstance{}, errors.New("instance name is required")
	}

	defaultSettings, _ := json.Marshal(InstanceSettings{
		SourceTagLabel: name, AllowedSendModes: []string{"all"},
	})
	defaultPolicy := []byte("{}")

	var id string
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO whatsapp_instances (organization_id, name, phone_number, status, pairing_state,
			settings, call_policy, auto_campaign, rating_settings, assignment_reset)
		VALUES ($1,$2,$3,'disconnected','needs_qr',$4,$5,$5,$5,$5)
		RETURNING id`, orgID, name, req.PhoneNumber, defaultSettings, defaultPolicy).Scan(&id)
	if err != nil {
		return WhatsAppInstance{}, err
	}

	// increment used_instances
	_, _ = s.db.Exec(s.ctx(), `UPDATE organizations SET used_instances = (SELECT COUNT(*) FROM whatsapp_instances WHERE organization_id = $1), updated_at = NOW() WHERE id = $1`, orgID)

	s.recordAudit(orgID, actorID, "", "instances.create", "instance", id, "Created a WhatsApp account.", nil)
	return s.getInstanceByID(id)
}

func (s *PGStore) getInstanceByID(id string) (WhatsAppInstance, error) {
	row := s.db.QueryRow(s.ctx(), `
		SELECT id, organization_id, name, phone_number, jid, status, pairing_state, qr_code, slot_blocked,
			settings, call_policy, auto_campaign, rating_settings, assignment_reset,
			health_status, health_uptime_label, health_queue_depth, health_sent_today,
			health_received_today, health_failed_today, health_error_rate, health_observed_at,
			created_at, updated_at
		FROM whatsapp_instances WHERE id = $1`, id)
	return scanInstance(row)
}

func (s *PGStore) UpdateInstanceName(orgID, actorID, instanceID, name string) (WhatsAppInstance, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return WhatsAppInstance{}, errors.New("name is required")
	}
	_, err := s.db.Exec(s.ctx(), `UPDATE whatsapp_instances SET name = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3`, name, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	s.recordAudit(orgID, actorID, "", "instances.rename", "instance", instanceID, "Renamed a WhatsApp account.", map[string]string{"name": name})
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) ConnectInstance(orgID, actorID, instanceID string) (WhatsAppInstance, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances
		SET status = 'connecting', pairing_state = 'qr_ready', qr_code = '',
			health_status = 'connecting', updated_at = NOW()
		WHERE id = $1 AND organization_id = $2`, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	s.recordAudit(orgID, actorID, "", "instances.connect", "instance", instanceID, "Initiated WhatsApp connection.", nil)
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) DisconnectInstance(orgID, actorID, instanceID string) (WhatsAppInstance, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances
		SET status = 'disconnected', pairing_state = 'needs_qr', qr_code = '',
			health_status = 'disconnected', health_uptime_label = '0m',
			health_sent_today = 0, health_received_today = 0, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2`, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	s.recordAudit(orgID, actorID, "", "instances.disconnect", "instance", instanceID, "Disconnected a WhatsApp account.", nil)
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) RecoverInstance(orgID, actorID, instanceID string) (WhatsAppInstance, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances
		SET status = 'disconnected', pairing_state = 'needs_qr', qr_code = '',
			slot_blocked = false, health_status = 'disconnected', updated_at = NOW()
		WHERE id = $1 AND organization_id = $2`, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	s.recordAudit(orgID, actorID, "", "instances.recover", "instance", instanceID, "Recovered a WhatsApp account slot.", nil)
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) UpdateInstanceSettings(orgID, actorID, instanceID string, req InstanceSettings) (WhatsAppInstance, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances SET settings = $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`, b, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) UpdateInstanceCallPolicy(orgID, actorID, instanceID string, req InstanceCallPolicy) (WhatsAppInstance, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances SET call_policy = $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`, b, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) UpdateInstanceAutoCampaign(orgID, actorID, instanceID string, req InstanceAutoCampaign) (WhatsAppInstance, error) {
	b, _ := json.Marshal(req)
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances SET auto_campaign = $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`, b, instanceID, orgID)
	if err != nil {
		return WhatsAppInstance{}, err
	}
	return s.getInstanceByID(instanceID)
}

func (s *PGStore) DeleteInstance(orgID, actorID, instanceID string) error {
	// check status
	var status string
	var name string
	err := s.db.QueryRow(s.ctx(), `SELECT status, name FROM whatsapp_instances WHERE id = $1 AND organization_id = $2`, instanceID, orgID).
		Scan(&status, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("instance not found")
		}
		return err
	}
	if status == "connected" {
		return errors.New("disconnect the account before deleting it")
	}

	// check linked contacts
	var linkedCount int
	_ = s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM contacts WHERE instance_id = $1 AND organization_id = $2`, instanceID, orgID).Scan(&linkedCount)
	if linkedCount > 0 {
		return fmt.Errorf("remove or reassign %d contacts linked to this account before deleting it", linkedCount)
	}

	_, err = s.db.Exec(s.ctx(), `DELETE FROM whatsapp_instances WHERE id = $1 AND organization_id = $2`, instanceID, orgID)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec(s.ctx(), `UPDATE organizations SET used_instances = (SELECT COUNT(*) FROM whatsapp_instances WHERE organization_id = $1), updated_at = NOW() WHERE id = $1`, orgID)
	s.recordAudit(orgID, actorID, "", "instances.delete", "instance", instanceID, "Deleted a WhatsApp account.", map[string]string{"name": name})
	return nil
}

// ---------- NOTIFICATIONS ----------

func (s *PGStore) ListNotifications(orgID string) ([]UserNotification, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, title, body, severity, related_contact_id, related_path, is_read, created_at
		FROM user_notifications WHERE organization_id = $1 ORDER BY created_at DESC LIMIT 50`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifs []UserNotification
	for rows.Next() {
		var n UserNotification
		_ = rows.Scan(&n.ID, &n.Title, &n.Body, &n.Severity, &n.RelatedContactID, &n.RelatedPath, &n.IsRead, &n.CreatedAt)
		notifs = append(notifs, n)
	}
	if notifs == nil {
		notifs = []UserNotification{}
	}
	return notifs, nil
}

func (s *PGStore) MarkAllNotificationsRead(orgID string) error {
	_, err := s.db.Exec(s.ctx(), `UPDATE user_notifications SET is_read = true WHERE organization_id = $1`, orgID)
	return err
}

func (s *PGStore) addNotification(orgID, title, body, severity, relatedContactID, relatedPath string) {
	_, _ = s.db.Exec(s.ctx(), `
		INSERT INTO user_notifications (organization_id, title, body, severity, related_contact_id, related_path)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		orgID, title, body, severity, relatedContactID, relatedPath)
}

// ---------- STATUS POSTS ----------

func (s *PGStore) ListStatuses(orgID string) ([]StatusPost, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, contact_id, contact_name, instance_id, instance_name, body, kind, created_at
		FROM status_posts WHERE organization_id = $1 ORDER BY created_at DESC LIMIT 20`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []StatusPost
	for rows.Next() {
		var p StatusPost
		_ = rows.Scan(&p.ID, &p.ContactID, &p.ContactName, &p.InstanceID, &p.InstanceName, &p.Body, &p.Kind, &p.CreatedAt)
		posts = append(posts, p)
	}
	if posts == nil {
		posts = []StatusPost{}
	}
	return posts, nil
}

func (s *PGStore) AddStatus(orgID string, req AddStatusRequest) (StatusPost, error) {
	var p StatusPost
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO status_posts (organization_id, contact_id, contact_name, instance_id, instance_name, body, kind)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, contact_id, contact_name, instance_id, instance_name, body, kind, created_at`,
		orgID, req.ContactID, req.ContactName, req.InstanceID, req.InstanceName, req.Body, req.Kind).
		Scan(&p.ID, &p.ContactID, &p.ContactName, &p.InstanceID, &p.InstanceName, &p.Body, &p.Kind, &p.CreatedAt)
	return p, err
}

// ---------- QUICK REPLIES ----------

func (s *PGStore) ListQuickReplies(orgID string) ([]QuickReply, error) {
	rows, err := s.db.Query(s.ctx(), `SELECT id, shortcut, title, body FROM quick_replies WHERE organization_id = $1 ORDER BY shortcut`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var qrs []QuickReply
	for rows.Next() {
		var q QuickReply
		_ = rows.Scan(&q.ID, &q.Shortcut, &q.Title, &q.Body)
		qrs = append(qrs, q)
	}
	return qrs, nil
}

// ---------- SHARED HELPERS ----------

func (s *PGStore) recordJob(orgID, kind, entityType, entityID, summary string) BackgroundJob {
	job := BackgroundJob{
		Kind:       kind,
		EntityType: entityType,
		EntityID:   entityID,
		Status:     "running",
		Summary:    summary,
		StartedAt:  time.Now(),
	}
	_ = s.db.QueryRow(s.ctx(), `
		INSERT INTO background_jobs (organization_id, kind, entity_type, entity_id, status, summary)
		VALUES ($1,$2,$3,$4,'running',$5) RETURNING id`,
		orgID, kind, entityType, entityID, summary).Scan(&job.ID)
	return job
}

func (s *PGStore) finishJob(jobID, status, reason string) {
	now := time.Now()
	_, _ = s.db.Exec(s.ctx(), `UPDATE background_jobs SET status = $1, failure_reason = $2, finished_at = $3 WHERE id = $4`,
		status, reason, now, jobID)
}

func (s *PGStore) recordAudit(orgID, actorID, actorName, action, entityType, entityID, summary string, metadata map[string]string) {
	if actorName == "" {
		_ = s.db.QueryRow(s.ctx(), `SELECT name FROM users WHERE id = $1`, actorID).Scan(&actorName)
	}
	b, _ := json.Marshal(metadata)
	if b == nil {
		b = []byte("{}")
	}
	_, _ = s.db.Exec(s.ctx(), `
		INSERT INTO audit_logs (organization_id, actor_user_id, actor_name, action, entity_type, entity_id, summary, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		orgID, actorID, actorName, action, entityType, entityID, summary, b)
}
func (s *PGStore) updateInstanceQRCode(orgID, instanceID, qrCode string) error {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances SET qr_code = $1, pairing_state = 'qr_ready', updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`, qrCode, instanceID, orgID)
	return err
}

func (s *PGStore) updateInstanceStatus(orgID, instanceID, status, pairingState string) error {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances SET status = $1, pairing_state = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4`, status, pairingState, instanceID, orgID)
	return err
}

func (s *PGStore) updateInstanceConnectionSuccess(orgID, instanceID, jid string) error {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE whatsapp_instances 
		SET status = 'connected', pairing_state = 'paired', jid = $1, qr_code = '', updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`, jid, instanceID, orgID)
	return err
}
