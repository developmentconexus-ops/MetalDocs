# Milestone 4b — Legacy schema cluster teardown

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** Spec approved
> **Authored:** 2026-06-15 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The
> end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Origin

This milestone exists because the M4 close gate (`milestone-4-systemic-ports/qa/milestone-qa.md`)
returned **FAIL** on F4.1a Gate #5 (`TestCreateDocumentTx_PopulatesAllSnapshotColumns` is
environment-coupled), and a `/systematic-debugging` pass proved the failure is a **symptom of a schema
defect**, not a harness bug:

- The runtime issues bare, unqualified `documents` SQL in 40+ sites; bare-name resolution is governed
  entirely by the connection `search_path` (no `ALTER DATABASE/ROLE SET search_path` exists anywhere
  in `db/`; the app config never injects one; production connects on the Postgres default → effective
  `public`).
- **Two** tables are duplicated across schemas — `documents` and `template_audit_log`. The
  `metaldocs.documents` copy is a dead legacy editor-era table (lacks `tenant_id` / `active_session_id`
  / `controlled_document_id`) that **shadows** the real `public.documents` whenever `search_path` is
  metaldocs-first (the operator/dev DSN's `search_path=metaldocs,public`). That shadow is harmless in
  production (public-first) but fails the test harness (`column "tenant_id" ... does not exist`,
  SQLSTATE 42703 — "Family A").
- `metaldocs.documents` anchors an **8-table dead FK satellite cluster** (`document_attachments`,
  `document_collaboration_presence`, `document_edit_locks`, `document_template_assignments`,
  `document_versions`, `document_versions_mddm`, `document_version_images`, `workflow_approvals`).
  The whole 10-object set (anchor + 8 satellites + `template_audit_log`) has **zero runtime Go
  references** (only one comment calling the schema "decommissioned"). Dead at the runtime layer —
  verified by F4b.1's census (`f4b.1-legacy-cluster-census/evidence.md`); `document_version_images` is
  a 2nd-level satellite caught during that census.

Operator decision 2026-06-15: **fix the root cause** (drop the dead duplicate cluster) rather than
adapt the harness `search_path`; and **fix the Family-B stale seeds now**. F4.1b
(`f4.1b-testdb-search-path-robustness`) is **superseded** by this milestone — no harness change is
made; `tests/integration/testdb/db.go` stays at HEAD, which is itself the proof this is a fix and not
an adaptation.

## Objective

Eliminate the dead `metaldocs.documents` legacy duplicate cluster (the table, its 8 FK-satellite
tables, their indexes/constraints/sequences) and the dead `metaldocs.template_audit_log` duplicate, so
that bare unqualified `documents` (and `template_audit_log`) resolves to the real `public.*` table
under **any** `search_path` ordering — permanently removing the shadow landmine. Separately, repair the
stale tripwire test seeds (Family B) so guarded-table writes assert their capability. The bar this
moves: **F4.1a Gate #5 re-greens deterministically with the operator DSN unchanged and the harness
untouched**, and the full integration suite is green from a clean curated baseline — clearing the M4
close-gate blocker and removing a latent correctness hazard ahead of the M5 re-audit.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4b.1 | `f4b.1-legacy-cluster-census` | Verify-dead census of every object slated for drop: `metaldocs.documents`, the 8 satellite tables, `metaldocs.template_audit_log`, and their indexes/constraints/sequences. Confirm zero dependents across **all** surfaces — Go runtime (done pre-flight: zero), curated baseline + migrations (views, triggers, RLS policies, FKs **inbound from any live table**), reference-data seeds, and any OpenAPI/contract anchor (cf. the `document_access_policies` keep-precedent in 0231). Produce a per-object verified-dead manifest. | A manifest listing every object to drop with its dependency proof (`0` inbound refs from any non-dropped object). **If any live dependent is found → HS-2 stop** (the cluster is not dead; report boundary, do not drop). |
| F4b.2 | `f4b.2-drop-migration` | A forward-only, idempotent `metaldocs-database` migration (`> 0239`) that `DROP`s the verified-dead objects from F4b.1 (`DROP TABLE IF EXISTS ... CASCADE`, satellites-before-anchor), with one `public.schema_migrations` row. The curated baseline (`db/baseline/0001_current_schema.sql`) is a **frozen snapshot** and is **left untouched** — per repo evidence (`document_subjects`/`templates_template.areas` dropped by 0236 still live in the baseline) the baseline is never re-mirrored; the migration tail carries forward state. A durable-decision **ADR** records the teardown (destructive-change rule). No production runtime wiring touched. | Migration applies clean on a fresh bootstrap (prerequisites→baseline→reference-data→migrations) — i.e. baseline creates the cluster, the new migration drops it; post-apply `metaldocs.documents` / `metaldocs.template_audit_log` / all 8 satellites are **absent**; bare `documents` resolves to `public.documents` under `search_path=metaldocs,public`; migration is idempotent (re-run = no-op). |
| F4b.3 | `f4b.3-family-b-seed-fix` | Repair the stale tripwire-guarded test seeds (Family B, SQLSTATE P0001 `ErrCapabilityNotAsserted`): `documents/approval/jobs` (`scheduled_publish_job_test.go`), `documents/approval/repository` (`postgres_approval_repository_test.go`), `iam/authz` (`authz_bypass_test.go`), `documents/repository` (`repository_revision_history_integration_test.go`) — set `metaldocs.asserted_caps` (or use the established bypass GUC) before writing `controlled_documents` / `iam_user_roles`. Test-only; no production change. | The named tests pass from a clean baseline under the operator DSN; the tripwire still fires for genuinely-unasserted writes (no weakening of `enforce_capability_asserted`). |
| F4b.4 | `f4b.4-integration-suite-green` | Prove the teardown re-greens the M4 blocker and introduces no regression, **without any harness change**. | F4.1a Gate #5 (`TestCreateDocumentTx_PopulatesAllSnapshotColumns`) passes deterministically with the operator DSN **including** `search_path=metaldocs,public`; `git diff tests/integration/testdb/db.go` is **empty** (db.go at HEAD — fix-not-adapt proof); full integration suite green from clean baseline under the operator DSN (captured output); Family A (42703) and Family B (P0001) classes both at 0. |

For each feature, "what to validate" is objectively checkable — a migration that applies, a schema
state observed, a named test that passes, an empty diff. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For this
milestone the gate enforces:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate"; each
   feature's `spec.md` consumer contract is honored.
2. **Workflow-class QA** — `wiki/quality/backend-api-qa-checklist.md` + the database rules in
   `wiki/database/` (migration policy, baseline-mirror, reference data) + `release-closeout-checklist.md`.
3. **Regression** — M0–M3 gates still pass; M4's built features (F4.1–F4.6) are unaffected (the drop
   touches only dead objects); no production route/contract change.
4. **Root-cause check (the bar)** — re-measured: bare `documents` resolves to `public.documents` under
   metaldocs-first `search_path` **with `db.go` unchanged from HEAD**. An empty `db.go` diff is the
   structural proof the defect was **fixed at the schema, not adapted at the harness**. A green achieved
   by editing the harness is a **FAIL**.
5. **No unplanned scope** — only verified-dead objects dropped; anything beyond F4b.1's manifest is a
   FAIL. Test-only changes in F4b.3 stay test-only.

## Dependencies & constraints

- Depends on: M4 features F4.1–F4.6 already built (this milestone unblocks M4's *close gate*, not its
  features). Routed through the **`metaldocs-database`** skill (migration + curated-baseline mirror +
  dictionary/reference-data rules).
- Constraints: **forward-only** migrations, mirrored into the curated baseline (no editing past
  migrations); drops restricted to F4b.1's verified-dead manifest; **no production runtime/contract
  change**; reads-stay-live and advisory-lock hazard rules are not in play (no read-path edits).
- Blocks: the **M4 close gate** (re-dispatch the M4 `milestone-validator` after 4b PASS + HS-1) and,
  transitively, **M5** (the terminal re-audit).

## Applicable hard-stops

- **HS-1** — 4b close gate: operator review; no M4-validator re-dispatch and no merge without approval.
- **HS-2** — if F4b.1 finds **any live dependent** of the cluster (inbound FK from a live table, a
  view/trigger/RLS policy, a reference-data seed, or an OpenAPI/contract anchor), the cluster is **not
  dead**: stop, report the boundary + minimum prerequisite plan, do **not** drop.
- **HS-4** — 4b `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle,
  re-dispatch the validator.
- **HS-6** — any scope drift beyond the verified-dead manifest or beyond test-only Family-B seed
  fixes → stop, surface, replan before continuing.
