# Design: Fix Build Tags

## Approach
Remove `//go:build bootstrap` from all production API files. The `bootstrap` tag has no legitimate use in this codebase — it was a temporary scaffold marker.

## Files to Modify
For each file listed below, **remove the first two lines** (the build tag comment and the blank line after it):
1. `backend/api/router.go` — remove `//go:build bootstrap`
2. `backend/api/chats.go` — remove `//go:build bootstrap`
3. `backend/api/users.go` — remove `//go:build bootstrap`
4. `backend/api/roles.go` — remove `//go:build bootstrap`
5. `backend/api/context.go` — remove `//go:build bootstrap`
6. `backend/api/utils.go` — remove `//go:build bootstrap`

## Files to Leave Unchanged
- `backend/api/phase11_16.go` — keep `//go:build ignore` (prototype)
- `backend/api/store.go` — keep `//go:build ignore` (prototype)

## Verification
After removal, run:
```bash
cd backend && go build ./...
go vet ./...
```
Both must pass with zero errors.

## Risk
Low — removing build tags is a mechanical change with no logic modification.
