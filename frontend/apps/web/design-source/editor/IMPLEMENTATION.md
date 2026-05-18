# Editor — Implementation Worksheet

> **Slug:** editor
> **Owning feature:** features/documents
> **Target route:** /documents-v2/:documentID
> **Reference:** ./editor.html + ./selected-editor-v2.jsx
> **Skill version:** 1.0
> **Started:** 2026-05-06
> **Completed:** 2026-05-06

---

## Open Questions Log

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 1 | 0 | Formatting toolbar — keep eigenpal's or build shell-level? | Keep eigenpal; add collapsible right sidebar | ✅ |
| 2 | 0 | Eigenpal width — shrinks when sidebar open? | Implicit yes (sidebar collapsible; eigenpal flex:1) | ✅ |
| 3 | 0 | "Submeter para revisão" = `handleFinalize()`? | Yes, label change only | ✅ |
| 4 | 3a | Toolbar style — feature-local `EditorDocBar` or unify with TemplateAuthorPage? | Unify; extract shared `EditorChrome` primitive | ✅ |
| 5 | 3a | Back button placement — overlay (left slot) or rail? | Rail (overlay collided with eigenpal File icon) | ✅ |
| 6 | 3a | Keep Revisões + Export buttons in doc bar? | No, remove both | ✅ |
| 7 | 3a | re2 template-annotation chip in doc editor sidebar | Remove; gate `templatePlugin` to `template-draft` mode in `MetalDocsEditor` | ✅ |

---

## Phase 0 — Audit (HARD GATE)

### 0.1 Element-by-element audit

| Element (HTML region) | Maps to (state / role / persona / data) | Keep / Cut / Defer | Reason |
|---|---|---|---|
| Slim doc bar — back button | `onDone()` callback → navigate back to library | Keep | Already exists as overlay pattern; redesign to proper bar |
| Slim doc bar — code chip (`POP-RH-001`) | `doc.Code` from `DocumentResponse` | Keep | Available in API response |
| Slim doc bar — doc name | `documentName` state (editable via `handleRename`) | Keep | Available + rename already wired |
| Slim doc bar — version+status (`v5 · rascunho`) | `doc.RevisionVersion` + `doc.Status` | Keep | Available in API response |
| Slim doc bar — autosave indicator (`Salvo · há 12s`) | `autosave.status` from `useDocumentAutosave` | Keep | Already exists, needs design port |
| Slim doc bar — "Revisões" button | Maps to existing Checkpoints dialog (`setCheckpointsOpen`) | Keep | Same concept, rename button label |
| Slim doc bar — "Salvar" button | `handleSave()` | Keep | Already exists |
| Slim doc bar — "Submeter para revisão" button | `handleFinalize()` → `POST /api/v1/documents/:id/finalize` | Keep (⚠️ Q3 open) | Likely same as "Finalizar". Label change. |
| Formatting toolbar (heading select, B/I/U, lists, link) | Eigenpal's own toolbar — already inside `ep-root.docx-editor` | CUT (⚠️ Q1 open) | Eigenpal renders its own toolbar. Shell-level duplicate = two toolbars |
| Word count (`1.247 palavras · 7 seções`) | Eigenpal internal — not exposed via `MetalDocsEditorRef` | CUT | No API to read word count from eigenpal. Can defer if eigenpal exposes it later |
| Paper canvas (eigenpal editor area) | `MetalDocsEditor` component — `buffer`, `isEditable`, session state | Keep | Just give it proper space. No implementation needed from us |
| Paper document header (code/version/status printed on page) | DOCX template content inside eigenpal's DOCX | CUT | Inside the DOCX, not our shell. Template controls this |
| Right sidebar — "Metadados" heading | Section header, no data | Keep (shell) | Static label |
| Right sidebar — Código row | `doc.Code` | Keep | Available |
| Right sidebar — Perfil row | Profile name — NOT in `DocumentResponse` | DEFER | Backend must extend `/api/v1/documents/:id` to return profile name |
| Right sidebar — Área row | Area name — NOT in `DocumentResponse` | DEFER | Backend must extend endpoint to return area name |
| Right sidebar — Vigência atual row | `doc.RevisionVersion` + approval date | Defer (partial) | RevisionVersion available, approval date not in response |
| Right sidebar — Próx. revisão row | Next review date — NOT in `DocumentResponse` | DEFER | Backend must extend endpoint |
| Right sidebar — Visibilidade row | Visibility scope — NOT in `DocumentResponse` | DEFER | Backend must extend endpoint |
| Right sidebar — "Revisões" timeline (list v1..v5) | Full revision history list — NOT available. Only `CurrentRevisionID` + `RevisionVersion` in current `DocumentResponse` | DEFER | Needs `GET /api/v1/documents/:id/revisions` returning list |
| Right sidebar — "Próximos aprovadores" section | Approval signoff list — exists in approval module but not wired to editor | DEFER | Needs `GET /api/v1/documents/:id/signoffs` or equivalent, wired from editor |

### 0.2 Cut list confirmed

- [x] User reviewed cut list
- [x] Cuts recorded in NOTES.md (cuts: formatting toolbar, word count, paper doc header)

---

## Phase 1 — Map (HARD GATE)

### 1.1 Reusability scan — backward

| Design element | Existing primitive | Path | Action |
|---|---|---|---|
| Code chip (`POP-RH-001`) | `CodeChip` | `components/ui/CodeChip.tsx` | Use |
| Status pill on doc bar | `StatusPill` | `components/ui/StatusPill.tsx` | Use — eliminates inline `statusPillClass` record in existing page |
| Approver avatars (sidebar, deferred) | `Avatar` | `components/ui/Avatar.tsx` | Use when wired |
| Revisões timeline (sidebar, deferred) | `TimelineRail` | `components/ui/TimelineRail.tsx` | Use when wired |

### 1.2 Reusability scan — forward

| Name | Generic? | Used by 2+ screens? | Placement | Rationale |
|---|---|---|---|---|
| `EditorChrome` (+ `VersionBadge`, `AutosaveStatus`) | Yes — slot-composition shell with eigenpal title-bar/formatting-bar overrides | Yes — `DocumentEditorPage` + `TemplateAuthorPage` | `features/shared/components/editor-chrome/` | Promoted from feature-local `EditorDocBar` after toolbar-unification. Pages pass left/center/right/alert slot content. |
| `EditorMetaSidebar` | No — metadata fields specific to document entity | No | `features/documents/components/EditorMetaSidebar.tsx` | Document-editor-specific sidebar |

**Drift note:** Q4 in Open Questions Log changed the original feature-local `EditorDocBar` plan. `EditorDocBar.tsx` deleted; both pages now consume `features/shared/components/editor-chrome`.

### 1.3 Component decomposition (post-implementation)

```
DocumentEditorPage (modified, features/documents/pages/DocumentEditorPage.tsx)
├── <aside className={styles.rail}>          (NEW — back button column; mirrors TemplateAuthorPage rail)
│   └── <button className={styles.railBackBtn}> chevron + tooltip → onDone()
├── <main className={styles.canvas}>
│   └── EditorChrome (shared, features/shared/components/editor-chrome/)
│       ├── center slot
│       │   ├── CodeChip — doc.Code
│       │   ├── <span> documentName (editable via handleRename)
│       │   ├── VersionBadge — `v${revNum}`
│       │   └── StatusPill — docStatus as DocumentStatus
│       ├── right slot
│       │   ├── AutosaveStatus — autosaveState (idle/saving/saved/error)
│       │   └── <button primaryBtn> "Submeter para revisão" → handleFinalize(), disabled if !isEditable
│       └── children
│           └── MetalDocsEditor (mode='document-edit' | 'readonly')
│               └── PluginHost — templatePlugin SKIPPED for non-template-draft modes (kills re2 chip)
└── EditorMetaSidebar (features/documents/components/EditorMetaSidebar.tsx)
    collapsible (open: 300px | closed: 0 + toggle strip)
    └── sections per Phase 1.3 mocks (TODO trail planted, see wiki/backlog/editor.md)
```

**Cuts from original plan (vs implemented):**
- `EditorDocBar` deleted — replaced by shared `EditorChrome`
- Revisões button + `CheckpointsDialog` mount removed (Q6)
- ExportMenuButton removed from doc bar (Q6)
- Back button moved from EditorChrome left slot → dedicated rail (Q5)

### 1.4 Status / enum meta SSOT

`StatusPill` (`components/ui/StatusPill.tsx`) already owns status → label+pillClass. No new file needed.

Remove the inline `statusPillClass` record from `DocumentEditorPage.tsx` — it duplicates StatusPill's own config. Replace with `<StatusPill status={docStatus as DocumentStatus} />` in `EditorDocBar`.

### 1.5 State design

| Type | Item | Notes |
|---|---|---|
| Server state | `getDocument(documentID)` via useEffect — keep as-is | Session system too complex to migrate to TanStack Query in this pass |
| Local state | `checkpointsOpen: boolean` — existing | Unchanged |
| Local state | `sidebarOpen: boolean` — NEW | Toggle sidebar visibility |
| Persisted | `editor-sidebar-open` localStorage key | lazy `useState<boolean>(() => localStorage.getItem('editor-sidebar-open') !== 'false')` — default open |
| Local state | `doc`, `documentName`, `buffer` — existing | Unchanged |

### 1.6 Backend contract

| Endpoint | Path | Status | Shape (if needed) | Backlog issue |
|---|---|---|---|---|
| Get document | `GET /api/v1/documents/:id` | Existing — but missing fields | Needs `ProfileName`, `AreaName`, `NextReviewAt`, `Visibility` added to response | wiki/backlog/editor.md |
| Revision history list | `GET /api/v1/documents/:id/revisions` | NEEDED | `[{ ID, VersionNum, Label, CreatedAt, CreatedBy }]` | wiki/backlog/editor.md |
| Approver list | `GET /api/v1/documents/:id/signoffs` | NEEDED | `[{ ActorID, ActorName, Role, Status: 'next'\|'wait' }]` | wiki/backlog/editor.md |

Mock strategy: all three "needed" fields/endpoints use hardcoded MOCK consts inside `EditorMetaSidebar` with `// TODO(editor:meta)` trail + `wiki/backlog/editor.md` row.

### 1.7 User review checkpoint

- [x] Reusability classifications reviewed
- [x] Backend contract reviewed
- [x] No open Phase-1 questions

---

## Phase 2 — Pre-flight (advisory)

- [x] OpenAPI codegen — not needed (no new endpoints in this pass)
- [x] Primitive fixes/extensions — not needed (all 4 primitives used as-is)
- [x] Status-meta — not needed (StatusPill owns it)
- [x] New shared atoms — none (both new components are domain-specific)
- [x] Route stub — not needed (route already exists)

---

## Phase 3a — Structure mirror (HARD GATE)

- [x] DOM tree mirrors design HTML
- [x] CSS Module class names = direct rename
- [x] No logic yet
- [x] Main agent confirmed match

---

## Phase 3b — Style port (HARD GATE)

### 3b.1 Token map

| Design value | Existing token | New token (if needed) |
|---|---|---|
| Brand maroon `#6b1f2a` | `--brand` | — |
| Brand soft hover | `--brand-soft` | — |
| Brand pale (rail active) | `--brand-pale` | — |
| Info blue (autosave saving dot) | `--info`, `--info-bg` | — |
| Success green (saved check) | `--success`, `--success-bg` | — |
| Danger (autosave error) | `--danger`, `--danger-bg` | — |
| Text scale | `--text`, `--text-soft`, `--text-muted`, `--text-faint` | — |
| Surface | `--surface`, `--surface-2` | — |
| Border | `--border` | — |
| Radius | `--r-1`, `--r-2` | — |
| Spacing | `--sp-2`, `--sp-3`, `--sp-4`, `--sp-5`, `--sp-7` | — |
| Mono font (CodeChip, VersionBadge) | `--font-mono` | — |
| Sans font | `--font-sans` | — |

- [x] All values mapped
- [x] Missing tokens added (none needed)
- [x] CSS Module tokens-only
- [x] User approved screenshot diff (manual smoke validated 2026-05-06)

---

## Phase 3c — State wiring (advisory)

- [x] Query hooks wired (existing useDocumentSession / useDocumentAutosave / useDocumentComments preserved)
- [x] Error UX wired (toasts via `sonner`; `resolveErrorMessage` for finalize errors)
- [x] Disabled CTAs (`Submeter para revisão` disabled when `!isEditable`)
- [x] Four states rendered (loading buffer / writer / readonly / lost)
- [x] Semantic HTML check (`<aside>` rail, `<main>` canvas, `aria-label="Voltar"`, `role="alert"` on errors)
- [ ] Sidebar metadata MOCKs replaced with real API — see `wiki/backlog/editor.md`

---

## Phase 4 — Verify (HARD GATE)

- [x] tsc green (zero errors in `features/documents`, `features/templates`, `features/shared/components/editor-chrome`, `packages/editor-ui`; pre-existing errors in audit/notifications/operations/registry/auth/shell unchanged)
- [x] vitest — no new regressions vs baseline (9 fails / 14 errors all jsdom-env / pre-existing)
- [x] Manual smoke (validated 2026-05-06: rail back button no collision, Revisões/Export gone, re2 chip gone, canvas centered, autosave + finalize work)
- [x] Screenshot diff approved (user validated)

---

## Phase 5 — Document (advisory)

- [x] `wiki/modules/editor-chrome.md` created (shared primitive)
- [x] `wiki/backlog/editor.md` created (deferred Metadados fields, Revisões, Aprovadores)
- [x] `wiki/modules/documents.md` updated for new layout (rail + EditorChrome; CheckpointsDialog + ExportMenuButton noted as unmounted)
- [x] `wiki/modules/editor-ui-eigenpal.md` updated for `templatePlugin` mode gating + three-value `EditorMode` type + cross-links to backlog/editor.md and editor-chrome.md
- [x] `wiki-curator` dispatched (2026-05-06)
- [ ] PR references worksheet

## 2026-05-17 Addendum - Review Comments Lifecycle

The original 2026-05-06 worksheet predates the approved review-comments lifecycle. Comments are now active review feedback: they survive rejection back to draft and stay out of clean released/PDF output. See `docs/superpowers/specs/2026-05-17-editor-review-comments-lifecycle-design.md` and `wiki/backlog/editor.md` Integration Audit (2026-05-17).

Current gate status for Plan 12 Task 1:

- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents` -> PASS
- `scripts/check-module-contract-sync.ps1 -Module documents` -> PASS (post-merge gate alignment to `documents.ts`)

Per stop rules, implementation can proceed only within the classified audit boundary in `wiki/backlog/editor.md` Integration Audit (2026-05-17).
