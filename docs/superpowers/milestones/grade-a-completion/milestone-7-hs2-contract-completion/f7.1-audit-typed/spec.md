# Feature F7.1 — Audit typed responses (close the confirmed Major)

> **Milestone:** 7 — HS-2 Contract Completion  ·  **Folder:** `f7.1-audit-typed`
> **Status:** Approved 2026-06-20 — code change may begin.
> **Approved before code:** 2026-06-20 / leandrotca.work — inherited from the M7 Phase-2 operator
> approval (README hard-stops 2026-06-20 + commit `45a03fa6`). This feature spec introduces **no new
> consumer contract** beyond `milestone.md` F7.1; it distills it and records one internal-structure
> decision (Q7) that does not change any wire shape, so it needs no fresh operator sign-off.

## Interview record (B1.5)

| # | Question | Answer / source |
|---|----------|----------------|
| 1 | Wire output changing? | **No.** F7.1 is a typed-parity refactor: the generated `auditapi.*` types' JSON tags equal today's hand-rolled map keys exactly. Byte-identical wire (with care on timestamp precision — see Q5). |
| 2 | Do the generated types already exist? | **Yes** — `internal/modules/audit/api/api.gen.go` already declares `AuditExportStatusResponse` (`:80`), `AuditExportResponse` (`:72`), `ListAuditEventsResponse` (`:115`), `CursorPage` (`:100`), `AuditEventItem` (`:50`). The hand-rolled mux handlers ignore them — that unused-generated-type drift **is** the confirmed Major. No OpenAPI change, no codegen regen needed in F7.1. |
| 3 | Which sites are response literals (kill) vs non-response (keep)? | Kill 4: `handler.go:120` `page := map[string]any{}`, `:127` events-list envelope, `:216` export-POST envelope, `:268` export-status `resp`. Keep 2: `:51` `EventResponse.Payload` field decl (domain-mirror — but see Q7), `:404` `payload := map[string]any{}` decode buffer (feeds the payload struct field; non-response). |
| 4 | Is the audit handler routed via codegen `ServerInterface`/`NewStrictHandler`? | **No** — it is a hand-rolled `http.ServeMux` (`RegisterRoutes`). F7.1 only swaps the response *bodies* to generated types; it does **not** rewire routing. (HS-2 boundary respected: no pipeline standup, no routing rewire.) |
| 5 | Timestamp wire-precision risk? | Real. Current emits `…UTC().Format(time.RFC3339)` (second precision). The generated types carry `time.Time` (`occurred_at`, `expires_at`), which marshal as RFC3339Nano. To stay byte-identical, build each `time.Time` field as `t.UTC().Truncate(time.Second)` so the marshaled value is second-precision `…Z`, equal to today. Verified: a second-truncated UTC `time.Time` marshals to the same string `Format(time.RFC3339)` produces. |
| 6 | Conditional fields (export-status `expires_at`, `error`)? | Preserved via the generated pointer+`omitempty` fields: set `ExpiresAt` only when `!job.ExpiresAt.IsZero()`, set `Error` only when `job.ErrorMessage != ""` — identical emit condition to today. |
| 7 | Events-list items: keep hand-rolled `EventResponse` or build generated `AuditEventItem`? | **Build `AuditEventItem`; retire `EventResponse`.** `ListAuditEventsResponse.Items` is `[]AuditEventItem`, so emitting the generated envelope requires items of that type. `AuditEventItem`'s JSON tags equal `EventResponse`'s exactly, and its `Payload map[string]interface{}` is the same allowlisted domain-mirror field — the substantive intent of milestone.md's "keep `EventResponse.Payload`" (don't gold-plate the arbitrary-JSON payload field) is preserved: the field survives as `AuditEventItem.Payload`. Keeping a structurally-identical hand-rolled twin solely to satisfy the literal line reference would be the redundancy a senior reviewer flags. `EventResponse` is referenced only inside `handler.go` (no test/external refs — verified), so retiring it is safe. **Wire-identical either way.** |
| 8 | Optional drive-by (§7 #11–13, 405 `Allow` header Minors)? | In the same file. Decision recorded in evidence: repair only if zero-risk and inside touched code; otherwise skip cleanly (not H-D, not required for the bar). |

## Consumer contract (FIRST — before any producer)

**Consumers:**
- FE TanStack-Query callers of the audit endpoints via the audit `api.gen.ts` codegen.
- Existing Go handler tests (`handler_test.go`, `handler_export_test.go`) that decode the wire JSON by struct.

**Contract — wire-identical to today, now emitted via the already-generated `auditapi.*` types.**

| Op | Path / method | Status | 200/202 body type emitted | Wire keys (unchanged) |
|----|---------------|--------|---------------------------|------------------------|
| list audit events | `GET /api/v1/audit/events` | 200 | `auditapi.ListAuditEventsResponse{Items []AuditEventItem, Page CursorPage}` | `{items:[{id,occurred_at,actor_id,action,resource_type,resource_id,payload,trace_id}], page:{next_cursor,has_more}}` |
| create export | `POST /api/v1/audit/events/export` | 202 | `auditapi.AuditExportResponse` | `{export_id,status,signed_url,expires_at}` |
| export status | `GET /api/v1/audit/events/export/{id}` | 200 | `auditapi.AuditExportStatusResponse` | `{export_id,status,signed_url,expires_at?,error?}` (**the confirmed Major**) |

`total` on `ListAuditEventsResponse` is `*int,omitempty` — left nil → omitted (no `total` today). Error
responses (problem+json) untouched.

**Source of truth for the contract:** the generated `auditapi.*` types (themselves the OpenAPI's Go
projection) + the existing handler-test decoders (`handler_test.go:117`, `handler_export_test.go:74,169`).

## What this feature implements

1. Import `auditapi "metaldocs/internal/modules/audit/api"` into `audit/delivery/http/handler.go`.
2. **events-list** (`handleEvents`): build `page` as `auditapi.CursorPage{NextCursor *string, HasMore bool}`;
   emit `auditapi.ListAuditEventsResponse{Items, Page}`. Retarget `buildEventResponses` to return
   `[]auditapi.AuditEventItem` (set `OccurredAt: item.OccurredAt.UTC().Truncate(time.Second)`; keep the
   `payload := map[string]any{}` decode buffer feeding `Payload`). Remove the now-unused `EventResponse` type.
3. **export-POST** (`handleExport`): emit `auditapi.AuditExportResponse{ExportId, Status, SignedUrl,
   ExpiresAt: job.ExpiresAt.UTC().Truncate(time.Second)}` at 202.
4. **export-status** (`handleExportSubresource`): build `auditapi.AuditExportStatusResponse{ExportId,
   Status, SignedUrl}`; set `ExpiresAt` (truncated) only when `!IsZero`; set `Error` only when non-empty;
   emit at 200.
5. NextCursor: keep the existing encode logic; assign into the `*string` (nil when no further page).

## Non-goals (mandatory)

- No change to wire JSON keys/values (byte-identical, incl. second-precision timestamps).
- No change to status codes (200 list / 202 export / 200 status; problem responses untouched).
- No routing rewire, no `NewStrictHandler`, no new codegen pipeline (HS-2 boundary).
- No OpenAPI change, no BE codegen regen (types already generated).
- No FE codegen regen in F7.1 (no schema change — nothing to regen).
- No change to authz/tenant scoping, cursor codec, pagination clamp, or export service logic.
- The `payload` `map[string]any` (decode buffer + `Payload` field) is **kept** — converting arbitrary
  JSON to a typed shape is out of scope and would be gold-plating.

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|------------------|
| 1 | Zero response-literal `map[string]any` in `handler.go` — only the `:404` decode buffer survives (non-response). | `grep -nE 'map\[string\]any' internal/modules/audit/delivery/http/handler.go` → exactly 1 hit, the decode buffer | real (grep) — **the red→green**: 5 hits (4 response literals + buffer) → 1 |
| 2 | Grep A blind-spot site closed too (the `httpresponse.WriteJSON(...map[string]any{})` calls gone). | `grep -nE '(WriteJSON\|writeJSON).*map\[string\]any' internal/modules/audit/delivery/http/handler.go` → 0 | real (grep) |
| 3 | The confirmed Major closed — export-status emits `auditapi.AuditExportStatusResponse`. | NEW `TestAuditHandler_ExportStatusTypedShape` strict-decodes the 200 body into `auditapi.AuditExportStatusResponse` with `DisallowUnknownFields` | real |
| 4 | export-POST + events-list emit the generated types, wire unchanged. | NEW `TestAuditHandler_ExportPOSTTypedShape`, `TestAuditHandler_ListEventsTypedShape` strict-decode into `auditapi.AuditExportResponse` / `auditapi.ListAuditEventsResponse` | real |
| 5 | No wire drift — all existing audit handler tests pass unmodified. | `go test -count=1 ./internal/modules/audit/...` → 0 FAIL | real |
| 6 | Build green; whole-repo tests green from clean. | `go build ./...` clean; `go test -count=1 ./...` → 0 FAIL | real |

## ADR needed?

- [x] No durable decision — F7.1 follows the contract-first / ADR 0012 posture already in force for audit
  (generated types exist); Q7 is an internal-structure call, wire-neutral, recorded above.
