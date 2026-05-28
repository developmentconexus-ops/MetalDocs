# PRP Plan: MetalDocs Go Backend Quality Gate Remediation

## Summary
Remediate the failed Go backend quality gate by fixing the smallest set of root-cause families that explain the blocking findings, instead of treating each finding as an isolated bug. The gate currently has `1 Critical` and `28 High` findings. These collapse into `6` fix waves with targeted re-review after each wave.

## Goal
Move the branch from `FAIL` to `PASS` or `PASS-WITH-WARNINGS` by:

1. Fixing all `Critical` and `High` findings.
2. Grouping changes by shared root cause and shared verification path.
3. Re-running only the affected review slices after each wave.
4. Avoiding another broad destabilizing sweep across the whole backend.

## Problem -> Solution
Large finding sweeps followed by broad fixes have been introducing regressions because the work is being handled as too many local edits instead of a few architectural failure classes.

Solution:
- Fix by root-cause family.
- Keep each wave bounded to a small module set.
- Require targeted validation before moving to the next wave.
- Re-review the exact affected slice before declaring the wave complete.

## Metadata
- **Source**: Go Backend Quality Gate review on branch `codex/metaldocs-go-backend-quality-bar`
- **Status**: Ready for execution
- **Complexity**: Large but bounded
- **Blocking findings**: `1 Critical`, `28 High`
- **Execution model**: sequential waves, targeted validation, targeted re-review

---

## Root-Cause Groups

### Group A: Tier-2 authz transaction context is not seeded before `authz.Require`
This is the most repeated backend correctness bug in the current gate.

Affected areas:
- `internal/modules/documents/repository`
- `internal/modules/iam/infrastructure/postgres`
- related service-owned transaction paths

Why it matters:
- production writes fail closed at runtime
- tests can miss it when mocks/stubs bypass real transaction state

### Group B: Preconditions are parsed but not enforced
This is concentrated in approval flows.

Affected areas:
- `internal/modules/documents/approval/http`
- `internal/modules/documents/approval/application`

Why it matters:
- stale clients can overwrite newer state
- integrity checks exist in shape only, not in effect

### Group C: Authorization/read-path enforcement is inconsistent
This includes the only Critical.

Affected areas:
- `internal/modules/documents/approval/application/read_service.go`
- selected read/update helpers in taxonomy/search/documents-adjacent flows

Why it matters:
- authenticated users can observe data they should not see
- module contracts diverge between similar entry points

### Group D: Observability and failure surfaces report success-shaped output on failure
Affected areas:
- `internal/platform/observability`
- `internal/modules/audit`

Why it matters:
- broken dependencies look healthy
- malformed audit payloads become silently normalized
- operators get false confidence during incidents

### Group E: Outbound and async infrastructure lacks defensive bounds
Affected areas:
- `internal/platform/config/gotenberg.go`
- `internal/platform/render/gotenberg`
- `internal/platform/servicebus`
- `internal/platform/worker`

Why it matters:
- SSRF/misrouting risk
- unbounded memory reads
- retry and batch processing behavior degrades under failure

### Group F: Search/template/migration correctness drift
Affected areas:
- `internal/modules/search`
- `internal/modules/templates`
- `db/migrations/0211_editor_sessions_tenant_id.sql`
- `apps/api/cmd/metaldocs-api`

Why it matters:
- search can return false negatives and incomplete data
- template lifecycle updates are not concurrency-safe
- migration leaves sentinel-tenant risk behind

---

## Fix Waves

### Wave 1: Approval read/authz and precondition enforcement
**Priority**: P0  
**Reason**: contains the only `Critical`

**Scope**
- `internal/modules/documents/approval/application/read_service.go`
- `internal/modules/documents/approval/http/doc_approval_handler.go`
- `internal/modules/documents/approval/http/route_admin_handler.go`
- `internal/modules/documents/approval/application/decision_service.go`

**Findings closed**
- `1 Critical`
- `3 High`

**Root causes**
- Group B
- Group C

**Implementation targets**
- enforce `document.view` path symmetry between `LoadInstance` and `LoadActiveInstanceByDocument`
- require and propagate `If-Match` consistently across mutating approval entry points
- enforce server-vs-client `content_hash` comparison before signoff persistence
- add optimistic concurrency to route update/deactivate operations

**Validation**
- approval application tests for unauthorized read path
- handler tests for missing/stale `If-Match`
- signoff tests for stale or mismatched `content_hash`
- targeted reviewer rerun: approval slice only

**Exit condition**
- no remaining Criticals in approval
- no stale-write approval Highs remain

---

### Wave 2: Shared authz transaction seeding
**Priority**: P0  
**Reason**: repeated root cause across documents and IAM/auth-adjacent writes

**Scope**
- `internal/modules/documents/repository`
- `internal/modules/documents/application` transaction owners
- `internal/modules/iam/infrastructure/postgres/user_area_repository.go`

**Findings closed**
- `6 High` directly
- likely collapses follow-on behavior bugs

**Root causes**
- Group A

**Implementation targets**
- introduce one canonical helper for seeding `metaldocs.actor_id` and `metaldocs.tenant_id` on transaction-local scope
- apply it before every tier-2 `authz.Require`
- standardize service/repository ownership so transaction creation and GUC seeding are not split inconsistently

**Validation**
- Postgres-backed integration tests, not memory-only tests
- positive-path mutation tests for create/rename/status/archive/membership writes
- negative-path tests for missing actor/tenant context
- targeted reviewer rerun: documents core + IAM/Auth affected paths

**Exit condition**
- all known missing-GUC Highs closed
- no write path depends on caller remembering ad hoc authz seeding

---

### Wave 3: IAM/Auth contract and tenant-scope repair
**Priority**: P1

**Scope**
- `internal/modules/auth/infrastructure/postgres/repository.go`
- `internal/modules/auth/infrastructure/memory/repository.go`
- `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`

**Findings closed**
- `3 High`
- `2 Medium` likely in the same pass

**Root causes**
- Group A
- tenant scoping and contract drift between memory and Postgres

**Implementation targets**
- enforce `RowsAffected` checks on user update paths
- add `tenant_id` scoping to online-user listing
- make role replacement contract match migration `0166`
- align memory adapter sentinel errors with Postgres behavior

**Validation**
- contract tests shared across memory and Postgres adapters
- tenant-isolation tests for online-user listing
- targeted reviewer rerun: IAM + Auth slice

**Exit condition**
- Postgres and memory behavior match for sentinel errors and tenant scope on reviewed paths

---

### Wave 4: Platform observability and outbound safety
**Priority**: P1

**Scope**
- `internal/platform/observability`
- `internal/platform/config/gotenberg.go`
- `internal/platform/render/gotenberg`
- `internal/platform/servicebus/docgen_v2_validate.go`
- `internal/platform/worker/service.go`
- `apps/worker/cmd/metaldocs-worker/main.go`

**Findings closed**
- `8 High`
- `1 Medium`

**Root causes**
- Group D
- Group E

**Implementation targets**
- readiness must degrade on dependency failure
- metrics must not publish fake zero-success states
- validate outbound URLs at config load
- cap response-body reads and propagate read errors
- make worker batch processing continue safely after per-item persistence failure
- clamp retry backoff arithmetic
- make worker shutdown context-aware

**Validation**
- focused unit tests for readiness/metrics failure behavior
- response-size bound tests for Gotenberg/servicebus
- worker tests for batch-continue and backoff overflow
- targeted reviewer rerun: platform layer slice

**Exit condition**
- no success-shaped failure behavior remains in reviewed observability paths
- no unbounded outbound body reads remain in reviewed paths

---

### Wave 5: Search and templates correctness
**Priority**: P1

**Scope**
- `internal/modules/search`
- `internal/modules/templates`
- related taxonomy helper if touched for pre-write authz sequencing

**Findings closed**
- `4 High`
- `2 Medium`

**Root causes**
- Group C
- Group F

**Implementation targets**
- make search reader hydrate all fields used by filters/response
- stop limiting before authorization/secondary filtering
- add proper optimistic concurrency to template version state changes
- align template route capability checks with service-layer contract

**Validation**
- search filter correctness tests with authorized results beyond first SQL page
- template lifecycle concurrency tests
- targeted reviewer rerun: templates + taxonomy + search slice

**Exit condition**
- search no longer produces false negatives due to partial hydration or early limit
- template lifecycle transitions are stale-write safe

---

### Wave 6: Audit surfaces, entrypoint hardening, and migration 0211 cleanup
**Priority**: P2

**Scope**
- `internal/modules/audit`
- `apps/api/cmd/metaldocs-api/main.go`
- `db/migrations/0211_editor_sessions_tenant_id.sql`
- any migration verification helper/scripts needed for proof

**Findings closed**
- `3 High`
- remaining Mediums tied to the same surfaces

**Root causes**
- Group D
- Group F

**Implementation targets**
- fail loudly on nil audit reader wiring
- stop silently rewriting malformed audit payloads
- bound integrity validation work
- fail startup or provide explicit safe behavior if templates authz dependency is absent
- convert `0211` from sentinel-default migration to verified backfill-and-tighten migration

**Validation**
- audit tests for malformed payload behavior and nil dependency handling
- migration verification against representative pre-0211 data
- direct invariant check: `tenant_id` present, backfilled, indexed, no sentinel fallback left behind
- targeted reviewer rerun: audit + apps/migrations slice

**Exit condition**
- migration `0211` is trustworthy as runtime truth
- audit and entrypoint behavior fail explicitly, not silently

---

## Execution Rules

### Rule 1: One wave at a time
Do not mix Wave 1 and Wave 4 in the same patch set. Each wave must remain reviewable and reversible.

### Rule 2: Fix by helper or contract first
If three findings share one missing helper or invariant, introduce the helper once and migrate all touched callsites in the same wave.

### Rule 3: Prefer integration proof where the bug is transactional
Any bug involving:
- `authz.Require`
- transaction-local GUCs
- `RowsAffected`
- migration data repair

must be validated against real Postgres behavior.

### Rule 4: Re-review the affected slice before starting the next wave
Do not trust local reasoning alone after structural fixes.

### Rule 5: Do not reopen the whole backend gate until all six waves are complete
Use targeted re-review between waves, then one final full quality-gate rerun at the end.

---

## Step-by-Step Tasks

### Task 1: Build a finding-to-wave mapping
- assign every Critical/High finding to exactly one wave
- note shared helper opportunities before code changes begin
- freeze wave boundaries to avoid scope creep

### Task 2: Execute Wave 1
- implement approval authz/precondition fixes
- add targeted tests
- rerun approval review slice
- record closed findings

### Task 3: Execute Wave 2
- introduce shared authz transaction seeding helper
- migrate documents and IAM membership write paths
- add Postgres integration coverage
- rerun documents core + IAM/Auth affected review slices

### Task 4: Execute Wave 3
- repair auth contract drift and tenant leakage
- align memory vs Postgres behavior where relevant
- rerun IAM/Auth slice

### Task 5: Execute Wave 4
- fix platform observability and outbound safety
- add bounds/failure-behavior tests
- rerun platform slice

### Task 6: Execute Wave 5
- fix search hydration/limit semantics
- add template optimistic concurrency
- rerun templates/taxonomy/search slice

### Task 7: Execute Wave 6
- repair audit fail-open behavior
- harden API entrypoint dependency behavior
- repair and verify migration `0211`
- rerun audit + apps/migrations slice

### Task 8: Final backend gate rerun
- rerun the full professional-patterns quality gate
- compare against current blocking list
- declare `PASS`, `PASS-WITH-WARNINGS`, or next bounded follow-up

---

## Validation Strategy

### Per-wave validation
Every wave must include:
- targeted unit tests for local logic
- integration tests for DB/authz/migration behavior where applicable
- targeted reviewer rerun for the affected ownership slice

### Final validation
- rerun full Go backend quality gate
- recheck required invariants:
  - `internal/platform/config/postgres.go` sslmode default remains `require`
  - `internal/modules/iam/infrastructure/postgres/role_provider.go` still returns `ErrNoRolesAssigned`
  - `internal/modules/iam/delivery/http/admin_handler.go` still reads `X-Trace-Id` from request header
  - `internal/modules/documents/approval/application/decision_service.go` still calls `authz.WithCapCache(ctx)` before `db.BeginTx`
  - `internal/modules/documents/approval/application/scheduler_service.go` still sets `OccurredAt` from `s.clock.Now()`
  - `internal/modules/documents/delivery/http/handler.go` still logs `parseListOptions` errors server-side and returns `bad_request`
  - `db/migrations/0211_editor_sessions_tenant_id.sql` remains present and trustworthy, with verified backfill and index

---

## Acceptance Criteria
- all `Critical` findings are closed
- all `High` findings are closed
- each wave has a validation record and targeted re-review result
- no wave introduces a new Critical/High in an already-reviewed earlier wave
- final full backend quality gate returns `PASS` or `PASS-WITH-WARNINGS`

## Not Doing
- broad “fix everything in backend” patch
- full-medium sweep before Critical/High closure
- frontend/browser QA before backend gate clears
- speculative refactors unrelated to a root-cause family

## Recommended Start
Start with `Wave 1`.

Reason:
- it removes the only Critical
- it is narrow
- it tests the execution model before the larger shared-helper waves
