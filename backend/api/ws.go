package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type WSClaims struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	jwt.RegisteredClaims
}

type WSMessage[T any] struct {
	Type           string `json:"type"`
	EventID        string `json:"event_id,omitempty"`
	Sequence       int64  `json:"sequence,omitempty"`
	OccurredAt     string `json:"occurred_at,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	Payload        T      `json:"payload"`
}

type clientMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type realtimeClient struct {
	conn           *websocket.Conn
	userID         string
	organizationID string
}

type RealtimeHub struct {
	upgrader websocket.Upgrader
	mu       sync.RWMutex
	clients  map[*realtimeClient]struct{}
	sequence int64
}

func NewRealtimeHub() *RealtimeHub {
	return &RealtimeHub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			Subprotocols: []string{"whm.v1"},
		},
		clients: map[*realtimeClient]struct{}{},
	}
}

func (s *Server) WSToken(w http.ResponseWriter, r *http.Request) {
	claims, err := currentClaims(r)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &WSClaims{
		UserID:         claims.UserID,
		OrganizationID: currentOrgID(r),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "failed to sign ws token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":           tokenString,
		"user_id":         claims.UserID,
		"organization_id": currentOrgID(r),
		"expires_at":      expirationTime.Format(time.RFC3339),
	})
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.hub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &realtimeClient{conn: conn}
	s.hub.register(client)
	defer s.hub.unregister(client)
	defer conn.Close()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var message clientMessage
		if err := json.Unmarshal(payload, &message); err != nil {
			_ = conn.WriteJSON(WSMessage[map[string]string]{
				Type:    "error",
				Payload: map[string]string{"error": "invalid ws payload"},
			})
			continue
		}

		switch message.Type {
		case "auth":
			tokenValue, _ := message.Payload["token"].(string)
			if tokenValue == "" {
				return
			}

			claims := &WSClaims{}
			token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtSecretKey, nil
			})
			if err != nil || !token.Valid {
				return
			}

			client.userID = claims.UserID
			client.organizationID = claims.OrganizationID

			_ = conn.WriteJSON(WSMessage[map[string]string]{
				Type:           "auth_ok",
				OrganizationID: claims.OrganizationID,
				Payload: map[string]string{
					"user_id":         claims.UserID,
					"organization_id": claims.OrganizationID,
				},
			})
		case "ping":
			_ = conn.WriteJSON(WSMessage[map[string]string]{
				Type:    "pong",
				Payload: map[string]string{},
			})
		}
	}
}

func (h *RealtimeHub) register(client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}
}

func (h *RealtimeHub) unregister(client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
}

func (h *RealtimeHub) Publish(organizationID, eventType string, payload any) {
	message := WSMessage[any]{
		Type:           eventType,
		EventID:        eventType + "-" + time.Now().Format("150405"),
		Sequence:       atomic.AddInt64(&h.sequence, 1),
		OccurredAt:     time.Now().Format(time.RFC3339),
		OrganizationID: organizationID,
		Payload:        payload,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.organizationID != "" && client.organizationID != organizationID {
			continue
		}
		_ = client.conn.WriteJSON(message)
	}
}
