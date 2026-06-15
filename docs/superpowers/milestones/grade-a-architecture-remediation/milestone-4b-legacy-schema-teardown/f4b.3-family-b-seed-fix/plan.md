# Feature F4b.3 — plan (the "how")

1. Read the tripwire CASE mapping (`db/baseline/0001_current_schema.sql:426-486`) → cap per guarded
   table. Read the canonical assert idiom (`create_document_snapshot_integration_test.go:156-173`).
2. RED: reproduce P0001 on a representative target (approval repo test) under the operator DSN.
3. For each target seed site, assert the required cap(s) **in the same tx/pinned connection** as the
   guarded writes, mirroring the canonical idiom:
   - approval repo (5 seed-stmt loops): prepend a `set_config('metaldocs.asserted_caps', …, true)` stmt
     as the first slice element of each loop (per-loop caps: +`document.submit` where the tx writes
     `approval_instances`; +`document.edit` where it later UPDATEs `documents`).
   - scheduled publish job: prepend the same assert stmt to its seed loop.
   - authz_bypass: extend the existing tx-local `set_config` call to also assert `user.manage` before the
     `iam_user_roles` insert.
   - revision_history (2nd test): add the missing `controlled_documents.create` to its session-scoped
     asserted set (test pins `SetMaxOpenConns(1)` + `is_local=false`).
4. GREEN: re-run the four packages under the operator DSN; vet clean; confirm test-only diff.
5. evidence.md — RED + GREEN + diff-stat.

No production / schema / migration / tripwire change.
