package workers

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPollInterval = 10 * time.Second
	maxJobsPerPoll      = 5
)

type DBProvider interface {
	DB() *pgxpool.Pool
}

type CampaignSender interface {
	SendCampaignMessage(instanceID, phone, body string) error
}

type jobStore interface {
	claimPendingJobs(ctx context.Context, limit int) ([]backgroundJob, error)
	markJobFailed(ctx context.Context, jobID, reason string) error
	markJobCompleted(ctx context.Context, jobID string) error

	loadCampaign(ctx context.Context, orgID, campaignID string) (campaignRecord, error)
	createCampaignRun(ctx context.Context, orgID, campaignID, jobID, trigger string) (string, error)
	listCampaignTargets(ctx context.Context, orgID string, campaign campaignRecord) ([]campaignTarget, error)
	saveCampaignRecipient(ctx context.Context, runID string, recipient campaignRecipientOutcome) error
	updateCampaignRun(ctx context.Context, runID string, recipientTotal, delivered, failed int) error
	completeCampaign(ctx context.Context, orgID, campaignID, summary string) error
}

type WorkerPool struct {
	store    jobStore
	sender   CampaignSender
	interval time.Duration

	quit     chan struct{}
	stopOnce sync.Once
}

func New(store DBProvider, sender CampaignSender) *WorkerPool {
	var adapter jobStore
	if store != nil {
		adapter = newPGJobStore(store.DB())
	}

	return &WorkerPool{
		store:    adapter,
		sender:   sender,
		interval: defaultPollInterval,
		quit:     make(chan struct{}),
	}
}

func (p *WorkerPool) SetPollInterval(interval time.Duration) {
	if interval > 0 {
		p.interval = interval
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	ticker := time.NewTicker(p.pollInterval())
	defer ticker.Stop()

	p.pollOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.quit:
			return
		case <-ticker.C:
			p.pollOnce(ctx)
		}
	}
}

func (p *WorkerPool) Stop() {
	p.stopOnce.Do(func() {
		close(p.quit)
	})
}

func (p *WorkerPool) pollInterval() time.Duration {
	if p.interval <= 0 {
		return defaultPollInterval
	}
	return p.interval
}

func (p *WorkerPool) pollOnce(ctx context.Context) {
	if p.store == nil {
		return
	}

	jobs, err := p.store.claimPendingJobs(ctx, maxJobsPerPoll)
	if err != nil {
		log.Printf("worker: claim pending jobs: %v", err)
		return
	}

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return
		case <-p.quit:
			return
		default:
		}

		p.executeJob(ctx, job)
	}
}

func (p *WorkerPool) executeJob(ctx context.Context, job backgroundJob) {
	defer func() {
		if recovered := recover(); recovered != nil {
			stack := debug.Stack()
			log.Printf("worker panic job=%s kind=%s entity=%s/%s panic=%v\n%s",
				job.ID, job.Kind, job.EntityType, job.EntityID, recovered, stack)

			if p.store != nil {
				reason := fmt.Sprintf("panic: %v", recovered)
				if err := p.store.markJobFailed(ctx, job.ID, reason); err != nil {
					log.Printf("worker: mark failed after panic job=%s: %v", job.ID, err)
				}
			}
		}
	}()

	if p.store == nil {
		return
	}

	var err error
	switch {
	case job.Kind == "campaign_run" || job.EntityType == "campaign":
		err = p.processCampaignJob(ctx, job)
	default:
		err = fmt.Errorf("unsupported job kind %q for entity type %q", job.Kind, job.EntityType)
	}

	if err != nil {
		if markErr := p.store.markJobFailed(ctx, job.ID, err.Error()); markErr != nil {
			log.Printf("worker: mark failed job=%s: %v", job.ID, markErr)
		}
		return
	}

	if err := p.store.markJobCompleted(ctx, job.ID); err != nil {
		log.Printf("worker: mark completed job=%s: %v", job.ID, err)
	}
}

type backgroundJob struct {
	ID            string
	OrganizationID string
	Kind          string
	EntityType    string
	EntityID      string
	Status        string
	Summary       string
	FailureReason string
	StartedAt     time.Time
	FinishedAt    *time.Time
}

type pgJobStore struct {
	db *pgxpool.Pool
}

func newPGJobStore(db *pgxpool.Pool) *pgJobStore {
	return &pgJobStore{db: db}
}

func (s *pgJobStore) claimPendingJobs(ctx context.Context, limit int) ([]backgroundJob, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil || limit <= 0 {
		return []backgroundJob{}, nil
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, kind, entity_type, entity_id, status, summary, failure_reason, started_at, finished_at
		FROM background_jobs
		WHERE status = 'pending'
		ORDER BY started_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []backgroundJob
	for rows.Next() {
		var job backgroundJob
		if err := rows.Scan(
			&job.ID,
			&job.OrganizationID,
			&job.Kind,
			&job.EntityType,
			&job.EntityID,
			&job.Status,
			&job.Summary,
			&job.FailureReason,
			&job.StartedAt,
			&job.FinishedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if _, err := tx.Exec(ctx, `
			UPDATE background_jobs
			SET status = 'in_progress',
				started_at = NOW(),
				failure_reason = '',
				finished_at = NULL
			WHERE id = $1`, job.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	if jobs == nil {
		jobs = []backgroundJob{}
	}
	return jobs, nil
}

func (s *pgJobStore) markJobFailed(ctx context.Context, jobID, reason string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return nil
	}

	_, err := s.db.Exec(ctx, `
		UPDATE background_jobs
		SET status = 'failed',
			failure_reason = $1,
			finished_at = NOW()
		WHERE id = $2`, reason, jobID)
	return err
}

func (s *pgJobStore) markJobCompleted(ctx context.Context, jobID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil || s.db == nil {
		return nil
	}

	_, err := s.db.Exec(ctx, `
		UPDATE background_jobs
		SET status = 'completed',
			failure_reason = '',
			finished_at = NOW()
		WHERE id = $1`, jobID)
	return err
}
