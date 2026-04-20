package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ---------- CAMPAIGNS ----------

func (s *PGStore) ListCampaigns(orgID string) ([]Campaign, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at
		FROM campaigns WHERE organization_id = $1 ORDER BY updated_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var campaigns []Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	if campaigns == nil {
		campaigns = []Campaign{}
	}
	return campaigns, nil
}

func scanCampaign(row interface {
	Scan(dest ...interface{}) error
}) (Campaign, error) {
	var c Campaign
	var filtersB, scheduleB []byte
	err := row.Scan(&c.ID, &c.Name, &c.Status, &c.Source, &c.LinkedInstanceID,
		&c.Content, &filtersB, &scheduleB, &c.LastRunSummary, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	_ = json.Unmarshal(filtersB, &c.Filters)
	_ = json.Unmarshal(scheduleB, &c.Schedule)
	return c, nil
}

func (s *PGStore) CampaignDetail(orgID, campaignID string) (CampaignRecord, error) {
	row := s.db.QueryRow(s.ctx(), `
		SELECT id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at
		FROM campaigns WHERE id = $1 AND organization_id = $2`, campaignID, orgID)
	camp, err := scanCampaign(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CampaignRecord{}, errors.New("campaign not found")
		}
		return CampaignRecord{}, err
	}

	// runs
	runs, _ := s.ListCampaignRuns(orgID, campaignID)
	// recipients per run
	recipients := map[string][]CampaignRecipient{}
	for _, run := range runs {
		recs, _ := s.listRecipientsForRun(run.ID)
		recipients[run.ID] = recs
	}

	return CampaignRecord{Campaign: camp, Runs: runs, Recipients: recipients}, nil
}

func (s *PGStore) CreateCampaign(orgID, actorID string, req CampaignUpsertRequest) (Campaign, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Campaign{}, errors.New("campaign name is required")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "manual"
	}
	filtersB, _ := json.Marshal(req.Filters)
	scheduleB, _ := json.Marshal(req.Schedule)

	var c Campaign
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO campaigns (organization_id, name, status, source, content, filters, schedule)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at`,
		orgID, name, status, source, strings.TrimSpace(req.Content), filtersB, scheduleB).
		Scan(&c.ID, &c.Name, &c.Status, &c.Source, &c.LinkedInstanceID, &c.Content,
			&filtersB, &scheduleB, &c.LastRunSummary, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Campaign{}, err
	}
	_ = json.Unmarshal(filtersB, &c.Filters)
	_ = json.Unmarshal(scheduleB, &c.Schedule)
	s.recordAudit(orgID, actorID, "", "campaigns.create", "campaign", c.ID, "Created a campaign.", nil)
	return c, nil
}

func (s *PGStore) UpdateCampaign(orgID, actorID, campaignID string, req CampaignUpsertRequest) (Campaign, error) {
	filtersB, _ := json.Marshal(req.Filters)
	scheduleB, _ := json.Marshal(req.Schedule)
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	var c Campaign
	err := s.db.QueryRow(s.ctx(), `
		UPDATE campaigns SET name = $1, content = $2, filters = $3, schedule = $4, status = $5, updated_at = NOW()
		WHERE id = $6 AND organization_id = $7
		RETURNING id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at`,
		strings.TrimSpace(req.Name), strings.TrimSpace(req.Content), filtersB, scheduleB, status, campaignID, orgID).
		Scan(&c.ID, &c.Name, &c.Status, &c.Source, &c.LinkedInstanceID, &c.Content,
			&filtersB, &scheduleB, &c.LastRunSummary, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Campaign{}, err
	}
	_ = json.Unmarshal(filtersB, &c.Filters)
	_ = json.Unmarshal(scheduleB, &c.Schedule)
	s.recordAudit(orgID, actorID, "", "campaigns.edit", "campaign", campaignID, "Updated the campaign.", nil)
	return c, nil
}

func (s *PGStore) DeleteCampaign(orgID, actorID, campaignID string) error {
	res, err := s.db.Exec(s.ctx(), `DELETE FROM campaigns WHERE id = $1 AND organization_id = $2`, campaignID, orgID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("campaign not found")
	}
	s.recordAudit(orgID, actorID, "", "campaigns.delete", "campaign", campaignID, "Deleted a campaign.", nil)
	return nil
}

func (s *PGStore) LaunchCampaign(orgID, actorID, campaignID string) (CampaignRun, []CampaignRecipient, error) {
	camp, err := s.CampaignDetail(orgID, campaignID)
	if err != nil {
		return CampaignRun{}, nil, err
	}

	job := s.recordJob(orgID, "campaign_run", "campaign", campaignID, "Executing a campaign run.")
	now := time.Now()
	var runID string
	err = s.db.QueryRow(s.ctx(), `
		INSERT INTO campaign_runs (campaign_id, organization_id, trigger, status, job_id, started_at)
		VALUES ($1,$2,'manual','running',$3,$4) RETURNING id`,
		campaignID, orgID, job.ID, now).Scan(&runID)
	if err != nil {
		return CampaignRun{}, nil, err
	}

	// target contacts
	contacts, _ := s.targetContacts(orgID, camp.Campaign)
	delivered := 0
	failed := 0
	var storedRecs []CampaignRecipient
	for _, contact := range contacts {
		var instStatus string
		_ = s.db.QueryRow(s.ctx(), `SELECT status FROM whatsapp_instances WHERE id = $1`, contact.InstanceID).Scan(&instStatus)
		rec := CampaignRecipient{
			ContactID:      contact.ID,
			ContactName:    contact.Name,
			PhoneNumber:    contact.PhoneNumber,
			MessagePreview: camp.Campaign.Content,
		}
		if instStatus == "connected" {
			rec.Status = "delivered"
			t := time.Now()
			rec.DeliveredAt = &t
			delivered++
		} else {
			rec.Status = "failed"
			rec.FailureReason = "Instance not connected"
			failed++
		}
		var recID string
		_ = s.db.QueryRow(s.ctx(), `
			INSERT INTO campaign_recipients (run_id, contact_id, contact_name, phone_number, status, failure_reason, message_preview, delivered_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
			runID, rec.ContactID, rec.ContactName, rec.PhoneNumber, rec.Status, rec.FailureReason, rec.MessagePreview, rec.DeliveredAt).Scan(&recID)
		rec.ID = recID
		storedRecs = append(storedRecs, rec)
	}

	endAt := time.Now()
	summary := fmt.Sprintf("%d recipients, %d delivered, %d failed", len(storedRecs), delivered, failed)
	_, _ = s.db.Exec(s.ctx(), `
		UPDATE campaign_runs SET status = 'completed', finished_at = $1, recipient_total = $2, delivered = $3, failed = $4 WHERE id = $5`,
		endAt, len(storedRecs), delivered, failed, runID)
	_, _ = s.db.Exec(s.ctx(), `UPDATE campaigns SET status = 'active', last_run_summary = $1, updated_at = NOW() WHERE id = $2`, summary, campaignID)
	s.finishJob(job.ID, "completed", "")
	s.recordAudit(orgID, actorID, "", "campaigns.launch", "campaign", campaignID, "Launched a campaign run.", map[string]string{"run_id": runID})

	run := CampaignRun{
		ID:             runID,
		CampaignID:     campaignID,
		Trigger:        "manual",
		Status:         "completed",
		JobID:          job.ID,
		StartedAt:      now,
		FinishedAt:     &endAt,
		RecipientTotal: len(storedRecs),
		Delivered:      delivered,
		Failed:         failed,
	}
	if storedRecs == nil {
		storedRecs = []CampaignRecipient{}
	}
	return run, storedRecs, nil
}

func (s *PGStore) targetContacts(orgID string, camp Campaign) ([]ChatContact, error) {
	q := `SELECT c.id, c.name, c.phone_number, c.avatar, c.status,
		c.assigned_user_id, c.assigned_user_name, c.instance_id, c.instance_name,
		c.instance_source_label, c.last_message_preview, c.last_message_at,
		c.last_inbound_at, c.closed_at, c.is_public, c.is_read, c.unread_count,
		c.tags, c.metadata, c.organization_id, c.created_at, false, false, ''
		FROM contacts c WHERE c.organization_id = $1`
	args := []interface{}{orgID}
	idx := 2
	if camp.Filters.InstanceID != "" {
		q += fmt.Sprintf(" AND c.instance_id = $%d", idx)
		args = append(args, camp.Filters.InstanceID)
		idx++
	}
	if camp.Filters.Status != "" {
		q += fmt.Sprintf(" AND c.status = $%d", idx)
		args = append(args, camp.Filters.Status)
		idx++
	}
	if !camp.Filters.IncludeClosed {
		q += " AND c.status != 'closed'"
	}
	rows, err := s.db.Query(s.ctx(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContacts(rows)
}

func (s *PGStore) PauseCampaign(orgID, actorID, campaignID string) (Campaign, error) {
	var c Campaign
	var filtersB, scheduleB []byte
	err := s.db.QueryRow(s.ctx(), `
		UPDATE campaigns SET status = 'paused', updated_at = NOW() WHERE id = $1 AND organization_id = $2
		RETURNING id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at`,
		campaignID, orgID).
		Scan(&c.ID, &c.Name, &c.Status, &c.Source, &c.LinkedInstanceID, &c.Content,
			&filtersB, &scheduleB, &c.LastRunSummary, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Campaign{}, errors.New("campaign not found")
	}
	_ = json.Unmarshal(filtersB, &c.Filters)
	_ = json.Unmarshal(scheduleB, &c.Schedule)
	return c, nil
}

func (s *PGStore) ResumeCampaign(orgID, actorID, campaignID string) (Campaign, error) {
	var c Campaign
	var filtersB, scheduleB []byte
	err := s.db.QueryRow(s.ctx(), `
		UPDATE campaigns SET status = 'scheduled', updated_at = NOW() WHERE id = $1 AND organization_id = $2
		RETURNING id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at`,
		campaignID, orgID).
		Scan(&c.ID, &c.Name, &c.Status, &c.Source, &c.LinkedInstanceID, &c.Content,
			&filtersB, &scheduleB, &c.LastRunSummary, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Campaign{}, errors.New("campaign not found")
	}
	_ = json.Unmarshal(filtersB, &c.Filters)
	_ = json.Unmarshal(scheduleB, &c.Schedule)
	return c, nil
}

func (s *PGStore) ListCampaignRuns(orgID, campaignID string) ([]CampaignRun, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, campaign_id, trigger, status, job_id, started_at, finished_at, recipient_total, delivered, failed
		FROM campaign_runs WHERE campaign_id = $1 ORDER BY started_at DESC`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []CampaignRun
	for rows.Next() {
		var r CampaignRun
		_ = rows.Scan(&r.ID, &r.CampaignID, &r.Trigger, &r.Status, &r.JobID,
			&r.StartedAt, &r.FinishedAt, &r.RecipientTotal, &r.Delivered, &r.Failed)
		runs = append(runs, r)
	}
	if runs == nil {
		runs = []CampaignRun{}
	}
	return runs, nil
}

func (s *PGStore) listRecipientsForRun(runID string) ([]CampaignRecipient, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, run_id, contact_id, contact_name, phone_number, status, failure_reason, message_preview, delivered_at
		FROM campaign_recipients WHERE run_id = $1`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recs []CampaignRecipient
	for rows.Next() {
		var r CampaignRecipient
		_ = rows.Scan(&r.ID, &r.RunID, &r.ContactID, &r.ContactName, &r.PhoneNumber,
			&r.Status, &r.FailureReason, &r.MessagePreview, &r.DeliveredAt)
		recs = append(recs, r)
	}
	if recs == nil {
		recs = []CampaignRecipient{}
	}
	return recs, nil
}

func (s *PGStore) ListCampaignRecipients(orgID, campaignID, runID string) ([]CampaignRecipient, error) {
	if runID == "" {
		// pick last run
		_ = s.db.QueryRow(s.ctx(), `SELECT id FROM campaign_runs WHERE campaign_id = $1 ORDER BY started_at DESC LIMIT 1`, campaignID).Scan(&runID)
	}
	if runID == "" {
		return []CampaignRecipient{}, nil
	}
	return s.listRecipientsForRun(runID)
}

// ---------- JOBS ----------

func (s *PGStore) ListJobs(orgID string) ([]BackgroundJob, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, kind, entity_type, entity_id, status, summary, failure_reason, started_at, finished_at
		FROM background_jobs WHERE organization_id = $1 ORDER BY started_at DESC LIMIT 30`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []BackgroundJob
	for rows.Next() {
		var j BackgroundJob
		_ = rows.Scan(&j.ID, &j.Kind, &j.EntityType, &j.EntityID, &j.Status, &j.Summary, &j.FailureReason, &j.StartedAt, &j.FinishedAt)
		jobs = append(jobs, j)
	}
	if jobs == nil {
		jobs = []BackgroundJob{}
	}
	return jobs, nil
}

func (s *PGStore) JobByID(orgID, jobID string) (BackgroundJob, error) {
	var j BackgroundJob
	err := s.db.QueryRow(s.ctx(), `
		SELECT id, kind, entity_type, entity_id, status, summary, failure_reason, started_at, finished_at
		FROM background_jobs WHERE id = $1 AND organization_id = $2`, jobID, orgID).
		Scan(&j.ID, &j.Kind, &j.EntityType, &j.EntityID, &j.Status, &j.Summary, &j.FailureReason, &j.StartedAt, &j.FinishedAt)
	if err != nil {
		return j, errors.New("job not found")
	}
	return j, nil
}

// ---------- WEBHOOKS ----------

func (s *PGStore) ListWebhooks(orgID string) ([]WebhookEndpoint, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, name, target_url, active FROM webhook_endpoints WHERE organization_id = $1 ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var webhooks []WebhookEndpoint
	for rows.Next() {
		var w WebhookEndpoint
		_ = rows.Scan(&w.ID, &w.Name, &w.TargetURL, &w.Active)
		webhooks = append(webhooks, w)
	}
	if webhooks == nil {
		webhooks = []WebhookEndpoint{}
	}
	return webhooks, nil
}

func (s *PGStore) ListWebhookDeliveries(orgID, webhookID string) ([]WebhookDelivery, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, webhook_id, event_id, status, attempt, last_attempt_at, next_retry_at, response_code, response_body
		FROM webhook_deliveries WHERE webhook_id = $1 AND organization_id = $2 ORDER BY last_attempt_at DESC`, webhookID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deliveries []WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		_ = rows.Scan(&d.ID, &d.WebhookID, &d.EventID, &d.Status, &d.Attempt,
			&d.LastAttemptAt, &d.NextRetryAt, &d.ResponseCode, &d.ResponseBody)
		deliveries = append(deliveries, d)
	}
	if deliveries == nil {
		deliveries = []WebhookDelivery{}
	}
	return deliveries, nil
}

func (s *PGStore) RetryWebhookDelivery(orgID, actorID, webhookID, deliveryID string) (WebhookDelivery, error) {
	now := time.Now()
	nextRetry := now.Add(5 * time.Minute)
	_, err := s.db.Exec(s.ctx(), `
		UPDATE webhook_deliveries SET status = 'retry_scheduled', attempt = attempt + 1,
			last_attempt_at = $1, next_retry_at = $2, response_code = 0
		WHERE id = $3 AND webhook_id = $4 AND organization_id = $5`,
		now, nextRetry, deliveryID, webhookID, orgID)
	if err != nil {
		return WebhookDelivery{}, err
	}
	var d WebhookDelivery
	_ = s.db.QueryRow(s.ctx(), `
		SELECT id, webhook_id, event_id, status, attempt, last_attempt_at, next_retry_at, response_code, response_body
		FROM webhook_deliveries WHERE id = $1`, deliveryID).
		Scan(&d.ID, &d.WebhookID, &d.EventID, &d.Status, &d.Attempt,
			&d.LastAttemptAt, &d.NextRetryAt, &d.ResponseCode, &d.ResponseBody)
	return d, nil
}

// ---------- AUDIT LOGS ----------

func (s *PGStore) ListAuditLogs(orgID string) ([]AuditLogEntry, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, actor_user_id, actor_name, action, entity_type, entity_id, summary, metadata, occurred_at
		FROM audit_logs WHERE organization_id = $1 ORDER BY occurred_at DESC LIMIT 50`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []AuditLogEntry
	for rows.Next() {
		var e AuditLogEntry
		var metaB []byte
		_ = rows.Scan(&e.ID, &e.ActorUserID, &e.ActorName, &e.Action, &e.EntityType, &e.EntityID, &e.Summary, &metaB, &e.OccurredAt)
		_ = json.Unmarshal(metaB, &e.Metadata)
		if e.Metadata == nil {
			e.Metadata = map[string]string{}
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []AuditLogEntry{}
	}
	return entries, nil
}

// ---------- LICENSE ----------

func (s *PGStore) LicenseBootstrap(orgID string) (LicenseBootstrapView, error) {
	var l LicenseBootstrapView
	var activatedAt, expiresAt *time.Time
	err := s.db.QueryRow(s.ctx(), `
		SELECT status, hwid, short_id, last_key_hint, message, activate_url, cleanup_url,
			activated_at, expires_at, max_contacts, max_campaigns, max_instances, tier, kind
		FROM license_records WHERE organization_id = $1`, orgID).
		Scan(&l.Status, &l.HWID, &l.ShortID, &l.LastKeyHint, &l.Message,
			&l.ActivateURL, &l.CleanupURL, &activatedAt, &expiresAt,
			&l.Entitlements.MaxContacts, &l.Entitlements.MaxCampaigns, &l.Entitlements.MaxInstances,
			&l.Tier, &l.Kind)
	if err != nil {
		// return default if no license record
		l = LicenseBootstrapView{
			Status:      "active",
			HWID:        "ENCANTO-DEV",
			ShortID:     "ENCDEV",
			Message:     "License is active.",
			ActivateURL: "/settings/license",
			CleanupURL:  "/license-cleanup",
			Tier:        "growth",
			Kind:        "offline-signed",
			Entitlements: LicenseEntitlements{MaxContacts: 100, MaxCampaigns: 10, MaxInstances: 5},
		}
	}
	l.ActivatedAt = activatedAt
	l.ExpiresAt = expiresAt

	// compute quotas
	var contactCount, campaignCount, instanceCount int
	_ = s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM contacts WHERE organization_id = $1`, orgID).Scan(&contactCount)
	_ = s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM campaigns WHERE organization_id = $1`, orgID).Scan(&campaignCount)
	_ = s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM whatsapp_instances WHERE organization_id = $1`, orgID).Scan(&instanceCount)

	maxContacts := l.Entitlements.MaxContacts
	maxCampaigns := l.Entitlements.MaxCampaigns
	maxInstances := l.Entitlements.MaxInstances
	if maxContacts < 1 {
		maxContacts = 100
	}
	if maxCampaigns < 1 {
		maxCampaigns = 10
	}
	if maxInstances < 1 {
		maxInstances = 5
	}

	l.Quotas = []LicenseQuota{
		{Resource: "contacts", Label: "Contacts", Current: contactCount, Limit: maxContacts, OverQuota: contactCount > maxContacts},
		{Resource: "campaigns", Label: "Campaigns", Current: campaignCount, Limit: maxCampaigns, OverQuota: campaignCount > maxCampaigns},
		{Resource: "instances", Label: "WhatsApp Accounts", Current: instanceCount, Limit: maxInstances, OverQuota: instanceCount > maxInstances},
	}
	for _, q := range l.Quotas {
		if q.OverQuota {
			l.RestrictedCleanup = true
			l.Status = "over_limit"
			l.Message = "Usage is above the current license entitlements."
			break
		}
	}
	return l, nil
}

func (s *PGStore) ActivateLicense(orgID, actorID, securityKey string) (LicenseBootstrapView, error) {
	key := strings.TrimSpace(securityKey)
	if key == "" {
		return LicenseBootstrapView{}, errors.New("security key is required")
	}
	now := time.Now()
	expiresAt := now.Add(45 * 24 * time.Hour)
	hint := key
	if len(hint) > 6 {
		hint = strings.ToUpper(hint[len(hint)-6:])
	}
	_, _ = s.db.Exec(s.ctx(), `
		UPDATE license_records SET status = 'active', last_key_hint = $1, activated_at = $2, expires_at = $3,
			message = 'License activated successfully.', updated_at = NOW()
		WHERE organization_id = $4`, hint, now, expiresAt, orgID)
	s.recordAudit(orgID, actorID, "", "license.activate", "license", orgID, "Applied a new offline license key.", map[string]string{"key_hint": hint})
	return s.LicenseBootstrap(orgID)
}

func (s *PGStore) requireLicensedWrite(orgID string) error {
	view, err := s.LicenseBootstrap(orgID)
	if err != nil {
		return nil
	}
	if view.RestrictedCleanup {
		return errors.New("usage is over the license limits; cleanup is required before creating new records")
	}
	return nil
}

// ---------- ANALYTICS ----------

func (s *PGStore) AgentAnalyticsSummary(orgID string, filters AnalyticsFilters) (AgentAnalyticsSummaryResponse, error) {
	q := `SELECT c.status, c.assigned_user_id FROM contacts c WHERE c.organization_id = $1`
	args := []interface{}{orgID}
	idx := 2
	if filters.InstanceID != "" {
		q += fmt.Sprintf(" AND c.instance_id = $%d", idx)
		args = append(args, filters.InstanceID)
		idx++
	}
	if filters.AgentID != "" {
		q += fmt.Sprintf(" AND c.assigned_user_id = $%d", idx)
		args = append(args, filters.AgentID)
		_ = idx
	}

	rows, err := s.db.Query(s.ctx(), q, args...)
	if err != nil {
		return AgentAnalyticsSummaryResponse{}, err
	}
	defer rows.Close()
	active, closed := 0, 0
	for rows.Next() {
		var status string
		var uid *string
		_ = rows.Scan(&status, &uid)
		if status == "closed" {
			closed++
		} else {
			active++
		}
	}

	return AgentAnalyticsSummaryResponse{
		Cards: []AnalyticsMetricCard{
			{Key: "active_conversations", Label: "Active Conversations", Value: strconv.Itoa(active), Hint: "Open conversations.", EvidenceCount: active + closed},
			{Key: "closed_conversations", Label: "Closed Conversations", Value: strconv.Itoa(closed), Hint: "Closed conversation events.", EvidenceCount: closed},
			{Key: "average_queue_minutes", Label: "Average Queue", Value: "N/A", Hint: "Computed from assignment events.", EvidenceCount: 0},
			{Key: "average_resolution_minutes", Label: "Average Resolution", Value: "N/A", Hint: "Computed from close events.", EvidenceCount: 0},
			{Key: "transfer_count", Label: "Transfers", Value: "0", Hint: "Assignment/reopen events.", EvidenceCount: 0},
			{Key: "average_rating", Label: "Average Rating", Value: "N/A", Hint: "Customer ratings.", EvidenceCount: 0},
		},
		Validation: []AnalyticsMetricEvidence{},
		Generated:  time.Now(),
	}, nil
}

func (s *PGStore) AgentTransferTrends(orgID string, filters AnalyticsFilters) ([]AnalyticsPoint, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT to_char(occurred_at, 'Mon DD') as label, COUNT(*) as val
		FROM timeline_events te
		JOIN contacts c ON c.id = te.contact_id
		WHERE c.organization_id = $1 AND te.event_type IN ('assigned','reopened','unassigned')
		GROUP BY label ORDER BY label`, orgID)
	if err != nil {
		return []AnalyticsPoint{}, nil
	}
	defer rows.Close()
	var points []AnalyticsPoint
	for rows.Next() {
		var p AnalyticsPoint
		var val int
		_ = rows.Scan(&p.Label, &val)
		p.Value = val
		points = append(points, p)
	}
	if points == nil {
		points = []AnalyticsPoint{}
	}
	return points, nil
}

func (s *PGStore) AgentSourceBreakdown(orgID string, filters AnalyticsFilters) ([]AnalyticsPoint, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT COALESCE(NULLIF(instance_source_label,''), instance_name, 'Unknown') as label, COUNT(*) as val
		FROM contacts WHERE organization_id = $1 GROUP BY label ORDER BY val DESC`, orgID)
	if err != nil {
		return []AnalyticsPoint{}, nil
	}
	defer rows.Close()
	var points []AnalyticsPoint
	for rows.Next() {
		var p AnalyticsPoint
		var val int
		_ = rows.Scan(&p.Label, &val)
		p.Value = val
		points = append(points, p)
	}
	if points == nil {
		points = []AnalyticsPoint{}
	}
	return points, nil
}

func (s *PGStore) AgentComparison(orgID string, filters AnalyticsFilters) ([]AgentComparisonRow, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT u.id, u.name,
			COUNT(CASE WHEN c.status != 'closed' THEN 1 END) as active,
			COUNT(CASE WHEN c.status = 'closed' THEN 1 END) as closed
		FROM users u
		LEFT JOIN contacts c ON c.assigned_user_id = u.id AND c.organization_id = $1
		WHERE u.organization_id = $1
		GROUP BY u.id, u.name ORDER BY u.name`, orgID)
	if err != nil {
		return []AgentComparisonRow{}, nil
	}
	defer rows.Close()
	var result []AgentComparisonRow
	for rows.Next() {
		var r AgentComparisonRow
		_ = rows.Scan(&r.AgentID, &r.AgentName, &r.ActiveConversations, &r.ClosedConversations)
		result = append(result, r)
	}
	if result == nil {
		result = []AgentComparisonRow{}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].AgentName < result[j].AgentName })
	return result, nil
}

func (s *PGStore) AgentRatings(orgID string, filters AnalyticsFilters) ([]CustomerRating, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, contact_id, contact_name, phone_number, agent_user_id, agent_name, score, message, rated_at, chat_path, source_event_id
		FROM customer_ratings WHERE organization_id = $1 ORDER BY rated_at DESC`, orgID)
	if err != nil {
		return []CustomerRating{}, nil
	}
	defer rows.Close()
	var ratings []CustomerRating
	for rows.Next() {
		var r CustomerRating
		_ = rows.Scan(&r.ID, &r.ContactID, &r.ContactName, &r.PhoneNumber, &r.AgentUserID, &r.AgentName, &r.Score, &r.Message, &r.RatedAt, &r.ChatPath, &r.SourceEventID)
		ratings = append(ratings, r)
	}
	if ratings == nil {
		ratings = []CustomerRating{}
	}
	return ratings, nil
}

func (s *PGStore) ExportAgentAnalyticsCSV(orgID string, filters AnalyticsFilters) (string, error) {
	rows, _ := s.AgentComparison(orgID, filters)
	var sb strings.Builder
	sb.WriteString("agent_name,active_conversations,closed_conversations,transfers,average_queue_minutes,average_resolution_minutes,average_rating\n")
	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("%s,%d,%d,%d,0,0,0\n",
			r.AgentName, r.ActiveConversations, r.ClosedConversations, r.Transfers))
	}
	return sb.String(), nil
}
