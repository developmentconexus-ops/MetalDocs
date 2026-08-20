---
id: documentation-governance
kind: authority
owner: engineering
summary: Defines MetalDocs documentation placement, naming, lifecycle, navigation, and temporary-work rules.
---

# Documentation governance

## One root

First-party maintained documentation lives under `docs/`.

Platform-conventional files may remain at repository root: `README.md`, `AGENTS.md`, `CLAUDE.md`, `.gitignore`, `mkdocs.yml`, and `.github/*`.

The live tree does not contain a second `wiki/` root, documentation archives, tombstones, superseded candidates, completed review artifacts, or historical roadmaps.

Git history and closed pull requests are the archive.

## Names

Durable paths use lowercase kebab-case semantic subjects, for example:

```text
product/contract.md
architecture/persistence.md
decisions/repository-reset.md
```

Durable filenames do not encode dates, R10/T-stage identifiers, versions, `final`, `candidate`, `review`, `adjudication`, `amendment`, `legacy`, or `historical` unless that token is genuinely part of the subject identity.

Ratified authorities imported during the clean-slate reset may retain their existing internal title/status/provenance block until the owning subject is next rewritten. Their **new semantic path** is the current navigation identity; cosmetic rewriting is not allowed to risk decision loss during reset.

New or substantively rewritten maintained Markdown pages use minimal frontmatter:

```yaml
---
id: unique-document-id
kind: authority | work
owner: owner-name
summary: One-sentence purpose.
---
```

`authority` is durable current truth for its declared subject. `work` is temporary non-authoritative material inside a Draft PR.

## Navigation

- `docs/index.md` is the human/agent intent router.
- `docs/status.md` is the sole stage / implementation-gate authority.
- `docs/decisions/index.md` is the compact decision registry.
- `mkdocs.yml` is the durable navigation manifest; publication infrastructure is optional.
- `docs/work/` is excluded from durable navigation.

Indexes link to authority; they do not repeat authority prose.

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

A new page must either own a unique current meaning or be temporary work. Do not keep compatibility stubs unless a real external consumer cannot be repaired in the same gate.

Generated maintained pages, if introduced later, receive metadata from their generator and are not hand-edited.

## Repository reset consequence

The previous documentation estate was intentionally removed together with the superseded implementation. Old paths, accepted ADR labels, QA artifacts, or roadmap statuses do not regain authority from Git history.

Only a current authority may name historical material as evidence.