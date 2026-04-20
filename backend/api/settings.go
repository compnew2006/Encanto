package api

import "net/http"



func (s *Server) GetProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := s.store.ProfileForOrg(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	profile, err := s.store.UpdateProfile(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) GetSettingsSummary(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.SettingsForOrg(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) GetGeneralSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.SettingsForOrg(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings.General)
}

func (s *Server) UpdateGeneralSettings(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	var req GeneralSettings
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	settings, err := s.store.UpdateGeneral(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) GetAppearanceSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.SettingsForOrg(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings.Appearance)
}

func (s *Server) UpdateAppearanceSettings(w http.ResponseWriter, r *http.Request) {
	var req AppearanceSettings
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	settings, err := s.store.UpdateAppearance(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) GetChatSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.SettingsForOrg(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings.Chat)
}

func (s *Server) UpdateChatSettings(w http.ResponseWriter, r *http.Request) {
	var req ChatSettings
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	settings, err := s.store.UpdateChatSettings(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) GetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.SettingsForOrg(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings.Notifications)
}

func (s *Server) UpdateCleanupSettings(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CleanupSettings
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	settings, err := s.store.UpdateCleanupSettings(currentOrgID(r), claims.UserID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	var req NotificationSettings
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	settings, err := s.store.UpdateNotificationsSettings(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) RunCleanup(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	job, err := s.store.RunCleanup(currentOrgID(r), claims.UserID)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "notification", map[string]any{"job": job})
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) isAdmin(orgID string) bool {
	// Verify that an admin user exists for this org (fast DB check)
	return s.store.IsUserAdmin(orgID, "")
}
