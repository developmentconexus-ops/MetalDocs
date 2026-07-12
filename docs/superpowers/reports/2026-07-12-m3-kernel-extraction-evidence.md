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
