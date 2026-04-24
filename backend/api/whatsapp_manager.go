package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type WhatsAppManager struct {
	server  *Server
	clients map[string]*whatsmeow.Client
	mu      sync.RWMutex
}

func NewWhatsAppManager(server *Server) *WhatsAppManager {
	return &WhatsAppManager{
		server:  server,
		clients: make(map[string]*whatsmeow.Client),
	}
}

func (m *WhatsAppManager) GetClient(instanceID string) (*whatsmeow.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[instanceID]
	return client, ok
}

func (m *WhatsAppManager) StartInstance(orgID, instanceID string, device *store.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[instanceID]; ok {
		return nil // Already running
	}

	client := whatsmeow.NewClient(device, nil)
	client.AddEventHandler(func(evt interface{}) {
		m.handleEvent(orgID, instanceID, client, evt)
	})

	m.clients[instanceID] = client

	// Connect in background
	go func() {
		if err := client.Connect(); err != nil {
			fmt.Printf("Failed to connect instance %s: %v\n", instanceID, err)
		}
	}()

	return nil
}

func (m *WhatsAppManager) StopInstance(instanceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[instanceID]; ok {
		client.Disconnect()
		delete(m.clients, instanceID)
	}
}

// LoadSessions queries all instances from DB and starts those that should be connected.
func (m *WhatsAppManager) LoadSessions() error {
	instances, err := m.server.store.ListInstances("") // List all across orgs (internal use)
	if err != nil {
		return err
	}

	for _, inst := range instances {
		if inst.Status == "connected" && inst.JID != "" {
			jid, err := types.ParseJID(inst.JID)
			if err != nil {
				continue
			}
			device, _ := m.server.store.waContainer.GetDevice(context.Background(), jid)
			if device != nil {
				_ = m.StartInstance(inst.OrganizationID, inst.ID, device)
			}
		}
	}
	return nil
}

func (m *WhatsAppManager) handleEvent(orgID, instanceID string, client *whatsmeow.Client, evt interface{}) {
	switch v := evt.(type) {
	case *events.QR:
		// Update DB with new QR code
		_ = m.server.store.updateInstanceQRCode(orgID, instanceID, v.Codes[0])
		// Broadcast to Hub
		m.server.hub.Publish(orgID, "qr_updated", map[string]any{
			"instance_id": instanceID,
			"qr_code":     v.Codes[0],
		})

	case *events.Connected:
		// Scan successful or reconnected!
		if client.Store.ID != nil {
			jid := client.Store.ID.String()
			_ = m.server.store.updateInstanceConnectionSuccess(orgID, instanceID, jid)
			
			// Fetch instance to broadcast complete state
			inst, err := m.server.store.getInstanceByID(instanceID)
			if err == nil {
				m.server.hub.Publish(orgID, "instance_connected", map[string]any{
					"instance": inst,
				})
			}
		}

	case *events.Message:
		// Ignore messages from self (outbound)
		if v.Info.IsFromMe {
			return
		}

		// Extract text body
		body := ""
		if v.Message.GetConversation() != "" {
			body = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetText() != "" {
			body = v.Message.GetExtendedTextMessage().GetText()
		}

		if body == "" {
			return // skip media/stickers for now to keep it simple
		}

		senderJID := m.resolveSenderJID(client, v.Info)
		groupName := ""
		if v.Info.IsGroup {
			groupName = m.resolveGroupName(client, v.Info.Chat)
		}

		msg, err := m.server.store.HandleInboundMessage(orgID, instanceID, v.Info.Chat.String(), senderJID.String(), groupName, body, "text")
		if err == nil {
			// Broadcast to frontend
			m.server.hub.Publish(orgID, "new_message", map[string]any{
				"instance_id": instanceID,
				"contact_id":  msg.ContactID,
				"message":     msg,
			})
		}

	case *events.LoggedOut:
		_ = m.server.store.updateInstanceStatus(orgID, instanceID, "disconnected", "needs_qr")
		m.server.hub.Publish(orgID, "instance_disconnected", map[string]any{
			"instance_id": instanceID,
		})
		m.StopInstance(instanceID)
	}
}

func (m *WhatsAppManager) resolveSenderJID(client *whatsmeow.Client, info types.MessageInfo) types.JID {
	if !info.SenderAlt.IsEmpty() && info.SenderAlt.Server == types.DefaultUserServer {
		return info.SenderAlt
	}

	if info.Sender.Server != types.HiddenUserServer {
		return info.Sender
	}

	if client != nil && client.Store != nil && client.Store.LIDs != nil {
		if pn, err := client.Store.LIDs.GetPNForLID(context.Background(), info.Sender); err == nil && !pn.IsEmpty() {
			return pn
		}
	}

	return info.Sender
}

func (m *WhatsAppManager) resolveGroupName(client *whatsmeow.Client, chatJID types.JID) string {
	if client == nil || chatJID.Server != types.GroupServer {
		return ""
	}

	info, err := client.GetGroupInfo(context.Background(), chatJID)
	if err != nil || info == nil {
		return ""
	}

	return strings.TrimSpace(info.Name)
}

func (m *WhatsAppManager) sendText(instanceID, phone, body string) error {
	client, ok := m.GetClient(instanceID)
	if !ok {
		return errors.New("WhatsApp instance not found or not running")
	}

	server := types.DefaultUserServer
	if len(phone) > 10 && !strings.Contains(phone, "@") && (phone[0:3] == "149" || len(phone) > 12) {
		// Heuristic or check if it was previously identified as LID
		// For now, most are DefaultUserServer. The store resolved LID to PN earlier.
	}

	targetJID := types.NewJID(phone, server)

	_, err := client.SendMessage(context.Background(), targetJID, &waE2E.Message{
		Conversation: proto.String(body),
	})
	return err
}

func (m *WhatsAppManager) SendCampaignMessage(instanceID, phone, body string) error {
	return m.sendText(instanceID, phone, body)
}

func (m *WhatsAppManager) SendMessage(orgID, instanceID, phone, messageID, body string) {
	err := m.sendText(instanceID, phone, body)

	status := "sent"
	reason := ""
	if err != nil {
		status = "failed"
		reason = err.Error()
	}

	updatedMsg, err := m.server.store.UpdateMessageStatus(orgID, messageID, status, reason)
	if err == nil {
		// Broadcast update to frontend
		m.server.hub.Publish(orgID, "status_update", map[string]any{
			"contact_id": updatedMsg.ContactID,
			"message":    updatedMsg,
		})
	}
}

// updateInstanceQRCode, updateInstanceStatus, and updateInstanceConnectionSuccess 
// will need to be added to PGStore.
