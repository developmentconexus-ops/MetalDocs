# Feature F8.1 — Evidence (presence snapshot typed body)

> **Milestone:** 8  ·  **Feature:** `f8.1-presence-typed-body`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (approved 2026-06-20). Plan: `plan.md`.
> **Commit:** recorded at commit time below.

## What was implemented

- `handleSnapshot` ([`iam/presence/handler.go:81`](../../../../../internal/modules/iam/presence/handler.go))
  now emits the generated `iamapi.PresenceSnapshotResponse` typed body via `httpresponse.WriteJSON`,
  replacing the `map[string]any{"items": items}` response literal (was `:83`).
- New unexported `toPresenceSnapshotResponse([]Item) iamapi.PresenceSnapshotResponse` maps each
  `presence.Item` → `iamapi.OnlinePresenceItem`, always setting the `Status` pointer (snapshot items
  are never "off" — excluded upstream — so the `*omitempty` field always serialises and `status` stays
  on the wire).
- The generated `iamapi.PresenceSnapshotResponse` / `OnlinePresenceItem` models were **dead code**
  (unused); this feature makes them live. No OpenAPI change, no codegen regen (fields already aligned —
  verified `api.gen.go:453-467` vs `model.go:54`).

**Producer matches consumer contract:** consumer is the web presence view expecting
`200 application/json` `{ "items": OnlinePresenceItem[] }` per OpenAPI `PresenceSnapshotResponse`. The
producer now emits exactly that generated type. The WS `/stream` route and its `Event` envelope are
untouched (non-goal honored).

## Wire-equivalence (not behavior change)

Wire-equivalent, not byte-identical (one honest caveat, identical to the operator-ratified M7 F7.1):
- **Key ordering** shifts from `presence.Item` declaration order (`user_id, username, display_name,
  last_seen_at, status`) to `OnlinePresenceItem` order (`display_name, last_seen_at, status, user_id,
  username`). Semantically irrelevant per JSON (RFC 8259) and the OpenAPI contract; invisible to every
  key-based decoder — all existing presence tests pass unmodified.
- **`status` presence** is preserved: always set, asserted non-nil in the parity-lock test.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Parity-lock (characterization) test green pre-refactor | `go test -run TestHandler_Snapshot_TypedBody ./internal/modules/iam/presence/` | `ok` (3.023s) — generated type already matches the wire key-set | real |
| H-D red→green (response literal removed) | `grep -nE '(Encode\|WriteJSON)\(.*map\[string\]any' internal/modules/iam/presence/handler.go` | baseline **1 hit** (`:83`) → after **0 hits (exit 1)** | real |
| Parity-lock test green post-refactor (no drift) | `go test -count=1 ./internal/modules/iam/presence/` | `ok` (2.204s) — incl. `TestHandler_Snapshot_TenantScoped`, `TestStream_SnapshotMatchesHTTP` unmodified | real |
| Static (build) | `GOFLAGS=-mod=mod go build ./...` | exit 0 | — |
| iam module regression | `GOFLAGS=-mod=mod go test -count=1 ./internal/modules/iam/...` | all packages `ok` (application, authz, delivery/http, domain, infra/memory, infra/postgres, presence) | real |

> The feature's genuine red→green is the H-D grep (1→0). The strict-decode test is a parity-lock
> characterization test (green before and after) — its teeth are catching any field drop / renamed field /
> lost `status` the typed swap could introduce. No fabricated red (a pure typed-parity refactor has no
> observable contract change to drive a red behavioral test — same honest stance as M7 F7.1).

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| No response `map[string]any` in presence handler | yes | grep 1→0 (only a code comment mentions the old pattern, not a response literal) |
| Wire-equivalent: strict-decodes into `PresenceSnapshotResponse`, same items/values, `status` present, 200 + `application/json` | yes | `TestHandler_Snapshot_TypedBody` (DisallowUnknownFields, asserts fields + `status` non-nil + content-type) |
| Build + tests green | yes | `go build ./...` exit 0; `go test ./internal/modules/iam/...` all `ok` |

## Review disposition

- Spec-compliance review: PASS — producer emits the contracted generated type; consumer contract honored;
  non-goals (WS route, OpenAPI schema, StrictServerInterface) untouched.
- Code-quality review: PASS — mapping extracted to a documented helper; dead generated model now live;
  no new `map[string]any`; import set minimal (`encoding/json` retained — still used by WS `writeJSON`).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| None | | |
