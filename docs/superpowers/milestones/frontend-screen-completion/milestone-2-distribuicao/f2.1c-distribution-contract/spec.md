# Feature F2.1c — Spec (distribution-contract)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.1c-distribution-contract`
> **Status:** Approved (pre-code) — 2026-06-21 (re-decomposition gate, operator).
> **Approved before code:** 2026-06-21 / operator (leandrotca) — *new `distribution` module; dedicated `CapDistributionRead` (tenant-scope, deferredCaps — operator grants to roles separately); denominator-only contract; reads composed over F2.1a + F2.1b views + ADR-0029 iam port.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Re-scope note (2026-06-21)

This file supersedes the earlier `f2.1-coverage-contract/spec.md` (same folder, renamed). The
original spec assumed distribution could read `metaldocs.v_cd_grantee` + the existing
`v_active_user_areas` directly. Recon surfaced an ADR-0039 blocker: that view doesn't carry
`area_code`/`source` and excludes company-scope CDs by design (search-semantic contract).
Re-decomposed (HS-6, operator-chosen Option A) — F2.1a now publishes
`metaldocs.v_cd_obligated_readers` (CD-owned, carries the obligation shape); F2.1b publishes
`metaldocs.v_process_area_name` (taxonomy-owned, carries area labels). **This feature (F2.1c) is the
distribution-module contract; F2.2 implements over the two new views + the ADR-0029 iam display-name
port.** The consumer contract below is unchanged in shape from the operator-approved 2026-06-21
version — `source` simply gains a third variant `'company_scope'` to mirror F2.1a's leg.

## Interview record (fail-closed gate)

Contract discovered by reading the **consumer** (`DocumentDistributionPage.tsx` + its mock
`lib/distributionMeta.ts`) and read-only backend recon (OpenAPI flow, authz pattern, the denominator
source via the two new published views from F2.1a + F2.1b). Genuine contract questions + how each
resolved:

| # | Question | Answer |
|---|----------|--------|
| 1 | What exact shape does the consumer (`DocumentDistributionPage`) need that M2 can serve truthfully? | The **denominator** only: total obligated count (`KPIStrip`/`DonutCard` headline), the obligated recipient list with area (`RecipientsCard`), and by-area obligated totals (`CoverageByArea`). Numerator fields in `DistributionMock` (`read`, `acknowledged`, `pending`, `overdue`, `dailyReads`, `deadline`, `policy`, `channel`, `reminders*`) have **no producer** → excluded (rendered "tracking pending" in F2.3). |
| 2 | Does the recipient `role` (job title, e.g. "Op. Empilhadeira") have a real producer? | **No.** `metaldocs.iam_users` has `display_name` only — no job-title/position column. `user_process_areas.role` is an area-membership role (viewer/editor/approver/…), **not** a job title. Showing it as "role" would mislabel data → **omit `role`** from the contract (truthfulness, quality-goal 1). |
| 3 | How is the obligated reader set resolved? | Per `(tenant_id, controlled_document_id)` from **`metaldocs.v_cd_obligated_readers`** (F2.1a): three legs UNIONed + DISTINCT-by-user with source precedence `user_grant` > `area_grant` > `company_scope`. `area_name` resolved from **`metaldocs.v_process_area_name`** (F2.1b). User display name resolved via the **ADR-0029 iam display-name read-port**. Distribution **never** reads CD/taxonomy/iam base tables (`hgcrossmodule` = 0). |
| 4 | Source values in the response? | `'user_grant'` \| `'area_grant'` \| `'company_scope'` (mirrors F2.1a's leg discriminator). FE can label company-scope as "todos os colaboradores" or similar in F2.3. |
| 5 | Does the by-area `coverage` total dedupe against the global total? | **No — documented semantic.** Each coverage row counts active members of that area (`source='area_grant'` rows of `v_cd_obligated_readers` grouped by `area_code`); a user in two granted areas counts in both, and `user_grant`-only / `company_scope`-only users belong to no area. So `Σ coverage.total ≠ total_targets` by design. Stated in the schema description. Company-scope CDs: the coverage endpoint returns an empty array (no per-area breakdown applies); FE renders the "todos os colaboradores" label instead of a by-area card for those docs. |
| 6 | authz capability + scope? | **RESOLVED (operator, 2026-06-21): dedicated `CapDistributionRead = "distribution.read"`, tenant-scope, registered in `validCapabilities` + `capability_scope.go`, added to `scripts/api-lint/registry_rules.go:37 deferredCaps`** (operator grants to roles separately later — never pre-granted by the agent). Sensitive surface (reader coverage); not granted to mere doc-viewers. Tenant-scope matches `CapMetricsView`/`CapAuditRead` precedent (cross-area rollup). |
| 7 | Where does the endpoint live? | **New read-only module `internal/modules/distribution`** — greenfield (no `/distribution` route exists); co-locates with the future parked-mission write-path. |
| 8 | Pagination shape? | **Cursor (keyset) per the backend's canonical `CursorPage` convention** — *not* `X-Total-Count`/`Link` (Grade-A backend + `api-lint` use opaque forward cursors; recon confirmed at `api/openapi/v1/openapi.yaml:2069-2122` `listControlledDocuments`). Keyset/order: `area_name`, `name`, `user_id` (deterministic, stable for the cursor). Total lives in the summary endpoint (`total_targets`), not the list. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx` and its
  child cards (`KPIStrip`, `DonutCard`, `CoverageByArea`, `RecipientsCard`), via new TanStack Query
  hooks in `features/documents/queries/` (built in F2.3). The FE consumes the **generated** types from
  `frontend/apps/web/src/lib/api-types/index.d.ts` — never hand-rolled.
- **Contract** (denominator-only; all under `/api/v1`, tag `distribution`, gated by `CapDistributionRead`):

  **`GET /documents/{documentId}/distribution`** → `200 DistributionSummaryResponse`
  ```
  DistributionSummaryResponse { total_targets: integer }   // distinct obligated readers
  ```

  **`GET /documents/{documentId}/distribution/recipients?cursor={opaque}&limit={int 1..100}`**
  → `200 DistributionRecipientsResponse`. Keyset pagination; keyset order `area_name`, `name`, `user_id`.
  ```
  DistributionRecipientsResponse {
    items: DistributionRecipient[],
    page:  CursorPage                        // existing schema { next_cursor: string|null, has_more: bool }
  }
  DistributionRecipient {
    user_id:   string,
    name:      string,                        // iam display-name via ADR-0029 port
    area_code: string | null,                 // null when source != 'area_grant'
    area_name: string | null,                 // resolved via v_process_area_name (F2.1b)
    source:    "area_grant" | "user_grant" | "company_scope"
  }
  ```

  **`GET /documents/{documentId}/distribution/coverage`** → `200 DistributionAreaCoverage[]`
  Order: `area_name`. Counts `source='area_grant'` rows of `v_cd_obligated_readers` grouped by area.
  Empty array for company-scope CDs (no per-area breakdown).
  ```
  DistributionAreaCoverage { area_code: string, area_name: string, total: integer }
  ```

  **Errors:** RFC 9457 problem+json on every path (404 unknown/unauthorized document, 403 missing
  cap), matching the existing error envelope.
- **Source of truth for the contract:** the new OpenAPI operations in `api/openapi/v1/openapi.yaml`
  (tag `distribution`) → `oapi-codegen` server types → `openapi-typescript` FE types
  (`lib/api-types/index.d.ts`). This feature **authors that source of truth**; F2.2 implements the
  handlers; F2.3 consumes the generated FE types.

## What this feature implements

The contract artifacts only: the OpenAPI operations + schemas for the three denominator-only
endpoints (tag `distribution`), the regenerated Go server types + FE types, the new cap registration
(`CapDistributionRead` in iam's `model.go` + `capability_scope.go` + `deferredCaps` entry), and
**ADR-0042** recording the durable decisions (new module, new cap, denominator-only +
additive-extension commitment, recipient distinct/source rule, omit `role`). **No handler logic, no
SQL, no read-source wiring** — that is F2.2.

## Non-goals (mandatory)

- **No numerator anywhere:** no `read`, `acknowledged`, `pending`, `overdue`, `last_event_at`,
  `read_at`, `acknowledged_at`, deadline, policy, channel, reminder, or `timeline` field in any
  schema. The parked mission extends the contract **additively** later.
- **No timeline endpoint** (`/distribution/timeline`) — numerator-only.
- **No recipient `role`/job-title field** — no honest producer.
- **No new table, no migration in this feature, no change to `PublishApproved()`** — F2.1c is
  contract-only.
- **No action endpoints** (reminders/export/add-recipients/fanout-policy) — parked mission.
- **No handler/repository implementation** — that is F2.2.
- **No pre-grant of `CapDistributionRead` to any role** — `deferredCaps` is the right home;
  operator owns role-grant separately.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `api-lint -strict` parses the new `distribution` paths with **0** violations | `scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .` → exit 0 | real |
| The new response schemas contain **no** numerator field | `grep -nE "read\|acknowledg\|overdue\|pending\|deadline\|timeline\|reminder"` over the new `Distribution*` schema blocks in `openapi.yaml` = 0 | real |
| Generated **FE types** exist and are denominator-only | `npm run gen:api` clean; `grep -n "DistributionSummaryResponse\|DistributionRecipient\|DistributionAreaCoverage" frontend/apps/web/src/lib/api-types/index.d.ts` present; no read/ack fields in those types | real |
| Generated **Go server types** exist | `go build ./...` clean after `go generate` on the new module's `gen.go` | real |
| Cap registered + scoped + deferred | `grep -n "CapDistributionRead" internal/modules/iam/domain/model.go` ≥ 2 hits (const + validCapabilities); `grep -n "CapDistributionRead" internal/modules/iam/domain/capability_scope.go` = 1 hit (`ScopeTenant`); `grep -n "distribution.read" scripts/api-lint/registry_rules.go` = 1 hit in `deferredCaps` | real |
| Durable decisions recorded | ADR-0042 file present under `wiki/decisions/` and linked below | real |

> F2.1c has no runtime handler, so its "tests" are the contract gates above (api-lint + codegen +
> schema-grep + cap-registration greps). The handler TDD (failing integration test → green) lives in
> **F2.2**.

## ADR needed?

- [x] **Durable decision made** → ADR-0042 (new `distribution` module + `CapDistributionRead`
  tenant-scoped + denominator-only contract + additive-extension commitment for the parked mission +
  recipient distinct/`source` precedence rule + omit `role`). Link: _wiki/decisions/0042-distribution-module-and-cap.md (to be authored during execution)_.
  Companion ADRs ADR-0040 (F2.1a) + ADR-0041 (F2.1b) declare the read sources this feature's contract
  composes over.

---

### Approval-gate outcome (operator, 2026-06-21)

1. **Contract:** approved as specified (denominator-only).
2. **authz cap (interview #6):** **dedicated `CapDistributionRead`, tenant-scope, `deferredCaps`** —
   coverage is sensitive; not granted to mere doc-viewers; operator wires role grants separately.
3. **module placement (interview #7):** **new `internal/modules/distribution`** read module.
4. **`role` omission (interview #2):** accepted — the recipient job-title column **disappears** from
   the screen (no honest producer).
5. **Read sources (re-decomposition, HS-6):** `metaldocs.v_cd_obligated_readers` (F2.1a) +
   `metaldocs.v_process_area_name` (F2.1b) + ADR-0029 iam display-name port. **No** raw base-table
   reads.
