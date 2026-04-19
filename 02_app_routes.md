# Whatomate App Routes

هذه الوثيقة تمثل **خطة المسارات بعد تدقيق الواجهة الحية** لشاشات `/chat` و`/analytics/agents` و`/settings` و`/settings/license` و`/settings/instances` و`/settings/contacts` و`/settings/closed-chats`.

التحقق الفعلي من الواجهة أكد أن المنتج الحالي يحتوي على تنقل رئيسي أوسع من الخطة المختزلة السابقة، لذلك يجب الإبقاء على:

- `Dashboard`
- `Chat`
- `Chatbot`
- `Agent Analytics`
- `Campaigns`
- `Settings`

مع بقاء `Meta Cloud API` خارج النطاق المؤكد حالياً.

## مبادئ التوجيه

- كل صفحة رئيسية تمثل `+page.svelte`.
- الجلب الأولي للبيانات يتم عبر `+page.ts` أو `+layout.ts`.
- التحقق من الجلسة يتم عبر `hooks.server.ts` و`+layout.server.ts`.
- الاتصال الفوري يتم عبر WebSocket client داخل `frontend/src/lib/realtime/ws.ts`.
- العناصر التي تظهر كـ drawers أو dialogs داخل `/chat` لا تحتاج route مستقل، لكنها تحتاج state URL واضح أو store موحد.

## المسارات المؤكدة من الواجهة

| المسار | ملف SvelteKit | الغرض | أبرز الـ API | WebSocket |
| :--- | :--- | :--- | :--- | :---: |
| `/login` | `src/routes/login/+page.svelte` | تسجيل الدخول | `/api/auth/login` | ❌ |
| `/register` | `src/routes/register/+page.svelte` | إنشاء حساب | `/api/auth/register` | ❌ |
| `/auth/sso/callback` | `src/routes/auth/sso/callback/+page.svelte` | إكمال SSO | `/api/auth/sso/*/callback` | ❌ |
| `/activate` | `src/routes/activate/+page.svelte` | صفحة عامة لإدخال security key عندما يكون التطبيق غير مفعّل أو مغلق | `/api/license/bootstrap`, `/api/license/activate` | ❌ |
| `/dashboard` | `src/routes/(app)/dashboard/+page.svelte` | نظرة عامة تشغيلية | `/api/dashboard/*` | ✅ |
| `/chat` | `src/routes/(app)/chat/+page.svelte` | inbox المحادثات | `/api/chats`, `/api/contacts`, `/api/notifications`, `/api/statuses` | ✅ |
| `/chat/[contactId]` | `src/routes/(app)/chat/[contactId]/+page.svelte` | محادثة محددة مع الأدوات الجانبية | `/api/contacts/{id}/messages`, `/api/contacts/{id}/notes`, `/api/contacts/{id}/collaborators` | ✅ |
| `/chatbot` | `src/routes/(app)/chatbot/+page.svelte` | وحدة chatbot المؤكدة في التنقل | `/api/chatbot/*` | ❌ |
| `/analytics/agents` | `src/routes/(app)/analytics/agents/+page.svelte` | تحليلات الوكلاء مع KPI cards وcharts وجدول ratings وتصدير CSV | `/api/analytics/agents/summary`, `/api/analytics/agents/transfers`, `/api/analytics/agents/sources`, `/api/analytics/agents/comparison`, `/api/analytics/agents/ratings`, `/api/analytics/agents/export` | ✅ |
| `/campaigns` | `src/routes/(app)/campaigns/+page.svelte` | إدارة الحملات | `/api/campaigns*` | ✅ |
| `/profile` | `src/routes/(app)/profile/+page.svelte` | الملف الشخصي | `/api/me`, `/api/me/settings` | ❌ |
| `/settings` | `src/routes/(app)/settings/+page.svelte` | إعدادات عامة + i18n locale bootstrap + Tailwind appearance presets + chat preferences + notifications preferences + summary للـ quotas والاستهلاك | `/api/settings/general`, `/api/settings/limits`, `/api/settings/appearance`, `/api/settings/chat`, `/api/settings/notifications`, `/api/settings/uploads-cleanup/run` | ❌ |
| `/settings/chatbot` | `src/routes/(app)/settings/chatbot/+page.svelte` | إعدادات chatbot من داخل مركز الإعدادات | `/api/chatbot/*` | ❌ |
| `/settings/users` | `src/routes/(app)/settings/users/+page.svelte` | إدارة المستخدمين + availability + send/contact visibility overrides | `/api/users/*`, `/api/users/{id}/send-restrictions`, `/api/users/{id}/contact-visibility` | ❌ |
| `/settings/roles` | `src/routes/(app)/settings/roles/+page.svelte` | إدارة الأدوار + مصفوفة CRUD/read-only + unclaimed view/send + contact visibility scope | `/api/roles/*`, `/api/permissions` | ❌ |
| `/settings/teams` | `src/routes/(app)/settings/teams/+page.svelte` | الفرق | `/api/teams/*` | ❌ |
| `/settings/instances` | `src/routes/(app)/settings/instances/+page.svelte` | بطاقات إدارة instances مع create / rename / delete / connect / disconnect وسياسات التشغيل | `/api/instances/*`, `/api/settings/limits`, `/api/instances/{id}/settings`, `/api/instances/{id}/call-auto-reject`, `/api/instances/{id}/auto-campaign` | ✅ |
| `/settings/instances/health` | `src/routes/(app)/settings/instances/health/+page.svelte` | لوحة صحة instances مع refresh وmetrics يومية وتشغيلية | `/api/instances/health`, `/api/instances/{id}/health` | ❌ |
| `/settings/canned-responses` | `src/routes/(app)/settings/canned-responses/+page.svelte` | الردود الجاهزة | `/api/canned-responses/*` | ❌ |
| `/settings/contacts` | `src/routes/(app)/settings/contacts/+page.svelte` | إدارة جهات الاتصال مع البحث والإضافة والاستيراد/التصدير وفتح المحادثة | `/api/contacts/*`, `/api/contacts/export`, `/api/contacts/import` | ❌ |
| `/settings/closed-chats` | `src/routes/(app)/settings/closed-chats/+page.svelte` | المحادثات المغلقة مع الفلاتر و`Refresh` و`Reopen` | `/api/chats/closed`, `/api/chats/{id}/reopen` | ❌ |
| `/settings/tags` | `src/routes/(app)/settings/tags/+page.svelte` | إدارة التاغات | `/api/tags/*` | ❌ |
| `/settings/api-keys` | `src/routes/(app)/settings/api-keys/+page.svelte` | إنشاء/حذف مفاتيح API | `/api/api-keys/*` | ❌ |
| `/settings/webhooks` | `src/routes/(app)/settings/webhooks/+page.svelte` | Webhooks + events + URL + secret اختياري + headers اختيارية | `/api/webhook-events`, `/api/webhooks/*` | ❌ |
| `/settings/sso` | `src/routes/(app)/settings/sso/+page.svelte` | إعدادات SSO | `/api/settings/sso/*` | ❌ |
| `/settings/license` | `src/routes/(app)/settings/license/+page.svelte` | ترخيص deployment مع offline activation, HWID, quota cards, expiry/grace status، وتجديد المفتاح. هذا المسار admin-only | `/api/license/bootstrap`, `/api/license/activate` | ❌ |
| `/license-cleanup` | `src/routes/(app)/license-cleanup/+page.svelte` | وضع تنظيف خاص يظهر عند وجود `quota overage` عام ويقصر الاستخدام على حذف الموارد حتى العودة ضمن الحد | `/api/license/bootstrap`, `/api/organizations`, `/api/users`, `/api/accounts`, `/api/instances` | ❌ |
| `/settings/custom-actions` | `src/routes/(app)/settings/custom-actions/+page.svelte` | الإجراءات المخصصة | `/api/custom-actions/*` | ❌ |
| `/(fallback)` | `src/routes/+error.svelte` | صفحة الخطأ | لا يوجد | ❌ |

## عناصر مدمجة داخل `/chat`

هذه ليست routes مستقلة، لكنها surfaces مؤكدة في الواجهة الحالية ويجب أن تنعكس في التصميم:

- `Notifications dialog`
  مرتبط بزر الجرس ويعرض unread messages والتنبيهات النظامية مع `Mark all as read`.

- `Statuses drawer`
  مرتبط بزر `Statuses` ويعرض feed للحالات مع `Add status`.

- `Start New Chat dialog`
  يُفتح من زر `Add Contact` ويطلب:
  - رقم دولي
  - اسم اختياري
  - instance للإرسال

- `Assign Contact dialog`
  يُفتح من ترويسة المحادثة لاختيار agent للمحادثة.

- `Invite Collaborator dialog`
  يُفتح من لوحة `Contact Info`.

- `Notes panel`
  panel جانبي للملاحظات الداخلية.

- `Contact Info panel`
  panel جانبي للمعلومات العامة، التاغات، المتعاونين، والبيانات الإضافية.

- `Message composer dropzone`
  surface داخل مساحة المحادثة يدعم drag & drop للملفات مع معاينة أولية قبل الإرسال.

## عناصر مدمجة داخل `/settings`

- `Settings navigation`
  يحتوي على:
  - `General`
  - `Chatbot`
  - `WhatsApp`
  - `Contacts`
  - `Closed Chats`
  - `Canned Responses`
  - `Tags`
  - `Teams`
  - `Users`
  - `Roles`
  - `API Keys`
  - `Webhooks`
  - `Custom Actions`
  - `SSO`
  - `License`

- `General settings tabs`
  داخل `/settings` نفسها توجد تبويبات:
  - `General`
  - `Appearance`
  - `Chat`
  - `Notifications`

- `Role permissions matrix`
  surface داخل `Roles` يسمح بضبط كل وحدة كـ `create` / `update` / `delete` / `read-only`
  مع مفاتيح مستقلة لـ unclaimed chat view/send وcontact visibility scope.

- `User visibility overrides drawer`
  surface داخل `Users` لتحديد scope الرؤية: كل الجهات، جهات instances محددة،
  أو أرقام مسموح بها فقط.

- `Webhook editor`
  داخل `Webhooks` يدعم اختيار events, target URL, secret اختياري,
  وcustom headers اختيارية.

- `Tenant usage and quotas summary`
  surface داخل `General` أو `License` يعرض:
  - active members out of 5
  - reserved and used slots
  - storage used out of 5 GiB
  - `tenant_status`

- `Uploads Cleanup`
  surface داخلي يؤكد الحاجة إلى scheduler تنظيمي مع زر `Run Cleanup Now`، ويجب أن يكون للأدمن فقط.

- `Appearance tab`
  تحتوي على:
  - `Color Mode` بقيم `Light`, `Dark`, `System`
  - `Theme Style`
  - theme presets مبنية على Tailwind مثل `Twitter`, `Ocean Breeze`, `Soft Pop`, `Amber Minimal`

- `Chat tab`
  تحتوي على:
  - `Media Grouping Window`
  - `Sidebar Contact View`
  - `Sidebar Hover Expand`
  - `Pin Sidebar`
  - `Chat Background`
  - `Show Print Buttons`
  - `Show Download Buttons`
  - رابط أو surface إلى `Closed Chats`

- `Notifications tab`
  تحتوي على:
  - `Email Notifications`
  - `New Message Alerts`
  - `Notification Sound`
  - `Play`
  - `Campaign Updates`

## عناصر مؤكدة داخل `/settings/license`

- رأس الصفحة يعرض:
  - `License`
  - `Manage offline activation, quotas, and renewal status for this deployment.`

- الإجراءات المؤكدة:
  - `Refresh`
  - `Copy HWID`
  - `Activate license`
  - `Refresh status`

- بطاقات الحصة الظاهرة:
  - `Organizations`
  - `Users / Org`
  - `WA Endpoints / Org`
  - `Storage / Org`
  - `Subscription Days`

- قسم `Server identity`
  يحتوي:
  - `HWID`
  - `Short ID`
  - زر نسخ الـ HWID

- قسم `Activate or renew`
  يحتوي:
  - حقل `Security Key`
  - تفعيل أو تجديد المفتاح الموقع offline

- من التدقيق الحي في `2026-04-19` ظهرت الحالة `Disabled` مع شرح أن checks غير مفعلة على هذا deployment.
- من الكود المحلي تأكد أن هذا route admin-only وأنه يعتمد على `GET /api/license/bootstrap` و`POST /api/license/activate`.

## عناصر مؤكدة من الكود حول `/license-cleanup`

- عند وجود `quota overage` عام، router يحوّل المستخدمين المصادقين إلى `/license-cleanup`.
- باقي صفحات التطبيق تبقى محجوبة حتى تنخفض:
  - organizations
  - users per org
  - WA endpoints per org
  - storage per org

- الشاشة نفسها تركز على الحذف والتنظيف فقط وتستخدم البيانات الحالية من `license bootstrap`.

## عناصر مؤكدة داخل `/analytics/agents`

- شريط فلاتر يحتوي:
  - `All Agents`
  - `All`
  - `This month`
  - `Any`
  - `Export CSV`

- بطاقات KPI تشمل:
  - `Transfers Handled`
  - `Completed conversations`
  - `Active Conversations`
  - `Avg Resolution Time`
  - `Avg Queue Time`
  - `Break Time`
  - `Average Rating`

- مساحات تحليلات:
  - `Transfer Trends`
  - `Conversation Sources`
  - `Agent Comparison`
  - `Customer Ratings`

- جدول `Customer Ratings`
  يحتوي:
  - `Agent`
  - `Phone Number`
  - `Contact`
  - `Rating`
  - `Rated At`
  - `Closing Agent`
  - `Rating Message`
  - `Context Messages`

- الضغط على رقم الهاتف يفتح `/chat/[contactId]` مباشرة.
- لم تظهر أي خاصية assign داخل شاشة التحليلات نفسها.

## عناصر مؤكدة داخل `/settings/contacts`

- الأزرار العلوية:
  - `Import/Export`
  - `Add Contact`

- الفلاتر والأدوات:
  - مربع بحث `Search contacts...`
  - فلتر `All instances`

- الجدول يحتوي:
  - `Name`
  - `Phone Number`
  - `WhatsApp Instance`
  - `Tags`
  - `Last message`
  - `Created`
  - `Actions`

- إجراءات الصف:
  - `Open chat`
  - `Edit`
  - `Delete`

- `Add Contact dialog`
  يحتوي:
  - `Phone Number`
  - `Profile Name`
  - `WhatsApp Instance`

- `Import/Export Contacts dialog`
  يدعم:
  - `Export CSV`
  - `Import CSV`
  - `Download sample CSV`
  - `Update existing records if duplicate found`

- لا يوجد زر assign مباشر داخل الصفحة، لكن `Assigned User ID` يظهر ضمن أعمدة التصدير.

## عناصر مؤكدة داخل `/settings/closed-chats`

- شريط الفلاتر يحتوي:
  - `All Agents`
  - `All instances`
  - `25`
  - `Refresh`

- الجدول يحتوي:
  - `Contact Name`
  - `Closed By`
  - `Date Closed`
  - `Actions`

- الإجراء المؤكد:
  - `Reopen`

- التنقل:
  - `Previous`
  - `Next`

- لم تظهر خاصية assign داخل هذه الشاشة.

## عناصر مدمجة داخل `/settings/instances`

- `Add WhatsApp Account dialog`
  لإنشاء instance جديدة بالحد الأدنى عبر اسم الحساب، مع إظهار الـ remaining slots قبل التأكيد.

- `Edit Account Name dialog`
  لإعادة تسمية instance موجودة.

- `Call Auto-Reject Settings dialog`
  لإدارة reject policies للمكالمات الفردية والجماعية وجدولة الرد.

- `Auto Campaign Settings dialog`
  لجدولة إنشاء حملات دورية تلقائية per-instance مع delays, status, message body, media.

- `Chat Close Rating Settings dialog`
  لإعداد templates متعددة اللغات ورسالة تقييم بعد إغلاق المحادثة.

- `Assigned Chat Reset dialog`
  لإعادة المحادثات المسندة إلى `pending` بحسب schedule يومي.

- `Chat Source Tag settings`
  لتخصيص label, display mode, color لكل instance داخل inbox.

- `Slot exhaustion state`
  تنبيه أو disabled state يوضح أن create/connect لن ينجح إذا استُهلكت كل slots المخصصة للمنظمة.

## Layouts المقترحة

- `src/routes/+layout.svelte`
  تهيئة عامة، i18n dictionaries من ملفات JSON، theme bootstrap، وbootstrap config.

- `src/routes/(app)/+layout.svelte`
  الغلاف المحمي: sidebar hover-expand/pin، topbar، org switcher، user menu، WebSocket bootstrap.

- `src/routes/(app)/+layout.server.ts`
  تحميل الجلسة، المستخدم، المنظمة الحالية، وإعدادات الشريط الجانبي.

## Notes From Live Audit

- الشريط الجانبي يحتوي زر `Pin sidebar closed` ويمكن أن يتسع تلقائياً عند اقتراب المؤشر.
- في SaaS mode الافتراضي سيملك المستخدم العميل org واحدة غالباً، لذلك org switcher يبقى موجوداً أساساً للحسابات الداخلية أو الحالات الخاصة.
- قائمة المستخدم تحتوي:
  - availability status menu
  - `/profile`
  - theme switcher
  - language selector
  - logout
- لوحة `Contact Info` أشارت إلى أن بعض البيانات الإضافية يتم ضبطها من `chatbot flow settings`، لذلك لا يجوز حذف chatbot من خطة المنتج كلياً.
- صفحة `WhatsApp Instances` تستخدم card layout وتعرض:
  - status badge
  - phone number + JID
  - uptime / queue / sent / received / error rate
  - toggles لـ auto-sync history وauto-download incoming media
- `/settings/instances/health` تعرض بطاقة health summary لكل instance مع `Refresh`.
- `/settings/contacts` تعرض إدارة CRUD فعلية للجهات مع import/export وفتح chat من الصف.
- `/settings/closed-chats` تعرض إدارة مراجعة وتشغيل للمحادثات المغلقة عبر `Refresh` و`Reopen`.
- `/settings/license` تعرض offline activation + quota status + HWID copy، وهي admin-only من الكود المحلي.
- `/license-cleanup` route خاص يفرضه النظام عند وجود overage عام على مستوى الترخيص.
- `/analytics/agents` توفر drill-down إلى المحادثة من جدول التقييمات، من دون surface assign خاص بها.
- لم تظهر أي خاصية assign داخل `/settings` نفسها؛ assignment بقيت في `/chat`.
