# ADR 0081 — Per-Profile Signature Policy (Governance Class → Route Policy, DB-Enforced, Submit Relies on the Write-Boundary Invariant)

> **Status:** Accepted
> **Date:** 2026-07-10
> **Scope:** A document profile carries a **governance class** {controlado, simples, livre} that decides
> whether its approval route MUST contain a signature (approval-kind) stage. The rule is owned by
> `taxonomy` as a pure domain derivation (`RoutePolicy()`), consumed by `documents/approval` through a
> narrow published port, enforced friendly-first at the route-admin write boundaries, and authoritative-last
> by a DEFERRABLE bidirectional DB constraint trigger. **Submit performs structural-only route validation and
> relies on the write-boundary invariant** — a deliberate, recorded deviation from the locked design's
> literal "enforce at submit" wording (see Decision §4).
> **ROADMAP unit:** 2.1 (G1). One gated feature; G2/G3 out of scope.
> **Governing spec:** `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (R1 + G1)
> **Implementation design (LOCKED):** `docs/superpowers/specs/2026-07-10-g1-per-profile-signature-policy-design.md`
> **System-impact (Yellow):** `docs/superpowers/analysis/2026-07-10-review-approval-workflow-model-system-impact.md`
> **Related:** ADR 0022 (capabilities-not-roles), ADR 0038 (taxonomy owns document profiles),
> ADR 0072 (approval is a nested kernel inside documents), ADR 0073 (submit resolves route/profile in-tx),
> ADR 0035 (generated DTOs), H-PRE-1 (no authz-recording read inside a lock-holding tx),
> RouteVersionSnapshot pinning (in-flight instance protection)
> **Key files:**
> - `internal/modules/taxonomy/domain/governance_class.go` — `GovernanceClass` enum + `RoutePolicy()` derivation
> - `internal/modules/taxonomy/domain/profile.go` — `governance_class` field, construction default, validation
> - `internal/modules/taxonomy/application/profile_service.go` — `Reclassify` (the SOLE class-write use-case)
> - `internal/modules/taxonomy/infrastructure/repository.go` — `SetGovernanceClassTx`; generic `Update` never writes the class
> - `db/migrations/0295_profile_governance_class.sql` — column + backfill + bidirectional DEFERRABLE triggers
> - `api/openapi/v1/openapi.yaml` — `governance_class` on `DocumentProfileItem` + `TaxonomyProfileUpsertRequest`
> - `internal/modules/documents/approval/application/ports.go` — `ProfilePolicyReader` (narrow published port)
> - `internal/modules/documents/approval/infrastructure/profile_policy_reader.go` — taxonomy adapter (fail-closed)
> - `internal/modules/documents/approval/domain/route.go` — `Route.Validate(policy)` + `hasApprovalStage`
> - `internal/modules/documents/approval/application/route_admin_service.go` — off-tx policy resolve + validate
> - `internal/modules/documents/approval/application/submit_service.go` — structural-only `Validate("")`

## Context

The review/approval workflow model (ratified 2026-07-10) requires that a profile's governance strictness
decide the shape of its approval route. Not every document needs a signed sign-off: a `nota fiscal` or an
internal `relatório` should route through review without forcing a signature stage, while a `POP` or an `IT`
must carry one; a `rascunho` should not route at all. Before G1 this was operator lore, unenforced — any
profile could be wired to any route shape.

The M2b substrate already made this expressible: migration 0286 added `StageKind{review, approval}` (default
`approval`) and `AdvanceStage()` is kind-agnostic. G1 is additive parameterization of that substrate, not a
patch — the foundation verdict is SOUND.

Three constraints shaped where enforcement can live:

1. **Capabilities-not-roles (ADR 0022).** The policy is a *domain* rule about route shape, not an authz rule.
   It introduces NO new capability and no new PDP tier. `taxonomy.view` / `taxonomy.manage` already gate the
   reads and writes involved.
2. **Cross-module boundary (ADR 0072).** `approval` may consume `taxonomy` only through a published surface,
   never its repository/SQL/domain internals.
3. **H-PRE-1 + ADR 0073 collide at submit.** ADR 0073 requires submit to resolve route/profile *in-tx* to close
   the wrapper-era TOCTOU. H-PRE-1 forbids an authz-recording read (the taxonomy policy read preserves
   `CapTaxonomyView`) inside a lock-holding tx. There is no single tx in which submit can both hold its route
   lock and perform the policy read. This is a structural contradiction, not an implementation gap.

## Decision

**1 — `taxonomy` OWNS the classification (pure domain).** A `GovernanceClass` enum {controlado, simples, livre}
lives on `document_profiles`. A pure `RoutePolicy()` method derives a 3-valued published-language policy —
`RequireApprovalStage` (controlado), `ApprovalOptional` (simples), `NoRoutePermitted` (livre). Invalid/empty
class **fails closed** to `RequireApprovalStage` in the derivation and to `ErrInvalidGovernanceClass` in
construction/validation. The derivation carries no approval internals; the policy type is the entire contract
crossing the module edge (Published Language).

**2 — `approval` CONSUMES via a narrow published port.** `ProfilePolicyReader.RoutePolicy(ctx, tenantID,
profileCode) → RoutePolicy` mirrors the existing `taxonomy_reader.go` port; the adapter goes through the
taxonomy profile getter (preserving `CapTaxonomyView`), **NOT** through `CDFieldReader`. Approval never sees the
raw `GovernanceClass` enum — only the derived policy value. The reader **fails closed** (empty policy + error,
never a permissive default). `Route.Validate` stays pure domain: the policy is passed IN as a value, never
fetched inside the domain.

**3 — Enforce friendly-first at the route-admin write boundaries.** `Route.Validate(policy)` gains the policy
parameter: `NoRoutePermitted → ErrRouteNotPermittedForProfile`; `RequireApprovalStage → ErrApprovalStageRequired`
unless the route `hasApprovalStage` (a stage with `Kind == approval`, and — fail-closed — an unset/empty Kind
counts as approval); `ApprovalOptional`/empty policy → structural-only. Structural validation runs BEFORE the
policy check, so a malformed route fails structurally regardless of policy. Route-admin `createTx`/`updateTx`
resolve the policy **off-tx** (hoisted before `runner.Do`, per H-PRE-1) and call `Validate(policy)`.

**4 — Submit performs structural-only validation and relies on the write-boundary invariant (recorded
deviation).** The locked design's literal wording said "enforce at submit." Submit instead calls
`route.Validate("")` (structural-only; empty Kind ≡ approval, fail-closed) and does **not** perform the
per-profile policy read. This is a conscious deviation, ratified by the operator on 2026-07-10, forced by the
H-PRE-1 × ADR 0073 contradiction in Context §3.

The invariant that makes submit-time policy re-checking unnecessary:

> **Every active route, at submit time, already conforms to its profile's current governance class.**

held by construction at the write boundaries:

- **Route writes (direction A).** A route can only become/stay active through route-admin `Validate(policy)`
  (friendly first line) AND the DB direction-A trigger (last line) — neither admits a non-conforming active
  route.
- **Reclassification (direction B).** A profile's class can only change through the `Reclassify` use-case
  (friendly first line) AND the DB direction-B trigger (last line), which re-validates every active route
  against the new class and rejects the reclassification if any active route would then conflict.
- **The reclassify-vs-submit race is closed at the DB.** The bidirectional trigger is `DEFERRABLE INITIALLY
  DEFERRED`, so it evaluates the **final committed state** at COMMIT — no interleaving of a reclassification
  and a route write can commit an inconsistent pair. In-flight instances are additionally protected by
  **RouteVersionSnapshot pinning**: a submitted instance binds the route version it captured, so a later
  reclassification cannot retroactively invalidate a route the instance already resolved. `livre` profiles have
  no active route, so there is nothing for submit to bind.

Because the invariant is total at the write boundaries, a submit-time policy read would be redundant work that
also cannot be done correctly (it would have to be either in-tx — violating H-PRE-1 — or off-tx — violating
ADR 0073). **Option 2 (an in-tx taxonomy policy port at submit) is rejected permanently, not deferred.**

**5 — Contract-first API + friendly 409.** `governance_class` is added to `DocumentProfileItem` (required) and
`TaxonomyProfileUpsertRequest` (optional; server defaults to `controlado`) via `api/openapi` + `oapi-codegen`
regen — never a hand-edit of `api.gen.go`. The generic profile `Update` path deliberately **never writes**
`governance_class`; a class change is routed exclusively through `Reclassify`, so an ordinary edit can never
silently reclassify. `Reclassify` maps the DB `ErrClassChangeRouteConflict` P0001 token to a friendly HTTP 409
(`PROFILE_CLASS_ROUTE_CONFLICT`); an invalid class maps to 400.

**6 — DB is the last line (expand-only migration).** Migration 0295 adds the column with default `controlado`,
backfills, and installs the two DEFERRABLE INITIALLY DEFERRED constraint triggers (directions A and B). All
trigger queries are tenant-scoped. The DB — not the app — is the authoritative guarantor of the invariant;
the app checks are the friendly first line.

## Consequences

- Governance strictness is enforced end-to-end: a controlado profile cannot have an active review-only route,
  a livre profile cannot have any active route, and neither a route write nor a reclassification can commit a
  violating pair — the DEFERRABLE trigger is the final arbiter.
- The policy is a pure Published-Language value crossing one narrow port. `taxonomy` owns the classification;
  `approval` never learns the enum, only the derived 3-valued policy. Adding a fourth class or changing a
  derivation is a taxonomy-local edit; approval recompiles unchanged.
- Submit stays inside ADR 0073 (in-tx route/profile resolve) AND H-PRE-1 (no authz read in a lock-holding tx)
  simultaneously — the two are no longer in conflict because submit does not perform the policy read at all.
- No new capability, no new PDP tier, no authz allowlist edit — the module edge is auto-conformant because the
  policy is a domain rule, not an authz rule (ADR 0022 preserved).
- The class-write path is single-sourced through `Reclassify`; a future generic `Update` cannot regress into a
  silent reclassification, and the DB trigger backstops it regardless.

## Alternatives rejected

- **Option 2 — in-tx taxonomy policy port at submit (enforce policy at submit literally).** Rejected
  **permanently, not deferred.** It cannot satisfy H-PRE-1 (the `CapTaxonomyView`-preserving read is
  authz-recording and would run inside submit's lock-holding tx) and ADR 0073 (the resolve must be in-tx) at the
  same time. The write-boundary invariant + DEFERRABLE trigger already make the check redundant.
- **Enforce only in app code, leave the DB unconstrained.** App checks are the friendly first line, not the
  last; a future write path could reintroduce a non-conforming active route. The invariant belongs in the DB
  (DB-enforces-invariants).
- **A one-direction trigger (route writes only).** Leaves the reclassification path free to strand an active
  route in a class it no longer satisfies. Both directions are required for the invariant to be total.
- **A non-deferred (row-level) trigger.** Would reject valid multi-statement intermediate states (e.g. inserting
  a route then its approval stage) and could not evaluate the final committed pair. DEFERRABLE INITIALLY
  DEFERRED is load-bearing.
- **Consume the policy through `CDFieldReader`.** That port is a different, broader surface; reusing it would
  widen the module edge beyond the single `RoutePolicy` fact. A dedicated narrow `ProfilePolicyReader` keeps
  the boundary minimal.
- **Route policy keyed by role/capability instead of governance class.** Would reintroduce role-reasoning into a
  domain rule (ADR 0022 violation). The class is a per-profile taxonomy fact; the policy is derived from it,
  never from an actor's role.
