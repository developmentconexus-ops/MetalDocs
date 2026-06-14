# v1 Blocker Remediation — Coordinated Fix Design

- **Date:** 2026-06-13
- **Release target:** 2026-06-14 (tomorrow)
- **Branch:** `qa/iam-area-membership` (HEAD `eae488e7c`)
- **Trigger:** Independent adversarial re-verification returned **NOT READY for v1** — 92 Wave-H holds re-confirmed, 26 NEW defects (2 critical, 8 major, 16 minor).
- **Disposition decisions:** defer appetite = **Maximal**; commit shape = **pure blocker commits + separate bounded hygiene commits**.

## 1. Mission & Success Criteria

Remediate the 6 v1 blockers and clear every fix-now-eligible defer so the branch reaches **READY for v1**, without merging — present evidence for the operator review gate.

**Success criteria (per the project close-out + evidence rules):**
- All 6 blockers fixed; each observable fix runtime-verified by us (never asked of the operator).
- One bounded commit per defect family; trackers (`wiki/backend/roadmap.md` row + `wiki/backend/_artifacts/architecture-audit-2026-06-13.md` disposition) updated in the **same** commit.
- Static gates green and **recorded** per commit: `go build ./...`, `go vet ./...`, `go test -p 2 ./...`, `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .`, `go run ./tools/cilint ./...`. Contract changes → `cd frontend/apps/web; pnpm gen:api; npx tsc --noEmit` green.
- Every TRIGGER carries a written trigger; every KEEP cites its deferred boundary.
- Do **NOT** merge. Stage only named files (`git add <path>`), never `-A`; never stage `.gitnexus/`, the `.agents/skills` worktree deletions, `AGENTS.md`, or `CLAUDE.md`.

**Hard-stop rule (carried):** if any fix balloons into shared-API / authz-internal / storage-provider / workflow-semantic redesign — STOP, record the trigger, do not symptom-patch.

## 2. The 6 Blockers → 5 Bounded Commits

### Commit 1 — `fix(presence): snake_case wire keys + username (B1+B2)`
B1 (FE casing) and B2 (BE username) ride together: same module, FE consumes BE.

**Backend (B2):**
- `internal/modules/iam/presence/model.go` — add `Username string \`json:"username"\`` to `Item`; add `Username string \`json:"username,omitempty"\`` to `Event`.
- `internal/modules/iam/presence/repository.go` — `Snapshot` SQL → `SELECT user_id, username, display_name, last_seen_at`; scan `&item.Username` in order.
- `internal/modules/iam/delivery/http/admin_handler.go` — `h.presence != nil` branch (lines 225-232) add `"username": item.Username` to the emitted map (matches the `else`/`onlineUsers` branch which already emits it).
- `internal/modules/iam/presence/hub.go` — `diff()` thread `Username: it.Username` into join (and the status-transition) events; `Username: old.Username` into leave.

**Frontend (B1):**
- `frontend/apps/web/src/features/iam/queries/usePresenceStream.ts` — replace hand camelCase `PresenceItem` with `type PresenceItem = OnlinePresenceItem & { status?: PresenceStatus }` (import `OnlinePresenceItem` from generated types); rewrite `PresenceEvent` union + `applyEvent` to snake_case (`user_id` / `display_name` / `last_seen_at` / `username`); fix the `p.userId` filter to `p.user_id`.
- `frontend/apps/web/src/features/iam/components/PresencePanel.tsx` lines 69-81 → `u.user_id` / `u.display_name` / `u.username` / `u.last_seen_at`.

**Contract:** none — `OnlinePresenceItem` already requires `username` (openapi.yaml 4058-4067).

**Verify:** `go test -p 2 ./internal/modules/iam/...`; `tsc --noEmit`; runtime — login, hit `GET /api/v1/iam/presence/snapshot`, confirm each item carries `username`; open IAM Admin Center, confirm presence panel renders names/@usernames.

### Commit 2 — `fix(audit): FE consumes snake_case AuditEventItem (B3)`
FE-only — BE already emits snake_case (commit e6423a24e).
- `useAuditEventsQuery.ts` — alias `export type AuditEventItem = components["schemas"]["AuditEventItem"]`; fix `adaptPage` bare cast (line 44) — items are already snake_case, the cast just needs the aliased type.
- `AuditEventsTable.tsx` — `ev.occurredAt→occurred_at`, `actorId→actor_id`, `resourceType→resource_type`, `resourceId→resource_id`, `traceId→trace_id` (lines 34, 44-47, 125-235). `ev.action` / `ev.payload` unchanged (same in both casings).
- `ActivityEventRow.tsx` — its **own** local type (lines 10-14) + read sites (39-67) → snake_case (or import the aliased type).
- `ActivityPanel.tsx` line 67 — `prev.actorId === ev.actorId` → `actor_id`.

**Verify:** `tsc --noEmit`; runtime — IAM Activity panel + Audit table render rows (no `undefined` cells), relative-time + trace render.

### Commit 3 — `fix(documents): form_data_json object not base64 (B4)`
- `api/openapi/v1/openapi.yaml` `DocumentSummary.form_data_json` (4292-4294) → `type: object` + `additionalProperties: true` (mirror `DocumentDetailResponse` 4349-4351).
- `go generate` → regen `api.gen.go` (`FormDataJson` becomes `map[string]interface{}`).
- `internal/modules/documents/delivery/http/handler.go` `toDocumentSummary` (403-430) → `json.Unmarshal` form bytes into a map like `toDocumentDetailResponse` (435-443); change signature to `(DocumentSummary, error)`; propagate the error in the `listDocuments` loop (276).
- `cd frontend/apps/web; pnpm gen:api; npx tsc --noEmit`.

**Verify:** `go build/vet/test`; api-lint strict; `tsc`; runtime — `GET /api/v1/documents` returns `form_data_json` as a JSON object (not a base64 string).

### Commit 4 — `fix(documents): duplicate preserves visibility (B5)`
- `apps/api/internal/wiring/documents_adapters.go` `CreateControlledDocumentCmd` literal (47-65) — add `VisibilityScope: string(source.Visibility.Scope)`, `VisibilityAreaCodes: source.Visibility.AreaCodes`, `VisibilityUserIDs: source.Visibility.UserIDs`. Field names confirmed: cmd has `VisibilityScope string`/`VisibilityAreaCodes []string`/`VisibilityUserIDs []string` (service.go 69-71); `domain.Visibility` = `{Scope VisibilityScope, AreaCodes []string, UserIDs []string}`.
- Add a regression test: duplicating a `restricted` doc with user grants preserves scope + user IDs (today `NewVisibility` defaults to `restricted`/single-area and drops user grants).

**Verify:** new test fails before / passes after; `go test -p 2 ./...`.

### Commit 5 — `fix(search): escape LIKE/ILIKE wildcards (B6)`
Mirror the canonical pattern at `internal/modules/documents/repository/repository.go:417-423` (`"%"+sqlescape.LikeEscape(q)+"%"` + `ILIKE $n ESCAPE '\'`).
- `internal/modules/search/infrastructure/v2documents/reader.go` — line 56 `LIKE '%' || $2 || '%'`: escape the bound `$2` value (line 127, `strings.ToLower(strings.TrimSpace(query.Text))`) via `sqlescape.LikeEscape` and add `ESCAPE '\'`. Preserve the `$2 = ''` empty-guard (escaping `''` stays `''`).
- `internal/modules/controlleddocuments/infrastructure/repository.go` — lines 128-129 `code ILIKE $n OR title ILIKE $n`: wrap bound value (129) with `LikeEscape`, add `ESCAPE '\'` to both predicates.

**Verify:** unit/repo tests; `go test -p 2 ./...`; a query containing `%`/`_` matches literally, not as wildcard.

## 3. Hygiene Commits (Maximal scope) — fix-now defers

Pure-family commits, separate from the blocker commits. Each updates the audit-artifact disposition in-commit.

- **H-A — `fix(audit): declare exportAuditEvents 202 status field (F-4b)`** — add the undeclared `status` field to the 202 response schema; regen; FE regen if consumed.
- **H-B — `fix(observability): RFC9457 error bodies (OBS-2/3/4)`** — bare 405 → problem+json body; `http.NotFound` plain-text (audit ~234/244) → `problem.Write`; raw `"INTERNAL_ERROR"` literals (presence/handler.go 72/78/89) → `problem.CodeInternalError` constant.
- **H-C — `fix(config): env guards (OTEL none + E2E consistency)`** — `OTEL_TRACES_EXPORTER=none` must skip SDK install (otel.go ~42); unify `METALDOCS_E2E` read (`main.go` `TrimSpace` vs `e2e_seed.go` raw `os.Getenv`).
- **H-D — `fix(iam): remove phantom POST /iam/users permission row`** — drop the catalog row (permissions.go ~103) that maps to no real route. **Gate:** verify removal trips no permission-seed CI guard before committing.
- **H-E — `chore: remove dead code (H-5.1/5.2 orphans)`** — `decision_service.go` `hasUnresolvedComments` / `loadActiveDocumentContentHash` (~554/573); `documents service.go` `ListDocuments` / `ListDocumentsForUser` / `SyncArtifactMetadata`. **Gate:** confirm zero callers (grep + GitNexus impact) before deleting.
- **H-F — `fix(metrics): declare p50/p95/p99 in MetricItem (OBS-1)`** — *contingent:* only if the handler already emits these and the schema merely under-declares them. If emitting them requires new computation → downgrade to TRIGGER (D-H1 ring-buffer boundary).
- **H-G — `fix(documents): profileDefaults real template status (profileDefaults)`** — *contingent:* `documents_adapters.go:103` hardcodes `status="published"`. Fix only if the real status is cheaply readable from the profile/template the adapter already loads. If it needs new template-status plumbing → TRIGGER.

## 4. TRIGGER (leave with written trigger) — even under Maximal

| Item | Trigger reason |
|------|----------------|
| **F-001** forged `assigned_by` (H-3b) | authz hard-stop boundary (ADR 0022). Fix = tier-2 "principal-from-`assigned_by`" redesign. Primary audit row is correct; no privilege gained. Symptom-patching authz is forbidden. |
| **F-4a** dead audit `ServerInterface` gen | D-H3 deferred boundary (raw-mux → ServerInterface migration). Removing/migrating is an architectural sweep, not a release-eve fix. |
| **Concurrency** — `presenceHub.Run/RunHeartbeat` + `startAuditRetention` goroutines not joined; `jobs.stop()` skipped on `os.Exit(1)`; `capCache assertedByTxPtr` no eviction | Not cheap-and-safe: changing shutdown ordering risks new races; `capCache` is the D-H4 boundary (`do NOT re-key capCache`). Low severity, observability/shutdown-tidiness only. |

## 5. KEEP AS-IS (pre-existing Wave-H deferred boundaries)

Do not touch unless forced: **D-H1** metrics ring-buffer; **D-H2** approval→peer module; **D-H3** raw-mux→ServerInterface; **D-H4** authz `*sql.Tx`→`db.Tx` (keep `authz.Require`/`SeedTxIdentity` taking `*sql.Tx`; do NOT re-key capCache); **D-H5** `CanDo` TTL cache; **H-5.3-D1** SQL filter+paginate; **H-3b** tier-2 principal-from-`assigned_by`.

## 6. Execution Protocol

- **Orchestration:** ultracode ON. Delegate code/doc edits to subagents (sonnet implement/review, haiku mechanical; never fable; ≤15 concurrent). Git staging/commits and static gates run directly by the orchestrator.
- **Per commit:** implement inside bounded task → run static + targeted verification for the touched slice → code review → product QA against the relevant canonical checklist → fix by root-cause family → record evidence → commit (named files only) with footer `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`.
- **Order:** B1+B2 → B3 → B4 → B5 → B6 → H-A → H-B → H-C → H-D → H-E → (H-F, H-G if verification passes). Contract-changing commits (B4, H-A, H-F) carry FE regen + `tsc` in the same commit.
- **Close-out:** append a close-out entry to `wiki/references/current-agent-handoff.md`; do NOT merge; present the commit ledger + per-commit gate evidence + TRIGGER/KEEP register for the operator review gate.

## 7. Out of Scope

Anything in §4 TRIGGER / §5 KEEP; merge to integration branch; touching `.gitnexus/`, `.agents/skills` deletions, `AGENTS.md`, `CLAUDE.md`; any redesign that crosses a §1 hard-stop boundary.
