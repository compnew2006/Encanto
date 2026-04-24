# Bugfix Requirements Document

## Introduction

Six production API files in `backend/api/` carry `//go:build bootstrap` build tags. Because the `bootstrap` tag is never satisfied by a standard `go build ./...` invocation, these files are silently excluded from compilation. The result is a binary that is either uncompilable (due to missing symbols) or missing all core route handlers, making the application non-functional in any normal build environment.

The fix is purely mechanical: remove the `//go:build bootstrap` line from each affected production file. Two prototype files (`phase11_16.go`, `store.go`) carry `//go:build ignore` intentionally and must not be changed.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN `go build ./...` is run from `backend/` THEN the system excludes `router.go`, `chats.go`, `users.go`, `roles.go`, `context.go`, and `utils.go` from compilation due to unsatisfied `//go:build bootstrap` constraints

1.2 WHEN the binary is produced without the `bootstrap` tag THEN the system is missing all route registrations, context helpers, and utility functions, rendering the API non-functional

1.3 WHEN a file carries `//go:build bootstrap` THEN the system treats it as a conditional build artifact rather than a required production source file

### Expected Behavior (Correct)

2.1 WHEN `go build ./...` is run from `backend/` THEN the system SHALL compile all six production API files unconditionally, with no build tag filtering applied

2.2 WHEN the binary is produced THEN the system SHALL include all route handlers (`chats`, `users`, `roles`), the router, context helpers, and utilities, producing a fully functional API binary

2.3 WHEN a production API file has no build tag THEN the system SHALL include it in every standard build without requiring special flags or tag injection

### Unchanged Behavior (Regression Prevention)

3.1 WHEN `go build ./...` is run THEN the system SHALL CONTINUE TO exclude `backend/api/phase11_16.go` from compilation, as it carries `//go:build ignore` intentionally

3.2 WHEN `go build ./...` is run THEN the system SHALL CONTINUE TO exclude `backend/api/store.go` from compilation, as it carries `//go:build ignore` intentionally

3.3 WHEN the production files are compiled THEN the system SHALL CONTINUE TO exhibit identical runtime behavior — no logic, signatures, or data flows are altered by this fix

3.4 WHEN `go test ./...` is run THEN the system SHALL CONTINUE TO pass all existing tests without modification

---

## Bug Condition

### Bug Condition Function

```pascal
FUNCTION isBugCondition(F)
  INPUT: F of type GoSourceFile
  OUTPUT: boolean

  // Returns true when a production API file carries a build tag
  // that prevents it from being included in a standard build
  RETURN F.package = "api"
    AND F.path MATCHES "backend/api/*.go"
    AND F.buildTag = "bootstrap"
    AND F.isPrototype = false
END FUNCTION
```

Concrete instances where `isBugCondition` is true:

| File | Build Tag |
|---|---|
| `backend/api/router.go` | `//go:build bootstrap` |
| `backend/api/chats.go` | `//go:build bootstrap` |
| `backend/api/users.go` | `//go:build bootstrap` |
| `backend/api/roles.go` | `//go:build bootstrap` |
| `backend/api/context.go` | `//go:build bootstrap` |
| `backend/api/utils.go` | `//go:build bootstrap` |

### Fix Checking Property

```pascal
// Property: Fix Checking — production files must be included in standard builds
FOR ALL F WHERE isBugCondition(F) DO
  result ← buildOutput(F_fixed)   // F' = file after build tag removal
  ASSERT result.includedInBuild = true
  ASSERT result.compilationError = false
END FOR
```

### Preservation Checking Property

```pascal
// Property: Preservation — prototype files and runtime behavior must be unchanged
FOR ALL F WHERE NOT isBugCondition(F) DO
  ASSERT buildOutput(F_original) = buildOutput(F_fixed)
END FOR

// Specifically:
// phase11_16.go  → STILL excluded (//go:build ignore preserved)
// store.go       → STILL excluded (//go:build ignore preserved)
// All other backend packages → compiled output identical
```

### Root Cause

Files were tagged `//go:build bootstrap` during an initial scaffolding phase to defer compilation until the bootstrap tooling was ready. The tags were never removed before the code was merged to the main branch, violating the project convention that build tags must not appear on production files.
