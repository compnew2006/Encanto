# Milestones

This file turns the execution guide into delivery milestones. Each milestone has a goal, scope, dependencies, and exit criteria.

## Milestone 1: Stabilize the Final Picture

- Objective: Lock one final interpretation of scope, visible product behavior, and first-release success.
- Includes:
  - Final scope approval
  - Unified screen and behavior reference
  - First-release definition of done
- Depends on: existing planning documents and audited surfaces
- Exit criteria:
  - In-scope, later-scope, and out-of-scope lists are approved
  - Screen behavior no longer conflicts across documents
  - First-release success is measurable

## Milestone 2: Establish the Project Foundation

- Objective: Create the buildable project base and execution order.
- Includes:
  - Project structure
  - Core conventions
  - Full build sequence
- Depends on: Milestone 1
- Exit criteria:
  - Structure is agreed
  - Naming and organization rules are documented
  - Work order is dependency-safe

## Milestone 3: Identity and Context

- Objective: Build access, current-user context, and safe context switching.
- Includes:
  - Login, logout, session continuity
  - Current user surface
  - Context switching where allowed
- Depends on: Milestone 2
- Exit criteria:
  - Users can enter and leave safely
  - Current-user context loads consistently
  - Context switching changes permissions and data correctly

## Milestone 4: Permissions and Visibility

- Objective: Decide who can do what and who can see what before operational buildout.
- Includes:
  - Action-based permissions
  - Read-only behavior
  - Visibility scope rules
  - UI enforcement alignment
- Depends on: Milestone 3
- Exit criteria:
  - Permission catalog exists
  - Visibility decisions are enforced in lists and detail views
  - UI and backend behavior no longer diverge

## Milestone 5: Core Models

- Objective: Lock the main entities and their supporting records.
- Includes:
  - Primary models
  - Supporting collaboration models
  - Personal state and event history models
- Depends on: Milestone 4
- Exit criteria:
  - Main entities are stable
  - Collaboration records are defined
  - Personal state and event history are part of the baseline model

## Milestone 6: Conversation Workspace

- Objective: Deliver the operational inbox and open-conversation experience.
- Includes:
  - Conversation list
  - Open conversation screen
  - Day-to-day conversation actions
- Depends on: Milestone 5
- Exit criteria:
  - Users can browse and open conversations
  - Notes, info, and timeline are visible
  - Assignment and related actions affect live state and history

## Milestone 7: Sending and Receiving

- Objective: Make inbound and outbound conversation flow reliable.
- Includes:
  - Inbound processing
  - Outbound text flow
  - Media flow and failure handling
- Depends on: Milestone 6
- Exit criteria:
  - Inbound items appear in the correct conversation
  - Text sends use the correct pre-send behavior
  - Media sends bypass text-only logic
  - Failure and retry paths are traceable

## Milestone 8: Realtime and Notifications

- Objective: Keep operational surfaces live and current.
- Includes:
  - Live updates channel
  - Notifications center
  - Status drawer
- Depends on: Milestone 7
- Exit criteria:
  - Live changes appear in the correct surfaces
  - Reconnect and resync behavior is defined
  - Notifications and statuses are operational

## Milestone 9: Settings Center

- Objective: Deliver general and personal settings with stable effect on the product.
- Includes:
  - General settings
  - Personal settings
  - Cleanup scheduling and manual run
- Depends on: Milestone 8
- Exit criteria:
  - Settings persist correctly
  - Settings affect the intended screens and behavior
  - Cleanup controls are restricted appropriately

## Milestone 10: Account Operations and Health

- Objective: Treat sending accounts as live operational units.
- Includes:
  - Account card catalog
  - Connect, disconnect, recover flows
  - Health and account-specific policies
- Depends on: Milestone 9
- Exit criteria:
  - Accounts can be managed operationally
  - Connection lifecycle is visible and controlled
  - Health and policy behavior are saved and reviewable

## Milestone 11: Contacts and Closed Conversations

- Objective: Complete the confirmed administrative surfaces.
- Includes:
  - Contacts management
  - Import and export
  - Closed conversations review and reopen
- Depends on: Milestone 10
- Exit criteria:
  - Contacts CRUD works
  - Import/export is safe and reviewable
  - Closed conversations can be filtered and reopened

## Milestone 12: Licensing and Limits

- Objective: Build controlled operation under license and resource limits.
- Includes:
  - License page and activation
  - Usage and limit visibility
  - Restricted cleanup mode
- Depends on: Milestone 11
- Exit criteria:
  - License state is visible and actionable
  - Limits are enforced with understandable behavior
  - Over-limit flow moves users into restricted cleanup safely

## Milestone 13: Analytics

- Objective: Deliver trustworthy reporting from operational facts.
- Includes:
  - Core metric derivation
  - Analytics screen
  - Event-based metric validation
- Depends on: Milestone 12
- Exit criteria:
  - Analytics derive from recorded facts
  - Filters and exports work
  - Drill-down to conversation is supported

## Milestone 14: Campaigns

- Objective: Add campaign definition, execution, and recipient-level review.
- Includes:
  - Campaign definition
  - Campaign runs
  - Recipient outcomes
  - Link from operational automation into campaigns
- Depends on: Milestone 13
- Exit criteria:
  - Campaigns are reusable entities
  - Runs are stored independently
  - Recipient outcomes are inspectable

## Milestone 15: Reliability and Audit

- Objective: Close the gaps around background execution, delivery reliability, and traceability.
- Includes:
  - Background job history
  - Reliable outbound event delivery
  - General audit log
- Depends on: Milestone 14
- Exit criteria:
  - Long-running work leaves a reviewable record
  - Important post-save events are not silently lost
  - Sensitive actions are auditable

## Milestone 16: Verification, Cleanup, and Handoff

- Objective: Make the project safe to declare complete and safe to continue later.
- Includes:
  - Verification plan
  - Final cleanup
  - Handoff guide
- Depends on: Milestone 15
- Exit criteria:
  - Verification coverage is explicit
  - Final cleanup is complete
  - Another engineer or AI assistant can continue from the handoff pack

## Recommended Order

Follow the milestones strictly in numeric order. Do not move to a later milestone unless the current one has a clear exit state.

## Working Rule

At the end of each milestone:

- Mark completed tasks in [checklist.md](/Users/noiemany/Downloads/Encanto/checklist.md)
- Update progress and notes in [plan.html](/Users/noiemany/Downloads/Encanto/plan.html)
- Record any scope change before beginning the next milestone
