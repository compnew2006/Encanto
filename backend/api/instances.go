package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UpdateInstanceNameRequest struct {
	Name string `json:"name"`
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

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.requireLicensedWrite(w, r, "instances") {
		return
	}

	var req CreateInstanceRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.CreateInstance(currentOrgID(r), claims.UserID, req)
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

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req UpdateInstanceNameRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceName(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"), req.Name)
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

	instance, err := s.store.ConnectInstance(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "instance_connected", map[string]any{"instance": instance})
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}

func (s *Server) DisconnectInstance(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	instance, err := s.store.DisconnectInstance(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"))
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

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	instance, err := s.store.RecoverInstance(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"))
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

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req InstanceSettingsRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceSettings(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"), req)
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

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CallPolicyRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceCallPolicy(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"), req)
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

	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req AutoCampaignRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	instance, err := s.store.UpdateInstanceAutoCampaign(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"instance": instance})
}
