package api

import "time"

// ---------- WORKSPACE / USERS ----------

type WorkspaceUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Avatar   string `json:"avatar"`
	IsActive bool   `json:"is_active"`
}

// ---------- CHAT / CONTACTS ----------

type ChatContact struct {
	ID                  string            `json:"id"`
	OrganizationID      string            `json:"organization_id"`
	Name                string            `json:"name"`
	PhoneNumber         string            `json:"phone_number"`
	PhoneDisplay        string            `json:"phone_display"`
	Avatar              string            `json:"avatar"`
	Status              string            `json:"status"`
	AssignedUserID      string            `json:"assigned_user_id"`
	AssignedUserName    string            `json:"assigned_user_name"`
	InstanceID          string            `json:"instance_id"`
	InstanceName        string            `json:"instance_name"`
	InstanceSourceLabel string            `json:"instance_source_label"`
	LastMessagePreview  string            `json:"last_message_preview"`
	LastMessageAt       time.Time         `json:"last_message_at"`
	LastInboundAt       time.Time         `json:"last_inbound_at"`
	ClosedAt            *time.Time        `json:"closed_at,omitempty"`
	IsPublic            bool              `json:"is_public"`
	IsRead              bool              `json:"is_read"`
	IsPinned            bool              `json:"is_pinned"`
	IsHidden            bool              `json:"is_hidden"`
	LastReadMessageID   string            `json:"last_read_message_id"`
	UnreadCount         int               `json:"unread_count"`
	Tags                []string          `json:"tags"`
	Metadata            map[string]string `json:"metadata"`
	CreatedAt           time.Time         `json:"created_at"`
}

type ChatMessage struct {
	ID            string     `json:"id"`
	ContactID     string     `json:"contact_id"`
	Direction     string     `json:"direction"`
	Type          string     `json:"type"`
	Body          string     `json:"body"`
	Status        string     `json:"status"`
	FileName      string     `json:"file_name,omitempty"`
	FileSizeLabel string     `json:"file_size_label,omitempty"`
	MediaURL      string     `json:"media_url,omitempty"`
	FailureReason string     `json:"failure_reason,omitempty"`
	RetryCount    int        `json:"retry_count"`
	TypedForMs    int        `json:"typed_for_ms"`
	CreatedAt     time.Time  `json:"created_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CanRetry      bool       `json:"can_retry"`
	CanRevoke     bool       `json:"can_revoke"`
	Reaction      string     `json:"reaction,omitempty"`
}

type ConversationNote struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Collaborator struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Status    string    `json:"status"`
	InvitedAt time.Time `json:"invited_at"`
}

type TimelineEvent struct {
	ID          string            `json:"id"`
	EventType   string            `json:"event_type"`
	ActorUserID string            `json:"actor_user_id"`
	ActorName   string            `json:"actor_name"`
	Summary     string            `json:"summary"`
	OccurredAt  time.Time         `json:"occurred_at"`
	Metadata    map[string]string `json:"metadata"`
}

type QuickReply struct {
	ID       string `json:"id"`
	Shortcut string `json:"shortcut"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

type UserNotification struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	Severity         string    `json:"severity"`
	RelatedContactID string    `json:"related_contact_id,omitempty"`
	RelatedPath      string    `json:"related_path,omitempty"`
	IsRead           bool      `json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}

type StatusPost struct {
	ID           string    `json:"id"`
	ContactID    string    `json:"contact_id"`
	ContactName  string    `json:"contact_name"`
	InstanceID   string    `json:"instance_id"`
	InstanceName string    `json:"instance_name"`
	Body         string    `json:"body"`
	Kind         string    `json:"kind"`
	CreatedAt    time.Time `json:"created_at"`
}

// ---------- SETTINGS ----------

type GeneralSettings struct {
	OrganizationName  string `json:"organization_name"`
	Slug              string `json:"slug"`
	Timezone          string `json:"timezone"`
	DateFormat        string `json:"date_format"`
	Locale            string `json:"locale"`
	MaskPhoneNumbers  bool   `json:"mask_phone_numbers"`
	TenantStatus      string `json:"tenant_status"`
	ActiveMembers     int    `json:"active_members"`
	MaxMembers        int    `json:"max_members"`
	UsedInstances     int    `json:"used_instances"`
	MaxInstances      int    `json:"max_instances"`
	StorageUsedLabel  string `json:"storage_used_label"`
	StorageLimitLabel string `json:"storage_limit_label"`
}

type AppearanceSettings struct {
	ColorMode   string `json:"color_mode"`
	ThemePreset string `json:"theme_preset"`
}

type ChatSettings struct {
	MediaGroupingWindowMinutes int    `json:"media_grouping_window_minutes"`
	SidebarContactView         string `json:"sidebar_contact_view"`
	SidebarHoverExpand         bool   `json:"sidebar_hover_expand"`
	PinSidebar                 bool   `json:"pin_sidebar"`
	ChatBackground             string `json:"chat_background"`
	ShowPrintButtons           bool   `json:"show_print_buttons"`
	ShowDownloadButtons        bool   `json:"show_download_buttons"`
}

type NotificationSettings struct {
	EmailNotifications bool   `json:"email_notifications"`
	NewMessageAlerts   bool   `json:"new_message_alerts"`
	NotificationSound  string `json:"notification_sound"`
	CampaignUpdates    bool   `json:"campaign_updates"`
}

type CleanupSettings struct {
	RetentionDays int        `json:"retention_days"`
	RunHour       int        `json:"run_hour"`
	Timezone      string     `json:"timezone"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	LastJobStatus string     `json:"last_job_status"`
}

type SettingsSummary struct {
	General       GeneralSettings      `json:"general"`
	Appearance    AppearanceSettings   `json:"appearance"`
	Chat          ChatSettings         `json:"chat"`
	Notifications NotificationSettings `json:"notifications"`
	Cleanup       CleanupSettings      `json:"cleanup"`
	Team          []WorkspaceUser      `json:"team"`
	QuickReplies  []QuickReply         `json:"quick_replies"`
}

// ---------- WHATSAPP INSTANCES ----------

type InstanceSettings struct {
	AutoSyncHistory           bool     `json:"auto_sync_history"`
	AutoDownloadIncomingMedia bool     `json:"auto_download_incoming_media"`
	SourceTagLabel            string   `json:"source_tag_label"`
	SourceTagDisplayMode      string   `json:"source_tag_display_mode"`
	SourceTagColor            string   `json:"source_tag_color"`
	AllowedSendModes          []string `json:"allowed_send_modes"`
}

type InstanceHealth struct {
	Status        string    `json:"status"`
	UptimeLabel   string    `json:"uptime_label"`
	QueueDepth    int       `json:"queue_depth"`
	SentToday     int       `json:"sent_today"`
	ReceivedToday int       `json:"received_today"`
	FailedToday   int       `json:"failed_today"`
	ErrorRate     string    `json:"error_rate"`
	ObservedAt    time.Time `json:"observed_at"`
}

type InstanceCallPolicy struct {
	Enabled               bool     `json:"enabled"`
	RejectIndividualCalls bool     `json:"reject_individual_calls"`
	RejectGroupCalls      bool     `json:"reject_group_calls"`
	ReplyMode             string   `json:"reply_mode"`
	ScheduleMode          string   `json:"schedule_mode"`
	EmergencyBypass       []string `json:"emergency_bypass"`
	ReplyMessage          string   `json:"reply_message"`
}

type InstanceAutoCampaign struct {
	Enabled            bool   `json:"enabled"`
	CampaignNamePrefix string `json:"campaign_name_prefix"`
	ScheduleEveryDays  int    `json:"schedule_every_days"`
	DelayFromMinutes   int    `json:"delay_from_minutes"`
	DelayToMinutes     int    `json:"delay_to_minutes"`
	CampaignStatus     string `json:"campaign_status"`
	MessageBody        string `json:"message_body"`
}

type InstanceRatingSettings struct {
	Enabled               bool   `json:"enabled"`
	FollowUpWindowMinutes int    `json:"follow_up_window_minutes"`
	TemplateAR            string `json:"template_ar"`
	TemplateEN            string `json:"template_en"`
}

type InstanceAssignmentReset struct {
	Enabled      bool   `json:"enabled"`
	ScheduleMode string `json:"schedule_mode"`
	Timezone     string `json:"timezone"`
}

type WhatsAppInstance struct {
	ID              string                  `json:"id"`
	OrganizationID  string                  `json:"organization_id"`
	Name            string                  `json:"name"`
	PhoneNumber     string                  `json:"phone_number"`
	JID             string                  `json:"jid"`
	Status          string                  `json:"status"`
	PairingState    string                  `json:"pairing_state"`
	QRCode          string                  `json:"qr_code,omitempty"`
	SlotBlocked     bool                    `json:"slot_blocked"`
	Settings        InstanceSettings        `json:"settings"`
	Health          InstanceHealth          `json:"health"`
	CallPolicy      InstanceCallPolicy      `json:"call_policy"`
	AutoCampaign    InstanceAutoCampaign    `json:"auto_campaign"`
	RatingSettings  InstanceRatingSettings  `json:"rating_settings"`
	AssignmentReset InstanceAssignmentReset `json:"assignment_reset"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type InstanceHealthSummary struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	UptimeLabel   string    `json:"uptime_label"`
	QueueDepth    int       `json:"queue_depth"`
	SentToday     int       `json:"sent_today"`
	ReceivedToday int       `json:"received_today"`
	FailedToday   int       `json:"failed_today"`
	ErrorRate     string    `json:"error_rate"`
	ObservedAt    time.Time `json:"observed_at"`
}

// ---------- CONVERSATIONS ----------

type ConversationDetail struct {
	Contact       ChatContact        `json:"contact"`
	Messages      []ChatMessage      `json:"messages"`
	Notes         []ConversationNote `json:"notes"`
	Collaborators []Collaborator     `json:"collaborators"`
	Events        []TimelineEvent    `json:"events"`
}

type WorkspaceView struct {
	CurrentTab    string               `json:"current_tab"`
	TabCounts     map[string]int       `json:"tab_counts"`
	Filters       map[string]string    `json:"filters"`
	Conversations []ChatContact        `json:"conversations"`
	Selected      *ConversationDetail  `json:"selected,omitempty"`
	Users         []WorkspaceUser      `json:"users"`
	Instances     []WhatsAppInstance   `json:"instances"`
	Statuses      []StatusPost         `json:"statuses"`
	Notifications []UserNotification   `json:"notifications"`
	QuickReplies  []QuickReply         `json:"quick_replies"`
	Settings      SettingsSummary      `json:"settings"`
}

// ---------- CONTACTS ----------

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

// ---------- LICENSE ----------

type LicenseEntitlements struct {
	MaxContacts  int    `json:"max_contacts"`
	MaxCampaigns int    `json:"max_campaigns"`
	MaxInstances int    `json:"max_instances"`
	Tier         string `json:"tier"`
	Kind         string `json:"kind"`
}

type LicenseQuota struct {
	Resource  string `json:"resource"`
	Label     string `json:"label"`
	Current   int    `json:"current"`
	Limit     int    `json:"limit"`
	OverQuota bool   `json:"over_quota"`
}

type LicenseBootstrapView struct {
	Status            string               `json:"status"`
	Tier              string               `json:"tier"`
	Kind              string               `json:"kind"`
	HWID              string               `json:"hwid"`
	ShortID           string               `json:"short_id"`
	LastKeyHint       string               `json:"last_key_hint"`
	Message           string               `json:"message"`
	ActivateURL       string               `json:"activate_url"`
	CleanupURL        string               `json:"cleanup_url"`
	ActivatedAt       *time.Time           `json:"activated_at,omitempty"`
	ExpiresAt         *time.Time           `json:"expires_at,omitempty"`
	RestrictedCleanup bool                 `json:"restricted_cleanup"`
	Quotas            []LicenseQuota       `json:"quotas"`
	Entitlements      LicenseEntitlements  `json:"entitlements"`
}

type LicenseActivationRequest struct {
	SecurityKey string `json:"security_key"`
}

// ---------- ANALYTICS ----------

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

// ---------- CAMPAIGNS ----------

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

// ---------- JOBS / EVENTS / WEBHOOKS / AUDIT ----------

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

// ---------- HTTP REQUEST TYPES ----------

type SendMessageRequest struct {
	Type          string `json:"type"`
	Body          string `json:"body"`
	FileName      string `json:"file_name"`
	FileSizeLabel string `json:"file_size_label"`
	MediaURL      string `json:"media_url"`
}

type CreateDirectChatRequest struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	InstanceID  string `json:"instance_id"`
}

type AddStatusRequest struct {
	ContactID    string `json:"contact_id"`
	ContactName  string `json:"contact_name"`
	InstanceID   string `json:"instance_id"`
	InstanceName string `json:"instance_name"`
	Body         string `json:"body"`
	Kind         string `json:"kind"`
}

type CreateInstanceRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type InstanceSettingsRequest struct {
	Settings        InstanceSettings        `json:"settings"`
	RatingSettings  InstanceRatingSettings  `json:"rating_settings"`
	AssignmentReset InstanceAssignmentReset `json:"assignment_reset"`
}

type CallPolicyRequest struct {
	CallPolicy InstanceCallPolicy `json:"call_policy"`
}

type AutoCampaignRequest struct {
	AutoCampaign InstanceAutoCampaign `json:"auto_campaign"`
}

type UpdateProfileRequest struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Language    string `json:"language"`
	ThemePreset string `json:"theme_preset"`
}
