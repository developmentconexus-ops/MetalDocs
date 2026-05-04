# Reference: How to Run Tests

> **Last verified:** 2026-05-03
> **Status:** Stub. Verify exact commands + paths against current scripts.
> **Scope:** Backend (Go), frontend (Vitest), e2e (Playwright if present).

## Backend (Go)

```powershell
# all tests
go test ./...

# single module
go test ./internal/modules/documents/...

# verbose
go test -v ./internal/modules/approval/...

# with race detector
go test -race ./...

# coverage
go test -cover ./internal/modules/...
```

## Frontend (Vitest)

```powershell
cd frontend\apps\web
npm.cmd run test           # one-shot
npm.cmd run test -- --watch
npm.cmd run test -- --ui   # vitest UI
```

## E2E (Playwright — verify presence)

```powershell
cd frontend\apps\web
npm.cmd run test:e2e
```

System acceptance test (manual e2e): see [`wiki/tests/system-acceptance-test.md`](../tests/system-acceptance-test.md).

## CI

TBD — fill in once CI pipeline is locked.

## See also

- [references/local-dev-startup.md](local-dev-startup.md)
- [`wiki/tests/system-acceptance-test.md`](../tests/system-acceptance-test.md) — full manual acceptance run, Groups A–E; derived from [workflows/user-onboarding.md](../workflows/user-onboarding.md)
