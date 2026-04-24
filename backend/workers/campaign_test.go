package workers

import (
	"context"
	"testing"
	"time"

	"encanto/api"
)

type campaignStoreStub struct {
	campaignLoaded   bool
	runCreated       bool
	targetsListed    bool
	recipientSaved   bool
	runCompleted     bool
	campaignComplete bool

	savedRecipients []campaignRecipientOutcome
	createdRunID    string
	lastSummary     string
}

func (s *campaignStoreStub) claimPendingJobs(ctx context.Context, limit int) ([]backgroundJob, error) {
	return []backgroundJob{}, nil
}

func (s *campaignStoreStub) markJobFailed(ctx context.Context, jobID, reason string) error {
	return nil
}

func (s *campaignStoreStub) markJobCompleted(ctx context.Context, jobID string) error {
	return nil
}

func (s *campaignStoreStub) loadCampaign(ctx context.Context, orgID, campaignID string) (campaignRecord, error) {
	s.campaignLoaded = true
	return campaignRecord{
		Campaign: api.Campaign{
			ID:              campaignID,
			Name:            "Launch campaign",
			Status:          "scheduled",
			Source:          "manual",
			LinkedInstanceID: "instance-1",
			Content:         "Hello from the worker",
			Filters: api.CampaignFilters{
				InstanceID:    "instance-1",
				IncludeClosed:  false,
			},
			Schedule: api.CampaignSchedule{
				Mode: "manual",
			},
			LastRunSummary: "",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		},
	}, nil
}

func (s *campaignStoreStub) createCampaignRun(ctx context.Context, orgID, campaignID, jobID, trigger string) (string, error) {
	s.runCreated = true
	s.createdRunID = "run-1"
	return s.createdRunID, nil
}

func (s *campaignStoreStub) listCampaignTargets(ctx context.Context, orgID string, campaign campaignRecord) ([]campaignTarget, error) {
	s.targetsListed = true
	return []campaignTarget{
		{
			ContactID:   "contact-1",
			ContactName: "Alice",
			PhoneNumber: "15550001",
			InstanceID:  "instance-1",
		},
	}, nil
}

func (s *campaignStoreStub) saveCampaignRecipient(ctx context.Context, runID string, recipient campaignRecipientOutcome) error {
	s.recipientSaved = true
	s.savedRecipients = append(s.savedRecipients, recipient)
	return nil
}

func (s *campaignStoreStub) updateCampaignRun(ctx context.Context, runID string, recipientTotal, delivered, failed int) error {
	s.runCompleted = true
	return nil
}

func (s *campaignStoreStub) completeCampaign(ctx context.Context, orgID, campaignID, summary string) error {
	s.campaignComplete = true
	s.lastSummary = summary
	return nil
}

type campaignSenderStub struct {
	calls []struct {
		instanceID string
		phone      string
		body       string
	}
}

func (s *campaignSenderStub) SendCampaignMessage(instanceID, phone, body string) error {
	s.calls = append(s.calls, struct {
		instanceID string
		phone      string
		body       string
	}{
		instanceID: instanceID,
		phone:      phone,
		body:       body,
	})
	return nil
}

func TestProcessCampaignJobDispatchesRecipientsAndCompletesCampaign(t *testing.T) {
	store := &campaignStoreStub{}
	sender := &campaignSenderStub{}
	pool := &WorkerPool{
		store:  store,
		sender: sender,
		quit:   make(chan struct{}),
	}

	err := pool.processCampaignJob(context.Background(), backgroundJob{
		ID:             "job-1",
		OrganizationID: "org-1",
		Kind:           "campaign_run",
		EntityType:     "campaign",
		EntityID:       "campaign-1",
		Status:         "in_progress",
		StartedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("processCampaignJob returned error: %v", err)
	}

	if !store.campaignLoaded || !store.runCreated || !store.targetsListed || !store.recipientSaved || !store.runCompleted || !store.campaignComplete {
		t.Fatalf("expected campaign workflow to call all persistence steps, got store=%+v", store)
	}

	if len(sender.calls) != 1 {
		t.Fatalf("expected one send call, got %d", len(sender.calls))
	}
	if got := sender.calls[0]; got.instanceID != "instance-1" || got.phone != "15550001" || got.body != "Hello from the worker" {
		t.Fatalf("unexpected send payload: %+v", got)
	}

	if len(store.savedRecipients) != 1 {
		t.Fatalf("expected one saved recipient, got %d", len(store.savedRecipients))
	}
	if got := store.savedRecipients[0]; got.Status != "sent" || got.ContactID != "contact-1" {
		t.Fatalf("unexpected recipient outcome: %+v", got)
	}
	if store.lastSummary == "" {
		t.Fatalf("expected completion summary to be recorded")
	}
}
