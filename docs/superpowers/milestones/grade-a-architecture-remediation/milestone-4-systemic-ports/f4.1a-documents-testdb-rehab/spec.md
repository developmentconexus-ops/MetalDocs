# Feature F4.1a — Documents integration testdb rehab (fix-feature)

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.1a-documents-testdb-rehab`
> **Type:** fix-feature (pulled from spawned `task_2d29ac7d`), enabling F4.1's deferred Gate #5.
> **Approved:** operator chose "fix now" (option 2) 2026-06-15 — build a shared tripwire-aware testdb
> seed helper, then prove F4.1 Gate #5 green with a **non-Noop** reader value assertion, then F4.2.
> **Scope class:** test-harness only. **No production code, no schema, no migration changes.**

## Why this feature exists

F4.1 closed with one bounded defer: the live documents create → `created_by_display_name_snapshot`
read-back (`TestCreateDocumentTx_PopulatesAllSnapshotColumns`) was RED, so the F4.1 port's value was
never proven end-to-end through a real `UserDisplayNameReader` against a live create. The original
defer hypothesized "missing dev-seeds for governed FK parents." Investigation found that hypothesis
**partly wrong** — the real root cause is a cluster of **pre-existing test-harness defects** (production
code is correct):

1. **Wrong `search_path` ordering.** The documents integration tests set
   `SET search_path TO metaldocs, public`. The baseline ships **two** `documents` tables: the real
   runtime table `public.documents` (has `controlled_document_id` + all snapshot columns) and a **dead
   legacy** `metaldocs.documents` (no such columns). With `metaldocs` first, every bare `documents`
   reference resolved to the dead table → `column "controlled_document_id" does not exist` and silent
   no-op UPDATEs. The create path's production SQL uses bare names and correctly relies on
   `public.documents`; the tests were mis-pathed.
2. **Missing app-level `system_admin` seed.** The create path runs `authz.Require(document.create)`
   (ADR 0022 tier-2). The test actor had no role/area, so every create was denied. The legitimate seed
   is a tenant-scoped `system_admin` (tier-2 inheritance bypass) — seeding **required membership**, not
   symptom-patching authz.
3. **`SetMaxOpenConns(1)` pool deadlock.** Gate #5's create tx holds the only connection while the
   off-tx `UserDisplayNameReader` read **and** the templates `GetPublishedVersion` read (both on the
   pool, H-PRE-1) wait for a connection → 600 s hang.
4. **Session-level cap assertions in shared seed helpers** were not pool-safe once the pool has >1
   connection.

## Consumer contract (what the helpers must provide)

The consumers are the documents integration tests. Required testdb helpers in
`tests/integration/testdb/fixtures.go`:

- `SeedGovernedTaxonomy(t, db, tenantID, profileCode, processAreaCode)` — seed the governed FK-parent
  chain (`document_families` / `document_process_areas` / `document_profiles`) **pool-safely**, with
  the `taxonomy.manage` tripwire asserted **transaction-locally** (never leaking into the caller's
  session). Idempotent.
- `SeedSystemAdmin(t, db, tenantID, userID, displayName)` — seed the tenant parent (`metaldocs.tenants`),
  upsert `iam_users` (with `display_name` for the port to resolve), and grant tenant-scoped
  `system_admin` in `iam_user_roles` (under a `user.manage` tx-local assertion). Satisfies the
  create-path `authz.Require` via the tier-2 system_admin bypass without per-area grants.
- `SupersedeActiveDocumentForCD(t, db, controlledDocumentID)` — walk the single active document for a CD
  through the **legal** status lifecycle (`draft → under_review → approved → published → superseded`,
  stubbing the six snapshot columns the `enforce_snapshot_on_submit` trigger requires) so a successor
  revision can be created. Mirrors production: a prior revision is published then superseded before its
  successor exists. (`trg_documents_legal_transition` permits no shortcut out of `draft`.)
- `seedWithCaps(t, db, capsJSON, fn)` — internal: run `fn` in its own tx with `metaldocs.asserted_caps`
  asserted tx-locally (mirrors production `authz.appendAssertedCap`), pool-safe.

## Non-goals

- **No production code, schema, migration, or `Qualified`/`metaldocsOwnedObjects` map change.** The
  harness map disagreements (e.g. `documents` collision, missing `tenants` entry) are worked around
  locally (explicit `metaldocs.tenants`), **not** "corrected" — reconciling the map is out of scope and
  would be an HS-6 expansion.
- **The dead `metaldocs.documents` legacy table is left in place** (noticed, not deleted — CLAUDE.md
  §5.3).
- **Not** rehabbing the rest of the documents integration suite (`fillin`, `commit_upload`,
  `snapshot_repository`, `revision_history`) — those carry the **same pre-existing harness-rot class**
  but are a materially larger effort (per-target-table investigation, not mechanical). Recorded as a
  bounded defer in `evidence.md` (HS-6: surfaced to operator, not silently absorbed).

## Validation Gate

| # | Acceptance | Named test / command | Proof |
|---|------------|----------------------|-------|
| 1 | F4.1 Gate #5 green with a **real** reader value assertion | `TestCreateDocumentTx_PopulatesAllSnapshotColumns` under `iampg.NewUserDisplayNameRepository` asserting `created_by_display_name_snapshot == "Snapshot Author"` | `go test -tags integration -count=1 -p 2 ./internal/modules/documents/application/` PASS |
| 2 | `repository_create` suite green (4 tests / 5 cases) on the real `public.documents` | `TestCreateDocumentTx_StorageKeyInvariant` (Non/Empty), `…_RevisionNumberIncrementsForSameCD`, `…_RejectsEmptyName`, `TestGetDocument_ReturnsSnapshotMetadata` | `go test -tags integration -count=1 -p 2 -run '<those>' ./internal/modules/documents/repository/` PASS |
| 3 | No production change | `git diff --stat` shows only `tests/integration/testdb/fixtures.go` + the two touched integration test files + milestone docs | diff review |
| 4 | Static clean | `go vet -tags integration ./internal/modules/documents/... ./tests/integration/testdb/` | VET OK |

## Interview record

No operator interview required beyond the option-2 "fix now" decision (2026-06-15). Root cause was
established by live-DB probe + baseline-source inspection (the two `documents` tables, the trigger
graph, the `ux_documents_cd_active` index).
