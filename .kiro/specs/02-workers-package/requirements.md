# Requirements: Background Workers Package

## Overview
The architecture specifies a `backend/workers/` package for background job processing, but this package does not exist. Campaign execution, scheduled cleanup, and long-running tasks currently have no execution engine.

## Requirements

### REQ-1: Worker Package Foundation
WHEN the backend starts
THE SYSTEM SHALL initialize a worker pool that reads pending jobs from the `background_jobs` table

### REQ-2: Campaign Execution Worker
WHEN a campaign's `scheduled_at` time is reached and its status is `scheduled`
THE SYSTEM SHALL dispatch messages to all recipients via the WhatsApp manager
AND update each recipient's outcome in `campaign_recipients` to `sent` or `failed`
AND update the campaign status to `completed` when all recipients are processed

### REQ-3: Job Status Tracking
WHEN a worker starts processing a job
THE SYSTEM SHALL update the job's status to `in_progress` and set `started_at`

WHEN a worker finishes a job successfully
THE SYSTEM SHALL update the job's status to `completed` and set `finished_at`

WHEN a worker encounters an unrecoverable error
THE SYSTEM SHALL update the job's status to `failed`, set `failure_reason`, and set `finished_at`

### REQ-4: Graceful Shutdown
WHEN the server receives SIGTERM or SIGINT
THE SYSTEM SHALL allow in-progress jobs to finish before exiting
AND shall NOT pick up new jobs after the shutdown signal is received

### REQ-5: Retry on Panic
WHEN a worker goroutine panics
THE SYSTEM SHALL recover the panic, mark the job as `failed`, log the stack trace, and continue processing other jobs
