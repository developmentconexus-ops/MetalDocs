---
id: documentation-governance
kind: authority
owner: engineering
summary: Defines MetalDocs documentation placement, naming, lifecycle, navigation, checkpoint, and temporary-work rules.
---

# Documentation governance

## One root

First-party maintained documentation lives under `docs/`.

Platform-conventional files may remain at repository root: `README.md`, `AGENTS.md`, `CLAUDE.md`, `.gitignore`, `.gitattributes`, and `.github/*`.

The live tree does not contain a second `wiki/` root, documentation archives, tombstones, superseded candidates, completed review artifacts, or historical roadmaps.

Git history, closed PRs, and explicitly retained provenance refs preserve historical process; they are not part of routine agent routing.

## Names

Durable paths use lowercase kebab-case semantic subjects, for example:

```text
product/contract.md
architecture/persistence.md
decisions/repository-reset.md
```

Durable filenames do not encode dates, R10/T-stage identifiers, versions, `final`, `candidate`, `review`, `adjudication`, `amendment`, `legacy`, or `historical` unless that token is genuinely part of the subject identity.

Ratified authorities imported during the clean-slate reset may retain their existing internal title/status/provenance block until the owning subject is next substantively rewritten. Their new semantic path is the current navigation identity; cosmetic rewriting must not risk decision loss during reset.

New or substantively rewritten maintained Markdown pages use minimal frontmatter:

```yaml
---
id: unique-document-id
kind: authority | checkpoint | work
owner: owner-name
summary: One-sentence purpose.
---
```

Meanings:

- `authority` — durable current truth for its declared subject;
- `checkpoint` — durable, non-authoritative accepted work snapshot intentionally preserved across merges until its owning gate resumes or is ratified;
- `work` — temporary, non-authoritative material inside a Draft PR and deleted before merge.

A checkpoint is never edited as active work. Resume by copying it into `docs/work/current/` and let the later ratified authority supersede it.

## Navigation

- `docs/index.md` is the human/agent intent router.
- `docs/status.md` is the sole stage / implementation-gate authority.
- `docs/decisions/index.md` routes durable decisions, remaining-stage ownership, and forward obligations.
- `docs/reference/` may contain explicitly routed durable checkpoints or bounded reference truth with a named consumer.
- `docs/work/` is temporary and excluded from durable navigation.

No static-site navigation manifest survives without a named publication/build consumer. `docs/index.md` is the current durable navigation surface.

Indexes point to authority/checkpoint; they do not copy detailed decision prose.

## Carried pre-reset authority text

Copied ratified authorities intentionally retain some old provenance strings. Any embedded `wiki/...` path is **non-navigational provenance**, even if old prose calls it `current`, `program authority`, or `decision baseline`.

Current routing replacements are:

```text
wiki/architecture/r10-technical-architecture.md
→ docs/status.md

wiki/architecture/r10-post-t6-implementation-readiness-program.md
→ docs/decisions/stage-program.md

wiki/architecture/rebaseline-decision-registry.md
→ current semantic authorities + docs/decisions/forward-obligations.md
```

Agents MUST NOT follow deleted `wiki/...` paths to determine current truth. This centralized routing law is stronger than rewriting thousands of provenance-only citations during the reset.

## Active work

A governed architecture PR may use:

```text
docs/work/current/index.md
docs/work/current/proposal.md
docs/work/current/plan.md
docs/work/current/ai-dialog.md
```

Use only the files actually needed by the gate. Proposal/plan are edited in place. `ai-dialog.md` exists only for the final independent review, Lead adjudication, bounded Round 2 when necessary, and operator decision.

Temporary work files are deleted before a governance/architecture PR is merged. Review provenance remains in Git/PR history.

## One meaning, one authority

A new durable page must own a unique current meaning or a uniquely named durable checkpoint. Do not keep compatibility stubs unless a real external consumer cannot be repaired in the same gate.

Generated maintained pages, if introduced later, receive metadata from their generator and are not hand-edited.

## Repository reset consequence

The previous documentation estate was intentionally removed together with the superseded implementation. Old paths, accepted ADR labels, QA artifacts, or roadmap statuses do not regain authority from Git history.

Only a current authority may name historical material as evidence.