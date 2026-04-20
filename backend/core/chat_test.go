package core

import "testing"

func TestMaybeMaskPhone(t *testing.T) {
	t.Parallel()

	if got := maybeMaskPhone("+201000000001", false); got != "+2****01" {
		t.Fatalf("maybeMaskPhone() = %q, want %q", got, "+2****01")
	}
	if got := maybeMaskPhone("+201000000001", true); got != "+201000000001" {
		t.Fatalf("maybeMaskPhone() with unmasked access = %q", got)
	}
}

func TestComposerStatePermissionParity(t *testing.T) {
	t.Parallel()

	service := &ChatService{access: &AccessService{}}

	readOnly := CurrentUserContext{
		EffectiveAccess: EffectiveAccess{
			PermissionKeys: []PermissionKey{"messages.view"},
		},
	}
	got := service.composerState(readOnly, "assigned")
	if got.Allowed || !got.Disabled {
		t.Fatalf("read-only composer state = %+v, want disabled", got)
	}
	if got.DenialReason != service.access.PermissionDeniedReason("messages.send") {
		t.Fatalf("read-only denial = %q", got.DenialReason)
	}

	pendingRestricted := CurrentUserContext{
		EffectiveAccess: EffectiveAccess{
			PermissionKeys: []PermissionKey{"messages.send"},
		},
	}
	got = service.composerState(pendingRestricted, "pending")
	if got.DenialReason != service.access.PermissionDeniedReason("chats.unclaimed.send") {
		t.Fatalf("pending denial = %q", got.DenialReason)
	}

	fullyAllowed := CurrentUserContext{
		EffectiveAccess: EffectiveAccess{
			PermissionKeys: []PermissionKey{"messages.send", "chats.unclaimed.send"},
		},
	}
	got = service.composerState(fullyAllowed, "pending")
	if !got.Allowed || got.Disabled || got.DenialReason != "" {
		t.Fatalf("fully allowed composer state = %+v, want enabled", got)
	}
}
