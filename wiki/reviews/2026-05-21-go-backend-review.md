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
| 2  | `internal/platform/*`                           | Pending | -        | -    | -      | -   | -        | -    | -        |
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
- Platform packages reviewed as a single batch (#2) because they share orchestration concerns; if any one balloons in findings, split into its own row.
