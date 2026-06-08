# ADR — Audit Events Cursor Shape Drift

> **Date:** 2026-06-03
> **Status:** **CLOSED 2026-06-08** (api-contract-hardening Phase F closing re-audit). The runtime handler now emits the canonical nested `{items, page:{next_cursor, has_more}}` CursorPage envelope matching the spec; the FE dual-shape adapter (`useAuditEventsQuery.adaptPage`) was simplified to read only the nested shape. Drift reconciled at root — no FE workaround remains. Commit `4806167ac`. (Casing drift was resolved earlier by Phase E1.)
> **Owner:** Backend audit module
> **Scope:** `GET /audit/events` response shape

## Context

The generated OpenAPI client declares the paged response as:

```jsonc
{
  "items": [ /* ... */ ],
  "page": {
    "next_cursor": "…",
    "has_more": true
  }
}
```

But the runtime server (current `main`) emits a flat shape:

```jsonc
{
  "items": [ /* ... */ ],
  "next_cursor": "…",
  "has_more": true
}
```

> **Phase E1 note (2026-06-08):** The flat-shape field names were `nextCursor`/`hasMore` at time of writing; Phase E1 snake_case big-bang renamed them to `next_cursor`/`has_more`. The nesting drift (flat vs. nested `page.{...}`) remains open.

The frontend audit hook (`features/iam/queries/useAuditEventsQuery.ts`) sits on
both shapes via an `adaptPage()` helper. This unblocks PR-12 Fase 5 (Audit
tab) without renegotiating the contract under deadline.

## Decision

Adapt at the frontend for now. Keep `adaptPage()` tolerant of either shape and
expose a single normalized `AuditEventsPage` to callers.

Do **not** change `/audit/events` route, query params, or OpenAPI spec from the
frontend side. The backend audit module owns reconciliation.

## Consequences

- FE callers stay on a stable internal shape; no churn when backend converges.
- One adapter is the single source of brittleness — keep it covered by tests
  before the audit module starts emitting the nested shape.
- Generated TS types under `lib/api-types/` still reflect the spec, so any
  attempt to access `data.page.next_cursor` directly will type-check while
  failing at runtime. The adapter is the only safe access path.

## Closure Plan

1. Backend audit module decides which shape is canonical.
2. Backend either updates the spec to match the flat runtime, or updates the
   handler to emit the nested `page.{next_cursor, has_more}`.
3. Run codegen, update `adaptPage()` to the single canonical path (or remove
   it if direct access is safe).
4. Delete this ADR or move to Accepted/Superseded with the resolved shape.

## Related

- `frontend/apps/web/src/features/iam/queries/useAuditEventsQuery.ts`
- `wiki/architecture/api-contract.md`
- `wiki/architecture/api-design-system.md`
