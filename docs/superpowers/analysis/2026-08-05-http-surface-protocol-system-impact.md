# System-impact analysis — Unified HTTP surface protocol per module

**Date:** 2026-08-05
**Intent (one line):** Replace the hand-written delivery wiring (`routeHandlers` struct + `routeFamilies` mount table + `main.go` keyed literal) and the hand-maintained `permissions.go` `routeRules` mirror with a single per-module declaration of HTTP surface, from which both mux registration and the tier-1 PDP table are derived.
**Work type:** feature (cross-cutting platform framework — births a platform package, not a bounded-context module)
**Author:** developing-new-work skill
**Verdict:** 🔴 Red — **AS-2 unresolved** (see §10)

---

## 1. Classify & own

- **Work type:** feature. It creates a platform framework and rewires the composition root; it does not birth a bounded-context module. §5 is therefore N/A.

- **Owning "module(s)":** none of the 15. This lives at two seams neither of which is a bounded context:
  - `internal/platform/httprouter` — the protocol itself (already exists as of `54cc496b`, currently holding only the `Muxer` interface).
  - `apps/api/cmd/metaldocs-api` — the composition root that consumes it (`router.go`, `permissions.go`, `main.go`).

- **Stakeholder with veto:** `iam`. The tier-1 PDP contract is iam's — `iamdelivery.PermissionResolver`, `iamdelivery.Visibility`, `iamdomain.Capability`. Anything that changes how tier-1 is populated changes an iam-owned contract, governed by ADR 0022. iam does not *own* the wiring, but the wiring may not redefine iam's vocabulary.

- **Explicitly NOT owning:**
  - The 15 bounded-context modules. They are *consumers* of the protocol — they declare their surface; they do not define the declaration format. If any module ends up with a bespoke variant, the protocol has failed.
  - `documents` in particular. It is the largest delivery surface and the natural place to "just fix it locally"; a documents-private route descriptor would be the exact local maximum this work exists to remove.

- **Cross-module edges (with direction):**
  | Edge | Through what | Status |
  |---|---|---|
  | every module `→` `platform/httprouter` | implements the published protocol interface | new, clean (platform is the correct home for a primitive with 15 consumers — catalog §11) |
  | composition root `→` every module | already exists (constructs + mounts) | unchanged in direction; **shrinks** in surface (one interface instead of five call shapes) |
  | every module `→` `iam` (`iamdomain.Capability`, `iamdelivery.Visibility`) | published types | **new, 15 edges — flagged.** If the descriptor carries capability/visibility, every module's delivery layer gains a compile dependency on iam. Design must decide whether that is acceptable, or whether the capability binding stays out of the module and is derived from the spec at the composition root. **This is a first-order design question, not an implementation detail.** |
  | composition root `→` `iam` (PermissionResolver) | published | already exists |

- **Ambiguity?** No AS-3. The absence of a bounded-context owner is the *correct* answer here, not ambiguity: a concern with 15 consumers belongs to platform by the repo's own promotion rule.

## 2. Foundation verdict

**Base you'd build on: NOT sound. Three enumerations of the same truth, none of which agree, none of which is derived from another.**

Measured this run:

| Enumeration | Size | Source of truth? |
|---|---|---|
| `api/openapi/v1/openapi.yaml` operations | **147** `operationId`s | Declared as route truth by the contract-first invariant |
| Go mux registrations | 17 mount operations across 5 different call shapes (`RegisterRoutes` method, `Register`, `RegisterGenerated`, two package-level `RegisterRoutes` functions, one bare `mux.Handle`) | Partly generated from the spec, partly hand-written |
| `permissions.go` `routeRules` | **119** hand-typed rules | Hand-maintained; sole home of route→capability |

**Three facts that make this a patch, not a base:**

1. **The spec does not carry capability metadata at all.** Zero `x-` extensions across the whole spec (`grep -rn "x-metaldocs\|x-capability\|x-required-cap\|x-visibility" api/openapi/` → no matches). The route→capability binding exists *only* in `routeRules`. So the "generate tier-1 from the spec" idea is not a refactor of existing data — the data does not exist yet and must be authored.

2. **Five route families are not in the spec at all.** Verified: `/api/v1/auth/login`, `/api/v1/search`, `/api/v1/security/lockouts`, `/healthz`, `/api/v1/feature-flags` — all **NOT-in-spec**. These are the same five bare-pattern families (`auth`, `health`, `featureFlags`, `search`, `security`) that the route-coverage guard can only check loosely. **The contract-first invariant is therefore already violated today, pre-existing, for a material part of the business API surface.** `/healthz` and `/api/v1/metrics` are defensible infra exceptions; `/api/v1/auth/login`, `/api/v1/search` and `/api/v1/security/*` are not — they are ordinary API surface that was hand-added in Go.

3. **The guard I built in `54cc496b..e846d295` is a drift detector, not a fix.** By CLAUDE.md's own definition that is a local maximum: it optimizes inside a structure it never questioned. Two independent gate arms then found two MAJORs that are *not* bugs in the guard but consequences of the structure:
   - `conditionalRouteFamilies` (`router.go:85`) — a hand-typed fail-open exemption set that no test validates and that is currently a dead branch (`permissions_test.go:351` always populates `presence`). One added line silently disarms the completeness check.
   - `main.go:817-837` is a **keyed** struct literal — a field added to `routeHandlers` + `routeFamilies` + the fixture but omitted *there* compiles clean, is nil in production, and the test stays green. Worse, a nil pointer handler still registers its routes (method values bind without dereferencing), so the failure is a per-request 500, not a loud boot failure — which means `router.go:78-80`'s claim to the contrary is false.

   Both are the *same* root cause as the original defect: the same truth enumerated in more than one place, synchronized by hand.

**Global-maximum structure (named):** a **published HTTP-surface port in `internal/platform/httprouter`** — each module declares its routes once as data (method, path template, handler, and the authorization binding), the composition root iterates `[]httprouter.Module`, and **both** the mux registration **and** the tier-1 PDP table are *derived* from that single declaration. Construction and mounting become the same act, so "constructed but never mounted" stops being a detectable defect and becomes an unrepresentable state.

What that dissolves rather than patches: the `routeHandlers` struct · the `routeFamilies` table · `conditionalRouteFamilies` · the `main.go` literal hole · `routeRules`↔routes drift · HEAD method semantics declared in two places · and the transitional loose-check in `TestRouteCoverage`.

**Trade-off, stated honestly:** the declaration must be *generated from the spec*, not hand-written. A hand-written descriptor would be a **fourth** enumeration and would repeat the original defect at larger scale, across all 15 modules at once. Generating it requires (a) authoring capability/visibility metadata into the spec, and (b) the five off-spec families actually being in the spec. (b) is the CON-07/ARC-02 migration already named as the deleting structure for the current transitional label.

**⇒ AS-2 FIRES.** Not because the proposed work optimizes inside the patch — it explicitly does not — but because **the work as scoped is blocked on a prerequisite that is itself outside the scope.** If we build the protocol only for the codegen families and leave the five off-spec families on the old path, we ship a fourth enumeration and a two-regime delivery layer. That is optimizing inside the patch by another name. **The operator decides the scope.** See §10.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|---|---|---|---|
| **AuthZ = capabilities, never roles** | **Yes — centrally.** This work changes *how tier-1 is populated*, not what it decides. | Capabilities stay the vocabulary; `Visibility` stays iam's enum. Tier-2 `authz.Require` and the DB tripwire are untouched — the two-tier PDP shape is preserved, only tier-1's data source changes from hand-typed to derived. **Constraint: tier-1 must remain fail-closed** — an operation with no declared capability resolves to `VisibilitySessionRequired`, never public (`permissions.go:336-338` today). | `iamdelivery.PermissionResolver`, `iamdomain.Capability`, `iamdelivery.Visibility`. ADR 0022 governs. |
| **Contract-first (OpenAPI + oapi-codegen)** | **Yes — this is the invariant the work restores.** | Routes and now also their authz binding become spec-derived. | `api/openapi/v1/openapi.yaml`, per-module `api/cfg.yaml` + `gen.go`. **Pre-existing violation recorded in §2.2 — this work's whole value is closing it.** |
| **Multi-tenant pooled** | No | Registration/mounting is tenant-agnostic; tenant resolution stays in the middleware chain and in-tx GUCs. | `tenant.FromContext`, `authz.SeedTxIdentity` — unchanged. |
| **Async = transactional outbox** | No | No state writes, no external side effects. | N/A |
| **DB enforces invariants** | No | No schema change. The capability *format* tripwires (`ck_cap_format`, `ck_cap_not_legacy`) already constrain any capability string this work would carry. | `db/baseline/0001_current_schema.sql` — unchanged. |
| **Cross-module via published interface only** | **Yes.** | The protocol interface must live in `platform`, not in any module. 15 consumers ⇒ platform promotion is mandatory, not optional (catalog §11). **Risk flagged in §1:** the descriptor carrying capability creates 15 new module→iam edges; those go through published iam types, so they are legal — but 15 new edges is a design cost that must be weighed, not absorbed silently. | `internal/platform/httprouter` (exists). |

**No AS-1.** The work violates no invariant; it *restores* one. The pre-existing contract-first violation is recorded as a finding, not attributed to this work.

## 4. Capability wiring

**N/A** — no new capability is created. The work changes the *mechanism* by which existing capabilities are bound to routes, not the capability set. `TestCapabilityRegistrySize` is unaffected.

**One touchpoint does apply and is load-bearing:** touchpoint 3, *tier-1 route→cap*. It is exactly what this work replaces. Any design must keep touchpoints 4 (tier-2 `authz.Require` in-tx), 6 (DB tripwire) and 9 (CI capability-coherence, REQ-AUTHZ-5) intact and independently enforced — **tier-1 must not become the only gate.** A derived tier-1 that is more convenient must not tempt anyone to treat it as sufficient (CLAUDE.md: "tier-1 is not treated as a floor").

## 5. Module wiring

**N/A** — no bounded-context module is born. The work adds to an existing platform package (`internal/platform/httprouter`) and rewires the composition root.

## 6. Frameworks to reuse, not reinvent

| Primitive | Reuse decision |
|---|---|
| `internal/platform/httprouter.Muxer` | **Reuse and extend.** Already exists, already threaded through all ~31 registration signatures as of `54cc496b`. The protocol interface belongs here, next to it. |
| oapi-codegen `HandlerWithOptions` / `StdHTTPServerOptions{BaseRouter}` | **Reuse.** The generated per-module `ServeMux` interface needs only `HandleFunc` + `http.Handler`; `Muxer` already satisfies it. The protocol must not replace codegen — it must wrap it. |
| `iamdelivery.PermissionResolver` | **Reuse the port, replace the implementation.** The resolver signature `(method, path) → (Capability, Visibility)` stays; only how the table behind it is built changes. |
| `problem.New` / `problem.Write` | **Reuse.** Any new failure mode (e.g. unmountable module at boot) that surfaces over HTTP uses RFC 9457. Boot-time failures are not HTTP and should panic loudly instead. |
| `chain.go` middleware chain | **Do not touch.** `apps/api/cmd/metaldocs-api/chain.go:25`. The fixed request lifecycle is inherited; this work changes registration, never the chain. |
| `testdb` factory | Not needed — this is not DB integration work. |

**Genuinely new cross-cutting concern?** Yes — "a module's published HTTP surface" has no existing framework row. Per the catalog's own rule that is the signal to design a *new platform framework*, which is precisely what §2 names. It is not a one-off inline.

## 7. Contract & data

- **OpenAPI-first:** this is where the real work is.
  - **Add authorization metadata to the spec.** No `x-` extensions exist today. A vendor extension per operation (`x-metaldocs-capability`, `x-metaldocs-visibility`) is the contract-first-compatible home for the route→capability binding currently hand-typed in `routeRules`' 119 entries.
  - **Bring the five off-spec families onto the spec** (`auth`, `search`, `security`, `featureFlags`, and the business parts of `health`) — the prerequisite named in §2. `/healthz` and `/api/v1/metrics` stay deliberate infra exceptions and must be *declared* as such somewhere machine-readable, not implied by absence.
  - **Regeneration is all-or-nothing.** Any spec edit churns every module's embedded `swaggerSpec`; partial regeneration is forbidden drift.
- **Migration:** **none.** No schema change, no new table, no `tenant_id` question.
- **Destructive change?** Yes, in the wire-adjacent sense: `routeRules` is deleted, not deprecated. Per the house rule (*tudo fallback legacy é extermínio*) it is dropped clean, not kept as a fallback beside the derived table. But it may only be dropped **after** parity is proven — the expand/contract shape here is: derive the new table → assert byte-equal resolution against the old one for every (method, path) the mux registers → then delete. A parity gate, not a compatibility layer.

## 8. Test & QA plan

- **Canonical framework:** mostly *not* integration. This is composition-root and platform work — package tests in `apps/api/cmd/metaldocs-api` and `internal/platform/httprouter`. Any live-drive verification uses the real stack per the Docker methodology, not a bespoke harness. `//go:build integration` + `testdb` applies only if a test needs a live PDP against the DB tripwire.
- **QA gates that apply:** contract (spec regeneration + parity), authz (tier-1 derivation must not weaken any current rule — the parity gate *is* this gate), docs. **N/A:** multi-tenant isolation, async/idempotency, DB-invariant — nothing in those planes changes.
- **The parity gate is the definition-of-done, and it must be exhaustive, not sampled.** For every pattern the mux registers, the derived resolver and the current `routeRules` resolver must return identical `(Capability, Visibility)`. Any intentional difference (e.g. HEAD canonicalization) is listed explicitly as an accepted delta with its justification, in the commit.
- **Evidence shape:** `go build ./...` · `go vet -tags integration ./...` (mandatory — seam signatures change again) · `gofmt -l` · `go test ./...` · `.\scripts\check-system-runnable.ps1` · live drive of at least one route per visibility class. Plus a TDD red-proof for the parity gate itself.
- **Regression watch:** `go vet -tags integration` is non-optional here. Untagged `go test` does not compile `//go:build integration` files, and this work changes registration seams across all 15 modules.

## 9. Docs / ADR

- **Wiki:** `wiki/architecture/backend-api-structure.md` and `wiki/architecture/api-contract.md` both describe the current registration model and would become false — both need updating with refreshed `Last verified`. `wiki/architecture/api-design-system.md` likely too.
- **REQ IDs:** **not cited — deliberately.** `wiki/architecture/backend-target-architecture.md` was not read this run (targeted-verify budget spent on the two spec anchors, which were higher-value). Fabricating REQ IDs here would be exactly catalog Class 27, which this very work stream just committed once in a code comment. **Locked constraint: the design must open the target spec and cite real REQ IDs before any task list is written.**
- **ADR required? YES — two, and they are not the same decision:**
  1. **Authorization metadata in the OpenAPI spec.** Making the spec the source of truth for route→capability is a standing-policy change and touches ADR 0022's boundary (capability-oriented authz). ADR 0022 is *amended or referenced*, not superseded — the PDP shape is unchanged; only tier-1's data source moves.
  2. **The module HTTP-surface protocol** as a platform framework, with the five-call-shape status quo recorded as what it replaces.
  A third may be needed if the operator rules that the off-spec families stay off-spec — that would be a formal, recorded MUST-deviation from contract-first rather than the current silent one.

## 10. Verdict & locked constraints

**Verdict: 🔴 Red.**

**Open hard-stop: AS-2, unresolved.** The foundation is legacy (three unsynchronized enumerations, §2), and the proposed work is blocked on a prerequisite outside its own scope. Design cannot begin until the operator rules on **one question**:

> **Does this work include bringing the five off-spec route families (`auth`, `search`, `security`, `featureFlags`, `health`) onto the OpenAPI spec — or not?**
>
> - **(A) Yes — one program.** The protocol is built once, over the whole surface. Larger, slower, and it closes the pre-existing contract-first violation permanently. No second regime is ever created.
> - **(B) No — protocol first, migration later.** Faster to first value, but it ships a **two-regime delivery layer**: spec-derived for the codegen families, hand-declared for five. That hand-declared side is a fourth enumeration. Legal **only** under CLAUDE.md §2(b): explicitly labelled transitional, naming the deleting milestone, with deletion as that milestone's definition-of-done. Without a named milestone this is an unlabelled local maximum and therefore a defect.
> - **(C) Neither — the off-spec families stay off-spec permanently.** A formal, recorded MUST-deviation from contract-first via its own ADR, rather than today's silent violation. Honest, but it caps what the protocol can ever guarantee.

Two decisions already standing from the round-4 gate, unchanged and independent of the above:

- **HEAD bypasses tier-1** (`permissions.go:35-38`). No `routeRules` row carries HEAD; `matches` demands exact method equality; Go's `ServeMux` routes HEAD to GET patterns. Result: HEAD reaches the handler with only `VisibilitySessionRequired`, skipping the route's capability. *(Nuance the gate added: methodless rows — `r.method == ""` — do match HEAD, so the fall-through is confined to paths covered only by method-qualified rows.)* Pre-existing, invisible to route-coverage tooling by construction. The unified protocol dissolves it (method semantics declared once); until then it needs its own fix.
- **The transitional label in `permissions_test.go:424-431` has no deleting milestone.** CLAUDE.md §2(b) requires label + global-max structure + milestone; the first two are in-repo, the third is the operator's. The gate additionally found the label's discharge criterion is **wrong**: it names five families, but `metrics` is also mounted bare (`router.go:124`), so migrating five would not actually delete the loose branch. Six, not five.

**Locked constraints handed to brainstorming (binding, once AS-2 clears):**

1. **The declaration is generated from the spec, never hand-written.** A hand-authored descriptor is a fourth enumeration and repeats the original defect across 15 modules at once. This is the single most important constraint and the one most likely to be quietly relaxed under implementation pressure.
2. **Tier-1 stays fail-closed.** An operation with no declared capability resolves to `VisibilitySessionRequired` — never public, never "permissive by default because the table is now derived".
3. **Tier-2 and the DB tripwire are untouched and remain independently sufficient.** A more convenient tier-1 must not become the only gate (ADR 0022).
4. **The protocol interface lives in `internal/platform/httprouter`.** 15 consumers ⇒ platform, not a module. Any module-private variant is a failure of the whole exercise.
5. **`routeRules` is deleted, not deprecated** — after an exhaustive parity gate over every registered pattern, with every intentional delta listed explicitly. No compatibility layer, no fallback.
6. **The middleware chain (`chain.go:25`) is not touched.**
7. **Regeneration is all-or-nothing** — partial `swaggerSpec` regeneration is forbidden drift.
8. **Cite real REQ IDs** from `wiki/architecture/backend-target-architecture.md` before writing any task list. Do not invent them.
9. **Decide the 15 new module→iam edges deliberately** (§1). Either accept them with a stated reason, or keep the capability binding out of the modules and derive it at the composition root. Do not let this be settled by whichever is easier to type.
10. **Two ADRs are in scope** (§9), and the work is not done when the code compiles — it is done when the spec, the ADRs, and the three affected `wiki/architecture/*` docs are all true.

**Do not invoke `superpowers:brainstorming` until AS-2 is resolved.**
