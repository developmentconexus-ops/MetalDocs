# Feature F0.4 — Evidence

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Feature:** `f0.4-backlog-hygiene`  ·  **Closed:** 2026-06-14
> A feature is closed only when every row below is filled with real output.

## What was implemented

Backlog hygiene pass under the operator's **fully-closed-files-only** rule. Read every
`wiki/backlog/*.md` and classified active-deferred vs fully-closed. Exactly **one** file is
fully closed with zero active deferred work — `api-contract-hardening.md` (PROGRAM CLOSED).
Marked it CLOSED in place and reorganized `backlog/index.md` to separate active from
closed/superseded. Physical relocation to `wiki/_archive/` is F0.5's job (archive dir does not
exist yet) — handed off with a trigger, nothing deleted.

| File | Change |
|------|--------|
| `wiki/backlog/api-contract-hardening.md:3` | Top ✅ CLOSED banner (completed program; → `_archive/` in F0.5); body intact |
| `wiki/backlog/index.md:9-20` | Split program-level list into **active** (`planned-endpoints.md`) vs **Closed / superseded** (`api-contract-hardening` CLOSED, `contract-first-followups` superseded, `roadmap.md` historical); `Last verified` 2026-05-27 → 2026-06-14 |

Not yet committed — staged for the M0 close-out commit batch (operator gate HS-1).

## Census (the hygiene judgment, evidence-backed)

Classification done by Explore sweep + per-row verification. Operator rule: a file stays unless
it has **zero** active `open` rows **and** zero live deferred-stub items.

- **Fully closed (archive):** `api-contract-hardening.md` — "Phase F shipped, PROGRAM CLOSED";
  closing 4-dimension re-audit 0 CRITICAL / 0 HIGH; all audit-ledger findings `closed`. Verified
  the single `open` grep hit at line 260 is description text (`open-by-default`) inside a row whose
  status cell = `closed 2026-06-05` — **not** an active row.
- **Superseded (→ F0.5):** `contract-first-followups.md` — no independent active work; "folded
  into api-contract-hardening Phase C/E" (which is itself now closed).
- **Stay — active `open` rows:** approval, audit, auth, controlled-documents, documents,
  editor-chrome, editor-ui-eigenpal, frontend-primitives, iam, novo-documento-wizard,
  render-fanout, search, taxonomy, templates (all `*-refactor.md`).
- **Stay — implemented but carry live deferred items (intentional stubs / missing-backend):**
  caixa-aprovacao, distribuicao, documento-publicado, editor, library-screen, novo-documento,
  novo-template-wizard, template-editor, templates.
- **Stay — forward sketches:** planned-endpoints.

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Gate A — fully-closed file marked CLOSED | `grep -n "CLOSED — completed program" api-contract-hardening.md` | line 3 banner present |
| Gate B — index separates active vs closed/superseded | `grep -nE "Program-level backlog \(active\)|Closed / superseded|CLOSED|superseded" index.md` | active section (l.9) + closed/superseded section (l.16-20) present; api-contract CLOSED, contract-first superseded, roadmap historical |
| Gate C — no backlog file deleted; docs-only | `ls -1 wiki/backlog/*.md | wc -l` + `git status` | **28** files intact (none removed); only 2 `wiki/backlog/**` modified; 0 code files |

> Docs-only feature — no build/test/runtime surface.

## Acceptance vs milestone spec

From `../milestone.md` F0.4: *"`wiki/backlog/*` contains active deferred work only; closed items
archived, not deleted-without-trace."*

| Acceptance criterion | Met? | Evidence |
|----------------------|------|----------|
| `wiki/backlog/*` contains active deferred work only | yes (judgment complete) | Census: every retained file has ≥1 active `open` row or live defer; the lone fully-closed file is marked CLOSED + queued for F0.5 relocation. `index.md` no longer presents any closed item as active. |
| Closed items archived, not deleted-without-trace | yes (no trace lost) | CLOSED banner + index "Closed/superseded" section preserve the trace; **physical move to `wiki/_archive/` is F0.5** (archive dir doesn't exist yet — bounded defer w/ trigger below). Zero files deleted. |

## Review disposition

- **Spec-compliance review:** ✅ compliant with operator rule A. No active deferred work
  archived (rule A held — every retained file verified to carry active work). No file deleted.
  HS-6 guard: classification is judgment-only; no backlog *content/decision* rewritten.
- **Code-quality review:** N/A — docs-only markdown. Index links verified against on-disk files.

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| Physical relocation of `api-contract-hardening.md` (CLOSED) + `contract-first-followups.md` (superseded) to `wiki/_archive/` | `wiki/_archive/` does not exist until **F0.5**; F0.4 owns the closure judgment, F0.5 owns the archive convention + move + governance-map rows. CLOSED banner + index section preserve the trace meanwhile | **F0.5** (`f0.5-archive-convention`) — relocate both + add governance migration-map rows |
