# Whatomate Navigation Guide

Updated from the live navigation observed on 2026-04-19 across `/chat`, `/analytics/agents`, `/settings`, `/settings/license`, `/settings/instances`, `/settings/contacts`, and `/settings/closed-chats`. The earlier “reduced scope” map was too narrow and did not match the actual product menu.

## Primary Navigation

| Route | Purpose | Status |
| :--- | :--- | :--- |
| `/dashboard` | Operational overview cards and counters | Confirmed |
| `/chat` | Inbox and chat workspace | Confirmed |
| `/chat/[contactId]` | Conversation detail view | Confirmed |
| `/chatbot` | Chatbot module | Confirmed |
| `/analytics/agents` | Agent analytics module | Confirmed |
| `/campaigns` | Campaign management module | Confirmed |
| `/settings` | Settings hub | Confirmed |
| `/profile` | User profile | Confirmed through user menu |

## Sidebar Utility Surfaces

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Sidebar pin toggle | Pin sidebar closed | Keep persistent sidebar state in user settings |
| Organization switcher | Current org selector, create org, delete selected org | Keep org context and org CRUD in active scope |
| User menu | Availability, profile, theme, language, logout | Keep personal settings and availability in active scope |

## Settings Navigation

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Settings sidebar | General, Chatbot, WhatsApp, Contacts, Closed Chats, Canned Responses, Tags, Teams, Users, Roles, API Keys, Webhooks, Custom Actions, SSO, License | Keep settings hub broad; do not collapse it into one page |
| General settings tabs | General, Appearance, Chat, Notifications | Model tabbed settings under the same route |
| Uploads Cleanup | Retention days, daily cleanup hour, Run Cleanup Now | Keep cleanup scheduling and manual trigger in active scope, but gate it to admins only |
| Appearance tab | Color Mode, Theme Style, and presets such as Twitter, Ocean Breeze, Soft Pop, Amber Minimal | Keep appearance presets as first-class settings, not just theme toggles |
| Chat tab | Media Grouping Window, Sidebar Contact View, Chat Background, Show Print Buttons, Show Download Buttons, Closed Chats | Keep chat-behavior preferences in the settings plan |
| Notifications tab | Email Notifications, New Message Alerts, Notification Sound with Play, Campaign Updates | Keep user/org notification preferences in active scope |
| Assignment on settings page | No assign or ownership-transfer control was visible on `/settings` | Keep assignment logic scoped to chat/contact workflows |

## WhatsApp Instances Screen

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Page actions | Health Dashboard and Add Account | Keep dedicated routes for instance catalog and health view |
| Instance card header | Status badge, phone, JID, Edit instance name, Delete instance | Model instances as rich operational records, not just sessions |
| Connection actions | Disconnect for paired numbers, Connect / Scan QR for unpaired numbers | Keep explicit pairing and reconnect flows in the plan |
| Assignment on instance screen | No assign or ownership-transfer control was visible on this page | Keep assignment logic scoped to chat/contact workflows, not instance management |
| Instance metrics | Uptime, Queue, Sent / Received, Error Rate | Keep live health metrics and snapshots per instance |
| Quick toggles | Auto-sync history and Auto-download incoming media | Keep per-instance operational settings |
| Policy dialogs | Call auto-reject, Auto campaign, Chat Close Rating Settings, Assigned Chat Reset | Keep instance-level automation and rating flows in active scope |
| Chat Source Tag | Custom label, display mode, color palette, save button | Keep source-tag rendering config for inbox and contact chips |

## License Screen

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Page header | `License` and `Manage offline activation, quotas, and renewal status for this deployment.` | Keep a dedicated deployment-license page in active scope |
| Top action | `Refresh` | Keep manual bootstrap refresh in the UI |
| Status card | On 2026-04-19 the live deployment showed `Disabled` | Model status states such as disabled, unlicensed, active, grace, and locked |
| Quota cards | Organizations, Users / Org, WA Endpoints / Org, Storage / Org, Subscription Days | Keep license usage snapshots and entitlement limits in the bootstrap payload |
| Current organization usage | Per-org usage summary was visible | Keep per-organization usage details in the data model |
| Server identity | `HWID`, `Short ID`, and `Copy HWID` | Keep explicit host fingerprint and copy flow |
| Activation panel | `Security Key`, `Activate license`, `Refresh status` | Keep offline token paste/activate/renew flow |
| Access policy | Code confirms `/settings/license` is admin-only | Keep license management restricted to admins |

## License Cleanup Route

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Redirect behavior | Code confirms authenticated users get redirected to `/license-cleanup` when quota is over | Keep a dedicated overage-resolution route |
| Allowed actions | Cleanup view focuses on deleting organizations, users, accounts, or instances until usage drops | Keep cleanup mode narrower than normal app access |
| Exit condition | Once overage clears, the user returns to normal app flow | Keep bootstrap re-checks and automatic recovery |

## Agent Analytics Screen

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Filter bar | `All Agents`, `All`, `This month`, `Any`, `Export CSV` | Keep analytics filters and CSV export in active scope |
| KPI cards | Transfers Handled, Completed conversations, Active Conversations, Avg Resolution Time, Avg Queue Time, Break Time, Average Rating | Keep agent-performance summary endpoints and materialized datasets |
| Analytics sections | Transfer Trends, Conversation Sources, Agent Comparison | Keep chart datasets and aggregated analytics views |
| Customer Ratings table | Agent, Phone Number, Contact, Rating, Rated At, Closing Agent, Rating Message, Context Messages | Keep closure ratings and message snapshots in the data model |
| Phone drill-down | Clicking the phone number opened the related chat | Keep chat deep-linking from analytics |
| Assignment on analytics page | No assign or ownership-transfer control was visible on `/analytics/agents` | Keep assignment logic scoped to chat/contact workflows |

## Contacts Screen

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Top actions | `Import/Export` and `Add Contact` | Keep contacts CRUD and CSV workflows in active scope |
| Search and filter | `Search contacts...` and `All instances` | Keep contact list filtering first-class in the route design |
| Table columns | Name, Phone Number, WhatsApp Instance, Tags, Last message, Created | Keep a denormalized last-message preview in planning |
| Row actions | `Open chat`, `Edit`, `Delete` | Keep direct chat navigation and in-place contact editing |
| Create dialog | Phone Number, Profile Name, WhatsApp Instance | Keep explicit instance selection on manual contact creation |
| Import/Export dialog | Export CSV, Import CSV, Download sample CSV, Update existing records if duplicate found | Keep CSV import/export with dedupe behavior |
| Assignment on contacts page | No direct assign button was visible, but `Assigned User ID` appears in export columns | Keep assignment data in the model without forcing an assignment UI on this page |

## Closed Chats Screen

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Filter bar | `All Agents`, `All instances`, `25`, `Refresh` | Keep a dedicated closed-chats listing endpoint with filters |
| Table columns | Contact Name, Closed By, Date Closed, Actions | Keep closed-chat metadata queryable without reopening the chat |
| Row action | `Reopen` | Keep explicit reopen workflow in the plan |
| Pagination | `Previous` and `Next` | Keep stable filtered pagination for operational review |
| Assignment on closed chats page | No assign or ownership-transfer control was visible | Keep assignment logic scoped to live chat workflows |

## Chat Workspace Overlays

| Surface | Observed Capability | Backend Need |
| :--- | :--- | :--- |
| Notifications dialog | Unread items and Mark all as read | `/api/notifications*` |
| Statuses drawer | Add status and status feed | `/api/statuses*` |
| Start New Chat | Phone number + profile name + instance | `/api/chats/direct` |
| Assign Contact | Searchable user picker | `/api/contacts/{id}/assign` |
| Invite Collaborator | Searchable collaborator picker | `/api/contacts/{id}/collaborators*` |
| Notes panel | Internal notes feed and create note box | `/api/contacts/{id}/notes*` |
| Contact Info panel | Tags, collaborators, general data, chatbot-configured data panel | Tags + collaborators + chatbot flow config retention |

## Planning Hardening Additions

هذه ليست عناصر تنقل مؤكدة من الواجهة، لكنها أضيفت إلى الخطة لأن غيابها كان يترك فجوات تنفيذية واضحة:

| Surface | Why It Must Exist | Suggested Placement |
| :--- | :--- | :--- |
| Conversation timeline | Required for assignment history, queue/resolution analytics, and support debugging | Drawer inside `/chat/[contactId]` |
| Webhook delivery log | Required because plan already includes retries, tests, and failure handling | Modal or tab inside `/settings/webhooks` |
| Background job progress | Required for cleanup/import/reconnect/campaign visibility | Modal or admin drawer from settings/actions |
| Audit log view | Required for license, roles, cleanup, and destructive actions | Admin-only surface under settings or operations |

## Vendor License Studio

| Surface | Observed Capability | Plan Impact |
| :--- | :--- | :--- |
| Generate tab | Issue offline license from HWID + private key + entitlements | Keep vendor console separate from the main app |
| Verify tab | Validate token and show tracked/untracked/invalid | Keep a verification workflow for support and ops |
| Registry tab | Filter registry, copy token, re-issue, remove | Keep a vendor-side registry of issued licenses |
| Customer metadata | `customer_name` is merged by the guard into registry entries | Keep vendor-side customer metadata outside the main app DB |
| Guard layer | Login, sessions, audit log, customer metadata, guarded delete | Model the vendor console as a proxied companion service, not part of the product UI |

## Still Unverified

| Area | Reason | Status |
| :--- | :--- | :--- |
| Meta Cloud API | Did not appear in the audited chat workflow | Not confirmed |
| `/templates` | Not visible in the audited navigation | Not confirmed |
| `/widgets` | Not visible in the audited navigation | Not confirmed |
| AI / SLA / transfer queue | No direct evidence from the current live audit | Not confirmed |
| Keyword rules | Did not appear in the audited settings and instance flows | Not confirmed |
