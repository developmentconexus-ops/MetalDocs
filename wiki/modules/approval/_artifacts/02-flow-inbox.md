# Phase 2 — Flow trace: GET /api/v2/approval/inbox

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | n/a — handler not generated from spec (legacy `mux.HandleFunc` route) | — |
| Generated server stub | n/a | — |
| Handler | `Handler.InboxHandler` | `internal/modules/documents/approval/http/inbox_handler.go:15` |

## 2. Call chain

```
1. inbox_handler.go:15  Handler.InboxHandler — extracts tenantID, actorID (`iamdomain.UserIDFromContext`), area_code, limit, offset
   → calls: read_service.go:153 ReadService.ListInboxItems
   → calls: read_service.go:224 ReadService.CountPendingForActor

2. read_service.go:153  ListInboxItems — opens NO transaction (uses `db.QueryContext` directly); marshals actorID → JSON array; runs single SELECT joining approval_instances ai + approval_stage_instances asi (active) + LEFT JOIN documents d, with signoff-count subquery. Returns []InboxView.
   → calls: pgx driver via *sql.DB

3. read_service.go:224  CountPendingForActor — opens NO transaction; runs second SELECT COUNT(DISTINCT ai.id) over the same join (no documents LEFT JOIN). Returns int.
   → calls: pgx driver via *sql.DB

4. inbox_handler.go:50-69  builds []contracts.InboxItem from views; writes WriteJSON 200 with InboxResponse{Items, Total}.
```

No service-layer transaction. No authz call. No idempotency. No pagination cursor — limit/offset only.

## 3. State changes

`none` — read-only.

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg |
|---|---|---|---|
| `read_service.go:163-198` | SELECT | `approval_instances ai`, `approval_stage_instances asi`, `documents d` (LEFT JOIN), subquery on `approval_signoffs s` | none — filters by `ai.tenant_id = $1` and `asi.eligible_actor_ids @> $2::jsonb` (actor scoping via JSON containment) |
| `read_service.go:231-241` | SELECT COUNT | `approval_instances ai`, `approval_stage_instances asi` | same tenant + JSONB-containment scoping |

**Tripwire pairing:** N/A — both statements are SELECT. Tripwire (`enforce_capability_asserted`) fires only on INSERT/UPDATE/DELETE per `migrations/0142b:200-209`.

**`authz.Require` called?** NO — neither for inbox nor for count. Reads rely on:
- HTTP middleware tier-1 capability gate (verify in Phase 3 cross-deps).
- Tenant-id JSONB containment scoping in the WHERE clause.

## 5. Response shape

- 200: `contracts.InboxResponse{ Items: []InboxItem, Total: int }` — `inbox_handler.go:66-69`; `InboxItem` declared at `contracts/instance_read.go:33-44`.
- Errors: `WriteError(w, reqID, err)` — funnelled through `errors.go:147` → `MapErrorToResponse` (errors.go:22). Envelope: legacy `contracts.ErrorResponse{Error: ErrorBody{Code, Message, Details, TraceID}}` (`contracts/errors.go:3-12`). **NOT RFC 9457.**
- Validation errors from `parseInboxLimit`/`parseInboxOffset` (e.g. `"limit must be between 1 and 100"`) currently flow through `looksLikeValidationError` (`errors.go:181`); substring match `" must be "` matches → status 422.

## 6. Cross-references

- Idempotency: no.
- Pagination: yes — `limit` (1..100, default 25) + `offset` (≥0). No cursor. Two SQL round-trips per page (data + count) — same shape as documents list per `wiki/architecture/data-model.md` "two-query LIMIT/OFFSET+COUNT pattern".
- Audit log emission: no — read path; no `governance_events` write.

## Notes / open observations (raw — composer to interpret)

- `InboxHandler` does NOT begin a tx; the two SELECTs run on different snapshots. A signoff committed between the two queries can produce `Total < len(Items) + 1` or vice versa. Existing stub does not flag this.
- Actor matching uses JSONB containment (`@>`) over `eligible_actor_ids` — relies on the snapshot populated by `submit_service.go:299 resolveEligibleActors` (workflow doc, J1 fix).
- `area_code` filter compares against `asi.area_code_snapshot` (snapshot column). Matches the Phase-0 hypothesis: snapshot at submit time, not live.
- Frontend wraps both reads in `useInboxQuery` (`features/approval/queries/useInboxQuery.ts:6`).
