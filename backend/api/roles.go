package api

import (
	"context"
	"net/http"

	"encanto/audit"
	"encanto/core"
	"encanto/data/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rolePayload struct {
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	PermissionKeys    []string              `json:"permissionKeys"`
	DefaultVisibility roleVisibilityPayload `json:"defaultVisibility"`
}

type roleVisibilityPayload struct {
	Mode                 core.VisibilityScopeMode `json:"mode"`
	AllowedInstanceIDs   []string                 `json:"allowedInstanceIds"`
	AllowedPhoneNumbers  []string                 `json:"allowedPhoneNumbers"`
	CanViewUnmaskedPhone bool                     `json:"canViewUnmaskedPhone"`
}

func listPermissionsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"permissions": deps.AccessService.Catalog().All(),
		})
	}
}

func listRolesHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "roles.manage"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Role management is not allowed.", err.Error())
			return
		}

		roles, err := deps.Store.Queries.ListRoles(r.Context(), user.CurrentOrganization.ID)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "roles_list_failed", "Unable to load roles.")
			return
		}

		serialized := make([]core.RoleDefinition, 0, len(roles))
		for _, role := range roles {
			keys, err := deps.Store.Queries.ListRolePermissionKeys(r.Context(), role.ID)
			if err != nil {
				audit.WriteError(w, http.StatusInternalServerError, "role_permissions_failed", "Unable to load role permissions.")
				return
			}
			serialized = append(serialized, core.RoleDefinition{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				IsSystem:    role.IsSystem,
				IsDefault:   role.IsDefault,
				PermissionKeys: func() []core.PermissionKey {
					result := make([]core.PermissionKey, 0, len(keys))
					for _, key := range keys {
						result = append(result, core.PermissionKey(key))
					}
					return result
				}(),
				DefaultVisibility: core.VisibilityScope{
					Mode:                 core.VisibilityScopeMode(role.DefaultScopeMode),
					AllowedInstanceIDs:   coreDecodeUUIDStrings(role.DefaultAllowedInstanceIds),
					AllowedPhoneNumbers:  coreDecodeStringValues(role.DefaultAllowedPhoneNumbers),
					CanViewUnmaskedPhone: role.CanViewUnmaskedPhone,
				},
			})
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"roles": serialized})
	}
}

func createRoleHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "roles.manage"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Role management is not allowed.", err.Error())
			return
		}

		var payload rolePayload
		if err := audit.DecodeJSON(r, &payload); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid role payload.")
			return
		}

		role, err := persistRole(r.Context(), deps, user, payload, nil)
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "role_create_failed", err.Error())
			return
		}
		audit.WriteJSON(w, http.StatusCreated, map[string]any{"role": role})
	}
}

func updateRoleHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "roles.manage"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Role management is not allowed.", err.Error())
			return
		}

		roleID, err := parseUUIDParam(chi.URLParam(r, "roleID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_role", "Invalid role identifier.")
			return
		}

		var payload rolePayload
		if err := audit.DecodeJSON(r, &payload); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid role payload.")
			return
		}

		role, err := persistRole(r.Context(), deps, user, payload, &roleID)
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "role_update_failed", err.Error())
			return
		}
		audit.WriteJSON(w, http.StatusOK, map[string]any{"role": role})
	}
}

func deleteRoleHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "roles.manage"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Role management is not allowed.", err.Error())
			return
		}

		roleID, err := parseUUIDParam(chi.URLParam(r, "roleID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_role", "Invalid role identifier.")
			return
		}

		if err := deps.Store.Queries.SoftDeleteRole(r.Context(), sqlc.SoftDeleteRoleParams{
			ID:             roleID,
			OrganizationID: user.CurrentOrganization.ID,
		}); err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "role_delete_failed", "Unable to delete the selected role.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
	}
}

func persistRole(ctx context.Context, deps Dependencies, user core.CurrentUserContext, payload rolePayload, roleID *uuid.UUID) (core.RoleDefinition, error) {
	if payload.Name == "" {
		return core.RoleDefinition{}, pgx.ErrNoRows
	}

	normalizedKeys := normalizeRolePermissionKeys(payload.PermissionKeys, payload.DefaultVisibility.Mode)
	permissionRows, err := deps.Store.Queries.GetPermissionsByKeys(ctx, normalizedKeys)
	if err != nil {
		return core.RoleDefinition{}, err
	}

	var storedRole sqlc.CustomRole
	if roleID == nil {
		row, err := deps.Store.Queries.CreateRole(ctx, sqlc.CreateRoleParams{
			OrganizationID:             user.CurrentOrganization.ID,
			Name:                       payload.Name,
			Description:                payload.Description,
			IsSystem:                   false,
			IsDefault:                  false,
			DefaultScopeMode:           string(payload.DefaultVisibility.Mode),
			DefaultAllowedInstanceIds:  jsonBytes(payload.DefaultVisibility.AllowedInstanceIDs),
			DefaultAllowedPhoneNumbers: jsonBytes(payload.DefaultVisibility.AllowedPhoneNumbers),
			CanViewUnmaskedPhone:       payload.DefaultVisibility.CanViewUnmaskedPhone,
		})
		if err != nil {
			return core.RoleDefinition{}, err
		}
		storedRole = row
	} else {
		row, err := deps.Store.Queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
			ID:                         *roleID,
			OrganizationID:             user.CurrentOrganization.ID,
			Name:                       payload.Name,
			Description:                payload.Description,
			DefaultScopeMode:           string(payload.DefaultVisibility.Mode),
			DefaultAllowedInstanceIds:  jsonBytes(payload.DefaultVisibility.AllowedInstanceIDs),
			DefaultAllowedPhoneNumbers: jsonBytes(payload.DefaultVisibility.AllowedPhoneNumbers),
			CanViewUnmaskedPhone:       payload.DefaultVisibility.CanViewUnmaskedPhone,
		})
		if err != nil {
			return core.RoleDefinition{}, err
		}
		storedRole = row
	}

	if err := deps.Store.Queries.DeleteRolePermissionsByRole(ctx, storedRole.ID); err != nil {
		return core.RoleDefinition{}, err
	}
	for _, permission := range permissionRows {
		if err := deps.Store.Queries.InsertRolePermission(ctx, sqlc.InsertRolePermissionParams{
			CustomRoleID: storedRole.ID,
			PermissionID: permission.ID,
		}); err != nil {
			return core.RoleDefinition{}, err
		}
	}

	return core.RoleDefinition{
		ID:          storedRole.ID,
		Name:        storedRole.Name,
		Description: storedRole.Description,
		IsSystem:    storedRole.IsSystem,
		IsDefault:   storedRole.IsDefault,
		PermissionKeys: func() []core.PermissionKey {
			result := make([]core.PermissionKey, 0, len(normalizedKeys))
			for _, key := range normalizedKeys {
				result = append(result, core.PermissionKey(key))
			}
			return result
		}(),
		DefaultVisibility: core.VisibilityScope{
			Mode:                 payload.DefaultVisibility.Mode,
			AllowedInstanceIDs:   coreDecodeUUIDStrings(jsonBytes(payload.DefaultVisibility.AllowedInstanceIDs)),
			AllowedPhoneNumbers:  payload.DefaultVisibility.AllowedPhoneNumbers,
			CanViewUnmaskedPhone: payload.DefaultVisibility.CanViewUnmaskedPhone,
		},
	}, nil
}
