# Tasks: Password Reset Flow

- [ ] 1. Create migration `backend/db/migrations/002_password_reset.sql` with `password_reset_tokens` table
- [ ] 2. Add SMTP config keys (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`) to `config/config.go`
- [ ] 3. Add sqlc queries for `password_reset_tokens`: `InsertResetToken`, `FindResetToken`, `DeleteResetToken`, `DeleteExpiredResetTokens`
- [ ] 4. Run `sqlc generate` to regenerate `backend/data/sqlc/`
- [ ] 5. Implement `forgotPasswordHandler` in `backend/api/auth.go`
- [ ] 6. Implement `validateResetTokenHandler` in `backend/api/auth.go`
- [ ] 7. Implement `resetPasswordHandler` in `backend/api/auth.go`
- [ ] 8. Create `backend/api/email.go` with SMTP sender (fallback: log to stdout if unconfigured)
- [ ] 9. Register the 3 new routes in the router (no auth middleware)
- [ ] 10. Add token expiry cleanup to the workers polling loop
- [ ] 11. Create `frontend/src/routes/forgot-password/+page.svelte` with email form
- [ ] 12. Create `frontend/src/routes/reset-password/+page.svelte` with new password form + token validation on mount
- [ ] 13. Add "Forgot password?" link to `frontend/src/routes/login/+page.svelte`
- [ ] 14. Write E2E test in Playwright covering the full reset flow
