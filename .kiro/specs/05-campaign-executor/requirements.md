# Requirements: Campaign Execution Engine

## Overview
Campaigns can be created and scheduled via the API, but no execution engine exists to send messages to recipients. This spec wires the campaign lifecycle from `scheduled` → `running` → `completed`.

## Requirements

### REQ-1: Campaign Dispatch Trigger
WHEN a campaign's `scheduled_at <= NOW()` and its status is `scheduled`
THE SYSTEM SHALL change the campaign status to `running` and create a `background_job` of kind `campaign_dispatch`

### REQ-2: Message Sending
WHEN a campaign job is being processed
THE SYSTEM SHALL send the campaign message to each recipient with status `pending`
AND respect a configurable delay between messages (default: 1 second) to avoid WhatsApp rate-limit bans

### REQ-3: Outcome Tracking
WHEN a message is sent successfully to a recipient
THE SYSTEM SHALL update `campaign_recipients.outcome` to `sent` and set `sent_at = NOW()`

WHEN a message fails for a recipient
THE SYSTEM SHALL update `campaign_recipients.outcome` to `failed` and store the error reason

### REQ-4: Completion
WHEN all recipients have been processed
THE SYSTEM SHALL update the campaign status to `completed` and set `completed_at = NOW()`

### REQ-5: Pause / Resume
WHEN a campaign status is set to `paused` by an admin
THE SYSTEM SHALL stop processing remaining recipients until status is changed back to `running`
