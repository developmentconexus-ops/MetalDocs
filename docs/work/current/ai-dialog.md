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

