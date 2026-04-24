# Bugfix Spec: Duplicate Migration Files Cause Startup Failure

## Bug Summary
Two migration files exist that both define core tables like `organizations`, `users`, and `contacts`:
- `backend/db/migrations/000001_phase_1_4.sql`
- `backend/db/migrations/001_schema.sql`

Running both migrations in sequence causes `ERROR: relation already exists`.

## Current Behavior
On a fresh database, the migration runner applies both files, fails on duplicate `CREATE TABLE` statements, and the server cannot start.

## Expected Behavior
Exactly one canonical migration file defines the initial schema. All subsequent changes are in sequentially numbered files (002, 003, etc.).

## Unchanged Behavior
The schema content defined in `001_schema.sql` is the source of truth and must be preserved exactly.

## Root Cause
`000001_phase_1_4.sql` is a leftover from an earlier scaffolding phase. Its content is a subset of `001_schema.sql`.
