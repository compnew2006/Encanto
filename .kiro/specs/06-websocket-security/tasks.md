# Tasks: WebSocket Origin Validation

- [ ] 1. Update `NewRealtimeHub()` signature to accept `frontendOrigin string`
- [ ] 2. Replace `return true` with allowed-origins check (FrontendOrigin + localhost variants)
- [ ] 3. Update all call sites of `NewRealtimeHub()` to pass `config.FrontendOrigin`
- [ ] 4. Write a test verifying that a request with a foreign origin is rejected (HTTP 403)
- [ ] 5. Write a test verifying that a request with the configured origin is accepted
