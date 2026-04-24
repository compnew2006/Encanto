# Design: Resolve Migration Conflict

## Step 1: Compare the Two Files
Diff `000001_phase_1_4.sql` vs `001_schema.sql`:
- If `001_schema.sql` is a superset (contains everything in `000001_phase_1_4.sql` plus more), then `000001_phase_1_4.sql` is redundant.
- Confirm there are no tables/columns in `000001_phase_1_4.sql` that are NOT in `001_schema.sql`.

## Step 2: Remove the Duplicate
Delete `backend/db/migrations/000001_phase_1_4.sql`.

## Step 3: Verify Migration Runner Order
Confirm that `backend/data/migrations.go` applies files in alphabetical/numeric order. With only `001_schema.sql` present, the runner applies it once cleanly.

## Step 4: Test on Fresh DB
```bash
docker compose up -d postgres
go run main.go  # AUTO_MIGRATE=true
# Expect: migrations applied, server starts, /api/health returns 200
```

## Note on Already-Applied DBs
If a development DB already has `000001_phase_1_4.sql` applied, dropping and recreating is the cleanest path. For production, this is a non-issue since no production DB exists yet.
