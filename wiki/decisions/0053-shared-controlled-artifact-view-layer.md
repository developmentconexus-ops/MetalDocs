# ADR 0053 — Shared controlled-artifact frontend view layer (one presentational shell, per-kind adapters)

- **Status:** Accepted
- **Last verified:** 2026-06-30
- **Date:** 2026-06-30
- **Scope:** Establishes a single, kind-agnostic frontend view layer that renders both controlled **documents** and controlled **templates** through one set of purely presentational components. The two kinds converge on a normalized `ArtifactViewModel` produced by per-kind **adapters**; thin **route wrappers** own all kind-specific state, dialogs, and navigation. Also mandates that clicking a template opens a **detail screen** (parity with documents), not the eigenpal editor directly.
- **Depends on:** ADR 0052 — template lifecycle parity (manual versioning + `in_review`→`under_review`). Without a shared lifecycle semantic the shell would have to branch on kind, which this decision forbids.

---

## Context

Documents and templates are both **governed artifacts** and were converging on the same screens — a detail/overview page, an approval decision screen, a metadata sidebar, a hero bar, a revision history. Yet each kind had grown its own implementation:

- Documents rendered through `ControlledDocumentDetailPanel` and a document-specific approval surface.
- Templates rendered approval through an inline `VersionActionPanel`, and **clicking a template jumped straight into the eigenpal editor** — there was no template detail screen at all.
- The hero doc-card, badges, sidebar, and approval chain were inlined and duplicated across the two kinds, and had already begun to drift (divergent labels, one surface reaching across into another's CSS module).

Two forces made this untenable. First, **duplication guarantees drift**: every parity fix had to be made twice and inevitably wasn't. Second, the missing template detail screen broke the mental model — a template is a governed artifact and should be *inspected* before being *edited*, exactly as a document is.

The wrong fix would have been to merge the backend modules or to add a `kind` switch inside a single mega-component. The former violates the bounded-context boundary (templates and documents stay distinct modules); the latter just relocates the duplication into conditionals. The right fix is a **presentational shell + data-shaping adapters** boundary.

---

## Decision

### 1. `ArtifactViewModel` is the single boundary

`features/shared/controlled-artifact/types.ts` defines a kind-agnostic, fully-normalized view-model (`ArtifactViewModel`) plus its supporting types (`ArtifactHeroModel`, `ArtifactMetaModel`, `ArtifactKpiCell`, `ApprovalChainItem`, `VersionHistoryItem`, `ArtifactTab`, `ArtifactAction`). Template-divergent fields (`code`, `subtitle`, `areaLabel`, `fileSizeBytes`, `pageCount`, `effectiveFrom`, `nextReviewAt`, `approvalChain`) are **nullable** so the shell renders a reduced surface for templates **without branching on kind**. The lifecycle vocabulary aliases the canonical `DocumentStatus` union (`LifecycleStatus`) rather than forking a parallel one.

### 2. The shell is purely presentational — four hard rules

Every component under `features/shared/controlled-artifact/` (`ArtifactDetailView`, `ArtifactDetailLayout`, `ArtifactHero`, `ArtifactHeroCard`, `ArtifactMetaSidebar`, `ArtifactApprovalScreen`, `VersionTimeline`) obeys, without exception:

1. **No data fetching** — no react-query hooks, no API imports, no `useParams`. It receives a finished `ArtifactViewModel`.
2. **No `kind` branching** — the component never inspects `model.kind` to choose layout. Divergence is expressed as data (null fields, empty arrays, ordered `actions`), never as conditionals.
3. **No cross-feature imports** — the shell imports only from `shared/` and shared primitives; it never reaches into `documents/`, `templates/`, or `approval/`, and no component imports another component's CSS module.
4. **Composition via `ReactNode` slots** — kind-specific chrome (hero action buttons, the document coverage aside, error banners) is injected by the route as `heroActions` / `aside` / `extras` slots. The shell owns none of that behavior.

### 3. Per-kind adapters compose the model

Adapters are the *only* place kind-specific API shapes and business rules live:

- Detail: `documents/adapters/useDocumentArtifact.ts`, `templates/adapters/useTemplateArtifact.ts`.
- Approval: `documents/adapters/useDocumentApprovalArtifact.ts`, `templates/adapters/useTemplateApprovalArtifact.ts`.

Each fetches kind-specific data and maps it to `ArtifactViewModel`, keeping every kind-specific rule (the document scheduled→published-head "current version" override, live coverage denominators, template revision labels, the ordered `ArtifactAction[]`) out of the shell.

### 4. Thin route wrappers own state, dialogs, and navigation

Route components stay thin: they call one adapter, own kind-specific `useState` + dialogs, and pass slots into the shell. `TemplateDetailRoute`, `TemplateApprovalRoute`, and the document approval route (`approval/pages/SignoffDetailPage`) are the wrappers. Workflow writes live in route-owned dialogs (e.g. `CancelInstanceDialog`, `SupersedePublishDialog`), each of which owns its own mutation + error state; the shell only fires `action.run()`.

### 5. Cross-kind duplication is deduplicated to one home

Logic shared by more than one adapter is extracted, not copied: `shared/controlled-artifact/resolveOwnerDisplay.ts` (owner-name resolution with current-user fallback) and `documents/lib/approvalWorkflow.ts#mapApprovalChain` (stage→`ApprovalChainItem` mapping).

### 6. Templates get a detail screen

Clicking a template in the list now routes to `TemplateDetailRoute` (the shared `ArtifactDetailView`), matching documents. The eigenpal editor is reached only via an explicit "Editar modelo" / "Ver no editor" action from that detail screen — a template is inspected before it is edited.

---

## Consequences

### Positive
- **Parity by construction.** Both kinds render through the same shell; a fix to the hero, sidebar, or approval screen lands for documents and templates simultaneously. Drift is structurally impossible, not merely discouraged.
- **Isolated, testable seams.** Adapters are plain hooks tested against fixtures; the shell is tested with hand-built view-models. Neither needs the other's runtime.
- **Enforceable purity.** The four rules give reviewers a bright-line test. The post-implementation review confirmed the shell components carry zero data-fetching, zero `useParams`, and zero `kind` branching.
- **No module merge.** The bounded-context boundary is untouched — templates and documents remain distinct backend modules; convergence is a frontend view concern only.

### Negative / trade-offs
- **An indirection layer.** Rendering now goes route → adapter → view-model → shell instead of a component reading its own query. The `ArtifactViewModel` boundary is one more artifact to keep in sync when a genuinely new field is added.
- **Nullability discipline.** Because divergence is modeled as nullable fields rather than kind branches, adapters must be careful to populate or null every field intentionally; a forgotten field renders as an em-dash rather than a type error.
- **Slot contracts are implicit.** Kind-specific chrome passed as `ReactNode` slots is not type-checked against "what belongs here" — the route and shell agree by convention on what `heroActions`/`aside`/`extras` mean.

---

## References

- ADR 0052 — template manual versioning + `in_review`→`under_review` (the lifecycle-parity precondition this layer consumes).
- ADR 0022 — AuthZ = capabilities, never roles (unchanged; the shell renders actions the adapter already authorized, it does not itself gate).
- `frontend/apps/web/src/features/shared/controlled-artifact/types.ts` — the `ArtifactViewModel` contract.
- Shell components — `ArtifactDetailView.tsx`, `ArtifactDetailLayout.tsx`, `ArtifactHero.tsx`, `ArtifactHeroCard.tsx`, `ArtifactMetaSidebar.tsx`, `ArtifactApprovalScreen.tsx`, `VersionTimeline.tsx` (all under the same directory).
- Adapters — `documents/adapters/useDocumentArtifact.ts`, `documents/adapters/useDocumentApprovalArtifact.ts`, `templates/adapters/useTemplateArtifact.ts`, `templates/adapters/useTemplateApprovalArtifact.ts`.
- Route wrappers — `templates/pages/TemplateDetailRoute.tsx`, `templates/pages/TemplateApprovalRoute.tsx`, `approval/pages/SignoffDetailPage.tsx`.
- Dedup helpers — `shared/controlled-artifact/resolveOwnerDisplay.ts`, `documents/lib/approvalWorkflow.ts` (`mapApprovalChain`).
- `wiki/architecture/frontend-structure.md` — the canonical frontend architecture doc (updated alongside this ADR).
