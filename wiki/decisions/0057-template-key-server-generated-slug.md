# ADR 0057 — Template `key`: server-generated slug + uniqueness suffix, drop client-supplied key

- **Status:** Accepted (decision only — implementation is a follow-up)
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Resolves the `key-generation` blocker at `wiki/backlog/novo-template-wizard.md:39-44` for `POST /api/v1/templates`. Records the decision only. Does **not** change `api/openapi/v1/openapi.yaml`, the templates application/domain code, or the wizard UI — those are follow-up work tracked by `wiki/modules/templates-tech-debt.md` T-015 (added by this ADR).
- **Depends on:** none.

---

## Context

### Verified runtime facts

- **Contract today requires a client-supplied `key`.** `api/openapi/v1/openapi.yaml:1159-1172` — the `POST /templates` request body schema has `required: [key, name]` (line 1165) with `key: { type: string }` (line 1167) and no format/pattern constraint, no server-derivation hint.
- **Application layer trusts the client value verbatim and gates on a pre-check + DB conflict.** `internal/modules/templates/application/create.go:14-35` (`CreateTemplateCmd.Key string`, line 18) — `CreateTemplate` calls `s.repo.GetTemplateByKey(ctx, cmd.TenantID, cmd.Key)` (line 31) and returns `domain.ErrKeyConflict` (`internal/modules/templates/domain/template.go:30`) if a row already exists for that `(tenant_id, key)`. The check-then-insert is not itself atomic against a concurrent duplicate `key` (a unique DB constraint is the actual backstop at insert time); either way, the key's *content* is 100% client-chosen.
- **Contract already returns 409 on conflict.** `api/openapi/v1/openapi.yaml:1186-1187` — `'409': $ref: '#/components/responses/Conflict'`, consistent with `ErrKeyConflict`.
- **Design has no key input.** `wiki/backlog/novo-template-wizard.md:40` — "Design has no key input — derived auto from name." The wizard UI was never built with a key field; the backend requirement is what's blocking Step 5, not a missing UI affordance.
- **No existing slug-generation utility was found in the templates module** (`internal/modules/templates/application/*.go` has no `slug`/`Slugify` helper) — this is greenfield for the module, not a rename of existing logic.

## Decision

**Drop the client-supplied `key` from the create-template request. The server generates the key from `name` server-side, with a uniqueness suffix on collision.**

This resolves the backlog's option (b) ("backend generates key from name server-side") over option (a) ("advanced Identificador técnico field with manual override"). Rationale: the design (source of truth for the wizard UX) already has no key input (`novo-template-wizard.md:40`), so option (a) would require adding UI surface the design never called for, just to work around a backend requirement. Option (b) matches the design as-is and removes the fragility the backlog itself flagged ("Auto-slug from name is fragile (collisions, edits break links)") by making collision handling a defined server rule instead of a client guess.

Binding shape for the follow-up implementation:

1. **`key` is removed from the `POST /templates` request body.** `name` remains the only required identifying input. `key` becomes purely a server-computed, response-only field (already present in `CreateTemplateResponse`).
2. **Slug algorithm:** lowercase `name`, transliterate/strip diacritics, replace runs of non-`[a-z0-9]` with a single `-`, trim leading/trailing `-`, cap length to a fixed bound (e.g. 63 chars, matching common slug/DNS-label conventions and leaving room for a suffix).
3. **Uniqueness suffix on collision:** if the base slug collides with an existing `(tenant_id, key)` row, the server appends a short deterministic disambiguator (e.g. `-2`, `-3`, ... or a short random/opaque suffix) and retries the existing `GetTemplateByKey` check-then-insert (or, better, races it against the DB unique constraint and retries on conflict — the follow-up should evaluate whether to keep the pre-check race or replace it with catch-and-retry-on-constraint-violation, since the same TOCTOU gap that exists today would remain otherwise).
4. **`key` is immutable after creation**, matching the existing pattern for other identifier-shaped fields elsewhere in the schema (e.g. taxonomy's `code` column, made immutable by `reject_code_update()` — `archive/migrations/0123_taxonomy_extend_process_areas.sql:61-75`). Renaming a template updates `name` only; `key` does not follow.
5. **Client never supplies or edits `key`.** No "advanced technical identifier" field is added to the wizard (option (a) is explicitly not chosen); if a future need for a stable, human-chosen technical identifier emerges (e.g. for external cross-references), that is a new decision, not a reopening of this one.

## Consequences

- Removes the wizard Step 5 blocker without adding UI scope the design didn't call for.
- `POST /templates` becomes a breaking contract change (`key` required → absent) — the follow-up implementation must update `api/openapi/v1/openapi.yaml`, regenerate via `oapi-codegen` (contract-first, per the repo's non-negotiable invariant), update `CreateTemplateCmd`/`CreateTemplate`, and update any FE caller currently sending `key`.
- The 409-on-conflict response path likely becomes unreachable from client input (server no longer accepts a client key to conflict on) — the follow-up should confirm whether 409 stays reserved for the internal collision-retry exhaustion case or is dropped from the contract for this endpoint.
- `wiki/modules/templates-tech-debt.md` gets a new row (T-015, added by this ADR) linking here so the follow-up implementation has a tracked entry point.

## References

- `wiki/backlog/novo-template-wizard.md:39-44` — `key-generation` blocker this ADR resolves.
- `api/openapi/v1/openapi.yaml:1151-1189` — current `POST /templates` contract (required `key`, 409 Conflict response).
- `internal/modules/templates/application/create.go:14-35` — `CreateTemplateCmd`/`CreateTemplate` (client-key trust + conflict check).
- `internal/modules/templates/domain/template.go:30` — `ErrKeyConflict`.
- `wiki/modules/templates-tech-debt.md` T-015 — follow-up implementation tracking row (added by this ADR).
- `archive/migrations/0123_taxonomy_extend_process_areas.sql:61-75` — `reject_code_update()` precedent for immutable-identifier-after-create pattern.
