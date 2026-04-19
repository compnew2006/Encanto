# Product Requirements Document

## Document Info

- Product: Encanto / Whatomate planning baseline
- Document type: PRD
- Status: Draft for execution
- Source of truth: [plan.html](/Users/noiemany/Downloads/Encanto/plan.html), [01_api_endpoints.md](/Users/noiemany/Downloads/Encanto/01_api_endpoints.md), [02_app_routes.md](/Users/noiemany/Downloads/Encanto/02_app_routes.md), [03_database_schema.md](/Users/noiemany/Downloads/Encanto/03_database_schema.md), [05_business_logic.md](/Users/noiemany/Downloads/Encanto/05_business_logic.md)

## 1. Product Summary

This product is a team-based messaging operations platform centered on live conversations, operational assignment, internal collaboration, account operations, reporting, licensing, and controlled growth through phased delivery.

The first release is not a stripped demo. It must support real operational work with:

- Secure access and stable user context
- Team inbox and conversation handling
- Human assignment and collaboration
- Text and media messaging
- Realtime updates and notifications
- Operational settings and account management
- Contact and closed-conversation administration
- Licensing and quota enforcement
- Analytics and campaigns
- Reliable background processing and auditability

## 1.1 Tech Stack

**Frontend**
- Svelte 5
- SvelteKit

**Backend**
- Go
- chi
- pgx
- sqlc
- WebSocket داخل Go

**Infra**
- PostgreSQL
- Redis

**WhatsApp providers**
- whatsmeow لاحقًا

## 2. Problem Statement

Teams need one workspace to:

- Receive and manage conversations in real time
- Assign ownership clearly
- Preserve operational history
- Control what each user can see and do
- Manage sending accounts and their health
- Review performance and outcomes
- Enforce tenant limits safely
- Continue operating without losing critical events or state

Without a structured build order, these needs create drift between UI, logic, data, permissions, and operations. This project must be built in an order that avoids rework.

## 3. Product Goals

### Primary Goals

1. Deliver a production-capable team conversation workspace.
2. Make assignment, visibility, and operational state consistent everywhere.
3. Preserve full traceability for important user and system actions.
4. Support controlled tenant operation through limits, licensing, and cleanup flows.
5. Enable measurable operations through analytics and reviewable execution history.

### Secondary Goals

1. Make the project easy to continue by another engineer or AI assistant.
2. Keep the architecture modular enough to add later modules without redesign.
3. Ensure progress can be tracked phase by phase with clear completion criteria.

## 4. Non-Goals

The following are not part of the core committed scope unless explicitly reintroduced later:

- Template management
- Widget management
- AI-generated replies
- SLA automation
- Keyword-rule automation
- Transfer queue workflows as a dedicated module

## 5. Users and Roles

### Organization Admin

- Manages users, roles, settings, licensing, limits, accounts, cleanup, and integrations
- Needs full visibility and control of operational surfaces

### Agent / Operator

- Works in the inbox
- Sends and receives messages
- Assigns, collaborates, writes notes, and closes or reopens conversations
- Needs filtered visibility based on role and explicit allowances

### Team Lead / Supervisor

- Reviews assignments, queue behavior, performance, and closed conversations
- Needs analytics, drill-down, and operational visibility

### Internal Support / Elevated User

- May switch context across organizations when allowed
- Needs safe access without breaking tenant isolation

## 6. In-Scope Product Modules

### Core Workspace

- Login, session continuity, current-user context
- Organization context switching when allowed
- Team inbox with tabs, filters, search, and personal state
- Conversation view with header actions, notes, info panel, and timeline

### Messaging

- Inbound message processing
- Outbound text and media sending
- Retry, revoke, read state, and failure visibility
- Rich handling of personal and operational state

### Realtime Surfaces

- Live updates for conversation state
- Notifications center
- Status drawer
- Reliable reconnect and resync behavior

### Operational Admin

- Settings hub
- Contact management
- Closed conversation review
- Account catalog, pairing lifecycle, health view, and account-specific policies

### Governance and Control

- Roles and permissions
- Visibility rules
- Licensing and quota enforcement
- Cleanup mode when over limits

### Measurement and Growth

- Agent analytics
- Campaign definition, runs, and recipient-level outcomes

### Reliability and Audit

- Background jobs
- Reliable outbound event delivery
- Delivery attempt tracking
- Audit logging

## 7. Product Surfaces

### Primary Navigation

- Dashboard
- Chat
- Chatbot-adjacent configuration surface
- Agent analytics
- Campaigns
- Settings
- Profile

### Embedded Operational Surfaces

- Notifications dialog
- Statuses drawer
- Start new chat dialog
- Assign contact dialog
- Invite collaborator dialog
- Notes panel
- Contact info panel
- Conversation timeline drawer

### Settings Surfaces

- General
- Appearance
- Chat behavior
- Notifications
- Users
- Roles
- Teams
- Accounts
- Contacts
- Closed conversations
- Canned responses
- Tags
- API keys
- Webhooks
- Custom actions
- SSO
- License

## 8. Functional Requirements

### 8.1 Access and Session

- Users must be able to log in, log out, and maintain a stable session.
- Current-user data must be available globally after login.
- Allowed context switching must update data and permissions consistently.
- Personal preferences must persist between sessions.

### 8.2 Roles and Visibility

- Access control must be action-based, not role-name based only.
- Read-only modes must be supported where viewing is allowed but mutation is not.
- Visibility must support full scope, narrowed scope, and explicit allowed-item scope.
- The same visibility rules must apply to lists, search, detail views, and exports.

### 8.3 Core Data Model

- The system must represent organizations, users, teams, contacts, conversations, messages, media, notes, collaborators, tags, canned responses, and operational settings.
- Personal inbox state must be stored per user.
- Operational conversation history must be stored as explicit events.
- Message send attempts must be tracked independently from final message state.

### 8.4 Inbox and Conversation Handling

- The inbox must support operational tabs, search, filters, and personal state.
- Opening a conversation must expose header actions, notes, info, and event timeline.
- Assignment, unassignment, pin, hide, close, reopen, and visibility changes must be reflected immediately and recorded historically.

### 8.5 Sending and Receiving

- Incoming items must create or update the correct conversation reliably.
- Outgoing text must support natural pre-send behavior before actual delivery.
- Non-text sends must bypass the text-only pre-send behavior.
- Failures must remain inspectable and retryable.

### 8.6 Realtime Behavior

- Important operational changes must reach the right open surface with minimal delay.
- Reconnect behavior must support continuation or controlled resync when needed.
- Notifications and statuses must reflect live changes.

### 8.7 Settings and Admin

- Organization and personal settings must be editable and persistent.
- Cleanup scheduling and manual cleanup must exist and be restricted to high-privilege users.
- User, role, team, integration, and SSO administration must be represented in the product.

### 8.8 Account Operations

- Accounts must be managed as operational entities, not just passive records.
- Connection state, health, and policy behavior must be visible and editable.
- Pairing, reconnect, disconnect, delete, and rename flows must be controlled and traceable.

### 8.9 Contacts and Closed Conversations

- Contacts must support CRUD, import/export, filtering, and opening the linked conversation.
- Closed conversations must support review, filtering, refresh, and reopen.

### 8.10 Licensing and Limits

- The product must show current license state, identity, usage, and limits.
- Activation and renewal must be possible from the product.
- Exceeding shared limits must trigger constrained behavior instead of silent breakage.
- General overage must move the user to a restricted cleanup mode until usage returns within bounds.

### 8.11 Analytics

- Analytics must derive from explicit operational facts, not opaque estimates.
- Core metrics must support drill-down back to the underlying conversation.
- Exports must honor the same filters shown on screen.

### 8.12 Campaigns

- Campaigns must be definable, editable, schedulable, and runnable.
- Each run must be stored independently.
- Recipient-level outcomes must be inspectable.
- Operational automation that launches campaigns must route through the same campaign domain.

### 8.13 Reliability and Audit

- Long-running or retryable work must have visible execution records.
- Important post-save events must not be lost.
- Sensitive actions must be written to a general audit log.

## 9. Operational and Quality Requirements

### Consistency

- UI state, permission enforcement, and recorded history must agree.

### Recoverability

- Interruptions must not silently drop important operational events.

### Traceability

- Assignment changes, send failures, destructive actions, cleanup, and licensing operations must all be reviewable later.

### Safe Degradation

- When limits are reached, the product must move into controlled behavior rather than broad failure.

### Delivery Readiness

- Each phase must end with an observable output and a clear next step.

## 10. Release Success Criteria

The first release is considered successful when all of the following are true:

- A user can enter the product, load current context, and work within the correct scope.
- Conversations can be received, opened, assigned, replied to, and reviewed.
- Text and media sending behave correctly, including failure handling.
- Notifications and live updates keep the operational surfaces current.
- Settings, contacts, closed conversations, and account operations work as expected.
- License state and usage limits are visible and enforced safely.
- Analytics and campaigns are available on top of operational data.
- Background work, outbound event delivery, and sensitive actions are traceable.

## 11. Delivery Approach

Execution follows the 16-phase sequence defined in the plan:

1. Stabilize the final picture
2. Establish the project foundation
3. Build identity and context
4. Build permissions and visibility
5. Lock the core models
6. Build the conversation workspace
7. Build sending and receiving
8. Build realtime and notifications
9. Build the settings center
10. Build account operations and health
11. Build contacts and closed-conversation admin
12. Build licensing and limits
13. Build analytics
14. Build campaigns
15. Build reliability and audit layers
16. Verify, clean, and hand off

For the exact phase structure and completion path, use:

- [mailstone.md](/Users/noiemany/Downloads/Encanto/mailstone.md)
- [checklist.md](/Users/noiemany/Downloads/Encanto/checklist.md)
- [plan.html](/Users/noiemany/Downloads/Encanto/plan.html)
