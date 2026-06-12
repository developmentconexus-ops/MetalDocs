# Wave 2 Fallback Elimination & Single-Mode — Design Spec

> **Date:** 2026-06-12
> **Status:** APPROVED (brainstorm complete, user sign-off in session)
> **Owners:** Leandro (decisions) · Claude (execution)
> **Scope:** Roadmap rows **2.12** and **2.13** of the backend professionalization program (`wiki/backend/roadmap.md`). Extends the Wave 2 card; the Wave 2 close-out gate blocks on both rows. Parent contract: `docs/superpowers/specs/2026-06-11-backend-professionalization-design.md` (D-1..D-5 remain binding).
> **Out of scope:** Wave 3 trigger-gated items (RLS expansion, OTel), frontend, anything not listed in §3/§4.

---

## 1. Problem

Wave 2 items 2.1–2.11 landed, but the review pass identified surviving **fallback paths** — deprecated escape hatches and silent degradation branches. Root-cause analysis shows most share one cause: **dual-mode production code** — services carry a "no-database mode" so unit tests can construct them without Postgres, even though memory mode is dead in production (`main.go` fatal-exits for any non-postgres repository mode; F-08). Each fallback cleaned so far (controlled-documents govLogger 2.11, TenantMemberChecker warn-log) was an instance of this family. Fixing instances without killing the pattern keeps minting new ones.

A second, smaller family is **dead schema**: columns/tables verified dead in Stage-1 that code stopped writing but migrations never dropped.

## 2. Decisions (binding)

| ID | Decision | Rationale |
|----|----------|-----------|
| FE-1 | **Transactional seam = narrow `db.Tx` interface** in `internal/platform/db`. Domain ports never see `*sql.Tx`. | Industry-standard hexagonal seam; `*sql.Tx` satisfies it structurally (zero adapters); future pgx swap is wiring-only. |
| FE-2 | **Seam rollout scope = all audit/governance/authz-path `*Tx` ports** (Wave 2 additions + pre-existing same-module ports). Not a full-repo sweep. | One convention on every cross-module transactional seam; infra-internal `*sql.Tx` use is fine. |
| FE-3 | **`UserActiveInTenant` joins the `RoleProvider` interface.** `TenantMemberChecker` sibling port, type-assert wiring, and `ListUsers` full-scan fallback are deleted. | RoleProvider is already the IAM read port; every implementor must implement membership; no silent degradation path can exist. |
| FE-4 | **Full single-mode: production services are Postgres-only.** All `db == nil` / port-nil branches deleted; `RepositoryMemory` production path deleted; memory repositories become test fixtures; tests inject fakes at ports. | The fallback *class* dies, not just instances. Memory mode is already unreachable in production (F-08). |
| FE-5 | **Full verified-dead schema sweep** in one migration (0236): `templates.areas/visibility/specific_areas`, `document_profiles.is_active`, `document_subjects` table. | Definitive end of registered dead schema; precedent: 0231 dead-schema migration. |
| FE-6 | **Both seams CI-locked**: cilint `nosqltxindomain` (no `*sql.Tx` params in `internal/modules/**/domain/`) and `nodualmode` (no `db == nil` conditionals in `internal/modules/**/application/`). Empty frozen baselines. | Fixed classes of error become machine-blocked, not memory-dependent (parent spec §9.3). |
| FE-7 | Lands **inside Wave 2** as rows 2.12 + 2.13; wave close-out blocks on them. | The wave that introduced the `*Tx` ports also lands the clean shape; no half-finished state crosses the review gate. |

## 3. Row 2.12 — Fallback elimination (ports)

### 3.1 `db.Tx` / `db.DB` interfaces (commit 1)

New file `internal/platform/db/tx.go`:

```go
// DB runs SQL outside a transaction. *sql.DB and *sql.Tx satisfy it.
type DB interface {
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Tx runs SQL inside a transaction. *sql.Tx satisfies it.
// Structurally identical to DB today; the names document the contract
// (a port taking Tx REQUIRES transactional context) and let Tx grow
// independently (savepoints) later.
type Tx interface { /* same three methods */ }
```

Compile-assert test: `var _ db.Tx = (*sql.Tx)(nil)` etc. No callers yet.

### 3.2 Domain ports speak `db.Tx`; Unwrap chain deleted (commit 2)

- All FE-2-scope `*Tx` port methods change `*sql.Tx` → `db.Tx`: taxonomy `GovernanceLogger.LogTx`, `*Repository.CreateTx`/`UpdateTx`-family; audit `Writer.RecordTx`; documents `WriteTx`/`ForceReleaseSessionTx`/`MarkArchivedTx`; templates `AppendAuditTx`; approval repository tx methods; controlled-documents `LogTx`.
- **Deleted:** `sqlTxFromFamilyTx` + inline `unwrapper` interface (`taxonomy/application/governance_payload.go`), `taxonomyTx.Unwrap()`, `familyTx.Unwrap()`. Application services pass the live `*sql.Tx` straight through (structural satisfaction).
- `database/sql` import leaves `internal/modules/taxonomy/domain/port.go` (closes 2.2 reviewer finding M-1).
- Behavior change: none. Pure type-level refactor; compiler is the gate.

### 3.3 `DBGovernanceLogger` deletion (commit 3)

- `taxonomy/module.go`: delete the `deps.AuditWriter == nil` fallback branch; add fail-loud guard `panic("taxonomy.module: AuditWriter is required")` (pattern: 2.11 controlled-documents). Drop the import that pulled the deprecated type.
- Delete `internal/modules/taxonomy/application/governance_logger.go` entirely (`DBGovernanceLogger`, `NewDBGovernanceLogger`, `Log`, `LogTx`, the `// Deprecated:` notice).
- `AuditGovernanceAdapter` is the **only** `GovernanceLogger` implementation. Its nil-tx error guard stays (contract assertion, not fallback).
- Pre-deletion gate: grep proves zero non-test references; surprises = stop and report.
- Tests constructing taxonomy services with nil AuditWriter switch to the existing in-memory audit writer fake.

### 3.4 `RoleProvider.UserActiveInTenant` (commit 4)

- `iam/domain/port.go`: add `UserActiveInTenant(ctx, tenantID, userID string) (bool, error)` to `RoleProvider`, doc-commented as "identical semantics to appearing in ListUsers(tenantID)".
- Implementors (compile is the gate): postgres provider (EXISTS query from 2.9 moves under the interface); `CachedRoleProvider` delegates uncached (boolean, deactivation-sensitive — one-line comment states the no-cache decision); memory repository implements honestly **including the tenantID predicate** (also fixes the memory `RolesByUserIDs` ignoring tenantID — review finding M-6, one line, same file); `DevRoleProvider`; test mocks as compile errors surface them.
- `PeopleService`: `VerifyUserInTenant` calls `s.roles.UserActiveInTenant` directly. **Deleted:** `TenantMemberChecker` interface, `tenantChecker` field, `WithTenantMemberChecker`, the `ListUsers` full-scan fallback, and the `main.go` type-assert block + warn-log.

## 4. Row 2.13 — Single-mode + dead schema

### 4.1 Fact-check gate (commit-less, before any deletion)

Read-only verification, results recorded in the roadmap row:
1. `RepositoryMemory` reachable only from tests (grep `RepositoryMemory`, read bootstrap/config call graph; check scripts/CI for memory-mode launches).
2. Per dead-schema artifact: fresh zero-reference grep (Go + SQL + frontend) for `areas`/`visibility`/`specific_areas` (templates table), `is_active` (document_profiles), `document_subjects`.
3. Inventory every `db == nil` / port-nil branch in `internal/modules/**/application/` (grep) — the deletion worklist, attached to the row.

Any surprise (live reference found) = stop and report; do not improvise.

### 4.2 Single-mode refactor (commit 5)

- Delete every dual-mode branch on the §4.1 worklist: templates application (autosave/create/lifecycle/schema/approval_config no-db else-branches **and their `//cilint:allow-post-commit-audit` directives**), documents service (ForceRelease/Archive else-paths), controlled-documents `Create` best-effort loop disposition re-checked (post-commit loop was 2.11-annotated; under single-mode it must either move in-tx or carry a written rationale — decide from the code, report which).
- `auth.Service`: `loginCtxPort` becomes a **required** constructor dependency (fail-loud); the `!= nil` guard dies; tests inject a fake port.
- Reauth: PG limiter wired unconditionally; **delete `InMemoryAuthFailureRateLimiter`**; its window-reset unit test rewrites against the Postgres impl via sqlmock backdated rows (live integration probe already covers reset end-to-end).
- Delete the `RepositoryMemory` config value + bootstrap branch. Memory repositories are kept but only test code constructs them (move under a `testsupport` package or document as test fixtures — match repo conventions; report choice).
- Services whose constructors previously tolerated nil deps now validate fail-loud, matching the module-construction conventions established in 2.11/3.3.

### 4.3 Dead-schema migration 0236 (commit 6)

- Remove the `CreateTemplateTx` writes of `areas`/`visibility`/`specific_areas` (same commit).
- Migration `0236_dead_schema_drop.sql` (next free number; follow `metaldocs-database` skill: ledger insert, idempotent guards, grants untouched): `ALTER TABLE ... DROP COLUMN` ×4 (`templates` ×3, `document_profiles.is_active`), `DROP TABLE document_subjects`.
- Live-apply on Docker PG; post-drop API smoke (template create + document profile read through the API).
- Dictionary pages updated; `document_subjects` page retired per skill rules. Applied prior migrations stay byte-stable.

### 4.4 CI locks (commit 7)

Two new cilint analyzers (siblings of `platformboundary`/`postcommitaudit`, wired into invariants.yml, unit-tested, **empty baselines**):
- `nosqltxindomain` — flags `*sql.Tx` (or `database/sql` import) in `internal/modules/**/domain/`.
- `nodualmode` — flags `== nil` conditionals on db/port-typed fields guarding alternate execution paths in `internal/modules/**/application/` (heuristic: comparisons of `*sql.DB`/port-interface fields against nil followed by an else-branch; allow-directive mechanism per house style for any future justified case — none expected).

## 5. Verification & evidence

- Per commit: `go build ./...` · `go vet ./...` · targeted module tests · `go run ./tools/cilint/...` exit 0 · api-lint strict 0 (contract untouched — verify).
- Runtime proofs (Docker up): after §3.3 — Wave 2 Proof A re-run (taxonomy mutation → in-tx `audit_events` row on the rewritten path); after §4.2 — login + failed-reauth smoke (single-mode broke nothing on auth paths); after §4.3 — migration live-applied + post-drop API smoke.
- Tracker rows 2.12/2.13 updated same-commit per item; Wave 2 close-out (full tests, /code-review, wiki sync, evidence block, register notes) runs **after** both rows.

## 6. End state (the closure claim)

Zero `// Deprecated` markers in backend production code · zero nil-port/nil-db fallbacks · zero registered dead schema · domain layer free of `database/sql` · both seams CI-locked. Remaining DB-area items are Wave 3 **feature triggers** (RLS expansion at first external tenant), not refactors.
