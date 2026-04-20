package api

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ContactMutationRequest struct {
	Name        string   `json:"name"`
	PhoneNumber string   `json:"phone_number"`
	InstanceID  string   `json:"instance_id"`
	Tags        []string `json:"tags"`
}

type ContactsView struct {
	Contacts   []ChatContact      `json:"contacts"`
	Instances  []WhatsAppInstance `json:"instances"`
	Search     string             `json:"search"`
	InstanceID string             `json:"instance_id"`
}

type ContactImportRequest struct {
	CSV               string `json:"csv"`
	UpdateOnDuplicate bool   `json:"update_on_duplicate"`
}

type ContactImportPreviewRow struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Instance    string `json:"instance"`
	Action      string `json:"action"`
}

type ContactImportResult struct {
	Created         int                       `json:"created"`
	Updated         int                       `json:"updated"`
	Skipped         int                       `json:"skipped"`
	DuplicatePhones []string                  `json:"duplicate_phones"`
	Preview         []ContactImportPreviewRow `json:"preview"`
	Job             BackgroundJob             `json:"job"`
}

type ClosedConversationRow struct {
	ID               string    `json:"id"`
	ContactName      string    `json:"contact_name"`
	PhoneDisplay     string    `json:"phone_display"`
	InstanceID       string    `json:"instance_id"`
	InstanceName     string    `json:"instance_name"`
	AssignedUserName string    `json:"assigned_user_name"`
	ClosedBy         string    `json:"closed_by"`
	ClosedAt         time.Time `json:"closed_at"`
}

type ClosedChatPage struct {
	Items       []ClosedConversationRow `json:"items"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
	Total       int                     `json:"total"`
	HasNext     bool                    `json:"has_next"`
	HasPrevious bool                    `json:"has_previous"`
	AgentID     string                  `json:"agent_id"`
	InstanceID  string                  `json:"instance_id"`
}

type LicenseEntitlements struct {
	MaxContacts  int    `json:"max_contacts"`
	MaxCampaigns int    `json:"max_campaigns"`
	MaxInstances int    `json:"max_instances"`
	Tier         string `json:"tier"`
	Kind         string `json:"kind"`
}

type LicenseRecord struct {
	Status       string              `json:"status"`
	HWID         string              `json:"hwid"`
	ShortID      string              `json:"short_id"`
	LastKeyHint  string              `json:"last_key_hint"`
	Message      string              `json:"message"`
	ActivateURL  string              `json:"activate_url"`
	CleanupURL   string              `json:"cleanup_url"`
	ActivatedAt  *time.Time          `json:"activated_at,omitempty"`
	ExpiresAt    *time.Time          `json:"expires_at,omitempty"`
	Entitlements LicenseEntitlements `json:"entitlements"`
}

type LicenseQuota struct {
	Resource  string `json:"resource"`
	Label     string `json:"label"`
	Current   int    `json:"current"`
	Limit     int    `json:"limit"`
	OverQuota bool   `json:"over_quota"`
}

type LicenseBootstrapView struct {
	Status            string         `json:"status"`
	Tier              string         `json:"tier"`
	Kind              string         `json:"kind"`
	HWID              string         `json:"hwid"`
	ShortID           string         `json:"short_id"`
	LastKeyHint       string         `json:"last_key_hint"`
	Message           string         `json:"message"`
	ActivateURL       string         `json:"activate_url"`
	CleanupURL        string         `json:"cleanup_url"`
	ActivatedAt       *time.Time     `json:"activated_at,omitempty"`
	ExpiresAt         *time.Time     `json:"expires_at,omitempty"`
	RestrictedCleanup bool           `json:"restricted_cleanup"`
	Quotas            []LicenseQuota `json:"quotas"`
}

type LicenseActivationRequest struct {
	SecurityKey string `json:"security_key"`
}

type AnalyticsFilters struct {
	AgentID    string
	InstanceID string
}

type AnalyticsMetricCard struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Value         string `json:"value"`
	Hint          string `json:"hint"`
	EvidenceCount int    `json:"evidence_count"`
}

type AnalyticsMetricEvidence struct {
	MetricKey    string   `json:"metric_key"`
	Explanation  string   `json:"explanation"`
	ContactIDs   []string `json:"contact_ids"`
	SourceEvents []string `json:"source_events"`
}

type AgentAnalyticsSummaryResponse struct {
	Cards      []AnalyticsMetricCard     `json:"cards"`
	Validation []AnalyticsMetricEvidence `json:"validation"`
	Generated  time.Time                 `json:"generated"`
}

type AnalyticsPoint struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type AgentComparisonRow struct {
	AgentID               string  `json:"agent_id"`
	AgentName             string  `json:"agent_name"`
	ActiveConversations   int     `json:"active_conversations"`
	ClosedConversations   int     `json:"closed_conversations"`
	Transfers             int     `json:"transfers"`
	AverageQueueMinutes   float64 `json:"average_queue_minutes"`
	AverageResolutionMins float64 `json:"average_resolution_minutes"`
	AverageRating         float64 `json:"average_rating"`
}

type CustomerRating struct {
	ID            string    `json:"id"`
	ContactID     string    `json:"contact_id"`
	ContactName   string    `json:"contact_name"`
	PhoneNumber   string    `json:"phone_number"`
	AgentUserID   string    `json:"agent_user_id"`
	AgentName     string    `json:"agent_name"`
	Score         int       `json:"score"`
	Message       string    `json:"message"`
	RatedAt       time.Time `json:"rated_at"`
	ChatPath      string    `json:"chat_path"`
	SourceEventID string    `json:"source_event_id"`
}

type CampaignFilters struct {
	InstanceID    string `json:"instance_id"`
	Tag           string `json:"tag"`
	Status        string `json:"status"`
	Search        string `json:"search"`
	IncludeClosed bool   `json:"include_closed"`
}

type CampaignSchedule struct {
	Mode      string `json:"mode"`
	EveryDays int    `json:"every_days"`
	TimeOfDay string `json:"time_of_day"`
}

type Campaign struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Status           string           `json:"status"`
	Source           string           `json:"source"`
	LinkedInstanceID string           `json:"linked_instance_id"`
	Content          string           `json:"content"`
	Filters          CampaignFilters  `json:"filters"`
	Schedule         CampaignSchedule `json:"schedule"`
	LastRunSummary   string           `json:"last_run_summary"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type CampaignRecord struct {
	Campaign   Campaign                       `json:"campaign"`
	Runs       []CampaignRun                  `json:"runs"`
	Recipients map[string][]CampaignRecipient `json:"recipients"`
}

type CampaignUpsertRequest struct {
	Name     string           `json:"name"`
	Content  string           `json:"content"`
	Status   string           `json:"status"`
	Source   string           `json:"source"`
	Filters  CampaignFilters  `json:"filters"`
	Schedule CampaignSchedule `json:"schedule"`
}

type CampaignRun struct {
	ID             string     `json:"id"`
	CampaignID     string     `json:"campaign_id"`
	Trigger        string     `json:"trigger"`
	Status         string     `json:"status"`
	JobID          string     `json:"job_id"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	RecipientTotal int        `json:"recipient_total"`
	Delivered      int        `json:"delivered"`
	Failed         int        `json:"failed"`
}

type CampaignRecipient struct {
	ID             string     `json:"id"`
	RunID          string     `json:"run_id"`
	ContactID      string     `json:"contact_id"`
	ContactName    string     `json:"contact_name"`
	PhoneNumber    string     `json:"phone_number"`
	Status         string     `json:"status"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	MessagePreview string     `json:"message_preview"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
}

type BackgroundJob struct {
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	EntityType    string     `json:"entity_type"`
	EntityID      string     `json:"entity_id"`
	Status        string     `json:"status"`
	Summary       string     `json:"summary"`
	FailureReason string     `json:"failure_reason,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

type OutboxEvent struct {
	ID                 string            `json:"id"`
	EventType          string            `json:"event_type"`
	EntityType         string            `json:"entity_type"`
	EntityID           string            `json:"entity_id"`
	Status             string            `json:"status"`
	OccurredAt         time.Time         `json:"occurred_at"`
	Payload            map[string]string `json:"payload"`
	DeliveryCount      int               `json:"delivery_count"`
	LastDeliveryStatus string            `json:"last_delivery_status"`
}

type WebhookEndpoint struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TargetURL string `json:"target_url"`
	Active    bool   `json:"active"`
}

type WebhookDelivery struct {
	ID            string     `json:"id"`
	WebhookID     string     `json:"webhook_id"`
	EventID       string     `json:"event_id"`
	Status        string     `json:"status"`
	Attempt       int        `json:"attempt"`
	LastAttemptAt time.Time  `json:"last_attempt_at"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	ResponseCode  int        `json:"response_code"`
	ResponseBody  string     `json:"response_body"`
}

type AuditLogEntry struct {
	ID          string            `json:"id"`
	ActorUserID string            `json:"actor_user_id"`
	ActorName   string            `json:"actor_name"`
	Action      string            `json:"action"`
	EntityType  string            `json:"entity_type"`
	EntityID    string            `json:"entity_id"`
	Summary     string            `json:"summary"`
	Metadata    map[string]string `json:"metadata"`
	OccurredAt  time.Time         `json:"occurred_at"`
}

func seedLicenseRecord(now time.Time, orgID, orgName string, maxContacts, maxCampaigns, maxInstances int) LicenseRecord {
	activatedAt := now.Add(-14 * 24 * time.Hour)
	expiresAt := now.Add(45 * 24 * time.Hour)

	return LicenseRecord{
		Status:      "active",
		HWID:        strings.ToUpper(strings.ReplaceAll(orgID+"-"+orgName, " ", "-")),
		ShortID:     strings.ToUpper(strings.ReplaceAll(orgID, "-", "")),
		LastKeyHint: "DEMO42",
		Message:     "License is active and within limits.",
		ActivateURL: "/settings/license",
		CleanupURL:  "/license-cleanup",
		ActivatedAt: &activatedAt,
		ExpiresAt:   &expiresAt,
		Entitlements: LicenseEntitlements{
			MaxContacts:  maxContacts,
			MaxCampaigns: maxCampaigns,
			MaxInstances: maxInstances,
			Tier:         "growth",
			Kind:         "offline-signed",
		},
	}
}

func seedPrimaryCampaigns(now time.Time) map[string]*CampaignRecord {
	finishedAt := now.Add(-18 * time.Hour)
	return map[string]*CampaignRecord{
		"camp-1": {
			Campaign: Campaign{
				ID:               "camp-1",
				Name:             "Winback Weekly",
				Status:           "scheduled",
				Source:           "instance_auto_campaign",
				LinkedInstanceID: "inst-1",
				Content:          "We have a new offer ready for you.",
				Filters:          CampaignFilters{InstanceID: "inst-1", Tag: "renewal", Status: "assigned"},
				Schedule:         CampaignSchedule{Mode: "every_n_days", EveryDays: 7, TimeOfDay: "10:00"},
				LastRunSummary:   "2 recipients, 1 delivered, 1 failed",
				CreatedAt:        now.Add(-20 * 24 * time.Hour),
				UpdatedAt:        now.Add(-18 * time.Hour),
			},
			Runs: []CampaignRun{
				{
					ID:             "run-1",
					CampaignID:     "camp-1",
					Trigger:        "instance_auto_campaign",
					Status:         "completed",
					JobID:          "job-1",
					StartedAt:      now.Add(-18*time.Hour - 2*time.Second),
					FinishedAt:     &finishedAt,
					RecipientTotal: 2,
					Delivered:      1,
					Failed:         1,
				},
			},
			Recipients: map[string][]CampaignRecipient{
				"run-1": {
					{ID: "recipient-1", RunID: "run-1", ContactID: "contact-1", ContactName: "Mina Salah", PhoneNumber: "+201111111111", Status: "delivered", MessagePreview: "We have a new offer ready for you.", DeliveredAt: &finishedAt},
					{ID: "recipient-2", RunID: "run-1", ContactID: "contact-2", ContactName: "Laila Hassan", PhoneNumber: "+201222222222", Status: "failed", FailureReason: "Instance not connected", MessagePreview: "We have a new offer ready for you."},
				},
			},
		},
		"camp-2": {
			Campaign: Campaign{
				ID:             "camp-2",
				Name:           "April Follow-up",
				Status:         "draft",
				Source:         "manual",
				Content:        "Checking if you still need help with your request.",
				Filters:        CampaignFilters{Status: "pending"},
				Schedule:       CampaignSchedule{Mode: "manual", EveryDays: 0, TimeOfDay: ""},
				LastRunSummary: "Draft only",
				CreatedAt:      now.Add(-3 * 24 * time.Hour),
				UpdatedAt:      now.Add(-2 * time.Hour),
			},
			Runs:       []CampaignRun{},
			Recipients: map[string][]CampaignRecipient{},
		},
	}
}

func seedSecondaryCampaigns(now time.Time) map[string]*CampaignRecord {
	return map[string]*CampaignRecord{
		"camp-3": {
			Campaign: Campaign{
				ID:             "camp-3",
				Name:           "Store Walk-ins",
				Status:         "scheduled",
				Source:         "manual",
				Content:        "Thank you for visiting our store.",
				Filters:        CampaignFilters{InstanceID: "inst-3"},
				Schedule:       CampaignSchedule{Mode: "every_n_days", EveryDays: 14, TimeOfDay: "18:00"},
				LastRunSummary: "No runs yet",
				CreatedAt:      now.Add(-24 * time.Hour),
				UpdatedAt:      now.Add(-12 * time.Hour),
			},
			Runs:       []CampaignRun{},
			Recipients: map[string][]CampaignRecipient{},
		},
	}
}

func seedPrimaryJobs(now time.Time) []BackgroundJob {
	finishedAt := now.Add(-18 * time.Hour)
	return []BackgroundJob{
		{
			ID:         "job-1",
			Kind:       "campaign_run",
			EntityType: "campaign",
			EntityID:   "camp-1",
			Status:     "completed",
			Summary:    "Processed the scheduled winback campaign.",
			StartedAt:  now.Add(-18*time.Hour - 2*time.Second),
			FinishedAt: &finishedAt,
		},
	}
}

func seedDefaultWebhooks() map[string]*WebhookEndpoint {
	return map[string]*WebhookEndpoint{
		"webhook-1": {
			ID:        "webhook-1",
			Name:      "Primary Ops Sink",
			TargetURL: "https://ops.example.com/webhooks/encanto",
			Active:    true,
		},
	}
}

func seedPrimaryDeliveries(now time.Time) []WebhookDelivery {
	nextRetry := now.Add(-10 * time.Minute)
	return []WebhookDelivery{
		{
			ID:            "delivery-1",
			WebhookID:     "webhook-1",
			EventID:       "outbox-2",
			Status:        "retry_scheduled",
			Attempt:       1,
			LastAttemptAt: now.Add(-15 * time.Minute),
			NextRetryAt:   &nextRetry,
			ResponseCode:  502,
			ResponseBody:  "upstream timeout",
		},
		{
			ID:            "delivery-2",
			WebhookID:     "webhook-1",
			EventID:       "outbox-1",
			Status:        "delivered",
			Attempt:       1,
			LastAttemptAt: now.Add(-18 * time.Hour),
			ResponseCode:  202,
			ResponseBody:  "accepted",
		},
	}
}

func seedPrimaryOutbox(now time.Time) []OutboxEvent {
	return []OutboxEvent{
		{
			ID:                 "outbox-1",
			EventType:          "campaign_run.completed",
			EntityType:         "campaign",
			EntityID:           "camp-1",
			Status:             "delivered",
			OccurredAt:         now.Add(-18 * time.Hour),
			Payload:            map[string]string{"campaign_id": "camp-1", "run_id": "run-1"},
			DeliveryCount:      1,
			LastDeliveryStatus: "delivered",
		},
		{
			ID:                 "outbox-2",
			EventType:          "cleanup.scheduled",
			EntityType:         "cleanup",
			EntityID:           "org-1",
			Status:             "retry_scheduled",
			OccurredAt:         now.Add(-15 * time.Minute),
			Payload:            map[string]string{"job_kind": "cleanup"},
			DeliveryCount:      1,
			LastDeliveryStatus: "retry_scheduled",
		},
	}
}

func seedPrimaryAudit(now time.Time) []AuditLogEntry {
	return []AuditLogEntry{
		{
			ID:          "audit-1",
			ActorUserID: "1",
			ActorName:   "Admin Encanto",
			Action:      "campaign.launch",
			EntityType:  "campaign",
			EntityID:    "camp-1",
			Summary:     "Launched the Winback Weekly campaign.",
			Metadata:    map[string]string{"run_id": "run-1"},
			OccurredAt:  now.Add(-18 * time.Hour),
		},
		{
			ID:          "audit-2",
			ActorUserID: "1",
			ActorName:   "Admin Encanto",
			Action:      "cleanup.schedule.update",
			EntityType:  "cleanup",
			EntityID:    "org-1",
			Summary:     "Adjusted the uploads cleanup schedule.",
			Metadata:    map[string]string{"run_hour": "3"},
			OccurredAt:  now.Add(-2 * time.Hour),
		},
	}
}

func seedPrimaryRatings(now time.Time) []CustomerRating {
	return []CustomerRating{
		{
			ID:            "rating-1",
			ContactID:     "contact-3",
			ContactName:   "Omar Group",
			PhoneNumber:   "+201333333333",
			AgentUserID:   "3",
			AgentName:     "Omar Care",
			Score:         5,
			Message:       "Fast follow-up and clear answers.",
			RatedAt:       now.Add(-4 * time.Hour),
			ChatPath:      "/chat/contact-3",
			SourceEventID: "event-4",
		},
		{
			ID:            "rating-2",
			ContactID:     "contact-1",
			ContactName:   "Mina Salah",
			PhoneNumber:   "+201111111111",
			AgentUserID:   "2",
			AgentName:     "Maha Support",
			Score:         4,
			Message:       "Helpful proposal follow-up.",
			RatedAt:       now.Add(-6 * time.Hour),
			ChatPath:      "/chat/contact-1",
			SourceEventID: "event-2",
		},
	}
}

func seedSecondaryRatings(now time.Time) []CustomerRating {
	return []CustomerRating{
		{
			ID:            "rating-3",
			ContactID:     "contact-4",
			ContactName:   "Store Visitor",
			PhoneNumber:   "+201444444444",
			AgentUserID:   "1",
			AgentName:     "Admin Encanto",
			Score:         5,
			Message:       "Quick reply with store hours.",
			RatedAt:       now.Add(-90 * time.Minute),
			ChatPath:      "/chat/contact-4",
			SourceEventID: "msg-7",
		},
	}
}

func (s *Store) recordJobUnlocked(org *OrgData, kind, entityType, entityID, summary string) BackgroundJob {
	job := BackgroundJob{
		ID:         s.next("job"),
		Kind:       kind,
		EntityType: entityType,
		EntityID:   entityID,
		Status:     "running",
		Summary:    summary,
		StartedAt:  time.Now(),
	}
	org.Jobs = append([]BackgroundJob{job}, org.Jobs...)
	if len(org.Jobs) > 30 {
		org.Jobs = org.Jobs[:30]
	}
	return job
}

func (s *Store) finishJobUnlocked(org *OrgData, jobID, status, failureReason string) {
	for i := range org.Jobs {
		if org.Jobs[i].ID != jobID {
			continue
		}
		finishedAt := time.Now()
		org.Jobs[i].Status = status
		org.Jobs[i].FailureReason = failureReason
		org.Jobs[i].FinishedAt = &finishedAt
		return
	}
}

func (s *Store) recordAuditUnlocked(org *OrgData, actorID, actorName, action, entityType, entityID, summary string, metadata map[string]string) {
	entry := AuditLogEntry{
		ID:          s.next("audit"),
		ActorUserID: actorID,
		ActorName:   actorName,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Summary:     summary,
		Metadata:    metadata,
		OccurredAt:  time.Now(),
	}
	org.Audit = append([]AuditLogEntry{entry}, org.Audit...)
	if len(org.Audit) > 50 {
		org.Audit = org.Audit[:50]
	}
}

func (s *Store) recordOutboxUnlocked(org *OrgData, eventType, entityType, entityID string, payload map[string]string, forceRetry bool) {
	event := OutboxEvent{
		ID:         s.next("outbox"),
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		OccurredAt: time.Now(),
		Payload:    payload,
	}

	status := "delivered"
	responseCode := 202
	responseBody := "accepted"
	var nextRetryAt *time.Time
	if forceRetry {
		status = "retry_scheduled"
		responseCode = 502
		responseBody = "upstream timeout"
		retryAt := time.Now().Add(5 * time.Minute)
		nextRetryAt = &retryAt
	}

	event.Status = status
	event.LastDeliveryStatus = status
	webhookIDs := sortedWebhookIDs(org.Webhooks)
	event.DeliveryCount = len(webhookIDs)

	org.Outbox = append([]OutboxEvent{event}, org.Outbox...)
	if len(org.Outbox) > 50 {
		org.Outbox = org.Outbox[:50]
	}

	for _, webhookID := range webhookIDs {
		delivery := WebhookDelivery{
			ID:            s.next("delivery"),
			WebhookID:     webhookID,
			EventID:       event.ID,
			Status:        status,
			Attempt:       1,
			LastAttemptAt: time.Now(),
			NextRetryAt:   nextRetryAt,
			ResponseCode:  responseCode,
			ResponseBody:  responseBody,
		}
		org.Deliveries = append([]WebhookDelivery{delivery}, org.Deliveries...)
	}
	if len(org.Deliveries) > 80 {
		org.Deliveries = org.Deliveries[:80]
	}
}

func sortedWebhookIDs(webhooks map[string]*WebhookEndpoint) []string {
	ids := make([]string, 0, len(webhooks))
	for id, webhook := range webhooks {
		if webhook == nil || !webhook.Active {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (s *Store) ListAdminContacts(orgID, userID, search, instanceID string) (ContactsView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ContactsView{}, errors.New("organization not found")
	}

	searchValue := strings.ToLower(strings.TrimSpace(search))
	contacts := make([]ChatContact, 0, len(org.Contacts))
	for _, record := range org.Contacts {
		contact := s.decoratedContact(org, record, userID)
		if instanceID != "" && contact.InstanceID != instanceID {
			continue
		}
		if searchValue != "" {
			if !strings.Contains(strings.ToLower(contact.Name), searchValue) &&
				!strings.Contains(strings.ToLower(contact.PhoneNumber), searchValue) &&
				!strings.Contains(strings.ToLower(strings.Join(contact.Tags, " ")), searchValue) {
				continue
			}
		}
		contacts = append(contacts, contact)
	}

	sort.Slice(contacts, func(i, j int) bool {
		if contacts[i].Name == contacts[j].Name {
			return contacts[i].LastMessageAt.After(contacts[j].LastMessageAt)
		}
		return contacts[i].Name < contacts[j].Name
	})

	return ContactsView{
		Contacts:   contacts,
		Instances:  sortedInstancesForOrg(org),
		Search:     search,
		InstanceID: instanceID,
	}, nil
}

func sortedInstancesForOrg(org *OrgData) []WhatsAppInstance {
	instances := make([]WhatsAppInstance, 0, len(org.Instances))
	for _, instance := range org.Instances {
		instances = append(instances, *instance)
	}
	sort.Slice(instances, func(i, j int) bool { return instances[i].Name < instances[j].Name })
	return instances
}

func (s *Store) CreateContact(orgID, actorID string, req ContactMutationRequest) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}

	instance, err := resolveInstanceForContact(org, req.InstanceID)
	if err != nil {
		return ChatContact{}, err
	}

	phone := normalizePhoneNumber(req.PhoneNumber)
	if phone == "" {
		return ChatContact{}, errors.New("phone number is required")
	}
	for _, record := range org.Contacts {
		if normalizePhoneNumber(record.Contact.PhoneNumber) == phone && record.Contact.InstanceID == instance.ID {
			return ChatContact{}, errors.New("contact already exists for this number and account")
		}
	}

	now := time.Now()
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = phone
	}
	contactID := s.next("contact")
	record := &ConversationRecord{
		Contact: ChatContact{
			ID:                  contactID,
			OrganizationID:      orgID,
			Name:                name,
			PhoneNumber:         phone,
			Avatar:              "https://i.pravatar.cc/150?u=" + contactID,
			Status:              "pending",
			InstanceID:          instance.ID,
			InstanceName:        instance.Name,
			InstanceSourceLabel: instance.Settings.SourceTagLabel,
			LastMessagePreview:  "Contact added from the contacts management surface.",
			LastMessageAt:       now,
			LastInboundAt:       now,
			IsPublic:            true,
			IsRead:              true,
			Tags:                append([]string{}, req.Tags...),
			Metadata:            map[string]string{"created_via": "contacts_admin"},
		},
		Messages:      []ChatMessage{},
		Notes:         []ConversationNote{},
		Collaborators: []Collaborator{},
		Events: []TimelineEvent{
			{
				ID:          s.next("event"),
				EventType:   "contact_created",
				ActorUserID: actorID,
				ActorName:   s.actorName(org, actorID),
				Summary:     "Created a new contact from the contacts screen",
				OccurredAt:  now,
				Metadata:    map[string]string{"instance_id": instance.ID},
			},
		},
		UserStates: map[string]*ContactUserState{actorID: {}},
	}

	org.Contacts[contactID] = record
	s.addOrgNotification(org, UserNotification{
		ID:               s.next("notif"),
		Title:            "Contact created",
		Body:             fmt.Sprintf("%s was added to the contacts directory.", record.Contact.Name),
		Severity:         "success",
		RelatedContactID: contactID,
		RelatedPath:      "/settings/contacts",
		CreatedAt:        now,
	})
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "contacts.create", "contact", contactID, "Created a contact from the contacts screen.", map[string]string{"instance_id": instance.ID})
	s.recordOutboxUnlocked(org, "contacts.created", "contact", contactID, map[string]string{"instance_id": instance.ID}, false)

	return s.decoratedContact(org, record, actorID), nil
}

func resolveInstanceForContact(org *OrgData, instanceID string) (*WhatsAppInstance, error) {
	if instanceID != "" {
		instance, ok := org.Instances[instanceID]
		if !ok {
			return nil, errors.New("instance not found")
		}
		return instance, nil
	}

	for _, instance := range org.Instances {
		return instance, nil
	}

	return nil, errors.New("no WhatsApp accounts are available")
}

func (s *Store) UpdateContact(orgID, actorID, contactID string, req ContactMutationRequest) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("contact not found")
	}
	instance, err := resolveInstanceForContact(org, req.InstanceID)
	if err != nil {
		return ChatContact{}, err
	}

	phone := normalizePhoneNumber(req.PhoneNumber)
	if phone == "" {
		return ChatContact{}, errors.New("phone number is required")
	}
	for _, existing := range org.Contacts {
		if existing.Contact.ID == contactID {
			continue
		}
		if normalizePhoneNumber(existing.Contact.PhoneNumber) == phone && existing.Contact.InstanceID == instance.ID {
			return ChatContact{}, errors.New("another contact already uses this number on the selected account")
		}
	}

	record.Contact.Name = strings.TrimSpace(req.Name)
	if record.Contact.Name == "" {
		record.Contact.Name = phone
	}
	record.Contact.PhoneNumber = phone
	record.Contact.InstanceID = instance.ID
	record.Contact.InstanceName = instance.Name
	record.Contact.InstanceSourceLabel = instance.Settings.SourceTagLabel
	record.Contact.Tags = append([]string{}, req.Tags...)
	record.Contact.LastMessageAt = time.Now()
	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "contact_updated",
		ActorUserID: actorID,
		ActorName:   s.actorName(org, actorID),
		Summary:     "Updated contact details",
		OccurredAt:  time.Now(),
		Metadata:    map[string]string{"instance_id": instance.ID},
	})

	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "contacts.edit", "contact", contactID, "Updated contact details from the contacts screen.", map[string]string{"instance_id": instance.ID})
	s.recordOutboxUnlocked(org, "contacts.updated", "contact", contactID, map[string]string{"instance_id": instance.ID}, false)

	return s.decoratedContact(org, record, actorID), nil
}

func (s *Store) DeleteContact(orgID, actorID, contactID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return errors.New("contact not found")
	}

	delete(org.Contacts, contactID)
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "contacts.delete", "contact", contactID, "Deleted a contact from the contacts screen.", map[string]string{"name": record.Contact.Name})
	s.recordOutboxUnlocked(org, "contacts.deleted", "contact", contactID, map[string]string{"name": record.Contact.Name}, false)
	return nil
}

func (s *Store) ExportContactsCSV(orgID string, columns []string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return "", errors.New("organization not found")
	}

	if len(columns) == 0 {
		columns = []string{"name", "phone_number", "instance_name", "status", "assigned_user_name", "tags"}
	}

	contacts := make([]ChatContact, 0, len(org.Contacts))
	for _, record := range org.Contacts {
		contacts = append(contacts, record.Contact)
	}
	sort.Slice(contacts, func(i, j int) bool { return contacts[i].Name < contacts[j].Name })

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write(columns); err != nil {
		return "", err
	}
	for _, contact := range contacts {
		row := make([]string, 0, len(columns))
		for _, column := range columns {
			switch column {
			case "name":
				row = append(row, contact.Name)
			case "phone_number":
				row = append(row, contact.PhoneNumber)
			case "instance_name":
				row = append(row, contact.InstanceName)
			case "instance_id":
				row = append(row, contact.InstanceID)
			case "status":
				row = append(row, contact.Status)
			case "assigned_user_name":
				row = append(row, contact.AssignedUserName)
			case "tags":
				row = append(row, strings.Join(contact.Tags, "|"))
			case "last_message_at":
				row = append(row, contact.LastMessageAt.Format(time.RFC3339))
			default:
				row = append(row, "")
			}
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (s *Store) SampleContactsCSV(orgID string, columns []string) (string, error) {
	if len(columns) == 0 {
		columns = []string{"name", "phone_number", "instance_name", "status", "assigned_user_name", "tags"}
	}
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write(columns); err != nil {
		return "", err
	}
	sample := map[string]string{
		"name":               "Sample Contact",
		"phone_number":       "+201500000000",
		"instance_name":      "Sales WA",
		"instance_id":        "inst-1",
		"status":             "pending",
		"assigned_user_name": "",
		"tags":               "sample|import",
	}
	row := make([]string, 0, len(columns))
	for _, column := range columns {
		row = append(row, sample[column])
	}
	if err := writer.Write(row); err != nil {
		return "", err
	}
	writer.Flush()
	return buffer.String(), writer.Error()
}

func (s *Store) ImportContactsCSV(orgID, actorID, csvData string, updateOnDuplicate bool) (ContactImportResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ContactImportResult{}, errors.New("organization not found")
	}

	reader := csv.NewReader(strings.NewReader(strings.TrimSpace(csvData)))
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return ContactImportResult{}, errors.New("invalid CSV payload")
	}
	if len(rows) < 2 {
		return ContactImportResult{}, errors.New("the CSV import requires a header row and at least one data row")
	}

	header := map[string]int{}
	for index, column := range rows[0] {
		header[strings.ToLower(strings.TrimSpace(column))] = index
	}
	if _, ok := header["phone_number"]; !ok {
		return ContactImportResult{}, errors.New("phone_number column is required")
	}

	result := ContactImportResult{
		DuplicatePhones: []string{},
		Preview:         []ContactImportPreviewRow{},
	}
	job := s.recordJobUnlocked(org, "contacts_import", "contacts", orgID, "Imported contacts from CSV.")

	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}
		name := csvCell(row, header, "name")
		phone := normalizePhoneNumber(csvCell(row, header, "phone_number"))
		instanceID := csvCell(row, header, "instance_id")
		instanceName := csvCell(row, header, "instance_name")
		tags := splitImportTags(csvCell(row, header, "tags"))

		instance, err := resolveImportInstance(org, instanceID, instanceName)
		if err != nil || phone == "" {
			result.Skipped++
			result.Preview = append(result.Preview, ContactImportPreviewRow{Name: name, PhoneNumber: phone, Instance: instanceName, Action: "skipped"})
			continue
		}

		existing := findContactByPhoneAndInstance(org, phone, instance.ID)
		if existing != nil {
			if !updateOnDuplicate {
				result.Skipped++
				result.DuplicatePhones = append(result.DuplicatePhones, phone)
				result.Preview = append(result.Preview, ContactImportPreviewRow{Name: name, PhoneNumber: phone, Instance: instance.Name, Action: "duplicate_skipped"})
				continue
			}
			existing.Contact.Name = strings.TrimSpace(name)
			if existing.Contact.Name == "" {
				existing.Contact.Name = phone
			}
			existing.Contact.Tags = tags
			existing.Contact.InstanceID = instance.ID
			existing.Contact.InstanceName = instance.Name
			existing.Contact.InstanceSourceLabel = instance.Settings.SourceTagLabel
			existing.Contact.LastMessageAt = time.Now()
			s.addConversationEvent(existing, TimelineEvent{
				ID:          s.next("event"),
				EventType:   "contact_import_updated",
				ActorUserID: actorID,
				ActorName:   s.actorName(org, actorID),
				Summary:     "Updated contact through CSV import",
				OccurredAt:  time.Now(),
				Metadata:    map[string]string{"instance_id": instance.ID},
			})
			result.Updated++
			result.Preview = append(result.Preview, ContactImportPreviewRow{Name: existing.Contact.Name, PhoneNumber: phone, Instance: instance.Name, Action: "updated"})
			continue
		}

		contactID := s.next("contact")
		createdAt := time.Now()
		record := &ConversationRecord{
			Contact: ChatContact{
				ID:                  contactID,
				OrganizationID:      orgID,
				Name:                strings.TrimSpace(name),
				PhoneNumber:         phone,
				Avatar:              "https://i.pravatar.cc/150?u=" + contactID,
				Status:              "pending",
				InstanceID:          instance.ID,
				InstanceName:        instance.Name,
				InstanceSourceLabel: instance.Settings.SourceTagLabel,
				LastMessagePreview:  "Imported from CSV.",
				LastMessageAt:       createdAt,
				LastInboundAt:       createdAt,
				IsPublic:            true,
				IsRead:              true,
				Tags:                tags,
				Metadata:            map[string]string{"created_via": "csv_import"},
			},
			Messages:      []ChatMessage{},
			Notes:         []ConversationNote{},
			Collaborators: []Collaborator{},
			Events: []TimelineEvent{
				{
					ID:          s.next("event"),
					EventType:   "contact_import_created",
					ActorUserID: actorID,
					ActorName:   s.actorName(org, actorID),
					Summary:     "Created contact through CSV import",
					OccurredAt:  createdAt,
					Metadata:    map[string]string{"instance_id": instance.ID},
				},
			},
			UserStates: map[string]*ContactUserState{actorID: {}},
		}
		if record.Contact.Name == "" {
			record.Contact.Name = phone
		}
		org.Contacts[contactID] = record
		result.Created++
		result.Preview = append(result.Preview, ContactImportPreviewRow{Name: record.Contact.Name, PhoneNumber: phone, Instance: instance.Name, Action: "created"})
	}

	s.finishJobUnlocked(org, job.ID, "completed", "")
	for i := range org.Jobs {
		if org.Jobs[i].ID == job.ID {
			job = org.Jobs[i]
			break
		}
	}

	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "contacts.import", "contacts", orgID, "Imported contacts from CSV.", map[string]string{
		"created": strconv.Itoa(result.Created),
		"updated": strconv.Itoa(result.Updated),
		"skipped": strconv.Itoa(result.Skipped),
	})
	s.recordOutboxUnlocked(org, "contacts.import.completed", "contacts", orgID, map[string]string{
		"created": strconv.Itoa(result.Created),
		"updated": strconv.Itoa(result.Updated),
	}, false)

	result.Job = job
	if len(result.Preview) > 6 {
		result.Preview = result.Preview[:6]
	}

	return result, nil
}

func csvCell(row []string, header map[string]int, column string) string {
	index, ok := header[column]
	if !ok || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func splitImportTags(value string) []string {
	if value == "" {
		return []string{}
	}
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == '|' || r == ',' })
	tags := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			tags = append(tags, field)
		}
	}
	return tags
}

func resolveImportInstance(org *OrgData, instanceID, instanceName string) (*WhatsAppInstance, error) {
	if instanceID != "" {
		return resolveInstanceForContact(org, instanceID)
	}
	if instanceName != "" {
		for _, instance := range org.Instances {
			if strings.EqualFold(instance.Name, instanceName) {
				return instance, nil
			}
		}
	}
	return resolveInstanceForContact(org, "")
}

func findContactByPhoneAndInstance(org *OrgData, phone, instanceID string) *ConversationRecord {
	for _, record := range org.Contacts {
		if normalizePhoneNumber(record.Contact.PhoneNumber) == phone && record.Contact.InstanceID == instanceID {
			return record
		}
	}
	return nil
}

func (s *Store) ListClosedChats(orgID string, page, pageSize int, agentID, instanceID string) (ClosedChatPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ClosedChatPage{}, errors.New("organization not found")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	rows := make([]ClosedConversationRow, 0)
	for _, record := range org.Contacts {
		if record.Contact.Status != "closed" || record.Contact.ClosedAt == nil {
			continue
		}
		if agentID != "" && record.Contact.AssignedUserID != agentID {
			continue
		}
		if instanceID != "" && record.Contact.InstanceID != instanceID {
			continue
		}
		rows = append(rows, ClosedConversationRow{
			ID:               record.Contact.ID,
			ContactName:      record.Contact.Name,
			PhoneDisplay:     maskedPhoneForOrg(org, record.Contact.PhoneNumber),
			InstanceID:       record.Contact.InstanceID,
			InstanceName:     record.Contact.InstanceName,
			AssignedUserName: record.Contact.AssignedUserName,
			ClosedBy:         actorForEvent(record, "closed"),
			ClosedAt:         *record.Contact.ClosedAt,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ClosedAt.After(rows[j].ClosedAt) })

	start := (page - 1) * pageSize
	if start > len(rows) {
		start = len(rows)
	}
	end := min(len(rows), start+pageSize)

	return ClosedChatPage{
		Items:       append([]ClosedConversationRow{}, rows[start:end]...),
		Page:        page,
		PageSize:    pageSize,
		Total:       len(rows),
		HasNext:     end < len(rows),
		HasPrevious: page > 1,
		AgentID:     agentID,
		InstanceID:  instanceID,
	}, nil
}

func actorForEvent(record *ConversationRecord, eventType string) string {
	for _, event := range record.Events {
		if event.EventType == eventType {
			return event.ActorName
		}
	}
	return "Unknown"
}

func maskedPhoneForOrg(org *OrgData, phone string) string {
	if org.General.MaskPhoneNumbers {
		return maskPhoneNumber(phone)
	}
	return phone
}

func (s *Store) LicenseBootstrap(orgID string) (LicenseBootstrapView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return LicenseBootstrapView{}, errors.New("organization not found")
	}
	return s.licenseBootstrapUnlocked(org), nil
}

func (s *Store) licenseBootstrapUnlocked(org *OrgData) LicenseBootstrapView {
	quotas := []LicenseQuota{
		{
			Resource:  "contacts",
			Label:     "Contacts",
			Current:   len(org.Contacts),
			Limit:     max(1, org.License.Entitlements.MaxContacts),
			OverQuota: len(org.Contacts) > max(1, org.License.Entitlements.MaxContacts),
		},
		{
			Resource:  "campaigns",
			Label:     "Campaigns",
			Current:   len(org.Campaigns),
			Limit:     max(1, org.License.Entitlements.MaxCampaigns),
			OverQuota: len(org.Campaigns) > max(1, org.License.Entitlements.MaxCampaigns),
		},
		{
			Resource:  "instances",
			Label:     "WhatsApp Accounts",
			Current:   len(org.Instances),
			Limit:     max(1, org.License.Entitlements.MaxInstances),
			OverQuota: len(org.Instances) > max(1, org.License.Entitlements.MaxInstances),
		},
	}

	restrictedCleanup := false
	for _, quota := range quotas {
		if quota.OverQuota {
			restrictedCleanup = true
			break
		}
	}

	status := org.License.Status
	message := org.License.Message
	if restrictedCleanup {
		status = "over_limit"
		message = "Usage is above the current license entitlements. Only cleanup actions are available until usage returns within limits."
	}

	return LicenseBootstrapView{
		Status:            status,
		Tier:              org.License.Entitlements.Tier,
		Kind:              org.License.Entitlements.Kind,
		HWID:              org.License.HWID,
		ShortID:           org.License.ShortID,
		LastKeyHint:       org.License.LastKeyHint,
		Message:           message,
		ActivateURL:       org.License.ActivateURL,
		CleanupURL:        org.License.CleanupURL,
		ActivatedAt:       org.License.ActivatedAt,
		ExpiresAt:         org.License.ExpiresAt,
		RestrictedCleanup: restrictedCleanup,
		Quotas:            quotas,
	}
}

func (s *Store) ActivateLicense(orgID, actorID, securityKey string) (LicenseBootstrapView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return LicenseBootstrapView{}, errors.New("organization not found")
	}

	key := strings.TrimSpace(securityKey)
	if key == "" {
		return LicenseBootstrapView{}, errors.New("security key is required")
	}

	now := time.Now()
	org.License.Status = "active"
	org.License.ActivatedAt = &now
	expiresAt := now.Add(45 * 24 * time.Hour)
	org.License.ExpiresAt = &expiresAt
	org.License.LastKeyHint = lastKeyHint(key)

	loweredToCleanup := strings.Contains(strings.ToLower(key), "cleanup") || strings.Contains(strings.ToLower(key), "restrict")
	if loweredToCleanup {
		org.License.Message = "The key is valid, but the entitlements are lower than the current usage. Cleanup is required before normal work resumes."
		org.License.Entitlements.MaxContacts = max(1, len(org.Contacts)-1)
		org.License.Entitlements.MaxCampaigns = max(1, len(org.Campaigns))
		org.License.Entitlements.MaxInstances = max(1, len(org.Instances))
	} else {
		org.License.Message = "License activated successfully."
		org.License.Entitlements.MaxContacts = max(len(org.Contacts)+8, 12)
		org.License.Entitlements.MaxCampaigns = max(len(org.Campaigns)+3, 4)
		org.License.Entitlements.MaxInstances = max(len(org.Instances)+1, 3)
	}
	org.General.MaxInstances = org.License.Entitlements.MaxInstances

	actorName := s.actorName(org, actorID)
	s.addOrgNotification(org, UserNotification{
		ID:          s.next("notif"),
		Title:       "License updated",
		Body:        org.License.Message,
		Severity:    "success",
		RelatedPath: "/settings/license",
		CreatedAt:   now,
	})
	s.recordAuditUnlocked(org, actorID, actorName, "license.activate", "license", orgID, "Applied a new offline license key.", map[string]string{"key_hint": org.License.LastKeyHint})
	s.recordOutboxUnlocked(org, "license.activated", "license", orgID, map[string]string{"key_hint": org.License.LastKeyHint}, false)

	return s.licenseBootstrapUnlocked(org), nil
}

func lastKeyHint(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 6 {
		return strings.ToUpper(key)
	}
	return strings.ToUpper(key[len(key)-6:])
}

func analyticsConversationRecords(org *OrgData, filters AnalyticsFilters) []*ConversationRecord {
	records := make([]*ConversationRecord, 0, len(org.Contacts))
	for _, record := range org.Contacts {
		if filters.InstanceID != "" && record.Contact.InstanceID != filters.InstanceID {
			continue
		}
		if filters.AgentID != "" && record.Contact.AssignedUserID != filters.AgentID && actorForEvent(record, "closed") != userNameForID(org, filters.AgentID) {
			continue
		}
		records = append(records, record)
	}
	return records
}

func userNameForID(org *OrgData, userID string) string {
	if user, ok := org.Users[userID]; ok {
		return user.Name
	}
	return ""
}

func firstMessageAt(record *ConversationRecord) time.Time {
	if len(record.Messages) == 0 {
		return record.Contact.LastMessageAt
	}
	earliest := record.Messages[0].CreatedAt
	for _, message := range record.Messages[1:] {
		if message.CreatedAt.Before(earliest) {
			earliest = message.CreatedAt
		}
	}
	return earliest
}

func eventByType(record *ConversationRecord, eventType string) *TimelineEvent {
	for _, event := range record.Events {
		if event.EventType == eventType {
			copy := event
			return &copy
		}
	}
	return nil
}

func resolutionMinutes(record *ConversationRecord) float64 {
	if closeEvent := eventByType(record, "closed"); closeEvent != nil {
		return closeEvent.OccurredAt.Sub(firstMessageAt(record)).Minutes()
	}
	return 0
}

func queueMinutes(record *ConversationRecord) float64 {
	for _, event := range record.Events {
		if event.EventType == "assigned" || event.EventType == "reopened" {
			return event.OccurredAt.Sub(firstMessageAt(record)).Minutes()
		}
	}
	return 0
}

func transferCount(record *ConversationRecord) int {
	count := 0
	for _, event := range record.Events {
		switch event.EventType {
		case "assigned", "reopened", "unassigned":
			count++
		}
	}
	return count
}

func ratingsForFilters(org *OrgData, filters AnalyticsFilters) []CustomerRating {
	ratings := make([]CustomerRating, 0, len(org.Ratings))
	for _, rating := range org.Ratings {
		record, ok := org.Contacts[rating.ContactID]
		if !ok {
			continue
		}
		if filters.InstanceID != "" && record.Contact.InstanceID != filters.InstanceID {
			continue
		}
		if filters.AgentID != "" && rating.AgentUserID != filters.AgentID {
			continue
		}
		ratings = append(ratings, rating)
	}
	return ratings
}

func (s *Store) AgentAnalyticsSummary(orgID string, filters AnalyticsFilters) (AgentAnalyticsSummaryResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return AgentAnalyticsSummaryResponse{}, errors.New("organization not found")
	}

	records := analyticsConversationRecords(org, filters)
	ratings := ratingsForFilters(org, filters)

	activeCount := 0
	closedCount := 0
	totalQueue := 0.0
	queueSources := 0
	totalResolution := 0.0
	resolutionSources := 0
	totalTransfers := 0
	totalRating := 0.0
	for _, record := range records {
		if record.Contact.Status == "closed" {
			closedCount++
		} else {
			activeCount++
		}
		if value := queueMinutes(record); value > 0 {
			totalQueue += value
			queueSources++
		}
		if value := resolutionMinutes(record); value > 0 {
			totalResolution += value
			resolutionSources++
		}
		totalTransfers += transferCount(record)
	}
	for _, rating := range ratings {
		totalRating += float64(rating.Score)
	}

	avgQueue := 0.0
	if queueSources > 0 {
		avgQueue = totalQueue / float64(queueSources)
	}
	avgResolution := 0.0
	if resolutionSources > 0 {
		avgResolution = totalResolution / float64(resolutionSources)
	}
	avgRating := 0.0
	if len(ratings) > 0 {
		avgRating = totalRating / float64(len(ratings))
	}

	evidenceContacts := make([]string, 0, min(3, len(records)))
	evidenceEvents := make([]string, 0, min(3, len(records)))
	for _, record := range records {
		evidenceContacts = append(evidenceContacts, record.Contact.ID)
		if event := eventByType(record, "closed"); event != nil {
			evidenceEvents = append(evidenceEvents, event.ID)
		}
		if len(evidenceContacts) == 3 {
			break
		}
	}

	return AgentAnalyticsSummaryResponse{
		Cards: []AnalyticsMetricCard{
			{Key: "active_conversations", Label: "Active Conversations", Value: strconv.Itoa(activeCount), Hint: "Derived from live chat status values.", EvidenceCount: len(records)},
			{Key: "closed_conversations", Label: "Closed Conversations", Value: strconv.Itoa(closedCount), Hint: "Derived from `closed` conversation events.", EvidenceCount: closedCount},
			{Key: "average_queue_minutes", Label: "Average Queue", Value: fmt.Sprintf("%.1f min", avgQueue), Hint: "First assignment or reopen minus first recorded message.", EvidenceCount: queueSources},
			{Key: "average_resolution_minutes", Label: "Average Resolution", Value: fmt.Sprintf("%.1f min", avgResolution), Hint: "Close event time minus the first recorded message.", EvidenceCount: resolutionSources},
			{Key: "transfer_count", Label: "Transfers", Value: strconv.Itoa(totalTransfers), Hint: "Count of assignment, reopen, and unassign events.", EvidenceCount: totalTransfers},
			{Key: "average_rating", Label: "Average Rating", Value: fmt.Sprintf("%.1f / 5", avgRating), Hint: "Average of stored customer ratings linked to conversations.", EvidenceCount: len(ratings)},
		},
		Validation: []AnalyticsMetricEvidence{
			{MetricKey: "average_queue_minutes", Explanation: "Queue time is calculated from the first message until the first `assigned` or `reopened` event.", ContactIDs: evidenceContacts, SourceEvents: evidenceEvents},
			{MetricKey: "average_resolution_minutes", Explanation: "Resolution time is calculated from the first message until the `closed` event on the same conversation.", ContactIDs: evidenceContacts, SourceEvents: evidenceEvents},
			{MetricKey: "transfer_count", Explanation: "Transfer count is the total number of `assigned`, `unassigned`, and `reopened` events that passed the active filters.", ContactIDs: evidenceContacts, SourceEvents: evidenceEvents},
			{MetricKey: "average_rating", Explanation: "Average rating comes from stored rating rows linked back to conversation IDs and drill-down paths.", ContactIDs: ratingContactIDs(ratings), SourceEvents: ratingEventIDs(ratings)},
		},
		Generated: time.Now(),
	}, nil
}

func ratingContactIDs(ratings []CustomerRating) []string {
	ids := make([]string, 0, min(3, len(ratings)))
	for _, rating := range ratings {
		ids = append(ids, rating.ContactID)
		if len(ids) == 3 {
			break
		}
	}
	return ids
}

func ratingEventIDs(ratings []CustomerRating) []string {
	ids := make([]string, 0, min(3, len(ratings)))
	for _, rating := range ratings {
		ids = append(ids, rating.SourceEventID)
		if len(ids) == 3 {
			break
		}
	}
	return ids
}

func (s *Store) AgentTransferTrends(orgID string, filters AnalyticsFilters) ([]AnalyticsPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	counts := map[string]int{}
	for _, record := range analyticsConversationRecords(org, filters) {
		for _, event := range record.Events {
			if event.EventType != "assigned" && event.EventType != "reopened" && event.EventType != "unassigned" {
				continue
			}
			label := event.OccurredAt.Format("Jan 02")
			counts[label]++
		}
	}

	points := make([]AnalyticsPoint, 0, len(counts))
	for label, value := range counts {
		points = append(points, AnalyticsPoint{Label: label, Value: value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Label < points[j].Label })
	return points, nil
}

func (s *Store) AgentSourceBreakdown(orgID string, filters AnalyticsFilters) ([]AnalyticsPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	counts := map[string]int{}
	for _, record := range analyticsConversationRecords(org, filters) {
		label := record.Contact.InstanceSourceLabel
		if label == "" {
			label = record.Contact.InstanceName
		}
		if label == "" {
			label = "Unknown"
		}
		counts[label]++
	}
	points := make([]AnalyticsPoint, 0, len(counts))
	for label, value := range counts {
		points = append(points, AnalyticsPoint{Label: label, Value: value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Value > points[j].Value })
	return points, nil
}

func (s *Store) AgentComparison(orgID string, filters AnalyticsFilters) ([]AgentComparisonRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	rows := make([]AgentComparisonRow, 0, len(org.Users))
	for _, user := range org.Users {
		if filters.AgentID != "" && filters.AgentID != user.ID {
			continue
		}

		assignedRecords := make([]*ConversationRecord, 0)
		for _, record := range org.Contacts {
			if filters.InstanceID != "" && record.Contact.InstanceID != filters.InstanceID {
				continue
			}
			if record.Contact.AssignedUserID == user.ID {
				assignedRecords = append(assignedRecords, record)
			}
		}

		active := 0
		closed := 0
		transfers := 0
		totalQueue := 0.0
		queueCount := 0
		totalResolution := 0.0
		resolutionCount := 0
		for _, record := range assignedRecords {
			if record.Contact.Status == "closed" {
				closed++
			} else {
				active++
			}
			transfers += transferCount(record)
			if value := queueMinutes(record); value > 0 {
				totalQueue += value
				queueCount++
			}
			if value := resolutionMinutes(record); value > 0 {
				totalResolution += value
				resolutionCount++
			}
		}

		avgQueue := 0.0
		if queueCount > 0 {
			avgQueue = totalQueue / float64(queueCount)
		}
		avgResolution := 0.0
		if resolutionCount > 0 {
			avgResolution = totalResolution / float64(resolutionCount)
		}

		agentRatings := make([]CustomerRating, 0)
		for _, rating := range org.Ratings {
			if rating.AgentUserID == user.ID {
				if filters.InstanceID != "" {
					record, ok := org.Contacts[rating.ContactID]
					if !ok || record.Contact.InstanceID != filters.InstanceID {
						continue
					}
				}
				agentRatings = append(agentRatings, rating)
			}
		}
		avgRating := 0.0
		if len(agentRatings) > 0 {
			total := 0.0
			for _, rating := range agentRatings {
				total += float64(rating.Score)
			}
			avgRating = total / float64(len(agentRatings))
		}

		rows = append(rows, AgentComparisonRow{
			AgentID:               user.ID,
			AgentName:             user.Name,
			ActiveConversations:   active,
			ClosedConversations:   closed,
			Transfers:             transfers,
			AverageQueueMinutes:   avgQueue,
			AverageResolutionMins: avgResolution,
			AverageRating:         avgRating,
		})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].AgentName < rows[j].AgentName })
	return rows, nil
}

func (s *Store) AgentRatings(orgID string, filters AnalyticsFilters) ([]CustomerRating, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	ratings := ratingsForFilters(org, filters)
	sort.Slice(ratings, func(i, j int) bool { return ratings[i].RatedAt.After(ratings[j].RatedAt) })
	return ratings, nil
}

func (s *Store) ExportAgentAnalyticsCSV(orgID string, filters AnalyticsFilters) (string, error) {
	rows, err := s.AgentComparison(orgID, filters)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"agent_name", "active_conversations", "closed_conversations", "transfers", "average_queue_minutes", "average_resolution_minutes", "average_rating"}); err != nil {
		return "", err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.AgentName,
			strconv.Itoa(row.ActiveConversations),
			strconv.Itoa(row.ClosedConversations),
			strconv.Itoa(row.Transfers),
			fmt.Sprintf("%.2f", row.AverageQueueMinutes),
			fmt.Sprintf("%.2f", row.AverageResolutionMins),
			fmt.Sprintf("%.2f", row.AverageRating),
		}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	return buffer.String(), writer.Error()
}

func (s *Store) ListCampaigns(orgID string) ([]Campaign, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	campaigns := make([]Campaign, 0, len(org.Campaigns))
	for _, record := range org.Campaigns {
		campaigns = append(campaigns, record.Campaign)
	}
	sort.Slice(campaigns, func(i, j int) bool { return campaigns[i].UpdatedAt.After(campaigns[j].UpdatedAt) })
	return campaigns, nil
}

func (s *Store) CampaignDetail(orgID, campaignID string) (CampaignRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return CampaignRecord{}, errors.New("organization not found")
	}
	record, ok := org.Campaigns[campaignID]
	if !ok {
		return CampaignRecord{}, errors.New("campaign not found")
	}
	copied := CampaignRecord{
		Campaign:   record.Campaign,
		Runs:       append([]CampaignRun{}, record.Runs...),
		Recipients: map[string][]CampaignRecipient{},
	}
	for runID, recipients := range record.Recipients {
		copied.Recipients[runID] = append([]CampaignRecipient{}, recipients...)
	}
	return copied, nil
}

func (s *Store) CreateCampaign(orgID, actorID string, req CampaignUpsertRequest) (Campaign, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return Campaign{}, errors.New("organization not found")
	}

	campaign := Campaign{
		ID:             s.next("campaign"),
		Name:           strings.TrimSpace(req.Name),
		Status:         strings.TrimSpace(req.Status),
		Source:         strings.TrimSpace(req.Source),
		Content:        strings.TrimSpace(req.Content),
		Filters:        req.Filters,
		Schedule:       req.Schedule,
		LastRunSummary: "No runs yet",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if campaign.Name == "" {
		return Campaign{}, errors.New("campaign name is required")
	}
	if campaign.Status == "" {
		campaign.Status = "draft"
	}
	if campaign.Source == "" {
		campaign.Source = "manual"
	}
	org.Campaigns[campaign.ID] = &CampaignRecord{
		Campaign:   campaign,
		Runs:       []CampaignRun{},
		Recipients: map[string][]CampaignRecipient{},
	}

	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "campaigns.create", "campaign", campaign.ID, "Created a reusable campaign definition.", map[string]string{"status": campaign.Status})
	s.recordOutboxUnlocked(org, "campaigns.created", "campaign", campaign.ID, map[string]string{"status": campaign.Status}, false)
	return campaign, nil
}

func (s *Store) UpdateCampaign(orgID, actorID, campaignID string, req CampaignUpsertRequest) (Campaign, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return Campaign{}, errors.New("organization not found")
	}
	record, ok := org.Campaigns[campaignID]
	if !ok {
		return Campaign{}, errors.New("campaign not found")
	}

	record.Campaign.Name = strings.TrimSpace(req.Name)
	record.Campaign.Content = strings.TrimSpace(req.Content)
	record.Campaign.Filters = req.Filters
	record.Campaign.Schedule = req.Schedule
	if strings.TrimSpace(req.Status) != "" {
		record.Campaign.Status = strings.TrimSpace(req.Status)
	}
	if strings.TrimSpace(req.Source) != "" {
		record.Campaign.Source = strings.TrimSpace(req.Source)
	}
	record.Campaign.UpdatedAt = time.Now()

	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "campaigns.edit", "campaign", campaignID, "Updated the campaign definition.", map[string]string{"status": record.Campaign.Status})
	s.recordOutboxUnlocked(org, "campaigns.updated", "campaign", campaignID, map[string]string{"status": record.Campaign.Status}, false)
	return record.Campaign, nil
}

func (s *Store) DeleteCampaign(orgID, actorID, campaignID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return errors.New("organization not found")
	}
	record, ok := org.Campaigns[campaignID]
	if !ok {
		return errors.New("campaign not found")
	}

	delete(org.Campaigns, campaignID)
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "campaigns.delete", "campaign", campaignID, "Deleted a campaign definition.", map[string]string{
		"name": record.Campaign.Name,
	})
	s.recordOutboxUnlocked(org, "campaigns.deleted", "campaign", campaignID, map[string]string{
		"name": record.Campaign.Name,
	}, false)
	return nil
}

func (s *Store) LaunchCampaign(orgID, actorID, campaignID string) (CampaignRun, []CampaignRecipient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return CampaignRun{}, nil, errors.New("organization not found")
	}
	record, ok := org.Campaigns[campaignID]
	if !ok {
		return CampaignRun{}, nil, errors.New("campaign not found")
	}

	recipients := targetCampaignRecipients(org, record.Campaign)
	job := s.recordJobUnlocked(org, "campaign_run", "campaign", campaignID, "Executing a campaign run.")
	startedAt := time.Now()
	finishedAt := startedAt
	run := CampaignRun{
		ID:             s.next("run"),
		CampaignID:     campaignID,
		Trigger:        "manual",
		Status:         "completed",
		JobID:          job.ID,
		StartedAt:      startedAt,
		FinishedAt:     &finishedAt,
		RecipientTotal: len(recipients),
	}

	delivered := 0
	failed := 0
	storedRecipients := make([]CampaignRecipient, 0, len(recipients))
	for _, contact := range recipients {
		recipient := CampaignRecipient{
			ID:             s.next("recipient"),
			RunID:          run.ID,
			ContactID:      contact.ID,
			ContactName:    contact.Name,
			PhoneNumber:    contact.PhoneNumber,
			MessagePreview: record.Campaign.Content,
		}
		instance := org.Instances[contact.InstanceID]
		if instance != nil && instance.Status == "connected" {
			recipient.Status = "delivered"
			deliveredAt := time.Now()
			recipient.DeliveredAt = &deliveredAt
			delivered++
		} else {
			recipient.Status = "failed"
			recipient.FailureReason = "Instance not connected"
			failed++
		}
		storedRecipients = append(storedRecipients, recipient)
	}

	run.Delivered = delivered
	run.Failed = failed
	record.Runs = append([]CampaignRun{run}, record.Runs...)
	if record.Recipients == nil {
		record.Recipients = map[string][]CampaignRecipient{}
	}
	record.Recipients[run.ID] = storedRecipients
	record.Campaign.Status = "active"
	record.Campaign.LastRunSummary = fmt.Sprintf("%d recipients, %d delivered, %d failed", len(storedRecipients), delivered, failed)
	record.Campaign.UpdatedAt = time.Now()
	s.finishJobUnlocked(org, job.ID, "completed", "")

	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "campaigns.launch", "campaign", campaignID, "Launched a campaign run.", map[string]string{"run_id": run.ID})
	s.recordOutboxUnlocked(org, "campaigns.launched", "campaign", campaignID, map[string]string{"run_id": run.ID}, true)
	return run, storedRecipients, nil
}

func targetCampaignRecipients(org *OrgData, campaign Campaign) []ChatContact {
	recipients := make([]ChatContact, 0)
	for _, record := range org.Contacts {
		if campaign.Filters.InstanceID != "" && record.Contact.InstanceID != campaign.Filters.InstanceID {
			continue
		}
		if campaign.Filters.Status != "" && record.Contact.Status != campaign.Filters.Status {
			continue
		}
		if campaign.Filters.Tag != "" && !containsFold(record.Contact.Tags, campaign.Filters.Tag) {
			continue
		}
		if campaign.Filters.Search != "" {
			search := strings.ToLower(campaign.Filters.Search)
			if !strings.Contains(strings.ToLower(record.Contact.Name), search) && !strings.Contains(strings.ToLower(record.Contact.PhoneNumber), search) {
				continue
			}
		}
		if !campaign.Filters.IncludeClosed && record.Contact.Status == "closed" {
			continue
		}
		recipients = append(recipients, record.Contact)
	}
	sort.Slice(recipients, func(i, j int) bool { return recipients[i].Name < recipients[j].Name })
	return recipients
}

func containsFold(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func (s *Store) PauseCampaign(orgID, actorID, campaignID string) (Campaign, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return Campaign{}, errors.New("organization not found")
	}
	record, ok := org.Campaigns[campaignID]
	if !ok {
		return Campaign{}, errors.New("campaign not found")
	}
	record.Campaign.Status = "paused"
	record.Campaign.UpdatedAt = time.Now()
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "campaigns.pause", "campaign", campaignID, "Paused a campaign.", nil)
	s.recordOutboxUnlocked(org, "campaigns.paused", "campaign", campaignID, map[string]string{"status": "paused"}, false)
	return record.Campaign, nil
}

func (s *Store) ResumeCampaign(orgID, actorID, campaignID string) (Campaign, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return Campaign{}, errors.New("organization not found")
	}
	record, ok := org.Campaigns[campaignID]
	if !ok {
		return Campaign{}, errors.New("campaign not found")
	}
	record.Campaign.Status = "scheduled"
	record.Campaign.UpdatedAt = time.Now()
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "campaigns.resume", "campaign", campaignID, "Resumed a campaign.", nil)
	s.recordOutboxUnlocked(org, "campaigns.resumed", "campaign", campaignID, map[string]string{"status": "scheduled"}, false)
	return record.Campaign, nil
}

func (s *Store) ListCampaignRuns(orgID, campaignID string) ([]CampaignRun, error) {
	record, err := s.CampaignDetail(orgID, campaignID)
	if err != nil {
		return nil, err
	}
	return record.Runs, nil
}

func (s *Store) ListCampaignRecipients(orgID, campaignID, runID string) ([]CampaignRecipient, error) {
	record, err := s.CampaignDetail(orgID, campaignID)
	if err != nil {
		return nil, err
	}
	if runID == "" && len(record.Runs) > 0 {
		runID = record.Runs[0].ID
	}
	recipients := record.Recipients[runID]
	return append([]CampaignRecipient{}, recipients...), nil
}

func (s *Store) syncAutoCampaignUnlocked(org *OrgData, instanceID string) {
	instance, ok := org.Instances[instanceID]
	if !ok {
		return
	}
	for _, record := range org.Campaigns {
		if record.Campaign.LinkedInstanceID != instanceID {
			continue
		}
		record.Campaign.Name = fmt.Sprintf("%s %s", instance.AutoCampaign.CampaignNamePrefix, instance.Name)
		record.Campaign.Content = instance.AutoCampaign.MessageBody
		record.Campaign.Filters.InstanceID = instance.ID
		record.Campaign.Schedule.Mode = "every_n_days"
		record.Campaign.Schedule.EveryDays = max(1, instance.AutoCampaign.ScheduleEveryDays)
		if instance.AutoCampaign.Enabled {
			record.Campaign.Status = "scheduled"
		} else {
			record.Campaign.Status = "disabled"
		}
		record.Campaign.Source = "instance_auto_campaign"
		record.Campaign.UpdatedAt = time.Now()
		return
	}

	campaignStatus := "disabled"
	if instance.AutoCampaign.Enabled {
		campaignStatus = "scheduled"
	}
	record := &CampaignRecord{
		Campaign: Campaign{
			ID:               s.next("campaign"),
			Name:             fmt.Sprintf("%s %s", instance.AutoCampaign.CampaignNamePrefix, instance.Name),
			Status:           campaignStatus,
			Source:           "instance_auto_campaign",
			LinkedInstanceID: instance.ID,
			Content:          instance.AutoCampaign.MessageBody,
			Filters:          CampaignFilters{InstanceID: instance.ID},
			Schedule:         CampaignSchedule{Mode: "every_n_days", EveryDays: max(1, instance.AutoCampaign.ScheduleEveryDays), TimeOfDay: "10:00"},
			LastRunSummary:   "Linked from account automation",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		Runs:       []CampaignRun{},
		Recipients: map[string][]CampaignRecipient{},
	}
	org.Campaigns[record.Campaign.ID] = record
}

func (s *Store) DeleteInstance(orgID, actorID, instanceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return errors.New("instance not found")
	}
	if instance.Status == "connected" {
		return errors.New("disconnect the account before deleting it")
	}
	for _, record := range org.Contacts {
		if record.Contact.InstanceID == instanceID {
			return errors.New("remove or reassign contacts linked to this account before deleting it")
		}
	}

	delete(org.Instances, instanceID)
	org.General.UsedInstances = len(org.Instances)
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.delete", "instance", instanceID, "Deleted a WhatsApp account.", map[string]string{"name": instance.Name})
	s.recordOutboxUnlocked(org, "instances.deleted", "instance", instanceID, map[string]string{"name": instance.Name}, false)
	return nil
}

func (s *Store) ListJobs(orgID string) ([]BackgroundJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return append([]BackgroundJob{}, org.Jobs...), nil
}

func (s *Store) JobByID(orgID, jobID string) (BackgroundJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return BackgroundJob{}, errors.New("organization not found")
	}
	for _, job := range org.Jobs {
		if job.ID == jobID {
			return job, nil
		}
	}
	return BackgroundJob{}, errors.New("job not found")
}

func (s *Store) ListWebhooks(orgID string) ([]WebhookEndpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	webhooks := make([]WebhookEndpoint, 0, len(org.Webhooks))
	for _, webhook := range org.Webhooks {
		webhooks = append(webhooks, *webhook)
	}
	sort.Slice(webhooks, func(i, j int) bool { return webhooks[i].Name < webhooks[j].Name })
	return webhooks, nil
}

func (s *Store) ListWebhookDeliveries(orgID, webhookID string) ([]WebhookDelivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	deliveries := make([]WebhookDelivery, 0)
	for _, delivery := range org.Deliveries {
		if webhookID == "" || delivery.WebhookID == webhookID {
			deliveries = append(deliveries, delivery)
		}
	}
	return deliveries, nil
}

func (s *Store) RetryWebhookDelivery(orgID, actorID, webhookID, deliveryID string) (WebhookDelivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WebhookDelivery{}, errors.New("organization not found")
	}
	for i := range org.Deliveries {
		if org.Deliveries[i].ID != deliveryID || org.Deliveries[i].WebhookID != webhookID {
			continue
		}
		now := time.Now()
		org.Deliveries[i].Attempt++
		org.Deliveries[i].LastAttemptAt = now
		org.Deliveries[i].NextRetryAt = nil
		org.Deliveries[i].Status = "delivered"
		org.Deliveries[i].ResponseCode = 202
		org.Deliveries[i].ResponseBody = "manual retry succeeded"

		for j := range org.Outbox {
			if org.Outbox[j].ID == org.Deliveries[i].EventID {
				org.Outbox[j].Status = "delivered"
				org.Outbox[j].LastDeliveryStatus = "delivered"
			}
		}
		s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "deliveries.retry", "delivery", deliveryID, "Retried a failed webhook delivery.", map[string]string{"webhook_id": webhookID})
		return org.Deliveries[i], nil
	}
	return WebhookDelivery{}, errors.New("delivery not found")
}

func (s *Store) ListAuditLogs(orgID string) ([]AuditLogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return append([]AuditLogEntry{}, org.Audit...), nil
}
