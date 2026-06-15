# Feature F4b.2 — drop migration for the legacy MDDM document cluster

> **Milestone:** 4b (Legacy schema cluster teardown)  ·  **Feature:** `f4b.2-drop-migration`
> **Approved:** 2026-06-15 (operator standing authorization; consumer = F4b.4 integration-suite-green; producer = the migration).

## Consumer contract

The **consumer** is F4b.4 (and, transitively, the M4 close gate): it needs a database state in which
the legacy `metaldocs.documents` / `metaldocs.template_audit_log` cluster is **gone**, so that bare
unqualified `documents` resolves to `public.documents` under a `metaldocs`-first `search_path`. The
contract on this feature's producer (a `metaldocs-database` forward migration):

1. **Drops exactly F4b.1's verified-dead manifest** — the 10 objects, no more, no less:
   `document_version_images`, `document_attachments`, `document_collaboration_presence`,
   `document_edit_locks`, `document_template_assignments`, `document_versions`, `document_versions_mddm`,
   `workflow_approvals`, `metaldocs.documents` (anchor), `metaldocs.template_audit_log`.
2. **Forward-only, idempotent.** New migration number `> 0239` (= `0240`); `DROP TABLE IF EXISTS ...
   CASCADE`; satellites before anchor; one `public.schema_migrations` row with
   `ON CONFLICT DO NOTHING`; re-run is a no-op.
3. **Baseline left untouched.** `db/baseline/0001_current_schema.sql` is a frozen snapshot (never
   re-mirrored — objects dropped by 0236 still live there). The baseline creates the cluster; `0240`
   drops it on the migration pass.
4. **Durable decision recorded** — ADR `0032-drop-legacy-mddm-document-cluster.md` (destructive-change
   rule: data-loss caveat + rollback note + maintenance-window posture).
5. **No production runtime / route / contract change.**

## Non-goals

- **Not** running the migration against production (HS-1 — operator-gated; this feature authors + tests
  against a fresh testdb bootstrap only).
- **Not** touching `tests/integration/testdb/db.go` or the operator DSN (the empty `db.go` diff is 4b's
  fix-not-adapt proof — owned by F4b.4).
- **Not** the Family-B tripwire seed repair (that is F4b.3).
- **Not** dropping or altering the live `public.documents` / `public.template_audit_log` tables.
- **Not** re-mirroring the curated baseline.

## Validation Gate

| # | Acceptance | Proof |
|---|-----------|-------|
| A | Migration applies clean on a fresh bootstrap (prerequisites→baseline→reference-data→migrations) | bootstrap a fresh template DB; migration log shows `0240` applied, no error |
| B | Post-apply: all 10 cluster objects absent from `metaldocs` schema | `\dt metaldocs.*` / catalog query → none of the 10 present |
| C | Post-apply: live `public.documents` + `public.template_audit_log` still present and intact | catalog query → both present with their columns |
| D | Bare `documents` resolves to `public.documents` under `search_path=metaldocs,public` | `SET search_path TO metaldocs,public; SELECT to_regclass('documents')` → `public.documents` |
| E | Idempotent: re-running `0240` is a no-op (no error, no second `schema_migrations` row) | re-apply → success; `schema_migrations` has one `0240` row |
| F | Exactly the manifest dropped — no kept object lost (CASCADE blast contained) | post-apply catalog diff vs pre-apply shows only the 10 removed |

## Interview record

No operator interview needed — the consumer contract is read directly from F4b.4's need (cluster
absent) and F4b.1's verified-dead manifest. Operator decision (2026-06-15) selected the root-cause path
("New milestone: drop legacy cluster"). Destructive-change handling (ADR + data-loss/rollback/window)
follows `wiki/database/migration-policy.md` and the 0236/0231 precedent.
