package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Local request types (not shared externally)

type AssignRequest struct {
	AssigneeID string `json:"assignee_id"`
}

type AddNoteRequest struct {
	Body string `json:"body"`
}

type AddCollaboratorRequest struct {
	UserID string `json:"user_id"`
}

type TogglePinRequest struct {
	Pinned bool `json:"pinned"`
}

type ToggleHideRequest struct {
	Hidden bool `json:"hidden"`
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	snapshot, err := s.store.Workspace(currentOrgID(r), claims.UserID)
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	// If a specific conversation is requested, attach it
	contactID := chi.URLParam(r, "contactID")
	if contactID == "" {
		contactID = r.URL.Query().Get("contact_id")
	}
	if contactID != "" {
		detail, err := s.store.GetConversation(currentOrgID(r), claims.UserID, contactID)
		if err == nil {
			snapshot.Selected = &detail
		}
	}

	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) CreateDirectChat(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !s.requireLicensedWrite(w, r, "contacts") {
		return
	}

	var req CreateDirectChatRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	contact, err := s.store.CreateDirectChat(currentOrgID(r), claims.UserID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hub.Publish(currentOrgID(r), "conversation_event", map[string]any{
		"contact_id": contact.ID,
		"contact":    contact,
	})
	writeJSON(w, http.StatusCreated, map[string]any{"contact": contact})
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

	orgID := currentOrgID(r)
	contactID := chi.URLParam(r, "contactID")

	contact, err := s.store.getContactByID(contactID, orgID, claims.UserID)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "contact not found")
		return
	}

	message, err := s.store.SendOutgoingMessage(orgID, claims.UserID, contactID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Trigger async send via WhatsApp
	go s.WhatsApp.SendMessage(orgID, contact.InstanceID, contact.PhoneNumber, message.ID, req.Body)

	s.hub.Publish(orgID, "new_message", map[string]any{
		"contact_id": contactID,
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

	var req TogglePinRequest
	_ = decodeJSON(r, &req)

	contact, err := s.store.TogglePin(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req.Pinned)
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

	var req ToggleHideRequest
	_ = decodeJSON(r, &req)

	contact, err := s.store.ToggleHide(currentOrgID(r), claims.UserID, chi.URLParam(r, "contactID"), req.Hidden)
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
	if err := s.store.MarkAllNotificationsRead(currentOrgID(r)); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	notifications, _ := s.store.ListNotifications(currentOrgID(r))
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

	var req AddStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	status, err := s.store.AddStatus(currentOrgID(r), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = claims
	s.hub.Publish(currentOrgID(r), "status_feed_update", map[string]any{"status": status})
	writeJSON(w, http.StatusCreated, map[string]any{"status": status})
}
