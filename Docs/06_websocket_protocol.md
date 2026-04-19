# WebSocket Protocol Specification: Whatomate

هذه الوثيقة تصف بروتوكول الـ WebSocket بعد **تدقيق الواجهة الحية**.

التركيز الحالي على:

- الرسائل
- تحديثات جهات الاتصال
- حالات الرسائل
- حالات توفر المستخدمين
- حالة instances الخاصة بـ `whatsmeow`
- الإشعارات التشغيلية
- تحديثات `Statuses` و`Dashboard` عند الحاجة
- مع الإبقاء على provider-side typing simulation كمنطق backend داخلي لا يعتمد على WebSocket

## 1. Connection & Authentication

### Endpoint

- `ws://<domain>/ws`
- أو `wss://<domain>/ws` في الإنتاج
- subprotocol: `whm.v1`

### Token Acquisition

قبل الاتصال، يحصل العميل على token قصير العمر:

- `GET /api/auth/ws-token`

### Mandatory Post-Handshake Auth

```json
{
  "type": "auth",
  "payload": { "token": "JWT_TOKEN_HERE" }
}
```

## 2. Message Schema

```ts
interface WSMessage<T = unknown> {
  type: string;
  event_id?: string;
  sequence?: number;
  occurred_at?: string;
  organization_id?: string;
  payload: T;
}
```

## 3. Client -> Server Events

| Event | Payload | Description |
| :--- | :--- | :--- |
| `auth` | `{"token": string}` | Authenticate connection |
| `set_contact` | `{"contact_id": "UUID"}` | Subscribe to a specific chat context |
| `resume` | `{"last_sequence": number}` | Ask server to resume from the last acknowledged sequence when possible |
| `ping` | `{}` | Keepalive heartbeat |
| `set_dashboard_scope` | `{"organization_id": "UUID"}` | Optional dashboard subscription |

## 4. Server -> Client Events

| Event | Description | Data Context |
| :--- | :--- | :--- |
| `new_message` | New inbound or outbound message | Contact + message payload |
| `status_update` | Message status changed | Message ID + status |
| `typing_state` | Optional chat typing/presence update for UI | Contact ID + actor + state |
| `contact_update` | Contact assignment or state changed | Contact payload |
| `conversation_event` | Lifecycle event such as assign, close, reopen, or unassign | Contact ID + event payload |
| `instance_qr_code` | Pairing QR changed | Instance ID + QR string |
| `instance_connected` | Instance connected successfully | Instance ID + phone/JID |
| `instance_disconnected` | Instance disconnected | Instance ID + reason |
| `reaction_update` | Reaction changed on message | Message ID + reaction payload |
| `notification` | Generic user/org notification | Text + severity |
| `notification_read` | Notification read state changed | Notification ID + read state |
| `availability_update` | User availability changed | User ID + status (`available`, `unavailable`, `busy`) |
| `status_feed_update` | Status drawer feed changed | Status summary payload |
| `dashboard_update` | Dashboard counters/cards changed | Aggregate payload |
| `job_update` | Background job status changed | Job ID + status + progress |
| `snapshot_required` | Resume gap was too large or lost | Client should reload current route data |

## 5. Delivery Source & Ordering

- كل حدث websocket يجب أن يخرج من `outbox_events` لا من handler مباشر بعد الحفظ.
- كل channel يحصل على `sequence` متزايد يمكن للعميل تخزينه مؤقتاً.
- عند reconnect، العميل يرسل `resume`.
- إذا لم تعد events القديمة متاحة أو ظهرت فجوة في التسلسل، الخادم يرسل `snapshot_required`.
- Redis تستخدم فقط كـ fanout layer بين العقد؛ source of truth للأحداث يبقى PostgreSQL + outbox.

## 6. Room & Context Management

### Organization Room

- كل المستخدمين داخل نفس `organization_id`.
- يستخدم لأحداث مثل:
  - `new_message`
  - `contact_update`
  - `notification`
  - `status_feed_update`

### User Room

- كل التبويبات المفتوحة لنفس المستخدم.
- يستخدم للإشعارات الخاصة بالمستخدم أو حالات availability (`available`, `unavailable`, `busy`).

### Contact Room

- يفعّل عندما يرسل العميل `set_contact`.
- يستخدم لتقليل الضوضاء في المحادثات غير المفتوحة حالياً.

## 7. Frontend Resilience

عميل `SvelteKit` يتعامل مع الانقطاع كالتالي:

- `ping` كل 30 ثانية.
- إعادة ربط بمحاولات متدرجة.
- بعد نجاح إعادة الربط:
  - إرسال `resume` بآخر `sequence` معروف
  - إعادة الاشتراك في contact الحالي
  - إعادة تحميل notifications وstatuses عند الحاجة

## 8. Explicitly Removed Or Unverified Events

ملاحظة:

- محاكاة `typing event` المطلوبة قبل إرسال النصوص عبر `whatsmeow` تعتبر سلوك backend-to-provider داخلي.
- لذلك لا تعتمد الخطة على WebSocket لكي تصل إلى واتساب نفسه، حتى لو اختير لاحقاً بث `typing_state` داخل الواجهة المحلية.

لم يتم تأكيد هذه الأحداث من التدقيق الحالي:

- `meta_provider_update`
- `template_sync_update`
- `widget_update`
- `ai_reply_generated`
- `agent_transfer`
- `sla_breach`
