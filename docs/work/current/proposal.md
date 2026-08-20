---
id: work-repository-information-architecture
kind: work
status: active
owner: architecture
summary: Defines the proposed docs-only information architecture, agent context model, review protocol, PR lifecycle, and legacy-deletion strategy for MetalDocs.
---

# Repository documentation and agent context architecture

> **Non-authoritative proposal.** This file exists only while the repository-governance change is under review. It must be deleted before the final PR is merged.

## 1. Decision requested

Adopt a single documentation root and rebuild the repository's live documentation around stable subjects rather than historical stages, dates, review rounds, or implementation modules.

```text
ONE docs/ ROOT
+
SEMANTIC STABLE FILENAMES
+
TASK-ORIENTED INDEX
+
EXPLICIT NAVIGATION
+
SHORT AGENT BOOTSTRAP
+
ONE TEMPORARY PROPOSAL
+
ONE TEMPORARY AI DIALOGUE
+
ONE PR PER RATIFIABLE GATE
+
GIT AS THE ONLY ARCHIVE
+
ALLOWLIST-BASED LEGACY DELETION
-
wiki/
-
docs/superpowers/
-
stage/date/version filenames
-
amendment chains
-
per-round permanent review artifacts
-
live-tree archives
-
duplicated status and authority
```

## 2. Evidence and root cause

The current R10 PR is a useful provenance source, but it has become an unsafe unit of continued work:

```text
PR #131
575 commits
1,018 changed files
39,173 additions
128,852 deletions
```

The live tree also contains multiple entry points that repeat or disagree about current status, while completed candidates, reviews, adjudications, tombstones, old roadmaps, module pages, and historical milestone trees remain searchable beside current authority.

The root cause is not merely a large number of files. It is that the live repository performs three incompatible jobs at once:

```text
current maintained truth
+
active working material
+
historical archive
```

This forces humans and agents to reconstruct authority from filenames and chronology. It increases token use, stale-context risk, Git conflicts, and the probability of implementing from an internally coherent but superseded document.

## 3. Target invariants

The restructured repository must keep these properties true:

1. One meaning has one current authority.
2. A new agent can locate the right document without knowing R10 stage codes.
3. Routine orientation requires a small deterministic read set.
4. Historical provenance remains available in Git and closed PRs, not in the live tree.
5. Work-in-progress cannot be confused with durable authority.
6. Review artifacts cannot survive merge accidentally.
7. Current runtime documentation is retained only when a named current consumer still needs it.
8. No accepted Product/R10 decision is lost during consolidation.
9. The reorganization changes documentation and governance only; it does not authorize product implementation.
10. The structure can be reused as a DevelopmentConexus repository profile without centralizing each product's truth.

## 4. Industry-aligned basis

This design deliberately reuses established documentation practices rather than inventing a MetalDocs-specific archive system.

### Docs as code

Backstage TechDocs uses a repository-root `docs/` directory, `docs/index.md` as the entry point, and `mkdocs.yml` for navigation. It keeps Markdown beside code and updates it through ordinary Git and pull-request workflows.

- https://backstage.io/docs/features/techdocs/
- https://backstage.io/docs/features/techdocs/creating-and-publishing/

arc42 recommends architecture documentation as plain-text docs-as-code, stored beside source, diffed in Git, reviewed through pull requests, and published with a static-site generator when useful.

- https://arc42.org/documentation/

### Reader-oriented information architecture

Diátaxis separates documentation by reader need—explanation, reference, how-to, and tutorial—rather than by the internal history of the team that produced it.

- https://diataxis.fr/

MetalDocs does not need to copy Diátaxis folder-for-folder. It does use the core separation:

```text
product/architecture explanation and authority
reference facts
how-to development and operations guides
active work kept separately
```

### Agent instructions

GitHub supports repository-wide and path-specific agent instructions, including `AGENTS.md`, and recommends path-specific instructions when context applies only to one area. This supports a short root bootstrap instead of loading all repository policy into every task.

- https://docs.github.com/en/copilot/how-tos/configure-custom-instructions-in-your-ide/add-repository-instructions-in-your-ide
- https://docs.github.com/en/copilot/concepts/prompting/response-customization

These references are implementation evidence and design guidance. They do not create MetalDocs product requirements.

## 5. One documentation root

Select:

```text
docs/
```

Delete after parity proof:

```text
wiki/
docs/superpowers/
docs/operator/
repository-local archive/tombstone trees
```

Files with platform-defined locations remain at repository root when applicable:

```text
README.md
AGENTS.md
CLAUDE.md
LICENSE
SECURITY.md
CONTRIBUTING.md
mkdocs.yml
.github/*
```

### Why `docs/`

`docs/` is immediately recognizable to people, GitHub tooling, MkDocs, and Backstage TechDocs. Keeping both `docs/` and `wiki/` creates an avoidable authority question. The target makes that question unrepresentable.

## 6. Target tree

```text
MetalDocs/
├── README.md
├── AGENTS.md
├── CLAUDE.md
├── mkdocs.yml
│
└── docs/
    ├── index.md
    ├── status.md
    │
    ├── product/
    │   ├── index.md
    │   ├── contract.md
    │   ├── journeys.md
    │   └── glossary.md
    │
    ├── architecture/
    │   ├── index.md
    │   ├── overview.md
    │   ├── context.md
    │   ├── ownership.md
    │   ├── domain-model.md
    │   ├── lifecycle.md
    │   ├── authorization.md
    │   ├── audit.md
    │   ├── content-integrity.md
    │   ├── async-and-search.md
    │   ├── backend.md
    │   ├── interfaces.md
    │   ├── persistence.md
    │   ├── api.md
    │   ├── frontend.md
    │   ├── runtime.md
    │   └── transition.md
    │
    ├── decisions/
    │   ├── index.md
    │   └── adr-0001-short-title.md
    │
    ├── development/
    │   ├── index.md
    │   ├── setup.md
    │   ├── testing.md
    │   ├── verification.md
    │   └── contributing.md
    │
    ├── operations/
    │   ├── index.md
    │   └── runbooks/
    │
    ├── reference/
    │   ├── index.md
    │   ├── repository-map.md
    │   └── configuration.md
    │
    └── work/
        └── current/
            ├── index.md
            ├── proposal.md
            └── ai-dialog.md
```

Only files with a real consumer survive. Empty folders and speculative documents are not created merely because they appear in this target map.

## 7. Stable naming law

Durable filenames describe the subject, not the process that produced the content.

Use:

```text
lowercase
kebab-case
short semantic nouns or noun phrases
```

Do not use in durable filenames:

```text
dates
R10/T-stage codes
v1/v2/revision numbers
final/new/old
candidate/corrected
review/adjudication
amendment/tombstone
legacy/historical
```

Exceptions are allowed only when time or numbering is the identity of the document:

```text
ADR numbers
incident dates
release notes
```

Examples:

```text
product/contract.md
architecture/persistence.md
architecture/api.md
operations/runbooks/restore.md
```

not:

```text
launch-v1-product-contract-rev001.md
r10-t8d-persistence-realization.md
2026-08-20-final-api-candidate.md
```

Versions, status, provenance, and review dates belong in frontmatter, the owning decision index, PRs, and Git history—not in stable paths.

## 8. Document classes and metadata

Every live Markdown document under `docs/` has minimal frontmatter:

```yaml
---
id: architecture-persistence
kind: authority
status: current
owner: architecture
summary: Defines the target relational model, constraints, and concurrency rules.
---
```

Closed vocabularies:

```text
kind:
  authority
  guide
  reference
  work

status:
  current
  active
```

The live tree does not retain documents with status:

```text
legacy
historical
superseded
tombstone
closed-candidate
```

When a page is replaced:

```text
promote replacement
→ repair links/navigation
→ delete replaced page
→ rely on Git history
```

## 9. Authority map

| Meaning | Sole current authority |
|---|---|
| Current project/stage status | `docs/status.md` |
| Documentation discovery | `docs/index.md` |
| Product boundary and launch contract | `docs/product/contract.md` |
| User/system journeys | `docs/product/journeys.md` |
| Semantic ownership | `docs/architecture/ownership.md` |
| Domain state and invariants | `docs/architecture/domain-model.md` |
| Lifecycle/effectivity | `docs/architecture/lifecycle.md` |
| Authorization | `docs/architecture/authorization.md` |
| Audit | `docs/architecture/audit.md` |
| Exact content/storage/restore | `docs/architecture/content-integrity.md` |
| Async effects and Search | `docs/architecture/async-and-search.md` |
| Backend topology | `docs/architecture/backend.md` |
| Internal communication contracts | `docs/architecture/interfaces.md` |
| Persistence | `docs/architecture/persistence.md` |
| Executable API contract | `docs/architecture/api.md` |
| Frontend realization | `docs/architecture/frontend.md` |
| Runtime/deployment realization | `docs/architecture/runtime.md` |
| Transition/cutover | `docs/architecture/transition.md` |
| Compact decision registry | `docs/decisions/index.md` |
| Active proposal | `docs/work/current/proposal.md` |
| Lead↔Fable review dialogue | `docs/work/current/ai-dialog.md` |
| Historical process | Git commits and closed PRs |

Indexes point; they do not duplicate the owned decision prose.

## 10. R10 consolidation map

The existing stage files are source material for consolidation, not target paths.

| Existing authority family | Target destination |
|---|---|
| Product Contract | `product/contract.md` |
| Whole-Product GCR conclusions | merge into owning product/architecture pages; delete review artifact |
| Ownership topology | `architecture/ownership.md` |
| T1 semantic state | `architecture/domain-model.md` |
| T2 lifecycle/effectivity | `architecture/lifecycle.md` |
| T3 authorization and audit | split into `authorization.md` and `audit.md` |
| T4 exact content/storage/restore | `content-integrity.md` |
| T5 async/Search/effects | `async-and-search.md` |
| T6 journeys/API | split into `product/journeys.md` and `architecture/api.md` |
| T7 migration truth | `architecture/transition.md` |
| T8-A disposition conclusions | merge into `transition.md` and relevant target pages; delete census staging |
| T8-B backend topology | `architecture/backend.md` |
| T8-C communication contracts | `architecture/interfaces.md` |
| T8-D persistence | `architecture/persistence.md` |
| T8-E accepted work | continue in `work/current/proposal.md`, then promote to `architecture/api.md` |
| Registry + amendments | one compact `decisions/index.md` |
| Router | `status.md` |
| Agent handoff | `work/current/index.md` |

Stage boundaries remain in decision provenance. They do not determine the information architecture.

## 11. Index and navigation

### `docs/index.md`

The landing page is task-oriented:

| I need to know… | Read |
|---|---|
| Where the project is now | `status.md` |
| What belongs to the product | `product/contract.md` |
| Which journeys must work | `product/journeys.md` |
| Who owns each meaning | `architecture/ownership.md` |
| How lifecycle works | `architecture/lifecycle.md` |
| How authorization works | `architecture/authorization.md` |
| How persistence works | `architecture/persistence.md` |
| What the API contract is | `architecture/api.md` |
| How to run or verify the repository | `development/index.md` |
| What proposal is active | `work/current/index.md` |

### `mkdocs.yml`

`mkdocs.yml` declares the durable published navigation. `docs/work/` is excluded from published navigation.

Build verification must fail when:

```text
a durable page is orphaned
an internal link is broken
navigation references a missing page
a work artifact is published
```

Adopting the directory and navigation format does not require deploying Backstage now. It preserves compatibility with MkDocs/TechDocs while remaining useful as plain Markdown in GitHub.

## 12. Agent context model

### Root `AGENTS.md`

The root file is a bootstrap/router only. It contains:

```text
repository purpose
read docs/index.md
read docs/status.md when status matters
read docs/work/current/index.md only for active governed work
how to locate task-specific authority
build/test/verify commands
stable Git and security rules
canonical Method pointer for material decisions
short Fable protocol pointer
```

It does not contain:

```text
R10 stage summaries
decision ledgers
architecture prose
review history
Fable prompts
lists of every document
```

### Path-specific instructions

Default:

```text
one root AGENTS.md
```

A nested `AGENTS.md` is introduced only after repeated evidence proves that one path requires materially different instructions. Path-specific guidance must never duplicate global architecture authority.

### `CLAUDE.md`

Retain the current minimal pattern:

```text
Read AGENTS.md first.
This file has no independent authority.
```

## 13. Active work protocol

A repository has at most one active governed architecture proposal in:

```text
docs/work/current/
```

### `index.md`

Contains only:

```text
gate/topic
branch and PR
expected HEAD — revalidate
proposal path
current checkpoint
open questions
exact next action
```

### `proposal.md`

One coherent proposal, edited in place until convergence.

Do not create:

```text
candidate-1.md
corrected-candidate.md
final-candidate.md
lead-adjudication.md
review-request.md
delta-review.md
tombstone.md
```

### `ai-dialog.md`

Created only when the proposal is ready for independent review.

```markdown
# AI dialogue

## Review request
## Fable review
## Lead adjudication
## Bounded round 2
## Operator decision
```

The file is temporary, non-authoritative, and deleted before merge.

## 14. Fable protocol

Fable is used once at the end of a coherent material decision package.

```text
Lead analysis
→ coherent proposal
→ operator convergence
→ create ai-dialog.md
→ final independent Fable challenge
→ Lead adjudication in the same file
→ Round 2 only if a material contradiction survives
→ operator ratification
→ promote durable authority
→ delete proposal and ai-dialog
```

The chat handoff stays compact:

```text
repository / branch / PR / expected HEAD
read AGENTS.md
review docs/work/current/proposal.md
write only in docs/work/current/ai-dialog.md
```

The canonical DevelopmentConexus Method and Fable workflow remain organizational authority. MetalDocs does not copy them into another local skill.

## 15. Pull-request lifecycle

One coherent ratifiable gate uses:

```text
one branch
one Draft PR
one active proposal
one final independent review
one operator decision
one squash merge
```

Do not create one PR per conversation or one permanent file per review round.

Architecture/governance PR flow:

```text
1. branch from current clean main
2. open Draft PR
3. maintain one proposal
4. converge with operator
5. final Fable review through ai-dialog.md
6. Lead adjudication in the same file
7. operator ratification
8. promote durable docs
9. delete temporary work files
10. required checks green
11. squash merge
12. delete branch
13. next gate starts from updated main
```

A stage may be split only into independently coherent, ratifiable, merge-safe gates.

## 16. Legacy deletion strategy

Use an allowlist, not file-by-file preservation by inertia.

A document survives only if all are true:

1. It has a named current consumer.
2. It contains unique current meaning.
3. Its authority class is explicit.
4. It is reachable through the new index/navigation.
5. It has an owner.
6. Its lifecycle is clear.

Otherwise:

```text
DELETE FROM LIVE TREE
```

### Default deletion set

After accepted truth has been consolidated and links/tooling repaired, delete:

```text
wiki/**
docs/superpowers/**
docs/operator/**
old roadmaps and milestone trees
old candidates/reviews/adjudications/tombstones
historical reports and review requests
superseded architecture pages
Decision Registry amendment chain after consolidation
repository-local archive directories
skills that duplicate the canonical Method/Fable workflow
```

### Possible retention set

Retain only after current-consumer proof:

```text
runbooks needed to operate the current runtime
local setup and verification instructions used by tools/operators
security and recovery procedures
current schema/reference material required before T10 cutover
ADRs still referenced by current code or tooling
```

No compatibility stub is retained unless an external consumer or unchangeable tool requires the old path.

## 17. PR #131 disposition

PR #131 is frozen as provenance. It must not receive additional architecture stages.

The clean replacement sequence is:

```text
G0 — repository information architecture and governance
G1 — current R10 authority consolidation into semantic docs/
T8-E — fresh executable API-contract PR from the clean merged baseline
```

Before PR #131 is closed as superseded, G1 must prove:

```text
all operator-ratified Product/R10 decisions through T8-D are present
all accepted T8-E checkpoint decisions are transferred
no product-code behavior is silently changed
no required current-runtime runbook is lost
all live links and tool consumers are repaired
```

PR #131 remains available as historical provenance and archaeology.

## 18. T8-E checkpoint preservation

The governance reset must preserve the already-approved T8-E direction, including:

```text
one OpenAPI application SSOT
one generated Go wire boundary
one generated TypeScript boundary
purpose-built response schemas
semantic operationIds
strong ETag / If-Match rules
bounded T6 resource precision
Idempotency-Key operation matrix
session-bound CSRF
stateless opaque pagination
RFC 9457 closed Problem catalog
errors.conexus.fun/{product}/{code}
create-only direct upload
server-authoritative completion
exact-byte response contract
Submission/Governance/Release/Rendition wire shapes
```

The exact upload/document corpus measurement remains open and must be carried into the fresh T8-E proposal rather than guessed during repository cleanup.

## 19. Mechanical controls

The final governance implementation must introduce a repository documentation verifier that can be demonstrated to fail.

### Structure

```text
wiki/ exists                         → FAIL
docs/superpowers/ exists             → FAIL
unapproved Markdown root             → FAIL
archive/legacy/tombstone directory   → FAIL
```

### Naming

For durable docs, prohibited filename tokens include:

```text
date prefixes
r10-t*
candidate
corrected
final
review
adjudication
amendment
legacy
tombstone
v1/v2
```

ADR and incident exceptions are explicit.

### Metadata

```text
missing required frontmatter
unknown kind/status
duplicate document id
→ FAIL
```

### Navigation

```text
broken link
orphan durable page
page absent from mkdocs navigation
work page included in published navigation
→ FAIL
```

### Merge hygiene

```text
docs/work/current/ai-dialog.md exists
unpromoted proposal remains in merge-ready PR
status authority conflicts with current work
→ FAIL
```

### Agent bloat

Bootstrap files may point to authority but must not copy durable decision prose. The verifier should prefer structural duplicate detection/allowlists over an arbitrary line-count limit.

## 20. Proof strategy

The rebaseline is complete only when all of these are demonstrated:

1. A fresh agent can identify the current task from `AGENTS.md` and `docs/index.md` without reading historical stages.
2. Every durable document is reachable from task-oriented navigation.
3. No live document has ambiguous class or owner.
4. `wiki/`, `docs/superpowers/`, archives, tombstones, and completed review artifacts are absent.
5. Accepted Product/R10 authority parity against PR #131 passes.
6. Current runtime operational documentation retained by allowlist still supports its named consumer.
7. The docs verifier has negative fixtures proving each major rule fires.
8. The final PR contains no `proposal.md` or `ai-dialog.md`.
9. Required repository CI is green without weakening an existing guard.
10. PR #131 can be closed as superseded without losing current authority.

## 21. Non-goals

This governance change does not:

```text
implement R10 product code
change product semantics
change target backend/persistence/API decisions
migrate production business data
introduce Backstage as a runtime dependency
build a central Conexus documentation platform
create a shared cross-product architecture authority
preserve old paths for convenience
rewrite Git history or force-push PR #131
```

## 22. DevelopmentConexus reuse

The reusable organizational profile is the structure and lifecycle:

```text
docs/
product/
architecture/
decisions/
development/
operations/
reference/
work/
```

Each product owns its own content and decisions. MetalDocs does not become the shared authority for Marketplace Central or other Conexus products.

After the profile is proven in at least two repositories, a small organization-level template may be extracted. Until then, reuse is manual and evidence-driven.

## 23. Reopen triggers

Reopen this design only if evidence shows one of these:

- a real external consumer requires stable historical documentation URLs;
- a tool cannot operate without a second documentation root;
- one active-work directory prevents safe parallel independent gates;
- MkDocs navigation cannot represent a required documentation consumer;
- deleting a current-state document makes the existing runtime unsafe to operate;
- a product requires a different information architecture that cannot be expressed through the common profile;
- agent measurements show the new routing still requires excessive or ambiguous context.

## 24. Proposed outcome

```text
RESTRUCTURE NOW
```

The current documentation shape preserves the defect class. Updating a few indexes would be a local maximum. The selected structure removes duplicate authority, makes historical artifacts non-live by default, and establishes mechanically enforceable lifecycle rules.
