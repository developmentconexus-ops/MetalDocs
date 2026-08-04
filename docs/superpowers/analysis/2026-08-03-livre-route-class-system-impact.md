# System-impact analysis — Rota de aprovação classe "livre" (configured no-approval route)

**Date:** 2026-08-03
**Intent (one line):** governance_class='livre' passa a significar rota CONFIGURADA que não exige aprovação (ruling operador 2026-07-29); ausência de rota = sempre misconfiguração; rework do trigger de policy 0295; submit auto-aprova.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10 — ADR required; two design constraints locked)*

---

## 1. Classify & own

- **Work type:** feature (vocabulary + policy + submit semantics inside existing modules; no new module).
- **Owning module(s):**
  - `taxonomy` — owns `GovernanceClass` and the derived `RoutePolicy` published-language
    (`internal/modules/taxonomy/domain/governance_class.go`). The semantic change starts here:
    `RoutePolicyNoRoutePermitted` dies; a new policy value ("route required, zero approval burden")
    is born.
  - `approval` — owns route shape validation, the profile-policy DB trigger pair
    (`enforce_profile_route_policy` / `assert_route_shape`, ex-0295, now folded into baseline),
    route-admin write boundary, and `submit_service` resolution/binding semantics.
- **Explicitly NOT owning:** `documents` / `templates` — their creation gates and submit entry
  points consume approval's published behavior unchanged (gates just see "an active route exists");
  `iam` — no capability changes.
- **Cross-module edges (direction):**
  - `approval → taxonomy` via the published `RoutePolicy` value (already crosses the boundary today
    through `profile_policy_reader.go`; stays interface-only).
  - `documents/templates → approval` via existing submit/preview application services (unchanged
    signatures expected).
- **Ambiguity?** None — the exact boundary already exists and is documented in code comments
  (`submit_service.go:204-216`). No AS-3.

## 2. Foundation verdict

- **Base:** the 0295 architecture — class→policy derivation pure in taxonomy domain, enforcement by
  construction at write boundaries (route-admin `Validate(policy)` off-tx) + DEFERRABLE
  both-directions DB trigger, submit binding only active routes. This is a sound, ratified design
  (operator 2026-07-10; ADR 0073-era), not a patch.
- **Sound or patchy?** Sound. The current "livre = no route permitted" is a *deliberate* stance the
  operator has now reversed as product policy — this is a vocabulary/semantics change inside a good
  structure, not optimization inside a workaround. The change *removes* a special case (livre as
  route-less exception to config-first) and makes config-first universal — global maximum by
  construction.
- **AS-2?** No.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Marginal | No new/changed capability. Submit keeps its existing caps; auto-approve happens in the same submit tx under the same asserted caps (instance tripwires already satisfied by the submit path). | `authz.SeedTxIdentity`, `authz.Require` |
| Contract-first | Yes | Any route/DTO surface change (e.g. exposing that a route is "livre"/zero-stage) enters via `api/openapi/v1/openapi.yaml` + full regen. Likely minimal: stage list already serializes empty. | oapi-codegen, full `go generate ./...` |
| Multi-tenant pooled | Yes (inherited) | All touched tables already tenant-scoped; no new tables expected. | `tenant.FromContext`, `SeedTxTenant` |
| Async = transactional outbox | Marginal | If auto-approve triggers publish-side effects, they go through the existing outbox consumers exactly as a normally-approved instance does — no new inline network calls. | existing outbox repo |
| DB enforces invariants | **Yes — core** | Trigger rework ships as **forward migration 0316** (baseline is frozen post-fold 2026-07-29; never edit baseline). New rule: livre profile ⇒ active route allowed AND must carry **zero stages**; controlado/simples rules unchanged; both-directions (route writes AND reclassification) semantics preserved, DEFERRABLE kept. 55006 lesson applies to any data purge in 0316. | migration + trigger pattern from archived 0295 |
| Cross-module via published interface | Yes | `RoutePolicy` published-language value is the only thing that crosses taxonomy→approval; rename/replace the value, never leak `GovernanceClass` into approval. | `taxonomy/domain` port |

No AS-1.

## 4. Capability wiring

**N/A** — no capability added or changed. Submit/approve/publish capabilities remain as-is; the
livre path must NOT mint a bypass capability (it is a route *shape*, not a permission).

## 5. Module wiring

**N/A** — feature inside existing modules `taxonomy` + `approval`.

## 6. Frameworks to reuse, not reinvent

- `TxRunner` (`Do`) — submit auto-approve runs inside the existing submit tx boundary.
- `authz.SeedTxIdentity` + `authz.Require` — unchanged submit path.
- `problem.Write` + codes — any new error (e.g. route-admin rejecting stages on a livre route)
  is a problem+json code, not ad-hoc.
- `audit.RecordTx` — auto-approve MUST leave an audit event in the submit tx (instant approval is
  still a state change of record).
- `testdb` factory — all integration tests; `//go:build integration`; R1–R4.
- Migration self-registration idiom — 0316 inserts its own `schema_migrations` row (0309 lesson).

## 7. Contract & data

- **OpenAPI:** expected minimal — route create/update already carries stages[]; a livre route is a
  route with zero stages bound to a livre profile. Possible additions: problem codes for
  (a) stages present on livre-profile route, (b) submit-instant-approve response shape if it differs.
  Taxonomy `has_active_template_route` semantics unchanged.
- **Migration 0316** (first post-fold forward migration — also validates the new empty-tail
  pipeline end-to-end): replace `enforce_profile_route_policy`/`assert_route_shape` livre arm
  ("no active route") with "active route must have zero stages"; keep DEFERRABLE INITIALLY DEFERRED;
  keep reclassification direction. No destructive data change expected (livre profiles today have
  zero routes by construction — nothing to purge; if a purge is ever added, disable the deferred
  policy triggers around it, 55006).
- **Destructive change?** No wire-contract break expected. `RoutePolicy` value replacement is an
  internal published-language change (Go-level), handled atomically in one commit.

## 8. Test & QA plan

- **Framework:** `testdb` factory, `//go:build integration`.
- **Gates that apply (feature subset):** DB-invariant gate (trigger both-directions matrix: livre
  route with stages rejected; controlado zero-stage rejected; reclassify livre→controlado with a
  zero-stage active route rejected; reclassify controlado→livre with staged active route rejected),
  contract gate (spec↔generated↔handlers), authz gate (no new caps, tripwires still pass on
  auto-approved instances), async/idempotency gate (submit idempotency unchanged; instant-approve
  idempotent on retry).
- **Evidence:** `go build ./...`, `go test ./...`, `go vet -tags integration ./...`,
  `test-integration.ps1` on approval+taxonomy+migrations packages, live QA on :80 (create livre
  profile → create zero-stage route → create doc → submit → instantly approved/publishable),
  review disposition + bounded defers.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/approval.md` + `wiki/modules/taxonomy.md` (+ `Last verified`);
  route-admin frontend doc if UI changes land in the same unit.
- **REQ IDs:** REQ-AUTHZ (capability-only, unchanged), REQ-DB-INVARIANT class (trigger as
  authoritative line) — cite from `wiki/architecture/backend-target-architecture.md` in review.
- **ADR required?** **Yes** — this reverses a standing ratified policy (0295's "livre ⇒ no route
  permitted", ADR 0073-era signature-policy model) and closes ADR 0086's open item (livre profiles
  can never own templates). New ADR (next number in `wiki/decisions/`), superseding the livre arm
  of the per-profile signature policy; must state: config-first universal, absence of route =
  misconfiguration always, livre route = configured zero-stage route, submit semantics =
  instant approve in the submit tx with audit event, creation gates unchanged.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to design; ADR mandatory; two semantic constraints locked.
- **Open hard-stops:** none (AS-1/AS-2/AS-3 all clear).
- **Locked constraints handed to brainstorming/design:**
  1. **ADR obrigatório** superseding the livre arm of 0295/ADR-0073-era policy + closing ADR 0086's
     open item; implementation blocked until the ADR is ratified by the operator.
  2. **No bypass path:** livre submit flows through the normal submit service and instance model
     (instance created and instantly approved, audit event recorded in-tx) — never a side door that
     skips instance/audit/tripwires. Uniformity is the point of the ruling.
  3. **Zero-stage shape, not a flag:** livre-ness of a route is derived from its profile's
     governance class + empty stages; do not add a redundant `auto_approve` boolean that can drift.
  4. **Migration 0316 forward-only**, baseline untouched; trigger stays DEFERRABLE
     both-directions; 55006 discipline if any purge appears.
  5. `RoutePolicy` published-language value replaced (`no_route_permitted` →
     route-required-zero-stages semantics); `GovernanceClass` never leaks into approval.
