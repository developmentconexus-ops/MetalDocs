# Phase 2 — Data Flow: Autosave (write path)

> Operation: DOCX edit → autosave → parent callback → upload.
> Source: `packages/editor-ui/src/MetalDocsEditor.tsx:30-47`

## Trace

1. **Eigenpal `DocxEditor`** fires `onChange` on every editor mutation.
2. **`handleChange()`** (MetalDocsEditor.tsx:30):
   - early-return if `props.mode === 'readonly'` (line 31)
   - early-return if no `onAutoSaveRef.current` (line 32)
   - clear previous `timerRef` (line 33)
   - schedule `setTimeout(..., AUTOSAVE_DEBOUNCE_MS=1500)` (line 34)
3. **Inside debounced callback:**
   - guard: return if `inFlightRef.current` (concurrent save protection) — line 35
   - guard: return if `inner.current` (DocxEditorRef) is null — line 36
   - read latest cb from ref (line 37–38)
   - set `inFlightRef = true`, await `inner.current.save()` (returns `Uint8Array | null`) — lines 40-41
   - if buffer non-null, `await cb(buf)` — parent receives bytes (line 42)
   - `finally`: `inFlightRef = false` (line 44)
4. **Parent (e.g. `DocumentEditorPage.tsx:252`)** owns:
   - upload to API/S3
   - failure handling, retry, 409/etag conflicts
   - surfacing save state via `AutosaveStatus` in `EditorChrome` right slot

## State machine

None. The adapter has no save-state of its own — only `inFlightRef` (boolean) for concurrent-save guard. Status pill is rendered by the parent.

## Token semantics on this path

DOCX bytes serialized by eigenpal contain **literal `{name}` tokens** (writer mode). No substitution. Server freeze pipeline performs substitution. Confirmed: `applyVariables` is never called from `MetalDocsEditor.tsx`. See `concepts/placeholders.md`.

## Failure modes (adapter-side)

| Mode | Behavior |
|---|---|
| `inner.current === null` | save skipped silently; next change schedules another save |
| `save()` returns `null` | cb not invoked; no upload attempt |
| `cb(buf)` throws | error propagates out of debounced callback (uncaught — relies on parent try/catch in `onAutoSave`) |
| Concurrent change during in-flight save | new timer scheduled; new save will run on next `handleChange` tick after `inFlightRef` clears |

## Cleanup

Component unmount → `useEffect` cleanup clears `timerRef` (MetalDocsEditor.tsx:26-28). Any in-flight `cb(buf)` is not cancelled (caller may resolve after unmount; safe because caller owns state).

## Edge: `mode === 'document-edit'` switch to `readonly` mid-session

`handleChange` checks `props.mode` each call. If parent flips mode to `readonly`, scheduled timer still runs but the guard at line 31 short-circuits later change events. The already-scheduled timer also resolves and may save once more after the flip (timer captured cb ref, not mode). **Minor latent race** — not severe because writer sessions are server-gated by `isEditable` before mounting the editor.
