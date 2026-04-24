package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"encanto/api"

	"github.com/jackc/pgx/v5"
)

type campaignRecord struct {
	Campaign api.Campaign
}

type campaignTarget struct {
	ContactID   string
	ContactName string
	PhoneNumber string
	InstanceID  string
}

type campaignRecipientOutcome struct {
	ContactID      string
	ContactName    string
	PhoneNumber    string
	Status         string
	FailureReason  string
	MessagePreview string
	DeliveredAt    *time.Time
}

func (p *WorkerPool) processCampaignJob(ctx context.Context, job backgroundJob) error {
	if p == nil || p.store == nil {
		return errors.New("worker store not configured")
	}
	if p.sender == nil {
		return errors.New("campaign sender not configured")
	}

	if strings.TrimSpace(job.OrganizationID) == "" {
		return errors.New("job organization is required")
	}
	campaignID := strings.TrimSpace(job.EntityID)
	if campaignID == "" {
		return errors.New("job campaign entity id is required")
	}

	campaign, err := p.store.loadCampaign(ctx, job.OrganizationID, campaignID)
	if err != nil {
		return err
	}
	if campaign.Campaign.Status != "scheduled" {
		return fmt.Errorf("campaign must be scheduled before worker execution, got %q", campaign.Campaign.Status)
	}

	instanceID := strings.TrimSpace(campaign.Campaign.LinkedInstanceID)
	if instanceID == "" {
		instanceID = strings.TrimSpace(campaign.Campaign.Filters.InstanceID)
	}
	if instanceID == "" {
		return errors.New("campaign instance is required")
	}

	body := strings.TrimSpace(campaign.Campaign.Content)
	if body == "" {
		return errors.New("campaign content is required")
	}

	runID, err := p.store.createCampaignRun(ctx, job.OrganizationID, campaignID, job.ID, "worker")
	if err != nil {
		return err
	}

	targets, err := p.store.listCampaignTargets(ctx, job.OrganizationID, campaign)
	if err != nil {
		return err
	}

	delivered := 0
	failed := 0
	for _, target := range targets {
		outcome := campaignRecipientOutcome{
			ContactID:      target.ContactID,
			ContactName:    target.ContactName,
			PhoneNumber:    target.PhoneNumber,
			Status:         "failed",
			FailureReason:  "",
			MessagePreview: body,
		}

		if err := p.sender.SendCampaignMessage(instanceID, target.PhoneNumber, body); err != nil {
			outcome.FailureReason = err.Error()
			failed++
		} else {
			outcome.Status = "sent"
			now := time.Now()
			outcome.DeliveredAt = &now
			delivered++
		}

		if err := p.store.saveCampaignRecipient(ctx, runID, outcome); err != nil {
			return err
		}
	}

	total := delivered + failed
	if err := p.store.updateCampaignRun(ctx, runID, total, delivered, failed); err != nil {
		return err
	}

	summary := fmt.Sprintf("%d recipients, %d delivered, %d failed", total, delivered, failed)
	if err := p.store.completeCampaign(ctx, job.OrganizationID, campaignID, summary); err != nil {
		return err
	}

	return nil
}

func (s *pgJobStore) loadCampaign(ctx context.Context, orgID, campaignID string) (campaignRecord, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return campaignRecord{}, errors.New("database not configured")
	}

	var record campaignRecord
	var filtersB, scheduleB []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, name, status, source, linked_instance_id, content, filters, schedule, last_run_summary, created_at, updated_at
		FROM campaigns
		WHERE id = $1 AND organization_id = $2`,
		campaignID, orgID).
		Scan(
			&record.Campaign.ID,
			&record.Campaign.Name,
			&record.Campaign.Status,
			&record.Campaign.Source,
			&record.Campaign.LinkedInstanceID,
			&record.Campaign.Content,
			&filtersB,
			&scheduleB,
			&record.Campaign.LastRunSummary,
			&record.Campaign.CreatedAt,
			&record.Campaign.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaignRecord{}, errors.New("campaign not found")
		}
		return campaignRecord{}, err
	}

	if len(filtersB) > 0 {
		_ = json.Unmarshal(filtersB, &record.Campaign.Filters)
	}
	if len(scheduleB) > 0 {
		_ = json.Unmarshal(scheduleB, &record.Campaign.Schedule)
	}
	return record, nil
}

func (s *pgJobStore) createCampaignRun(ctx context.Context, orgID, campaignID, jobID, trigger string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return "", errors.New("database not configured")
	}

	if strings.TrimSpace(trigger) == "" {
		trigger = "worker"
	}

	var runID string
	err := s.db.QueryRow(ctx, `
		INSERT INTO campaign_runs (campaign_id, organization_id, trigger, status, job_id, started_at)
		VALUES ($1, $2, $3, 'running', $4, NOW())
		RETURNING id`,
		campaignID, orgID, trigger, jobID).Scan(&runID)
	if err != nil {
		return "", err
	}
	return runID, nil
}

func (s *pgJobStore) listCampaignTargets(ctx context.Context, orgID string, campaign campaignRecord) ([]campaignTarget, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return nil, errors.New("database not configured")
	}

	q := `
		SELECT c.id, c.name, c.phone_number, c.instance_id
		FROM contacts c
		WHERE c.organization_id = $1`
	args := []any{orgID}
	idx := 2

	if instanceID := strings.TrimSpace(campaign.Campaign.Filters.InstanceID); instanceID != "" {
		q += fmt.Sprintf(" AND c.instance_id = $%d", idx)
		args = append(args, instanceID)
		idx++
	}

	if status := strings.TrimSpace(campaign.Campaign.Filters.Status); status != "" {
		q += fmt.Sprintf(" AND c.status = $%d", idx)
		args = append(args, status)
		idx++
	}

	if tag := strings.TrimSpace(campaign.Campaign.Filters.Tag); tag != "" {
		q += fmt.Sprintf(" AND c.tags::jsonb ? $%d", idx)
		args = append(args, tag)
		idx++
	}

	if search := strings.TrimSpace(campaign.Campaign.Filters.Search); search != "" {
		q += fmt.Sprintf(" AND (c.name ILIKE $%d OR c.phone_number ILIKE $%d)", idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}

	if !campaign.Campaign.Filters.IncludeClosed {
		q += " AND c.status != 'closed'"
	}

	q += " ORDER BY c.last_message_at DESC, c.name ASC"

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []campaignTarget
	for rows.Next() {
		var target campaignTarget
		if err := rows.Scan(&target.ContactID, &target.ContactName, &target.PhoneNumber, &target.InstanceID); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if targets == nil {
		targets = []campaignTarget{}
	}
	return targets, nil
}

func (s *pgJobStore) saveCampaignRecipient(ctx context.Context, runID string, recipient campaignRecipientOutcome) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return errors.New("database not configured")
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO campaign_recipients (
			run_id, contact_id, contact_name, phone_number, status, failure_reason, message_preview, delivered_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		runID,
		recipient.ContactID,
		recipient.ContactName,
		recipient.PhoneNumber,
		recipient.Status,
		recipient.FailureReason,
		recipient.MessagePreview,
		recipient.DeliveredAt,
	)
	return err
}

func (s *pgJobStore) updateCampaignRun(ctx context.Context, runID string, recipientTotal, delivered, failed int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return errors.New("database not configured")
	}

	_, err := s.db.Exec(ctx, `
		UPDATE campaign_runs
		SET status = 'completed',
			finished_at = NOW(),
			recipient_total = $1,
			delivered = $2,
			failed = $3
		WHERE id = $4`,
		recipientTotal, delivered, failed, runID)
	return err
}

func (s *pgJobStore) completeCampaign(ctx context.Context, orgID, campaignID, summary string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return errors.New("database not configured")
	}

	_, err := s.db.Exec(ctx, `
		UPDATE campaigns
		SET status = 'completed',
			last_run_summary = $1,
			updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`,
		summary, campaignID, orgID)
	return err
}
