# Feature F1.2 — Evidence (presence/stream contract declaration)

> **Milestone:** 1 (Reach-A Blockers)  ·  **Feature:** `f1.2-presence-stream-spec`  ·  **Closed:** 2026-06-14
> **Strategy:** A (operator-approved 2026-06-14) — declare the live WebSocket route in `openapi.yaml`
> for documentation + exclude it from `strict-server` codegen so the hand-written 101 upgrade handler
> stays authoritative. Closes Grade-A blocker #3 (H-D contract tri-source drift, stronger form):
> `GET /iam/presence/stream` was live + FE-consumed but absent from openapi.yaml / stub / FE-types.

## What was implemented

The route `GET /iam/presence/stream` (a `coder/websocket` 101 upgrade, `handler.go:97`) is now
**contract-declared without touching the runtime handler**:

1. **Declared** in canonical `api/openapi/v1/openapi.yaml`: the path (`operationId: streamPresence`,
   `tags:[iam]`, `101` response, `x-websocket-message` ref) + two documentary schemas
   `PresenceStreamEvent` / `PresenceStreamItem` mirroring the producer `model.go` exactly. The frame
   `type` enum (`snapshot|join|leave|online|idle|heartbeat`) was re-derived from the real producer
   (`handler.go:117` + `hub.go:230,349,360,372`), not guessed.
2. **Excluded from server codegen**: `internal/modules/iam/api/cfg.yaml` gained
   `exclude-operation-ids: [streamPresence]` (kept `include-tags: [iam]`); `api.gen.go` regenerated.
   No `StreamPresence` ServerInterface method, no mux registration → the hand-written handler is not
   shadowed. The documentary schemas are codegen-pruned from the Go stub.
3. **FE types regenerated**: `frontend/apps/web/src/lib/api-types/index.d.ts` gained the path + both
   schemas (openapi-typescript ignores the codegen exclusion → the consumer-type half of tri-source).

The WS handler, `usePresenceStream.ts`, and `OnlinePresenceItem` were **not** modified (non-goals).

Commits (all on `main`, baseline `821c09e0` = F1.1 close):
- `e0a3b081` — feat(openapi): declare GET /iam/presence/stream (WS 101) + frame schemas
- `6136c85e` — fix(openapi): add 500 resp, clarify event-frame desc, block-style date-time (review fixes)
- `040e67b5` — build(iam): exclude streamPresence from oapi-codegen; regen embedded spec
- `7664cdc1` — chore(fe): regen api-types — add /iam/presence/stream path + frame schemas

## AC1 — Route-truth-table for `/iam/presence/stream`

| Source | Present? | Evidence |
|--------|----------|----------|
| runtime | yes (`GET`, guard `user.view`) | registration `internal/modules/iam/presence/handler.go:66` (`mux.HandleFunc("GET /api/v1/iam/presence/stream", h.handleStream)`); guard `apps/api/cmd/metaldocs-api/permissions.go:247` (`CapUserView`, `VisibilityPermissionGuarded`) |
| openapi.yaml | yes | path block + `PresenceStreamEvent`/`PresenceStreamItem` schemas (commits `e0a3b081`+`6136c85e`) |
| FE types | yes | `index.d.ts:368` path key + `:2146` `PresenceStreamItem` + `:2156` `PresenceStreamEvent` (commit `7664cdc1`) |
| server stub | **excluded (documented)** | `cfg.yaml exclude-operation-ids: [streamPresence]`; `grep StreamPresence api.gen.go` → no ServerInterface method / no registration; path survives only in the encoded embedded-spec blob |

The three sources that *can* agree (openapi / FE-type / runtime) agree on method `GET` + guard
`user.view`; the stub exclusion is **explicit and visible**, not silent drift. (Was: runtime-only.)

## Verification

| Check | Command / action | Result | Real vs fixture |
|-------|------------------|--------|-----------------|
| AC2 — handler-clean stub | `grep -nE 'StreamPresence\|streamPresence' internal/modules/iam/api/api.gen.go` | no ServerInterface method, no registration; the only readable presence route is `GetPresenceSnapshot`/`/iam/presence/snapshot`. Decoded embedded blob: contains the path key `/iam/presence/stream` but `streamPresence`+both schemas codegen-pruned | real |
| AC2 — deterministic regen | `go generate ./internal/modules/iam/api/...` ×2 → `git status --porcelain api.gen.go` | empty on 2nd run (byte-identical); `go build ./...` exit 0; `go.sum` not churned | real |
| AC3 — FE codegen + typecheck | `node_modules/.bin/openapi-typescript … -o index.d.ts`; `tsc --noEmit -p tsconfig.build.json` | path + both schemas land; **tsc 0 errors** (re-run independently by reviewer) | real |
| AC4 — no collateral shift | `git diff 821c09e0 HEAD -- openapi.yaml index.d.ts` | **pure additive: 116 insertions, 0 deletions**; only the stream path + 2 schemas; no other path/op/schema changed; `OnlinePresenceItem` untouched | real |
| AC5 — runtime 101 live | `start-api.ps1 -Build` → login → `ClientWebSocket` (session cookie + `Origin`) to `/api/v1/iam/presence/stream` | **HTTP handshake = `SwitchingProtocols` (101)**, state Open, first frame a real `snapshot` (see below) | real |

### AC5 runtime proof (live `:8081`, observed by us)

```
LOGIN status: 200    SESSION cookie: metaldocs_session
WS HTTP handshake status: SwitchingProtocols      WS State: Open
FIRST FRAME (204 bytes, Text):
{"type":"snapshot","last_seen_at":"0001-01-01T00:00:00Z","presence":[{"user_id":"admin",
 "username":"admin","display_name":"Administrator","last_seen_at":"2026-06-14T22:43:40.215842Z",
 "status":"online"}]}
```

This closes the long-standing `[runtime-unverified]` residual (`wiki/backend/wave-h-plan.md:32`): proven
**101, not 501**, on a live build, with a `snapshot` frame whose shape matches the declared
`PresenceStreamEvent` (`type` + `presence[]` of `PresenceStreamItem` = user_id/username/display_name/
last_seen_at/status) — acceptance is the successful upgrade + first frame, not "looks wired".

## Acceptance vs spec Validation Gate

| AC (from spec.md) | Met? | Evidence |
|-------------------|------|----------|
| AC1 — route-truth-table; openapi↔FE↔runtime agree; stub intentionally excluded | yes | AC1 table above |
| AC2 — oapi-codegen deterministic & handler-clean | yes | verification rows; both Task-2 reviews SPEC-COMPLIANT + clean |
| AC3 — FE gen clean; type present; tsc 0 | yes | verification row; both Task-3 reviews SPEC-COMPLIANT + clean (independent tsc 0) |
| AC4 — no collateral contract shift | yes | pure-additive diff, 0 deletions |
| AC5 — runtime 101 still live | yes | runtime proof block (101 + snapshot frame) |

## Review disposition (subagent-driven, two-stage per task)

- **Task 1** (openapi declaration): spec-compliance ✅; code-quality found 4 (1🔴 2🟡 1🔵) → classified by
  root-cause family: **applied 3** (add `500` response, reword event-frame description, block-style
  `date-time` to match `OnlinePresenceItem`) as `6136c85e`; **rejected 2** with rationale ($allOf for
  PresenceStreamItem, oneOf/discriminator for PresenceStreamEvent — producer structs are flat with
  omitempty; FE uses a hand-written union over the raw WebSocket; documentary schemas are pruned from
  Go). Re-review clean.
- **Task 2** (codegen exclude + regen): spec-compliance **SPEC-COMPLIANT** (cfg knob correct; no stub
  method/registration; deterministic; build 0; 2-file commit, no go.sum churn); code-quality **clean**
  (diff is pure embedded-blob recompress; `exclude-operation-ids` is the correct idiomatic v2 knob).
- **Task 3** (FE regen + tsc): spec-compliance **SPEC-COMPLIANT** (single-file commit; types mirror
  model.go + Task-1 yaml exactly; independent tsc 0; consumer untouched); code-quality **clean**
  (genuine openapi-typescript output, no hand-edits).

## Bounded defers / notes

| Item | Why bounded | Trigger / owner |
|------|-------------|-----------------|
| **Environment (not a code change):** the FE pnpm virtual store had ~1008 missing directory junctions on this machine (degraded-SSD), so the installed binaries could not run until the implementer restored the junctions with `mklink /J`. This is gitignored `node_modules` only — **no tracked-file collateral**, not in any commit. It restored the plan's "tooling already installed" precondition that had drifted. | If FE tooling fails to run in a future session, re-materialize the pnpm store (`pnpm install --frozen-lockfile`) once the eigenpal `file:` defer is resolved; owner = whoever next does FE codegen |
| Snapshot frame carries a top-level zero-value `last_seen_at` (`0001-01-01…`) alongside the populated per-user `last_seen_at`. Within the declared contract (`PresenceStreamEvent.last_seen_at` is optional). | Producer quirk, not a contract defect; F1.2 is declaration-only and does not modify the producer | If undesired, trim in a producer-side cleanup (out of F1.2 scope); owner = backend agent |
| Lightweight decision record (WS/streaming routes = declare-for-docs + codegen-exclude) to be folded into wiki decisions / `api-contract.md` | Heavy wiki refresh is deferred to `wiki-curator` at milestone close per the milestone workflow | M1 close (`wiki-curator`); the archived wont-fix note is already retired (`api-contract-hardening.md:129`) |
