# ADR 0084 — Approval route-stage actor selectors: a discriminated union (M4 ActorSelector)

> **Status:** Accepted 2026-07-13
> **Extends:** [ADR 0082](0082-approval-kernel-extraction.md) (approval as a subject-generic kernel),
> [ADR 0081](0081-per-profile-signature-policy.md) (per-profile route governance).
> **Relates to:** [ADR 0022](0022-authz-capability-coherence.md) (area-role registry
> binding), [ADR 0035](0035-flat-typed-responses-and-presign-status.md) (flat wire DTOs — no `oneOf`/discriminator),
> [ADR 0073](0073-remove-finalize-wrapper-canonical-submit.md) (server-resolved active route at
> submit).
> **Scope:** ROADMAP unit 3.2 / approval-remediation M4 — how a route stage names *who* may act on it.

## Context

Before M4 a route stage named its actors with two flat columns on `approval_route_stages`:
`required_role` + `area_code`. That expresses exactly one shape — "the holders of role R in area A" —
and cannot represent the actor patterns the ratified review/approval model (spec §6, B1–B6) requires:

- a **named individual** (one specific user), independent of role;
- a role in a **fixed** area chosen at route-authoring time;
- a role in the **document's own** area, resolved live per subject;
- a role×area pool from which the **submitter picks** the concrete actors at submit time.

Flat `required_role`/`area_code` is a **local maximum**: every richer pattern would have to be
smuggled through sentinel values or parallel side-tables, each an invariant the DB could not enforce.
The global-maximum structure is a first-class **discriminated union of actor selectors**, one row per
selector, with the DB enforcing per-kind field consistency — the same shape a BPMN engine uses to bind
a lane/task to its potential owners.

## Decision

A stage owns an ordered list of **actor selectors**. Each selector is a discriminated union over four
kinds; `kind` decides which of `{user_id, role, area_code}` are meaningful. Selectors are the **sole**
source of truth for stage actors — the flat `required_role`/`area_code` columns are removed, not
shadowed (see Migration).

### The four kinds

| kind | meaningful fields | resolves to |
|---|---|---|
| `named_user` | `user_id` | exactly that user |
| `role_in_fixed_area` | `role`, `area_code` | holders of `role` in the fixed `area_code` |
| `role_in_document_area` | `role` | holders of `role` in the subject document's own area (resolved live) |
| `submit_choice` | `role`, `area_code` | the submitter's chosen subset of holders of `role` in `area_code` |

### Three-way parity (DB ⋈ domain ⋈ wire)

The field-presence contract is written **once** and enforced in three places that must never drift:

1. **DB (last line):** child table `public.approval_route_stage_selectors`
   (migration `0303`) — `(tenant_id, route_stage_id, selector_order, kind, user_id?, role?, area_code?)`
   with `ON DELETE CASCADE` from the stage, a `UNIQUE (route_stage_id, selector_order)`, and a
   `CHECK` that is the exact truth table:
   ```
   (kind='named_user'           AND user_id NOT NULL AND role NULL     AND area_code NULL)
   OR (kind='role_in_fixed_area'    AND role NOT NULL AND area_code NOT NULL AND user_id NULL)
   OR (kind='role_in_document_area' AND role NOT NULL AND area_code NULL     AND user_id NULL)
   OR (kind='submit_choice'         AND role NOT NULL AND area_code NOT NULL AND user_id NULL)
   ```
2. **Domain (friendly first line):** `domain.ActorSelector.Validate()`
   (`internal/modules/approval/domain/selector.go`) enforces byte-for-byte the same presence/absence
   rules; an unknown kind → `ErrInvalidSelectorKind`, wrong fields → `ErrSelectorFieldsInvalid`.
3. **Wire (contract-first):** the OpenAPI `ActorSelector` schema is a **flat** object
   `{ kind, user_id?, role?, area_code? }` — deliberately **no** `oneOf`/`discriminator`, per ADR 0035.
   Route-admin request/response and the submit-preview response all reuse this one schema.

### Area-role registry binding (ADR 0022)

The area-role registry binding (a `role` may only be paired with an `area_code` the registry
permits) is re-homed from the old flat `required_role` onto **each selector's `role`** for the three
role-bearing kinds (`role_in_fixed_area`, `role_in_document_area`, `submit_choice`). `named_user`
carries no role and is excluded. The check lives at the HTTP contract boundary (`validateAreaRole`),
the friendly first line ahead of the DB.

### `submit_choice` requires a read half (submitter preview)

`submit_choice` shifts actor selection to submit time: the submitter must be shown the role×area pool
and pick a subset, sent as `chosen_actors: [{stage_order, user_ids}]`. The submit path already
resolves the active route server-side (ADR 0073) and fail-closes (`ErrSubmitChoiceRequired` /
`ErrSubmitChoiceConstraintViolated`, 422) — but that is a **write**, so a submitter had no way to learn
the route's `submit_choice` stages *before* committing. That asymmetry is the gap M4 closes with a
read half:

- **`GET /documents/{id}/approval-preview`** and
  **`GET /templates/{id}/versions/{n}/approval-preview`** return the resolved active route's stages +
  selectors (the flat `ActorSelector` schema), so the frontend can render a per-`submit_choice`-stage
  picker filtered by `role`×`area_code`.
- Gated by the submitter's **existing** capability (`CapDocumentSubmit` / `CapTemplateSubmit`) — no new
  capability, least-privilege preserved.
- **Read-only and resolution-parity-bound:** `PreviewRoute` reuses the *same* `resolveActiveRoute` /
  `LoadActiveRouteIDBySubject` logic the submit path uses, so preview and submit can never resolve
  different routes. An integration parity test pins this. The server remains authoritative — the FE
  picker and its client-side "every submit_choice stage has ≥1 actor" check are the friendly first
  line only; the backend 422 is the hard line.
- No active route → empty preview (200), never a fabricated route.

### Drift policy is per selector

Each selector carries its own drift policy (how a stage reacts when its resolved actor pool changes
after a route is bound), rather than one policy for the whole stage. Instance-side audit snapshots
(`RequiredRoleSnapshot`/`AreaCodeSnapshot`) are retained on the instance for historical fidelity and
are out of scope here.

## Migration (expand → backfill → contract)

1. **Expand** (`0303`): add `approval_route_stage_selectors`; backfill one `role_in_fixed_area`
   selector per existing stage from its flat `required_role`/`area_code`.
2. **Cutover:** domain, repository, resolver, and HTTP contract read/write selectors; flat→selector
   synthesis relocated to the HTTP boundary as a byte-stable shim, then removed once the wire contract
   dropped the flat fields (slice 7a).
3. **Contract** (`0305`): drop `approval_route_stages.required_role`/`area_code`. Selectors are now the
   sole source of truth. No relax-to-optional compat is kept (standing "no legacy fallback" rule).

## Consequences

- **Positive:** actor binding is now expressive enough for the full B1–B6 model; the DB enforces every
  per-kind invariant; `submit_choice` is usable end-to-end (authorable *and* submittable) via a
  capability-correct, parity-bound read endpoint; the flat wire stays ADR-0035-clean (no discriminated
  wire union).
- **Cost:** a stage's actors are now a child-table read (`loadRouteStageSelectors` batch-loads them);
  the three-way parity (DB CHECK ⋈ domain `Validate` ⋈ wire schema) must be kept in lockstep by hand —
  a change to the kind set touches all three. This ADR is the pin.
- **Rejected — widen the flat columns / sentinels:** encoding four kinds into `required_role`/
  `area_code` sentinels would put invariants beyond the DB's reach and lock in the local maximum.
- **Rejected — `oneOf`/discriminator wire union:** breaks ADR 0035's flat-envelope rule and the
  hand-written-consumer safety it buys; the flat `{kind, …?}` object is validated by the domain and DB
  regardless.
- **Rejected — reactive 422-driven picker (no preview endpoint):** would couple the picker to a failed
  submit and carry business data on the error channel; leaves the read/write asymmetry unresolved.
