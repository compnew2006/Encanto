package api

import (
	"net/http"
)

func (s *Server) GetLicenseBootstrap(w http.ResponseWriter, r *http.Request) {
	bootstrap, err := s.store.LicenseBootstrap(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bootstrap)
}

func (s *Server) ActivateLicense(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req LicenseActivationRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	bootstrap, err := s.store.ActivateLicense(currentOrgID(r), claims.UserID, req.SecurityKey)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bootstrap)
}

func (s *Server) GetSettingsLimits(w http.ResponseWriter, r *http.Request) {
	bootstrap, err := s.store.LicenseBootstrap(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bootstrap)
}

func (s *Server) requireLicensedWrite(w http.ResponseWriter, r *http.Request, resource string) bool {
	bootstrap, err := s.store.LicenseBootstrap(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return false
	}

	if bootstrap.Status == "locked" {
		writeJSON(w, http.StatusLocked, map[string]any{
			"error":        bootstrap.Message,
			"activate_url": bootstrap.ActivateURL,
		})
		return false
	}

	for _, quota := range bootstrap.Quotas {
		if quota.Resource != resource {
			continue
		}
		if bootstrap.RestrictedCleanup || quota.Current >= quota.Limit {
			writeJSON(w, http.StatusPaymentRequired, map[string]any{
				"error":       bootstrap.Message,
				"resource":    quota.Resource,
				"current":     quota.Current,
				"limit":       quota.Limit,
				"over_quota":  quota.Current >= quota.Limit,
				"cleanup_url": bootstrap.CleanupURL,
			})
			return false
		}
	}

	if bootstrap.RestrictedCleanup {
		writeJSON(w, http.StatusLocked, map[string]any{
			"error":       bootstrap.Message,
			"cleanup_url": bootstrap.CleanupURL,
		})
		return false
	}

	return true
}
