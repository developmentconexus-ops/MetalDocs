# Plan 12.1 Templates List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finalize the Templates list screen without fake data and close or preserve its deferred backlog items honestly.

**Architecture:** This is one screen PR for `/templates`. The fresh implementation session must run `metaldocs-screen-implementation` Phase 0 before code changes, then implement only confirmed Keep items in the existing `features/templates` and shared `components/ui` structure.

**Tech Stack:** React 18, TypeScript, TanStack Query, CSS Modules, Vitest, Playwright/manual Vite smoke.

---

## Required Fresh-Session Skills

- `metaldocs-frontend`
- `metaldocs-screen-implementation`
- `metaldocs-tanstack-query` if any query/API wrapper changes are made
- `verification-before-completion`

Do not implement this screen in the planning session. Start a fresh clean implementation session for the PR.

## Source Files To Read First

- Global spec: `docs/superpowers/specs/2026-05-13-plan-12-screens.md`
- Roadmap: `wiki/backlog/roadmap.md`
- Frontend architecture: `wiki/architecture/frontend-structure.md`
- API contract: `wiki/architecture/api-contract.md`
- Audit rules: `wiki/concepts/design-workflow-audit.md`
- Screen backlog: `wiki/backlog/templates.md`
- Design notes: `frontend/apps/web/design-source/templates/NOTES.md`
- Design HTML: `frontend/apps/web/design-source/templates/templates.html`

## Current Evidence From Planning

- Design source exists at `frontend/apps/web/design-source/templates/`.
- Backlog lists 5 deferred items.
- `frontend/apps/web/src/lib/api-types/index.d.ts` already contains `updated_at?: string` for a template list shape.
- `frontend/apps/web/src/features/templates/api/templatesV2.ts` has `TemplateListRow.updated_at?: string`, but `TemplateDTO` lacks `updated_at`.
- `TemplatesListPage.tsx` still renders `updated: formatRelative(dto.created_at)`.
- `TabBar.module.css` has no horizontal overflow handling.
- `NOTES.md` already cuts profile pills, the review tab, and card description text.

## Candidate Keep/Cut/Defer Before Phase 0

These are planning hypotheses only. Phase 0 must confirm them against the design and real contracts before code changes.

| Backlog item | Likely decision | Reason |
|---|---|---|
| `updated_at` timestamp | Keep if the runtime list response exposes it; otherwise Defer | Generated types suggest it may now exist, but `TemplateDTO` and the page still use `created_at`. |
| `created_by` display name | Defer | Requires user lookup or API join; no existing source was identified in planning. |
| Card grid gap | Keep if a token-supported adjustment preserves visual parity | Current CSS uses `var(--sp-4)`; backlog says design delta was accepted unless a token exists. |
| Mobile tab clipping | Keep | `TabBar` can add horizontal scrolling without backend support. |
| `formatRelative` promotion | Defer unless a second real caller exists in the fresh session | Backlog says promote only when a second caller appears. |

## File Structure

- Modify: `frontend/apps/web/design-source/templates/NOTES.md`
  - Append or update Plan 12.1 Phase 0 Keep/Cut/Defer findings.
- Create/update: `frontend/apps/web/design-source/templates/IMPLEMENTATION.md`
  - Copy from the screen implementation template if missing.
- Create/update artifacts under: `frontend/apps/web/design-source/templates/artifacts/`
  - `phase0-audit.md`
  - `phase1-map.md`
  - `phase4-behavior.md`
  - `phase4-review.md`
  - screenshot files under `artifacts/screenshots/`
- Modify only if Phase 0 confirms Keep:
  - `frontend/apps/web/src/features/templates/TemplatesListPage.tsx`
  - `frontend/apps/web/src/features/templates/api/templatesV2.ts`
  - `frontend/apps/web/src/components/ui/TabBar.module.css`
  - focused tests added near the touched code
- Modify: `wiki/backlog/templates.md`
  - Close completed items and preserve deferred items with reason.

## Task 1: Phase 0 Audit And Artifact Setup

**Files:**
- Modify: `frontend/apps/web/design-source/templates/NOTES.md`
- Create/update: `frontend/apps/web/design-source/templates/IMPLEMENTATION.md`
- Create: `frontend/apps/web/design-source/templates/artifacts/phase0-audit.md`

- [ ] **Step 1: Confirm design-source files exist**

Run:

```powershell
Get-ChildItem -LiteralPath frontend/apps/web/design-source/templates -Force
```

Expected:

- `NOTES.md`
- `templates.html`
- `artifacts/`

- [ ] **Step 2: Run Phase 0 audit**

Read the screen notes and backlog, then classify each visible UI element:

```powershell
Get-Content -LiteralPath frontend/apps/web/design-source/templates/NOTES.md
Get-Content -LiteralPath wiki/backlog/templates.md
```

Record these sections in `frontend/apps/web/design-source/templates/artifacts/phase0-audit.md`:

```markdown
# Phase 0 Audit - Templates

## Keep

- Header, subtitle, and "Novo template" action if route/action already exists.
- Tab bar statuses that map to real statuses: all, published, draft, archived.
- Card grid fields backed by real API response fields.
- Loading, empty, and error states.

## Cut

- Profile pills, unless an existing endpoint returns bound profile associations.
- Review tab, unless the current template list domain exposes a review status for this screen.
- Description text, unless list response contains real descriptions.

## Defer

- Created-by display name if no existing user lookup or eager list field exists.
- Any timestamp field not present in the actual list response.
- Any visual element that would require mock counts or hardcoded backend state.

## User Checkpoint

- Await user confirmation before TSX/CSS edits.
```

- [ ] **Step 3: Update `NOTES.md` with the Phase 0 result**

Append a dated `Plan 12.1 audit` section that mirrors the Keep/Cut/Defer decisions from `phase0-audit.md`.

Expected:

- `NOTES.md` names the backend dependency for every Defer item.
- No implementation files are changed before this point.

## Task 2: Phase 1 Map And Scope Lock

**Files:**
- Create: `frontend/apps/web/design-source/templates/artifacts/phase1-map.md`

- [ ] **Step 1: Map code placement**

Record this placement map in `phase1-map.md`, updating it only with evidence found in the fresh session:

```markdown
# Phase 1 Map - Templates

## Route

- `/templates`
- Route page: `frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx`
- Page: `frontend/apps/web/src/features/templates/TemplatesListPage.tsx`

## Data

- Query hook: `frontend/apps/web/src/features/templates/queries/useTemplatesQuery.ts`
- API wrapper: `frontend/apps/web/src/features/templates/api/templatesV2.ts`
- Query key: `QK.templates.list()`

## UI

- Feature card: `frontend/apps/web/src/features/templates/components/TemplateCard.tsx`
- Shared tabs: `frontend/apps/web/src/components/ui/TabBar.tsx`
- Shared tab styles: `frontend/apps/web/src/components/ui/TabBar.module.css`

## Confirmed Keep Items

- Copy the exact Keep bullets from `phase0-audit.md`. If this section would be empty, write `None; stop before implementation because this PR has no supported Keep items.`

## Confirmed Defer Items

- Copy the exact Defer bullets from `phase0-audit.md`, including the missing endpoint, field, or role contract for each item. If this section would be empty, write `None.`

## Tier

- Light if only timestamp wiring and TabBar CSS are changed.
- Heavy if Phase 0 requires new responsive layout, new shared primitive, or form/input work.
```

- [ ] **Step 2: Stop if backend shape is ambiguous**

If the implementation would require guessing whether `updated_at`, display names, descriptions, or profile bindings exist in the real list response, stop and report:

```text
BLOCKER: Templates list requires <field/endpoint>, but no contract evidence exists.
Files checked:
- wiki/backlog/templates.md
- frontend/apps/web/src/features/templates/api/templatesV2.ts
- frontend/apps/web/src/lib/api-types/index.d.ts
```

## Task 3: Implement Confirmed Local Keeps Only

**Files:**
- Potential modify after Phase 0: `frontend/apps/web/src/features/templates/TemplatesListPage.tsx`
- Potential modify after Phase 0: `frontend/apps/web/src/features/templates/api/templatesV2.ts`
- Potential modify after Phase 0: `frontend/apps/web/src/components/ui/TabBar.module.css`
- Test: add a focused test for any behavior changed

- [ ] **Step 1: Add a test for timestamp preference if `updated_at` is confirmed**

If the real contract exposes `updated_at`, add a focused test that verifies the page prefers `updated_at` over `created_at`.

Test target:

```text
frontend/apps/web/src/features/templates/__tests__/TemplatesListPage.test.tsx
```

Test structure:

```tsx
it("uses updated_at for the relative timestamp when the list response provides it", async () => {
  vi.setSystemTime(new Date("2026-05-13T12:00:00Z"));
  vi.mocked(listTemplates).mockResolvedValueOnce({
    templates: [
      {
        id: "tpl-1",
        tenant_id: "tenant-1",
        doc_type_code: "POP",
        key: "pop-001",
        name: "Procedimento Operacional",
        description: null,
        areas: [],
        visibility: "public",
        specific_areas: [],
        latest_version: 3,
        published_version_id: "ver-3",
        created_by: "Ana Silva",
        created_at: "2026-04-01T12:00:00Z",
        updated_at: "2026-05-12T12:00:00Z",
        archived_at: null,
      },
    ],
    meta: { limit: 50, offset: 0 },
  });

  render(
    <QueryClientProvider client={new QueryClient()}>
      <TemplatesListPage onOpenTemplate={vi.fn()} onCreate={vi.fn()} />
    </QueryClientProvider>,
  );

  expect(await screen.findByText("ontem")).toBeInTheDocument();
});
```

Run:

```powershell
cd frontend/apps/web
pnpm test -- TemplatesListPage
```

Expected before implementation:

- The test fails because the page uses `created_at`.

- [ ] **Step 2: Wire the minimal timestamp fix if the test was added**

Update `TemplateDTO` and mapping only if the contract supports `updated_at`:

```ts
updated_at?: string;
```

```ts
updated: formatRelative(dto.updated_at ?? dto.created_at),
```

Run:

```powershell
cd frontend/apps/web
pnpm test -- TemplatesListPage
```

Expected:

- The timestamp test passes.

- [ ] **Step 3: Add mobile tab overflow if Phase 0 keeps the clipping fix**

Update `frontend/apps/web/src/components/ui/TabBar.module.css`:

```css
.bar {
  display: flex;
  gap: 0;
  overflow-x: auto;
  scrollbar-width: none;
  border-bottom: 1px solid var(--border);
}

.bar::-webkit-scrollbar {
  display: none;
}

.tab {
  flex: 0 0 auto;
}
```

Run the focused component test suite if a TabBar test is added, otherwise verify through screenshot/smoke in Task 5.

- [ ] **Step 4: Avoid unsupported backlog items**

Do not implement these unless Phase 0 finds real contract evidence:

- `created_by` to display name.
- Description text.
- Bound profile pills.
- New backend fields or endpoints.

## Task 4: Backlog And Notes Update

**Files:**
- Modify: `wiki/backlog/templates.md`
- Modify: `frontend/apps/web/design-source/templates/NOTES.md`

- [ ] **Step 1: Close only completed backlog items**

For each implemented item, change its checklist row to closed and include a short evidence note:

```markdown
- [x] Mobile tab clipping at 375px resolved in `frontend/apps/web/src/components/ui/TabBar.module.css` by horizontal overflow; verified by screenshot at 375px.
```

- [ ] **Step 2: Preserve deferred items**

For each deferred item, keep it unchecked and add a reason:

```markdown
- [ ] Resolve `created_by` user_id -> display name. Deferred in Plan 12.1 because no existing user lookup, batch endpoint, or eager `created_by_display_name` field is available.
```

Expected:

- No backlog item is closed without code or contract evidence.

## Task 5: Verification And Review

**Files:**
- Create/update: `frontend/apps/web/design-source/templates/artifacts/phase4-behavior.md`
- Create/update: `frontend/apps/web/design-source/templates/artifacts/phase4-review.md`
- Create screenshots under: `frontend/apps/web/design-source/templates/artifacts/screenshots/`

- [ ] **Step 1: Run frontend checks**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Expected:

- TypeScript exits `0`.
- Vitest exits `0` or reports only pre-existing skipped tests.

- [ ] **Step 2: Start the app and smoke `/templates`**

Run API from repo root:

```powershell
.\scripts\start-api.ps1
```

Run frontend:

```powershell
cd frontend/apps/web
pnpm dev
```

Smoke:

- Log in with the local credentials from `wiki/references/local-dev-credentials.md`.
- Open `/templates`.
- Check loading, list, tab filtering, empty tab behavior, and create navigation.
- Capture desktop and 375px mobile screenshots.

- [ ] **Step 3: Run frontend-screen-reviewer**

Use the reviewer required by `metaldocs-screen-implementation` with:

```text
slug: templates
page path: /templates
worksheet: frontend/apps/web/design-source/templates/IMPLEMENTATION.md
screenshots: frontend/apps/web/design-source/templates/artifacts/screenshots/
```

Expected:

- No unresolved Critical or Major findings.
- Minor findings are either fixed or preserved in `wiki/backlog/templates.md`.

## Task 6: Commit Screen PR

**Files:**
- Stage only files touched for the Templates screen PR.

- [ ] **Step 1: Review final diff**

Run:

```powershell
git diff --check
git status --short
git diff -- frontend/apps/web/src/features/templates frontend/apps/web/src/components/ui/TabBar.module.css frontend/apps/web/design-source/templates wiki/backlog/templates.md
```

Expected:

- No unrelated files.
- No mock data paths introduced.
- `NOTES.md`, artifacts, and backlog match implementation.

- [ ] **Step 2: Commit**

Run:

```powershell
git add -- frontend/apps/web/src/features/templates frontend/apps/web/src/components/ui/TabBar.module.css frontend/apps/web/design-source/templates wiki/backlog/templates.md
git commit -m "feat: finalize templates list screen"
```

Expected:

- Commit contains only the Templates screen PR changes.
