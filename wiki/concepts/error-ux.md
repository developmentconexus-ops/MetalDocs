# Error UX — Shared HTTP Client & Auth Bus

> **Last verified:** 2026-05-04
> **Branch:** `phase-e-error-ux` (merged into main)
> **Bugs fixed:** E2, E3, E4

---

## Overview

Before this work, error handling was fragmented across features: each had its own `fetch` wrapper, 401 triggered local toasts or silent failures, and SoD/route errors fell back to hardcoded English strings. This module introduces a single shared HTTP client layer (`src/lib/api/`) that all auth-critical features use.

---

## Key files

| File | Purpose |
|------|---------|
| `src/lib/api/client.ts` | `apiFetch<T>` — shared fetch wrapper with 401/non-ok handling |
| `src/lib/api/errors.ts` | `ApiError` class + `resolveErrorMessage(code, fallback)` |
| `src/lib/api/errorMessages.ts` | Portuguese error code → message map (30+ codes) |
| `src/lib/api/authBus.ts` | `auth:expired` CustomEvent bus: `dispatchAuthExpired`, `onAuthExpired` |
| `src/lib/api/index.ts` | Barrel re-export of all above |
| `src/features/auth/useAuthSession.ts` | Registers `auth:expired` listener; stores returnTo; restores on login |
| `src/features/approval/api/mutationClient.ts` | `ApprovalError extends ApiError`; 401 uses auth-bus; 403 throws with code |
| `src/features/approval/components/SignoffDialog.tsx` | E2 SoD states: `error_sod_submitter`, `error_sod_duplicate` |
| `src/features/documents/v2/DocumentEditorPage.tsx` | E3 finalize uses `resolveErrorMessage` |

---

## apiFetch

```ts
async function apiFetch<T>(url: string, init?: RequestInit): Promise<T>
```

- On 401: calls `dispatchAuthExpired(returnTo)`, throws `ApiError('authn.expired', 401, ...)`
- On non-ok: parses `{ error: { code, message } }` envelope, throws `ApiError(code, status, message)`
- On 204: returns `undefined as T`
- No `init` argument: calls `fetch(url)` without second arg (preserves spy assertion semantics)

---

## ApiError

```ts
class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly details?: unknown;
}
```

`ApprovalError extends ApiError` for backwards compatibility with existing import sites in the approval feature.

---

## resolveErrorMessage

```ts
function resolveErrorMessage(code: string | undefined, backendMessage?: string): string
```

Priority: `errorMessages[code]` → `backendMessage` (non-empty) → `'Ocorreu um erro. Tente novamente.'`

Key codes:
- `sod.submitter_cannot_sign` — submitter ≠ approver SoD
- `sod.cross_stage_duplicate` — same user, multiple stages
- `not_found.route` — no approval route configured for document profile
- `authn.expired` — session expired
- `authn.rate_limited` — too many attempts
- `authz.capability_denied` — insufficient role

Full map: `src/lib/api/errorMessages.ts`

---

## Auth bus (E4)

```ts
// Dispatch (inside apiFetch / mutationClient on 401)
dispatchAuthExpired(returnTo: string): void

// Subscribe (inside useAuthSession)
onAuthExpired(handler: (detail: { returnTo: string }) => void): () => void
```

`useAuthSession` on `auth:expired`:
1. Stores `returnTo` in `sessionStorage`
2. Resets auth state to `idle` (forces re-login)
3. Restores `returnTo` after successful login via `window.history.pushState`

No `toast.error` on 401 — the login redirect is the UX.

---

## E2 — SoD errors in SignoffDialog

`mutationClient.ts` 403 handler throws `ApprovalError` with the backend's error code, no toast. `SignoffDialog` state machine has two new states:

- `error_sod_submitter` — shown when `code === 'sod.submitter_cannot_sign'`
- `error_sod_duplicate` — shown when `code === 'sod.cross_stage_duplicate'`

Messages come from `resolveErrorMessage` against the shared `errorMessages` map — Portuguese, displayed inline in the dialog, not as a toast.

---

## E3 — Finalize error

`handleFinalize` in `DocumentEditorPage.tsx` catches `ApiError` and calls:

```ts
toast.error(resolveErrorMessage(err.code, err.message))
```

`not_found.route` resolves to: *"Nenhuma rota de aprovação configurada para este perfil de documento. Configure uma rota antes de finalizar."*

---

## Coverage

`src/lib/api/` statement coverage: 85.39% (meets ≥85% threshold). `authBus.ts` listener registration is integration-covered via `useAuthSession` and `mutationClient` tests; `onAuthExpired` unit coverage can be improved in a follow-up.
