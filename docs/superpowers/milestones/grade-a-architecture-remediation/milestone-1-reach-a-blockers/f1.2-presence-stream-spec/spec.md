# Feature F1.2 — presence/stream contract declaration — Spec

> **Milestone:** 1 (Reach-A Blockers)  ·  **Folder:** `f1.2-presence-stream-spec`
> **Closes:** Grade-A blocker #3 (governing spec line 97/132) — `GET /iam/presence/stream`
> live + FE-consumed but absent from `openapi.yaml`/stub/FE-types (H-D contract tri-source drift, stronger form).

## Interview record (fail-closed gate)

The consumer contract was **read from source, not guessed**. The one genuine decision — *how* a
WebSocket-upgrade route reaches contract-consistency without breaking the live 101 — was a real
design fork (OAS3/oapi-codegen `strict-server` cannot model a hijacked 101 upgrade), so it was
**escalated to the operator** rather than silently picked.

| Q | A | Source |
|---|---|--------|
| What does the route do at runtime? | `GET /api/v1/iam/presence/stream` → WebSocket **101 upgrade** via `coder/websocket` `websocket.Accept`, streams JSON `Event` frames | `internal/modules/iam/presence/handler.go:66,97,117` |
| Who consumes it, and how? | FE `usePresenceStream` opens a **raw native `WebSocket`** (deliberately outside the codegen client) at `ws://…/api/v1/iam/presence/stream` | `frontend/apps/web/src/features/iam/queries/usePresenceStream.ts:31-39,105` |
| What is the on-wire message shape? | discriminated union keyed on `type`: `snapshot` (carries `presence[]`), `join`/`leave`/`online`/`idle` (per-user delta), `heartbeat` | producer `internal/modules/iam/presence/model.go:64-71` (`Event`), `:55-59` (`Item`) |
| Auth guard? | capability `user.view`, tier-1 permission middleware | `apps/api/cmd/metaldocs-api/permissions.go:247` |
| Can a WS 101 be a `strict-server` op? | **No** — strict codegen models request→response; a generated stub would shadow/replace the live WS handler (breaks 101) | oapi-codegen `strict-server` semantics; milestone "101 upgrade stays live" |
| Chosen strategy? | **A — declare for documentation + exclude from strict codegen** (operator-approved 2026-06-14) | this spec §"What this implements" |

**Supersedes:** the archived wont-fix note `wiki/_archive/backlog/api-contract-hardening.md:129`
("WS not representable in OAS3 → leave undocumented"). That note is **retired** by this feature:
the route is now *declared-for-docs* and *codegen-excluded*, which is the industry-default for
streaming endpoints (document the route + frame shape; never auto-generate the upgrade handler).

## Consumer contract (FIRST — before any producer change)

- **Consumer:** the FE presence client `usePresenceStream.ts` (raw `WebSocket`) **and** the generated
  FE type surface (`lib/api-types/index.d.ts`) that the audit/route-truth-table reads for consistency.
- **Contract shape (producer must match — read from `model.go:64-71`):** a JSON frame
  ```
  Event = {
    type: "snapshot"|"join"|"leave"|"online"|"idle"|"heartbeat",  // required
    user_id?, username?, display_name?, last_seen_at?(date-time), status?,  // delta frames
    presence?: PresenceStreamItem[]   // snapshot frame only
  }
  PresenceStreamItem = { user_id, username, display_name, last_seen_at(date-time), status }
  ```
  `status` is the presence `Status` enum (`model.go:59`). `PresenceStreamItem` is `OnlinePresenceItem`
  **plus `status`** — it must **not** mutate the existing `OnlinePresenceItem` (snapshot uses that).
- **Transport:** `GET` that returns **`101 Switching Protocols`** (WebSocket upgrade), guard `user.view`.
- **Source of truth (read, not guessed):** `model.go` (frame), `handler.go` (transport/guard), the FE
  client + generated types (consumer). The producer change is purely the **contract declaration**; the
  runtime handler is already correct and is **not modified**.

## What this implements (Strategy A)

1. **Declare** `GET /iam/presence/stream` in `api/openapi/v1/openapi.yaml`: `tags: [iam]`,
   `operationId: streamPresence`, a documented **`101`** response (description = WS upgrade; the streamed
   frame documented via a new `components/schemas/PresenceStreamEvent` + `PresenceStreamItem` that mirror
   `model.go` exactly), the `user.view` security, and standard `401/403` error refs. Mirror the existing
   `getPresenceSnapshot` block style (`openapi.yaml:527`).
2. **Exclude from server codegen:** add `exclude-operation-ids: [streamPresence]` to
   `internal/modules/iam/api/cfg.yaml` `output-options`. The op stays `iam`-tagged (documented/grouped),
   but oapi-codegen emits **no** `PresenceStream` server-interface method and **no** mux registration for
   it; with default pruning, `PresenceStreamEvent`/`PresenceStreamItem` (referenced only by the excluded
   op) are pruned from the Go stub. → the hand-written WS handler stays authoritative.
3. **Regenerate, canonical order:** `go generate` the iam stub (oapi-codegen) → then FE `gen:api`
   (openapi-typescript). The only expected `api.gen.go` change is the **embedded-spec blob** now
   containing the new path (correct — it embeds `openapi.yaml`); **no** handler/route-registration change.
   FE `index.d.ts` gains the `/iam/presence/stream` path + the two schemas (openapi-typescript ignores
   tags/exclusions → FE type is generated, satisfying the consumer-type half of tri-source).

## Non-goals (mandatory)

- **No change to the WS handler / upgrade path** — `presence/handler.go` (incl. `statusWriter.Unwrap()`)
  is untouched. This is a contract-declaration feature, not a delivery redesign.
- **No refactor of `usePresenceStream`** to route through the codegen client — raw `WebSocket` is the
  correct pattern; forcing the generated client would be wrong and is out of scope.
- **No mutation of `OnlinePresenceItem`** or the `getPresenceSnapshot` contract.
- **No other route's openapi/FE-type may shift** in the regen (HS-6 guard).
- **No `strict-server` stub for the WS route** (Strategy B) — that breaks the 101 (HS-2).

## Validation Gate (acceptance — objectively checkable)

| # | Criterion | Named proof | Real vs fixture |
|---|-----------|-------------|-----------------|
| AC1 | **Route-truth-table** built for `/iam/presence/stream` → openapi.yaml ↔ FE types ↔ runtime **agree** (method `GET`, guard `user.view`); server-stub **intentionally excluded** (documented, not silent drift). Was: runtime-only. | the table itself (4 columns) + the `exclude-operation-ids` line as the documented reason | real |
| AC2 | oapi-codegen regen **deterministic & handler-clean** | `go generate ./internal/modules/iam/api/...` then re-run → `git diff` empty on 2nd run; `grep PresenceStream internal/modules/iam/api/api.gen.go` → **no** ServerInterface method / **no** mux registration (only the embedded-spec blob mentions the path) | real |
| AC3 | FE codegen clean; FE type present; types compile | `pnpm --filter web gen:api` → `index.d.ts` contains `"/iam/presence/stream"`; `pnpm --filter web tsc --noEmit` → **0 errors** | real |
| AC4 | **No collateral contract shift** | `git diff api/openapi/v1/openapi.yaml` and `git diff …/index.d.ts` show **only** the stream path + the two new schemas — no other path/operation/schema changed | real |
| AC5 | **Runtime 101 still live** | API on `:8081` (`.\scripts\start-api.ps1 -Build`); a WS client (authenticated, `Origin` set) to `/api/v1/iam/presence/stream` observes **HTTP 101** + a `snapshot` frame — by us, not deferred | real |

> AC5 closes the long-standing `[runtime-unverified]` residual (`wiki/backend/wave-h-plan.md:32`): prove
> 101 (not 501) on a live build. A successful upgrade + first frame is the acceptance, not "looks wired".

## ADR / decision record needed?

- [x] **Lightweight decision record** — retire the archived wont-fix and record the precedent:
  *"WebSocket/streaming routes are declared in `openapi.yaml` for documentation + excluded from
  `strict-server` codegen via `exclude-operation-ids`; the hand-written upgrade handler stays
  authoritative."* Captured at close (wiki decisions / api-contract doc via `wiki-curator`); no heavy
  architectural ADR — the governing spec already mandated declaring the route; A is the *how*.

## Approval

Strategy **A approved by the operator 2026-06-14** (after a detailed walkthrough of the
WebSocket-vs-OpenAPI constraint and the industry-default pattern). The consumer contract is read from
source; the governing spec (line 132) already mandated declaring the route. No implementation began
before this line.
