# G1 — Per-profile signature policy — implementation design (LOCKED)

**Date:** 2026-07-10
**ROADMAP unit:** 2.1 (G1). One gated feature. G2/G3 out of scope.
**Status:** Design LOCKED by operator 2026-07-10 (research-validated). This doc CAPTURES the
locked design for implementation; it does NOT renegotiate architecture.
**Governing spec (LOCKED):** `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (R1 + G1).
**System-impact (Yellow, binding constraints):** `docs/superpowers/analysis/2026-07-10-review-approval-workflow-model-system-impact.md`.
**Process:** `docs/superpowers/HARNESS.md` (unit loop P1→P7, §4 anti-slop, §5 ladder).

---

## 0. One-line intent

A document profile's **governance class** decides whether its approval route MUST contain a
signature (approval-kind) stage. Controlado ⇒ ≥1 approval stage required. Simples ⇒ review-only
route allowed. Livre ⇒ no route permitted at all. Enforced friendly-first in domain
`Route.Validate(policy)`, authoritative-last in a DB deferrable constraint trigger.

## 1. Orientation (whole-system)

- **Owning modules:**
  - `taxonomy` — OWNS the governance classification. New `GovernanceClass` enum + column on
    `document_profiles`; new pure `RoutePolicy()` derivation. (ADR 0038 profile owner.)
  - `documents/approval` (nested kernel, ADR 0072) — CONSUMES the policy. `Route.Validate` gains a
    policy parameter; enforced at route-admin create/update and at submit.
- **Cross-module edge (NEW):** `documents/approval → taxonomy` (published surface only). approval
  reads the profile's `RoutePolicy` through a NEW narrow read port + adapter that preserves
  taxonomy repo authz (`CapTaxonomyView`). NOT via `CDFieldReader`. `Route.Validate` stays **pure
  domain** — policy is passed IN as a value, never fetched inside the domain.
- **Invariants touched:** capabilities-not-roles (policy is a DOMAIN rule, not an authz rule — no
  new capability); contract-first (governance_class on taxonomy profile API via openapi + regen);
  multi-tenant pooled (column on a tenant table, every read tenant-scoped, cross-tenant → 404);
  DB-enforces-invariants (deferrable constraint trigger is the last line); cross-module via
  published interface only (narrow port). H-PRE-1 respected: the policy read runs in the taxonomy
  repo's OWN short tx (separate connection), never inside a lock-holding approval tx.
- **Foundation verdict:** SOUND. M2b left `StageKind{review,approval}` substrate (migration 0286)
  and a kind-agnostic `AdvanceStage()`. G1 wires that substrate as designed — additive
  parameterization, not a patch. No AS-1/AS-2/AS-3.

## 2. Locked design — the seven pieces

### P1. taxonomy OWNS governance classification (domain)

`internal/modules/taxonomy/domain/profile.go`:

```go
// GovernanceClass classifies how strictly a document profile is governed. It
// drives the per-profile route-signature policy (RoutePolicy). DB-mirrored by
// document_profiles.governance_class CHECK ('controlado','simples','livre').
type GovernanceClass string

const (
    GovernanceControlado GovernanceClass = "controlado" // POP, IT, desenho, FMEA
    GovernanceSimples    GovernanceClass = "simples"    // nota fiscal, orçamento, relatório
    GovernanceLivre      GovernanceClass = "livre"      // rascunhos, material não governado
)

func (c GovernanceClass) Validate() error { … ErrInvalidGovernanceClass }

// RoutePolicy is the narrow, approval-facing consequence of a GovernanceClass.
// It is intentionally 3-valued and carries NO taxonomy internals — the only
// thing that crosses the approval boundary.
type RoutePolicy string

const (
    RoutePolicyRequireApprovalStage RoutePolicy = "require_approval_stage" // controlado
    RoutePolicyApprovalOptional     RoutePolicy = "approval_optional"      // simples
    RoutePolicyNoRoutePermitted     RoutePolicy = "no_route_permitted"     // livre
)

// RoutePolicy derives the route-signature policy for this profile. Pure; total
// over the three valid classes; unknown/empty class fail-closes to
// RequireApprovalStage (safest — never silently drops the signature requirement).
func (p *DocumentProfile) RoutePolicy() RoutePolicy { … }
```

- `DocumentProfile` gains `GovernanceClass GovernanceClass` (JSON `governance_class`).
- `NewDocumentProfile` defaults empty class → `GovernanceControlado` (behavior-preserving; matches
  the DB default and the pre-G1 world where every profile required signatures) and validates it.
- Empty/unknown class in `RoutePolicy()` fail-closes to `RequireApprovalStage` (no-fallback
  principle: an integrity gate never weakens on ambiguous input).
- **CI guard:** taxonomy domain keeps ZERO `database/sql` import (`nosqltxindomain`). This file
  imports nothing new.

### P2. approval → taxonomy narrow read port + adapter

- **Port** (`internal/modules/documents/approval/application/ports.go`, new or appended):
  ```go
  // ProfilePolicyReader is the narrow approval→taxonomy read port: given a
  // tenant + profile code it returns the profile's route-signature policy.
  // Mirrors the interface-segregation pattern in
  // controlleddocuments/infrastructure/taxonomy_reader.go.
  type ProfilePolicyReader interface {
      RoutePolicy(ctx context.Context, tenantID, profileCode string) (taxonomydomain.RoutePolicy, error)
  }
  ```
- **Adapter** (`internal/modules/documents/approval/infrastructure/profile_policy_reader.go`, new):
  a narrow getter slice `{ GetByCode(ctx, tenantID, ProfileCode) (*DocumentProfile, error) }`
  satisfied by the canonical taxonomy `ProfileRepository`; `RoutePolicy` delegates to
  `GetByCode(...).RoutePolicy()`. GetByCode runs in taxonomy's own short tx and enforces
  `CapTaxonomyView` + GUC — authz preserved, and being a separate tx keeps it off any approval
  lock-holding tx (H-PRE-1).
- **Wiring:** `NewServices(...).WithProfilePolicyReader(reader)` option (mirrors
  `WithScheduledPublishEnqueuer`) so existing call sites/tests stay valid; production composition
  root (`apps/api/cmd/metaldocs-api/main.go`, `apps/jobs/.../main.go`) wires the adapter over the
  taxonomy `profileRepo`. A **nil** reader ⇒ the friendly policy check is skipped (treated as
  `ApprovalOptional`); the DB trigger remains the authoritative last line, so a misconfigured
  composition root degrades to DB-enforced, never to unenforced. Production wiring is asserted
  present.

### P3. `Route.Validate(policy)` — pure domain, both call sites

`internal/modules/documents/approval/domain/route.go`:

```go
func (r Route) Validate(policy taxonomydomain.RoutePolicy) error {
    // …existing structural checks (dense order, quorum, unique names)…
    switch policy {
    case taxonomydomain.RoutePolicyNoRoutePermitted:
        return ErrRouteNotPermittedForProfile        // livre: route creation actively blocked
    case taxonomydomain.RoutePolicyRequireApprovalStage:
        if !hasApprovalStage(r.Stages) { return ErrApprovalStageRequired } // controlado ⇒ ≥1 approval
    case taxonomydomain.RoutePolicyApprovalOptional, "":
        // simples (and unset in tests): review-only route allowed — no extra constraint.
    }
    return nil
}
```

New domain errors: `ErrRouteNotPermittedForProfile`, `ErrApprovalStageRequired`.

Enforced at **both** service call sites, each resolving the policy from the port and passing it in:
- `route_admin_service.go` `createTx` (~:172) and `updateTx` (~:304) — policy resolved from
  `in.ProfileCode` (create) / the locked route's profile (update) via the port BEFORE/independent of
  the write, passed to `route.Validate(policy)`.
- `submit_service.go` (~:128) — after the profile code is resolved in-tx, resolve its policy via the
  port and pass to `route.Validate(policy)`.

Livre naturally has no active route, so submit already fails at route resolution; the explicit
`NoRoutePermitted` check is belt-and-suspenders and drives the route-admin block.

### P4. DB last line — expand-only migration + deferrable constraint trigger (BOTH directions)

New migration `db/migrations/0295_profile_governance_class.sql` (expand-only, forward-only,
idempotent — idiom from 0286/0287):

1. `ALTER TABLE metaldocs.document_profiles ADD COLUMN IF NOT EXISTS governance_class text NOT NULL DEFAULT 'controlado'`
   + named CHECK `document_profiles_governance_class_check CHECK (governance_class IN ('controlado','simples','livre'))`.
   Behavior-preserving backfill: every existing profile becomes `controlado` (today's implicit
   "signature required" world) — no live route silently becomes invalid.
2. **Deferrable constraint trigger** `CONSTRAINT TRIGGER … DEFERRABLE INITIALLY DEFERRED` firing in
   BOTH directions to a shared `enforce_profile_route_policy()` plpgsql function:
   - **Route direction:** AFTER INSERT/UPDATE/DELETE on `public.approval_route_stages` (and on
     `public.approval_routes`) — deferred to commit so it sees the FULL stage set of a multi-statement
     route write. For the affected route's profile class: `controlado` ⇒ route must have ≥1
     `stage_kind='approval'` stage; `livre` ⇒ an ACTIVE route must not exist; `simples` ⇒ no
     constraint. Violation ⇒ `RAISE EXCEPTION … ERRCODE 'P0001'` with a stable message prefix.
   - **Reclassify direction:** AFTER UPDATE OF `governance_class` on
     `metaldocs.document_profiles` — deferred — re-validates the profile's active route(s) against
     the NEW class; conflict ⇒ same P0001 with a distinct message token so the app maps it to the
     reclassify-conflict error.
   Only ACTIVE routes are considered (`approval_routes.active`), so historical/superseded versions
   never trip the guard. In-flight instances are already protected by `RouteVersionSnapshot`
   pinning — **freeze logic is NOT touched**.
3. **App-side friendly 409** `ErrClassChangeRouteConflict`: taxonomy's reclassify path maps the DB
   trigger's reclassify-token P0001 → this domain error → RFC 9457 `409 problem+json`. No taxonomy→
   approval read is introduced — the friendly layer is pure error-mapping over the constraint.

### P5. Contract-first — governance_class on taxonomy profile API + governance event

- `api/openapi/v1/openapi.yaml`: add `governance_class` (enum `[controlado,simples,livre]`) to
  `TaxonomyProfileUpsertRequest` (optional; server defaults `controlado`) and to
  `DocumentProfileItem` (response). Regenerate via `oapi-codegen` ONLY — never hand-edit
  `api.gen.go`.
- taxonomy handler maps the new field on create/update/get/list.
- **Persistence:** taxonomy `ProfileRepository` Create/Update/Get/List read & write
  `governance_class`. `ProfileService` emits a governance event on class change via the existing
  `LogTx` pattern — new `GovernanceEventTypeProfileGovernanceClassChange = "profile.governance_class_change"`.
  Class change is detected against the locked prior value (FOR UPDATE), mirroring
  `SetDefaultTemplate`.

### P6. ADR (required)

New ADR (next number) ratifying:
- **Per-profile, NEVER system-global.** Operator explicitly killed the global "must have signature"
  invariant — record that this ADR SUPERSEDES/rejects it; never reintroduce.
- **Published-Language / Knowledge-Level rationale:** `RoutePolicy` is the published language across
  the approval↔taxonomy boundary (3-valued, no taxonomy internals); `GovernanceClass` is
  taxonomy's knowledge-level type that derives it.
- **Predicate-rule traceability caution:** the class→policy→route-shape rule is enforced in three
  places (domain Validate, DB trigger both directions); record the single-source-of-truth intent
  and the drift risk (a fourth consumer must reuse `RoutePolicy()`/the port, not re-derive).
- **Type-object-table evolution note:** `governance_class` is a small closed enum today; if classes
  grow attributes it evolves toward a type-object table — recorded so a future author doesn't
  bolt booleans onto `document_profiles`.

### P7. module-boundaries

The conformance model (`scripts/check-module-boundaries.ps1`) already permits cross-module imports
that target a module's published `domain`/`application`/`api` layers. approval importing
`taxonomy/domain` (for `RoutePolicy`) is therefore conformant with no allow-list edit required —
this is verified at L0. The NEW `approval → taxonomy` edge is recorded in the ADR and the wiki
module docs (documentation of the edge is the deliverable; the lint needs no new entry). If a future
narrower taxonomy published package is introduced for this, it is added to `$publishedPackages`.

## 3. Slices (TDD, failing test first, two-stage review each)

| # | Slice | Files | First failing test |
|---|-------|-------|--------------------|
| S1 | taxonomy domain: GovernanceClass + RoutePolicy + field + derivation | `taxonomy/domain/profile.go`, new `governance_class.go`+`_test.go`, `errors` | `RoutePolicy()`/`Validate` table test |
| S2 | migration + deferrable trigger (both directions) | `db/migrations/0295_*.sql` | testdb integration: controlado review-only route rejected at COMMIT; livre active route rejected; simples accepted; reclassify-with-conflict rejected |
| S3 | taxonomy persistence + service + governance event + 409 mapping | `taxonomy/infrastructure/repository.go`, `application/profile_service.go`, `domain/port.go` (event type), `domain/errors` | testdb: create/get round-trips class; reclassify emits event; conflict → ErrClassChangeRouteConflict |
| S4 | approval read port + adapter + wiring | `approval/application/ports.go`, `approval/infrastructure/profile_policy_reader.go`, `application/services.go`, both `main.go` | adapter maps GetByCode→RoutePolicy; `WithProfilePolicyReader` |
| S5 | `Route.Validate(policy)` + 3 call sites | `approval/domain/route.go`, `route_admin_service.go`, `submit_service.go`, domain errors | domain table test (require/optional/no-route); route-admin create rejects controlado review-only; submit rejects likewise |
| S6 | contract-first API surface | `api/openapi/v1/openapi.yaml`, regen, taxonomy handler | contract/integration: POST profile with governance_class persists + returns it |
| S7 | ADR + wiki + evidence | `wiki/decisions/<n>-*.md`, `wiki/modules/taxonomy.md`, `wiki/modules/documents.md` | n/a (docs) |

Dependencies: S1 → {S3,S4,S5}; S2 → S3; S1+S4 → S5; S3 → S6. Commit per green slice; NO batch merge.

## 4. Verification ladder (P4)

- **L0:** `go build ./...`; `.\scripts\check-module-boundaries.ps1` (approval→taxonomy/domain
  conformant); api-lint (openapi↔gen↔handler parity); check-test-discipline; capability-coherence
  (expect no-op — no new cap).
- **L1:** `go test ./...` + `go test -tags=integration ./...` vs docker postgres (trigger both
  directions, reclassify 409, persistence round-trip).
- **L2:** container stack via gateway :80 smoke — (a) Livre profile route creation blocked 4xx
  problem+json; (b) Controlado review-only route rejected at route-admin AND at submit; (c) Simples
  review-only route accepted.

## 5. Non-negotiables re-checked (miss-nothing)

Tenant predicate on every new query · cross-tenant → 404 · problem+json for every new error
(`ErrRouteNotPermittedForProfile`, `ErrApprovalStageRequired`, `ErrClassChangeRouteConflict`,
`ErrInvalidGovernanceClass`) · no new capability (policy is domain) · DB trigger is the last line ·
H-PRE-1 (policy read off the lock-holding tx) · openapi↔gen↔handler parity · taxonomy domain zero
`database/sql` · testdb factory for all DB integration · freeze logic untouched.
