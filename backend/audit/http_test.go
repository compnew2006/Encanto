package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorWithReason(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteErrorWithReason(rec, http.StatusForbidden, "forbidden", "access denied", "policy")

	if got, want := rec.Code, http.StatusForbidden; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}

	var envelope Envelope
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil {
		t.Fatal("response error = nil, want non-nil")
	}
	if got, want := envelope.Error.Code, "forbidden"; got != want {
		t.Fatalf("Error.Code = %q, want %q", got, want)
	}
	if got, want := envelope.Error.Message, "access denied"; got != want {
		t.Fatalf("Error.Message = %q, want %q", got, want)
	}
	if got, want := envelope.Error.DenialReason, "policy"; got != want {
		t.Fatalf("Error.DenialReason = %q, want %q", got, want)
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"encanto","unexpected":true}`))

	var dst struct {
		Name string `json:"name"`
	}
	if err := DecodeJSON(req, &dst); err == nil {
		t.Fatal("DecodeJSON() error = nil, want non-nil")
	}
}