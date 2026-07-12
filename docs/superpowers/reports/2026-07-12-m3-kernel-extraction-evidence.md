# M3 — Approval kernel extraction — EVIDENCE

**Unit:** ROADMAP 3.1 (approval-remediation M3) · **Branch:** `claude/nice-wu-353cd4`
**Spec:** `docs/superpowers/specs/2026-07-12-m3-approval-kernel-extraction-plan.md`
**P0 gate:** `docs/superpowers/analysis/2026-07-12-approval-kernel-extraction-system-impact.md` (🟡 Yellow)
**Ratification:** all 4 items APPROVED by operator via hub (as recommended) — logged below.

## Ratification log
- 2026-07-12 — ESCALATION sent (commit ffe604c6 base). ACK: R1 additive route-admin contract; R2(a)
  thin template entry points, retire parallel path; R3 3-phase relocate-then-generalize; R4 count
  in-flight first then hard-cutover/drain. openapi edits authorized within R1/R2 shapes.

## Dispatch ledger (HARNESS §4.4)

| Slice | Dispatch | Implementer | Reviewer | Gates | Commit | Status |
|-------|----------|-------------|----------|-------|--------|--------|
| P1.S1 relocate tree + imports | 2026-07-12 | sonnet | sonnet ACCEPT | go build+vet green; staged rename 0/0 byte-identical | b37c46d0 | DONE |
| P1.S2 re-port audit edge | 2026-07-12 | main (verify) | — | boundary-lint GREEN | b37c46d0 | VOID — no real violation |
| P1.S3 composition + codegen | 2026-07-12 | main (config align) | self+gates | api-lint 0 violations; unit tests green | 092a79d2 | DONE |
| P1.S4 supersede ADR 0072 + guard | 2026-07-12 | main | negative-plant proof | boundary GREEN; plant RED→revert-clean→GREEN | 7f407646 | DONE |
| P2.S1 migration + backfill | 2026-07-12 | sonnet | sonnet ACCEPT-WITH-NITS | build+vet; 8 testMig0296 GREEN (canonical); api-lint 0; check-db-bootstrap PASS | 82b897f1 | DONE |
| P2.S2 domain generalize | 2026-07-12 | sonnet | sonnet ACCEPT-WITH-NITS | byte-equal doc path; version-copy col-order CONFIRMED (indep + reviewer) | fe581164 | DONE |
| P2.S3 route-admin contract delta | 2026-07-12 | sonnet | sonnet ACCEPT | additive-only diff proven; regen-clean; byte-equal default | 9062f169 | DONE |
| P3.S1 in-flight count | 2026-07-12 | main | — | 0 under_review, 0 approved, 0 config → HARD CUTOVER | (evidence) | DONE |
| P3.S2a repository subject truth | 2026-07-12 | sonnet | sonnet ACCEPT | hard-require + 5-site hydration scan-order verified; template round-trip real; doc byte-equal | ec922e97 | DONE |
| P3.S2b template entry points | — | sonnet | sonnet (indep) | tier-1 caps reuse, subject-aware authz-area, kernel wire | — | pending |
| P3.S3 config→route migration | — | sonnet | sonnet (indep) | cutover rule applied | — | pending |
| P3.S4 retire parallel path | — | sonnet | sonnet (indep) | contract diff | — | pending |

## Gate results (fill per slice)

### P1.S2 finding — audit edge was a FALSE POSITIVE
The earlier-flagged `audit/delivery/http/handler.go → approval/http/router` edge is a COMMENT
reference (lines 3/74/76/80), NOT a real Go import. `sed -n '/^import (/,/^)/p'` on the audit
handler confirms NO approval import. `check-module-boundaries.ps1` → `[module-boundaries] OK` on
the relocated tree: the layer allow-list (`domain`/`application`/`api`) already covers every real
cross-module edge (documents↔approval, jobs→approval — all on allowed layers). No re-port required.
The stale `documents/approval` nested-family exception in the guard is now DEAD config (path no
longer exists) → cleaned in P1.S4 with negative-plant proof.

### P1.S3 gates
- `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → **0 violation(s)** (was 16:
  8 tripwire-allowlist entries stale at old path + 8 new-path violations unmatched; fixed by path
  rewrite in `scripts/api-lint/tripwire-allowlist.txt`).
- `go test ./internal/modules/{approval,documents,templates,audit}/...` → all **ok** (unit).
- `go build ./...` + `go vet ./...` → green.
- Pre-existing (NOT introduced here, out of ladder): `check-module-contract-sync.ps1 -Module approval`
  reports route-enumeration DRIFT; my diff to that script is paths-only and the grepped files
  (router.go/openapi/frontend) are byte-identical to main, so the drift predates this work
  (router.go has 0 occurrences of the literal `/api/v1/approval/inbox` pattern the check greps).

### P1.S4 gates — boundary guard realignment + negative-plant proof
1. Realigned guard on real tree → `[module-boundaries] OK`.
2. Plant `_ "metaldocs/internal/modules/approval/infrastructure"` in
   `internal/modules/jobs/stuck_instance_watchdog/job.go` (external module) → guard **FAIL**, names
   exactly `... job.go -> metaldocs/internal/modules/approval/infrastructure`.
3. `git checkout` → `git diff --exit-code` CLEAN; guard `[module-boundaries] OK` again.
- ADR 0082 written (supersedes ADR 0072 ruling (a)); ADR 0072 header stamped superseded.

### Phase-1 close — integration suite (L1)
- Run 1 (commit 7f407646): 11 failing = 9 accepted RED + **2 NEW regressions** I introduced.
  - 9 accepted (unchanged): controlleddocuments `TestTenantIsolation_SequenceCounters_CrossTenant`
    ×1 (E-PROD-2 document_profiles_pkey); jobs/approval_sla_surfacer ×4 (status ambiguous);
    scenarios ×3 (`TestGrantAreaMembershipFn`, `TestGrantAreaMembershipIdempotent`,
    `TestTriggerBypassBlocked`); tenantdata `TestTenantDataPortCoverage` ×1
    (approval_delegations/review_verdicts unregistered — pre-existing gap).
  - 2 NEW (FIXED, commit dc45f360): `TestReflect_RepositoryNoBeginTx`, `TestHTTPHandlers_NoBeginTx`
    in `tests/integration/scenarios/tx_ownership_test.go` — walk dir built from split
    `filepath.Join(root,"internal","modules","documents","approval",...)` args, so P1.S1's
    string-grep import sweep could not see it; the moved dir vanished. Repointed to
    `internal/modules/approval/{infrastructure,http}` (tx-ownership invariant guards — repaired,
    not deleted). `go vet -tags=integration ./tests/integration/scenarios/` clean.
- Run 2/3 (commit dc45f360): **INVALID** — Postgres OOM-crashed mid-suite (3 back-to-back heavy
  integration runs); every failure was `FATAL: the database system is in recovery mode (SQLSTATE
  57P03)` / `driver: bad connection`, NOT assertion failures. Root cause: **429 orphan
  `metaldocs_test_*` databases** left by the killed runs bloated WAL crash-recovery (`resetting
  unlogged relations` 600s+). Not a code regression.
- **Run 4 — CLEAN CLOSE (commit dc45f360):** waited for recovery (ready after ~605s), dropped all
  429 orphan test DBs (→ 0), ran canonical `test-integration.ps1` once. Full failure set =
  **exactly 9 tests / 4 pkgs, byte-match to accepted baseline, zero NEW:**
  - `controlleddocuments/application` `TestTenantIsolation_SequenceCounters_CrossTenant` ×1 (E-PROD-2)
    — re-ran (13.2s, not cached), identical accepted failure.
  - `jobs/approval_sla_surfacer` ×4 (`FullTick`, `Writer_TenantSeed`, `Idempotent_SecondRunNoOp`,
    `AlertOnly`) — error `column reference "status" is ambiguous (SQLSTATE 42702)`, a **pre-existing
    query defect**. Decisive proof of behavior-neutrality: this package imports `approval/domain`
    (repathed → build-cache invalidated → **forced re-run**) and reproduced its *identical*
    pre-existing error against the moved import path.
  - `scenarios` ×3 (`TestGrantAreaMembershipFn` invalid area_code, `TestGrantAreaMembershipIdempotent`,
    `TestTriggerBypassBlocked` session_replication_role/BYPASSRLS dev-env).
  - `tenantdata` `TestTenantDataPortCoverage` ×1 (`approval_delegations`/`approval_review_verdicts`
    unregistered ports — pre-existing gap, predates this work).
  - tx_ownership regressions (`TestReflect_RepositoryNoBeginTx`, `TestHTTPHandlers_NoBeginTx`) —
    **gone** (P1.S1 fix confirmed). `test-integration.ps1` exit 1 = throws on any RED incl. accepted.

**PHASE 1 (pure relocate) — CLOSED.** Boundary-lint GREEN + negative-plant; byte-equal behavior
proven (every re-compiled dependent reproduces its exact pre-existing accepted-RED; zero new). ADR
0072 ruling (a) superseded by ADR 0082 here.

## Phase 2 — generalize (subject_kind, subject_key)

### P2.S1 gates — migration 0296 (expand phase)
- Files: `db/migrations/0296_approval_subject_generalization.sql` + `tests/integration/migrations/migration_0296_test.go` (8 tests).
- `go build ./...` + `go vet -tags=integration ./tests/integration/migrations/...` → green.
- `.\scripts\test-integration.ps1 -Package ./tests/integration/migrations/... -Run TestMigration0296`
  → **PASS** (canonical runner, DATABASE_URL derived from .env — never hand-set). 8/8 GREEN.
  Coverage: backfill values both tables, CHECK reject (23514), unique reject (23505),
  partial-index allows inactive dup (route-versioning regression guard), template-subject
  insertable both tables, kept-index existence.
- `api-lint -strict` → **0 violations** (DB-only slice, no contract change).
- `check-db-bootstrap.ps1` → forward-migration execution on clean bootstrap (fresh volume).
- Implementer runtime-truth correction: baseline `approval_routes_tenant_profile_key` was already
  dropped by 0287 route-versioning → migration keeps the ACTUAL post-0287 constraints
  (`approval_routes_active_profile_uq`/`_profile_version_uq`). `ux_approval_routes_tenant_subject`
  made PARTIAL `WHERE active` to match (superseded rows share subject_key).
- Independent reviewer: **ACCEPT-WITH-NITS**, zero must-fix correctness defects. Nit-1 (compat-trigger
  removal not tracked) FIXED — debt now recorded in migration header + P2.S2 plan line. Nit-2
  (index existence-only assertion) left: partial/unique semantics covered by dedicated tests.
- Expand-phase compat shim: `default_approval_subject()` BEFORE-INSERT trigger backfills subject
  cols from legacy document cols when omitted → existing Go/testdb INSERTs work under new NOT NULL
  without a Go cutover. **Contract-phase debt (drop in P2.S2) tracked.**

### P2.S2 gates — domain Subject(kind,key) + explicit persistence (commit fe581164)
- New: `internal/modules/approval/domain/subject.go` (`SubjectKind` enum document|template, `Subject{Kind,Key}`,
  `NewDocumentSubject`, `Validate`/`Equal`/`String`) + `subject_test.go` (value object + Route/Instance
  projection invariants). `Route`/`Instance` gain `Subject` field.
- Production repo now writes `subject_kind`/`subject_key` EXPLICITLY on all 3 INSERT paths:
  `InsertInstance` (from `inst.Subject`, zero-value fallback → `NewDocumentSubject(document_id)`),
  route create (`route_admin_service.go` ~211), route version-copy/supersede (SQL `INSERT ... SELECT`
  copies source row's own subject cols). Compat trigger now a **production no-op** (still backstops testdb).
- Read hydration: `Subject` DERIVED in Go from legacy col (`NewDocumentSubject(profile_code/document_id)`)
  at all Route/Instance hydration sites — lower risk, provably equivalent for document rows
  (backfill set subject_key=document_id/profile_code). Diverges only if non-document subject_key ever
  differs → safe this phase (document-only); revisit at P3 template rows.
- Gates: `go build`/`go vet` clean · `go test ./internal/modules/approval/...` green ·
  consumers (documents, templates) green · full `go test ./...` (non-integration) no FAIL ·
  canonical `test-integration.ps1 -Package ./internal/modules/approval/...` PASS + `./tests/integration/approval/...`
  PASS (document submit/signoff byte-equal) · api-lint **0** · `check-module-boundaries.ps1` **OK**.
  No accepted-RED baseline test touched/newly broken.
- Reviewer **ACCEPT-WITH-NITS**, zero must-fix. Highest-risk item (version-copy `INSERT...SELECT`
  column order) CONFIRMED correct by BOTH an independent orchestrator read and the reviewer
  (8-col positional match, copies source route's own subject). Two nits (both "document-only-safe,
  breaks at P3 templates"): InsertInstance zero-Subject fallback + derive-from-legacy hydration —
  now tracked as explicit P3.S2 must-close items in the plan.

### P2.S3 gates — route-admin subject contract delta (commit 9062f169)
- **R1 additive-only, zero breaking change** — `git show 9062f169 -- api/openapi/v1/openapi.yaml`:
  `CreateRouteRequest.required` unchanged `[profile_code, name, stages]`;
  `RouteSummary.required` unchanged `[id, name, tenant_id, profile_code, active, version, stages,
  created_at, updated_at]`. `subject_kind` (enum document|template) + `subject_key` added as OPTIONAL
  props outside both `required` lists. No field removed, no new required entry, no type/pattern/enum
  narrowing on any existing field. **Reviewer determination: NO breaking change.**
- **Regen-clean** — `go generate ./internal/modules/approval/api/...` reproduces committed `api.gen.go`
  byte-for-byte (working tree clean after regen). New fields pointer-typed/optional
  (`SubjectKind *CreateRouteRequestSubjectKind`, `SubjectKey *string`). Not hand-edited.
- **Byte-equal default** — `resolveCreateRouteSubject` (route_admin_service.go): both fields absent →
  `Subject=(document, profile_code)`, identical persisted row to pre-slice. Proven two ways:
  fake-driver INSERT-arg capture unit test + real-Postgres `SELECT subject_kind, subject_key`
  integration test (`route_admin_service_subject_integration_test.go`). `profile_code` stays required
  throughout (contract, input struct, INSERT).
- **Enum validation at both layers** — HTTP `http/contracts/route.go` `CreateRouteRequest.Validate()`
  (reject unknown → 400) + domain `domain/subject.go` `Subject.Validate()` (`ErrInvalidSubjectKind`).
  Hand-written `contracts` package + generated `api.gen.go` both carry the fields, verified in sync
  (module decodes via hand-written contracts; gen types drive the strict-server route guard only).
- **Scope guard held** — grep of full commit diff: NO template routes, NO template governance, NO new
  capability, NO G1-policy change, NO P3.S4 retired-path work. `subject_kind=template` accepted +
  persisted faithfully with zero governance (Phase-3 wiring deferred).
- Gates (reviewer ran fresh, all green): `go build`/`go vet` clean · `go test
  ./internal/modules/approval/...` (8 pkgs) ok · consumers documents+templates ok · full `go test ./...`
  no FAIL · api-lint **0** · `check-module-boundaries.ps1` **OK** · canonical
  `test-integration.ps1 -Package ./internal/modules/approval/...` PASS, zero NEW failures (approval-only
  scope, none of the 9 baseline-RED pkgs exercised).
- Reviewer **ACCEPT**, zero must-fix, zero nits.

### Phase-2 close
- P2.S1 (DB expand + backfill + compat trigger) · P2.S2 (domain Subject + explicit persistence) ·
  P2.S3 (route-admin additive contract) all DONE + committed. Existing document routes/instances
  byte-equal; contract diff additive-only; kernel now keyed by `(subject_kind, subject_key)` with
  document as the projection. Ready for Phase 3 (templates onto kernel).

## Phase 3 — templates onto kernel

### P3.S1 gates — R4 in-flight template-approval count (2026-07-12)
- **State location (runtime truth):** templates approval state is templates-OWNED, NOT the kernel.
  It is `public.templates_template_version.status text` (CHECK ∈ {draft, under_review, approved,
  published, obsolete}). Static role config is `public.templates_approval_config`
  (`template_id` PK, `reviewer_role?`, `approver_role`). No `approval_instances` row / no
  `subject_kind='template'` exists yet. In-flight = `under_review` (submitted, undecided);
  reviewer-accepted-but-unpublished = `approved`. Rejected collapses to `draft` (no `rejected` state).
- **Counts (dev DB, cross-tenant; POSTGRES_USER is superuser+BYPASSRLS so RLS inert — genuine total):**
  - `SELECT status, count(*) FROM public.templates_template_version GROUP BY status;` → only `published|1`.
  - `SELECT count(*) FROM public.templates_template_version WHERE status = 'under_review';` → **0**
  - `SELECT count(*) FROM public.templates_template_version WHERE status IN ('under_review','approved');` → **0**
  - `SELECT count(*) FROM public.templates_approval_config;` → **0**
- **R4 decision: HARD CUTOVER.** Dev shows 0 in-flight approvals AND 0 approval-config rows — nothing
  to drain, nothing to migrate. No silent state loss possible.
- **Safety backstop (design constraint carried into P3.S3/P3.S4):** the P3.S3 data migration migrates
  ONLY the static `templates_approval_config` → kernel routes and MUST NOT mutate/destroy any
  `templates_template_version.status` row (the version enum is left intact), so even a non-dev target
  with nonzero in-flight versions is not silently lost. The actual old-path retirement (P3.S4) is the
  behavior switch; its deploy MUST re-run the `under_review`/`approved` count above on the target DB
  and only hard-cutover when both return 0 — nonzero → drain (old path finishes in-flight, new
  submissions route through the kernel). This preserves R4 for any DB regardless of what dev showed.

### P3.S2 design resolution — template subject-key semantics (CONTRADICTION RESOLVED)
- **Contradiction:** plan draft P3.S2 line said template `subject_key=doc_type`. Schema truth
  contradicts: `templates_approval_config` PK=`template_id` (per-template roles) and `doc_type_code`
  is NON-unique per tenant (`idx_templates_template_tenant_doctype`) → many templates share a doc_type.
  A doc_type-keyed route cannot honor the per-`template_id` approver/reviewer roles the config stores.
- **Resolution (schema truth beats plan wording; mirrors document two-level keying):**
  - ROUTE.subject_key = `template_id::text` (governance selector, ≙ document `profile_code`).
  - INSTANCE.subject_key = `templates_template_version.id::text` (artifact under approval, ≙ `document_id`).
  Both are `uuid::text` — the document instance path already casts `document_id::text` (0296:93).
- **Not a ratified-rail deviation** — R2's ratified shape only pinned "thin subject-scoped entry points
  onto the shared kernel"; `doc_type` was an under-specified impl detail, now corrected.
- **Kernel read-side gap for P3.S2b:** route SELECTION is still document-hard-coded
  (`LoadActiveRouteIDByProfile(tenantID, profileCode)`, SQL `WHERE profile_code=$2`, no subject_kind
  predicate). Template submit needs a subject-generic `LoadActiveRouteIDBySubject(tenantID, kind, key)`
  over the `ux_approval_routes_tenant_subject (tenant_id, subject_kind, subject_key) WHERE active`
  index; the profile_code method becomes a document specialization. (Route/Instance HYDRATION reading
  real subject columns is closed by P3.S2a.)

### P3.S2b authz pre-design — capability reuse + area-scope constraint (VERIFIED)
- **Tier-1 caps: REUSE, no new capability.** New routes map:
  `POST /templates/{id}/versions/{n}/submit-for-approval` → `CapDocumentSubmit` (`document.submit`);
  `/signoff` → `CapDocumentSignoff` (`document.signoff`). These ARE the kernel's caps — DB tripwire
  arms (`internal/platform/tripwire/arms.go:58-67`) force `approval_instances` INSERT→`document.submit`,
  `approval_signoffs` INSERT→`document.signoff`, so kernel-routed template submit/signoff is bound to
  them at tier-2 regardless. Minting `template.signoff`/`approval.*` would require widening arms +
  re-render + catalog/scope/seed edits — outside M3, not warranted. Registry edit: 2 rows in
  `apps/api/cmd/metaldocs-api/permissions.go` `routeRules` (templates block) + fixtures in
  `permissions_test.go` (`TestPermissionResolver` cases, `TestRouteCoverage` rows). NO `model.go`
  catalog, NO `capability_scope.go`, NO seed, NO tripwire edits.
- **Grant matrix: personas already hold the caps (no seed rows to add).** From
  `db/reference-data/0001_product_reference_data.sql`: author holds `document.submit` (L98) — covers
  the submit→`approval_instances` arm (author also holds `template.submit` L103); approver
  (`document.signoff` L82), qms_admin (`document.signoff` L113), system_admin (both) — cover the
  signoff→`approval_signoffs` arm (these hold `template.review`/`template.approve`). Overlap exact;
  no persona holds a template-approval cap but lacks the corresponding document cap.
- **HARD CONSTRAINT for the S2b kernel wiring (area scope):** `document.submit`/`document.signoff` are
  `ScopeArea` (area-grade, `capability_scope.go:41-42`); a template has NO process area. `authz.Require`
  resolves area-grade caps with `($2='tenant' OR upa.area_code=$2)` (`internal/modules/iam/authz/authz.go:155`).
  The kernel submit/signoff path MUST assert these caps for a template subject using the **`'tenant'`
  sentinel** (area filter OFF), NOT a derived area — templates have no `user_process_areas` area row,
  so a real areaCode fail-closes template approvers who DO hold the cap. Document subjects keep passing
  their real derived area. ⇒ P3.S2b needs **subject-aware authz-area resolution** in the kernel app
  service (document → derived area; template → `'tenant'`). The DB tripwire is area-blind (checks cap
  string only), so the sentinel path still satisfies it.
- **DEBT (post-M3, tracked in approval-tech-debt):** the kernel generalized SUBJECTS but its capability
  names + scope (`document.*`, area-grade) still read document-shaped. A future generalization to
  subject-agnostic `approval.submit`/`approval.signoff` (subject-appropriate scope) would remove the
  `'tenant'`-sentinel accommodation. Large cross-cutting change (new caps → 10-touchpoint walk +
  tripwire re-render + grant migration); explicitly out of M3 scope.

### P3.S2a gates — repository subject truth (commit ec922e97)
- **(a) hard-require:** `InsertInstance` zero-`Subject` fallback removed → `inst.Subject.Validate()`
  errors (ErrInvalidSubjectKind/ErrEmptySubjectKey); binds `inst.Subject.Kind/.Key` as-is. Only
  production caller `submit_service.go:202` already sets Subject — no caller broken. Compat trigger
  kept (backstops raw-SQL testdb factory, which omits subject cols).
- **(b) hydration:** 5 sites (LoadInstance, LoadActiveInstanceByDocument, LoadInstanceByDocumentForView,
  LoadInstancesByIDs, LoadRoute) now SELECT real `subject_kind`/`subject_key` (appended last) + Scan
  aligned (reviewer verified column-order↔Scan-order at EACH site); Subject built from real columns,
  legacy `profile_code`/`document_id` retained.
- **Tests (new `postgres_approval_repository_subject_integration_test.go`):** template round-trip
  (instance + route, key ≠ document_id/profile_code) RED under old derivation → GREEN; zero-Subject
  error test (`errors.Is` sentinel, zero rows); document byte-equal control. RED→GREEN proven via
  git-stash A/B, not inferred.
- Gates (reviewer ran fresh): build/vet clean (+ `-tags integration`); approval+documents+templates
  unit `ok`; api-lint **0**; boundaries **OK**; `test-integration.ps1 -Package
  ./internal/modules/approval/...` PASS (4 new tests individually PASS); `./tests/integration/approval/...`
  `ok`. Reviewer **ACCEPT**, zero must-fix/nits.
- **ENVIRONMENT FLAG (not a slice defect — pre-M3-close action):** broad `./tests/integration/...`
  run shows iam/scenarios/tenantdata failures with test-DB SCHEMA DRIFT (`relation
  "metaldocs.tenant_keys" does not exist`, `column "governance_class" of relation "document_profiles"
  does not exist`, `iam_user_roles_pkey` dup). Reviewer A/B (revert 2 files, re-run) → IDENTICAL with &
  without the slice → conclusively pre-existing stale/orphan test-DB template drift, ZERO new attributable
  to this slice. The recorded 9-baseline is stale vs the current environment. **ACTION before
  milestone-validator: rebuild the testdb template + drop orphan `metaldocs_test_*` DBs so the C-gate
  clean re-run reflects code truth, not env drift.**

### P3.S2b-1 STOP → P3.S2b-0 prerequisite migration (architecture contradiction, resolved in-boundary)
- **STOP (correct):** S2b-1 implementer refused to write the template-insert path — `approval_instances.document_id`
  is `uuid NOT NULL` FK→`documents(id)` (baseline :1971/:4129) and `approval_routes.profile_code` is NOT NULL
  FK→`document_profiles` (:4161). A template subject has no document row → template INSERT is impossible;
  the TDD suite could never go green. Migration 0296 EXPLICITLY deferred relaxing `document_id` NOT NULL to
  "a later phase" — my plan's P3.S2b never scheduled it. Gap in the plan, not authorization to schema-patch
  mid-slice. No code changed, no commit.
- **Resolution (in-boundary — 0296 pre-declared this, no ratified rail crossed):** insert P3.S2b-0, a relax
  migration BEFORE S2b-1. `DROP NOT NULL` on `document_id`/`profile_code`; KEEP both FKs (NULL-tolerant:
  single-col FK skips NULL, composite MATCH SIMPLE skips any-NULL → document rows stay integrity-checked,
  template rows NULL the legacy col); add projection CHECKs (document rows require the legacy key, template
  rows forbid it). Verify + cover template submit idempotency. Full design in plan §P3.S2b-0.
- **Not escalated to a new operator gate:** relaxing a pre-declared NOT NULL/FK to complete the milestone's
  committed expand/contract is within P3 scope; it fulfills 0296's own note rather than deviating from a rail.
  Flagged here for the HS-1 close review.

### P3.S2b-0 — relax migration 0297 (DONE, commit c70d890e)
Stalled implementer wrote both files (migration + migration_0297_test.go) but never gated/committed
(same stall as P2.S1). Recovered by running the ladder myself. Running the tests surfaced real defects
the implementer never caught by running them — fixed in-slice:
- **migration 0297 DDL: correct as written.** DROP NOT NULL on document_id/profile_code; both legacy FKs
  KEPT (NULL-tolerant under MATCH SIMPLE); 4 projection CHECKs (document/template × instances/routes);
  subject-scoped partial unique idempotency index alongside the kept legacy one; schema_migrations '0297'
  ON CONFLICT DO NOTHING. E-PROD-2 (document_profiles) untouched. All 10 migration_0297_test.go cases green.
- **0296 tests broke (4)** — they seeded template rows with legacy cols POPULATED (forced by the old
  NOT NULL); 0297's projection CHECK correctly rejects that shape (23514). Fix (0297 owns the test updates
  its invariant necessitates): removed the two now-superseded template-insertability tests (0297 re-proves
  them with the correct NULL-legacy-col shape); switched the two subject-unique-index tests to NULL
  profile_code so template rows satisfy the projection CHECK.
- **0297 `AllowsDistinctIdempotencyKeys` mis-designed** — two ACTIVE rows same subject collided on the
  pre-existing `ux_approval_instances_active_subject` (one-active-per-subject) index, never reaching the
  idempotency index it meant to test. Fixed: drive the first row terminal ('approved') so the active-subject
  partial index (WHERE in_progress) is out of play, isolating the new idempotency index. Same isolation
  applied to `RejectsDuplicateTemplateSubmit` (was passing for the wrong reason).
- **P3.S2a template read tests → RED-deferred to P3.S2b-1 (2, t.Skip'd).** Landing 0297 exposed that the
  repo READ paths are not NULL-legacy-col tolerant: `LoadInstance` INNER JOINs `documents ON ai.document_id`
  (drops NULL-document_id template rows → ErrNoActiveInstance) and scans document_id into a non-nullable
  string; `LoadRoute` scans profile_code into a non-nullable string ("converting NULL to string is
  unsupported"). Making the repo read path subject-generic (LEFT JOIN + nullable scans + subject-keyed
  queries) is squarely P3.S2b-1, not this schema-only slice. Skipped with explicit "unskip in P3.S2b-1"
  markers; document byte-equal + zero-subject error tests stay green. This is a genuine latent P3.S2a gap
  (its read-hydration was only ever tested against template rows still carrying legacy cols).
- **Gates:** go build ✓, go vet -tags integration ✓, api-lint -strict → 0 ✓, check-module-boundaries → OK ✓,
  test-integration migrations → PASS ✓, test-integration ./internal/modules/approval/... → PASS ✓.
  No production Go changed (schema + tests only).
- Independent SONNET reviewer dispatched for this slice (verdict pending).

### P3.S2b-1 — repo read-path NULL-legacy-col tolerance (DONE, commit 3e8a48e6)
SONNET implementer, TDD RED→GREEN on the two P3.S2a tests deferred by S2b-0.
- **Change (infra-only, 2 files, 40+/24-):** LoadInstance (~315), LoadInstancesByIDs (~929), LoadRoute
  (~1758) switched INNER→LEFT JOIN documents + nullable scans (document_id/revision_version via
  sql.NullString/NullInt64 → ""/0; profile_code via sql.NullString → ""). Subject still hydrated from real
  subject_kind/subject_key (P3.S2a). LoadActiveInstanceByDocument / LoadInstanceByDocumentForView left
  INNER (document-keyed WHERE document_id=$x — template rows never match). No write/domain/contract change.
- **RED (unskipped, pre-fix):** LoadInstance → "no active approval instance for document" (INNER JOIN drop);
  LoadRoute → "converting NULL to string is unsupported". **GREEN (post-fix):** both pass; the two tests
  unskipped.
- **Gates:** go build ✓, go vet -tags integration approval/... ✓, api-lint -strict → 0 ✓, module-boundaries
  → OK ✓, test-integration ./internal/modules/approval/... → all 8 subpkgs ok ✓.
- Independent SONNET reviewer dispatched (verdict pending) — adversarial focus: does LEFT JOIN mask a
  dangling / cross-tenant document_id for DOCUMENT rows (regression axis)?
- **Reviewer verdict: ACCEPT** (5 axes cleared incl. document-path-regression). Non-blocking Finding #1:
  the LEFT JOIN's only document-row miss path is a `d.tenant_id != ai.tenant_id` split-brain, which used
  to error under INNER JOIN but now returns a row with RevisionVersion silently collapsed to 0 — a narrow
  deviation from the no-fallback fail-closed principle. Judged in-boundary + small → folded in now.
- **Hardening follow-on (commit 52742413):** fail-loud no-fallback guard in LoadInstance +
  LoadInstancesByIDs — `subjectKind=="document" && !revisionVersion.Valid` now returns an integrity error
  instead of emitting RevisionVersion=0. Template rows (NULL document/revision) unaffected. Gates: go build
  ✓, go vet -tags integration ✓, test-integration ./internal/modules/approval/... → all subpkgs ok ✓.

### P3.S2b-2 — subject-generic route selection (DONE, 3-commit unit: c0ae114a + 33aa6d10 + 73904846)
Add `LoadActiveRouteIDBySubject(tenantID, subject_kind, subject_key)`; `LoadActiveRouteIDByProfile` becomes a
document specialization delegating with `("document", profileCode)`. SONNET implementer + independent SONNET
review per commit (never the implementer).
- **c0ae114a (impl):** new selector on repo + `SubmitDefaultsResolver` interface; delegation rewrite; real-DB
  test (template selection + document delegation-equivalence). Gates all green. 0296 backfill confirmed
  (`subject_key = profile_code` for document routes) — delegation equivalence rests on it.
- **REGRESSION (reviewer, blocking):** delegation filters document lookup on `subject_key`, but
  `resolveCreateRouteSubject` permitted a document route with an explicit `subject_key != profile_code`
  (proven by the then-shipping `TestCreateRoute_SubjectFieldsPassedThrough` fixture) → such a route becomes
  unfindable → document submit regresses to `ErrApprovalRouteMissing`. Root cause = MISSING R1 alias invariant
  (`document ⇒ subject_key == profile_code`), not the delegation. Verified against source before acting.
- **33aa6d10 (fix, two-layer):** (a) app friendly-first-line — `resolveCreateRouteSubject` now returns
  `(Subject, error)`, rejects document+divergent-key with new sentinel `ErrDocumentSubjectKeyMismatch`;
  (b) DB last-line — migration `0298_approval_route_document_subject_key_alias.sql` normalizes any divergent
  dev rows then adds CHECK `subject_kind <> 'document' OR subject_key = profile_code` (0297 house style).
  Corrected 3 incoherent fixtures to coherent shapes (template subject / key==profile_code) + added negative
  tests (unit + real-DB) + delegation-equivalence real-DB test. Supersede path proven divergence-safe
  (INSERT…SELECT carries both cols forward from the locked row). Full ladder green.
- **RE-REVIEW (independent):** invariant closed at every write path (create/supersede/in-place/backfill/
  trigger) + DB CHECK — data-integrity regression genuinely closed. One remaining blocking gap:
  `ErrDocumentSubjectKeyMismatch` unmapped in `MapErrorToResponse` → surfaced as 500 not the intended 4xx.
- **73904846 (addendum):** wired `ErrDocumentSubjectKeyMismatch` → 422 + typed code
  `approvalCodeValidationDocumentSubjectKeyMismatch` (mirrors `ErrReasonForChangeRequired`/
  `ErrRevisionTitleRequired` convention); HTTP handler test RED (500) → GREEN (422). Full ladder green.
- **Disposition: ACCEPT** — 3-commit unit sound; regression closed; all reviewer findings resolved.

### P3.S2b-3 — subject-generic write path (decomposed: 3a area-resolution, 3b template path)
**Architecture (Global-Maximum):** the document `SubmitService`/`decision_service` orchestration is deeply
document-eQMS-shaped (draft→under_review transition, CD-link, content hash, revision title, reason-for-change)
— NOT generalized in place. The shared kernel = domain + repository + authz layer (already shared). Templates
get a THIN parallel write path reusing kernel primitives; document orchestration untouched beyond area
resolution. Rails-consistent (R2a "thin entry points on shared kernel").

#### P3.S2b-3a — subject-aware authz-area resolution (DONE, commit 6825afc3, reviewer ACCEPT)
- New `application/subject_area.go`: `resolveSubjectAreaCode(ctx, tx, cdRead, tenantID, subject)` — template →
  `"tenant"` sentinel (authz.go:155 `($2='tenant' OR upa.area_code=$2)` → area-blind, since templates have no
  process area); document → delegate to `docapp.LoadDocumentAreaCode(subject.Key)`; unknown kind → error.
- Call sites switched: `submit_service.go:95` (`NewDocumentSubject(req.DocumentID)`), `decision_service.go:264`
  (`instance.Subject`). Document-path byte-identical (subject.Key == documentID for document instances).
- Deviation (disclosed, legit): 4 pre-P3.S2a signoff-test fixtures never set Subject → zero-value rejected by
  the new helper; added `Subject: NewDocumentSubject(...)` (matches production hydration). Not a prod gap.
- **Reviewer ACCEPT (6 axes).** Axis-2 hammered: every prod signoff path re-runs LoadInstance internally +
  0296 NOT NULL/CHECK/backfill/trigger ⇒ no unhydrated Subject reachable. Axis-3: `area_code_not_tenant` CHECK
  (baseline :1061) makes a real 'tenant' area impossible — sentinel collision structurally excluded.
- Gates: build ✓, api-lint 0 ✓, module-boundaries OK ✓, vet ✓, unit ✓, integration ✓. Zero doc-path change.

#### P3.S2b-3b — thin template submit + signoff path — IN PROGRESS (decomposed i/ii/iii)

**3b-i — area-blind ResolveEligibleActors ('tenant' sentinel) — DONE, commit `182909c4`, main self-review ACCEPT**
- SQL delta: `AND area_code = $2` → `AND ($2 = 'tenant' OR area_code = $2)` in `postgres_approval_repository.go`
  ResolveEligibleActors; mirrors authz.go:155 `($2='tenant' OR upa.area_code=$2)` exactly. Only area
  filter in the fn. Reads `metaldocs.v_active_user_areas`.
- Identity-model check (STOP-candidate, resolved no-STOP): view = one row per (user,area,role);
  `$2='tenant'` makes predicate unconditionally true → returns the role holder in ANY area. No synthetic
  tenant-wide row needed. Real-area path: `$2<>'tenant'` → OR-left false → `area_code=$2` byte-identical.
- Test `eligible_actors_area_blind_integration_test.go` (+100): seeds tenantWideApprover (area 'quality')
  + scopedOnlyInX (area 'safety'); area-blind ('tenant') finds approver (RED empty pre-fix → GREEN post),
  area-scoped ('quality') assertion proves no real-area behavior change. Full gate ladder PASS; no deviations.
- **3b-ii — thin template submit path** — **STOP / ESCALATED** (architecture contradiction in authz tripwire kernel). Implementer work green at L1 (2 real-DB tests PASS) but blocked at L0 (api-lint). WIP preserved `git stash@{0}`. Details below.
- **3b-ii — thin template submit path** — **DONE**, commit `aed56af2`, independent reviewer **ACCEPT**. Resumed after kernel #26 landed: popped WIP stash cleanly, swapped assertion to `CapTemplateSubmit` (ScopeTenant, area-blind), generalized DRIFT lint (`unionArmCapsFor` — union caps across all arms matching a discriminated `(table,op)`; 2 new lint tests, no fixture weakened), STOP-candidate #1 re-confirmed (submit reads `stage.AreaCode` verbatim, no 'tenant' forcing). Reviewer definitively cleared the false-green risk: `authz.Require` short-circuits only the grant lookup for `system_admin` but still asserts the literal `template.submit` into `metaldocs.asserted_caps`; `testdb` connects as `metaldocs_ci` (NOSUPERUSER/NOBYPASSRLS) so the trigger fires; migration 0299 template arm requires exactly `template.submit`. Full integration ladder green (all 8 approval sub-pkgs). Non-blocking nit: `armFor` now dead code (remove in follow-on).
- **3b-iii — template signoff + completion port** — pending. Same kernel gap on `approval_signoffs` (arm hardcodes `document.signoff`), BUT signoff rows have no direct `subject_kind` column → discriminator strategy TBD (subquery vs denormalize). Being scouted before dispatch.

##### P3.S2b-3b-ii STOP — tripwire arm kernel is not subject-generic (ADR 0082 gap)
**Root cause.** ADR 0082 generalized `approval_instances` to a SHARED table (`subject_kind ∈ {document, template}`), touching the table's columns/indexes/constraints — but NOT the authz last-line-of-defense trigger `enforce_capability_asserted()` nor its Go source of truth `internal/platform/tripwire.TripwireArms`. Arm #1 still hardcodes `(approval_instances, INSERT) → [document.submit]`.

**Why it can't be fixed inside the slice (both attempts empirically confirmed):**
- Assert `CapDocumentSubmit` (`ScopeArea`) with the `"tenant"` area-blind sentinel (required for template stages, ratified 3b-i/3a) → integration GREEN but api-lint `authz-area-scope-binding` FAILS (ADR 0022 Phase 7: area-grade caps may not be enforced with literal `"tenant"`). Reproduced ×2.
- Assert `CapTemplateSubmit` (`ScopeTenant`, correct area-blind classification — already in the registry) → api-lint clean but DB trigger rejects: `ErrCapabilityNotAsserted: none of {document.submit} present … on approval_instances (P0001)`.

**Why the naive kernel widen is a SECURITY REGRESSION (rejected).** The trigger match is **match-one** (`arms.go`: "any one present in asserted_caps satisfies the branch"). Widening arm #1 to `[document.submit, template.submit]` would let a holder of ONLY `template.submit` authorize a **document**-subject INSERT. The `Arm` model `(table, op)→caps` has no column-value discriminator, so it cannot express "document rows require document.submit; template rows require template.submit" on a shared table. (Template-only tables like `templates_template_version` avoid this because they are not shared with documents.)

##### P3.S2b-3b-0 — tripwire arm subject-discriminator — DONE (commits `e0587887` + fix `3ae1d77f`), independent reviewer REQUEST-CHANGES→fixed→ACCEPT
Operator ratified Option A + ADR (2026-07-12). **ADR 0083** written (`ab497f94`), extends 0082.
- **`e0587887`** — `tripwire.Arm` gains optional `WhenColumn`/`WhenValue`; arm #1 split → document→`document.submit` / template→`template.submit`; `RenderMigration()` emits nested `CASE NEW.subject_kind` with fail-closed P0001 ELSE; golden migration + api-lint path + tests. PARITY+DRIFT green; every non-`approval_instances` branch byte-identical (independently diff-verified by reviewer).
- **Reviewer REQUEST CHANGES** (independent, adversarial): migration numbered `0284` collided with pre-existing `0284_ci_rls_role.sql` (committed `6e971e73`); `migrate.Apply` dedupes by 4-digit prefix → tripwire migration SILENTLY SKIPPED wherever 0284 already applied = feature no-op. Blocking. Arm logic itself verified sound (disjoint arrays, no union, fail-closed).
- **`3ae1d77f`** — mechanical renumber `0284→0299` (true next-free after 0298) across migration file (git rename) + `gen-tripwire` defaultRelPath + render.go header/ledger `VALUES('0299')` + api-lint `tripwireMigrationPath` + arms_test golden. Regenerated; rendered==committed 0299; full gate ladder green; `0284_ci_rls_role.sql` untouched. Verified by hub.
- **Follow-on (into #24):** DRIFT `armFor` returns only the FIRST arm for a discriminated `(table,op)` — harmless now (no committed template-writer of `approval_instances`) but #24 MUST generalize it (union caps across all arms matching `(table,op)`) before landing `TemplateSubmitService`, else false-positive DRIFT violation.

**Superseded plan (kept for history) — Global-maximum fix (ratified, HS-7):** extend the tripwire arm kernel with an optional subject discriminator (`WhenColumn`/`WhenValue`, e.g. `subject_kind`); `RenderMigration()` emits a nested `CASE NEW.subject_kind` for `approval_instances`; split arm #1 → document→`document.submit`, template→`template.submit`. Regenerate golden migration (new `db/migrations/0284_*.sql`), bump `tripwireMigrationPath`, keep TRIPWIRE-ARM-PARITY + TRIPWIRE-ARM-DRIFT green. Then 3b-ii resumes trivially (assert `CapTemplateSubmit`). Same discriminator (via join) later resolves the identical `approval_signoffs` gap for 3b-iii. **Blocked on operator ratification because it amends the binding GMR M2 validation-contract arm set (arms.go HS-7 clause) and touches the authz last line of defense.**

## Baseline (pre-work)
- Accepted RED on main: exactly 9 tests / 4 pkgs (E-PROD-1..5: sla_surfacer ×4, controlleddocuments
  cross-tenant sequence ×1, scenarios ×3, tenantdata ×1). Bar for every slice: zero NEW failures.
- approval subtree: 164 Go files. Coupling: 2 inbound production files (documents→approval), 24
  outbound (approval→documents domain/application), 1 true re-port (audit→approval/http/router),
  3 external consumers on allowed layers (audit, jobs/approval_sla_surfacer, jobs/stuck_instance_watchdog).

## Defers / notes
- E-PROD-2 (document_profiles PK) untouched — operator decision pending.

## HS-1
- Operator sign-off gate: pending (milestone close).
