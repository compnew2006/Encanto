package api

import (
	"net/http"
)

func analyticsFiltersFromRequest(r *http.Request) AnalyticsFilters {
	return AnalyticsFilters{
		AgentID:    r.URL.Query().Get("agent_id"),
		InstanceID: r.URL.Query().Get("instance_id"),
	}
}

func (s *Server) GetAgentAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.store.AgentAnalyticsSummary(currentOrgID(r), analyticsFiltersFromRequest(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) GetAgentTransferTrends(w http.ResponseWriter, r *http.Request) {
	points, err := s.store.AgentTransferTrends(currentOrgID(r), analyticsFiltersFromRequest(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"points": points})
}

func (s *Server) GetAgentSourceBreakdown(w http.ResponseWriter, r *http.Request) {
	points, err := s.store.AgentSourceBreakdown(currentOrgID(r), analyticsFiltersFromRequest(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"points": points})
}

func (s *Server) GetAgentComparison(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.AgentComparison(currentOrgID(r), analyticsFiltersFromRequest(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}

func (s *Server) GetAgentRatings(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.AgentRatings(currentOrgID(r), analyticsFiltersFromRequest(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}

func (s *Server) ExportAgentAnalytics(w http.ResponseWriter, r *http.Request) {
	csvBody, err := s.store.ExportAgentAnalyticsCSV(currentOrgID(r), analyticsFiltersFromRequest(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvBody))
}
