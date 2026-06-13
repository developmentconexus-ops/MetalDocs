# Wave H — Detailed Execution Plan (drift-proof)

> **Status:** ▶ in progress (2026-06-13). **Branch:** `qa/iam-area-membership`. **Do NOT merge** — operator review gate.
> **Source of truth:** [`_artifacts/architecture-audit-2026-06-13.md`](_artifacts/architecture-audit-2026-06-13.md) (the 23-defect audit) + the **7-plane global-optimum assessment** (synthesis in [`roadmap.md`](roadmap.md) §"Global-optimum assessment"). This file is the durable per-task spec so a fresh session executes without re-deriving — same precedent as [`wave-z-plan.md`](wave-z-plan.md).
> **Read order for a fresh session:** `CLAUDE.md` → `wiki/README.md` → `wiki/references/current-agent-handoff.md` → the audit → `roadmap.md` §Wave H → this file → the row you are executing.

## Anti-drift rules (read every session)

1. **One commit per sub-family.** Tracker row in [`roadmap.md`](roadmap.md) §Wave H **and** the audit disposition table updated **in the same commit**.
2. **Stage explicit paths** — `git add <path>`, never `git add -A` (the untracked `.gitnexus/` cache breaks it with an mmap error). The `D .agents/skills/*` deletions in `git status` are pre-existing, NOT ours — never stage them.
3. **Tests run `-p 2`** (C: SSD has degraded writes; `go test -p 2 ./...`).
4. **Per-commit verify gate (minimum):** `go build ./...` · `go vet ./...` · targeted `go test -p 2 ./internal/modules/<touched>/...`. **Family-close gate:** add `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` (if contract touched) + `go run ./tools/cilint/... ./...`.
5. **Models:** Opus orchestrates; **sonnet** implements/reviews; **haiku** mechanical only; **never fable** workers; ≤15 concurrent agents.
6. **Evidence rule:** no `done`/`green` without the command + output recorded in the tracker row.
7. **Hard-stop rule:** if a fix balloons into one of the 5 deferred boundaries below (shared-API / authz-internal / storage-provider / workflow-semantic redesign), STOP, record the trigger, do NOT symptom-patch.
8. **Semantics-preserving:** every Wave H change is a structural/quality refactor. Behavior identical; existing tests are the regression net. Where a test asserts the OLD structure (not behavior), update it and say so.

## Execution order (conflict-minimized) + why

`H-6a → H-3a → H-4 → H-1(a→b→c→d→e) → [H-3b, H-6b] → H-5 → H-2`  (H-3 split + H-6b resequenced after H-1d — see H-3/H-6 discovery notes)

- **H-6 first** — deletes dead code in `documents/application` → shrinks the surface H-1e/H-5 later touch.
- **H-3a, H-4** — medium, mostly delivery-layer; get them in before the big H-1 restructure churns the same files. (H-3b — the 4 sites that need *new* service-owned tx boundaries — resequenced after H-1d's TxRunner seam.)
- **H-1** — the big structural redesign (decision_service, CD repo, approval, documents delivery). Sub-ordered a→e so the mechanical/low-risk parts (setAuthzGUC, CD port, approval import) land before the TxRunner refactor (d) and the documents-delivery redesign (e).
- **H-5 after H-1** — RecordSignoff SQL-extraction works on the TxRunner-shaped service (d); the documents `Service` interface split works on the migrated handler (e).
- **H-2 LAST** — `main.go` extraction captures all prior wiring changes (CD→taxonomy port wiring, TxRunner wiring) in one final pass.

---

## Tier 1 — MUST-FIX (A1–A6): ✅ committed `1c31ffd70`(A1)…`ed1cf8376`(A6)

Only residual: **A1 runtime WS-upgrade proof** (`statusWriter.Unwrap()` confirmed at `internal/platform/observability/http.go:328`) — execute at Wave H runtime-QA close: `.\scripts\start-api.ps1 -Build`, open `/api/v1/iam/presence/stream`, confirm 101 Switching Protocols (not 501).

---

## Tier 2 — Architecture debt (the work)

### H-6 — Dead-code  ·  audit: Legacy/dead-code (B)  ·  **split: H-6a (now) + H-6b (resequenced after H-1d)**

**Discovery (2026-06-13, during execution):** the legacy `Service.CreateDocument` doc-comment ("renders via docx-renderer, uploads to S3") is **stale** — the code does template-passthrough (`service.go:331` `finalKey := docxKey`, no render). And `repo.CreateDocument` (`repository.go:100-118`) is already **DB-tx-atomic** (BeginTx→CreateDocumentTx→Commit). The genuine non-atomicity is **cross-operation**: `DuplicateDocument` (`service.go:485`) commits the CD duplication (`controlledDocumentDuplicator.DuplicateControlledDocument`) in one tx, then creates the document in a *separate* tx via legacy `CreateDocument` → a duplicated CD can orphan with no document. The atomic create path is `cd_initializer.CloneTemplate(ctx, tx, …)` → `svc.cloneIntoTx` (`service.go:368`), which **requires a caller-owned tx** the documents `Service` does not currently hold. ⇒ **H-6b is coupled to the TxRunner seam (H-1d), not pure dead-code** — resequenced to run after H-1d.

#### H-6a — delete dead `SnapshotFromTemplate` + cascade  ·  ✅ DONE (this commit)
**Done:** deleted `SnapshotFromTemplate`, `SnapshotWriter`/`PlaceholderValueSeeder` interfaces, `NewSnapshotServiceWithSeeder`, the `writer`/`seeder` fields + `w` param, and — completing the cascade to the repo layer exactly as specced below — the orphaned-by-deletion `SnapshotRepository.WriteSnapshot`/`ReadSnapshot` + `FillInRepository.SeedDefaults` (all grep-proven zero prod callers; only `ReadSnapshotWithFreezeAt` stays live). Kept `ResolveTemplate`/`parseRequiredPlaceholders` + all live snapshot/fillin methods. Removed dead-method test cases only (no live-method coverage lost — `ReadSnapshotWithFreezeAt` had no integration test either way). Gates: `go build ./...`=0 · `go vet ./...`=0 · `go vet -tags integration ./internal/modules/documents/...`=0 · `go test -p 2 ./internal/modules/documents/...`=ok · cilint=0.
`SnapshotFromTemplate` (`snapshot_service.go:48`) has only 2 test callers (no prod; GitNexus + grep agree). It is the **only** user of the `SnapshotService.writer`/`seeder` apparatus — `ResolveTemplate` (the live method) uses only the reader. So cascade-delete IF grep proves unused in prod: `SnapshotWriter` + `WriteSnapshot` impls, `PlaceholderValueSeeder` + `SeedDefaults` impls, `NewSnapshotServiceWithSeeder`, the `writer`/`seeder` fields (+ the `w` param of `NewSnapshotService` if no live caller passes a real writer). **KEEP** `ResolveTemplate` + `parseRequiredPlaceholders` (still used). Delete `snapshot_seeder_test.go` + the `SnapshotFromTemplate` integration test. Verify each symbol's prod callers via grep before deleting.
**Verify:** `go build ./...` · `go vet ./...` (+ `-tags integration` vet on the documents application pkg) · `go test -p 2 ./internal/modules/documents/...` + grep proof recorded.
**Commit:** `refactor(documents): delete dead SnapshotFromTemplate + collapsed snapshot writer/seeder apparatus (H-6a)`

#### H-6b — `DuplicateDocument` cross-operation atomicity (AFTER H-1d)
Once the TxRunner seam exists (H-1d), compose CD-duplicate + document-create in **one tx** via the `cd_initializer.CloneTemplate` pattern (mirror the atomic-create flow). Then delete the legacy `Service.CreateDocument` + `repo.CreateDocument` + `SetRevisionStorageKey` + their interface entries (`service.go:29`, `delivery/http/handler.go:32`) + the test fakes. Grep-prove zero remaining callers first.
**Verify:** build/vet + `go test -p 2 ./internal/modules/documents/... ./internal/modules/controlleddocuments/...` + runtime duplicate-document smoke.
**Commit:** `refactor(documents): atomic DuplicateDocument, delete legacy CreateDocument chain (H-6b)`

---

### H-3 — Persistence  ·  audit: Persistence (B)  ·  assessor: persistence-tx B− PATCH  ·  **split: H-3a (now) + H-3b (resequenced after H-1d)**

**Discovery (2026-06-13, during execution):** the plan's premise — "*move* 5 post-commit `audit.Record` calls in-tx" — is **false for 4 of the 5 sites**. Verified per-site:
- **Site 2** (`routes_memberships.go:361`, membership grant/revoke) — in-tx audit **already exists**: `AreaMembershipService.Grant/Revoke` call `logger.LogTx → RecordTx` inside the mutation tx before `Commit` (Z-6, `area_membership_service.go:152/178/237`), and the prod `membershipGovernanceLogger.buildEvent` (`main.go:1030`) maps `role.grant`→`iam.area_membership.granted` — the **same `Action`/`ResourceType`/`ResourceID`** the handler's post-commit `recordMembershipAudit` writes. ⇒ the handler call is a **duplicate audit row** (torn-write window + double-write), not a missing in-tx write. The only datum the post-commit row carries that the in-tx one lacks is `TraceID` (from the `X-Trace-Id` header) — and the `Recovery` middleware (`internal/platform/middleware/recovery.go:20-26`, outermost, REQ-MW-1) already seeds that trace into `ctx`, so the in-tx path can recover it via `requesttrace.Resolve(ctx)` (the auth-handler pattern, `auth/.../handler.go:195`). **Zero data loss on deletion.**
- **Sites 1/3/4/5** (`sessions_handler.go:213` RevokeSession · `people_handler.go:480` Patch/Reset/Unlock/Invite · `admin_handler.go:402` UpsertUserAndAssignRole/ReplaceUserRoles · `auth/.../handler.go:200` ChangePasswordForUser) — **have no service-owned tx**: each is one-or-more autocommit ops, or the only tx is repo-internal (`role_admin_repository.go` BeginTx/Commit, no handle exposed up). Making their audit in-tx requires **introducing a new service-owned transaction** wrapping the mutation + `RecordTx` — exactly the seam H-1d's `TxRunner` builds. Hand-rolling `BeginTx/Commit` in 4 services now = boilerplate H-1d deletes. ⇒ **resequenced after H-1d** (same rationale as the H-6b resequencing).

Also: the nil-tx allocator trap's only `nil`-passing caller (`controlleddocuments/application/service.go:258`) is an **unreachable dead `else` branch** (`createTx` is provably non-nil inside `if cmd.ManualCode == nil` — it's the very branch that opened `createTx`). So the allocator reject + dead-branch delete is clean now (the one real integration caller, `domain/sequence_test.go:80`, wraps each increment in its own `*sql.Tx`).

#### H-3a — site-2 duplicate-audit delete + allocator nil-tx reject  ·  **now**
1. **Extract `membershipGovernanceLogger` out of `main.go`** (`main.go:1015-1060`) into `internal/modules/iam/application/membership_governance_logger.go`, mirroring taxonomy's `AuditGovernanceAdapter` (`taxonomy/application/audit_governance_adapter.go`) — the iam equivalent sits untested in the composition root; this makes the `role.grant`→`iam.area_membership.granted` mapping unit-testable and removes one inline adapter from `main.go` (chips at H-2). Thread `TraceID` via `requesttrace.Resolve(ctx)`. Add a unit test for the mapping + trace.
2. **Delete the handler's duplicate** `recordMembershipAudit` calls (`routes_memberships.go:233` grant, `:287` revoke) + the now-unused helper (`:346`). Single in-tx audit row remains.
3. **Rewire the iam_memberships handler test** (`tests/unit/iam_memberships/area_memberships_handler_test.go`) to wire the **real** extracted adapter → the recording audit (the existing `EmitsAudit`/duplicate/areaadmin assertions now exercise the in-tx path); drop the unused `noopMembershipLogger`.
4. **Reject nil-tx in `PostgresSequenceAllocator.NextAndIncrement`** (`infrastructure/repository.go:656-687`): remove the `var exec db.Tx = a.db` fallback; nil tx → explicit error. **Delete the dead `else` branch** at `application/service.go:257-271` (+ redundant `if createTx != nil` guard). Update `domain/sequence_test.go` to open a per-goroutine `*sql.Tx`.

**What NOT to do (H-3a):** do not change audit semantics or add new audit events (the in-tx event is the survivor; it only *gains back* the `TraceID` the deleted duplicate carried). Do not touch the authz path. The `postcommitaudit` analyzer extension is **deferred to H-3b** — extending it now would false-flag the legitimately-deferred sites 1/3/4/5 (forcing `//cilint:allow` litter); after H-3b converts them it enforces cleanly.
**Verify:** `go build` · `go vet` · `go vet -tags integration ./internal/modules/controlleddocuments/...` · `go test -p 2 ./internal/modules/iam/... ./internal/modules/controlleddocuments/... ./tests/unit/iam_memberships/... ./apps/api/...` · `go run ./tools/cilint/... ./...` exit 0.
**Commit:** `refactor(persistence): dedupe membership audit to single in-tx write + extract governance logger + reject nil-tx in sequence allocator (H-3a)`

#### H-3b — service-owned tx + in-tx audit for sites 1/3/4/5 + analyzer (AFTER H-1d)
Once the `TxRunner` seam exists (H-1d), wrap each of `RevokeSession`, `people_handler` mutations (Patch/Reset/Unlock/Invite), `admin_handler` role ops, and `ChangePasswordForUser` in a service-owned tx and move their `audit.Record` → `RecordTx` inside it. Then **extend the `postcommitaudit` cilint analyzer** to catch the cross-function case (business `Commit()` in service + non-Tx `audit.Record` in delivery) + unit tests — now enforceable without allow-directives.
**Verify:** build/vet + `go test -p 2 ./internal/modules/iam/... ./internal/modules/auth/... ./tools/cilint/...` + `go run ./tools/cilint/... ./...` exit 0 + runtime audit-row smoke (one row per mutation, in-tx).
**Commit:** `refactor(persistence): service-owned tx + in-tx audit for 4 IAM/auth handlers + cross-function analyzer (H-3b)`

---

### H-4 — Contract  ·  audit: Contract (C)  ·  assessor: contract-api C PATCH  ·  ✅ DONE (commit pending push)

**Discovery during execution (2026-06-13):** the FE `error-codes.generated.json` already listed `MEMBERSHIP_EXISTS`, `MEMBERSHIP_NOT_FOUND`, `UNKNOWN_ROLE` (the dump tool scans literals, not catalog membership) — so the iam-package guard extension would catch 3 more off-catalog codes than the plan's "2". Resolution: promote all **5** to typed consts with their exact existing string values (not remap) → generated JSON stays byte-identical, zero FE/wire change. 164 literals converted; guard 5→10 pkgs. Status/message args left untouched: several "service not configured" sites pair 501 with `INTERNAL_ERROR` (a contract smell) — re-statusing is a wire change, **deferred** (not in codes-only scope). search/handler.go gained the `problem` import (it used `httpresponse.WriteError`).


**Goal:** close the typed `problem.Code` catalog over all delivery packages; kill raw error-code string literals.

**Steps:**
1. **Add the 2 off-catalog codes** to `internal/platform/problem/codes.go`: `CURSOR_EXPIRED` (domain pagination signal) and `NOT_IMPLEMENTED` (→ `CodeNotImplemented`). Off-catalog sites: `iam/.../people_handler.go:93` (CURSOR_EXPIRED), `people_handler.go:321` + `audit/.../handler.go:142,195,207,229` (NOT_IMPLEMENTED).
2. **Replace raw error-code string literals with typed `problem.Code` consts** across the ~163 sites in **auth, iam, audit, security, search** delivery handlers (iam ~115, audit ~25, auth ~17, security ~4).
3. **Extend the catalog CI guard** `internal/platform/problem/codes_catalog_guard_test.go:33-43` `guardedPackages` to cover `iam, audit, auth, search, security` (was 5 packages, now ~11).
4. **FE regen (criteria #4)** — route through **`metaldocs-tanstack-query`** skill: regenerate the FE error-code catalog (`dump-error-codes.go`), add PT-BR messages for the 2 new codes, run FE coverage test green.

**What NOT to do:** do not migrate IAM/auth/search raw-mux routing to ServerInterface here — that is **deferred boundary #3** (needs spec-prefix normalization). H-4 is codes-only.

**Verify:** `go build` · `go vet` · `go test -p 2 ./internal/platform/problem/... ./internal/modules/{auth,iam,audit,security,search}/...` · FE: `cd frontend/apps/web; pnpm gen:api; npx tsc --noEmit` + FE coverage test.
**Commit:** `refactor(contract): typed problem.Code across auth/iam/audit/security/search + catalog guard + 2 codes + FE regen (H-4)`

---

### H-1 — Module boundaries  ·  audit: Module boundaries (C)  ·  assessor: macro-topology B− REDESIGN (documents cluster)

5 sub-commits. **H-1a → H-1b → H-1c → H-1d → H-1e.**

#### H-1a — `setAuthzGUC` ×4 dedup → `authz.SeedTxIdentity`  ·  ✅ DONE (this commit)
Delete the 4 copies, replace all call sites (19 in approval) with the canonical `authz.SeedTxIdentity` (`internal/modules/iam/authz/context.go:48` — has the empty-string guard the copies skip; batches both GUCs in one query):
- `internal/modules/documents/approval/application/authz_guc.go:11`
- `internal/modules/templates/application/authz_guc.go:9`
- `internal/modules/taxonomy/infrastructure/authz_guc.go:14`
- `internal/modules/controlleddocuments/application/service.go` (inline `setAuthzGUC` ~:377)
**Verify:** build/vet + `go test -p 2 ./internal/modules/{documents/approval,templates,taxonomy,controlleddocuments}/...`. **Commit:** `refactor(authz): dedup setAuthzGUC x4 to authz.SeedTxIdentity (H-1a)`

**Discovery during execution (2026-06-13):**
- **approval (17 sites) / templates (11) / CD (4)** — pure textual swap to `authz.SeedTxIdentity(`; all files already import `authz`. Behavioral delta: canonical adds empty-string guards (`ErrTenantContextMissing`/`ErrActorContextMissing`) + `TrimSpace` + single batched query — strictly better; these are authenticated mutation flows so non-empty always holds.
- **taxonomy (24 sites) — deviation from the plan's literal "replace all call sites":** taxonomy's `setAuthzGUC(ctx, tx)` is a *ctx-resolving wrapper* (resolves tenant via `tenant.FromContext` + actor via `iamdomain.UserIDFromContext`), not a bare GUC writer. Inlining that resolution at 24 sites is worse DRY than keeping the one wrapper. **Resolution:** kept the thin wrapper, delegated only the duplicated GUC *SQL* to `authz.SeedTxIdentity`; the 24 call sites are untouched. The real duplication target (the two-query GUC write) is now centralized.
- **Test fixtures:** sqlmock helpers + custom `database/sql/driver` fakes asserted the old *two-separate-query* GUC form (`set_config(tenant_id)` then `set_config(actor_id)`, each 1-arg) → migrated to the canonical *single batched 2-arg* query (`SELECT set_config($1,…), set_config($2,…)`). Custom fakes that captured the actor from `args[0]` re-pointed to `args[1]` (actor is now the 2nd bind in the batched call).
- **Surfaced (out of H-1a scope, fixed separately):** the full-suite gate run flagged `TestPasswordChangePreservesSessionAndClearsMustChangePassword` (tests/unit) failing — a **pre-existing A3 stale test** (A3 revokes the current session; the older flow test still asserted survival). Confirmed pre-existing on clean HEAD (stash test). Reconciled + renamed in its own bounded commit `50e60e333` (A3 family, not H-1a).

**Gates (all green):** `go build ./...` 0 · `go vet ./...` 0 · `go test -p 2 ./...` 0 failures · `api-lint -strict` 0 violations · `cilint` exit 0. Contract-neutral (internal authz wiring only — no OpenAPI/route change).

#### H-1b — CD → taxonomy read port (closes authz-skip + stale-column drift)  ·  ✅ DONE (this commit)
Delete `controlleddocuments/infrastructure` `TaxonomyProfileReader` / `TaxonomyAreaReader` (`repository.go:720-799`) — they re-implement taxonomy `GetByCode` **without** the authz GUC + `CapTaxonomyView` check and **omit the `alias` column** the canonical reader has (`taxonomy/infrastructure/repository.go:103-126`). Define a **taxonomy read port** (interface in `taxonomy/domain` or a shared read-port package) that CD consumes via `controlleddocuments/module.go:34-35` wiring; CD calls taxonomy through a background-bypass context so the authz GUC + cap check + alias column are honored. **Regression-test the CD creation flow** (the missing-alias divergence is a real schema risk).
**Verify:** build/vet + `go test -p 2 ./internal/modules/controlleddocuments/... ./internal/modules/taxonomy/...`. **Commit:** `refactor(controlleddocuments): consume taxonomy read port, delete duplicate readers (H-1b)`

**Discovery during execution (2026-06-13):**
- **Deleted** the two duplicate readers (`repository.go` ~726-805) + the now-orphaned `setNullableStringPtrField` reflection helper (only caller was the deleted area reader) + the now-unused `reflect` and `taxonomydomain` imports from `repository.go`.
- **Deviation from the plan's "background-bypass context":** background-bypass is **fail-closed on request paths** (`ErrBypassNotBackground`, CWE-269) — it is *forbidden* on the CD create request path, so it cannot be the bridge. Empirically, **every role that can create a controlled document already holds `taxonomy.view`** (db/reference-data: `controlled_documents.create` → {author, editor, system_admin}; `taxonomy.view` → {system_admin, area_admin, approver, author, editor, viewer}). So routing the read through the canonical taxonomy reader with the **request ctx** enforces an already-held capability — zero behavior break, and the authz GUC + `CapTaxonomyView` + `alias` column are now honored. This is the correct global-optimum fix; the plan's bypass idea would have *weakened* the check.
- **Port shape:** rather than a new interface in `taxonomy/domain`, CD's existing consumer ports (`application.ProfileReader`/`AreaReader`, `GetByCode(ctx, tenantID, code string)`) are kept; two thin adapters in `controlleddocuments/infrastructure/taxonomy_reader.go` (`TaxonomyProfileReader`/`TaxonomyAreaReader`, same constructor names) now wrap **narrow read slices** of the canonical taxonomy repos (interface-segregated `taxonomyProfileGetter`/`taxonomyAreaGetter`) and bridge `code string` → `taxonomydomain.ProfileCode`/`AreaCode`. `module.go:34-35` constructs `taxonomyinfra.NewProfileRepository(deps.DB)`/`NewAreaRepository(deps.DB)` and injects them. CD's existing unit fakes (`fakeProfileReader`/`fakeAreaReader`) are unchanged — port signature is identical.
- **Regression test:** `taxonomy_reader_test.go` — adapter delegation, `string`→typed-code conversion, and error propagation for both readers.

**Gates (all green):** `go build ./...` 0 · `go vet ./...` 0 · `go test -p 2 ./...` 0 failures · `api-lint -strict` 0 violations · `cilint` exit 0. Contract-neutral (internal read wiring only — no OpenAPI/route change).

#### H-1c — approval delivery ↛ infrastructure  ·  ✅ DONE (this commit)
`approval/http/handler.go:13` imports `approvalinfra` for `SignoffReplayCommitter` + `SignoffReplay` — delivery importing infrastructure (hexagonal inversion). **Move those interfaces up** to approval application/domain; delivery imports application only; the postgres impl stays in infrastructure and satisfies the application-layer interface.
**Verify:** build/vet + `go test -p 2 ./internal/modules/documents/approval/...` + grep `approval/infrastructure` from `approval/http/` → 0. **Commit:** `refactor(approval): move SignoffReplay interfaces to application, delivery stops importing infrastructure (H-1c)`

**Execution (2026-06-13):**
- New `approval/application/signoff_idemp.go` holds `SignoffReplay` (struct) + `SignoffReplayCommitter` (interface) — mirrors the **existing** `route_admin_idemp.go` precedent (`RouteAdminReplay`/`RouteAdminReplayCommitter` already live in application; `PostgresRouteAdminIdempStore` returns them). Signoff was the lone holdout with the types in infra.
- `infrastructure/postgres_signoff_idemp_store.go` deletes its local type defs and now references `application.SignoffReplay`/`SignoffReplayCommitter` (infra→application is the established direction — root infra already imports application in `postgres_route_admin_idemp_store.go`, no cycle). `SignoffReplayHandle`/`PostgresSignoffIdempStore` concretes unchanged, still satisfy the application interface.
- Delivery (`handler.go`, `doc_approval_handler.go`, `signoff_handler_test.go`) now references `application.*`; the `approvalinfra` (root infrastructure) import is removed from all three. The consumer-defined `signoffIdempStore` port stays in delivery (legitimate — delivery may define the narrow port it needs; the inversion was only the infra *type* reference). CD's existing fakes unchanged.
- **Verify:** root-infra import from `approval/http/` (prod) → **0**; build/vet/`test -p 2 ./...`/api-lint -strict 0/cilint 0. Contract-neutral (no route/spec/codegen change).

**Bounded defer (recorded, not in H-1c scope):** `approval/http/errors.go:16` still imports `infrastructure/signature` to **map** its sentinel errors (`ErrInvalidCredentials`, `ErrRateLimited`, `ErrUnknownSignatureMethod`, `ErrRateLimiterConfig`) to HTTP codes. This is the standard edge error-mapping idiom (not type-coupling), and `application` already depends on `infrastructure/signature`. Relocating those sentinels to domain/application would touch `decision_service.go` + the signature sub-domain — a separate refactor. **Trigger to action:** when H-1d/H-1e or a future signature-verifier change reopens that sub-domain, lift the four sentinels to `approval/domain` (authn outcomes are domain-semantic) and have both delivery and the signature impl reference the domain errors.

#### H-1d — `*sql.DB` → `TxRunner` port (the hexagonal root fix)
Define `TxRunner` (`type TxRunner interface { Do(context.Context, func(db.Tx) error) error }`) + a postgres adapter wrapping `*sql.DB.BeginTx`. Replace the `*sql.DB` parameter in all approval (and CD/templates) **application** public methods with the runner so the application layer no longer receives the concrete DB type:
`decision_service.go:152, submit_service.go:48, publish_service.go:49, cancel_service.go:44, obsolete_service.go:44, supersede_service.go:43, scheduler_service.go:45, route_admin_service.go:115/241/393/540, read_service.go:44`.
**Boundary discipline:** keep `authz.Require` taking `*sql.Tx`; inside the runner callback, assert `db.Tx`→`*sql.Tx` at the authz call (existing `mustSQLTx` pattern). **Do NOT** lift `authz.Require`/`SeedTxIdentity` to `db.Tx` or re-key the capCache — that is **deferred boundary #4** (strongest area, risky). Update `main.go` wiring + test fakes (the custom `database/sql/driver` fakes in `coverage_boost_test.go` simplify to a synchronous runner fake).
**Verify:** build/vet + `go test -p 2 ./internal/modules/documents/approval/... ./internal/modules/controlleddocuments/... ./internal/modules/templates/... ./apps/api/...`. **Commit:** `refactor(approval): TxRunner port replaces *sql.DB in application signatures (H-1d)`

#### H-1e — documents delivery redesign (the one true REDESIGN trap)
Delete the `GeneratedServerAdapter` 29-method param-discard shim (`documents/delivery/http/generated_adapter.go`) + the `buildLegacyMux` wiring (`documents/module.go:118-173`). **Migrate `documents/delivery/http/Handler` to implement `documentsapi.ServerInterface` directly** (consume typed params; `handler.go:113-141`). **Collapse the second delivery subtree** `documents/http/` (fillin, placeholder_options, view, reconstruct, pdf webhook — ~5-6 files) into `delivery/http/` as sub-handlers behind the generated interface, with shared error helpers. **Route-truth-table before & after** (use **`metaldocs-backend-api`** skill — runtime registrations vs spec vs generated `ServerInterface`).
**What NOT to do:** leave `approval/http` (it's closest to target — `ServerInterfaceWrapper` + generated types; optional `HandlerWithOptions` upgrade is low-priority, not this commit). Do NOT change the public contract shape (no spec/regen unless a route is genuinely missing).
**Verify:** build/vet + `go test -p 2 ./internal/modules/documents/...` + `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` 0 + runtime route smoke. **Commit:** `refactor(documents): delete GeneratedServerAdapter, migrate Handler to ServerInterface, unify delivery (H-1e)`

---

### H-5 — Code quality  ·  audit: Code quality (B)  ·  assessor: async-workflow B− PATCH

**Steps:**
1. **`RecordSignoff` (407 lines, `decision_service.go:152-558`)** — **do NOT fragment the atomic transaction** (async assessor: the length is genuine indivisible approval-transaction complexity; splitting breaks atomicity). Instead **move the raw SQL helpers into `ApprovalRepository` methods**: `loadPriorSignoffs, loadStageSignoffs, loadActiveDocumentContentHash, hasUnresolvedComments, resolveEligibleActors, loadRoute` (`decision_service.go:582-741`, `submit_service.go:267-353`) → `ApprovalRepository` interface (`repository/approval_repository.go:59-81`) + postgres impl. After H-1d the BeginTx/Rollback boilerplate is already a callback, so this is the remaining work.
2. **Documents `Handler.Service` 28-method fat interface** (`delivery/http/handler.go:30-59`) — split into cohesive sub-interfaces; **drop the 4 unused methods**. After H-1e.
3. **`PeopleService.ListFiltered`** (`iam/application/people_service.go:511-581`) — rewrite to **filter + paginate in SQL**, killing the load-all-users + N+1 membership query + the swallowed error.

**Verify:** build/vet + `go test -p 2 ./internal/modules/documents/... ./internal/modules/iam/...`.
**Commit:** `refactor(quality): RecordSignoff SQL→repo, split documents Service interface, SQL-paginate PeopleService.ListFiltered (H-5)`

---

### H-2 — Composition / observability  ·  audit: Composition/obs (C)  ·  assessor: composition C PATCH ("C→A, no redesign") + observability C+ PATCH

**Steps (LAST — captures all prior main.go wiring changes):**
1. **Extract 13 inline adapters** from `apps/api/cmd/metaldocs-api/main.go` (lines 89, 93, 822, 833, 841, 882, 958, 972, 992, 1007, 1021, 1065, 1086) into `apps/api/internal/wiring/` files (`audit_adapters.go`, `documents_adapters.go`, `search_adapters.go`, `taxonomy_adapters.go`, `clock.go`) — ~300 lines out, beside the existing `wiring/documents.go`. Pure move; watch for import cycles (`go build` after each).
2. **Typed config loaders** in `internal/platform/config/`: `LoadFanoutConfig` (METALDOCS_FANOUT_URL + METALDOCS_DOCX_RENDERER_SERVICE_TOKEN), `LoadServerConfig` (APP_PORT), `LoadMigrationConfig` (METALDOCS_SKIP_STARTUP_MIGRATIONS + METALDOCS_MIGRATIONS_DIR), `LoadRetentionConfig` (AUDIT_RETENTION_DAYS). Delete the 9 bare `os.Getenv` in `main.go` (141, 171, 204, 205, 420, 424, 644, 694, 945) + 2 in `bootstrap/worker.go:50-51`. (METALDOCS_E2E gate may stay or become `LoadE2EConfig` — minor.)
3. **Decompose `main()`** (153-745) via extract-function (extend the existing `buildTaxonomyModule` pattern at 793-820): `buildApprovalPipeline`, `buildFanoutComponents`, `buildJobScheduler`, `buildPresenceSubsystem`, `buildDocumentModule`, `buildMux`, `buildServer` → `main()` ~100 lines.
4. **`slog.SetDefault(JSON)` per binary** (api/worker/jobs) before first log; **remove the private logger** at `observability/http.go:62` (use `slog.Default()`); replace every `log.Fatalf` → `slog.Error(...)` + `os.Exit(1)` in all 3 mains.
5. **Worker graceful drain** (`apps/worker/.../main.go` — drain the in-flight batch on signal before exit). **Jobs `Cleanup` on exit**: `apps/jobs/.../main.go:57` `log.Fatalf` skips `defer deps.Cleanup()` (os.Exit bypasses defers) → restructure so Cleanup always runs. **Per-binary OTel `service.name`**: worker/jobs call `SetupOTel` with `OTEL_SERVICE_NAME=metaldocs-worker` / `metaldocs-jobs`; `otel.go:50` hardcoded `"metaldocs-api"` → read from `OTEL_SERVICE_NAME` via `resource.Default()`.

**What NOT to do:** do NOT replace the bespoke `/api/v1/metrics` ring-buffer — that is **deferred boundary #1**. No DI framework (Wire/Fx) — manual wiring is correct.

**Verify:** build/vet + `go test -p 2 ./apps/... ./internal/platform/config/... ./internal/platform/observability/...` + runtime boot (`.\scripts\start-api.ps1 -Build`, login 200, structured JSON logs). **Commit:** `refactor(composition): extract main.go adapters+builders to wiring, typed config, slog/OTel/drain per binary (H-2)`

---

## Deferred boundaries (5) — bigger than Wave H; written triggers, NOT silent patches

| # | Boundary | Why deferred | Trigger |
|---|----------|--------------|---------|
| D-H1 | Bespoke `/api/v1/metrics` ring-buffer → OTel-metrics / Prometheus scrape (`observability/http.go:20-48,196-216,332-367`) | Introduces an ops/deploy dependency (collector); not OTel/Prometheus-scrapeable today. OTel SDK already vendored (Z-1) → path is clear. | Operator stands up a metrics collector / first SLO-alerting need. |
| D-H2 | Promote `documents/approval/` → peer module `internal/modules/approval/` (109 files) | Coupling to documents is already the canonical FK-based anti-corruption pattern → relocation is organizational, high churn / low correctness gain. A 109-file rename belongs in its own reviewed PR. | Next major approval feature / when the documents facade is otherwise refactored. |
| D-H3 | IAM/auth/search raw-mux → generated `ServerInterface` (`iam/.../people_handler.go:1-60`; auth/search have no `api/` package) | Needs spec-prefix normalization (`/iam` vs `/api/v1/iam`) + regen + 3-handler migration; FE-facing. A5 already fixed the one breaking endpoint. | Next IAM/auth contract change. |
| D-H4 | Lift `authz.Require`/`SeedTxIdentity` `*sql.Tx`→`db.Tx` + re-key capCache off `*sql.Tx` pointer identity (`iam/authz/authz.go:76`, `context.go:48`) | Strongest area; capCache pointer-keyed map (`assertedByTxPtr`) re-key is a real correctness risk. Operator directive: don't touch authz internals. H-1d keeps authz untouched via `mustSQLTx`. | Explicit authz-layer work. |
| D-H5 | Tier-1 `CanDo` per-request 4-union DB query → per-(user,tenant,cap) TTL cache (`iam/application/capability_service.go:40`) | Perf/load gap (uncached SELECT on every authenticated request). Touches the authz path. The "Redis authz cache" in the target diagram is aspirational — no Redis wired. | RF-3 / tenant-scale p95 regression (hundreds concurrent users). |

---

## Wave H DONE gate

All Tier-1 + Tier-2 rows ✅-or-deferred-with-trigger · static gates green (`go build` · `go vet` · `go test -p 2 ./...` · `api-lint -strict` 0 · `cilint` exit 0) · runtime QA (A1 WS upgrade 101, in-tx audit row, RLS NOSUPERUSER, 429/lockout, panic→500) · FE types regen + coverage green · audit dispositions + `roadmap.md` Wave H + `current-agent-handoff.md` close-out updated. **NOT merged — present evidence for operator review.**
