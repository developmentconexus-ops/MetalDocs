# T8-H Fable independent review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8h-global-coherence`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate branch: `arch/t8h-global-coherence`

Exact candidate HEAD under review: `96c99fa6edfed5b015c5f93dbec9a40d806b8c95`

Required candidate CI: **#1098 SUCCESS** on that exact HEAD.

Review branch: `review/t8h-fable`

Gate: **T8-H — Whole-T8 Global Coherence Review**

Canonical Method: `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0

Repository Standard: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0

## Fresh-actor route

Reconstruct authority independently:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t8h-global-coherence.md
→ only the smallest owning T8 authority required for a concrete attack
```

The Whole-T8 authorities are:

```text
T8-A  docs/architecture/technical-baseline.md
T8-B  docs/architecture/backend.md
T8-C  docs/architecture/interfaces.md
T8-D  docs/architecture/persistence.md
T8-E  docs/architecture/wire-contract.md
T8-F  docs/architecture/frontend.md
T8-G  docs/architecture/runtime.md
```

Do not broad-read all upstream Product/T1→T7 material by default. Expand only when a concrete contradiction requires the owning authority. Repository current authority wins over this handoff.

Do not use removed implementation, legacy branches, old review chronology or historical `wiki/...` paths as target authority unless a material falsifier requires exact provenance.

## Candidate under attack

T8-H claims T8-A→T8-G now form one coherent technical realization after three bounded corrections:

```text
H1  mutable status duplication removed
H2  effective DocumentOfficialView consolidated into wire-contract.md
H3  application/maintenance reflected in T8-B for existing T5-J GC
```

The candidate further claims the CI gate was corrected proportionally without weakening material repository protection:

```text
valid Draft review/* + candidate base + ai-dialog-only delta -> green
review/* ready or isolation violation -> red
Markdown whitespace -> warning only
forward-obligation counts -> document-owned declarations, not workflow constants
superseded same-PR runs -> cancelled
architecture/routing/provenance guards -> still blocking
```

Application census must remain exactly **78**. Operation 79 remains absent unless material evidence forces the owning Product/T6 reopen.

T9 remains NOT OPEN. Product implementation remains BLOCKED.

## Review objective

Try to falsify the candidate, not confirm it.

A MATERIAL finding must identify a real contradiction, missing/duplicated authority, incompatible assumption, removed required seam, unsafe failure mode, weakened enforcement, or accidental complexity that affects a protected property. Preference, cosmetic wording, framework taste, legacy familiarity or hypothetical scale are not material findings.

Attack at least these classes:

### 1. H1 — sole mutable status authority

- Verify T8-F/T8-G durable architecture and ratification records retain immutable facts without continuing to act as mutable roadmaps.
- Verify `docs/decisions/index.md` exposes current decision dispositions without duplicating stage/integration progression.
- Search for any remaining stale `INTEGRATION PENDING`, downstream `NOT OPEN`, or implementation-status snapshot outside roadmap that would mislead a fresh actor.
- Distinguish legitimate frozen historical evidence from live mutable status.

### 2. H2 — executable wire SSOT

- Verify `docs/architecture/wire-contract.md` alone contains the effective `DocumentOfficialView` schema and disclosure/presence laws.
- Verify `docs/decisions/frontend-read-symmetry.md` is provenance only and cannot still supersede executable meaning.
- Verify `docs/index.md`, frontend authority and T8-F ratification route consistently to one executable wire authority.
- Verify consolidation did not silently change operation 47 behavior, add a persisted current pointer, add a Permission, create a new Problem/header, or alter the 78-operation census.
- Attack disclosure: absence of `open_revision` / `active_obsolescence_request_id` must not leak protected non-existence.

### 3. H3 — backend/application topology

- Verify `internal/application/maintenance` fits the already-ratified application class rather than creating a sixth semantic owner or a hidden new dependency class.
- Trace the T5-J managed-content GC consumer through T8-C/T8-D/T8-G and check that the leaf is neither Product route nor storage authority.
- Verify T8-B's statement about T6-derived human/API leaves remains coherent after admitting non-semantic maintenance choreography.
- Look for any other later T8-C/T8-G application consumer omitted from T8-B's frozen projection.

### 4. Whole-T8 authority coherence regression

Attack the highest-risk cross-stage seams again:

```text
T8-B topology ↔ T8-C interfaces
T8-C transaction/Audit/AuthZ ↔ T8-D persistence
T8-D concurrency/state ↔ T8-E ETag/CSRF/idempotency/exact-byte wire
T8-E wire ↔ T8-F route/read-model/transport consumption
T8-F concrete consumers ↔ T8-G runtime topology
```

Look specifically for:

- second Authorization evaluator or permission matrix;
- application/transport/platform semantic authority leakage;
- cross-owner SQL or duplicate current truth;
- same-commit Audit or River intent that cannot share the accepted transaction;
- idempotency/replay mismatch across backend, persistence, wire and frontend retry;
- exact-content mismatch across descriptor, ManagedContent, wire and verified spool;
- renderer/scanner/provider state acquiring Product meaning;
- Search becoming a second authority or requiring an unadmitted operation;
- frontend global/entity truth store or lifecycle authority;
- any concrete accepted human flow that requires operation 79.

### 5. CI proportionality — do not equate more red with more rigor

Review `.github/workflows/ci.yml` against Repository Standard v1 and the Method.

Verify the changed `required` gate still blocks the properties that matter:

- implementation paths while architecture-only mode is active;
- missing/duplicate bootstrap/status routing surfaces;
- bootstrap budget breach;
- mutable roadmap leakage through root routers;
- temporary work entering a non-Draft merge candidate;
- permanent archive/review/handoff surfaces;
- missing required Product/R10 authority;
- inconsistent forward-obligation census;
- broken/unrouted durable docs;
- durable docs depending on temporary work;
- required archive/provenance refs disappearing.

Then challenge the reductions:

- Is a valid isolated Draft review PR safely green, or did green status create a realistic merge path that the guard fails to stop?
- Does requiring review PRs to remain Draft plus failing `ready_for_review` preserve the non-merge evidence-channel property with lower noise?
- Does making generic whitespace non-blocking leave any actual repository correctness property unprotected?
- Does the document-derived forward-obligation count proof still fail on a missing/extra/mismatched obligation without making workflow constants a second authority?
- Could concurrency cancellation hide the only useful failing evidence for the current HEAD, or does it cancel only superseded same-PR runs?
- Is any remaining gate merely ceremonial/presence-only and worth removing, or is any removed blocker actually material?

A finding that merely says “CI should be stricter” is insufficient. Name the protected property, failure class and smallest enforcement.

### 6. YAGNI / Global Maximum

Attack both under- and overengineering:

- Did H1-H3 add machinery rather than consolidate existing meaning?
- Did the CI correction introduce unnecessary scripts/jobs/frameworks?
- Is any surviving T8 mechanism present without a named current consumer?
- Is any necessary seam missing because the architecture over-subtracted capability?
- Would the candidate remain valid if the old implementation had been structurally opposite?

## Candidate claims to distrust and re-check

```text
application operations = 78 exactly
operation 79 = absent
semantic owners = Authentication / Organization / Authorization / Controlled Documents + supporting Audit only
application/maintenance = non-semantic T5-J GC choreography only
one wire SSOT = docs/architecture/wire-contract.md
frontend-read-symmetry = provenance only
roadmap = sole mutable program-status authority
transaction/Audit/River seam remains coherent
CSRF session bootstrap remains reconstructible
idempotency replay remains live-authorized, PII-free, <=2048 bytes and 24h semantic window
exact content remains descriptor-owned and application-origin verified
same-PDF required rendition produces no renderer/copy/job
Search remains canonical PostgreSQL behavior
runtime remains one modular-monolith app + one PostgreSQL DB + in-process River baseline
Redis / BFF / realtime / external Search / generic event bus remain absent
CI #1098 proves only repository-envelope conformance, not architecture convergence
```

## Output contract

Write the independent review result below this handoff in this file only.

For each finding use:

```text
F#  MATERIAL | MINOR | NON-FINDING
claim
exact current authority/evidence
why it falsifies or does not falsify the candidate
smallest owning correction or reopen, if any
```

Finish with exactly one verdict:

```text
CONVERGED — MATERIAL findings = 0
```

or

```text
NOT CONVERGED — MATERIAL findings = N
```

Do not edit candidate authority from the review branch. Do not create Product requirements. Reviewer output is Evidence only.

---

## Fable review output

> Reviewer: independent Fable adversarial pass (T8-H, Round 1 of the corrected candidate).
> Candidate HEAD verified: `96c99fa6edfed5b015c5f93dbec9a40d806b8c95`; review base ancestry and single-file delta re-verified locally (`git diff --name-only` vs candidate = `docs/work/current/ai-dialog.md` only).
> Method: falsification-first. CI #1098/#1099 SUCCESS was treated as envelope conformance only, never as convergence evidence. Every claim below was re-derived from the current tree, not from the handoff.

### Verified baseline (evidence, not trust)

```text
census                      78 rows, operationIds 1→78, unique (mechanical count of §6 ledger)
operation 79                absent
operation 47                GET /api/v1/documents/{document_id} → 200 DocumentOfficialView, SAFE_READ, JSON_NO_STORE, B+N — unchanged
Idempotency-Key creations   exact 10 = 9 × IDEMPOTENT_CREATE rows + 1 × SUBMISSION_CREATE (createSubmission)
exact-byte ledger rows      4
OpenRevisionState           pre-existing wire enum (wire-contract.md:560, draft|submitted); no dangling type from H2
disclosure laws             §3.5 presence/disclosure laws present; absence-never-proves-non-existence stated; references grant no access
idempotency seam            T8-E 24h/fingerprint/2048-byte ReplaySnapshot == T8-D expires_at = completed_at + 24h, 32-byte fingerprint, janitor-never-semantic
audit seam                  same-local-commit law intact in T8-C/T8-D (AdmissionClaim, official_rendition.completed, release.completed)
H3 seam                     interfaces.md §20 + D34 host T5-J in internal/application/maintenance; backend.md leaf + class-law sentence added; runtime.md GC scheduling/shutdown/KEEP rows consume it; no new owner/class/route/operation
speculative infra           Redis/BFF/event bus/service mesh/external Search appear only in exclusion/REMOVE censuses
forward obligations         21/3/27 bullets == declared headings == Count proof == TOTAL 51 (document-internal, three-way)
```

### F1 — Severity: MATERIAL

**Claim.** The H1 correction is incomplete: live-reading mutable stage/status snapshots survive in durable authority documents outside `docs/roadmap.md`, and at least one now asserts stage state that is factually false, directly contradicting the roadmap.

**Evidence.**

1. `docs/architecture/persistence.md` §31 "Stage boundary and next stage" (lines 1711–1718) asserts, with no frozen-evidence framing:

```text
T8-D Persistence Realization = CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-E Executable Wire Contract = ACTIVE
T8-F→T8-H                   = NOT OPEN
T9→T12                       = NOT OPEN
implementation               = BLOCKED
```

   `T8-E = ACTIVE` and `T8-F→T8-H = NOT OPEN` are false: roadmap records T8-E/T8-F/T8-G CLOSED / OPERATOR-RATIFIED / INTEGRATED and T8-H OPEN / ACTIVE. A fresh actor routed to T8-D by `docs/index.md` ("Persistence realization") reads a durable authority that contradicts the sole mutable status authority.

2. `docs/product/alignment.md` "## Current gate" (lines 199–210) asserts as current: `T1 ACTIVE NON-AUTHORITATIVE CANDIDATE`, `T2→T7 NOT OPEN`, `implementation BLOCKED`, then routes to an "Active T1 staging packet" at `docs/superpowers/analysis/2026-08-18-...md` (a path absent from the tracked tree and inside the CI-forbidden `docs/superpowers` surface class) and states the technical sequence "is now owned by `wiki/architecture/r10-technical-architecture.md`" — an ownership claim into a pre-reset `wiki/` path that `AGENTS.md` declares provenance-only, never current routing. This is a routed durable Product authority (`docs/index.md` → "Whole-product alignment") presenting stale live status and dead authority routing.

3. Same class, weaker instances the correction law should sweep consistently: `docs/architecture/interfaces.md` §29 "The next **active** stage is: T8-D … Implementation remains **BLOCKED**"; `docs/architecture/technical-baseline.md` header `> **Status:** ACTIVE …` / `> **Implementation:** BLOCKED`. These duplicate mutable state (currently true or ambiguous) exactly the way the candidate itself judged material when it removed the still-true `T8-G/T8-H NOT OPEN` and `Product implementation BLOCKED` lines from the T8-F/T8-G ratification records.

**Owning authority.** `docs/roadmap.md` (sole mutable stage/status/next-action authority — also an AGENTS.md hard stop) versus `docs/architecture/persistence.md` and `docs/product/alignment.md`.

**Why it is a real defect rather than preference.** T8-H's own Whole-T8 claim (attack class 7, and H1 as ratified) is that durable architecture/ratification docs contain only stable facts, never stale progression snapshots, and that the roadmap is the only mutable status authority. Candidate proof P1/P2 passed only because P2 was scoped to "affected T8-F/T8-G records"; the Whole-T8 property is repository-wide and is falsified by a durable doc asserting `T8-E = ACTIVE` today. This is a live contradiction between two authority documents — the precise defect class H1 classified as material — plus dead routing claims (`wiki/` ownership, `docs/superpowers` packet) that violate AGENTS.md provenance law. It misleads exactly the fresh actor the reading law optimizes for.

**Smallest correction.** Extend the already-approved H1 correction, on `arch/t8h-global-coherence` only, to the remaining durable docs: replace time-indexed stage-state blocks with immutable closure facts plus the standard deferral ("current stage, integration status, implementation permission and exact next action are owned exclusively by `../roadmap.md`"); in `alignment.md`, delete or reframe the "Current gate" block and replace the `wiki/` ownership sentence and the `docs/superpowers` packet pointer with provenance-only framing (the sequence's current owner is the roadmap/T-stage authorities). No semantic content of T8-D or Product changes; prose-status edits only.

**Upstream reopen.** None. No T8-D persistence semantics, Product meaning, census, Permission or topology is touched. The reopen is the T8-H H1 correction scope itself.

### F2 — Severity: MINOR

**Claim.** The whitespace demotion in `.github/workflows/ci.yml` is broader than the ratified scope: `git diff --check` also detects leftover merge-conflict markers, so demoting the whole check to `::warning` removed the only mechanical block against unresolved conflict-marker corruption entering durable authority text.

**Evidence.** `ci.yml` lines 224–229: `whitespace_issues=$(git diff --check "$BASE_SHA...$HEAD_SHA" 2>&1 || true)` → warning only. Roadmap/ratified correction says "Markdown whitespace -> warning only"; conflict markers are not whitespace. In a repository whose merged tree IS the product (docs-only), a `<<<<<<<`-bearing merge candidate now passes `required`.

**Why not MATERIAL.** No accepted invariant names conflict-marker blocking; the KEEP list is not contradicted; the warning still surfaces the defect; squash-merge remains operator-authorized after human review, which is the deliberate-change guard. The loss is of a mechanical accidental-corruption backstop, bounded.

**Smallest correction.** In `ci.yml` only: keep whitespace warning-only, but fail on conflict-marker patterns in changed files (e.g. scan the PR diff for `^<{7} |^={7}$|^>{7} `), restoring the accidental-corruption block without re-blocking hygiene noise.

**Upstream reopen.** None.

### F3 — Severity: MINOR

**Claim.** The sole-roadmap leakage guard has no firing mechanism over durable decision/architecture/product docs, so the H1 defect class can recur silently; F1 is existence proof (both F1 instances predate T8-H and were never mechanically caught — `alignment.md` even contains the guard's own `Current gate` pattern).

**Evidence.** `ci.yml` line 98 greps only `README.md AGENTS.md docs/index.md`. `docs/product/alignment.md:199` contains `## Current gate`, which the guard's regex would match if it were in scope.

**Why not MATERIAL.** The guard's ratified KEEP scope is the root routers; a blanket extension to all durable docs would false-positive on legitimate frozen ratification evidence (e.g. `Round 2 … PR #141`), so the enforcement design is genuinely non-trivial and belongs to the candidate, not to reviewer preference. Recorded because the recurrence risk is now demonstrated, not hypothetical.

**Smallest correction.** Candidate's choice; e.g. after the F1 sweep, extend the guard with a narrow pattern class (live-status headings such as `^## Current gate` or `=\s*ACTIVE$` stage lines) over `docs/architecture docs/product docs/decisions`, tuned against the post-F1 tree.

**Upstream reopen.** None.

### Attacked and NOT sustained (evidence recorded so the Lead need not re-derive)

- **Review-PR merge path (class 15's sharpest attack).** `on.pull_request.types` includes `ready_for_review` (ci.yml:5), so flipping a green Draft review PR to ready re-runs `required` with `PR_DRAFT=false` → hard FAIL (ci.yml:39–42); Draft PRs cannot merge on GitHub at all; `synchronize` while ready also fails; a manually cancelled run yields `cancelled`, not `success`, on the latest check. Defense in depth: even if a review branch were ever merged into a candidate, any main-targeting PR of that candidate fails the `ai-dialog` tracked-surface guard (ci.yml:108). Green-while-Draft created no realistic merge path. NO FINDING.
- **Review-path early `exit 0` skipping later guards.** Proportional: the skipped guards verify candidate/tree properties already enforced on the candidate's own CI; the review delta is mechanically pinned to one temporary file first. NO FINDING.
- **Concurrency cancellation.** Group `ci-pr-${{ pr.number }}`, `cancel-in-progress` — cancels only superseded runs of the same PR; cross-PR evidence and the current HEAD's run are untouched; a cancelled run never satisfies a required check. No evidence-loss class found. NO FINDING.
- **Document-derived forward-obligation counts.** The new proof requires three-way internal coherence (heading count == bullet count == Count-proof row, TOTAL == sum). It fails on any missing/extra/mismatched obligation bullet. The old hard-coded `21/3/27/51` lived in the same PR-editable file class as the document, so it was never an independent second approval surface; nothing enforceable was lost, and a real second-authority duplication was removed. Verified live: 21/3/27 bullets == headings == proof == 51. NO FINDING.
- **H2 single wire SSOT.** `wire-contract.md` §3.1/§3.5 now carry `OpenRevisionRoutingReference`, the extended `DocumentOfficialView`, and the exact presence/disclosure laws plus a conformance-fixture line (wire-contract.md:1669); `frontend-read-symmetry.md` declares itself provenance and explicitly "no longer supersedes any executable schema"; `docs/index.md` executable-wire row routes to the wire contract only; `decisions/index.md` T8-E-FR row and `t8f-ratification.md` agree. The schema text repeated inside the provenance record is a frozen, currently-identical, explicitly non-superseding snapshot — same class as ratification evidence. No second executable authority survives. NO FINDING.
- **H2 disclosure attack.** Absence of `open_revision`/`active_obsolescence_request_id` is explicitly decoupled from semantic non-existence for callers lacking disclosure authority; the references are derived read truth, never persisted pointers, grant no access, and every follow-up performs canonical AuthZ. No leak of protected non-existence, no persisted current pointer, no Permission, census unchanged. NO FINDING.
- **H3 class law.** `application/maintenance` is admitted by T8-C D34/§20 as a non-semantic leaf inside the existing `application` class; T8-B's amended class-law sentence conditions maintenance leaves on a named upstream consumer (T5-J GC) and preserves "an omitted inbound leaf never creates a transport → owner exception". Two-phase GC choreography keeps provider delete outside the semantic transaction; storage authority stays with ControlledDocs/mechanisms. No sixth semantic owner, no dependency class, no Product operation, no runtime shell. Searched for other T8-C/T8-G application consumers missing from T8-B: none found. NO FINDING.
- **Cross-stage regression spot-attacks** (T8-C↔T8-D transaction/Audit, T8-D↔T8-E idempotency/ETag laws, T8-E↔T8-F read symmetry, T8-F↔T8-G consumers): the T8-H delta touches none of these semantics; targeted seam checks (24h/fingerprint/ReplaySnapshot/same-commit Audit/GC/River, ETag 13/13 and exact-byte 4 preserved-property blocks) agree across stages. NO FINDING.
- **Census.** 78 exactly, IDs 1–78 unique, operation 79 absent; H1/H2/H3/CI corrections added no operation, Permission, route, owner or runtime capability. NO FINDING.

### Verdict

```text
NOT CONVERGED — MATERIAL findings = 1
```

F1 reopens **only** the T8-H H1 correction scope on `arch/t8h-global-coherence` (prose-status sweep of `docs/architecture/persistence.md` §31, `docs/product/alignment.md` tail, and consistent treatment of the weaker T8-A/T8-C instances). No Product, T1→T7, T8-A→T8-G semantic authority, census, Permission, or persistence/wire/frontend/runtime meaning reopens. F2/F3 are bounded MINOR corrections in `.github/workflows/ci.yml` at the candidate's discretion.

A second review round is justified only to confirm the bounded F1 sweep (and any F2/F3 uptake); no full Whole-T8 re-attack is warranted.
