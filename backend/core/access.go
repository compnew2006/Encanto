package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"encanto/data"
	"encanto/data/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AccessService struct {
	store   *data.Store
	catalog PermissionCatalog
}

func NewAccessService(store *data.Store, catalog PermissionCatalog) *AccessService {
	return &AccessService{store: store, catalog: catalog}
}

func (s *AccessService) Catalog() PermissionCatalog {
	return s.catalog
}

func (s *AccessService) ResolveCurrentUser(ctx context.Context, session CurrentSession) (CurrentUserContext, error) {
	user, err := s.store.Queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return CurrentUserContext{}, fmt.Errorf("load user: %w", err)
	}

	memberships, err := s.store.Queries.ListUserMemberships(ctx, session.UserID)
	if err != nil {
		return CurrentUserContext{}, fmt.Errorf("list memberships: %w", err)
	}
	if len(memberships) == 0 {
		return CurrentUserContext{}, fmt.Errorf("user has no active memberships")
	}

	var currentMembership sqlc.ListUserMembershipsRow
	found := false
	organizations := make([]UserOrganization, 0, len(memberships))
	for _, membership := range memberships {
		organizations = append(organizations, UserOrganization{
			ID:        membership.OrganizationID,
			Name:      membership.OrganizationName,
			RoleID:    membership.RoleID,
			RoleName:  membership.RoleName,
			IsDefault: membership.IsDefault,
		})
		if membership.OrganizationID == session.OrganizationID {
			currentMembership = membership
			found = true
		}
	}
	if !found {
		currentMembership = memberships[0]
	}

	role, err := s.store.Queries.GetRoleByID(ctx, sqlc.GetRoleByIDParams{
		ID:             currentMembership.RoleID,
		OrganizationID: currentMembership.OrganizationID,
	})
	if err != nil {
		return CurrentUserContext{}, fmt.Errorf("load role: %w", err)
	}

	rolePermissionKeys, err := s.store.Queries.ListRolePermissionKeys(ctx, role.ID)
	if err != nil {
		return CurrentUserContext{}, fmt.Errorf("load role permissions: %w", err)
	}

	overrideRows, err := s.store.Queries.ListUserPermissionOverrides(ctx, sqlc.ListUserPermissionOverridesParams{
		OrganizationID: currentMembership.OrganizationID,
		UserID:         session.UserID,
	})
	if err != nil {
		return CurrentUserContext{}, fmt.Errorf("load permission overrides: %w", err)
	}

	visibilityRule, err := s.store.Queries.GetUserVisibilityRule(ctx, sqlc.GetUserVisibilityRuleParams{
		OrganizationID: currentMembership.OrganizationID,
		UserID:         session.UserID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return CurrentUserContext{}, fmt.Errorf("load visibility rule: %w", err)
	}

	permissionKeys := applyPermissionOverrides(rolePermissionKeys, overrideRows)
	visibility := resolveVisibility(role, visibilityRule, err == nil)
	settings := decodeUserSettings(user.Settings)

	return CurrentUserContext{
		ID:                 user.ID,
		Email:              user.Email,
		FullName:           user.FullName,
		AvatarURL:          safeText(user.AvatarUrl),
		AvailabilityStatus: user.AvailabilityStatus,
		Settings:           settings,
		Organizations:      organizations,
		CurrentOrganization: UserOrganization{
			ID:        currentMembership.OrganizationID,
			Name:      currentMembership.OrganizationName,
			RoleID:    currentMembership.RoleID,
			RoleName:  currentMembership.RoleName,
			IsDefault: currentMembership.IsDefault,
		},
		EffectiveAccess: EffectiveAccess{
			PermissionKeys: permissionKeys,
			Visibility:     visibility,
		},
	}, nil
}

func (s *AccessService) HasPermission(user CurrentUserContext, key PermissionKey) bool {
	return slices.Contains(user.EffectiveAccess.PermissionKeys, key)
}

func (s *AccessService) PermissionDeniedReason(key PermissionKey) string {
	switch key {
	case "messages.send":
		return "You do not have permission to send messages in this workspace."
	case "chats.unclaimed.send":
		return "You cannot reply to pending chats until they are claimed."
	case "roles.manage":
		return "Only users with role management access can change role definitions."
	case "users.update":
		return "Only users with user-management access can change user overrides."
	default:
		return "Your current permissions do not allow this action."
	}
}

func applyPermissionOverrides(rolePermissionKeys []string, overrides []sqlc.ListUserPermissionOverridesRow) []PermissionKey {
	set := make(map[PermissionKey]bool, len(rolePermissionKeys))
	for _, key := range rolePermissionKeys {
		set[PermissionKey(key)] = true
	}
	for _, override := range overrides {
		key := PermissionKey(override.PermissionKey)
		if override.Mode == "deny" {
			delete(set, key)
			continue
		}
		set[key] = true
	}

	keys := make([]PermissionKey, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func resolveVisibility(role sqlc.CustomRole, rule sqlc.UserContactVisibilityRule, hasRule bool) VisibilityScope {
	scope := VisibilityScope{
		Mode:                 VisibilityScopeMode(role.DefaultScopeMode),
		AllowedInstanceIDs:   decodeUUIDSlice(role.DefaultAllowedInstanceIds),
		AllowedPhoneNumbers:  decodeStringSlice(role.DefaultAllowedPhoneNumbers),
		CanViewUnmaskedPhone: role.CanViewUnmaskedPhone,
	}
	if !hasRule {
		return scope
	}

	override := VisibilityScope{
		Mode:                 VisibilityScopeMode(rule.ScopeMode),
		AllowedInstanceIDs:   decodeUUIDSlice(rule.AllowedInstanceIds),
		AllowedPhoneNumbers:  decodeStringSlice(rule.AllowedPhoneNumbers),
		CanViewUnmaskedPhone: rule.CanViewUnmaskedPhone,
	}

	if !rule.InheritRoleScope {
		return override
	}

	scope.AllowedInstanceIDs = unionUUIDs(scope.AllowedInstanceIDs, override.AllowedInstanceIDs)
	scope.AllowedPhoneNumbers = unionStrings(scope.AllowedPhoneNumbers, override.AllowedPhoneNumbers)
	scope.CanViewUnmaskedPhone = scope.CanViewUnmaskedPhone || override.CanViewUnmaskedPhone
	return scope
}

func decodeUserSettings(raw []byte) UserSettings {
	settings := UserSettings{
		Theme:         "light",
		Language:      "en",
		SidebarPinned: true,
	}
	if len(raw) == 0 {
		return settings
	}
	var payload struct {
		Theme         string `json:"theme"`
		Language      string `json:"language"`
		SidebarPinned bool   `json:"sidebarPinned"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings
	}
	if payload.Theme != "" {
		settings.Theme = payload.Theme
	}
	if payload.Language != "" {
		settings.Language = payload.Language
	}
	settings.SidebarPinned = payload.SidebarPinned
	return settings
}

func decodeUUIDSlice(raw []byte) []uuid.UUID {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err == nil {
			result = append(result, parsed)
		}
	}
	return result
}

func decodeStringSlice(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return values
}

func unionUUIDs(left, right []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(left)+len(right))
	merged := make([]uuid.UUID, 0, len(left)+len(right))
	for _, item := range append(left, right...) {
		if _, exists := set[item]; exists {
			continue
		}
		set[item] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}

func unionStrings(left, right []string) []string {
	set := make(map[string]struct{}, len(left)+len(right))
	merged := make([]string, 0, len(left)+len(right))
	for _, item := range append(left, right...) {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := set[trimmed]; exists {
			continue
		}
		set[trimmed] = struct{}{}
		merged = append(merged, trimmed)
	}
	return merged
}

func safeText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
