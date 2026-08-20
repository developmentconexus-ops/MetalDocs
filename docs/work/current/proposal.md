---
id: work-repository-information-architecture
kind: work
status: active
owner: architecture
summary: Proposes the docs-only information architecture, agent context model, review protocol, PR lifecycle, and legacy-deletion strategy for MetalDocs.
---

# Repository documentation and agent context architecture

> **Non-authoritative proposal.** Delete this file before the final governance PR is merged.

## 1. Decision

Rebuild the live documentation around stable subjects rather than historical stages, dates, review rounds, or current implementation modules.

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

Method outcome:

```text
RESTRUCTURE NOW
```

Updating a few indexes would preserve the defect class.

## 2. Root cause and invariants

PR #131 is valuable provenance, but it is no longer a safe continuing workspace:

```text
575 commits
1,018 changed files
39,173 additions
128,852 deletions
```

The live repository currently performs three incompatible jobs:

```text
current maintained truth
+
active working material
+
historical archive
```

That forces humans and agents to infer authority from filenames and chronology, increasing token use, stale-context risk, Git conflicts, and wrong-target implementation.

The replacement must guarantee:

1. One meaning has one current authority.
2. A reader can find that authority without knowing R10 stage codes.
3. Routine orientation uses a small deterministic read set.
4. Work-in-progress cannot be mistaken for durable truth.
5. Git and closed PRs preserve history; the live tree does not.
6. Review artifacts cannot survive merge accidentally.
7. Current-runtime documentation survives only for a named current consumer.
8. No operator-ratified Product/R10 decision is lost.
9. The reorganization changes documentation/governance only and authorizes no product implementation.
10. The repository profile remains reusable by other Conexus products without centralizing their truth.

## 3. Industry-aligned basis

This design reuses established documentation practices:

- Backstage TechDocs: Markdown beside code, repository-root `docs/`, `docs/index.md`, and `mkdocs.yml` navigation.
- arc42: architecture docs as plain-text docs-as-code, Git versioning, pull-request review, optional static publishing.
- Diátaxis: organize by reader need rather than the history of the team that produced the document.
- GitHub agent instructions: keep repository-wide instructions bounded and use path-specific instructions only where local context genuinely differs.

Primary references:

- https://backstage.io/docs/features/techdocs/
- https://backstage.io/docs/features/techdocs/creating-and-publishing/
- https://arc42.org/documentation/
- https://diataxis.fr/
- https://docs.github.com/en/copilot/how-tos/configure-custom-instructions-in-your-ide/add-repository-instructions-in-your-ide
- https://docs.github.com/en/copilot/concepts/prompting/response-customization

These references inform the mechanism. They do not create MetalDocs product requirements.

## 4. One documentation root

Select:

```text
docs/
```

Delete after authority and consumer parity is proven:

```text
wiki/
docs/superpowers/
docs/operator/
repository-local archive/tombstone trees
```

Platform-defined files may remain at the root:

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

## 5. Target information architecture

```text
MetalDocs/
├── README.md
├── AGENTS.md
├── CLAUDE.md
├── mkdocs.yml
└── docs/
    ├── index.md
    ├── status.md
    ├── product/
    │   ├── index.md
    │   ├── contract.md
    │   ├── journeys.md
    │   └── glossary.md
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
    ├── decisions/
    │   ├── index.md
    │   └── adr-0001-short-title.md
    ├── development/
    │   ├── index.md
    │   ├── setup.md
    │   ├── testing.md
    │   ├── verification.md
    │   └── contributing.md
    ├── operations/
    │   ├── index.md
    │   └── runbooks/
    ├── reference/
    │   ├── index.md
    │   ├── repository-map.md
    │   └── configuration.md
    └── work/
        └── current/
            ├── index.md
            ├── proposal.md
            └── ai-dialog.md
```

This is a placement model, not a requirement to create empty pages. Every surviving file needs a current consumer and unique meaning.

## 6. Naming and metadata

Durable filenames describe stable subjects.

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

Exceptions are explicit for ADR numbers, incident dates, and release notes.

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

Every live Markdown page under `docs/` has minimal frontmatter:

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
kind: authority | guide | reference | work
status: current | active
```

The live tree does not retain `legacy`, `historical`, `superseded`, or `tombstone` documents. Replacement means repairing links and deleting the previous page; Git retains its history.

## 7. Authority and navigation

Sole owners:

| Meaning | Authority |
|---|---|
| Current project/stage status | `docs/status.md` |
| Documentation discovery | `docs/index.md` |
| Product boundary | `docs/product/contract.md` |
| Product journeys | `docs/product/journeys.md` |
| Semantic ownership | `docs/architecture/ownership.md` |
| Domain state/invariants | `docs/architecture/domain-model.md` |
| Lifecycle/effectivity | `docs/architecture/lifecycle.md` |
| Authorization | `docs/architecture/authorization.md` |
| Audit | `docs/architecture/audit.md` |
| Exact content/storage/restore | `docs/architecture/content-integrity.md` |
| Async effects/Search | `docs/architecture/async-and-search.md` |
| Backend topology | `docs/architecture/backend.md` |
| Internal contracts | `docs/architecture/interfaces.md` |
| Persistence | `docs/architecture/persistence.md` |
| Executable API | `docs/architecture/api.md` |
| Frontend realization | `docs/architecture/frontend.md` |
| Runtime/deployment | `docs/architecture/runtime.md` |
| Transition/cutover | `docs/architecture/transition.md` |
| Compact decision registry | `docs/decisions/index.md` |
| Active proposal | `docs/work/current/proposal.md` |
| Lead↔Fable dialogue | `docs/work/current/ai-dialog.md` |
| Historical process | Git and closed PRs |

`docs/index.md` maps reader intent to the owning page. `mkdocs.yml` publishes only durable documentation; `docs/work/` is excluded.

Indexes point to authorities and do not repeat their decision prose.

## 8. R10 consolidation

Existing stage documents are source material, not target paths.

| Existing family | Target |
|---|---|
| Product Contract | `product/contract.md` |
| Whole-Product review conclusions | merge into owning pages; delete review artifact |
| Ownership topology | `architecture/ownership.md` |
| T1 | `architecture/domain-model.md` |
| T2 | `architecture/lifecycle.md` |
| T3 | `architecture/authorization.md` + `architecture/audit.md` |
| T4 | `architecture/content-integrity.md` |
| T5 | `architecture/async-and-search.md` |
| T6 | `product/journeys.md` + `architecture/api.md` |
| T7 | `architecture/transition.md` |
| T8-A | merge surviving conclusions into target/transition pages; delete census staging |
| T8-B | `architecture/backend.md` |
| T8-C | `architecture/interfaces.md` |
| T8-D | `architecture/persistence.md` |
| T8-E accepted checkpoint | fresh `work/current/proposal.md`, then `architecture/api.md` |
| Registry amendments | one `decisions/index.md` |
| Router | `status.md` |
| Handoff | `work/current/index.md` |

Stage identifiers remain in PR/decision provenance. They do not remain the documentation navigation model.

## 9. Agent context model

Root `AGENTS.md` is a short bootstrap/router containing only:

```text
repository purpose
docs/index.md entrypoint
docs/status.md when current stage matters
docs/work/current/index.md for active governed work
task-specific authority lookup
build/test/verify commands
stable Git/security rules
canonical Method pointer for material decisions
short Fable protocol pointer
```

It does not copy architecture, stage summaries, decision ledgers, review history, or prompts.

Default:

```text
one root AGENTS.md
```

Nested `AGENTS.md` files require repeated evidence of genuinely path-specific needs. `CLAUDE.md` remains a minimal pointer to `AGENTS.md` with no independent authority.

Routine orientation becomes:

```text
AGENTS.md
→ docs/index.md
→ one status/work pointer when relevant
→ 1–3 task-specific authorities
→ concrete code evidence only as needed
```

## 10. Active work and Fable

A governed architecture PR maintains one active proposal in:

```text
docs/work/current/proposal.md
```

It is edited in place. Do not create permanent candidates, corrected candidates, review requests, delta reviews, adjudications, or tombstones.

`docs/work/current/index.md` contains only:

```text
topic/gate
branch and PR
expected HEAD — revalidate
proposal path
checkpoint/open questions
exact next action
```

`ai-dialog.md` is created only when the proposal is ready for final independent review:

```markdown
# AI dialogue

## Review request
## Fable review
## Lead adjudication
## Bounded round 2
## Operator decision
```

Flow:

```text
Lead analysis
→ coherent proposal
→ operator convergence
→ one final Fable challenge
→ Lead adjudication in the same file
→ Round 2 only if a material contradiction survives
→ operator ratification
→ promote durable docs
→ delete proposal and ai-dialog
```

The Fable chat handoff contains only repository/branch/PR/HEAD and pointers to `AGENTS.md`, `proposal.md`, and `ai-dialog.md`. The canonical DevelopmentConexus Method/Fable workflow is referenced, not copied into a local skill.

## 11. PR lifecycle

One coherent ratifiable gate uses:

```text
one branch
one Draft PR
one active proposal
one final independent review
one operator decision
one squash merge
```

Architecture/governance PR flow:

```text
branch from current main
→ Draft PR
→ proposal converges
→ final Fable review
→ Lead adjudication
→ operator ratification
→ promote durable docs
→ delete temporary work files
→ required checks green
→ squash merge
→ next gate starts from updated main
```

A stage may be split only into independently coherent, ratifiable, merge-safe gates—not per conversation.

## 12. PR #131 and T8-E

PR #131 is frozen as provenance and receives no further architecture stages.

Clean sequence:

```text
G0 — repository information architecture/governance
G1 — current Product/R10 authority consolidation into semantic docs/
T8-E — fresh executable API-contract PR from the clean merged baseline
```

Before #131 is closed as superseded, G1 proves:

```text
all operator-ratified Product/R10 decisions through T8-D are present
all accepted T8-E checkpoint decisions are transferred
no product behavior is silently changed
no required current-runtime runbook is lost
all live links/tool consumers are repaired
```

The T8-E checkpoint to preserve includes:

```text
one OpenAPI application SSOT
one generated Go wire boundary
one generated TypeScript boundary
purpose-built responses and semantic operationIds
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

Upload/document corpus measurement remains open and must be carried forward rather than guessed during cleanup.

## 13. Allowlist deletion

A document survives only when all are true:

1. Named current consumer.
2. Unique current meaning.
3. Explicit authority class.
4. Reachable from navigation.
5. Named owner.
6. Clear lifecycle.

Otherwise:

```text
DELETE FROM LIVE TREE
```

Default deletion set after consolidation and link/tool repair:

```text
wiki/**
docs/superpowers/**
docs/operator/**
old roadmaps and milestone trees
old candidates/reviews/adjudications/tombstones
historical reports and review requests
superseded architecture pages
Registry amendment chain after consolidation
repository-local archives
skills duplicating the canonical Method/Fable workflow
```

Possible retention requires proof of a current consumer:

```text
runtime runbooks
local setup/verification instructions
security/recovery procedures
current schema/reference required before T10
ADRs referenced by current code/tooling
```

No compatibility stub survives unless an external consumer or unchangeable tool requires the old path.

## 14. Mechanical controls and proof

The final implementation adds a verifier with negative fixtures.

It must fail on:

```text
wiki/ or docs/superpowers/ in the final tree
unapproved Markdown roots
archive/legacy/tombstone directories
prohibited durable filenames
missing/invalid frontmatter
duplicate document ids
broken links
orphan durable pages
pages absent from mkdocs navigation
work pages included in published navigation
ai-dialog.md or unpromoted proposal in a merge-ready PR
```

Completion proof:

1. Fresh-agent routing works from `AGENTS.md` and `docs/index.md` without historical stage reading.
2. Every durable page is indexed, owned, and unambiguous.
3. Product/R10 authority parity against PR #131 passes.
4. Retained operational docs still satisfy named current consumers.
5. Verifier negative fixtures demonstrate every major rule fires.
6. Final PR contains no temporary work files.
7. Existing repository gates are not weakened and required CI is green.
8. PR #131 can be closed as superseded without losing current truth.

## 15. Non-goals and reopen triggers

This work does not implement R10, change product semantics, migrate business data, deploy Backstage, create a central Conexus documentation service, preserve old paths for convenience, or rewrite PR #131 history.

The shared Conexus profile is the structure/lifecycle. Each repository owns its product and architecture truth.

Reopen only on material evidence such as:

- a real external consumer requires historical documentation URLs;
- a required tool cannot operate with one documentation root;
- safe parallel gates cannot use the proposed work lifecycle;
- deleting a current-state page makes the existing runtime unsafe to operate;
- another product cannot fit the common profile without semantic distortion;
- agent measurements show routing remains materially ambiguous or expensive.
