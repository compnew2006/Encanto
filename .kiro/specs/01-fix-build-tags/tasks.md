# Tasks: Fix Build Tags

- [x] 1. Remove `//go:build bootstrap` from `backend/api/router.go`
- [x] 2. Remove `//go:build bootstrap` from `backend/api/chats.go`
- [x] 3. Remove `//go:build bootstrap` from `backend/api/users.go`
- [x] 4. Remove `//go:build bootstrap` from `backend/api/roles.go`
- [x] 5. Remove `//go:build bootstrap` from `backend/api/context.go`
- [x] 6. Remove `//go:build bootstrap` from `backend/api/utils.go`
- [x] 7. Run `cd backend && go build ./...` and confirm zero errors
- [x] 8. Run `cd backend && go vet ./...` and confirm zero warnings
- [x] 9. Run `cd backend && go test ./...` and confirm all tests pass
