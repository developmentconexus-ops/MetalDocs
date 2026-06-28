# System-impact analysis — Tenant Token Dictionary (SP-1)

**Date:** 2026-06-28
**Intent (one line):** Add a tenant-scoped token dictionary backend — domain + storage + repo + new IAM capabilities (`token.view`, `token_dictionary.manage`) + capability-gated CRUD + audit.
**Work type:** module
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow

> Same ten sections for module and feature work. For a feature, mark module-only rows **N/A** with a
> one-line reason — do not delete them. Every row is a question the system forced you to answer.

---

## 1. Classify & own

- **Work type:** module — SP-1 births a new bounded-context module `tokens` under `internal/modules/tokens/`. It has its own domain entities (DictionaryEntry), its own tenant-scoped table, its own CRUD application service, its own delivery layer, and its own published port. A feature addition inside an existing module would not accommodate all of this without violating invariant 6 (cross-module via published interface).

- **Owning module(s):** `internal/modules/tokens/` (new) — this module owns the tenant dictionary concept: the domain entity (name→value pairs scoped to a tenant), the repo, the CRUD lifecycle, and the audit trail for dictionary changes. The spec (§4) explicitly labels it "new module surface".

- **Explicitly NOT owning:**
  - `templates` — owns template bodies, the computed-catalog endpoint, and tag validation. It consumes the dictionary (SP-2 will add that edge) but must not own the dictionary's storage or lifecycle. If templates owned the dictionary, the render module would transitively depend on templates to resolve values, creating an architectural coupling the spec calls out as a constraint ("render must not depend on templates").
  - `iam` — owns capability definitions and authz machinery. It does not own business domain entities like dictionary entries. SP-1 adds capabilities in `iam`, but the dictionary data lives in `tokens`.
  - `render` — owns freeze-time substitution. SP-1 does not touch render; SP-2 will inject dictionary values into `ResolvedValues` via a port on `tokens`.

- **Cross-module edges (with direction):**
  - `tokens` → `iam/authz` — uses `authz.SeedTxIdentity` + `authz.Require` for in-tx tier-2 enforcement. This is the standard platform authz path, not a module-to-module domain dependency.
  - `tokens` → `audit` — calls `audit.RecordTx` inside the business tx for any mutating operation. Goes through `audit`'s published `RecordTx` function, not its tables.
  - Future (SP-2): `render` → `tokens` (one direction only: render reads dictionary via `tokens`'s published port `domain/port.go`; `tokens` does NOT depend on `render`). This direction constraint is locked: it prevents a cycle.

- **Ambiguity?** None. Module boundary is unambiguous. No AS-3.

---

## 2. Foundation verdict

- **Base you'd build on:** The existing `render/resolvers/` computed-token infrastructure plus the `fillin_service` placeholder infrastructure — both are sound, production-stable, and grade-A audited (backend signed off 2026-06-21). The new `tokens` module does not modify those; it adds a parallel, orthogonal data source that SP-2 will inject at freeze time.

- **Sound, or legacy/patch/workaround?** Sound. The spec (§6) explicitly identifies the *frontend* authoring surface as the local maximum (dead `getAgent` pattern, schema corruption), but the *backend* resolver and fillin infrastructure is explicitly kept authoritative ("The mature parts stay authoritative; increments build on them"). SP-1 builds on the sound backend base. No AS-2.

- **Global-maximum structure:** A first-class bounded-context module following the exact same shape as `taxonomy` (smallest complete exemplar) or `templates`. This is already the global maximum for a new tenant-scoped dictionary: it has a domain layer, application service with `TxRunner`, an infra repo touching only its own tables, and a published port for SP-2 to consume. No workaround pattern in play.

---

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **Yes** — two new caps gate all CRUD | New caps `token.view` (ScopeTenant) and `token_dictionary.manage` (ScopeTenant) declared in `iam/domain/model.go`, route-mapped in `permissions.go`, `authz.Require` called in-tx in every mutating application-service method | `authz.SeedTxIdentity`, `authz.Require(ctx, tx, cap, areaCode)` (pattern: `templates/application/create.go:63`) |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** — new CRUD routes | Edit `api/openapi/v1/openapi.yaml` first (tag `tokens`); add `tokens` entry to `api/cfg.yaml`; run `go generate ./internal/modules/tokens/api/...` — never hand-add a Go route | `api/cfg.yaml` + `gen.go` per-module generation |
| Multi-tenant pooled (`tenant_id` / tx-local GUC / 404 cross-tenant) | **Yes** — dictionary entries are tenant-scoped | New table `token_dictionary_entries` has `tenant_id NOT NULL`; every query predicated on it; `tenant.FromContext` seeds the tenant; `authz.SeedTxIdentity` sets tx-local GUCs; cross-tenant lookup returns 404 | `tenant.FromContext` (`internal/platform/tenant/context.go:27`), `authz.SeedTxIdentity` (`internal/modules/iam/authz/context.go:58`) |
| Async = transactional outbox | **Not touched** — CRUD to own table is synchronous state-write with no external side effect. No network calls, no email, no blob, no webhook in SP-1 | N/A — no external side effects in SP-1; SP-2 (render injection) is a synchronous resolver read, not an outbox event | N/A |
| DB enforces invariants (triggers/constraints) | **Yes** — tenant_id NOT NULL + unique constraint on (tenant_id, name) prevents duplicate keys per tenant | Migration adds `NOT NULL tenant_id`, `UNIQUE (tenant_id, name)`, and a format check constraint on `name` (snake_case, no special chars) | `db/migrations/0NNN_*.sql` |
| Cross-module via published interface only | **Yes** — `tokens` publishes `domain/port.go`; SP-2's `render` → `tokens` edge goes through that port | `tokens/domain/port.go` exposes `DictionaryReader` interface; render will import the interface, not the repo or tables; `tokens` will never import `render` | `domain/port.go` (provider) / `application/ports.go` (consumer) pattern |

No invariant violation. No AS-1.

---

## 4. Capability wiring

SP-1 adds two capabilities: `token.view` and `token_dictionary.manage`.

Walk of the 10 ordered touchpoints:

1. **const + `validCapabilities`** — Add `CapTokenView Capability = "token.view"` and `CapTokenDictionaryManage Capability = "token_dictionary.manage"` to the const block in `internal/modules/iam/domain/model.go:90`, and append both to `validCapabilities` at line ~134.

2. **scope classify** — Both capabilities are `ScopeTenant` (dictionary is tenant-wide; no per-area scoping needed). Classify in `internal/modules/iam/domain/capability_scope.go:36`.

3. **tier-1 route→cap rule** — Map each new route to its capability in `apps/api/cmd/metaldocs-api/permissions.go`. `GET /tokens/*` → `CapTokenView`; `POST/PUT/DELETE /tokens/*` → `CapTokenDictionaryManage`. **Omitting this is silent privilege escalation** (default visibility is `VisibilitySessionRequired`, reachable by any authenticated user).

4. **tier-2 in-tx enforcement** — Call `authz.Require(ctx, tx, CapTokenView, area)` (read ops) and `authz.Require(ctx, tx, CapTokenDictionaryManage, area)` (mutating ops) inside the business tx, after `authz.SeedTxIdentity`. Follow the pattern at `internal/modules/templates/application/create.go:63`.

5. **seed grants** — Add grants for `token.view` and `token_dictionary.manage` to the appropriate roles in `db/reference-data/0001_product_reference_data.sql:17`. Likely: `token.view` granted broadly (authors, operators); `token_dictionary.manage` granted to `system_admin` and a designated template-manager role.

6. **DB tripwire** — The format constraint `ck_cap_format` and legacy constraint `ck_cap_not_legacy` in `db/baseline/0001_current_schema.sql` accept any `<noun>.<verb>` pattern — both new cap names conform. No schema change needed to the constraints themselves.

7. **guard tests stay green** — `TestEveryCapabilityClassified` and `TestAreaGradeCapabilitySet` in `internal/modules/iam/domain/capability_scope_test.go` must pass. They will once both caps are classified in step 2.

8. **bump `TestCapabilityRegistrySize`** — Current count (targeted-verified from `internal/modules/iam/domain/model_test.go:91`): `const want = 31`. After adding `token.view` + `token_dictionary.manage`: new `const want = 33`. This is the mandatory manual edit.

9. **CI capability-coherence (5-surface)** — const / classify / tier-1 / seed / test surfaces must agree; governed by REQ-AUTHZ-5 in `wiki/architecture/backend-target-architecture.md`. All 5 surfaces updated in steps 1–7 above satisfies this.

10. **H-PRE-1** — Do not call `authz.Require` inside a lock-holding atomic tx. The dictionary CRUD has no advisory locks, so H-PRE-1 is satisfied by design. Confirm no locks are introduced.

**Locked constraint from §4:** `TestCapabilityRegistrySize` must be bumped from 31 → 33.

---

## 5. Module wiring

SP-1 births the `tokens` module. Walk of the 11 ordered birth steps:

1. **Folders** — Create `internal/modules/tokens/{api,application,domain,delivery/http,infrastructure}/`.

2. **Domain** — Define `DictionaryEntry` entity (id, tenant_id, name, value, created_at, updated_at, created_by). Define `DictionaryReader` and `DictionaryWriter` interfaces in `domain/port.go` (the provider ports this module publishes — `DictionaryReader` is what SP-2's render will consume).

3. **Application** — `DictionaryService` application service in `application/service.go`; `ports.go` declares the consumer ports (only `TxRunner` and `AuditRecorder`). Orchestrates: `authz.SeedTxIdentity` → `authz.Require` → repo call → `audit.RecordTx`. Owns the tx boundary via `TxRunner`.

4. **Infrastructure** — `PostgresDictionaryRepository` in `infrastructure/repository.go`. Touches **only** `token_dictionary_entries`. Calls `authz.SeedTxIdentity` and `authz.Require` in-tx. Never reads another module's tables.

5. **Delivery** — `Handler` in `delivery/http/handler.go` + `RegisterRoutes(mux)`. Thin: decode via `contracts.Decode` → call `DictionaryService` → write response via `problem.Write` / `httpresponse`.

6. **api codegen** — Add `tokens` entry to `api/cfg.yaml` (`include-tags: [tokens]`); add `gen.go` in `internal/modules/tokens/api/`; run `go generate ./internal/modules/tokens/api/...` after spec is written.

7. **OpenAPI** — Add `tags: [{name: tokens}]` entry in `api/openapi/v1/openapi.yaml`. Tag **every** new route with `tokens`. Contract-first: spec written before any Go routes.

8. **`module.go` (optional)** — Recommended: `New(Dependencies) *Module` constructor; panic on nil deps (fail fast at composition). Follow `taxonomy`'s shape.

9. **Composition root** — Wire `tokens.New(...)` in `apps/api/cmd/metaldocs-api/main.go`. No worker or jobs binary needed (SP-1 has no async consumers or recurring janitors).

10. **Migration** — `db/migrations/0NNN_token_dictionary_entries.sql`: create `token_dictionary_entries` with `id UUID PK`, `tenant_id UUID NOT NULL REFERENCES tenants(id)`, `name TEXT NOT NULL`, `value TEXT NOT NULL`, `created_by UUID NOT NULL`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`; add `UNIQUE (tenant_id, name)` and a `CHECK` constraint on `name` format.

11. **Docs** — `wiki/modules/tokens.md` (12-section structure); `wiki/modules/tokens-tech-debt.md`; entry in `wiki/modules/index.md`.

---

## 6. Frameworks to reuse, not reinvent

| Platform primitive | SP-1 use | Confirm reuse |
|---|---|---|
| `TxRunner` (`Do`/`DoReadOnly`) — `internal/platform/db/runner.go:21` | Every `DictionaryService` method runs through `TxRunner.Do` or `DoReadOnly` | Yes — service depends on tx port, not `*sql.DB` |
| `tenant.FromContext` — `internal/platform/tenant/context.go:27` | Read current tenant in delivery layer before invoking service | Yes — never thread tenant ID by hand |
| `authz.SeedTxIdentity` — `internal/modules/iam/authz/context.go:58` | First call in every business tx | Yes |
| `authz.Require(ctx, tx, cap, area)` — pattern `templates/application/create.go:63` | Tier-2 in-tx cap check on every op | Yes |
| `problem.New`/`problem.Write` — `internal/platform/problem/problem.go:76` | All error responses in delivery layer | Yes — no bare `http.Error` |
| `httpresponse` helpers — `internal/platform/httpresponse/` | Success responses in delivery layer | Yes |
| `audit.NewEvent`/`RecordTx` — `internal/modules/audit/` | Record every mutating dictionary operation inside the business tx | Yes — `RecordTx` to stay in the same tx as the state write |
| Outbox repo — `internal/modules/render/fanout/staging_outbox.go:29` | Not used in SP-1 — no external side effects | N/A (SP-1 has no network side effects) |
| `contracts.Decode` — `internal/platform/contracts/` | Decode all incoming JSON request bodies in the delivery handler | Yes — rejects unknown fields |
| `testdb.Open` + factory builders — `tests/integration/testdb/` | All integration tests | Yes — `Open(t)`, `SeedWithCaps`, `Qualified` |

No hand-rolled equivalents. All cross-cutting concerns are covered by existing platform primitives.

---

## 7. Contract & data

**OpenAPI-first:**
- Add a `tokens` tag and the following routes to `api/openapi/v1/openapi.yaml` before writing any Go:
  - `GET /api/v1/tokens` — list dictionary entries (requires `token.view`)
  - `POST /api/v1/tokens` — create entry (requires `token_dictionary.manage`)
  - `GET /api/v1/tokens/{id}` — get single entry (requires `token.view`)
  - `PUT /api/v1/tokens/{id}` — update entry (requires `token_dictionary.manage`)
  - `DELETE /api/v1/tokens/{id}` — delete entry (requires `token_dictionary.manage`)
- New DTOs: `TokenDictionaryEntry` (response), `CreateTokenDictionaryEntryRequest`, `UpdateTokenDictionaryEntryRequest`. All generated via `oapi-codegen`; hand-crafted DTOs are a defect.

**Migration:**
- `db/migrations/0NNN_token_dictionary_entries.sql` (next sequence number after current highest in `db/migrations/`).
- Table: `token_dictionary_entries` — columns listed in §5 step 10.
- Constraints: `UNIQUE (tenant_id, name)`, format `CHECK` on `name` (e.g. `name ~ '^[a-z][a-z0-9_]*$'`).
- No existing table is modified in SP-1.

**Destructive change?** No. New table, new routes only. Zero impact on existing contracts. No expand/contract needed.

---

## 8. Test & QA plan

**Canonical framework:** `testdb` integration factory (`tests/integration/testdb/`); `Open(t)`, factory builders, `SeedWithCaps`; tag with `//go:build integration`; R1–R4 discipline enforced by `scripts/check-test-discipline.sh`.

**Which of the 6 QA gates apply (new module):**

| Gate | Applies? | Notes |
|------|----------|-------|
| Contract (OpenAPI / generated DTOs match runtime) | Yes | `go generate` passes; oapi-codegen output matches spec |
| AuthZ (capabilities gate all routes; no role-based shortcut) | Yes | Unit: both caps classified + in-registry; integration: `SeedWithCaps` grants cap → 200, without → 403; tier-1 omission → 401/403 |
| Multi-tenant isolation (no cross-tenant data leak) | Yes | Integration: tenant A cannot read/modify tenant B's entries (returns 404, not 403) |
| Async/idempotency | N/A | SP-1 has no outbox consumers |
| DB-invariant | Yes | Integration: duplicate `(tenant_id, name)` insert returns 409 (constraint enforced); invalid name format rejected |
| Docs | Yes | `wiki/modules/tokens.md`, `tokens-tech-debt.md`, `wiki/modules/index.md` entry |

**Evidence shape** (to report before saying done):
- `go build ./...` — passes
- `go test ./...` — passes (including `TestCapabilityRegistrySize` with bumped count 33)
- `go test -tags=integration ./...` — passes (authz, multi-tenant, DB-invariant gates above)
- `.\scripts\check-system-runnable.ps1` — passes
- Review/QA disposition: milestone-validator verdict
- Bounded defers: SP-2 (render injection), SP-3 (UI), SP-4, SP-5

---

## 9. Docs / ADR

**Wiki docs (module work):**
- Create `wiki/modules/tokens.md` (12-section structure; exemplar: `wiki/modules/taxonomy.md`)
- Create `wiki/modules/tokens-tech-debt.md`
- Add entry in `wiki/modules/index.md`
- Refresh `Last verified` stamp in `wiki/modules/templates.md` (the computed catalog endpoint is in templates; it gains a neighbour)

**REQ IDs to cite:**
- `REQ-AUTHZ-5` — capability-coherence CI gate (const/classify/tier-1/seed/test all agree)
- `REQ-AUTHZ-1` / `REQ-AUTHZ-2` — two-tier PDP (all new routes wire tier-1 + tier-2)
- `REQ-MT-1` — multi-tenant: every new table has `tenant_id`
- `REQ-CONTRACT-1` — contract-first: spec before Go routes
(All REQ IDs from `wiki/architecture/backend-target-architecture.md`)

**ADR required? Yes.**
- SP-1 is the point where the token catalog stops being fixed (ADR 0008's "fixed catalog, no author-defined tokens" stance is superseded). The spec (§7) explicitly states a formal ADR superseding 0008 should be written when SP-1 lands.
- New ADR number: next sequential in `wiki/decisions/` (check current highest; at time of analysis the ADR numbering is in the 0040s range based on ADR 0043 mentioned in test comments). The new ADR supersedes ADR 0008, preserves its correct half (computed tokens are server-resolved and secure), and documents the tenant dictionary extension.

---

## 10. Verdict & locked constraints

**Verdict:** 🟡 Yellow — proceed to brainstorming; the ADR requirement and `TestCapabilityRegistrySize` bump are flagged as locked constraints that must be honored in the design.

**Open hard-stops:** None. No AS-1, no AS-2, no AS-3.

**Locked constraints handed to brainstorming:**

1. **New module `tokens`** — a full bounded-context module under `internal/modules/tokens/`, not a feature inside `templates` or `iam`. All 11 birth steps in §5 must be followed in order.

2. **Two new capabilities** — `token.view` (ScopeTenant) and `token_dictionary.manage` (ScopeTenant). Full 10-touchpoint wiring (§4). Bump `TestCapabilityRegistrySize` from 31 → 33.

3. **ADR superseding ADR 0008** — required before or at SP-1 merge. The catalog is no longer fixed; the ADR documents the extension and preserves the correct parts of 0008.

4. **SP-2 direction constraint** — when render gains the dictionary injection, the dependency is `render` → `tokens` only (through `tokens/domain/port.go` `DictionaryReader`). `tokens` must never import `render`. This is a locked module-boundary constraint.

5. **Contract-first** — `api/openapi/v1/openapi.yaml` is edited and `oapi-codegen` is run before any Go handler is written.

6. **`tenant_id` on every row** — `token_dictionary_entries` has `tenant_id NOT NULL`; every query is predicated on it; cross-tenant access returns 404.

7. **`audit.RecordTx` in every mutating op** — audit inside the business tx, not after.

8. **`testdb` integration factory only** — no bespoke test harness; R1–R4 discipline enforced.

9. **No render/UI in SP-1** — scope boundary. SP-2 (render injection) and SP-3 (UI) are separate increments.
