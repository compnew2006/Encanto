# Requirements: Password Reset Flow

## Overview
No mechanism exists for users to reset a forgotten password. This is a blocking usability gap for any production deployment.

## Requirements

### REQ-1: Reset Request
WHEN a user submits their email to `POST /api/auth/forgot-password`
THE SYSTEM SHALL generate a secure random token, store it in `password_reset_tokens` with a 1-hour expiry
AND send a reset email to that address (if the user exists)
AND return HTTP 200 regardless of whether the email exists (to prevent user enumeration)

### REQ-2: Token Validation
WHEN a user submits a token to `GET /api/auth/reset-password?token=<token>`
THE SYSTEM SHALL return HTTP 200 with `{ "valid": true }` if the token exists and has not expired
AND return HTTP 400 with `{ "valid": false }` if the token is invalid or expired

### REQ-3: Password Update
WHEN a user submits a new password and valid token to `POST /api/auth/reset-password`
THE SYSTEM SHALL validate the token, hash the new password with bcrypt, update the user record
AND delete the used token from `password_reset_tokens`
AND return HTTP 200

WHEN the token is expired or invalid
THE SYSTEM SHALL return HTTP 400 with a descriptive error message

### REQ-4: Token Expiry Cleanup
WHEN the workers package polls for jobs
THE SYSTEM SHALL also delete `password_reset_tokens` WHERE `expires_at < NOW()`

### REQ-5: Frontend Reset Flow
WHEN a user clicks "Forgot password?" on the login page
THE SYSTEM SHALL show a form to enter their email and display a confirmation message after submission

WHEN a user visits `/reset-password?token=<token>`
THE SYSTEM SHALL show a form to enter and confirm a new password
