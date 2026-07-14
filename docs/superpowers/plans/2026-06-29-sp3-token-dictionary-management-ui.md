# SP-3 Token Dictionary Management UI — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a capability-gated CRUD screen at `templates/tokens` for managing the tenant token dictionary, consuming the existing SP-1 backend contract.

**Architecture:** New feature slice `frontend/apps/web/src/features/tokens/` following the canonical layout (`wiki/architecture/frontend-structure.md`) and the `features/approval` / `features/templates` exemplars: generated `lib/api-types` aliased per feature, `lib/api` `apiFetch` transport, central `QK.tokens.*` query keys, TanStack Query hooks, CSS-module components, `useHasCapability` write-gating. No backend change. Spec: `docs/superpowers/specs/2026-06-29-sp3-token-dictionary-management-ui-design.md`.

**Tech Stack:** React + TypeScript, TanStack Query v5, `openapi-typescript` generated types, `@metaldocs/shared-tokens` (grammar), `sonner` toasts, Vitest + Testing Library, CSS modules.

---

## File structure

| File | Responsibility |
|---|---|
| `frontend/apps/web/src/lib/api-types/index.d.ts` | regenerated — adds token dictionary schemas |
| `frontend/apps/web/src/lib/queryKeys.ts` | add `QK.tokens` namespace |
| `frontend/apps/web/src/features/templates/index.ts` | publish catalog query/fetch for cross-feature reuse |
| `features/tokens/api/tokensTypes.ts` | aliases of generated token DTOs |
| `features/tokens/api/tokens.ts` | HTTP calls (list/create/update/delete) |
| `features/tokens/validation.ts` | pure form validation (grammar + reserved + collision + lengths) |
| `features/tokens/queries/useTokensQuery.ts` | list query |
| `features/tokens/queries/useTokenMutations.ts` | create/update/delete mutations |
| `features/tokens/components/TokenList.tsx` (+ `.module.css`, `.test.tsx`) | table of entries + row actions |
| `features/tokens/components/TokenEditDialog.tsx` (+ `.module.css`, `.test.tsx`) | create/edit form |
| `features/tokens/pages/TokensRoutePage.tsx` | route entry; composes list + dialog |
| `features/tokens/routes.tsx` | `tokenRoutes` RouteObject[] |
| `features/tokens/index.ts` | feature barrel |
| `frontend/apps/web/src/app/AppRouter.tsx` | register `tokenRoutes` |
| `features/templates/TemplatesListPage.tsx` + `pages/TemplatesListRoutePage.tsx` | "Token Dictionary" entry-point button (gated) |
| `wiki/decisions/0049-...md`, `wiki/modules/tokens.md` | docs sync (§ADR gate) |

---

## Task 1: Regenerate API types + central query keys + publish catalog

**Files:**
- Modify: `frontend/apps/web/src/lib/api-types/index.d.ts` (regenerated)
- Modify: `frontend/apps/web/src/lib/queryKeys.ts`
- Modify: `frontend/apps/web/src/features/templates/index.ts`

- [ ] **Step 1: Regenerate the generated API types**

Run:
```bash
cd frontend/apps/web && npm run gen:api
```
Expected: `src/lib/api-types/index.d.ts` rewritten; `git diff` shows added `TokenDictionaryEntry`, `CreateTokenDictionaryEntryRequest`, `UpdateTokenDictionaryEntryRequest`, `ListTokenDictionaryEntriesResponse` and the `/tokens` + `/tokens/{id}` path entries, and **no unrelated drift**.

- [ ] **Step 2: Verify the token schemas are present**

Run:
```bash
grep -n "TokenDictionaryEntry" frontend/apps/web/src/lib/api-types/index.d.ts
```
Expected: matches in `components["schemas"]`.

- [ ] **Step 3: Add the `tokens` query-key namespace**

In `frontend/apps/web/src/lib/queryKeys.ts`, inside the `QK` object (add a sibling to `taxonomy`):

```ts
  tokens: {
    all: ['tokens'] as const,
    list: () => ['tokens', 'list'] as const,
  },
```

- [ ] **Step 4: Publish the placeholder-catalog query for cross-feature reuse**

In `frontend/apps/web/src/features/templates/index.ts`, append:

```ts
export { fetchPlaceholderCatalog, type PlaceholderCatalogEntry } from './api/catalog';
export { usePlaceholderCatalogQuery } from './queries/usePlaceholderCatalogQuery';
```

- [ ] **Step 5: Typecheck**

Run:
```bash
cd frontend/apps/web && npm run typecheck
```
Expected: PASS (no errors).

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/lib/api-types/index.d.ts frontend/apps/web/src/lib/queryKeys.ts frontend/apps/web/src/features/templates/index.ts
git commit -m "feat(tokens): regen api-types, add QK.tokens, publish catalog query"
```

---

## Task 2: API client + type aliases

**Files:**
- Create: `frontend/apps/web/src/features/tokens/api/tokensTypes.ts`
- Create: `frontend/apps/web/src/features/tokens/api/tokens.ts`
- Test: `frontend/apps/web/src/features/tokens/api/tokens.test.ts`

- [ ] **Step 1: Write the type aliases**

Create `features/tokens/api/tokensTypes.ts`:

```ts
import type { components } from '../../../lib/api-types';

export type TokenDictionaryEntry = components['schemas']['TokenDictionaryEntry'];
export type CreateTokenDictionaryEntryRequest =
  components['schemas']['CreateTokenDictionaryEntryRequest'];
export type UpdateTokenDictionaryEntryRequest =
  components['schemas']['UpdateTokenDictionaryEntryRequest'];
export type ListTokenDictionaryEntriesResponse =
  components['schemas']['ListTokenDictionaryEntriesResponse'];
```

- [ ] **Step 2: Write the failing API test**

Create `features/tokens/api/tokens.test.ts`:

```ts
import { afterEach, describe, expect, it, vi } from 'vitest';
import * as apiModule from '../../../lib/api';
import { listTokens, createToken, updateToken, deleteToken } from './tokens';

afterEach(() => vi.restoreAllMocks());

describe('tokens api', () => {
  it('listTokens GETs /api/v1/tokens and returns items', async () => {
    const spy = vi
      .spyOn(apiModule, 'apiFetch')
      .mockResolvedValue({ items: [{ id: '1', name: 'slogan', value: 'v', label: 'L', created_at: 'x', updated_at: 'y' }] } as never);
    const items = await listTokens();
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens', undefined);
    expect(items).toHaveLength(1);
  });

  it('createToken POSTs the body', async () => {
    const spy = vi.spyOn(apiModule, 'apiFetch').mockResolvedValue({ id: '1' } as never);
    await createToken({ name: 'slogan', value: 'v', label: 'L' });
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens', {
      method: 'POST',
      body: JSON.stringify({ name: 'slogan', value: 'v', label: 'L' }),
    });
  });

  it('updateToken PUTs to /api/v1/tokens/{id}', async () => {
    const spy = vi.spyOn(apiModule, 'apiFetch').mockResolvedValue({ id: '1' } as never);
    await updateToken('1', { name: 'slogan', value: 'v2', label: 'L' });
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens/1', {
      method: 'PUT',
      body: JSON.stringify({ name: 'slogan', value: 'v2', label: 'L' }),
    });
  });

  it('deleteToken DELETEs /api/v1/tokens/{id}', async () => {
    const spy = vi.spyOn(apiModule, 'apiFetch').mockResolvedValue(undefined as never);
    await deleteToken('1');
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens/1', { method: 'DELETE' });
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/api/tokens.test.ts`
Expected: FAIL — `tokens.ts` does not exist.

- [ ] **Step 3: Implement the API client**

Create `features/tokens/api/tokens.ts`:

```ts
import { apiFetch } from '../../../lib/api';
import type {
  CreateTokenDictionaryEntryRequest,
  ListTokenDictionaryEntriesResponse,
  TokenDictionaryEntry,
  UpdateTokenDictionaryEntryRequest,
} from './tokensTypes';

const BASE = '/api/v1/tokens';

export async function listTokens(): Promise<TokenDictionaryEntry[]> {
  const body = await apiFetch<ListTokenDictionaryEntriesResponse>(BASE);
  return body.items;
}

export async function createToken(
  req: CreateTokenDictionaryEntryRequest,
): Promise<TokenDictionaryEntry> {
  return apiFetch<TokenDictionaryEntry>(BASE, {
    method: 'POST',
    body: JSON.stringify(req),
  });
}

export async function updateToken(
  id: string,
  req: UpdateTokenDictionaryEntryRequest,
): Promise<TokenDictionaryEntry> {
  return apiFetch<TokenDictionaryEntry>(`${BASE}/${id}`, {
    method: 'PUT',
    body: JSON.stringify(req),
  });
}

export async function deleteToken(id: string): Promise<void> {
  await apiFetch<void>(`${BASE}/${id}`, { method: 'DELETE' });
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/api/tokens.test.ts`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/tokens/api
git commit -m "feat(tokens): api client + generated type aliases"
```

---

## Task 3: Form validation module

**Files:**
- Create: `frontend/apps/web/src/features/tokens/validation.ts`
- Test: `frontend/apps/web/src/features/tokens/validation.test.ts`

- [ ] **Step 1: Write the failing test**

Create `features/tokens/validation.test.ts`:

```ts
import { describe, expect, it } from 'vitest';
import { validateName, validateEntry } from './validation';

const COMPUTED = ['author', 'doc_code', 'effective_date'];

describe('validateName', () => {
  it('accepts a valid snake_case name', () => {
    expect(validateName('company_slogan', COMPUTED)).toBeNull();
  });
  it('rejects empty', () => {
    expect(validateName('', COMPUTED)).toBe('Nome obrigatório.');
  });
  it('rejects > 64 chars', () => {
    expect(validateName('a'.repeat(65), COMPUTED)).toBe('Nome deve ter no máximo 64 caracteres.');
  });
  it('rejects invalid grammar (leading digit)', () => {
    expect(validateName('1slogan', COMPUTED)).toBe('Nome inválido: use letras, números e _ , começando por letra ou _.');
  });
  it('rejects a reserved ident', () => {
    expect(validateName('constructor', COMPUTED)).toBe('Nome reservado pelo sistema.');
  });
  it('rejects a computed-catalog collision', () => {
    expect(validateName('author', COMPUTED)).toBe('Nome já é um token do sistema (catálogo computado).');
  });
});

describe('validateEntry', () => {
  it('returns no errors for a valid entry', () => {
    expect(
      validateEntry({ name: 'company_slogan', value: 'Qualidade', label: 'Slogan', description: '' }, COMPUTED),
    ).toEqual({});
  });
  it('flags required value and label', () => {
    expect(
      validateEntry({ name: 'company_slogan', value: '', label: '', description: '' }, COMPUTED),
    ).toEqual({ value: 'Valor obrigatório.', label: 'Rótulo obrigatório.' });
  });
  it('flags value over 4096 and description over 1024', () => {
    const errs = validateEntry(
      { name: 'company_slogan', value: 'a'.repeat(4097), label: 'L', description: 'b'.repeat(1025) },
      COMPUTED,
    );
    expect(errs.value).toBe('Valor deve ter no máximo 4096 caracteres.');
    expect(errs.description).toBe('Descrição deve ter no máximo 1024 caracteres.');
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/validation.test.ts`
Expected: FAIL — `validation.ts` does not exist.

- [ ] **Step 3: Implement validation**

Create `features/tokens/validation.ts`:

```ts
import { isReservedIdent, isValidIdent } from '@metaldocs/shared-tokens';

export interface TokenFormValues {
  name: string;
  value: string;
  label: string;
  description: string;
}

export type TokenFieldErrors = Partial<Record<keyof TokenFormValues, string>>;

export function validateName(name: string, computedKeys: string[]): string | null {
  if (name.length === 0) return 'Nome obrigatório.';
  if (name.length > 64) return 'Nome deve ter no máximo 64 caracteres.';
  if (!isValidIdent(name)) return 'Nome inválido: use letras, números e _ , começando por letra ou _.';
  if (isReservedIdent(name)) return 'Nome reservado pelo sistema.';
  if (computedKeys.includes(name)) return 'Nome já é um token do sistema (catálogo computado).';
  return null;
}

export function validateEntry(values: TokenFormValues, computedKeys: string[]): TokenFieldErrors {
  const errors: TokenFieldErrors = {};

  const nameErr = validateName(values.name, computedKeys);
  if (nameErr) errors.name = nameErr;

  if (values.value.length === 0) errors.value = 'Valor obrigatório.';
  else if (values.value.length > 4096) errors.value = 'Valor deve ter no máximo 4096 caracteres.';

  if (values.label.length === 0) errors.label = 'Rótulo obrigatório.';
  else if (values.label.length > 256) errors.label = 'Rótulo deve ter no máximo 256 caracteres.';

  if (values.description.length > 1024) errors.description = 'Descrição deve ter no máximo 1024 caracteres.';

  return errors;
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/validation.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/tokens/validation.ts frontend/apps/web/src/features/tokens/validation.test.ts
git commit -m "feat(tokens): form validation (grammar + reserved + collision + lengths)"
```

---

## Task 4: Query + mutation hooks

**Files:**
- Create: `frontend/apps/web/src/features/tokens/queries/useTokensQuery.ts`
- Create: `frontend/apps/web/src/features/tokens/queries/useTokenMutations.ts`
- Test: `frontend/apps/web/src/features/tokens/queries/useTokenMutations.test.tsx`

- [ ] **Step 1: Implement the list query**

Create `features/tokens/queries/useTokensQuery.ts`:

```ts
import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { listTokens } from '../api/tokens';

const STALE_ONE_MINUTE = 60 * 1000;

export function useTokensQuery() {
  return useQuery({
    queryKey: QK.tokens.list(),
    queryFn: listTokens,
    staleTime: STALE_ONE_MINUTE,
  });
}
```

- [ ] **Step 2: Implement the mutations**

Create `features/tokens/queries/useTokenMutations.ts`:

```ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { QK } from '../../../lib/queryKeys';
import { resolveErrorMessage } from '../../../lib/api';
import { createToken, deleteToken, updateToken } from '../api/tokens';
import type {
  CreateTokenDictionaryEntryRequest,
  UpdateTokenDictionaryEntryRequest,
} from '../api/tokensTypes';

export function useTokenMutations() {
  const qc = useQueryClient();
  const invalidate = () => qc.invalidateQueries({ queryKey: QK.tokens.list() });

  const create = useMutation({
    mutationFn: (req: CreateTokenDictionaryEntryRequest) => createToken(req),
    onSuccess: () => {
      toast.success('Token criado.');
      void invalidate();
    },
    onError: (err) => toast.error(resolveErrorMessage(err)),
  });

  const update = useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTokenDictionaryEntryRequest }) =>
      updateToken(id, req),
    onSuccess: () => {
      toast.success('Token atualizado.');
      void invalidate();
    },
    onError: (err) => toast.error(resolveErrorMessage(err)),
  });

  const remove = useMutation({
    mutationFn: (id: string) => deleteToken(id),
    onSuccess: () => {
      toast.success('Token removido.');
      void invalidate();
    },
    onError: (err) => toast.error(resolveErrorMessage(err)),
  });

  return { create, update, remove };
}
```

- [ ] **Step 3: Write the mutation test**

Create `features/tokens/queries/useTokenMutations.test.tsx`:

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import * as api from '../api/tokens';
import { useTokenMutations } from './useTokenMutations';

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
afterEach(() => vi.restoreAllMocks());

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('useTokenMutations', () => {
  it('create calls the api and resolves', async () => {
    const spy = vi.spyOn(api, 'createToken').mockResolvedValue({ id: '1' } as never);
    const { result } = renderHook(() => useTokenMutations(), { wrapper });
    result.current.create.mutate({ name: 'slogan', value: 'v', label: 'L' });
    await waitFor(() => expect(result.current.create.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledOnce();
  });

  it('remove calls deleteToken', async () => {
    const spy = vi.spyOn(api, 'deleteToken').mockResolvedValue(undefined as never);
    const { result } = renderHook(() => useTokenMutations(), { wrapper });
    result.current.remove.mutate('1');
    await waitFor(() => expect(result.current.remove.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith('1');
  });
});
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/queries/useTokenMutations.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/tokens/queries
git commit -m "feat(tokens): list query + create/update/delete mutations"
```

---

## Task 5: TokenEditDialog component

**Files:**
- Create: `frontend/apps/web/src/features/tokens/components/TokenEditDialog.tsx`
- Create: `frontend/apps/web/src/features/tokens/components/TokenEditDialog.module.css`
- Test: `frontend/apps/web/src/features/tokens/components/TokenEditDialog.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `features/tokens/components/TokenEditDialog.test.tsx`:

```tsx
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TokenEditDialog } from './TokenEditDialog';

const COMPUTED = ['author', 'doc_code'];

describe('TokenEditDialog', () => {
  it('blocks submit and shows a grammar error for an invalid name', () => {
    const onSubmit = vi.fn();
    render(
      <TokenEditDialog mode="create" computedKeys={COMPUTED} submitting={false} onSubmit={onSubmit} onClose={vi.fn()} />,
    );
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: '1bad' } });
    fireEvent.change(screen.getByLabelText('Valor'), { target: { value: 'v' } });
    fireEvent.change(screen.getByLabelText('Rótulo'), { target: { value: 'L' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByText(/Nome inválido/)).toBeInTheDocument();
  });

  it('blocks submit on a computed-catalog collision', () => {
    const onSubmit = vi.fn();
    render(
      <TokenEditDialog mode="create" computedKeys={COMPUTED} submitting={false} onSubmit={onSubmit} onClose={vi.fn()} />,
    );
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: 'author' } });
    fireEvent.change(screen.getByLabelText('Valor'), { target: { value: 'v' } });
    fireEvent.change(screen.getByLabelText('Rótulo'), { target: { value: 'L' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByText(/token do sistema/)).toBeInTheDocument();
  });

  it('submits a valid entry', () => {
    const onSubmit = vi.fn();
    render(
      <TokenEditDialog mode="create" computedKeys={COMPUTED} submitting={false} onSubmit={onSubmit} onClose={vi.fn()} />,
    );
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: 'company_slogan' } });
    fireEvent.change(screen.getByLabelText('Valor'), { target: { value: 'Qualidade' } });
    fireEvent.change(screen.getByLabelText('Rótulo'), { target: { value: 'Slogan' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(onSubmit).toHaveBeenCalledWith({ name: 'company_slogan', value: 'Qualidade', label: 'Slogan', description: '' });
  });

  it('disables the name field in edit mode', () => {
    render(
      <TokenEditDialog
        mode="edit"
        computedKeys={COMPUTED}
        submitting={false}
        initial={{ name: 'company_slogan', value: 'v', label: 'L', description: '' }}
        onSubmit={vi.fn()}
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByLabelText('Nome')).toBeDisabled();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/components/TokenEditDialog.test.tsx`
Expected: FAIL — component missing.

- [ ] **Step 3: Implement the styles**

Create `features/tokens/components/TokenEditDialog.module.css`:

```css
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}
.dialog {
  background: var(--bg);
  border-radius: 8px;
  padding: 24px;
  width: 480px;
  max-width: 92vw;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}
.title { margin: 0 0 16px; font-size: 18px; font-weight: 600; }
.field { display: flex; flex-direction: column; gap: 4px; margin-bottom: 14px; }
.field label { font-size: 13px; font-weight: 500; }
.field input, .field textarea {
  border: 1px solid var(--border, #d0d0d0);
  border-radius: 6px;
  padding: 8px 10px;
  font: inherit;
}
.error { color: var(--danger, #c0392b); font-size: 12px; }
.actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }
.primary {
  background: var(--brand-600, #7a1f3d);
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
}
.secondary {
  background: transparent;
  border: 1px solid var(--border, #d0d0d0);
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
}
```

- [ ] **Step 4: Implement the component**

Create `features/tokens/components/TokenEditDialog.tsx`:

```tsx
import { useState } from 'react';
import { validateEntry, type TokenFieldErrors, type TokenFormValues } from '../validation';
import styles from './TokenEditDialog.module.css';

export interface TokenEditDialogProps {
  mode: 'create' | 'edit';
  computedKeys: string[];
  submitting: boolean;
  initial?: TokenFormValues;
  onSubmit: (values: TokenFormValues) => void;
  onClose: () => void;
}

const EMPTY: TokenFormValues = { name: '', value: '', label: '', description: '' };

export function TokenEditDialog(props: TokenEditDialogProps) {
  const { mode, computedKeys, submitting, initial, onSubmit, onClose } = props;
  const [values, setValues] = useState<TokenFormValues>(initial ?? EMPTY);
  const [errors, setErrors] = useState<TokenFieldErrors>({});

  function set<K extends keyof TokenFormValues>(key: K, v: string) {
    setValues((prev) => ({ ...prev, [key]: v }));
  }

  function handleSubmit() {
    const errs = validateEntry(values, computedKeys);
    setErrors(errs);
    if (Object.keys(errs).length === 0) onSubmit(values);
  }

  return (
    <div className={styles.overlay} role="dialog" aria-modal="true">
      <div className={styles.dialog}>
        <h2 className={styles.title}>{mode === 'create' ? 'Novo token' : 'Editar token'}</h2>

        <div className={styles.field}>
          <label htmlFor="token-name">Nome</label>
          <input
            id="token-name"
            value={values.name}
            disabled={mode === 'edit'}
            onChange={(e) => set('name', e.target.value)}
          />
          {errors.name && <span className={styles.error}>{errors.name}</span>}
        </div>

        <div className={styles.field}>
          <label htmlFor="token-value">Valor</label>
          <textarea id="token-value" rows={3} value={values.value} onChange={(e) => set('value', e.target.value)} />
          {errors.value && <span className={styles.error}>{errors.value}</span>}
        </div>

        <div className={styles.field}>
          <label htmlFor="token-label">Rótulo</label>
          <input id="token-label" value={values.label} onChange={(e) => set('label', e.target.value)} />
          {errors.label && <span className={styles.error}>{errors.label}</span>}
        </div>

        <div className={styles.field}>
          <label htmlFor="token-description">Descrição</label>
          <textarea
            id="token-description"
            rows={2}
            value={values.description}
            onChange={(e) => set('description', e.target.value)}
          />
          {errors.description && <span className={styles.error}>{errors.description}</span>}
        </div>

        <div className={styles.actions}>
          <button type="button" className={styles.secondary} onClick={onClose}>Cancelar</button>
          <button type="button" className={styles.primary} disabled={submitting} onClick={handleSubmit}>Salvar</button>
        </div>
      </div>
    </div>
  );
}
```

> Note: `<label htmlFor>` + matching `id` is what makes `screen.getByLabelText('Nome')` resolve. `description` is optional and defaults to `''`.

- [ ] **Step 5: Run to verify it passes**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/components/TokenEditDialog.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/tokens/components/TokenEditDialog.tsx frontend/apps/web/src/features/tokens/components/TokenEditDialog.module.css frontend/apps/web/src/features/tokens/components/TokenEditDialog.test.tsx
git commit -m "feat(tokens): TokenEditDialog with client-side validation"
```

---

## Task 6: TokenList component

**Files:**
- Create: `frontend/apps/web/src/features/tokens/components/TokenList.tsx`
- Create: `frontend/apps/web/src/features/tokens/components/TokenList.module.css`
- Test: `frontend/apps/web/src/features/tokens/components/TokenList.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `features/tokens/components/TokenList.test.tsx`:

```tsx
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TokenList } from './TokenList';
import type { TokenDictionaryEntry } from '../api/tokensTypes';

const ENTRY: TokenDictionaryEntry = {
  id: '1', name: 'company_slogan', value: 'Qualidade desde 1990',
  label: 'Slogan', description: 'Slogan institucional',
  created_at: '2026-06-01T00:00:00Z', updated_at: '2026-06-01T00:00:00Z',
};

describe('TokenList', () => {
  it('renders entry rows', () => {
    render(<TokenList entries={[ENTRY]} canManage={true} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('company_slogan')).toBeInTheDocument();
    expect(screen.getByText('Slogan')).toBeInTheDocument();
  });

  it('shows edit/delete only when canManage', () => {
    const { rerender } = render(<TokenList entries={[ENTRY]} canManage={false} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.queryByRole('button', { name: 'Editar' })).not.toBeInTheDocument();
    rerender(<TokenList entries={[ENTRY]} canManage={true} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'Editar' })).toBeInTheDocument();
  });

  it('fires onEdit with the entry', () => {
    const onEdit = vi.fn();
    render(<TokenList entries={[ENTRY]} canManage={true} onEdit={onEdit} onDelete={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: 'Editar' }));
    expect(onEdit).toHaveBeenCalledWith(ENTRY);
  });

  it('renders an empty state', () => {
    render(<TokenList entries={[]} canManage={true} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('Nenhum token cadastrado.')).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/components/TokenList.test.tsx`
Expected: FAIL — component missing.

- [ ] **Step 3: Implement the styles**

Create `features/tokens/components/TokenList.module.css`:

```css
.table { width: 100%; border-collapse: collapse; }
.table th, .table td {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border, #ececec);
  font-size: 14px;
  vertical-align: top;
}
.table th { font-weight: 600; color: var(--muted, #666); font-size: 12px; text-transform: uppercase; }
.value { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.code { font-family: var(--mono, monospace); }
.actions { display: flex; gap: 8px; }
.link { background: none; border: none; color: var(--brand-600, #7a1f3d); cursor: pointer; padding: 0; font: inherit; }
.empty { padding: 32px; text-align: center; color: var(--muted, #666); }
```

- [ ] **Step 4: Implement the component**

Create `features/tokens/components/TokenList.tsx`:

```tsx
import type { TokenDictionaryEntry } from '../api/tokensTypes';
import styles from './TokenList.module.css';

export interface TokenListProps {
  entries: TokenDictionaryEntry[];
  canManage: boolean;
  onEdit: (entry: TokenDictionaryEntry) => void;
  onDelete: (entry: TokenDictionaryEntry) => void;
}

export function TokenList(props: TokenListProps) {
  const { entries, canManage, onEdit, onDelete } = props;

  if (entries.length === 0) {
    return <div className={styles.empty}>Nenhum token cadastrado.</div>;
  }

  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th>Nome</th>
          <th>Rótulo</th>
          <th>Valor</th>
          <th>Descrição</th>
          {canManage && <th aria-label="Ações" />}
        </tr>
      </thead>
      <tbody>
        {entries.map((e) => (
          <tr key={e.id}>
            <td className={styles.code}>{`{${e.name}}`}</td>
            <td>{e.label}</td>
            <td className={styles.value} title={e.value}>{e.value}</td>
            <td>{e.description ?? ''}</td>
            {canManage && (
              <td>
                <div className={styles.actions}>
                  <button type="button" className={styles.link} onClick={() => onEdit(e)}>Editar</button>
                  <button type="button" className={styles.link} onClick={() => onDelete(e)}>Excluir</button>
                </div>
              </td>
            )}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
```

- [ ] **Step 5: Run to verify it passes**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/components/TokenList.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/tokens/components/TokenList.tsx frontend/apps/web/src/features/tokens/components/TokenList.module.css frontend/apps/web/src/features/tokens/components/TokenList.test.tsx
git commit -m "feat(tokens): TokenList table with capability-gated row actions"
```

---

## Task 7: Route page (composition)

**Files:**
- Create: `frontend/apps/web/src/features/tokens/pages/TokensRoutePage.tsx`
- Create: `frontend/apps/web/src/features/tokens/pages/TokensRoutePage.module.css`
- Test: `frontend/apps/web/src/features/tokens/pages/TokensRoutePage.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `features/tokens/pages/TokensRoutePage.test.tsx`:

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import * as tokensApi from '../api/tokens';
import { Component as TokensRoutePage } from './TokensRoutePage';

vi.mock('../../templates', () => ({
  usePlaceholderCatalogQuery: () => ({ data: [{ key: 'author', label: 'Autor', description: '' }] }),
}));
const hasCapMock = vi.fn();
vi.mock('../../iam/hooks/useHasCapability', () => ({ useHasCapability: (c: string) => hasCapMock(c) }));

afterEach(() => vi.restoreAllMocks());

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('TokensRoutePage', () => {
  it('lists entries and shows the New button for managers', async () => {
    hasCapMock.mockReturnValue(true);
    vi.spyOn(tokensApi, 'listTokens').mockResolvedValue([
      { id: '1', name: 'company_slogan', value: 'v', label: 'Slogan', description: '', created_at: 'x', updated_at: 'y' },
    ]);
    render(<TokensRoutePage />, { wrapper });
    await waitFor(() => expect(screen.getByText('company_slogan')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'Novo token' })).toBeInTheDocument();
  });

  it('hides the New button without the manage capability', async () => {
    hasCapMock.mockReturnValue(false);
    vi.spyOn(tokensApi, 'listTokens').mockResolvedValue([]);
    render(<TokensRoutePage />, { wrapper });
    await waitFor(() => expect(screen.getByText('Nenhum token cadastrado.')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'Novo token' })).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/pages/TokensRoutePage.test.tsx`
Expected: FAIL — page missing.

- [ ] **Step 3: Implement the styles**

Create `features/tokens/pages/TokensRoutePage.module.css`:

```css
.page { padding: 24px 32px; }
.header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.title { margin: 0; font-size: 22px; font-weight: 600; }
.subtitle { margin: 4px 0 0; color: var(--muted, #666); font-size: 14px; }
.newBtn {
  background: var(--brand-600, #7a1f3d);
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
}
.state { padding: 32px; text-align: center; color: var(--muted, #666); }
```

- [ ] **Step 4: Implement the page**

Create `features/tokens/pages/TokensRoutePage.tsx`:

```tsx
import { useMemo, useState } from 'react';
import { usePlaceholderCatalogQuery } from '../../templates';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import { resolveQueryError } from '../../../lib/api/resolveQueryError';
import { TokenList } from '../components/TokenList';
import { TokenEditDialog } from '../components/TokenEditDialog';
import { useTokensQuery } from '../queries/useTokensQuery';
import { useTokenMutations } from '../queries/useTokenMutations';
import type { TokenDictionaryEntry } from '../api/tokensTypes';
import type { TokenFormValues } from '../validation';
import styles from './TokensRoutePage.module.css';

type DialogState =
  | { open: false }
  | { open: true; mode: 'create' }
  | { open: true; mode: 'edit'; entry: TokenDictionaryEntry };

export function Component() {
  const canManage = useHasCapability('token_dictionary.manage');
  const tokensQuery = useTokensQuery();
  const catalogQuery = usePlaceholderCatalogQuery();
  const { create, update, remove } = useTokenMutations();
  const [dialog, setDialog] = useState<DialogState>({ open: false });

  const computedKeys = useMemo(
    () => (catalogQuery.data ?? []).map((c) => c.key),
    [catalogQuery.data],
  );

  function handleSubmit(values: TokenFormValues) {
    const req = {
      name: values.name,
      value: values.value,
      label: values.label,
      description: values.description.length > 0 ? values.description : undefined,
    };
    if (dialog.open && dialog.mode === 'edit') {
      update.mutate({ id: dialog.entry.id, req }, { onSuccess: () => setDialog({ open: false }) });
    } else {
      create.mutate(req, { onSuccess: () => setDialog({ open: false }) });
    }
  }

  function handleDelete(entry: TokenDictionaryEntry) {
    if (window.confirm(`Excluir o token {${entry.name}}?`)) remove.mutate(entry.id);
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Dicionário de tokens</h1>
          <p className={styles.subtitle}>Constantes reutilizáveis preenchidas automaticamente nos documentos.</p>
        </div>
        {canManage && (
          <button type="button" className={styles.newBtn} onClick={() => setDialog({ open: true, mode: 'create' })}>
            Novo token
          </button>
        )}
      </div>

      {tokensQuery.isLoading && <div className={styles.state}>Carregando tokens...</div>}
      {tokensQuery.isError && (
        <div className={styles.state} role="alert">{resolveQueryError(tokensQuery.error)}</div>
      )}
      {tokensQuery.data && (
        <TokenList
          entries={tokensQuery.data}
          canManage={canManage}
          onEdit={(entry) => setDialog({ open: true, mode: 'edit', entry })}
          onDelete={handleDelete}
        />
      )}

      {dialog.open && (
        <TokenEditDialog
          mode={dialog.mode}
          computedKeys={computedKeys}
          submitting={create.isPending || update.isPending}
          initial={
            dialog.mode === 'edit'
              ? {
                  name: dialog.entry.name,
                  value: dialog.entry.value,
                  label: dialog.entry.label,
                  description: dialog.entry.description ?? '',
                }
              : undefined
          }
          onSubmit={handleSubmit}
          onClose={() => setDialog({ open: false })}
        />
      )}
    </div>
  );
}
```

> Note: the route loads this page lazily, so the exported symbol is `Component` (React Router lazy convention, matching `TemplatesListRoutePage`).

- [ ] **Step 5: Run to verify it passes**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/pages/TokensRoutePage.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/tokens/pages
git commit -m "feat(tokens): TokensRoutePage composition (list + dialog + gating)"
```

---

## Task 8: Routes, barrel, router registration, entry-point button

**Files:**
- Create: `frontend/apps/web/src/features/tokens/routes.tsx`
- Create: `frontend/apps/web/src/features/tokens/index.ts`
- Modify: `frontend/apps/web/src/app/AppRouter.tsx`
- Modify: `frontend/apps/web/src/features/templates/TemplatesListPage.tsx`
- Modify: `frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx`
- Test: `frontend/apps/web/src/features/tokens/routes.test.tsx`

- [ ] **Step 1: Create the routes**

Create `features/tokens/routes.tsx`:

```tsx
import type { RouteObject } from 'react-router-dom';

export const tokenRoutes: RouteObject[] = [
  {
    path: 'templates/tokens',
    handle: { workspaceView: 'templates', requiresCapability: 'token.view' },
    lazy: () => import('./pages/TokensRoutePage'),
  },
];
```

- [ ] **Step 2: Create the barrel**

Create `features/tokens/index.ts`:

```ts
export { tokenRoutes } from './routes';
```

- [ ] **Step 3: Write the routes test**

Create `features/tokens/routes.test.tsx`:

```tsx
import { describe, expect, it } from 'vitest';
import { tokenRoutes } from './routes';

describe('tokenRoutes', () => {
  it('gates templates/tokens on token.view', () => {
    const route = tokenRoutes.find((r) => r.path === 'templates/tokens');
    expect(route).toBeDefined();
    expect((route?.handle as { requiresCapability?: string }).requiresCapability).toBe('token.view');
    expect((route?.handle as { workspaceView?: string }).workspaceView).toBe('templates');
  });
});
```

- [ ] **Step 4: Run the routes test**

Run: `cd frontend/apps/web && npx vitest run src/features/tokens/routes.test.tsx`
Expected: PASS.

- [ ] **Step 5: Register the routes in AppRouter**

In `frontend/apps/web/src/app/AppRouter.tsx`, add the import near the other feature-route imports:

```ts
import { tokenRoutes } from '../features/tokens';
```

Then add the spread inside the `AppShell` `children` array, immediately after `...templatesRoutes,`:

```ts
          ...templatesRoutes,
          ...tokenRoutes,
```

- [ ] **Step 6: Add the gated entry-point button**

In `frontend/apps/web/src/features/templates/TemplatesListPage.tsx`:

(a) extend the props type:

```ts
export type TemplatesListPageProps = {
  onOpenTemplate: (templateId: string, versionNum: number) => void;
  onCreate: () => void;
  onOpenTokenDictionary?: () => void;
};
```

(b) add the button next to the existing "new" button (the element at line ~96 inside the header). Place it immediately before that button:

```tsx
            {props.onOpenTokenDictionary && (
              <button type="button" className={styles.newBtn} onClick={() => props.onOpenTokenDictionary?.()}>
                Dicionário de tokens
              </button>
            )}
```

In `frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx`, wire the callback (gate it on `token.view`):

```tsx
import { useNavigate } from "react-router-dom";
import { TemplatesListPage } from "../TemplatesListPage";
import { useHasCapability } from "../../iam/hooks/useHasCapability";

export function Component() {
  const navigate = useNavigate();
  const canViewTokens = useHasCapability("token.view");

  return (
    <TemplatesListPage
      onOpenTemplate={(templateId, versionNum) =>
        navigate(`/templates/${templateId}/versions/${versionNum}`)
      }
      onCreate={() => navigate("/templates/new")}
      onOpenTokenDictionary={canViewTokens ? () => navigate("/templates/tokens") : undefined}
    />
  );
}
```

- [ ] **Step 7: Typecheck + run the touched feature tests**

Run:
```bash
cd frontend/apps/web && npm run typecheck && npx vitest run src/features/tokens
```
Expected: typecheck PASS; all `features/tokens` tests PASS.

- [ ] **Step 8: Commit**

```bash
git add frontend/apps/web/src/features/tokens/routes.tsx frontend/apps/web/src/features/tokens/index.ts frontend/apps/web/src/features/tokens/routes.test.tsx frontend/apps/web/src/app/AppRouter.tsx frontend/apps/web/src/features/templates/TemplatesListPage.tsx frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx
git commit -m "feat(tokens): register templates/tokens route + gated entry point"
```

---

## Task 9: Docs sync (ADR 0049 gate + tokens module wiki)

**Files:**
- Modify: `wiki/decisions/0049-tenant-dictionary-token-substitution.md`
- Modify: `wiki/modules/tokens.md`

- [ ] **Step 1: Record the forensic-reconstruction decoupling in ADR 0049**

In `wiki/decisions/0049-tenant-dictionary-token-substitution.md`, replace the line:

```
**Named post-SP-2 owner:** unassigned (must be resolved before forensic audit features can be certified).
```

with:

```
**Named post-SP-2 owner:** Decoupled from SP-3 (operator decision 2026-06-29). SP-3 is the
dictionary management CRUD UI and does not read or reconstruct provenance, so it does not
depend on this fix. Forensic reconstruction (storing the dictionary entry name on
`source='dictionary'` pinned rows) is owned by the future forensic-audit epic and must be
resolved before any forensic-audit feature is certified.
```

- [ ] **Step 2: Add the SP-3 UI surface to the tokens module wiki**

In `wiki/modules/tokens.md`, under the section that lists SP increments / surfaces, add a row/line:

```
- **SP-3 — Token Dictionary Management UI** (`frontend/apps/web/src/features/tokens/`): capability-gated CRUD screen at `templates/tokens`, gated on `token.view` (route) and `token_dictionary.manage` (write actions). Consumes `GET/POST /tokens`, `GET/PUT/DELETE /tokens/{id}`. Reuses the templates placeholder-catalog query for the D4 collision check. Shipped 2026-06-29.
```

- [ ] **Step 3: Commit**

```bash
git add wiki/decisions/0049-tenant-dictionary-token-substitution.md wiki/modules/tokens.md
git commit -m "docs(tokens): SP-3 UI wiki surface; decouple ADR 0049 forensic gate from SP-3"
```

---

## Task 10: Full verification

**Files:** none (verification only)

- [ ] **Step 1: Typecheck**

Run: `cd frontend/apps/web && npm run typecheck`
Expected: PASS.

- [ ] **Step 2: Full frontend test suite**

Run: `cd frontend/apps/web && npx vitest run`
Expected: PASS, including all new `features/tokens` tests. (If the repo convention is `make test`, run that instead from repo root.)

- [ ] **Step 3: Production build**

Run: `cd frontend/apps/web && npm run build`
Expected: build succeeds (regenerated types + new feature compile).

- [ ] **Step 4: Manual smoke (optional, if a dev server is run)**

Sign in as a user holding `token.view` + `token_dictionary.manage`; from the templates list click "Dicionário de tokens"; create an entry `company_slogan`; verify it appears; edit its value; delete it; attempt to create `author` and confirm the collision error blocks submit.

- [ ] **Step 5: Final commit (if any verification fixups were needed)**

```bash
git add -A
git commit -m "chore(tokens): SP-3 verification fixups"
```

---

## Self-review notes (verified during planning)

- **Spec coverage:** §2 scope → Tasks 2–8; §3 placement/authz → Task 8 (route handle) + Task 7 (`useHasCapability`) + Task 8 (entry point); §4 feature tree → Tasks 2–8; §5 validation → Task 3 + Task 5; §6 data flow → Tasks 2, 4; §7 types → Task 1 (regen) + Task 2 (aliases); §8 styling → CSS modules in Tasks 5–7; §9 testing → tests in Tasks 2–8; §10 ADR gate → Task 9; §12 verification → Task 10.
- **Type consistency:** `TokenFormValues` (Task 3) used by Tasks 5 & 7; `TokenDictionaryEntry` alias (Task 2) used by Tasks 6 & 7; `QK.tokens.list()` (Task 1) used by Tasks 4; `Component` lazy export (Task 7) matched by `lazy: () => import('./pages/TokensRoutePage')` (Task 8); `usePlaceholderCatalogQuery` published (Task 1) consumed (Task 7).
- **Known assumption to confirm at execution:** the exact line of the existing "new" button in `TemplatesListPage.tsx` (~line 96) — place the new button adjacent to it; if the header markup differs, keep the same `styles.newBtn` class and gating logic.
