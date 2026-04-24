# Design: Campaign Execution Engine

## Scheduler (in workers package)
A new function `scheduleCampaigns()` runs every 60 seconds:
```sql
UPDATE campaigns SET status='running'
WHERE status='scheduled' AND scheduled_at <= NOW()
RETURNING id, organization_id
```
For each returned campaign, insert a `background_job` of kind `campaign_dispatch`.

## Executor (`workers/campaign.go`)
```go
func (p *WorkerPool) processCampaignJob(ctx context.Context, job BackgroundJob)
```
1. Load campaign + recipients WHERE outcome='pending'
2. Check campaign status — if `paused`, release lock and exit
3. Loop recipients with configurable delay (`CAMPAIGN_MESSAGE_DELAY_MS`, default 1000ms)
4. Call `WhatsAppManager.SendText(instanceID, recipientPhone, message)`
5. Update recipient outcome
6. After loop: UPDATE campaign SET status='completed', completed_at=NOW()

## New Config Key
`CAMPAIGN_MESSAGE_DELAY_MS` (default: `1000`) in `config/config.go`

## DB Changes Required
Verify `campaigns` table has `completed_at TIMESTAMPTZ` column — add via migration if missing.
