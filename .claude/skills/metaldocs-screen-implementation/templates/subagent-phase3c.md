# Subagent prompt — Phase 3c State wiring

You are a subagent dispatched in a fresh git worktree to perform Phase 3c (State wiring) of the MetalDocs screen-implementation workflow. Phase 3a + 3b produced the static page that visually matches the design. Your job is to wire data, state, error UX, and interactions.

## Inputs

- Worksheet path: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- Page file: `frontend/apps/web/src/features/<DOMAIN>/pages/<PAGENAME>.tsx`
- Worksheet §1.5 — state design
- Worksheet §1.6 — backend contract

## Required reading

- `wiki/concepts/error-ux.md` — `apiFetch`, `ApiError`, `resolveErrorMessage`, auth-bus
- `wiki/architecture/frontend-structure.md` — query hook layout, lazy `useState`, debounced inputs

## Steps

1. **Server state.** Create or reuse query hooks under `frontend/apps/web/src/features/<DOMAIN>/queries/`. Each hook follows the pattern:
   ```ts
   export function useXxxQuery(params: XxxParams) {
     return useQuery({
       queryKey: ['xxx', params],
       queryFn: () => xxxApi(params),
       placeholderData: keepPreviousData,
       staleTime: 30_000,
     });
   }
   ```
   Replace literal placeholder text in the page TSX with values from `query.data`.

2. **Loading / empty / error / success.** Render all four:
   - `query.isLoading` → loading skeleton.
   - `query.isError` → error message via `resolveErrorMessage(err.code, err.message)` inside an element with `role="alert"`. Use `instanceof ApiError`.
   - `query.data` empty → empty state copy from worksheet §0.1 or NOTES.md.
   - `query.data` present → success render.

3. **Local state.** `useState`/`useReducer` per worksheet §1.5.

4. **Persisted state.** Lazy initializer ONLY:
   ```ts
   const [pageSize, setPageSize] = useState<PageSize>(() => readStoredPageSize());
   ```
   Never `useState(readStoredPageSize())` — that fires every render.

5. **Debounced inputs.** `import { useDebouncedValue } from '@/lib/hooks/useDebouncedValue'`. Wire to query params per §1.5.

6. **Disabled CTAs.** Anything from §0 marked Defer:
   ```tsx
   <button type="button" disabled aria-disabled="true" title="Em breve">
     Exportar
   </button>
   ```
   Pair with worksheet `wiki/backlog/<screen>.md` row.

7. **Mock data with TODO trail.** Anything from §1.6 marked "needed" (no backend yet):
   ```tsx
   // TODO(<SLUG>:<feature>): replace with real query when GET <endpoint> ships.
   // Shape needed: <ShapeFromWorksheet>
   // Backlog: wiki/backlog/<screen>.md (row N)
   const MOCK = [...];
   ```

8. **Semantic HTML check.** Walk the JSX:
   - No `<button>` inside `<button>`.
   - Non-button rows that are click targets: `<div role="button" tabIndex={0} onClick={handle} onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') handle(); }}>`. Add `:focus-visible { outline: 2px solid var(--brand); outline-offset: -2px; }` to the row class.

9. Update worksheet §3c with `[x]` items.

## Output

Single commit: `feat(<DOMAIN>): wire <SLUG> state and error UX`. Report:
- Query hooks created/used (paths)
- States rendered (loading/empty/error/success — confirm all four)
- Persisted keys + lazy-init confirmed
- Debounced inputs (list)
- Disabled CTAs (list with backlog refs)
- Mock data blocks (list with TODO + backlog refs)
- Semantic HTML check pass/fail

## Verify before reporting done

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Both green. If red, fix before reporting done.

## Hard rules

- Never use raw `alert()`. Errors flow through `ApiError` + `resolveErrorMessage` + `role="alert"` only.
- Never `useState(readStored())` — must be `useState(() => readStored())`.
- Never wire a CTA without a corresponding backlog row when the backend is missing.
- Never invent endpoint shapes. If §1.6 says "needed" and shape is unspecified, STOP and ask main agent.
