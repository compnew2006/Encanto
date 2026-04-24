package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"encanto/audit"
	"encanto/config"
	"encanto/core"
	"encanto/data"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Dependencies holds all service-level dependencies injected into HTTP handlers.
type Dependencies struct {
	Config         config.Config
	Store          *data.Store
	SessionManager *core.SessionManager
	AccessService  *core.AccessService
	ChatService    *core.ChatService
}

// ---------- Context keys ----------

type contextKey string

const (
	contextKeyUser    contextKey = "current_user"
	contextKeySession contextKey = "current_session"
)

// ---------- Middleware ----------

// requireAuth validates the access token cookie and attaches the session to the request context.
func requireAuth(deps Dependencies) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(deps.Config.AccessCookieName)
			if err != nil || cookie.Value == "" {
				audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
				return
			}

			session, err := deps.SessionManager.ParseAccessToken(cookie.Value)
			if err != nil {
				audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "Session is invalid or expired.")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeySession, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// loadCurrentUser resolves the full user context from the session and attaches it to the request.
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
				audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unable to resolve user context.")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ---------- Context helpers ----------

func userFromContext(ctx context.Context) (core.CurrentUserContext, bool) {
	user, ok := ctx.Value(contextKeyUser).(core.CurrentUserContext)
	return user, ok
}

func sessionFromContext(ctx context.Context) (core.CurrentSession, bool) {
	session, ok := ctx.Value(contextKeySession).(core.CurrentSession)
	return session, ok
}

// ---------- Permission helpers ----------

func requirePermission(svc *core.AccessService, user core.CurrentUserContext, key string) error {
	if !svc.HasPermission(user, core.PermissionKey(key)) {
		return errors.New(svc.PermissionDeniedReason(core.PermissionKey(key)))
	}
	return nil
}

// ---------- Auth handlers ----------

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginPayload
		if err := audit.DecodeJSON(r, &req); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid login payload.")
			return
		}

		dbUser, err := deps.Store.Queries.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				audit.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
				return
			}
			audit.WriteError(w, http.StatusInternalServerError, "login_failed", "Unable to process login.")
			return
		}

		if !dbUser.IsActive {
			audit.WriteError(w, http.StatusUnauthorized, "account_inactive", "This account is not active.")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.Password)); err != nil {
			audit.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
			return
		}

		memberships, err := deps.Store.Queries.ListUserMemberships(r.Context(), dbUser.ID)
		if err != nil || len(memberships) == 0 {
			audit.WriteError(w, http.StatusUnauthorized, "no_membership", "User has no active organization memberships.")
			return
		}

		defaultOrg := memberships[0]
		for _, m := range memberships {
			if m.IsDefault {
				defaultOrg = m
				break
			}
		}

		accessToken, refreshToken, expiresAt, err := deps.SessionManager.Issue(r.Context(), dbUser.ID, defaultOrg.OrganizationID, dbUser.Email)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "session_failed", "Unable to create session.")
			return
		}

		deps.SessionManager.SetCookies(w, accessToken, refreshToken, expiresAt)
		_ = deps.Store.Queries.UpdateUserLastLogin(r.Context(), dbUser.ID)

		session, _ := deps.SessionManager.ParseAccessToken(accessToken)
		user, err := deps.AccessService.ResolveCurrentUser(r.Context(), session)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to load user context.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
	}
}

func refreshHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(deps.Config.RefreshCookieName)
		if err != nil || cookie.Value == "" {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "Refresh token is missing.")
			return
		}

		accessToken, refreshToken, expiresAt, session, err := deps.SessionManager.Rotate(r.Context(), cookie.Value)
		if err != nil {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "Refresh token is invalid or expired.")
			return
		}

		deps.SessionManager.SetCookies(w, accessToken, refreshToken, expiresAt)

		user, err := deps.AccessService.ResolveCurrentUser(r.Context(), session)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to reload the user context.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
	}
}

func logoutHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(deps.Config.RefreshCookieName)
		if err == nil && cookie.Value != "" {
			_ = deps.SessionManager.Invalidate(r.Context(), cookie.Value)
		}
		deps.SessionManager.ClearCookies(w)
		audit.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func switchOrgHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r.Context())
		if !ok {
			audit.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid session is required.")
			return
		}

		var payload struct {
			OrganizationID string `json:"organizationId"`
		}
		if err := audit.DecodeJSON(r, &payload); err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_payload", "Invalid switch-org payload.")
			return
		}

		orgID, err := uuid.Parse(payload.OrganizationID)
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_org", "Invalid organization identifier.")
			return
		}

		accessToken, refreshToken, expiresAt, err := deps.SessionManager.RotateForOrganization(r.Context(), session, orgID)
		if err != nil {
			audit.WriteError(w, http.StatusForbidden, "switch_failed", "Unable to switch organization.")
			return
		}

		deps.SessionManager.SetCookies(w, accessToken, refreshToken, expiresAt)

		newSession, err := deps.SessionManager.ParseAccessToken(accessToken)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "session_parse_failed", "Unable to parse new session.")
			return
		}

		user, err := deps.AccessService.ResolveCurrentUser(r.Context(), newSession)
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "user_context_failed", "Unable to reload the user context.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
	}
}

// ---------- Utility helpers ----------

// parseUUIDParam parses a UUID from a URL parameter string.
func parseUUIDParam(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

// writeNotFound writes a standard 404 response.
func writeNotFound(w http.ResponseWriter) {
	audit.WriteError(w, http.StatusNotFound, "not_found", "The requested resource was not found.")
}

// jsonBytes marshals a value to JSON bytes, returning nil on error.
func jsonBytes(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// normalizeRolePermissionKeys filters and deduplicates permission keys for a role.
func normalizeRolePermissionKeys(keys []string, _ core.VisibilityScopeMode) []string {
	seen := make(map[string]struct{}, len(keys))
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}
