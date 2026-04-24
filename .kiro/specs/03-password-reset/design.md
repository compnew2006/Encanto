# Design: Password Reset Flow

## New DB Migration
File: `backend/db/migrations/002_password_reset.sql`
```sql
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON password_reset_tokens(token_hash);
CREATE INDEX ON password_reset_tokens(expires_at);
```

## Token Generation
- Use `crypto/rand` to generate 32 bytes → hex-encode → store SHA-256 hash in DB
- Send raw token in the email link: `{FRONTEND_URL}/reset-password?token=<raw>`

## New API Endpoints (in `backend/api/auth.go`)
| Method | Path | Handler |
|---|---|---|
| POST | `/api/auth/forgot-password` | `forgotPasswordHandler` |
| GET | `/api/auth/reset-password` | `validateResetTokenHandler` |
| POST | `/api/auth/reset-password` | `resetPasswordHandler` |

## Email Delivery
- Use `net/smtp` or an env-configured SMTP service
- New config keys: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`
- If SMTP is not configured, log the reset link to stdout (dev mode)

## Frontend Routes (SvelteKit)
- `frontend/src/routes/forgot-password/+page.svelte`
- `frontend/src/routes/reset-password/+page.svelte` (reads `?token` from URL)

## Sequence Diagram
```
User → POST /api/auth/forgot-password {email}
Backend → DB: find user by email
Backend → DB: INSERT password_reset_tokens (token_hash, expires_at=+1h)
Backend → SMTP: send email with raw token link
Backend → User: HTTP 200 (always)

User → GET /api/auth/reset-password?token=xxx
Backend → DB: find token by SHA256(xxx), check expires_at
Backend → User: { valid: true/false }

User → POST /api/auth/reset-password {token, new_password}
Backend → DB: validate token
Backend → DB: UPDATE users SET password_hash=bcrypt(new_password)
Backend → DB: DELETE FROM password_reset_tokens WHERE token_hash=...
Backend → User: HTTP 200
```
