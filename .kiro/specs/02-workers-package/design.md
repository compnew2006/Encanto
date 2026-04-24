# Design: Background Workers Package

## Package Location
`backend/workers/`

## Core Components

### `workers/worker.go` — Worker Pool
```go
type WorkerPool struct {
    store    *data.Store
    whatsapp *api.WhatsAppManager
    interval time.Duration
    quit     chan struct{}
    wg       sync.WaitGroup
}

func New(store *data.Store, whatsapp *api.WhatsAppManager) *WorkerPool
func (p *WorkerPool) Start(ctx context.Context)
func (p *WorkerPool) Stop()
```

### `workers/campaign.go` — Campaign Executor
```go
func (p *WorkerPool) processCampaignJob(ctx context.Context, job data.BackgroundJob)
```

## Polling Loop
The worker pool polls `background_jobs` WHERE `status = 'pending'` ORDER BY `created_at ASC` LIMIT 5 every **10 seconds**. Uses `SELECT ... FOR UPDATE SKIP LOCKED` to prevent double-processing in future multi-instance deployments.

## Integration in `main.go`
```go
pool := workers.New(store, server.WhatsApp)
go pool.Start(ctx)
defer pool.Stop()
```

## Sequence: Campaign Job Execution
```
WorkerPool → DB: SELECT pending campaign jobs
WorkerPool → DB: UPDATE job SET status='in_progress'
WorkerPool → DB: SELECT campaign recipients WHERE status='pending'
loop for each recipient:
  WorkerPool → WhatsAppManager: SendMessage(instanceID, phone, message)
  WorkerPool → DB: UPDATE recipient SET outcome='sent'|'failed'
WorkerPool → DB: UPDATE campaign SET status='completed'
WorkerPool → DB: UPDATE job SET status='completed', finished_at=NOW()
```

## Error Handling
- Per-recipient failures do NOT abort the campaign — log and mark recipient as `failed`
- Job-level panic → recover, mark job `failed`, continue loop
