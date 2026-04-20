# Phase Checklist

Use this file as the markdown companion to the interactive guide. Check items only when the expected result is actually complete.

## Phase 1: Stabilize the Final Picture

- [x] `1-1` Final scope approved
  Result: in-scope, later-scope, and out-of-scope lists are explicit
- [x] `1-2` Screen and behavior reference unified
  Result: no conflicting behavior across planning files
- [x] `1-3` First-release success defined
  Result: release readiness is measurable

Notes:

- Done during the initial project breakdown.

## Phase 2: Establish the Project Foundation

- [x] `2-1` Project structure mapped
  Result: clean boundaries and folder structure
- [x] `2-2` Conventions fixed
  Result: naming, error handling, and organization rules documented
- [x] `2-3` Build order defined
  Result: dependency-safe delivery sequence exists

Notes:

- Done using SvelteKit 5 / Tailwind v4 and Go chi backend.

## Phase 3: Identity and Context

- [x] `3-1` Access flow completed
  Result: login, logout, and session continuity work
- [x] `3-2` Current-user context completed
  Result: current-user data, presence, and personal menu are stable
- [x] `3-3` Context switching completed
  Result: switching updates scope, visibility, and data correctly

Notes:

-

## Phase 4: Permissions and Visibility

- [x] `4-1` Action-based permission model defined
  Result: actions are clear and reusable
- [x] `4-2` Visibility rules implemented
  Result: users see only what they should see
- [x] `4-3` UI aligned with enforcement
  Result: buttons, screens, and backend decisions match

Notes:

-

## Phase 5: Core Models

- [ ] `5-1` Main entities stabilized
  Result: primary records are agreed and usable
- [ ] `5-2` Supporting collaboration entities added
  Result: notes, collaborators, tags, and quick replies are modeled
- [ ] `5-3` Personal state and event history added
  Result: per-user inbox state and operational event history exist

Notes:

-

## Phase 6: Conversation Workspace

- [ ] `6-1` Conversation list completed
  Result: tabs, search, filters, and counts are usable
- [ ] `6-2` Open conversation screen completed
  Result: header, message area, notes, info, and timeline work
- [ ] `6-3` Daily conversation actions completed
  Result: assign, unassign, pin, hide, and related actions work

Notes:

-

## Phase 7: Sending and Receiving

- [ ] `7-1` Inbound handling completed
  Result: incoming items appear in the correct conversation
- [ ] `7-2` Outbound text flow completed
  Result: text sends use the intended pre-send behavior and track attempts
- [ ] `7-3` Media flow and failure handling completed
  Result: attachments, retry, revoke, and failure details work

Notes:

-

## Phase 8: Realtime and Notifications

- [ ] `8-1` Live update channel completed
  Result: operational changes reach the correct open surfaces
- [ ] `8-2` Notifications center completed
  Result: notification review and read actions work
- [ ] `8-3` Status drawer completed
  Result: statuses can be added, viewed, and updated live

Notes:

-

## Phase 9: Settings Center

- [ ] `9-1` General and personal settings completed
  Result: settings are editable and persistent
- [ ] `9-2` Scheduled and manual cleanup completed
  Result: cleanup is configurable, runnable, and restricted
- [ ] `9-3` Settings effect propagation completed
  Result: saved settings actually change product behavior

Notes:

-

## Phase 10: Account Operations and Health

- [ ] `10-1` Operational account catalog completed
  Result: each account is visible as a live operational unit
- [ ] `10-2` Connect, disconnect, and recovery completed
  Result: lifecycle actions are controlled and traceable
- [ ] `10-3` Health and policy surfaces completed
  Result: health and account-specific policies are visible and saved

Notes:

-

## Phase 11: Contacts and Closed Conversations

- [x] `11-1` Contacts screen completed
  Result: CRUD, search, filter, and open-conversation actions work
- [x] `11-2` Import/export completed
  Result: import, export, and duplicate handling are safe
- [x] `11-3` Closed conversations screen completed
  Result: review, filter, refresh, and reopen work

Notes:

- Added `/settings/contacts` and `/settings/closed-chats` with end-to-end CRUD, CSV import/export, and reopen-to-inbox flow backed by new contacts/closed-chat APIs.

## Phase 12: Licensing and Limits

- [x] `12-1` License page and activation completed
  Result: current state, identity, and activation flow work
- [x] `12-2` Limits and enforcement completed
  Result: usage and limit behavior are visible and enforced
- [x] `12-3` Restricted cleanup mode completed
  Result: over-limit flow safely narrows the usable surface

Notes:

- Added `/settings/license` and `/license-cleanup`, enforced quota checks on contact/instance/campaign creation paths, and redirected restricted tenants into cleanup mode until usage returned within limits.

## Phase 13: Analytics

- [x] `13-1` Metric derivation completed
  Result: metrics derive from recorded facts
- [x] `13-2` Analytics screen completed
  Result: filters, cards, tables, and export work
- [x] `13-3` Event-based validation completed
  Result: every important number can be explained from its source records

Notes:

- Added analytics derivation from conversation events, ratings, and assignment history with `/analytics/agents` cards, breakdowns, comparison rows, drill-down links, and CSV export.

## Phase 14: Campaigns

- [x] `14-1` Campaign definition completed
  Result: campaigns can be created, saved, edited, and scheduled
- [x] `14-2` Runs and recipient tracking completed
  Result: each run and recipient outcome is inspectable
- [x] `14-3` Operational automation linked to campaigns
  Result: automatic launch paths route through the same campaign domain

Notes:

- Added reusable campaign definitions, run history, recipient outcomes, delete support, and linked instance auto-campaign syncing so operational automations flow through the same campaign model.

## Phase 15: Reliability and Audit

- [x] `15-1` Background execution history completed
  Result: long-running work is reviewable
- [x] `15-2` Reliable outbound event delivery completed
  Result: important events are not silently lost after save
- [x] `15-3` General audit log completed
  Result: sensitive actions are traceable later

Notes:

- Added job history, outbox/webhook delivery tracking with retry, and audit records across cleanup, messaging, account operations, licensing, contacts, and campaigns, surfaced in `/settings/audit`.

## Phase 16: Verification, Cleanup, and Handoff

- [x] `16-1` Verification plan completed
  Result: normal, failure, and recovery paths are covered
- [x] `16-2` Final cleanup completed
  Result: naming, structure, and leftovers are cleaned up
- [x] `16-3` Handoff pack completed
  Result: another engineer or AI assistant can continue safely

Notes:

- Verified backend, frontend type checks, Chromium smoke, and the full Playwright cross-browser suite; refreshed the checklist and overwrote `summary.md` with the handoff record for the next session.
