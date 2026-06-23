# Feature F5.1 — Spec (Detalhe Signoff)

> **Milestone:** 5 — Detalhe Signoff + Taxonomy Admin restyle  ·  **Folder:** `f5.1-signoff-detail`
> **Status:** Approved — implementation may begin
> **Approved before code:** 2026-06-23 — operator (leandrotca) APPROVE.

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Engine: `superpowers:brainstorming` (run this session). Recon-first: the consumer surface was read
from live app + code before the contract was written, which **corrected the milestone assumption**
(see README HS-6 2026-06-23 — the decision surface already exists as an orphan component).

| # | Question | Answer |
|---|----------|--------|
| 1 | Route shape + how the screen is reached from the inbox? | **New route `/approvals/:documentId`, replaces the modal.** `InboxStack` approve/reject navigates to the detail screen; decision happens there. Keyed by `documentId` (reuses `getInstance`/`signoff` directly — no new lookup). |
| 2 | A4 document-body source? (`/documents/{id}/view` returns a rendered-PDF pointer, not structured blocks; no diff backend exists) | **Embed the rendered PDF** from `GET /documents/{id}/view` (`pdf_url`/`signed_url`; honest loading when `pdf_status` pending). Inline structured-diff overlay is **deferred-with-trigger** (no diff endpoint). |
| 3 | Which tabs ship live vs defer? | **Comentários live** ← `GET /documents/{id}/comments`. **Mudanças vs vX (diff) deferred-with-trigger** (no diff backend; honest absent state, backlog row). **Trilha dropped** — the approval-flow timeline covers audit. |
| 4 | Build the decision form from the design HTML, or reuse existing? | **Reuse.** Recon found `ControlledDocumentDetailPanel` (exported, unit-tested, **mounted nowhere**) already implements the full decision surface (per-state policy, Assinar/Cancelar/Publicar, integrity, inline `ApprovalTimelinePanel`, lock/stale banners). It internally reuses `ApprovalTimelinePanel` + `SignoffDialog`. F5.1 **mounts the orphan**, does not rebuild. |
| 5 | Given the decision surface already exists + A4 already renders in the editor — which approach? | **Option α — new route, assemble existing parts.** Beats β (extend editor review view) because it keeps the approver's decision cockpit separate from the author's edit surface, is deep-linkable, and matches the design-source standalone screen. Operator-approved 2026-06-23. |

## Consumer contract (FIRST — before any producer)

This feature is consumer-side assembly: no new producer/endpoint is built. The "contract" is the set
of existing producer shapes the screen consumes, plus the navigation contract it exposes to the inbox.

- **Consumer(s):**
  - The new route `/approvals/:documentId` (a page component) — renders the cockpit.
  - `InboxStack`/`InboxPage` — the **caller**: its approve/reject actions now `navigate` to the route
    instead of opening the in-inbox `SignoffDialog` modal.
- **Contract (existing shapes consumed, read from current callers — not invented):**
  - **Entry resolution:** `getActiveDocumentContext(controlled_document_id)` →
    `ActiveDocumentResponse` providing `document_id`, `content_hash`, `approval_instance_id`,
    `revision_version` (exactly as `InboxPage.openDecisionFlow` consumes today,
    `frontend/apps/web/src/features/approval/pages/InboxPage.tsx:70`).
  - **Decision surface props:** `ControlledDocumentDetailPanel` requires
    `{ documentId, approvalState, contentHash, revisionVersion, lockedByInstanceId? }` — all sourced
    from the active-document context above (`approvalState` = the document's `under_review` status;
    `lockedByInstanceId` = `approval_instance_id`, which the panel requires truthy to enable Assinar).
  - **Approval instance:** `getInstance(documentId)` → `ApprovalInstance` (the panel fetches this
    itself; the timeline + per-stage signoffs render from `ApprovalInstance.stages`).
  - **A4 body:** `GET /documents/{id}/view` → `ViewDocumentResponse { pdf_status, signed_url?, pdf_url? }`.
  - **Comments:** `GET /documents/{id}/comments` (existing endpoint + hook).
  - **Decision out:** `signoff(documentId, { decision, reason?, password, content_hash }, { ifMatch:
    "\"v{revisionVersion}\"" })` — fired by the reused `SignoffDialog` inside the panel; unchanged.
- **Source of truth for the contract:** the existing callers — `InboxPage.tsx`,
  `approval/components/ControlledDocumentDetailPanel.tsx`, `approval/api/approvalApi.ts` — and the
  generated FE types in `lib/api-types` (`ActiveDocumentResponse`, `ViewDocumentResponse`). No
  hand-authored response types.

## What this feature implements

A new route `/approvals/:documentId` rendering the **Detalhe Signoff cockpit**, assembled from
existing tested parts and matching `design-source/detalhe-signoff/detalhe-signoff.html`:

- **Left column (read the document):** doc header (code/title/version/status from the document-detail
  query) → tab strip → body.
  - **Documento tab:** embed the rendered PDF from `GET /documents/{id}/view` in an A4 frame; honest
    loading state while `pdf_status` is pending; honest absent state if no artifact.
  - **Comentários tab:** live `GET /documents/{id}/comments`.
  - **Mudanças vs vX tab:** honest deferred state (no faked diff); a backlog row in
    `wiki/backlog/detalhe-signoff.md` names the unblock trigger (a document-diff endpoint).
- **Right column (make the decision):** **mount `ControlledDocumentDetailPanel`** with props from the
  active-document context. For an `under_review` document its policy surfaces **Assinar** (→ reused
  `SignoffDialog`) + **Cancelar instância**, the inline `ApprovalTimelinePanel`, and the integrity
  block.
- **Inbox rewire:** `InboxStack` approve/reject (`InboxPage.openDecisionFlow`) resolves the
  active-document context and `navigate(\`/approvals/${document_id}\`)` instead of `setDialogState`.
  The in-inbox `SignoffDialog` modal is removed **for that approve/reject entry path** (the dialog
  component itself stays — it is reused by the panel).
- On a successful sign-off, the existing panel/`SignoffDialog` cache invalidation (`['approval']`,
  `['documents']`) runs; navigating back to the inbox shows the updated stack.

## Non-goals (mandatory)

Anything here that later appears in the diff is scope drift (validator C6).

- **No new backend** — no new endpoint, field, migration, or contract change. All consumed endpoints
  already ship.
- **No document-diff backend and no faked diff** — the "Mudanças vs vX" tab is deferred-with-trigger,
  never invented data.
- **No fork / re-implementation** of `ControlledDocumentDetailPanel`, `ApprovalTimelinePanel`, or
  `SignoffDialog`. No second decision form, timeline, or sign-off path. (If design parity *requires*
  forking the panel → **HS-2 stop + replan**, not a silent rewrite — see R4.)
- **No change to approval semantics / state machine / route-admin** — consume only.
- **Not** extending the editor review view (Option β was rejected) and **not** removing the
  `SignoffDialog` component (still reused by the panel).
- **No fabricated header fields** — the design mock's `vence hoje 18:00` SLA chip and illustrative
  author/stage strings are not rendered unless a real field backs them (no due-date field exists →
  omitted).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Route `/approvals/:documentId` mounts the cockpit (left A4 + right decision panel) | `SignoffDetailPage.test.tsx` — renders given a mocked active-doc context + instance | fixture (vitest, mocked query layer) |
| A4 region embeds the real PDF from `GET /documents/{id}/view` (no `MOCK_`/illustrative literal); honest loading when `pdf_status` pending | `SignoffDetailPage.test.tsx` — asserts the embed targets the `view` `pdf_url`/`signed_url`; loading branch on pending | fixture |
| Decision surface is the mounted `ControlledDocumentDetailPanel` (not a new form) — Assinar present for `under_review` | `SignoffDetailPage.test.tsx` — asserts the panel renders Assinar; reuse assertion | fixture |
| Recording a decision fires `signoff(documentId, …)` with `If-Match` and invalidates caches | reuse of existing `SignoffDialog`/panel behavior; `grep` confirms no second `signoff(` call site added | fixture + grep |
| Inbox approve/reject navigates to `/approvals/:documentId` (modal path removed for that flow) | `InboxPage.test.tsx` — asserts `navigate` to the route on approve/reject; no `setDialogState` for that path | fixture |
| Comentários tab renders live from `GET /documents/{id}/comments` | `SignoffDetailPage.test.tsx` — comments branch | fixture |
| "Mudanças vs vX" diff tab shows an honest deferred state; a backlog row exists with a trigger | `grep -n "em breve" SignoffDetailPage.tsx` = 0 unbacked; `wiki/backlog/detalhe-signoff.md` has the diff row | grep + file |
| No forked timeline/decision/sign-off component; generated types consumed directly | `grep` for new `ApprovalTimelinePanel`/`SignoffDialog`/signoff-form definitions = 0; `tsc` clean | grep + `tsc` |
| Visual parity with `detalhe-signoff.html` | `frontend-screen-reviewer` APPROVE on record | real (reviewer) |
| Architecture / maintainability | `frontend-code-reviewer` APPROVE on record | real (reviewer) |
| Type + test health | `pnpm.cmd tsc --noEmit` clean; `vitest run` for the new + touched suites green | real |

> TDD: write the failing `SignoffDetailPage.test.tsx` (and the `InboxPage` navigation assertion)
> first, then implement to green. Fixture-only proof is labeled — the reviewer-APPROVE + `tsc`/vitest
> rows are the real-provider/real-tool proof. A live end-to-end sign-off is gated on seeding an
> `under_review` document (dev DB is currently empty) and is covered by the reviewer drive, not a
> fabricated value.

## ADR needed?

- [x] No durable backend/contract decision — skip. F5.1 is consumer-side assembly of existing parts
  (no new endpoint/capability). The notable decision (mount the orphan `ControlledDocumentDetailPanel`
  behind a new `/approvals/:documentId` route, replacing the inbox modal for the approve/reject path)
  is recorded in the program README hard-stops table (HS-6, 2026-06-23) and in this spec, not an ADR.
