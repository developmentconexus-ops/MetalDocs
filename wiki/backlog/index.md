# Backlog

> **Last verified:** 2026-06-21 (verify-and-archive sweep — see [`_cleanup-2026-06-21.md`](_cleanup-2026-06-21.md))
> **Scope:** Governed deferred work, refactor queues, and screen-specific follow-ups.

> **2026-06-21 hygiene sweep:** all refactor/screen backlogs were re-verified against HEAD `d477e9f0`
> (post Grade-A). Solved rows were pruned, `documents-refactor.md` was fully archived, and the
> Distribuição fanout/read-tracking domain was parked as a designed mission
> ([`document-distribution-mission.md`](document-distribution-mission.md)). Full audit trail:
> [`_cleanup-2026-06-21.md`](_cleanup-2026-06-21.md).

> **Forward roadmap:** the single canonical progression surface is [`wiki/roadmap.md`](../roadmap.md).
> The refactor roadmap below is **historical** (superseded); plan from `wiki/roadmap.md`.

## Program-level backlog (active)

- [planned-endpoints.md](planned-endpoints.md) - spec-ready sketches for endpoints removed from the live contract in Phase C but planned for future build (notifications, workflow transitions/approvals)

## Parked missions (designed, not started)

- [document-distribution-mission.md](document-distribution-mission.md) — **DESIGN-ONLY.** The full document distribution / read-tracking / fanout domain (read+ack evidence write-path, reader-side acknowledge surface, snapshot-vs-derive decision, reminders/export/fanout worker). Scoped out of frontend-screen-completion M2 (which builds only the derive-on-read coverage-*scope* subset). **Execute only after the frontend-screen-completion mission completes;** do not run `/mission` until then.

> Forward program progression lives in [`wiki/roadmap.md`](../roadmap.md). The refactor
> roadmap below is historical (superseded — see F0.3).

### Closed / superseded (archived under [`wiki/_archive/`](../_archive/README.md))

- [api-contract-hardening.md](../_archive/backlog/api-contract-hardening.md) - **CLOSED** completed program (Phase F shipped; closing re-audit 0 CRITICAL/0 HIGH). Archived at `wiki/_archive/backlog/`.
- [contract-first-followups.md](../_archive/backlog/contract-first-followups.md) - **superseded** — folded into api-contract-hardening Phase C/E. Archived at `wiki/_archive/backlog/`.
- [roadmap.md](roadmap.md) - ordered refactor roadmap (**historical** — superseded by [`wiki/roadmap.md`](../roadmap.md); its open Plan 12 carried forward)

## Module and platform refactors

- `approval`, `audit`, `auth`, `controlled-documents`, `iam`, `taxonomy`, `templates` (`documents` fully closed → archived `wiki/_archive/backlog/documents-refactor.md`)
- supporting frontend and infrastructure refactors such as `editor-chrome`, `editor-ui-eigenpal`, `frontend-primitives`, `render-fanout`, `search`

## Screen and workflow follow-ups

- library, template editor, novo-documento, novo-template-wizard, caixa-aprovacao, documento-publicado, distribuicao, editor

Backlog items stay in `wiki/` because they are governed active memory, not disposable planning notes.
