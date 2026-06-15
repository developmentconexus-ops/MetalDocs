# Feature F4.1b — evidence (SUPERSEDED, not implemented)

> **Outcome:** SUPERSEDED by milestone **4b** (`milestone-4b-legacy-schema-teardown`). No code shipped
> under F4.1b. The harness-`search_path` adaptation specced in `spec.md` was rejected as a symptom
> patch after `/systematic-debugging`. This file records the investigation that led to the supersession.

## Root-cause investigation (systematic-debugging, 2026-06-15)

**Symptom:** F4.1a Gate #5 `TestCreateDocumentTx_PopulatesAllSnapshotColumns` fails under the operator
DSN (which carries `?search_path=metaldocs,public`):

```
column "tenant_id" of relation "documents" does not exist (SQLSTATE 42703)
```

**Evidence gathered (commands + findings):**

| # | Probe | Finding |
|---|-------|---------|
| 1 | grep duplicate tables in baseline | **Two** tables exist in both schemas: `documents` and `template_audit_log`. `metaldocs.documents` = legacy editor-era table (no `tenant_id`/`active_session_id`/`controlled_document_id`); `public.documents` = real governance table. |
| 2 | grep runtime bare `documents` | 40+ non-test SQL sites use **bare** `documents` (never `public.documents`). Only ref to `metaldocs.documents` is a comment at `internal/modules/search/infrastructure/v2documents/reader.go:25` calling it "the decommissioned metaldocs.documents schema". |
| 3 | how runtime sets `search_path` | **Nothing sets it.** No `ALTER DATABASE/ROLE SET search_path` anywhere in `db/`; `internal/platform/config/postgres.go` never injects it; `.env` has no `DATABASE_URL`/`search_path`; `.env.example` DSN omits `search_path`. ⇒ production connects on Postgres default `"$user",public` → effective **`public`** (role `metaldocs_app` has no own schema). Runtime qualifies `metaldocs.*`, leaves `public.*` bare. |
| 4 | why tests differ | The operator test DSN carries `search_path=metaldocs,public` (metaldocs **first**). pgx `ParseConfig` puts it in `RuntimeParams` → sent as a connection **startup param** that overrides any per-DB `ALTER DATABASE` default. ⇒ bare `documents` resolves to **`metaldocs.documents`** (legacy) → 42703. |
| 5 | neutrality | PRE-EDIT HEAD + operator DSN reproduces the identical Family-A (42703) + Family-B (P0001) failures (background run `beapjnrb8`). F4.1b edits cause none of them — pre-existing. |
| 6 | satellite liveness | `metaldocs.documents` anchors FK satellites `document_attachments`, `document_collaboration_presence`, `document_edit_locks`, `document_template_assignments`, `document_versions`, `document_versions_mddm`, `workflow_approvals` — **all zero runtime Go refs**. `metaldocs.template_audit_log` — zero refs. Dead cluster, not a live subsystem. |

**Root cause:** the dead `metaldocs.documents` duplicate (and `template_audit_log`) **shadows** the real
`public.*` table whenever `search_path` is metaldocs-first. Harmless in production (public-first), fatal
to the test harness (metaldocs-first DSN).

**Family B** is a *separate* root cause: tripwire-guarded test seeds (`controlled_documents`,
`iam_user_roles` — both single-schema, search-path-invariant) write without setting
`metaldocs.asserted_caps`. Tripwire working as designed; seeds are stale.

## Fix-vs-adapt decision

- **Adapt (this feature's spec):** manipulate the harness `search_path` in `db.go`. Rejected — symptom
  patch; also the stashed default `metaldocs, public` is still metaldocs-first so it would not even fix
  bare-`documents` tests generally. CLAUDE.md hard-stop: symptom-patching forbidden.
- **Fix (chosen):** drop the dead duplicate cluster so bare `documents` resolves to `public.documents`
  under any `search_path`. Operator chose this (2026-06-15) → **milestone 4b**.

## Artifacts

- Rejected harness edits parked in `git stash@{0}` ("WIP on main: 30503533"); `tests/integration/testdb/db.go`
  on disk is at **HEAD** (clean). Stash retained until 4b is green, then dropped.
- Resolution: `../../../milestone-4b-legacy-schema-teardown/milestone.md`.
- On 4b PASS + HS-1 → re-dispatch the M4 `milestone-validator`; F4.1a Gate #5 re-greens with the
  operator DSN unchanged and `db.go` unchanged.
