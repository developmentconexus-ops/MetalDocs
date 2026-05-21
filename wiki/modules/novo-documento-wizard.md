# Novo-Documento Wizard

> **Last verified:** 2026-05-15 (blank-template create runtime smoke)
> **Scope:** 4-step document-creation wizard at `/documents/new` Ã¢â‚¬â€ state machine, step components, new primitives (`DocPaperPreview`, `WizardFooter`), sub-controls, and helpers (`resolveQueryError`, `STALE_FIVE_MINUTES`, `QK.templates.byProfile`).

> **Maturity:** L2
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:43` Ã¢â‚¬â€ wizard entry point; `useReducer(wizardReducer)` + URL step sync + `useMutation` submit
> - `frontend/apps/web/src/features/documents/state/wizard.reducer.ts:62` Ã¢â‚¬â€ `wizardReducer` Ã¢â‚¬â€ pure reducer; `selectProfile` clamps step; `clampStep` / `maxReachableStep` / `canAdvance` exported
> - `frontend/apps/web/src/features/documents/state/__tests__/wizard.reducer.test.ts:1` Ã¢â‚¬â€ vitest unit suite for reducer + helpers
> - `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:1` Ã¢â‚¬â€ doc-wizard adapter wrapping `features/shared/components/wizard/WizardShell`; supplies hardcoded 4-step copy and converts `WizardStepNumber` (1Ã¢â‚¬â€œ4) to Stepper string IDs
> - `frontend/apps/web/src/features/shared/components/wizard/WizardFooter.tsx:16` Ã¢â‚¬â€ shared footer row (Back/Cancel + primary CTA); promoted from documents wizard in Phase 2; used by both document and template wizards
> - `frontend/apps/web/src/features/documents/components/wizard/DocPaperPreview.tsx:18` Ã¢â‚¬â€ decorative paper-preview tile; `lines`, `code`, `variant` props
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepProfile.tsx:1` Ã¢â‚¬â€ Step 1: profile radio cards
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/index.tsx:40` Ã¢â‚¬â€ Step 2: area + title + visibility; delegates to `PeopleSubcontrols` / `ExternalSubcontrols`
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/PeopleSubcontrols.tsx:12` Ã¢â‚¬â€ disabled people-sharing sub-control (deferred)
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/ExternalSubcontrols.tsx:11` Ã¢â‚¬â€ disabled external-sharing sub-control (deferred)
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:1` Ã¢â‚¬â€ Step 3: template selector; filters to `published_version_id` only
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:1` Ã¢â‚¬â€ Step 4: read-only summary + consent + submit
> - `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:1` Ã¢â‚¬â€ `{profile}-{area}-???` preview banner
> - `frontend/apps/web/src/features/taxonomy/queries/useProfilesQuery.ts:7` Ã¢â‚¬â€ profiles query; `STALE_FIVE_MINUTES`; moved from documents queries in Phase 5 (template wizard reuse)
> - `frontend/apps/web/src/features/documents/queries/useAreasQuery.ts:1` Ã¢â‚¬â€ areas query; `STALE_FIVE_MINUTES`
> - `frontend/apps/web/src/features/documents/queries/useTemplatesByProfileQuery.ts:6` Ã¢â‚¬â€ templates by profile; `QK.templates.byProfile`
> - `frontend/apps/web/src/features/documents/queries/_constants.ts:2` Ã¢â‚¬â€ `STALE_FIVE_MINUTES = 5 * 60 * 1000`
> - `frontend/apps/web/src/lib/api/resolveQueryError.ts:10` Ã¢â‚¬â€ `resolveQueryError(err, fallback)` helper
> - `frontend/apps/web/src/lib/api/index.ts:3` Ã¢â‚¬â€ re-exports `resolveQueryError`
> - `frontend/apps/web/src/lib/queryKeys.ts:51` Ã¢â‚¬â€ `QK.templates.byProfile(profileCode)` key

---

## Overview

The 4-step wizard replaces the old `DocumentCreatePage` single-step flow. It creates the controlled-document slot and first draft revision in a **single atomic call**:

`POST /api/v1/controlled-documents` (`Idempotency-Key` required) Ã¢â‚¬â€ inserts the CD row, increments the per-(profile, area) sequence counter, and clones the template into the first draft document revision, all within a single DB transaction. Returns the CD with server-resolved code (e.g. `DC-RH-001`) plus the new document ID.

The blank-template path is a real runtime path, not fake frontend data: the wizard resolves `GET /api/v1/templates/system/blank`, uses the returned `templateId` and `templateVersionId`, and submits that version ID through the same controlled-documents-owned atomic create endpoint.

All wizard form state lives in a single `useReducer(wizardReducer)` call inside `NewDocumentWizardPage`. Step is mirrored to `?step=1..4` in the URL (with `replace: true`) so the browser back button works.

---

## Steps

| Step | Component | Server call | Gate to advance |
|------|-----------|-------------|-----------------|
| 1 Ã¢â‚¬â€ Profile | `StepProfile` | `GET /api/v1/taxonomy/profiles` | `profileCode !== null` |
| 2 Ã¢â‚¬â€ Area + Title + Visibility | `StepAreaCodeVisibility` | `GET /api/v1/taxonomy/areas` | `areaCode !== ''` and `title.trim() !== ''` |
| 3 Ã¢â‚¬â€ Template | `StepTemplate` | `GET /api/v1/templates?profileCode=Ã¢â‚¬Â¦` | `templateVersionID !== null` |
| 4 Ã¢â‚¬â€ Confirm + Create | `StepConfirm` | `POST /api/v1/controlled-documents` (atomic) | `consent && !submitting` |

---

## State machine (`wizard.reducer.ts`)

### `WizardState`

```typescript
// frontend/apps/web/src/features/documents/state/wizard.reducer.ts:16
type WizardState = {
  step: WizardStep;           // 1 | 2 | 3 | 4
  profileCode: string | null;
  areaCode: string;
  title: string;
  visibility: VisibilityKey;  // 'area' | 'people' | 'company' | 'external'
  invitees: WizardInvitee[];
  external: WizardExternalConfig;
  templateID: string | null;
  templateVersionID: string | null;
  consent: boolean;
  submitting: boolean;
  error: string | null;
};
```

### Key reducer behaviors

**`selectProfile` Ã¢â‚¬â€ step clamp on profile change.**
When the user changes their profile selection at any step > 2, the reducer resets `templateID` / `templateVersionID` and clamps `step` back to 2. This is self-contained Ã¢â‚¬â€ no effect in the page component needed.

```typescript
// wizard.reducer.ts:66Ã¢â‚¬â€œ83
case 'selectProfile': {
  if (!action.code.trim()) return state;   // guard: whitespace is not valid
  const nextStep: WizardStep = state.step > 2 ? 2 : state.step;
  return { ...state, step: nextStep, profileCode: action.code,
           templateID: null, templateVersionID: null, error: null };
}
```

**`clearProfile`** Ã¢â‚¬â€ resets to step 1, clears profile + template.

### Exported helpers

| Helper | Signature | Purpose |
|--------|-----------|---------|
| `maxReachableStep` | `(state) => WizardStep` | Highest step the user may jump to |
| `clampStep` | `(requested, state) => WizardStep` | Caps a requested step at `maxReachableStep` |
| `canAdvance` | `(state) => boolean` | Whether the primary CTA is enabled for current step |

`clampStep` is used in `NewDocumentWizardPage` to validate URL-seeded step at mount and stepper clicks.

---

## Submit flow (`NewDocumentWizardPage.tsx:111`)

```typescript
// Simplified Ã¢â‚¬â€ see NewDocumentWizardPage.tsx:111Ã¢â‚¬â€œ144
const createMutation = useMutation({
  mutationFn: async (input) => {
    return createControlledDocumentAtomic(
      { profileCode, processAreaCode, title, ownerUserId, documentName, templateVersionId },
      input.idempotencyKey,  // POST /api/v1/controlled-documents with Idempotency-Key header
    );
  },
  onError: (err) => {
    const message = resolveQueryError(err, 'Falha ao criar o documento.');
    dispatch({ type: 'submitError', message });
    toast.error(message);
  },
});
```

The `idempotencyKey` is generated via `crypto.randomUUID()` in `handleCreate` (`NewDocumentWizardPage.tsx:160`) immediately before the mutation call. Replay of the same key returns the stored 201 response Ã¢â‚¬â€ safe to retry on network timeout. Orphan slots are structurally impossible (CD + document commit atomically or both roll back). See ADR 0011 and `backlog/novo-documento.md#slot-rollback`.

---

## `profileNotFound` derived flag

`NewDocumentWizardPage` derives this flag instead of dispatching from an effect:

```typescript
// NewDocumentWizardPage.tsx:89Ã¢â‚¬â€œ92
const profileNotFound =
  profilesQuery.isSuccess &&
  state.profileCode !== null &&
  selectedProfile === null;
```

When true (URL `?profile=X` but X is not in the loaded list), Step 1 renders an alert card with a "Limpar seleÃƒÂ§ÃƒÂ£o" button rather than silently resetting state.

---

## New primitives (major-findings refactor)

### `WizardFooter` (`features/shared/components/wizard/WizardFooter.tsx:16`)

Shared footer row extracted from step components. Renders the Back/Cancel button on the left and the primary CTA on the right. Uses `WizardShell.module.css` tokens (`footerDivider`, `footerRow`).

```typescript
type WizardFooterProps = {
  stepLabel: string;
  primaryLabel?: string;          // default: 'AvanÃƒÂ§ar Ã¢â€ â€™'
  primaryDisabled?: boolean;
  primaryVariant?: 'advance' | 'submit';
  showBack?: boolean;             // false Ã¢â€ â€™ shows Cancel instead of Back
  onCancel?: () => void;
  onBack?: () => void;
  onAdvance?: () => void;
};
```

### `DocPaperPreview` (`wizard/DocPaperPreview.tsx:18`)

Decorative paper-document thumbnail used in Step 3 (template tiles) and Step 4 (confirm summary). Aria-hidden Ã¢â‚¬â€ purely visual. Line widths are deterministic (index-based formula) so no flicker between renders.

```typescript
type DocPaperPreviewProps = {
  lines: number;
  code?: string;       // optional label in top-right corner
  variant?: 'thumbnail' | 'template';  // default: 'thumbnail'
};
```

---

## Visibility sub-controls (`StepAreaCodeVisibility/`)

Step 2 renders conditional sub-controls based on the selected `VisibilityKey`:

| Key | Sub-control | Status |
|-----|------------|--------|
| `'people'` | `PeopleSubcontrols` Ã¢â‚¬â€ invitee list + add button | Rendered **disabled** (`aria-disabled="true"`) Ã¢â‚¬â€ deferred |
| `'external'` | `ExternalSubcontrols` Ã¢â‚¬â€ password/watermark/expiry | Rendered **disabled** (`aria-disabled="true"`) Ã¢â‚¬â€ deferred |
| `'area'` / `'company'` | none | No sub-control |

Both sub-controls are no-ops: callbacks are accepted but prefixed `_` to mark unused. Sharing/ACL endpoints do not exist yet. See `backlog/novo-documento.md#visibility`.

---

## Query hooks

### `useTemplatesByProfileQuery` (`queries/useTemplatesByProfileQuery.ts:6`)

```typescript
useQuery({
  queryKey: profileCode ? QK.templates.byProfile(profileCode) : QK.templates.list(),
  queryFn: () => listTemplates({ doc_type: profileCode! }),
  enabled: profileCode !== null,
  staleTime: STALE_FIVE_MINUTES,
})
```

When `profileCode` is null the query is disabled. The key uses `QK.templates.byProfile` (added in this refactor) so profile-scoped results can be invalidated independently.

### `STALE_FIVE_MINUTES` (`queries/_constants.ts:2`)

```typescript
export const STALE_FIVE_MINUTES = 5 * 60 * 1000;
```

Shared constant for slow-moving taxonomy and template lookups (profiles, areas, templates). Avoids per-call magic numbers.

---

## `resolveQueryError` (`lib/api/resolveQueryError.ts:10`)

```typescript
export function resolveQueryError(err: unknown, fallback: string): string {
  if (err instanceof ApiError) return resolveErrorMessage(err.code, err.message);
  if (err instanceof Error)    return err.message;
  return fallback;
}
```

Extracts the user-facing string from any TanStack Query error. Replaces the inline `ApiError` instanceof triad that was duplicated at every `onError` callsite. Re-exported from `lib/api/index.ts`. Used in `NewDocumentWizardPage` (`onError`) and `StepAreaCodeVisibility` (areas error inline alert).

---

## `QK.templates.byProfile` (`lib/queryKeys.ts:51`)

```typescript
templates: {
  list: () => ['templates', 'list'] as const,
  byProfile: (profileCode: string) => ['templates', 'by-profile', profileCode] as const,
},
```

Added to allow profile-scoped invalidation without busting the full templates list cache.

---

## Deferred items

See `wiki/backlog/novo-documento.md` for all deferred items. Short list (closed items omitted):

- `visibility` Ã¢â‚¬â€ selected value not submitted (no backend column)
- `template-versions` Ã¢â‚¬â€ no version picker
- `profile-counts` Ã¢â‚¬â€ no document count per profile card

Closed by feat/cd-atomic-create: `sequence-preview` (preview endpoint shipped), `slot-rollback` (atomic create eliminates orphan slots). See ADR 0011.
Closed by Novo Documento blank-template repair: `blank-template` (system blank endpoint supplies a valid `templateVersionId`; StepTemplate keeps the card selectable even when no profile templates exist).

---

## See also

- `wiki/modules/documents.md` Ã¢â‚¬â€ library + editor flow; wizard overview table
- `wiki/backlog/novo-documento.md` Ã¢â‚¬â€ all deferred wizard items
- `wiki/concepts/error-ux.md` Ã¢â‚¬â€ `ApiError`, `resolveErrorMessage`; `resolveQueryError` extends this pattern
- `wiki/concepts/placeholders.md` Ã¢â‚¬â€ 7-token catalog used inside documents created by this wizard

---

## Governance Links

- [novo-documento-wizard-tech-debt.md](novo-documento-wizard-tech-debt.md)
- [backlog/novo-documento-wizard-refactor.md](../backlog/novo-documento-wizard-refactor.md)


## 11. Risks & Technical Debt

- Critical: 0
- Major: 2
- Minor: 1

Refactor backlog: [../backlog/novo-documento-wizard-refactor.md](../backlog/novo-documento-wizard-refactor.md)
