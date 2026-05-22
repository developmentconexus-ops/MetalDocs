# Go Backend Review — 2026-05-21

**Initiative:** Multi-session ECC review of the MetalDocs Go backend.
**Started:** 2026-05-21
**Mode:** Append-only. One row per module. Findings live in `2026-05-21-go-backend-review/<module>.md`.
**Cursor:** see `MEMORY.md` → `project_go_backend_review` for next-up module.

## Scope

- In: all Go code under `apps/api/` and `internal/` (hand-written).
- Out: generated code (`*/api/api.gen.go`, `frontend/apps/web/src/lib/api-types/`), vendored editor (`frontend/apps/web/eigenpal/`), wiki/spec docs unless code contradicts them.
- Tests reviewed alongside owning module.

## Severity Convention

| Level    | Meaning                          | Action                                |
|----------|----------------------------------|---------------------------------------|
| Critical | Security flaw or data-loss risk  | Open issue + spawn fix task at once   |
| High     | Bug or correctness gap           | Track in module backlog               |
| Medium   | Maintainability / design smell   | Note for opportunistic refactor       |
| Low      | Style / minor                    | Note only                             |

## Per-Session Loop

1. Read `MEMORY.md` → `project_go_backend_review` → confirm cursor.
2. Pick next `Pending` module in table below.
3. Run `metaldocs-module-doc` skill to confirm module surface.
4. Spawn ECC agents in parallel (single message, multiple Agent blocks):
   - `ecc:go-reviewer` (idioms, errors, concurrency)
   - `ecc:security-reviewer` (OWASP, auth boundary)
   - `ecc:database-reviewer` (only if module touches Postgres)
   - `ecc:silent-failure-hunter` (swallowed errors)
   - `ecc:type-design-analyzer` (invariants, encapsulation)
5. Consolidate digests → append to `2026-05-21-go-backend-review/<module>.md`.
6. Update tracker row: status, severity counts, date, link.
7. Update cursor in `MEMORY.md`.
8. Commit: `docs(review): <module> findings`.

## Status Legend

- `Pending` — not started
- `In Progress` — agents dispatched, findings being consolidated
- `Done` — findings committed, tracker row updated
- `Skipped` — explicit decision, reason in findings file

## Module Tracker

| #  | Module                                          | Status  | Critical | High | Medium | Low | Reviewer | Date | Findings |
|----|-------------------------------------------------|---------|----------|------|--------|-----|----------|------|----------|
| 1  | `apps/api/cmd/metaldocs-api`                    | Done    | 4        | 5    | 6      | 6   | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer | 2026-05-21 | [cmd-metaldocs-api.md](2026-05-21-go-backend-review/cmd-metaldocs-api.md) |
| 2a | `platform/{authn,security,idempotency,ratelimit,tenant,problem,httpresponse}` | Done | 5 | 11 | 11 | 8 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-21 | [platform-2a-security.md](2026-05-21-go-backend-review/platform-2a-security.md) |
| 2b | `platform/{db,migrate,bootstrap,objectstore,storage,messaging,servicebus,jobs,worker}` | Done | 10 | 24 | 16 | 8 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [platform-2b-data-infra.md](2026-05-21-go-backend-review/platform-2b-data-infra.md) |
| 2c | `platform/{config,observability,cache,featureflags,formval,httpclient,pagination,docgenv2,render}` | Done | 1 | 14 | 25 | 18 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [platform-2c-support-observability.md](2026-05-21-go-backend-review/platform-2c-support-observability.md) |
| 3  | `internal/modules/auth`                         | Pending | -        | -    | -      | -   | -        | -    | -        |
| 4  | `internal/modules/iam`                          | Pending | -        | -    | -      | -   | -        | -    | -        |
| 5  | `internal/modules/documents`                    | Pending | -        | -    | -      | -   | -        | -    | -        |
| 6  | `internal/modules/controlleddocuments`          | Pending | -        | -    | -      | -   | -        | -    | -        |
| 7  | `internal/modules/taxonomy`                     | Pending | -        | -    | -      | -   | -        | -    | -        |
| 8  | `internal/modules/templates`                    | Pending | -        | -    | -      | -   | -        | -    | -        |
| 9  | `internal/modules/audit`                        | Pending | -        | -    | -      | -   | -        | -    | -        |
| 10 | `internal/modules/render`                       | Pending | -        | -    | -      | -   | -        | -    | -        |
| 11 | `internal/modules/search`                       | Pending | -        | -    | -      | -   | -        | -    | -        |
| 12 | `internal/modules/jobs`                         | Pending | -        | -    | -      | -   | -        | -    | -        |
| 13 | Shared infra (`internal/test`, `objectstore`, `docgenv2`, fanout) | Pending | - | - | - | - | - | - | - |

## Notes

- Module order revised from initial plan: `approval` module does not exist in repo. Added `render`, `search`, `jobs` (real modules under `internal/modules/`).
- Platform row split into #2a/#2b/#2c on 2026-05-21 — 25 packages / ~5200 LoC too big for a single agent pass. Grouped by concern: security boundary (#2a), data + infra (#2b), support + observability (#2c).

## Critical Backlog (G3 handoff)

Per plan §6 G3: each Critical needs owner + ETA + reserved fix-branch before cursor advances. ETA = TBC means owner to set on branch-cut. Land in stated order — `fix/migrate-2b-c6-c7` first (silent data bomb, standalone, smallest).

### Module #2a (carried open)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 2a-C3 | idempotency two-phase (in progress on branch) | Critical | leandrotca | TBC | `fix/h11-idempotency-schema-v2` | WIP |
| 2a-C4 | idempotency two-phase (in progress on branch) | Critical | leandrotca | TBC | `fix/h11-idempotency-schema-v2` | WIP |

### Module #2b (10 Criticals, 4 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 2b-C6 | `migrations/0042_*` duplicate prefix | Critical | leandrotca | TBC | `fix/migrate-2b-c6-c7` | Backlog (land first) |
| 2b-C7 | `migrations/0130_*` duplicate prefix | Critical | leandrotca | TBC | `fix/migrate-2b-c6-c7` | Backlog (land first) |
| 2b-C1 | `internal/platform/migrate/migrate.go:74-78` swallow | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C5 | `internal/platform/migrate/migrate.go:24-69` no advisory lock | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C8 | `migrations/0176_pdf_dispatch_outbox.sql:1-24` no tx | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C9 | `migrations/0111_docx_v2_exports.sql:1,6` unqualified FK + no tx | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C2 | `internal/platform/storage/local/store.go:20,31,40` path containment | Critical | leandrotca | TBC | `fix/storage-2b-c2` | Backlog |
| 2b-C3 | `internal/platform/config/docgen_v2.go:20-32` SSRF | Critical | leandrotca | TBC | `fix/docgen-2b-c3-c4` | Backlog |
| 2b-C4 | `internal/platform/config/docgen_v2.go:24` empty token | Critical | leandrotca | TBC | `fix/docgen-2b-c3-c4` | Backlog |
| 2b-C10 | `internal/platform/messaging/events.go:7-18` typed boundary | Critical | leandrotca | TBC | `fix/messaging-2b-c10` | Backlog (cascades H14, H15, H19) |

### Module #2c (1 Critical, 1 fix branch reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 2c-C1 | `internal/platform/pagination/cursor.go:37-43` HMAC-less cursor, anchor tamperable | Critical | leandrotca | TBC | `fix/cursor-2c-c1` | Backlog (cascades H7, H8, H9, M26, M27, L15, L16 — coordinated cursor rewrite) |
