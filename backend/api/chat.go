package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SendMessageRequest struct {
	Type     string `json:"type"`
	Body     string `json:"body"`
	FileName string `json:"file_name"`
	MediaURL string `json:"media_url"`
}

type AssignRequest struct {
	AssigneeID string `json:"assignee_id"`
}

type AddNoteRequest struct {
	Body string `json:"body"`
}

type AddCollaboratorRequest struct {
	UserID string `json:"user_id"`
}

type CreateStatusRequest struct {
	ContactID  string `json:"contact_id"`
	InstanceID string `json:"instance_id"`
	Body       string `json:"body"`
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	contactID := chi.URLParam(r, "contactID")
	snapshot, err := s.store.Workspace(
		currentOrgID(r),
		claims.UserID,
		contactID,
		r.URL.Query().Get("tab"),
		r.URL.Query().Get("search"),
		r.URL.Query().Get("instance_id"),
		r.URL.Query().Get("tag"),
	)
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req SendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	message, err := s.store.SendOutgoingMessage(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "new_message", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"message":    message,
	})
	writeJSON(w, http.StatusCreated, map[string]any{"message": message})
}

func (s *Server) RetryMessage(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	message, err := s.store.RetryMessage(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), chi.URLParam(r, "messageID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "status_update", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"message":    message,
	})
	writeJSON(w, http.StatusOK, map[string]any{"message": message})
}

func (s *Server) RevokeMessage(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	message, err := s.store.RevokeMessage(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), chi.URLParam(r, "messageID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "status_update", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"message":    message,
	})
	writeJSON(w, http.StatusOK, map[string]any{"message": message})
}

func (s *Server) AssignContact(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req AssignRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	contact, err := s.store.Assign(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req.AssigneeID)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"contact":    contact,
	})
	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) UnassignContact(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	contact, err := s.store.Unassign(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"contact":    contact,
	})
	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) TogglePin(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	contact, err := s.store.TogglePin(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) ToggleHide(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	contact, err := s.store.ToggleHide(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) CloseChat(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	contact, err := s.store.Close(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"contact":    contact,
	})
	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) ReopenChat(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	contact, err := s.store.Reopen(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"contact":    contact,
	})
	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *Server) AddNote(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req AddNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	note, err := s.store.AddNote(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req.Body)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id": chi.URLParam(r, "contactID"),
		"note":       note,
	})
	writeJSON(w, http.StatusCreated, map[string]any{"note": note})
}

func (s *Server) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req AddCollaboratorRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	collaborator, err := s.store.AddCollaborator(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req.UserID)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id":   chi.URLParam(r, "contactID"),
		"collaborator": collaborator,
	})
	writeJSON(w, http.StatusCreated, map[string]any{"collaborator": collaborator})
}

func (s *Server) ListNotifications(w http.ResponseWriter, r *http.Request) {
	notifications, err := s.store.ListNotifications(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"notifications": notifications})
}

func (s *Server) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	notifications, err := s.store.MarkAllNotificationsRead(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "notification_read", map[string]any{"notifications": notifications})
	writeJSON(w, http.StatusOK, map[string]any{"notifications": notifications})
}

func (s *Server) ListStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := s.store.ListStatuses(currentOrgID(r))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"statuses": statuses})
}

func (s *Server) CreateStatus(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CreateStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	status, err := s.store.AddStatus(currentOrgID(r), claims.UserID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "status_feed_update", map[string]any{"status": status})
	writeJSON(w, http.StatusCreated, map[string]any{"status": status})
}
