# Feature F8.1 — Spec (APPROVED)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.1-presence-typed-body`
> **Status:** Approved 2026-06-20 (execution session) — seed confirmed against runtime truth; one acceptance criterion corrected (see Q3).
> **Approved before code:** ✅ 2026-06-20 — seed claims verified against `handler.go:83`, `model.go:54`, `api.gen.go:453-467`; criterion #2 corrected from "byte-identical" to "wire-equivalent" per established M7 precedent. No code written before this line.

> Seed contract from the post-M7 re-audit (Contract Major #4). The execution session must run the
> brainstorming interview, confirm/adjust below, then fill the approval line before code.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Does `OnlinePresenceItem` (generated) already carry every field presence `Item` emits? | **Yes** — verified `display_name, last_seen_at, status, user_id, username` present (`iam/api/api.gen.go:453-459`); presence `Item` matches (`presence/model.go:54`). No OpenAPI change needed. |
| 2 | Must the wire shape stay byte-identical? | Seed said "byte-identical." **Corrected (Q3).** |
| 3 | Does switching `presence.Item` → generated `OnlinePresenceItem` preserve literal bytes? | **No** — `presence.Item` field order is `user_id, username, display_name, last_seen_at, status` (`model.go:54`); generated `OnlinePresenceItem` is `display_name, last_seen_at, status, user_id, username` (`api.gen.go:453`). `encoding/json` emits in struct-declaration order, so per-item **key order shifts**. Key order is **not** part of the JSON (RFC 8259) or OpenAPI contract and is invisible to every key-based decoder. Acceptance corrected to **wire-equivalent** (same keys/values, `status` always present), matching the operator-ratified M7 F7.1 precedent (same caveat, same program). Also: `OnlinePresenceItem.Status` is `*OnlinePresenceItemStatus` (omitempty) — snapshot items always carry a status (off excluded), so the pointer is always set → `status` stays present. |

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
| Wire-equivalent (not byte-identical): response decodes into `iamapi.PresenceSnapshotResponse` with `DisallowUnknownFields`; same items/field-values as pre-change; `status` always present; 200 + `application/json` unchanged. Key order may shift (struct order) — semantically irrelevant per RFC 8259 + OpenAPI; M7 F7.1 precedent. | strict-decode parity-lock test on a fixed snapshot + existing `TestHandler_Snapshot_*` pass unmodified | real |
| Build + tests green | `GOFLAGS=-mod=mod go build ./... && go test -count=1 ./internal/modules/iam/...` | real |

## ADR needed?

- [x] No durable decision — skip.
