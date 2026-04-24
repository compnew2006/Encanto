# Design: WebSocket Origin Validation

## Fix Location
`backend/api/ws.go` — `NewRealtimeHub()` function

## Current Code
```go
CheckOrigin: func(r *http.Request) bool {
    return true
},
```

## Fixed Code
```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    allowed := []string{
        deps.Config.FrontendOrigin,
        "http://localhost:5173",
        "http://127.0.0.1:5173",
    }
    for _, o := range allowed {
        if origin == o {
            return true
        }
    }
    return false
},
```

## Config Integration
Pass `deps.Config` (or just `FrontendOrigin`) to `NewRealtimeHub()` so the production origin is configurable via `FRONTEND_ORIGIN` env var.

## Updated Signature
```go
func NewRealtimeHub(frontendOrigin string) *RealtimeHub
```
