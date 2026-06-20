# Feature F8.1 — Spec (SEED — approval pending)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.1-presence-typed-body`
> **Status:** Drafting (seed authored 2026-06-20 from post-M7 re-audit; **interview + approval gate pending in execution session**)
> **Approved before code:** PENDING — *no implementation begins until this line is filled (Phase 3, fresh session).*

> Seed contract from the post-M7 re-audit (Contract Major #4). The execution session must run the
> brainstorming interview, confirm/adjust below, then fill the approval line before code.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Does `OnlinePresenceItem` (generated) already carry every field presence `Item` emits? | **Yes** — verified `display_name, last_seen_at, status, user_id, username` present (`iam/api/api.gen.go:453-459`); presence `Item` matches (`presence/model.go:54`). No OpenAPI change needed. |
| 2 | Must the wire shape stay byte-identical? | Yes — `{"items":[…]}`; consumer is the web presence view. Map 1:1; assert byte-identical in test. |

## Consumer contract (FIRST)

- **Consumer(s):** web presence UI consuming `GET /api/v1/iam/presence/snapshot`.
- **Contract:** `200 application/json` body `{ "items": OnlinePresenceItem[] }` where each item = `{user_id, username, display_name, last_seen_at, status?}`.
- **Source of truth:** OpenAPI `PresenceSnapshotResponse` (`openapi.yaml:527-538`) + generated `iamapi.PresenceSnapshotResponse` (`api.gen.go:465`).

## What this feature implements

`handleSnapshot` ([`iam/presence/handler.go:83`](../../../../../internal/modules/iam/presence/handler.go)) maps `[]presence.Item` → `[]iamapi.OnlinePresenceItem` and emits `iamapi.PresenceSnapshotResponse` via `httpresponse.WriteJSON`, replacing the `map[string]any{"items":…}` response literal. The generated model is currently unused.

## Non-goals (mandatory)

- No change to the WS `/stream` route or its `Event` envelope.
- No OpenAPI schema change (fields already aligned).
- No StrictServerInterface adoption for iam.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| No response `map[string]any` in presence handler | `grep -n 'map\[string\]any' internal/modules/iam/presence/handler.go` → none on the snapshot path | real |
| Wire bytes byte-identical to pre-change | presence handler test marshaling a fixed snapshot, asserting exact JSON | real |
| Build + tests green | `GOFLAGS=-mod=mod go build ./... && go test -count=1 ./internal/modules/iam/...` | real |

## ADR needed?

- [x] No durable decision — skip.
