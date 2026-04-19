package core

import (
	"time"

	"github.com/google/uuid"
)

type PermissionKey string

type VisibilityScopeMode string

const (
	ScopeAllContacts                 VisibilityScopeMode = "all_contacts"
	ScopeInstancesOnly               VisibilityScopeMode = "instances_only"
	ScopeAllowedNumbersOnly          VisibilityScopeMode = "allowed_numbers_only"
	ScopeInstancesPlusAllowedNumbers VisibilityScopeMode = "instances_plus_allowed_numbers"
)

type UserSettings struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	SidebarPinned bool   `json:"sidebarPinned"`
}

type UserOrganization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	RoleID    uuid.UUID `json:"roleId"`
	RoleName  string    `json:"roleName"`
	IsDefault bool      `json:"isDefault"`
}

type VisibilityScope struct {
	Mode                 VisibilityScopeMode `json:"mode"`
	AllowedInstanceIDs   []uuid.UUID         `json:"allowedInstanceIds"`
	AllowedPhoneNumbers  []string            `json:"allowedPhoneNumbers"`
	CanViewUnmaskedPhone bool                `json:"canViewUnmaskedPhone"`
}

type EffectiveAccess struct {
	PermissionKeys []PermissionKey `json:"permissionKeys"`
	Visibility     VisibilityScope `json:"visibility"`
}

type CurrentUserContext struct {
	ID                  uuid.UUID          `json:"id"`
	Email               string             `json:"email"`
	FullName            string             `json:"fullName"`
	AvatarURL           string             `json:"avatarUrl"`
	AvailabilityStatus  string             `json:"availabilityStatus"`
	Settings            UserSettings       `json:"settings"`
	Organizations       []UserOrganization `json:"organizations"`
	CurrentOrganization UserOrganization   `json:"currentOrganization"`
	EffectiveAccess     EffectiveAccess    `json:"effectiveAccess"`
}

type CurrentSession struct {
	SessionID      string
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	ExpiresAt      time.Time
}

type RoleDefinition struct {
	ID                uuid.UUID       `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	IsSystem          bool            `json:"isSystem"`
	IsDefault         bool            `json:"isDefault"`
	PermissionKeys    []PermissionKey `json:"permissionKeys"`
	DefaultVisibility VisibilityScope `json:"defaultVisibility"`
}

type UserPermissionOverride struct {
	PermissionKey PermissionKey `json:"permissionKey"`
	Mode          string        `json:"mode"`
}

type UserContactVisibilityRule struct {
	UserID               uuid.UUID           `json:"userId"`
	ScopeMode            VisibilityScopeMode `json:"scopeMode"`
	AllowedInstanceIDs   []uuid.UUID         `json:"allowedInstanceIds"`
	AllowedPhoneNumbers  []string            `json:"allowedPhoneNumbers"`
	InheritRoleScope     bool                `json:"inheritRoleScope"`
	CanViewUnmaskedPhone bool                `json:"canViewUnmaskedPhone"`
}

type ChatListItem struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	PhoneNumber        string     `json:"phoneNumber"`
	VisiblePhone       string     `json:"visiblePhone"`
	Status             string     `json:"status"`
	LastMessagePreview string     `json:"lastMessagePreview"`
	LastMessageAt      *time.Time `json:"lastMessageAt"`
	InstanceName       string     `json:"instanceName"`
	IsHidden           bool       `json:"isHidden"`
	IsPinned           bool       `json:"isPinned"`
}

type ConversationMessage struct {
	ID           uuid.UUID  `json:"id"`
	ContactID    uuid.UUID  `json:"contactId"`
	Direction    string     `json:"direction"`
	Type         string     `json:"type"`
	Body         string     `json:"body"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	SentByUserID *uuid.UUID `json:"sentByUserId,omitempty"`
}

type ConversationNote struct {
	ID           uuid.UUID `json:"id"`
	ContactID    uuid.UUID `json:"contactId"`
	AuthorUserID uuid.UUID `json:"authorUserId"`
	AuthorName   string    `json:"authorName"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ComposerState struct {
	Allowed      bool   `json:"allowed"`
	Disabled     bool   `json:"disabled"`
	DenialReason string `json:"denialReason"`
}

type ChatDetail struct {
	Chat     ChatListItem          `json:"chat"`
	Messages []ConversationMessage `json:"messages"`
	Notes    []ConversationNote    `json:"notes"`
	Composer ComposerState         `json:"composer"`
}
