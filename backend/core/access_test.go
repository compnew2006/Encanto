package core

import (
	"encoding/json"
	"testing"

	"encanto/data/sqlc"

	"github.com/google/uuid"
)

func TestApplyPermissionOverrides(t *testing.T) {
	t.Parallel()

	overrides := []sqlc.ListUserPermissionOverridesRow{
		{PermissionKey: "messages.send", Mode: "deny"},
		{PermissionKey: "chats.unclaimed.send", Mode: "allow"},
	}

	got := applyPermissionOverrides([]string{"messages.view", "messages.send"}, overrides)
	want := []PermissionKey{"chats.unclaimed.send", "messages.view"}

	if len(got) != len(want) {
		t.Fatalf("applyPermissionOverrides() length = %d, want %d (%v)", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("applyPermissionOverrides()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestResolveVisibilityWithRoleInheritance(t *testing.T) {
	t.Parallel()

	roleInstanceID := uuid.New()
	overrideInstanceID := uuid.New()
	role := sqlc.CustomRole{
		DefaultScopeMode:           string(ScopeInstancesOnly),
		DefaultAllowedInstanceIds:  mustJSONBytes([]string{roleInstanceID.String()}),
		DefaultAllowedPhoneNumbers: mustJSONBytes([]string{"+201000000001"}),
		CanViewUnmaskedPhone:       false,
	}
	rule := sqlc.UserContactVisibilityRule{
		ScopeMode:            string(ScopeAllowedNumbersOnly),
		AllowedInstanceIds:   mustJSONBytes([]string{overrideInstanceID.String()}),
		AllowedPhoneNumbers:  mustJSONBytes([]string{"+201000000003"}),
		InheritRoleScope:     true,
		CanViewUnmaskedPhone: true,
	}

	got := resolveVisibility(role, rule, true)

	if got.Mode != ScopeInstancesOnly {
		t.Fatalf("resolveVisibility() mode = %q, want %q", got.Mode, ScopeInstancesOnly)
	}
	if len(got.AllowedInstanceIDs) != 2 {
		t.Fatalf("resolveVisibility() instance count = %d, want 2", len(got.AllowedInstanceIDs))
	}
	if len(got.AllowedPhoneNumbers) != 2 {
		t.Fatalf("resolveVisibility() phone count = %d, want 2", len(got.AllowedPhoneNumbers))
	}
	if !got.CanViewUnmaskedPhone {
		t.Fatal("resolveVisibility() should preserve unmasked phone access when either scope grants it")
	}
}

func mustJSONBytes(values []string) []byte {
	encoded, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	return encoded
}
