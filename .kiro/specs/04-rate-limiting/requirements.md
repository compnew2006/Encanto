# Requirements: API Rate Limiting

## Overview
The API has no rate limiting. The `/api/auth/login` endpoint is vulnerable to brute-force attacks. All API endpoints are vulnerable to abuse.

## Requirements

### REQ-1: Login Brute-Force Protection
WHEN more than 10 login attempts are made from the same IP within 1 minute
THE SYSTEM SHALL return HTTP 429 with `{ "error": "Too many login attempts. Try again in X seconds." }`
AND include a `Retry-After` header

### REQ-2: General API Rate Limit
WHEN any authenticated API endpoint receives more than 300 requests per minute from the same user ID
THE SYSTEM SHALL return HTTP 429

### REQ-3: Rate Limit Headers
WHEN any rate-limited endpoint responds
THE SYSTEM SHALL include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` headers in every response

### REQ-4: Whitelist Internal Health Check
WHEN a request hits `GET /api/health`
THE SYSTEM SHALL NOT apply rate limiting
