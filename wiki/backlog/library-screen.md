# Library Screen — Deferred Backlog

> **Last verified:** 2026-05-06
> **Scope:** Deferred implementation items for the `/documents` Library screen. These are intentional stubs — the screen ships with mock data in places while backend endpoints are designed.
> **Out of scope:** Bug fixes (see `bugs/`), future screens.
> **Key files:**
> - `frontend/apps/web/src/features/documents/components/ActivityPanel.tsx:1` — TODO block for inbox + audit wiring
> - `frontend/apps/web/src/features/documents/components/LibraryStatCards.tsx:1` — TODO block for mocked stat cards
> - `frontend/apps/web/src/features/documents/components/LibraryFilterTabs.tsx:44` — Filtros/Exportar disabled buttons

---

## Status

| Item | Priority | Depends on | Status |
|---|---|---|---|
| ActivityPanel → Inbox (real pending approvals) | High | New backend endpoint | Deferred |
| ActivityPanel → Audit trail (real events) | High | OpenAPI codegen for existing endpoint | Deferred |
| StatCard: Aprovação pendente | Medium | Backend stats extension | Deferred |
| StatCard: Frozen este mês | Low | Backend stats extension | Deferred |
| StatCard: Próx. revisão | Low | New query param on `/documents` | Deferred |
| Filtros panel | Low | Design + backend filter spec | Deferred |
| Exportar action | Low | Export job endpoint | Deferred |

---

## Item 1 — ActivityPanel: Inbox (sua caixa)

**File:** `frontend/apps/web/src/features/documents/components/ActivityPanel.tsx:19`

**Current state:** Hardcoded `INBOX` array with 3 mock items. "Abrir caixa de aprovação" button is `disabled`.

**What it should do:** Show documents pending the **current user's** approval action, sorted by due date ascending.

**Backend work needed:**
- New endpoint: `GET /api/v2/users/me/inbox`
- Returns: list of `{ documentID, code, name, dueAt, dueLabel }` — documents where `current_user` is an eligible approver, filtered to `under_review` status.
- Might be a specialized query on the `signoffs` + `documents` join, scoped to `eligible_actor_ids` containing the current user.
- No endpoint defined yet. File a backend issue before frontend work starts.

**Frontend work when ready:**
1. Add `GET /api/v2/users/me/inbox` to OpenAPI spec → run codegen.
2. Create `features/documents/queries/useInboxQuery.ts`.
3. Replace `INBOX` const in `ActivityPanel.tsx` with query result.
4. Wire count badge: `SUA CAIXA · {data.length} PENDENTES`.
5. Enable "Abrir caixa de aprovação" button → navigate to `/approval` (or inbox route when built).

---

## Item 2 — ActivityPanel: Audit trail

**File:** `frontend/apps/web/src/features/documents/components/ActivityPanel.tsx:26`

**Current state:** Hardcoded `AUDIT` array with 7 mock events. "Ver tudo →" button is `disabled`.

**What it should do:** Show the last 8 hours of document audit events for the current user's tenant (not scoped to current user's actions only — it's a team feed).

**Backend work needed:**
- Endpoint exists: likely `GET /api/v2/audit` (check `internal/modules/audit/` or equivalent).
- Not yet surfaced in OpenAPI spec / codegen types.
- Needed query params: `scope=documents`, `limit=8h` (or `since=<ISO>`)
- Response shape needed: `[{ actorID, actorName, action, targetCode, occurredAt }]`
- Confirm with backend whether `isSystem` flag exists (for system-generated events like "gerou PDF").

**Frontend work when ready:**
1. Surface audit endpoint in OpenAPI spec → codegen.
2. Create `features/documents/queries/useDocumentAuditQuery.ts` with `staleTime: 30_000` and `refetchInterval: 60_000` (near-real-time feel).
3. Replace `AUDIT` const. Map `actorID === 'system'` → `isSystem: true`.
4. Enable "Ver tudo →" button → navigate to `/audit?scope=documents` when audit screen exists.

---

## Item 3 — LibraryStatCards: Aprovação pendente

**File:** `frontend/apps/web/src/features/documents/components/LibraryStatCards.tsx:20`

**Current state:** Hardcoded value `3`, trend `"2 vencendo"`.

**What it should do:** Count of documents currently awaiting the **current user's** approval signature. Distinct from `under_review` (system-wide) — this is the user's personal action queue size.

**Backend work needed:**
- Extend `GET /api/v2/documents/stats` response schema:
  ```json
  {
    "pendingMyApproval": 3,
    "pendingMyApprovalDueSoon": 2
  }
  ```
- Alternatively: derive from `useInboxQuery` result once Item 1 lands (count + due-soon filter).

**Frontend work when ready:**
- Wire `value` function to `statsQuery.data?.pendingMyApproval ?? 0`.
- Wire `trend` to `"${pendingMyApprovalDueSoon} vencendo"` or hide if 0.

---

## Item 4 — LibraryStatCards: Frozen este mês

**File:** `frontend/apps/web/src/features/documents/components/LibraryStatCards.tsx:21`

**Current state:** Hardcoded value `47`, trend `"+12% vs anterior"`.

**What it should do:** Count of documents frozen (status `frozen` + `published`) this calendar month, plus month-over-month delta.

**Backend work needed:**
- Extend `GET /api/v2/documents/stats`:
  ```json
  {
    "frozenThisMonth": 47,
    "frozenThisMonthDelta": 12
  }
  ```
- Backend calculates via `WHERE frozen_at >= date_trunc('month', NOW())`.
- File a backend issue; this is a reporting query, low complexity.

**Frontend work when ready:**
- Wire `value` to `statsQuery.data?.frozenThisMonth ?? 0`.
- Wire `trend` to `"+${delta}% vs anterior"` or `"=${delta}% vs anterior"`.

---

## Item 5 — LibraryStatCards: Próx. revisão

**File:** `frontend/apps/web/src/features/documents/components/LibraryStatCards.tsx:22`

**Current state:** Hardcoded value `23`, trend `"Maio · Jun"`.

**What it should do:** Count of documents whose next scheduled review date falls within the next 60 days, plus the months covered.

**Backend work needed:**
- Option A: New query param on list endpoint: `GET /api/v2/documents?nextReviewWithin=60d&pageSize=0` → response `total`.
- Option B: Extend stats endpoint with `upcomingReview60d: number, upcomingReviewMonths: string[]`.
- Option B is cleaner (single stats call, no phantom list fetch).

**Frontend work when ready:**
- Wire `value` to `statsQuery.data?.upcomingReview60d ?? 0`.
- Wire `trend` to joined month names from `upcomingReviewMonths`.

---

## Item 6 — Filtros panel

**File:** `frontend/apps/web/src/features/documents/components/LibraryFilterTabs.tsx:44`

**Current state:** "Filtros" button is `disabled aria-disabled="true"`.

**What it should do:** Open a sidebar/panel with advanced filter options: date range, area multi-select, profile, author, revision range.

**Depends on:** Design spec for the filter panel + backend support for each filter param in `GET /api/v2/documents`. Most params already exist (`areaCode`, `status`, `q`). Need: `authorID`, `profileCode`, `createdAfter`, `createdBefore`, `revisionMin`, `revisionMax`.

---

## Item 7 — Exportar action

**File:** `frontend/apps/web/src/features/documents/components/LibraryFilterTabs.tsx:58`

**Current state:** "Exportar" button is `disabled aria-disabled="true"`.

**What it should do:** Export current filtered list as CSV or XLSX. Probably a background job (`POST /api/v2/documents/export`) that returns a download URL.

**Depends on:** Backend export job endpoint. No spec yet.

---

## How to pick up a deferred item

1. Check the `Key files:` block at the top for the exact file + line with the in-code TODO.
2. File a backend issue (or confirm endpoint exists) before writing any frontend query.
3. Implement as a standard TanStack Query hook in `features/documents/queries/`.
4. Remove the `// MOCK` comment and the `disabled` attribute on the CTA once wired.
5. Update `Last verified` stamp on `modules/documents.md` and this file.
