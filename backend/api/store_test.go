package api

import "testing"

func TestCreateDirectChatCreatesAssignedConversation(t *testing.T) {
	store := NewStore()

	contact, err := store.CreateDirectChat("org-1", "1", StartDirectChatRequest{
		PhoneNumber: "+201566677788",
		ProfileName: "QA Direct",
		InstanceID:  "inst-1",
	})
	if err != nil {
		t.Fatalf("CreateDirectChat returned error: %v", err)
	}

	if contact.Name != "QA Direct" {
		t.Fatalf("expected contact name to persist, got %q", contact.Name)
	}
	if contact.Status != "assigned" {
		t.Fatalf("expected new direct chat to be assigned, got %q", contact.Status)
	}
	if contact.AssignedUserID != "1" {
		t.Fatalf("expected creator to own the chat, got %q", contact.AssignedUserID)
	}

	snapshot, err := store.Workspace("org-1", "1", contact.ID, "assigned", "", "", "")
	if err != nil {
		t.Fatalf("Workspace returned error: %v", err)
	}
	if snapshot.Selected == nil || snapshot.Selected.Contact.ID != contact.ID {
		t.Fatalf("expected created chat to be selectable from workspace")
	}
}

func TestUpdateCleanupSettingsValidatesAndPersists(t *testing.T) {
	store := NewStore()

	if _, err := store.UpdateCleanupSettings("org-1", "1", CleanupSettings{RetentionDays: -1, RunHour: 3, Timezone: "Africa/Cairo"}); err == nil {
		t.Fatalf("expected negative retention to fail validation")
	}
	if _, err := store.UpdateCleanupSettings("org-1", "1", CleanupSettings{RetentionDays: 7, RunHour: 25, Timezone: "Africa/Cairo"}); err == nil {
		t.Fatalf("expected invalid cleanup hour to fail validation")
	}

	settings, err := store.UpdateCleanupSettings("org-1", "1", CleanupSettings{
		RetentionDays: 21,
		RunHour:       6,
		Timezone:      "Africa/Cairo",
	})
	if err != nil {
		t.Fatalf("UpdateCleanupSettings returned error: %v", err)
	}

	if settings.RetentionDays != 21 || settings.RunHour != 6 {
		t.Fatalf("expected cleanup schedule to persist, got %+v", settings)
	}
	if got := store.orgs["org-1"].NotificationsFeed[0].Title; got != "Cleanup schedule updated" {
		t.Fatalf("expected schedule update notification, got %q", got)
	}
}

func TestSendOutgoingMessagePreservesMediaMetadata(t *testing.T) {
	store := NewStore()

	message, err := store.SendOutgoingMessage("org-1", "1", "contact-1", SendMessageRequest{
		Type:          "media",
		Body:          "Attached quote draft",
		FileName:      "quote.txt",
		FileSizeLabel: "22 B",
		MediaURL:      "data:text/plain;base64,ZHJhZnQgcXVvdGU=",
	})
	if err != nil {
		t.Fatalf("SendOutgoingMessage returned error: %v", err)
	}

	if message.FileName != "quote.txt" {
		t.Fatalf("expected media filename to persist, got %q", message.FileName)
	}
	if message.FileSizeLabel != "22 B" {
		t.Fatalf("expected media file size label to persist, got %q", message.FileSizeLabel)
	}
	if message.TypedForMS != 0 {
		t.Fatalf("expected media sends to bypass typing simulation, got %dms", message.TypedForMS)
	}
}
