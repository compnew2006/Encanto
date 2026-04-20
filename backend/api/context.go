//go:build bootstrap

package api

import (
	"net/http"

	"encanto/audit"
)

type updateSettingsRequest struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	SidebarPinned bool   `json:"sidebarPinned"`
}

type updateAvailabilityRequest struct {
	AvailabilityStatus string `json:"availabilityStatus"`
}

func meHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
			return
		}
		audit.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
	}
}

func meOrganizationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
			return
		}
		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"organizations": user.Organizations,
		})
	}
}

func updateSettingsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
			return
		}

		var request updateSettingsRequest
		if err := audit.DecodeJSON(r, &request); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid settings payload.")
			return
		}

		if request.Theme == "" {
			request.Theme = user.Settings.Theme
		}
		if request.Language == "" {
			request.Language = user.Settings.Language
		}

		if err := deps.Store.Queries.UpdateUserSettings(r.Context(), dataSQLCUpdateUserSettingsParams(user.ID, request)); err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "settings_update_failed", "Unable to save the current settings.")
			return
		}

		refreshedUser, err := deps.AccessService.ResolveCurrentUser(r.Context(), mustSession(r))
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to reload the user context.")
			return
		}
		audit.WriteJSON(w, http.StatusOK, map[string]any{"user": refreshedUser})
	}
}

func updateAvailabilityHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
			return
		}

		var request updateAvailabilityRequest
		if err := audit.DecodeJSON(r, &request); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid availability payload.")
			return
		}

		if request.AvailabilityStatus != "available" && request.AvailabilityStatus != "unavailable" && request.AvailabilityStatus != "busy" {
			audit.WriteError(w, http.StatusBadRequest, "invalid_availability", "Availability must be available, unavailable, or busy.")
			return
		}

		if err := deps.Store.Queries.UpdateUserAvailability(r.Context(), dataSQLCUpdateUserAvailabilityParams(user.ID, request.AvailabilityStatus)); err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "availability_update_failed", "Unable to update the availability status.")
			return
		}

		refreshedUser, err := deps.AccessService.ResolveCurrentUser(r.Context(), mustSession(r))
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to reload the user context.")
			return
		}
		audit.WriteJSON(w, http.StatusOK, map[string]any{"user": refreshedUser})
	}
}
