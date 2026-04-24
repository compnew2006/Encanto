# Tasks: Campaign Execution Engine

- [ ] 1. Add `scheduleCampaigns()` ticker in `workers/worker.go` running every 60 seconds
- [ ] 2. Implement the SQL UPDATE to transition `scheduled` → `running` campaigns and create jobs
- [ ] 3. Create `workers/campaign.go` with `processCampaignJob()` function
- [ ] 4. Add pause/resume check: re-fetch campaign status before each recipient
- [ ] 5. Add configurable delay between messages via `CAMPAIGN_MESSAGE_DELAY_MS`
- [ ] 6. Update recipient outcomes (`sent`/`failed`) after each dispatch
- [ ] 7. Mark campaign as `completed` after all recipients processed
- [ ] 8. Add migration for `completed_at` column if missing from `campaigns` table
- [ ] 9. Add sqlc query `UpdateCampaignStatus` and `UpdateRecipientOutcome`
- [ ] 10. Write integration test simulating a 3-recipient campaign
