# Feature F1.2 — presence/stream contract declaration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:subagent-driven-development.
> Steps use checkbox (`- [ ]`) syntax. Contract-first; the consumer contract is in `spec.md`
> (read from source). **Strategy A** (operator-approved): declare for docs + exclude from codegen.

**Goal:** Make `GET /iam/presence/stream` tri-source-consistent (openapi ↔ FE types ↔ runtime) by
declaring it in `openapi.yaml` and excluding it from `strict-server` codegen — without touching the
live WebSocket 101 handler.

**Architecture:** Add the path + two documentary schemas to `openapi.yaml`; add
`exclude-operation-ids: [streamPresence]` to the iam oapi-codegen config so no server stub shadows the
hand-written upgrade handler; regen server stub (embedded-spec blob only) + FE types (path + schemas).
Runtime handler unchanged → 101 stays live.

**Tech Stack:** OpenAPI 3.0.3, oapi-codegen v2.7.0 (`go generate`), openapi-typescript (FE `gen:api`),
Go `net/http`, `coder/websocket`.

**Prerequisite note (do NOT skip):** FE tooling is already installed at
`frontend/apps/web/node_modules/.bin/` (`openapi-typescript`, `tsc`). **Run those binaries directly.
Do NOT run `pnpm install`** — a fresh install risks the carried HS-2 eigenpal `file:` path defer.
`pnpm` is not on PATH in bash; invoke `node_modules/.bin/<tool>` (or pwsh `pnpm` if present).

---

### Task 1: Declare the route + frame schemas in openapi.yaml

**Files:**
- Modify: `api/openapi/v1/openapi.yaml` (add path after the `getPresenceSnapshot` block ~line 544; add
  two schemas near `OnlinePresenceItem` ~line 4068)

- [ ] **Step 1: Add the two documentary schemas** (in `components/schemas`, adjacent to `OnlinePresenceItem`). They mirror `internal/modules/iam/presence/model.go:55-71` exactly. `PresenceStreamItem` = `OnlinePresenceItem` **+ `status`** — do NOT edit `OnlinePresenceItem` itself.

```yaml
    PresenceStreamItem:
      type: object
      required: [user_id, username, display_name, last_seen_at, status]
      properties:
        user_id: { type: string }
        username: { type: string }
        display_name: { type: string }
        last_seen_at: { type: string, format: date-time }
        status: { type: string, enum: [online, idle] }
    PresenceStreamEvent:
      type: object
      description: One frame on the presence WebSocket. `type` discriminates; `presence` is present only on the initial `snapshot` frame.
      required: [type]
      properties:
        type: { type: string, enum: [snapshot, join, leave, online, idle, heartbeat] }
        user_id: { type: string }
        username: { type: string }
        display_name: { type: string }
        last_seen_at: { type: string, format: date-time }
        status: { type: string, enum: [online, idle] }
        presence:
          type: array
          items: { $ref: '#/components/schemas/PresenceStreamItem' }
```

- [ ] **Step 2: Add the path** (mirror the `getPresenceSnapshot` block style at `openapi.yaml:527`). A `101` response carries no body in OAS3 — the frame shape is documented via the operation `description` + the `x-websocket-message` extension (documentary; ignored by codegen).

```yaml
  /iam/presence/stream:
    get:
      summary: Online presence live stream (WebSocket upgrade)
      description: >
        Upgrades to a WebSocket (HTTP 101) and streams presence frames: an initial
        `snapshot`, then `join`/`leave`/`online`/`idle` deltas and periodic `heartbeat`.
        Each frame conforms to PresenceStreamEvent. Not a request/response operation —
        excluded from server codegen (exclude-operation-ids: streamPresence); the
        hand-written upgrade handler is authoritative. HTTP fallback: /iam/presence/snapshot.
      tags: [iam]
      operationId: streamPresence
      x-websocket-message:
        $ref: '#/components/schemas/PresenceStreamEvent'
      responses:
        '101':
          description: Switching Protocols — WebSocket established; server streams PresenceStreamEvent frames.
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
```

- [ ] **Step 3: Verify the spec parses** (before any regen):

Run (from repo root): `go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --version` then a dry parse: `go generate ./internal/modules/iam/api/...`
Expected: no YAML/parse error. (Generation result is checked in Task 2.) If it errors on the spec, fix the YAML before continuing.

- [ ] **Step 4: Re-derive the emitted `type` values from the producer** to confirm the enum matches reality (consumer-contract-first = match the producer): read `internal/modules/iam/presence/handler.go` for every `Event{Type: "…"}` / type string written. The enum in Step 1 must equal the set the handler actually emits (`snapshot`, `join`, `leave`, `online`, `idle`, `heartbeat`). Adjust the enum if the handler differs; do not invent values.

- [ ] **Step 5: Commit**

```bash
git add api/openapi/v1/openapi.yaml
git commit -m "feat(openapi): declare GET /iam/presence/stream (WS 101) + frame schemas"
```

---

### Task 2: Exclude from server codegen + regenerate the iam stub

**Files:**
- Modify: `internal/modules/iam/api/cfg.yaml`
- Regenerate (do not hand-edit): `internal/modules/iam/api/api.gen.go`

- [ ] **Step 1: Add the exclusion** to `cfg.yaml` `output-options` (keep `include-tags: [iam]`):

```yaml
output-options:
  include-tags:
    - iam
  exclude-operation-ids:
    - streamPresence
```

- [ ] **Step 2: Regenerate**

Run: `go generate ./internal/modules/iam/api/...`
Expected: completes with no error.

- [ ] **Step 3: Assert the server stub is handler-clean** (the 101 handler must NOT be generated/shadowed):

Run: `grep -nE 'PresenceStream|streamPresence|presence/stream' internal/modules/iam/api/api.gen.go`
Expected: matches appear ONLY inside the embedded-spec blob (the embedded `openapi.yaml` string) — there is **no** `StreamPresence` ServerInterface method and **no** `m.HandleFunc(... "/iam/presence/stream" ...)` registration. If a ServerInterface method or a mux registration appears, the exclusion failed — STOP (do not hand-delete generated code; fix `cfg.yaml`).

- [ ] **Step 4: Build + determinism check**

Run: `go build ./...`  → Expected: exit 0.
Run again: `go generate ./internal/modules/iam/api/... && git diff --stat internal/modules/iam/api/api.gen.go`
Expected: second generate produces **no** diff (deterministic).

- [ ] **Step 5: Commit**

```bash
git add internal/modules/iam/api/cfg.yaml internal/modules/iam/api/api.gen.go
git commit -m "build(iam): exclude streamPresence (WS) from oapi-codegen; regen embedded spec"
```

---

### Task 3: Regenerate FE types + typecheck

**Files:**
- Regenerate (do not hand-edit): `frontend/apps/web/src/lib/api-types/index.d.ts`

- [ ] **Step 1: Regen FE types** (run the installed binary directly — NO `pnpm install`):

Run (from `frontend/apps/web`): `node_modules/.bin/openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts`
Expected: writes `index.d.ts`, no error.

- [ ] **Step 2: Assert the path + schemas landed**

Run: `grep -nE '"/iam/presence/stream"|PresenceStreamEvent|PresenceStreamItem' frontend/apps/web/src/lib/api-types/index.d.ts`
Expected: the path key `"/iam/presence/stream"` is present, and both schemas appear under `components["schemas"]`.

- [ ] **Step 3: Typecheck (FE tsc 0)**

Run (from `frontend/apps/web`): `node_modules/.bin/tsc --noEmit -p tsconfig.build.json`
Expected: **0 errors**. (The hand-written `usePresenceStream.ts` is NOT refactored; it must still compile against the regenerated types — `PresenceItem = OnlinePresenceItem & { status? }` is unaffected since `OnlinePresenceItem` is unchanged.)

- [ ] **Step 4: Commit**

```bash
git add frontend/apps/web/src/lib/api-types/index.d.ts
git commit -m "chore(fe): regen api-types — add /iam/presence/stream path + frame schemas"
```

---

### Task 4: Route-truth-table, collateral-diff guard, runtime 101 proof, retire wont-fix

**Files:**
- Create: (evidence captured by controller into `evidence.md`)
- Modify (light): retire the archived wont-fix note `wiki/_archive/backlog/api-contract-hardening.md:129`
  (or add a superseding line) — record Strategy A as the precedent. Heavy wiki refresh is deferred to
  `wiki-curator` at milestone close.

- [ ] **Step 1: AC4 — collateral-diff guard.** Confirm the regen changed ONLY the stream additions.

Run: `git diff <pre-F1.2-sha> -- api/openapi/v1/openapi.yaml frontend/apps/web/src/lib/api-types/index.d.ts`
Expected: only the `/iam/presence/stream` path + `PresenceStreamEvent`/`PresenceStreamItem` schemas added; **no** other path/operation/schema changed. If any other route moved → HS-6 STOP.

- [ ] **Step 2: AC1 — build the route-truth-table** for `/iam/presence/stream`:

| Source | Present? | Evidence |
|--------|----------|----------|
| runtime | yes | `presence/handler.go:66` registration; `permissions.go:247` guard |
| openapi.yaml | yes | Task 1 path |
| FE types | yes | Task 3 grep |
| server stub | **excluded (documented)** | `cfg.yaml exclude-operation-ids: [streamPresence]`; grep shows no handler |

All three that *can* agree (openapi/FE-type/runtime) agree on method `GET` + guard `user.view`; the stub exclusion is explicit, not silent drift. (Was: runtime-only.)

- [ ] **Step 3: AC5 — runtime 101 proof (controller-run).** Start the API and prove the upgrade still works.

Run: `.\scripts\start-api.ps1 -Build`; login (`POST /api/v1/auth/login`, capture session cookie); open a WebSocket to `ws://localhost:8081/api/v1/iam/presence/stream` with the session cookie + `Origin: http://localhost:8081`.
Expected: server responds **HTTP 101 Switching Protocols** and sends an initial `{"type":"snapshot",...}` frame. Capture the 101 handshake + first frame. (Use a WS client: `websocat`, a short Go client, or node `ws`. A plain `curl --include --no-buffer -H 'Upgrade: websocket' -H 'Connection: Upgrade' -H 'Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==' -H 'Sec-WebSocket-Version: 13'` with cookie+Origin will show the `HTTP/1.1 101` status line.)

- [ ] **Step 4: Retire the wont-fix note.** Edit `wiki/_archive/backlog/api-contract-hardening.md:129` to record that `/iam/presence/stream` is now **declared-for-docs + codegen-excluded (Strategy A, F1.2, 2026-06-14)**, superseding "leave undocumented".

- [ ] **Step 5: Commit**

```bash
git add wiki/_archive/backlog/api-contract-hardening.md
git commit -m "docs(wiki): retire presence/stream wont-fix — superseded by F1.2 declare+exclude"
```

---

## Self-Review (run before dispatching)

- **Spec coverage:** AC1 (route-truth-table) → Task 4 Step 2. AC2 (handler-clean regen) → Task 2 Steps 3-4.
  AC3 (FE gen + tsc 0) → Task 3. AC4 (no collateral shift) → Task 4 Step 1. AC5 (runtime 101) → Task 4 Step 3. All covered.
- **No placeholders:** every step has the exact YAML/command + expected output.
- **Type consistency:** schema field names/json tags match `model.go:55-71`; enum `type` values re-derived
  from the handler in Task 1 Step 4; `operationId: streamPresence` is the single string used in both
  `openapi.yaml` and `cfg.yaml exclude-operation-ids`.
- **Non-goals guarded:** no WS-handler edit, no `usePresenceStream` refactor, no `OnlinePresenceItem`
  mutation, no `pnpm install`.
