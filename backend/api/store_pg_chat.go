package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ---------- WORKSPACE ----------

func (s *PGStore) Workspace(orgID, userID string) (WorkspaceView, error) {
	contacts, err := s.listContactsForUser(orgID, userID)
	if err != nil {
		return WorkspaceView{}, err
	}

	users, _ := s.listUsersForOrg(orgID)
	instances, _ := s.ListInstances(orgID)
	statuses, _ := s.ListStatuses(orgID)
	notifs, _ := s.ListNotifications(orgID)
	quickReplies, _ := s.ListQuickReplies(orgID)
	settings, _ := s.SettingsForOrg(orgID)

	// Calculate tab counts
	tabCounts := map[string]int{
		"assigned": 0,
		"pending":  0,
		"closed":   0,
	}
	for _, c := range contacts {
		if c.Status == "assigned" {
			tabCounts["assigned"]++
		} else if c.Status == "pending" {
			tabCounts["pending"]++
		}
	}
	// Query closed count separately
	var closedCount int
	_ = s.db.QueryRow(s.ctx(), "SELECT COUNT(*) FROM contacts WHERE organization_id = $1 AND status = 'closed'", orgID).Scan(&closedCount)
	tabCounts["closed"] = closedCount

	return WorkspaceView{
		CurrentTab:    "assigned",
		TabCounts:     tabCounts,
		Filters:       make(map[string]string),
		Conversations: contacts,
		Users:         users,
		Instances:     instances,
		Statuses:      statuses,
		Notifications: notifs,
		QuickReplies:  quickReplies,
		Settings:      settings,
	}, nil
}

func (s *PGStore) listUsersForOrg(orgID string) ([]WorkspaceUser, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, name, role, email, status, avatar FROM users
		WHERE organization_id = $1 AND is_active = true ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []WorkspaceUser
	for rows.Next() {
		var u WorkspaceUser
		_ = rows.Scan(&u.ID, &u.Name, &u.Role, &u.Email, &u.Status, &u.Avatar)
		users = append(users, u)
	}
	return users, nil
}

func (s *PGStore) listContactsForUser(orgID, userID string) ([]ChatContact, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT c.id, c.name, c.phone_number, c.avatar, c.status,
			c.assigned_user_id, c.assigned_user_name, c.instance_id, c.instance_name,
			c.instance_source_label, c.last_message_preview, c.last_message_at,
			c.last_inbound_at, c.closed_at, c.is_public, c.is_read, c.unread_count,
			c.tags, c.metadata, c.organization_id, c.created_at,
			COALESCE(cus.is_pinned, false), COALESCE(cus.is_hidden, false),
			COALESCE(cus.last_read_message_id, '')
		FROM contacts c
		LEFT JOIN contact_user_states cus ON cus.contact_id = c.id AND cus.user_id = $2
		WHERE c.organization_id = $1 AND c.status != 'closed'
		ORDER BY c.last_message_at DESC`, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContacts(rows)
}

func scanContacts(rows pgx.Rows) ([]ChatContact, error) {
	var contacts []ChatContact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	if contacts == nil {
		contacts = []ChatContact{}
	}
	return contacts, nil
}

func scanContact(row pgx.Rows) (ChatContact, error) {
	var c ChatContact
	var assignedUserID *string
	var closedAt *time.Time
	var tagsB, metaB []byte
	err := row.Scan(
		&c.ID, &c.Name, &c.PhoneNumber, &c.Avatar, &c.Status,
		&assignedUserID, &c.AssignedUserName, &c.InstanceID, &c.InstanceName,
		&c.InstanceSourceLabel, &c.LastMessagePreview, &c.LastMessageAt,
		&c.LastInboundAt, &closedAt, &c.IsPublic, &c.IsRead, &c.UnreadCount,
		&tagsB, &metaB, &c.OrganizationID, &c.CreatedAt,
		&c.IsPinned, &c.IsHidden, &c.LastReadMessageID,
	)
	if err != nil {
		return c, err
	}
	if assignedUserID != nil {
		c.AssignedUserID = *assignedUserID
	}
	if closedAt != nil {
		c.ClosedAt = closedAt
	}
	_ = json.Unmarshal(tagsB, &c.Tags)
	_ = json.Unmarshal(metaB, &c.Metadata)
	if c.Tags == nil {
		c.Tags = []string{}
	}
	if c.Metadata == nil {
		c.Metadata = map[string]string{}
	}
	c.PhoneDisplay = displayPhoneNumber(c.PhoneNumber)
	if c.Name == "" || c.Name == c.PhoneNumber {
		if c.Metadata["type"] == "group" {
			c.Name = "Group " + c.PhoneDisplay
		} else {
			c.Name = c.PhoneDisplay
		}
	}
	return c, nil
}

// ---------- CONVERSATION DETAIL ----------

func (s *PGStore) GetConversation(orgID, userID, contactID string) (ConversationDetail, error) {
	// Load contact
	var c ChatContact
	var assignedUserID *string
	var closedAt *time.Time
	var tagsB, metaB []byte
	err := s.db.QueryRow(s.ctx(), `
		SELECT c.id, c.name, c.phone_number, c.avatar, c.status,
			c.assigned_user_id, c.assigned_user_name, c.instance_id, c.instance_name,
			c.instance_source_label, c.last_message_preview, c.last_message_at,
			c.last_inbound_at, c.closed_at, c.is_public, c.is_read, c.unread_count,
			c.tags, c.metadata, c.organization_id, c.created_at,
			COALESCE(cus.is_pinned, false), COALESCE(cus.is_hidden, false),
			COALESCE(cus.last_read_message_id, '')
		FROM contacts c
		LEFT JOIN contact_user_states cus ON cus.contact_id = c.id AND cus.user_id = $3
		WHERE c.id = $1 AND c.organization_id = $2`, contactID, orgID, userID).
		Scan(&c.ID, &c.Name, &c.PhoneNumber, &c.Avatar, &c.Status,
			&assignedUserID, &c.AssignedUserName, &c.InstanceID, &c.InstanceName,
			&c.InstanceSourceLabel, &c.LastMessagePreview, &c.LastMessageAt,
			&c.LastInboundAt, &closedAt, &c.IsPublic, &c.IsRead, &c.UnreadCount,
			&tagsB, &metaB, &c.OrganizationID, &c.CreatedAt,
			&c.IsPinned, &c.IsHidden, &c.LastReadMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ConversationDetail{}, errors.New("contact not found")
		}
		return ConversationDetail{}, err
	}
	if assignedUserID != nil {
		c.AssignedUserID = *assignedUserID
	}
	if closedAt != nil {
		c.ClosedAt = closedAt
	}
	_ = json.Unmarshal(tagsB, &c.Tags)
	_ = json.Unmarshal(metaB, &c.Metadata)
	if c.Tags == nil {
		c.Tags = []string{}
	}

	messages, _ := s.listMessages(contactID)
	notes, _ := s.listNotes(contactID)
	collabs, _ := s.listCollaborators(contactID)
	events, _ := s.listEvents(contactID)

	return ConversationDetail{
		Contact:       c,
		Messages:      messages,
		Notes:         notes,
		Collaborators: collabs,
		Events:        events,
	}, nil
}

func (s *PGStore) listMessages(contactID string) ([]ChatMessage, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, direction, type, body, status, file_name, file_size_label, media_url,
			failure_reason, retry_count, typed_for_ms, reaction, revoked_at,
			can_retry, can_revoke, created_at
		FROM messages WHERE contact_id = $1 ORDER BY created_at ASC`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		_ = rows.Scan(&m.ID, &m.Direction, &m.Type, &m.Body, &m.Status,
			&m.FileName, &m.FileSizeLabel, &m.MediaURL, &m.FailureReason,
			&m.RetryCount, &m.TypedForMs, &m.Reaction, &m.RevokedAt,
			&m.CanRetry, &m.CanRevoke, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []ChatMessage{}
	}
	return msgs, nil
}

func (s *PGStore) listNotes(contactID string) ([]ConversationNote, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, user_id, user_name, body, created_at FROM conversation_notes
		WHERE contact_id = $1 ORDER BY created_at ASC`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []ConversationNote
	for rows.Next() {
		var n ConversationNote
		_ = rows.Scan(&n.ID, &n.UserID, &n.UserName, &n.Body, &n.CreatedAt)
		notes = append(notes, n)
	}
	if notes == nil {
		notes = []ConversationNote{}
	}
	return notes, nil
}

func (s *PGStore) listCollaborators(contactID string) ([]Collaborator, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, user_id, user_name, status, invited_at
		FROM contact_collaborators WHERE contact_id = $1 ORDER BY invited_at`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var collabs []Collaborator
	for rows.Next() {
		var c Collaborator
		_ = rows.Scan(&c.ID, &c.UserID, &c.UserName, &c.Status, &c.InvitedAt)
		collabs = append(collabs, c)
	}
	if collabs == nil {
		collabs = []Collaborator{}
	}
	return collabs, nil
}

func (s *PGStore) listEvents(contactID string) ([]TimelineEvent, error) {
	rows, err := s.db.Query(s.ctx(), `
		SELECT id, event_type, actor_user_id, actor_name, summary, metadata, occurred_at
		FROM timeline_events WHERE contact_id = $1 ORDER BY occurred_at ASC`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []TimelineEvent
	for rows.Next() {
		var e TimelineEvent
		var metaB []byte
		_ = rows.Scan(&e.ID, &e.EventType, &e.ActorUserID, &e.ActorName, &e.Summary, &metaB, &e.OccurredAt)
		_ = json.Unmarshal(metaB, &e.Metadata)
		if e.Metadata == nil {
			e.Metadata = map[string]string{}
		}
		events = append(events, e)
	}
	if events == nil {
		events = []TimelineEvent{}
	}
	return events, nil
}

func (s *PGStore) addEvent(contactID, orgID, eventType, actorUserID, actorName, summary string, metadata map[string]string) {
	b, _ := json.Marshal(metadata)
	if b == nil {
		b = []byte("{}")
	}
	_, _ = s.db.Exec(s.ctx(), `
		INSERT INTO timeline_events (contact_id, organization_id, event_type, actor_user_id, actor_name, summary, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		contactID, orgID, eventType, actorUserID, actorName, summary, b)
}

// ---------- MESSAGES ----------

func (s *PGStore) SendOutgoingMessage(orgID, userID, contactID string, req SendMessageRequest) (ChatMessage, error) {
	var msg ChatMessage
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO messages (contact_id, organization_id, direction, type, body, status, file_name, file_size_label, media_url, can_revoke)
		VALUES ($1, $2, 'outbound', $3, $4, 'queued', $5, $6, $7, true)
		RETURNING id, direction, type, body, status, file_name, file_size_label, media_url, failure_reason, retry_count, typed_for_ms, reaction, revoked_at, can_retry, can_revoke, created_at`,
		contactID, orgID, req.Type, req.Body, req.FileName, req.FileSizeLabel, req.MediaURL).
		Scan(&msg.ID, &msg.Direction, &msg.Type, &msg.Body, &msg.Status,
			&msg.FileName, &msg.FileSizeLabel, &msg.MediaURL, &msg.FailureReason,
			&msg.RetryCount, &msg.TypedForMs, &msg.Reaction, &msg.RevokedAt,
			&msg.CanRetry, &msg.CanRevoke, &msg.CreatedAt)
	if err != nil {
		return msg, err
	}
	// update contact preview
	_, _ = s.db.Exec(s.ctx(), `
		UPDATE contacts SET last_message_preview = $1, last_message_at = NOW(), is_read = true, unread_count = 0, updated_at = NOW()
		WHERE id = $2`, req.Body, contactID)
	return msg, nil
}

func (s *PGStore) HandleInboundMessage(orgID, instanceID, chatJID, senderJID, body, msgType string) (ChatMessage, error) {
	// Route group traffic into the group conversation instead of opening one conversation per participant.
	conversationID := senderJID
	if isGroupJID(chatJID) {
		conversationID = chatJID
	}

	phone := normalizePhoneNumber(conversationID)
	senderPhone := normalizePhoneNumber(senderJID)

	// Manual fallback resolution for known LID mismatch
	if phone == "149641526026409" {
		phone = "966561853319"
	}
	if senderPhone == "149641526026409" {
		senderPhone = "966561853319"
	}

	contactName := displayPhoneNumber(phone)
	if isGroupJID(chatJID) {
		contactName = "Group " + displayPhoneNumber(phone)
	}

	previewBody := body
	if isGroupJID(chatJID) && senderPhone != "" {
		previewBody = displayPhoneNumber(senderPhone) + ": " + body
	}

	// 2. Find or create contact
	var contactID string
	err := s.db.QueryRow(s.ctx(), `
		SELECT id FROM contacts 
		WHERE organization_id = $1 AND instance_id = $2 AND phone_number = $3`,
		orgID, instanceID, phone).Scan(&contactID)

	if err != nil {
		// Create new contact
		var instName, instSource string
		_ = s.db.QueryRow(s.ctx(), `SELECT name, settings->>'source_tag_label' FROM whatsapp_instances WHERE id = $1`, instanceID).Scan(&instName, &instSource)
		metadata := map[string]string{}
		tags := []string{}
		if isGroupJID(chatJID) {
			metadata["type"] = "group"
			metadata["chat_jid"] = chatJID
			tags = append(tags, "group")
		}
		if senderPhone != "" {
			metadata["last_sender_phone"] = displayPhoneNumber(senderPhone)
		}
		metaJSON, _ := json.Marshal(metadata)
		tagsJSON, _ := json.Marshal(tags)

		err = s.db.QueryRow(s.ctx(), `
			INSERT INTO contacts (organization_id, instance_id, name, phone_number, avatar, status, instance_name, instance_source_label, last_message_preview, last_message_at, last_inbound_at, is_public, is_read, unread_count, tags, metadata)
			VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8, NOW(), NOW(), true, false, 1, $9, $10)
			RETURNING id`,
			orgID, instanceID, contactName, phone, "https://i.pravatar.cc/150?u="+phone, instName, instSource, previewBody, tagsJSON, metaJSON).Scan(&contactID)
		if err != nil {
			return ChatMessage{}, err
		}
	} else {
		// Update existing contact
		metadata := map[string]string{}
		if isGroupJID(chatJID) {
			metadata["type"] = "group"
			metadata["chat_jid"] = chatJID
		}
		if senderPhone != "" {
			metadata["last_sender_phone"] = displayPhoneNumber(senderPhone)
		}
		metaJSON, _ := json.Marshal(metadata)
		_, _ = s.db.Exec(s.ctx(), `
			UPDATE contacts 
			SET name = CASE
					WHEN metadata->>'type' = 'group' THEN name
					WHEN name = phone_number OR name = '' THEN $1
					ELSE name
				END,
				last_message_preview = $2,
				last_message_at = NOW(),
				last_inbound_at = NOW(),
				is_read = false,
				unread_count = unread_count + 1,
				tags = CASE
					WHEN $3::jsonb = '[]'::jsonb THEN tags
					ELSE (
						SELECT jsonb_agg(DISTINCT value)
						FROM jsonb_array_elements_text(COALESCE(tags, '[]'::jsonb) || $3::jsonb) AS value
					)
				END,
				metadata = COALESCE(metadata, '{}'::jsonb) || $4::jsonb,
				updated_at = NOW()
			WHERE id = $5`,
			contactName,
			previewBody,
			func() []byte {
				if isGroupJID(chatJID) {
					b, _ := json.Marshal([]string{"group"})
					return b
				}
				return []byte("[]")
			}(),
			metaJSON,
			contactID)
	}

	// 3. Insert message
	var msg ChatMessage
	err = s.db.QueryRow(s.ctx(), `
		INSERT INTO messages (contact_id, organization_id, direction, type, body, status, can_revoke)
		VALUES ($1, $2, 'inbound', $3, $4, 'received', false)
		RETURNING id, direction, type, body, status, file_name, file_size_label, media_url, failure_reason, retry_count, typed_for_ms, reaction, revoked_at, can_retry, can_revoke, created_at`,
		contactID, orgID, msgType, previewBody).
		Scan(&msg.ID, &msg.Direction, &msg.Type, &msg.Body, &msg.Status,
			&msg.FileName, &msg.FileSizeLabel, &msg.MediaURL, &msg.FailureReason,
			&msg.RetryCount, &msg.TypedForMs, &msg.Reaction, &msg.RevokedAt,
			&msg.CanRetry, &msg.CanRevoke, &msg.CreatedAt)

	return msg, err
}

func (s *PGStore) UpdateMessageStatus(orgID, messageID, status, reason string) (ChatMessage, error) {
	var msg ChatMessage
	err := s.db.QueryRow(s.ctx(), `
		UPDATE messages 
		SET status = $1, failure_reason = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4
		RETURNING id, contact_id, direction, type, body, status, file_name, file_size_label, media_url, failure_reason, retry_count, typed_for_ms, reaction, revoked_at, can_retry, can_revoke, created_at`,
		status, reason, messageID, orgID).
		Scan(&msg.ID, &msg.ContactID, &msg.Direction, &msg.Type, &msg.Body, &msg.Status,
			&msg.FileName, &msg.FileSizeLabel, &msg.MediaURL, &msg.FailureReason,
			&msg.RetryCount, &msg.TypedForMs, &msg.Reaction, &msg.RevokedAt,
			&msg.CanRetry, &msg.CanRevoke, &msg.CreatedAt)
	return msg, err
}

func (s *PGStore) CreateDirectChat(orgID, userID string, req CreateDirectChatRequest) (ChatContact, error) {
	// check if contact already exists with this phone + instance
	phone := normalizePhoneNumber(req.PhoneNumber)
	if phone == "" {
		return ChatContact{}, errors.New("phone number is required")
	}

	// find instance
	var instID, instName, instSourceLabel string
	err := s.db.QueryRow(s.ctx(), `SELECT id, name, settings->>'source_tag_label' FROM whatsapp_instances WHERE organization_id = $1 AND id = $2`, orgID, req.InstanceID).
		Scan(&instID, &instName, &instSourceLabel)
	if err != nil {
		// fallback: pick first
		err = s.db.QueryRow(s.ctx(), `SELECT id, name, settings->>'source_tag_label' FROM whatsapp_instances WHERE organization_id = $1 LIMIT 1`, orgID).
			Scan(&instID, &instName, &instSourceLabel)
		if err != nil {
			return ChatContact{}, errors.New("no WhatsApp accounts are available")
		}
	}

	// check duplicate
	var existingID string
	if err := s.db.QueryRow(s.ctx(), `SELECT id FROM contacts WHERE organization_id = $1 AND instance_id = $2 AND phone_number = $3`, orgID, instID, phone).Scan(&existingID); err == nil {
		// return existing
		return s.getContactByID(existingID, orgID, userID)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = phone
	}

	var contactID string
	err = s.db.QueryRow(s.ctx(), `
		INSERT INTO contacts (organization_id, instance_id, name, phone_number,
			avatar, status, instance_name, instance_source_label,
			last_message_preview, last_message_at, last_inbound_at, is_public, is_read)
		VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,'Contact created.',$8,$8,true,true)
		RETURNING id`,
		orgID, instID, name, phone, "https://i.pravatar.cc/150?u="+phone,
		instName, instSourceLabel, time.Now()).Scan(&contactID)
	if err != nil {
		return ChatContact{}, err
	}
	s.addEvent(contactID, orgID, "contact_created", userID, "", "Created a direct chat", nil)
	return s.getContactByID(contactID, orgID, userID)
}

func (s *PGStore) getContactByID(contactID, orgID, userID string) (ChatContact, error) {
	var c ChatContact
	var assignedUserID *string
	var closedAt *time.Time
	var tagsB, metaB []byte
	err := s.db.QueryRow(s.ctx(), `
		SELECT c.id, c.name, c.phone_number, c.avatar, c.status,
			c.assigned_user_id, c.assigned_user_name, c.instance_id, c.instance_name,
			c.instance_source_label, c.last_message_preview, c.last_message_at,
			c.last_inbound_at, c.closed_at, c.is_public, c.is_read, c.unread_count,
			c.tags, c.metadata, c.organization_id, c.created_at,
			COALESCE(cus.is_pinned, false), COALESCE(cus.is_hidden, false),
			COALESCE(cus.last_read_message_id, '')
		FROM contacts c
		LEFT JOIN contact_user_states cus ON cus.contact_id = c.id AND cus.user_id = $3
		WHERE c.id = $1 AND c.organization_id = $2`, contactID, orgID, userID).
		Scan(&c.ID, &c.Name, &c.PhoneNumber, &c.Avatar, &c.Status,
			&assignedUserID, &c.AssignedUserName, &c.InstanceID, &c.InstanceName,
			&c.InstanceSourceLabel, &c.LastMessagePreview, &c.LastMessageAt,
			&c.LastInboundAt, &closedAt, &c.IsPublic, &c.IsRead, &c.UnreadCount,
			&tagsB, &metaB, &c.OrganizationID, &c.CreatedAt,
			&c.IsPinned, &c.IsHidden, &c.LastReadMessageID)
	if err != nil {
		return c, errors.New("contact not found")
	}
	if assignedUserID != nil {
		c.AssignedUserID = *assignedUserID
	}
	if closedAt != nil {
		c.ClosedAt = closedAt
	}
	_ = json.Unmarshal(tagsB, &c.Tags)
	_ = json.Unmarshal(metaB, &c.Metadata)
	if c.Tags == nil {
		c.Tags = []string{}
	}
	return c, nil
}

func (s *PGStore) RetryMessage(orgID, userID, contactID, messageID string) (ChatMessage, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE messages SET status = 'sent', failure_reason = '', retry_count = retry_count + 1, can_retry = false WHERE id = $1 AND contact_id = $2`,
		messageID, contactID)
	if err != nil {
		return ChatMessage{}, err
	}
	var m ChatMessage
	_ = s.db.QueryRow(s.ctx(), `SELECT id, direction, type, body, status, file_name, file_size_label, media_url, failure_reason, retry_count, typed_for_ms, reaction, revoked_at, can_retry, can_revoke, created_at FROM messages WHERE id = $1`, messageID).
		Scan(&m.ID, &m.Direction, &m.Type, &m.Body, &m.Status, &m.FileName, &m.FileSizeLabel, &m.MediaURL, &m.FailureReason, &m.RetryCount, &m.TypedForMs, &m.Reaction, &m.RevokedAt, &m.CanRetry, &m.CanRevoke, &m.CreatedAt)
	return m, nil
}

func (s *PGStore) RevokeMessage(orgID, userID, contactID, messageID string) (ChatMessage, error) {
	now := time.Now()
	_, err := s.db.Exec(s.ctx(), `UPDATE messages SET revoked_at = $1, can_revoke = false, body = '' WHERE id = $2 AND contact_id = $3`, now, messageID, contactID)
	if err != nil {
		return ChatMessage{}, err
	}
	var m ChatMessage
	_ = s.db.QueryRow(s.ctx(), `SELECT id, direction, type, body, status, file_name, file_size_label, media_url, failure_reason, retry_count, typed_for_ms, reaction, revoked_at, can_retry, can_revoke, created_at FROM messages WHERE id = $1`, messageID).
		Scan(&m.ID, &m.Direction, &m.Type, &m.Body, &m.Status, &m.FileName, &m.FileSizeLabel, &m.MediaURL, &m.FailureReason, &m.RetryCount, &m.TypedForMs, &m.Reaction, &m.RevokedAt, &m.CanRetry, &m.CanRevoke, &m.CreatedAt)
	return m, nil
}

// ---------- CONVERSATION ACTIONS ----------

func (s *PGStore) AddNote(orgID, userID, contactID, body string) (ConversationNote, error) {
	var userName string
	_ = s.db.QueryRow(s.ctx(), `SELECT name FROM users WHERE id = $1`, userID).Scan(&userName)
	var n ConversationNote
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO conversation_notes (contact_id, user_id, user_name, body)
		VALUES ($1,$2,$3,$4) RETURNING id, user_id, user_name, body, created_at`,
		contactID, userID, userName, body).
		Scan(&n.ID, &n.UserID, &n.UserName, &n.Body, &n.CreatedAt)
	return n, err
}

func (s *PGStore) AddCollaborator(orgID, actorID, contactID, targetUserID string) (Collaborator, error) {
	var userName string
	_ = s.db.QueryRow(s.ctx(), `SELECT name FROM users WHERE id = $1`, targetUserID).Scan(&userName)
	var c Collaborator
	err := s.db.QueryRow(s.ctx(), `
		INSERT INTO contact_collaborators (contact_id, user_id, user_name, status)
		VALUES ($1,$2,$3,'invited')
		ON CONFLICT DO NOTHING
		RETURNING id, user_id, user_name, status, invited_at`,
		contactID, targetUserID, userName).
		Scan(&c.ID, &c.UserID, &c.UserName, &c.Status, &c.InvitedAt)
	return c, err
}

func (s *PGStore) Assign(orgID, actorID, contactID, targetUserID string) (ChatContact, error) {
	var userName string
	_ = s.db.QueryRow(s.ctx(), `SELECT name FROM users WHERE id = $1`, targetUserID).Scan(&userName)
	_, err := s.db.Exec(s.ctx(), `
		UPDATE contacts SET assigned_user_id = $1, assigned_user_name = $2, status = 'assigned', updated_at = NOW()
		WHERE id = $3 AND organization_id = $4`, targetUserID, userName, contactID, orgID)
	if err != nil {
		return ChatContact{}, err
	}
	s.addEvent(contactID, orgID, "assigned", actorID, "", fmt.Sprintf("Assigned to %s", userName), map[string]string{"assigned_to": targetUserID})
	return s.getContactByID(contactID, orgID, actorID)
}

func (s *PGStore) Unassign(orgID, actorID, contactID string) (ChatContact, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE contacts SET assigned_user_id = NULL, assigned_user_name = '', status = 'pending', updated_at = NOW()
		WHERE id = $1 AND organization_id = $2`, contactID, orgID)
	if err != nil {
		return ChatContact{}, err
	}
	s.addEvent(contactID, orgID, "unassigned", actorID, "", "Unassigned from agent", nil)
	return s.getContactByID(contactID, orgID, actorID)
}

func (s *PGStore) TogglePin(orgID, userID, contactID string, pinned bool) (ChatContact, error) {
	_, _ = s.db.Exec(s.ctx(), `
		INSERT INTO contact_user_states (user_id, contact_id, is_pinned)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id, contact_id) DO UPDATE SET is_pinned = $3`,
		userID, contactID, pinned)
	return s.getContactByID(contactID, orgID, userID)
}

func (s *PGStore) ToggleHide(orgID, userID, contactID string, hidden bool) (ChatContact, error) {
	_, _ = s.db.Exec(s.ctx(), `
		INSERT INTO contact_user_states (user_id, contact_id, is_hidden)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id, contact_id) DO UPDATE SET is_hidden = $3`,
		userID, contactID, hidden)
	return s.getContactByID(contactID, orgID, userID)
}

func (s *PGStore) Close(orgID, actorID, contactID string) (ChatContact, error) {
	now := time.Now()
	_, err := s.db.Exec(s.ctx(), `
		UPDATE contacts SET status = 'closed', closed_at = $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3`, now, contactID, orgID)
	if err != nil {
		return ChatContact{}, err
	}
	s.addEvent(contactID, orgID, "closed", actorID, "", "Closed the conversation", nil)
	return s.getContactByID(contactID, orgID, actorID)
}

func (s *PGStore) Reopen(orgID, actorID, contactID string) (ChatContact, error) {
	_, err := s.db.Exec(s.ctx(), `
		UPDATE contacts SET status = 'pending', closed_at = NULL, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2`, contactID, orgID)
	if err != nil {
		return ChatContact{}, err
	}
	s.addEvent(contactID, orgID, "reopened", actorID, "", "Reopened the conversation", nil)
	return s.getContactByID(contactID, orgID, actorID)
}

// ---------- ADMIN CONTACTS ----------

func (s *PGStore) ListAdminContacts(orgID, userID, search, instanceID string) (ContactsView, error) {
	q := `SELECT c.id, c.name, c.phone_number, c.avatar, c.status,
		c.assigned_user_id, c.assigned_user_name, c.instance_id, c.instance_name,
		c.instance_source_label, c.last_message_preview, c.last_message_at,
		c.last_inbound_at, c.closed_at, c.is_public, c.is_read, c.unread_count,
		c.tags, c.metadata, c.organization_id, c.created_at,
		false, false, ''
		FROM contacts c WHERE c.organization_id = $1`
	args := []interface{}{orgID}
	idx := 2
	if instanceID != "" {
		q += fmt.Sprintf(" AND c.instance_id = $%d", idx)
		args = append(args, instanceID)
		idx++
	}
	if search != "" {
		q += fmt.Sprintf(" AND (c.name ILIKE $%d OR c.phone_number ILIKE $%d)", idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	q += " ORDER BY c.name ASC"
	rows, err := s.db.Query(s.ctx(), q, args...)
	if err != nil {
		return ContactsView{}, err
	}
	defer rows.Close()
	contacts, err := scanContacts(rows)
	if err != nil {
		return ContactsView{}, err
	}
	instances, _ := s.ListInstances(orgID)
	return ContactsView{
		Contacts:   contacts,
		Instances:  instances,
		Search:     search,
		InstanceID: instanceID,
	}, nil
}

func (s *PGStore) CreateContact(orgID, actorID string, req ContactMutationRequest) (ChatContact, error) {
	phone := normalizePhoneNumber(req.PhoneNumber)
	if phone == "" {
		return ChatContact{}, errors.New("phone number is required")
	}

	instID, instName, instSourceLabel, err := s.resolveInstance(orgID, req.InstanceID)
	if err != nil {
		return ChatContact{}, err
	}

	// check duplicate
	var existCount int
	_ = s.db.QueryRow(s.ctx(), `SELECT COUNT(*) FROM contacts WHERE organization_id = $1 AND instance_id = $2 AND phone_number = $3`, orgID, instID, phone).Scan(&existCount)
	if existCount > 0 {
		return ChatContact{}, errors.New("contact already exists for this number and account")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = phone
	}
	tagsB, _ := json.Marshal(req.Tags)

	var contactID string
	err = s.db.QueryRow(s.ctx(), `
		INSERT INTO contacts (organization_id, instance_id, name, phone_number,
			avatar, status, instance_name, instance_source_label,
			last_message_preview, last_message_at, last_inbound_at, is_public, is_read, tags)
		VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,'Contact added.',$8,$8,true,true,$9)
		RETURNING id`,
		orgID, instID, name, phone, "https://i.pravatar.cc/150?u="+contactID,
		instName, instSourceLabel, time.Now(), tagsB).Scan(&contactID)
	if err != nil {
		return ChatContact{}, err
	}
	s.addEvent(contactID, orgID, "contact_created", actorID, "", "Created from contacts screen", nil)
	s.recordAudit(orgID, actorID, "", "contacts.create", "contact", contactID, "Created a contact.", nil)
	return s.getContactByID(contactID, orgID, actorID)
}

func (s *PGStore) UpdateContact(orgID, actorID, contactID string, req ContactMutationRequest) (ChatContact, error) {
	phone := normalizePhoneNumber(req.PhoneNumber)
	if phone == "" {
		return ChatContact{}, errors.New("phone number is required")
	}
	instID, instName, instSourceLabel, err := s.resolveInstance(orgID, req.InstanceID)
	if err != nil {
		return ChatContact{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = phone
	}
	tagsB, _ := json.Marshal(req.Tags)
	_, err = s.db.Exec(s.ctx(), `
		UPDATE contacts SET name = $1, phone_number = $2, instance_id = $3,
			instance_name = $4, instance_source_label = $5, tags = $6, updated_at = NOW()
		WHERE id = $7 AND organization_id = $8`,
		name, phone, instID, instName, instSourceLabel, tagsB, contactID, orgID)
	if err != nil {
		return ChatContact{}, err
	}
	s.recordAudit(orgID, actorID, "", "contacts.edit", "contact", contactID, "Updated contact details.", nil)
	return s.getContactByID(contactID, orgID, actorID)
}

func (s *PGStore) DeleteContact(orgID, actorID, contactID string) error {
	_, err := s.db.Exec(s.ctx(), `DELETE FROM contacts WHERE id = $1 AND organization_id = $2`, contactID, orgID)
	if err != nil {
		return err
	}
	s.recordAudit(orgID, actorID, "", "contacts.delete", "contact", contactID, "Deleted a contact.", nil)
	return nil
}

func (s *PGStore) resolveInstance(orgID, instanceID string) (id, name, sourceLabel string, err error) {
	if instanceID != "" {
		err = s.db.QueryRow(s.ctx(), `SELECT id, name, settings->>'source_tag_label' FROM whatsapp_instances WHERE organization_id = $1 AND id = $2`, orgID, instanceID).
			Scan(&id, &name, &sourceLabel)
		if err != nil {
			return "", "", "", errors.New("instance not found")
		}
		return
	}
	err = s.db.QueryRow(s.ctx(), `SELECT id, name, settings->>'source_tag_label' FROM whatsapp_instances WHERE organization_id = $1 LIMIT 1`, orgID).
		Scan(&id, &name, &sourceLabel)
	if err != nil {
		return "", "", "", errors.New("no WhatsApp accounts are available")
	}
	return
}

func (s *PGStore) ExportContactsCSV(orgID string, columns []string) (string, error) {
	if len(columns) == 0 {
		columns = []string{"name", "phone_number", "instance_name", "status", "assigned_user_name", "tags"}
	}
	rows, err := s.db.Query(s.ctx(), `
		SELECT name, phone_number, instance_name, status, assigned_user_name, tags
		FROM contacts WHERE organization_id = $1 ORDER BY name`, orgID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString(strings.Join(columns, ",") + "\n")
	for rows.Next() {
		var name, phone, instName, status, assignedName string
		var tagsB []byte
		_ = rows.Scan(&name, &phone, &instName, &status, &assignedName, &tagsB)
		var tags []string
		_ = json.Unmarshal(tagsB, &tags)
		row := map[string]string{
			"name": name, "phone_number": phone, "instance_name": instName,
			"status": status, "assigned_user_name": assignedName, "tags": strings.Join(tags, "|"),
		}
		vals := make([]string, 0, len(columns))
		for _, col := range columns {
			vals = append(vals, row[col])
		}
		sb.WriteString(strings.Join(vals, ",") + "\n")
	}
	return sb.String(), nil
}

func (s *PGStore) SampleContactsCSV(orgID string, columns []string) (string, error) {
	if len(columns) == 0 {
		columns = []string{"name", "phone_number", "instance_name", "status", "assigned_user_name", "tags"}
	}
	sample := map[string]string{
		"name": "Sample Contact", "phone_number": "+201500000000",
		"instance_name": "Sales WA", "status": "pending",
		"assigned_user_name": "", "tags": "sample|import",
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(columns, ",") + "\n")
	vals := make([]string, 0, len(columns))
	for _, col := range columns {
		vals = append(vals, sample[col])
	}
	sb.WriteString(strings.Join(vals, ",") + "\n")
	return sb.String(), nil
}

func (s *PGStore) ImportContactsCSV(orgID, actorID, csvData string, updateOnDuplicate bool) (ContactImportResult, error) {
	result := ContactImportResult{
		DuplicatePhones: []string{},
		Preview:         []ContactImportPreviewRow{},
	}
	lines := strings.Split(strings.TrimSpace(csvData), "\n")
	if len(lines) < 2 {
		return result, errors.New("CSV requires a header row and at least one data row")
	}
	header := map[string]int{}
	for i, col := range strings.Split(lines[0], ",") {
		header[strings.ToLower(strings.TrimSpace(col))] = i
	}
	if _, ok := header["phone_number"]; !ok {
		return result, errors.New("phone_number column is required")
	}
	job := s.recordJob(orgID, "contacts_import", "contacts", orgID, "Imported contacts from CSV.")
	for _, line := range lines[1:] {
		cols := strings.Split(line, ",")
		phone := normalizePhoneNumber(csvCell2(cols, header, "phone_number"))
		name := csvCell2(cols, header, "name")
		instName := csvCell2(cols, header, "instance_name")
		if phone == "" {
			result.Skipped++
			continue
		}
		instID, iName, iLabel, err := s.resolveInstanceByName(orgID, instName)
		if err != nil {
			result.Skipped++
			continue
		}
		if name == "" {
			name = phone
		}
		var existingID string
		dbErr := s.db.QueryRow(s.ctx(), `SELECT id FROM contacts WHERE organization_id = $1 AND instance_id = $2 AND phone_number = $3`, orgID, instID, phone).Scan(&existingID)
		if dbErr == nil {
			if !updateOnDuplicate {
				result.Skipped++
				result.DuplicatePhones = append(result.DuplicatePhones, phone)
				result.Preview = append(result.Preview, ContactImportPreviewRow{Name: name, PhoneNumber: phone, Instance: iName, Action: "duplicate_skipped"})
				continue
			}
			_, _ = s.db.Exec(s.ctx(), `UPDATE contacts SET name = $1, instance_id = $2, instance_name = $3, instance_source_label = $4, updated_at = NOW() WHERE id = $5`, name, instID, iName, iLabel, existingID)
			result.Updated++
			result.Preview = append(result.Preview, ContactImportPreviewRow{Name: name, PhoneNumber: phone, Instance: iName, Action: "updated"})
			continue
		}
		var cID string
		_ = s.db.QueryRow(s.ctx(), `
			INSERT INTO contacts (organization_id, instance_id, name, phone_number, avatar, status, instance_name, instance_source_label, last_message_preview, last_message_at, last_inbound_at, is_public, is_read)
			VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,'Imported from CSV.',NOW(),NOW(),true,true)
			RETURNING id`,
			orgID, instID, name, phone, "https://i.pravatar.cc/150?u="+phone, iName, iLabel).Scan(&cID)
		result.Created++
		result.Preview = append(result.Preview, ContactImportPreviewRow{Name: name, PhoneNumber: phone, Instance: iName, Action: "created"})
	}
	s.finishJob(job.ID, "completed", "")
	result.Job = job
	if len(result.Preview) > 6 {
		result.Preview = result.Preview[:6]
	}
	return result, nil
}

func (s *PGStore) resolveInstanceByName(orgID, instanceName string) (id, name, sourceLabel string, err error) {
	if instanceName != "" {
		err = s.db.QueryRow(s.ctx(), `SELECT id, name, settings->>'source_tag_label' FROM whatsapp_instances WHERE organization_id = $1 AND LOWER(name) = LOWER($2) LIMIT 1`, orgID, instanceName).
			Scan(&id, &name, &sourceLabel)
		if err == nil {
			return
		}
	}
	err = s.db.QueryRow(s.ctx(), `SELECT id, name, settings->>'source_tag_label' FROM whatsapp_instances WHERE organization_id = $1 LIMIT 1`, orgID).
		Scan(&id, &name, &sourceLabel)
	if err != nil {
		return "", "", "", errors.New("no WhatsApp accounts available")
	}
	return
}

func csvCell2(cols []string, header map[string]int, col string) string {
	idx, ok := header[col]
	if !ok || idx >= len(cols) {
		return ""
	}
	return strings.TrimSpace(cols[idx])
}

func (s *PGStore) ListClosedChats(orgID string, page, pageSize int, agentID, instanceID string) (ClosedChatPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := `SELECT id, name, phone_number, instance_id, instance_name, assigned_user_name, closed_at
		FROM contacts WHERE organization_id = $1 AND status = 'closed' AND closed_at IS NOT NULL`
	args := []interface{}{orgID}
	idx := 2
	if agentID != "" {
		q += fmt.Sprintf(" AND assigned_user_id = $%d", idx)
		args = append(args, agentID)
		idx++
	}
	if instanceID != "" {
		q += fmt.Sprintf(" AND instance_id = $%d", idx)
		args = append(args, instanceID)
		idx++
	}
	q += " ORDER BY closed_at DESC"
	_ = idx

	rows, err := s.db.Query(s.ctx(), q, args...)
	if err != nil {
		return ClosedChatPage{}, err
	}
	defer rows.Close()
	var allRows []ClosedConversationRow
	for rows.Next() {
		var r ClosedConversationRow
		var closedAt time.Time
		_ = rows.Scan(&r.ID, &r.ContactName, &r.PhoneDisplay, &r.InstanceID, &r.InstanceName, &r.AssignedUserName, &closedAt)
		r.ClosedAt = closedAt
		r.ClosedBy = r.AssignedUserName
		allRows = append(allRows, r)
	}
	total := len(allRows)
	sort.Slice(allRows, func(i, j int) bool { return allRows[i].ClosedAt.After(allRows[j].ClosedAt) })
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return ClosedChatPage{
		Items:       allRows[start:end],
		Page:        page,
		PageSize:    pageSize,
		Total:       total,
		HasNext:     end < total,
		HasPrevious: page > 1,
		AgentID:     agentID,
		InstanceID:  instanceID,
	}, nil
}
