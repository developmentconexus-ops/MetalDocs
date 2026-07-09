# System-impact analysis — Delegation-aware instance-view visibility (F2d.1)

**Date:** 2026-07-08
**Intent (one line):** `requireInstanceVisible` (documents/approval read path) composes on `domain.CheckEligibility` + `domain.ResolveEligibleIdentity` (ADR 0077 primitive) so an active delegate of a stage-pool member can load the instance they are eligible to act on — required by F2d.1's operator-approved Validation Gate (`TestViewerBlock_Delegate`, ADR 0078 viewer facts).
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(proceed; ADR amendment + one named risk carried — see §10)*

---

## 1. Classify & own

- **Work type:** feature (read-path change inside an existing module; no new routes for this change itself).
- **Owning module(s):** `documents/approval` (nested exception inside `documents`, ADR 0072) — it owns `requireInstanceVisible`, the instance aggregate, the delegation table (`approval_delegations`, migration 0293), and the eligibility/SoD domain predicates. The change is entirely within this module's application + domain layers.
- **Explicitly NOT owning:**
  - `iam` — no capability is added or changed; `authz.Require`/`MustActorID` are consumed via their published API exactly as before (ADR 0077 decision 2: delegation grants no capability).
  - `documents` (parent) — document lifecycle, publish, edit paths untouched.
  - `controlleddocuments` — consumed only via the existing `CDFieldReader` port (area resolution), unchanged.
- **Cross-module edges (with direction):** `documents/approval → iam/authz` (published functions `Require`, `MustActorID` — pre-existing edge, unchanged); `documents/approval → controlleddocuments` (published `CDFieldReader` port — pre-existing, unchanged). `LoadActiveDelegationsFor` is the approval module's **own** repository method (`infrastructure/approval_repository.go:196`) — no new cross-module edge.
- **Ambiguity?** None. AS-3 not tripped.

## 2. Foundation verdict

- **Base you'd build on:** `requireInstanceVisible` (`read_service.go`), the F8/ADR 0075 visibility gate: hand-rolled membership loop over `eligible_actor_ids` + capability fallbacks (oversee, edit-in-area).
- **Sound, or legacy/patch/workaround?** **The visibility SET is sound** (author ∪ pool ∪ oversee ∪ edit — matches ServiceNow/GitHub visibility semantics). **The implementation is a pre-delegation artifact**: ADR 0075 (F8) shipped before ADR 0077 (F9, delegation); F9 widened the write path's eligibility input but never taught the visibility gate. Proven live-DB consequence: a delegate **can sign** a stage (`decision_service.go:288-292`, delegation-aware) but gets `ErrInstanceNotVisible` → 404 **loading** the instance (`TestViewerBlock_Delegate`, real Postgres, 2026-07-08). Can-act-but-cannot-see. Second field defect of the same class as the M2c 412 (eligibility re-derived per surface → divergence).
- **Global-maximum structure (named):** one relationship-evaluation kernel per aggregate — `EvaluateAccess(viewer, inst, delegations) → AccessDecision{CanView, CanAct, OnBehalfOf, IsAuthor, HasSigned}` — with act/view/DTO-facts/worklist as projections of that ONE evaluation. Market-validated invariant (Google Zanzibar/Drive `capabilities`, Camunda IdentityLinks + `delegateTask`, ServiceNow `sysapproval_approver` + `sys_user_delegate`, AWS Cedar single evaluator, Veeva Vault `lifecycle_actions`): eligible-to-act ⟹ able-to-see through one evaluation, divergence structurally impossible.
- **AS-2 judgment:** NOT tripped. This work does not optimize *inside* the patch — it **removes the patch's divergent rule** and substitutes the single primitive (`CheckEligibility` direct fast path, `ResolveEligibleIdentity` delegation fallback) at the one chokepoint all three view readers share. Full kernel unification (collapsing the remaining three composition points + pinning the worklist SQL) is **M3's chartered scope** (milestone.md Rabbit holes: "Approval kernel extraction … that is M3"; governing spec §5). Trade-off recorded: until M3, three call sites still *compose* the shared primitives (write services, visibility, viewer-facts) — composition duplication, not rule duplication; no divergent rule remains after this change.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes | No new rule, no role reasoning. Pool membership evaluated via the single existing eligibility primitive; capability fallbacks (`CapApprovalOversee` tenant-grade, `CapDocumentEdit` area-grade) unchanged via `authz.Require`. Delegation grants no capability (ADR 0077 §2). | `authz.Require`, `authz.MustActorID`, `domain.CheckEligibility`, `domain.ResolveEligibleIdentity` |
| Contract-first (OpenAPI + oapi-codegen) | Indirect | The visibility change adds/changes NO route or DTO. (F2d.1's `viewer` block — the consumer of this fix — is already OpenAPI-first + regenerated, per its own spec.md.) | n/a for this change |
| Multi-tenant pooled | Yes | `LoadActiveDelegationsFor(ctx, tx, tenantID, actorID, asOf)` is tenant-scoped (tenant-isolation proven: `delegation_repository_integration_test.go:259-297`); cross-boundary remains `ErrInstanceNotVisible` → 404, never 403. Runs in the view tx whose GUCs are seeded by the TxRunner chokepoint. | existing repo method; TxRunner GUC seeding |
| Async = transactional outbox | No | Read-only change; no side effects, no network calls. | n/a |
| DB enforces invariants (triggers/constraints) | No | Read path; no writes, no new tables/columns/triggers. Write-side tripwires (SoD trigger 0290, signoff caps arm) untouched. | n/a |
| Cross-module via published interface only | Yes | No new edges; own-module repository read + published iam/authz functions. | existing ports |

**Supporting invariants:** H-PRE-1 — the delegation SELECT is a plain, non-recording read inside the plain view tx; no advisory lock is held on this path; `authz.Require` inside the view tx is the pre-existing, RW-tx-compliant pattern (api-lint `authz-require-rw-tx` green). TxRunner — all three callers already run under `runner.Do`. AS-1 not tripped.

## 4. Capability wiring

**N/A** — no capability added or changed. Delegation is eligibility-widening, never a capability (ADR 0077 decision 2); `TestCapabilityRegistrySize` unaffected.

## 5. Module wiring

**N/A** — no new module. Change lives in `documents/approval` application (+1 doc-comment in domain).

## 6. Frameworks to reuse, not reinvent

| Primitive | Used? | How |
|---|---|---|
| `TxRunner` (`Do`) | Yes | All three view paths already open the tx via `runner.Do`; visibility runs inside it. No new tx handling. |
| `tenant.FromContext` / GUC seeding | Yes | Identity flows via the TxRunner chokepoint (`seedTxIdentityFromContext`) → `authz.MustActorID` reads the tx GUC. No hand-threaded ids. |
| `authz.Require` | Yes | Capability fallbacks unchanged. |
| `problem.Write` / error mapping | Yes (unchanged) | `ErrInstanceNotVisible` → existing 404 mapping (`http/errors.go`). No new error shape. |
| `testdb` factory | Yes | The 7 real-DB viewer scenarios + the 9 pre-existing tenant-grade visibility tests all on `testdb.Open` + builders + `SeedWithCaps`/`SetCapsOnTx`, `-tags integration`. |
| Outbox / audit / strictjson | N/A | No side effects, no state change, no request decoding. |

No hand-rolled equivalents introduced. The change *deletes* a hand-rolled membership loop in favor of the domain primitive — the reuse direction the catalog demands.

## 7. Contract & data

- **OpenAPI-first:** no route/DTO change in this fix. (Viewer block already spec'd + regenerated under F2d.1.)
- **Migration:** none. No schema change.
- **Destructive change?** No. Visibility **widens** monotonically: only active delegates of current-or-past pool members gain view — exactly the set that can already ACT on the instance (write path). No existing consumer loses access; deny behavior (404) unchanged for everyone else. Security posture: no information becomes visible beyond what acting already requires — this is the eligible-to-act ⟹ able-to-see closure, not a broadening beyond it.

## 8. Test & QA plan

- **Canonical framework:** `testdb` integration factory, `//go:build integration` (R1–R4). Drive-by repair recorded: `ctxWithIdentity` in `read_service_tenant_grade_view_integration_test.go` was latently broken (seeded only the iam ctx key, never the platform actor key the TxRunner chokepoint reads → every test in that file failed on real DB); repaired to seed `tenant.WithActorID` — pre-existing defect, canonical-framework-compliant repair.
- **QA gates applying (feature subset):** authz gate (visibility = authz-adjacent read gate — regression suite below), multi-tenant isolation gate (delegation repo isolation test exists), contract gate (F2d.1 regen/lint — owned by the feature), docs gate (§9). Async/idempotency + DB-invariant gates N/A (read-only).
- **Evidence (commands + expected):**
  - `go build ./...` — PASS (already verified, BUILD_EXIT=0).
  - `go vet ./internal/modules/documents/approval/...` + package unit tests — compile/regression of test fakes.
  - Real-DB: `go test -tags integration ./internal/modules/documents/approval/application/... -run 'TestViewerBlock'` — all 7 scenarios (6/7 already green; Delegate is the one this change closes).
  - Real-DB regression: `-run 'TestLoadInstance|TestLoadActiveInstanceByDocument'` — the 9 F8 visibility tests must stay green (bare-viewer DENIED, wrong-area edit DENIED, author/oversee/edit-own-area GRANTED, system_admin GRANTED) — proves the widening did not loosen any deny case.
  - Full evidence lands in `f1-viewer-contract/evidence.md`, fixture-vs-real labeled.

## 9. Docs / ADR

- **Wiki:** feature-class — update the approval section of the owning module doc + `Last verified` refresh at milestone close (wiki-curator pass); no new module doc.
- **REQ IDs cited:** authz-coherence family of `wiki/architecture/backend-target-architecture.md` (ADR 0022 boundary: single source of "who may act"); no MUST-deviation — this change *restores* coherence between two ratified decisions (0075 visibility set × 0077 delegation).
- **ADR required?** No **new** ADR. **Amendment required:** ADR 0078 (viewer-facts contract, created by F2d.1) gains a paragraph recording that `requireInstanceVisible` now composes on the same primitive — completing ADR 0075's visibility set with ADR 0077's delegation semantics (eligible-to-act ⟹ able-to-see). Policy is unchanged; two accepted policies are being made mutually consistent. Flagged as the Yellow condition.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed. No hard-stop; one ADR amendment and one named risk carried.
- **Open hard-stops:** none. AS-1 clean (no invariant violated), AS-2 judged not-tripped (§2 — removes the divergent rule rather than optimizing inside it; kernel = M3 charter), AS-3 clean (owner unambiguous).
- **Locked constraints handed forward (into F2d.1 plan.md execution notes / M3):**
  1. **Never a second membership/SoD rule.** Visibility, viewer-facts, and act-eligibility MUST all compose `CheckEligibility`/`ResolveEligibleIdentity` (+ `CheckSoD` where SoD applies). Any new surface answering "how does this user relate to this instance" consumes the primitive.
  2. **No new capability, no write-path change** (M2d appetite; ADR 0077 §2).
  3. **ADR 0078 amended** to record the visibility convergence (this analysis is the source).
  4. **Named risk → M3:** the worklist/inbox SQL is still an unpinned re-expression of membership (the 4th projection); M3 kernel extraction must either derive it from the kernel or pin it with a Go≡SQL parity test (precedent: SoD app-predicate ↔ DB-trigger mirror, tripwire parity lints). Until then it is the next most likely drift-defect site.
  5. **F8 deny-regression is part of the gate:** the 9 tenant-grade visibility tests must stay green alongside the 7 viewer scenarios.
