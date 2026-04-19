package api

import (
	"net/http"

	"encanto/audit"
	"encanto/data/sqlc"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type switchOrgRequest struct {
	OrganizationID string `json:"organizationId"`
}

func loginHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request loginRequest
		if err := audit.DecodeJSON(r, &request); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid login payload.")
			return
		}

		user, err := deps.Store.Queries.GetUserByEmail(r.Context(), request.Email)
		if err != nil {
			audit.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
			return
		}
		if !user.IsActive {
			audit.WriteError(w, http.StatusForbidden, "inactive_user", "This user is inactive.")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
			audit.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
			return
		}

		memberships, err := deps.Store.Queries.ListUserMemberships(r.Context(), user.ID)
		if err != nil || len(memberships) == 0 {
			audit.WriteError(w, http.StatusForbidden, "no_membership", "No active organization membership is available.")
			return
		}

		activeMembership := memberships[0]
		accessToken, refreshToken, expiresAt, err := deps.SessionManager.Issue(r.Context(), user.ID, activeMembership.OrganizationID, user.Email)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "session_issue_failed", "Unable to start a new session.")
			return
		}
		deps.SessionManager.SetCookies(w, accessToken, refreshToken, expiresAt)

		session, err := deps.SessionManager.ParseAccessToken(accessToken)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "session_parse_failed", "Unable to load the new session.")
			return
		}
		_ = deps.Store.Queries.UpdateUserLastLogin(r.Context(), user.ID)

		currentUser, err := deps.AccessService.ResolveCurrentUser(r.Context(), session)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to load the user context.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"user": currentUser,
		})
	}
}

func refreshHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshCookie, err := r.Cookie(deps.Config.RefreshCookieName)
		if err != nil {
			audit.WriteError(w, http.StatusUnauthorized, "missing_refresh_cookie", "No refresh session is available.")
			return
		}

		accessToken, refreshToken, expiresAt, session, err := deps.SessionManager.Rotate(r.Context(), refreshCookie.Value)
		if err != nil {
			audit.WriteError(w, http.StatusUnauthorized, "refresh_failed", "The session refresh failed.")
			return
		}
		deps.SessionManager.SetCookies(w, accessToken, refreshToken, expiresAt)

		currentUser, err := deps.AccessService.ResolveCurrentUser(r.Context(), session)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to load the user context.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"user": currentUser,
		})
	}
}

func logoutHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if refreshCookie, err := r.Cookie(deps.Config.RefreshCookieName); err == nil {
			_ = deps.SessionManager.Invalidate(r.Context(), refreshCookie.Value)
		}
		deps.SessionManager.ClearCookies(w)
		audit.WriteJSON(w, http.StatusOK, map[string]any{"loggedOut": true})
	}
}

func switchOrgHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r.Context())
		if !ok {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
			return
		}

		var request switchOrgRequest
		if err := audit.DecodeJSON(r, &request); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid organization switch payload.")
			return
		}

		organizationID, err := parseUUIDParam(request.OrganizationID)
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_organization", "The target organization is invalid.")
			return
		}

		if _, err := deps.Store.Queries.GetMembershipByUserAndOrg(r.Context(), sqlc.GetMembershipByUserAndOrgParams{
			UserID:         session.UserID,
			OrganizationID: organizationID,
		}); err != nil {
			audit.WriteError(w, http.StatusForbidden, "organization_denied", "You do not have access to that organization.")
			return
		}

		accessToken, refreshToken, expiresAt, err := deps.SessionManager.RotateForOrganization(r.Context(), session, organizationID)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "switch_failed", "Unable to switch the active organization.")
			return
		}
		deps.SessionManager.SetCookies(w, accessToken, refreshToken, expiresAt)

		newSession, err := deps.SessionManager.ParseAccessToken(accessToken)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "session_parse_failed", "Unable to load the switched session.")
			return
		}

		currentUser, err := deps.AccessService.ResolveCurrentUser(r.Context(), newSession)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to load the switched user context.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{
			"user": currentUser,
		})
	}
}

func requireAuth(deps Dependencies) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(deps.Config.AccessCookieName)
			if err != nil {
				audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
				return
			}

			session, err := deps.SessionManager.ParseAccessToken(cookie.Value)
			if err != nil {
				audit.WriteError(w, http.StatusUnauthorized, "invalid_session", "The current session is invalid.")
				return
			}

			next.ServeHTTP(w, r.WithContext(withSession(r.Context(), session)))
		})
	}
}

func loadCurrentUser(deps Dependencies) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := sessionFromContext(r.Context())
			if !ok {
				audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
				return
			}

			user, err := deps.AccessService.ResolveCurrentUser(r.Context(), session)
			if err != nil {
				audit.WriteError(w, http.StatusUnauthorized, "user_context_failed", "Unable to load the current user context.")
				return
			}

			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
		})
	}
}

func _routerGuard(_ chi.Router) {}
