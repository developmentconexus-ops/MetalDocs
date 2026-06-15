# Feature F4b.2 — plan (the "how")

Producer = one forward-only migration; no production run (HS-1).

1. **ADR** `wiki/decisions/0032-drop-legacy-mddm-document-cluster.md` — record the durable decision
   (destructive drop, data-loss caveat, rollback note, maintenance-window posture). → done.
2. **Migration** `db/migrations/0240_drop_legacy_mddm_document_cluster.sql`:
   - `BEGIN; ... COMMIT;`
   - `DROP TABLE IF EXISTS metaldocs.<obj> CASCADE` for the 10 manifest objects, satellites before
     anchor (`document_version_images` first → … → `documents` → `template_audit_log`).
   - one `INSERT INTO public.schema_migrations (version, description) VALUES ('0240', …) ON CONFLICT DO NOTHING`.
   - Baseline NOT edited (frozen-snapshot policy).
3. **Verify** against a fresh bootstrap that mirrors `testdb.ApplyCuratedBootstrap`
   (prerequisites → baseline → reference-data → migrations sorted) on a scratch DB in the dev Postgres
   container, then assert gates A–F. Drop the scratch DB after.
4. **evidence.md** — record the commands + real output.

No code/runtime/contract change; no harness change (that empty `db.go` diff belongs to F4b.4).
