package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) ListContacts(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	view, err := s.store.ListAdminContacts(
		currentOrgID(r),
		claims.UserID,
		r.URL.Query().Get("search"),
		r.URL.Query().Get("instance_id"),
	)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, view)
}

func (s *Server) CreateContact(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.requireLicensedWrite(w, r, "contacts") {
		return
	}

	var req ContactMutationRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	contact, err := s.store.CreateContact(currentOrgID(r), claims.UserID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"contact": contact})
}

func (s *Server) UpdateContact(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req ContactMutationRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	contact, err := s.store.UpdateContact(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) DeleteContact(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := s.store.DeleteContact(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID")); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) ExportContacts(w http.ResponseWriter, r *http.Request) {
	columns := []string{}
	if raw := strings.TrimSpace(r.URL.Query().Get("columns")); raw != "" {
		for _, column := range strings.Split(raw, ",") {
			column = strings.TrimSpace(column)
			if column != "" {
				columns = append(columns, column)
			}
		}
	}

	var (
		csvBody string
		err     error
	)
	if r.URL.Query().Get("sample") == "1" {
		csvBody, err = s.store.SampleContactsCSV(currentOrgID(r), columns)
	} else {
		csvBody, err = s.store.ExportContactsCSV(currentOrgID(r), columns)
	}
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvBody))
}

func (s *Server) ImportContacts(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.requireLicensedWrite(w, r, "contacts") {
		return
	}

	var req ContactImportRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	result, err := s.store.ImportContactsCSV(currentOrgID(r), claims.UserID, req.CSV, req.UpdateOnDuplicate)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) ListClosedChats(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	result, err := s.store.ListClosedChats(
		currentOrgID(r),
		page,
		pageSize,
		r.URL.Query().Get("agent_id"),
		r.URL.Query().Get("instance_id"),
	)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
