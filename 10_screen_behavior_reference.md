# المرجع المعتمد للشاشات والسلوك

هذا الملف هو المرجع الحاكم لوصف الشاشات والسلوك المتوقع داخل المنتج.
عند وجود تعارض بين ملفات المسارات أو الـ API أو سير العمل، يُعتمد هذا الملف للفصل في:

- ما هي الشاشة المعتمدة فعلاً
- ما الأزرار أو العناصر المؤكدة داخلها
- ما الفعل الناتج عن كل عنصر
- ما النتيجة المتوقعة على الواجهة والبيانات

## مستوى الثقة

- `مؤكد بصرياً`: ظهر في التدقيق الحي للواجهة.
- `مؤكد وظيفياً`: ثبت من المسارات والـ API وسير العمل، لكن ليس كل تفاصيله مرئية في التدقيق الحي.
- `مضاف لسد فجوة`: ليس عنصراً تنقلياً مؤكداً، لكنه معتمد لأن غيابه يكسر التشغيل أو التتبع.

## قواعد حاكمة قبل قراءة الشاشات

- assignment وownership transfer يبقيان داخل chat/contact workflows فقط، ولا يُفترضان داخل analytics أو settings أو instance management.
- `pin` و`hide` وآخر قراءة هي state شخصية لكل مستخدم في `contact_user_states` وليست mutation عامة على contact.
- `conversation timeline` معتمدة كسطح لازم داخل المحادثة، حتى لو لم تكن route مستقلة.
- `Dashboard` و`Agent Analytics` و`Closed Chats` تعتمد على الوقائع التشغيلية و`conversation_events`، وليس على source of truth منفصل.
- `Campaigns` route معتمدة، لكن لا يجوز اختراع عناصر UI غير مثبتة؛ يعتمد فقط الحد الأدنى المؤكد من listing, create, detail, runs, recipients, launch/pause/resume.
- `Chatbot` يبقى في المنتج كمسار وإعدادات مجاورة لأنه يؤثر على data panels داخل `Contact Info`.
- `Vendor License Studio` يبقى خارج UI التطبيق الرئيسي وخارج قاعدة بياناته.
- العناصر غير المؤكدة لا تدخل هذا المرجع: `Meta`, `templates`, `widgets`, `AI replies`, `SLA`, `keyword rules`, و`transfer queue` كوحدة مستقلة.

## الشاشات العامة وشاشات الوصول

### `/login`

- الحالة: `مؤكد وظيفياً`
- الغرض: إدخال المستخدم إلى التطبيق بجلسة مستقرة.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| نموذج الدخول | إدخال بيانات الاعتماد ثم الإرسال | التحقق من المستخدم، إصدار session/tokens، ثم تحميل سياق المستخدم |
| فشل التحقق | إرسال بيانات غير صالحة | فشل منضبط من دون إنشاء جلسة |

### `/register`

- الحالة: `مؤكد وظيفياً`
- الغرض: إنشاء حساب عندما يكون التسجيل مسموحاً.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| نموذج التسجيل | إدخال بيانات التسجيل ثم الإرسال | إنشاء الحساب ثم بدء تدفق الجلسة أو إعادة التوجيه حسب سياسة التطبيق |

### `/auth/sso/callback`

- الحالة: `مؤكد وظيفياً`
- الغرض: إتمام عودة SSO وتثبيت الجلسة.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| callback handler | استقبال رد المزود | تثبيت الجلسة أو إظهار فشل منضبط إذا كانت العودة غير صالحة |

### `/activate`

- الحالة: `مؤكد وظيفياً`
- الغرض: صفحة عامة لتفعيل الترخيص عندما يكون التطبيق غير مفعل أو مقفلاً.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Security Key` | لصق المفتاح الموقع | إرسال طلب التفعيل إلى `/api/license/activate` |
| `Activate license` | بدء التفعيل | تثبيت الترخيص ثم إعادة تحميل bootstrap |
| قراءة bootstrap | تحميل حالة الترخيص الحالية | إظهار status و`HWID` والحدود الحالية |

## غلاف التطبيق العام

### `/(app)` layout

- الحالة: `مؤكد وظيفياً`
- الغرض: غلاف التطبيق المحمي والسياق التنظيمي والشخصي.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Sidebar pin toggle | تثبيت أو فك تثبيت الشريط | حفظ حالة الشريط في إعدادات المستخدم |
| Organization switcher | تبديل المنظمة الحالية | تبديل البيانات والصلاحيات والقوائم إلى `organization_id` الصحيح |
| User menu | فتح القائمة الشخصية | إظهار availability و`/profile` والثيم واللغة وlogout |
| Availability status | تغيير حالة التوفر | تحديث الحالة الشخصية وبثها عند الحاجة |
| Theme switcher | تغيير الثيم | تطبيق المظهر وحفظه ضمن الإعدادات الشخصية |
| Language selector | تغيير اللغة | تحميل `i18n JSON` المناسب مع fallback واضح |
| Logout | إنهاء الجلسة | تسجيل الخروج والعودة إلى شاشة الوصول |

## الشاشات الرئيسية

### `/dashboard`

- الحالة: `مؤكد وظيفياً`
- الغرض: نظرة تشغيلية مجمعة فوق الدومينات القائمة.
- ملاحظة: لا تملك dashboard source of truth مستقلاً.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Overview cards | فتح الصفحة | تحميل counters من chats, instances, jobs, quotas, recent events |
| Inbox summary | فتح الصفحة أو تحديث البيانات | إظهار assigned/pending/unread/recent failures |
| Instances summary | فتح الصفحة أو تحديث البيانات | إظهار حالة الحسابات والـ queue depth والاتصال |
| Live refresh | وصول `dashboard_update` | تحديث البطاقات من دون كسر السياق الحالي |

### `/chat`

- الحالة: `مؤكد بصرياً`
- الغرض: inbox المحادثات ومساحة التشغيل اليومية.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Assigned` / `Pending` tabs | التبديل بين التبويبين | تحديث القائمة وفق حالة الإسناد وصلاحيات `unclaimed` |
| Search | البحث بالنص أو الهاتف | ترشيح النتائج داخل النطاق المسموح للمستخدم |
| Filters | الفلترة حسب instance/chat type/tags | تحديث النتائج مع الحفاظ على قواعد الرؤية |
| `Add Contact` | فتح بدء محادثة جديدة | إظهار dialog إدخال رقم واسم اختياري وinstance |
| Refresh | إعادة التحميل | إعادة جلب القائمة والعدادات |
| Row click | فتح محادثة | الانتقال إلى `/chat/[contactId]` |
| `Hide chat` | إخفاء المحادثة للمستخدم الحالي | تحديث `contact_user_states` وإخفاء العنصر من inbox لهذا المستخدم فقط |
| `Delete chat` | حذف الكيان عند توفر الصلاحية | إزالة المحادثة/الجهة وفق قواعد الحذف المعتمدة، وليس كإخفاء شخصي |

### `/chat/[contactId]`

- الحالة: `مؤكد بصرياً` مع أفعال lifecycle `مؤكدة وظيفياً`
- الغرض: عرض المحادثة المفتوحة مع كل أدوات التعاون والإرسال.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Conversation header | فتح المحادثة | إظهار هوية الجهة، cluster الأفعال، وحالة المحادثة |
| `Assign` | فتح اختيار المسؤول | إسناد المحادثة وتحديث الحالة والـ timeline |
| `Unassign` / responsibility reset | إعادة المحادثة إلى pending | تسجيل `unassign` وإرجاع المسؤولية إلى وضع الانتظار |
| `Pin` | تثبيت المحادثة | تحديث الحالة الشخصية للمستخدم في inbox |
| `Notes` | فتح panel الملاحظات | عرض الملاحظات الداخلية وإتاحة الإنشاء |
| `Info` | فتح `Contact Info` | عرض البيانات العامة، tags، collaborators، والبيانات الإضافية |
| `Close` | إغلاق المحادثة | نقلها إلى closed chats وتسجيل الحدث |
| `Reopen` | إعادة فتح المحادثة | إعادتها إلى inbox وتسجيل الحدث |

- موضع أزرار `Close` و`Reopen` داخل الترويسة لم يكن محور التدقيق البصري، لكنه معتمد وظيفياً لأن lifecycle نفسه مثبت في المسارات والـ API وscreen flows.

### Message composer داخل `/chat/[contactId]`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Emoji | إدراج emoji | تحديث حقل الرسالة قبل الإرسال |
| Quick replies | فتح الردود الجاهزة | إدراج رد جاهز في composer |
| Attachment | اختيار ملف أو وسائط | تجهيز إرسال media/file من دون typing simulation |
| Dropzone | سحب وإفلات ملف | معاينة أولية ثم تجهيز الإرسال |
| Print | طباعة محتوى قابل للطباعة | إتاحة طباعة الرسائل/الوسائط المدعومة |
| Send | إرسال الرسالة | إذا كانت نصية: pending ثم typing simulation ثم send. إذا كانت media/file: إرسال مباشر بلا typing simulation |

### Message history داخل `/chat/[contactId]`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Retry sending` | إعادة محاولة الرسالة الفاشلة | إنشاء attempt جديد في `message_delivery_attempts` |
| `Revoke message` | سحب رسالة صادرة | تحديث حالة الرسالة وتسجيل الحدث إذا نجح المزود |
| `Download` | تنزيل ملف أو وسائط | قراءة الملف من مصدر التخزين |
| `Download All` | تنزيل batch كاملة | تنزيل كل الملفات المجمعة المدعومة |
| `Print` | طباعة عنصر وسائط أو رسالة قابلة للطباعة | فتح سلوك طباعة مناسب |
| Reactions | إضافة أو عرض تفاعل | تحديث الرسالة وإظهار reaction |

### `/chatbot`

- الحالة: `مؤكد وظيفياً`
- الغرض: وحدة chatbot موجودة في التنقل، لكن مرجعها الحالي يقتصر على كونها تؤثر على `Contact Info` panels.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Chatbot module route | فتح المسار | الوصول إلى سطح chatbot المعتمد في التنقل |
| Chatbot-configured data panels | تعديل إعدادات panel | انعكاس البيانات الإضافية داخل `Contact Info` عندما تكون المعرفة موجودة |

- لا توجد في المراجع الحالية أزرار مدققة بصرياً داخل هذه الشاشة أكثر من وجودها كمسار ووحدة إعدادات.

### `/analytics/agents`

- الحالة: `مؤكد بصرياً`
- الغرض: تقارير العاملين مع drill-down إلى المحادثة.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Filter bar | تغيير agent/period/status/source | تحديث KPIs والجداول بنفس الفلاتر |
| `Export CSV` | تصدير البيانات | إخراج نفس النتيجة المفلترة الحالية |
| KPI cards | فتح الصفحة | عرض `Transfers Handled`, `Completed conversations`, `Active Conversations`, `Avg Resolution Time`, `Avg Queue Time`, `Break Time`, `Average Rating` |
| `Transfer Trends` | عرض القسم | تحميل dataset الخاص بالتحويلات |
| `Conversation Sources` | عرض القسم | تحميل dataset الخاص بالمصادر |
| `Agent Comparison` | عرض القسم | تحميل dataset المقارنات |
| `Customer Ratings` table | عرض الجدول | إظهار agent, phone, contact, rating, rated at, closing agent, message, context |
| Phone drill-down | الضغط على رقم الهاتف | فتح `/chat/[contactId]` المرتبط مباشرة |

- لا يوجد assign workflow داخل هذه الشاشة.

### `/campaigns`

- الحالة: `مؤكد وظيفياً`
- الغرض: قائمة الحملات مع الإنشاء والحالة وآخر تشغيل.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Campaign list | فتح الصفحة | تحميل الحملات مع status/source/schedule/last run summary |
| Filters | تغيير الفلاتر | تحديث القائمة وفق المرشحات الحالية |
| `Create campaign` | إنشاء draft أو حملة مجدولة | حفظ تعريف حملة reusable داخل domain الحملات |
| `Launch` | تشغيل حملة يدوياً | إنشاء `campaign_run` جديد |
| `Pause` | إيقاف حملة مجدولة أو نشطة | نقل الحالة إلى paused |
| `Resume` | استئناف حملة paused | استعادة الجدولة أو التنفيذ |

- تفاصيل الـ UI الدقيقة للحملة غير مدققة بصرياً بالكامل، لذلك يعتمد الحد الأدنى فقط.

### `/campaigns/[campaignId]`

- الحالة: `مؤكد وظيفياً`
- الغرض: تفاصيل الحملة.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Campaign definition | فتح الصفحة | إظهار المحتوى والفلاتر والجدولة |
| Update campaign | تعديل التعريف | حفظ التعديلات على الحملة |
| Runs view | فتح runs | إظهار سجل تشغيلات الحملة |
| Recipients view | فتح recipients | إظهار حالة كل مستهدف وfailure reason عند الحاجة |

### `/profile`

- الحالة: `مؤكد وظيفياً`
- الغرض: الملف الشخصي للمستخدم.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Profile route | فتح الصفحة | تحميل بيانات المستخدم الحالية وإعداداته الشخصية |
| Personal settings save | تعديل الإعدادات الشخصية | حفظ الإعدادات وربطها بالغلاف العام للتطبيق |

- تفاصيل الحقول داخل الصفحة غير مدققة بصرياً بالكامل.

## السطوح المدمجة داخل chat

### Notifications dialog

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Notifications button | فتح dialog | عرض unread messages والتنبيهات النظامية |
| Item click | فتح العنصر المرتبط | الانتقال إلى السطح أو المحادثة المرتبطة |
| `Mark all as read` | تعليم الكل كمقروء | تحديث حالة الإشعارات للمستخدم الحالي |

### Statuses drawer

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Statuses button | فتح drawer | عرض feed الحالات |
| `Add status` | إنشاء status جديد | إضافة status وربطها بالجهة أو الـ instance |
| Feed updates | وصول تحديثات حية | تحديث drawer عبر realtime |

### Start New Chat dialog

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Phone Number` | إدخال الرقم | تحديد الجهة المستهدفة |
| `Profile Name` | إدخال اسم اختياري | حفظ الاسم إن تم توفيره |
| `WhatsApp Instance` | اختيار حساب الإرسال | ربط الإنشاء بنطاق الإرسال الصحيح |
| Create/start action | تأكيد البداية | إنشاء direct chat إذا كانت الصلاحية والحدود تسمح |

### Assign Contact dialog

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Searchable assignee list | البحث عن مستخدم | تضييق قائمة المكلفين |
| Select assignee | اختيار المسؤول | تحديث الإسناد وتسجيل الحدث تاريخياً |

### Invite Collaborator dialog

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Searchable user list | البحث عن متعاون | تضييق قائمة المستخدمين |
| Invite action | دعوة المتعاون | إضافة collaborator أو إنشاء دعوة/حالة قبول |

### Notes panel

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Notes list | فتح اللوحة | عرض الملاحظات السابقة |
| Note input | كتابة ملاحظة | إنشاء note داخل المحادثة |

### Contact Info panel

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| General fields | فتح اللوحة | عرض الهاتف والبيانات العامة |
| Tags | تعديل التاغات | تحديث ربط التاغات بالجهة |
| Collaborators | إضافة أو إزالة متعاونين | تحديث قائمة المتعاونين |
| Chatbot-configured area | عرض panel إضافية | قراءة تكوينها من `chatbot flow settings` لا من منطق مستقل داخل chat |

### Conversation timeline drawer

- الحالة: `مضاف لسد فجوة`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Timeline open action | فتح drawer | عرض assignment/claim/close/reopen/collaborator changes من `conversation_events` |
| Event list | مراجعة السجل | دعم debugging والتحليلات وclosed chats من نفس المصدر |

## مركز الإعدادات

### `/settings`

- الحالة: `مؤكد بصرياً`
- الغرض: hub واسع، وليس صفحة إعدادات واحدة مسطحة.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Settings sidebar | التنقل بين الأقسام | تحميل القسم المختار داخل مركز الإعدادات |
| `General` tab | فتح التبويب | إظهار organization name, slug, timezone, date format, locale, mask phone numbers, usage summary |
| `Appearance` tab | فتح التبويب | إظهار color mode وtheme presets |
| `Chat` tab | فتح التبويب | إظهار chat behavior preferences |
| `Notifications` tab | فتح التبويب | إظهار notification preferences |
| Save action | حفظ تبويب حالي | تطبيق التغيير على الإعدادات العامة أو الشخصية بحسب الحقل |

### General tab داخل `/settings`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Organization fields | تعديل بيانات المنظمة | حفظ الإعدادات الأساسية |
| Usage summary | عرض limits والاستهلاك | تمكين المستخدم من فهم الحصة قبل الفشل في create/upload |
| `Uploads Cleanup` schedule | ضبط retention/day/hour | حفظ الجدولة الإدارية للتنظيف |
| `Run Cleanup Now` | تشغيل cleanup فوري | إنشاء background job مع audit واضح |

- `Uploads Cleanup` و`Run Cleanup Now` admin-only.

### Appearance tab داخل `/settings`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Color Mode` | اختيار `Light` أو `Dark` أو `System` | تطبيق النمط وحفظه |
| `Theme Style` / presets | اختيار preset | تطبيق preset وحفظه للمستخدم |

### Chat tab داخل `/settings`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Media Grouping Window` | تعديل القيمة | تغيير سلوك تجميع الوسائط |
| `Sidebar Contact View` | تعديل العرض | تغيير شكل عرض جهات الاتصال |
| `Sidebar Hover Expand` | تفعيل أو تعطيل | تغيير تمدد الشريط الجانبي |
| `Pin Sidebar` | تفعيل أو تعطيل | حفظ ثبات الشريط الجانبي |
| `Chat Background` | تعديل الخلفية | تغيير مظهر مساحة المحادثة |
| `Show Print Buttons` | تفعيل أو تعطيل | إظهار أو إخفاء أزرار الطباعة |
| `Show Download Buttons` | تفعيل أو تعطيل | إظهار أو إخفاء أزرار التنزيل |
| `Closed Chats` link/surface | فتح المحادثات المغلقة | الانتقال إلى مراجعة closed chats |

### Notifications tab داخل `/settings`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Email Notifications` | تفعيل أو تعطيل | حفظ تفضيل البريد |
| `New Message Alerts` | تفعيل أو تعطيل | حفظ تفضيل التنبيه بالرسائل |
| `Notification Sound` | اختيار الصوت | حفظ الصوت المحدد |
| `Play` | تجربة الصوت | تشغيل معاينة للصوت |
| `Campaign Updates` | تفعيل أو تعطيل | حفظ تفضيل تنبيهات الحملات |

### `/settings/chatbot`

- الحالة: `مؤكد وظيفياً`
- الغرض: إعدادات chatbot ضمن settings hub.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Chatbot settings route | فتح المسار | تعديل إعدادات chatbot ذات الصلة بالمنتج |
| Flow/panel configuration | حفظ الإعدادات | انعكاسها على `Contact Info` panels عند الحاجة |

### `/settings/users`

- الحالة: `مؤكد وظيفياً`
- الغرض: إدارة المستخدمين وتجاوزات الرؤية والإرسال.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Users list | فتح الصفحة | عرض المستخدمين الحاليين |
| Availability controls | تعديل الحالة | تحديث حالة توفر المستخدم |
| Send restrictions | تعديل القيود | ضبط من يمكنه الإرسال وفي أي نطاق |
| Contact visibility overrides | ضبط scope أو allowed numbers | توحيد الرؤية بين list/search/detail |

### `/settings/roles`

- الحالة: `مؤكد وظيفياً`
- الغرض: إدارة الأدوار ومصفوفة الصلاحيات.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Roles list | فتح الصفحة | عرض الأدوار الحالية |
| Create/update/delete role | تعديل الدور | حفظ التغييرات على role definitions |
| Permission matrix | ضبط CRUD/read-only | توحيد الإنفاذ على الأزرار والـ endpoints والواجهات |
| `unclaimed view/send` controls | تعديل الصلاحية | ضبط التعامل مع pending/unclaimed chats |
| Contact visibility scope | تعديل النطاق | تحديد مستوى الرؤية للمستخدمين الخاضعين للدور |

### `/settings/teams`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Teams list | فتح الصفحة | عرض الفرق الحالية |
| Team CRUD | إنشاء أو تعديل أو حذف | حفظ بنية الفرق وعضويتها |

### `/settings/instances`

- الحالة: `مؤكد بصرياً`
- الغرض: إدارة حسابات WhatsApp كبطاقات تشغيلية.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Add Account` | فتح dialog الإضافة | إدخال الحد الأدنى وإنشاء instance إذا كانت slots تسمح |
| Instance card | عرض البطاقة | إظهار status, phone, JID, uptime, queue, sent/received, error rate |
| `Edit instance name` | إعادة التسمية | تحديث الاسم المعروض |
| `Delete instance` | حذف الحساب | إزالة الحساب وتحرير slot المحجوزة |
| `Disconnect` | فصل الحساب المربوط | إبقاء instance موجودة مع فقدان الربط |
| `Connect / Scan QR` | بدء أو استكمال الربط | إظهار QR أو تشغيل pairing flow |
| Quick toggles | تفعيل/تعطيل auto-sync وauto-download | حفظ إعدادات التشغيل السريعة |
| `Health Dashboard` | فتح لوحة الصحة | الانتقال إلى `/settings/instances/health` |

### Dialogs داخل `/settings/instances`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Add WhatsApp Account` | إدخال اسم الحساب ثم الحفظ | إنشاء instance جديدة مع فحص remaining slots |
| `Edit Account Name` | تعديل الاسم | حفظ الاسم الجديد |
| `Call Auto-Reject Settings` | ضبط السياسة | حفظ قواعد رفض المكالمات |
| `Auto Campaign Settings` | ضبط auto campaign | إنشاء أو تعديل ربط التشغيل الآلي بالحملات |
| `Chat Close Rating Settings` | ضبط template التقييم | حفظ إعدادات رسالة التقييم بعد الإغلاق |
| `Assigned Chat Reset` | ضبط schedule | إعادة المحادثات المسندة إلى pending وفق السياسة |
| `Chat Source Tag settings` | ضبط label/display/color | تحديث عرض مصدر المحادثة في inbox |
| Slot exhaustion state | محاولة إنشاء أو ربط بدون slots | إظهار disabled/error state واضح يمنع العملية |

### `/settings/instances/health`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Health cards | فتح الصفحة | عرض health summary لكل instance |
| `Refresh` | تحديث اللقطة | جلب snapshot أحدث للـ metrics |

### `/settings/canned-responses`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Canned responses list | فتح الصفحة | عرض الردود الجاهزة |
| CRUD actions | إنشاء أو تعديل أو حذف | حفظ الردود الجاهزة لاستخدامها من quick replies |

### `/settings/contacts`

- الحالة: `مؤكد بصرياً`
- الغرض: CRUD للجهات مع import/export وفتح المحادثة.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Import/Export` | فتح dialog الاستيراد/التصدير | إظهار أدوات CSV |
| `Add Contact` | فتح dialog الإضافة | إنشاء جهة اتصال جديدة |
| Search | البحث | ترشيح النتائج |
| `All instances` filter | اختيار instance | تضييق النتائج حسب الحساب |
| `Open chat` | فتح المحادثة | الانتقال إلى `/chat/[contactId]` |
| `Edit` | تعديل الجهة | حفظ تغييرات الاسم أو الرقم |
| `Delete` | حذف الجهة | حذف contact وفق الصلاحيات |

### Dialogs داخل `/settings/contacts`

- الحالة: `مؤكد بصرياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Add Contact` dialog | إدخال phone/profile/instance | إنشاء جهة مرتبطة بالحساب المختار |
| `Import CSV` | استيراد ملف | إدخال جهات جديدة أو تحديث المكرر عند التفعيل |
| `Export CSV` | تصدير البيانات | إخراج الملف بالأعمدة المختارة |
| `Download sample CSV` | تنزيل العينة | توفير ملف مرجعي للاستيراد |
| `Update existing records if duplicate found` | تفعيل الخيار | دمج السجلات المكررة بدل رفضها |

- لا يوجد assign button مباشر داخل هذه الشاشة.

### `/settings/closed-chats`

- الحالة: `مؤكد بصرياً`
- الغرض: مراجعة المحادثات المغلقة وإعادة فتحها.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Filters | تغيير agent/instance/page size | تحديث الجدول مع الحفاظ على الفلاتر |
| `Refresh` | إعادة التحميل | إعادة طلب الصفحة الحالية من دون تغيير الحالة |
| `Reopen` | إعادة فتح المحادثة | إرجاعها إلى inbox التشغيلي وتسجيل الحدث |
| `Previous` / `Next` | التنقل بين الصفحات | الحفاظ على استقرار pagination مع الفلاتر الحالية |

- لا يوجد assign workflow داخل هذه الشاشة.

### `/settings/tags`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Tags list | فتح الصفحة | عرض التاغات الحالية |
| Tag CRUD | إنشاء أو تعديل أو حذف | حفظ التاغات لاستخدامها في contacts/chat |

### `/settings/api-keys`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| API keys list | فتح الصفحة | عرض المفاتيح الحالية |
| Create key | إنشاء مفتاح | حفظ API key مع scope مناسب |
| Delete key | حذف مفتاح | إزالة المفتاح وتسجيل العملية عند الحاجة |

### `/settings/webhooks`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Webhooks list | فتح الصفحة | عرض webhooks الحالية |
| Event selector | اختيار event واحد أو أكثر | تحديد الاشتراكات المربوطة |
| `Webhook URL` | إدخال الرابط | حفظ target URL |
| `Secret` | إدخال قيمة اختيارية | تفعيل التوقيع عند الاستخدام |
| `Custom headers` | إدخال key/value | حفظ headers المخصصة |
| `Test webhook` | تنفيذ اختبار | إرسال طلب تجريبي وتسجيل النتيجة |
| CRUD actions | إنشاء أو تعديل أو حذف | حفظ إعدادات webhook |

### `/settings/custom-actions`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Custom actions list | فتح الصفحة | عرض الإجراءات المخصصة |
| CRUD actions | إنشاء أو تعديل أو حذف | حفظ custom actions المرتبطة بالـ UI |

### `/settings/sso`

- الحالة: `مؤكد وظيفياً`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Providers list | فتح الصفحة | عرض مزودي SSO المهيئين |
| Update provider | تعديل إعداد مزود | حفظ الإعدادات |
| Delete provider | حذف المزود | إزالة الإعداد |

### `/settings/license`

- الحالة: `مؤكد بصرياً`
- الغرض: إدارة ترخيص deployment داخل التطبيق.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| `Refresh` | إعادة تحميل bootstrap | تحديث حالة الترخيص والاستهلاك |
| Status card | فتح الصفحة | إظهار الحالة مثل `disabled`, `unlicensed`, `active`, `grace`, `locked` |
| Quota cards | عرض الصفحة | إظهار Organizations, Users/Org, WA Endpoints/Org, Storage/Org, Subscription Days |
| Usage summary | عرض بيانات المنظمات | إظهار الاستخدام الفعلي مقابل الحدود |
| `Copy HWID` | نسخ هوية الخادم | تمكين إصدار مفتاح التفعيل خارجياً |
| `Security Key` | لصق المفتاح | تجهيز طلب التفعيل أو التجديد |
| `Activate license` | تنفيذ التفعيل | تثبيت الترخيص وتحديث `license_records` |
| `Refresh status` | إعادة الفحص | سحب bootstrap الجديد بعد التفعيل |

- هذه الصفحة admin-only.

### `/license-cleanup`

- الحالة: `مؤكد بصرياً`
- الغرض: وضع محدود لحل `quota overage`.

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Redirect to cleanup | وجود overage عام | منع بقية التطبيق وتحويل المستخدم إلى cleanup route |
| Cleanup actions | حذف organizations/users/accounts/instances | خفض الاستهلاك نحو العودة للحد الطبيعي |
| Bootstrap re-check | إعادة الفحص بعد كل حذف | إذا زال overage يعود التطبيق إلى المسار العادي |

## سطوح لازمة غير مستقلة في التنقل

### Webhook delivery log

- الحالة: `مضاف لسد فجوة`
- الموقع المعتمد: modal أو tab داخل `/settings/webhooks`

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Deliveries list | فتح السطح | عرض HTTP status, attempts, next retry, response metadata |
| Manual retry | إعادة محاولة delivery فاشلة | إنشاء retry أو تحديث السجل وفق السياسة |

### Background job progress

- الحالة: `مضاف لسد فجوة`
- الموقع المعتمد: modal أو admin drawer من settings/actions

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Job status view | فتح السطح | عرض حالة cleanup/import/reconnect/webhook replay/campaign runs |
| Progress refresh | تحديث السطح | إبقاء المستخدم واعياً بحالة العمل الطويل |

### Audit log view

- الحالة: `مضاف لسد فجوة`
- الموقع المعتمد: admin-only surface تحت settings أو operations

| العنصر | الفعل | النتيجة المتوقعة |
| :--- | :--- | :--- |
| Audit list | فتح السطح | عرض license activation, role edits, API keys, instance delete/disconnect, cleanup, campaign control وغيرها من الأفعال الحساسة |

## التعارضات التي تم حسمها

### 1. اتساع التنقل الرئيسي

- الحسم: يعتمد تنقل واسع يضم `Dashboard`, `Chat`, `Chatbot`, `Agent Analytics`, `Campaigns`, `Settings`.
- السبب: التدقيق الحي أثبت أن الخطة المختزلة الأقدم كانت أضيق من المنتج الفعلي.

### 2. موقع assignment

- الحسم: assignment وunassignment يبقيان داخل chat فقط.
- السبب: لم يظهر أي assign control داخل `analytics`, `settings`, `instances`, `contacts`, أو `closed chats`.

### 3. حالة `pin/hide/read`

- الحسم: تعتبر state شخصية لكل مستخدم.
- السبب: المرجع التشغيلي والـ schema يربطانها صراحةً بـ `contact_user_states` وليس بخصائص عامة على contact.

### 4. `Conversation timeline`

- الحسم: تعتمد كسطح لازم داخل `/chat/[contactId]`.
- السبب: غيابها يكسر تتبع assignment history وqueue/resolution analytics وsupport debugging، حتى لو لم تكن route مستقلة.

### 5. مكانة `Chatbot`

- الحسم: لا تُحذف من المنتج، لكن لا تُوسّع وظيفياً خارج ما هو مثبت الآن.
- السبب: `Contact Info` تعتمد على `chatbot flow settings` لعرض panel data، بينما تفاصيل chatbot الداخلية لم تُدقق بالكامل.

### 6. `Dashboard`

- الحسم: تبقى route معتمدة، لكن كمشتق من الدومينات الأخرى لا كـ domain مستقل أو table مستقلة.
- السبب: المراجع تثبت وجودها وتثبت أيضاً أنها تبنى من aggregates تشغيلية.

### 7. `Campaigns`

- الحسم: تبقى route وdomain معتمدة، لكن UI التفصيلي لا يتجاوز create/list/detail/runs/recipients/launch/pause/resume.
- السبب: route وAPI والدومين مثبتة، أما التفاصيل البصرية الكاملة فلم تُدقق بما يكفي لاختراع عناصر إضافية.

### 8. `Vendor License Studio`

- الحسم: يبقى خارج التطبيق الرئيسي.
- السبب: هو companion service خاص بالـ vendor وله registry وguard وmetadata منفصلة.

### 9. العناصر غير المؤكدة

- الحسم: تبقى خارج هذا المرجع.
- السبب: لا توجد أدلة كافية من التدقيق الحالي على `Meta`, `templates`, `widgets`, `AI`, `SLA`, `keyword rules`, أو `transfer queue`.
