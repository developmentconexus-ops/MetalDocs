# Phase 2 — Data-flow trace: Autosave status rendering

**Operation:** Page-level autosave hook reports state; `<AutosaveStatus>` renders visual indicator.
**Path traced:** `DocumentEditorPage` → `useDocumentAutosave` → `AutosaveStatus`.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| Hook (state owner) | `useDocumentAutosave` | `frontend/apps/web/src/features/documents/hooks/v2/useDocumentAutosave.ts:22` |
| Local hook status type | `AutosaveStatus` (NAME COLLISION with component) | `useDocumentAutosave.ts:5` |
| Adapter | inline `autosaveState` ternary | `DocumentEditorPage.tsx:184-188` (and `TemplateEditorPage.tsx:217-221`) |
| Component | `AutosaveStatus` | `parts/AutosaveStatus.tsx:28` |
| Component state type | `AutosaveState` | `parts/AutosaveStatus.tsx:3` |

### 2. Call chain

```
1. useDocumentAutosave.ts:22       useDocumentAutosave(args)
   → setStatus('saving') | 'saved' | 'error' | 'stale' | 'session_lost' | 'dirty' | 'idle'
   → returns { status: AutosaveStatus (7-state), queue, flush }

2. DocumentEditorPage.tsx:184      const autosaveState: AutosaveState =
                                     autosave.status === 'saving' ? 'saving' :
                                     autosave.status === 'error'  ? 'error'  :
                                     autosave.status === 'saved'  ? 'saved'  :
                                     'idle';
   ↓ 7-state → 4-state collapse. 'dirty', 'stale', 'session_lost' all become 'idle'.

3. DocumentEditorPage.tsx:228      <AutosaveStatus status={autosaveState} />
   → parts/AutosaveStatus.tsx:28   function AutosaveStatus({status, labels, className})
       merges DEFAULT_LABELS (pt-BR) with optional override
       branches on status:
         'saving' → <span><dot.dotSaving aria-hidden/>{labels.saving}</span>   (line :33–40)
         'error'  → <span><dot.dotError aria-hidden/>{labels.error}</span>     (line :41–48)
         'saved'  → <span><CheckIcon/>{labels.saved}</span>                    (line :49–56)
         default  → <span><dot.dotIdle aria-hidden/>{labels.idle}</span>       (line :57–62)

4. CSS animation                 AutosaveStatus.module.css:28–31
   .dotSaving { animation: pulse 1.2s ease-in-out infinite }
   @media (prefers-reduced-motion: reduce) — animation disabled (line :49)
```

### 3. State changes

| Entity | From | To | Trigger | Capability |
|---|---|---|---|---|
| Local React `status` (hook) | `idle` | `dirty` | user edits, `queue()` called | n/a |
| Local React `status` (hook) | `dirty` | `saving` | 15s debounce fires `flush()` | n/a |
| Local React `status` (hook) | `saving` | `saved` | server `commitAutosave` 200 OK | n/a |
| Local React `status` (hook) | `saving` | `error` | server 410 / 422 / network error | n/a |
| Local React `status` (hook) | `saving` | `stale` | server 409 `stale_base` | n/a |
| Local React `status` (hook) | `saving` | `session_lost` | server 409 `session_inactive` / `session_not_holder` | n/a |
| Visible to user (`AutosaveStatus`) | (4-state only) | — | adapter ternary at `DocumentEditorPage.tsx:184` | n/a |

**Observation (fact, not prescription):** the 4-state visual enum `AutosaveState` cannot represent `stale`, `session_lost`, or `dirty`. The page-level ternary collapses all three to `idle` — which renders as the green `Salvo` (saved) dot + label, identical to a successful save. The hook also calls `onSessionLost(...)` separately at lines 68, 70 so the page is notified; whether that callback updates `AutosaveStatus` props another way is out of scope of this artifact (it does not — the ternary at `:184` is the sole feed to the component).

### 4. SQL touched

n/a in trace — autosave hook calls `presignAutosave`/`commitAutosave` REST endpoints; those touch `documents` table via the `documents` module. Not this module's surface.

### 5. Response shape

n/a — UI render, no HTTP boundary inside the chrome module.

### 6. Cross-references

- **Idempotency:** the hook computes `content_hash` (SHA-256 of buffer) but sends `pending_upload_id` to commit; idempotency story lives in documents module, not editor-chrome.
- **Pagination:** n/a.
- **Audit log emission:** n/a in this trace.
- **A11y:** `AutosaveStatus` wrapper `<span>` has no `role="status"` / `aria-live="polite"`. Dots and check SVG carry `aria-hidden="true"`. Text content is the only accessible name; screen readers do not get a live announcement when state flips saving → saved.
- **`min-width: 60px`** on `.status` (`AutosaveStatus.module.css:10`) prevents adjacent-button shift when text length changes (`Salvando…` longer than `Salvo`). Hardcoded; not token-driven.
