package workers

import (
	"context"
	"strings"
	"testing"
	"time"
)

type panicStore struct {
	failedReason   string
	completedCount  int
	claimedCount    int
}

func (s *panicStore) claimPendingJobs(ctx context.Context, limit int) ([]backgroundJob, error) {
	s.claimedCount++
	return []backgroundJob{}, nil
}

func (s *panicStore) markJobFailed(ctx context.Context, jobID, reason string) error {
	s.failedReason = reason
	return nil
}

func (s *panicStore) markJobCompleted(ctx context.Context, jobID string) error {
	s.completedCount++
	return nil
}

func (s *panicStore) loadCampaign(ctx context.Context, orgID, campaignID string) (campaignRecord, error) {
	panic("boom")
}

func (s *panicStore) createCampaignRun(ctx context.Context, orgID, campaignID, jobID, trigger string) (string, error) {
	return "", nil
}

func (s *panicStore) listCampaignTargets(ctx context.Context, orgID string, campaign campaignRecord) ([]campaignTarget, error) {
	return nil, nil
}

func (s *panicStore) saveCampaignRecipient(ctx context.Context, runID string, recipient campaignRecipientOutcome) error {
	return nil
}

func (s *panicStore) updateCampaignRun(ctx context.Context, runID string, recipientTotal, delivered, failed int) error {
	return nil
}

func (s *panicStore) completeCampaign(ctx context.Context, orgID, campaignID, summary string) error {
	return nil
}

type noopCampaignSender struct{}

func (noopCampaignSender) SendCampaignMessage(instanceID, phone, body string) error {
	return nil
}

func TestWorkerPoolExecuteJobRecoversFromPanic(t *testing.T) {
	store := &panicStore{}
	pool := &WorkerPool{
		store:  store,
		sender: noopCampaignSender{},
		quit:   make(chan struct{}),
	}

	pool.executeJob(context.Background(), backgroundJob{
		ID:             "job-1",
		OrganizationID: "org-1",
		Kind:           "campaign_run",
		EntityType:     "campaign",
		EntityID:       "campaign-1",
		Status:         "in_progress",
		StartedAt:      time.Now(),
	})

	if store.failedReason == "" {
		t.Fatalf("expected failed reason to be recorded")
	}
	if !strings.Contains(store.failedReason, "boom") {
		t.Fatalf("expected panic reason to include boom, got %q", store.failedReason)
	}
	if store.completedCount != 0 {
		t.Fatalf("expected job not to be marked completed, got %d", store.completedCount)
	}
}
