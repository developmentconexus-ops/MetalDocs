# F0.2 — Binding cross-module read census (against ADR-0039)

> **Milestone:** M0 · **Feature:** F0.2 · **Date:** 2026-06-20 · **Tree:** `main` @ working copy (forks `d7d53590`)
> **Rule applied:** ADR-0039 D1/D3 — raw `SELECT`/`JOIN`/subquery/`EXISTS` against **another module's owned
> base table** = violation; published view / owner read-port / own table = compliant.
> **Method:** mechanical owner-map diff (Step A) + per-site source inspection (Step B), over
> `internal/modules/**` non-`_test.go`. This is the skeptic re-run of `../discovery-brief.md`.

## Method (reproducible)

1. **Owner map** — for every `INSERT INTO` / `UPDATE` / `DELETE FROM` target in `internal/modules/**`,
   record the mutating module = the table's owner (where the writes live). Tables, normalized over
   `public.`/`metaldocs.` prefixes.
2. **Read map** — for every `FROM` / `JOIN` / `EXISTS` target, record (file:line, table, reader module).
3. **Cross-module diff** — a read where `reader ≠ owner` and the table is in the owned set = candidate H-G.
4. **Inspect** — read each candidate's SQL + tx context in source; classify per ADR-0039; correct ownership.

Owner map (real tables; grep noise like `for`/`on`/`affects` removed):

| Owner module | Owned base tables (mutated by that module) |
|---|---|
| controlleddocuments (CD) | `controlled_documents`, `controlled_document_area_grants`, `controlled_document_user_grants`, `cd_sequence_counters` |
| documents (DOC) | `documents`, `document_revisions`, `document_comments`, `document_checkpoints`, `document_exports`, `document_placeholder_values`, `editor_sessions`, `autosave_pending_uploads`, `auth_failure_counters` |
| documents/approval (APP) | `approval_instances`, `approval_routes`, `approval_route_stages`, `approval_stage_instances`, `approval_signoffs`, `governance_events` |
| taxonomy (TAX) | `document_process_areas`, `document_profiles`, `document_families` |
| iam | `iam_users`, `iam_user_roles`, `user_process_areas` |
| auth | `auth_identities`, `auth_sessions` |
| audit | `audit_events`, `audit_export_jobs` |
| templates | `templates_template`, `templates_template_version`, `templates_approval_config` |
| jobs | `idempotency_keys`, `job_leases` |

> Ownership corrections vs the re-audit, re-confirmed here: `document_process_areas` + `document_profiles` =
> **taxonomy** (not iam). `auth_failure_counters` = **documents/approval** (the signature rate limiter
> INSERTs at `postgres_auth_failure_rate_limiter.go:64`, DELETEs at `:86`) — the Step-A grep's audit-like
> name is a red herring; it is a documents-owned table, so the limiter's reads of it are **same-module,
> compliant** (a false positive the diff initially raised; dropped).

## Part 1 — Mission-scoped in-scope sites (the "~20", reproduced)

All discovery-brief sites confirmed present at the cited `file:line` (re-grepped this session). **None dropped.**
Verdict column = ADR-0039 classification *today*; "→ exemption" = the compliant target after its milestone ports it.

### Category A — foreign domain-state literals (no SQL; typed-constant fix) → M1

| # | Site | Note | Verdict | Home |
|---|------|------|---------|------|
| A1 | `controlleddocuments/domain/resolution.go:42` | `*Status != "published"` (in-memory `*string`) | not a base-table read (literal coupling) | M1 |
| A2 | `controlleddocuments/domain/resolution.go:55` | `*Status == "obsolete"` | not a base-table read | M1 |
| A3 | `controlleddocuments/domain/resolution.go:58` | `*Status != "published"` | not a base-table read | M1 |

### Category B — clean foreign point-reads → owner read-ports → M2

| # | Site | Reads (owner) | tx | Verdict → target |
|---|------|---------------|----|------------------|
| B1 | `documents/repository/repository.go:1701` | `controlled_documents` (CD) | off-tx | **violation** → CD read-port |
| B2 | `controlleddocuments/infrastructure/repository.go:532` | `document_revisions` (DOC) | off-tx | **violation** → DOC read-port |
| B3 | `controlleddocuments/infrastructure/repository.go:539,545` | `documents` (DOC) | off-tx | **violation** → DOC read-port |
| B4 | `controlleddocuments/infrastructure/repository.go:593` | `approval_instances` (APP) | off-tx | **violation** → APP read-port |
| B5 | `documents/application/document_area.go:37` | `controlled_documents` (CD) | **in-tx** | **violation** → CD read-port (tx-aware) |
| B6 | `documents/approval/application/read_service.go:355` | `controlled_documents` (CD) | **in-tx** | **violation** → CD read-port (tx-aware) |
| B7 | `documents/repository/repository.go:154` | `document_process_areas` (TAX) | **in-tx** | **violation** → TAX area read-port (tx-aware) |
| B8 | `iam/infrastructure/postgres/area_catalog_reader.go:28` | `document_process_areas` (TAX) | off-tx | **violation** → TAX area read-port |

### Category C — authz-visibility membership reads → published view (C-α) → M3

| # | Site | Reads (owner) | tx | Verdict → target |
|---|------|---------------|----|------------------|
| C1 | `controlleddocuments/infrastructure/repository.go:150` | `user_process_areas` (iam) | off-tx | **violation** → iam membership view |
| C2 | `controlleddocuments/infrastructure/repository.go:492` | `user_process_areas` (iam) | off-tx | **violation** → iam membership view |
| C3 | `documents/approval/repository/postgres_approval_repository.go:1136` | `user_process_areas` (iam) | **in-tx (H-PRE-1)** | **violation** → iam membership view (in-tx, non-recording) |

### Category C4 — search inlines CD's whole visibility predicate (CD+iam) → M4

| # | Site | Reads (owner) | Verdict → target |
|---|------|---------------|------------------|
| C4a | `search/infrastructure/v2documents/reader.go:69` | `documents` (DOC) | **violation** → consume search/CD read contract (M4) |
| C4b | `search/infrastructure/v2documents/reader.go:70` | `controlled_documents` (CD) | **violation** → CD visibility projection (M4) |
| C4c | `search/infrastructure/v2documents/reader.go:97` | `controlled_document_area_grants` (CD) | **violation** → CD visibility projection (M4) |
| C4d | `search/infrastructure/v2documents/reader.go:102` | `user_process_areas` (iam) | **violation** → iam membership view (M4) |
| C4e | `search/infrastructure/v2documents/reader.go:111` | `controlled_document_user_grants` (CD) | **violation** → CD visibility projection (M4) |

> `reader.go:69` (`FROM public.documents`) is a foreign read the brief folded into the C4 narrative but did
> not enumerate; recorded here for completeness. It is part of the same M4 redesign.

## Part 2 — NEW sites surfaced by the widen (beyond the named tokens)

The brief deferred "owned-table classes beyond the named H-G tokens" to this census (its coverage statement,
"NOT swept"). The widen found:

### 2a — NEW **document-domain** site (fits the mission's shape)

| # | Site | Reads (owner) | tx | Verdict | Proposed home |
|---|------|---------------|----|---------|---------------|
| N1 | `documents/application/fillin_service.go:225` (`TemplateVersionSchemaReader.LoadFillInSchema`) | `templates_template_version` (templates) — `SELECT tv.placeholder_schema FROM templates_template_version tv JOIN documents d …` | off-tx | **violation** | **M2** (add a templates `placeholder_schema` read-port; ADR 0030 `TemplateVersionPort` precedent) |

N1 is the **same class** as Category B (a clean foreign point-read → owner read-port). It is genuinely new
(not in the brief's "~20"). Routing it to M2 **adds one read-port to M2's shape** → an HS-6 surface (below).

### 2b — Cross-module reads of **auth / audit / platform** tables (the contested bucket)

The mission §2 Non-Goal records "No touching auth/security cross-module reads — already ported (M4 work
holds)." The widen shows these reads are **not ported** — they are still raw `FROM <foreign base table>` SQL.
Each carries prior disposition (inline `(M4/F4.x)` comments + ADR 0029/0031), but under **ADR-0039 D1 (broad
reading)** they are violations, and the mission's **terminal §8 bar measures the broad reading**.

| # | Site(s) | Reads (owner) | Nature | Prior disposition |
|---|---------|---------------|--------|-------------------|
| X1 | `security/infrastructure/postgres/repository.go:121,185,236` | `auth_identities` (auth) | lockout / suspicious-activity projections, scoped `WHERE user_id = ANY($1)` (ids from a port) | parent grade-a M4 F4.6; ADR 0031 (security scopes auth via `= ANY(ids)`) |
| X2 | `security/infrastructure/postgres/repository.go:262,269` | `auth_sessions` (auth) | new-device / session projections, tenant-scoped | parent M4 F4.6 (inline note) |
| X3 | `security/infrastructure/postgres/repository.go:340` | `audit_events` (audit) | security event projection | — (platform sink read) |
| X4 | `iam/infrastructure/postgres/observability_repository.go:63,75,93,152` | `audit_events` (audit) | metrics COUNT projections | — (platform sink read) |
| X5 | `iam/infrastructure/postgres/observability_repository.go:133,174` | `auth_identities` (auth) | active-users / lockout metrics (JOIN iam_users) | — |
| X6 | `iam/presence/repository.go:64–65` | `auth_identities` (auth) — `JOIN` iam_users (own) for `username` | presence snapshot username | — |
| X7 | `templates/repository/postgres.go:695` (`ListAudit`) | `audit_events` (audit) | template audit-history read from the canonical sink | Wave 1.8 (`AppendAudit` write-sink note) |
| X8 | `jobs/stuck_instance_watchdog/job.go:147,148` | `approval_instances`, `approval_stage_instances` (APP) | watchdog reads in-flight instances (worker layer) | — |

Observations that bear on classification:
- **`audit_events` is a platform append-sink** — the `audit` module owns it, but it is a cross-cutting sink
  every module *writes* via `AppendAudit[Tx]` and a few *read* for projections (templates `ListAudit`, iam
  metrics, security events). This is architecturally distinct from a domain module's private table.
- **`auth_identities`/`auth_sessions` reads in security/iam** were audited and accepted in the **parent
  grade-a-completion M4** (ADR 0029 `UserDisplayNameReader`, ADR 0031 `TenantUserReader` — the latter
  explicitly sanctions security reading `auth_identities` via `= ANY(ids)` because the table has no
  `tenant_id`). They are "already dispositioned," not "already ported to a view/port."
- **`jobs`** is a worker layer (`internal/modules/jobs/...`), not a domain module; whether it is subject to
  the same boundary rule (or is infrastructure operating on the approval domain) is undecided by the mission.

## Coverage statement

- **Swept:** all of `internal/modules/**` non-`_test.go` `.go`, for every owned base table in the owner map
  (Step A), via `FROM`/`JOIN`/`EXISTS` read-target diff (Steps B–C). This is **wider** than the brief's six
  named tokens — it ranges over the **full owned-table set** (incl. `iam_users`, `iam_user_roles`,
  `templates_template(_version)`, `document_comments`, `document_families`, `auth_*`, `audit_events`,
  approval tables).
- **Assumed / NOT swept (recorded, not engineered away):**
  - Dynamically-assembled SQL or table names behind Go variables/aliases — the grep matches **literal table
    tokens** only (same residual as the H-D guard). No evidence of such a site was found, but absence is not
    proof; the F0.3 guard inherits this limitation.
  - `_test.go` files — excluded by design (tests may legitimately cross modules).
  - Runtime/Docker (:5433) verification — **not run** (M0 is static analysis; Docker may be down). No false
    green: nothing here is runtime-reproduced.
- **Unclassified: 0.** Every cross-module SQL read above carries an ADR-0039 verdict. The one non-SQL coupling
  (A1–A3) is recorded as out-of-D1-range (literal fix, M1). The one false positive (`auth_failure_counters`)
  is recorded as same-module-compliant.

## Delta vs the discovery brief

- **Confirmed (kept):** all ~20 brief sites (A1–A3, B1–B8, C1–C4) reproduced at the cited lines — none dropped.
- **Added — document-domain:** N1 (`fillin_service.go:225` → `templates_template_version`). Routed to M2,
  **changes M2's shape** (one extra read-port) → HS-6.
- **Added — auth/audit/platform (X1–X8):** a class the mission Non-Goal excluded as "already ported", but
  which the widen shows are **raw base-table reads** still in force. Under ADR-0039's broad reading they are
  violations the **terminal §8 bar would flag**. → **HS-6** (scope-shape decision required from the operator).
- **Corrected:** `auth_failure_counters` owner = documents/approval (false-positive drop). Ownership
  corrections `document_process_areas`/`document_profiles` = taxonomy re-confirmed.

## HS-6 — RESOLVED 2026-06-20 (operator ruling)

The census surfaced sites that change milestone shape (N1) and a Non-Goal/terminal-bar contradiction (X1–X8).
Per mission §9 HS-6 these were surfaced to the operator before F0.3 finalized. **Operator ruling
(`./hs-6-scope-decision.md`):** **1a** — fold N1 into M2 (one templates read-port); **2a** — resolve X1–X8 via
principled ADR-0039 exemptions **D3(d)** platform append-sink (`audit_events`), **D3(e)**
parent-ADR-dispositioned auth (`auth_identities`/`auth_sessions`, ADR 0029/0031), **D3(f)** worker-layer
(`jobs`). ADR-0039 amended accordingly; the F0.3 guard allowlists X1–X8 with per-site justification.

**Post-ruling state:** mission appetite held (M1–M4 shape unchanged; M2 +1 port). Terminal "H-G=0 under both
readings" redefined honestly as **0 violations outside the recorded allowlist** (the (d)–(f) carve-outs are
enumerated, justified, not pretended absent). Census remains **0 unclassified**. F0.3 unblocked.
