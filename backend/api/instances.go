package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
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
	orgID := currentOrgID(r)
	if !s.isAdmin(orgID) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	instanceID := chi.URLParam(r, "instanceID")
	inst, err := s.store.getInstanceByID(instanceID)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "instance not found")
		return
	}

	// 1. Get or create deviceStore
	var device *store.Device
	if inst.JID != "" {
		jid, err := types.ParseJID(inst.JID)
		if err == nil {
			device, _ = s.store.waContainer.GetDevice(context.Background(), jid)
		}
	}
	if device == nil {
		device = s.store.waContainer.NewDevice()
	}

	// 2. Start the instance via manager
	err = s.WhatsApp.StartInstance(orgID, instanceID, device)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Mark as connecting in DB
	updated, err := s.store.ConnectInstance(orgID, claims.UserID, instanceID)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"instance": updated})
}

func (s *Server) DisconnectInstance(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	orgID := currentOrgID(r)
	if !s.isAdmin(orgID) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	instanceID := chi.URLParam(r, "instanceID")

	// 1. Stop via manager
	s.WhatsApp.StopInstance(instanceID)

	// 2. Mark as disconnected in DB
	instance, err := s.store.DisconnectInstance(orgID, claims.UserID, instanceID)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

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

	var req InstanceSettings
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

	var req InstanceCallPolicy
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

	var req InstanceAutoCampaign
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
