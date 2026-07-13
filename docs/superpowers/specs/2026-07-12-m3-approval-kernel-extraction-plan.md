# M3 — Approval kernel extraction + templates unification (ROADMAP unit 3.1)

**Date:** 2026-07-12
**Branch:** `claude/nice-wu-353cd4`
**Status:** RATIFIED (operator via hub, all 4 items as recommended) — implementation authorized.
**Rails:** `docs/superpowers/analysis/2026-07-12-approval-kernel-extraction-system-impact.md` (P0 gate, Yellow).
**Spec source:** `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md:177-205` (Milestone C).

## Goal

Extract `internal/modules/documents/approval` → top-level `internal/modules/approval` (15th bounded
context). Generalize the kernel to `(subject_kind, subject_key)`. Rewire templates approval onto the
kernel, retiring the parallel templates-local approval path. Supersede ADR 0072. G1/G2/G3 behavior
byte-equal.

## Ratified contract shapes (locked)

**R1 — route-admin contract (additive-only; zero breaking change is a done-criterion):**
- `CreateRouteRequest` gains optional `subject_kind` (enum `document|template`, default `document`)
  + `subject_key` (string). `profile_code` **stays** as backward-compat alias =
  `(subject_kind=document, subject_key=<profile_code>)`.
- `RouteResponse` / `ListRoutesResponse` gain `subject_kind` + `subject_key` (additive).
- Contract diff MUST be additive-only (no removed/narrowed fields on existing document paths).

**R2 — template entry points (option a):**
- Add thin subject-scoped entry points mirroring the documents pattern:
  `/templates/{id}/versions/{n}/submit-for-approval` + `/templates/{id}/versions/{n}/signoff`
  onto the shared kernel.
- **Retire** the parallel path: `/templates/{id}/approval-config` +
  `/templates/{id}/versions/{n}/approve`.
- Each new route gets a tier-1 route→capability mapping (capabilities-not-roles, ADR 0022).

**R3 — sequencing (3 phases, each provable-closed before the next; commit per green slice):**
- **P1 pure relocate** — move files, fix import paths, re-port the one audit edge, boundary-lint
  GREEN, ALL tests green, ZERO behavior change. **Supersede ADR 0072 in this phase.**
- **P2 generalize** — add `subject_kind`/`subject_key` columns + route-admin contract delta +
  backfill existing rows to `(document, profile_code)`. Expand/contract.
- **P3 templates onto kernel** — new entry points + retire parallel path + `ApprovalConfig`→route
  data migration + in-flight cutover.

**R4 — in-flight cutover:** count live in-flight template approvals FIRST (record count + exact query
in evidence.md). Zero → hard cutover. Nonzero → drain (old path finishes in-flight, new template
versions on kernel). No silent state loss.

## Carried constraints (re-confirmed)

- **AuthZ capabilities never roles** (ADR 0022); two-tier PDP; tripwire arms regenerated for the new
  module path. No new capability unless a real gap surfaces (then full 10-touchpoint walk).
- **H-PRE-1** — authz-recording reads off lock-holding tx.
- **Migrations expand/contract**; `tenant_id` preserved; **idempotency route templates (signoff +
  fast-forward) keep working across the move**.
- **E-PROD-2 DO-NOT-TOUCH** — `document_profiles` PK=(code) not tenant-scoped is an open operator
  decision; note if approached, do not repair, do not collide.
- **Direction:** `render`/`search`/`notifications` must not gain an approval dependency.

## Phase 1 — pure relocate (behavior byte-equal)

Slices (one sonnet TDD dispatch each; independent sonnet reviewer per slice; commit per green):

- **P1.S1 — relocate tree + package rename.** `git mv internal/modules/documents/approval/*` →
  `internal/modules/approval/`; update the `package` clauses only where the path changes package
  identity (subpackages keep their names). Fix ALL import paths `.../documents/approval/...` →
  `.../approval/...` repo-wide (production + tests + composition root + worker/jobs binaries). No
  behavior change. Gate: `go build ./...` green.
- **P1.S2 — re-port the one violation-class edge.** `audit/delivery/http/handler.go` imports
  `approval/http/router` (not a published surface). Re-port to consume approval's published
  application/api surface (or expose the needed router hook via a published port). Gate:
  `check-module-boundaries.ps1` GREEN with `approval` as a first-class module (drop the
  `documents/approval` nested-family exception in the guard script).
- **P1.S3 — composition-root + codegen rewire.** New `api/cfg.yaml` include-tags path; regenerate;
  wire the module in `metaldocs-api` + worker + jobs binaries at the new path. Gate: `api-lint
  -strict` + `go build ./...` + full `go test ./...` green (accepted RED = the 9 E-PROD baseline).
- **P1.S4 — supersede ADR 0072 + boundary-guard realignment.** New ADR (kernel extraction; records
  the fired promotion trigger; realigns `check-module-boundaries.ps1` — `approval` first-class, drop
  nested exception). Negative-plant proof (external module importing `approval/infrastructure` →
  guard RED) + revert-clean, mirroring ADR 0072's proof discipline.

**Phase-1 close criterion:** `go build ./...` · `api-lint -strict` · `check-module-boundaries.ps1`
GREEN · `go test ./...` + `test-integration.ps1` = only the 9 accepted RED, ZERO new failures ·
document approval lifecycle unchanged. Behavior byte-equal proven BEFORE any P2 semantic change.

## Phase 2 — generalize `(subject_kind, subject_key)`

- **P2.S1** — migration: add `subject_kind`/`subject_key` to the approval route/instance tables,
  backfill existing rows to `('document', profile_code)`, CHECK constraints (DB enforces). Keep
  `profile_code`/document keying working (expand). testdb integration proving backfill + old rows
  still resolve.
- **P2.S2** — domain generalization: route/instance value objects keyed by `(subject_kind,
  subject_key)`; `profile_code` becomes a document-subject projection. Byte-equal for document path.
  **Includes (contract-phase debt from P2.S1):** once production repositories AND the
  `tests/integration/testdb` factory write `subject_kind`/`subject_key` explicitly, DROP the
  temporary `public.default_approval_subject()` compatibility triggers/function introduced by
  migration 0296 (a follow-on contract migration). Leaving them in place would silently default a
  forgotten subject write to `'document'`, masking the very bug the CHECK constraint guards.
- **P2.S3** — route-admin contract delta (R1): openapi `CreateRouteRequest` + `RouteResponse` +
  `ListRoutesResponse` additive fields; regenerate; service accepts subject_kind/subject_key,
  defaults to document when only profile_code sent. Contract-diff additive-only proof.

**Phase-2 close:** L0+L1 green; existing document routes byte-equal; contract diff additive-only.

## Phase 3 — templates onto kernel

- **P3.S1** — in-flight count (R4): record live in-flight template-approval count + exact SQL in
  evidence.md; decide hard-cutover vs drain.
- **P3.S2a** (repository truth — DONE-as-own-slice) — close two P2.S2-deferred document-only
  assumptions that break with template rows: (a) `InsertInstance` zero-`Subject` fallback →
  `NewDocumentSubject(document_id)` (`postgres_approval_repository.go` ~49-52) removed / hard-require,
  else a forgotten template Subject is silently mis-tagged `document`. (b) Route/Instance
  read-hydration DERIVES `Subject` from the legacy `profile_code`/`document_id` column
  (`postgres_approval_repository.go` LoadRoute/LoadInstance/LoadActiveInstanceByDocument/
  LoadInstanceByDocumentForView/LoadInstancesByIDs) — MUST SELECT the real `subject_kind`/`subject_key`
  columns (only equivalent while all rows are document rows).
- **P3.S2b-0** (PREREQUISITE migration — surfaced by the P3.S2b-1 STOP; 0296 pre-declared it as "a
  later phase"). The template-INSERT path is blocked by legacy document-only NOT NULL + FK on the
  approval tables: `approval_instances.document_id uuid NOT NULL` FK→`documents(id)`
  (baseline :1971 / :4129); `approval_routes.profile_code NOT NULL` FK→`document_profiles(tenant_id,code)`
  (:4161). A `(template, version_id)` instance / `(template, template_id)` route has no such document row.
  RELAX (expand, forward-only, idempotent):
  - `ALTER COLUMN document_id DROP NOT NULL` (approval_instances) + `profile_code DROP NOT NULL`
    (approval_routes). **KEEP both FKs** — they are NULL-tolerant (single-col FK skips NULL rows;
    composite MATCH SIMPLE skips any-NULL), so document rows stay fully integrity-checked while template
    rows set the legacy col NULL. Do NOT drop the FKs.
  - Projection CHECKs (DB enforces the subject invariant): instances
    `CHECK (subject_kind <> 'document' OR document_id IS NOT NULL)` +
    `CHECK (subject_kind <> 'template' OR document_id IS NULL)`; routes the same on `profile_code`.
  - Template route uniqueness already handled by 0296 `ux_approval_routes_tenant_subject` (NULL
    profile_codes are distinct in the kept `approval_routes_active_profile_uq`, so no false collision).
  - **Verify submit idempotency:** if approval submit dedups via the DB index
    `approval_instances_document_id_idempotency_key_key (document_id, idempotency_key)`, that leaves
    template rows (NULL document_id → NULLs distinct) UN-deduped — add a subject-scoped unique
    (`(tenant_id, subject_kind, subject_key, idempotency_key) WHERE idempotency_key IS NOT NULL`) so
    template submit idempotency holds. If submit idempotency is via the platform `idempotency_keys`
    store instead, no approval-table change needed — confirm which and act accordingly.
  - testdb integration: template instance + route now INSERTable; document rows still rejected when
    legacy key missing; document path byte-equal. This migration UNBLOCKS P3.S2b-1.
- **P3.S2b** — template entry points (R2): openapi
  `/templates/{id}/versions/{n}/submit-for-approval` + `/signoff`; tier-1 caps; handlers onto the
  kernel application service (`subject_kind=template`). Adds a subject-generic route-selection method
  `LoadActiveRouteIDBySubject(tenantID, subject_kind, subject_key)` — the read/selection side is still
  document-hard-coded (`LoadActiveRouteIDByProfile` takes `profile_code`, no subject_kind predicate);
  `LoadActiveRouteIDByProfile` becomes a document specialization of it.
  **Tier-1 caps REUSE (verified, no new cap/grant):** submit-for-approval → `CapDocumentSubmit`,
  signoff → `CapDocumentSignoff` (the kernel's caps, forced by tripwire arms on
  `approval_instances`/`approval_signoffs` INSERT). Registry: 2 rows in `permissions.go` `routeRules`
  + `permissions_test.go` fixtures. Personas already hold both (author/approver/qms_admin/system_admin).
  **HARD CONSTRAINT:** `document.submit`/`document.signoff` are `ScopeArea`; templates have no process
  area, so the kernel MUST assert them for template subjects with the `'tenant'` sentinel (not a derived
  area) or tier-2 (`authz.go:155`) fail-closes template approvers. ⇒ subject-aware authz-area resolution
  in the kernel app service (document → derived area; template → `'tenant'`). Capability
  naming/scope generalization (`approval.*`) is post-M3 debt.

  **SUBJECT-KEY SEMANTICS — CORRECTED (2026-07-12, schema truth beats plan wording).** The earlier
  plan draft said `subject_key=doc_type`; that is WRONG and would lose per-template roles. Resolution,
  mirroring the document two-level keying:
  - **ROUTE.subject_key = `template_id::text`** (governance selector). Evidence: `templates_approval_config`
    PK = `template_id` (baseline :3121) with FK → `templates_template(id)`; per-template
    `reviewer_role`/`approver_role`. `doc_type_code` is NON-unique per tenant
    (`idx_templates_template_tenant_doctype`, baseline :3578) → many templates per doc_type, so a
    doc_type-keyed route could not honor the per-`template_id` roles the config already stores.
  - **INSTANCE.subject_key = `templates_template_version.id::text`** (artifact under approval), the direct
    analogue of `document_id`. Both template ids are `uuid`; `uuid::text` is the established pattern
    (document instance already casts `document_id::text`, migration 0296:93).
  This is NOT a ratified-rail deviation — the R2 ratified shape only pinned "thin subject-scoped entry
  points onto the shared kernel"; `subject_key=doc_type` was an under-specified impl detail, now
  corrected to the schema-truth grain.
- **P3.S3** — data migration: `templates_approval_config{reviewer_role?, approver_role}` → kernel routes
  (2-stage, or 1-stage when no reviewer) **per `template_id`** (subject_key=template_id::text; corrected
  from "per doc_type" per P3.S2b resolution); cutover rule from P3.S1 (hard cutover, 0 in-flight).
- **P3.S4** — retire parallel path: remove `/templates/{id}/approval-config` + `/approve`;
  `templates/domain/approval.go` (`CheckSegregation`, `HasReviewer`) + `approval_config.go` deleted;
  SoD/state-machine/audit now sourced from the kernel. Contract-diff shows the retired paths.

**Phase-3 close:** template lifecycle live QA (config→route migration verified; review+approve+publish
a template version through the kernel; worklist shows it; SoD + delegation enforced); document approval
lifecycle regression (M2b F8-class walkthrough) unchanged.

## Verification ladder (every slice)

- **L0:** `go build ./...` · `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` ·
  `.\scripts\check-module-boundaries.ps1` (core proof).
- **L1:** `go test ./...` + `.\scripts\test-integration.ps1` (canonical; NEVER hand-set DATABASE_URL).
  Accepted RED = exactly 9 tests / 4 pkgs (E-PROD-1..5). Bar: zero NEW failures.
- New integration tests: `tests/integration/testdb` (`seedWithCapsIdentity`), `//go:build integration`.
- FE: if generated api-types shift, `pnpm exec tsc --noEmit -p tsconfig.build.json` clean.

## Docs / ADR

- New ADR supersedes ADR 0072 (P1.S4).
- wiki-curator: `wiki/modules/approval.md` (12-section) + `approval-tech-debt.md` +
  `wiki/modules/index.md`; update `documents.md`/`templates.md` (approval consumed, not owned);
  `backend-target-architecture.md` module count 14→15 + REQ-TOP-1 reference.

## Close

- milestone-validator PASS before claiming done.
- Evidence: `docs/superpowers/reports/2026-07-12-m3-kernel-extraction-evidence.md` with dispatch
  ledger (HARNESS §4.4). CLOSED event to hub. Commits on chip branch. NEVER push.
