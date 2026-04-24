# Tasks: Background Workers Package

- [x] 1. Create `backend/workers/` directory
- [x] 2. Create `backend/workers/worker.go` with `WorkerPool` struct, `New()`, `Start()`, `Stop()`
- [x] 3. Implement polling loop: `SELECT ... FOR UPDATE SKIP LOCKED` on `background_jobs` WHERE `status='pending'`
- [x] 4. Implement panic recovery middleware wrapping each job handler
- [x] 5. Create `backend/workers/campaign.go` with `processCampaignJob()`
- [x] 6. Implement per-recipient message dispatch via `WhatsAppManager.SendMessage()`
- [x] 7. Implement recipient outcome update (`sent` / `failed`) per dispatch result
- [x] 8. Implement campaign completion: update campaign status to `completed` after all recipients processed
- [x] 9. Wire `WorkerPool` into `backend/main.go` with context-based graceful shutdown
- [x] 10. Add `WORKER_POLL_INTERVAL` env var (default: `10s`) to `config/config.go`
- [x] 11. Write unit test for panic recovery in `workers/worker_test.go`
- [x] 12. Write integration test for campaign job execution in `workers/campaign_test.go`
