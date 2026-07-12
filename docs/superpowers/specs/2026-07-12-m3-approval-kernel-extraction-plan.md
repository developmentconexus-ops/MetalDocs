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
- **P2.S3** — route-admin contract delta (R1): openapi `CreateRouteRequest` + `RouteResponse` +
  `ListRoutesResponse` additive fields; regenerate; service accepts subject_kind/subject_key,
  defaults to document when only profile_code sent. Contract-diff additive-only proof.

**Phase-2 close:** L0+L1 green; existing document routes byte-equal; contract diff additive-only.

## Phase 3 — templates onto kernel

- **P3.S1** — in-flight count (R4): record live in-flight template-approval count + exact SQL in
  evidence.md; decide hard-cutover vs drain.
- **P3.S2** — template entry points (R2): openapi
  `/templates/{id}/versions/{n}/submit-for-approval` + `/signoff`; tier-1 caps; handlers onto the
  kernel application service (`subject_kind=template`, `subject_key=doc_type`).
- **P3.S3** — data migration: `ApprovalConfig{ReviewerRole?, ApproverRole}` → kernel routes (2-stage,
  or 1-stage when no reviewer) per doc_type; cutover rule from P3.S1.
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
