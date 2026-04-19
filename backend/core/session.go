package core

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encanto/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionManager struct {
	config config.Config
	redis  *redis.Client
}

type refreshState struct {
	SessionID      string    `json:"sessionId"`
	UserID         string    `json:"userId"`
	OrganizationID string    `json:"organizationId"`
	Email          string    `json:"email"`
	SecretHash     string    `json:"secretHash"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

type accessClaims struct {
	SessionID      string `json:"sid"`
	UserID         string `json:"uid"`
	OrganizationID string `json:"orgId"`
	Email          string `json:"email"`
	jwt.RegisteredClaims
}

func NewSessionManager(cfg config.Config, redisClient *redis.Client) *SessionManager {
	return &SessionManager{config: cfg, redis: redisClient}
}

func (m *SessionManager) Issue(ctx context.Context, userID, organizationID uuid.UUID, email string) (string, string, time.Time, error) {
	sessionID := uuid.NewString()
	secret, err := randomSecret(32)
	if err != nil {
		return "", "", time.Time{}, err
	}

	expiresAt := time.Now().Add(m.config.AccessTokenTTL)
	accessToken, err := m.signAccessToken(sessionID, userID, organizationID, email, expiresAt)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken := sessionID + "." + secret
	state := refreshState{
		SessionID:      sessionID,
		UserID:         userID.String(),
		OrganizationID: organizationID.String(),
		Email:          email,
		SecretHash:     hashSecret(secret),
		ExpiresAt:      time.Now().Add(m.config.RefreshTokenTTL),
	}

	if err := m.storeRefreshState(ctx, state); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, expiresAt, nil
}

func (m *SessionManager) Rotate(ctx context.Context, refreshToken string) (string, string, time.Time, CurrentSession, error) {
	state, secret, err := m.loadRefreshState(ctx, refreshToken)
	if err != nil {
		return "", "", time.Time{}, CurrentSession{}, err
	}
	if hashSecret(secret) != state.SecretHash {
		return "", "", time.Time{}, CurrentSession{}, fmt.Errorf("refresh token mismatch")
	}

	userID, err := uuid.Parse(state.UserID)
	if err != nil {
		return "", "", time.Time{}, CurrentSession{}, err
	}
	organizationID, err := uuid.Parse(state.OrganizationID)
	if err != nil {
		return "", "", time.Time{}, CurrentSession{}, err
	}

	if err := m.redis.Del(ctx, refreshRedisKey(state.SessionID)).Err(); err != nil {
		return "", "", time.Time{}, CurrentSession{}, fmt.Errorf("delete old refresh state: %w", err)
	}

	accessToken, newRefreshToken, expiresAt, err := m.Issue(ctx, userID, organizationID, state.Email)
	if err != nil {
		return "", "", time.Time{}, CurrentSession{}, err
	}

	session, err := m.ParseAccessToken(accessToken)
	if err != nil {
		return "", "", time.Time{}, CurrentSession{}, err
	}

	return accessToken, newRefreshToken, expiresAt, session, nil
}

func (m *SessionManager) RotateForOrganization(ctx context.Context, session CurrentSession, newOrganizationID uuid.UUID) (string, string, time.Time, error) {
	if err := m.InvalidateBySessionID(ctx, session.SessionID); err != nil {
		return "", "", time.Time{}, err
	}
	return m.Issue(ctx, session.UserID, newOrganizationID, session.Email)
}

func (m *SessionManager) Invalidate(ctx context.Context, refreshToken string) error {
	state, _, err := m.loadRefreshState(ctx, refreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "missing refresh token") {
			return nil
		}
		return err
	}
	return m.InvalidateBySessionID(ctx, state.SessionID)
}

func (m *SessionManager) InvalidateBySessionID(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	if err := m.redis.Del(ctx, refreshRedisKey(sessionID)).Err(); err != nil {
		return fmt.Errorf("invalidate refresh state: %w", err)
	}
	return nil
}

func (m *SessionManager) ParseAccessToken(token string) (CurrentSession, error) {
	claims := &accessClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (any, error) {
		return []byte(m.config.AccessTokenSecret), nil
	})
	if err != nil || !parsed.Valid {
		return CurrentSession{}, fmt.Errorf("parse access token: %w", err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return CurrentSession{}, fmt.Errorf("parse access user id: %w", err)
	}
	organizationID, err := uuid.Parse(claims.OrganizationID)
	if err != nil {
		return CurrentSession{}, fmt.Errorf("parse access org id: %w", err)
	}

	return CurrentSession{
		SessionID:      claims.SessionID,
		UserID:         userID,
		OrganizationID: organizationID,
		Email:          claims.Email,
		ExpiresAt:      claims.ExpiresAt.Time,
	}, nil
}

func (m *SessionManager) SetCookies(w http.ResponseWriter, accessToken, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.AccessCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})

	refreshExpiresAt := time.Now().Add(m.config.RefreshTokenTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  refreshExpiresAt,
		MaxAge:   int(time.Until(refreshExpiresAt).Seconds()),
	})
}

func (m *SessionManager) ClearCookies(w http.ResponseWriter) {
	for _, name := range []string{m.config.AccessCookieName, m.config.RefreshCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
}

func (m *SessionManager) signAccessToken(sessionID string, userID, organizationID uuid.UUID, email string, expiresAt time.Time) (string, error) {
	claims := accessClaims{
		SessionID:      sessionID,
		UserID:         userID.String(),
		OrganizationID: organizationID.String(),
		Email:          email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.config.AccessTokenSecret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (m *SessionManager) storeRefreshState(ctx context.Context, state refreshState) error {
	serialized, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal refresh state: %w", err)
	}
	if err := m.redis.Set(ctx, refreshRedisKey(state.SessionID), serialized, m.config.RefreshTokenTTL).Err(); err != nil {
		return fmt.Errorf("store refresh state: %w", err)
	}
	return nil
}

func (m *SessionManager) loadRefreshState(ctx context.Context, refreshToken string) (refreshState, string, error) {
	parts := strings.Split(refreshToken, ".")
	if len(parts) != 2 {
		return refreshState{}, "", fmt.Errorf("missing refresh token")
	}

	raw, err := m.redis.Get(ctx, refreshRedisKey(parts[0])).Bytes()
	if err != nil {
		return refreshState{}, "", fmt.Errorf("load refresh state: %w", err)
	}

	var state refreshState
	if err := json.Unmarshal(raw, &state); err != nil {
		return refreshState{}, "", fmt.Errorf("decode refresh state: %w", err)
	}
	return state, parts[1], nil
}

func refreshRedisKey(sessionID string) string {
	return "refresh:" + sessionID
}

func randomSecret(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("random secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
