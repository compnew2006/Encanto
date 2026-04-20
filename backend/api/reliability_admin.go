package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) ListJobs(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}
	jobs, err := s.store.ListJobs(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

func (s *Server) GetJob(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}
	job, err := s.store.JobByID(currentOrgID(r), chi.URLParam(r, "jobID"))
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}
	webhooks, err := s.store.ListWebhooks(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"webhooks": webhooks})
}

func (s *Server) ListWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}
	deliveries, err := s.store.ListWebhookDeliveries(currentOrgID(r), chi.URLParam(r, "webhookID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deliveries": deliveries})
}

func (s *Server) RetryWebhookDelivery(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	delivery, err := s.store.RetryWebhookDelivery(currentOrgID(r), claims.UserID, chi.URLParam(r, "webhookID"), chi.URLParam(r, "deliveryID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, delivery)
}

func (s *Server) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}
	entries, err := s.store.ListAuditLogs(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

func (s *Server) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.isAdmin(currentOrgID(r)) {
		errorJSON(w, http.StatusForbidden, "admin privileges required")
		return
	}

	if err := s.store.DeleteInstance(currentOrgID(r), claims.UserID, chi.URLParam(r, "instanceID")); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
