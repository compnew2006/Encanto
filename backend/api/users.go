//go:build bootstrap

package api

import (
	"errors"
	"net/http"
	"slices"

	"encanto/audit"
	"encanto/core"
	"encanto/data/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type sendRestrictionsPayload struct {
	AllowPermissionKeys []string `json:"allowPermissionKeys"`
	DenyPermissionKeys  []string `json:"denyPermissionKeys"`
}

type visibilityPayload struct {
	ScopeMode            core.VisibilityScopeMode `json:"scopeMode"`
	AllowedInstanceIDs   []string                 `json:"allowedInstanceIds"`
	AllowedPhoneNumbers  []string                 `json:"allowedPhoneNumbers"`
	InheritRoleScope     bool                     `json:"inheritRoleScope"`
	CanViewUnmaskedPhone bool                     `json:"canViewUnmaskedPhone"`
}

type updateUserPayload struct {
	AvailabilityStatus string `json:"availabilityStatus"`
}

func listUsersHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User listing is not allowed.", err.Error())
			return
		}

		rows, err := deps.Store.Queries.ListUsersByOrganization(r.Context(), user.CurrentOrganization.ID)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "users_list_failed", "Unable to load users.")
			return
		}

		users := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			users = append(users, map[string]any{
				"id":                 row.ID,
				"email":              row.Email,
				"fullName":           row.FullName,
				"avatarUrl":          pgText(row.AvatarUrl),
				"isActive":           row.IsActive,
				"availabilityStatus": row.AvailabilityStatus,
				"roleId":             row.RoleID,
				"roleName":           row.RoleName,
				"settings":           coreDecodeSettings(row.Settings),
			})
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"users": users})
	}
}

func getUserHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User detail is not allowed.", err.Error())
			return
		}

		targetID, err := parseUUIDParam(chi.URLParam(r, "userID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_user", "Invalid user identifier.")
			return
		}

		rows, err := deps.Store.Queries.ListUsersByOrganization(r.Context(), user.CurrentOrganization.ID)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_lookup_failed", "Unable to load the user.")
			return
		}
		for _, row := range rows {
			if row.ID == targetID {
				audit.WriteJSON(w, http.StatusOK, map[string]any{
					"user": map[string]any{
						"id":                 row.ID,
						"email":              row.Email,
						"fullName":           row.FullName,
						"avatarUrl":          pgText(row.AvatarUrl),
						"isActive":           row.IsActive,
						"availabilityStatus": row.AvailabilityStatus,
						"roleId":             row.RoleID,
						"roleName":           row.RoleName,
						"settings":           coreDecodeSettings(row.Settings),
					},
				})
				return
			}
		}
		writeNotFound(w)
	}
}

func updateUserHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.update"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User updates are not allowed.", err.Error())
			return
		}

		targetID, err := parseUUIDParam(chi.URLParam(r, "userID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_user", "Invalid user identifier.")
			return
		}

		var payload updateUserPayload
		if err := audit.DecodeJSON(r, &payload); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid user payload.")
			return
		}
		if payload.AvailabilityStatus != "available" && payload.AvailabilityStatus != "busy" && payload.AvailabilityStatus != "unavailable" {
			audit.WriteError(w, http.StatusBadRequest, "invalid_availability", "Availability must be available, unavailable, or busy.")
			return
		}

		if err := deps.Store.Queries.UpdateUserAvailability(r.Context(), dataSQLCUpdateUserAvailabilityParams(targetID, payload.AvailabilityStatus)); err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_update_failed", "Unable to update the user.")
			return
		}

		rows, err := deps.Store.Queries.ListUsersByOrganization(r.Context(), user.CurrentOrganization.ID)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_lookup_failed", "Unable to reload the user.")
			return
		}
		for _, row := range rows {
			if row.ID == targetID {
				audit.WriteJSON(w, http.StatusOK, map[string]any{
					"user": map[string]any{
						"id":                 row.ID,
						"email":              row.Email,
						"fullName":           row.FullName,
						"avatarUrl":          pgText(row.AvatarUrl),
						"isActive":           row.IsActive,
						"availabilityStatus": row.AvailabilityStatus,
						"roleId":             row.RoleID,
						"roleName":           row.RoleName,
						"settings":           coreDecodeSettings(row.Settings),
					},
				})
				return
			}
		}

		writeNotFound(w)
	}
}

func getSendRestrictionsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User detail is not allowed.", err.Error())
			return
		}

		targetID, err := parseUUIDParam(chi.URLParam(r, "userID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_user", "Invalid user identifier.")
			return
		}

		rows, err := deps.Store.Queries.ListUserPermissionOverrides(r.Context(), sqlc.ListUserPermissionOverridesParams{
			OrganizationID: user.CurrentOrganization.ID,
			UserID:         targetID,
		})
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "override_load_failed", "Unable to load send restrictions.")
			return
		}

		allow := []string{}
		deny := []string{}
		for _, row := range rows {
			if row.PermissionKey != "messages.send" && row.PermissionKey != "chats.unclaimed.send" {
				continue
			}
			if row.Mode == "allow" {
				allow = append(allow, row.PermissionKey)
			} else {
				deny = append(deny, row.PermissionKey)
			}
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"allowPermissionKeys": allow,
			"denyPermissionKeys":  deny,
		})
	}
}

func updateSendRestrictionsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.update"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User updates are not allowed.", err.Error())
			return
		}

		targetID, err := parseUUIDParam(chi.URLParam(r, "userID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_user", "Invalid user identifier.")
			return
		}

		var payload sendRestrictionsPayload
		if err := audit.DecodeJSON(r, &payload); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid send restriction payload.")
			return
		}

		allowedKeys := []string{"messages.send", "chats.unclaimed.send"}
		if err := deps.Store.Queries.DeleteUserPermissionOverrides(r.Context(), sqlc.DeleteUserPermissionOverridesParams{
			OrganizationID: user.CurrentOrganization.ID,
			UserID:         targetID,
		}); err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "override_clear_failed", "Unable to clear old overrides.")
			return
		}

		for _, key := range payload.AllowPermissionKeys {
			if !slices.Contains(allowedKeys, key) {
				continue
			}
			if err := deps.Store.Queries.InsertUserPermissionOverride(r.Context(), sqlc.InsertUserPermissionOverrideParams{
				OrganizationID: user.CurrentOrganization.ID,
				UserID:         targetID,
				PermissionKey:  key,
				Mode:           "allow",
			}); err != nil {
				audit.WriteError(w, http.StatusInternalServerError, "override_save_failed", "Unable to save send overrides.")
				return
			}
		}
		for _, key := range payload.DenyPermissionKeys {
			if !slices.Contains(allowedKeys, key) {
				continue
			}
			if err := deps.Store.Queries.InsertUserPermissionOverride(r.Context(), sqlc.InsertUserPermissionOverrideParams{
				OrganizationID: user.CurrentOrganization.ID,
				UserID:         targetID,
				PermissionKey:  key,
				Mode:           "deny",
			}); err != nil {
				audit.WriteError(w, http.StatusInternalServerError, "override_save_failed", "Unable to save send overrides.")
				return
			}
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"saved": true})
	}
}

func getContactVisibilityHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User detail is not allowed.", err.Error())
			return
		}

		targetID, err := parseUUIDParam(chi.URLParam(r, "userID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_user", "Invalid user identifier.")
			return
		}

		rule, err := deps.Store.Queries.GetUserVisibilityRule(r.Context(), sqlc.GetUserVisibilityRuleParams{
			OrganizationID: user.CurrentOrganization.ID,
			UserID:         targetID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				audit.WriteJSON(w, http.StatusOK, map[string]any{
					"rule": map[string]any{
						"scopeMode":            core.ScopeAllContacts,
						"allowedInstanceIds":   []string{},
						"allowedPhoneNumbers":  []string{},
						"inheritRoleScope":     true,
						"canViewUnmaskedPhone": true,
					},
				})
				return
			}
			audit.WriteError(w, http.StatusInternalServerError, "visibility_load_failed", "Unable to load contact visibility.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"rule": map[string]any{
				"scopeMode":            rule.ScopeMode,
				"allowedInstanceIds":   coreEncodeUUIDStrings(rule.AllowedInstanceIds),
				"allowedPhoneNumbers":  coreDecodeStringValues(rule.AllowedPhoneNumbers),
				"inheritRoleScope":     rule.InheritRoleScope,
				"canViewUnmaskedPhone": rule.CanViewUnmaskedPhone,
			},
		})
	}
}

func updateContactVisibilityHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "users.update"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "User updates are not allowed.", err.Error())
			return
		}

		targetID, err := parseUUIDParam(chi.URLParam(r, "userID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_user", "Invalid user identifier.")
			return
		}

		var payload visibilityPayload
		if err := audit.DecodeJSON(r, &payload); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid contact visibility payload.")
			return
		}

		rule, err := deps.Store.Queries.UpsertUserVisibilityRule(r.Context(), sqlc.UpsertUserVisibilityRuleParams{
			OrganizationID:       user.CurrentOrganization.ID,
			UserID:               targetID,
			ScopeMode:            string(payload.ScopeMode),
			AllowedInstanceIds:   jsonBytes(payload.AllowedInstanceIDs),
			AllowedPhoneNumbers:  jsonBytes(payload.AllowedPhoneNumbers),
			InheritRoleScope:     payload.InheritRoleScope,
			CanViewUnmaskedPhone: payload.CanViewUnmaskedPhone,
		})
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "visibility_update_failed", "Unable to save contact visibility.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"rule": map[string]any{
				"scopeMode":            rule.ScopeMode,
				"allowedInstanceIds":   coreEncodeUUIDStrings(rule.AllowedInstanceIds),
				"allowedPhoneNumbers":  coreDecodeStringValues(rule.AllowedPhoneNumbers),
				"inheritRoleScope":     rule.InheritRoleScope,
				"canViewUnmaskedPhone": rule.CanViewUnmaskedPhone,
			},
		})
	}
}
