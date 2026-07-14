# ADR 0042 — new `distribution` module + `CapDistributionRead` + denominator-only coverage contract

> **Status:** Accepted 2026-06-21
> **Last verified:** 2026-06-22
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M2 (Distribuição coverage-scope) · Feature F2.1c.
> **Supersedes:** none.
> **Related ADRs:** [0040 — `v_cd_obligated_readers`](./0040-cd-obligated-readers-view.md) (the obligated-set read source this contract projects); [0041 — `v_process_area_name`](./0041-taxonomy-process-area-name-view.md) (the area-label read source); [0029 — `UserDisplayNameReader`](./0029-user-display-name-reader-port.md) (the iam display-name read-port supplying `name`); [0039 — Cross-module read boundary](./0039-cross-module-base-table-read-boundary.md) (distribution is ADR-0039-compliant by reading only those published views + the iam port); [0024 — Single base path](./0024-openapi-single-base-path.md); [0012 — Contract-first API](./0012-contract-first-api.md); [0022 — Authz tiers](./0022-authz-capability-coherence.md) (cap registry + scope).
> **Related code (Last verified 2026-06-22):**
> - `api/openapi/v1/openapi.yaml` — tag `distribution`; ops `getDocumentDistribution`, `listDocumentDistributionRecipients`, `getDocumentDistributionCoverage`; schemas `DistributionSummaryResponse`, `DistributionRecipient`, `DistributionRecipientsResponse`, `DistributionAreaCoverage`.
> - `internal/modules/distribution/api/{cfg.yaml,gen.go,api.gen.go}` — generated server types (`include-tags: [distribution]`).
> - `internal/modules/iam/domain/model.go` — `CapDistributionRead = "distribution.read"` (const + validCapabilities).
> - `internal/modules/iam/domain/capability_scope.go` — `CapDistributionRead: ScopeTenant`.
> - `scripts/api-lint/registry_rules.go` — `deferredCaps[CapDistributionRead]`.
> - `wiki/backlog/document-distribution-mission.md` — the parked numerator + action mission this contract extends additively.

## Context

M2 builds the Distribuição & Cobertura screen against a real backend. Runtime truth (recon, HEAD `d477e9f0`): the **denominator** (the obligated audience of a controlled document) is derivable, but the **numerator** (any read/acknowledge event, distribution target, reminder job) has no producer anywhere. The operator split M2 (HS-6): build the read-only denominator now; park the numerator + action layer as a designed mission. F2.1a/F2.1b published the two views the read composes over; this feature (F2.1c) authors the consumer-facing contract; F2.2 implements the handlers.

## Decision

### D1 — A new read-only `distribution` module

`internal/modules/distribution` is greenfield (no `/distribution` route exists). It co-locates with the future parked-mission write-path. It is a non-owner of CD/taxonomy/iam, so per ADR-0039 it reads **only** published views (`v_cd_obligated_readers`, `v_process_area_name`) + the ADR-0029 iam display-name port — never base tables.

### D2 — Denominator-only contract (three GET reads)

- `GET /documents/{id}/distribution` → `DistributionSummaryResponse { total_targets }`.
- `GET /documents/{id}/distribution/recipients?cursor=&limit=` → `DistributionRecipientsResponse { items: DistributionRecipient[], page: CursorPage }`, keyset order `area_name, name, user_id`.
- `GET /documents/{id}/distribution/coverage` → `DistributionAreaCoverage[]` (bare array; empty for company-scope docs).

`DistributionRecipient` carries `{ user_id, name, area_code|null, area_name|null, source }`. **No `role`** — `iam_users` has no job-title column and `user_process_areas.role` is a membership role, not a title; surfacing it would mislabel data (truthfulness).

### D3 — `CapDistributionRead`, tenant-scope, deferred

New cap `distribution.read`, registered in the iam registry, classified `ScopeTenant` (cross-area rollup, matching `CapMetricsView`/`CapAuditRead`). Added to `deferredCaps`: it is a sensitive coverage surface, deliberately **not** seeded to any tenant role by the agent — the operator grants it to roles separately. Handlers (F2.2) gate with `authz.Require` + the `trg_require_cap_asserted` tripwire.

### D4 — `source` precedence + by-area coverage semantics

Recipients are DISTINCT by user with precedence `user_grant > area_grant > company_scope` (inherited from `v_cd_obligated_readers`, ADR-0040). By-area `coverage` counts `source='area_grant'` members per area; `Σ coverage.total ≠ total_targets` by design (user-grant-only and company-scope users belong to no area; multi-area users count once per area). Documented in the schema.

### D5 — Forward-compatible, additive-only commitment

The contract is denominator-only and **forward-compatible**: the parked `document-distribution-mission` extends it **additively** (new numerator fields/endpoints), never breaking the shapes minted here. No numerator vocabulary (`read`/`acknowledged`/`overdue`/`pending`/`deadline`/`timeline`/`reminder`) appears in any schema.

## Consequences

- The FE (F2.3) consumes generated types only; the denominator surfaces render live, the numerator renders an honest "tracking pending" state.
- The `distribution` module is ADR-0039-compliant by construction (`hgcrossmodule` = 0 once F2.2 lands).
- One new cap in the registry; zero new role grants until the operator acts.
- This feature ships **no handler, no SQL, no migration** — F2.1c is contract-only; F2.2 implements.

## Verification

- `api-lint -strict` = 0 over the new `distribution` paths; numerator-grep = 0 over the new schema blocks.
- Generated Go types (`internal/modules/distribution/api/api.gen.go`) + FE types (`lib/api-types/index.d.ts`) present and denominator-only; `go build ./...` green.
- Cap registered (`model.go`), classified (`capability_scope.go`), deferred (`registry_rules.go`); `TestEveryCapabilityClassified` + `TestEveryCapSeededOrDeferred` + api-lint registry lints green.
