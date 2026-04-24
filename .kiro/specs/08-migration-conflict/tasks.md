# Tasks: Resolve Migration Conflict

- [ ] 1. Diff `000001_phase_1_4.sql` vs `001_schema.sql` — confirm `001_schema.sql` is a superset
- [ ] 2. Delete `backend/db/migrations/000001_phase_1_4.sql`
- [ ] 3. Review `backend/data/migrations.go` — confirm it reads migration files in sorted order
- [ ] 4. Drop local dev database and recreate: `dropdb encanto && createdb encanto`
- [ ] 5. Run `go run main.go` with `AUTO_MIGRATE=true` and confirm clean startup
- [ ] 6. Verify all tables exist: `\dt` in psql should show organizations, users, contacts, messages, etc.
- [ ] 7. Run `go test ./...` to confirm nothing references the deleted migration file
