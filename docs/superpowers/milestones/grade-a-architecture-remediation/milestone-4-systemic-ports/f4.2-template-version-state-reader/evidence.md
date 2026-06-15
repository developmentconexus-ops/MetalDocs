# Feature F4.2 — Evidence

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Feature:** `f4.2-template-version-state-reader`  ·  **Closed:** 2026-06-15
> **Contract:** [`spec.md`](spec.md) (consumer contract + Validation Gate this proves against).

## What was implemented

Producer built to the **existing consumer contracts** (CD's `application.TemplateVersionChecker`
and the documents-create profile-default resolver) — not the reverse:

- **Extended the templates-owned port** (`templates/domain.TemplateVersionPort`) with the raw-state
  primitive `GetTemplateVersionState(ctx, tenantID, versionID) (*string, string, error)`; `IsPublished`
  retained unchanged (taxonomy's Wave-Z contract). Impl
  (`templates/infrastructure.TemplateVersionReader.GetTemplateVersionState`) reuses the single existing
  `templateVersionQuery` with the same NullString / not-found `(nil,"",nil)` semantics CD's checker had.
- **Deleted** `controlleddocuments/infrastructure.PostgresTemplateVersionChecker` (struct + constructor +
  method) — the cross-module `templates_*` reach. `controlleddocuments/module.go` now wires
  `templatesinfra.NewTemplateVersionReader(deps.DB)` as `tplCheck`; it satisfies CD's
  `application.TemplateVersionChecker` directly (same name + signature). `service.go:209/308` unchanged.
- **Replaced the `status := "published"` hardcode** in `wiring/documents_adapters.go`:
  `profileDefaultsAdapter` now takes the templates reader injected (`NewProfileDefaults` panics if nil)
  and reads the real status via `GetTemplateVersionState(ctx, tenantID, *defaultTemplateVersionID)`.
  `main.go` wires `templatesinfra.NewTemplateVersionReader(deps.SQLDB)`. This corrects a latent bug: an
  obsolete/non-published profile-default template version previously resolved as publishable.
- Reads stay **live** (no snapshot/denormalization); call sites unmoved (H-PRE-1 not in play, see below).

Committed in this feature's commit (subject: `feat(milestone-4): F4.2 TemplateVersionStateReader port — close CD templates_* reach + real-status profile default`).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `TestProfileDefaults_ReadsRealStatusFromPort` (asserts injected port's "obsolete" flows through, NOT hardcoded "published") | RED before adapter wired to port → GREEN after | fixture |
| TDD — port raw-state live | `TestTemplateVersionReader_GetTemplateVersionState_Live` (4 subtests) | RED (no method) → **PASS** all 4: published_returns_raw_status_and_doctype, obsolete_returns_raw_status_not_bool, absent_returns_nil_empty_nil, other_tenant_returns_nil | **real (live PG)** |
| Static — build | `go build ./...` | `BUILD-OK` (no output) | — |
| Static — vet (plain) | `go vet ./internal/modules/templates/... ./internal/modules/controlleddocuments/... ./apps/api/internal/wiring/...` | `VET-OK` | — |
| Static — vet (`-tags integration`) | `go vet -tags integration ./internal/modules/templates/... ./internal/modules/controlleddocuments/...` | `VET-INT-OK` | — |
| Targeted suites | `go test ./apps/api/internal/wiring/ ./internal/modules/templates/... ./internal/modules/controlleddocuments/... ./internal/modules/taxonomy/...` | all `ok` (wiring 3.4s; templates+CD+taxonomy green incl. CD override `service_test` + delivery contract test) | real + fixture |
| Integration (live PG) | `go test -tags integration -run TestTemplateVersionReader_GetTemplateVersionState_Live ./internal/modules/templates/infrastructure/` | `ok ... 199.9s` — `--- PASS` (4/4 subtests) | **real (live PG)** |
| Class proof — CD reach closed | `grep -rn "templates_template" internal/modules/controlleddocuments/` | **0 SQL** (only a comment pointer at `repository.go:695`) | real |
| Class proof — hardcode gone | `grep -rn 'status := "published"' apps/api/internal/wiring/` | **0 matches** | real |
| H-PRE-1 intact | code inspection of `service.go:308` (auto-path read inside lock-holding tx) | read runs on the reader's own `r.db` **pool conn**, not `tx`; plain non-authz `SELECT v.status, t.doc_type_code` — **identical connection topology to the deleted checker** (both used the `deps.DB` pool); call site unmoved → not authz-recording, H-PRE-1 not in play and not regressed | real (structural) |

> The live-PG cleanup `drop isolated test database ... timeout` line is a best-effort teardown flake on
> the degraded C: SSD (per machine memory), **not** a test failure — the test reports `--- PASS` and the
> package reports `ok`.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Port returns raw status present / NULL→nil / not-found→nil, tenant-scoped, with `doc_type_code` | yes | live-PG integration 4/4 PASS |
| `IsPublished` still works (no taxonomy regression) | yes | taxonomy + templates suites `ok` |
| CD override validation behavior-identical via the port | yes | CD `application/service_test.go` override cases green (reader injected as `tplCheck`) |
| 0 `templates_template(_version)` SQL under `controlleddocuments/` | yes | grep → 0 SQL (comment only) |
| 0 `status := "published"` in `wiring/` | yes | grep → 0 |
| Adapter returns the **real** status (obsolete default no longer falsely resolves) | yes | `TestProfileDefaults_ReadsRealStatusFromPort` asserts "obsolete" from injected port |
| Status read stays off the lock-holding tx (H-PRE-1 intact) | yes | structural proof row above — pool conn, non-authz, call site unmoved |
| `go build ./...` + `go vet` (incl. `-tags integration`) clean | yes | BUILD-OK / VET-OK / VET-INT-OK |
| backend-api + workflow-async QA checklists | yes | dispositioned below |

### QA checklist disposition

- **backend-api-qa-checklist:** internal module-port swap — **no** route/OpenAPI/generated-surface/authz
  change, so the route-drift / contract-alignment / shared-consumer-break items are N/A by construction.
  Shared consumers (`CD.TemplateVersionChecker`, taxonomy `IsPublished`) explicitly preserved and
  green. The single behavioral delta — documents-create now feeds *real* status into
  `resolveDefaultTemplate` so an obsolete profile-default no longer falsely resolves as publishable —
  is the intended correctness fix and is unit-proven.
- **workflow-async-qa-checklist:** F4.2 adds **no** async/worker/outbox/scheduler behavior; CD-create
  remains a synchronous lock-bearing tx. Only the lock-interaction item applies and is covered by the
  H-PRE-1 structural proof (read on pool conn, non-authz, call site unmoved).

## Review disposition

- Spec-compliance review: **PASS** — producer matches both consumer contracts exactly (CD interface
  shape unchanged; documents-create now returns real status as the contract requires). Non-goals
  respected: `IsPublished` untouched, no OpenAPI/route change, no snapshot, no adjacent refactor.
- Code-quality review: **PASS** — single-owning-adapter (one type owns `templates_*` SQL + tenant
  scoping), CDC preserved, query reused (no duplication), nil-reader fail-fast panic in
  `NewProfileDefaults`. No root-cause family of findings.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Migrate taxonomy's `IsPublished` consumer to explicit-tenant for consistency | Pure consistency cleanup; current ctx-tenant form is correct and tested | Next structural touch of the templates port (recorded in spec Non-goals) |
| Two port ADRs (`UserDisplayNameReader` + `TemplateVersionStateReader`) | ADR authoring is its own feature | **F4.3** (next feature in this milestone) |
