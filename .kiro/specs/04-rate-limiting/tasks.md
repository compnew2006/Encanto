# Tasks: API Rate Limiting

- [ ] 1. Create `backend/api/middleware_ratelimit.go` with `RateLimitByIP()` and `RateLimitByUserID()` functions
- [ ] 2. Implement sliding window counter using Redis `INCR` + `EXPIRE`
- [ ] 3. Add `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After` headers
- [ ] 4. Return HTTP 429 with JSON error body when limit exceeded
- [ ] 5. Implement fail-open fallback when Redis is unavailable
- [ ] 6. Apply `RateLimitByIP(10, 1min)` middleware to `POST /api/auth/login` in the router
- [ ] 7. Apply `RateLimitByUserID(300, 1min)` middleware to the authenticated route group
- [ ] 8. Exclude `/api/health` from rate limiting
- [ ] 9. Add config keys `RATE_LIMIT_LOGIN_MAX` (default: 10) and `RATE_LIMIT_API_MAX` (default: 300) to `config/config.go`
- [ ] 10. Write unit tests for the rate limit middleware using a mock Redis client
