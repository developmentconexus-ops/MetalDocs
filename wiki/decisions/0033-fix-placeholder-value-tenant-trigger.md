# ADR 0033 — Fix the `document_placeholder_values` tenant-consistency trigger

> **Status:** Accepted 2026-06-15
> **Last verified:** 2026-06-15
> **Scope:** Correct the `BEFORE INSERT OR UPDATE` trigger function
> `public.enforce_placeholder_value_tenant_consistent()` so it resolves the row's owning tenant via
> `document_revisions → documents` instead of looking the `revision_id` up directly against `documents.id`.
> Forward-only `CREATE OR REPLACE` migration; trigger binding unchanged.
> **Out of scope:** the placeholder fill-in repository code (already correct — it writes a real
> `documents.current_revision_id`); the FK target (already `document_revisions`, set by archived 0191); the
> frozen baseline snapshot (left untouched — the migration tail carries forward state, per ADR 0032 precedent).
> **Key files:**
> - `db/migrations/0241_fix_placeholder_value_tenant_trigger.sql` — the forward fix
> - `db/baseline/0001_current_schema.sql:615-630` — frozen snapshot of the **buggy** function; **left untouched**
> - `internal/modules/documents/repository/fillin_repository.go` — `FillInRepository.UpsertValue` (consumer)
> - `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4c-test-fixture-framework/f4c.2-migrate-blocker-files/evidence.md` — the HS-2 escalation that surfaced this

## Context

`public.document_placeholder_values.revision_id` is a **`document_revisions.id`**: the FK
`document_placeholder_values_revision_id_fkey` references `document_revisions` (re-pointed there by archived
migration `0191_document_placeholder_values_revision_fk.sql`), and `FillInRepository.UpsertValue` writes
`documents.current_revision_id`.

The tenant-consistency trigger function (added in archived `0153_placeholder_values_tenant_consistency.sql`)
resolves the tenant the wrong way:

```sql
SELECT tenant_id INTO doc_tenant FROM documents WHERE id = NEW.revision_id;   -- revision id ≠ documents.id
IF doc_tenant IS NULL THEN
    RAISE EXCEPTION 'document % not found' USING ERRCODE = 'foreign_key_violation';
END IF;
```

A revision id never equals a `documents.id`, so `doc_tenant` is always NULL and **every** placeholder fill-in
insert/update raises `document % not found` (SQLSTATE 23503). The 0153 trigger was never updated when 0191
re-pointed the FK at `document_revisions`. Verified **byte-identical** in `db/baseline/0001_current_schema.sql`
and the live dev DB (`docker exec metaldocs-postgres psql … pg_get_functiondef`). The function is currently
attached **only** to `public.document_placeholder_values` (the 0153 binding on `document_editable_zone_content`
no longer exists — runtime truth wins over migration archaeology).

Surfaced as an **HS-2** escalation during milestone-4c feature F4c.2 (migrating the placeholder fill-in
integration test onto the unified `testdb` factory): the test failed on the curated baseline not because of a
seed gap but because of this production-schema defect. Per the F4c.2 spec's pre-committed non-goal — "No
production-source change except an HS-2-escalated, operator-approved schema fix" — the operator approved the
prerequisite fix before any further test migration.

## Decision

Replace the function body to walk the real relationship:

```sql
SELECT d.tenant_id INTO doc_tenant
  FROM document_revisions r JOIN documents d ON d.id = r.document_id
 WHERE r.id = NEW.revision_id;
IF doc_tenant IS NULL THEN
    RAISE EXCEPTION 'revision % not found' USING ERRCODE = 'foreign_key_violation';
END IF;
IF doc_tenant <> NEW.tenant_id THEN
    RAISE EXCEPTION 'tenant mismatch: document=% value=%' USING ERRCODE = 'check_violation';
END IF;
```

Delivered as forward-only, idempotent `CREATE OR REPLACE` (no trigger DDL — the binding is unchanged). The
frozen baseline keeps the buggy snapshot; the migration tail carries the corrected state, matching the
ADR-0032 baseline-frozen convention.

## Consequences

- **Restores a real security control.** The cross-tenant consistency check (reject a placeholder value whose
  `tenant_id` differs from its revision's owning document) now actually executes. Before the fix it failed
  **closed** — raised on every write — so there was no cross-tenant leak, but the fill-in write path was dead
  on any environment built from the canonical migrations.
- **Unblocks placeholder fill-in.** `UpsertValue` / `ListValues` now succeed; F4c.2 migrates the integration
  test onto the factory with this as a prerequisite.
- **Rollback:** re-applying the old body restores the (broken) prior state; no data migration, no destructive
  change. The trigger is `BEFORE` and side-effect-free, so the swap is safe online.
- **Error-message change:** the not-found path now says `revision % not found` (was `document % not found`).
  No code matches on that text (the SQLSTATE `23503` is unchanged).
