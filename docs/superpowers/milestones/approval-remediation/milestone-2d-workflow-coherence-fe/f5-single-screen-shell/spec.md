# Feature F2d.5 — Spec (consumer-contract-first)

> **Milestone:** 2d · **Feature:** `f5-single-screen-shell`
> **Governing:** `specs/2026-07-08-approval-workflow-coherence-design.md` §4 (A3) ·
> `specs/2026-07-08-single-screen-design-brief.md` (RATIFIED) — §2 matrix, §5 layout, §6 states, §9 decisions.
> **Approved (pre-code):** 2026-07-09 (destination + mode matrix + states operator-ratified; the one
> composition-structure call below is a code-grounded senior decision, recorded per the Global-Maximum rule).

## Consumer contract (what the end user + worklist require)

**Consumer:** end users (all roles) + worklist deep links. **Producer:** ONE mode-adaptive working screen
at the canonical URL `/documents/:documentId`.

The screen MUST, per derived `deriveWorkspaceMode(doc, instance, viewer)` (F2d.3), present the §2 matrix:

| Mode | Canvas | Sidebar footer | Contextual panel |
|---|---|---|---|
| `author-editing` | edits (autosave) | — | meta + route preview |
| `author-changes-requested` | edits | — | RequestedChangesPanel (F6 verdict reason) |
| `author-waiting` | read (comment replies = F2d.6, NOT here) | — | timeline + stage status |
| `reviewing` | read/suggest | **verdict CTAs** ("Pronto para aprovação" / "Solicitar mudanças") | timeline |
| `approving` | **frozen** read + integrity disclosure | **signature panel** (password + legal; delegation disclosure when `via_delegation_from`) | timeline |
| `observing` | read | — (only "Cancelar instância" when capability allows) | timeline |
| `lifecycle` | read | Publish/Schedule per `TRANSITION_POLICY` | timeline + lineage |

Structural requirements (§5/§6/§9):
- **Constant `DocumentShell`** (header: code, StateBadge, revision; central canvas; right sidebar). Shell
  never changes — the screen switches mode, not screens.
- **Unified right sidebar in ALL modes** incl. `author-editing` — the `ArtifactMetaSidebar` composition is
  retired (§9.2); meta + route preview become sidebar panels alongside the accountability timeline + footer.
- **Header mode chip** (PT-BR: Editando / Revisando / Aprovando / Visualizando / Aguardando revisão), Wine
  tokens, tooltip explaining WHY (from server `viewer` facts).
- **`DecisionFooter` variant = `stage_kind` + `viewer.eligible_for_active_stage`** (F4 contract honored) —
  three-way: `approving`→signature, `reviewing`(eligible)→verdict CTAs, else→nothing. Today it is binary
  (`decision != null ? signature : verdictCTAs`, `DecisionFooter.tsx:224`) which shows verdict CTAs to an
  ineligible observer — a defect this feature closes.
- **`changes_requested`**: highlight banner atop canvas + F6 panel.
- **Loading** = shell skeleton (no central spinner); **instance error** = sidebar error panel, canvas stays
  readable; **empty/edge** (no active stage, cancelled) = PT-BR teaching copy.
- **`?decision=approve|reject`** seeds the `approving` decision panel's `defaultOptionKey`.
- **Route restructuring (owned here):** (a) `DocumentDetailLayout`+children move `documents/:id` →
  `documents/:id/details`; (b) `documents/:id/edit` → `<Navigate>` redirect to `/documents/:id`;
  (c) new screen mounts at `documents/:id` (leaf); (d) `DocumentDistributionPage.tsx:95` breadcrumb href →
  `/documents/${id}/details`; (e) sidebar meta panel carries a discoverable link to `/documents/:id/details`;
  (f) `?decision=` survives — `InboxPage.openDecisionFlow` (`:110`) targets `/documents/:id?decision=…`, and
  the `/approvals/:documentId` route becomes a redirect forwarding `location.search`; (g) approver/observer
  (read-only) modes **lazy-load** the editing chunk (no editor payload for read-only modes).

## Composition structure (senior decision — Global-Maximum rule)

The governing spec says "`DocumentEditorPage` adapts by workspace mode." Taken literally, a 661-line editor
file growing to serve 7 modes is a local maximum. **Global-maximum structure:** a NEW thin owner
`DocumentWorkspacePage` (mounted at `/documents/:id`) owns (1) mode derivation, (2) the constant
`DocumentShell` + unified sidebar composition, (3) the header mode chip, (4) the mode→canvas/footer/panel
switch. The existing editor body is extracted into a lazily-imported **`EditorCanvas`** rendered ONLY in
`author-editing` / `author-changes-requested` (satisfies §9.3 bundle discipline for free). Approval canvas
(frozen read + disclosure), timeline, and footer compose in from `features/approval`. This honors "the
editor adapts by mode" (the editor becomes one mode's canvas within the unified screen) without ballooning
one file, and keeps each unit independently testable — the brief's "modular domains, one surface."

## Interview record (code-grounded)

| Q | Resolution | Grounding |
|---|------------|-----------|
| Evolve `DocumentEditorPage` in place or new owner? | New `DocumentWorkspacePage` owner + lazy `EditorCanvas` extraction (above) | `DocumentEditorPage.tsx` 661 lines; §9.2/§9.3; Global-Maximum rule |
| Does F2d.5 build author comment replies? | **No** — that is F2d.6 (`author-waiting` shows the read timeline only here) | milestone.md F2d.6 owns §9.1 |
| Footer variant source? | `stage_kind` + `viewer.eligible_for_active_stage` (equivalently the WorkspaceMode) — three-way | brief §5; F4 contract; `DecisionFooter.tsx:209-224` |
| Where do meta + route-preview panels live now? | Sidebar panels inside the unified sidebar (ArtifactMetaSidebar composition retired) | §9.2; `DocumentEditorPage.tsx:530` |
| Lazy-load mechanism? | `React.lazy` for `EditorCanvas` (none exists today under features/documents) + `Suspense` skeleton | investigator §7 |
| `/approvals/:documentId` redirect preserves query? | Yes — forward `location.search` (worklist `?decision=`) | brief §9.3; `InboxPage.tsx:110` |

## Non-goals (mandatory, written refusals)

- **Author comment replies / resolve** — F2d.6 (this feature renders only the read timeline in `author-waiting`).
- **Deleting `ApprovalCockpitPage` / `useDocumentApprovalArtifact` / the `onRefetchInstance` thread** — F2d.7
  (cockpit retirement). F2d.5 STOPS mounting the cockpit at a route (via the `/approvals/:id` redirect) but the
  file deletions + worklist-target cleanup are F2d.7. The old cockpit becomes unreachable, not yet deleted.
- **No** OpenAPI/DTO change (FE composition only).
- **No** new palette (Wine tokens only, §3).
- **No** mutation/If-Match/etag changes (F2d.4 owns instance state).

## Validation Gate (acceptance + named tests + proof)

Executed as ordered sub-slices (see plan.md). Aggregate acceptance:

| Acceptance | Named test | Proof |
|------------|-----------|-------|
| Footer: review-kind active stage renders verdict CTAs, NEVER the signature panel | `DecisionFooter` variant test (reviewing/approving/observing) | `vitest run …/DecisionFooter*.test.tsx` |
| Footer: approval-kind renders signature panel only when `eligible_for_active_stage` | same | same |
| Footer: ineligible observer on an active stage → neither variant (no CTAs) | new observer case | same |
| Screen renders the correct canvas+footer+panel per §2 for all 7 modes | `DocumentWorkspacePage` per-mode component tests (§2 matrix) | `vitest run …/DocumentWorkspacePage.test.tsx` |
| §6 states: loading skeleton, instance error (canvas readable), empty/edge | state tests | same |
| `changes_requested` banner + F6 panel present | mode test | same |
| `?decision=` preselects the approving decision panel | preselect regression test (M2c precedent) | same |
| `/documents/:id/details` renders `DocumentDetailLayout`; `/documents/:id` renders the workspace | route test | `vitest run …/routes*.test.tsx` |
| `/documents/:id/edit` redirects to `/documents/:id` preserving `:id` | route test | same |
| `/approvals/:documentId` redirects to `/documents/:id` preserving `location.search` | route test | same |
| Lazy: editing chunk dynamically imported, absent in `observing`/`approving` initial load | lazy-load assertion | same |
| Breadcrumb + deep-link retargeted (`DocumentDistributionPage`, `InboxPage`) | grep + route test | grep + vitest |
| build/types clean | `npx tsc --noEmit` | 0 errors |

## ADR?

Yes — **"Single artifact destination"** (governing spec §8 item 2): one mode-adaptive screen per artifact;
cockpit pattern retired. Recorded under `wiki/decisions/` and linked here at close. (ADR 0078 viewer-facts
and 0079 verdict-history already exist from F2d.1/F2d.2.)
