# Feature F2.1c — distribution-module contract (OpenAPI + codegen + cap + ADR-0042)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.1c-distribution-contract`
> **Status:** Planning → ready for execution

## Source

- Spec (`./spec.md`, approved pre-code 2026-06-21): consumer-contract-first; three **denominator-only** read endpoints under tag `distribution`; new tenant-scope cap `CapDistributionRead` in `deferredCaps`; regen Go + FE types; ADR-0042. **No handler/SQL/migration** — that is F2.2.
- Milestone row (F2.1c, `../milestone.md`): authors the OpenAPI + ADR + codegen; reads compose over F2.1a `v_cd_obligated_readers` + F2.1b `v_process_area_name` + ADR-0029 iam display-name port (wired in F2.2).
- Consumer: `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx` + child cards, via TanStack hooks built in **F2.3**, consuming the **generated** FE types only.

### Verified codebase facts (recon this session — embed, do not re-derive)

- **Path-param convention:** every sibling `/documents/{...}` op uses `{id}` (e.g. `openapi.yaml:2404 /documents/{id}`), **not** `{documentId}`. The spec's `{documentId}` was illustrative; the URL path-param name is cosmetic (the FE constructs the URL; it is not part of the response shape the generated types carry). Use `{id}` for sibling consistency.
- **Global security:** `openapi.yaml:12-13` declares `security: [ - sessionCookie: [] ]`. Read GETs inherit it; `listControlledDocuments` (a GET) carries no per-op `security` and passes `api-lint -strict` today. → **no per-op `security` block needed**; AUTHZ-DRIFT is satisfied by inheritance.
- **api-lint `-strict` rules that bind here** (`scripts/api-lint/spec_rules.go`):
  - `PAGINATION-DRIFT` fires **only when `operationId` starts with `list`** (`spec_rules.go:248`). → `listDocumentDistributionRecipients` MUST carry query `cursor` (string, not-required) + `limit` (integer, not-required) and a `page` with `next_cursor` + `has_more`. `getDocumentDistribution` + `getDocumentDistributionCoverage` use a `get…` prefix, so the **bare-array** coverage response is allowed (no pagination forced).
  - `ENVELOPE-DRIFT` applies only to **error** responses (must `$ref` a Problem response). Success bodies (incl. a top-level array) are unconstrained.
  - `CASING-DRIFT` — every schema property must be snake_case (all of ours are).
  - `AUTHZ-DRIFT` area markers (`x-authz-area*`) are required only on **state-transition POSTs**; our ops are GET reads → none.
- **Numerator-grep gate** (`spec.md` Validation Gate): `grep -nE "read|acknowledg|overdue|pending|deadline|timeline|reminder"` over the new `Distribution*` schema blocks must be **0**. `read` matches the substring in "reader"/"read" — so **every new description + path `summary` says "obligated audience", never "obligated reader"**, and no numerator vocabulary appears anywhere in the new content.
- **Codegen wiring** (`wiki/references/oapi-codegen.md`; `internal/modules/controlleddocuments/api/{cfg.yaml,gen.go}`): per-module `cfg.yaml` (`include-tags: [<tag>]`) + `gen.go` `//go:generate`; run with `GOFLAGS=-mod=mod go generate ./internal/modules/<x>/api/...` (vendor-mode gotcha); commit `api.gen.go`. `internal/modules/distribution/api/` is 4 dirs deep → `../../../../api/openapi/v1/openapi.yaml` (identical depth to controlleddocuments).
- **Cap registry** (`internal/modules/iam/domain/model.go`): const block ends `CapSessionManage` (`:121`); `validCapabilities` map ends `CapSessionManage: {}` (`:161`). `capability_scope.go:36 capabilityScopes` classifies every cap — `TestEveryCapabilityClassified` fails the build if a registry cap is unclassified. `scripts/api-lint/registry_rules.go:37 deferredCaps` is currently `{}`; a registry cap seeded to no role MUST appear there or `seed-registry-parity` (api-lint) + `TestEveryCapSeededOrDeferred` (`apps/api/cmd/metaldocs-api/permissions_test.go`) both red.
- **FE codegen** (`frontend/apps/web/package.json:13`): `gen:api` = `openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts`.

---

## Plan

# Feature F2.1c Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the contract source-of-truth for the new `distribution` module: three denominator-only read endpoints (tag `distribution`) in `openapi.yaml`, the regenerated Go server types (`internal/modules/distribution/api/api.gen.go`) + FE types (`lib/api-types/index.d.ts`), the new tenant-scope cap `CapDistributionRead` registered + scoped + deferred, and ADR-0042. **No handler logic, no SQL, no migration** (that is F2.2).

**Architecture:** Contract-first (ADR-0012). The OpenAPI operations are the single source of truth; `oapi-codegen` (Go) + `openapi-typescript` (FE) derive types from it. The new module owns no path mux yet (no handler), so the generated Go interface compiles unimplemented and `go build ./...` stays green. The cap is minted in the iam registry, classified tenant-grade (matching `CapMetricsView`/`CapAuditRead` cross-area-rollup precedent), and parked in `deferredCaps` — the agent never seeds it to a role. The contract is **denominator-only and forward-compatible**: the parked `document-distribution-mission` extends it additively (ADR-0042).

**Tech Stack:** OpenAPI 3 (`api/openapi/v1/openapi.yaml`); `oapi-codegen` v2.7.0 (vendor-mode, `GOFLAGS=-mod=mod`); `openapi-typescript` ^7.13.0; Go capability registry under `internal/modules/iam/domain`; `scripts/api-lint` strict linter + Go parity tests.

---

### Task 1: Author the OpenAPI contract (tag + paths + schemas)

**Files:**
- Modify: `api/openapi/v1/openapi.yaml` — add the `distribution` tag (after the `documents` tag block, currently `:31-32`); add the four schemas (immediately after the `CursorPage` schema block, currently ending `:3989`, before `ListUsersResponse:` `:3990`); add the three path items (at the end of the `paths:` map, immediately before the `components:` key, currently `:3701`).

- [ ] **Step 1: Add the `distribution` tag**

In the top-level `tags:` list, immediately after the `documents` entry (`:31-32`), insert:

```yaml
  - name: distribution
    description: Controlled-document distribution coverage — the obligated audience (denominator only; numerator parked, ADR-0042).
```

- [ ] **Step 2: Add the four response schemas**

In `components.schemas`, immediately after the `CursorPage` block (after `:3989`, before `ListUsersResponse:`), insert. **Every description avoids the tokens `read|acknowledg|overdue|pending|deadline|timeline|reminder` — "obligated audience", never "obligated reader":**

```yaml
    DistributionSummaryResponse:
      type: object
      required: [total_targets]
      properties:
        total_targets:
          type: integer
          minimum: 0
          description: Distinct count of users obligated by the controlled document — the coverage denominator. The numerator is out of M2 scope (ADR-0042).
    DistributionRecipient:
      type: object
      required: [user_id, name, area_code, area_name, source]
      properties:
        user_id:
          type: string
        name:
          type: string
          description: User display name, resolved via the iam UserDisplayNameReader port (ADR-0029).
        area_code:
          type: string
          nullable: true
          description: Granting area code; null unless source is area_grant.
        area_name:
          type: string
          nullable: true
          description: Area label resolved via metaldocs.v_process_area_name (ADR-0041); null unless source is area_grant.
        source:
          type: string
          enum: [area_grant, user_grant, company_scope]
          description: How the user became obligated. Distinct-by-user precedence user_grant > area_grant > company_scope (ADR-0040 / ADR-0042).
    DistributionRecipientsResponse:
      type: object
      required: [items, page]
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/DistributionRecipient'
        page:
          $ref: '#/components/schemas/CursorPage'
    DistributionAreaCoverage:
      type: object
      required: [area_code, area_name, total]
      properties:
        area_code:
          type: string
        area_name:
          type: string
        total:
          type: integer
          minimum: 0
          description: Active members of this area obligated via area_grant. The sum across areas need not equal total_targets (user_grant-only and company_scope users belong to no area; multi-area users count once per area). Empty array for company-scope documents.
```

- [ ] **Step 3: Add the three path items**

At the end of the `paths:` map, immediately before the `components:` key (`:3701`), insert. Note `operationId` prefixes: `get…` for the two non-paginated ops, `list…` for the paginated recipients op:

```yaml
  /documents/{id}/distribution:
    get:
      summary: Distribution coverage summary — total obligated audience for a controlled document
      tags: [distribution]
      operationId: getDocumentDistribution
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DistributionSummaryResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
  /documents/{id}/distribution/recipients:
    get:
      summary: Obligated audience for a controlled document (denominator only)
      tags: [distribution]
      operationId: listDocumentDistributionRecipients
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
        - name: cursor
          in: query
          description: Opaque forward keyset cursor from a prior page's page.next_cursor.
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, minimum: 1, maximum: 100 }
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DistributionRecipientsResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
  /documents/{id}/distribution/coverage:
    get:
      summary: By-area obligated totals for a controlled document (denominator only)
      tags: [distribution]
      operationId: getDocumentDistributionCoverage
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/DistributionAreaCoverage'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
```

- [ ] **Step 4: Run `api-lint -strict` — expect 0 violations on the new paths**

The api-lint binary is committed at `scripts/api-lint/api-lint.exe`; if absent, build it: `go build -o scripts/api-lint/api-lint.exe ./scripts/api-lint`.

Run:
```bash
./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .
echo "api-lint exit: $?"
```
Expected: exit `0`, no `PAGINATION-DRIFT` / `AUTHZ-DRIFT` / `CASING-DRIFT` / `ENVELOPE-DRIFT` on `getDocumentDistribution`, `listDocumentDistributionRecipients`, `getDocumentDistributionCoverage`.

If `PAGINATION-DRIFT` appears on a `get…` op, an operationId was mis-prefixed; if it appears on the recipients op, re-check the `cursor`/`limit` query params + `page.next_cursor`/`page.has_more`.

- [ ] **Step 5: Numerator-grep gate — 0 hits in the new schema blocks**

Slice the new schema span (`DistributionSummaryResponse` through the end of `DistributionAreaCoverage`, before `ListUsersResponse:`) and grep it for numerator vocabulary:

```bash
sed -n '/    DistributionSummaryResponse:/,/    ListUsersResponse:/p' api/openapi/v1/openapi.yaml \
  | grep -nE "read|acknowledg|overdue|pending|deadline|timeline|reminder"
```
Expected: **0** hits. Also grep the three new path summaries directly:
```bash
grep -nE "read|acknowledg|overdue|pending|deadline|timeline|reminder" api/openapi/v1/openapi.yaml \
  | grep -iE "distribution|total_targets|area_grant|DistributionRecipient"
```
Expected: **0** hits. ("reader" must not appear — that is why summaries/descriptions say "audience".)

- [ ] **Step 6: Commit the contract**

```bash
git add api/openapi/v1/openapi.yaml
git commit -m "feat(M2/F2.1c): OpenAPI distribution tag — 3 denominator-only read endpoints"
```

---

### Task 2: New `distribution` api package + Go codegen

**Files:**
- Create: `internal/modules/distribution/api/cfg.yaml`
- Create: `internal/modules/distribution/api/gen.go`
- Generated (committed): `internal/modules/distribution/api/api.gen.go`

- [ ] **Step 1: Write `cfg.yaml`** (mirrors `internal/modules/controlleddocuments/api/cfg.yaml`)

```yaml
package: distributionapi
generate:
  models: true
  std-http-server: true
  strict-server: true
  embedded-spec: true
output: api.gen.go
output-options:
  include-tags:
    - distribution
```

- [ ] **Step 2: Write `gen.go`** (4 dirs deep → `../../../../`, identical to controlleddocuments)

```go
package distributionapi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../../../api/openapi/v1/openapi.yaml
```

- [ ] **Step 3: Generate the Go types**

```bash
GOFLAGS=-mod=mod go generate ./internal/modules/distribution/api/...
```
Expected: creates `internal/modules/distribution/api/api.gen.go` containing only the three `distribution`-tagged ops + their schemas (`DistributionSummaryResponse`, `DistributionRecipient`, `DistributionRecipientsResponse`, `DistributionAreaCoverage`) — nothing from other tags (include-tags filter).

- [ ] **Step 4: Build — generated interface compiles unimplemented**

```bash
go build ./...
echo "build exit: $?"
```
Expected: exit `0`. The generated `ServerInterface`/`StrictServerInterface` are declarations only; no handler references them yet (F2.2), so the build is green.

- [ ] **Step 5: Verify the generated types are denominator-only**

```bash
grep -nE "DistributionSummaryResponse|DistributionRecipient|DistributionAreaCoverage|TotalTargets" internal/modules/distribution/api/api.gen.go | head
grep -nE "Read|Acknowledg|Overdue|Pending|Deadline|Timeline|Reminder" internal/modules/distribution/api/api.gen.go
```
Expected: the three response types present; the numerator grep returns **0** hits.

- [ ] **Step 6: Commit the generated package**

```bash
git add internal/modules/distribution/api/cfg.yaml internal/modules/distribution/api/gen.go internal/modules/distribution/api/api.gen.go
git commit -m "feat(M2/F2.1c): generate distribution Go server types (oapi-codegen)"
```

---

### Task 3: Register `CapDistributionRead` — registry + scope + deferred

**Files:**
- Modify: `internal/modules/iam/domain/model.go` (const block + `validCapabilities`)
- Modify: `internal/modules/iam/domain/capability_scope.go` (`capabilityScopes`)
- Modify: `scripts/api-lint/registry_rules.go` (`deferredCaps` + its comment)

- [ ] **Step 1: Add the const + registry entry in `model.go`**

In the `const (...)` capability block, immediately after `CapSessionManage Capability = "session.manage"` (`:121`), add:

```go
	CapSessionManage               Capability = "session.manage"
	CapDistributionRead            Capability = "distribution.read"
```

In `validCapabilities`, immediately after `CapSessionManage: {},` (`:161`), add:

```go
	CapSessionManage:               {},
	CapDistributionRead:            {},
```

- [ ] **Step 2: Classify it tenant-grade in `capability_scope.go`**

In `capabilityScopes`, in the tenant-grade section immediately after `CapSessionManage: ScopeTenant,` (`:68`), add:

```go
	CapSessionManage:   ScopeTenant,
	CapDistributionRead: ScopeTenant,
```

(Tenant-grade matches `CapMetricsView`/`CapAuditRead` — cross-area rollup reads, per spec interview #6. `TestAreaGradeCapabilitySet` is unaffected; `TestEveryCapabilityClassified` now passes because the registry cap is classified.)

- [ ] **Step 3: Defer it in `registry_rules.go`**

Replace the `deferredCaps` declaration (`:31-37`) — update the comment and add the entry:

```go
// deferredCaps lists registry capabilities intentionally seeded to no tenant
// role (enforced today only via the system_admin tier-2 bypass). Mirrors the
// allow-list in apps/api/cmd/metaldocs-api/permissions_test.go
// (TestEveryCapSeededOrDeferred). A routed-but-unseeded cap goes here with a
// documented deferral.
//
//   distribution.read — minted in mission frontend-screen-completion M2/F2.1c
//   (ADR-0042). Sensitive coverage surface; deliberately NOT seeded to any
//   tenant role by the agent — the operator grants it to roles separately.
var deferredCaps = map[iamdomain.Capability]struct{}{
	iamdomain.CapDistributionRead: {},
}
```

- [ ] **Step 4: Build + run the parity tests**

```bash
go build ./...
go test ./internal/modules/iam/domain/... -run "TestEveryCapabilityClassified|TestAreaGradeCapabilitySet" -v
go test ./apps/api/cmd/metaldocs-api/... -run "TestEveryCapSeededOrDeferred" -v
go test ./scripts/api-lint/... -v
```
Expected: all PASS. The iam-domain tests confirm classification coverage; `TestEveryCapSeededOrDeferred` confirms `distribution.read` is registry-known and deferred (not orphaned); the api-lint package tests confirm `seed-registry-parity` + `no-inline-capability` hold.

- [ ] **Step 5: Re-run `api-lint -strict` (registry-binding lints over the live tree)**

```bash
./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .
echo "api-lint exit: $?"
```
Expected: exit `0` — `seed-registry-parity` does not flag `distribution.read` (covered by `deferredCaps`); `wiki-capability-parity` does not flag it (no `cap:distribution.read` marker exists in the scanned authz docs, and the rule is one-directional — markers must be in the registry, not the reverse).

- [ ] **Step 6: Cap-registration greps (spec Validation Gate rows)**

```bash
grep -n "CapDistributionRead" internal/modules/iam/domain/model.go
grep -n "CapDistributionRead" internal/modules/iam/domain/capability_scope.go
grep -n "distribution.read" scripts/api-lint/registry_rules.go
```
Expected: `model.go` ≥ 2 hits (const + validCapabilities); `capability_scope.go` = 1 hit (`ScopeTenant`); `registry_rules.go` ≥ 1 hit in `deferredCaps`.

- [ ] **Step 7: Commit the cap registration**

```bash
git add internal/modules/iam/domain/model.go internal/modules/iam/domain/capability_scope.go scripts/api-lint/registry_rules.go
git commit -m "feat(M2/F2.1c): register CapDistributionRead (tenant-scope, deferred)"
```

---

### Task 4: Regenerate FE types

**Files:**
- Generated (committed): `frontend/apps/web/src/lib/api-types/index.d.ts`

- [ ] **Step 1: Regenerate**

```bash
cd frontend/apps/web && npm run gen:api
```
Expected: clean exit; `src/lib/api-types/index.d.ts` rewritten with the new `distribution` paths + `DistributionSummaryResponse` / `DistributionRecipient` / `DistributionRecipientsResponse` / `DistributionAreaCoverage` under `components["schemas"]`.

> If this fails on the pnpm junction drift (`[[fe-node-modules-junction-drift]]`), tactically repoint the `openapi-typescript` bin / `entities` + `lru-cache` junctions as previously, regenerate, then proceed. Do not attempt the full pnpm install here — that is out of feature scope.

- [ ] **Step 2: Verify the FE types exist and are denominator-only**

```bash
grep -n "DistributionSummaryResponse\|DistributionRecipient\|DistributionAreaCoverage\|total_targets" frontend/apps/web/src/lib/api-types/index.d.ts | head
grep -nE "read|acknowledg|overdue|pending|deadline|timeline|reminder" frontend/apps/web/src/lib/api-types/index.d.ts | grep -iE "distribution|total_targets|area_grant"
```
Expected: the new schema names present; the numerator grep returns **0** hits attributable to distribution. `source` is the union `"area_grant" | "user_grant" | "company_scope"`; `area_code`/`area_name` are `string | null`.

- [ ] **Step 3: Commit the FE types**

```bash
git add frontend/apps/web/src/lib/api-types/index.d.ts
git commit -m "feat(M2/F2.1c): regenerate FE distribution types (openapi-typescript)"
```

---

### Task 5: ADR-0042 + index entry + spec link

**Files:**
- Create: `wiki/decisions/0042-distribution-module-and-cap.md`
- Modify: `wiki/decisions/index.md` (insert the 0042 entry — only if hand-maintained)
- Modify: `docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1c-distribution-contract/spec.md` (fill the ADR-0042 link line)

- [ ] **Step 1: Write ADR-0042**

```markdown
# ADR 0042 — new `distribution` module + `CapDistributionRead` + denominator-only coverage contract

> **Status:** Accepted 2026-06-21
> **Last verified:** 2026-06-22
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M2 (Distribuição coverage-scope) · Feature F2.1c.
> **Supersedes:** none.
> **Related ADRs:** [0040 — `v_cd_obligated_readers`](./0040-cd-obligated-readers-view.md) (the obligated-set read source this contract projects); [0041 — `v_process_area_name`](./0041-taxonomy-process-area-name-view.md) (the area-label read source); [0029 — `UserDisplayNameReader`](./0029-user-display-name-reader-port.md) (the iam display-name read-port supplying `name`); [0039 — Cross-module read boundary](./0039-cross-module-base-table-read-boundary.md) (distribution is ADR-0039-compliant by reading only those published views + the iam port); [0024 — Single base path](./0024-openapi-single-base-path.md); [0012 — Contract-first API](./0012-contract-first-api.md); [0022 — Authz tiers](./0022-authz-capability-registry.md) (cap registry + scope).
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
```

- [ ] **Step 2: Insert the index entry (if hand-maintained)**

Read `wiki/decisions/index.md`. If it lists ADRs in order, insert a `0042` row after `0041`. **If a banner says it is auto-generated, skip.**

- [ ] **Step 3: Fill the spec's ADR link line**

In `spec.md` §"ADR needed?", replace `_wiki/decisions/0042-distribution-module-and-cap.md (to be authored during execution)_` with `[wiki/decisions/0042-distribution-module-and-cap.md](../../../../../wiki/decisions/0042-distribution-module-and-cap.md)` (verify the relative depth resolves from the feature folder; adjust `../` count if needed).

- [ ] **Step 4: Verify + commit**

```bash
ls wiki/decisions/0042-distribution-module-and-cap.md
git add wiki/decisions/0042-distribution-module-and-cap.md docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1c-distribution-contract/spec.md
git add wiki/decisions/index.md  # only if Step 2 modified it
git commit -m "docs(adr): ADR-0042 distribution module + CapDistributionRead (M2/F2.1c)"
```

---

### Task 6: Full Validation Gate + evidence

Mirrors `spec.md` §Validation Gate row-by-row. No code changes — evidence collection only.

- [ ] **Step 1: api-lint -strict = 0**

```bash
./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .
echo "exit: $?"
```
Expected: exit `0`.

- [ ] **Step 2: Numerator-grep = 0 over the new schema blocks**

```bash
grep -nE "read|acknowledg|overdue|pending|deadline|timeline|reminder" api/openapi/v1/openapi.yaml | grep -iE "distribution|total_targets|area_grant|DistributionRecipient"
```
Expected: no hits.

- [ ] **Step 3: Generated FE types present + denominator-only**

```bash
grep -n "DistributionSummaryResponse\|DistributionRecipient\|DistributionAreaCoverage" frontend/apps/web/src/lib/api-types/index.d.ts | head
```
Expected: present; no read/ack fields.

- [ ] **Step 4: Generated Go types present + build green**

```bash
ls internal/modules/distribution/api/api.gen.go
go build ./...
echo "build exit: $?"
```
Expected: file present; build exit `0`.

- [ ] **Step 5: Cap registered + scoped + deferred**

```bash
grep -n "CapDistributionRead" internal/modules/iam/domain/model.go
grep -n "CapDistributionRead" internal/modules/iam/domain/capability_scope.go
grep -n "distribution.read" scripts/api-lint/registry_rules.go
```
Expected: model.go ≥ 2; capability_scope.go = 1; registry_rules.go ≥ 1.

- [ ] **Step 6: ADR-0042 present**

```bash
ls wiki/decisions/0042-distribution-module-and-cap.md
```
Expected: file exists.

- [ ] **Step 7: Regression — vet + tests + boundary scope unchanged**

```bash
go vet ./...
go test ./...
git diff --stat origin/main -- db/migrations            # expect: empty (no migration this feature)
git diff --stat origin/main -- internal/modules/documents/approval/application/publish_service.go  # expect: empty
git diff --stat origin/main -- internal/modules/search   # expect: empty
go run ./tools/cilint/...                                 # expect: exit 0 (no handler yet → no new cross-module read)
```
Expected: vet + tests green; the three scope diffs empty; cilint exit `0`.

- [ ] **Step 8: Capture evidence**

Copy `.claude/skills/milestone/templates/feature-evidence.md` → this folder's `evidence.md`; fill each row with the exact command + real output from Steps 1–7. Label every row `real`. Map each `spec.md` Validation Gate criterion to its evidence row (a suite-level "green" without per-criterion mapping is not acceptance).

- [ ] **Step 9: Commit evidence + close the feature**

```bash
git add docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1c-distribution-contract/evidence.md
git commit -m "docs(M2/F2.1c): close evidence — distribution contract gate green"
```

---

## Execution notes

Filled during `superpowers:subagent-driven-development`:

- Model: Sonnet 4.6 per `[[workflow-model-balancing]]` + operator directive 2026-06-21.
- Deviations from plan: <none yet>.
- Open questions answered: <none yet>.
