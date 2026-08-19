# R10-T8B — Adjudicated Corrected Candidate — Independent Fable Delta Review (Round 2)

```text
INDEPENDENT DELTA REVIEW
EVIDENCE ONLY
NOT TARGET AUTHORITY
```

> **Status:** BOUNDED ROUND-2 DELTA REVIEW / NON-AUTHORITATIVE
> **Date:** 2026-08-19
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed remote HEAD:** `0b2b9fe9be48282d8f6cdf6f22e52d22aa8a6f66`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Workflow:** canonical Standard Fable review workflow, `developmentconexus-ops/conexus-methodology/README.md`
> **Stage:** T8-B ACTIVE
> **Implementation:** BLOCKED

This document is reviewer evidence. It is not target authority, not T8-B promotion, not T8-C opening and not implementation authorization. The Lead adjudicates; the operator ratifies.

---

## 0. Provenance

### 0.1 HEAD revalidation

The task prompt asserted an expected HEAD. It was not trusted; it was independently revalidated before any judgment:

```text
gh pr view 131 --json headRefOid   → 0b2b9fe9be48282d8f6cdf6f22e52d22aa8a6f66
                                     state OPEN / base main
git fetch origin docs/a8-authz-approval-redesign-ledger
git rev-parse origin/…             → 0b2b9fe9be48282d8f6cdf6f22e52d22aa8a6f66
```

The assertion was correct. **Reviewed HEAD = `0b2b9fe9be48282d8f6cdf6f22e52d22aa8a6f66`.**

Round-1 reviewed `2153cbdb`. The intervening commits are `254d98a9` (Round-1 review materialized) and `0b2b9fe9` (corrected candidate materialized). No upstream authority file changed between the two reviewed HEADs, so the Round-1 authority reconstruction remains valid and was re-verified rather than re-derived.

`L-01` from Round 1 is **CLOSED**: the candidate is now repository-reproducible at
`docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-adjudicated-corrected-candidate.md`.

### 0.2 Repository authority reconstructed before judging the delta

Read in the order `AGENTS.md` mandates:

```text
AGENTS.md
docs/engineering/standards/root-cause-global-maximum-method.md  (= METHOD.md 1.0.0)
wiki/references/current-agent-handoff.md
wiki/architecture/r10-technical-architecture.md                 (current R10 router)
docs/superpowers/analysis/…-t8b-…-bootstrap.md
wiki/architecture/launch-v1-ownership-topology.md               (§2, §4 re-read)
wiki/architecture/r10-t3-authorization-audit-enforcement.md     (§2, §9, §12, §13 re-read)
wiki/architecture/r10-t5-durable-async-search-external-effects.md (§1, §2, §6 re-read)
wiki/architecture/r10-t6-canonical-api-frontend-journeys.md      (§2.2, §3, §28, §29 re-read)
wiki/architecture/r10-post-t6-implementation-readiness-program.md (T8-B/C/D/E scope re-read)
wiki/architecture/rebaseline-decision-registry-t8a-amendment.md   (§8 re-read)
```

**Active-authority contradictions found: none.** Router, handoff, bootstrap and T8-A authority agree on stage state (`T8-B ACTIVE`, `T8-C NOT OPEN`, `implementation BLOCKED`) and on the T8-B/T8-C boundary.

### 0.3 Scope of this review

Per §22 of the corrected candidate and the canonical workflow step 7, this is a **delta review only**. The Round-1 conclusion that the one-module owner-first modular-monolith class is the Global Maximum, and that no materially superior alternative topology exists, is **not re-derived**. No already-converged decision is reopened. Nothing below constitutes a new requirement proposal; every finding is a defect against authority the candidate itself already accepts.

---

## 1. Primary verdict

```text
APPROVE CORRECTED T8-B DELTA
WITH MATERIAL FIXES
```

```text
BLOCKER   0
MAJOR     3
LOW       4
```

Stated without hedging:

- the Lead adjudication is **substantively correct**. Every Round-1 BLOCKER and MAJOR is closed at the right abstraction level;
- where the Lead **narrowed or rejected** a Round-1 prescription (B-01 `AuditAppender`, M-08 purge/redact), the Lead is **right and this reviewer's Round-1 prescription was wrong** — those prescriptions were T8-C/T8-D contracts wearing the costume of a T8-B correction. §9 below attacks that reasoning directly rather than assuming Round-1 authority;
- **no corrected statement trespasses into T8-C contracts or T8-D persistence realization.** Two near-misses were tested and both hold;
- the three remaining MAJORs are **omission and internal-inconsistency defects in the corrected text itself**, not defects in any decision. Each is closable by a bounded wording/law correction with **no decision reversal**;
- **another review round is not required.** Final Lead adjudication may proceed directly.

### Sub-verdicts

| Question | Verdict |
|---|---|
| Global Maximum class | **CONFIRMED — unchanged** |
| Corrected candidate still the Global Maximum | **YES** |
| Materially superior alternative found in Round 2 | **NONE** |
| Dependency graph non-contradictory | **YES** (§12/§13 are internally consistent) |
| Dependency graph complete | **NO — two gaps** (`D-M-01`, `D-M-02`) |
| Document internally consistent | **NO — one contradiction** (`D-M-03`) |
| T8-C trespass | **NONE** |
| T8-D trespass | **NONE** |
| New material finding introduced by the correction | **YES — `D-M-01`, `D-M-03`** |
| Upstream reopen required | **NONE** |
| Another broad review | **NO** |
| Another delta round | **NO** |
| Ready for final Lead adjudication | **YES** |
| Ready for operator ratification/promotion | **YES, after the three MAJOR fixes land** |

---

## 2. Delta findings — MAJOR

### D-M-01 — No application leaf exists for the ratified Session/AuthN inbound surface, and `transport → application ONLY` leaves it unroutable

**Classification:** real T8-B defect (omission), introduced by the corrected candidate.

**Evidence.**

T6 ratifies five inbound operations whose semantic owner is Authentication:

```text
T6 §2.2   GET /auth/login                                    (outside the /api/v1 census)
T6 §2.2   GET /auth/callback                                 (outside the /api/v1 census)
T6 §29    GET    /api/v1/session
T6 §29    DELETE /api/v1/session
T6 §29    GET    /api/v1/authentication/provider-subjects?query=…
```

T6 §3 makes `/auth/callback` unambiguously owner work, not transport translation:

```text
verify state/code/provider → resolve issuer + subject → ProviderSubjectBinding
→ require existing ENABLED MetalDocs User → create fresh ApplicationSession
```

The corrected candidate freezes:

```text
§6   transport → semantic owner public surface        FORBIDDEN
§6   transport → application                          the one inbound door
§3   application leaves = library / mywork / documentofficial / documentwork /
                          governancecase / history / audit / admin
```

None of those eight leaves covers session/authentication. Round-1's proposed leaf list contained `session/`; the corrected list correctly **added** `mywork` and `documentofficial` (a genuine improvement in T6 §1/§5 lens fidelity) but **dropped** `session`.

**Known / Inferred / Unknown.** Known: five ratified operations require an inbound path to Authentication. Known: transport may not reach an owner. Known: T6 §28's ratified feature vocabulary **does** include `auth/shell`, so an auth leaf is inside already-ratified lens meaning. Unknown: whether §3's leaf list is exhaustive or illustrative — the candidate never says, and unlike T6 §28 it carries no `e.g.` marker.

**Root cause.** The leaf set was derived from the *product-content* lenses and the authentication/session lens was dropped in transcription. The tree in §3 and the platform-facing summary in §21 are the artifacts an operator ratifies, and neither states that the enumeration is open.

**Target invariant.**

> Every ratified T6 inbound operation has exactly one legal path from transport to its semantic owner through a named application leaf, and no operation requires a transport→owner exception.

**Local vs Global Maximum.** Local fix: allow `transport → authentication` as a narrow exception for the OIDC routes. Rejected — that reintroduces the second inbound door `B-03` exists to close, at precisely the surface where session issuance and provider anti-corruption happen. Global fix: add the leaf.

**Essential vs Accidental.** Adding one leaf is essential completeness, not new structure. It creates no semantic owner: the leaf is stateless choreography over `authentication`, identical in kind to the other eight.

**YAGNI / Future cost.** No speculative capability is added. Cost of omission is concrete: the likeliest Writer resolution is the transport-bypass exception.

**Authority / Boundary.** Entirely inside T8-B (repository/package layout + inbound law). Not T8-C: no contract is named. Not T8-F: T6 §28 already disclaims 1:1 frontend coupling and the candidate repeats that disclaimer.

**Enforcement.** Covered by the existing rule set once the leaf exists — `transport → application` only, `application/<a> → application/<b>` forbidden.

**Proof strategy.** Falsifiable now: enumerate T6 §2.2 + §29 operation families against the leaf set and require a total mapping. A family with no leaf fails.

**Decision.** **MATERIAL FIX.** Add `session/` (or `auth/`) to the §3 leaf set and to §21; state that the leaf set is derived from — and must totally cover — T6's ratified lens/use-case vocabulary including `auth/shell`.

**Why MAJOR and not BLOCKER.** Deliberately not inflated. §5 already states the completion rule ("leaf names may be refined only within already-ratified lens/use-case meaning"), and `auth/shell` is ratified, so a Writer completing the set is applying a stated rule rather than inventing a decision. It is MAJOR rather than LOW because §3/§21 are the ratification artifacts and carry no non-exhaustive marker, and because the omitted surface is exactly where a transport-bypass exception is most tempting.

**Reopen trigger.** If T8-E proves an OIDC browser route needs a transport-level path with no owner interaction at all, revisit whether it needs a leaf — not whether transport may reach owners.

---

### D-M-02 — The allowed-edge matrix reads as exhaustive but the enforcement spec implements deny-listing; at least one harmful edge is permitted by silence

**Classification:** real T8-B defect (enforcement completeness), materialized by the corrected candidate.

**Evidence.**

§12 enumerates allowed edges per class with the precision of a total matrix — e.g. the owner row is exactly three entries:

```text
semantic owner → its own private implementation
semantic owner → its own consumer-declared technical seams
semantic owner → platform/txscope abstraction where transaction participation requires it
```

That row is only correct as a **closed** list: it is what forces `owner → platform/identityprovider`, `owner → platform/managedcontent` and `owner → platform/observability` through consumer-declared ports (dependency inversion) instead of direct import. Read as an *open* list, the entire inversion discipline evaporates.

But §14.3 specifies the analyzer as a rejection list:

```text
Target analyzers must reject at minimum:
  forbidden import edges
  semantic-owner SCCs
  …
```

and §13 is the finite forbidden list. An edge in neither list passes.

The candidate applied closed-world reasoning to *classification* (§14.2: "An unmapped package fails verification. Silence is not a classification.") and open-world reasoning to *edges*. Those are not equivalent, and the repository has already paid for exactly this asymmetry — Round-1 evidence, unchanged at this HEAD:

```text
tools/cilint/internal/analyzers/table_ownership.go:23-30
  "…that direction cannot see an omission: 19 of the 56 live base tables were simply
   absent, and because hgViolation treats an unknown table as a non-violation, every one
   of them was unguarded. An omission is not a classification."

tools/verify/registry.go:41-45
  "…forgetting to opt IN is silent, and silence is how coverage rots."
```

**Concrete hole.** `application leaf → platform/postgres` is in neither list. §13 forbids only `semantic owner → connection pool / driver concrete type`. §9.1 forbids only *transport* from opening a scope. So an application leaf may legally import the pool and acquire a connection outside `platform/txscope` — defeating the single-scope invariant §9 exists to establish, and reintroducing "shared write authority hidden in platform code" (§13) through the one class that is allowed to touch platform directly. Secondary instances: `transport → platform/idempotency`, `transport → platform/observability`, `composition → platform/*` beyond wiring.

**Root cause.** Two laws expressed in two forms whose readings are not equivalent — structurally the same defect class as Round-1 `B-03`, now relocated from the inbound rule to the enforcement rule.

**Target invariant.**

> The allowed dependency graph is total. Any import edge not affirmatively permitted for the importing package's architecture class is forbidden, and the analyzer implements permission, not prohibition.

**Local vs Global Maximum.** Local: add `application → platform/postgres` to §13. Rejected — it closes one instance and leaves the class open, which is the defect. Global: invert the policy to default-deny, consistent with the candidate's own default-DENY authorization posture and closed-world classification.

**Essential vs Accidental.** Removes accidental complexity: with a total matrix, §13 becomes explanatory rather than load-bearing, so the rule set shrinks rather than grows.

**YAGNI.** No new mechanism. The classifier that §14.2 already requires is the same classifier that evaluates a class-pair matrix.

**Authority / Boundary.** T8-B: allowed/forbidden dependency graph and mechanical-enforcement strategy are both explicitly in scope (bootstrap §4, §10.9).

**Enforcement.** Default-deny over the class×class matrix, with negative fixtures per denied pair. Firing is demonstrable.

**Proof strategy.** A fixture adding `application/library → platform/postgres` must fail. Under the corrected candidate as written it passes. That is the falsification.

**Decision.** **MATERIAL FIX.** (a) State that §12 is exhaustive and any edge outside it is forbidden; (b) complete the matrix so the inversion does not forbid legitimate edges — minimally: mechanism access to `platform/observability` and `platform/config` for non-owner classes, and intra-`platform` edges, which §8.2's own realization statement already requires (`platform/postgres` realizes `platform/txscope`); (c) restate §14.3 as allow-list enforcement.

**Reopen trigger.** If the matrix proves to require more than a small, stable set of per-pair exceptions, reconsider the class granularity — not the default-deny direction.

---

### D-M-03 — Owner-private package layout is simultaneously ungated (D04) and assigned to T8-D (§8.2, §18), partially reversing the M-03 correction

**Classification:** real T8-B defect (internal contradiction), introduced by the corrected candidate.

**Evidence.**

§4.1 / `T8B-D04`:

```text
No architecture approval/evidence gate is required for purely owner-private decomposition…
The exact private decomposition is implementation-local and may evolve without reopening T8-B
```

§8.2:

```text
Positive corollary: owner-specific persistence adapters are owner-private
Their exact locations/schema/query design remain T8-D.
```

§18 — T8-D owns:

```text
persistence adapter exact package/file layout inside owner-private code
```

Persistence adapters are the largest single category of owner-private code. If T8-D *owns* their package/file layout, the evidence gate M-03 deleted is not deleted — it is deferred to a later stage for the category where it bites hardest.

**Known / Inferred / Unknown.** Known: the two statements are not reconcilable as written. Known: `T8B-D04` is a decision; §18 is a recorded successor obligation. Unknown: which the Lead intended to bind — and that is the defect.

**Root cause.** "Where adapters live" (layout — T8-B, and freed by D04) was conflated with "what adapters map to" (schema/query/constraint design — T8-D). §8.2 bundles "locations/schema/query design" into one clause and hands the bundle to T8-D.

**Target invariant.**

> Owner-private package layout is unconstrained by any stage. T8-D owns persistence *semantics* — schema, constraints, query and transaction mapping — and may describe adapter placement but may not gate it.

**Local vs Global Maximum.** Local: delete the §18 line. Insufficient — §8.2 carries the same conflation. Global: separate the two concerns explicitly in both places.

**Essential vs Accidental.** Purely accidental complexity: the contradiction adds a stage dependency that protects nothing. Go's nested `internal` already makes owner-private layout invisible to every cross-owner invariant — which is precisely the argument the Lead accepted when adopting M-03.

**YAGNI / Future cost.** Left as written, the first Writer to structure `internal/controlleddocs/internal/…` persistence code must either wait for T8-D or decide against a stated stage boundary. That is the post-T6 program's prohibited failure mode arriving through a documentation seam rather than a design one.

**Authority / Boundary.** Both halves are T8-B's to state: `T8B-D04` is a T8-B decision, and correcting a successor-obligation line that contradicts it is T8-B housekeeping, not T8-D design.

**Enforcement.** None needed — this is a text correction. The mechanical rule (`one public surface per owner`) is unaffected either way.

**Proof strategy.** Coherence review: no two statements in the promoted artifact may assign the same decision to two owners. This is the only surviving instance found.

**Decision.** **MATERIAL FIX.** Reword §8.2 to "their schema/query/constraint design remains T8-D; their package placement inside the owner is owner-private and ungated per D04", and amend §18's T8-D bullet to "persistence adapter *mapping*", dropping "exact package/file layout".

**Reopen trigger.** T8-D producing concrete evidence that a required serialization or constraint realization is only expressible by fixing adapter package boundaries.

---

## 3. Delta findings — LOW

**D-L-01 — §14.5 exception decay says "fail/alert"; the "/alert" branch reintroduces the rot channel M-05 closed.**
Round-1 M-05 required a stale exception to **FAIL**. The corrected text permits either. The repository's own evidence at this HEAD is that alert-grade signals do not get acted on: `tools/cilint/internal/analyzers/platformboundary.go:19` still exempts `internal/platform/objectstore`, which imports zero module packages — a frozen exception that has outlived its violation. Per METHOD, a control counts only when its firing is demonstrable; an alert nobody must act on is not that control. Delete `/alert`. Blast radius is limited by the candidate's own preference for no allowlist, hence LOW.

**D-L-02 — Authorization's fail-closed behaviour on a missing required predicate fact is implied but not stated.**
Under §11.1 the application gathers and routes owner-authored facts. If Authorization treats an absent required predicate as "not required", the T3 §9 permission→predicate mapping migrates to the application by default — a second authorization authority arriving through omission rather than through code. T3 §2's ratified `otherwise default DENY` already closes this, so it is a clarification, not a defect: state once that a required predicate fact that is absent or unverifiable yields DENY. One sentence, no contract, no T8-C content.

**D-L-03 — Classification of in-package `_test.go` files under the closed-world classifier is unstated.**
§14.2 requires every Go package to map to exactly one of nine classes and names `test` as one of them, and §3 adds a `tests/` root. But a `_test.go` file inside `internal/controlleddocs/` is not a separate package for classification purposes, and integration wiring legitimately crosses boundaries. Left unstated, a Writer will either exempt all test files (a hole in a closed-world classifier) or break cross-boundary test wiring. One line closes it: in-package test files inherit the owner class and are bound by the same edge law; cross-boundary assembly belongs in `tests/` or `composition`.

**D-L-04 — Backend attachment point for the interactive editor/viewer is deferred to stages that may not own it.**
§8.6 correctly separates `platform/officialrendition` from interactive editor/viewer realization (L-03 closed) and defers the latter to "later frontend/runtime stages". But T6 records that the selected editor may prove a bounded session/lease is required, and T5 §2 leaves renderer/viewer product selection deliberately unfrozen. A bounded editor session/lease would be a **backend** mechanism, i.e. T8-B layout, not T8-F or T8-G. Attachment is additive under the existing platform law, so nothing is stranded structurally — but stating it prevents a later stage from creating an editor semantic owner. One sentence: if the selected editor requires a backend mechanism, it attaches as an additional `platform/` mechanism under §8.1 and never as a semantic owner.

---

## 4. §22 bounded contract — question-by-question disposition

| # | §22 question | Disposition |
|---|---|---|
| 1 | B-01: same-tx Audit guaranteed without freezing `AuditAppender`? | **YES — correction accepted; Lead's narrowing is superior.** See §9.1 |
| 2 | B-02: `platform/txscope` + application-owned lifecycle correct level, no T8-D theft? | **YES.** §9.2 defers exactly T8-D's ratified set (isolation, lock modes/order, serialization primitive, PostgreSQL mapping). Named home + direction is layout, which is T8-B |
| 3 | B-03: `transport → application ONLY` — Global Maximum or pass-through ceremony? | **GLOBAL MAXIMUM.** Re-tested against the full T6 §29 census: no ratified operation is a pure pass-through. Every read is authorization-composed (T3 §9 predicates) and T6 §26 requires `allowed_actions` to share derivation. The "pass-through" hop is the composition site, not ceremony. Lens-scoped leaves prevent the single-God-interface cost |
| 4 | B-04: fact routing preserves Authorization as sole authority without designing T8-C? | **YES.** Authority partition matches ownership-topology §4 verbatim; §11.2 defers interface, fact vocabulary, actor/scope/result shapes and tx-participation signature. One clarification: `D-L-02` |
| 5 | M-01/M-02: platform sufficiently non-semantic incl. IdP separation? | **YES.** The import ban is strictly stronger than the Round-1 type ban and closes all four proven current leaks (`platform/authn`, `platform/docgenv2`, `platform/bootstrap`), which are import leaks with no SQL |
| 6 | M-02: IdP protocol vs anti-corruption boundary correct? | **YES.** `platform/identityprovider` = discovery/token exchange/JWKS/raw claims; Authentication = provider-subject meaning, binding, session issuance/revocation, assurance. Matches ownership-topology §2 clause-for-clause. The added leak clause (raw claims may not escape the protocol+Authentication boundary) is a genuine strengthening over Round-1 |
| 7 | M-03/M-04: one public surface + free private decomposition + stateless leaves? | **YES — superior to Round-1's framing.** Freezes the essential property (surface) and frees the accidental one (package count). God-surface risk is bounded by the single inbound door; fragmentation risk by §4.3. Verified Go semantics: `internal/<owner>/internal/x` is importable only from the `internal/<owner>/…` subtree. One contradiction: `D-M-03` |
| 8 | M-04: stateless leaves + no app→app constrain God growth without artificial owners? | **YES.** `application/<a> → application/<b>` is trivially machine-checkable and is the control that actually prevents an `application/core`. Escape hatches tested: a shared `application/shared` is caught by the same rule; pushing choreography into `platform` is caught by §8.1; into `composition` by §7. Both bad outlets are already mechanically closed. The residual cost — mechanical boilerplate duplicated across leaves — is the correct trade against a God orchestrator |
| 9 | M-05/M-06: cilint + verify roles, closed-world, bidirectional, fixtures, decay sufficient and minimal? | **NEARLY.** Roles, closed-world classification, bidirectional proof and RED-proven fixtures are correct and minimal. Two gaps: `D-M-02` (edge policy still open-world) and `D-L-01` ("fail/alert") |
| 10 | M-07: `cmd/` + `tools/codegen` + `db/` + `tests/` without deciding T8-G/T10? | **YES — no stealth decision.** `cmd/<runtime-shells>` with names/count/topology explicitly T8-G; `db/` root as class with content T8-D; moves explicitly T10 (§15). The "roots not shown are not implicitly deleted or redesigned" clause closes M-07 more cleanly than either Round-1 option |
| 11 | M-08: is "erasure-safe" the right T8-B invariant? | **YES — Lead's narrowing is superior.** See §9.2 |
| 12 | L-03: `officialrendition` separation correct? | **YES.** Matches T5 §2's ratified `preview/viewer mechanism != OfficialRendition`. One clarification: `D-L-04` |
| 13 | Has any corrected statement crossed into T8-C or T8-D? | **NO.** Two near-misses tested and cleared — see §5 |

---

## 5. Trespass audit — T8-C and T8-D

Registry amendment §8 is the governing permission:

> "It may name required seams/interfaces only to the degree necessary to prove package ownership and dependency direction; detailed communication contracts remain T8-C."

Each of the three seam classes was tested against the post-T6 program's literal T8-C and T8-D scope lists.

| Corrected statement | Trespass? | Basis |
|---|---|---|
| §9.1 `txscope` home = `platform/txscope`, mechanism only | **NO** | "location of shared mechanisms" is T8-B's frozen list |
| §9.1 application owns scope lifecycle; owners participate explicitly | **NO** | Dependency direction, not signature. §9.2 defers the signature |
| §9.1 provider types forbidden on owner public surfaces | **NO** | Public package surface law — T8-B's frozen list |
| §9.2 defers isolation / lock modes / serialization primitive / PG mapping | **CORRECT DEFER** | Exactly T8-D's ratified "transaction and serialization/lock mapping" |
| §10.1 owner authors evidence; application appends in-tx before commit | **NO** | Authority partition + ordering invariant restated from T3 §12, not a contract |
| §10.2 defers handoff interface / port pattern / fact types | **CORRECT DEFER** | T8-C "consumer/producer contract ownership" |
| §11.1 Authorization sole decider; application routes owner-authored facts | **NO** | Ownership-topology §4 restated |
| §11.1 "decision path must be able to participate in the same transaction scope" | **NEAR-MISS — CLEARED** | A capability requirement derived from T3 §11 + T6 §26, not a signature. §11.2 explicitly defers "exact transaction participation signature" |
| §8.2 "owner-specific persistence adapters are owner-private" | **NEAR-MISS — CLEARED as a decision** | Adapter *placement* is layout (T8-B) and the foreclosure of a shared persistence layer is correctly made visible. The accompanying "exact locations … remain T8-D" clause is the `D-M-03` contradiction — a wording defect, not a trespass |
| §8.8 idempotency: opaque, never authorizes, erasure-safe | **NO** | Authority + invariant. Record schema, retention and PII-free-vs-purge realization deferred |
| §14.2 nine-class closed-world classification | **NO** | Bootstrap §10.9 assigns mechanical-enforcement/proof strategy to T8-B |
| §3 `tools/codegen`, `db/`, `tests/` roots | **NO** | Repository layout is T8-B's first frozen item; content deferred to T8-D/T8-E/T9, moves to T10 |

**Result: zero trespass.** The candidate is disciplined at both boundaries — noticeably more so than the Round-1 review it is answering.

---

## 6. Dependency graph — completeness and consistency

**Non-contradictory: YES.** §12 and §13 were checked pairwise. No edge is both allowed and forbidden. The owner row of §12 (private impl + own consumer-declared seams + `txscope`) is consistent with §8.1's platform import ban: owners reach `identityprovider`, `managedcontent`, `river` and `observability` only through consumer-declared ports bound by composition. `application → Audit public surface` is consistent with `owner → audit FORBIDDEN` because application is not an owner.

**Complete: NO — two gaps.**

```text
D-M-01   no inbound leaf for the ratified Session/AuthN operation family
D-M-02   allowed matrix not declared total; §14.3 enforces the deny-list,
         so application → platform/postgres passes
```

**One document-level contradiction outside the graph:**

```text
D-M-03   owner-private layout: ungated (D04) vs T8-D-owned (§8.2, §18)
```

After those three corrections the graph is total, non-contradictory and mechanically enforceable.

---

## 7. Global Maximum — reconfirmation

Not re-derived. Re-tested only for whether the Lead's corrections moved the candidate off the class:

```text
class                    ONE GO MODULE / OWNER-FIRST MODULAR MONOLITH /
                         NON-SEMANTIC APPLICATION / SINGLE INBOUND DOOR /
                         NON-SEMANTIC MECHANISMS / WIRING-ONLY COMPOSITION /
                         COMPLETE MECHANICAL ENFORCEMENT
Round-1 conclusion       class = Global Maximum
Round-2 delta            corrections are realization-level, all inside the class
alternatives A, C, D, E  rejection reasoning unchanged and unweakened
new alternative found    NONE
```

The corrected candidate is **strictly closer** to the Global Maximum than the Round-1 candidate on every axis reviewed, and strictly closer than the Round-1 *review's* own §13 restatement on three axes (T6 lens fidelity via `mywork`/`documentofficial`; T8-C discipline on the audit seam; the "roots not shown" scope clause). No structural element is deletable: the Round-1 subtractive pass produced exactly one deletion (the D04 evidence gate) and the Lead executed it.

---

## 8. New material findings introduced by the correction

```text
D-M-01   YES — leaf set lost the session/auth lens present in the Round-1 restatement
D-M-03   YES — the M-03 fix and the T8-D obligation list disagree
D-M-02   PARTLY — the asymmetry existed in Round 1 but became material only once
         §12 became a precise per-class matrix that reads as total
D-L-01   YES — Round-1 required FAIL; corrected text permits "fail/alert"
D-L-03   YES — a closed-world classifier makes test-file classification load-bearing
D-L-02   NO  — pre-existing implicit property, clarification only
D-L-04   NO  — pre-existing defer, clarification only
```

No new finding requires a decision reversal. All are text-level.

---

## 9. Direct attack on the Lead's rejections of Round-1 prescriptions

The Method requires this reviewer to attack the Lead's technical reasoning rather than assume Round-1 authority. Both rejections were tested to destruction and **both survive**.

### 9.1 B-01 — the Lead rejected the `AuditAppender` consumer-owned port

**Round-1 prescription:** each owner declares `type AuditAppender interface{ Append(ctx, tx, Event) error }`, `internal/audit` implements it, composition injects, the owner appends **at the mutation site**.

**Lead's substitute:** owner authors required evidence facts/results; application coordinates the Audit append in the same tx scope; commit only after the required append succeeds.

**Attack 1 — does the Lead's model lose the same-commit guarantee for owner-internal cascades?**
This was Round-1's core argument: a T2 §9 same-transaction Release, or a T3 §10 multi-event offboarding teardown, could commit with no append because application "cannot see inside the owner". **The argument does not survive.** T3 §12's ratified form is four ordered steps, not three:

```text
BEGIN
business/security semantic mutation
required owning-domain evidence          ← a distinct step
required AuditEvent(s)
COMMIT
```

The Lead's model maps to this **more literally** than the Round-1 prescription did: owner produces owning-domain evidence, application appends the AuditEvent(s), commit. T3 §12 never requires the append to occur at the owner's call site — that was a Round-1 realization detail promoted to an invariant without warrant. T3 §15's offboarding note further confirms the multi-event case is expressible as data ("Multiple semantic events may be appended in the same transaction plus the final User offboarding event"), and §10.1's plural "facts/result" carries it.

**Attack 2 — does the extra hop create a drop channel?**
It creates one, and it is closed by two properties the corrected candidate holds simultaneously: §10.1's "application commit may occur only after required Audit append succeeds" makes discarding owner-produced required evidence a commit-time violation, and §6's single inbound door means **every** mutation is initiated by an application leaf, so there is no path that reaches an owner without passing the coordination site. B-01 and B-03 reinforce each other; Round-1 treated them as independent.

**Attack 3 — is the Lead's model less enforceable?**
Marginally, and the candidate says so honestly rather than concealing it: §14.3 declines to name a fake analyzer for a contract that does not yet exist, records the obligation, and §10.2 hands the proof to T9 ("a required governed/security mutation cannot commit without its required Audit evidence"). Against that, the Lead's model puts append and commit **in the same function** in the same package, which is a strictly easier static-ordering proof than the Round-1 model's cross-package ordering between an owner-side append and an application-side commit.

**Attack 4 — boundary.** `type AuditAppender interface{ Append(ctx, tx, Event) error }` is a method set, a parameter list and a type. That is a T8-C communication contract by the post-T6 program's own definition ("owner capabilities", "consumer/producer contract ownership"). Freezing it in T8-B would have been the exact stealth the bootstrap §5 prohibits.

**Conclusion: the Lead is right and Round-1 was wrong.** The rejection is accepted without reservation. All four properties Round-1 actually needed — no owner→audit import, append inside the mutating transaction, owner owns event meaning, application may not invent it — are preserved in §10.1.

### 9.2 M-08 — the Lead rejected the mandated purge/redact operation

**Round-1 prescription:** "the replay store exposes a purge/redact operation reachable from the UserProfile-erasure use case."

**Lead's substitute:** "replay persistence must be erasure-safe and must not become an unintended PII retention root"; whether that is achieved by PII-free representation or by purge/redaction remains T8-C/T8-D.

**Attack.** Round-1 named an *operation on a mechanism* — that is a contract (T8-C) plus a retention realization (T8-D). It also silently foreclosed the strictly better option: a replay record that stores no PII at all needs no purge path, and the T6 §19 requirement that every replay pass a **live** T3 permission/scope + visibility check before disclosure makes the PII-free design viable. Round-1's prescription would have foreclosed the superior realization while trespassing two stage boundaries at once. The invariant the finding actually established — erasure must not be defeated by a mechanism cache — is fully preserved.

**Conclusion: the Lead is right and Round-1 was over-specified.**

### 9.3 M-04 — the Lead altered the leaf set

Adding `mywork` and `documentofficial` is a genuine improvement in fidelity to T6 §1/§5's ratified lens vocabulary. Dropping `session` is the one regression — `D-M-01`.

---

## 10. Round-1 finding disposition after Lead correction

| ID | Round-1 finding | Post-correction disposition |
|---|---|---|
| **B-01** | same-tx Audit unenforceable under `owner → audit FORBIDDEN` | **CLOSED.** Seam class frozen; exact contract → T8-C. Round-1's prescription correctly rejected (§9.1) |
| **B-02** | transaction-scope carrier undefined | **CLOSED.** `platform/txscope`, application-owned lifecycle, explicit owner participation, provider leakage banned; T8-D defers exact |
| **B-03** | `transport → owner` internally contradictory | **CLOSED.** `transport → application ONLY`; owner public surfaces explicitly closed to transport. Re-tested against the full T6 §29 census: no pass-through ceremony |
| **B-04** | authorization decision input seam unnamed | **CLOSED.** Authorization sole ALLOW/default-DENY; owner-authored facts; application routes only; second evaluator forbidden. Clarification `D-L-02` |
| **M-01** | platform law stated over SQL, not types | **CLOSED — strengthened.** Outright import ban is stronger than the proposed type ban and covers all four proven current leaks |
| **M-02** | `identityprovider` vs Authentication ownership | **CLOSED — strengthened.** Matches ownership-topology §2; adds a raw-claim non-escape clause Round-1 did not require |
| **M-03** | wrong unit — freeze surfaces, not packages | **CLOSED**, except `D-M-03` partially re-imposes the deleted gate through §8.2/§18 |
| **M-04** | `application` God-orchestrator seat | **CLOSED**, except the dropped session leaf (`D-M-01`) |
| **M-05** | rejection-only enforcement, no completeness or decay | **PARTIALLY CLOSED.** Classification is closed-world and bidirectional; **edges are not** (`D-M-02`); decay weakened to "fail/alert" (`D-L-01`) |
| **M-06** | enforcement host misnamed | **CLOSED.** `tools/cilint` = analyzer host, `tools/verify` = registry/profile SSOT, `tools/codegen` added; harness reuse only, policy/fixtures rewritten; five-part gate applied |
| **M-07** | missing repository homes | **CLOSED — better than either Round-1 option.** `db/`, `tests/`, `tools/codegen` homed; "roots not shown are not implicitly deleted or redesigned" closes the exhaustiveness question |
| **M-08** | idempotency replay authority/privacy law | **CLOSED.** Invariant frozen, realization deferred. Round-1 over-specified (§9.2) |
| **L-01** | candidate not in repository | **CLOSED.** Materialized at `0b2b9fe9` |
| **L-02** | "one Go module" scope wording | **CLOSED** |
| **L-03** | rendering conflation | **CLOSED.** `platform/officialrendition`; clarification `D-L-04` |
| **L-04** | managedcontent non-semantic law | **CLOSED.** T4 law restated incl. malware-inspection seam and the no-`owner_type/owner_id` prohibition |
| **L-05** | Library/Search read-model home | **CLOSED.** `application/library`; no Search owner |
| **L-06** | D10's positive corollary | **CLOSED as a decision**; the accompanying clause is `D-M-03` |

```text
CLOSED                    15
CLOSED WITH RESIDUE        2   (M-03 → D-M-03 ; M-04 → D-M-01)
PARTIALLY CLOSED           1   (M-05 → D-M-02, D-L-01)
REOPENED                   0
```

---

## 11. Upstream reopen assessment

```text
Product Contract REV001         NO REOPEN REQUIRED
Whole-Product GCR A1–A10        NO REOPEN REQUIRED
4+1 ownership topology          NO REOPEN REQUIRED
T1 … T7                         NO REOPEN REQUIRED
T8-A + registry amendment       NO REOPEN REQUIRED
Decision Registry               NO REOPEN REQUIRED
```

No concrete contradiction with upstream authority was found. Every Round-2 finding is a defect against authority the corrected candidate already accepts. Per METHOD §3, none is a proposal creating new requirement authority.

---

## 12. Reopen triggers for this delta review

Reopen these conclusions only on material evidence that:

- a ratified T6 operation cannot be served through any application leaf without a transport→owner exception (⇒ reconsider `B-03`, not `D-M-01`);
- the total allowed-edge matrix requires more than a small, stable set of per-pair entries (⇒ reconsider class granularity, not default-deny);
- T8-C proves the owner-authored evidence handoff cannot express a T3 §15 census event without leaking owner internals (⇒ reconsider the B-01 correction, and Round-1's port model returns as a candidate);
- T8-D proves provider-neutral explicit transaction participation is structurally impossible (⇒ reconsider `D25`'s shape, never the atomicity requirement);
- a real external Go consumer, release train or trust boundary appears (⇒ reconsider `D01`).

Preference, ceremony aversion, sunk cost and hypothetical futures are not triggers.

---

## 13. Adjudication guidance

```text
PRIMARY VERDICT     APPROVE CORRECTED T8-B DELTA
                    WITH MATERIAL FIXES

BLOCKER   0
MAJOR     3         D-M-01  add the session/auth application leaf; state the leaf
                            set must totally cover T6 ratified lens vocabulary
                    D-M-02  declare the §12 allowed matrix total, complete it, and
                            restate §14.3 as allow-list enforcement
                    D-M-03  separate owner-private layout (ungated, D04) from
                            persistence mapping (T8-D) in §8.2 and §18
LOW       4         D-L-01  delete "/alert" from §14.5 exception decay
                    D-L-02  state Authorization DENYs on absent required facts
                    D-L-03  classify in-package _test.go under the closed-world rule
                    D-L-04  state that a required backend editor mechanism attaches
                            as a platform mechanism, never a semantic owner

T8-C TRESPASS                NONE
T8-D TRESPASS                NONE
GRAPH NON-CONTRADICTORY      YES
GRAPH COMPLETE               NO — closed by D-M-01 + D-M-02
GLOBAL MAXIMUM               CONFIRMED — corrected candidate remains it
UPSTREAM REOPEN              NONE
ANOTHER BROAD REVIEW         NO
ANOTHER DELTA ROUND          NO
FINAL LEAD ADJUDICATION      MAY PROCEED DIRECTLY
READY FOR RATIFICATION       YES, after the three MAJOR fixes land
```

None of the three MAJOR fixes reverses a decision, changes the topology class, opens T8-C, or touches T8-D. They are completeness and consistency corrections to the text of a candidate whose architecture is sound.

T8-B remains **ACTIVE**. T8-C remains **NOT OPEN**. Implementation remains **BLOCKED**.

---

**End of independent delta review. Reviewer findings are evidence, not authority.**
