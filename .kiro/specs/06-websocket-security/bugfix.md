# Bugfix Spec: WebSocket CheckOrigin Allows Any Domain

## Bug Summary
`backend/api/ws.go` initializes the WebSocket upgrader with `CheckOrigin: func(r *http.Request) bool { return true }`. This disables the browser's same-origin protection for WebSocket connections, enabling CSRF attacks from any malicious website.

## Current Behavior
Any webpage on any domain can open a WebSocket connection to the Encanto server and receive real-time events from authenticated sessions.

## Expected Behavior
Only requests originating from the configured `FRONTEND_ORIGIN` (and localhost in dev mode) should be allowed to upgrade to WebSocket.

## Unchanged Behavior
- WebSocket authentication via JWT token remains unchanged
- All existing WebSocket message types and handlers remain unchanged

## Root Cause
`return true` was used as a placeholder during development and was never replaced with a proper origin check.
