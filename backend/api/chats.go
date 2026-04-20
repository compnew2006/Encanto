//go:build bootstrap

package api

import (
	"errors"
	"net/http"

	"encanto/audit"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func listChatsHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "chats.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Chat visibility is not allowed.", err.Error())
			return
		}

		items, err := deps.ChatService.ListChats(r.Context(), user, r.URL.Query().Get("search"))
		if err != nil {
			audit.WriteError(w, http.StatusInternalServerError, "chats_list_failed", "Unable to load chats.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"chats": items})
	}
}

func getChatHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "chats.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Chat visibility is not allowed.", err.Error())
			return
		}

		contactID, err := parseUUIDParam(chi.URLParam(r, "contactID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_chat", "Invalid chat identifier.")
			return
		}

		detail, err := deps.ChatService.GetChatDetail(r.Context(), user, contactID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeNotFound(w)
				return
			}
			audit.WriteError(w, http.StatusInternalServerError, "chat_detail_failed", "Unable to load the selected chat.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"chat": detail})
	}
}

func listMessagesHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "messages.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Message history is not allowed.", err.Error())
			return
		}

		contactID, err := parseUUIDParam(chi.URLParam(r, "contactID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_chat", "Invalid chat identifier.")
			return
		}

		detail, err := deps.ChatService.GetChatDetail(r.Context(), user, contactID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeNotFound(w)
				return
			}
			audit.WriteError(w, http.StatusInternalServerError, "messages_failed", "Unable to load the selected message history.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"messages": detail.Messages})
	}
}

func listNotesHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustUser(r)
		if err := requirePermission(deps.AccessService, user, "notes.view"); err != nil {
			audit.WriteErrorWithReason(w, http.StatusForbidden, "permission_denied", "Conversation notes are not allowed.", err.Error())
			return
		}

		contactID, err := parseUUIDParam(chi.URLParam(r, "contactID"))
		if err != nil {
			audit.WriteError(w, http.StatusBadRequest, "invalid_chat", "Invalid chat identifier.")
			return
		}

		detail, err := deps.ChatService.GetChatDetail(r.Context(), user, contactID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeNotFound(w)
				return
			}
			audit.WriteError(w, http.StatusInternalServerError, "notes_failed", "Unable to load conversation notes.")
			return
		}

		audit.WriteJSON(w, http.StatusOK, map[string]any{"notes": detail.Notes})
	}
}
