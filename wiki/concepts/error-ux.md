# Error UX — Shared HTTP Client & Auth Bus

> **Migration note.** The legacy `ApiErrorEnvelope` shape documented below (`{ error: { code, message } }`) will be replaced by RFC 9457 Problem in Plan 2. See [`architecture/api-design-system.md`](../architecture/api-design-system.md) for the incoming contract.

> **Audit envelope note.** `GET /api/v1/audit/events` emits the same legacy `{error:{code,message,details,trace_id}}` envelope on errors (audit T-002 in [`modules/audit-tech-debt.md`](../modules/audit-tech-debt.md)). The success body is `{"items":[...]}` — not a `data` wrapper. See [`modules/audit.md §8.2`](../modules/audit.md) for the full envelope spec.

> **Last verified:** 2026-05-10
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
| `src/lib/api/resolveQueryError.ts` | `resolveQueryError(err, fallback)` — wraps the `ApiError`/`Error`/unknown triad for TanStack Query `onError` callbacks |
| `src/lib/api/index.ts` | Barrel re-export of all above including `resolveQueryError` |
| `src/features/auth/useAuthSession.ts` | Registers `auth:expired` listener; stores returnTo; restores on login |
| `src/features/approval/api/mutationClient.ts` | `ApprovalError extends ApiError`; 401 uses auth-bus; 403 throws with code |
| `src/features/approval/components/SignoffDialog.tsx` | E2 SoD states: `error_sod_submitter`, `error_sod_duplicate` |
| `src/features/documents/pages/DocumentEditorPage.tsx` | E3 finalize uses `resolveErrorMessage` |

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
- `approval.unresolved_comments` — final approval/release is blocked until active document comments are resolved; handled inline in `SignoffDialog`, not as a generic outage message
- `signoff.not_eligible` — actor not in `eligible_actor_ids` snapshot frozen at submit time; HTTP 403. Mapped from `domain.ErrActorNotEligible` at `internal/modules/documents/approval/http/errors.go:48-50`. Handle analogously to SoD codes in `SignoffDialog` — inline dialog error, not a toast.
- `not_found.route` — no approval route configured for document profile
- `authn.expired` — session expired
- `authn.rate_limited` — too many attempts
- `authz.capability_denied` — insufficient role

Full map: `src/lib/api/errorMessages.ts`

---

## resolveQueryError

```ts
// src/lib/api/resolveQueryError.ts:10
function resolveQueryError(err: unknown, fallback: string): string
```

Convenience wrapper for TanStack Query `onError` callbacks. Replaces the per-callsite `ApiError` instanceof triad:

```ts
// Before (per-callsite pattern)
const message = err instanceof ApiError
  ? resolveErrorMessage(err.code, err.message)
  : err instanceof Error ? err.message : fallback;

// After
const message = resolveQueryError(err, fallback);
```

Re-exported from `lib/api/index.ts`. Used in `NewDocumentWizardPage` (`onError`) and `StepAreaCodeVisibility` (inline areas error). See `modules/novo-documento-wizard.md` for the wizard usage.

Additional callsite: `TemplateEditorPage` (`frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`) uses a local `resolveError(err, fallback)` helper that funnels `ApiError → resolveErrorMessage(code, message)` for both `submitForReview` and `importDocx` errors. `submitForReview` specifically routes through `apiFetch` (wired at `frontend/apps/web/src/features/templates/api/templates.ts`) which throws `ApiError` on non-ok responses. Error codes relevant to review submission (e.g. `authz.capability_denied`, future `template.review_gap`) would resolve via the shared `errorMessages` map. See `wiki/backlog/template-editor.md#submitForReview-error-codes` for deferred error-code coverage.

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

The same dialog now also resolves mapped business conflicts such as `approval.unresolved_comments` through `resolveErrorMessage(code)` before falling back to the generic `error_server` copy. Unmapped backend codes still keep the safe generic message.

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
