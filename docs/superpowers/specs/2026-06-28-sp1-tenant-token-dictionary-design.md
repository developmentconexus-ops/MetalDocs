# SP-1 — Tenant Token Dictionary (backend)

**Date:** 2026-06-28
**Parent:** `2026-06-27-template-tokens-north-star.md` (§5 SP-1)
**Orientation:** `docs/superpowers/analysis/2026-06-28-tokens-system-impact.md` (Yellow; no hard-stops)
**Scope:** New backend module `tokens`. Domain + tenant-scoped storage + repo + 2 IAM capabilities +
capability-gated CRUD + audit. **No** render, **no** freeze substitution, **no** UI.
**Supersedes:** ADR 0008 (fixed placeholder catalog) — formal ADR written when SP-1 lands (§10).

---

## 1. Goal

Give a tenant author-defined, reusable tokens (`{company_slogan}`, …): tenant-scoped `name → value`
constants, capability-governed, audited. This is north-star intent (2) ("author-defined reusable
tokens"). SP-1 builds the **backend** only; freeze substitution (SP-2), management UI (SP-3), and the
palette merge (SP-4) are later increments.

## 2. Boundaries (non-goals)

- No freeze-time substitution and no injection into `ResolvedValues` — SP-2.
- No tag-validation / computed-vs-dictionary collision reconciliation — SP-2 (see §8).
- No management UI — SP-3. No palette merge — SP-4.
- No render dependency in either direction at SP-1; the only future edge is `render → tokens` (SP-2).
- No token parsing in Go (binding invariant `wiki/concepts/token-syntax.md` §Node-owned-grammar).

## 3. Architecture & layering

New bounded-context module `internal/modules/tokens/` with the standard layout
(`{api, application, domain, delivery/http, infrastructure}`). Exemplars: `taxonomy` (smallest
complete module, has `module.go`), `templates`.

```
delivery/http (Handler, RegisterRoutes)         generated server iface (oapi-codegen)
        │  decode (contracts.Decode) → service → problem.Write / httpresponse
        ▼
application.Service  (Create/Get/List/Update/Delete; owns the tx boundary via TxRunner)
        │  authz.SeedTxIdentity → authz.Require → repo → audit.RecordTx
        ▼
infrastructure.PostgresRepository  (touches ONLY token_dictionary_entries; sets tenant GUC)
        ▲
domain  (Entry entity; port.go publishes DictionaryReader + the repo write port)
```

**Dependencies (all via published interfaces):** `platform/db` (`TxRunner`), `platform/tenant`,
`iam/authz`, `audit`. **`tokens` imports nothing from `render` or `templates`** — this preserves the
locked acyclic direction (`render → tokens` only, arriving in SP-2). No worker/jobs binary (CRUD is
synchronous; no async side effects).

## 4. Data model

Migration `db/migrations/0NNN_token_dictionary_entries.sql` (next sequence number at impl). New table,
no existing table modified.

| column | type | constraint |
|--------|------|------------|
| `id` | UUID | PK |
| `tenant_id` | UUID | NOT NULL, FK → `tenants(id)` |
| `name` | TEXT | NOT NULL; `CHECK (name ~ '^[A-Za-z0-9_]+$')`; `CHECK (char_length(name) BETWEEN 1 AND 64)` |
| `value` | TEXT | NOT NULL; `CHECK (char_length(value) BETWEEN 1 AND 4096)` |
| `label` | TEXT | NOT NULL; `CHECK (char_length(label) BETWEEN 1 AND 256)` |
| `description` | TEXT | NULL; `CHECK (description IS NULL OR char_length(description) <= 1024)` |
| `created_by` | UUID | NOT NULL |
| `updated_by` | UUID | NOT NULL |
| `created_at` | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| `updated_at` | TIMESTAMPTZ | NOT NULL DEFAULT now() |

- `UNIQUE (tenant_id, name)`.
- **The `name` CHECK is anti-corruption storage hygiene, NOT the token grammar.** It rejects the
  genuinely-corrupting set (`{}`, `.`, `-`, whitespace, unicode) so the server is safe at SP-1 when no
  UI exists yet. The *canonical* charset (`IDENT_RE = /^[A-Za-z_][A-Za-z0-9_]*$/`, incl. the
  leading-char rule) remains **Node-owned** (`@metaldocs/shared-tokens`) and is enforced at the edges
  (SP-3 UI, editor detection). Go never claims to be the grammar — see §7 and the decision record.
- **`name` is immutable after create.** Renaming a token would orphan every template body referencing
  the old `{name}` (`wiki/concepts/token-syntax.md`). To "rename": delete + create.

## 5. Domain & ports

`domain/`:
- `Entry{ ID, TenantID, Name, Value, Label, Description, CreatedBy, UpdatedBy, CreatedAt, UpdatedAt }`.
- `port.go` (provider ports this module publishes):
  - `DictionaryReader` — `GetByName(ctx, tenantID, name) (Entry, error)`, `List(ctx, tenantID) ([]Entry, error)`.
    **This is SP-2's consumption surface** (render reads dictionary values through it).
  - the repository write port (`Create`/`Update`/`Delete`/`GetByID`) consumed by the application service.

Pure domain — no SQL, no HTTP.

## 6. Application service & tx flow

`application.Service` with `Create / Get / List / Update / Delete`. `ports.go` declares consumer ports:
`TxRunner` and `audit.Recorder`.

Both caps are `ScopeTenant` (tenant-wide, not area-scoped), so the tier-2 check is the tenant-wide
`authz.Require` form — follow the call convention of an existing `ScopeTenant` capability's call site
(verify the exact signature/area-sentinel at impl; do not pass a real area code).

- **Mutating** (`Create`/`Update`/`Delete`):
  `TxRunner.Do` → `authz.SeedTxIdentity` → `authz.Require(ctx, tx, CapTokenDictionaryManage, …)`
  → repo write → `audit.RecordTx` (records actor, action, prior/next state inside the same tx).
- **Read** (`Get`/`List`):
  `TxRunner.DoReadOnly` → `authz.SeedTxIdentity` → `authz.Require(ctx, tx, CapTokenView, …)` → repo read.
- No advisory locks anywhere → **H-PRE-1 satisfied** (no authz-recording read inside a lock-holding tx).
- `Update` rejects any attempt to change `name` (immutable; §4).

## 7. Validation (Decision A — Node owns the grammar)

Go does **anti-corruption + membership** only; it does not re-implement the token grammar.

- `name`: non-empty, `[A-Za-z0-9_]`, ≤ 64 — app check (friendly) + DB CHECK (enforcement). Framed and
  documented as storage hygiene.
- `value`: non-empty, ≤ 4096. `label`: non-empty, ≤ 256. `description`: ≤ 1024 (nullable).
- Canonical token charset (leading-char rule, dotted/hyphenated rejection semantics) is **Node-owned**
  and applied at the edges (SP-3 UI, editor). SP-1's only client is tests; SP-3's UI will run
  `@metaldocs/shared-tokens` for the authoritative gate.
- **Why not full `IDENT_RE` in Go:** it would re-fork the just-unified grammar (binding invariant
  `token-syntax.md:65`), be the *unguarded* copy (no parity gate like `token-parity.test.ts`), and the
  orientation's own hand-copy already drifted (`^[a-z]…` vs canonical `^[A-Za-z_]…`). Charset is not
  Go's gate — membership is. Accepting full `IDENT_RE` would require an ADR; there is no payoff here.

## 8. Computed-key collision — deferred to SP-2

The authoritative registry is `computed ∪ dictionary`. A tenant creating a name equal to a computed
key (e.g. `{author}`) is reconciled at **freeze / tag-validation, which is SP-2's defined job**
(north-star §5 SP-2: "validate template tokens ⊆ registry"). SP-1 does **not** check computed keys,
because sourcing the computed-key list into `tokens` would either create a `tokens → render` edge
(risking a package cycle with SP-2's `render → tokens`), couple `tokens → templates`, or add a third
hardcoded copy of the computed keys (the catalog-duplication smell the orientation scoped out). No UI
until SP-3, so the footgun window is tests-only. **Logged as an explicit SP-2 responsibility** in
`tokens-tech-debt.md`.

## 9. Delivery & contract (OpenAPI-first)

Edit `api/openapi/v1/openapi.yaml` first (tag `tokens`); add `tokens` to `api/cfg.yaml`
(`include-tags`) + `gen.go`; `go generate ./internal/modules/tokens/api/...` before any handler. Every
route tagged `tokens`.

| route | capability | success |
|-------|-----------|---------|
| `GET /api/v1/tokens` | `token.view` | 200 list |
| `POST /api/v1/tokens` | `token_dictionary.manage` | 201 entry |
| `GET /api/v1/tokens/{id}` | `token.view` | 200 entry |
| `PUT /api/v1/tokens/{id}` | `token_dictionary.manage` | 200 entry (value/label/description only) |
| `DELETE /api/v1/tokens/{id}` | `token_dictionary.manage` | 204 |

DTOs (generated, never hand-written): `TokenDictionaryEntry` (id, name, value, label, description,
createdAt, updatedAt), `CreateTokenDictionaryEntryRequest` (name, value, label, description),
`UpdateTokenDictionaryEntryRequest` (value, label, description — no `name`). Decode with
`contracts.Decode` (strict, rejects unknown fields).

`module.go` `New(Dependencies)` constructor (panic on nil deps; follow `taxonomy`). Wire in the
composition root `apps/api/cmd/metaldocs-api/main.go`.

## 10. Capabilities (10-touchpoint wiring)

Two new capabilities, both `ScopeTenant`:
- `CapTokenView = "token.view"`
- `CapTokenDictionaryManage = "token_dictionary.manage"`

1. const + `validCapabilities` — `internal/modules/iam/domain/model.go`.
2. scope classify (both `ScopeTenant`) — `capability_scope.go`.
3. tier-1 route→cap — `apps/api/cmd/metaldocs-api/permissions.go` (omission = silent escalation).
4. tier-2 `authz.Require` in-tx — §6.
5. seed grants — `db/reference-data/0001_product_reference_data.sql`: `token.view` to
   template-authoring roles; `token_dictionary.manage` to admin-tier (exact role set per the existing
   per-role grant pattern in that file at impl).
6. DB tripwire — both names conform to `ck_cap_format` / `ck_cap_not_legacy`; no constraint change.
7. guard tests stay green — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet`.
8. **bump `TestCapabilityRegistrySize` `const want` 31 → 33** — `model_test.go` (targeted-verified: current 31).
9. CI capability-coherence (5-surface) — REQ-AUTHZ-5.
10. H-PRE-1 — satisfied (no locks).

## 11. Errors (RFC 9457 `problem+json`)

`problem.New` / `problem.Write`; never bare `http.Error`.
- 401 no session · 403 missing capability · 404 not found / cross-tenant (never 403 for cross-tenant)
- 409 duplicate `(tenant_id, name)` · 422 validation (charset / length / empty / attempted `name` change on PUT).

## 12. Multi-tenant

`token_dictionary_entries.tenant_id NOT NULL`; every query predicated on the tenant; `tenant.FromContext`
+ `authz.SeedTxIdentity` set the tx-local GUC; a cross-tenant id resolves to **404**.

## 13. Frameworks reused (no hand-rolled equivalents)

`TxRunner` · `tenant.FromContext` · `authz.SeedTxIdentity`/`Require` · `audit.RecordTx` ·
`problem.New`/`Write` · `httpresponse` · `contracts.Decode` · `oapi-codegen` (DTOs) · `testdb` factory.

## 14. Testing

Canonical `testdb` integration factory (`tests/integration/testdb/`); `//go:build integration`;
R1–R4 discipline (`scripts/check-test-discipline.sh`).

- **Unit:** both caps classified + in registry; `TestCapabilityRegistrySize` = 33.
- **Integration — AuthZ:** granted cap → 2xx; ungranted → 403; missing session → 401; tier-1 mapping present.
- **Integration — Multi-tenant:** tenant A cannot read/update/delete tenant B's entry → 404.
- **Integration — DB-invariant:** duplicate `(tenant_id, name)` → 409; bad `name` charset → 422; empty
  `value` → 422; PUT changing `name` → 422.
- **Audit:** a mutating op records an audit event in the same tx (asserted via the audit read surface).
- **Evidence:** `go build ./...`, `go test ./...`, `go test -tags=integration ./...`,
  `.\scripts\check-system-runnable.ps1` — all green; review/QA disposition; bounded defers (SP-2..SP-5).

## 15. Docs / ADR

- `wiki/modules/tokens.md` (12-section, exemplar `wiki/modules/taxonomy.md`) + `wiki/modules/tokens-tech-debt.md`
  (record the SP-2 collision-deferral and the catalog-duplication smell) + entry in `wiki/modules/index.md`.
- Refresh `Last verified` in `wiki/modules/templates.md` (computed catalog gains a dictionary neighbour).
- **ADR superseding ADR 0008** — the catalog stops being fixed at SP-1. The ADR preserves 0008's correct
  half (computed tokens are server-resolved and secure) and documents the tenant-dictionary extension +
  the Node-owned-grammar / Go-anti-corruption split (Decision A).
- **Numbering reconciliation:** `wiki/concepts/token-syntax.md` calls the tenant dictionary "SP-2"; the
  north-star calls it SP-1. The ADR / wiki update aligns the wiki to the north-star numbering.
- REQ-IDs cited: REQ-AUTHZ-1, REQ-AUTHZ-2, REQ-AUTHZ-5, REQ-MT-1, REQ-CONTRACT-1
  (`wiki/architecture/backend-target-architecture.md`).

## 16. Verdict carried from orientation

Yellow — proceed. Locked constraints honored above: new module `tokens`; two `ScopeTenant` caps with
full wiring + registry bump 31→33; ADR superseding 0008; `render → tokens` is the only future edge;
contract-first; `tenant_id` everywhere; `audit.RecordTx` in every mutation; `testdb` only; no render/UI.
