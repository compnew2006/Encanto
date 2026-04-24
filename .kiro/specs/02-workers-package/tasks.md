# Tasks: Background Workers Package

- [ ] 1. Create `backend/workers/` directory
- [ ] 2. Create `backend/workers/worker.go` with `WorkerPool` struct, `New()`, `Start()`, `Stop()`
- [ ] 3. Implement polling loop: `SELECT ... FOR UPDATE SKIP LOCKED` on `background_jobs` WHERE `status='pending'`
- [ ] 4. Implement panic recovery middleware wrapping each job handler
- [ ] 5. Create `backend/workers/campaign.go` with `processCampaignJob()`
- [ ] 6. Implement per-recipient message dispatch via `WhatsAppManager.SendMessage()`
- [ ] 7. Implement recipient outcome update (`sent` / `failed`) per dispatch result
- [ ] 8. Implement campaign completion: update campaign status to `completed` after all recipients processed
- [ ] 9. Wire `WorkerPool` into `backend/main.go` with context-based graceful shutdown
- [ ] 10. Add `WORKER_POLL_INTERVAL` env var (default: `10s`) to `config/config.go`
- [ ] 11. Write unit test for panic recovery in `workers/worker_test.go`
- [ ] 12. Write integration test for campaign job execution in `workers/campaign_test.go`
