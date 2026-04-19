# Whatomate API Endpoints

هذه الوثيقة تمثل **سطح الـ API بعد تدقيق الواجهة الحية** لشاشات `/chat` و`/analytics/agents` و`/settings` و`/settings/license` و`/settings/instances` و`/settings/contacts` و`/settings/closed-chats`.

الاستنتاج الأهم: الخطة لا يجب أن تبقى محصورة في `chat core` فقط، لأن الواجهة الفعلية أكدت وجود:

- `Dashboard`
- `Chat`
- `Chatbot`
- `Agent Analytics`
- `Campaigns`
- `Notifications`
- `Statuses`
- `Start New Chat`
- `Assign / Notes / Contact Info / Tags / Collaborators`
- `WhatsApp instance cards`
- `Health Dashboard`
- `Call auto-reject / Auto campaign / Rating message / Assigned chat reset`
- `Appearance / Chat / Notifications tabs`
- `Text-message typing simulation via whatsmeow`

مع بقاء `Meta Cloud API` خارج النطاق المؤكد حالياً.

## Base URL

عادة يكون:
`http://localhost:8080`

## Health & System

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/health` | Basic health check |
| GET | `/ready` | Readiness check |
| GET | `/metrics` | Prometheus metrics |
| GET | `/ws` | WebSocket endpoint |
| GET | `/api/config` | Runtime config and feature flags |

## Authentication & SSO

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| POST | `/api/auth/login` | User login |
| POST | `/api/auth/register` | User registration |
| POST | `/api/auth/refresh` | Refresh JWT token |
| POST | `/api/auth/logout` | User logout |
| POST | `/api/auth/switch-org` | Switch active organization for elevated or explicitly multi-org accounts |
| GET | `/api/auth/ws-token` | Short-lived token for WebSocket |
| POST | `/api/auth/register/invite` | Generate registration invite |
| GET | `/api/auth/sso/providers` | List public SSO providers |
| GET | `/api/auth/sso/{provider}/init` | Start SSO flow |
| GET | `/api/auth/sso/{provider}/callback` | SSO callback |

## Current User & Sidebar Context

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/me` | Get current user profile |
| GET | `/api/auth/me` | Legacy alias for `/api/me` |
| PUT | `/api/me/settings` | Update locale, theme selection, sidebar preferences, and personal settings |
| GET | `/api/me/organizations` | List user organizations; usually a single active org in SaaS mode |
| PUT | `/api/me/password` | Change password |
| PUT | `/api/me/availability` | Update availability status (`available`, `unavailable`, `busy`) |
| GET | `/api/me/chat-background` | Get chat background |
| POST | `/api/me/chat-background` | Upload chat background |

## Settings Hub

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/settings/general` | Get general organization settings |
| PUT | `/api/settings/general` | Update org name, slug, timezone, default locale, language, date format, and mask-phone setting |
| GET | `/api/settings/limits` | Get tenant quotas, current usage, slot allocation, storage usage, and tenant status |
| PUT | `/api/settings/limits` | Update tenant quotas and runtime guardrails (platform admin only) |
| GET | `/api/settings/appearance` | Get personal appearance settings like color mode (`light`/`dark`/`system`) and Tailwind theme preset |
| PUT | `/api/settings/appearance` | Update color mode, follow-system behavior, and Tailwind theme style across devices |
| GET | `/api/settings/chat` | Get chat preferences like media grouping, sidebar contact view, sidebar hover-expand/pin, background, and print/download visibility |
| PUT | `/api/settings/chat` | Update chat behavior settings |
| GET | `/api/settings/notifications` | Get email/new-message/campaign notification preferences and sound selection |
| PUT | `/api/settings/notifications` | Update organization or user notification settings |
| POST | `/api/settings/uploads-cleanup/run` | Trigger uploads cleanup and storage reconciliation manually, admin only |

## Organizations & Teams

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/organizations` | List organizations |
| POST | `/api/organizations` | Create organization and seed default quota configuration |
| DELETE | `/api/organizations/{id}` | Delete selected organization |
| GET | `/api/organizations/current` | Get current organization |
| GET | `/api/organizations/members` | List current org members |
| POST | `/api/organizations/members` | Add member if the tenant is still below the 5-user cap |
| PUT | `/api/organizations/members/{member_id}` | Update member role |
| DELETE | `/api/organizations/members/{member_id}` | Remove member |
| GET | `/api/teams` | List teams |
| POST | `/api/teams` | Create team |
| GET | `/api/teams/{id}` | Get team |
| PUT | `/api/teams/{id}` | Update team |
| DELETE | `/api/teams/{id}` | Delete team |
| GET | `/api/teams/{id}/members` | List team members |
| POST | `/api/teams/{id}/members` | Add team member |
| DELETE | `/api/teams/{id}/members/{member_user_id}` | Remove team member |

## User Management

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/users` | List users |
| POST | `/api/users` | Create user within the tenant member cap |
| GET | `/api/users/{id}` | Get user |
| PUT | `/api/users/{id}` | Update user |
| DELETE | `/api/users/{id}` | Delete user |
| GET | `/api/users/{id}/send-restrictions` | Get effective send permissions and read-only rules |
| PUT | `/api/users/{id}/send-restrictions` | Update user-level send overrides on top of role defaults |
| GET | `/api/users/{id}/contact-visibility` | Get contact scope (`all`, `instances_only`, `allowed_numbers`) and phone whitelist |
| PUT | `/api/users/{id}/contact-visibility` | Update allowed instances, phone whitelist, and mask-phone behavior overrides |
| GET | `/api/roles` | List roles |
| POST | `/api/roles` | Create role |
| GET | `/api/roles/{id}` | Get role |
| PUT | `/api/roles/{id}` | Update role |
| DELETE | `/api/roles/{id}` | Delete role |
| GET | `/api/permissions` | List permission catalog including CRUD/read-only actions, unclaimed chat policies, and visibility scopes |
| GET | `/api/api-keys` | List API keys |
| POST | `/api/api-keys` | Create API key with scoped permissions |
| DELETE | `/api/api-keys/{id}` | Delete API key |

## Contacts & Chats

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/contacts` | List contacts after applying role/user visibility scope and allowed-phone overrides |
| POST | `/api/contacts` | Create contact |
| GET | `/api/contacts/{id}` | Get contact |
| PUT | `/api/contacts/{id}` | Update contact |
| DELETE | `/api/contacts/{id}` | Delete contact |
| GET | `/api/contacts/export` | Export contacts to CSV with selectable columns like phone, name, WhatsApp account, tags, assigned user, and timestamps |
| POST | `/api/contacts/import` | Import contacts from CSV with optional update-on-duplicate behavior |
| POST | `/api/contacts/{id}/soft-delete` | Hide contact for current user |
| PUT | `/api/contacts/{id}/assign` | Assign contact to user/team |
| GET | `/api/contacts/{id}/collaborators` | List collaborators |
| POST | `/api/contacts/{id}/collaborators` | Invite collaborator |
| PUT | `/api/contacts/{id}/collaborators/{user_id}/accept` | Accept collaboration |
| PUT | `/api/contacts/{id}/collaborators/{user_id}/decline` | Decline collaboration |
| DELETE | `/api/contacts/{id}/collaborators/{user_id}` | Remove collaborator |
| PUT | `/api/contacts/{id}/tags` | Update tags |
| GET | `/api/chats` | List chats with `Assigned` / `Pending` split, filtered by role-based unclaimed visibility |
| POST | `/api/chats/direct` | Start direct chat from phone number + instance if sender is allowed to use that scope |
| GET | `/api/chats/closed` | List closed chats with filters for agent, instance, page size, and pagination |
| PUT | `/api/chats/{id}/claim` | Claim chat |
| PUT | `/api/chats/{id}/close` | Close chat |
| PUT | `/api/chats/{id}/reopen` | Reopen chat |
| PUT | `/api/chats/{id}/public` | Set public/private |
| PUT | `/api/chats/{id}/pin` | Pin or unpin chat in inbox |
| GET | `/api/chats/{id}/messages` | Get chat messages |
| GET | `/api/contacts/{id}/notes` | List notes |
| POST | `/api/contacts/{id}/notes` | Create note |
| PUT | `/api/contacts/{id}/notes/{note_id}` | Update note |
| DELETE | `/api/contacts/{id}/notes/{note_id}` | Delete note |

## Messaging

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/contacts/{id}/messages` | Get messages |
| POST | `/api/contacts/{id}/messages` | Send text message after permission checks, role gating, and provider-side typing simulation proportional to text length |
| POST | `/api/messages/media` | Send media/file message from picker or drag-drop without typing simulation after permission and storage quota checks |
| POST | `/api/contacts/{id}/typing` | Send UI-level typing indicator; separate from provider-side pre-send typing simulation |
| POST | `/api/contacts/{id}/messages/{message_id}/reaction` | Send reaction |
| POST | `/api/contacts/{id}/messages/{message_id}/revoke` | Revoke message |
| POST | `/api/messages/{id}/retry` | Retry failed outbound message |
| PUT | `/api/messages/{id}/read` | Mark message as read |
| GET | `/api/media/{message_id}` | Download or serve media from object storage or bounded disk cache |

## Notifications & Statuses

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/notifications` | List notification center items |
| PUT | `/api/notifications/{id}/read` | Mark notification as read |
| PUT | `/api/notifications/read-all` | Mark all notifications as read |
| GET | `/api/statuses` | List WhatsApp statuses for the status drawer |
| POST | `/api/statuses` | Create a new status |
| GET | `/api/statuses/{id}` | Open status detail |
| DELETE | `/api/statuses/{id}` | Remove or expire a status |

## Agent Analytics

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/analytics/agents/summary` | KPI cards like handled transfers, completed conversations, active conversations, average resolution time, average queue time, break time, and average rating |
| GET | `/api/analytics/agents/transfers` | Dataset for `Transfer Trends` |
| GET | `/api/analytics/agents/sources` | Dataset for `Conversation Sources` |
| GET | `/api/analytics/agents/comparison` | Agent comparison table or chart dataset |
| GET | `/api/analytics/agents/ratings` | Customer ratings table with agent, phone, contact, score, rated time, closing agent, rating message, and context messages |
| GET | `/api/analytics/agents/export` | Export the filtered analytics view to CSV |

## WhatsApp Instances (whatsmeow)

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/instances` | List instances |
| GET | `/api/instances/health` | Summary dataset for the health dashboard |
| POST | `/api/instances` | Create instance and reserve tenant/global slot if available |
| GET | `/api/instances/{id}` | Get instance |
| PUT | `/api/instances/{id}` | Update instance |
| PUT | `/api/instances/{id}/name` | Rename instance from `Edit Account Name` dialog |
| DELETE | `/api/instances/{id}` | Delete instance and release its reserved slot |
| GET | `/api/instances/{id}/health` | Get instance health |
| GET | `/api/instances/{id}/qr` | Get pairing QR code |
| POST | `/api/instances/{id}/connect` | Connect instance using an existing reservation |
| POST | `/api/instances/{id}/pair-phone` | Pair by phone/code |
| POST | `/api/instances/{id}/disconnect` | Disconnect instance while preserving reservation until delete or reassignment |
| POST | `/api/instances/{id}/reconnect` | Reconnect instance with guarded backoff policy |
| GET | `/api/instances/{id}/settings` | Load per-instance toggles and display settings |
| PUT | `/api/instances/{id}/settings` | Update auto-sync, media download, and source-tag settings |
| GET | `/api/instances/{id}/call-auto-reject` | Load call auto-reject policy |
| PUT | `/api/instances/{id}/call-auto-reject` | Update call auto-reject policy |
| GET | `/api/instances/{id}/auto-campaign` | Load auto-campaign configuration |
| PUT | `/api/instances/{id}/auto-campaign` | Update auto-campaign configuration |
| GET | `/api/instances/{id}/close-rating` | Load close-rating message configuration |
| PUT | `/api/instances/{id}/close-rating` | Update close-rating message configuration |
| GET | `/api/instances/{id}/assignment-reset` | Load assigned-chat reset policy |
| PUT | `/api/instances/{id}/assignment-reset` | Update assigned-chat reset policy |
| GET | `/api/instances/{id}/source-tag` | Load chat source tag settings |
| PUT | `/api/instances/{id}/source-tag` | Update custom label, display mode, and color |

## Operational Settings

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/canned-responses` | List canned responses |
| POST | `/api/canned-responses` | Create canned response |
| GET | `/api/canned-responses/{id}` | Get canned response |
| PUT | `/api/canned-responses/{id}` | Update canned response |
| DELETE | `/api/canned-responses/{id}` | Delete canned response |
| GET | `/api/tags` | List tags |
| POST | `/api/tags` | Create tag |
| PUT | `/api/tags/{id}` | Update tag |
| DELETE | `/api/tags/{id}` | Delete tag |
| GET | `/api/webhook-events` | List supported outbound webhook events |
| GET | `/api/webhooks` | List webhooks and their subscribed events |
| POST | `/api/webhooks` | Create webhook with event subscriptions, target URL, optional secret, and optional custom headers |
| GET | `/api/webhooks/{id}` | Get webhook |
| PUT | `/api/webhooks/{id}` | Update subscribed events, target URL, secret, headers, or activation state |
| DELETE | `/api/webhooks/{id}` | Delete webhook |
| POST | `/api/webhooks/{id}/test` | Test webhook |
| GET | `/api/custom-actions` | List custom actions |
| POST | `/api/custom-actions` | Create custom action |
| GET | `/api/custom-actions/{id}` | Get custom action |
| PUT | `/api/custom-actions/{id}` | Update custom action |
| DELETE | `/api/custom-actions/{id}` | Delete custom action |
| GET | `/api/settings/sso` | Get SSO settings |
| PUT | `/api/settings/sso/{provider}` | Update SSO provider |
| DELETE | `/api/settings/sso/{provider}` | Delete SSO provider |
| GET | `/api/license/bootstrap` | Public bootstrap payload with license state, HWID, tier, kind, expiry/grace status, quota overages, and per-organization usage snapshot |
| POST | `/api/license/activate` | Install or renew a signed offline security key bound to the current host HWID |

## License Enforcement Notes

- عند غياب الترخيص أو فساد الـ token أو عدم تطابق الـ `HWID`، الـ backend يعيد `423 Locked` مع `activate_url`.
- عند وجود `quota overage` عام، الـ backend يعيد `423 Locked` مع `cleanup_url=/license-cleanup`.
- عند فحص resource محدد أثناء create/upload/delete-sensitive flows، يمكن أن يعيد الـ backend `402 Payment Required` مع:
  - `resource`
  - `current`
  - `limit`
  - `over_quota`

## Vendor License Studio API

هذه ليست ضمن API التطبيق الرئيسي، لكنها جزء من منظومة الترخيص التي يعتمد عليها المنتج.

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| POST | `/auth/login` | Authenticate to the vendor-only license console |
| POST | `/auth/logout` | End the vendor console session |
| GET | `/guard/healthz` | Guard health check |
| GET | `/api/bootstrap` | Studio bootstrap with defaults, registry summary, known KIDs, and storage paths |
| POST | `/api/issue` | Issue a new offline license token from uploaded private key + HWID + entitlements |
| POST | `/api/verify` | Verify a token and report `valid_tracked`, `valid_untracked`, or `invalid` |
| GET | `/api/licenses` | List local registry entries with filters for HWID, tier, kind, and status |
| GET | `/api/licenses/{id}/token` | Retrieve the stored token for copy/preview/reissue flows |
| DELETE | `/api/v1/licenses/{id}` | Remove a license from the vendor registry |

## Retained Product Modules Confirmed in Navigation

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| GET | `/api/dashboard/*` | Dashboard summaries and cards |
| GET | `/api/chatbot/*` | Chatbot management surface retained in product |
| GET | `/api/analytics/agents*` | Agent analytics views |
| GET | `/api/campaigns*` | Campaign listing and execution |

## Removed Or Still Unverified

لم يؤكد التدقيق الحالي لهذه الشاشات هذه المسارات، لذلك لا تبقى ضمن الخطة الأساسية إلا إذا ظهرت في تدقيق مخصص:

- `/api/meta/*`
- `/api/templates/*`
- `/api/widgets/*`
- `/api/chatbot/keywords/*`
- `/api/chatbot/ai/*`
- `/api/transfers/*`
- `/api/sla/*`
