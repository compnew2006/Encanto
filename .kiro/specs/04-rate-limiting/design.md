# Design: API Rate Limiting

## Implementation Strategy
Use Redis (already available via `backend/cache/redis.go`) with a **sliding window** counter:
- Key pattern: `ratelimit:{type}:{identifier}` (e.g., `ratelimit:login:192.168.1.1`)
- Use `INCR` + `EXPIRE` for atomic counter increment

## Middleware: `backend/api/middleware_ratelimit.go`
```go
func RateLimitByIP(redisClient *cache.Client, limit int, window time.Duration) func(http.Handler) http.Handler
func RateLimitByUserID(redisClient *cache.Client, limit int, window time.Duration) func(http.Handler) http.Handler
```

## Router Integration
```go
// Login: 10 requests/minute per IP
router.With(RateLimitByIP(deps.Cache, 10, time.Minute)).Post("/api/auth/login", loginHandler(deps))

// All authenticated routes: 300 requests/minute per user
protected.Use(RateLimitByUserID(deps.Cache, 300, time.Minute))
```

## Response on Limit Exceeded
```json
HTTP 429 Too Many Requests
Retry-After: 45
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1714000000

{ "error": "Too many requests. Try again in 45 seconds." }
```

## Fallback
If Redis is unavailable, log a warning and allow the request (fail-open to prevent outage).
