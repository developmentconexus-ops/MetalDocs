# Documentation Governance

> **Last verified:** 2026-06-11 (added "Secrets in documentation" rule per design decision D-4a)
> **Scope:** Ownership model for durable project knowledge, section boundaries, safe promotion rules, and the secrets-in-documentation rule.
> **Out of scope:** Rewriting every legacy document or mass-moving existing trees in one session.

## ADR status field (F9.1 — permanent rule)

An ADR's `> **Status:**` block — from the `> **Status:**` line through any following `>`-continuation
lines up to (not including) the next `> **Field:**` marker — MUST be **≤3 physical lines AND ≤400
characters total**. This is the "mega-status anti-pattern" guard: architecture review `778f494a` finding
105 flagged ADR 0022's status field growing to a 2757-character, 13-phase execution changelog, making the
decision state unreadable at a glance.

**Canonical vocabulary** for the status line: `Proposed | Accepted | Accepted (amended YYYY-MM-DD by
NNNN) | Superseded by NNNN | Deprecated | Historical`. The status line MAY be followed by one optional
date-and-scope clause and one optional history-pointer line (e.g. `Execution history:
[NNNN-execution-history.md]`) — still counted inside the 3-line/400-char budget.

**Execution history, phase-by-phase changelogs, and amendment narratives live OUTSIDE the status
field** — either in the ADR's own body (a `## Status history` / `## Amendment` section) or, when the
history is long enough to itself risk sprawl, in a companion doc `wiki/decisions/NNNN-execution-history.md`
linked from the status block's history-pointer line. Relocating history must not lose information —
restructure into dated entries, never summarize away facts.

**Repeatable sweep** — run the single-source gate script (reports every file whose status block
exceeds the budget; prints `adr-status: clean` and exits 0 when all pass):

```bash
bash scripts/check-adr-status.sh
```

**CI-enforced (F-R2).** This gate runs on every pull request as a **blocking** step in
`.github/workflows/governance-check.yml` (the `check` job invokes `scripts/check-adr-status.sh`). A PR
that introduces an over-budget ADR status block fails CI. The script — not an inline one-liner — is
the single source of the sweep, so the local command above and the CI gate can never diverge.

## Secrets in documentation (D-4a — permanent rule)

Secrets are **referenced by location, never quoted** — in any doc, report, commit message, audit artifact, or chat-derived summary. Write `<redacted — see .env>` (or the owning store, e.g. "see the CI secret `X`") instead of the value. This applies to *audit and security findings as well*: a report about a leaked credential must not itself reproduce the credential (the Stage-1/2 backend audit made exactly this mistake across 5 committed files — that is the incident this rule comes from, decision D-4a in `docs/superpowers/specs/2026-06-11-backend-professionalization-design.md`).

Enforcement: the `secret-scan` CI workflow (gitleaks) blocks checked-in secret values; this rule covers what scanners cannot reliably catch — deliberate quotation in prose.

## Ownership rule

- `wiki/` is the durable source of truth for maintained project knowledge.
- `docs/` is the staging and draft workspace for plans, specs, imported notes, prompts, templates, and non-canonical material that may later be promoted.
- Promotion goes by ownership and durability, not convenience. Durable material moves into the owning wiki domain folder. It does not get dumped at wiki root.

## Classification model

- `canonical wiki content` - maintained truth that should be the first-stop reference for the topic
- `staging/draft content` - design, planning, or research material that is still being shaped or not yet promoted
- `reference/archive content` - supporting evidence, handoffs, external-package notes, historical review packets, or archived context kept for re-entry
- `workflow/tooling gap` - missing index, governance rule, promotion path, or execution guidance that makes docs harder to trust
- `migration risk` - path, link, or ownership instability that makes a move unsafe without a phased pass
- `defer` - intentionally postponed cleanup or promotion

## Durable section boundaries inside `wiki/`

- `architecture/` - durable system structure, contracts, and boundary rules
- `backend/` - backend detail layer; produced by Stage-1 backend audit; owned by backend audit / professionalization program
- `modules/` - module-local living docs, debt registers, maturity evidence
- `database/` - schema ownership, dictionary, migration policy, reference data
- `standards/` - cross-cutting standards, including documentation governance
- `workflows/` - durable end-to-end flows
- `tests/` - repeatable acceptance and validation procedures
- `quality/` - QA operating-system home, release-quality rules, scenario-proof governance
- `decisions/` - durable ADRs
- `vision/` - durable product intent and target users
- `backlog/` - governed deferred work that remains active product memory
- `reviews/` - review packets and audit evidence
- `references/` - supporting operator references and historical aids that are useful but not primary canonical product truth

## What stays in `docs/`

- specs and plans under `docs/superpowers/`
- imported or exploratory research
- prompts and templates
- non-canonical runbooks that are phase-specific, transient, or still being normalized
- legacy ADR/history sets that have not been reconciled with the current wiki structure

`docs/` should not compete with `wiki/` for durable truth. If a `docs/` page becomes maintained truth, promote it into the owning wiki section and update indexes.

## Promotion rules

1. Confirm the content is durable enough to maintain.
2. Place it in the owning wiki domain folder.
3. Add it to the local folder index and root [../index.md](../index.md).
4. Leave a stable breadcrumb if the old path still has known consumers.
5. Avoid mass renames unless link impact is bounded and verified.

## Index rules

- `wiki/index.md` is the canonical root landing page.
- Each major domain folder should have its own canonical `index.md`.
- `README.md` inside wiki folders is optional and should exist only as a compatibility stub when older instructions or links still point there.
- Do not let any `README.md` evolve into a second full catalog once `index.md` exists.

## Current migration map

| Path | Classification | Decision |
|---|---|---|
| `wiki/references/ai-operating-system.md` | `reference/archive content` + `migration risk` + `compatibility bridge` | Keep path stable for now because `AGENTS.md`, `CLAUDE.md`, and handoff docs point here directly. It should explicitly bridge agents to the canonical QA loop in `wiki/quality/` until references are fully normalized. |
| `docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md` | `staging/draft content` | Keep in `docs/` as design input. It is not the live canonical wiki page. |
| `wiki/quality/deep-qa/index.md` | `canonical wiki content` | Canonical home for the promoted deep-QA artifact set. New durable links should point here. |
| `wiki/quality/deep-qa/runbook.md` | `canonical wiki content` | Canonical deep-QA execution runbook. |
| `wiki/quality/deep-qa/matrix.md` | `canonical wiki content` | Canonical deep-QA scenario matrix. |
| `wiki/quality/deep-qa/fixtures.md` | `canonical wiki content` | Canonical deep-QA fixture registry. |
| `wiki/quality/screen-qa-checklist.md` | `canonical wiki content` | Canonical reusable checklist for user-facing screen QA. |
| `wiki/quality/backend-api-qa-checklist.md` | `canonical wiki content` | Canonical reusable checklist for backend/API QA. |
| `wiki/quality/workflow-async-qa-checklist.md` | `canonical wiki content` | Canonical reusable checklist for async and workflow-owned QA. |
| `wiki/quality/release-closeout-checklist.md` | `canonical wiki content` | Canonical reusable checklist for final merge/release close-out. |
| `wiki/references/documents-approval-deep-qa/README.md` | `reference/archive content` + `compatibility breadcrumb` | Keep path stable because startup docs, module docs, and prompts still reference it directly. |
| `wiki/references/documents-approval-deep-qa/runbook.md` | `reference/archive content` + `compatibility breadcrumb` | Redirect-style compatibility page to the canonical QA location. |
| `wiki/references/documents-approval-deep-qa/matrix.md` | `reference/archive content` + `compatibility breadcrumb` | Redirect-style compatibility page to the canonical QA location. |
| `wiki/references/documents-approval-deep-qa/fixtures.md` | `reference/archive content` + `compatibility breadcrumb` | Redirect-style compatibility page to the canonical QA location. |
| `docs/superpowers/specs/2026-05-20-documents-approval-product-plus-qa-system-design.md` | `staging/draft content` | Keep in `docs/` as design/proposal input. Its durable operating outcome should land under `wiki/quality/`, not replace the draft spec. |
| `docs/runbooks/release-readiness.md` (removed) | `removed at re-baseline` | **Removed** at the v1 re-baseline (commit `c7f06f2e`). Its durable release gate was already promoted to the canonical [`wiki/quality/release-readiness.md`](../quality/release-readiness.md); the legacy staging draft no longer exists. |
| `docs/adr/*` (removed) | `removed at re-baseline` | **Removed** at the v1 re-baseline (commit `c7f06f2e`). The legacy ADR tree was superseded by [`wiki/decisions/`](../decisions/index.md); per-ADR reconciliation was completed in M0 F0.1 (see `decisions/index.md`). |
| `docs/ck5-wiki/*` (removed) | `removed at re-baseline` | **Removed** at the v1 re-baseline (commit `c7f06f2e`). External/reference CK5 knowledge was not retained in-repo. |
| `wiki/backlog/api-contract-hardening.md` → `wiki/_archive/backlog/api-contract-hardening.md` | `archived (closed program)` | **Relocated** in M0 F0.5. Closed program (Phase F shipped; closing re-audit 0 CRITICAL/0 HIGH); no active deferred work. Archived under [`wiki/_archive/`](../_archive/README.md), not deleted; inbound links repointed. |
| `wiki/backlog/contract-first-followups.md` → `wiki/_archive/backlog/contract-first-followups.md` | `archived (superseded)` | **Relocated** in M0 F0.5. Superseded — folded into the api-contract-hardening program (Phases C/E). Archived under [`wiki/_archive/`](../_archive/README.md), not deleted. |
| `wiki/backend/roadmap.md` | `historical — retained in place` | **Not relocated.** De-staled in M0 F0.3 with a top-of-file HISTORICAL banner (COMPLETE + Wave Z sealed). Kept in place because it is living-history cross-referenced by backend tracker docs; the forward surface is [`wiki/roadmap.md`](../roadmap.md). |
| `wiki/backlog/roadmap.md` | `historical — retained in place` | **Not relocated.** De-staled in M0 F0.3 with a top-of-file HISTORICAL banner (Refactor Roadmap; its open Plan 12 carried forward). Forward surface is [`wiki/roadmap.md`](../roadmap.md). |

## Explicit defers

- Do not mass-rename `README.md` files to `index.md` where existing path consumers are unknown.
- Do not delete `wiki/references/documents-approval-deep-qa/` while startup docs, prompts, and module memory still point to it.
- ~~Do not reconcile the legacy `docs/adr/` tree into `wiki/decisions/`~~ — obsolete: the legacy ADR tree was removed at the v1 re-baseline (commit `c7f06f2e`) and reconciled into `wiki/decisions/` during M0 F0.1.
