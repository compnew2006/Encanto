package audit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	DenialReason string `json:"denialReason,omitempty"`
}

type Envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteErrorWithReason(w, status, code, message, "")
}

func WriteErrorWithReason(w http.ResponseWriter, status int, code, message, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Error: &APIError{
			Code:         code,
			Message:      message,
			DenialReason: reason,
		},
	})
}

func DecodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}
