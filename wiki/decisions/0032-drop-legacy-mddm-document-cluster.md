# ADR 0032 — Drop the dead legacy `metaldocs.documents` (MDDM) document cluster

> **Status:** Accepted 2026-06-15
> **Last verified:** 2026-06-15
> **Scope:** Removal of the early editor-era *MetalDocs Document Model* (MDDM) tables that survive,
> unused, in the `metaldocs` schema and **shadow** the live `public.documents` governance model under a
> `metaldocs`-first `search_path`. The destructive forward-only migration, the satellites-before-anchor
> drop order, the data-loss caveat, and the maintenance-window / rollback posture.
> **Out of scope:** the live `public.documents` + `public.controlled_documents` governance model
> (untouched); the **MDDM export feature flag** (`MDDMNativeExportRolloutPercent` — alive, not a table);
> the test-harness `search_path` (deliberately **not** changed — see milestone 4b fix-not-adapt proof);
> the Family-B tripwire seed repair (separate feature F4b.3).
> **Key files:**
> - `db/migrations/0240_drop_legacy_mddm_document_cluster.sql` — the destructive forward migration
> - `db/baseline/0001_current_schema.sql` — frozen snapshot; **left untouched** (migration tail carries forward state)
> - `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4b-legacy-schema-teardown/f4b.1-legacy-cluster-census/evidence.md` — the verified-dead manifest authorizing this drop

## Context

Two tables are duplicated across schemas: `documents` and `template_audit_log`. The `metaldocs.*`
copies are dead, early editor-era MDDM tables; the live content lives in `public.*`. The runtime issues
**bare, unqualified** `documents` SQL in 40+ sites, so bare-name resolution is governed entirely by the
connection `search_path` (nothing in `db/` sets one; production connects on the Postgres default →
effective `public`). When the connection `search_path` is `metaldocs`-first (the operator/dev test DSN
carries `search_path=metaldocs,public`), bare `documents` resolves to the **dead** `metaldocs.documents`
— which lacks `tenant_id` / `active_session_id` / `controlled_document_id` — producing
`column "tenant_id" ... does not exist` (SQLSTATE 42703). Harmless in production (public-first), fatal to
the test harness.

A `/systematic-debugging` pass (recorded in
`milestone-4-systemic-ports/f4.1b-testdb-search-path-robustness/evidence.md`) proved this is a **schema
defect, not a harness bug**. Adapting the harness `search_path` was rejected as a symptom patch
(CLAUDE.md hard-stop: no symptom-patching). The root-cause fix is to remove the dead duplicate so bare
`documents` resolves to `public.documents` under **any** `search_path` ordering.

F4b.1's verify-dead census proved the cluster is dead across every surface: **zero runtime Go
references** (only one comment calling the schema "decommissioned"), every inbound FK originates
**inside** the cluster, and no view / trigger / RLS policy / stored function / sequence / reference-data
seed references any cluster object. HS-2 not tripped.

MDDM disambiguation: the legacy MDDM **tables** are dead; the MDDM **export format** survives only as the
`MDDMNativeExportRolloutPercent` feature flag (client-side DOCX export rollout) — no flag code touches
these tables. The dead `canonical_mddm_snapshot` column rides on the dead `metaldocs.document_versions`
and drops with it.

## Decision

Drop the entire dead legacy MDDM document cluster — **10 objects**, satellites before anchor — in a
single forward-only, idempotent migration (`0240`), `DROP TABLE IF EXISTS ... CASCADE`, with one
`public.schema_migrations` row:

1. `metaldocs.document_version_images` (2nd-level satellite — FK → `document_versions_mddm`; leaf)
2. `metaldocs.document_attachments`
3. `metaldocs.document_collaboration_presence`
4. `metaldocs.document_edit_locks`
5. `metaldocs.document_template_assignments`
6. `metaldocs.document_versions`
7. `metaldocs.document_versions_mddm`
8. `metaldocs.workflow_approvals`
9. `metaldocs.documents` — **anchor**
10. `metaldocs.template_audit_log` — independent dead duplicate of `public.template_audit_log`

Indexes, FK constraints, and sequences attached to these tables drop with them (CASCADE-safe — every
inbound FK originates inside the manifest; no kept table references the cluster).

The curated baseline (`db/baseline/0001_current_schema.sql`) is a **frozen snapshot** and is **left
untouched** — per repo evidence (objects dropped by 0236 still live in the baseline) the baseline is
never re-mirrored; the migration tail carries forward state. On a fresh bootstrap the baseline creates
the cluster and `0240` drops it.

No production runtime, route, or contract wiring changes.

## Consequences

- Bare unqualified `documents` / `template_audit_log` resolves to the real `public.*` table under **any**
  `search_path` ordering — the shadow landmine is permanently removed.
- The M4 close-gate blocker (F4.1a Gate #5) re-greens with the operator DSN unchanged **and** the
  harness (`tests/integration/testdb/db.go`) unchanged from HEAD — the empty `db.go` diff is the
  structural fix-not-adapt proof.
- **Destructive change (data loss).** Any historical rows still in these `metaldocs.*` tables are
  abandoned by this migration. They are believed empty/dead in production (zero runtime writers since the
  schema was decommissioned), but the operator **MUST dump these tables before applying `0240` to
  production** if any forensic/historical value is suspected. Apply within a **maintenance window**;
  gated at HS-1 (operator runs against prod — author + test only here).
- **Rollback:** forward-only repo policy. To restore, re-create the tables from the baseline's
  `CREATE TABLE metaldocs.*` definitions (lines noted in F4b.1) plus a restore of any pre-drop dump.
  No automated down-migration is provided (consistent with `0236`/`0231` precedent).
- Migration is idempotent (`IF EXISTS`): re-run is a no-op; `schema_migrations` insert is
  `ON CONFLICT DO NOTHING`.

## References
- Milestone 4b — `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4b-legacy-schema-teardown/milestone.md`
- Verified-dead manifest — `.../milestone-4b-legacy-schema-teardown/f4b.1-legacy-cluster-census/evidence.md`
- Root-cause investigation — `.../milestone-4-systemic-ports/f4.1b-testdb-search-path-robustness/evidence.md`
- Precedent dead-schema drops — `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql`, `db/migrations/0236_dead_schema_drop.sql`
- Migration policy — `wiki/database/migration-policy.md`; `db/migrations/README.md`
- Governing spec — `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
