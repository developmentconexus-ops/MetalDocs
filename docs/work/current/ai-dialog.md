# AI Dialog — MetalDocs working-method review

candidate_branch: `chore/fable-git-review-protocol`
candidate_head: `20d53b9dec5beffbc047fa1a1bcfcbb171811236`
review_branch: `review/repository-method-fable`
status: `OPEN`

> This file is temporary review transport only. The review branch must differ from the candidate only by this file. Never merge this branch or this file.

## REVIEW_REQUEST R1 — LEAD

### Review objective

READ-ONLY review of the **MetalDocs working method as an operating system for finishing the Product**, not a generic opinion on whether the methodology “makes sense”.

The question to falsify is:

> Given the actual stage of MetalDocs, does the current method reliably and efficiently take an already deeply planned Product into concrete user-operable frontend validation, allow new frontend Evidence to falsify/refine prior Product/API/architecture decisions when genuinely necessary, find the Global Maximum instead of patching locally, and then return to Product work without review/governance loops becoming the work itself?

Do not edit any candidate file. Append your answer only under `## FABLE_RESPONSE R1 — FABLE`, commit this file, and push this review branch.

### Where MetalDocs actually is

MetalDocs is **not** at initial discovery.

Already planned/ratified in substantial depth:

- Product scope/concepts and whole-product alignment;
- domain state/invariants;
- lifecycle/governance;
- authorization/audit;
- content integrity;
- async/search/effects;
- canonical journeys;
- semantic ownership;
- backend topology/interfaces/persistence;
- executable API/wire contract;
- frontend realization architecture;
- runtime/process/deployment;
- validation baseline;
- transition/cutover.

Current application census = **89 operations**.

Product/runtime implementation remains **BLOCKED** until the Product readiness gates close.

Current phase = **T11 / FP1 — block-by-block Product Experience concretization**.

B01–B10 are accepted/LOCKED/integrated.

B11 — Access Administration — is the next clean Product Experience increment.

Integrated upstream B11 authority:

- `docs/decisions/access-assignment-read.md` — op31 read precision.

### What the frontend phase is for

The functional low-fidelity frontend is not decoration and is not production implementation.

It is a **Product simulator / falsification surface**:

```text
accepted Product + architecture
→ human need
→ IA / flow / interaction hypothesis
→ functional low-fi P8
→ operator actually uses it as the user
→ concrete Evidence appears
→ validate or falsify prior assumptions
```

Expected outcomes when a gap appears:

```text
frontend representation defect
→ local frontend correction

current authority already sufficient but frontend used it incorrectly
→ bounded correction

real user need cannot be satisfied by current accepted authority
→ stop only affected scope
→ Engineering Method
→ root cause / invariant / alternatives
→ Local vs Global Maximum
→ essential vs accidental complexity / YAGNI
→ correct owning authority
→ proof / decision
→ reopen smallest upstream owner if justified
→ bounded frontend rebaseline
→ continue Product concretization
```

Both extremes are failures:

1. every frontend issue reopens Product/architecture;
2. accepted planning is never allowed to be challenged by concrete user Evidence.

### Current operating route

Read first:

- `AGENTS.md`
- `docs/roadmap.md`
- `docs/development/engineering-method.md`
- `docs/development/repository-method.md`
- `docs/development/frontend-product-experience-planning-method.md`
- `docs/development/engineering-rules.md`
- `docs/decisions/repository-readiness.md`
- `docs/index.md`

Then expand only when it can materially falsify a claim.

For B11, likely relevant:

- `docs/decisions/access-assignment-read.md`
- `docs/architecture/authorization-and-audit.md`
- `docs/decisions/api-operation-census.md`
- exact Access sections/operations in `docs/product/journeys.md`
- exact operations/global laws in `docs/architecture/wire-contract.md`
- relevant realization law in `docs/architecture/frontend.md`

Do not recursively read the repository merely for ceremony.

### B11 as a real falsification case

The superseded old B11 workspace exposed four known failure classes now carried by current authority:

1. **Pagination** — paginated member/selector reads must expose real server-page traversal; no hidden all-page crawl.
2. **op6 `listUsers`** — preserve raw server page boundaries; DISABLED Users stay visible but unavailable; no enabled-only pre-filter before pagination.
3. **Group membership** — add-member UX cannot assume complete GroupMembership knowledge from paginated op27; idempotent op28 PUT reconciles 201 newly-added vs 204 already-existed.
4. **Idempotency** — repeated grant confirmation with the same logical Idempotency-Key must produce zero second semantic mutation.

Use these to test whether the method naturally reaches the right owner/outcome without patch-on-patch or unnecessary upstream redesign.

### Properties to test

For each, return `PASS | PARTIAL | FAIL` with repository evidence, strongest attempted counterexample, and correction if needed.

**A — Authority recovery**
Can a fresh actor recover current stage, gates, methods and semantic owners without chat/old-PR archaeology?

**B — Context efficiency without blindness**
Can normal work start small/targeted while still expanding whenever another source can materially falsify the conclusion?

**C — Frontend as Product validation**
Does P8 actually test whether a human can perform the real Product job and expose missing information/actions/API/Product assumptions?

**D — Gap triage / ownership**
Can the method distinguish UX defect, frontend realization defect, existing-authority misuse, API/read insufficiency, Product gap and architecture gap, reopening only the smallest correct owner?

**E — Global Maximum behavior**
Does a material gap go through root cause/invariant/alternatives/Global Maximum/YAGNI/authority/proof instead of simply closing the nearest finding?

**F — Proportionality**
Do cheap frontend iterations stay cheap while Product/AuthZ/API/persistence/trust-boundary changes receive proportional rigor?

**G — Review convergence**
Does the ClaudeCode/FABLE posture prevent `review → finding → patch → review` forever? Attack the stop conditions and the new Git-mediated `ai-dialog.md` protocol.

**H — Product flow**
After resolving a finding, can the team return quickly to Product concretization instead of repository administration becoming the work?

**I — End-state viability**
Can this method plausibly take MetalDocs from current deep planning + B01–B10 through remaining FP1, integrated low-fi P11, whole-product challenge P12, visual/readiness stages, T11/T12 closure and eventual implementation authorization without either implementing an unvalidated UX or planning forever?

### Required counterexamples

Invent at least **two realistic future frontend discoveries**:

- one where the correct outcome **must be an upstream Product/API/architecture reopen**;
- one where the correct outcome **must remain a local frontend correction**.

Walk both through the current method and try to make it misclassify them.

### Review doctrine to challenge

The useful historical ClaudeCode/FABLE doctrine is intended to be:

```text
verify anchors/premises first
→ root cause before patch
→ Local Maximum vs Global Maximum
→ /simplify + YAGNI
→ reviewer finding = Evidence, not authority
→ dispose prior findings explicitly
→ attack only new material uncertainty
→ repeated same-altitude findings = structural signal
→ converge and STOP
```

The old `.claude/skills` framework is intentionally not restored as another methodology.

Current Git transport candidate rule is:

```text
candidate exact HEAD
→ review branch from exact HEAD
→ only delta = docs/work/current/ai-dialog.md
→ Lead pushes request
→ FABLE appends response + pushes
→ Lead appends adjudication + pushes
→ corrections occur on candidate, never review branch
→ changed candidate HEAD => new review branch
→ review branch never merges
```

Try to falsify whether this is sufficient and minimal.

### Structural findings format

Number findings most severe first:

```text
SEVERITY: BLOCKER | MAJOR | MINOR
CLAIM:
EVIDENCE:
ROOT CAUSE:
WHAT MAKES THIS DEFECT CLASS IMPOSSIBLE:
LOCAL VS GLOBAL MAXIMUM:
SMALLEST SUSTAINABLE CORRECTION:
BLOCKS B11: YES | NO
```

Apply `/simplify` aggressively. Do not solve context overflow or review quality by inventing more context machinery/frameworks unless a concrete defect requires it.

### B11 classification output

For each of the four known B11 failure classes state:

- gap classification;
- correct owner;
- expected method path;
- whether current method reaches the correct outcome;
- remaining patch-on-patch/replanning risk.

### Final verdict

End with exactly one:

```text
VERDICT: PROCEED
VERDICT: PROCEED WITH METHOD CORRECTIONS
VERDICT: DO NOT PROCEED
```

Then:

```text
BIGGEST REMAINING RISK: <one sentence>
```

If corrections exist, explicitly separate:

```text
BLOCKS B11
vs
CAN BE CORRECTED WITHOUT STOPPING B11
```

---

## FABLE_RESPONSE R1 — FABLE

<!-- FABLE: append your response below this line, then commit and push ONLY this file on review/repository-method-fable. -->
