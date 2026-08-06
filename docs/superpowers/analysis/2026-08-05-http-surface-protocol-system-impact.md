# System-impact analysis — Unified HTTP surface protocol per module

**Date:** 2026-08-05
**Intent (one line):** Replace the hand-written delivery wiring (`routeHandlers` struct + `routeFamilies` mount table + `main.go` keyed literal) and the hand-maintained `permissions.go` `routeRules` mirror with a single per-module declaration of HTTP surface, from which both mux registration and the tier-1 PDP table are derived.
**Work type:** feature (cross-cutting platform framework — births a platform package, not a bounded-context module)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow — AS-2 **resolved by operator ruling A**, 2026-08-05 (see §10). Two ADRs required; ten locked constraints carried into design.

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

> ### ⚠️ CORRECTION — 2026-08-05, after the ruling
>
> **Two of the three measurements below were wrong when first committed (`c19f1e46`), and the operator's ruling A was made on that faulty basis.** Both errors were mine and both were the same mistake: I grepped the spec for runtime paths without accounting for `servers: - url: /api/v1` (spec line 7). The spec writes `/auth/login`; I searched for `/api/v1/auth/login` and read the absence as evidence. The tell was that `/api/v1/documents` also came back "NOT-in-spec" — a route that is unquestionably generated — which is what exposed the error.
>
> **What is actually true is stated below.** The core thesis — three unsynchronized enumerations, structure is the defect, restructure is the answer — **survives intact and is unchanged**. What changes is the *kind* and the *size* of the contract-first violation, and therefore the cost of ruling A. See §10 for how the ruling is affected.

**Base you'd build on: NOT sound. Three enumerations of the same truth, none of which agree, none of which is derived from another.**

Measured this run (corrected):

| Enumeration | Size | Source of truth? |
|---|---|---|
| `api/openapi/v1/openapi.yaml` | **147** `operationId`s across **125** paths and **16** tags | Declared as route truth by the contract-first invariant |
| Go mux registrations | 17 mount operations across 5 different call shapes (`RegisterRoutes` method, `Register`, `RegisterGenerated`, two package-level `RegisterRoutes` functions, one bare `mux.Handle`) | Partly generated from the spec, partly hand-written |
| `permissions.go` `routeRules` | **119** hand-typed rules | Hand-maintained; sole home of route→capability |

**Three facts that make this a patch, not a base:**

1. **The spec carries authz metadata, but not the capability.** *(Corrected — my original claim of "zero `x-` extensions" was false.)* The spec already uses vendor extensions freely: `x-pagination-exempt` + `x-pagination-exempt-reason`, `x-websocket-message`, and — directly relevant — **`x-authz-area`** and **`x-authz-area-none`** (lines 1108, 1144, 1705). So there is an established convention for authz metadata in the spec, and the design does not have to invent one.

   More importantly, **the spec already expresses two of the three visibility states natively**: a global `security: - sessionCookie: []` default (line 11-12) with per-operation `security: []` overrides marking the genuinely public endpoints. That is `VisibilitySessionRequired` and `VisibilityPublic` in standard OpenAPI, already authored, already reviewed.

   What is missing is exactly one thing: **the capability name for `VisibilityPermissionGuarded` operations.** That is the only datum that lives solely in `routeRules` and must be authored into the spec. This is a substantially smaller authoring job than "the data does not exist yet".

2. **Every production route is already in the spec. The violation is that the Go side ignores it for six tags.** *(Corrected — my original claim that five families were absent from the spec was false; all of them are present.)*

   | Tag | In spec | `api/cfg.yaml` | Generated server wired |
   |---|---|---|---|
   | approval, audit, controlled-documents, distribution, documents, iam, notifications, taxonomy, templates, tokens (10) | ✅ | ✅ | ✅ |
   | **security** | ✅ | ✅ | ❌ **dead generated code** — a 29,750-byte `internal/modules/security/api/api.gen.go` exists and `grep -rn "securityapi"` finds **zero** references outside its own package. The handler hand-registers bare patterns beside it. |
   | **auth** | ✅ | ❌ | hand-written |
   | **configuration** (`/feature-flags`) | ✅ | ❌ | hand-written |
   | **health** (`/health/live`, `/health/ready`) | ✅ | ❌ | hand-written |
   | **observability** (`/metrics`) | ✅ | ❌ | bare `mux.Handle` in the composition root |
   | **search** (`/search/documents`) | ✅ | ❌ | hand-written |

   The only runtime routes genuinely absent from the spec are `/internal/test/*` — the e2e scaffolding, deliberately gated behind `METALDOCS_E2E` and deliberately excluded.

   **So the contract-first violation is real but is a different defect than I reported.** It is not "routes were hand-added in Go without a spec entry" (nothing is). It is: **the spec declares these operations, and for six tags the Go side ignores that declaration and hand-writes the routes anyway — including one tag where the generated server already exists and is simply dead code.** That is a *drift* violation, not an *omission* violation, and it is materially cheaper to close: five `cfg.yaml` files plus wiring six handlers onto their generated `ServerInterface`, rather than authoring new spec surface.

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

- **OpenAPI-first:** this is where the real work is. *(Corrected per §2 — the spec is complete; the Go side is what diverges.)*
  - **Add the capability to the spec, and only the capability.** Visibility is already expressible and already authored: the global `security: - sessionCookie: []` default (spec lines 11-12) plus per-operation `security: []` overrides give `VisibilitySessionRequired` and `VisibilityPublic` in standard OpenAPI. One vendor extension is needed for the third state — the capability name on `VisibilityPermissionGuarded` operations. It follows the existing `x-authz-area` / `x-authz-area-none` convention (spec lines 1108, 1144, 1705) rather than inventing one. **Open design question, not settled here:** whether visibility is *derived* purely from `security` or is *also stated explicitly*. Deriving is the DRY answer; stating is the auditable answer. Pick one deliberately.
  - **Close the six-tag Go-side drift** — give `auth`, `configuration`, `health`, `observability` and `search` an `api/cfg.yaml` + `gen.go`, and wire `security` onto the generated server that already exists as dead code. `/internal/test/*` is the only deliberate spec exclusion and is already gated behind `METALDOCS_E2E`.
  - **Regeneration is all-or-nothing.** Any spec edit churns every module's embedded `swaggerSpec`; partial regeneration is forbidden drift.
- **Migration:** **none.** No schema change, no new table, no `tenant_id` question.
- **Destructive change?** Yes, in the wire-adjacent sense: `routeRules` is deleted, not deprecated. Per the house rule (*tudo fallback legacy é extermínio*) it is dropped clean, not kept as a fallback beside the derived table. But it may only be dropped **after** parity is proven — the expand/contract shape here is: derive the new table → assert byte-equal resolution against the old one for every (method, path) the mux registers → then delete. A parity gate, not a compatibility layer.

  > **AMENDED 2026-08-05 by the design's §10.1 — this constraint was wrong.** The parity gate
  > described above treats `routeRules` as an oracle. It is not one: it is one *attempt to express*
  > which capability each route requires, and the spec annotations are a second attempt. Neither is
  > the authority; the authority is a decision, and a decision has a home rather than a proof.
  > Seven adversarial rounds landed on this gate, each finding a real defect and none reducing the
  > count or the altitude, because comparing two non-authoritative artifacts yields differences
  > forever and truth never. Worse, the gate was a **fourth** hand-synced enumeration of route truth
  > compared against two of the three this program exists to delete. The deletion license is now
  > positive conformance against the single authority ruling A establishes — see the design §10.
  > **The four operator rulings are untouched.** This constraint was authored in this program,
  > before any design existed, and is corrected in it.

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

**Verdict: 🟡 Yellow.** AS-2 raised and **resolved by operator ruling, 2026-08-05.**

### AS-2 — raised and resolved

The foundation is legacy (three unsynchronized enumerations, §2), and the work was blocked on a prerequisite outside its own scope. The question put to the operator:

> **Does this work include bringing the five off-spec route families (`auth`, `search`, `security`, `featureFlags`, `health`) onto the OpenAPI spec — or not?**
>
> - **(A) Yes — one program.** The protocol is built once, over the whole surface. Larger, slower, and it closes the pre-existing contract-first violation permanently. No second regime is ever created.
> - **(B) No — protocol first, migration later.** Faster to first value, but it ships a **two-regime delivery layer**: spec-derived for the codegen families, hand-declared for five. That hand-declared side is a fourth enumeration. Legal **only** under CLAUDE.md §2(b): explicitly labelled transitional, naming the deleting milestone, with deletion as that milestone's definition-of-done. Without a named milestone this is an unlabelled local maximum and therefore a defect.
> - **(C) Neither — the off-spec families stay off-spec permanently.** A formal, recorded MUST-deviation from contract-first via its own ADR, rather than today's silent violation. Honest, but it caps what the protocol can ever guarantee.

**RULING: (A) — one program, whole surface.** Operator, 2026-08-05.

> **⚠️ The ruling was made on a faulty measurement.** See the correction box in §2. The question put to the operator said the five families were "off-spec" and framed ruling A as the larger, slower path that would have to *author new spec surface*. That was wrong: every one of those families is already in the spec. Ruling A's actual cost is five `cfg.yaml` files plus wiring six handlers onto generated servers — one of which (`security`) is already generated and merely unwired.
>
> **The ruling stands and is unaffected in direction.** It chose "one program over the whole surface, no two-regime state" — a scope decision that remains correct and is now *cheaper* than presented. Option (C) (formalize a permanent deviation) is moot: there is nothing off-spec to formalize. Option (B)'s premise — that including the five families is what makes the program large — was the faulty part.
>
> **What the operator should re-confirm, and nothing more:** ruling A was chosen against a cost that was overstated. It is not being re-opened. It is recorded here that the choice would have been at least as attractive under the true numbers, so no re-ruling is required unless the operator wants one.

**What ruling A binds, beyond clearing the block:**

- **The five off-spec families are in scope, not adjacent to it.** `auth`, `search`, `security`, `featureFlags`, and the business surface of `health` are brought onto `api/openapi/v1/openapi.yaml` as part of this work. Their absence is the pre-existing contract-first violation recorded in §2.2; ruling A closes it rather than routing around it.
- **No second regime may be created, at any point, including intermediate commits.** There is no "codegen families first, legacy families after" phase that ships. Intermediate states may exist inside the work; none of them is a release boundary that leaves two declaration mechanisms live.
- **CLAUDE.md §2(b) does not apply and must not be invoked.** Ruling A is outcome **(a) restructure now**. Any proposal during design to label something transitional and defer it is a re-litigation of a settled decision and must be surfaced to the operator, not absorbed.
- **The existing transitional label in `permissions_test.go:424-431` is discharged by this work, not carried.** Its deleting milestone — the open item from the round-4 gate — is **this program**. Its discharge criterion is corrected from five families to six (`metrics` is also mounted bare, `router.go:124`).
- **The HEAD tier-1 defect is absorbed, not deferred.** Ruling A makes method semantics a single declaration, which dissolves it. It stops being a standing operator-decision item and becomes an acceptance criterion: HEAD must resolve to its GET row's capability, and that must be an explicitly listed delta in the parity gate (§8), not a silent behavior change.
- **Two MAJORs from the round-4 gate are closed by dissolution, not by patch.** `conditionalRouteFamilies` (`router.go:85`) and the `main.go:817-837` keyed-literal hole both cease to be representable once construction and mounting are the same act. **Constraint: they may not be patched in the interim** — that would be the third consecutive patch on the same construct, which the adversarial-review ratchet (§1) forbids. If the program is interrupted before they dissolve, they revert to open findings requiring their own disposition.
- **Scope cost is accepted knowingly.** Ruling A is the larger and slower path. That was stated when the question was put and is not grounds for re-scoping later without a new ruling.

### Residual items after the ruling

- **Three factual errors in comments authored by this work stream, unfixed and independent of the ruling.** They are cheap, they are wrong regardless of which path was chosen, and they should be corrected before or alongside the first design commit: `router.go:72` names `buildPresence`, which does not exist (the function is `startPresence`, `main.go:1168`) — Class 27, committed in the change that edits the Class 27 catalog; `permissions_test.go:415-417,431` says five bare-pattern families when `metrics` makes six; `router.go:78-80` claims a missing family fails loudly at boot, which is false for pointer-receiver handlers (method values bind without dereferencing, so the failure is a per-request 500).
- **Catalog residue** flagged by the round-4 gate: `defect-class-catalog.md:12-15` still frames all of Part II as prevented by review-protocol mechanisms, contradicting Class 34's rung-4 structural prevention (the Part II preamble at `:1089-1092` reconciles this; the header was not updated), and the "four rounds" headline at `:1100`/`:1445` silently excludes work item 2's three rounds.
- **R3-5 remains open and is now a design input, not a defect:** the per-family floor is `>0`, so route *deletion* is invisible (the escalation direction — a registered pattern with no rule — is strictly caught). Under ruling A a derived declaration makes deletion a spec diff, which is reviewed by construction. Confirm during design that this is genuinely dissolved rather than assumed.

Two decisions already standing from the round-4 gate, unchanged and independent of the above:

- **HEAD bypasses tier-1** (`permissions.go:35-38`). No `routeRules` row carries HEAD; `matches` demands exact method equality; Go's `ServeMux` routes HEAD to GET patterns. Result: HEAD reaches the handler with only `VisibilitySessionRequired`, skipping the route's capability. *(Nuance the gate added: methodless rows — `r.method == ""` — do match HEAD, so the fall-through is confined to paths covered only by method-qualified rows.)* Pre-existing, invisible to route-coverage tooling by construction. The unified protocol dissolves it (method semantics declared once); until then it needs its own fix.
- **The transitional label in `permissions_test.go:424-431` has no deleting milestone.** CLAUDE.md §2(b) requires label + global-max structure + milestone; the first two are in-repo, the third is the operator's. The gate additionally found the label's discharge criterion is **wrong**: it names five families, but `metrics` is also mounted bare (`router.go:124`), so migrating five would not actually delete the loose branch. Six, not five.

**Locked constraints handed to brainstorming (binding — AS-2 cleared by ruling A):**

0. **Ruling A governs scope.** One program over the whole surface; the five off-spec families are in scope; no two-regime state ships. This constraint outranks the rest — if any other constraint appears to require a transitional split, that is a contradiction to surface, not to resolve locally.

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

**Handoff: AS-2 cleared by ruling A. `superpowers:brainstorming` is invoked with this document as its locked constraints.**
