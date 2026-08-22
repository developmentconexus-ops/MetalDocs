# Fable Review — Frontend Product Experience Planning Method v2

## Review target

Candidate repository: `developmentconexus-ops/MetalDocs`  
Candidate branch: `arch/t11-implementation-program`  
Exact candidate HEAD: `a9e6f3b3ae2b8e56c65d8114e1551e40ec1d7161`  
Required CI on candidate: `#1218 SUCCESS`  
Review branch: `review/t11-frontend-method-fable`

Primary review subject:

```text
docs/development/functional-html-wireframe-method.md
```

Read repository governance/routing only as needed:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ methodology
```

Do **not** review or redesign MetalDocs product screens yet. The prior whole-product HTML prototype was explicitly rejected/superseded because it skipped deliberate IA/layout exploration. This review is solely an adversarial review of the reusable methodology that must prevent that failure across products/repositories.

## Required reviewer stance

Act simultaneously as:

```text
Principal Product Designer
Information Architect
Senior UX Research / Service Design practitioner
Senior Frontend Architect
Design-system / component-architecture specialist
Accessibility-conscious interaction designer
Adversarial software/product architecture reviewer
```

Be rigorous, skeptical and evidence-oriented. Do not approve because the document is detailed. Attack whether the process would actually produce a coherent, usable, implementation-ready frontend while preserving accepted backend/product authority.

## Core question

> Does this methodology leave any material product, UX, information-architecture, layout, interaction, component-pattern, responsive/accessibility or backend↔frontend decision to be improvised during production implementation?

A second required question:

> Is the method itself proportionate and YAGNI, or has it introduced ceremony/analysis that would slow ordinary blocks without protecting a real failure mode?

## Context / failure that triggered v2

The previous method proved Product/API coverage and then moved too quickly from Screen Contracts to a complete HTML prototype. That produced a structurally poor result because it skipped deliberate discussion of:

```text
user goals / jobs
information architecture
navigation mental model
reference study
layout alternatives
screen hierarchy
position / relative size / density
cards vs tables vs lists vs master-detail
progressive disclosure
screen-by-screen operator walkthrough
pattern derivation after evidence of repetition
```

The operator requires a process in which assistant + human discuss and visually evaluate each important screen/block before it is locked. The method must remain reusable and not contain MetalDocs-specific ontology.

## External-practice principles already considered

The candidate deliberately tries to incorporate these established ideas; validate whether it uses them correctly rather than merely naming them:

```text
user need/problem before solution
whole end-to-end user flow
prototype/test before committing to build
multiple design approaches when genuine ambiguity exists
information architecture / mental-model validation when material
patterns chosen by user task rather than decoration
data tables for dense structured comparison/find-specific-record tasks
alternate views only when distinct tasks justify them
progressive disclosure without hiding decision-critical information
accessibility and responsive behavior as structural concerns
prototype code as evidence, not production code
```

## Mandatory attack dimensions

### A. User-centered discovery

Check whether the method adequately distinguishes:

```text
accepted Product capability
user need / job
persona/access role
operator assumption
actual user evidence
```

Find any place where screen design could still begin from backend nouns rather than user goals.

### B. Information Architecture

Attack:

```text
global vs contextual navigation
browse/search/filter strategy
object/task grouping
mental models
findability
growth/future seams without speculative features
sequential journeys vs unordered repeated-task navigation
```

Does P4 occur at the right time? Is it falsifiable enough? Is there a risk of doing IA theater without evidence?

### C. Screen/block decomposition

Attack whether:

```text
blocks are too large
blocks are too small
screen inventory happens too early/late
global IA dependencies can invalidate later locked blocks
one-block-at-a-time can create local optimum inconsistent with whole product
```

If a whole-product coherence checkpoint is required between blocks, say exactly where and why.

### D. Reference study

Attack whether reference research can degrade into:

```text
copying popular SaaS patterns
confirmation bias
research theater
unbounded browsing
wrong task analogies
```

Evaluate whether 3–6 references is appropriately optional/bounded and whether the evidence matrix is sufficient.

### E. Competing layout hypotheses

Attack whether forcing alternatives can create ceremony, and whether the escape clause for conventional patterns is strong enough.

Review criteria for:

```text
table vs cards/grid
list
master-detail
search-first
category-first browse
alternate list/grid views
```

Identify missing decision variables if any.

### F. Human operator collaboration

This is critical.

Determine whether the method truly guarantees:

```text
assistant presents a bounded screen/block
human can visualize it
human and assistant discuss layout/position/size/hierarchy
findings are recorded
block becomes LOCKED only after explicit adjudication
next material block is not silently generated
```

Attack ambiguity around what "sufficiently coherent" or "LOCKED" means.

### G. Backend/frontend coherence

Attack:

```text
screen-shaped API risk
missing identity sources
read-model sufficiency
frontend permission/authority duplication
concurrency/idempotency/recovery behavior
projection vs source-truth confusion
```

Evaluate the order `structural wireframe → Screen Contract/vertical trace`. Is that correct, or should a lightweight backend-sufficiency gate happen before layout with the full contract after? This is an especially important point to adjudicate.

### H. Component / interaction pattern derivation

Attack both duplication and premature abstraction.

Does P10 happen at the right time? Should pattern vocabulary evolve incrementally after each locked block rather than wait until all screens? How should local patterns graduate into shared patterns without building a speculative design system?

### I. HTML prototype role

Confirm the method prevents HTML from becoming the design-thinking phase.

Attack whether P11:

```text
is too late to discover important interaction problems
is early enough for realistic flow testing
preserves reviewed layout
supports local negative-state simulation
remains throwaway evidence
```

### J. Visual design handoff

Attack the boundary between structural UX and visual design. Identify any area where “visual styling” could materially alter hierarchy/task behavior without triggering a reopen.

### K. Accessibility + responsive behavior

Check whether these are integrated early enough and whether requirements are concrete enough to prevent an inaccessible structural layout from being locked.

### L. YAGNI / overengineering

Assume this method will be reused on both complex enterprise products and smaller products.

Identify ceremony that should be conditional rather than mandatory. Prefer clear trigger conditions over blanket steps where possible.

### M. Implementation by LLM

This is a primary consumer.

Attack whether a capable coding LLM, given the completed artifacts, could still legitimately ask or invent:

```text
what screen to build
what information comes first
which components/patterns to reuse
where actions belong
how navigation works
which backend call owns an action
what state/error behavior to show
what changes responsively
```

Any materially unanswered question is a readiness defect.

## Review severity

Use exactly:

```text
MATERIAL
  method can plausibly produce materially wrong UX/product architecture or still force implementation-time invention

IMPORTANT
  meaningful weakness to correct before method convergence, but does not invalidate the entire approach

OPTIONAL
  useful improvement; method remains correct without it

UNSUPPORTED PREFERENCE
  taste/style/ceremony preference without a protected user/system property
```

## Output contract

Append your response below `## Fable response` in this same file.

Required structure:

```text
VERDICT: CONVERGED | NOT CONVERGED
MATERIAL: <count>
IMPORTANT: <count>
OPTIONAL: <count>
UNSUPPORTED_PREFERENCE: <count>

1. Executive assessment
2. Material findings
3. Important findings
4. Optional findings
5. Unsupported preferences
6. Phase-order critique
7. YAGNI / proportionality critique
8. LLM implementation-readiness critique
9. Exact recommended corrections, smallest-first
10. Re-review scope
```

For every MATERIAL/IMPORTANT finding provide:

```text
finding id
method section(s)
concrete failure scenario
protected property / user harm / implementation risk
why current text is insufficient
smallest recommended correction
whether correction is generic methodology or product-specific (product-specific findings are out of scope here)
```

Do not modify any file other than this dialogue. Do not create Product code. Do not begin T12. Do not treat the review branch as a merge candidate.

---

## Fable response

<!-- Fable: append the adversarial review below this line. -->

```text
Reviewer                  Fable (independent, isolated)
Basis                     candidate a9e6f3b3ae2b8e56c65d8114e1551e40ec1d7161 (CI #1218 SUCCESS, verified live)
Subject                   docs/development/functional-html-wireframe-method.md (v2.0, read whole)
Routing followed          AGENTS.md → docs/index.md → docs/roadmap.md → methodology
Review delta              docs/work/current/ai-dialog.md only
```

```text
VERDICT: NOT CONVERGED
MATERIAL: 2
IMPORTANT: 10
OPTIONAL: 6
UNSUPPORTED_PREFERENCE: 0
```

### 1. Executive assessment

The v2 method is a substantial and mostly correct correction of the failure that triggered it. Its laws (§4) attack the right enemies: backend-noun navigation, all-at-once generation, screen-shaped APIs, parallel frontend authority, premature abstraction, silent visual redesign. The phase order is fundamentally sound: coverage before layout, IA before screen inventory, references and competing hypotheses before commitment, structure before contract, contract before prototype, prototype before walkthrough. The evidence posture (§5) and the three-layer research discipline (§23) are genuinely better than most industrial practice, and the finding-classification ladder (§29) correctly makes upstream reopen the last resort rather than the first.

But the method's core promise — *the operator sees and adjudicates every important block before it becomes baseline* — is not yet mechanically guaranteed by the text. Two defects re-open the exact failure mode v2 exists to prevent: the lock/progression semantics are stated three different, mutually inconsistent ways, none of which actually requires operator adjudication before progression; and the artifact the operator is supposed to "visually evaluate" at P8 is never defined as a viewable thing at all. A capable, well-intentioned LLM following this document literally can still race ahead of the human on self-judged coherence, presenting prose descriptions as "wireframes" — which is precisely the v1 failure with better paperwork. Both defects have small textual corrections. The remaining findings are meaningful but bounded weaknesses in an approach that is directionally right and should converge in one bounded correction round.

### 2. Material findings

**FM-1 — Lock authority and block progression are underspecified; silent progression remains lawful.**

- Method sections: §4.6, §6, §5.4, §15.1, §22.
- Concrete failure scenario: An LLM assistant completes block B02's candidate wireframe, judges it "sufficiently coherent to serve as context" (§15.1 — its own judgment, no definition), notes that its open findings do not "materially affect global IA/patterns" (§22 — its own judgment again), and proceeds to generate B03 and B04 before the operator has adjudicated B02. Nothing in the text forbids this. Later the operator reworks B02's structure, and B03/B04 — built on B02 as "context" — inherit the discarded structure. This is the v1 failure (assistant outruns human review) reconstructed inside the block protocol.
- Protected property / risk: the operator-adjudication guarantee that is the method's reason to exist (attack dimension F; trigger context: "discuss and visually evaluate each important screen/block **before it is locked**").
- Why current text is insufficient: three progression rules coexist and none requires a lock: §4.6 says "record decision → only then progress" (a recorded decision can be CANDIDATE or FINDING); §15.1 says do not start the next high-impact block until the current one is "sufficiently coherent" (undefined, assistant-judged, weaker than LOCKED); §22 blocks progression only when unresolved findings "materially affect global IA/patterns" (assistant-judged materiality). Separately, §6 defines LOCKED as "accepted for the current planning baseline" without ever naming *who* may set it; §5.4 gives the operator decision ownership only "when the planning process requires a choice," and the text never states that locking is such a choice. The weakest reading is always available, and an LLM under progress pressure will find it.
- Smallest recommended correction: (a) in §6, define LOCKED as an operator-only status — no assistant, reviewer, or tool may set it; (b) collapse the three progression rules into one: *the next material block may not be presented or generated as baseline until the current material block is LOCKED, or the operator explicitly authorizes parallel progression with the current block held CANDIDATE*; (c) delete or subordinate "sufficiently coherent" to that rule.
- Classification: generic methodology.

**FM-2 — The P8 structural wireframe has no defined viewable medium; "operator visual evaluation" is not mechanically realizable.**

- Method sections: §15, §15.1, §18 (boundary), trigger context in this brief.
- Concrete failure scenario: An LLM produces, for each screen, a markdown table of regions plus a bullet list of hierarchy decisions and labels it "CANDIDATE STRUCTURAL WIREFRAME." The operator — required by the method to discuss "layout, hierarchy, position, size" — cannot actually see relative size, position, reading order, or density from prose, approves what they imagine rather than what will be built, and the structure is LOCKED. The mismatch surfaces at P11, when the assembled HTML finally makes the structure visible — exactly the "structurally poor result discovered too late" that v2 was written to prevent. The dual failure is also lawful: to make the wireframe visible, the LLM jumps to styled HTML at P8, and HTML silently becomes the design-exploration phase again, because the P8/P11 boundary is drawn by phase name, not by artifact definition.
- Protected property / risk: the human's ability to *visually* adjudicate structure before lock (attack dimension F: "human can visualize it"); the P11 premise that HTML realizes already-reviewed structure (§18: "HTML is not where layout exploration starts").
- Why current text is insufficient: §15 defines what a structural wireframe *decides* but never what it *is*. No sentence in the document names an acceptable representation (schematic drawing, annotated box layout, image, unstyled HTML skeleton), and no sentence requires the operator-reviewed artifact to be rendered/viewable rather than described. The §18 boundary ("avoid the production framework"; "HTML is not where exploration starts") is about P11 and, read defensively, discourages the one medium an LLM-plus-operator pair can most easily share at P8.
- Smallest recommended correction: add to §15 a wireframe-medium rule: a candidate structural wireframe MUST be a rendered visual artifact the operator can view (schematic image, SVG/box diagram, or a deliberately unstyled grayscale HTML/CSS skeleton with placeholder content); prose or tables alone do not satisfy P8 review. State explicitly that a per-screen structural HTML *skeleton* is a permitted P8 wireframe medium, and that the P8/P11 boundary is scope and fidelity (single-screen structural schematic, no brand styling, no cross-flow navigation) — not the presence or absence of HTML technology.
- Classification: generic methodology.

### 3. Important findings

**FI-1 — Global IA decision status at P4 exit is ambiguous; the highest-impact structure can lock on the least evidence.**

- Sections: §11 (P4 exit), §13 (P6 per-block), §22 (B01).
- Failure scenario: P4 exits with "a reviewed model" of navigation/grouping before any reference study exists (P6 is per-block and comes after P5). If that model is treated as LOCKED, the product's single highest-leverage structure — global navigation — was decided with zero reference evidence and no competing hypotheses, while every lesser screen gets both. §22's B01 ("App shell + global IA") then re-covers the same ground with no stated relationship to the P4 output.
- Protected property: IA quality; "IA theater without evidence" (attack dimension B).
- Insufficiency: P4 exit does not assign a §6 status to its output, and the P4↔B01 relationship is unstated.
- Smallest correction: state that P4 exits with the IA as **CANDIDATE**; it becomes LOCKED only through the first block cycle (B01 or equivalent: reference study + hypotheses + operator adjudication of the global shell/navigation). This also answers dimension C's checkpoint question: the first mandatory whole-product coherence point is the B01 lock, because every later block inherits its navigation frame; the terminal one is P12.
- Generic methodology.

**FI-2 — Phase-to-loop scope is never stated; §32 reads as product-wide linear.**

- Sections: §32 vs §13/§15/§22.
- Failure scenario: an LLM follows the §32 diagram literally and runs each phase product-wide: all reference studies, then all hypotheses, then all wireframes for all blocks, then all contracts. §4.7 prohibits one-pass *generation-then-review*, but a phase-global execution with per-phase reviews still defeats block-by-block adjudication and §22 ordering — while claiming compliance with §32.
- Protected property: block-by-block operator loop; LLM executability (dimension M).
- Insufficiency: §13 says P6 is per-block, §15 embeds P6–P8 in the block cycle, §22 puts P9/P11 per block — but this scoping is scattered and §32, the artifact an implementer will actually follow, contradicts it visually.
- Smallest correction: add a small scope table to §32: P0–P5 global (P4 exiting CANDIDATE per FI-1); P6–P9 per block inside the block cycle; P10 incremental per block + terminal reconciliation; P11 per block after lock, assembled; P12–P14 global.
- Generic methodology.

**FI-3 — P7 "backend truth fit" is a table cell with no check behind it.**

- Sections: §14.1, §16, adjudication requested by dimension G.
- Failure scenario: the operator locks a master-detail structure at P8; at P9 the contract reveals the list read model lacks the summary fields the detail pane presumes, or the card grid presumes preview/thumbnail truth no owner supplies. The block reopens after lock — lawful under §32's loop, but the rework was cheaply preventable.
- Protected property: locked structure durability; avoidance of late screen-shaped-API pressure (an implementer squeezed at P9 is exactly who invents a convenience endpoint).
- Insufficiency: "backend truth fit" appears as one criterion among sixteen with no procedure; nothing requires a hypothesis to name its data dependencies before lock.
- Smallest correction: require each *leading* hypothesis to carry a data-feasibility line before P8 lock: the material read/write truths it presumes (fields, scale, sort/filter, preview truth), each marked present-in-authority or named as a FINDING. Explicit adjudication of dimension G's order question: the current order (P3 coarse sufficiency → P8 structure → P9 exact contract) is **correct** — a full contract before layout would be waterfall waste, since contract content depends on chosen structure; the gap is only this missing lightweight feasibility line at P7, not a reordering.
- Generic methodology.

**FI-4 — Pattern-derivation cadence (P10) is unspecified relative to the block loop.**

- Sections: §17, §22, §32.
- Failure scenario: read one way ("after enough screens"), P10 runs once after all blocks — B05–B08 are wireframed with no shared vocabulary and produce divergent drawer/table/form semantics that a late P10 must reconcile by reopening locked blocks. Read the other way, vocabulary consolidates continuously with no stated rule, and early two-screen coincidences graduate into premature abstractions.
- Protected property: pattern consistency without premature abstraction (dimension H).
- Insufficiency: §17's "after enough screens have been reviewed" is the only timing statement and it is unfalsifiable.
- Smallest correction: run a bounded pattern-consolidation pass after **each** block lock (candidate patterns from ≥2 locked screens; §17's semantic-sameness rule unchanged), with P10 proper as the terminal reconciliation. Local patterns graduate to shared only at a consolidation pass, never mid-block.
- Generic methodology.

**FI-5 — Assumption lifecycle is untracked; labeled assumptions can silently become locked structure.**

- Sections: §5.2, §8, §19.
- Failure scenario: P1 needs are written from operator/domain assumption (lawful under §5.2: "labeled... and tested during prototype review"). Frequency and comparison-vs-recognition assumptions drive P7 choices and P8 locks. P12 arrives; its attack questions never mention the assumption register, so nothing forces the walkthrough to retest the specific labeled assumptions. The product ships table-first for users who actually needed recognition-first — with every review checkbox green.
- Protected property: user-evidence grounding (dimension A); §5.2's own promise.
- Insufficiency: assumptions are labeled at creation but no later phase is obligated to revisit them; the promised "tested during prototype review" has no hook in §19.
- Smallest correction: make the assumption register a required artifact (assumption → phase(s) it influenced → how P12 will probe it), and add one P12 requirement: every material assumption is explicitly probed or carried as an open FINDING into P14, where unresolved material assumptions block closure.
- Generic methodology.

**FI-6 — Accessibility has a principle chapter but no P8 hook; an inaccessible structure can be LOCKED.**

- Sections: §24 vs §15/§15.2.
- Failure scenario: §15's list of what wireframes decide includes responsive rules but not one accessibility item; §15.2's operator walkthrough questions contain zero accessibility questions. A hover-revealed action model or a drag-only reorder interaction is locked at P8; the accessibility reviewer first appears at P12, after structure is frozen across blocks, and the fix now reopens locked blocks — the exact "accessibility deferred as polish" failure §30 tells reviewers to attack.
- Protected property: §24's own claim that accessibility is structural.
- Insufficiency: principle without a mechanism in the phase where structure is decided.
- Smallest correction: add accessibility to §15's decided-by-wireframe list (keyboard/focus order plausibility, non-drag alternative, heading/label structure, disclosure reachability) and 2–3 accessibility questions to §15.2. Responsive already has this treatment; mirror it.
- Generic methodology.

**FI-7 — Orphan-operation zero-count pressure invites manufactured screens.**

- Sections: §10 (P3 exit), §21 ("orphan backend human operations 0").
- Failure scenario: a backend operation has human trust class but no discovered user need. P14 requires orphans = 0 to close. The lawful dispositions are unstated, so the path of least resistance is to invent a user need and an admin junk-drawer screen to house the operation — screen design beginning from a backend noun, laundered through the coverage matrix. This is the inversion of dimension A's attack: the method firmly blocks *need→missing-API* invention (§4.8) but not *API→manufactured-need* invention.
- Protected property: user-goal-first design; honesty of the closure counts.
- Insufficiency: no finding class or disposition exists for excess/unneeded backend capability; §19's UPSTREAM covers only *missing* truth.
- Smallest correction: state that an orphan human-facing operation resolves by exactly one of: (a) a real evidenced user need and home; (b) an operator-adjudicated NOT-HUMAN-FACING/DEFERRED disposition recorded in the ledger; (c) an UPSTREAM finding of excess capability. The P14 count then reconciles as orphans-without-disposition = 0.
- Generic methodology.

**FI-8 — UX language (terminology, labels, message intent) has no owner; the LLM invents all user-facing words at implementation time.**

- Sections: §11.1 (partial: "name in user language"), §16, §21, dimension M.
- Failure scenario: artifacts complete, method closed. The implementing LLM must still write every button label, empty-state sentence, permission-denied explanation, and concurrency-conflict message. Error language for material failures ("stale write", "ambiguous outcome") *is* decision-critical UX — a wrong conflict message causes wrong user action — yet contracts specify the state, never the communicative intent. Terminology drifts per screen ("archive"/"retire"/"deactivate") because 11.1's user-language names are never consolidated into a durable artifact.
- Protected property: dimension M's "what state/error behavior to show" — behavior includes what the user is told; cross-screen coherence.
- Insufficiency: no phase, artifact, or checklist item owns copy; the a11y chapter's "labels/instructions" is the only mention.
- Smallest correction: (a) make 11.1's object/action names a durable glossary artifact carried through P8–P11; (b) add one line to the Screen Contract's MATERIAL STATES: the *message intent* of each material failure (what the user must understand and be able to do next), not final copy. Full content design remains out of scope; unnamed intent does not.
- Generic methodology.

**FI-9 — No structural-conformance check after visual design; the handoff contract is enforced by self-report only.**

- Sections: §20, §21, dimension J.
- Failure scenario: visual design legitimately exercises its allowed list — spacing scale, typography, "visual hierarchy emphasis" (§4.11 allows it) — and the accumulation quietly demotes the primary region, changes effective density (§26 calls density a task decision), and buries the primary action, without any single change feeling structural. The must-not list is violated in aggregate; nobody is required to look. §20's only safeguard is that the designer "must raise a finding" — self-report by the party incentivized not to.
- Protected property: locked hierarchy/task behavior surviving to implementation.
- Insufficiency: P14's closure list verifies traces and findings but never verifies that visually designed screens still match locked wireframes.
- Smallest correction: add one P14 (or post-P13) requirement: a structural-conformance review of visual-design output against locked wireframes — reading order, region priority, action placement, density class per screen — with mismatches classified via §19. Where visual design happens during implementation (LLM-built products), state that this check transfers to implementation review as an explicit obligation.
- Generic methodology.

**FI-10 — No proportionality/tailoring rule; the method's own weight becomes an adoption risk on small scopes.**

- Sections: whole document; dimension L; the brief's second required question.
- Failure scenario: a team must add one screen to a converged product. Read literally, the method demands P0–P14, evidence matrices, and a block ledger for one screen. Teams then either perform ceremony (analysis theater) or — worse and more likely — skip the method entirely because it is "for big things", losing its laws exactly where a quick screen-shaped API is most tempting. Conditionality currently exists but is scattered in adjectives ("non-trivial", "consequential", "when material", "when available") that are judged by the party doing the work.
- Protected property: the method surviving contact with ordinary increments; YAGNI.
- Insufficiency: no section states what scales down and what never does.
- Smallest correction: add a short proportionality section: the laws in §4 plus the Screen Contract (P9) and operator adjudication are **unconditional at any scope**; P4/P6/P7 activate on named triggers (new navigation context; new collection presentation; genuine structural ambiguity; new actor or need); a single-block change collapses to P0→P1-delta→P8→P9→P11-delta with the block ledger as the only bookkeeping.
- Generic methodology.

### 4. Optional findings

- **FO-1** (§14): when the single-hypothesis escape clause is invoked, require a one-line ledger justification ("conventional pattern X because task Y; no genuine ambiguity"). Keeps the clause honest and auditable without adding work to the common case.
- **FO-2** (§13/§23): add a disconfirming-reference prompt (at least one reference examined for why the leading pattern *fails* elsewhere) and a soft timebox note; both are the cheapest defenses against confirmation bias and unbounded browsing.
- **FO-3** (§14.1): add three missing decision variables — data volatility/refresh semantics (live-updating collections punish some layouts), realistic cold-start/empty-collection state (a grid of one item), and export/print/reporting needs (common in enterprise; favors tables).
- **FO-4** (§32): a generic minimum-artifact inventory per phase (what documents exist when the phase exits) would remove the last "what do I actually write down" invention for LLM consumers; currently only some phases imply their artifact.
- **FO-5** (§17): pattern entries could name validation timing semantics (inline vs on-submit; client-mirror vs server-owned) for form-class patterns — a recurring implementation-time invention seam not covered elsewhere.
- **FO-6** (§12): "exact-content behavior" as a material-surface trigger reads as consuming-product vocabulary escaping into the generic method; rephrase generically (e.g., "content-integrity/exactness guarantees") to keep the document ontology-clean.

### 5. Unsupported preferences

None recorded. Candidate complaints considered and discarded as taste: mandating a specific wireframe tool, demanding a design-token step before implementation, preferring fewer/renamed phases. No protected property behind any of them.

### 6. Phase-order critique

The macro order is correct and is the document's strongest feature: needs (P1) before flows (P2) before coverage (P3) before IA (P4) before inventory (P5) before evidence (P6) before alternatives (P7) before structure (P8) before contract (P9) before vocabulary (P10) before prototype (P11) before adversarial review (P12). Specifically adjudicated, as the brief requires: **structural wireframe → Screen Contract is the right order**, because P3 already provides the lightweight backend-sufficiency gate before any layout, and a full pre-layout contract would be waterfall waste — its content depends on the structure chosen. The correction needed is not a reorder but the P7 data-feasibility line (FI-3). The order defects that do exist are scoping, not sequencing: P4's exit status is ambiguous against the B01 block (FI-1), and §32's linear rendering contradicts the per-block loop (FI-2). The whole-product coherence checkpoints, once FI-1/FI-2 land, sit correctly at B01-lock (frame) and P12 (assembled product).

### 7. YAGNI / proportionality critique

For its designed case — whole-product planning of a non-trivial enterprise product with an LLM assistant — the method is proportionate: nearly every heavy instrument is already conditioned on materiality or ambiguity, and each targets a real recorded failure mode rather than ceremony. Three genuine YAGNI risks remain: no scaling rule for small scopes (FI-10, the largest); artifact proliferation without a stated minimum set (FO-4); and self-judged conditionality adjectives that let the diligent over-produce and the hurried under-produce. Nothing in the method warrants deletion; the fix is triggers, not amputation. The declined alternatives (mandatory design system upfront, mandatory user research program, production-framework prototyping) were all correctly declined.

### 8. LLM implementation-readiness critique

Against the brief's eight invention questions: *what screen to build* — answered (inventory + ledger); *what information comes first* — answered (P8 hierarchy) contingent on FM-2, since prose wireframes leave reading order unverified by the human; *which components/patterns to reuse* — answered once FI-4 fixes cadence; *where actions belong* — answered (P8 + contracts); *how navigation works* — answered (P4/P11) contingent on FI-1; *which backend call owns an action* — answered strongly (P9 bidirectional trace + §18.5 metadata, the document's best LLM affordance); *what state/error behavior to show* — half-answered: states and fixtures yes, user-facing language no (FI-8); *what changes responsively* — answered (§15/§25). Two meta-defects dominate LLM consumption: the loop-scope ambiguity (FI-2) lets a literal reader execute the wrong process shape while compliant, and the lock semantics (FM-1) let it self-certify progress. Both are text fixes, not redesigns.

### 9. Exact recommended corrections, smallest-first

1. §6: LOCKED is operator-only; no other party may set it. (FM-1a)
2. §4.6/§15.1/§22: single progression rule — next material block is not generated as baseline until the current one is LOCKED or the operator explicitly authorizes parallel progression; delete "sufficiently coherent" as a gate. (FM-1b)
3. §15: define the wireframe medium — rendered visual artifact required for P8 review; unstyled structural HTML skeleton explicitly permitted; P8/P11 boundary defined by scope/fidelity, not technology. (FM-2)
4. §11 exit: P4 IA exits CANDIDATE; LOCKED only via the global-shell block cycle. (FI-1)
5. §32: add the phase↔loop scope table. (FI-2)
6. §14: leading hypothesis carries a data-feasibility line before P8 lock. (FI-3)
7. §17/§22: pattern-consolidation pass after each block lock; P10 terminal reconciliation. (FI-4)
8. §5.2/§19: assumption register + mandatory P12 probe of material assumptions. (FI-5)
9. §15/§15.2: accessibility items in the wireframe-decides list and walkthrough questions. (FI-6)
10. §21: lawful orphan-operation dispositions; count = orphans-without-disposition. (FI-7)
11. §11.1/§16: durable glossary + failure message-intent line in Screen Contracts. (FI-8)
12. §21: post-visual-design structural-conformance check (or explicit transfer to implementation review). (FI-9)
13. New short section: proportionality triggers; §4 laws + P9 + operator adjudication unconditional at any scope. (FI-10)
14. FO-1…FO-6 at the Lead's discretion.

### 10. Re-review scope

Round 2 should be a **bounded confirmation**, not a fresh review: verify corrections 1–13 against FM-1/FM-2/FI-1…FI-10 as written here, plus a regression sweep that the corrections introduce no new mandatory ceremony (guarding the brief's second question) and no product-specific ontology. No re-litigation of phase order, severity vocabulary, or the declined alternatives — all adjudicated here as correct. If corrections 1–3 land as specified, no plausible Round-3 trigger is visible from this review.

```text
VERDICT = NOT CONVERGED
MATERIAL findings = 2
Round 2 justified = YES
```
