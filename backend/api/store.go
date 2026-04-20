//go:build ignore

package api

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WorkspaceUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Avatar   string `json:"avatar"`
	IsActive bool   `json:"is_active"`
}

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
	UnreadCount         int               `json:"unread_count"`
	Tags                []string          `json:"tags"`
	Metadata            map[string]string `json:"metadata"`
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
	TypedForMS    int        `json:"typed_for_ms"`
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
}

type InstanceSettings struct {
	AutoSyncHistory           bool   `json:"auto_sync_history"`
	AutoDownloadIncomingMedia bool   `json:"auto_download_incoming_media"`
	SourceTagLabel            string `json:"source_tag_label"`
	SourceTagDisplayMode      string `json:"source_tag_display_mode"`
	SourceTagColor            string `json:"source_tag_color"`
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
}

type CleanupJob struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

type ConversationDetail struct {
	Contact       ChatContact        `json:"contact"`
	Messages      []ChatMessage      `json:"messages"`
	Notes         []ConversationNote `json:"notes"`
	Collaborators []Collaborator     `json:"collaborators"`
	Events        []TimelineEvent    `json:"events"`
}

type WorkspaceSnapshot struct {
	CurrentTab    string              `json:"current_tab"`
	TabCounts     map[string]int      `json:"tab_counts"`
	Filters       map[string]string   `json:"filters"`
	Conversations []ChatContact       `json:"conversations"`
	Selected      *ConversationDetail `json:"selected,omitempty"`
	Notifications []UserNotification  `json:"notifications"`
	Statuses      []StatusPost        `json:"statuses"`
	QuickReplies  []QuickReply        `json:"quick_replies"`
	Instances     []WhatsAppInstance  `json:"instances"`
	Users         []WorkspaceUser     `json:"users"`
	Settings      SettingsSummary     `json:"settings"`
}

type ProfileView struct {
	User     UserResponse    `json:"user"`
	Settings SettingsSummary `json:"settings"`
}

type ContactUserState struct {
	IsHidden          bool
	IsPinned          bool
	LastReadMessageID string
	LastOpenedAt      time.Time
	LastSeenAt        time.Time
}

type ConversationRecord struct {
	Contact       ChatContact
	Messages      []ChatMessage
	Notes         []ConversationNote
	Collaborators []Collaborator
	Events        []TimelineEvent
	UserStates    map[string]*ContactUserState
}

type OrgData struct {
	General           GeneralSettings
	Appearance        AppearanceSettings
	Chat              ChatSettings
	Notifications     NotificationSettings
	Cleanup           CleanupSettings
	Contacts          map[string]*ConversationRecord
	NotificationsFeed []UserNotification
	Statuses          []StatusPost
	Instances         map[string]*WhatsAppInstance
	QuickReplies      []QuickReply
	Users             map[string]*WorkspaceUser
	License           LicenseRecord
	Campaigns         map[string]*CampaignRecord
	Jobs              []BackgroundJob
	Webhooks          map[string]*WebhookEndpoint
	Deliveries        []WebhookDelivery
	Outbox            []OutboxEvent
	Audit             []AuditLogEntry
	Ratings           []CustomerRating
}

type Store struct {
	mu     sync.RWMutex
	nextID int
	orgs   map[string]*OrgData
}

func NewStore() *Store {
	now := time.Now()

	return &Store{
		nextID: 100,
		orgs: map[string]*OrgData{
			"org-1": seedPrimaryOrg(now),
			"org-2": seedSecondaryOrg(now),
		},
	}
}

func seedPrimaryOrg(now time.Time) *OrgData {
	users := map[string]*WorkspaceUser{
		"1": {ID: "1", Name: "Admin Encanto", Email: mockEmail, Role: "admin", Status: "online", Avatar: "https://i.pravatar.cc/150?u=admin", IsActive: true},
		"2": {ID: "2", Name: "Maha Support", Email: "maha@example.com", Role: "agent", Status: "busy", Avatar: "https://i.pravatar.cc/150?u=maha", IsActive: true},
		"3": {ID: "3", Name: "Omar Care", Email: "omar@example.com", Role: "agent", Status: "available", Avatar: "https://i.pravatar.cc/150?u=omar", IsActive: true},
	}

	inst1 := &WhatsAppInstance{
		ID:             "inst-1",
		OrganizationID: "org-1",
		Name:           "Sales WA",
		PhoneNumber:    "+20 100 200 3001",
		JID:            "201002003001@s.whatsapp.net",
		Status:         "connected",
		PairingState:   "paired",
		Settings: InstanceSettings{
			AutoSyncHistory:           true,
			AutoDownloadIncomingMedia: true,
			SourceTagLabel:            "Sales",
			SourceTagDisplayMode:      "label",
			SourceTagColor:            "emerald",
		},
		Health: InstanceHealth{
			Status:        "connected",
			UptimeLabel:   "18h 24m",
			QueueDepth:    2,
			SentToday:     41,
			ReceivedToday: 57,
			FailedToday:   1,
			ErrorRate:     "1.7%",
			ObservedAt:    now.Add(-5 * time.Minute),
		},
		CallPolicy: InstanceCallPolicy{
			Enabled:               true,
			RejectIndividualCalls: true,
			RejectGroupCalls:      true,
			ReplyMode:             "reject_with_message",
			ScheduleMode:          "always_on",
			EmergencyBypass:       []string{"+201234567890"},
			ReplyMessage:          "We reply faster in chat. Please leave a message.",
		},
		AutoCampaign: InstanceAutoCampaign{
			Enabled:            true,
			CampaignNamePrefix: "Winback",
			ScheduleEveryDays:  7,
			DelayFromMinutes:   1,
			DelayToMinutes:     3,
			CampaignStatus:     "active",
			MessageBody:        "We have a new offer ready for you.",
		},
		RatingSettings: InstanceRatingSettings{
			Enabled:               true,
			FollowUpWindowMinutes: 15,
			TemplateAR:            "كيف كانت تجربتك معنا؟",
			TemplateEN:            "How was your experience with us?",
		},
		AssignmentReset: InstanceAssignmentReset{
			Enabled:      true,
			ScheduleMode: "midnight",
			Timezone:     "Africa/Cairo",
		},
	}

	inst2 := &WhatsAppInstance{
		ID:             "inst-2",
		OrganizationID: "org-1",
		Name:           "Care WA",
		PhoneNumber:    "+20 100 200 3002",
		Status:         "disconnected",
		PairingState:   "needs_qr",
		QRCode:         "QR-CODE-DEMO-123456",
		Settings: InstanceSettings{
			AutoSyncHistory:           false,
			AutoDownloadIncomingMedia: false,
			SourceTagLabel:            "Care",
			SourceTagDisplayMode:      "instance_name",
			SourceTagColor:            "amber",
		},
		Health: InstanceHealth{
			Status:        "disconnected",
			UptimeLabel:   "0m",
			QueueDepth:    0,
			SentToday:     5,
			ReceivedToday: 11,
			FailedToday:   3,
			ErrorRate:     "21.4%",
			ObservedAt:    now.Add(-9 * time.Minute),
		},
		CallPolicy: InstanceCallPolicy{
			Enabled:               false,
			RejectIndividualCalls: true,
			RejectGroupCalls:      true,
			ReplyMode:             "reject_without_message",
			ScheduleMode:          "always_on",
		},
		AutoCampaign: InstanceAutoCampaign{
			Enabled:            false,
			CampaignNamePrefix: "Dormant",
			ScheduleEveryDays:  14,
			DelayFromMinutes:   3,
			DelayToMinutes:     8,
			CampaignStatus:     "draft",
		},
		RatingSettings: InstanceRatingSettings{
			Enabled:               false,
			FollowUpWindowMinutes: 15,
			TemplateAR:            "ما تقييمك للمحادثة؟",
			TemplateEN:            "How would you rate this conversation?",
		},
		AssignmentReset: InstanceAssignmentReset{
			Enabled:      false,
			ScheduleMode: "midnight",
			Timezone:     "Africa/Cairo",
		},
	}

	contact1 := &ConversationRecord{
		Contact: ChatContact{
			ID:                  "contact-1",
			OrganizationID:      "org-1",
			Name:                "Mina Salah",
			PhoneNumber:         "+201111111111",
			Avatar:              "https://i.pravatar.cc/150?u=mina",
			Status:              "assigned",
			AssignedUserID:      "2",
			AssignedUserName:    "Maha Support",
			InstanceID:          "inst-1",
			InstanceName:        "Sales WA",
			InstanceSourceLabel: "Sales",
			LastMessagePreview:  "Can you send the proposal PDF?",
			LastMessageAt:       now.Add(-12 * time.Minute),
			LastInboundAt:       now.Add(-12 * time.Minute),
			IsPublic:            false,
			IsRead:              false,
			UnreadCount:         2,
			Tags:                []string{"vip", "renewal"},
			Metadata:            map[string]string{"city": "Cairo"},
		},
		Messages: []ChatMessage{
			{ID: "msg-1", ContactID: "contact-1", Direction: "inbound", Type: "text", Body: "Hi, I need the updated proposal.", Status: "received", CreatedAt: now.Add(-38 * time.Minute), Reaction: "👍"},
			{ID: "msg-2", ContactID: "contact-1", Direction: "outbound", Type: "text", Body: "Absolutely. I will send it now.", Status: "sent", TypedForMS: 1180, CreatedAt: now.Add(-33 * time.Minute), CanRevoke: true},
			{ID: "msg-3", ContactID: "contact-1", Direction: "inbound", Type: "text", Body: "Can you send the proposal PDF?", Status: "received", CreatedAt: now.Add(-12 * time.Minute)},
		},
		Notes:         []ConversationNote{{ID: "note-1", UserID: "2", UserName: "Maha Support", Body: "Customer is asking specifically about the annual package.", CreatedAt: now.Add(-20 * time.Minute)}},
		Collaborators: []Collaborator{{ID: "col-1", UserID: "3", UserName: "Omar Care", Status: "accepted", InvitedAt: now.Add(-28 * time.Minute)}},
		Events: []TimelineEvent{
			{ID: "event-1", EventType: "assigned", ActorUserID: "1", ActorName: "Admin Encanto", Summary: "Assigned to Maha Support", OccurredAt: now.Add(-40 * time.Minute), Metadata: map[string]string{"to_user_id": "2"}},
			{ID: "event-2", EventType: "note_created", ActorUserID: "2", ActorName: "Maha Support", Summary: "Added an internal note", OccurredAt: now.Add(-20 * time.Minute), Metadata: map[string]string{}},
		},
		UserStates: map[string]*ContactUserState{
			"1": {IsPinned: true, LastReadMessageID: "msg-2", LastOpenedAt: now.Add(-15 * time.Minute), LastSeenAt: now.Add(-15 * time.Minute)},
		},
	}

	contact2 := &ConversationRecord{
		Contact: ChatContact{
			ID:                  "contact-2",
			OrganizationID:      "org-1",
			Name:                "Laila Hassan",
			PhoneNumber:         "+201222222222",
			Avatar:              "https://i.pravatar.cc/150?u=laila",
			Status:              "pending",
			InstanceID:          "inst-2",
			InstanceName:        "Care WA",
			InstanceSourceLabel: "Care",
			LastMessagePreview:  "The image failed to send earlier.",
			LastMessageAt:       now.Add(-55 * time.Minute),
			LastInboundAt:       now.Add(-64 * time.Minute),
			IsPublic:            true,
			IsRead:              false,
			UnreadCount:         1,
			Tags:                []string{"lead"},
			Metadata:            map[string]string{"city": "Alexandria"},
		},
		Messages: []ChatMessage{
			{ID: "msg-4", ContactID: "contact-2", Direction: "inbound", Type: "text", Body: "Can you send the product image again?", Status: "received", CreatedAt: now.Add(-64 * time.Minute)},
			{ID: "msg-5", ContactID: "contact-2", Direction: "outbound", Type: "media", Body: "Catalog preview", FileName: "catalog-preview.png", FileSizeLabel: "1.2 MB", Status: "failed", FailureReason: "Instance not connected", RetryCount: 1, CreatedAt: now.Add(-55 * time.Minute), CanRetry: true, CanRevoke: true, MediaURL: "https://picsum.photos/seed/catalog/640/420"},
		},
		Events: []TimelineEvent{
			{ID: "event-3", EventType: "public_changed", ActorUserID: "1", ActorName: "Admin Encanto", Summary: "Conversation made visible to the team", OccurredAt: now.Add(-56 * time.Minute), Metadata: map[string]string{"is_public": "true"}},
		},
		UserStates: map[string]*ContactUserState{
			"1": {LastOpenedAt: now.Add(-66 * time.Minute), LastSeenAt: now.Add(-66 * time.Minute)},
		},
	}

	contact3 := &ConversationRecord{
		Contact: ChatContact{
			ID:                  "contact-3",
			OrganizationID:      "org-1",
			Name:                "Omar Group",
			PhoneNumber:         "+201333333333",
			Avatar:              "https://i.pravatar.cc/150?u=omar-group",
			Status:              "closed",
			AssignedUserID:      "3",
			AssignedUserName:    "Omar Care",
			InstanceID:          "inst-1",
			InstanceName:        "Sales WA",
			InstanceSourceLabel: "Sales",
			LastMessagePreview:  "Thanks, we will confirm internally.",
			LastMessageAt:       now.Add(-4 * time.Hour),
			LastInboundAt:       now.Add(-4 * time.Hour),
			IsPublic:            false,
			IsRead:              true,
			Tags:                []string{"group"},
			Metadata:            map[string]string{"type": "group"},
		},
		Messages: []ChatMessage{{ID: "msg-6", ContactID: "contact-3", Direction: "inbound", Type: "text", Body: "Thanks, we will confirm internally.", Status: "received", CreatedAt: now.Add(-4 * time.Hour)}},
		Events:   []TimelineEvent{{ID: "event-4", EventType: "closed", ActorUserID: "3", ActorName: "Omar Care", Summary: "Closed the conversation after resolution", OccurredAt: now.Add(-4 * time.Hour), Metadata: map[string]string{}}},
		UserStates: map[string]*ContactUserState{
			"1": {LastReadMessageID: "msg-6", LastOpenedAt: now.Add(-4 * time.Hour), LastSeenAt: now.Add(-4 * time.Hour)},
		},
	}

	return &OrgData{
		General: GeneralSettings{
			OrganizationName:  "Global Corp",
			Slug:              "global-corp",
			Timezone:          "Africa/Cairo",
			DateFormat:        "DD MMM YYYY",
			Locale:            "en",
			MaskPhoneNumbers:  false,
			TenantStatus:      "active",
			ActiveMembers:     3,
			MaxMembers:        5,
			UsedInstances:     2,
			MaxInstances:      3,
			StorageUsedLabel:  "1.8 GiB",
			StorageLimitLabel: "5 GiB",
		},
		Appearance: AppearanceSettings{ColorMode: "light", ThemePreset: "ocean-breeze"},
		Chat: ChatSettings{
			MediaGroupingWindowMinutes: 5,
			SidebarContactView:         "comfortable",
			SidebarHoverExpand:         true,
			PinSidebar:                 true,
			ChatBackground:             "paper-grid",
			ShowPrintButtons:           true,
			ShowDownloadButtons:        true,
		},
		Notifications: NotificationSettings{
			EmailNotifications: true,
			NewMessageAlerts:   true,
			NotificationSound:  "soft-bell",
			CampaignUpdates:    true,
		},
		Cleanup: CleanupSettings{
			RetentionDays: 30,
			RunHour:       3,
			Timezone:      "Africa/Cairo",
			LastJobStatus: "never-run",
		},
		Contacts: map[string]*ConversationRecord{
			contact1.Contact.ID: contact1,
			contact2.Contact.ID: contact2,
			contact3.Contact.ID: contact3,
		},
		NotificationsFeed: []UserNotification{
			{ID: "notif-1", Title: "Conversation needs a retry", Body: "Laila Hassan has a failed outbound media message.", Severity: "warning", RelatedContactID: "contact-2", RelatedPath: "/chat/contact-2", IsRead: false, CreatedAt: now.Add(-50 * time.Minute)},
			{ID: "notif-2", Title: "Cleanup schedule active", Body: "Uploads cleanup runs daily at 03:00 Africa/Cairo.", Severity: "info", RelatedPath: "/settings", IsRead: true, CreatedAt: now.Add(-2 * time.Hour)},
		},
		Statuses: []StatusPost{
			{ID: "status-1", ContactID: "contact-1", ContactName: "Mina Salah", InstanceID: "inst-1", InstanceName: "Sales WA", Body: "Sharing the proposal with the client now.", Kind: "text", CreatedAt: now.Add(-16 * time.Minute)},
			{ID: "status-2", ContactID: "contact-2", ContactName: "Laila Hassan", InstanceID: "inst-2", InstanceName: "Care WA", Body: "Reconnect Care WA before retrying the media send.", Kind: "text", CreatedAt: now.Add(-10 * time.Minute)},
		},
		Instances: map[string]*WhatsAppInstance{
			inst1.ID: inst1,
			inst2.ID: inst2,
		},
		QuickReplies: []QuickReply{
			{ID: "qr-1", Shortcut: "/proposal", Title: "Proposal follow-up", Body: "Absolutely. I am sending the proposal right now."},
			{ID: "qr-2", Shortcut: "/thanks", Title: "Thank you", Body: "Thank you for the update. I am checking this now."},
			{ID: "qr-3", Shortcut: "/reconnect", Title: "Reconnect note", Body: "We are reconnecting the sending account and will retry shortly."},
		},
		Users:      users,
		License:    seedLicenseRecord(now, "org-1", "Global Corp", 12, 4, 3),
		Campaigns:  seedPrimaryCampaigns(now),
		Jobs:       seedPrimaryJobs(now),
		Webhooks:   seedDefaultWebhooks(),
		Deliveries: seedPrimaryDeliveries(now),
		Outbox:     seedPrimaryOutbox(now),
		Audit:      seedPrimaryAudit(now),
		Ratings:    seedPrimaryRatings(now),
	}
}

func seedSecondaryOrg(now time.Time) *OrgData {
	users := map[string]*WorkspaceUser{
		"1": {ID: "1", Name: "Admin Encanto", Email: mockEmail, Role: "agent", Status: "online", Avatar: "https://i.pravatar.cc/150?u=admin", IsActive: true},
	}

	inst := &WhatsAppInstance{
		ID:             "inst-3",
		OrganizationID: "org-2",
		Name:           "Local Store WA",
		PhoneNumber:    "+20 100 200 4001",
		JID:            "201002004001@s.whatsapp.net",
		Status:         "connected",
		PairingState:   "paired",
		Settings: InstanceSettings{
			AutoSyncHistory:           true,
			AutoDownloadIncomingMedia: true,
			SourceTagLabel:            "Store",
			SourceTagDisplayMode:      "label",
			SourceTagColor:            "sky",
		},
		Health: InstanceHealth{
			Status:        "connected",
			UptimeLabel:   "9h 12m",
			QueueDepth:    0,
			SentToday:     8,
			ReceivedToday: 14,
			FailedToday:   0,
			ErrorRate:     "0.0%",
			ObservedAt:    now.Add(-3 * time.Minute),
		},
	}

	record := &ConversationRecord{
		Contact: ChatContact{
			ID:                  "contact-4",
			OrganizationID:      "org-2",
			Name:                "Store Visitor",
			PhoneNumber:         "+201444444444",
			Avatar:              "https://i.pravatar.cc/150?u=visitor",
			Status:              "pending",
			InstanceID:          "inst-3",
			InstanceName:        "Local Store WA",
			InstanceSourceLabel: "Store",
			LastMessagePreview:  "Are you open today?",
			LastMessageAt:       now.Add(-22 * time.Minute),
			LastInboundAt:       now.Add(-22 * time.Minute),
			IsPublic:            true,
			IsRead:              false,
			UnreadCount:         1,
			Tags:                []string{"walk-in"},
			Metadata:            map[string]string{"city": "Giza"},
		},
		Messages:   []ChatMessage{{ID: "msg-7", ContactID: "contact-4", Direction: "inbound", Type: "text", Body: "Are you open today?", Status: "received", CreatedAt: now.Add(-22 * time.Minute)}},
		UserStates: map[string]*ContactUserState{"1": {}},
	}

	return &OrgData{
		General: GeneralSettings{
			OrganizationName:  "Local Store",
			Slug:              "local-store",
			Timezone:          "Africa/Cairo",
			DateFormat:        "DD/MM/YYYY",
			Locale:            "ar",
			MaskPhoneNumbers:  true,
			TenantStatus:      "active",
			ActiveMembers:     1,
			MaxMembers:        5,
			UsedInstances:     1,
			MaxInstances:      2,
			StorageUsedLabel:  "640 MB",
			StorageLimitLabel: "5 GiB",
		},
		Appearance: AppearanceSettings{ColorMode: "light", ThemePreset: "amber-minimal"},
		Chat: ChatSettings{
			MediaGroupingWindowMinutes: 3,
			SidebarContactView:         "compact",
			SidebarHoverExpand:         true,
			PinSidebar:                 false,
			ChatBackground:             "linen",
			ShowPrintButtons:           false,
			ShowDownloadButtons:        true,
		},
		Notifications: NotificationSettings{
			EmailNotifications: false,
			NewMessageAlerts:   true,
			NotificationSound:  "none",
			CampaignUpdates:    false,
		},
		Cleanup: CleanupSettings{
			RetentionDays: 14,
			RunHour:       1,
			Timezone:      "Africa/Cairo",
			LastJobStatus: "never-run",
		},
		Contacts:  map[string]*ConversationRecord{record.Contact.ID: record},
		Instances: map[string]*WhatsAppInstance{inst.ID: inst},
		QuickReplies: []QuickReply{
			{ID: "qr-4", Shortcut: "/hours", Title: "Working hours", Body: "We are open from 10 AM to 10 PM every day."},
		},
		Users:      users,
		License:    seedLicenseRecord(now, "org-2", "Local Store", 6, 2, 2),
		Campaigns:  seedSecondaryCampaigns(now),
		Jobs:       []BackgroundJob{},
		Webhooks:   seedDefaultWebhooks(),
		Deliveries: []WebhookDelivery{},
		Outbox:     []OutboxEvent{},
		Audit:      []AuditLogEntry{},
		Ratings:    seedSecondaryRatings(now),
	}
}

func (s *Store) next(prefix string) string {
	s.nextID++
	return fmt.Sprintf("%s-%d", prefix, s.nextID)
}

func (s *Store) getUserResponseUnlocked(activeOrgID string) UserResponse {
	orgs := getMockOrganizations()
	active := orgs[0]

	for _, org := range orgs {
		if org.ID == activeOrgID {
			active = org
			break
		}
	}

	activeData, ok := s.orgs[active.ID]
	if !ok {
		return UserResponse{
			ID:     "1",
			Email:  mockEmail,
			Name:   "Admin Encanto",
			Avatar: "https://i.pravatar.cc/150?u=admin",
			Status: "online",
			Role:   active.Role,
			Settings: UserSettings{
				Theme:         "light",
				Language:      "en",
				SidebarPinned: active.Role == "admin",
			},
			Organizations:       orgs,
			CurrentOrganization: active,
		}
	}

	user := activeData.Users["1"]
	if user == nil {
		user = &WorkspaceUser{
			ID:       "1",
			Email:    mockEmail,
			Name:     "Admin Encanto",
			Avatar:   "https://i.pravatar.cc/150?u=admin",
			Status:   "online",
			Role:     active.Role,
			IsActive: true,
		}
	}

	language := activeData.General.Locale
	if language == "" {
		language = "en"
	}

	return UserResponse{
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Avatar: user.Avatar,
		Status: user.Status,
		Role:   user.Role,
		Settings: UserSettings{
			Theme:         "light",
			Language:      language,
			SidebarPinned: user.Role == "admin",
		},
		Organizations:       orgs,
		CurrentOrganization: active,
	}
}

func (s *Store) GetUserResponse(activeOrgID string) UserResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getUserResponseUnlocked(activeOrgID)
}

func getMockOrganizations() []Organization {
	return []Organization{
		{ID: "org-1", Name: "Global Corp", Role: "admin"},
		{ID: "org-2", Name: "Local Store", Role: "agent"},
	}
}

func (s *Store) OrgAccessible(orgID string) bool {
	for _, org := range getMockOrganizations() {
		if org.ID == orgID {
			return true
		}
	}
	return false
}

func (s *Store) SettingsForOrg(orgID string) (SettingsSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return SettingsSummary{}, errors.New("organization not found")
	}

	return SettingsSummary{
		General:       org.General,
		Appearance:    org.Appearance,
		Chat:          org.Chat,
		Notifications: org.Notifications,
		Cleanup:       org.Cleanup,
	}, nil
}

func (s *Store) ProfileForOrg(orgID string) (ProfileView, error) {
	settings, err := s.SettingsForOrg(orgID)
	if err != nil {
		return ProfileView{}, err
	}

	return ProfileView{
		User:     s.GetUserResponse(orgID),
		Settings: settings,
	}, nil
}

func (s *Store) Workspace(orgID, userID, contactID, tab, search, instanceID, tag string) (WorkspaceSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WorkspaceSnapshot{}, errors.New("organization not found")
	}

	if tab == "" {
		tab = "assigned"
	}

	conversations := make([]ChatContact, 0, len(org.Contacts))
	counts := map[string]int{"assigned": 0, "pending": 0, "closed": 0}

	for _, record := range org.Contacts {
		state := s.ensureUserState(record, userID)
		contact := s.decoratedContact(org, record, userID)

		counts[contact.Status]++
		if state.IsHidden {
			continue
		}
		if tab != "" && tab != "all" && contact.Status != tab {
			continue
		}
		if search != "" {
			searchValue := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(contact.Name), searchValue) &&
				!strings.Contains(strings.ToLower(contact.PhoneNumber), searchValue) &&
				!strings.Contains(strings.ToLower(contact.LastMessagePreview), searchValue) {
				continue
			}
		}
		if instanceID != "" && contact.InstanceID != instanceID {
			continue
		}
		if tag != "" {
			found := false
			for _, contactTag := range contact.Tags {
				if strings.EqualFold(contactTag, tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		conversations = append(conversations, contact)
	}

	sort.Slice(conversations, func(i, j int) bool {
		if conversations[i].IsPinned != conversations[j].IsPinned {
			return conversations[i].IsPinned
		}
		return conversations[i].LastMessageAt.After(conversations[j].LastMessageAt)
	})

	var selected *ConversationDetail
	if contactID != "" {
		record, ok := org.Contacts[contactID]
		if !ok {
			return WorkspaceSnapshot{}, errors.New("conversation not found")
		}

		state := s.ensureUserState(record, userID)
		if len(record.Messages) > 0 {
			lastMessage := record.Messages[len(record.Messages)-1]
			state.LastReadMessageID = lastMessage.ID
			record.Contact.IsRead = true
			record.Contact.UnreadCount = 0
			state.LastOpenedAt = time.Now()
			state.LastSeenAt = time.Now()
		}

		selected = &ConversationDetail{
			Contact:       s.decoratedContact(org, record, userID),
			Messages:      append([]ChatMessage{}, record.Messages...),
			Notes:         append([]ConversationNote{}, record.Notes...),
			Collaborators: append([]Collaborator{}, record.Collaborators...),
			Events:        append([]TimelineEvent{}, record.Events...),
		}
	}

	notifications := slices.Clone(org.NotificationsFeed)
	sort.Slice(notifications, func(i, j int) bool { return notifications[i].CreatedAt.After(notifications[j].CreatedAt) })

	statuses := slices.Clone(org.Statuses)
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].CreatedAt.After(statuses[j].CreatedAt) })

	instances := make([]WhatsAppInstance, 0, len(org.Instances))
	for _, instance := range org.Instances {
		instances = append(instances, *instance)
	}
	sort.Slice(instances, func(i, j int) bool { return instances[i].Name < instances[j].Name })

	users := make([]WorkspaceUser, 0, len(org.Users))
	for _, user := range org.Users {
		users = append(users, *user)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].Name < users[j].Name })

	return WorkspaceSnapshot{
		CurrentTab:    tab,
		TabCounts:     counts,
		Filters:       map[string]string{"search": search, "instance_id": instanceID, "tag": tag},
		Conversations: conversations,
		Selected:      selected,
		Notifications: notifications,
		Statuses:      statuses,
		QuickReplies:  slices.Clone(org.QuickReplies),
		Instances:     instances,
		Users:         users,
		Settings: SettingsSummary{
			General:       org.General,
			Appearance:    org.Appearance,
			Chat:          org.Chat,
			Notifications: org.Notifications,
			Cleanup:       org.Cleanup,
		},
	}, nil
}

func (s *Store) decoratedContact(org *OrgData, record *ConversationRecord, userID string) ChatContact {
	state := s.ensureUserState(record, userID)
	contact := record.Contact
	contact.IsPinned = state.IsPinned
	contact.IsHidden = state.IsHidden
	if org.General.MaskPhoneNumbers {
		contact.PhoneDisplay = maskPhoneNumber(contact.PhoneNumber)
	} else {
		contact.PhoneDisplay = contact.PhoneNumber
	}
	return contact
}

func (s *Store) ensureUserState(record *ConversationRecord, userID string) *ContactUserState {
	state, ok := record.UserStates[userID]
	if ok {
		return state
	}
	state = &ContactUserState{}
	record.UserStates[userID] = state
	return state
}

func maskPhoneNumber(phone string) string {
	if len(phone) < 6 {
		return phone
	}
	return phone[:4] + strings.Repeat("*", max(0, len(phone)-6)) + phone[len(phone)-2:]
}

func (s *Store) addOrgNotification(org *OrgData, notification UserNotification) {
	org.NotificationsFeed = append([]UserNotification{notification}, org.NotificationsFeed...)
	if len(org.NotificationsFeed) > 12 {
		org.NotificationsFeed = org.NotificationsFeed[:12]
	}
}

func (s *Store) addConversationEvent(record *ConversationRecord, event TimelineEvent) {
	record.Events = append([]TimelineEvent{event}, record.Events...)
	if len(record.Events) > 30 {
		record.Events = record.Events[:30]
	}
}

func previewForMessage(message ChatMessage) string {
	if message.Type == "media" {
		if message.FileName != "" {
			return "Attachment: " + message.FileName
		}
		return "Attachment sent"
	}
	body := strings.TrimSpace(message.Body)
	if body == "" {
		return "Message sent"
	}
	return body
}

func normalizePhoneNumber(phone string) string {
	var builder strings.Builder
	for i, r := range strings.TrimSpace(phone) {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if r == '+' && i == 0 {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func (s *Store) actorName(org *OrgData, userID string) string {
	if user, ok := org.Users[userID]; ok {
		return user.Name
	}
	return "Unknown User"
}

func (s *Store) SendOutgoingMessage(orgID, userID, contactID string, req SendMessageRequest) (ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatMessage{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatMessage{}, errors.New("conversation not found")
	}
	instance, ok := org.Instances[record.Contact.InstanceID]
	if !ok {
		return ChatMessage{}, errors.New("instance not found")
	}

	message := ChatMessage{
		ID:            s.next("msg"),
		ContactID:     contactID,
		Direction:     "outbound",
		Type:          req.Type,
		Body:          strings.TrimSpace(req.Body),
		FileName:      strings.TrimSpace(req.FileName),
		FileSizeLabel: strings.TrimSpace(req.FileSizeLabel),
		MediaURL:      strings.TrimSpace(req.MediaURL),
		CreatedAt:     time.Now(),
		CanRetry:      true,
		CanRevoke:     true,
	}
	if message.Type == "" {
		message.Type = "text"
	}
	if message.Type == "text" {
		message.TypedForMS = min(4000, 420+len(message.Body)*38)
	}
	if message.Type == "media" && message.FileName == "" {
		message.FileName = "attachment.bin"
	}
	if message.Type == "media" && message.FileSizeLabel == "" {
		message.FileSizeLabel = "Preview"
	}
	if instance.Status != "connected" {
		message.Status = "failed"
		message.FailureReason = "Instance not connected"
		message.RetryCount = 1
	} else {
		message.Status = "sent"
	}

	record.Messages = append(record.Messages, message)
	record.Contact.LastMessagePreview = previewForMessage(message)
	record.Contact.LastMessageAt = message.CreatedAt
	record.Contact.IsRead = true

	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "message_sent",
		ActorUserID: userID,
		ActorName:   s.actorName(org, userID),
		Summary:     fmt.Sprintf("Sent a %s message", message.Type),
		OccurredAt:  message.CreatedAt,
		Metadata:    map[string]string{"message_id": message.ID, "status": message.Status},
	})
	s.recordAuditUnlocked(org, userID, s.actorName(org, userID), "messages.send", "message", message.ID, "Queued an outbound message.", map[string]string{
		"contact_id": contactID,
		"status":     message.Status,
		"type":       message.Type,
	})
	s.recordOutboxUnlocked(org, "messages.outbound", "message", message.ID, map[string]string{
		"contact_id": contactID,
		"status":     message.Status,
		"type":       message.Type,
	}, message.Status == "failed")

	return message, nil
}

func (s *Store) CreateDirectChat(orgID, userID string, req StartDirectChatRequest) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}

	phone := normalizePhoneNumber(req.PhoneNumber)
	if phone == "" {
		return ChatContact{}, errors.New("phone number is required")
	}

	instanceID := strings.TrimSpace(req.InstanceID)
	if instanceID == "" {
		for _, instance := range org.Instances {
			instanceID = instance.ID
			break
		}
	}

	instance, ok := org.Instances[instanceID]
	if !ok {
		return ChatContact{}, errors.New("instance not found")
	}

	for _, record := range org.Contacts {
		if normalizePhoneNumber(record.Contact.PhoneNumber) != phone || record.Contact.InstanceID != instanceID {
			continue
		}
		state := s.ensureUserState(record, userID)
		state.IsHidden = false
		record.Contact.Status = "assigned"
		record.Contact.AssignedUserID = userID
		record.Contact.AssignedUserName = s.actorName(org, userID)
		record.Contact.LastMessageAt = time.Now()
		s.recordAuditUnlocked(org, userID, s.actorName(org, userID), "chat.direct.resume", "contact", record.Contact.ID, "Re-used an existing direct chat.", map[string]string{
			"instance_id": instanceID,
		})
		s.recordOutboxUnlocked(org, "chat.direct.resumed", "contact", record.Contact.ID, map[string]string{
			"instance_id": instanceID,
		}, false)
		return s.decoratedContact(org, record, userID), nil
	}

	now := time.Now()
	name := strings.TrimSpace(req.ProfileName)
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
			Status:              "assigned",
			AssignedUserID:      userID,
			AssignedUserName:    s.actorName(org, userID),
			InstanceID:          instance.ID,
			InstanceName:        instance.Name,
			InstanceSourceLabel: instance.Settings.SourceTagLabel,
			LastMessagePreview:  "Direct chat ready for the first outbound message.",
			LastMessageAt:       now,
			IsPublic:            true,
			IsRead:              true,
			Tags:                []string{"direct"},
			Metadata:            map[string]string{"created_via": "direct_chat"},
		},
		Messages:      []ChatMessage{},
		Notes:         []ConversationNote{},
		Collaborators: []Collaborator{},
		Events: []TimelineEvent{
			{
				ID:          s.next("event"),
				EventType:   "direct_chat_created",
				ActorUserID: userID,
				ActorName:   s.actorName(org, userID),
				Summary:     "Started a direct chat",
				OccurredAt:  now,
				Metadata: map[string]string{
					"instance_id": instance.ID,
				},
			},
		},
		UserStates: map[string]*ContactUserState{
			userID: {
				LastOpenedAt: now,
				LastSeenAt:   now,
			},
		},
	}

	org.Contacts[contactID] = record
	s.addOrgNotification(org, UserNotification{
		ID:               s.next("notif"),
		Title:            "Direct chat started",
		Body:             fmt.Sprintf("A new direct chat with %s was created on %s.", name, instance.Name),
		Severity:         "info",
		RelatedContactID: contactID,
		RelatedPath:      "/chat/" + contactID,
		CreatedAt:        now,
	})
	s.recordAuditUnlocked(org, userID, s.actorName(org, userID), "chat.direct.create", "contact", contactID, "Started a direct chat.", map[string]string{
		"instance_id": instance.ID,
	})
	s.recordOutboxUnlocked(org, "chat.direct.created", "contact", contactID, map[string]string{
		"instance_id": instance.ID,
	}, false)

	return s.decoratedContact(org, record, userID), nil
}

func (s *Store) RetryMessage(orgID, userID, contactID, messageID string) (ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatMessage{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatMessage{}, errors.New("conversation not found")
	}
	instance, ok := org.Instances[record.Contact.InstanceID]
	if !ok {
		return ChatMessage{}, errors.New("instance not found")
	}

	for i := range record.Messages {
		message := &record.Messages[i]
		if message.ID != messageID {
			continue
		}

		message.RetryCount++
		if instance.Status == "connected" {
			message.Status = "sent"
			message.FailureReason = ""
			message.CanRetry = false
			message.TypedForMS = min(4000, 420+len(message.Body)*38)
		} else {
			message.Status = "failed"
			message.FailureReason = "Instance not connected"
			message.CanRetry = true
		}
		record.Contact.LastMessagePreview = previewForMessage(*message)
		record.Contact.LastMessageAt = time.Now()

		s.addConversationEvent(record, TimelineEvent{
			ID:          s.next("event"),
			EventType:   "message_retried",
			ActorUserID: userID,
			ActorName:   s.actorName(org, userID),
			Summary:     "Retried a failed message",
			OccurredAt:  time.Now(),
			Metadata:    map[string]string{"message_id": message.ID, "status": message.Status},
		})

		return *message, nil
	}

	return ChatMessage{}, errors.New("message not found")
}

func (s *Store) RevokeMessage(orgID, userID, contactID, messageID string) (ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatMessage{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatMessage{}, errors.New("conversation not found")
	}

	for i := range record.Messages {
		message := &record.Messages[i]
		if message.ID != messageID {
			continue
		}
		now := time.Now()
		message.Status = "revoked"
		message.RevokedAt = &now
		message.CanRevoke = false

		s.addConversationEvent(record, TimelineEvent{
			ID:          s.next("event"),
			EventType:   "message_revoked",
			ActorUserID: userID,
			ActorName:   s.actorName(org, userID),
			Summary:     "Revoked an outbound message",
			OccurredAt:  now,
			Metadata:    map[string]string{"message_id": message.ID},
		})

		return *message, nil
	}

	return ChatMessage{}, errors.New("message not found")
}

func (s *Store) AddNote(orgID, userID, contactID, body string) (ConversationNote, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ConversationNote{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ConversationNote{}, errors.New("conversation not found")
	}

	note := ConversationNote{ID: s.next("note"), UserID: userID, UserName: s.actorName(org, userID), Body: strings.TrimSpace(body), CreatedAt: time.Now()}
	record.Notes = append([]ConversationNote{note}, record.Notes...)
	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "note_created",
		ActorUserID: userID,
		ActorName:   note.UserName,
		Summary:     "Added an internal note",
		OccurredAt:  note.CreatedAt,
		Metadata:    map[string]string{"note_id": note.ID},
	})

	return note, nil
}

func (s *Store) AddCollaborator(orgID, userID, contactID, collaboratorID string) (Collaborator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return Collaborator{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return Collaborator{}, errors.New("conversation not found")
	}
	user, ok := org.Users[collaboratorID]
	if !ok {
		return Collaborator{}, errors.New("user not found")
	}
	for _, existing := range record.Collaborators {
		if existing.UserID == collaboratorID {
			return existing, nil
		}
	}

	collaborator := Collaborator{ID: s.next("collab"), UserID: collaboratorID, UserName: user.Name, Status: "invited", InvitedAt: time.Now()}
	record.Collaborators = append([]Collaborator{collaborator}, record.Collaborators...)
	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "collaborator_invited",
		ActorUserID: userID,
		ActorName:   s.actorName(org, userID),
		Summary:     "Invited a collaborator to the conversation",
		OccurredAt:  collaborator.InvitedAt,
		Metadata:    map[string]string{"collaborator_user_id": collaboratorID},
	})

	return collaborator, nil
}

func (s *Store) Assign(orgID, actorID, contactID, assigneeID string) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("conversation not found")
	}
	user, ok := org.Users[assigneeID]
	if !ok {
		return ChatContact{}, errors.New("assignee not found")
	}

	record.Contact.AssignedUserID = assigneeID
	record.Contact.AssignedUserName = user.Name
	record.Contact.Status = "assigned"

	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "assigned",
		ActorUserID: actorID,
		ActorName:   s.actorName(org, actorID),
		Summary:     "Assigned the conversation",
		OccurredAt:  time.Now(),
		Metadata:    map[string]string{"to_user_id": assigneeID, "to_user_name": user.Name},
	})

	s.addOrgNotification(org, UserNotification{
		ID:               s.next("notif"),
		Title:            "Conversation assigned",
		Body:             fmt.Sprintf("%s is now assigned to %s.", record.Contact.Name, user.Name),
		Severity:         "info",
		RelatedContactID: contactID,
		RelatedPath:      "/chat/" + contactID,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})

	return s.decoratedContact(org, record, actorID), nil
}

func (s *Store) Unassign(orgID, actorID, contactID string) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("conversation not found")
	}

	record.Contact.AssignedUserID = ""
	record.Contact.AssignedUserName = ""
	record.Contact.Status = "pending"

	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "unassigned",
		ActorUserID: actorID,
		ActorName:   s.actorName(org, actorID),
		Summary:     "Returned the conversation to pending",
		OccurredAt:  time.Now(),
		Metadata:    map[string]string{},
	})

	return s.decoratedContact(org, record, actorID), nil
}

func (s *Store) TogglePin(orgID, userID, contactID string) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("conversation not found")
	}

	state := s.ensureUserState(record, userID)
	state.IsPinned = !state.IsPinned
	return s.decoratedContact(org, record, userID), nil
}

func (s *Store) ToggleHide(orgID, userID, contactID string) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("conversation not found")
	}

	state := s.ensureUserState(record, userID)
	state.IsHidden = !state.IsHidden
	return s.decoratedContact(org, record, userID), nil
}

func (s *Store) Close(orgID, actorID, contactID string) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("conversation not found")
	}

	now := time.Now()
	record.Contact.Status = "closed"
	record.Contact.ClosedAt = &now
	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "closed",
		ActorUserID: actorID,
		ActorName:   s.actorName(org, actorID),
		Summary:     "Closed the conversation",
		OccurredAt:  now,
		Metadata:    map[string]string{},
	})
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "chat.close", "contact", contactID, "Closed a conversation.", nil)
	s.recordOutboxUnlocked(org, "chat.closed", "contact", contactID, map[string]string{
		"closed_at": now.Format(time.RFC3339),
	}, false)

	return s.decoratedContact(org, record, actorID), nil
}

func (s *Store) Reopen(orgID, actorID, contactID string) (ChatContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatContact{}, errors.New("organization not found")
	}
	record, ok := org.Contacts[contactID]
	if !ok {
		return ChatContact{}, errors.New("conversation not found")
	}

	now := time.Now()
	record.Contact.Status = "pending"
	record.Contact.ClosedAt = nil
	s.addConversationEvent(record, TimelineEvent{
		ID:          s.next("event"),
		EventType:   "reopened",
		ActorUserID: actorID,
		ActorName:   s.actorName(org, actorID),
		Summary:     "Reopened the conversation",
		OccurredAt:  now,
		Metadata:    map[string]string{},
	})
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "chat.reopen", "contact", contactID, "Reopened a conversation.", nil)
	s.recordOutboxUnlocked(org, "chat.reopened", "contact", contactID, map[string]string{
		"reopened_at": now.Format(time.RFC3339),
	}, false)

	return s.decoratedContact(org, record, actorID), nil
}

func (s *Store) ListNotifications(orgID string) ([]UserNotification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return slices.Clone(org.NotificationsFeed), nil
}

func (s *Store) MarkAllNotificationsRead(orgID string) ([]UserNotification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	for i := range org.NotificationsFeed {
		org.NotificationsFeed[i].IsRead = true
	}
	return slices.Clone(org.NotificationsFeed), nil
}

func (s *Store) ListStatuses(orgID string) ([]StatusPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return slices.Clone(org.Statuses), nil
}

func (s *Store) AddStatus(orgID, userID string, req CreateStatusRequest) (StatusPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return StatusPost{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[req.InstanceID]
	if !ok {
		return StatusPost{}, errors.New("instance not found")
	}
	record, ok := org.Contacts[req.ContactID]
	if !ok {
		return StatusPost{}, errors.New("conversation not found")
	}

	status := StatusPost{
		ID:           s.next("status"),
		ContactID:    req.ContactID,
		ContactName:  record.Contact.Name,
		InstanceID:   req.InstanceID,
		InstanceName: instance.Name,
		Body:         strings.TrimSpace(req.Body),
		Kind:         "text",
		CreatedAt:    time.Now(),
	}
	org.Statuses = append([]StatusPost{status}, org.Statuses...)
	return status, nil
}

func (s *Store) UpdateProfile(orgID string, req UpdateProfileRequest) (ProfileView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ProfileView{}, errors.New("organization not found")
	}

	user := org.Users["1"]
	user.Name = strings.TrimSpace(req.Name)
	user.Status = req.Status
	if req.Language != "" {
		org.General.Locale = req.Language
	}
	if req.ThemePreset != "" {
		org.Appearance.ThemePreset = req.ThemePreset
	}

	return ProfileView{
		User:     s.getUserResponseUnlocked(orgID),
		Settings: SettingsSummary{General: org.General, Appearance: org.Appearance, Chat: org.Chat, Notifications: org.Notifications, Cleanup: org.Cleanup},
	}, nil
}

func (s *Store) UpdateGeneral(orgID string, req GeneralSettings) (GeneralSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return GeneralSettings{}, errors.New("organization not found")
	}
	previousTimezone := org.General.Timezone
	org.General.OrganizationName = strings.TrimSpace(req.OrganizationName)
	org.General.Slug = strings.TrimSpace(req.Slug)
	org.General.Timezone = req.Timezone
	org.General.DateFormat = req.DateFormat
	org.General.Locale = req.Locale
	org.General.MaskPhoneNumbers = req.MaskPhoneNumbers
	if org.Cleanup.Timezone == "" || org.Cleanup.Timezone == previousTimezone {
		org.Cleanup.Timezone = req.Timezone
	}
	return org.General, nil
}

func (s *Store) UpdateAppearance(orgID string, req AppearanceSettings) (AppearanceSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return AppearanceSettings{}, errors.New("organization not found")
	}
	org.Appearance = req
	return org.Appearance, nil
}

func (s *Store) UpdateChatSettings(orgID string, req ChatSettings) (ChatSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return ChatSettings{}, errors.New("organization not found")
	}
	org.Chat = req
	return org.Chat, nil
}

func (s *Store) UpdateNotificationsSettings(orgID string, req NotificationSettings) (NotificationSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return NotificationSettings{}, errors.New("organization not found")
	}
	org.Notifications = req
	return org.Notifications, nil
}

func (s *Store) UpdateCleanupSettings(orgID, actorID string, req CleanupSettings) (CleanupSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return CleanupSettings{}, errors.New("organization not found")
	}

	if req.RetentionDays < 0 {
		return CleanupSettings{}, errors.New("retention days must be 0 or greater")
	}
	if req.RunHour < 0 || req.RunHour > 23 {
		return CleanupSettings{}, errors.New("cleanup hour must be between 0 and 23")
	}

	org.Cleanup.RetentionDays = req.RetentionDays
	org.Cleanup.RunHour = req.RunHour
	if strings.TrimSpace(req.Timezone) != "" {
		org.Cleanup.Timezone = strings.TrimSpace(req.Timezone)
	} else if strings.TrimSpace(org.Cleanup.Timezone) == "" {
		org.Cleanup.Timezone = org.General.Timezone
	}

	s.addOrgNotification(org, UserNotification{
		ID:          s.next("notif"),
		Title:       "Cleanup schedule updated",
		Body:        fmt.Sprintf("Uploads cleanup now runs daily at %02d:00 with %d-day retention.", org.Cleanup.RunHour, org.Cleanup.RetentionDays),
		Severity:    "info",
		RelatedPath: "/settings",
		CreatedAt:   time.Now(),
	})
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "settings.cleanup.update", "cleanup", orgID, "Updated the uploads cleanup schedule.", map[string]string{
		"retention_days": strconv.Itoa(org.Cleanup.RetentionDays),
		"run_hour":       strconv.Itoa(org.Cleanup.RunHour),
		"timezone":       org.Cleanup.Timezone,
	})
	s.recordOutboxUnlocked(org, "settings.cleanup.updated", "cleanup", orgID, map[string]string{
		"retention_days": strconv.Itoa(org.Cleanup.RetentionDays),
		"run_hour":       strconv.Itoa(org.Cleanup.RunHour),
	}, false)

	return org.Cleanup, nil
}

func (s *Store) RunCleanup(orgID, actorID string) (CleanupJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return CleanupJob{}, errors.New("organization not found")
	}

	start := time.Now()
	finish := start.Add(2 * time.Second)
	job := s.recordJobUnlocked(org, "uploads_cleanup", "cleanup", orgID, "Running uploads cleanup.")
	org.Cleanup.LastRunAt = &finish
	org.Cleanup.LastJobStatus = "completed"
	s.addOrgNotification(org, UserNotification{
		ID:          s.next("notif"),
		Title:       "Uploads cleanup finished",
		Body:        "Cleanup completed successfully and archived expired uploads.",
		Severity:    "success",
		RelatedPath: "/settings",
		CreatedAt:   finish,
	})
	s.finishJobUnlocked(org, job.ID, "completed", "")
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "settings.cleanup.run", "cleanup", orgID, "Ran uploads cleanup manually.", map[string]string{
		"job_id": job.ID,
	})
	s.recordOutboxUnlocked(org, "settings.cleanup.completed", "cleanup", orgID, map[string]string{
		"job_id": job.ID,
		"status": "completed",
	}, false)

	return CleanupJob{ID: job.ID, Status: "completed", StartedAt: start, FinishedAt: finish}, nil
}

func (s *Store) ListInstances(orgID string) ([]WhatsAppInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	instances := make([]WhatsAppInstance, 0, len(org.Instances))
	for _, instance := range org.Instances {
		instances = append(instances, *instance)
	}
	sort.Slice(instances, func(i, j int) bool { return instances[i].Name < instances[j].Name })
	return instances, nil
}

func (s *Store) ListInstanceHealth(orgID string) ([]InstanceHealthSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, errors.New("organization not found")
	}

	summaries := make([]InstanceHealthSummary, 0, len(org.Instances))
	for _, instance := range org.Instances {
		summaries = append(summaries, InstanceHealthSummary{
			ID:            instance.ID,
			Name:          instance.Name,
			Status:        instance.Health.Status,
			UptimeLabel:   instance.Health.UptimeLabel,
			QueueDepth:    instance.Health.QueueDepth,
			SentToday:     instance.Health.SentToday,
			ReceivedToday: instance.Health.ReceivedToday,
			FailedToday:   instance.Health.FailedToday,
			ErrorRate:     instance.Health.ErrorRate,
			ObservedAt:    instance.Health.ObservedAt,
		})
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Name < summaries[j].Name })
	return summaries, nil
}

func (s *Store) CreateInstance(orgID, actorID string, req CreateInstanceRequest) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	if len(org.Instances) >= org.General.MaxInstances {
		return WhatsAppInstance{}, errors.New("no remaining instance slots")
	}

	instance := WhatsAppInstance{
		ID:             s.next("inst"),
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		PhoneNumber:    req.PhoneNumber,
		Status:         "disconnected",
		PairingState:   "needs_qr",
		QRCode:         "QR-CODE-DEMO-NEW",
		Settings: InstanceSettings{
			AutoSyncHistory:           true,
			AutoDownloadIncomingMedia: true,
			SourceTagLabel:            strings.TrimSpace(req.Name),
			SourceTagDisplayMode:      "label",
			SourceTagColor:            "violet",
		},
		Health: InstanceHealth{
			Status:      "disconnected",
			UptimeLabel: "0m",
			ErrorRate:   "0.0%",
			ObservedAt:  time.Now(),
		},
		CallPolicy: InstanceCallPolicy{
			ReplyMode:    "reject_without_message",
			ScheduleMode: "always_on",
		},
		AutoCampaign:    InstanceAutoCampaign{CampaignStatus: "draft"},
		RatingSettings:  InstanceRatingSettings{FollowUpWindowMinutes: 15},
		AssignmentReset: InstanceAssignmentReset{ScheduleMode: "midnight", Timezone: org.General.Timezone},
	}

	org.Instances[instance.ID] = &instance
	org.General.UsedInstances = len(org.Instances)
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.create", "instance", instance.ID, "Created a WhatsApp account.", map[string]string{
		"name": instance.Name,
	})
	s.recordOutboxUnlocked(org, "instances.created", "instance", instance.ID, map[string]string{
		"name": instance.Name,
	}, false)
	return instance, nil
}

func (s *Store) UpdateInstanceName(orgID, actorID, instanceID, name string) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, errors.New("instance not found")
	}
	instance.Name = strings.TrimSpace(name)
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.rename", "instance", instanceID, "Updated the account name.", map[string]string{
		"name": instance.Name,
	})
	s.recordOutboxUnlocked(org, "instances.renamed", "instance", instanceID, map[string]string{
		"name": instance.Name,
	}, false)
	return *instance, nil
}

func (s *Store) ConnectInstance(orgID, actorID, instanceID string) (WhatsAppInstance, *ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, nil, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, nil, errors.New("instance not found")
	}

	instance.Status = "connected"
	instance.PairingState = "paired"
	instance.QRCode = ""
	instance.JID = strings.ReplaceAll(instance.PhoneNumber, " ", "") + "@s.whatsapp.net"
	instance.Health.Status = "connected"
	instance.Health.ObservedAt = time.Now()
	instance.Health.UptimeLabel = "1m"

	s.addOrgNotification(org, UserNotification{
		ID:          s.next("notif"),
		Title:       "Instance connected",
		Body:        fmt.Sprintf("%s is connected and ready to send.", instance.Name),
		Severity:    "success",
		RelatedPath: "/settings/instances",
		CreatedAt:   time.Now(),
	})
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.connect", "instance", instanceID, "Connected a WhatsApp account.", map[string]string{
		"name": instance.Name,
	})
	s.recordOutboxUnlocked(org, "instances.connected", "instance", instanceID, map[string]string{
		"name": instance.Name,
	}, false)

	var inbound *ChatMessage
	for _, record := range org.Contacts {
		if record.Contact.InstanceID == instanceID && record.Contact.Status != "closed" {
			msg := ChatMessage{
				ID:        s.next("msg"),
				ContactID: record.Contact.ID,
				Direction: "inbound",
				Type:      "text",
				Body:      "We are back online. Can you retry the media now?",
				Status:    "received",
				CreatedAt: time.Now(),
			}
			record.Messages = append(record.Messages, msg)
			record.Contact.LastMessagePreview = msg.Body
			record.Contact.LastMessageAt = msg.CreatedAt
			record.Contact.LastInboundAt = msg.CreatedAt
			record.Contact.UnreadCount++
			record.Contact.IsRead = false
			inbound = &msg
			break
		}
	}

	return *instance, inbound, nil
}

func (s *Store) DisconnectInstance(orgID, actorID, instanceID string) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, errors.New("instance not found")
	}

	instance.Status = "disconnected"
	instance.PairingState = "needs_qr"
	instance.QRCode = "QR-CODE-DEMO-RECONNECT"
	instance.Health.Status = "disconnected"
	instance.Health.ObservedAt = time.Now()
	instance.Health.UptimeLabel = "0m"
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.disconnect", "instance", instanceID, "Disconnected a WhatsApp account.", map[string]string{
		"name": instance.Name,
	})
	s.recordOutboxUnlocked(org, "instances.disconnected", "instance", instanceID, map[string]string{
		"name": instance.Name,
	}, false)
	return *instance, nil
}

func (s *Store) RecoverInstance(orgID, actorID, instanceID string) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, errors.New("instance not found")
	}

	instance.Status = "recovering"
	instance.PairingState = "reconnecting"
	instance.Health.Status = "recovering"
	instance.Health.ObservedAt = time.Now()
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.recover", "instance", instanceID, "Started account recovery.", map[string]string{
		"name": instance.Name,
	})
	s.recordOutboxUnlocked(org, "instances.recovery_started", "instance", instanceID, map[string]string{
		"name": instance.Name,
	}, false)
	return *instance, nil
}

func (s *Store) UpdateInstanceSettings(orgID, actorID, instanceID string, req InstanceSettings) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, errors.New("instance not found")
	}
	instance.Settings = req
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.settings.update", "instance", instanceID, "Updated instance settings.", map[string]string{
		"source_tag_label": instance.Settings.SourceTagLabel,
	})
	s.recordOutboxUnlocked(org, "instances.settings.updated", "instance", instanceID, map[string]string{
		"source_tag_label": instance.Settings.SourceTagLabel,
	}, false)
	return *instance, nil
}

func (s *Store) UpdateInstanceCallPolicy(orgID, actorID, instanceID string, req InstanceCallPolicy) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, errors.New("instance not found")
	}
	instance.CallPolicy = req
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.call_policy.update", "instance", instanceID, "Updated the call auto-reject policy.", map[string]string{
		"reply_mode": instance.CallPolicy.ReplyMode,
	})
	s.recordOutboxUnlocked(org, "instances.call_policy.updated", "instance", instanceID, map[string]string{
		"reply_mode": instance.CallPolicy.ReplyMode,
	}, false)
	return *instance, nil
}

func (s *Store) UpdateInstanceAutoCampaign(orgID, actorID, instanceID string, req InstanceAutoCampaign) (WhatsAppInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return WhatsAppInstance{}, errors.New("organization not found")
	}
	instance, ok := org.Instances[instanceID]
	if !ok {
		return WhatsAppInstance{}, errors.New("instance not found")
	}
	instance.AutoCampaign = req
	s.syncAutoCampaignUnlocked(org, instanceID)
	s.recordAuditUnlocked(org, actorID, s.actorName(org, actorID), "instances.auto_campaign.update", "instance", instanceID, "Updated linked account automation.", map[string]string{
		"enabled": strconv.FormatBool(instance.AutoCampaign.Enabled),
	})
	s.recordOutboxUnlocked(org, "instances.auto_campaign.updated", "instance", instanceID, map[string]string{
		"enabled": strconv.FormatBool(instance.AutoCampaign.Enabled),
	}, false)
	return *instance, nil
}
