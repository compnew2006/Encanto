package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"encanto/audit"
	"encanto/config"
	"encanto/core"
	"encanto/data"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Dependencies struct {
	Config         config.Config
	Store          *data.Store
	SessionManager *core.SessionManager
	AccessService  *core.AccessService
	ChatService    *core.ChatService
}

type contextKey string

const (
	sessionContextKey contextKey = "session"
	userContextKey    contextKey = "user"
)

func sessionFromContext(ctx context.Context) (core.CurrentSession, bool) {
	session, ok := ctx.Value(sessionContextKey).(core.CurrentSession)
	return session, ok
}

func userFromContext(ctx context.Context) (core.CurrentUserContext, bool) {
	user, ok := ctx.Value(userContextKey).(core.CurrentUserContext)
	return user, ok
}

func withSession(ctx context.Context, session core.CurrentSession) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

func withUser(ctx context.Context, user core.CurrentUserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func requirePermission(access *core.AccessService, user core.CurrentUserContext, key core.PermissionKey) error {
	if access.HasPermission(user, key) {
		return nil
	}
	return errors.New(access.PermissionDeniedReason(key))
}

func parseUUIDParam(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid param: %w", err)
	}
	return parsed, nil
}

func jsonBytes(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}

func normalizeRolePermissionKeys(keys []string, scopeMode core.VisibilityScopeMode) []string {
	filtered := make([]string, 0, len(keys)+2)
	for _, key := range keys {
		if key == "contacts.scope.all" || key == "contacts.scope.instance_only" || key == "contacts.scope.allowed_numbers" {
			continue
		}
		filtered = append(filtered, key)
	}

	switch scopeMode {
	case core.ScopeAllContacts:
		filtered = append(filtered, "contacts.scope.all")
	case core.ScopeInstancesOnly:
		filtered = append(filtered, "contacts.scope.instance_only")
	case core.ScopeAllowedNumbersOnly:
		filtered = append(filtered, "contacts.scope.allowed_numbers")
	case core.ScopeInstancesPlusAllowedNumbers:
		filtered = append(filtered, "contacts.scope.instance_only", "contacts.scope.allowed_numbers")
	}

	slices.Sort(filtered)
	return slices.Compact(filtered)
}

func writeNotFound(w http.ResponseWriter) {
	audit.WriteError(w, http.StatusNotFound, "not_found", "The requested record could not be found.")
}

func isNoRows(err error) bool {
	return err == nil || err == pgx.ErrNoRows
}
