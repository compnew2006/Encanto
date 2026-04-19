package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type CreateInstanceRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type UpdateInstanceNameRequest struct {
	Name string `json:"name"`
}

type InstanceHealthSummary struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	UptimeLabel   string    `json:"uptime_label"`
	QueueDepth    int       `json:"queue_depth"`
	SentToday     int       `json:"sent_today"`
	ReceivedToday int       `json:"received_today"`
	FailedToday   int       `json:"failed_today"`
	ErrorRate     string    `json:"error_rate"`
	ObservedAt    time.Time `json:"observed_at"`
}

func (s *Server) ListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := s.store.ListInstances(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"instances": instances})
}

func (s *Server) ListInstanceHealth(w http.ResponseWriter, r *http.Request) {
	health, err := s.store.ListInstanceHealth(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"health": health})
}

func (s *Server) CreateInstance(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	var req CreateInstanceRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.CreateInstance(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"instance": instance})
}

func (s *Server) UpdateInstanceName(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	var req UpdateInstanceNameRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceName(currentOrgID(r), chi.URLParam(r, "instanceID"), req.Name)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}

func (s *Server) ConnectInstance(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	instance, inbound, err := s.store.ConnectInstance(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "instance_connected", map[string]any{"instance": instance})
	if inbound != nil {
		s.hub.Publish(currentOrgID(r), "new_message", map[string]any{"contact_id": inbound.ContactID, "message": inbound})
	}
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance, "inbound_message": inbound})
}

func (s *Server) DisconnectInstance(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	instance, err := s.store.DisconnectInstance(currentOrgID(r), chi.URLParam(r, "instanceID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "instance_disconnected", map[string]any{"instance": instance})
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}

func (s *Server) RecoverInstance(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	instance, err := s.store.RecoverInstance(currentOrgID(r), chi.URLParam(r, "instanceID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "instance_recovering", map[string]any{"instance": instance})
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}

func (s *Server) UpdateInstanceSettings(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	var req InstanceSettings
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceSettings(currentOrgID(r), chi.URLParam(r, "instanceID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}

func (s *Server) UpdateInstanceCallPolicy(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	var req InstanceCallPolicy
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceCallPolicy(currentOrgID(r), chi.URLParam(r, "instanceID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}

func (s *Server) UpdateInstanceAutoCampaign(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	var req InstanceAutoCampaign
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceAutoCampaign(currentOrgID(r), chi.URLParam(r, "instanceID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}
