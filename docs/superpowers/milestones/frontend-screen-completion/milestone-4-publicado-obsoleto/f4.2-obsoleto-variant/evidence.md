# Feature F4.2 — Evidence

> **Milestone:** 4 — Documento Publicado completion + Documento Obsoleto  ·  **Feature:** `f4.2-obsoleto-variant`  ·  **Closed:** 2026-06-23
> **Contract:** `spec.md` (consumer-contract-first; reuses `DocumentResponse.status` + `document.obsolete` capability — no new producer).
> A feature is closed only when every row below is filled with real, honestly-labeled output.

## What was implemented

Frontend-only. No Go/backend change (Non-goal honored — reuses the existing `document.obsolete` capability;
a dedicated "view-obsolete" cap was explicitly refused as HS-2 out-of-boundary). The "Documento Obsoleto"
state is a `status === 'obsolete'` **branch of the existing `DocumentPublishedPage`** — no second page file,
no forked route (reuse is itself an acceptance criterion, R2).

By outcome (each consumes a **frozen, already-shipped** input — F4.2 added no producer):

- **Page-wide dim** — `.rootObsolete { filter: grayscale(0.65); opacity: 0.85; }` applied to the root via a
  space-guarded class join `${styles.root}${isObsolete ? ` ${styles.rootObsolete}` : ''}`. Faithful port of
  `onda1-v5.jsx:746-747`. Absent when not obsolete (negative-control test).
- **"Vigente" pill hidden** — the green `vigenteBadge` is wrapped `{!isObsolete && ( … )}` (a doc cannot be
  both obsolete and "vigente"). Mirrors design `{!obsolete && <vigente pill>}` at `onda1-v5.jsx:190`.
- **"Visualizar documento" capability-gated** — `disabled={isObsolete && !canViewObsolete}` where
  `const canViewObsolete = useHasCapability('document.obsolete')` is called **unconditionally at the top-level
  hook block** (before both early returns — hooks-order safe). When disabled-for-obsolete a Portuguese `title`
  names the missing capability. **UX hint only** — backend `document.view` remains the sole authz boundary
  (inline comment cites `wiki/concepts/authz-tiers.md`; precedent `features/templates/lib/canActOnVersion.ts`).
- **"Baixar PDF" + "Copiar link" disabled when obsolete** — `disabled={isObsolete || …}` / `disabled={isObsolete}`.
  Mirrors design `disabled={obsolete}` on hero buttons (`onda1-v5.jsx:220-230`).
- **Rotated OBSOLETO watermark** — pre-existing `.obsoleteBanner`/`.obsoleteStamp` overlay (faithful design
  port) confirmed still rendered, unchanged.

## Validation Gate — acceptance mapped to proof

| Acceptance criterion (spec.md) | Proof | Real vs fixture | Result |
|---|---|---|---|
| `status:'obsolete'` ⇒ OBSOLETO watermark + obsolete status presentation, driven by real `status` | `DocumentPublishedPage.test.tsx` `F4.2` watermark+status case | fixture (vitest) | PASS |
| Root applies dim class only when obsolete | vitest: obsolete ⇒ root carries `.rootObsolete`; published ⇒ not | fixture (vitest) | PASS |
| "Vigente" pill hidden when obsolete | vitest: obsolete ⇒ no vigente; published ⇒ present | fixture (vitest) | PASS |
| Visualizar enabled with `document.obsolete` cap, disabled without (obsolete only) | vitest: obsolete + caps `['document.obsolete']` ⇒ enabled; caps `[]` ⇒ disabled + tooltip | fixture (vitest) | PASS |
| Baixar PDF + Copiar link disabled when obsolete | vitest: obsolete ⇒ both `disabled` | fixture (vitest) | PASS |
| Reuse — no new page file | `ls src/features/documents/pages` shows no obsolete page; `grep` confirms branch in `DocumentPublishedPage.tsx` (9 hits `isObsolete`/`canViewObsolete`/`rootObsolete`) | real (fs + grep) | PASS |
| Generated types consumed directly; capability via `useHasCapability`; no hand mapper | `npx tsc --noEmit` → exit 0 | real | PASS |
| No F4.1 regression | re-run `DocumentPublishedPage.test.tsx` + `documentDetailMeta.test.ts` ⇒ all green | fixture (vitest) | PASS |
| Both reviewer agents APPROVE | `frontend-screen-reviewer` + `frontend-code-reviewer` reports on record | real (review) | PASS (both APPROVE WITH NITS) |

## Commands + real output

- **TDD RED** — before implementation, 4 of the new `F4.2` cases failed as designed (dim class absent, vigente
  not hidden, Visualizar not gated, Baixar/Copiar not disabled); the watermark+status case passed (pre-existing
  branch). RED confirmed the tests bind behavior, not the implementation.
- **TDD GREEN** — `npx vitest run DocumentPublishedPage documentDetailMeta`:
  `Test Files  2 passed (2)` · `Tests  43 passed (43)` (DocumentPublishedPage 35 + documentDetailMeta 8). Zero
  F4.1 regression.
- **Static** — `npx tsc --noEmit` → exit 0.
- **Reuse** — `ls src/features/documents/pages | grep -i obsolet` → none (no separate obsolete page);
  `grep -c "isObsolete\|canViewObsolete\|rootObsolete" DocumentPublishedPage.tsx` → 9.

## Review disposition

Both required reviewers ran on the F4.2 diff and returned **APPROVE WITH NITS** (gate = both APPROVE → met):

- **`frontend-screen-reviewer`** — APPROVE WITH NITS. No Critical/Major. Confirmed field-by-field CSS parity of
  the OBSOLETO stamp and the page-wide dim against `design-source/documento-obsoleto`, and the no-fork reuse.
  - *Minor M1* — "Iniciar revisão" uses `aria-disabled` not native `disabled` for obsolete. **Accepted
    divergence:** spec §What-this-implements item 3 explicitly scoped this button as "already disabled for
    obsolete status (not published/approved); kept consistent" — pre-existing pattern, click handler guards on
    `canCreateRevision`, no functional regression. Not a gate failure. Tracked (see Defers).
  - *Minor M2* — mojibake `'Sem permissÃ£o para publicar'` tooltip. Pre-existing, outside the F4.2 obsolete
    branch (Publicar tooltip for `approved` docs). Tracked (see Defers).
- **`frontend-code-reviewer`** — APPROVE WITH NITS. No Critical. All three Majors (god-component 824 LOC,
  local `GovernedRevisionHistoryItem` type shadow, hardcoded CSS color literals) are **pre-existing F4.1
  inheritance**; the F4.2 delta (+113 LOC, 0 new files, 0 new exports) worsens none and is the wrong scope to
  fix them. Confirmed hooks-order safety (top-level `useHasCapability` before both early returns), real-status
  derivation, the UX-hint-only authz boundary comment, and full spec/contract cross-check (8/8 PASS).
  - *Minors n1–n3* (tooltip names a phrase vs the raw cap string; `beforeEach` mock self-documentation; fixture
    field symmetry) — taste-level; both reviewers approved without requiring them. Left as-is.

## Bounded defers (with triggers)

1. **Mojibake `'Sem permissÃ£o para publicar'`** (`DocumentPublishedPage.tsx` Publicar tooltip) — pre-existing
   double-encoding, outside the F4.2 obsolete branch. Tracked under the existing UTF-8-sweep task
   (`task_2b429f03`). *Trigger:* next edit to the `heroActions` Publicar branch, or a repo-wide mojibake sweep.
2. **"Iniciar revisão" `aria-disabled` → native `disabled` for obsolete** — accepted spec divergence (item 3).
   *Trigger:* a dedicated a11y pass on Publicado hero actions, or if AT-inconsistency is reported.
3. **God-component split** (`DocumentPublishedPage.tsx` 824 LOC > 400 LOC Major gate, `frontend-structure.md
   §15`) — pre-existing F4.1 inheritance. *Trigger:* next substantive feature touching this page; extract
   `ObsoleteBanner`/`HeroActions`/`SignoffPipeline`/`RevisionComposer`/`AboutCard`.
4. **Local `GovernedRevisionHistoryItem` type shadow** (`:42-49`) → replace with
   `components["schemas"]["DocumentRevisionHistoryItem"]`. Pre-existing. *Trigger:* the god-component split PR.
5. **Hardcoded CSS color literals** (`DocumentPublishedPage.module.css` lines ~245–297) → token-map.
   Pre-existing. *Trigger:* the god-component split PR / a styling-token sweep.

All defers are honest empty-scope or pre-existing tech debt with a written trigger; none is a silent F4.2 stub.

## Closure statement

F4.2 closed: every Validation Gate row PASS, TDD RED→GREEN on record, reuse + tsc are real checks, both
reviewers APPROVE, and all five defers are pre-existing/accepted with triggers. No new producer, no backend, no
new capability, no fork — the obsolete state is a faithful, real-status-driven variant of the one
`DocumentPublishedPage`. Ready for the milestone-validator (Phase 4).
