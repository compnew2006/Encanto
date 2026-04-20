package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	store *PGStore
	hub   *RealtimeHub
}

func NewServer(store *PGStore) *Server {
	hub := NewRealtimeHub()
	return &Server{
		store: store,
		hub:   hub,
	}
}

func (s *Server) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/auth/login", s.Login)
	r.Post("/auth/logout", s.Logout)

	r.Group(func(r chi.Router) {
		r.Use(RequireAuth)

		r.Get("/me", s.Me)
		r.Post("/auth/switch-org", s.SwitchOrg)
		r.Get("/auth/ws-token", s.WSToken)

		r.Get("/chats", s.GetWorkspace)
		r.Get("/chats/{contactID}", s.GetWorkspace)
		r.Get("/chats/closed", s.ListClosedChats)
		r.Post("/chats/direct", s.CreateDirectChat)
		r.Post("/chats/{contactID}/messages", s.SendMessage)
		r.Post("/chats/{contactID}/messages/{messageID}/retry", s.RetryMessage)
		r.Post("/chats/{contactID}/messages/{messageID}/revoke", s.RevokeMessage)
		r.Post("/chats/{contactID}/assign", s.AssignContact)
		r.Post("/chats/{contactID}/unassign", s.UnassignContact)
		r.Post("/chats/{contactID}/pin", s.TogglePin)
		r.Post("/chats/{contactID}/hide", s.ToggleHide)
		r.Post("/chats/{contactID}/close", s.CloseChat)
		r.Post("/chats/{contactID}/reopen", s.ReopenChat)
		r.Put("/chats/{contactID}/reopen", s.ReopenChat)
		r.Post("/chats/{contactID}/notes", s.AddNote)
		r.Post("/chats/{contactID}/collaborators", s.AddCollaborator)

		r.Get("/contacts", s.ListContacts)
		r.Post("/contacts", s.CreateContact)
		r.Get("/contacts/export", s.ExportContacts)
		r.Post("/contacts/import", s.ImportContacts)
		r.Put("/contacts/{contactID}", s.UpdateContact)
		r.Delete("/contacts/{contactID}", s.DeleteContact)

		r.Get("/notifications", s.ListNotifications)
		r.Post("/notifications/read-all", s.MarkAllNotificationsRead)
		r.Get("/statuses", s.ListStatuses)
		r.Post("/statuses", s.CreateStatus)

		r.Get("/profile", s.GetProfile)
		r.Put("/profile", s.UpdateProfile)

		r.Get("/settings/summary", s.GetSettingsSummary)
		r.Get("/settings/general", s.GetGeneralSettings)
		r.Put("/settings/general", s.UpdateGeneralSettings)
		r.Get("/settings/limits", s.GetSettingsLimits)
		r.Get("/settings/appearance", s.GetAppearanceSettings)
		r.Put("/settings/appearance", s.UpdateAppearanceSettings)
		r.Get("/settings/chat", s.GetChatSettings)
		r.Put("/settings/chat", s.UpdateChatSettings)
		r.Get("/settings/notifications", s.GetNotificationSettings)
		r.Put("/settings/uploads-cleanup", s.UpdateCleanupSettings)
		r.Put("/settings/notifications", s.UpdateNotificationSettings)
		r.Post("/settings/uploads-cleanup/run", s.RunCleanup)

		r.Get("/license/bootstrap", s.GetLicenseBootstrap)
		r.Post("/license/activate", s.ActivateLicense)

		r.Get("/analytics/agents/summary", s.GetAgentAnalyticsSummary)
		r.Get("/analytics/agents/transfers", s.GetAgentTransferTrends)
		r.Get("/analytics/agents/sources", s.GetAgentSourceBreakdown)
		r.Get("/analytics/agents/comparison", s.GetAgentComparison)
		r.Get("/analytics/agents/ratings", s.GetAgentRatings)
		r.Get("/analytics/agents/export", s.ExportAgentAnalytics)

		r.Get("/campaigns", s.ListCampaigns)
		r.Post("/campaigns", s.CreateCampaign)
		r.Get("/campaigns/{campaignID}", s.GetCampaign)
		r.Put("/campaigns/{campaignID}", s.UpdateCampaign)
		r.Delete("/campaigns/{campaignID}", s.DeleteCampaign)
		r.Post("/campaigns/{campaignID}/launch", s.LaunchCampaign)
		r.Post("/campaigns/{campaignID}/pause", s.PauseCampaign)
		r.Post("/campaigns/{campaignID}/resume", s.ResumeCampaign)
		r.Get("/campaigns/{campaignID}/runs", s.ListCampaignRuns)
		r.Get("/campaigns/{campaignID}/recipients", s.ListCampaignRecipients)

		r.Get("/instances", s.ListInstances)
		r.Get("/instances/health", s.ListInstanceHealth)
		r.Post("/instances", s.CreateInstance)
		r.Delete("/instances/{instanceID}", s.DeleteInstance)
		r.Put("/instances/{instanceID}/name", s.UpdateInstanceName)
		r.Post("/instances/{instanceID}/connect", s.ConnectInstance)
		r.Post("/instances/{instanceID}/disconnect", s.DisconnectInstance)
		r.Post("/instances/{instanceID}/recover", s.RecoverInstance)
		r.Put("/instances/{instanceID}/settings", s.UpdateInstanceSettings)
		r.Put("/instances/{instanceID}/call-auto-reject", s.UpdateInstanceCallPolicy)
		r.Put("/instances/{instanceID}/auto-campaign", s.UpdateInstanceAutoCampaign)

		r.Get("/jobs", s.ListJobs)
		r.Get("/jobs/{jobID}", s.GetJob)
		r.Get("/webhooks", s.ListWebhooks)
		r.Get("/webhooks/{webhookID}/deliveries", s.ListWebhookDeliveries)
		r.Post("/webhooks/{webhookID}/deliveries/{deliveryID}/retry", s.RetryWebhookDelivery)
		r.Get("/audit-logs", s.ListAuditLogs)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func currentOrgID(r *http.Request) string {
	cookie, err := r.Cookie("org_context")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return "org-1"
}

func currentClaims(r *http.Request) (*Claims, error) {
	claims, ok := r.Context().Value("user").(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("unauthorized")
	}
	return claims, nil
}
