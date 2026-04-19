# Whatomate Business Logic Documentation

هذه الوثيقة تصف المنطق التشغيلي **بعد تدقيق الواجهة الحية**.

الواجهة الحالية أكدت أن المنتج يحتاج منطقاً لتشغيل:

- inbox للمحادثات
- notifications center
- statuses drawer
- start new chat
- assignment / notes / collaborators / tags / contact info
- settings hub
- contacts management وclosed chats management
- whatsapp instance cards + health dashboard + per-instance policies
- tenant quotas + slot allocation + storage limits + graceful backpressure
- offline licensing + quota cleanup mode + worker caps
- role-based access matrix with read-only / send / unclaimed visibility / contact scope
- drag-and-drop media upload + outbound webhooks
- text-message typing simulation before send
- dashboard / chatbot / analytics / campaigns كمسارات باقية في المنتج

مع بقاء `Meta Cloud API` خارج النطاق المؤكد.

## 1. Inbound Message Flow

**Path**: `whatsmeow Event -> Persistence -> Chat UI`

1. `whatsmeow` يستقبل الحدث.
2. يتم التطبيع إلى payload داخلي موحد.
3. يتم إيجاد أو إنشاء contact حسب `organization_id + phone_number + instance_id`.
4. قبل بث contact في الواجهة، يتم تطبيق visibility filters الخاصة بالدور والمستخدم.
5. يتم منع التكرار عبر `whatsapp_message_id`.
6. يتم حفظ الرسالة وتحديث contact.
7. يتم إنشاء notification عند الحاجة.
8. يتم بث `new_message` و`contact_update` و`notification`.

## 2. Outbound Message Flow

**Path**: `User Sends Message -> API -> whatsmeow -> UI Update`

1. التحقق من صلاحية المستخدم على المحادثة وعلى الإرسال تحديداً:
   - contact visibility
   - `read_only`
   - `chats.unclaimed.send`
2. إذا لم يملك المستخدم حق الإرسال، ترفض العملية مع بقاء القراءة متاحة عند الحاجة.
3. إنشاء رسالة بحالة `pending`.
4. إذا كانت الرسالة نصية، يحسب backend مدة كتابة تقريبية متناسبة مع طول النص.
5. يرسل backend `typing/presence` عبر `whatsmeow` قبل الإرسال.
6. ينتظر المدة المحسوبة ضمن حدود دنيا وعليا معقولة ثم يرسل النص.
7. إذا كانت الرسالة ملفاً أو وسائط من picker أو drag & drop، يتم تجاوز typing simulation بالكامل.
8. تحديث الحالة إلى `sent` أو `failed`.
9. عند الفشل، تخزين سبب الفشل لاستخدام `Retry sending`.
10. بث `new_message` أو `status_update`.

## 3. Direct Chat Bootstrap

**Path**: `Start New Chat Dialog -> Contact Bootstrap -> Open Conversation`

1. المستخدم يفتح `Start New Chat`.
2. يحدد:
   - phone number
   - profile name اختياري
   - WhatsApp instance
3. backend ينشئ أو يستعيد contact.
4. backend يربط contact بالـ instance المختار.
5. الواجهة تنتقل مباشرة إلى `/chat/[contactId]`.

## 4. Chat Workspace Flow

**Path**: `Inbox -> Select Chat -> Collaborate`

1. inbox يعرض `Assigned` و`Pending` فقط إذا كان للمستخدم حق `chats.unclaimed.view`.
2. يمكن البحث والتصفية حسب instance ونوع المحادثة والتاغات ضمن scope المسموح للمستخدم.
3. المستخدم يختار المحادثة.
4. من الترويسة يمكن:
   - assign
   - pin
   - فتح notes
   - فتح contact info
5. من side panels يمكن:
   - كتابة note
   - إضافة collaborator
   - إضافة tags
   - قراءة البيانات العامة للجهة
6. message composer يدعم drag & drop للملفات مع validation قبل الإرسال.

## 5. Notifications Center

**Path**: `New Event -> User Notification -> Notification Bell`

1. أي inbound message أو event تشغيلي مهم ينشئ notification.
2. notification ترتبط بمستخدم أو بمنظمة مع audience محدد.
3. الجرس يعرض unread count.
4. dialog الإشعارات يعرض:
   - الرسائل غير المقروءة
   - التنبيهات النظامية
   - `Mark all as read`

## 6. Status Drawer

**Path**: `Status Publish -> Status Feed -> Drawer`

1. المستخدم يفتح drawer الحالات.
2. يمكن إنشاء status جديد عبر `Add status`.
3. feed يعرض الحالات بحسب الجهة وinstance.
4. status item يرتبط بـ media أو نص ويخضع لصلاحيات المنظمة والinstance.

## 7. Contact Info Panel

**Path**: `Contact -> General Data + Operational Metadata`

1. panel يعرض:
   - الاسم
   - الهاتف الخام أو masked حسب الصلاحية
   - tags
   - collaborators
   - البيانات العامة مثل avatar sync
2. panel قد يعرض data blocks إضافية ناتجة من `chatbot flow settings`.
3. لذلك chatbot لم تعد قابلة للحذف من الخطة كلياً، حتى لو تأجل تدقيقها التفصيلي.

## 8. User Context Flow

**Path**: `Sidebar -> Organization/User Context -> Scoped Queries`

1. المستخدم العميل في SaaS mode يعمل عادة داخل org واحدة فقط.
2. org switcher يبقى متاحاً للحسابات الداخلية أو أي مستخدم لديه أكثر من عضوية مصرح بها.
3. backend يغيّر `organization_id` النشط فقط عندما يملك المستخدم أكثر من منظمة فعالة.
4. user menu يتحكم في:
   - availability status (`available`, `unavailable`, `busy`)
   - theme mode (`light`, `dark`, `system`)
   - Tailwind theme preset
   - language / locale المحمّل من JSON dictionaries
   - profile
   - logout
5. كل query في chat, notifications, statuses يعتمد على هذا السياق.
6. sidebar يتسع عند hover ويمكن pin لتثبيت حالته.

## 9. Retained Adjacent Modules

هذه الوحدات ظهرت في التنقل الحي، لذلك تبقى في الخطة:

- `Dashboard`
- `Chatbot`
- `Agent Analytics`
- `Campaigns`

لكن تفاصيلها الداخلية تحتاج audit منفصل قبل تثبيت منطقها النهائي.

## 10. Settings Hub Flow

**Path**: `Settings Navigation -> Tabbed Settings -> Save / Run Cleanup`

1. المستخدم يدخل `/settings`.
2. تظهر شجرة إعدادات جانبية مستقلة عن التنقل الرئيسي.
3. صفحة `General` تحتوي:
   - organization name
   - slug
   - timezone
   - date format
   - language / default locale
   - mask phone numbers
   - usage summary for members, slots, storage, and tenant mode
4. الواجهة تحمّل ملفات `i18n JSON` بحسب `locale` المختار مع fallback واضح.
5. تبويب `Appearance` يحفظ color mode (`light`, `dark`, `system`) وTailwind theme preset على مستوى المستخدم.
6. تبويب `Chat` يحفظ media grouping, sidebar contact view, hover-expand, pin state, background, وإظهار أزرار print/download.
7. تبويب `Notifications` يحفظ email notifications, new-message alerts, sound choice, وcampaign updates.
8. يوجد قسم `Uploads Cleanup` بجدولة retention + daily hour.
9. `Uploads Cleanup` و`Run Cleanup Now` يجب أن يكونا متاحين للأدمن فقط.
10. واجهة الإعدادات تحتاج surface واضحاً لعرض `remaining quota` قبل أن يفشل المستخدم في عملية create أو upload.
11. زر `Run Cleanup Now` يطلق job تشغيلي فوري.

## 11. WhatsApp Instance Catalog Flow

**Path**: `Settings > WhatsApp -> Card List -> Instance Action`

1. backend يعرض instances كبطاقات.
2. كل بطاقة تحتوي:
   - status
   - phone number
   - JID pairing state
   - uptime / queue / sent-received / error rate
3. المستخدم يمكنه:
   - `Add Account`
   - `Edit instance name`
   - `Delete instance`
   - `Disconnect`
   - `Connect / Scan QR`
4. create أو activate لا يكتملان إلا إذا كان لدى المنظمة slot متاح ضمن حصة tenant والمخزون العام.
5. disconnected instances تبقى ظاهرة مع حالة `Not paired` ويمكن أن تبقي slot محجوزاً حتى يتم حذفها أو إعادة تخصيصها إدارياً.

## 12. Instance Policy Flows

**Path**: `Instance Card -> Config Dialog -> Save`

1. `Auto-sync history` و`Auto-download incoming media` تحفظ كإعدادات سريعة per-instance.
2. `Call auto-reject` يضبط:
   - individual vs group calls
   - reply mode
   - schedule
   - bypass contacts
3. `Auto campaign` ينشئ cycle دوري لتوليد campaigns drafts أو campaigns فعالة لهذا instance.
4. `Chat Close Rating Settings` تضبط follow-up window وtemplates متعددة اللغات.
5. `Assigned Chat Reset` يعيد المحادثات المسندة إلى `pending` وفق schedule تنظيمي.
6. `Chat Source Tag` يحدد كيف يظهر مصدر المحادثة في inbox عبر label وdisplay mode وcolor.

## 13. Instance Health Dashboard

**Path**: `Health Dashboard -> Metrics Snapshot -> Refresh`

1. النظام يجمع snapshot دوري لكل instance.
2. dashboard تعرض:
   - uptime
   - sent today
   - received today
   - failed today
   - error rate
   - queue depth
3. زر `Refresh` يعيد طلب آخر snapshot أو يشغل refresh خفيف.

## 14. Tenant Quota & Admission Flow

**Path**: `Provision Tenant -> Seed Limits -> Operate Safely`

1. عند إنشاء `organization`، backend ينشئ `organization_configs` بالقيم الأساسية للخطة مثل:
   - `max_users = 5`
   - `storage_quota_bytes = 5 GiB`
   - `max_whatsapp_instances` حسب الباقة أو التخصيص
   - `max_concurrent_jobs`
   - `max_queue_size`
   - `tenant_status = active`
2. العضو العميل العادي يحصل على default org واحدة فقط.
3. كل write path يستهلك موارد يجب أن يقرأ quotas الحالية قبل التنفيذ.
4. إذا تحولت المنظمة إلى `read_only` تبقى القراءة وعرض البيانات متاحين، لكن النظام يمنع العمليات التي تستهلك موارد جديدة.

## 15. Member Limit Flow

**Path**: `Invite Member -> Count Active Seats -> Accept / Reject`

1. دعوة أو إنشاء عضو جديد تتم داخل transaction.
2. backend يعدّ فقط الأعضاء النشطين وغير المحذوفين المرتبطين بالمنظمة.
3. إذا وصل العدد إلى `5` يتم رفض create/invite بخطأ واضح بدلاً من إدخال صف زائد.
4. الحسابات الداخلية غير المرتبطة كأعضاء tenant لا تدخل في هذا العد.

## 16. Slot Allocation Flow

**Path**: `Request Instance Capacity -> Lock Inventory -> Reserve / Reject`

1. المنصة تحتفظ بمخزون عام للـ slots عبر `slot_inventory`.
2. كل منظمة تملك reservation خاصاً عبر `organization_slot_allocations`.
3. عند create أو activate لـ instance جديد، backend ينفذ row lock على سجل المخزون والتخصيص قبل الحجز حتى لا يحدث overbooking.
4. إذا لم تتوفر slot في حصة المنظمة أو في المخزون العام، العملية تُرفض بينما بقية النظام يبقى سليماً.

## 17. Media Upload & Storage Quota Flow

**Path**: `Upload Media -> Check Quota -> Persist -> Reconcile`

1. `media_assets.file_size` هو المصدر الأساسي لحساب استهلاك التخزين.
2. backend يقارن `incoming_file_size + storage_used_bytes` مع `storage_quota_bytes` قبل كتابة الملف.
3. عند النجاح، يتم تحديث `storage_used_bytes` داخل transaction مرتبطة بتسجيل `media_assets`.
4. عند الحذف أو cleanup، يتم إنقاص الاستهلاك بدلاً من إعادة scan كامل للديسك في كل مرة.
5. job يومي للمراجعة يقوم بعمل reconciliation للتخزين ويصحح أي drift.
6. التخزين المفضل هو `object storage`، أما الديسك المحلي فيبقى cache أو volume محدوداً فقط.

## 18. Shared Worker Pool & Backpressure

**Path**: `Job Creation -> Shared Queue -> Tenant Guardrails`

1. لا يتم تخصيص worker أو process مستقل لكل tenant.
2. كل jobs الثقيلة تبقى async مثل:
   - media download
   - uploads cleanup
   - campaign evaluation
   - reconnect handling
3. التنفيذ يستخدم shared worker pool مع limits على مستوى المنظمة:
   - `max_concurrent_jobs`
   - `max_queue_size`
4. عند امتلاء queue، السلوك يتبع `queue_backpressure_mode` مثل:
   - defer
   - reject new jobs
   - pause
5. reconnect logic للـ WhatsApp instances يجب أن يستخدم exponential backoff + jitter لتجنب reconnect storms بعد restart.

## 19. Role & Permission Resolution Flow

**Path**: `Auth Context -> Role Matrix -> Effective Capability Check`

1. كل request يجمع role permissions مع user overrides قبل الوصول إلى handler الفعلي.
2. كل module يجب أن يدعم على الأقل:
   - `read`
   - `create`
   - `update`
   - `delete`
   - أو حالة `read_only`
3. الإرسال لا يعامل كامتداد ضمني لـ read؛ `messages.send` و`chats.unclaimed.send` مفاتيح مستقلة.
4. إذا كان المستخدم `read_only` على chat module تبقى الرسائل مرئية لكن composer معطل.
5. أي تغيير في role أو permission يجب أن ينعكس على queries والـ buttons والـ endpoints معاً.

## 20. Contact Visibility & Allowed Numbers Flow

**Path**: `Login -> Resolve Scope -> Filter Contacts/Chats`

1. role تحدد ما إذا كان المستخدم يرى:
   - كل جهات الاتصال
   - جهات instances المصرح بها فقط
   - أرقام مسموح بها فقط
2. user-level override يستطيع إضافة allowed phone numbers حتى لو لم يملك رؤية كاملة للinstance.
3. إذا كانت الرؤية محدودة، backend يفلتر `GET /api/contacts` و`GET /api/chats` قبل pagination النهائية.
4. إذا كان كشف الرقم غير مسموح، الواجهة تعرض نسخة masked وتفتح الرقم الكامل فقط لصلاحية صريحة.
5. pending/unclaimed chats لا تظهر أو لا ترسل بحسب `chats.unclaimed.view` و`chats.unclaimed.send`.

## 21. Integration Access Flow

**Path**: `Settings -> API Keys / Webhooks -> Create / Delete / Deliver`

1. admin أو role مصرح له يمكنه إنشاء أو حذف API keys من صفحة `API Keys`.
2. المفتاح يظهر مرة واحدة عند الإنشاء بينما التخزين يكون على شكل hash + prefix.
3. صفحة `Webhooks` تسمح باختيار event واحد أو أكثر، وإدخال `Webhook URL`.
4. `secret` اختياري ويستخدم للتوقيع أو HMAC عند التفعيل.
5. `custom headers` اختيارية وتحفظ كـ key/value pairs.
6. delivery failures تسجل مع retry policy واختبار يدوي عبر `Test webhook`.

## 22. Agent Analytics Flow

**Path**: `Filter Analytics -> Read KPIs -> Drill Down To Chat`

1. صفحة `Agent Analytics` تقرأ فلاتر الوكيل والفترة والحالة والمصدر أو ما يعادلها.
2. backend يرجع بطاقات KPI مثل:
   - handled transfers
   - completed conversations
   - active conversations
   - average resolution time
   - average queue time
   - break time
   - average rating
3. backend يرجع datasets مستقلة لأقسام:
   - `Transfer Trends`
   - `Conversation Sources`
   - `Agent Comparison`
4. جدول `Customer Ratings` يجب أن يرجع:
   - agent
   - phone number
   - contact
   - rating
   - rated at
   - closing agent
   - rating message
   - context messages
5. الضغط على رقم الهاتف يفتح المحادثة المرتبطة مباشرة.
6. لا يوجد assign workflow داخل هذه الشاشة.
7. `Export CSV` يطبق نفس الفلاتر الحالية على التصدير.

## 23. Contacts Management Flow

**Path**: `Contacts Settings -> Search/Filter -> CRUD / Import / Export / Open Chat`

1. صفحة `Contacts` تعرض النتائج مع search وفلتر instance قبل pagination.
2. `Add Contact` ينشئ جهة جديدة من:
   - phone number
   - profile name
   - WhatsApp instance
3. `Edit Contact` يحدث الاسم أو الرقم من دون مغادرة الصفحة.
4. `Open chat` يفتح `/chat/[contactId]` مباشرة.
5. `Import/Export` يدعم:
   - export columns selection
   - import from CSV
   - update existing records on duplicate
6. preview آخر رسالة يبقى متاحاً في الجدول بما في ذلك markers مثل `[image]` أو رسائل النظام.
7. لا يوجد زر assign مباشر داخل هذه الصفحة، لكن النموذج يبقى مرتبطاً بحقل الإسناد.

## 24. Closed Chats Flow

**Path**: `Closed Chats -> Filter -> Refresh / Reopen`

1. صفحة `Closed Chats` تعرض فلاتر:
   - agent
   - instance
   - page size
2. زر `Refresh` يعيد طلب الصفحة الحالية دون أي تغيير في الحالة.
3. الجدول يعرض:
   - contact name
   - closed by
   - date closed
4. `Reopen` تعيد المحادثة إلى inbox التشغيلي.
5. pagination عبر `Previous` و`Next` يجب أن تكون مستقرة مع الفلاتر نفسها.
6. لا يوجد assign workflow داخل شاشة المحادثات المغلقة.

## 25. Offline License Activation Flow

**Path**: `Copy HWID -> Vendor Issues Security Key -> Activate License -> Refresh State`

1. backend يحسب `HWID` كامل و`short id` و`hwid_hash` من fingerprint sources على مستوى الخادم.
2. صفحة `License` تعرض الـ `HWID` وتسمح بنسخه.
3. الـ vendor يصدر `Ed25519-signed JWT` يحتوي:
   - `license_id`
   - `license_family_id`
   - `revision`
   - `tier`
   - `license_kind`
   - entitlements limits
4. المستخدم يلصق `Security Key` داخل `/settings/license` أو `/activate`.
5. backend يتحقق من:
   - signature + `kid`
   - issuer / audience / product
   - `HWID hash`
   - `not_before`
   - `expires_at`
   - freshness مقارنة بـ `license_family_id + revision`
6. token تخزن مشفرة at rest.
7. backend يحدث `license_records` ويسجل `license_events`.
8. الواجهة تعيد تحميل bootstrap وتعرض الحالة الجديدة.

## 26. License State Refresh & Enforcement Flow

**Path**: `Stored License -> Periodic Refresh -> App Gatekeeping`

1. الخدمة تقرأ السجل النشط وتتحقق من سلامته وتوقيعه.
2. الحالة الممكنة:
   - `disabled`
   - `unlicensed`
   - `active`
   - `grace`
   - `locked`
3. إذا انتهت الرخصة لكن ما زالت داخل `grace period` تبقى الحالة `grace`.
4. إذا حصل:
   - `hwid_mismatch`
   - `stored_token_invalid`
   - `time_rollback`
   - `expired` بعد انتهاء الـ grace
   تصبح الحالة `locked`.
5. الخدمة تعيد حساب `quota_overages` دورياً وتنشر state عبر Redis invalidation عندما تتغير.
6. startup يقيّد `worker count` إذا تجاوز `licensed max_workers`.

## 27. License Quota Cleanup Flow

**Path**: `Quota Overage -> App Redirect -> Delete Resources -> Re-check`

1. عندما تتجاوز المنظمة حدود:
   - organizations
   - users per org
   - WhatsApp endpoints per org
   - storage per org
   يدخل التطبيق وضع `quota cleanup`.
2. الـ frontend يحول المستخدمين المصادقين إلى `/license-cleanup`.
3. الـ backend يسمح فقط بطلبات عامة أو طلبات cleanup اللازمة مثل:
   - auth refresh/login/logout
   - قراءة state الأساسية
   - حذف organizations/users/accounts/instances
4. بقية المسارات التشغيلية تبقى محجوبة حتى تعود usage ضمن الحدود.
5. بعد كل حذف، الواجهة تعيد طلب `license bootstrap`.
6. عند اختفاء overage، المستخدم يعود تلقائياً إلى التطبيق العادي.

## 28. Vendor License Studio Flow

**Path**: `Vendor Console -> Generate / Verify / Registry`

1. يوجد console منفصل للـ vendor يحتوي tabs:
   - `Generate`
   - `Verify`
   - `Registry`
2. `Generate` يقبل:
   - HWID
   - private key file
   - tier
   - paid/trial mode
   - orgs/users/WA/workers limits
   - `customer_name`
3. `Verify` يعيد:
   - `valid_tracked`
   - `valid_untracked`
   - `invalid`
4. `Registry` يدعم:
   - filters by HWID/tier/kind/status
   - copy token
   - re-issue
   - remove
   - customer name display
5. guard layer تضيف:
   - login session
   - audit logging
   - customer metadata
   - delete synchronization between registry and metadata store

## 29. Removed Or Still Unverified Logic

المنطق التالي لم يظهر بوضوح في تدقيق `/chat` و`/analytics/agents` و`/settings/license` و`/settings/instances` و`/settings/contacts` و`/settings/closed-chats` الحالي:

- Meta webhook ingestion
- template synchronization
- widgets rendering logic
- AI replies
- agent transfer queue
- SLA processor
- keyword-rule automation
