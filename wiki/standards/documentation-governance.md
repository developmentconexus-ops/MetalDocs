# Documentation Governance

> **Last verified:** 2026-08-14
> **Scope:** Ownership, authority, promotion, historical retention and anti-conflict rules for MetalDocs documentation.

## Core rule

- `wiki/` = **durable maintained truth**.
- `docs/` = **active staging / transient working material**.
- Git history = **archive of superseded/completed staging**.

The live working tree is not an archive. Do not keep obsolete plans/specs/reports beside current authority merely because they may be useful someday.

## Why

MetalDocs accumulated multiple roadmaps, milestone trees, specs, reports and living module pages that remained readable after their assumptions had been replaced. A new agent could follow an internally-consistent but obsolete path and implement the wrong architecture.

Documentation is therefore governed by **authority clarity**, not by maximum retention in the current checkout.

## Classification

Every maintained page should be understandable as one of:

- **ACTIVE/CANONICAL** — current maintained truth for its domain.
- **ACTIVE WIP/STAGING** — current design/plan not yet promoted; must identify its canonical parent/exit condition.
- **CURRENT-STATE REFERENCE** — accurately describes what runs now but is not target authority.
- **LEGACY/HISTORICAL** — retained only because live consumers still need the path or because the evidence is unusually valuable in-place.
- **SUPERSEDED/DEPRECATED** — replaced; must point to the successor if retained.

If a page's status cannot be determined quickly, that is documentation debt.

## Authority rules

1. One topic has one active authority.
2. Index/entrypoint pages MUST distinguish target truth from current-state/history.
3. A historical ADR can remain `Accepted` as a record of what was decided at that time, but active redesign programs may classify its **current relevance** separately until formal supersession/amendment lands.
4. Runtime/code/schema truth wins for the question “what runs today”.
5. An active, operator-approved redesign authority wins for “what target are we designing” when it explicitly reopens the historical boundary.
6. Do not use an implementation artifact as proof that the target abstraction should survive.

## `docs/` staging retention

### Allowed in the live tree

Keep only material that is actively being used to produce a current outcome, for example:

- the active design ledger;
- an active implementation plan after design approval;
- transient research directly feeding the current decision;
- current execution evidence until promoted/closed.

Every active staging artifact should name:

- status;
- owning program/design;
- authoritative parent;
- exit/promotion/deletion condition.

### Remove when finished/superseded

When a staged plan/spec/report/analysis is no longer active:

1. promote durable conclusions into the owning `wiki/` page/ADR if needed;
2. ensure current handoff/indexes point to the promoted authority;
3. delete the staging artifact from the live tree;
4. rely on Git history for archaeology.

Do **not** create an `archive/` copy of every deleted staging artifact. That reproduces the same authority problem under another folder name.

A live compatibility stub is justified only when a verified current consumer still requires the path. It must be short and redirect to the canonical successor.

## Cohesive Platform Redesign reset

On 2026-08-14 the old `docs/superpowers` accumulation was intentionally removed. The live directory now contains only:

- `docs/superpowers/README.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Do not restore the deleted roadmaps/milestones/plans/reports/specs/analyses from history as a normal workflow.

Current redesign entrypoint:

- `wiki/architecture/cohesive-platform-redesign.md`

Current recovery pointer:

- `wiki/references/current-agent-handoff.md`

## Wiki placement

Durable truth belongs under the owning domain:

- `wiki/architecture/` — system structure and target boundaries;
- `wiki/decisions/` — ADRs;
- `wiki/vision/` — product intent/personas;
- `wiki/modules/` — current module implementation references, not automatically target contexts;
- `wiki/database/` — current schema/migration truth;
- `wiki/workflows/` — durable end-to-end flows after lifecycle semantics are settled;
- `wiki/quality/` — QA/close-out governance;
- `wiki/standards/` — cross-cutting engineering/documentation standards;
- `wiki/references/` — operator/recovery references, not primary target truth;
- `wiki/backlog/` — governed deferred requirements; not a forward roadmap unless explicitly declared active.

`wiki/index.md` is the root canonical landing page. Major folders should have an `index.md` that exposes status/authority, not merely a file list.

## Legacy/current-state wiki pages

When a large current-state page would mislead target work but its path still has value:

- replace or prepend an obvious `LEGACY` / `CURRENT-STATE REFERENCE` marker;
- point to the active target authority;
- explain what remains useful about the old page;
- prefer Git history for the verbose previous narrative if the live body itself is the source of confusion.

Do not rewrite a legacy implementation page to pretend the new architecture is already implemented.

## ADR status field

An ADR's `> **Status:**` block MUST remain concise (≤3 physical lines and ≤400 characters).

Canonical status vocabulary:

```text
Proposed
Accepted
Accepted (amended YYYY-MM-DD by NNNN)
Superseded by NNNN
Deprecated
Historical
```

Long execution history belongs in the ADR body or a separate historical companion, not in the status field.

Run the existing ADR-status guard when ADRs change.

## Secrets

Secrets are referenced by location, never quoted in documentation, reports, commit messages or evidence.

Use forms such as:

```text
<redacted — see .env>
<redacted — see CI secret NAME>
```

Security findings must not reproduce the secret they report.

## Promotion checklist

Before promoting a WIP decision to durable wiki truth:

1. the decision is operator-approved/final for the current program;
2. authority/boundary is explicit;
3. conflicting active pages are updated, marked legacy or deleted;
4. indexes/handoff route to the promoted page;
5. implementation state is represented honestly (design ≠ implemented);
6. staging source is deleted when no longer active.

## New-document test

Before creating a new plan/spec/architecture page, ask:

> Does an existing active authority already own this information?

If yes, update that authority instead of creating another document.

A new document is justified when it has a distinct lifecycle/owner and a clear promotion/deletion condition.
