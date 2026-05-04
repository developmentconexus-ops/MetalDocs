# Group E (sub-plan 1) — Error UX Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix E2/E3/E4 by introducing a shared frontend HTTP client (`apiFetch`), structured `ApiError`, error-code → Portuguese-message map, and global 401 interceptor with returnTo preservation. Migrate auth-critical features (documents/v2, registry, approval, iam) to the shared layer.

**Architecture:** Three-layer error stack. `apiFetch` handles fetch + parse + 401 dispatch. `errorMessages.ts` translates known codes; falls back to backend `error.message`. Auth-bus `CustomEvent` decouples HTTP layer from React auth context. Backend untouched (codes already correct).

**Tech Stack:** React 18 + TypeScript + Vite, vitest for unit, msw for HTTP mocking, sonner for toasts, react-router-dom for navigation, Zustand for auth store.

**Spec:** `docs/superpowers/specs/2026-05-03-group-e-error-ux-design.md`

**Model selection:**
- Codex: Phase 1 (shared library + auth listener), Phase 2 (documents/v2 migration — many call sites)
- Sonnet: Phase 3 (registry, iam — mechanical migrations), Phase 4 (mutationClient refactor)
- Haiku: none (every task touches multiple files)
- Opus: phase reviews only — never coding

**Codex parallelism:**
- Phase 3: registry ‖ iam (no file overlap)
- Phase 0/1/2/4/5/6: sequential

**Wiki maintenance:** `wiki-curator` subagent dispatched in Phase 6. Creates `wiki/concepts/error-ux.md`, refreshes any `wiki/modules/frontend-*.md` stamps, updates audit doc closure.

**Caveman prompts:** subagent prompts drop articles/filler. Code/commits/security written normally per plan content.

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `frontend/apps/web/src/lib/api/client.ts` | `apiFetch<T>` — fetch + parse error envelope + 401 dispatch | Create |
| `frontend/apps/web/src/lib/api/errors.ts` | `ApiError` class + `resolveErrorMessage` | Create |
| `frontend/apps/web/src/lib/api/errorMessages.ts` | Code → Portuguese map | Create |
| `frontend/apps/web/src/lib/api/authBus.ts` | `dispatchAuthExpired` + listener registration helper | Create |
| `frontend/apps/web/src/lib/api/index.ts` | Barrel export | Create |
| `frontend/apps/web/src/lib/api/__tests__/client.test.ts` | vitest for apiFetch behavior | Create |
| `frontend/apps/web/src/lib/api/__tests__/errors.test.ts` | vitest for resolveErrorMessage | Create |
| `frontend/apps/web/src/features/auth/useAuthSession.ts` | Add `auth:expired` listener; honor returnTo on login success | Modify |
| `frontend/apps/web/src/features/documents/v2/**/*.{ts,tsx}` | Replace raw `fetch` with `apiFetch`, replace inline error toasts | Modify |
| `frontend/apps/web/src/features/registry/**/*.{ts,tsx}` | Same migration | Modify |
| `frontend/apps/web/src/features/iam/**/*.{ts,tsx}` | Same migration | Modify |
| `frontend/apps/web/src/features/approval/api/mutationClient.ts` | Refactor to delegate auth/error to `apiFetch`; keep ETag/Idempotency | Modify |
| `wiki/concepts/error-ux.md` | New concept doc | Create (Phase 6) |

---

## Phase 0 — Worktree + Reconnaissance

### Task 0.1: Create worktree

**Files:** none (git op)

- [ ] **Step 1:** Create worktree

```bash
git worktree add ../MetalDocs-group-e-error -b group-e-error main
cd ../MetalDocs-group-e-error
```

- [ ] **Step 2:** Verify clean state

```bash
git status
```
Expected: `nothing to commit, working tree clean` on `group-e-error`.

### Task 0.2: Codex spec validation

- [ ] **Step 1: Dispatch codex auditor**

Prompt (caveman):
```
Validate spec docs/superpowers/specs/2026-05-03-group-e-error-ux-design.md
against code reality. Confirm (a) backend MapErrorToResponse codes match
spec listing, (b) mutationClient.ts:70 toast hardcoded as spec claims,
(c) "Failed to finalize" string exists in documents/v2, (d) auth state
machine in useAuthSession.ts uses "idle"/"loading"/"ready"/"error".
PASS/FAIL per claim.
```

- [ ] **Step 2:** Block on FAIL

### Task 0.3: Locate finalize call site (E3 prerequisite)

- [ ] **Step 1: Grep for hardcoded string**

```bash
grep -rn "Failed to finalize\|Falha ao finalizar" frontend/apps/web/src/
```

- [ ] **Step 2: Record exact file:line in plan log for Phase 2**

(Likely `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`. Confirm before starting Phase 2.)

---

## Phase 1 — Shared `lib/api/` Module + Auth Listener (Codex)

### Task 1.1: Add `errorMessages.ts`

**Files:**
- Create: `frontend/apps/web/src/lib/api/errorMessages.ts`
- Test: `frontend/apps/web/src/lib/api/__tests__/errors.test.ts`

- [ ] **Step 1: Write failing test**

```ts
// frontend/apps/web/src/lib/api/__tests__/errors.test.ts
import { describe, it, expect } from 'vitest';
import { resolveErrorMessage } from '../errors';

describe('resolveErrorMessage', () => {
  it('returns mapped message for known code', () => {
    expect(resolveErrorMessage('sod.submitter_cannot_sign', 'fallback')).toContain('submeteu este documento');
  });

  it('returns backend fallback for unknown code', () => {
    expect(resolveErrorMessage('made.up.code', 'backend says X')).toBe('backend says X');
  });

  it('returns backend fallback when code is undefined', () => {
    expect(resolveErrorMessage(undefined, 'backend says X')).toBe('backend says X');
  });
});
```

- [ ] **Step 2: Run to verify FAIL**

```bash
cd frontend/apps/web && npx vitest run src/lib/api/__tests__/errors.test.ts
```
Expected: FAIL — module not found.

- [ ] **Step 3: Implement errorMessages.ts**

```ts
// frontend/apps/web/src/lib/api/errorMessages.ts
// Backend codes (see internal/modules/documents/approval/http/errors.go)
// → Portuguese user-facing copy. Frontend takes priority; falls back to
// backend error.message when code is unknown. Falls back to a generic
// string only when both are missing.
export const errorMessages: Record<string, string> = {
  // Authn / Authz
  'authn.expired': 'Sessão expirada. Por favor, autentique novamente.',
  'authn.signature_invalid': 'Credenciais inválidas para assinatura.',
  'authn.signature_rate_limited': 'Muitas tentativas de assinatura. Aguarde 30 segundos.',
  'authz.capability_denied': 'Você não tem permissão para esta ação.',

  // Segregation of Duties
  'sod.submitter_cannot_sign': 'Você submeteu este documento e não pode aprová-lo. Outro usuário precisa assinar.',
  'sod.cross_stage_duplicate': 'Você já assinou este documento em uma etapa anterior.',

  // Conflict / state
  'conflict.stale_revision': 'O documento foi atualizado por outra pessoa. Recarregue a página.',
  'conflict.duplicate_submission': 'Submissão duplicada. Aguarde a anterior concluir.',
  'signoff.duplicate': 'Você já assinou este documento.',
  'state.instance_completed': 'Esta aprovação já foi concluída.',
  'route.in_use': 'Esta rota está em uso por uma aprovação ativa e não pode ser alterada.',
  'route.duplicate_profile': 'Já existe uma rota para este perfil de documento.',

  // Not found
  'not_found.instance': 'Aprovação não encontrada.',
  'not_found.route': 'Nenhuma rota de aprovação configurada para este perfil de documento. Configure uma rota antes de finalizar.',

  // Precondition
  'precondition.if_match_required': 'Cabeçalho If-Match obrigatório.',
  'precondition.content_hash_mismatch': 'O conteúdo do documento mudou. Recarregue a página antes de finalizar.',
  'idempotency.key_required': 'Chave de idempotência obrigatória.',

  // Validation
  'validation.if_match_malformed': 'Cabeçalho If-Match inválido.',
  'validation.reason_required': 'Justificativa é obrigatória para esta ação.',
  'validation.json_decode': 'Corpo da requisição inválido.',
  'validation.json_type_error': 'Tipo de campo inválido na requisição.',
  'validation.empty_body': 'Corpo da requisição vazio.',
  'validation.content_type': 'Tipo de conteúdo não suportado.',
  'validation.body_too_large': 'Corpo da requisição muito grande.',
  'validation.duplicate_key': 'Chave duplicada na requisição.',
  'validation.request_invalid': 'Requisição inválida.',

  // Persistence
  'db.fk_violation': 'Operação rejeitada: referência inválida.',
  'db.check_violation': 'Operação rejeitada: validação do banco falhou.',

  // Network / timeout
  'timeout': 'Tempo de resposta excedido. Tente novamente.',
};
```

- [ ] **Step 4: Implement errors.ts**

```ts
// frontend/apps/web/src/lib/api/errors.ts
import { errorMessages } from './errorMessages';

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    public readonly status: number,
    message: string,
    public readonly details?: unknown,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

const GENERIC_FALLBACK = 'Ocorreu um erro. Tente novamente.';

export function resolveErrorMessage(code: string | undefined, backendMessage?: string): string {
  if (code && errorMessages[code]) return errorMessages[code];
  if (backendMessage && backendMessage.trim() !== '') return backendMessage;
  return GENERIC_FALLBACK;
}
```

- [ ] **Step 5: Run to verify PASS**

```bash
npx vitest run src/lib/api/__tests__/errors.test.ts
```
Expected: PASS, all 3 tests green.

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/lib/api/errorMessages.ts frontend/apps/web/src/lib/api/errors.ts frontend/apps/web/src/lib/api/__tests__/errors.test.ts
git commit -m "feat(api): E2/E3 add errorMessages map and ApiError class"
```

### Task 1.2: Add `authBus.ts`

**Files:**
- Create: `frontend/apps/web/src/lib/api/authBus.ts`

- [ ] **Step 1: Implement**

```ts
// frontend/apps/web/src/lib/api/authBus.ts
// Decouples lib/api from auth context. lib/api dispatches; the auth
// context registers a listener and decides what to do (logout + restore
// returnTo on next login).
export const AUTH_EXPIRED_EVENT = 'auth:expired';

export interface AuthExpiredDetail {
  returnTo: string;
}

export function dispatchAuthExpired(returnTo: string): void {
  window.dispatchEvent(
    new CustomEvent<AuthExpiredDetail>(AUTH_EXPIRED_EVENT, { detail: { returnTo } }),
  );
}

export function onAuthExpired(handler: (detail: AuthExpiredDetail) => void): () => void {
  const wrapped = (e: Event) => handler((e as CustomEvent<AuthExpiredDetail>).detail);
  window.addEventListener(AUTH_EXPIRED_EVENT, wrapped);
  return () => window.removeEventListener(AUTH_EXPIRED_EVENT, wrapped);
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/apps/web/src/lib/api/authBus.ts
git commit -m "feat(api): E4 add auth:expired CustomEvent bus"
```

### Task 1.3: Add `client.ts` (`apiFetch`)

**Files:**
- Create: `frontend/apps/web/src/lib/api/client.ts`
- Test: `frontend/apps/web/src/lib/api/__tests__/client.test.ts`

- [ ] **Step 1: Write failing tests**

```ts
// frontend/apps/web/src/lib/api/__tests__/client.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiFetch } from '../client';
import { ApiError } from '../errors';
import { AUTH_EXPIRED_EVENT, type AuthExpiredDetail } from '../authBus';

describe('apiFetch', () => {
  beforeEach(() => {
    vi.spyOn(global, 'fetch');
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('returns parsed JSON on 200', async () => {
    (global.fetch as any).mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    const result = await apiFetch<{ ok: boolean }>('/api/test');
    expect(result).toEqual({ ok: true });
  });

  it('returns undefined on 204', async () => {
    (global.fetch as any).mockResolvedValue(new Response(null, { status: 204 }));
    const result = await apiFetch<undefined>('/api/test', { method: 'DELETE' });
    expect(result).toBeUndefined();
  });

  it('throws ApiError with parsed code on 4xx', async () => {
    (global.fetch as any).mockResolvedValue(new Response(JSON.stringify({ error: { code: 'sod.submitter_cannot_sign', message: 'submitter cannot sign' } }), { status: 403 }));
    await expect(apiFetch('/api/test')).rejects.toMatchObject({
      name: 'ApiError',
      code: 'sod.submitter_cannot_sign',
      status: 403,
    });
  });

  it('dispatches auth:expired on 401 and throws ApiError', async () => {
    Object.defineProperty(window, 'location', { value: { pathname: '/documents/abc', search: '?tab=foo' }, writable: true });
    (global.fetch as any).mockResolvedValue(new Response(null, { status: 401 }));
    const captured: AuthExpiredDetail[] = [];
    const handler = (e: Event) => captured.push((e as CustomEvent<AuthExpiredDetail>).detail);
    window.addEventListener(AUTH_EXPIRED_EVENT, handler);

    await expect(apiFetch('/api/test')).rejects.toBeInstanceOf(ApiError);
    expect(captured).toHaveLength(1);
    expect(captured[0].returnTo).toBe('/documents/abc?tab=foo');

    window.removeEventListener(AUTH_EXPIRED_EVENT, handler);
  });

  it('falls back to http_<status> code when error envelope absent', async () => {
    (global.fetch as any).mockResolvedValue(new Response('plain text', { status: 500 }));
    await expect(apiFetch('/api/test')).rejects.toMatchObject({ code: 'http_500', status: 500 });
  });
});
```

- [ ] **Step 2: Run to verify FAIL**

```bash
npx vitest run src/lib/api/__tests__/client.test.ts
```
Expected: FAIL — `apiFetch` not exported.

- [ ] **Step 3: Implement client.ts**

```ts
// frontend/apps/web/src/lib/api/client.ts
import { ApiError } from './errors';
import { dispatchAuthExpired } from './authBus';

interface BackendErrorEnvelope {
  error?: { code?: string; message?: string; details?: unknown };
}

/**
 * Single HTTP entry point for the frontend. Wraps fetch with:
 *  - Structured error parsing into ApiError
 *  - 401 auth-expired event dispatch (with returnTo)
 *  - 204 handled as undefined
 *  - JSON body parsing for 2xx
 *
 * Callers catch ApiError and toast resolveErrorMessage(err.code, err.message).
 */
export async function apiFetch<T = unknown>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);

  if (res.status === 401) {
    dispatchAuthExpired(window.location.pathname + window.location.search);
    throw new ApiError('authn.expired', 401, 'Sessão expirada');
  }

  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as BackendErrorEnvelope;
    const code = body.error?.code ?? `http_${res.status}`;
    const message = body.error?.message ?? 'Erro interno';
    throw new ApiError(code, res.status, message, body.error?.details);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}
```

- [ ] **Step 4: Add barrel export**

```ts
// frontend/apps/web/src/lib/api/index.ts
export { apiFetch } from './client';
export { ApiError, resolveErrorMessage } from './errors';
export { dispatchAuthExpired, onAuthExpired, AUTH_EXPIRED_EVENT } from './authBus';
export type { AuthExpiredDetail } from './authBus';
```

- [ ] **Step 5: Run to verify PASS**

```bash
npx vitest run src/lib/api/
```
Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/lib/api/client.ts frontend/apps/web/src/lib/api/index.ts frontend/apps/web/src/lib/api/__tests__/client.test.ts
git commit -m "feat(api): E4 add apiFetch shared HTTP client with 401 dispatch"
```

### Task 1.4: Wire auth listener in `useAuthSession`

**Files:**
- Modify: `frontend/apps/web/src/features/auth/useAuthSession.ts`

- [ ] **Step 1: Add listener + returnTo logic**

In `useAuthSession.ts`, add imports near top:

```ts
import { useEffect } from 'react';
import { onAuthExpired } from '../../lib/api';
```

Inside `useAuthSession` body (after the existing `useUiStore` hook, before `clearWorkspaceAfterPasswordChange`):

```ts
// E4: react to auth:expired events from apiFetch (any 401 across the app).
// Stores returnTo in sessionStorage so post-login can restore it.
useEffect(() => {
  return onAuthExpired(({ returnTo }) => {
    if (returnTo && returnTo !== '/' && !returnTo.startsWith('/login')) {
      sessionStorage.setItem('auth:returnTo', returnTo);
    }
    // Mirror handleLogout body without showing an error toast.
    setUser(null);
    setAuthState('idle');
  });
}, [setAuthState, setUser]);
```

In `handleLogin` after `setAuthState("ready")` and `await onAuthenticated(...)` (line 156), append returnTo restoration:

```ts
const returnTo = sessionStorage.getItem('auth:returnTo');
if (returnTo) {
  sessionStorage.removeItem('auth:returnTo');
  // Use react-router navigate via window.location since useAuthSession
  // does not currently consume react-router context; this is
  // intentionally side-effecting so the SPA restores deep-link state.
  window.history.pushState({}, '', returnTo);
  window.dispatchEvent(new PopStateEvent('popstate'));
}
```

> **Note:** `App.tsx` listens to popstate via `useLocation`. If verification (Phase 0.3 follow-up) shows useAuthSession already has access to a router navigate function via props, replace the `window.history.pushState` block with `navigate(returnTo)`.

- [ ] **Step 2: Add unit test for listener**

```ts
// frontend/apps/web/src/features/auth/__tests__/useAuthSession.returnTo.test.tsx
import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAuthSession } from '../useAuthSession';
import { dispatchAuthExpired } from '../../../lib/api';

vi.mock('../../../lib.api', () => ({ api: { me: vi.fn(), login: vi.fn(), logout: vi.fn(), changePassword: vi.fn() } }));

describe('useAuthSession auth:expired listener', () => {
  beforeEach(() => sessionStorage.clear());

  it('stores returnTo on auth:expired event', () => {
    const onAuthenticated = vi.fn();
    renderHook(() => useAuthSession({ onAuthenticated }));
    act(() => dispatchAuthExpired('/documents/abc'));
    expect(sessionStorage.getItem('auth:returnTo')).toBe('/documents/abc');
  });

  it('ignores root path and login paths', () => {
    const onAuthenticated = vi.fn();
    renderHook(() => useAuthSession({ onAuthenticated }));
    act(() => dispatchAuthExpired('/'));
    expect(sessionStorage.getItem('auth:returnTo')).toBeNull();
    act(() => dispatchAuthExpired('/login'));
    expect(sessionStorage.getItem('auth:returnTo')).toBeNull();
  });
});
```

- [ ] **Step 3: Run to verify PASS**

```bash
npx vitest run src/features/auth/__tests__/useAuthSession.returnTo.test.tsx
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/apps/web/src/features/auth/useAuthSession.ts frontend/apps/web/src/features/auth/__tests__/useAuthSession.returnTo.test.tsx
git commit -m "feat(auth): E4 listen for auth:expired and restore returnTo on login"
```

---

## Phase 1 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Review group-e-error commits since Phase 0. Confirm: (a) ApiError carries
code+status+message+details, (b) apiFetch dispatches auth:expired with
correct returnTo (pathname+search), (c) errorMessages.ts covers all
backend codes from MapErrorToResponse, (d) useAuthSession listener clears
on unmount, (e) test coverage ≥85% for lib/api. PASS/FAIL.
```

Block on FAIL.

---

## Phase 2 — Migrate `documents/v2` (Codex, sequential)

### Task 2.1: Inventory call sites

- [ ] **Step 1: List every fetch call in documents/v2**

```bash
grep -rn "await fetch(\|fetch(['\"]" frontend/apps/web/src/features/documents/v2/
```

- [ ] **Step 2: Record list**

Expected categories: editor save (PUT/PATCH revisions), freeze (POST), rename (PUT name), archive (POST), get document, get revisions, signed URL fetch.

### Task 2.2: Migrate editor save flow (failing test → passing)

**Files:**
- Modify: each `documents/v2/**/*.{ts,tsx}` mutation site (specific paths from inventory)
- Test: `frontend/apps/web/src/features/documents/v2/__tests__/<feature>.test.tsx`

For each mutation call site identified in 2.1, perform the bite-sized loop:

- [ ] **Step 1: Write failing test**

Example for freeze flow (template — adapt path/name per call site):

```tsx
// frontend/apps/web/src/features/documents/v2/__tests__/freeze.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { toast } from 'sonner';
import { FreezeButton } from '../components/FreezeButton';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

describe('FreezeButton E3', () => {
  beforeEach(() => vi.spyOn(global, 'fetch'));
  afterEach(() => vi.restoreAllMocks());

  it('shows mapped error message when finalize fails with not_found.route', async () => {
    (global.fetch as any).mockResolvedValue(new Response(
      JSON.stringify({ error: { code: 'not_found.route', message: 'no route' } }),
      { status: 404 },
    ));
    render(<FreezeButton documentId="doc-1" />);
    fireEvent.click(screen.getByRole('button', { name: /finalizar/i }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(expect.stringContaining('Nenhuma rota de aprovação'));
    });
  });
});
```

- [ ] **Step 2: Run to verify FAIL**

```bash
npx vitest run src/features/documents/v2/__tests__/freeze.test.tsx
```
Expected: FAIL — current code shows "Failed to finalize document".

- [ ] **Step 3: Migrate component**

Replace the freeze handler:

```ts
// before:
const res = await fetch(`/api/v1/documents/${documentId}/freeze`, { method: 'POST' });
if (!res.ok) { toast.error('Failed to finalize document'); return; }

// after:
import { apiFetch, ApiError, resolveErrorMessage } from '../../../lib/api';

try {
  await apiFetch<void>(`/api/v1/documents/${documentId}/freeze`, { method: 'POST' });
  toast.success('Documento finalizado.');
} catch (err) {
  if (err instanceof ApiError) {
    toast.error(resolveErrorMessage(err.code, err.message));
  } else {
    toast.error('Erro inesperado ao finalizar.');
  }
}
```

- [ ] **Step 4: Run to verify PASS**

```bash
npx vitest run src/features/documents/v2/__tests__/freeze.test.tsx
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/v2/components/FreezeButton.tsx frontend/apps/web/src/features/documents/v2/__tests__/freeze.test.tsx
git commit -m "refactor(documents): E3 freeze uses apiFetch + resolveErrorMessage"
```

Repeat steps 1-5 for each remaining mutation site identified in Task 2.1 (editor save, rename, archive, etc.). Each is its own commit.

### Task 2.3: Verify documents/v2 GET sites

- [ ] **Step 1: Migrate GET calls used in mutation flows**

GET calls inside mutation flows (e.g., refetch after save) should also use `apiFetch` so a 401 mid-flow triggers the listener. List-page initial loads (GET-only) can defer; flag in audit doc as follow-up if not migrated.

- [ ] **Step 2: Commit per file**

Same TDD loop. One commit per migrated file.

---

## Phase 2 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Review Phase 2 commits. Confirm: (a) every documents/v2 mutation uses
apiFetch, (b) every catch block uses resolveErrorMessage, (c) no
hardcoded error strings remain in catch blocks (grep "toast.error\(['\"]"
in documents/v2 should yield only branded fallbacks), (d) tests run
green. PASS/FAIL.
```

---

## Phase 3 — Parallel: Migrate registry ‖ iam (Sonnet)

### Task 3.1: Migrate registry features (Sonnet, parallel)

**Files:**
- Modify: `frontend/apps/web/src/features/registry/**/*.{ts,tsx}` mutation sites

- [ ] **Step 1: Inventory**

```bash
grep -rn "await fetch(" frontend/apps/web/src/features/registry/
```

- [ ] **Step 2-N: Per-file TDD loop**

For each mutation site: write failing test using same template as Task 2.2 (sub `registry` for `documents/v2`), implement migration, run to PASS, commit per file.

Common registry sites: list page (GET — defer), detail page (mutations — migrate), create dialog (POST — migrate), rename (PUT — migrate).

- [ ] **Step Final: Verify**

```bash
npx vitest run src/features/registry/
go test -mod=mod ./...  # backend regression
```

### Task 3.2: Migrate iam features (Sonnet, parallel with 3.1)

**Files:**
- Modify: `frontend/apps/web/src/features/iam/**/*.{ts,tsx}` mutation sites

- [ ] **Step 1: Inventory**

```bash
grep -rn "await fetch(" frontend/apps/web/src/features/iam/
```

- [ ] **Step 2-N: Per-file TDD loop**

Same template as Task 2.2.

Common iam sites: user list, create user, set role, unlock user.

- [ ] **Step Final: Verify**

```bash
npx vitest run src/features/iam/
```

---

## Phase 3 Review (Opus)

Prompt:
```
Review Phase 3 commits (registry, iam). Same checklist as Phase 2.
PASS/FAIL.
```

---

## Phase 4 — Refactor `mutationClient` to delegate (Sonnet)

### Task 4.1: Rewire mutationClient over apiFetch

**Files:**
- Modify: `frontend/apps/web/src/features/approval/api/mutationClient.ts`
- Test: `frontend/apps/web/src/features/approval/api/__tests__/mutationClient.test.ts`

- [ ] **Step 1: Update existing tests if needed; add regression test**

```ts
// frontend/apps/web/src/features/approval/api/__tests__/mutationClient.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mutate, ApprovalError } from '../mutationClient';

describe('mutate (post-refactor)', () => {
  beforeEach(() => vi.spyOn(global, 'fetch'));
  afterEach(() => vi.restoreAllMocks());

  it('still returns parsed JSON for 200', async () => {
    (global.fetch as any).mockResolvedValue(new Response(JSON.stringify({ id: '1' }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    const result = await mutate<unknown, { id: string }>('POST', '/api/v1/approval/x', {});
    expect(result).toEqual({ id: '1' });
  });

  it('throws ApprovalError-shaped error on 412 and calls on412', async () => {
    (global.fetch as any).mockResolvedValue(new Response(JSON.stringify({ error: { code: 'conflict.stale' } }), { status: 412, headers: { 'ETag': 'W/"1"' } }));
    const on412 = vi.fn();
    await expect(mutate('PUT', '/x', {}, { resourceId: 'r1', on412 })).rejects.toMatchObject({ code: 'conflict.stale', status: 412 });
    expect(on412).toHaveBeenCalledWith('r1');
  });
});
```

- [ ] **Step 2: Refactor mutationClient.ts**

```ts
// frontend/apps/web/src/features/approval/api/mutationClient.ts
// @ts-expect-error pacote uuid não tipado neste app
import { v4 as uuidv4 } from 'uuid';
import { toast } from 'sonner';

import { etagCache } from './etagCache';
import { apiFetch, ApiError } from '../../../lib/api';

// Re-export for backwards compatibility with existing import sites.
export class ApprovalError extends ApiError {
  constructor(code: string, status: number, message: string) {
    super(code, status, message);
    this.name = 'ApprovalError';
  }
}

export interface MutateOptions {
  idempotencyKey?: string;
  resourceId?: string;
  ifMatch?: string;
  on412?: (resourceId: string) => void;
}

/**
 * Approval-module mutation helper. Adds ETag/Idempotency-Key concerns
 * on top of apiFetch. Auth/error parsing delegated to apiFetch.
 */
export async function mutate<TReq, TRes>(
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  url: string,
  body?: TReq,
  opts: MutateOptions = {},
): Promise<TRes> {
  const idempotencyKey = opts.idempotencyKey ?? uuidv4();
  const ifMatch = opts.ifMatch ?? (opts.resourceId ? etagCache.get(opts.resourceId) : undefined);

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Idempotency-Key': idempotencyKey,
  };
  if (ifMatch) headers['If-Match'] = ifMatch;

  try {
    // We need access to the Response for ETag header capture, so we use
    // a thin wrapper rather than apiFetch directly.
    const res = await fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    const newETag = res.headers.get('ETag');
    if (newETag && opts.resourceId) {
      etagCache.set(opts.resourceId, newETag);
    }

    if (res.status === 401) {
      // Delegate to lib/api by re-using its dispatch path.
      const { dispatchAuthExpired } = await import('../../../lib/api');
      dispatchAuthExpired(window.location.pathname + window.location.search);
      throw new ApprovalError('authn.expired', 401, 'Sessão expirada');
    }

    if (res.status === 412) {
      const responseBody = (await res.json().catch(() => ({}))) as { error?: { code?: string } };
      if (opts.on412 && opts.resourceId) {
        opts.on412(opts.resourceId);
      } else {
        toast.error('Documento foi alterado. Por favor, atualize a página.');
      }
      throw new ApprovalError(responseBody.error?.code ?? 'conflict.stale', 412, 'Stale resource');
    }

    if (!res.ok) {
      const responseBody = (await res.json().catch(() => ({}))) as { error?: { code?: string; message?: string } };
      throw new ApprovalError(
        responseBody.error?.code ?? `http_${res.status}`,
        res.status,
        responseBody.error?.message ?? 'Erro interno',
      );
    }

    if (res.status === 204) return undefined as TRes;
    return (await res.json()) as TRes;
  } catch (err) {
    if (err instanceof ApprovalError) throw err;
    if (err instanceof ApiError) {
      throw new ApprovalError(err.code, err.status, err.message);
    }
    throw err;
  }
}
```

- [ ] **Step 3: Run vitest for approval module**

```bash
npx vitest run src/features/approval/
```
Expected: all PASS, including pre-existing tests.

- [ ] **Step 4: Commit**

```bash
git add frontend/apps/web/src/features/approval/api/mutationClient.ts frontend/apps/web/src/features/approval/api/__tests__/mutationClient.test.ts
git commit -m "refactor(approval): mutationClient delegates 401 dispatch to apiFetch"
```

### Task 4.2: Update SignoffDialog toast to use resolver (E2)

**Files:**
- Modify: `frontend/apps/web/src/features/approval/components/SignoffDialog.tsx`

- [ ] **Step 1: Find error toast inside SignoffDialog**

```bash
grep -n "toast.error" frontend/apps/web/src/features/approval/components/SignoffDialog.tsx
```

- [ ] **Step 2: Wrap with resolveErrorMessage**

For each catch block, replace generic toast with:

```ts
import { ApiError, resolveErrorMessage } from '../../../lib/api';
// ...
catch (err) {
  if (err instanceof ApiError) {
    toast.error(resolveErrorMessage(err.code, err.message));
  } else {
    toast.error('Erro ao processar assinatura.');
  }
}
```

- [ ] **Step 3: Add E2 acceptance test**

```tsx
// frontend/apps/web/src/features/approval/components/__tests__/SignoffDialog.E2.test.tsx
import { render, fireEvent, waitFor, screen } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { toast } from 'sonner';
import { SignoffDialog } from '../SignoffDialog';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

describe('SignoffDialog E2', () => {
  beforeEach(() => vi.spyOn(global, 'fetch'));
  afterEach(() => vi.restoreAllMocks());

  it('shows specific message when submitter tries to sign', async () => {
    (global.fetch as any).mockResolvedValue(new Response(
      JSON.stringify({ error: { code: 'sod.submitter_cannot_sign', message: 'submitter cannot sign' } }),
      { status: 403 },
    ));
    render(<SignoffDialog instanceId="i-1" stageId="s-1" open onClose={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: /assinar/i }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(expect.stringContaining('submeteu este documento'));
    });
  });
});
```

- [ ] **Step 4: Run + commit**

```bash
npx vitest run src/features/approval/components/__tests__/SignoffDialog.E2.test.tsx
git add frontend/apps/web/src/features/approval/components/SignoffDialog.tsx frontend/apps/web/src/features/approval/components/__tests__/SignoffDialog.E2.test.tsx
git commit -m "feat(approval): E2 SignoffDialog uses resolveErrorMessage for SoD codes"
```

---

## Phase 4 Review (Opus)

Prompt:
```
Review Phase 4 commits. Confirm: (a) ApprovalError extends ApiError so
existing instanceof checks keep working, (b) etagCache writes preserved
on 200/204, (c) on412 callback still fires, (d) E2 SignoffDialog shows
specific Portuguese message. PASS/FAIL.
```

---

## Phase 5 — Full Verification

### Task 5.1: Frontend test suite

- [ ] **Step 1**

```bash
cd frontend/apps/web && npx vitest run
```
Expected: all PASS.

### Task 5.2: Backend regression check

- [ ] **Step 1**

```bash
go test -mod=mod ./...
```
Expected: all PASS (zero backend changes expected).

### Task 5.3: Smoke tests

- [ ] **Step 1: Start API**

```powershell
.\scripts\start-api.ps1 -Build
```

- [ ] **Step 2: E4 smoke — session expiry returnTo**

```
1. Login as admin / AdminMetalDocs123!
2. Navigate to /documents/v2/<some-doc-id>
3. In a separate psql window:
   UPDATE metaldocs.auth_sessions SET expires_at = now() - interval '1h' WHERE user_id = (SELECT id FROM metaldocs.iam_users WHERE username='admin');
4. In the SPA, click any mutation (rename, save, etc.)
5. Verify: toast "Sessão expirada", AuthShell appears, log in again
6. Verify: navigated back to /documents/v2/<same-doc-id>
```

- [ ] **Step 3: E2 smoke — submitter SoD message**

```
1. Login as approver
2. Submit a document for review (creates instance where actor=submitter)
3. Try to sign your own submission
4. Verify: toast contains "submeteu este documento e não pode aprová-lo"
```

- [ ] **Step 4: E3 smoke — finalize without route**

```
1. Login as admin
2. Create document with profile that has no approval route
3. Click Finalizar
4. Verify: toast contains "Nenhuma rota de aprovação configurada"
```

### Task 5.4: Codex independent audit

- [ ] **Dispatch codex auditor**

Prompt (caveman):
```
Independent audit Group E sub-plan 1 (E2 E3 E4) on branch group-e-error.
Per bug PASS/FAIL with file:line evidence.
- E2: SignoffDialog catch uses resolveErrorMessage; errorMessages
  includes sod.submitter_cannot_sign in Portuguese
- E3: documents/v2 finalize call uses apiFetch + resolveErrorMessage,
  no "Failed to finalize" string remains
- E4: apiFetch dispatches auth:expired on 401 with returnTo;
  useAuthSession listener stores returnTo, handleLogin restores it
No fixes. Report only.
```

Block on FAIL.

### Task 5.5: Lint + coverage

- [ ] **Step 1**

```bash
cd frontend/apps/web && npm run lint
```
Expected: zero new warnings vs main.

- [ ] **Step 2**

```bash
npx vitest run --coverage src/lib/api/
```
Expected: ≥85% line coverage in `lib/api/`.

---

## Phase 5 Review (Opus)

Prompt:
```
Final review Group E sub-plan 1. Confirm: tests, smoke, codex audit, lint,
coverage all green. Identify any spec acceptance criterion not yet met.
PASS/FAIL.
```

---

## Phase 6 — Wiki + Audit + Merge

### Task 6.1: Concept doc

**Files:**
- Create: `wiki/concepts/error-ux.md`

- [ ] **Step 1: Write doc**

```markdown
# Error UX — frontend ↔ backend contract

> Last verified: 2026-05-03

## Backend contract

`MapErrorToResponse` (`internal/modules/documents/approval/http/errors.go:21`)
emits `{error:{code, message, details?}, request_id}`. Codes are stable
identifiers (e.g., `sod.submitter_cannot_sign`); messages are Portuguese
fallbacks for codes the frontend does not yet translate.

## Frontend layers

- `frontend/apps/web/src/lib/api/client.ts` — `apiFetch<T>`. Throws
  `ApiError{code, status, message, details}`. Dispatches `auth:expired`
  CustomEvent with `returnTo` on 401.
- `frontend/apps/web/src/lib/api/errorMessages.ts` — code → Portuguese
  map. Single source of truth for user-facing copy.
- `frontend/apps/web/src/lib/api/errors.ts` — `resolveErrorMessage(code,
  fallback)`. Frontend map first, backend message second, generic third.
- `frontend/apps/web/src/lib/api/authBus.ts` — `auth:expired` dispatch +
  listener helper. Decouples HTTP layer from React auth context.

## Auth-expired flow

1. Any mutation hits 401
2. `apiFetch` dispatches `auth:expired` with `returnTo = location.pathname + search`
3. `useAuthSession` listener stores returnTo in `sessionStorage`,
   resets auth state to `idle` (AuthShell renders)
4. User logs in; `handleLogin` post-success reads sessionStorage,
   navigates back, clears the key

## Adding a new code

1. Backend: add case in `MapErrorToResponse`, return new code
2. Frontend: add Portuguese copy in `errorMessages.ts`
3. Component catch blocks already use `resolveErrorMessage` — no further change

## Migration status

Auth-critical features migrated to `apiFetch`: documents/v2, registry,
approval, iam. Pending follow-up: taxonomy admin, templates_v2 admin,
audit-logs viewer.

## References

- Spec: `docs/superpowers/specs/2026-05-03-group-e-error-ux-design.md`
- Plan: `docs/superpowers/plans/2026-05-03-group-e-error-ux.md`
- ADR: none (pure pattern, no architectural inversion)
```

- [ ] **Step 2: Commit**

```bash
git add wiki/concepts/error-ux.md
git commit -m "docs(wiki): error UX frontend ↔ backend contract"
```

### Task 6.2: Close audit entries

**Files:**
- Modify: `wiki/bugs/audit-2026-05-03.md`

- [ ] **Step 1: Mark E2/E3/E4 closed**

For each of E2/E3/E4 in the audit doc, add `[x]` and a closure line referencing actual commit SHAs from `git log --oneline group-e-error ^main`:

```markdown
| E2 | ISO segregation not communicated... | 🟠 high | fixed | <sha> |
| E3 | No active approval route → opaque... | 🟠 high | fixed | <sha> |
| E4 | No global 401 interceptor... | 🟠 high | fixed | <sha> |
```

Add a "Pending follow-up" sub-section listing taxonomy/templates_v2/audit-logs migration.

- [ ] **Step 2: Commit**

```bash
git add wiki/bugs/audit-2026-05-03.md
git commit -m "docs(audit): close E2/E3/E4 with commit SHAs"
```

### Task 6.3: Wiki-curator subagent dispatch

- [ ] **Dispatch wiki-curator agent**

Prompt:
```
Group E sub-plan 1 (error UX) merged. Refresh wiki:
- wiki/concepts/error-ux.md verified file:line anchors
- wiki/modules/frontend-*.md (if exists) Last verified stamps
- wiki/README.md index entry for new concept doc
- Cross-link from wiki/concepts/error-ux.md to backend
  internal/modules/documents/approval/http/errors.go (with line ref)
Commit changes.
```

### Task 6.4: Finish branch

- [ ] **Step 1: Invoke finishing-a-development-branch skill**

Run tests one more time, choose Option 1 (merge locally) or Option 2 (PR), per user preference.

---

## Self-Review Checklist

**Spec coverage:**
- [x] E2 → Tasks 1.1 (errorMessages), 4.2 (SignoffDialog)
- [x] E3 → Tasks 1.1 (errorMessages), 2.2 (freeze migration)
- [x] E4 → Tasks 1.2 (authBus), 1.3 (apiFetch 401), 1.4 (useAuthSession listener), 2.x/3.x/4.1 (migrations)
- [x] errorMessages map → Task 1.1
- [x] apiFetch shared client → Task 1.3
- [x] mutationClient refactor → Task 4.1
- [x] returnTo preservation → Task 1.4

**Placeholder scan:** none — every step contains exact code or commands. Per-mutation-site loop in Tasks 2.2, 3.1, 3.2 references the same template explicitly (does not say "similar to Task X").

**Type consistency:**
- `ApiError(code, status, message, details?)` — Task 1.1 + 1.3 + 4.1 match
- `apiFetch<T>(url, init?) → Promise<T>` — Task 1.3 + every migration site match
- `resolveErrorMessage(code, fallback)` — Task 1.1 signature; consumers in Task 2.2, 3.x, 4.2 match
- `dispatchAuthExpired(returnTo: string)` and `onAuthExpired(handler)` — Task 1.2 definitions; Task 1.3 + 1.4 consumers match
- `AuthExpiredDetail.returnTo: string` — Task 1.2 type; Task 1.3 dispatch + Task 1.4 listener match
- `ApprovalError extends ApiError` — Task 4.1; existing approval call sites unchanged

---

## Acceptance Criteria (from spec)

- [ ] E2: submitter clicks Approve → toast contains "submeteu este documento"
- [ ] E3: finalize without route → toast contains "Nenhuma rota de aprovação"
- [ ] E4: any auth-critical page hits 401 → returns to original page after re-login
- [ ] `apiFetch` is the single mutation entry point in migrated features
- [ ] All vitest pass
- [ ] Backend `go test ./...` unchanged
- [ ] Codex audit returns 3/3 PASS
- [ ] Audit doc updated, E2/E3/E4 closed with commit SHAs
- [ ] Follow-up tasks documented for non-migrated features

---

## References

- Spec: `docs/superpowers/specs/2026-05-03-group-e-error-ux-design.md`
- Audit: `wiki/bugs/audit-2026-05-03.md` (E2 line 216, E3 line 217, E4 line 218)
- Backend errors: `internal/modules/documents/approval/http/errors.go`
- Existing helper: `frontend/apps/web/src/features/approval/api/mutationClient.ts`
- Auth state: `frontend/apps/web/src/features/auth/useAuthSession.ts`
- Sub-plan 2: `docs/superpowers/specs/2026-05-03-group-e-editor-design.md` (TBD)
- Sub-plan 3: `docs/superpowers/specs/2026-05-03-group-e-misc-design.md` (TBD)
