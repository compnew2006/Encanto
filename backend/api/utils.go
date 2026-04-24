package api

import (
	"encoding/json"
	"net/http"

	"encanto/core"
	"encanto/data/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func mustUser(r *http.Request) core.CurrentUserContext {
	user, _ := userFromContext(r.Context())
	return user
}

func mustSession(r *http.Request) core.CurrentSession {
	session, _ := sessionFromContext(r.Context())
	return session
}

func dataSQLCUpdateUserSettingsParams(userID uuid.UUID, request updateSettingsRequest) sqlc.UpdateUserSettingsParams {
	return sqlc.UpdateUserSettingsParams{
		ID:       userID,
		Settings: jsonBytes(map[string]any{"theme": request.Theme, "language": request.Language, "sidebarPinned": request.SidebarPinned}),
	}
}

func dataSQLCUpdateUserAvailabilityParams(userID uuid.UUID, availability string) sqlc.UpdateUserAvailabilityParams {
	return sqlc.UpdateUserAvailabilityParams{
		ID:                 userID,
		AvailabilityStatus: availability,
	}
}

func coreDecodeUUIDStrings(raw []byte) []uuid.UUID {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	_ = json.Unmarshal(raw, &values)
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if parsed, err := uuid.Parse(value); err == nil {
			result = append(result, parsed)
		}
	}
	return result
}

func coreEncodeUUIDStrings(values []byte) []string {
	if len(values) == 0 {
		return nil
	}
	var result []string
	_ = json.Unmarshal(values, &result)
	return result
}

func coreDecodeStringValues(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	_ = json.Unmarshal(raw, &values)
	return values
}

func coreDecodeSettings(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	_ = json.Unmarshal(raw, &payload)
	return payload
}

func pgText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
