# نموذج الصلاحيات القائم على الأفعال (Action-Based Permissions Model)

هذه الوثيقة تصف نموذج الصلاحيات المعتمد في نظام Whatomate. بدلاً من الاعتماد على أسماء الأدوار المغلقة (مثل Manager أو Agent)، يعتمد النظام على مصفوفة صلاحيات دقيقة مبنية على "الأفعال" (Actions) و"الموارد" (Resources)، مما يمنح مرونة تامة لتغطية كافة الحالات وربطها مباشرة بالواجهات والخوادم.

## 1. بنية الصلاحية (Permission Structure)

كل صلاحية تتكون من مقطعين أساسيين: `[Resource].[Action]`
- **Resource (المورد):** الكيان المراد التحكم به (مثال: `contacts`, `messages`, `instances`).
- **Action (الفعل):** الإجراء المراد تنفيذه (مثال: `view`, `create`, `edit`, `delete`, `send`).

### الأفعال المعيارية (Standard Actions):
- `view`: يسمح بقراءة وعرض البيانات (القراءة).
- `create`: يسمح بإضافة كيانات جديدة (الإنشاء).
- `edit`: يسمح بتعديل كيانات موجودة (التعديل).
- `delete`: يسمح بحذف كيانات موجودة (الحذف).

## 2. تحقيق حالة "المشاهدة بلا كتابة" (Read-Only Mode)

يتم تحقيق الحالات التي تسمح بالمشاهدة وتمنع الكتابة من خلال **الفصل الصريح للأفعال**. 
على سبيل المثال، للمشاهدة فقط، يتم منح المستخدم صلاحية الـ `view` وسحب صلاحيات الـ `create`, `edit`, `delete`. 

**مثال تطبيقي في المحادثات:**
- مشاهدة المحادثات فقط: `chats.view`
- منع الكتابة (إرسال رسائل): عدم وجود `messages.send`
- النتيجة: يستطيع المستخدم الدخول إلى `/chat/[contactId]` ورؤية الرسائل السابقة، لكن حقل إدخال الرسالة (Composer) وزر الإرسال سيكونان مخفيين (أو Disabled).

## 3. مصفوفة الصلاحيات المعتمدة (Permission Matrix)

فيما يلي قائمة الصلاحيات وربطها المباشر بالأفعال والشاشات:

### نطاق جهات الاتصال والمحادثات (Contacts & Chats)
| رمز الصلاحية | الوصف الدقيق (Action) | الربط مع الواجهة / الزر |
| :--- | :--- | :--- |
| `contacts.view` | قراءة قائمة جهات الاتصال | دخول صفحة `/settings/contacts`، ظهور صف الجهة |
| `contacts.create` | إضافة جهة جديدة / استيراد CSV | تفعيل أزرار `Add Contact` و `Import CSV` |
| `contacts.edit` | تعديل اسم أو تفاصيل الجهة | تفعيل زر `Edit` في لوحة `Contact Info` |
| `contacts.delete` | إزالة جهة الاتصال بشكل كامل | تفعيل زر `Delete` للجهة |
| `contacts.export` | رفع وتصدير قائمة الجهات | زر `Export CSV` |

### نطاق الرسائل والمحادثات (Messages & Inbox)
| رمز الصلاحية | الوصف الدقيق (Action) | الربط مع الواجهة / الزر |
| :--- | :--- | :--- |
| `chats.view` | القدرة على فتح أي محادثة | النقر على row في الـ inbox، رؤية المحادثة المفتوحة |
| `messages.send` | القدرة على إرسال رسائل جديدة أو ملفات | ظهور تفعيل الـ Composer وزر الإرسال `Send` المباشر |
| `messages.revoke` | القدرة على تراجع / سحب رسالة | خيار `Revoke message` ضمن خيارات الرسالة الفردية |
| `notes.view` | قراءة الملاحظات الداخلية | إظهار التبويب `Notes` ومحتوياته داخل المحادثة |
| `notes.create` | كتابة ملاحظة داخلية للزملاء | تفعيل حقل الإدخال وزر "إضافة ملاحظة" داخل تبويب `Notes` |
| `chats.unclaimed.view` | رؤية المحادثات في طابور الانتظار (Pending) | رؤية تبويب المحادثات غير المسندة `Pending` |
| `chats.unclaimed.send` | التدخل بكتابة رسالة قبل الإسناد الرسمي | التحكم بتفعيل Composer الخاص بالمحادثات المنتظرة |

### نطاق تحديد الرؤية (Visibility Scopes)
هذه الصلاحيات هي "معدلات" (Modifiers) تحدد نطاق الـ `view` داخل `contacts.view` أو `chats.view`:
- `contacts.scope.all`: رؤية تفاصيل كافة المحادثات / الجهات بالمنظمة.
- `contacts.scope.instance_only`: رؤية المحادثات المرتبطة فقط بـ Instances المسموحة للمستخدم.
- `contacts.scope.allowed_numbers`: الرؤية مقيدة حصراً بأرقام محددة.

### نطاق التشغيل والإعدادات (Operations & Settings)
| رمز الصلاحية | الوصف الدقيق | الربط |
| :--- | :--- | :--- |
| `instances.view` | رؤية حالة الربط وحسابات WhatsApp | دخول المشاهدة في `/settings/instances` |
| `instances.manage` | إضافة، تعديل، فصل، ومسح حسابات WhatsApp | تفعيل أزرار `Add Account`, `Edit`, `Delete`, `Disconnect` |
| `campaigns.view` | الاطلاع على قائمة وتاريخ الحملات | رؤية جدول الحملات وتقارير `Runs` |
| `campaigns.create` | تخطيط أو إنشاء حملات جديدة | تفعيل `Create campaign` |
| `campaigns.launch` | التنفيذ المباشر (إطلاق الحملة) | تفعيل أزرار `Launch`, `Pause`, `Resume` |
| `api_keys.manage` | إدارة الـ API Keys | القدرة الكاملة لإضافة وإنهاء مفاتيح برمجية |
| `webhooks.manage` | تكوين وإدارة الـ Webhooks | صلاحيات تعديل نقاط الربط |
| `settings.manage` | تعديل إعدادات المنظمة العامة (اللغة الوصف..) | الوصول والحفظ لتبويب `General`, `Chat settings` |
| `settings.uploads_cleanup.manage` | حذف الملفات المرفوعة (Admin Action) | تفعيل جدولة وضغط زر `Run Cleanup Now` الفوري |
| `roles.manage` | إنشاء وتعديل مصفوفات الأدوار | القدرة على إجراء التغييرات في `/settings/roles` |

## 4. الربط المباشر مع مسارات واجهة الاستخدام (UI Binding)

جميع الشاشات والأزرار سيتم حمايتها برمجياً استناداً إلى هذا النموذج. في واجهة المستخدم (الـ Frontend)، لا نتحقق من اسم الدور `role_name === 'admin'`، بل نتحقق من الفعل المسموح:

مثال تنفيذي للأزرار:
```html
<!-- سيظهر حقل الإدخال وزر الإرسال فقط إذا كان مقيّم الصلاحيات يحتوي messages.send -->
<div v-if="hasPermission('messages.send')" class="message-composer">
  <input type="text" placeholder="Type a message..." />
  <button>Send</button>
</div>
<div v-else class="read-only-notice">
  ليس لديك صلاحيات للرد في هذه المحادثة
</div>
```

مثال تنفيذي للشاشات (الروابط):
```javascript
// حماية الدخول إلى مسار الحملات
{
  path: '/campaigns',
  component: CampaignsView,
  meta: { requiresPermission: 'campaigns.view' }
}
```

## 5. الربط المستند على الـ API (Server & Endpoints Binding)

في واجهة الخادم (Backend)، يقوم نظام التحقق من الجلسة (Middleware) بتحليل الـ Tokens وإسناد الصلاحية كالتالي:

| مسار الطلب (API Endpoint) | الفعل (HTTP Method) | الصلاحية المطلوبة (Required Permission) |
| :--- | :--- | :--- |
| `/api/contacts` | `GET` | `contacts.view` |
| `/api/contacts` | `POST` | `contacts.create` |
| `/api/contacts/{id}` | `PUT` | `contacts.edit` |
| `/api/contacts/{id}` | `DELETE` | `contacts.delete` |
| `/api/chats/{id}/messages` | `POST` | `messages.send` |
| `/api/settings/instances` | `POST` / `DELETE` | `instances.manage` |

بهذا النموذج، يمكن تركيب أو تخصيص أي دور (Custom Role) بشكل ديناميكي بحيث يغطي الصلاحيات بصورة مرنة ومستقلة لكل شاشة وزر من دون الحاجة إلى تعديل الكود البرمجي مستقبلاً.
