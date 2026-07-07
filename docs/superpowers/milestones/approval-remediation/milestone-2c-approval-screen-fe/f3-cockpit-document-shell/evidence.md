# F3 — Evidence

## Commands + real output

- `npx vitest run src/features/documents/components/DocumentShell.test.tsx
  src/features/approval/pages/ApprovalCockpitPage.test.tsx
  src/features/documents/pages/DocumentEditorPage.test.tsx` →
  **Test Files 3 passed (3) · Tests 41 passed (41)** (DocumentShell 6, ApprovalCockpitPage 13,
  DocumentEditorPage 22). Re-run independently by the reviewer from clean state — same result.
- Broader regression sweep (`src/features/approval src/features/documents`): **48 files / 317
  tests passed** — no collateral breakage.
- `npx tsc --noEmit -p tsconfig.build.json` → **0 errors, clean** (implementer + reviewer both).
- `grep 'useDocumentSession|useDocumentAutosave' frontend/apps/web/src/features/approval`
  (excluding test mocks) → **ZERO** (main-session verified). The only match is the `vi.fn()` mock
  declarations in `ApprovalCockpitPage.test.tsx` that assert those hooks are **never called**.
- `grep 'ReviewDocumentCanvas' frontend/apps/web/src/features/approval` → **ZERO** (main-session
  verified). Deleted files (`SignoffDetailPage.tsx/.test.tsx`, `ReviewDocumentCanvas.tsx/.test.tsx`)
  confirmed absent via `ls`.
- Route swap verified: `routes.tsx:13-14` lazy-imports `./pages/ApprovalCockpitPage` →
  `module.ApprovalCockpitPage` for `approvals/:documentId`.

## TDD proof

- Implementer confirmed RED-first: both new test files failed on import resolution before the
  components existed — `Failed to resolve import "./ApprovalCockpitPage"` /
  `"./DocumentShell"`, `Test Files 2 failed (2)`. GREEN after implementing. Tests authored first.

## Runtime proof (observable change) + fixture-vs-real

- **W2 fix (the point of C1/W2), verified structurally + behaviorally:** the approval feature
  mounts **no** writer session and **no** autosave after F3. The cockpit passes DocumentShell no
  `onAutoSave` (`ApprovalCockpitPage.tsx:265-271`), and DocumentShell imports neither hook. A
  reviewer in `review` mode can suggest (client-side track-changes, surfaced via
  `onTrackedChangesChange`) and comment (comments API) but **cannot persist a document revision**.
  Asserted in test: session/autosave mocks `not.toHaveBeenCalled()` on **every** case including the
  eligible-review branch. Labeled **fixture/mock** (vitest + mocked editor/hooks).
- **Mode resolution, real field chain (not compiles-but-wrong):** eligibility compares
  `currentUser.userId === actor.user_id`. Main-session traced the chain — `auth.ts:43` maps
  `userId: value?.user_id ?? ''` from the API `CurrentUser.user_id`, and the DTO
  `ApprovalStageActorResponse.user_id` is the same TEXT `iam_users.user_id` (memory
  `tokens-actor-id-text-contract`). So review mode **actually activates** for an eligible
  reviewer — this is not a silently-always-readonly defect. Reviewer independently confirmed.
- **Author page zero-regression:** reviewer diffed against `HEAD~1` — buffer state/effect moved
  verbatim into DocumentShell, `skipInitialEditorChangeRef` dirty-guard retained, autosave /
  `onArtifactMetadata` / submit / rename all still page-owned and functional, chrome header renders
  during buffer-load exactly as before. `DocumentEditorPage.test.tsx` 22/22 green.
- **Real end-to-end** (a live approval-stage readonly cockpit and a review-stage suggesting cockpit
  against the running stack, plus author-page autosave still writing revisions) is exercised in the
  **F8 live-QA walkthrough**. Registered there, not claimed here.

## Key design decisions (verified against runtime truth)

- **Shell = editor-canvas region, not the outer page frame.** The only near-verbatim duplication
  was buffer-fetch + `MetalDocsEditor` mount (`DocumentEditorPage.tsx:109-154,519-536` ≈
  `ReviewDocumentCanvas.tsx:59-143`). Author frame (rail + `ArtifactMetaSidebar`) and cockpit frame
  (`ArtifactApprovalScreen` two-pane) genuinely differ and stay page-owned. Extracting the canvas
  kills the dup AND the W2 vector in one move; forcing a shared outer frame would have been a
  local-max contortion. `ArtifactApprovalScreen` stays in the cockpit — F4 owns the sidebar rebuild
  (and removes the duplicated "Fluxo de aprovação" band there).
- **`'review'` mode carries no autosave.** Reviewer suggestions are a live-session affordance
  feeding the verdict (F2 `reviewVerdict`) + comments (which persist). Durable suggestion
  persistence remains the program-level HS-2 bounded defer.
- **Fail-safe mode default `readonly`** — no writable/suggesting affordance without positive
  eligibility (approval stage / non-eligible / oversee / no active stage / unknown user all →
  readonly).

## Review / QA disposition

- Independent reviewer subagent (separate from implementer): **APPROVE**, **0 Critical, 0 Major,
  0 Minor**. All 8 adversarial checks PASS (W2-real, mode-correct incl. field-chain, author
  zero-regression via HEAD~1 diff, 404-not-toast, test-non-tautology, shell-contract-fidelity,
  generated-DTO discipline, gates-from-clean-state). Reviewer re-ran vitest (41/41) + tsc (clean)
  itself.

## Disclosed deviations (both accepted)

- **Additive prop `DocumentShell.onDocumentNameChange`** beyond spec §1's literal prop list —
  required to preserve the author page's existing inline-rename affordance (`handleRename`,
  exercised by the "E9 rename rollback" regression test) via `MetalDocsEditor`'s own callback
  rather than inventing a second mechanism. Optional, inert when omitted, **not** used by the
  cockpit → does not touch the W2 / mode-resolution contract. Reviewer judged it legitimate.
- **CSS module kept as `SignoffDetailPage.module.css`** (not renamed) and imported by
  `ApprovalCockpitPage.tsx` — the plan's sanctioned "pick one" option. Cosmetic; no behavior change.

## Bounded defers

- None new. Durable server-authoritative suggestion persistence remains the F0/HS-2 program-level
  bounded defer (client-authoritative track-changes + comments + freeze hash chain cover today).
- F4 will render `trackedChanges` (currently wired to state but unrendered) via `SuggestionList`
  and remove the duplicated timeline band in `ArtifactApprovalScreen`.
