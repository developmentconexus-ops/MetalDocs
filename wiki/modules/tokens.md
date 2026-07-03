# Module: tokens

> Living architecture doc. Arc42 (12 sections) + C4 (Context / Container) Mermaid diagrams + ADR links.

**Last verified:** 2026-06-29 (ADR 0050 symmetry — computed tokens now publish from `render/domain.ComputedCatalog()`, mirroring `tokens/domain.DictionaryReader`; both catalog kinds are architecturally symmetric: each publishes from its owning module’s domain layer, composed at the editor palette; prior 2026-06-29: read-write-tx authz fix — Get/GetByName/List switched `DoReadOnly`→`Do` so the system_admin bypass audit (ADR 0022 F8) can INSERT; new api-lint `authz-require-rw-tx` rule; SP-3 UI delivered — capability-gated CRUD screen at templates/tokens; see SP-2 for prior history — creation-time dictionary token pinning; reserved-name guard; `PHDictionary` placeholder type in templates; migration 0249 widens `source` CHECK; ADR 0049) | **Prior:** 2026-06-28 (SP-1 initial publish — module fully implemented: migration 0248, IAM capabilities, OpenAPI + oapi-codegen, domain/application/infrastructure/delivery layers, module assembly, integration tests) | **Owner:** unassigned | **Status:** active (SP-1 + SP-2 + SP-3 feature-complete) | **Maturity:** L3

> **Key files:**
> - `internal/modules/tokens/domain/entry.go` — `Entry` aggregate, `NewEntry`, `ApplyUpdate`, `nameRe`, sentinel errors
> - `internal/modules/tokens/domain/port.go` — `Repository` (storage port), `DictionaryReader` (published SP-2 provider port)
> - `internal/modules/tokens/application/service.go` — `Service` (tx owner, authz + audit); `var _ domain.DictionaryReader = (*Service)(nil)` compile-time proof
> - `internal/modules/tokens/application/ports.go` — `txRunner`, `auditRecorder` internal ports
> - `internal/modules/tokens/infrastructure/repository.go` — `PostgresRepository` (all ops via caller-supplied `*sql.Tx`)
> - `internal/modules/tokens/delivery/http/handler.go` — `Handler` implementing `tokensapi.ServerInterface`; RFC 9457 error mapping
> - `internal/modules/tokens/api/api.gen.go` — oapi-codegen generated types + `StdHTTP` router
> - `internal/modules/tokens/module.go` — `Module` struct + `New` composition root; exposes `Handler` + `Reader`
> - `apps/api/cmd/metaldocs-api/permissions.go:190-194` — tier-1 dispatcher entries (token.view on GET, token_dictionary.manage on POST/PUT/DELETE)
> - `db/migrations/0248_token_dictionary_entries.sql` — table + indexes + CHECK constraints + NULL-permissive RLS (ADR 0247 pattern)

---

## 1. Introduction & Goals

`internal/modules/tokens` owns the **per-tenant author-defined token dictionary**: a flat `name → value` constant table that makes tenant-specific strings (e.g., company name, legal entity, standard boilerplate) available as named tokens in documents. SP-1 delivered the CRUD dictionary with full authz + audit. SP-2 delivered creation-time pinning: dictionary values are resolved off-tx at document creation and pinned in `document_placeholder_values` with `source='dictionary'`; render transports pre-resolved values and does not merge catalogs.

This module introduces a **new class of token** alongside the computed-token catalog owned by `templates`. Dictionary tokens are tenant-governed, mutable at runtime, and require no template schema change to add. Computed tokens are system-defined and driven by template authoring. The two catalogs are complementary. **Collision is prevented by the reserved-name guard (SP-2 D4/D5):** a dictionary entry name cannot equal a native/computed resolver key; `PHDictionary` schema references are validated the same way. TD-1 (collision at render time) is closed — see `wiki/modules/tokens-tech-debt.md`. **Architectural symmetry (ADR 0050):** computed tokens now publish their catalog from `render/domain.ComputedCatalog()`, mirroring the `tokens/domain.DictionaryReader` port in this module. Both token kinds share the same structure: each publishes a catalog from its owning module’s domain layer; neither crosses into the other’s storage or resolvers. The two catalogs are composed client-side at the editor palette by `kind` (`‘computed’ | ‘dictionary’`). Promotion trigger: if a second backend consumer appears or the merge acquires non-trivial logic, lift the composition to a unified backend endpoint (see ADR 0050 §Deferred).

### 1.1 Requirements overview

- **Per-tenant dictionary CRUD** (`name`, `value`, `label`, optional `description`) with UUID primary key.
- **Name immutability post-create** — `name` is the render key; mutation is rejected at the application layer and never reaches the DB. Value, label, description are mutable.
- **Capability-gated** — `token.view` (ScopeTenant) for reads; `token_dictionary.manage` (ScopeTenant) for writes — both two-tier: tier-1 path-prefix dispatcher + tier-2 `authz.Require` inside tx + DB tripwire.
- **Audited writes** — every Create/Update/Delete records a `tokens.entry.created/updated/deleted` event in `metaldocs.audit_events` via `audit.Writer.RecordTx` inside the same transaction.
- **Published port for SP-2** — `domain.DictionaryReader` (`GetByName` + `List`, no tx coupling) is implemented by `application.Service` and exported from `Module.Reader`.
- **Token grammar is Node-owned** — `@metaldocs/shared-tokens` owns the leading-char rule and reserved-word list. Go does anti-corruption storage hygiene only (`^[A-Za-z0-9_]+$`, 1–64 chars). Grammar enforcement lives at the SP-3 UI edge; Go never re-parses token syntax.
- **Multi-tenant isolation** — `tenant_id UUID NOT NULL` on every row; NULL-permissive RLS per ADR 0027 pattern (migration `0248_token_dictionary_entries.sql`); no FK to the tenants table.

### 1.2 Quality Goals

| Rank | Goal | How verified |
|---|---|---|
| 1 | Multi-tenant isolation | `(tenant_id, name)` unique index; tenant always read from `tenant.FromContext` (session-bound, not client header); RLS NULL-permissive policy on `token_dictionary_entries` |
| 2 | Authz on every operation | Tier-1 dispatcher wired; `authz.Require` inside every tx, all via `Do` (read-WRITE — the F8 bypass audit may INSERT, so `DoReadOnly` is banned on authz-gated paths; see [authz-tiers.md](../concepts/authz-tiers.md) Common pitfalls); DB tripwire from `trg_require_cap_asserted` pattern |
| 3 | Audit completeness | `audit.Writer.RecordTx` called for every state-changing operation inside the same tx; reads not audited per convention |
| 4 | Published surface stability | `domain.DictionaryReader` is the only inter-module Go contract; it has no tx coupling so SP-2 callers never see the storage implementation |

### 1.3 Stakeholders

| Role | Expectation |
|---|---|
| Tenant admin / QMS operator | CRUD own token dictionary via `/api/v1/tokens`; use entries in controlled documents |
| documents / controlleddocuments (SP-2) | Consume `DictionaryReader.GetByName` / `List` off-tx at document creation to pin values into `document_placeholder_values` (`source='dictionary'`) |
| SP-3 UI team | Enforce full Node grammar at the editor edge; surface token panel for authors |
| Audit / compliance | Every write operation produces a dated, actor-attributed audit event |

---

## 2. Architecture Constraints

- Language / runtime: Go 1.25
- Persistence: Postgres; 1 owned table (`metaldocs.token_dictionary_entries`) created by forward migration `0248_token_dictionary_entries.sql`
- HTTP routing: oapi-codegen generated router (`tokensapi.HandlerWithOptions`); spec at `internal/modules/tokens/api/cfg.yaml` → `api.gen.go`; compile-time `ServerInterface` assertion in `api.gen.go`
- Error envelope: RFC 9457 `application/problem+json` via `httpresponse.WriteError` → `problem.Write`; domain-specific codes `ALREADY_EXISTS`, `NOT_FOUND`, `immutable_field`
- Authz: tier-1 path-prefix dispatcher (`permissions.go:190-194`); tier-2 `authz.Require` inside every `txRunner.Do` call (read-WRITE, including reads — the bypass audit may INSERT; `DoReadOnly` is banned on authz-gated paths, enforced by api-lint `authz-require-rw-tx`); no direct tier-3 tripwire row in this migration (DB CHECK constraints serve the same enforcement role)
- Tenant scoping: `tenant.FromContext` only — never client headers; panics/errors on missing context tenant per platform convention
- Token grammar: Node-owned (`@metaldocs/shared-tokens`); Go's `nameRe = ^[A-Za-z0-9_]+$` is anti-corruption hygiene, not grammar

---

## 3. System Scope & Context — module-scoped (C4 Level 1)

> System-level context lives in `wiki/diagrams/c4-context.md`. The diagram below is module-scoped: it shows the tokens module's external actors and the single owned Postgres table.

```mermaid
C4Context
    title System Context — tokens (module-scoped)
    Person(admin, "Tenant Admin / QMS Operator", "Web UI")
    System_Boundary(b1, "MetalDocs") {
        System(tokens, "tokens", "Per-tenant name→value dictionary")
        System_Ext(docs, "documents / controlleddocuments (SP-2)", "Resolves DictionaryReader off-tx at creation; pins values in document_placeholder_values")
        System_Ext(audit, "audit module", "Receives RecordTx calls for all writes")
    }
    System_Ext(pg, "Postgres", "metaldocs.token_dictionary_entries")
    Rel(admin, tokens, "HTTP /api/v1/tokens")
    Rel(tokens, pg, "SQL via *sql.Tx (txRunner)")
    Rel(tokens, audit, "audit.Writer.RecordTx in-tx")
    Rel(docs, tokens, "Go: domain.DictionaryReader.GetByName / List (off-tx, creation path only)")
```

### 3.1 Business Context

Authors of controlled documents need to embed tenant-specific constant values (company name, legal entity, standard clauses) without editing every template. The token dictionary is that namespace: the admin defines `{COMPANY_NAME} = "Acme Metalurgia"` once; SP-2 pins the value at document creation time so each document revision carries its own immutable snapshot. Render receives pre-resolved values — it does not call back into this module. The naming convention stays with the Node grammar; this module is the backend store.

### 3.2 Technical Context

Inbound interfaces (HTTP): 5 routes under `/api/v1/tokens` (§5.3).

Inbound interfaces (Go):
- `tokens.New(deps)` — composition root; returns `Module{Handler, Reader}`
- `Module.RegisterRoutes(mux)` — mounts `tokensapi.HandlerWithOptions` onto the platform mux

Outbound interfaces:
- DB: `metaldocs.token_dictionary_entries` (1 owned table; all ops within caller-supplied `*sql.Tx`)
- Go: `internal/platform/db.TxRunner` (tx boundary), `internal/modules/iam/authz` (Require + SeedTxIdentity), `internal/modules/audit/domain.Writer` (RecordTx), `internal/platform/authn` (UserIDFromContext), `internal/platform/tenant` (FromContext), `internal/platform/httpresponse` + `internal/platform/problem` (RFC 9457 envelope)

No outbound calls to other business modules. This module is a pure provider.

---

## 4. Solution Strategy

- **Tx-owner pattern.** The application service (`application.Service`) owns the transaction boundary via `db.TxRunner.Do` (read-WRITE for all operations, reads included — `authz.Require` may INSERT a bypass audit, so `DoReadOnly` is not used here). The repository receives the `*sql.Tx` as a parameter and never opens its own connection or tx. This is consistent with the MetalDocs standard for modules introduced after Wave 2.
- **SeedTxIdentity → authz.Require sequence.** Every tx starts with `authz.SeedTxIdentity(ctx, tx, tenantID, actorID)` (sets tx-local GUCs) followed immediately by `authz.Require(ctx, tx, capability, "tenant")`. Both read and write paths enforce capability before touching data.
- **Reads are not business-audited, but still run read-WRITE.** `Get`, `GetByName`, `List` run under `txRunner.Do` (not `DoReadOnly`): no business audit event is emitted for reads, but `authz.Require` may INSERT a system_admin bypass audit (ADR 0022 F8), which a read-only tx cannot do. `DoReadOnly` is therefore banned on authz-gated paths (enforced by api-lint `authz-require-rw-tx`).
- **Name is immutable.** `name` is the render key. `Update` fetches the existing entry first; if the caller sends a non-empty `name` that differs from the stored name, `domain.ErrImmutableName` is returned (422 `immutable_field`). The DB has no trigger for this — the app-layer check is the sole enforcement.
- **DictionaryReader as the published port.** The application service implements `domain.DictionaryReader` (compile-time proof: `var _ domain.DictionaryReader = (*Service)(nil)`). The module exports `Module.Reader domain.DictionaryReader` so SP-2 callers can inject only this narrow interface without depending on the full service.
- **Anti-corruption storage hygiene.** `nameRe = ^[A-Za-z0-9_]+$` in the domain is paired with an identical CHECK constraint in the migration. The DB rejects any row the app layer might wrongly admit. The canonical leading-char rule (`^[A-Za-z_]`) lives in Node and is a superset constraint applied at the UI edge.

---

## 5. Building Block View — module-scoped (C4 Level 2 — Container)

> System-level container topology lives in `wiki/diagrams/c4-container-backend.md`. The diagram below decomposes the internal Go packages of tokens.

### 5.1 Whitebox — tokens

```mermaid
C4Container
    title Container View — tokens (module-internal packages)
    Container(http, "HTTP Handler", "Go (http.ServeMux)", "5 routes: /api/v1/tokens[/{id}]")
    Container(svc, "Service", "Go", "Service: tx owner, authz, audit; implements DictionaryReader")
    Container(domain, "Domain", "Go", "Entry aggregate · NewEntry · ApplyUpdate · Repository port · DictionaryReader port · sentinel errors")
    Container(repo, "Repository", "Go + database/sql", "PostgresRepository (tx-parameterised)")
    ContainerDb(db, "metaldocs.token_dictionary_entries", "Postgres", "name→value rows, per-tenant, RLS NULL-permissive")
    System_Ext(iam, "iam/authz", "Go", "authz.Require + SeedTxIdentity")
    System_Ext(audit, "audit/domain", "Go", "Writer.RecordTx")
    Rel(http, svc, "calls TokenService interface")
    Rel(svc, domain, "uses Entry, NewEntry, ApplyUpdate")
    Rel(svc, repo, "calls domain.Repository interface with *sql.Tx")
    Rel(svc, iam, "SeedTxIdentity + Require inside every tx")
    Rel(svc, audit, "RecordTx inside write tx")
    Rel(repo, db, "SQL via *sql.Tx")
```

### 5.2 Public surface

| File | Symbol | Kind | Purpose |
|---|---|---|---|
| `domain/entry.go` | `Entry` | struct | Token dictionary entry aggregate (`ID, TenantID, Name, Value, Label, Description, CreatedBy, UpdatedBy, CreatedAt, UpdatedAt`) |
| `domain/entry.go` | `NewEntry(NewEntryInput)` | func | Validates input; returns `*Entry` or `*ValidationError` |
| `domain/entry.go` | `(*Entry).ApplyUpdate` | method | Validates value/label/description; mutates in-place; never touches `Name` |
| `domain/entry.go` | `ErrNotFound · ErrImmutableName` | sentinels | Storage + immutability errors |
| `domain/entry.go` | `ValidationError` | struct | Field-level validation failure (maps to 422) |
| `domain/port.go` | `Repository` | iface | Storage port: `Create · Update · Delete · GetByID · GetByName · List` (all take `*sql.Tx`) |
| `domain/port.go` | `DictionaryReader` | iface | **Published SP-2 provider port**: `GetByName(ctx, tenantID, name) · List(ctx, tenantID)` (no tx) |
| `application/service.go` | `Service` | struct | Tx owner; implements `TokenService` + `domain.DictionaryReader` |
| `application/service.go` | `NewService` | func | Production constructor (panics on nil deps); pins `authz.Require` + `authz.SeedTxIdentity` |
| `application/service.go` | `NewServiceForTest` | func | Test constructor with injectable authz stubs |
| `application/service.go` | `CreateCommand · UpdateCommand` | structs | Service input DTOs |
| `infrastructure/repository.go` | `PostgresRepository` | struct | `domain.Repository` implementation; all methods receive `*sql.Tx` |
| `delivery/http/handler.go` | `Handler` | struct | `tokensapi.ServerInterface` implementation |
| `delivery/http/handler.go` | `Handler.RegisterRoutes` | method | Mounts 5 routes via `tokensapi.HandlerWithOptions` with `/api/v1` base |
| `module.go` | `Module · Dependencies` | structs | Composition root |
| `module.go` | `Module.Handler` | field | `*tokenshttp.Handler` — HTTP delivery |
| `module.go` | `Module.Reader` | field | `domain.DictionaryReader` — published for SP-2 consumers |

### 5.3 HTTP operations

| Method | Path | OperationID | Handler | Tier-1 cap | Tier-2 cap |
|---|---|---|---|---|---|
| GET | `/api/v1/tokens` | `listTokens` | `h.ListTokens` | `token.view` | `CapTokenView` |
| POST | `/api/v1/tokens` | `createToken` | `h.CreateToken` | `token_dictionary.manage` | `CapTokenDictionaryManage` |
| GET | `/api/v1/tokens/{id}` | `getToken` | `h.GetToken` | `token.view` | `CapTokenView` |
| PUT | `/api/v1/tokens/{id}` | `updateToken` | `h.UpdateToken` | `token_dictionary.manage` | `CapTokenDictionaryManage` |
| DELETE | `/api/v1/tokens/{id}` | `deleteToken` | `h.DeleteToken` | `token_dictionary.manage` | `CapTokenDictionaryManage` |

All 5 routes aligned: spec `api/openapi/v1/openapi.yaml` → `internal/modules/tokens/api/api.gen.go` → handler.

---

## 6. Runtime View (selected scenarios)

### 6.1 listTokens — read path

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant M as Middleware chain
    participant H as Handler.ListTokens
    participant S as Service.List (Do, read-WRITE)
    participant R as PostgresRepository.List
    participant DB as token_dictionary_entries

    C->>M: GET /api/v1/tokens
    M->>M: authn (session) + tier-1 cap check (token.view)
    M->>H: pass
    H->>H: tenant.FromContext → tenantID
    H->>S: List(ctx, tenantID)
    S->>S: Do (read-WRITE): SeedTxIdentity → authz.Require(token.view)
    S->>R: List(ctx, tx, tenantID)
    R->>DB: SELECT ... WHERE tenant_id=$1 ORDER BY name
    DB-->>R: rows
    R-->>S: []Entry
    S-->>H: []Entry
    H-->>C: 200 {"items":[...]}
```

No business write; no business audit. The tx runs read-WRITE (`Do`, not `DoReadOnly`) because tier-2 `authz.Require(CapTokenView)` — still called inside the tx as the belt-and-suspenders pattern — may INSERT a system_admin bypass audit (ADR 0022 F8), which a read-only tx cannot do.

### 6.2 createToken — write path

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant M as Middleware chain
    participant H as Handler.CreateToken
    participant S as Service.Create (Do)
    participant R as PostgresRepository.Create
    participant A as audit.Writer.RecordTx
    participant DB as token_dictionary_entries

    C->>M: POST /api/v1/tokens {name,value,label}
    M->>M: authn + tier-1 cap check (token_dictionary.manage)
    M->>H: pass
    H->>H: JSON decode body
    H->>H: tenant.FromContext + authn.UserIDFromContext
    H->>S: Create(ctx, CreateCommand{TenantID, ActorID, ...})
    S->>S: domain.NewEntry — validate name/value/label/description
    S->>S: Do(tx): SeedTxIdentity → authz.Require(token_dictionary.manage)
    S->>R: Create(ctx, tx, entry)
    R->>DB: INSERT INTO token_dictionary_entries ... RETURNING *
    DB-->>R: inserted row
    R-->>S: *Entry
    S->>A: RecordTx(ctx, tx, tokens.entry.created event)
    A->>DB: INSERT INTO audit_events ...
    S->>S: tx.Commit
    S-->>H: *Entry
    H-->>C: 201 {TokenDictionaryEntry}
```

| Condition | HTTP | Code |
|---|---|---|
| Name already exists (tenant) | 409 | `ALREADY_EXISTS` |
| Validation failure (name/value/label) | 422 | `VALIDATION_ERROR` |
| DB CHECK violated | 422 | `VALIDATION_ERROR` |
| Missing/insufficient capability | 403 | (platform authz) |
| Internal error | 500 | `INTERNAL_ERROR` |

### 6.3 updateToken — name-immutability path

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler.UpdateToken
    participant S as Service.Update (Do)
    participant R as PostgresRepository
    participant DB as token_dictionary_entries

    C->>H: PUT /api/v1/tokens/{id} {name:"NEW", value:...}
    H->>S: Update(ctx, UpdateCommand{..., Name:"NEW"})
    S->>S: Do(tx): SeedTxIdentity → authz.Require
    S->>R: GetByID(ctx, tx, tenantID, id)
    R->>DB: SELECT WHERE tenant_id=$1 AND id=$2
    DB-->>R: existing Entry (name="ORIGINAL")
    S->>S: cmd.Name != "" && cmd.Name != existing.Name → ErrImmutableName
    S-->>H: ErrImmutableName
    H-->>C: 422 immutable_field "name is immutable after creation"
```

---

## 7. Deployment View

- Binary: `apps/api/cmd/metaldocs-api` (stateless MetalDocs API)
- Schema: `metaldocs.token_dictionary_entries` created by `db/migrations/0248_token_dictionary_entries.sql`; no dependency on the curated baseline schema (module added after baseline freeze)
- No module-specific env vars or config keys
- Tenant scoping is session-driven via `tenant.FromContext`; no fallback to `DevTenantID` at this layer

---

## 8. Cross-cutting Concepts

### 8.1 Authentication & Authorization

**Tier 1 (HTTP edge):** `apps/api/cmd/metaldocs-api/permissions.go:190-194` — four dispatcher entries:
- `GET /api/v1/tokens` (prefix) → `CapTokenView`
- `POST /api/v1/tokens` (exact) → `CapTokenDictionaryManage`
- `PUT /api/v1/tokens` (prefix) → `CapTokenDictionaryManage`
- `DELETE /api/v1/tokens` (prefix) → `CapTokenDictionaryManage`

**Tier 2 (in-tx):** `authz.SeedTxIdentity` + `authz.Require` called at the top of every `txRunner.Do` callback in `application/service.go` (read and write paths alike — all `Do`, never `DoReadOnly`, since the bypass audit may INSERT). The seam is injectable (`authzRequireFunc` / `seedTxFunc`) so unit tests run without a real DB.

**Postgres tripwire:** The tokens migration does not attach `trg_require_cap_asserted` (the trigger requires a `capabilities_asserted` column which is not part of the standard table schema). DB CHECK constraints (`CHECK (name ~ '^[A-Za-z0-9_]+$')`, `CHECK (length(name) BETWEEN 1 AND 64)`, etc.) enforce data integrity as the last line. The `trg_require_cap_asserted` pattern can be added in a follow-up migration if required.

See `wiki/decisions/0007-two-tier-authz.md`. Governed by ADR 0022 (capabilities, never roles). Capabilities are registered in `internal/modules/iam/domain/capability.go` (`CapTokenView`, `CapTokenDictionaryManage`) and granted via IAM reference data.

### 8.2 Tenant scoping

Tenant always read from `tenant.FromContext(r.Context())`. The middleware chain sets this from the authenticated session before the handler fires. The handler passes `tenantID` explicitly into every service command and read call. The repository uses `tenant_id = $N` predicates on all queries — there is no global scan. RLS policy on `token_dictionary_entries` is NULL-permissive (ADR 0247 pattern): the policy exists but does not block `system_admin` operations that run before the tenant GUC is set.

### 8.3 Error envelope

RFC 9457 `application/problem+json` via `internal/platform/httpresponse.WriteError` → `internal/platform/problem.Write`. Domain-specific codes in `delivery/http/handler.go`:
- `ALREADY_EXISTS` — PG `23505` unique violation on `(tenant_id, name)` index
- `NOT_FOUND` — `domain.ErrNotFound` (cross-tenant ID → same 404, never 403)
- `immutable_field` — `domain.ErrImmutableName`
- `VALIDATION_ERROR` — `*domain.ValidationError` or PG `23514` CHECK violation

### 8.4 Idempotency

No `Idempotency-Key` handling. `POST /tokens` with a duplicate `name` returns 409 `ALREADY_EXISTS`. Callers should use `GET /tokens` to discover existing entries before attempting create if idempotency matters.

### 8.5 Pagination

No pagination. `List` returns all entries for the tenant ordered by `name`. Token dictionaries are expected to remain small (< a few hundred entries per tenant). Pagination can be added as a non-breaking query parameter in a future SP.

### 8.6 Logging & Observability

`slog.Error` is emitted for unexpected errors in the handler (`"tokens: list"`, `"tokens: handler error"`). No structured observability beyond the standard platform trace middleware. Audit events for writes are the primary traceability surface.

### 8.7 Concurrency / Transactions

All operations run inside a single `*sql.Tx` supplied by `db.TxRunner.Do` (read-WRITE for every operation — `DoReadOnly` is not used on authz-gated paths). The repository never opens its own connection. The `Update` path fetches the existing entry inside the same tx before applying the update — no optimistic lock (last-write-wins for concurrent updates to the same entry). Duplicate `name` conflicts are serialisation-safe via the unique index `ux_token_dictionary_tenant_name`.

### 8.8 Token grammar boundary

Node (`@metaldocs/shared-tokens`) owns the canonical token grammar:
- Leading-char rule: `^[A-Za-z_]` (first char must be letter or underscore)
- Reserved words: defined in `packages/shared-tokens/src/grammar.ts`
- Dotted paths (e.g. `{a.b}`) and hyphenated names: rejected at UI edge

Go's `nameRe = ^[A-Za-z0-9_]+$` is anti-corruption **storage hygiene** only — it rejects characters the DB CHECK also rejects, but it does not enforce the leading-char rule. The leading-char constraint is applied at the SP-3 UI edge before the name reaches the API. This separation means Go's storage rules can never drift from Node's grammar; Go simply keeps the DB clean.

Go never parses token syntax (this is a binding invariant from `wiki/concepts/token-syntax.md`).

### 8.9 Cross-module data contracts

- **documents / controlleddocuments (SP-2)** consume `domain.DictionaryReader` (`GetByName` + `List`) off-tx at document creation to pin dictionary values into `document_placeholder_values`. The interface is published via `Module.Reader`. No direct SQL access to `token_dictionary_entries` from outside this module — ever. A composition-root adapter bridges `domain.DictionaryReader` → the consuming modules' `DictionaryValueReader` port (no cross-module type import).
- **audit module** — `audit.Writer.RecordTx` called inside every write tx. `ResourceType = "token_dictionary_entry"`. Actions: `tokens.entry.created`, `tokens.entry.updated`, `tokens.entry.deleted`.
- No FK relationship to any other module's tables.

---

## 9. Architecture Decisions

| Decision | Link / Status |
|---|---|
| Two-tier authz | `wiki/decisions/0007-two-tier-authz.md` — tokens module conformant at tier-1 + tier-2; no per-row tripwire trigger (see §8.1) |
| Capabilities, never roles | ADR 0022 — `CapTokenView` + `CapTokenDictionaryManage`; never "admin can manage tokens" |
| Contract-first API | ADR 0012 — spec at `api/openapi/v1/openapi.yaml`; oapi-codegen generates router and types |
| Multi-tenant pooled DB | ADR 0023 — `tenant_id NOT NULL` + NULL-permissive RLS per ADR 0027 pattern |
| Transactional outbox | ADR 0024 — not applicable to token CRUD (no async side effects; audit write is in-tx, not outbox) |
| Node-owns-grammar / Go-does-hygiene | **ADR 0048** — `wiki/decisions/0048-tenant-token-dictionary.md` |
| Soft-archive via timestamp | ADR 0010 — not applicable; tokens support hard delete (DELETE route) |

---

## 10. Quality Requirements

| Goal | Scenario | Pass criteria | Current state |
|---|---|---|---|
| Multi-tenant isolation | Tenant A user requests `/api/v1/tokens` with tenant B's token ID | 404 (not 403) | PASSES — `WHERE tenant_id=$1 AND id=$2` predicate; cross-tenant ID resolves to not-found |
| Authz: read gate | User without `token.view` calls GET `/tokens` | 403 | PASSES — tier-1 dispatcher + tier-2 `authz.Require(CapTokenView)` |
| Authz: write gate | User without `token_dictionary.manage` calls POST `/tokens` | 403 | PASSES — tier-1 dispatcher + tier-2 `authz.Require(CapTokenDictionaryManage)` |
| Duplicate name rejected | POST `/tokens` with name already in use for tenant | 409 `ALREADY_EXISTS` | PASSES — unique index `ux_token_dictionary_tenant_name` → PG `23505` → handler maps to 409 |
| Name immutability | PUT `/tokens/{id}` with changed `name` field | 422 `immutable_field` | PASSES — `application.Service.Update` checks before repo call |
| Audit on writes | Create/Update/Delete emit `tokens.entry.*` events | `audit_events` row with correct action + actor + tenant | PASSES — `record()` called inside every write tx; integration tests cover this |
| Migration discipline | New table appended as forward migration | file `0248_*` exists; does not modify curated baseline | PASSES — `db/migrations/0248_token_dictionary_entries.sql` |

---

## 11. Risks & Technical Debt

Pointer-only. Body in `wiki/modules/tokens-tech-debt.md`. Severity rubric: see that file.

- Critical: 0
- Major: 1 (TD-1: ~~closed~~ resolved by SP-2 reserved-name guard + creation-time pinning; see ADR 0049 — all rows counted by tally including closed)
- Minor: 1 (TD-2: ~~closed~~ strictjson promotion shim, resolved ARC-03 2026-07-02)

TD-1 closed: the render-time catalog merge problem is moot because dictionary values are pinned at creation time and the reserved-name guard (D4/D5) prevents naming collisions before any document is created. See `wiki/modules/tokens-tech-debt.md`.

---

## 12. Glossary

| Term | Definition |
|---|---|
| `token_dictionary_entries` | The single Postgres table owned by this module. Columns: `id UUID PK`, `tenant_id UUID NOT NULL`, `name TEXT NOT NULL`, `value TEXT NOT NULL`, `label TEXT NOT NULL`, `description TEXT NULL`, `created_by TEXT`, `updated_by TEXT`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`. Unique index on `(tenant_id, name)`. |
| `token.view` / `CapTokenView` | IAM capability required for all read operations. ScopeTenant. |
| `token_dictionary.manage` / `CapTokenDictionaryManage` | IAM capability required for create / update / delete. ScopeTenant. |
| `DictionaryReader` | The published provider port in `internal/modules/tokens/domain/port.go`. Consumed by the `documents` / `controlleddocuments` creation path (off-tx) at document creation to resolve and pin dictionary values. Two methods: `GetByName(ctx, tenantID, name)` and `List(ctx, tenantID)`. Does not expose the `*sql.Tx` coupling of `domain.Repository`. |
| Name immutability | `name` is the document-reference key. Once set on create it cannot change; `ErrImmutableName` (422) is returned on any PUT that tries to change it. `value`, `label`, `description` are freely mutable. |
| Anti-corruption storage hygiene | Go's `nameRe = ^[A-Za-z0-9_]+$` and length limits mirror the DB CHECK constraints. This is not grammar enforcement — the canonical grammar (leading-char rule, reserved words) is Node-owned. |
| SP-2 | Delivered: creation-time pinning. Dictionary values are resolved off-tx at document creation and seeded into `document_placeholder_values` with `source='dictionary'`. Render transports pre-resolved values. See ADR 0049. |
| SP-3 | Delivered: Token Dictionary Management UI (`frontend/apps/web/src/features/tokens/`). Capability-gated CRUD screen at `templates/tokens`, gated on `token.view` (route) and `token_dictionary.manage` (write actions). Consumes `GET/POST /tokens`, `GET/PUT/DELETE /tokens/{id}`. Reuses the templates placeholder-catalog query for the D4 collision check. Shipped 2026-06-29. Dictionary entries are also surfaced read-only in the **template-editor variable palette** (`Definido pela sua organização` section) via the unified `Token` catalog — see `wiki/modules/templates.md` "Editor variable palette" (ADR 0050). |
| NULL-permissive RLS | RLS policy pattern per ADR 0027. The `token_dictionary_entries` table has RLS enabled with a policy that permits rows where `tenant_id` matches the GUC or where the GUC is NULL. This allows system operations (migrations, bootstrap, health checks) that run before the tenant GUC is set. |

---

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Postgres unavailable | 500 on all token routes | Handler logs; `/healthz` | Restore Postgres; TxRunner will fail immediately on `BeginTx` |
| Duplicate name on create (concurrent) | 409 `ALREADY_EXISTS` on one of the concurrent requests | PG `23505` unique violation | Expected; caller retries with unique name or fetches existing |
| Name immutability violated | 422 `immutable_field` | `domain.ErrImmutableName` | Caller must not include a changed `name` in PUT body |
| DB CHECK violation (bad name chars) | 422 `VALIDATION_ERROR` | PG `23514` | Client bypassed UI grammar gate; fix-forward at UI edge |
| Token not found (cross-tenant probe) | 404 `NOT_FOUND` | `domain.ErrNotFound` | Correct — cross-tenant ID looks like not-found, never 403 |
| authz GUC not set (SeedTxIdentity skipped) | authz.Require returns error → 403 or 500 | Auth middleware or test harness misconfiguration | Ensure middleware chain seeds session before handler fires |
| SP-2 collision (dictionary vs computed token) | N/A — collision prevented by reserved-name guard (D4/D5) | Guard enforced at write-time and schema-save | TD-1 closed; see ADR 0049 |

## Cross-links

- Related ADRs: `wiki/decisions/0007-two-tier-authz.md` (two-tier pattern), `wiki/decisions/0022-capabilities-not-roles.md` (capability model), `wiki/decisions/0048-tenant-token-dictionary.md` (module ADR — SP-1), `wiki/decisions/0049-tenant-dictionary-token-substitution.md` (SP-2 ADR — creation-time pinning), `wiki/decisions/0050-computed-token-catalog-single-source.md` (ADR 0050 — symmetry: computed tokens publish from `render/domain`, mirroring `tokens/domain.DictionaryReader`)
- Related concepts: `wiki/concepts/token-syntax.md` (grammar boundary, Node ownership), `wiki/concepts/placeholders.md` (full placeholder concept), `wiki/concepts/authz-tiers.md`
- Cross-module: `wiki/modules/templates.md` (computed token catalog — complementary; `PHDictionary` type added SP-2), `wiki/modules/audit.md` (RecordTx sink), `wiki/modules/documents.md` (SP-2 creation-time pinning consumer), `wiki/modules/render-fanout.md` (render transport — receives pre-resolved values)
- Tech debt: `wiki/modules/tokens-tech-debt.md`
- OpenAPI spec: `api/openapi/v1/openapi.yaml`

## Changelog (this doc)

- 2026-06-29 — ADR 0050 symmetry: computed tokens now publish from `render/domain.ComputedCatalog()`, mirroring `tokens/domain.DictionaryReader`. §1 updated with architectural-symmetry note and deferred promotion trigger. `Last verified` bumped.
- 2026-06-28 — SP-2 delivered (ADR 0049): creation-time pinning; reserved-name guard; composition-root adapter; `PHDictionary` in templates; migration 0249 widens `source` CHECK; TD-1 closed. `DictionaryReader` doc-comment corrected (consumer = documents/controlleddocuments at creation, not render). Updated status, §1, §3 diagram, §8.9, §11, §12, cross-links.
- 2026-06-28 — Initial publish (SP-1). Module fully implemented: migration 0248, IAM capabilities, OpenAPI + oapi-codegen, domain/application/infrastructure/delivery layers, module assembly, integration tests. REQ-AUTHZ-1 (two-tier), REQ-AUTHZ-2 (capability scoping), REQ-AUTHZ-5 (DB tripwire — partial: CHECK constraints in place, `trg_require_cap_asserted` deferred), REQ-TEN-1 (tenant_id on every row + RLS), REQ-CONTRACT-1 (oapi-codegen generated surface).
