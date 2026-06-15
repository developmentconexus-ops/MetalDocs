# Feature F4b.3 — Family-B tripwire seed fix

> **Milestone:** 4b (Legacy schema cluster teardown)  ·  **Feature:** `f4b.3-family-b-seed-fix`
> **Approved:** 2026-06-15 (operator standing authorization; operator decision "fix the Family-B seeds now").

## Consumer contract

The **consumer** is the integration-test layer under the operator DSN. Family B is a **separate root
cause** from the schema shadow (Family A / F4b.2): stale integration-test **seeds** write to
tripwire-guarded tables (`controlled_documents`, `documents`, `approval_instances`, `iam_user_roles`)
**without** asserting the capability the `enforce_capability_asserted()` trigger requires, so the seed
INSERT/UPDATE fails with `ErrCapabilityNotAsserted` (SQLSTATE **P0001**). The seeds predate migration
0231 (which added the tripwire); the tripwire is working **as designed**. Required:

1. **Each guarded seed write asserts its capability** via `set_config('metaldocs.asserted_caps', '[…]')`
   **in the same transaction/connection** as the write, using the cap(s) the trigger's CASE branch maps
   for that table (`controlled_documents`→`controlled_documents.create` on INSERT; `documents`→
   `document.create` INSERT / `document.edit` UPDATE; `approval_instances`→`document.submit`;
   `iam_user_roles`→`user.manage`).
2. **Test-only.** No production code, schema, migration, or the tripwire itself changes.
3. **No weakening of the tripwire.** The trigger still fires (P0001) for genuinely-unasserted writes — the
   fix asserts the real capability, it does not disable or broaden the guard.
4. **Idiom matches the codebase** — the established `seedCreateDocumentSnapshotRows` pattern
   (`set_config(..., is_local)` then the guarded writes on the same tx/pinned connection).

## Non-goals

- **Not** the Family-A schema shadow (that is F4b.2 / the 0240 drop).
- **Not** the harness `tests/integration/testdb/db.go` (stays at HEAD — F4b.4's fix-not-adapt proof).
- **Not** touching tests that already assert correctly.
- **Not** changing the tripwire function, its CASE mapping, or any production authz path.

## Validation Gate

| # | Acceptance | Proof |
|---|-----------|-------|
| A | RED first: each target seed fails with P0001 `ErrCapabilityNotAsserted` pre-fix under operator DSN | captured pre-fix `go test` output |
| B | GREEN: the four target packages pass post-fix under the operator DSN | `go test -tags integration` on the four packages → `ok` |
| C | Test-only diff | `git diff --stat` touches only `*_test.go` files |
| D | Tripwire not weakened | the asserted caps equal the trigger's required caps for each table; no trigger/CASE edit (grep) |
| E | build/vet clean (integration tag) | `go vet -tags integration` on the four packages |

## Target seed sites

| File | Guarded write(s) | Cap(s) asserted |
|------|------------------|-----------------|
| `internal/modules/documents/approval/repository/postgres_approval_repository_test.go` | `controlled_documents`+`documents` (×5 tx), `approval_instances` (×2), `documents` UPDATE (×1) | `controlled_documents.create`,`document.create` (+`document.submit` / +`document.edit` per tx) |
| `internal/modules/documents/approval/jobs/scheduled_publish_job_test.go` | `controlled_documents`+`documents` | `controlled_documents.create`,`document.create` |
| `internal/modules/iam/authz/authz_bypass_test.go` | `iam_user_roles` | `user.manage` |
| `internal/modules/documents/repository/repository_revision_history_integration_test.go` | `controlled_documents` (2nd test was missing the cap) | add `controlled_documents.create` |

## Interview record

No operator interview needed — the operator decided "fix the Family-B seeds now" (2026-06-15). The
required caps are read directly from the tripwire's CASE mapping
(`db/baseline/0001_current_schema.sql:426-486`) and the canonical seed-assert idiom
(`create_document_snapshot_integration_test.go:156-173`).
