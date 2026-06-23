# Feature F4.2 — Plan

> The "how" for `spec.md`. TDD-ordered. Frontend-only; all changes in `DocumentPublishedPage.tsx` +
> `DocumentPublishedPage.module.css` + co-located test. No new page file (reuse is an acceptance criterion).

## Files touched

- `src/features/documents/pages/DocumentPublishedPage.tsx` — obsolete branch: root dim class, hide vigente
  pill, capability-gated Visualizar, disable Baixar PDF + Copiar link when obsolete.
- `src/features/documents/pages/DocumentPublishedPage.module.css` — add `.rootObsolete` (grayscale+opacity).
- `src/features/documents/pages/DocumentPublishedPage.test.tsx` — new `describe('F4.2 — obsolete variant')`.

## Reuse (no new code where it exists)

- `useHasCapability('document.obsolete')` — `features/iam/hooks/useHasCapability.ts` (reads `user.capabilities`).
- Existing `isObsolete` derived flag (`status === 'obsolete'`) + `.obsoleteBanner`/`.obsoleteStamp` watermark
  (already a faithful design port — unchanged).
- Existing `getDocumentStatusPresentation` `'obsolete'` case — unchanged.
- Reason-copy shape from `features/templates/lib/canActOnVersion.ts` (`Sua sessão não inclui a capacidade …`).

## Tasks (TDD order)

1. **RED — tests first.** Add `describe('F4.2 — obsolete variant')` to `DocumentPublishedPage.test.tsx`:
   - obsolete fixture (`status:'obsolete'`) ⇒ `OBSOLETO` text present; obsolete status subtitle present.
   - root carries the obsolete dim class when obsolete; not when published.
   - "vigente" text absent when obsolete; present when published.
   - obsolete + `useHasCapability` → true (caps `['document.obsolete']`) ⇒ Visualizar **enabled**.
   - obsolete + `useHasCapability` → false (caps `[]`) ⇒ Visualizar **disabled**.
   - obsolete ⇒ "Baixar PDF" and "Copiar link" buttons `disabled`.
   - Mock `useHasCapability` (`vi.mock('../../iam/hooks/useHasCapability')`), default to a controllable
     return; seed `useAuthStore` user with `capabilities` as needed. Verify the published-state tests
     (no obsolete class, vigente present, Visualizar enabled) still hold as the negative control.
2. **Derive the gate.** In `DocumentPublishedPage`, near `isObsolete`, add
   `const canViewObsolete = useHasCapability('document.obsolete');` (hook called unconditionally at top
   level — preserve hooks order; it lives with the other top-level hook calls, not inside the obsolete
   branch).
3. **Root dim.** Add `.rootObsolete { filter: grayscale(0.65); opacity: 0.85; }` to the module CSS; apply
   `className={`${styles.root}${isObsolete ? ' ' + styles.rootObsolete : ''}`}` (or `clsx`-style join if
   already used in the file).
4. **Hide vigente pill.** Wrap the `vigenteBadge` span in `{!isObsolete && ( … )}`.
5. **Gate Visualizar.** Add `disabled={isObsolete && !canViewObsolete}` to the Visualizar button; when
   disabled-for-obsolete, set a `title` naming the missing capability (reason-copy shape from
   `canActOnVersion`); keep its `onClick={handleView}` (no-op while disabled).
6. **Disable mutating actions when obsolete.** Baixar PDF: extend `disabled` to
   `… || isObsolete`. Copiar link: add `disabled={isObsolete}`. (Iniciar revisão/Publicar already
   aria-disabled for obsolete status — leave; no regression.)
7. **GREEN + static.** Run the page + formatter suites to green; `npx tsc --noEmit` exit 0.
8. **Review pass.** Dispatch `frontend-screen-reviewer` (visual parity vs `design-source/documento-obsoleto`)
   then `frontend-code-reviewer`; resolve findings by root-cause family; both APPROVE.

## Test strategy

Fixture-level vitest (mocked `useDocumentDetailQuery` + mocked `useHasCapability` + seeded `useAuthStore`).
Proves the obsolete branch + capability-gate wiring, not backend enforcement (backend is the real boundary
per `wiki/concepts/authz-tiers.md`). Published-state assertions act as the negative control (treatment
applies only when obsolete). Reuse + tsc are real checks.

## Ordering / risk

- Hooks order (R: conditional hook) — `useHasCapability` is added to the top-level hook block, never inside
  the obsolete branch.
- No-fork (R2) — zero new files under `pages/`; the validator/reviewer assert reuse.
- Scope drift (HS-6) — no new capability, no backend, no restyle beyond the obsolete treatment.
