package core

import (
	"context"
	"testing"
	"time"

	"encanto/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestSessionManagerIssueRotateAndInvalidate(t *testing.T) {
	t.Parallel()

	manager := newTestSessionManager(t)
	ctx := context.Background()
	userID := uuid.New()
	organizationID := uuid.New()

	accessToken, refreshToken, expiresAt, err := manager.Issue(ctx, userID, organizationID, "operator@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("Issue() expiry should be in the future, got %s", expiresAt)
	}

	session, err := manager.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if session.UserID != userID || session.OrganizationID != organizationID {
		t.Fatalf("ParseAccessToken() = %+v", session)
	}

	rotatedAccess, rotatedRefresh, _, rotatedSession, err := manager.Rotate(ctx, refreshToken)
	if err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}
	if rotatedRefresh == refreshToken {
		t.Fatal("Rotate() should rotate the refresh token")
	}
	if _, _, err := manager.loadRefreshState(ctx, refreshToken); err == nil {
		t.Fatal("old refresh token should be invalid after rotation")
	}

	parsedRotated, err := manager.ParseAccessToken(rotatedAccess)
	if err != nil {
		t.Fatalf("ParseAccessToken(rotated) error = %v", err)
	}
	if rotatedSession.SessionID != parsedRotated.SessionID {
		t.Fatalf("Rotate() returned session %q, want %q", rotatedSession.SessionID, parsedRotated.SessionID)
	}

	if err := manager.Invalidate(ctx, rotatedRefresh); err != nil {
		t.Fatalf("Invalidate() error = %v", err)
	}
	if _, _, err := manager.loadRefreshState(ctx, rotatedRefresh); err == nil {
		t.Fatal("Invalidate() should delete the rotated refresh token")
	}
}

func TestSessionManagerRotateForOrganization(t *testing.T) {
	t.Parallel()

	manager := newTestSessionManager(t)
	ctx := context.Background()
	userID := uuid.New()
	firstOrganizationID := uuid.New()
	secondOrganizationID := uuid.New()

	accessToken, refreshToken, _, err := manager.Issue(ctx, userID, firstOrganizationID, "switcher@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	session, err := manager.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	rotatedAccess, rotatedRefresh, _, err := manager.RotateForOrganization(ctx, session, secondOrganizationID)
	if err != nil {
		t.Fatalf("RotateForOrganization() error = %v", err)
	}
	if _, _, err := manager.loadRefreshState(ctx, refreshToken); err == nil {
		t.Fatal("RotateForOrganization() should invalidate the original refresh token")
	}

	parsedRotated, err := manager.ParseAccessToken(rotatedAccess)
	if err != nil {
		t.Fatalf("ParseAccessToken(rotated) error = %v", err)
	}
	if parsedRotated.OrganizationID != secondOrganizationID {
		t.Fatalf("rotated access org = %s, want %s", parsedRotated.OrganizationID, secondOrganizationID)
	}
	if _, _, err := manager.loadRefreshState(ctx, rotatedRefresh); err != nil {
		t.Fatalf("rotated refresh should exist: %v", err)
	}
}

func newTestSessionManager(t *testing.T) *SessionManager {
	t.Helper()

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(miniRedis.Close)

	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	return NewSessionManager(config.Config{
		AccessTokenSecret: "test-access-secret",
		AccessTokenTTL:    time.Minute,
		RefreshTokenTTL:   time.Hour,
		AccessCookieName:  "encanto_access",
		RefreshCookieName: "encanto_refresh",
	}, client)
}
