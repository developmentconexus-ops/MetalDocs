# Feature F4b.2 — evidence (drop migration)

> **Outcome:** PASS. Forward-only migration `0240` drops the 10-object verified-dead MDDM cluster on a
> fresh bootstrap; live `public.*` tables intact; bare `documents` → `public.documents` under
> `search_path=metaldocs,public`; idempotent. Baseline untouched. No production run (HS-1).

## Artifacts

- ADR: `wiki/decisions/0032-drop-legacy-mddm-document-cluster.md`
- Migration: `db/migrations/0240_drop_legacy_mddm_document_cluster.sql`
- Baseline `db/baseline/0001_current_schema.sql` — **not modified** (`git diff` empty; frozen snapshot).

## Verification harness

Fresh scratch DB in the dev Postgres container (`metaldocs-postgres`, :5433), applied in the exact order
`tests/integration/testdb.ApplyCuratedBootstrap` uses: `db/prerequisites/0001_extensions.sql` →
`db/baseline/0001_current_schema.sql` → `db/reference-data/0001_product_reference_data.sql` →
`db/migrations/*.sql` sorted (lexical; 4-digit zero-padded = apply order). `psql -v ON_ERROR_STOP=1`.

## Gate results

| # | Acceptance | Result |
|---|-----------|--------|
| A | Migration applies clean on fresh bootstrap | **PASS** — prereq/baseline/refdata OK; `ALL MIGRATIONS APPLIED OK (incl 0240)` with `ON_ERROR_STOP=1` |
| — | Pre-migration sanity: cluster present | baseline creates all **10** cluster tables (count = 10 before migrations) |
| B | Post-apply: all 10 cluster objects absent | **PASS** — `metaldocs` cluster count = **0** |
| C | Live `public.documents` + `public.template_audit_log` present & intact | **PASS** — both present; `public.documents` has `active_session_id,controlled_document_id,tenant_id` |
| D | Bare `documents` → `public.documents` under `search_path=metaldocs,public` | **PASS** — `'documents'::regclass` → `public.documents`; `'template_audit_log'::regclass` → `public.template_audit_log` |
| E | Idempotent re-run = no-op, one `schema_migrations` row | **PASS** — re-run: `table "documents" does not exist, skipping` (no error); `count(version='0240')` = **1** |
| F | Only the manifest dropped; kept objects intact | **PASS** — `metaldocs` schema retains **33** tables post-apply (cluster removed, nothing else) |

## Key transcript (real output)

```
=== PRE-migration cluster count (expect 10) ===
10
=== apply ALL migrations in sorted order ===
ALL MIGRATIONS APPLIED OK (incl 0240)
=== B: cluster absent (expect 0) ===
0
=== C: live public tables present ===
public.documents
public.template_audit_log
=== C2: public.documents tenant cols ===
active_session_id,controlled_document_id,tenant_id
=== D explicit: namespace of bare names under metaldocs-first ===
public.documents
public.template_audit_log
=== E: idempotent re-run ===
NOTICE:  table "documents" does not exist, skipping
NOTICE:  table "template_audit_log" does not exist, skipping
=== E2: count(0240 rows) ===
1
=== F: metaldocs tables remaining ===
33
```

## Notes / defers

- **Destructive / production:** `0240` drops production data. Per ADR 0032 + HS-1, the operator applies
  it against prod within a maintenance window (dump first if forensic value suspected). This feature
  only authored + tested against a throwaway scratch DB (dropped after).
- F4.1a Gate #5 re-green under the operator DSN with `db.go` unchanged is proven by **F4b.4** (this
  feature establishes the schema state it depends on).
- Family-B tripwire seed repair is **F4b.3** (separate root cause).
