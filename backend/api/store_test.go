package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerInitializesComponents(t *testing.T) {
	server := NewServer(nil)
	if server == nil {
		t.Fatal("expected server")
	}
	if server.hub == nil {
		t.Fatal("expected realtime hub to be initialized")
	}
	if server.WhatsApp == nil {
		t.Fatal("expected WhatsApp manager to be initialized")
	}
}

func TestWriteJSONEncodesPayload(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeJSON(recorder, http.StatusCreated, map[string]string{"status": "ok"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var payload map[string]string
	if err := json.NewDecoder(bytes.NewReader(recorder.Body.Bytes())).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected payload to round-trip, got %+v", payload)
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"encanto","extra":true}`))
	var got payload
	if err := decodeJSON(req, &got); err == nil {
		t.Fatal("expected decodeJSON to reject unknown fields")
	}
}

func TestCurrentOrgIDUsesCookieWhenPresent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := currentOrgID(req); got != "org-1" {
		t.Fatalf("expected default org-1, got %q", got)
	}

	req.AddCookie(&http.Cookie{Name: "org_context", Value: "org-2"})
	if got := currentOrgID(req); got != "org-2" {
		t.Fatalf("expected cookie org ID, got %q", got)
	}
}

func TestCurrentClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if _, err := currentClaims(req); err == nil {
		t.Fatal("expected missing claims to fail")
	}

	claims := &Claims{}
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))

	got, err := currentClaims(req)
	if err != nil {
		t.Fatalf("currentClaims returned error: %v", err)
	}
	if got != claims {
		t.Fatal("expected currentClaims to return the stored claims pointer")
	}

	req = req.WithContext(context.WithValue(req.Context(), "user", errors.New("not claims")))
	if _, err := currentClaims(req); err == nil {
		t.Fatal("expected wrong context type to fail")
	}
}
