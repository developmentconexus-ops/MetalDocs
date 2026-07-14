# PLAN 1 — F-001: Fix Tier-1 Authz Rule Conflation (CRITICAL)

## Context

F-001 is the audit finding that the central Tier-1 declarative authz table in `apps/api/cmd/metaldocs-api/permissions.go` conflates read and write capabilities on multiple writable prefixes, and in some cases declares method-less rows on prefixes that have write verbs. This violates ADR 0007 (`wiki/concepts/authz-tiers.md`) by causing GET requests to be gated on Manage/Submit caps and, conversely, allowing methodless prefix rows to shadow per-verb intent. Tier-2 area enforcement and the Postgres tripwire are unaffected.

## Chosen Approach — Option A (Surgical Per-Row Method Split)

Rewrite the offending rows in `permissions.go` so every writable prefix has at least one explicit read row (View-grade cap) and one or more explicit write rows (Manage/Submit-grade caps). No new helpers, no schema change, no codegen change, no `methodSet` field. Mirrors the already-correct legacy-alias pattern at `permissions.go:104-117`.

## Justification — Why Option A Beats B and C

| Criterion | A (winner, 23) | B (per-module packages, 20) | C (OpenAPI-derived authz, 12) |
|---|---|---|---|
| Canonical fit | 5 — preserves ADR 0007 Tier-1 declarative table | 5 — aligns ADR 0012 but redesign-grade | 2 — ADR 0007 explicitly rejects this |
| Blast radius | 5 — single file | 2 — new platform package + every module | 1 — codegen + middleware rewire |
| Recurrence prevention | 4 — CI shadowing invariant + positive matrix | 5 — colocated registry | 4 |
| Onboarding cost | 5 — one greppable file | 3 — new abstraction | 2 |
| SaaS long-vision | 4 — Option B can layer on later | 5 — scales with modules | 3 — blocked by ADR 0007 |

- F-001 is a bounded authz-rule-conflation bug, not an architecture problem. CLAUDE.md §5.3 (surgical changes) and §4 (hard-stop on redesign-grade work for symptom fixes) both favor A.
- B is strictly better for long-term scaling but its blast radius makes it a redesign-grade change for a focused fix; it should be filed as a follow-up ADR after A lands.
- C is blocked by ADR 0007's documented rejection (`*sql.Tx` unavailable pre-handler; Tier-2 + tripwire already provide the static guarantee). Adopting C would require an ADR reversal.

## Implementation Steps

1. **Freeze the audit as a truth table.** Add a comment block at the top of `permissions.go` (or a sibling `wiki/references/permissions-rules-truth-table.md`) listing all 13 evidence rows: route, current cap, intended read cap, intended write cap. The diff must satisfy this contract row-for-row.
2. **Verify read caps exist in the registry — do not invent.** Confirm presence of `CapDocumentView`, `CapMembershipView`, `CapUserView` (and any others needed). If a required read cap is missing, **STOP** and re-route to `runtime-contract-prereq` plus a new ADR proposing the cap. No silent invention (ADR 0007 registry rule, CLAUDE.md §4 hard-stop).
3. **Rewrite each F-001 row in place.** For every entry in the audit, replace the conflated row with: one `{method: GET, pattern, cap: <readCap>}` row plus one row per write verb (`POST|PUT|PATCH|DELETE`) with the Manage/Submit cap. Specifically address: `permissions.go:78` (metrics), `:82-83` (notifications, including inverse-F-001 POST `.../read`), `:86-87` (access-policies), `:94-101` (iam/users + subroutes), `:159-175` (taxonomy/{profiles,areas,families}), `:185` (iam/area-memberships — remove the methodless row, replace with per-verb rows).
4. **Add a shadowing-invariant test.** Create `TestPermissionsTable_NoMethodlessWriteShadowing` in `apps/api/cmd/metaldocs-api/permissions_test.go`. Iterate the rule slice and FAIL if any rule has empty `method` AND cap name contains `Manage` or `Submit`, OR if any methodless prefix overlaps a path declared with a write verb elsewhere.
5. **Add a positive (method, path) → cap matrix test.** One assertion per audit row, using the production rule-resolver function the middleware calls. No parallel implementation.
6. **Update `wiki/concepts/authz-tiers.md`** with a new "Tier-1 rule authoring rules" subsection: (a) never declare Manage/Submit on GET; (b) never omit `method` on a prefix that has any write verbs; (c) each writable resource declares ≥1 read row and ≥1 write row; (d) reference the new shadowing test as enforcement. Bump `Last verified`.
7. **Targeted verification.** Run `go test ./apps/api/cmd/metaldocs-api/... -run Permissions`. Start API via `.\scripts\start-api.ps1 -Build`. With a viewer-only test user, curl each affected GET (expect 200) and each write verb (expect 403 `FORBIDDEN_CAPABILITY`). Capture curl evidence in PR description.
8. **Dispatch `wiki-curator`** to refresh anchors in `wiki/concepts/authz-tiers.md` and module wikis for notifications, access-policies, taxonomy, iam/users, iam/area-memberships, metrics. Bump `Last verified` stamps.

## Risk Register

| ID | Severity | Risk | Mitigation |
|---|---|---|---|
| R1 | High | Required read cap missing from registry | Step 2 halts and re-routes to `runtime-contract-prereq` + ADR |
| R2 | Medium | Tier-2 `authz.Require` call sites assume the write cap is present on GET | Grep call sites for affected handlers; verify each requests correct cap; tripwire surfaces mismatches as `FORBIDDEN_AREA` on mutation, not on read |
| R3 | Medium | Frontend roles previously implicitly authorized for POSTs (notifications `.../read`, iam/users PATCH) lose access | Enumerate impacted roles in PR; coordinate with `metaldocs-tanstack-query` skill owner |
| R4 | Low | Future methodless prefix re-introduction | Step 4 invariant test |
| R5 | Low | Metrics cap is partly a product decision | Confirm with product owner before flipping; document explicitly |
| R6 | Low | Approval legacy mount (`:191-194`) is out of scope | Explicit defer with wiki note; do not touch in this PR |

## Regression Test Plan

- **Unit (`permissions_test.go`):** positive (method, path) → cap assertion for every audit row using the production rule-resolver.
- **Unit:** shadowing invariant test (Step 4).
- **Integration:** start API with `.\scripts\start-api.ps1 -Build`; with viewer-only user assert 200 on each affected GET and 403 `FORBIDDEN_CAPABILITY` on each write verb.
- **Cross-boundary regression:** run existing approval, taxonomy, and IAM handler tests to confirm Tier-2 area checks still pass for users who legitimately hold the write cap.
- **Frontend smoke:** with `metaldocs-tanstack-query` skill, confirm no caller relies on implicit GET-via-Manage-cap access.

## Rollback Plan

- Single-file diff; revert is a single `git revert <sha>` of the `permissions.go` + `permissions_test.go` + wiki commit.
- No data migration, no codegen artifact change, no client-visible schema change → instant rollback.
- If rollback occurs, re-open F-001 with notes from the failure mode that triggered rollback.

## Success Criteria

- All 13 audit rows have explicit per-method rows with correct read/write cap pairing.
- New shadowing invariant test exists and passes.
- New positive (method, path) → cap matrix test exists and passes.
- Viewer-only user gets 200 on listed GETs and 403 `FORBIDDEN_CAPABILITY` on listed writes (curl evidence in PR).
- `wiki/concepts/authz-tiers.md` updated with authoring rules and bumped `Last verified`.
- No regression in existing approval/taxonomy/IAM test suites.

## Files Touched

- `apps/api/cmd/metaldocs-api/permissions.go` (primary)
- `apps/api/cmd/metaldocs-api/permissions_test.go` (invariant + matrix tests)
- `wiki/concepts/authz-tiers.md` (authoring rules + `Last verified`)
- `wiki/README.md` (index touch if curator deems necessary)
- Module wikis listed in Step 8 (anchor + `Last verified` only, by `wiki-curator`)

## Wiki Updates Needed (per CLAUDE.md drift policy)

- `wiki/concepts/authz-tiers.md` — add Tier-1 authoring rules subsection; reference new shadowing test.
- Module wikis for notifications, access-policies, taxonomy, iam/users, iam/area-memberships, metrics — bump `Last verified` stamps and refresh route anchors.
- Optional: new `wiki/references/permissions-rules-truth-table.md` capturing the 13-row contract.

## Agent Dispatch

- **`runtime-contract-prereq`** — only if Step 2 reveals a missing read cap (prerequisite repair).
- **`code-reviewer`** — after `permissions.go` and tests are written.
- **`security-reviewer`** — mandatory; this is auth code (CLAUDE.md global security trigger).
- **`tdd-guide`** — for the invariant + matrix tests (write tests first against the desired truth table).
- **`wiki-curator`** — after merge, per CLAUDE.md drift policy.

## Spawn Task Prompt (self-contained)

```
Fix F-001 in MetalDocs: the central Tier-1 declarative authz table in apps/api/cmd/metaldocs-api/permissions.go conflates read and write capabilities on multiple writable prefixes, and contains methodless rows on prefixes with write verbs. Affected rows are at permissions.go:78 (metrics), :82-83 (notifications, including POST /api/v1/notifications/{id}/read), :86-87 (access-policies), :94-101 (iam/users and subroutes), :159-175 (taxonomy profiles/areas/families), :185 (iam/area-memberships methodless prefix). The correct precedent pattern is the legacy-alias block at permissions.go:104-117 — every writable prefix declares an explicit GET row with a View-grade cap and explicit per-verb rows for writes with Manage/Submit-grade caps.

Constraints: surgical change, single file plus tests. Do NOT introduce a methodSet field, do NOT regenerate authz from OpenAPI security (ADR 0007 rejects this), do NOT extract per-module authz packages (out of scope for this fix), do NOT touch Tier-2 enforcement or the Postgres tripwire, do NOT touch the approval legacy mount at permissions.go:191-194.

Required steps:
1. Read wiki/concepts/authz-tiers.md and ADR 0007 first.
2. Build a truth table of all 13 audit rows (route, current cap, intended read cap, intended write cap) as a comment block at the top of permissions.go or as wiki/references/permissions-rules-truth-table.md.
3. Verify every intended read cap exists in the capability registry. If any required read cap (e.g. CapMembershipView, CapUserView) is missing, STOP — do not invent caps. Re-route to runtime-contract-prereq skill and file an ADR proposing the cap before continuing.
4. Rewrite each affected row: one {method: GET, cap: readCap} row plus one row per write verb with the Manage/Submit cap. Preserve relative ordering.
5. Add TestPermissionsTable_NoMethodlessWriteShadowing in apps/api/cmd/metaldocs-api/permissions_test.go that fails if any rule has empty method AND cap name contains "Manage" or "Submit", or if any methodless prefix overlaps a path declared with a write verb elsewhere.
6. Add a positive (method, path) → cap matrix test using the production rule-resolver function the middleware calls. One assertion per audit row.
7. Update wiki/concepts/authz-tiers.md with a new "Tier-1 rule authoring rules" subsection and bump Last verified.
8. Verify: go test ./apps/api/cmd/metaldocs-api/... -run Permissions. Start the API via .\scripts\start-api.ps1 -Build. With a viewer-only user, curl each affected GET (expect 200) and each write verb (expect 403 FORBIDDEN_CAPABILITY). Capture curl evidence in PR description.
9. After merge, dispatch the wiki-curator agent to refresh anchors and Last verified stamps in module wikis for notifications, access-policies, taxonomy, iam/users, iam/area-memberships, metrics.

Honor CLAUDE.md gates: bias to caution (§5), surgical changes (§5.3), hard-stop on redesign-grade scope creep (§4), record verification evidence before claiming closure (§4 Evidence rule). Use security-reviewer and code-reviewer agents before opening the PR. Do not skip hooks or amend prior commits.

Success criteria: all 13 audit rows split correctly, new shadowing invariant test passes, positive matrix test passes, viewer-only user smoke test produces the expected 200/403 split, wiki/concepts/authz-tiers.md updated with authoring rules.
```

---

# PLAN 2 — F-002: Reorder Document-Scoped Signoff Idempotency Check (MEDIUM)

## Context

`SignoffByDocumentHandler` in `internal/modules/documents/approval/http/doc_approval_handler.go` runs state/eligibility validation BEFORE the idempotency replay check. As a result, a second call sharing `(tenant, actor, route, idempKey, payloadHash)` returns an error after the underlying instance has moved to a terminal state or the active stage has rotated past the original actor, instead of replaying the cached outcome. This violates RFC-style idempotency semantics. The stage-scoped handler is correct by construction and is out of scope.

## Chosen Approach (Verbatim from F-002 Design)

Reorder the document-scoped signoff handler so the idempotency replay check sits BETWEEN the instance load and the state/eligibility validation. This guarantees that any second call sharing `(tenant, actor, route, idempKey, payloadHash)` is served the cached outcome even after the underlying instance has moved to a terminal state or the active stage has rotated past the original actor.

The stage-scoped handler is left alone because (a) it has no instance load in the HTTP layer by design, (b) its `payloadHash` uses path-supplied IDs (empty docID + path instanceID/stageID) so replay-slot computation does not depend on a load, and (c) its state/eligibility validation is intentionally delegated to the application layer (`decisionSvc.RecordSignoff`). Its replay-before-`RecordSignoff` ordering is already correct.

Note: the existing test `TestSignoffByDocumentHandler_ReplayRequiresCurrentEligibility` (`signoff_handler_test.go:392-425`) ENCODES the bug as desired behavior — it asserts an eligibility failure must override a cached outcome. That assertion is the symptom we are removing; the test must be inverted to assert `was_replay: true` under the same setup.

Payload hash inputs are unchanged (still `docID + inst.ID + activeStage.ID + decision + reason + content_hash`), preserving cross-tenant/cross-document key isolation; computing it requires the load to stay where it is.

## Tradeoffs (Verbatim)

- **PRO:** Restores RFC-style idempotency semantics — once a `(key, payloadHash)` tuple has been recorded, any retry deterministically returns the recorded outcome regardless of intervening state changes. Eliminates the replay-after-terminal bug class without touching the two-layer state-validation model, the `parseIfMatch` contract, or signoff content-pin sources.
- **CON:** A cached failure outcome (Fail path) will be replayed for a key whose underlying instance has since become eligible — but this is the correct idempotency contract, not a regression; clients must use a fresh key for a fresh attempt.
- **CON:** The existing eligibility-overrides-replay assertion is reversed, which is a deliberate behavior change scoped to the bug.
- **Rejected alternative:** also adding state validation BEFORE `BeginDocumentReplay` (belt-and-suspenders) — reintroduces the exact bug, because a transient `ErrInstanceCompleted` on the first call would still return an error without consuming a slot, and a later retry under a now-eligible state would execute a second signoff under the same key.

## Implementation Steps (Verbatim Core, with Gates)

1. Edit `internal/modules/documents/approval/http/doc_approval_handler.go` lines 114-150. Keep `loadActiveInstanceByDocumentForMutation` (114-122) unchanged — the load is required because `payloadHash` mixes `inst.ID` and `activeStage.ID`, and `activeStage` must be resolved before hashing.
2. Move the state-validation block (current lines 124-132: `Active() == nil` + `CheckEligibility`) to sit AFTER the `BeginDocumentReplay` block.
3. Move the `payloadHash` computation (current line 134) and the replay block (current 135-150) so they run immediately after the load and BEFORE state validation.
4. New ordering inside `SignoffByDocumentHandler`:
   - (1) parse `tenantID/actorID/docID/idempKey` [73-86, unchanged]
   - (2) `parseIfMatch` [87-91, STABLE — untouched per constraint]
   - (3) decode + validate body [97-112, unchanged]
   - (4) load active instance [114-122, unchanged]
   - (5) resolve `activeStage := inst.Active()` but do NOT return on `nil` yet — derive `stageIDForHash`: if `activeStage != nil` use `activeStage.ID`, else use empty string (matches stage-scoped handler convention; keeps the key stable across terminal transitions for same `docID + instID`)
   - (6) compute `payloadHash := signoffPayloadHash(docID, inst.ID, stageIDForHash, decision, reason, contentHash)`
   - (7) `BeginDocumentReplay`; if replay != nil return cached 200 with `was_replay: true` immediately
   - (8) NOW perform state validation: if `activeStage == nil` return terminal-state error; else `CheckEligibility` and surface its result
5. **Invert the existing test:** `TestSignoffByDocumentHandler_ReplayRequiresCurrentEligibility` at `signoff_handler_test.go:392-425` must be rewritten to assert `was_replay: true` under the same setup (cached outcome replayed despite intervening terminal/eligibility change). Rename the test to reflect the corrected semantics, e.g. `TestSignoffByDocumentHandler_ReplayWinsOverPostLoadEligibilityChange`.
6. Add a new test that confirms a FIRST call to a terminal-state instance still produces the terminal-state error (no slot consumed) — guards against the rejected belt-and-suspenders alternative.
7. Add a test confirming `stageIDForHash` fallback to empty string when `activeStage == nil` produces a stable replay key across the terminal transition.
8. Confirm the stage-scoped handler is unchanged and its existing tests still pass.

## Risk Register

| ID | Severity | Risk | Mitigation |
|---|---|---|---|
| R1 | Medium | Inverting `TestSignoffByDocumentHandler_ReplayRequiresCurrentEligibility` is an explicit behavior contract change | Call it out in PR description; document the corrected RFC-style idempotency contract in `wiki/modules/approval/*` |
| R2 | Medium | Cached Fail outcomes replay under newly-eligible state | Correct per idempotency contract; document clearly in approval module wiki; clients must rotate `Idempotency-Key` for fresh attempts |
| R3 | Low | `stageIDForHash` empty-string fallback collides with a real stage ID | Impossible — stage IDs are non-empty UUIDs |
| R4 | Low | Stage-scoped handler accidentally edited | Explicit constraint; review-time check |

## Regression Test Plan

- **Unit:** inverted test (Step 5) asserts `was_replay: true` for the second call after intervening terminal/eligibility change.
- **Unit:** first-call terminal-state test (Step 6) asserts terminal error and no slot consumed.
- **Unit:** stage-fallback-empty-string test (Step 7) asserts stable replay key across terminal transitions.
- **Regression:** full `internal/modules/documents/approval/...` test suite must pass.
- **Regression:** stage-scoped handler tests must pass unchanged.
- **Manual integration:** start API via `.\scripts\start-api.ps1 -Build`; submit a doc-scoped signoff, transition the instance to terminal state out-of-band, retry the same signoff with same `Idempotency-Key` — expect 200 with `was_replay: true`. Capture curl evidence.

## Rollback Plan

- Single-file source diff plus test edits. `git revert <sha>` reverses the change.
- No schema change, no migration, no client contract change at the wire level (the corrected behavior matches the documented `Idempotency-Key` contract; rollback restores the buggy behavior).
- If rollback is needed, re-open F-002 with the failure mode notes.

## Success Criteria

- New handler ordering matches Step 4 exactly.
- Inverted test passes; new tests (Steps 6, 7) pass.
- Full approval module test suite passes.
- Stage-scoped handler and its tests untouched.
- Manual integration evidence (curl with `was_replay: true`) captured in PR.
- Approval module wiki updated with the corrected idempotency contract note.

## Files Touched

- `internal/modules/documents/approval/http/doc_approval_handler.go` (handler reordering)
- `internal/modules/documents/approval/http/signoff_handler_test.go` (invert existing test, add new tests)
- `wiki/modules/approval/*` (document the corrected idempotency contract, bump `Last verified`)

## Wiki Updates Needed (per CLAUDE.md drift policy)

- Approval module wiki page covering signoff endpoints — document that doc-scoped signoff now replays cached outcomes even after terminal transition or stage rotation; clients must rotate `Idempotency-Key` for fresh attempts.
- Bump `Last verified` stamp.

## Agent Dispatch

- **`tdd-guide`** — write the inverted test and new tests first; confirm they fail against the current handler.
- **`code-reviewer`** — after the reorder.
- **`security-reviewer`** — replay/idempotency boundary check.
- **`wiki-curator`** — after merge, refresh approval module wiki anchors and `Last verified`.

## Spawn Task Prompt (self-contained)

```
Fix F-002 in MetalDocs: SignoffByDocumentHandler in internal/modules/documents/approval/http/doc_approval_handler.go performs state/eligibility validation BEFORE the idempotency replay check, so a second call sharing (tenant, actor, route, Idempotency-Key, payloadHash) returns an error after the instance has moved to a terminal state or the active stage rotated, instead of replaying the cached outcome. The stage-scoped handler is correct and is OUT OF SCOPE — do not modify it.

Required behavior change: reorder the document-scoped handler so the BeginDocumentReplay check sits BETWEEN the instance load and the state/eligibility validation. New ordering inside SignoffByDocumentHandler:
  1. Parse tenantID/actorID/docID/idempKey (lines 73-86, unchanged).
  2. parseIfMatch (lines 87-91, STABLE — do not modify).
  3. Decode + validate body (lines 97-112, unchanged).
  4. Load active instance via loadActiveInstanceByDocumentForMutation (lines 114-122, unchanged) — the load is required because payloadHash mixes inst.ID and activeStage.ID.
  5. Resolve activeStage := inst.Active() but DO NOT return on nil yet. Derive stageIDForHash: if activeStage != nil use activeStage.ID, else use empty string (matches stage-scoped handler convention; keeps the replay key stable across terminal transitions for the same docID+instID).
  6. Compute payloadHash := signoffPayloadHash(docID, inst.ID, stageIDForHash, decision, reason, contentHash).
  7. BeginDocumentReplay; if replay != nil return cached 200 with was_replay: true immediately.
  8. NOW perform state validation: if activeStage == nil return terminal-state error; else CheckEligibility and surface its result.

Do NOT add state validation before BeginDocumentReplay (belt-and-suspenders) — that reintroduces the bug because a transient ErrInstanceCompleted on the first call would return without consuming a slot, allowing a later retry under newly-eligible state to execute a second signoff under the same key.

Tests:
  - Invert the existing TestSignoffByDocumentHandler_ReplayRequiresCurrentEligibility at signoff_handler_test.go:392-425 — it currently asserts that an eligibility failure overrides a cached outcome (this is the bug). Rewrite it to assert was_replay: true under the same setup, and rename it to TestSignoffByDocumentHandler_ReplayWinsOverPostLoadEligibilityChange (or equivalent).
  - Add a test confirming a FIRST call to a terminal-state instance still returns the terminal-state error and consumes no replay slot.
  - Add a test confirming the stageIDForHash empty-string fallback produces a stable replay key across the terminal transition.
  - Verify the stage-scoped handler and its tests are unchanged.

Run: go test ./internal/modules/documents/approval/... — all must pass. Then start API via .\scripts\start-api.ps1 -Build, manually verify by submitting a doc-scoped signoff, transitioning the instance to terminal state out-of-band, retrying the same signoff with the same Idempotency-Key, and confirming 200 with was_replay: true. Capture curl evidence in the PR description.

Wiki: update the approval module wiki page covering signoff endpoints to document the corrected idempotency contract (cached outcomes replay even after terminal transition or stage rotation; clients must rotate Idempotency-Key for fresh attempts). Bump Last verified. After merge, dispatch the wiki-curator agent.

Honor CLAUDE.md gates: surgical changes (§5.3), bias to caution (§5), record verification evidence before claiming closure (§4 Evidence rule). Use tdd-guide to write tests first, then code-reviewer and security-reviewer before opening the PR. Do not skip hooks. Do not amend prior commits.

Success criteria: handler ordering matches the spec above, inverted test plus two new tests pass, full approval module test suite passes, stage-scoped handler untouched, manual curl evidence of was_replay: true after terminal transition captured in PR, approval module wiki updated.
```