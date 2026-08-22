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
