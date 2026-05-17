# Caixa Aprovacao Reality-First Design

> **Date:** 2026-05-14
> **Scope:** Plan 12.2 implementation boundary for the `caixa-aprovacao` screen (`/approvals`) using the MetalDocs reality-first workflow.
> **Out of scope:** new backend endpoints, shared approval contract redesign, cross-screen refactors, and speculative inbox schema expansion.

## Goal

Finalize the `caixa-aprovacao` screen as a real product screen, not a design-era mock. Keep the designed experience where it can be supported by current runtime and contract truth, improve weak existing frontend implementation where the fix is screen-local, and explicitly preserve unsupported richness as backlog rather than fake UI.

## Current truth

Runtime and contract evidence verified on 2026-05-14:

- Screen route exists at `/approvals`.
- Inbox API exists at `GET /api/v1/approval/inbox`.
- Approval module contract surfaces are present in runtime, OpenAPI, generated backend code, generated frontend types, and the current feature wrapper.
- Registry detail remains the real review entrypoint at `/controlled-documents/:controlledDocumentId`.
- Registry exposes `GET /api/v1/controlled-documents/{id}/active-document`, which provides the active document and approval context needed by the existing signoff flow.
- Live inbox sample may legitimately return `items: []` and `total: 0`; therefore fake fallback data is not allowed.

Current inbox payload supports:

- `instance_id`
- `document_id`
- `controlled_document_id`
- `document_title`
- `area_code`
- `submitted_by`
- `submitted_at`
- `stage_label`
- `quorum_progress`

The payload does not currently prove support for:

- decision history heatmap data
- deadline or urgency
- code / kind
- summary
- changes count
- version transition

## Design decision

Keep both views from the designed screen:

- `Foco`
- `Linha do tempo`

But make both views fully honest and entirely driven by real data already available today.

This design deliberately rejects the prior mock-era strategy of enriching inbox items with fabricated fields or falling back to a fake inbox when the API is empty.

## Approach options considered

### Option 1 — recommended

Keep the full screen concept, but downgrade unsupported visual regions into honest real-data states and preserve backend-dependent richness in backlog.

Pros:

- preserves the screen
- stays truthful
- fits one-screen PR scope
- improves the existing weak implementation instead of freezing it

Cons:

- visual richness will be lower than the original mock in some regions

### Option 2

Expand backend support within this PR to preserve the original richness.

Pros:

- closest parity to the design

Cons:

- risks breaking the bounded one-screen scope
- turns this into shared contract work
- conflicts with the no-guessing rule unless every missing field can be proven from current module truth

### Option 3

Disable `Linha do tempo` until richer backend data exists.

Pros:

- simplest truthful implementation

Cons:

- removes a designed screen capability the user explicitly wants to keep

## Recommended design

Use Option 1.

Preserve the screen and improve the implementation, but strip out all fake runtime behavior. Unsupported richness remains explicitly documented as backend-dependent follow-up.

## Screen behavior

### Foco view

Keep `Foco` as the primary action view.

The queue rail and selected card must render only real fields:

- document title
- author
- area
- stage label
- submitted timestamp
- quorum progress

Remove mock-only live rendering for:

- code
- kind
- deadline
- urgency
- summary
- changes
- version

If the inbox is empty, render an honest empty state. Do not fall back to `MOCK_INBOX_ITEMS`.

### Linha do tempo view

Keep the tab and keep the timeline identity, but reinterpret it as a review stream rather than a deadline scheduler.

Timeline grouping should derive only from `submitted_at`, for example:

- `Hoje`
- `Ontem`
- `Últimos 7 dias`
- `Mais antigos`

Each row should render only real fields:

- title
- author
- area
- stage
- quorum progress
- submitted timestamp

Do not render fake urgency, countdown, or deadline-based grouping.

### Heatmap region

Keep the heatmap slot in the layout, but make it an explicit unavailable-data state, for example:

- `Histórico de decisões ainda não disponível`

No hardcoded bars. No simulated decision counts.

### Filters region

Keep the filter button disabled and explicitly deferred. The API supports `area_code`, but this PR does not expand the screen into an unaudited filter interaction flow.

## Actions

### Abrir documento

Wire to the real review entrypoint:

- `/controlled-documents/{controlled_document_id}`

### Aprovar e assinar

Use the existing approval flow:

1. Resolve the controlled document through the existing registry active-document endpoint.
2. Use returned `documentId`, `contentHash`, and `approvalInstanceId`.
3. Open the existing `SignoffDialog`.

### Devolver

Use the same entry flow as signoff, but open the dialog in reject mode.

This is screen-local because the backend mutation surface and the dialog already exist.

### Timeline row and `Revisar →`

Use the same real review entrypoint as the rest of the screen:

- navigate to `/controlled-documents/{controlled_document_id}`

This is preferred over bouncing back to the `Foco` card because it is simpler, consistent, and uses the established approval review path already documented in the repo.

## Implementation quality rules

This PR must improve weak current implementation when the fix is local to the screen.

That includes:

- removing fake inbox fallback
- removing mock enrichment from the real runtime path
- validating persisted `view` state as `'stack' | 'timeline'`
- replacing stubs with real action wiring
- aligning icon usage with the design where supported
- tightening tests around honest empty, real navigation, and review-flow entry behavior

This PR must not:

- invent unsupported backend fields
- silently preserve fake runtime values
- absorb shared contract redesign into the screen
- expand into unrelated approval or registry refactors

## Data flow

### Inbox read

- `useInboxQuery()` remains the source for `/approvals`
- query result drives both views
- empty API response renders empty UI, not fake content

### Signoff / reject entry

- screen action receives `controlled_document_id`
- fetch active-document context from registry
- if active document or approval context is missing, show an honest user-facing error state
- if context exists, open `SignoffDialog`

### Document open

- route directly to `/controlled-documents/{controlled_document_id}`

## Error handling

- empty inbox is a valid state
- inbox fetch failure remains a proper error state
- signoff/reject entry failure must surface an honest error, not silently no-op
- if a controlled document has no active review context, the screen must report that the approval action is unavailable rather than pretend it succeeded

## Testing strategy

Write tests before implementation for each behavior change:

- empty inbox renders honest empty state and not mock fallback content
- invalid persisted `view` falls back safely
- `Abrir documento` navigates to `/controlled-documents/{controlled_document_id}`
- timeline row and `Revisar →` navigate to the same route
- signoff/reject entry handles missing active-document context honestly

Verification before completion:

```powershell
cd frontend/apps/web
pnpm test -- InboxPage
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Add targeted screen smoke after the code changes.

## Backlog preserved after this PR

Keep the following as explicit follow-up work, not hidden debt:

- decision-history API for heatmap
- deadline and urgency inbox fields
- richer inbox metadata: code, kind, summary, changes, version
- approval API wrapper migration to the canonical API client when done as a bounded follow-up
- audited filter-panel behavior

## Success criteria

- `/approvals` keeps both `Foco` and `Linha do tempo`
- no fake inbox fallback remains in runtime UI
- no mock-only inbox enrichment remains on the live screen path
- all action buttons and review entrypoints use real routes or real existing approval flow
- unsupported richness is represented honestly and preserved in backlog
- screen-local implementation quality issues are improved instead of left in place
- verification commands pass
