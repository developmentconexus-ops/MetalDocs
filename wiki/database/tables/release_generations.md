# public.release_generations

**Schema:** `public.release_generations`
**Owner:** approval module (ADR 0085 — release generation identity/hold state machine)
**Last verified:** 2026-08-07

## Purpose

The release-generation ledger for the eQMS release-hold state machine (ADR
0085). Each row is one identity-pinned attempt to release a specific document
revision at a specific approval instance: `(tenant_id, subject_kind,
document_id, approval_instance_id, revision_id, revision_version,
frozen_content_hash)` is the full identity (`ux_release_generations_identity`
unique constraint), and `generation_seq` (a dedicated sequence) gives the
total order needed to answer "is G still D's freeze head" (`isFreezeHeadTx` —
a newer generation exists exactly when the document was returned to draft and
re-approved). A generation accumulates two independent facts over time —
the **approval fact** (`approval_fact_at`, `final_approver_id`,
`submitted_by`) and the **artifact fact** (`artifact_fact_at`,
`final_docx_s3_key`, `final_pdf_s3_key`) — via an `INSERT ... ON CONFLICT ...
DO UPDATE` upsert that only ever fills a fact in once (`COALESCE(existing,
new)`), never overwrites one already recorded. `released_at` is set only once
both facts are present (`ck_release_generations_released_implies_facts`).
`hold_reason` records why a generation is not yet released
(`awaiting_approval_fact`, `materializing`, `awaiting_effective_date`,
`supersede_conflict`, `plan_invalid`, `failed`).

## Columns

| Column                  | Type          | Notes |
|-------------------------|---------------|-------|
| `id`                    | `uuid` PK     | Default `gen_random_uuid()`. |
| `generation_seq`        | `bigint`      | Not null. Backed by dedicated sequence `release_generations_generation_seq_seq`; gives the freeze-head total order. |
| `tenant_id`             | `uuid`        | Not null. No declared FK to `metaldocs.tenants`. |
| `subject_kind`          | `text`        | Default `'document'`, not null. CHECK `subject_kind = 'document'` (only value supported today). |
| `document_id`           | `uuid`        | Not null. |
| `approval_instance_id`  | `uuid`        | Not null. Part of the identity constraint. |
| `revision_id`           | `uuid`        | Not null. Part of the identity constraint. |
| `revision_version`      | `integer`     | Not null. CHECK `>= 0`. Part of the identity constraint. |
| `frozen_content_hash`   | `text`        | Not null. CHECK matches `^[0-9a-f]{64}$` (sha256 hex). Part of the identity constraint. |
| `approval_fact_at`      | `timestamptz` | Nullable. Set once, never overwritten (COALESCE upsert). |
| `final_approver_id`     | `text`        | Nullable. Set with the approval fact. |
| `submitted_by`          | `text`        | Nullable. Set with the approval fact. |
| `final_docx_s3_key`     | `text`        | Nullable. Set with the artifact fact. |
| `final_pdf_s3_key`      | `text`        | Nullable. Set with the artifact fact. |
| `artifact_fact_at`      | `timestamptz` | Nullable. CHECK: non-null implies both S3 keys non-null (`ck_release_generations_artifact_fact_full_set`). |
| `hold_reason`           | `text`        | Nullable. CHECK is one of the six defined reasons or NULL. |
| `hold_detail`           | `text`        | Nullable. Free-text detail for `hold_reason`. |
| `released_at`           | `timestamptz` | Nullable. CHECK: non-null implies both `approval_fact_at` and `artifact_fact_at` are non-null. |
| `last_evaluated_at`     | `timestamptz` | Nullable. Drives the `ix_release_generations_open` partial index for the release-hold reconciler job. |
| `created_at`            | `timestamptz` | Default `now()`, not null. |
| `updated_at`            | `timestamptz` | Default `now()`, not null. |

## Migrations

Table, its CHECK constraints, and its identity/sequence infrastructure are
present in `db/baseline/0001_current_schema.sql` (folded baseline); no
post-baseline migration alters it as of 2026-08-07. Governed by ADR 0085
(release generation identity) referenced throughout `release_facts.go`,
`release.go`, and `release_hold_port.go`.

## Key callers

- `internal/modules/approval/application/release_facts.go::recordApprovalFact` (helper, unexported) — `INSERT ... ON CONFLICT ON CONSTRAINT ux_release_generations_identity DO UPDATE` upsert of the approval fact, in-tx.
- `internal/modules/approval/application/release_facts.go::LoadReleaseGenerationByIDTx` — `SELECT ... FOR UPDATE` by `id` + `tenant_id`, used by the artifact-fact producer before its own update.
- `internal/modules/approval/application/release_facts.go::loadReleaseGenerationByKeyTx` (unexported) — `SELECT ... FOR UPDATE` by the full ADR 0085 identity tuple.
- `internal/modules/approval/application/release_facts.go::isFreezeHeadTx` (unexported) — freeze-head predicate via `generation_seq` ordering.
- `internal/modules/approval/application/release_coordinator.go` — orchestrates the release-hold coordinator, reads/updates generation rows.
- `internal/modules/approval/infrastructure/release_hold_reader.go` — read-only queries backing the release-hold alert surface, all aliasing this table as `g`.
- `internal/modules/jobs/release_hold_reconciler/job.go` — periodic reconciler job; calls only application-layer functions, no raw SQL of its own against this table.
- `internal/platform/worker/materialize_job_runner.go`, `internal/platform/worker/pdf_job_runner.go` — workers that produce the artifact fact; they lock `release_generations` via the recorder before writing `documents`.
- `internal/modules/approval/infrastructure/tenant_data_port.go::TenantDataPort` — tenant export/erase port, keyed on `tenant_id`.

## Tenant scoping

`tenant_id` (uuid, not null) carries tenant scoping with no declared FK to
`metaldocs.tenants(id)`. Every generation read/write in `release_facts.go`
carries an explicit `tenant_id` predicate (identity tuple, `FOR UPDATE`
lookups). RLS `tenant_isolation` (`FORCE ROW LEVEL SECURITY`) backstops the
app-level predicate — a cross-tenant lookup by `id` alone would still be
scoped correctly because callers always pass `tenant_id` alongside `id`
(`LoadReleaseGenerationByIDTx`); RLS additionally prevents a caller that
omitted the predicate from seeing another tenant's row.
