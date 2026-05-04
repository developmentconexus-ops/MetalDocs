# Group E (sub-plan 1) — Error UX Layer Design

> **Status:** approved 2026-05-03
> **Scope:** Fix bugs E2, E3, E4 from `wiki/bugs/audit-2026-05-03.md` (lines 216–218). Establish a shared frontend HTTP client with structured error handling, message-code translation, and a global 401 interceptor that preserves the user's location.
> **Out of scope:** E1 / E10 / E11 (sub-plan 2: editor lifecycle). E5 / E7 / E9 / E12 (sub-plan 3: misc UI fixes). Backend error mapping changes — `MapErrorToResponse` already emits structured codes correctly.

---

## Why This Spec Exists

The backend already returns a clean error contract: `{error:{code,message,details?}, request_id}` with 30+ specific codes (`MapErrorToResponse` at `internal/modules/documents/approval/http/errors.go:21`). The frontend partially honors it: the approval module has `mutationClient.mutate()` that maps 401/403/412/429 and reads `error.code`. But this helper is approval-scoped. Documents, registry, and IAM pages call raw `fetch()` and surface generic strings.

Three concrete user-visible failures (E2/E3/E4) all trace to the same gap. This spec establishes the missing shared layer once.

---

## Architecture: Three-Layer Error Stack

```
Component  →  apiFetch  →  Backend MapErrorToResponse
   ↑             │
   │             └─ on 401: dispatch auth:expired (with returnTo)
   │             └─ on !ok: throw ApiError{code, status, message}
   │
   └─ catch: toast(resolveErrorMessage(err.code, err.message))
```

| Layer | Responsibility | Files |
|---|---|---|
| HTTP | Parse error envelope, dispatch auth events, throw typed `ApiError` | `frontend/apps/web/src/lib/api/client.ts` |
| Translation | Code → Portuguese map, fallback to backend message, fallback to generic | `frontend/apps/web/src/lib/api/errorMessages.ts` |
| Component | Catch `ApiError`, show toast | every feature module |
| Auth bus | Decouple HTTP from auth context via `auth:expired` CustomEvent | `frontend/apps/web/src/lib/api/authBus.ts` + listener in auth provider |

**Backend touch:** zero. Existing codes are correct.

**Migration discipline:** Phase 1 builds the shared module + auth listener. Phases 2–5 migrate auth-critical features (documents/v2, registry, approval verify, iam). Out of scope this plan: taxonomy admin, templates_v2 admin, audit-logs viewer — flagged as follow-up.

---

## Per-Bug Fix Design

### E2 — ISO segregation generic error

**Files:**
- Modify: `frontend/apps/web/src/lib/api/errorMessages.ts` (add codes)
- Modify: `frontend/apps/web/src/features/approval/api/mutationClient.ts:70` (use resolver)
- Verify: `frontend/apps/web/src/features/approval/components/SignoffDialog.tsx` toast path

**Problem:** Submitter clicks Approve → backend returns `403 sod.submitter_cannot_sign` with Portuguese message → frontend shows hardcoded "Permissão negada".

**Fix:** Add code mappings:
```ts
'sod.submitter_cannot_sign': 'Você submeteu este documento e não pode aprová-lo. Outro usuário precisa assinar.',
'sod.cross_stage_duplicate': 'Você já assinou este documento em uma etapa anterior.',
'authz.capability_denied': 'Você não tem permissão para esta ação.',
```

`mutationClient.ts:70` switches:
```ts
toast.error(resolveErrorMessage(responseBody.error?.code, responseBody.error?.message ?? 'Permissão negada.'));
```

**Test:** vitest renders SignoffDialog with mocked 403, asserts toast contains "submeteu este documento".

---

### E3 — Failed to finalize generic error

**Files:**
- Modify: finalize call site in `frontend/apps/web/src/features/documents/v2/` (locate via `grep -rn "Failed to finalize" frontend/`)
- Modify: `frontend/apps/web/src/lib/api/errorMessages.ts`

**Problem:** Backend returns specific code (`not_found.route`, `validation.reason_required`, etc.). Frontend catches error, shows hardcoded `"Failed to finalize document"`.

**Fix:** Replace hardcoded string with `resolveErrorMessage(err.code, err.message)`. Add codes:
```ts
'not_found.route': 'Nenhuma rota de aprovação configurada para este perfil de documento. Configure uma rota antes de finalizar.',
'validation.reason_required': 'Justificativa é obrigatória para esta ação.',
'precondition.content_hash_mismatch': 'O conteúdo do documento mudou. Recarregue a página antes de finalizar.',
```

**Test:** vitest mocks finalize 404 with `not_found.route`, click finalize, assert toast contains "Nenhuma rota".

---

### E4 — No global 401 interceptor

**Files (Phase 1 — new shared layer):**
- Create: `frontend/apps/web/src/lib/api/client.ts`
- Create: `frontend/apps/web/src/lib/api/errors.ts`
- Create: `frontend/apps/web/src/lib/api/errorMessages.ts`
- Create: `frontend/apps/web/src/lib/api/authBus.ts`
- Modify: auth provider component (locate via `grep -rn "useAuthSession\|AuthSession" frontend/apps/web/src/features/auth/`)
- Modify: login page component to honor `auth:returnTo` from sessionStorage

**Files (Phases 2–5 — feature migrations):**
- `frontend/apps/web/src/features/documents/v2/**/*.{ts,tsx}` — every `fetch()` call
- `frontend/apps/web/src/features/registry/**/*.{ts,tsx}`
- `frontend/apps/web/src/features/approval/api/mutationClient.ts` — refactor to delegate to `apiFetch` for auth/error
- `frontend/apps/web/src/features/iam/**/*.{ts,tsx}`

**Problem:** `mutate()` handles 401 in approval module only. Other features bypass it → generic errors or silent failures on session expiry.

**Fix:**

1. `client.ts` — single entry point:
```ts
export class ApiError extends Error {
  constructor(public code: string, public status: number, message: string, public details?: unknown) {
    super(message); this.name = 'ApiError';
  }
}

export async function apiFetch<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (res.status === 401) {
    dispatchAuthExpired(window.location.pathname + window.location.search);
    throw new ApiError('authn.expired', 401, 'Sessão expirada');
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({})) as { error?: { code?: string; message?: string; details?: unknown } };
    throw new ApiError(
      body.error?.code ?? `http_${res.status}`,
      res.status,
      body.error?.message ?? 'Erro interno',
      body.error?.details,
    );
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
```

2. `authBus.ts`:
```ts
export function dispatchAuthExpired(returnTo: string) {
  window.dispatchEvent(new CustomEvent('auth:expired', { detail: { returnTo } }));
}
```

3. Auth provider listener:
```ts
useEffect(() => {
  const handler = (e: Event) => {
    const detail = (e as CustomEvent).detail as { returnTo: string };
    if (detail.returnTo && detail.returnTo !== '/login') {
      sessionStorage.setItem('auth:returnTo', detail.returnTo);
    }
    logout();
    navigate('/login');
  };
  window.addEventListener('auth:expired', handler);
  return () => window.removeEventListener('auth:expired', handler);
}, [logout, navigate]);
```

4. Login page post-login:
```ts
const returnTo = sessionStorage.getItem('auth:returnTo') ?? '/';
sessionStorage.removeItem('auth:returnTo');
navigate(returnTo);
```

5. `mutationClient.ts` refactor: keep ETag/Idempotency-Key handling, delegate auth/error to `apiFetch`. `ApprovalError` becomes alias for `ApiError`.

6. Feature migrations: every `await fetch(url, ...)` becomes `await apiFetch<T>(url, ...)`. Inline 401 handling removed (now centralised). Inline error toasts use `resolveErrorMessage(err.code, err.message)`.

**Test:**
- Unit (vitest): `apiFetch` 401 dispatches event with `returnTo`
- Unit (vitest): `apiFetch` !ok throws `ApiError` with parsed code
- Integration (vitest + msw): trigger 401 from documents page action, assert event fired and listener navigates
- Smoke: log in, manually expire session in DB (`UPDATE auth_sessions SET expires_at = now() - interval '1h' WHERE user_id = '<uid>'`), click any documents action, observe redirect to `/login`, log in, land on original page

---

## Rollout Plan

| Phase | Tasks | Parallelism | Model |
|---|---|---|---|
| 0 | Worktree, codex spec validate, locate auth provider + login page | sequential | sonnet |
| 1 | Build `lib/api/` + auth listener + login returnTo | sequential | codex |
| 2 | Migrate `documents/v2` features | sequential | codex |
| 3 | Migrate `registry` ‖ Migrate `iam` | parallel | sonnet ‖ sonnet |
| 4 | Refactor `approval/api/mutationClient.ts` to delegate to `apiFetch`; verify regressions | sequential | sonnet |
| 5 | Verify: vitest + smoke (expiry + return-to), codex audit | sequential | sonnet → codex audit |
| 6 | Merge via `finishing-a-development-branch`, audit doc closure, wiki-curator | sequential | sonnet |

**Phase review after each:** Opus.

---

## Testing Strategy

**Per-bug:** see "Per-Bug Fix Design".

**Cross-cutting:**
- `npx vitest run` — full pass
- `go test -mod=mod ./...` — regression check (zero backend changes expected)
- Smoke flow: login → DB-expire session → page action → redirect → re-login → returnTo restored
- Smoke flow: submitter approves own doc → specific SoD toast
- Smoke flow: finalize without route → specific route toast
- Codex independent audit per-bug PASS/FAIL with file:line evidence
- Wiki-curator: refresh stamps on `wiki/modules/frontend-*.md` (if exists), create `wiki/concepts/error-ux.md`

**Coverage targets:** new `lib/api/` ≥85% line. No new lint warnings.

---

## Acceptance Criteria

- [ ] E2: submitter clicks Approve → toast contains "submeteu este documento"
- [ ] E3: finalize without route → toast contains "Nenhuma rota de aprovação"
- [ ] E4: any auth-critical page (documents/v2, registry, iam, approval) hits 401 → redirected to `/login`, post-login lands on original page
- [ ] `apiFetch` is the single mutation entry point in migrated features (zero raw `fetch` for mutations in those files)
- [ ] All vitest pass
- [ ] Backend `go test ./...` unchanged
- [ ] Codex audit returns 3/3 PASS
- [ ] Audit doc updated, E2/E3/E4 closed with commit SHAs
- [ ] Follow-up tasks documented for non-migrated features

---

## Open Questions

None.

---

## References

- Audit: `wiki/bugs/audit-2026-05-03.md` (E2 line 216, E3 line 217, E4 line 218)
- Backend error contract: `internal/modules/documents/approval/http/errors.go`, `internal/modules/documents/approval/http/contracts/errors.go`
- Existing partial helper: `frontend/apps/web/src/features/approval/api/mutationClient.ts`
- Sub-plan 2 (editor): `docs/superpowers/specs/2026-05-03-group-e-editor-design.md` (TBD)
- Sub-plan 3 (misc): `docs/superpowers/specs/2026-05-03-group-e-misc-design.md` (TBD)
