# Whatomate Database Schema Report

هذه الوثيقة تصف **المخطط بعد تدقيق الواجهة الحية** ليتوافق مع:

- `Go + chi + pgx + sqlc`
- `PostgreSQL + Redis`
- `whatsmeow`
- surfaces المؤكدة في واجهات `/chat` و`/analytics/agents` و`/settings/license` و`/settings/instances` و`/settings/contacts` و`/settings/closed-chats`

أهم فرق عن النسخة السابقة: الواجهة الحية أكدت أن المخطط يجب أن يحتفظ بكيانات للحالات والإشعارات،
وأن `chatbot` ما زالت موجودة على الأقل كطبقة إعدادات تؤثر على `Contact Info panel`،
وأن `whatsapp instances` تحتاج طبقة إعدادات تشغيلية مستقلة لكل instance.
كما أضيفت طبقة `quota/capacity` صريحة لإدارة حد المستخدمين، الـ slots، حدود التخزين،
والـ backpressure على مستوى كل `organization`.

## 1. Master Table Register

| Domain | Table Name | Logical Model | Multi-tenant | Soft Delete | Description |
| :--- | :--- | :--- | :---: | :---: | :--- |
| **Identity** | `organizations` | `Organization` | N/A | Yes | Root tenant entity |
| **Identity** | `organization_configs` | `OrganizationConfig` | Yes | No | Tenant-level limits + general settings + admin-only uploads cleanup + quota counters |
| **Capacity** | `slot_inventory` | `SlotInventory` | No | No | Platform-wide slot pool for scarce resources such as WhatsApp instance capacity |
| **Capacity** | `organization_slot_allocations` | `OrganizationSlotAllocation` | Yes | No | Tenant reservations and in-use counters against the global slot pool |
| **Identity** | `users` | `User` | No | Yes | Platform users with primary organization context in SaaS mode |
| **Identity** | `user_organizations` | `UserOrganization` | Yes | No | User to organization link; usually a single active membership for customer users |
| **Identity** | `permissions` | `Permission` | No | No | Permission catalog with CRUD/read-only actions and scope flags |
| **Identity** | `custom_roles` | `CustomRole` | Yes | Yes | Tenant roles |
| **Identity** | `role_permissions` | `RolePermission` | No | No | Role-permission mapping |
| **Identity** | `api_keys` | `APIKey` | Yes | No | Programmatic access with scoped permissions |
| **Identity** | `sso_providers` | `SSOProvider` | Yes | No | SSO configuration |
| **Operations** | `teams` | `Team` | Yes | Yes | Agent teams |
| **Operations** | `team_members` | `TeamMember` | No | No | Team membership |
| **Operations** | `user_availability_logs` | `UserAvailabilityLog` | Yes | No | Availability history |
| **Operations** | `user_contact_visibility_rules` | `UserContactVisibilityRule` | Yes | No | Role/user scoped contact visibility and allowed-phone overrides |
| **Operations** | `user_notifications` | `UserNotification` | Yes | No | Notification center and unread inbox |
| **Channels** | `whatsapp_instances` | `WhatsAppInstance` | Yes | Yes | whatsmeow session instances |
| **Channels** | `whatsapp_instance_settings` | `WhatsAppInstanceSettings` | Yes | No | Per-instance sync, media, and source-tag settings |
| **Channels** | `whatsapp_instance_health_snapshots` | `WhatsAppInstanceHealthSnapshot` | Yes | No | Health dashboard and time-series metrics |
| **Channels** | `instance_call_policies` | `InstanceCallPolicy` | Yes | No | Auto-reject call rules |
| **Channels** | `instance_auto_campaigns` | `InstanceAutoCampaign` | Yes | No | Per-instance recurring campaign automation |
| **Channels** | `instance_rating_settings` | `InstanceRatingSettings` | Yes | No | Chat close rating message templates |
| **Channels** | `instance_assignment_resets` | `InstanceAssignmentReset` | Yes | No | Daily reset rules for assigned chats |
| **Channels** | `instance_notifications` | `InstanceNotification` | Yes | No | Instance alerts |
| **Channels** | `whatsapp_statuses` | `WhatsAppStatus` | Yes | Yes | Status drawer items and published statuses |
| **Messaging** | `contacts` | `Contact` | Yes | Yes | Customer conversations |
| **Messaging** | `messages` | `Message` | Yes | Yes | Chat history |
| **Messaging** | `media_assets` | `MediaAsset` | Yes | No | Media metadata + storage accounting anchor |
| **Messaging** | `tags` | `Tag` | Yes | No | Contact labels |
| **Messaging** | `contact_user_deletions` | `ContactUserDeletion` | Yes | No | Per-user hide state |
| **Messaging** | `contact_collaborators` | `ContactCollaborator` | Yes | No | Shared chat access |
| **Messaging** | `conversation_notes` | `ConversationNote` | Yes | Yes | Internal notes |
| **Messaging** | `chat_closure_ratings` | `ChatClosureRating` | Yes | No | Post-close ratings |
| **Operations** | `canned_responses` | `CannedResponse` | Yes | Yes | Quick replies |
| **Chatbot** | `chatbot_flows` | `ChatbotFlow` | Yes | Yes | Flow config retained because contact info panel references chatbot flow settings |
| **Extensibility** | `webhooks` | `Webhook` | Yes | No | Outbound event delivery with optional secret and custom headers |
| **Extensibility** | `custom_actions` | `CustomAction` | Yes | No | UI-triggered integrations |
| **Licensing** | `license_records` | `LicenseRecord` | No | No | Installation entitlements |
| **Licensing** | `license_events` | `LicenseEvent` | No | No | License audit trail |

## 2. Core Table Notes

### `organizations`

- الجذر الأساسي لـ multi-tenancy.
- في SaaS mode الافتراضي، كل شركة تمتلك `organization` واحدة فقط.
- الواجهة الحية أكدت وجود org switcher في الشريط الجانبي، لذلك يجب أن يدعم:
  - إنشاء المنظمة
  - حذف المنظمة المحددة
  - تبديل السياق الفوري للمستخدم

### `organization_configs`

- تخزن الإعدادات التنظيمية العامة والحدود التشغيلية مثل:
  - default timezone
  - default locale
  - supported locale codes
  - date format
  - mask phone numbers
  - `max_users`
  - `max_queue_size`
  - `max_concurrent_jobs`
  - `max_whatsapp_instances`
  - `storage_quota_bytes`
  - `storage_used_bytes`
  - `tenant_status`
  - `queue_backpressure_mode`
  - uploads cleanup retention/hour/timezone
- `Uploads Cleanup` يجب أن تبقى مقيدة بصلاحية admin أو permission مخصصة مثل:
  - `settings.uploads_cleanup.manage`
- هذه القراءة يجب أن تكون cheap لأن معظم write paths ستتحقق منها قبل التنفيذ.

### `users`

- مستخدم النظام الأساسي.
- `organization_id` يمثل الـ primary org للمستخدم العادي في نمط SaaS الافتراضي.
- الحقول التشغيلية يجب أن تدعم:
  - `availability_status`
  - `settings.locale`
  - `settings.language`
  - `settings.theme`
  - `settings.appearance.color_mode` بقيم `light|dark|system`
  - `settings.appearance.theme_style` كمؤشر إلى Tailwind preset
  - `settings.sidebar.expand_on_hover`
  - `settings.sidebar.is_pinned`
  - تفضيلات notifications الشخصية عند الحاجة
- قواميس `i18n` نفسها وtokens الخاصة بالـ Tailwind تبقى في frontend assets،
  بينما قاعدة البيانات تخزن `locale` و`theme_style` فقط.

### `user_organizations`

- في نمط SaaS الافتراضي، العضو العميل يملك عضوية فعالة واحدة فقط وdefault org واحدة فقط.
- الجدول يبقى لدعم:
  - `super admin`
  - `support`
  - حالات enterprise الخاصة أو migration

### `permissions`

- الكتالوج يجب أن يكون action-based لا screen-based فقط.
- كل capability يجب أن تدعم على الأقل:
  - `read`
  - `create`
  - `update`
  - `delete`
- بعض الوحدات تحتاج مفاتيح إضافية مستقلة:
  - `messages.send`
  - `chats.unclaimed.view`
  - `chats.unclaimed.send`
  - `contacts.scope.all`
  - `contacts.scope.instance_only`
  - `contacts.scope.allowed_numbers`

### `custom_roles`

- كل role تربط مجموعة permissions قابلة للتركيب.
- الواجهة المطلوبة ليست مجرد checkbox عام، بل matrix تسمح بجعل كل وحدة:
  - `create`
  - `update`
  - `delete`
  - `read_only`
- `read_only` تعني بقاء `read` فعالاً مع تعطيل الكتابة والإرسال بحسب domain.

### `api_keys`

- تمثل مفاتيح وصول قابلة للإنشاء والحذف من الواجهة.
- الحقول المتوقعة:
  - `name`
  - `key_prefix`
  - `secret_hash`
  - `scopes JSONB` أو ربط بدور service role
  - `created_by_user_id`
  - `last_used_at`
  - `revoked_at`

### `user_availability_logs`

- تسجل انتقالات المستخدم بين:
  - `available`
  - `unavailable`
  - `busy`
- الحقول المتوقعة:
  - `status`
  - `changed_at`
  - `changed_by_user_id` أو `session_id` عند الحاجة

### `slot_inventory`

- مخزون المنصة للموارد المحدودة مثل slots الخاصة بالـ `whatsapp instances`.
- الحجز يتم transactionally مع row locking حتى لا يحدث over-allocation.

### `organization_slot_allocations`

- يربط كل tenant بمخزون الـ slots العام.
- الحقول المتوقعة:
  - `allocated_slots`
  - `used_slots`
  - `last_reconciled_at`

### `whatsapp_instances`

- تمثل جلسة `whatsmeow` مرتبطة بمنظمة.
- تظهر أيضاً كمرشح في chat filter وكخيار إلزامي في `Start New Chat`.
- create / delete يجب أن يحدّثا استهلاك الـ slots على مستوى المنظمة.
- الواجهة الحية أكدت حقولاً تشغيلية إضافية يجب أن تكون متاحة مباشرة:
  - `phone_number`
  - `jid`
  - `status`
  - last-known counters أو latest health summary

### `whatsapp_instance_settings`

- تغطي الإعدادات السريعة الظاهرة على البطاقة:
  - `auto_sync_history`
  - `auto_download_incoming_media`
  - `source_tag_label`
  - `source_tag_display_mode`
  - `source_tag_color`

### `whatsapp_instance_health_snapshots`

- مطلوبة لواجهة `Health Dashboard`.
- يجب أن تدعم على الأقل:
  - `status`
  - `uptime_seconds`
  - `queue_depth`
  - `sent_today`
  - `received_today`
  - `failed_today`
  - `error_rate`
  - `observed_at`

### `instance_call_policies`

- تمثل إعدادات `Call Auto-Reject Settings`.
- الحقول المتوقعة:
  - `enabled`
  - `reject_individual_calls`
  - `reject_group_calls`
  - `reply_mode`
  - `schedule_mode`
  - `emergency_bypass_contacts`

### `instance_auto_campaigns`

- الواجهة الحية أثبتت أن الحملات ليست فقط module مستقل، بل توجد أتمتة حملات على مستوى instance.
- الحقول المتوقعة:
  - `enabled`
  - `campaign_name_prefix`
  - `schedule_every_days`
  - `delay_from_minutes`
  - `delay_to_minutes`
  - `campaign_status`
  - `message_body`
  - `media_asset_id`
  - `last_evaluated_at`
  - `next_evaluation_at`

### `instance_rating_settings`

- تمثل `Chat Close Rating Settings`.
- الحقول المتوقعة:
  - `enabled`
  - `follow_up_window_minutes`
  - `template_ar`
  - `template_en`
  - `template_es`

### `instance_assignment_resets`

- تمثل `Assigned Chat Reset`.
- الحقول المتوقعة:
  - `enabled`
  - `schedule_mode`
  - `timezone`
  - `last_reset_at`
  - `next_reset_at`

### `whatsapp_statuses`

- تمثل منشورات الحالات المعروضة في `Statuses drawer`.
- الحقول المتوقعة:
  - `instance_id`
  - `created_by_user_id`
  - `media_asset_id`
  - `body`
  - `published_at`
  - `expires_at`
  - `status`
  - `metadata JSONB`

### `user_notifications`

- تمثل عناصر مركز الإشعارات.
- يجب أن تدعم:
  - unread message notifications
  - system notifications
  - `is_read`
  - `read_at`
  - payload مرن لفتح الجهة أو المحادثة المرتبطة

### `user_contact_visibility_rules`

- تحل التباين بين role defaults وبين الاستثناءات per-user.
- الحقول المتوقعة:
  - `user_id`
  - `scope_mode` = `all_contacts` | `instances_only` | `allowed_numbers_only` | `instances_plus_allowed_numbers`
  - `allowed_instance_ids JSONB`
  - `allowed_phone_numbers JSONB`
  - `inherit_role_scope`
  - `can_view_unmasked_phone`

### `contacts`

- تمثل جهة اتصال واحدة داخل المنظمة.
- الحقول الأهم:
  - `phone_number`
  - `name`
  - `assigned_user_id`
  - `instance_id`
  - `status`
  - `is_public`
  - `is_read`
  - `instance_source_label`
  - `last_message_preview`
  - `last_message_at`
  - `last_inbound_at`
  - `closed_at`
  - `tags JSONB`
  - `metadata JSONB`
- صفحة `Contacts` أكدت الحاجة إلى قراءة cheap وسريعة لآخر preview مع دعم:
  - search
  - filter by instance
  - open chat
  - add/edit/delete
  - import/export CSV
- وجود `Assigned User ID` ضمن التصدير يعني أن الإسناد جزء من النموذج حتى لو لم يظهر كزر مباشر داخل الصفحة.
- الرؤية لا تعتمد فقط على `organization_id`، بل أيضاً على:
  - `instance_id`
  - صلاحيات الدور
  - `user_contact_visibility_rules.allowed_phone_numbers`

### `messages`

- السجل الكامل للمحادثة.
- يجب أن يدعم:
  - `status = failed` مع سبب فشل
  - `revoke`
  - retry metadata
  - media references

### `media_assets`

- تمثل metadata layer وaccounting layer للتخزين.
- `file_size` يجب أن يكون موثوقاً لأنه يدخل مباشرة في حساب `storage_used_bytes`.
- التخزين الموصى به هو `object storage`، مع اعتبار الديسك المحلي cache أو volume محدود فقط.
- تمثل أيضاً مرجع كل upload صادر من picker أو drag & drop داخل chat composer.

### `conversation_notes`

- ملاحظات داخلية خاصة بالوكلاء، منفصلة عن الرسائل الفعلية.

### `contact_collaborators`

- تدعم واجهة `Invite Collaborator`.
- الحالات المتوقعة:
  - `invited`
  - `accepted`
  - `declined`

### `chat_closure_ratings`

- لم تعد مجرد كيان جانبي؛ شاشة `Agent Analytics` أكدت أن التقييمات تدخل مباشرة في التقارير.
- الحقول المتوقعة:
  - `contact_id`
  - `score`
  - `comment`
  - `closing_user_id`
  - `rated_at`
  - `rating_message_snapshot`
  - `context_messages_snapshot`
- شاشة `Closed Chats` أكدت أيضاً الحاجة إلى الاحتفاظ بعلاقة واضحة بين المحادثة المغلقة، من أغلقها، ووقت الإغلاق/التقييم.

### `chatbot_flows`

- لم يعد ممكناً حذف هذا المجال بالكامل من التخطيط لأن `Contact Info panel` تعرض رسالة:
  `Configure panel display in the chatbot flow settings.`
- في الحد الأدنى يجب أن تبقى طبقة إعدادات عرض البيانات محفوظة في هذا المجال أو في ما يعادله.

### `webhooks`

- تمثل outbound subscriptions على أي event مسموح من النظام.
- الحقول المتوقعة:
  - `name`
  - `target_url`
  - `subscribed_events TEXT[]`
  - `secret_encrypted` nullable
  - `custom_headers JSONB` nullable
  - `is_active`
  - `last_test_at`
  - `last_delivery_at`

### `license_records`

- تمثل **الترخيص النشط المثبت على هذا الـ deployment**، لا سجل إصدار الـ vendor.
- من الكود المحلي تأكد أن السجل يجب أن يحفظ:
  - `activation_token_encrypted`
  - `license_family_id`
  - `license_id`
  - `revision`
  - `key_id`
  - `issuer`
  - `audience`
  - `product`
  - `hwid_full`
  - `hwid_hash`
  - `tier`
  - `license_kind`
  - `trial_days`
  - `max_organizations`
  - `max_users_per_org`
  - `max_whatsapp_endpoints_per_org`
  - `max_workers`
  - `max_workers_per_org`
  - `max_storage_bytes_per_org`
  - `status`
  - `overages JSONB`
  - `issued_at`
  - `not_before`
  - `expires_at`
  - `grace_deadline`
  - `last_seen_at`
  - `activated_at`
  - `integrity_hmac`
- الواجهة الحية لصفحة `License` أكدت أن state لا تقتصر على active/expired فقط، بل يجب أن تدعم:
  - `disabled`
  - `unlicensed`
  - `active`
  - `grace`
  - `locked`
- الـ page نفسها تعرض usage snapshot لكل منظمة، لذلك السجل يجب أن يربط entitlements بعمليات حساب usage الحالية.

### `license_events`

- تمثل audit trail append-only لأحداث:
  - `license_activated`
  - state transitions
  - enforcement outcomes
- من الكود المحلي تأكد أن الجدول يجب أن يكون **مستقلاً** عن `license_records` لأن التطبيق قد يستبدل السجل النشط بالكامل عند التفعيل أو التجديد.
- الحقول المتوقعة:
  - `event_type`
  - `reason`
  - `status`
  - `license_family_id`
  - `license_id`
  - `hwid_hash`
  - `details JSONB`
  - `created_at`

### Vendor License Registry

- تحليل `whatomate-license-studio` أكد أن هناك طبقة vendor منفصلة عن DB التطبيق الرئيسي.
- هذه الطبقة تحفظ:
  - registry entries في `registry.json`
  - key ring في `keyring.json`
  - metadata/audit/session data في SQLite مثل:
    - `customer_meta`
    - `audit_events`
    - `revocations`
    - `hwid_transfers`
    - `key_ring_meta`
    - `sessions`
- لا يجب خلط هذه الجداول مع schema تطبيق Whatomate نفسه؛ هي تخص console الإصدار والإدارة لدى الـ vendor.

## 3. Multi-tenancy & Soft Delete

### Multi-tenancy Strategy

- الآلية الأساسية: `organization_id`.
- كل شركة عميل = `organization` واحدة فقط.
- المستخدم العادي في SaaS mode يملك منظمة فعالة واحدة فقط؛ org switcher يبقى للحسابات المرتفعة الصلاحية أو الحالات الخاصة.
- كل drawer أو dialog في الواجهة الحية يعتمد أيضاً على المنظمة الحالية:
  - chat list
  - notifications
  - statuses
  - users available for assignment/collaboration
- طبقة الحدود التشغيلية تقرأ من:
  - `organization_configs`
  - `organization_slot_allocations`

### Tables Usually Scoped By Organization

- `organization_configs`
- `organization_slot_allocations`
- `user_organizations`
- `custom_roles`
- `api_keys`
- `sso_providers`
- `teams`
- `user_notifications`
- `user_contact_visibility_rules`
- `whatsapp_instances`
- `whatsapp_instance_settings`
- `whatsapp_instance_health_snapshots`
- `instance_call_policies`
- `instance_auto_campaigns`
- `instance_rating_settings`
- `instance_assignment_resets`
- `instance_notifications`
- `whatsapp_statuses`
- `contacts`
- `messages`
- `media_assets`
- `tags`
- `contact_user_deletions`
- `contact_collaborators`
- `conversation_notes`
- `chat_closure_ratings`
- `canned_responses`
- `chatbot_flows`
- `webhooks`
- `custom_actions`
- `user_availability_logs`

### Tables Not Scoped By Organization

- `organizations`
- `slot_inventory`
- `users`
- `permissions`
- `role_permissions`
- `team_members`
- `license_records`
- `license_events`

### Capacity Control Defaults

- العميل SaaS العادي يعمل بحد `max_users = 5`.
- الحصة الافتراضية للتخزين هي `5 GiB` عبر `storage_quota_bytes = 5368709120`.
- حد الـ WhatsApp capacity يقرأ من `organization_configs.max_whatsapp_instances` ويجب ألا يتجاوز ما تم تخصيصه في `organization_slot_allocations`.
- عند تجاوز الحدود، المطلوب هو degradation موجه مثل:
  - رفض upload جديد
  - رفض إنشاء instance جديد
  - `read_only` mode
  - تأجيل jobs
  بدل استنزاف السيرفر أو إسقاط المنصة كاملة.

## 4. Key Indexes & Constraints

- `organizations.slug` unique
- `users.email` unique
- `users (organization_id, is_active)` index
- `custom_roles (organization_id, name)` unique
- `user_organizations (user_id, organization_id)` unique
- `user_organizations (user_id)` unique when `is_default = true`
- `api_keys (organization_id, key_prefix)` unique
- `teams (organization_id, name)` unique
- `slot_inventory.resource_type` unique
- `organization_slot_allocations (organization_id, slot_inventory_id)` unique
- `organization_slot_allocations (slot_inventory_id, organization_id)` index
- `contacts (organization_id, phone_number, instance_id)` unique
- `user_contact_visibility_rules (organization_id, user_id)` unique
- `user_notifications (user_id, is_read, created_at DESC)` index
- `media_assets (organization_id, created_at DESC)` index
- `whatsapp_instance_settings (instance_id)` unique
- `instance_call_policies (instance_id)` unique
- `instance_auto_campaigns (instance_id)` unique
- `instance_rating_settings (instance_id)` unique
- `instance_assignment_resets (instance_id)` unique
- `whatsapp_instance_health_snapshots (instance_id, observed_at DESC)` index
- `whatsapp_statuses (organization_id, published_at DESC)` index
- `webhooks (organization_id, name)` unique
- فهارس `GIN` على `contacts.tags` وحقول `metadata JSONB` و`webhooks.custom_headers`
  و`user_contact_visibility_rules.allowed_phone_numbers` عند الحاجة

### Permission Notes

- `permissions` يجب أن تغطي على الأقل صلاحيات من نوع:
  - `settings.manage`
  - `settings.uploads_cleanup.manage`
  - `instances.manage`
  - `contacts.read`
  - `contacts.create`
  - `contacts.update`
  - `contacts.delete`
  - `messages.read`
  - `messages.send`
  - `notes.read`
  - `notes.create`
  - `notes.update`
  - `notes.delete`
  - `chats.unclaimed.view`
  - `chats.unclaimed.send`
  - `contacts.scope.all`
  - `contacts.scope.instance_only`
  - `contacts.scope.allowed_numbers`
  - `api_keys.create`
  - `api_keys.delete`
  - `webhooks.create`
  - `webhooks.delete`
- في الحد الأدنى، `settings.uploads_cleanup.manage` تبقى محصورة على org admin حتى لو لم تُعرض كمفتاح مستقل في الواجهة الأولى.

## 5. Removed Or Still Unverified Domains

هذه الجداول لا تزال خارج النطاق المؤكد أو غير مثبتة من تدقيق `/chat` الحالي:

- `whatsapp_accounts`
- `chatbot_sessions`
- `keyword_rules`
- `chatbot_session_messages`
- `ai_contexts`
- `agent_transfers`
- `notification_rules`
- `templates`
- `catalogs`
- `catalog_products`
- `widgets`

ملاحظة:

- `campaigns` و`analytics` بقيتا ضمن الخطة لأنهما ظهرتا في الشريط الجانبي، لكن جداولهما الدقيقة مؤجلة إلى تدقيق مخصص ولا تُعامل هنا كجداول محذوفة.
