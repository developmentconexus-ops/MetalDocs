# ADR 0022 — Authz Capability Coherence (single source of truth + explicit scope + area-scoped administration)

> **Status:** Accepted (in execution) — direction + governance model approved 2026-06-03; **Phase 1 complete 2026-06-03** (registry hygiene, behavior-neutral); Phases 2–6 pending. Supersedes nothing; **extends** ADR [`0007`](0007-two-tier-authz.md) (two-tier model + lint harness) and ADR [`0021`](0021-tenant-vs-platform-admin-separation.md) (capabilities, not role names, are the enforced boundary).
> **Last verified:** 2026-06-03
> **Scope:** Coherence of the authorization model across all layers — Go capability registry, route table, DB role→capability seed, OpenAPI vendor annotations, tier-2 area enforcement, delivery handlers, and wiki. Establishes the target architecture and the phased path to reach it without symptom-patching.
> **Out of scope:** Authentication / sessions (tier-0, see `wiki/modules/auth.md`); platform-vs-tenant admin split (ADR 0021); changing any OpenAPI **shape** (paths, params, request/response schemas) — this ADR only adds `x-authz-*` vendor extensions and changes enforcement, never the contract shape.
> **Key files:**
> - `internal/modules/iam/domain/model.go:50` — capability registry (`Capability` consts + `validCapabilities`)
> - `apps/api/cmd/metaldocs-api/permissions.go:85` — tier-1 route → capability table
> - `internal/modules/iam/authz/authz.go:51` — tier-2 `Require(ctx, tx, cap, areaCode)`; `"tenant"` skips area filter
> - `internal/modules/iam/infrastructure/postgres/user_area_repository.go:100,152,196` — membership tier-2 call sites (currently `"tenant"`)
> - `internal/modules/iam/delivery/http/routes_memberships.go:365,374` — `isMembershipDirectoryAdmin` / `canManageMembershipTarget` (hardcoded `RoleSystemAdmin` string — the lone role-string gate in any handler)
> - `scripts/api-lint/code_rules.go:103` — `authz-call-present` (driven by `x-authz-area`, currently dormant) + `tripwire-pairing`
> - `db/reference-data/0001_product_reference_data.sql` — `role_capabilities` baseline seed

## Context

A cross-module audit (2026-06-03) of every authorization enforcement point in `internal/modules/` surfaced that the authz model is **not one coherent system**. It is three different dialects sharing one vocabulary, plus silent drift between the five places that declare capabilities. The IAM area-membership 403 defect (an `area_admin` holding `membership.manage` is blocked by a hardcoded `system_admin`-only handler gate) is the one place the incoherence throws a loud error; the rest fails silently.

### Finding 1 — three enforcement dialects

| Dialect | Mechanism | Modules |
|---|---|---|
| **A. Pure capability, tenant-wide** | tier-1 + tier-2 with literal `"tenant"` (area filter OFF) | templates, taxonomy, controlled-documents, iam user-roles, **iam area-membership** |
| **B. Genuine area-scoped** | tier-2 with real `areaCode` + frozen eligible-actor snapshot + SoD | documents / approval (submit, signoff, publish, supersede, obsolete) |
| **C. Hardcoded role string in handler** | `role == RoleSystemAdmin` in delivery code | **iam area-membership only** |

Role-string literals are compared inside delivery handlers in **exactly one place**: `routes_memberships.go` (`canManageMembershipTarget`, `isMembershipDirectoryAdmin`). Every other module trusts the capability model. IAM area-membership is doubly deviant — dialect A **and** C — and its name promises dialect-B (area) semantics it never implements. That contradiction is the root of the reported bug.

### Finding 2 — the area machinery is built but dormant

ADR 0007's codegen-rejection amendment shipped an annotation-driven lint: `x-authz-area` (`source` body|path, `field`, `multi`) drives `authz-call-present`, and `tripwire-pairing` requires every mutating repository SQL to pair with an `authz.Require`. **But 0 operations carry `x-authz-area` today**, and every tier-2 call passes the literal `"tenant"` — so the area rule never fires and the tripwire passes trivially (a call is present; its *scope* is unchecked). The scaffold for explicit area enforcement exists and is unused.

### Finding 3 — the five sources of truth drift silently

| Drift | Detail |
|---|---|
| Dead caps | `workflow.review`, `workflow.approve` — defined in Go, never seeded, never routed |
| Inline/undefined caps | `doc.publish`, `doc.obsolete`, `template.archive` used as raw `Capability("…")` strings in `permissions.go`, absent from `validCapabilities` → a typo compiles clean |
| Naming split | `doc.supersede` vs the `document.*` prefix used by every sibling |
| Orphan grant | `route.manage` seeded to roles, no route consumes it (ADR 0018 deferral) |
| Stale OpenAPI | governance schema enumerates 3 capabilities not in the Go model |
| Wiki drift | `local-dev-credentials.md` claims `area_admin` holds `membership.grant` — a capability that **does not exist** (real: `membership.manage`) |

### Root cause

> There is no single, machine-enforced source of truth for the capability model, and "area-scoped vs tenant-wide" is a per-call-site stringly-typed accident rather than a declared, enforced property of each capability.

Every layer re-declares authz in its own idiom (Go consts, raw route strings, duplicated SQL seeds, hand-written wiki, occasional handler role-checks) with nothing forcing agreement. The membership bug, the dormant area machinery, and the drift table are all symptoms of that one cause. Patching `canManageMembershipTarget` alone fixes the 403 and leaves the disease.

## Decision

Adopt five principles and one governance model. Capability remains the boundary (ADR 0021); this ADR makes that boundary *coherent and enforced*.

1. **Single source of truth = the Go capability registry.** `validCapabilities` in `model.go` is authoritative. No capability may be referenced as an inline string in `permissions.go` or elsewhere — every route, handler, and seed maps to a typed `Capability` const. CI enforces `routeRules.capability ∈ validCapabilities` and `every Capability is seeded to ≥1 role`.

2. **Scope is an explicit, declared property of each capability/operation — not a per-call-site literal.** Each capability is classified **tenant-grade** or **area-grade**. Area-grade operations annotate `x-authz-area` (reviving the dormant machinery) and pass the *real* target area to tier-2; tenant-grade operations pass `"tenant"` deliberately and are marked `x-authz-skip-area`. The choice is reviewed once, declared, and lint-enforced — never an accident of which string a repository author typed.

3. **Authorization lives only in the two tiers. Handlers hold business invariants, not authz.** The hardcoded `role == RoleSystemAdmin` checks in `routes_memberships.go` are removed. Per ADR 0021, the system pivots on capabilities, which the backend enforces at tier-1 (route) and tier-2 (area). Handlers retain only genuine domain invariants that are *not* capability questions: self-grant lockdown (`isSelf`) and the cross-tenant existence guard (404, not 403).

4. **Administration is least-privilege and area-scoped where the resource is an area.** `area_admin`'s `membership.manage` is **area-scoped**: an `area_admin` may grant/revoke membership only within areas where they themselves hold a `membership.manage`-granting role (i.e. their own `user_process_areas` rows). `system_admin` remains tenant-wide via the existing tier-2 bypass. This matches the ISO-segregation driver (`wiki/modules/iam.md`), the approval precedent (dialect B), and closes the previously-masked escalation path where an `area_admin` could mint `area_admin`/`qms_admin` across the whole tenant.

5. **The model is bound by CI, not by discipline.** Tests bind registry ↔ routes ↔ seed ↔ OpenAPI ↔ wiki so drift becomes a red build, not a latent rot.

### Governance model (decision 4) — concrete mechanics, no new schema

"Managed areas" need no new table. `authz.Require(ctx, tx, membership.manage, <targetArea>)` already returns true exactly when the actor has an active `user_process_areas` row in `<targetArea>` whose role grants `membership.manage` — and `area_admin` is the only non-system role granted that cap. So changing the membership tier-2 calls from `"tenant"` to the target area code **is** the area-scoping, for free. `system_admin` still bypasses (tenant-wide). This is the smallest correct change consistent with the existing tier-2 semantics.

## Consequences

**Wins.**
- One enforcement model across all modules; the membership module stops being the outlier.
- Area-grade vs tenant-grade is declared and lint-checked; new write surfaces can't silently default to tenant-wide.
- Escalation path closed; `area_admin` administration obeys least privilege and ISO segregation.
- Drift between Go / routes / seed / OpenAPI / wiki becomes CI-detectable.
- Reuses the dormant `x-authz-area` + tripwire harness instead of new infrastructure.

**Costs.**
- IAM membership/people handlers are hand-rolled (pre-codegen), so the `authz-call-present` lint (which assumes codegen `req.Body`/`req.Params` shapes) cannot directly cover them. Until/unless IAM migrates to codegen, area-grade IAM ops are covered by dedicated unit/integration tests + `x-authz-skip-area` on the spec with a documented exception — **not** by forcing a codegen migration into this program.
- A small behavior change for `area_admin` directory listing (Phase 4): from "self-only" to "managed areas".
- Several phases touch shared contract metadata; each must re-enter the `metaldocs-backend-api` skill and run its verification gates.

## Implementation — phased plan

Each phase is independently shippable, root-cause-scoped, and gated. No phase changes OpenAPI **shape**. Phases touching routes/handlers re-enter `metaldocs-backend-api`.

### Phase 1 — Registry hygiene (no behavior change) ✅ COMPLETE 2026-06-03
- ✅ Promoted inline caps to typed consts: `CapDocumentPublish` (`doc.publish`), `CapDocumentObsolete` (`doc.obsolete`), `CapTemplateArchive` (`template.archive`) in `model.go` + `catalog.go` descriptions; replaced the `Capability("…")` literals at `permissions.go:168,186,187,189`. Exact string values preserved (no DB/tripwire impact).
- ✅ Removed dead caps `workflow.review`/`workflow.approve` (const + `validCapabilities` + `capabilityDescriptions`). Grep confirmed zero non-definition Go references first (the `documents/_artifacts/00-context.md` claim that `fillin_authz.go` imports them was **stale** — it does not). Registry size guard bumped 28→29 (−2 dead, +3 promoted).
- ✅ Added CI tests in `permissions_test.go`: `TestEveryRouteCapInRegistry` (every `routeRules[i].capability` ∈ `validCapabilities`) and `TestEveryCapSeededOrDeferred` (every registry cap seeded to ≥1 role OR in the deferred allow-list = the 4 routed-but-unseeded system-admin-bypass-only write caps `doc.publish`/`doc.obsolete`/`doc.supersede`/`template.archive`). Both proven to bite on injected drift, then reverted.
- ✅ Fixed wiki drift: `local-dev-credentials.md` `membership.grant` → `membership.manage` + authoritative-source annotation (also dropped the incorrect `doc.publish` area_admin grant). OpenAPI governance enum reconciliation **deferred to Phase 5** (binding CI) — not isolated/trivial enough for Phase 1.
- **Branch:** `feat/authz-coherence-p1-registry`.
- **Gates (all green):** `go build ./...` ✓; `go test ./apps/api/... ./internal/modules/iam/... -count=1` ✓; `go run ./scripts/api-lint api/openapi/v1/openapi.yaml .` → 455 violations, **unchanged vs base** (pre-existing dormant tripwire/AUTHZ-DRIFT debt per Finding 2; zero new) ✓; `npx @redocly/cli lint` → valid ✓.

### Phase 2 — Declare scope explicitly
- Add a capability scope classification (tenant-grade vs area-grade) as typed data in `domain` with a test asserting every capability is classified.
- Annotate `x-authz-area` on codegen-served area-grade ops (documents/approval already pass real `areaCode`; this makes the spec match reality and re-activates `authz-call-present` for them). Mark tenant-grade ops `x-authz-skip-area`.
- **Gates:** redocly lint; `go run ./scripts/api-lint` (now non-dormant); `go build/test`.

### Phase 3 — Membership area-scoping (closes the original defect, root-cause)
- `user_area_repository.go` `Insert`/`CloseActive`/`GrantAtomic`: pass the membership's `areaCode` to `authz.Require` instead of `"tenant"`.
- `routes_memberships.go`: delete `canManageMembershipTarget` role-string gate; rely on tier-1 (cap) + tier-2 (area). Keep `isSelf` (self-grant 403) and the cross-tenant 404 guard.
- Annotate `grantAreaMembership`/`revokeAreaMembership` with `x-authz-area` (body `areaCode` / query `areaCode`); document the hand-rolled-handler lint exception.
- Tests (`tests/unit/iam_memberships/` + an integration test with a real DB and seeded `user_process_areas`): `area_admin` CAN grant within a managed area; CANNOT grant outside it; CANNOT self-grant; `system_admin` unchanged (tenant-wide); cross-tenant still 404.
- **Gates:** `go test ./internal/modules/iam/... ./tests/unit/iam_memberships/... ./tests/integration/iam/...`; `go build ./...`; api-lint.

### Phase 4 — Directory/list area-scoping
- Replace `isMembershipDirectoryAdmin` (`RoleSystemAdmin` string) with a managed-areas scope: `system_admin` → tenant-wide directory; `area_admin` → memberships in their managed areas; others → self only.
- **Gates:** handler tests for all three scopes; `go build/test`.

### Phase 5 — Bind the sources in CI
- Extend `scripts/api-lint` / add tests: every area-grade op is annotated; every route cap is typed; seed ↔ registry parity; wiki cap names ↔ registry parity.
- **Gates:** the new checks run red on injected drift (prove they bite).

### Phase 6 — Wiki sync
- Dispatch `wiki-curator`: refresh `authz-tiers.md`, `iam.md`, ADR 0007 amendment cross-link, bump `Last verified`, index this ADR.

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| Patch `canManageMembershipTarget` to allow tenant-wide `area_admin` | Rejected | Symptom patch; ships the escalation surface; leaves the root cause and the other dialects/drift untouched. Violates project hard-stop + no-symptom-patching rules. |
| Build a new managed-areas table/service | Rejected | Unnecessary — `user_process_areas` + tier-2 already express it. YAGNI. |
| Force IAM codegen migration to get lint coverage | Rejected (for now) | Large, orthogonal; covered by tests + documented exception instead. Revisit as its own work. |
| Per-operation generated authz wrappers | Rejected | Already rejected in ADR 0007 (tx-coupling); the tripwire is the static guarantee. |

## References
- Cross-module authz audit, 2026-06-03 (this session) — three-dialect map, tier-2 areaCode survey, drift inventory
- ADR [`0007-two-tier-authz.md`](0007-two-tier-authz.md) — two-tier model; `x-authz-area` + lint harness (codegen-rejection amendment)
- ADR [`0016-view-grade-capabilities.md`](0016-view-grade-capabilities.md) — view/manage split; §Security Boundary documented the (now-revised) handler gate
- ADR [`0021-tenant-vs-platform-admin-separation.md`](0021-tenant-vs-platform-admin-separation.md) — capabilities, not role names, are the boundary
- `wiki/concepts/authz-tiers.md` — tier-1/tier-2 reference
- `scripts/api-lint/code_rules.go` — `authz-call-present`, `tripwire-pairing`
