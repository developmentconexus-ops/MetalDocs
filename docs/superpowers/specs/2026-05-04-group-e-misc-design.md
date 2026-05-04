# Group E (sub-plan 3) — Misc UI Hardening Design

> **Status:** approved 2026-05-04
> **Scope:** Fix bugs E7, E9, E12 from `wiki/bugs/audit-2026-05-03.md` (lines 277, 279, 282). Three small consumer-facing UI fixes built on the shared HTTP/error primitives shipped in sub-plan 1.
> **Out of scope:** E5 (`document-profiles/{bundle,schema,governance}` endpoints) — promoted to a separate "E-admin" sub-plan, see Open Questions. E1/E2/E3/E4/E6/E8/E10/E11 are covered by other sub-plans or already shipped.

---

## Why This Spec Exists

Sub-plans 1 and 2 establish the foundation: shared `apiFetch`, `ApiError`, `resolveErrorMessage`, lifecycle-aware editor, registry surfaces published doc, async PDF readiness. Three remaining consumer-facing bugs ride on those primitives but were intentionally deferred to keep the prior plans focused:

- **E7** — `InboxPage.tsx:8` hardcodes `AREA_OPTIONS = ['','JUR','RH','FIN','TI','COM','ENG']`. Real area list lives in `process_areas` and is already exposed via `fetchAreas()`. Filter shows wrong codes for tenants whose areas don't match the hardcoded six.
- **E9** — `DocumentEditorPage.tsx:113` `handleRename` does optimistic `setDocumentName(name)` then `.catch` shows a generic toast — never rolls back. UI ends up displaying a name that never reached the server.
- **E12** — `RegistryCreateDialog.tsx:194` submit button gated only on `saving`. If user clicks before `currentUser` resolves, request fires with `ownerUserId=""`. Author field also displays empty string.

E5 (admin endpoints `/document-profiles/{code}/{bundle,schema,governance}` not registered on backend) was considered but is a multi-handler backend workstream — promoted to its own sub-plan to keep this one shippable in one sitting.

---

## Architecture: Defensive UI Hardening

```
┌──────────────────┐  ┌─────────────────────┐  ┌────────────────────┐
│ InboxPage (E7)   │  │ DocumentEditorPage  │  │ RegistryCreateDlg  │
│ areas from       │  │ (E9)                │  │ (E12)              │
│ fetchAreas()     │  │ rename rollback +   │  │ submit gated on    │
│                  │  │ ApiError toast      │  │ currentUser ready  │
└────────┬─────────┘  └──────────┬──────────┘  └─────────┬──────────┘
         │                       │                        │
         └────── shared primitives from sub-plan 1 ───────┘
              apiFetch · ApiError · resolveErrorMessage
```

**Backend touch:** zero.

**Frontend touch:**
- `frontend/apps/web/src/features/approval/pages/InboxPage.tsx` — fetch areas via existing taxonomy API (E7)
- `frontend/apps/web/src/features/approval/pages/InboxPage.test.tsx` — area-filter render test (E7)
- `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` — capture previous name, restore on `.catch`, surface `resolveErrorMessage` (E9)
- `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx` — rollback test (E9, append to file created in sub-plan 2)
- `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx` — derive `isAuthReady`, gate submit + author placeholder (E12)
- `frontend/apps/web/src/features/registry/RegistryCreateDialog.test.tsx` — auth-not-ready and ready cases (E12, new file)

**Dependency:** Merge order is sub-plan 1 → sub-plan 3. `resolveErrorMessage` and `ApiError` come from sub-plan 1's `lib/api/`. Sub-plan 2 is independent — sub-plan 3 can ship before or after sub-plan 2.

**Migration discipline:** Phase 1 runs all three fixes in parallel (different files, no overlap). Phase 2 verifies. Phase 3 closes audit + wiki.

---

## Per-Bug Fix Design

### E7 — Inbox area filter sourced from taxonomy

**File:** `frontend/apps/web/src/features/approval/pages/InboxPage.tsx`

**Problem:** Module-level constant `AREA_OPTIONS = ['', 'JUR', 'RH', 'FIN', 'TI', 'COM', 'ENG']` is rendered as `<option>` values. Tenants whose `process_areas` rows differ (most do) see wrong codes.

**Fix:**

1. Remove the constant.
2. Load areas via existing `fetchAreas()`:
   ```tsx
   import { fetchAreas } from '../../taxonomy/api';
   import type { ProcessArea } from '../../taxonomy/types';

   const [areas, setAreas] = useState<ProcessArea[]>([]);

   useEffect(() => {
     void fetchAreas().then(setAreas).catch(() => setAreas([]));
   }, []);
   ```
3. Render select:
   ```tsx
   <select value={areaFilter} onChange={(e) => setAreaFilter(e.target.value)}>
     <option value="">Todas as áreas</option>
     {areas.map((a) => (
       <option key={a.code} value={a.code}>{a.code} — {a.name}</option>
     ))}
   </select>
   ```

Empty `areaFilter` value continues to mean "all areas" (matches existing `area_code: areaFilter || undefined` behaviour at `InboxPage.tsx:45`).

Failure case (areas API errors): falls back to empty array → only "Todas as áreas" option visible. Acceptable degradation — user can still browse without area scoping.

**Test (`InboxPage.test.tsx`):** mounts page with mocked `fetchAreas` returning `[{code:'OPS',name:'Operações'},{code:'QA',name:'Qualidade'}]`, asserts:
- "Todas as áreas" option present with empty value
- `OPS — Operações` and `QA — Qualidade` options present
- Hardcoded `JUR/RH/FIN/TI/COM/ENG` no longer appear

---

### E9 — Rename rollback on server error

**File:** `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` (line ~112)

**Problem:** Optimistic update commits new name immediately; failure path only toasts. Stale UI state.

**Fix:**

```tsx
import { resolveErrorMessage } from '../../../lib/api/errorMessages';

const handleRename = useCallback((name: string) => {
  const prev = documentName;
  setDocumentName(name);
  void renameDocument(documentID, name).catch((err) => {
    setDocumentName(prev); // rollback
    const code = (err && typeof err === 'object' && 'code' in err) ? (err as { code?: string }).code : undefined;
    toast.error(resolveErrorMessage(code, 'Falha ao renomear documento.'));
  });
}, [documentID, documentName]);
```

`renameDocument` will throw `ApiError` (sub-plan 1) once the docs/v2 features migrate to `apiFetch`. Defensive `code` extraction handles pre-migration callsites that throw plain `Error`.

**Test (append to `DocumentEditorPage.test.tsx`):**

```tsx
it('rolls back document name on rename failure', async () => {
  const renameSpy = vi.spyOn(api, 'renameDocument').mockRejectedValueOnce(
    new ApiError('not_found', 404, 'Document not found'),
  );
  server.use(http.get('/api/v2/documents/:id', () => HttpResponse.json({
    Status: 'draft', CurrentRevisionID: 'r1', Name: 'Original', Code: 'C', RevisionVersion: 1,
  })));
  // mount, wait for initial load, programmatically invoke handleRename via test hook or by simulating editor onRename
  // assert displayed name reverts to "Original" after .catch
  // assert toast.error called with fallback string
  expect(renameSpy).toHaveBeenCalled();
});
```

(Test mounts page, waits for "Original" to render in title bar, programmatically triggers rename via the editor's `onChangeName` prop, awaits the rejected promise, then asserts displayed name === "Original".)

---

### E12 — RegistryCreateDialog submit gated on currentUser

**File:** `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx`

**Problem:** Submit button is `disabled={saving}` only. If user clicks during the brief window before `useAuthStore` populates `user`, request fires with `ownerUserId=""`. Author field displays empty string with no explanation.

**Fix:**

```tsx
const isAuthReady = !!currentUser?.userId;

// Author readonly field (replace lines ~123-128)
<input
  value={
    isAuthReady
      ? (currentUser!.displayName ?? currentUser!.userId)
      : 'Aguardando autenticação...'
  }
  readOnly
  style={{
    width: '100%', padding: '6px 8px', boxSizing: 'border-box',
    background: '#f5f5f5', color: isAuthReady ? '#666' : '#aaa', cursor: 'not-allowed',
  }}
/>

// Submit button (replace line ~194)
<button type="submit" disabled={saving || !isAuthReady} style={{ padding: '6px 14px' }}>
  {!isAuthReady ? 'Aguardando...' : saving ? 'Criando...' : 'Criar'}
</button>
```

No need to early-return or block dialog mount — the gate covers the race.

**Test (new file `RegistryCreateDialog.test.tsx`):**

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { RegistryCreateDialog } from './RegistryCreateDialog';
import { useAuthStore } from '../../store/auth.store';

vi.mock('../../store/auth.store');
vi.mock('../taxonomy/api', () => ({ fetchProfiles: () => Promise.resolve([]), fetchAreas: () => Promise.resolve([]) }));

describe('RegistryCreateDialog E12', () => {
  it('disables submit and shows placeholder while currentUser unresolved', () => {
    (useAuthStore as any).mockImplementation((sel: any) => sel({ user: null }));
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);
    expect(screen.getByRole('button', { name: /Aguardando/i })).toBeDisabled();
    expect(screen.getByDisplayValue('Aguardando autenticação...')).toBeInTheDocument();
  });

  it('enables submit when currentUser ready', () => {
    (useAuthStore as any).mockImplementation((sel: any) => sel({ user: { userId: 'u-1', displayName: 'Alice' } }));
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);
    expect(screen.getByRole('button', { name: /Criar/i })).not.toBeDisabled();
    expect(screen.getByDisplayValue('Alice')).toBeInTheDocument();
  });
});
```

---

## Rollout Plan

| Phase | Tasks | Parallelism | Model |
|---|---|---|---|
| 0 | Worktree, codex spec validate | sequential | sonnet |
| 1 | E7 (InboxPage area fetch) ‖ E9 (rename rollback) ‖ E12 (submit gate) | parallel ×3 | sonnet ‖ sonnet ‖ sonnet |
| 2 | Verify: vitest + tsc --noEmit + lint + smoke + codex audit 3/3 | sequential | sonnet → codex audit |
| 3 | Audit doc closure (E7/E9/E12 + flag E5 deferred) + wiki-curator + finishing-a-development-branch | sequential | sonnet + wiki-curator |

**Phase review after each:** opus.

Phase 1 is fully parallel — three independent files, three independent test files, no shared symbols changed.

---

## Testing Strategy

**Per-bug:** see "Per-Bug Fix Design".

**Cross-cutting:**
- `npx vitest run` — full pass
- `npx tsc --noEmit` — zero new errors
- `npm run lint` — zero new warnings
- Smoke E7: open Inbox in dev, observe area dropdown matches `process_areas` table content
- Smoke E9: rename a document with backend stub forced to 500, observe name visibly reverts
- Smoke E12: throttle network, open Create dialog before auth resolves, observe disabled state then enabled
- Codex independent audit: PASS/FAIL per bug with file:line evidence

**Coverage:** changed lines fully covered by new tests. No coverage regression on existing files.

---

## Acceptance Criteria

- [ ] E7: Inbox area filter populated from `fetchAreas()`; "Todas as áreas" option present
- [ ] E7: hardcoded `AREA_OPTIONS` constant removed
- [ ] E9: rename failure restores previous name in UI
- [ ] E9: error toast uses `resolveErrorMessage(err.code, fallback)`
- [ ] E12: submit button disabled until `currentUser?.userId` present
- [ ] E12: "Autor" field shows "Aguardando autenticação..." while loading
- [ ] All vitest pass
- [ ] `npx tsc --noEmit` passes
- [ ] No new lint warnings
- [ ] Codex audit returns 3/3 PASS
- [ ] Audit doc updated, E7/E9/E12 closed with commit SHAs
- [ ] Audit doc adds follow-up entry: "E-admin sub-plan deferred (E5 — bundle/schema/governance handlers)"

---

## Open Questions

**E5 promoted to its own sub-plan.** Decision recorded: backend has no handlers for `/document-profiles/{code}/{bundle,schema,governance}` (and ~10 sibling admin routes consumed by `useRegistryExplorer.ts`). Implementing is a multi-handler workstream including profile composition, schema versioning surface, and governance domain — too large for sub-plan 3. Tracked as future "E-admin" sub-plan; flag in audit doc closure.

---

## References

- Audit: `wiki/bugs/audit-2026-05-03.md` (E7 line 277, E9 line 279, E12 line 282; E5 line 275 deferred)
- Sub-plan 1 (error UX): `docs/superpowers/specs/2026-05-03-group-e-error-ux-design.md` — provides `apiFetch`, `ApiError`, `resolveErrorMessage`
- Sub-plan 2 (editor lifecycle): `docs/superpowers/specs/2026-05-04-group-e-editor-lifecycle-design.md` — provides `DocumentEditorPage.test.tsx` scaffold (sub-plan 3 appends to it)
- Inbox: `frontend/apps/web/src/features/approval/pages/InboxPage.tsx`
- Editor: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`
- Registry create: `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx`
- Taxonomy API: `frontend/apps/web/src/features/taxonomy/api.ts` (`fetchAreas`)
- Auth store: `frontend/apps/web/src/store/auth.store.ts`
