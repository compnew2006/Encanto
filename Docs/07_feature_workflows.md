# Whatomate Feature Workflows

هذه الوثيقة تمثل **الخريطة التشغيلية بعد تدقيق الواجهة الحية**.

المنتج المستهدف الآن يجب أن يغطي:

- الهوية والسياق متعدد المنظمات
- chat workspace الفعلي كما يظهر في `/chat`
- settings hub الفعلي كما يظهر في `/settings`
- license management كما يظهر في `/settings/license`
- instance operations كما تظهر في `/settings/instances`
- contacts management كما يظهر في `/settings/contacts`
- closed chats review كما يظهر في `/settings/closed-chats`
- agent analytics كما تظهر في `/analytics/agents`
- statuses وnotifications
- typing simulation قبل الرسائل النصية
- chatbot / analytics / campaigns كمسارات باقية في المنتج

## 1. Authentication & Authorization

**Endpoints**

- `POST /api/auth/login`
- `POST /api/auth/register`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`

**Flow**

1. قراءة بيانات الدخول.
2. تحميل المستخدم.
3. التحقق من كلمة المرور.
4. إصدار `access token`.
5. تسجيل `refresh token` في Redis.
6. إرجاع المستخدم والجلسة.

## 2. Current User, Org Switcher, and Sidebar Context

**Endpoints**

- `GET /api/me`
- `GET /api/me/organizations`
- `PUT /api/me/settings`
- `PUT /api/me/availability`
- `POST /api/auth/switch-org`
- `/api/organizations*`

**Flow**

1. استخراج المستخدم الحالي من الجلسة.
2. تحديد `organization_id` النشط.
3. تحميل إعدادات user menu وsidebar و`locale` وTailwind theme preset وavailability status.
4. عند التبديل، تُصدر tokens جديدة ضمن المنظمة المختارة.
5. إذا كان الشريط غير مثبّت، يتوسع عند hover وفق preference المستخدم.

## 3. Chat Inbox

**Endpoints**

- `GET /api/chats`
- `GET /api/contacts`
- `POST /api/chats/direct`
- `POST /api/contacts/{id}/soft-delete`
- `DELETE /api/contacts/{id}`

**Flow**

1. تحميل chat list مع `Assigned` و`Pending` وفق `chats.unclaimed.view`.
2. دعم البحث بالنص أو الهاتف.
3. دعم الفلترة حسب:
   - instance
   - chat type
   - tags
4. النتائج تمر أولاً عبر role/user visibility scope + allowed phone whitelist.
5. دعم `Hide chat` per-user.
6. دعم `Start New Chat` من رقم دولي + instance فقط إذا كان المستخدم يملك send/create permission على هذا scope.

## 4. Chat Collaboration Surface

**Endpoints**

- `PUT /api/contacts/{id}/assign`
- `PUT /api/chats/{id}/pin`
- `/api/contacts/{id}/notes*`
- `/api/contacts/{id}/collaborators*`
- `/api/contacts/{id}/tags`

**Flow**

1. المستخدم يفتح المحادثة.
2. يمكنه assign لوكيل.
3. يمكن فتح `Notes`.
4. يمكن فتح `Contact Info`.
5. من panel المعلومات:
   - إضافة tags
   - دعوة collaborators
   - عرض البيانات العامة

## 5. Messaging & Media

**Endpoints**

- `GET /api/contacts/{id}/messages`
- `POST /api/contacts/{id}/messages`
- `POST /api/messages/media`
- `POST /api/messages/{id}/retry`
- `POST /api/contacts/{id}/messages/{message_id}/revoke`
- `GET /api/media/{message_id}`

**Flow**

1. التحقق من صلاحية المستخدم على المحادثة وعلى الإرسال (`read_only`, `messages.send`, `chats.unclaimed.send`).
2. إنشاء سجل `pending`.
3. إذا كانت الرسالة نصية، يطلق backend presence/typing عبر `whatsmeow` لمدة محسوبة من عدد المحارف:
   - `typing_ms = clamp(char_count * 55, 900, 7000)`
4. يتم تطبيق حدود دنيا وعليا حتى لا تصبح المحاكاة غير واقعية أو بطيئة جداً.
5. يؤخذ lock قصير في Redis على `message_id` قبل الإرسال لتجنب التكرار عبر أكثر من worker.
6. إذا كانت الرسالة وسائط أو ملفات من picker أو drag & drop، لا يتم تشغيل typing simulation.
7. الإرسال عبر `whatsmeow`.
8. كل محاولة تحفظ في `message_delivery_attempts`.
9. تحديث الحالة إلى `sent` أو `failed`.
10. في الواجهة:
   - print/download للوسائط
   - download-all عند وجود batch
   - retry عند الفشل
   - revoke للرسائل الخارجة
   - dropzone للملفات مع preview أولي

## 6. Notifications Center

**Endpoints**

- `GET /api/notifications`
- `PUT /api/notifications/{id}/read`
- `PUT /api/notifications/read-all`

**Flow**

1. أي unread message أو حدث تشغيلي مهم يولد notification.
2. sidebar يعرض unread count.
3. المستخدم يفتح dialog الإشعارات.
4. يمكنه فتح العنصر المرتبط أو `Mark all as read`.

## 7. Statuses Drawer

**Endpoints**

- `GET /api/statuses`
- `POST /api/statuses`
- `GET /api/statuses/{id}`
- `DELETE /api/statuses/{id}`

**Flow**

1. drawer يعرض feed للحالات.
2. يمكن إنشاء status جديد عبر `Add status`.
3. status ترتبط بجهة وinstance أو على الأقل بالـ instance.
4. يتم تحديث drawer realtime عند الإضافة أو الانتهاء.

## 8. WhatsApp Instance Management

**Endpoints**

- `GET /api/instances`
- `GET /api/instances/health`
- `POST /api/instances`
- `PUT /api/instances/{id}/name`
- `GET /api/instances/{id}/qr`
- `POST /api/instances/{id}/connect`
- `POST /api/instances/{id}/disconnect`
- `POST /api/instances/{id}/reconnect`
- `GET /api/instances/{id}/health`
- `GET /api/instances/{id}/settings`
- `PUT /api/instances/{id}/settings`

**Flow**

1. إنشاء instance داخل المنظمة.
2. pairing عبر QR أو code.
3. عرض instance داخل card مع metrics تشغيلية.
4. دعم rename / delete / disconnect / connect.
5. تحديث الحالة.
6. بث الحالة عبر WebSocket.

## 9. Instance Operational Policies

**Endpoints**

- `GET /api/instances/{id}/call-auto-reject`
- `PUT /api/instances/{id}/call-auto-reject`
- `GET /api/instances/{id}/auto-campaign`
- `PUT /api/instances/{id}/auto-campaign`
- `GET /api/instances/{id}/close-rating`
- `PUT /api/instances/{id}/close-rating`
- `GET /api/instances/{id}/assignment-reset`
- `PUT /api/instances/{id}/assignment-reset`
- `GET /api/instances/{id}/source-tag`
- `PUT /api/instances/{id}/source-tag`

**Flow**

1. instance card تعرض toggles سريعة للـ sync والـ media download.
2. dialogs مستقلة تضبط:
   - call auto-reject
   - auto campaign
   - close-rating templates
   - assigned chat reset
   - source-tag display
3. `Auto campaign` يولد حملات دورية على مستوى instance، ما يعني أن route `Campaigns` ليست معزولة عن قسم `WhatsApp`.

## 10. Settings Hub

**Endpoints**

- `GET /api/settings/general`
- `PUT /api/settings/general`
- `GET /api/settings/appearance`
- `PUT /api/settings/appearance`
- `GET /api/settings/chat`
- `PUT /api/settings/chat`
- `GET /api/settings/notifications`
- `PUT /api/settings/notifications`
- `POST /api/settings/uploads-cleanup/run`

**Flow**

1. `/settings` تعرض تبويبات `General`, `Appearance`, `Chat`, `Notifications`.
2. `General` تحفظ إعدادات المنظمة الأساسية ومنها default locale.
3. `Appearance` تضبط color mode (`light`, `dark`, `system`) وTailwind theme preset.
4. `Chat` تضبط media grouping, sidebar contact view, hover-expand, pin state, background, وإظهار print/download.
5. client يحمّل translation dictionaries من ملفات JSON بحسب `locale` المختار.
6. `Notifications` تضبط email notifications, message alerts, sound, وcampaign updates.
7. قسم `Uploads Cleanup` يضبط retention وجدولة cleanup يومية.
8. `Uploads Cleanup` يجب أن يكون محصوراً على الأدمن فقط.
9. `Run Cleanup Now` يشغل job فوري مع audit مناسب.

## 11. Access Control & Visibility

**Endpoints**

- `GET /api/roles`
- `POST /api/roles`
- `PUT /api/roles/{id}`
- `DELETE /api/roles/{id}`
- `GET /api/permissions`
- `GET /api/users/{id}/send-restrictions`
- `PUT /api/users/{id}/send-restrictions`
- `GET /api/users/{id}/contact-visibility`
- `PUT /api/users/{id}/contact-visibility`

**Flow**

1. صفحات `Roles` و`Users` تعرض مصفوفة CRUD/read-only لكل capability.
2. role تحدد `messages.send` و`chats.unclaimed.view` و`chats.unclaimed.send` بشكل مستقل.
3. role أو user override تحدد ما إذا كان المستخدم يرى:
   - كل جهات الاتصال
   - جهات instances محددة
   - أرقام مسموح بها فقط
4. allowed phone numbers تعمل كاستثناءات صريحة حتى عند غياب رؤية كاملة للinstance.
5. كل queries والbuttons والendpoints يجب أن تطبق النتيجة نفسها دون اختلاف.

## 12. Integration Access & Webhooks

**Endpoints**

- `GET /api/api-keys`
- `POST /api/api-keys`
- `DELETE /api/api-keys/{id}`
- `GET /api/webhook-events`
- `GET /api/webhooks`
- `POST /api/webhooks`
- `PUT /api/webhooks/{id}`
- `DELETE /api/webhooks/{id}`
- `POST /api/webhooks/{id}/test`

**Flow**

1. صفحة `API Keys` تسمح بإنشاء وحذف المفاتيح وفق الصلاحية.
2. صفحة `Webhooks` تسمح بإنشاء أو حذف webhook وربطه بحدث واحد أو أكثر.
3. كل webhook تحتاج `Webhook URL` صالحاً واشتراك events واحداً على الأقل.
4. `secret` اختياري للتوقيع، و`custom headers` اختيارية كأزواج key/value.
5. زر `Test webhook` ينفذ طلباً تجريبياً مع audit للنتيجة.

## 13. Chatbot-Adjacent Configuration

**Endpoints**

- `/api/chatbot/*`
- أو أي surface يعادل `chatbot flow settings`

**Flow**

1. chatbot تظهر في الشريط الجانبي.
2. `Contact Info` تشير إلى data panels قابلة للضبط من chatbot flows.
3. لذلك chatbot تبقى في الخطة، حتى لو تأجل تدقيقها المفصل.

## 14. Instance Health Dashboard

**Endpoints**

- `GET /api/instances/health`
- `GET /api/instances/{id}/health`

**Flow**

1. health page تعرض summary cards per-instance.
2. كل بطاقة تعرض:
   - uptime
   - sent today
   - received today
   - failed today
   - error rate
   - queue depth
3. زر `Refresh` يجلب snapshot حديث.

## 15. Dashboard, Analytics, and Campaigns

**Endpoints**

- `/api/dashboard/summary`
- `/api/dashboard/inbox`
- `/api/dashboard/instances`
- `/api/analytics/agents*`
- `/api/campaigns*`

**Flow**

1. هذه الوحدات موجودة فعلياً في التنقل.
2. يجب أن تبقى ضمن الخطة المعمارية.
3. `Agent Analytics` أصبحت الآن مدققة بما يكفي لتثبيت KPIs وجدول ratings والتصدير.
4. `Dashboard` تبنى من aggregates فوق chats, instances, jobs, quota state, وrecent conversation events مع cache قصير في Redis.
5. `Campaigns` لم تعد تبقى route بلا domain؛ يجب أن تمتلك baseline data model واضحاً للحملات وruns والrecipients.

## 16. Contacts Management

**Endpoints**

- `GET /api/contacts`
- `POST /api/contacts`
- `PUT /api/contacts/{id}`
- `DELETE /api/contacts/{id}`
- `GET /api/contacts/export`
- `POST /api/contacts/import`

**Flow**

1. صفحة `Contacts` تعرض search وفلتر instance قبل pagination.
2. `Add Contact` ينشئ contact جديدة من الهاتف والاسم والـ instance.
3. `Edit` يحدّث بيانات الجهة من dialog داخل الصفحة.
4. `Open chat` ينتقل إلى شاشة المحادثة الفعلية.
5. `Import/Export` يدعم تصدير CSV بأعمدة اختيارية واستيراد CSV مع update-on-duplicate.
6. بيانات الإسناد تبقى ضمن النموذج وقابلة للتصدير حتى لو لم يوجد زر assign مباشر هنا.

## 17. Closed Chats

**Endpoints**

- `GET /api/chats/closed`
- `PUT /api/chats/{id}/reopen`

**Flow**

1. صفحة `Closed Chats` تقرأ الفلاتر حسب agent وinstance وpage size.
2. `Refresh` يعيد طلب الصفحة الحالية.
3. الجدول يعرض contact name وclosed by وdate closed.
4. `Reopen` يعيد المحادثة إلى inbox التشغيلي.
5. `Previous` و`Next` يحافظان على الفلاتر الحالية.

## 18. Agent Analytics

**Endpoints**

- `GET /api/analytics/agents/summary`
- `GET /api/analytics/agents/transfers`
- `GET /api/analytics/agents/sources`
- `GET /api/analytics/agents/comparison`
- `GET /api/analytics/agents/ratings`
- `GET /api/analytics/agents/export`

**Flow**

1. الصفحة تقرأ الفلاتر الحالية ثم تحمل بطاقات KPI.
2. تعرض charts لـ transfer trends وconversation sources وagent comparison.
3. تعرض جدول customer ratings مع drill-down إلى chat من رقم الهاتف.
4. `Export CSV` يصدر نفس النتيجة المفلترة.
5. لا يوجد assign workflow داخل analytics نفسها.

## 19. Offline License Activation

**Endpoints**

- `GET /api/license/bootstrap`
- `POST /api/license/activate`

**Flow**

1. صفحة `License` أو `/activate` تطلب bootstrap state أولاً.
2. الواجهة تعرض:
   - status
   - quota cards
   - HWID
   - short id
   - security key textarea
3. المستخدم ينسخ HWID ويرسله إلى الـ vendor.
4. vendor يصدر `signed offline security key`.
5. `POST /api/license/activate` يتحقق من signature وHWID والوقت والrevision.
6. backend يخزن token مشفرة ويحدث entitlements ثم يرجع state جديداً.

## 20. License Enforcement & Cleanup

**Endpoints**

- `GET /api/license/bootstrap`
- أخطاء `423 Locked` على المسارات المحجوبة
- أخطاء `402 Payment Required` عند تجاوز quota resource-specific

**Flow**

1. backend يعيد بناء state دورياً من `license_records`.
2. إذا كانت الحالة `locked` يسمح فقط بطلبات bootstrap/activate/public surfaces اللازمة.
3. إذا ظهرت `quota_overages` عامة، router يحول المستخدم إلى `/license-cleanup`.
4. شاشة cleanup تسمح فقط بعمليات الحذف اللازمة للعودة ضمن الحدود.
5. عند انخفاض الاستهلاك، يعود التطبيق إلى العمل العادي.

## 21. Vendor License Studio

**Endpoints**

- `POST /auth/login`
- `POST /auth/logout`
- `GET /guard/healthz`
- `GET /api/bootstrap`
- `POST /api/issue`
- `POST /api/verify`
- `GET /api/licenses`
- `GET /api/licenses/{id}/token`
- `DELETE /api/v1/licenses/{id}`

**Flow**

1. console منفصل للـ vendor يسجل الدخول ثم يحمّل bootstrap defaults والملخص.
2. `Generate` ينشئ token جديدة من HWID + private key + entitlements.
3. `Verify` يراجع token ويحدد tracked/untracked/invalid.
4. `Registry` يعرض السجل المحلي مع filters وcopy/reissue/remove.
5. guard layer تضيف customer metadata وaudit trail فوق backend الأصلي.

## 22. Contact State & Timeline

**Endpoints**

- `PUT /api/contacts/{id}/state`
- `GET /api/contacts/{id}/events`
- `POST /api/contacts/{id}/unassign`

**Flow**

1. `pin`, `hide`, وآخر رسالة مقروءة تحفظ على مستوى المستخدم في `contact_user_states`.
2. assign/unassign/claim/close/reopen تكتب `conversation_events`.
3. هذه الطبقة تصبح المرجع للtimeline التشغيلي ولتحليلات النقل والانتظار والحل.

## 23. Outbox & Delivery Reliability

**Endpoints**

- internal dispatcher over `outbox_events`
- `GET /api/webhooks/{id}/deliveries`
- `POST /api/webhooks/{id}/deliveries/{delivery_id}/retry`

**Flow**

1. كل mutation مهمة تكتب `outbox_event` داخل نفس transaction.
2. dispatcher ينشر event إلى websocket rooms وwebhook deliveries.
3. failures لا تضيع الحدث بل تنتج delivery record جديدة أو retry على السجل نفسه وفق السياسة.

## 24. Jobs, Redis, and Locks

**Endpoints**

- `GET /api/jobs/{id}`

**Flow**

1. Redis تستخدم للـ:
   - refresh tokens
   - websocket pub/sub
   - presence
   - rate limits
   - locks
   - delayed retries
2. أي job طويلة أو قابلة للإعادة تسجل في `job_runs`.
3. `Uploads Cleanup`, imports, reconnects, webhook replay, وcampaign runs تعتمد على هذا المسار.

## 25. Campaign Execution Baseline

**Endpoints**

- `GET /api/campaigns`
- `POST /api/campaigns`
- `POST /api/campaigns/{id}/launch`
- `GET /api/campaigns/{id}/runs`
- `GET /api/campaigns/{id}/recipients`

**Flow**

1. الحملة تحفظ كتعريف reusable.
2. كل تشغيل فعلي ينتج `campaign_run`.
3. كل جهة مستهدفة تنتج `campaign_recipient`.
4. `instance_auto_campaigns` يمكنها إنشاء أو جدولة runs ضمن نفس الدومين بدلاً من منطق جانبي منفصل.

## 26. Audit & Admin Safety

**Endpoints**

- `GET /api/audit-logs`

**Flow**

1. أي فعل حساس يكتب `audit_log`.
2. `Uploads Cleanup` تبقى admin-only ويجب توثيق استخدامها في الـ audit.
3. license activation, role edits, API keys, instance delete/disconnect, وcampaign control كلها تدخل في السجل نفسه.
