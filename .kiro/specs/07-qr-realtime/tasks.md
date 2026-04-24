# Tasks: QR Code Real-Time Delivery

- [ ] 1. In `whatsapp_manager.go` `handleEvent()`, handle `*events.QR` event type
- [ ] 2. On QR event: update `whatsapp_instances.qr_code` in DB
- [ ] 3. On QR event: call `hub.Broadcast()` with `instance.qr_updated` message (instance_id, qr_code, expires_at)
- [ ] 4. Handle `*events.Connected` event: update instance status to `connected` in DB
- [ ] 5. On Connected event: broadcast `instance.status_changed` with `{ instance_id, status: "connected" }`
- [ ] 6. Add `qrcode` npm package to frontend: `npm install qrcode`
- [ ] 7. Update `frontend/src/routes/settings/instances/+page.svelte` to open QR modal on "Connect" button click
- [ ] 8. In QR modal: subscribe to WebSocket and handle `instance.qr_updated` to re-render QR
- [ ] 9. Add 60-second countdown timer in QR modal
- [ ] 10. Handle `instance.status_changed` to auto-close modal and show success toast
