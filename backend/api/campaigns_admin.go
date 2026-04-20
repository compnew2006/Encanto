package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := s.store.ListCampaigns(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"campaigns": campaigns})
}

func (s *Server) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.requireLicensedWrite(w, r, "campaigns") {
		return
	}

	var req CampaignUpsertRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	campaign, err := s.store.CreateCampaign(currentOrgID(r), claims.UserID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"campaign": campaign})
}

func (s *Server) GetCampaign(w http.ResponseWriter, r *http.Request) {
	record, err := s.store.CampaignDetail(currentOrgID(r), chi.URLParam(r, "campaignID"))
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CampaignUpsertRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	campaign, err := s.store.UpdateCampaign(currentOrgID(r), claims.UserID, chi.URLParam(r, "campaignID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"campaign": campaign})
}

func (s *Server) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := s.store.DeleteCampaign(currentOrgID(r), claims.UserID, chi.URLParam(r, "campaignID")); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) LaunchCampaign(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	run, recipients, err := s.store.LaunchCampaign(currentOrgID(r), claims.UserID, chi.URLParam(r, "campaignID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"run": run, "recipients": recipients})
}

func (s *Server) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	campaign, err := s.store.PauseCampaign(currentOrgID(r), claims.UserID, chi.URLParam(r, "campaignID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"campaign": campaign})
}

func (s *Server) ResumeCampaign(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	campaign, err := s.store.ResumeCampaign(currentOrgID(r), claims.UserID, chi.URLParam(r, "campaignID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"campaign": campaign})
}

func (s *Server) ListCampaignRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ListCampaignRuns(currentOrgID(r), chi.URLParam(r, "campaignID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (s *Server) ListCampaignRecipients(w http.ResponseWriter, r *http.Request) {
	recipients, err := s.store.ListCampaignRecipients(currentOrgID(r), chi.URLParam(r, "campaignID"), r.URL.Query().Get("run_id"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recipients": recipients})
}
