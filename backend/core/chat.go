package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"encanto/data"
	"encanto/data/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChatService struct {
	store  *data.Store
	access *AccessService
}

func NewChatService(store *data.Store, access *AccessService) *ChatService {
	return &ChatService{store: store, access: access}
}

func (s *ChatService) ListChats(ctx context.Context, user CurrentUserContext, search string) ([]ChatListItem, error) {
	params := sqlc.ListVisibleChatsParams{
		OrganizationID:      user.CurrentOrganization.ID,
		ViewerUserID:        user.ID,
		ScopeMode:           string(user.EffectiveAccess.Visibility.Mode),
		AllowedInstanceIds:  mustJSON(user.EffectiveAccess.Visibility.AllowedInstanceIDs),
		AllowedPhoneNumbers: mustJSON(user.EffectiveAccess.Visibility.AllowedPhoneNumbers),
		IncludePending:      s.access.HasPermission(user, "chats.unclaimed.view"),
		Search:              search,
	}

	rows, err := s.store.Queries.ListVisibleChats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list chats: %w", err)
	}

	items := make([]ChatListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapChatRow(row, user.EffectiveAccess.Visibility.CanViewUnmaskedPhone))
	}
	return items, nil
}

func (s *ChatService) GetChatDetail(ctx context.Context, user CurrentUserContext, contactID uuid.UUID) (ChatDetail, error) {
	params := sqlc.GetVisibleChatByIDParams{
		OrganizationID:      user.CurrentOrganization.ID,
		ViewerUserID:        user.ID,
		ContactID:           contactID,
		ScopeMode:           string(user.EffectiveAccess.Visibility.Mode),
		AllowedInstanceIds:  mustJSON(user.EffectiveAccess.Visibility.AllowedInstanceIDs),
		AllowedPhoneNumbers: mustJSON(user.EffectiveAccess.Visibility.AllowedPhoneNumbers),
	}

	row, err := s.store.Queries.GetVisibleChatByID(ctx, params)
	if err != nil {
		return ChatDetail{}, err
	}

	messageRows, err := s.store.Queries.ListMessagesByContact(ctx, sqlc.ListMessagesByContactParams{
		OrganizationID: user.CurrentOrganization.ID,
		ContactID:      contactID,
	})
	if err != nil {
		return ChatDetail{}, fmt.Errorf("list messages: %w", err)
	}

	noteRows, err := s.store.Queries.ListNotesByContact(ctx, sqlc.ListNotesByContactParams{
		OrganizationID: user.CurrentOrganization.ID,
		ContactID:      contactID,
	})
	if err != nil {
		return ChatDetail{}, fmt.Errorf("list notes: %w", err)
	}

	messages := make([]ConversationMessage, 0, len(messageRows))
	for _, message := range messageRows {
		messages = append(messages, ConversationMessage{
			ID:           message.ID,
			ContactID:    message.ContactID,
			Direction:    message.Direction,
			Type:         message.Type,
			Body:         message.Body,
			Status:       message.Status,
			CreatedAt:    message.CreatedAt.Time,
			SentByUserID: optionalUUID(message.SentByUserID),
		})
	}

	notes := make([]ConversationNote, 0, len(noteRows))
	for _, note := range noteRows {
		notes = append(notes, ConversationNote{
			ID:           note.ID,
			ContactID:    note.ContactID,
			AuthorUserID: note.AuthorUserID,
			AuthorName:   note.AuthorName,
			Body:         note.Body,
			CreatedAt:    note.CreatedAt.Time,
		})
	}

	chat := mapChatByIDRow(row, user.EffectiveAccess.Visibility.CanViewUnmaskedPhone)
	return ChatDetail{
		Chat:     chat,
		Messages: messages,
		Notes:    notes,
		Composer: s.composerState(user, chat.Status),
	}, nil
}

func (s *ChatService) composerState(user CurrentUserContext, status string) ComposerState {
	if !s.access.HasPermission(user, "messages.send") {
		return ComposerState{Allowed: false, Disabled: true, DenialReason: s.access.PermissionDeniedReason("messages.send")}
	}
	if status == "pending" && !s.access.HasPermission(user, "chats.unclaimed.send") {
		return ComposerState{Allowed: false, Disabled: true, DenialReason: s.access.PermissionDeniedReason("chats.unclaimed.send")}
	}
	return ComposerState{Allowed: true, Disabled: false}
}

func mapChatRow(row sqlc.ListVisibleChatsRow, canViewUnmaskedPhone bool) ChatListItem {
	return ChatListItem{
		ID:                 row.ID,
		Name:               coalesceText(row.Name, row.PhoneNumber),
		PhoneNumber:        row.PhoneNumber,
		VisiblePhone:       maybeMaskPhone(row.PhoneNumber, canViewUnmaskedPhone),
		Status:             row.Status,
		LastMessagePreview: coalesceText(row.LastMessagePreview, ""),
		LastMessageAt:      optionalTime(row.LastMessageAt),
		InstanceName:       coalesceText(row.InstanceName, "No instance"),
		IsHidden:           row.IsHidden,
		IsPinned:           row.IsPinned,
	}
}

func mapChatByIDRow(row sqlc.GetVisibleChatByIDRow, canViewUnmaskedPhone bool) ChatListItem {
	return ChatListItem{
		ID:                 row.ID,
		Name:               coalesceText(row.Name, row.PhoneNumber),
		PhoneNumber:        row.PhoneNumber,
		VisiblePhone:       maybeMaskPhone(row.PhoneNumber, canViewUnmaskedPhone),
		Status:             row.Status,
		LastMessagePreview: coalesceText(row.LastMessagePreview, ""),
		LastMessageAt:      optionalTime(row.LastMessageAt),
		InstanceName:       coalesceText(row.InstanceName, "No instance"),
		IsHidden:           row.IsHidden,
		IsPinned:           row.IsPinned,
	}
}

func maybeMaskPhone(phone string, canViewUnmaskedPhone bool) string {
	if canViewUnmaskedPhone || len(phone) <= 4 {
		return phone
	}
	return phone[:2] + "****" + phone[len(phone)-2:]
}

func coalesceText(value pgtype.Text, fallback string) string {
	if !value.Valid || value.String == "" {
		return fallback
	}
	return value.String
}

func mustJSON(value any) []byte {
	switch typed := value.(type) {
	case []uuid.UUID:
		payload := make([]string, 0, len(typed))
		for _, item := range typed {
			payload = append(payload, item.String())
		}
		encoded, _ := json.Marshal(payload)
		return encoded
	case []string:
		encoded, _ := json.Marshal(typed)
		return encoded
	default:
		return []byte("[]")
	}
}

func optionalUUID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	identifier := uuid.UUID(value.Bytes)
	return &identifier
}

func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
