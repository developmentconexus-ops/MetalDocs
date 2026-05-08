# Module: approval

> **Last verified:** 2026-05-08
> **Scope:** Approval routes, signoffs, ISO segregation enforcement, eligibility enforcement, freeze trigger; Caixa de Aprovação inbox UI (`/approvals`).
> **Out of scope:** Freeze pipeline mechanics (see `workflows/freeze-and-fanout.md`).
> **Key files:**
> - `internal/modules/documents/approval/` — backend approval logic
> - `internal/modules/documents/approval/domain/eligibility.go:1` — pure `CheckEligibility` rule; `ErrActorNotEligible` sentinel
> - `internal/modules/documents/approval/application/decision_service.go:158-170` — eligibility check + audit event in RecordSignoff (step 5b)
> - `internal/modules/documents/approval/application/decision_service.go:284` — RecordSignoff QuorumRejectedStage branch (sets cancel GUC + transitions doc to draft)
> - `internal/modules/documents/approval/repository/postgres_approval_repository.go:305-316` — `loadStageInstances` SELECT … FOR UPDATE (prevents concurrent re-snapshot during signoff, J1)
> - `migrations/0180_signoff_eligibility_trigger.sql:1` — BEFORE INSERT trigger on `approval_signoffs`; DB defense in depth for J1
> - `internal/modules/documents/approval/application/read_service.go:152` — ListInboxItems (JOIN against documents + signoff-count subquery)
> - `internal/modules/documents/approval/application/read_service.go:222` — CountPendingForActor (global count for pagination)
> - `internal/modules/documents/approval/http/inbox_handler.go:15` — InboxHandler (calls ListInboxItems + CountPendingForActor)
> - `internal/modules/documents/approval/http/handler.go:26` — readService interface (widened to include ListInboxItems + CountPendingForActor)
> - `internal/modules/documents/approval/http/errors.go:174` — looksLikeValidationError (E4 gap)
> - `internal/modules/render/fanout/pdf_dispatcher.go:27` — PDFDispatcher.Dispatch (outbox idempotency_key bug)
> - `internal/modules/documents/repository/repository.go:37` — CreateDocument INSERT with MAX(revision_number)+1
> - `internal/modules/documents/approval/http/handler.go:65` — `NewHandler` — accepts `signoffIdempStore` positional param
> - `internal/modules/documents/approval/http/doc_approval_handler.go:51` — `SignoffByDocumentHandler` with idempotency replay
> - `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:1` — `PostgresSignoffIdempStore`
> - `frontend/apps/web/src/features/approval/pages/InboxPage.tsx:9` — `InboxPage`; view switcher (stack/timeline), localStorage persistence (`md.inbox.v`), keyboard nav (←/→), mock fallback strategy
> - `frontend/apps/web/src/features/approval/components/InboxStack.tsx:16` — `InboxStack`; two-panel queue-rail + card area; keyboard event wiring (ArrowLeft/ArrowRight)
> - `frontend/apps/web/src/features/approval/components/InboxApprovalCard.tsx:10` — `InboxApprovalCard`; urgency variant header, stats grid (author/changes/stage), action buttons (stub)
> - `frontend/apps/web/src/features/approval/components/InboxTimeline.tsx:63` — `InboxTimeline`; 4-bucket deadline grouping (Hoje/Amanhã/Esta semana/Próximo mês), animated rail, heatmap sparkline (hardcoded — deferred)
> - `frontend/apps/web/src/features/approval/lib/mockInboxData.ts:7` — `RichInboxItem` type extending `InboxItem`; `MOCK_INBOX_ITEMS`; `enrichInboxItem(item, idx)`
> - `frontend/apps/web/src/features/approval/queries/useInboxQuery.ts:6` — `useInboxQuery` TanStack Query hook wrapping `listInbox`; uses `QK.inbox(params)`
> - `frontend/apps/web/src/features/approval/pages/InboxPage.test.tsx:1` — 6 vitest tests: loading, error, empty→mock fallback, API items, localStorage persistence, next/prev nav
> - `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx:7` — `StageDraft` interface; `toDraft` at :49 maps existing stage fields
> - `frontend/apps/web/src/features/approval/api/approvalTypes.ts:16` — `RouteStage` (includes `required_role`, `required_capability`, `area_code`)
> - `internal/modules/documents/approval/http/contracts/route.go:119` — `ListStageItem` (includes `RequiredRole`, `RequiredCapability`, `AreaCode`)
> - `internal/modules/documents/approval/http/route_admin_handler.go:207` — `ListRoutesHandler` SQL (selects all stage fields)
> - `frontend/apps/web/src/features/approval/components/SignoffDialog.tsx:17` — `error_sod_submitter` / `error_sod_duplicate` states; `mapErrorToState` at :124
> - `frontend/apps/web/src/features/approval/api/mutationClient.ts:9` — `ApprovalError extends ApiError`; 401 dispatches auth-bus; 403 throws with SoD code

## Concepts

### Route

Sequence of stages. Each stage has steps. A step holds the quorum rule for that stage.

### Quorum kinds

- `any_1` — one signoff in the stage advances it.
- `m_of_n` — m of n named approvers.
- (others — verify against domain code)

### Signoff

A user's decision (approve / reject) recorded against a step. Stored with timestamp, user id, optional comment, password-confirmation flag.

## Domain rules

- **ISO segregation**: the document submitter cannot record a signoff on the same document version. Enforced at API, hidden in UI.
- **Idempotency**: re-submitting the same signoff with the same `Idempotency-Key` header returns `{"was_replay":true}` (HTTP 200). Backed by `PostgresSignoffIdempStore` which reads/writes `metaldocs.idempotency_keys`. Keys expire after 24 h.
- **Freeze trigger**: when the final stage's quorum is met, `decision_service.go` calls the freeze service inside the same transaction.
- **Signoff eligibility**: actor must be present in `eligible_actor_ids` frozen at submit time. Enforced at three layers — see _RecordSignoff eligibility check_ below.

## RecordSignoff eligibility check (J1 — fixed cb56e1e0..1cebea64)

`RecordSignoff` enforces eligibility at three layers:

1. **Domain (pure function):** `domain.CheckEligibility(actorUserID, activeStage.EligibleActorIDs)` (`eligibility.go:11`) — returns `ErrActorNotEligible` if actor absent. No DB, no globals. Mirrors `sod.go` shape.
2. **Service (in-tx):** called at `decision_service.go:159` — step 5b, after `authz.Require`, before `CheckSoD`. On failure: emits `signoff.rejected` governance event with `reason=not_eligible` (:160-168), rolls back, surfaces `ErrActorNotEligible`.
3. **DB trigger (defense in depth):** `enforce_signoff_eligibility_trg` (`migrations/0180_signoff_eligibility_trigger.sql`) — `BEFORE INSERT` on `approval_signoffs`; checks `eligible_actor_ids @> actor_user_id` on the parent stage row; raises `ERRCODE 23514` if absent. Belt to application braces.

**FOR UPDATE lock:** `loadStageInstances` (`postgres_approval_repository.go:305-316`) acquires `SELECT … FOR UPDATE` on all stage rows for the instance. This prevents a concurrent re-submit from re-snapshotting `eligible_actor_ids` while a signoff transaction is in flight.

**HTTP error:** `ErrActorNotEligible` → HTTP 403 `signoff.not_eligible` (`errors.go:48-50`). Frontend should handle this code analogously to the existing SoD codes — see `concepts/error-ux.md`.

## Reject path

Rejecting any required signoff returns the document to `draft` so the author can edit and resubmit immediately. The `approval_instance` retains `rejected` status for audit; only the document row reverts.

**Implementation (fixed commit 2977ef96 — B4):** `QuorumRejectedStage` in `RecordSignoff` (`decision_service.go:284`):

1. Marks stage as `rejected_here` and approval instance as `rejected`.
2. Sets `SET LOCAL metaldocs.cancel_in_progress = '<instance_id>'` — the same GUC used by the cancel flow, which authorises the DB trigger to permit the `under_review → draft` transition.
3. Issues `UPDATE documents SET status = 'draft', revision_version = revision_version + 1` within the same transaction.

The approval instance record keeps `status = 'rejected'` for the audit trail. Author can edit the document and call finalize again for a new approval round.

## Route edit — stage config preserved (fixed 41ca209d)

`ListStageItem` previously only returned `label`, `members`, `quorum_kind`, `drift_policy`. Opening a route for editing caused `toDraft` to call `defaultStage()` for every stage, silently wiping `required_role`, `required_capability`, and `area_code`.

**Fixed (F3):**

1. `route_admin_handler.go:207` SQL extended: `SELECT …, required_capability, area_code, …`.
2. `ListStageItem` gains `RequiredRole`, `RequiredCapability`, `AreaCode` JSON fields.
3. `RouteStage` (frontend type) gains `required_role`, `required_capability`, `area_code`.
4. `StageDraft` gains `requiredRole`, `requiredCapability`, `areaCode`.
5. `toDraft` at `RouteAdminPage.tsx:49` maps each existing stage's fields instead of substituting `defaultStage()`.
6. `toRouteStages` at `RouteAdminPage.tsx:118` writes all three fields back on save.

## Inbox area filter (E7 — fixed 6cc016f5)

`InboxPage` previously used a hardcoded `AREA_OPTIONS` array of 6 area codes. The filter now loads areas dynamically via `fetchAreas()` (`frontend/apps/web/src/features/taxonomy/api.ts`) on component mount (`InboxPage.tsx:42`). The `<select>` renders one `<option>` per `ProcessArea` returned by the taxonomy API. No backend change required — the existing `area_code` query param on `listInbox` was already correct.

---

## Known implementation gaps (as of 2026-05-03)

---

### E4 — Re-signoff on completed instance returns 500 instead of 409

**File:** `internal/modules/documents/approval/http/errors.go:174`

When a user attempts a second signoff on an already-approved document, the domain returns an error containing `"no active stage in this approval instance"`. The HTTP error mapper at `errors.go:174` calls `looksLikeValidationError`, which only matches substrings `" is required"`, `" must be "`, and `" must not be "`. The phrase `"no active stage"` does not match, so the error falls through to the 500 path and is returned as `internal.unknown`. The correct HTTP response should be `409 state.instance_completed`.

**Fix needed:** Either add `"no active stage"` (or a canonical sentinel error) to `looksLikeValidationError`, or switch to typed error matching in the HTTP layer.

---

### Outbox idempotency_key bug — PDFDispatcher silently no-ops on repeated dispatches

**File:** `internal/modules/render/fanout/pdf_dispatcher.go:27`

`PDFDispatcher.Dispatch` publishes a `messaging.Event` with no `IdempotencyKey` field set. If the platform messaging layer deduplicates by `IdempotencyKey` and the first call produced an empty key, all subsequent dispatches for different documents may be treated as duplicates and silently dropped. Symptom: freeze succeeds but no PDF is ever generated for documents after the first one.

**Fix needed:** Populate `IdempotencyKey` in `Dispatch`, e.g. `in.RevisionID` or a deterministic hash of `(tenant_id, revision_id)`.

---

### Nova Revisão — revision_number gap (FIXED in migration 0167)

**File:** `internal/modules/documents/repository/repository.go:37`

Previously `CreateDocument` did not include `revision_number` in its INSERT (defaulted to `1`), causing a `ux_documents_v2_cd_revision` unique-constraint violation on any controlled document that already had a document at `revision_number = 1`.

**Fixed by migration 0167:** `controlled_document_id` is now present on `public.documents`, and `CreateDocument` at `repository.go:37` computes `COALESCE(MAX(revision_number), 0) + 1` in the same INSERT via a subquery. The unique index `ux_documents_v2_cd_revision ON documents(controlled_document_id, revision_number)` (migration `0131`) now resolves correctly.

## Inbox UI (Caixa de Aprovação)

The `/approvals` route renders `InboxPage`, which implements a two-view inbox for actors with pending signoff decisions.

### Views

**Foco (stack view — default)**

`InboxStack` renders a two-panel layout:

- **Left rail (`aside.queueRail`):** scrollable list of all pending items, each showing code, title, area, deadline, and an urgency dot. Clicking selects the item.
- **Right card area (`main.cardArea`):** `InboxApprovalCard` for the selected item. Shows urgency-variant header (orange tint when `item.urgent`), summary text, a 3-cell stats grid (author, change count, stage label), and three action buttons (open doc / return / approve). The action buttons are stubs — see `wiki/backlog/caixa-aprovacao.md`.

Counter strip above card area shows `01 / 04` pagination + prev/next buttons.

**Linha do Tempo (timeline view)**

`InboxTimeline` groups items into 4 deadline buckets:

| Bucket | Boundary |
|---|---|
| Hoje | deadline ≤ end of today |
| Amanhã | deadline ≤ end of tomorrow |
| Esta semana | deadline ≤ today + 7 days |
| Próximo mês | everything else |

Each bucket is a two-column grid row: an animated rail column (dot + connecting line; urgent dot pulses orange) and a content column with item rows. Item rows show submitter avatar, code, title, change count, stage progress bars, and a "Revisar →" button (stub — deferred in backlog). Empty buckets render "Nada ainda. Continue assim."

A heatmap sparkline widget (14-day decision history) appears in the timeline header. Values are hardcoded — see `wiki/backlog/caixa-aprovacao.md` for real-data backend prereq.

### View switcher + persistence

`InboxPage` reads initial view from `localStorage.getItem('md.inbox.v')`, defaulting to `'stack'`. `handleViewChange` writes back on every switch. The `InboxToolbar` component renders the toggle buttons and calls `onViewChange`.

Known issue: `view` state is typed as `string` instead of `'stack' | 'timeline'` — tracked in backlog.

### Keyboard navigation (stack view only)

`InboxStack` attaches a `keydown` listener to `window`:

- `ArrowLeft` → `onPrev()` (moves to previous queue item)
- `ArrowRight` → `onNext()` (moves to next queue item)
- `A` / `D` → stubs (approve / return — deferred)

Listener is removed on unmount. Registered only when `items.length > 0`.

### Mock fallback strategy

`useInboxQuery` wraps `listInbox` via TanStack Query (`QK.inbox(params)`). `InboxPage` applies the fallback:

```
API data present + non-empty  → enrich each item with MOCK_EXTRAS (positional, wraps)
API loading or error           → empty list (InboxStack shows loading/error state)
API returned empty             → MOCK_INBOX_ITEMS (4 canonical mock items)
```

`enrichInboxItem(item, idx)` in `mockInboxData.ts:89` spreads `MOCK_EXTRAS[idx % 5]` over the API `InboxItem` to fill `RichInboxItem` fields (`code`, `kind`, `deadline`, `urgent`, `summary`, `changes`, `version`, `deadline_at`) that the backend does not yet return. All `TODO [BACKLOG]` comments in `mockInboxData.ts` and the component files track the removal condition.

### Deferred items

See `wiki/backlog/caixa-aprovacao.md` for all 7 deferred items: action button wiring, heatmap real data, timeline click handlers, "Revisar →" cross-view nav, `view` type narrowing, eye icon, `approvalApi.ts` migration to `lib/api/client.ts`.

---

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 8
- [workflows/approval.md](../workflows/approval.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [concepts/error-ux.md](../concepts/error-ux.md) — SoD dialog states (E2), auth-bus on 401 (E4), shared `ApprovalError`/`ApiError` hierarchy
- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md)
- [backlog/caixa-aprovacao.md](../backlog/caixa-aprovacao.md) — deferred inbox UI items
