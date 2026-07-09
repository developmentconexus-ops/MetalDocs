# Feature F2d.2 — Verdict History + Immutable Actor-Name Snapshot

> **Milestone:** 2d — Workflow Coherence FE + Viewer Contract  ·  **Folder:** `f2-verdicts-display-names`
> **Status:** Planning

Execution plan (the "how") for the contract in [`spec.md`](spec.md). Scope is **Option A** (global
maximum) per the milestone HS-2 clearance (milestone.md Amendment 2026-07-09) and
[ADR 0079](../../../../../wiki/decisions/0079-verdict-history-contract.md).

## Source

- Milestone spec row (F2d.2): implement — server projects the by-instance verdict history onto the
  instance-by-document DTO with immutable actor display names; validate — real-DB tests prove
  chronological ordering across stages, ready/request_changes rows, empty→`[]`, and the snapshot
  invariant (empty snapshot rejected at insert; no read fallback).
- HS-2 clearance: D1 (migration + unconditional write bind) and D2 (read-fallback removal) are in
  scope, overriding the milestone's "F2d.2 needs NO migration" line for this feature only.
- Governing spec: `2026-07-08-approval-workflow-coherence-design.md` §2 (D2), §8.1.

## Plan

TDD order: failing real-DB tests first, then D1 (storage invariant + write), then D2 (read-fallback
removal) + the new by-instance read, then the app/handler/OpenAPI projection. Build stays green after
each step; the snapshot-invariant test flips red→green at D1.

### Step 0 — Failing tests first (RED)

Application-layer, `-tags integration`, testdb factory, isolated DB per test:

- `TestInstanceVerdicts_Ready` — one review stage, one `ready` verdict → `verdicts[]` len 1, correct
  `actor_display_name` (snapshot), `verdict=ready`, `reason` = comment, `verdict_at` UTC.
- `TestInstanceVerdicts_RequestChanges` — one `request_changes` verdict → row present, enum value
  correct, reason carried.
- `TestInstanceVerdicts_MultiStageOrdered` — verdicts on two different stages of the same instance →
  returned chronological asc by `verdict_at`, spanning both stages (not stage-scoped).
- `TestInstanceVerdicts_None` — instance with no verdicts → `verdicts == []` (non-nil empty).
- `TestInsertVerdict_EmptySnapshot_Rejected` — inserting a verdict with an empty
  `actor_display_name_snapshot` → DB rejects (CHECK/NOT NULL violation surfaced as an error). This is
  the RED test that D1 turns GREEN by making the write bind unconditional + the column constrained.

Signoff parity (same package or infrastructure):
- `TestInsertSignoff_EmptySnapshot_Rejected` — mirror of the above on `approval_signoffs`.

### Step 1 — D1 storage invariant (migration `0294`)

`db/migrations/0294_actor_display_name_snapshot_not_null.sql` (idempotent, forward-only, BEGIN/COMMIT):

1. Backfill both tables where snapshot is NULL or `''`, from `iam_users.display_name`:
   ```sql
   UPDATE public.approval_review_verdicts v
      SET actor_display_name_snapshot = u.display_name
     FROM public.iam_users u
    WHERE u.user_id = v.actor_user_id
      AND u.tenant_id = v.actor_tenant_id
      AND (v.actor_display_name_snapshot IS NULL OR v.actor_display_name_snapshot = '');
   ```
   (mirror for `approval_signoffs`). `iam_users.display_name` is NOT NULL, so every matched row gets a
   non-empty value. Join keys per `iam_users` PK `(user_id text)` + tenant scope.
2. `ALTER TABLE … ALTER COLUMN actor_display_name_snapshot SET NOT NULL;` on both tables.
3. `ADD CONSTRAINT …_snapshot_nonempty CHECK (actor_display_name_snapshot <> '')` on both (DROP IF
   EXISTS first, 0288 idiom).
4. `schema_migrations` ledger insert (`ON CONFLICT DO NOTHING`).

Baseline (`db/baseline/0001_current_schema.sql`) stays frozen — migration is the change of record
(ADR 0032 baseline-frozen precedent). Note the drift in the migration comment.

Insert-binding change (unconditional): in
`infrastructure/postgres_approval_repository.go`
- `InsertVerdict` (line ~1225): delete the `if name := v.ActorDisplayNameSnapshot(); name != ""` guard;
  bind the value directly (still via `sql.NullString{…, Valid:true}` or a plain string — the column is
  now NOT NULL, so pass the raw string; an empty string now trips the CHECK and returns an error,
  which is the intended fail-closed behavior).
- `InsertSignoff` (line ~162): identical change.

> Root-cause note: the value is always non-empty in practice (`review_verdict_service.go:81` reads it
> from `iam_users.display_name` NOT NULL, off-tx, H-PRE-1). The unconditional bind + CHECK make the
> invariant explicit instead of silently NULL-tolerant.

### Step 2 — D2 read-fallback removal + by-instance read

`infrastructure/postgres_approval_repository.go`:
- **New** `LoadInstanceVerdicts(ctx, tx, tenantID, instanceID) ([]domain.ReviewVerdict, error)` —
  mirror `LoadStageVerdicts` but `WHERE v.approval_instance_id = $1` (spans all stages), same tenant
  guard, `ORDER BY v.verdict_at ASC`. Select `v.actor_display_name_snapshot` **directly** (no
  `coalesce`). Keep `coalesce(v.comment,'')` + `coalesce(v.on_behalf_of_user_id,'')` — those columns
  stay legitimately nullable.
- `LoadStageVerdicts` (line ~1330): drop the `coalesce(v.actor_display_name_snapshot,'')` → direct
  select (now dead-safe post-D1; keeps the no-fallback invariant uniform).
- Signoff load path: drop the snapshot `coalesce` on whichever signoff loader feeds the mapper.

Handler mapper (`http/get_instance_handler.go`, ~line 147-171): collapse the signoff
snapshot-else-live branch — `if displayName := sig.ActorDisplayNameSnapshot(); displayName != "" { … }
else { off-tx DisplayNames }` becomes **snapshot-only**: use `sig.ActorDisplayNameSnapshot()`
unconditionally. Delete the else-branch live lookup for signoff names. (The delegator-name off-tx
lookup for `via_delegation_from`, added in F2d.1, is unrelated and stays.)

Add the repo method to the `ApprovalRepository` port interface it belongs to; regenerate/adjust any
mock if one is hand-maintained (prefer the real repo in integration tests).

### Step 3 — App read extension

`application/read_service.go`: extend `LoadInstanceByDocumentForViewWithViewer` (the shared
`loadInstanceByDocumentForViewTx`) to also call `s.repo.LoadInstanceVerdicts(...)` in the **same view
tx**, returning the verdicts alongside the instance + viewer facts. Plain SELECT, H-PRE-1 safe (no
lock held on this path). Widen the return struct/tuple the handler consumes.

### Step 4 — Handler projection + contracts

- `http/contracts/instance_read.go`: add `VerdictRecord{ID, StageInstanceID, ActorUserID,
  ActorDisplayName, Verdict, Reason *string, VerdictAt}` and `Verdicts []VerdictRecord` on the
  by-document response struct.
- `http/get_instance_handler.go` `mapInstanceResponse`: map the loaded `[]domain.ReviewVerdict` →
  `[]VerdictRecord`, ordered as loaded, empty slice → `[]` (not nil). `reason` = comment, omitted/null
  when empty. `verdict_at` formatted RFC3339 UTC.

### Step 5 — OpenAPI + regen

- `api/openapi/v1/openapi.yaml`: add required `verdicts` array to `ApprovalInstanceByDocumentResponse`
  with the item schema from spec.md §Consumer-contract (ADR 0079 §Decision-1). `verdict` enum
  `[ready, request_changes]`.
- Run `oapi-codegen` regen (Go + TS). No hand-written `body.data.verdicts` reader (ADR 0035).
  Regenerate `frontend/apps/web/src/lib/api-types/index.d.ts`.

### Step 6 — Green + regression

- `go build ./...`, `go vet ./internal/modules/documents/approval/...` = 0.
- Real-DB: the 6 Step-0 tests GREEN, plus re-run the F2d.1 F8 visibility regression
  (`TestLoadInstance_*` / `TestLoadActiveInstanceByDocument_*`) and the verdict/signoff write suites to
  prove D1 didn't break existing inserts (existing tests seed real names, so they satisfy the new
  CHECK). Split runs with `-timeout 45m` (Windows teardown).

### Files touched

| File | Change |
|------|--------|
| `db/migrations/0294_actor_display_name_snapshot_not_null.sql` | **new** — backfill + NOT NULL + CHECK (both tables) |
| `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go` | unconditional insert bind (×2); `LoadInstanceVerdicts` (new); drop snapshot coalesce (LoadStageVerdicts + signoff load) |
| `internal/modules/documents/approval/application/read_service.go` | load verdicts in the view tx; widen return |
| `internal/modules/documents/approval/http/contracts/instance_read.go` | `VerdictRecord` + `Verdicts[]` |
| `internal/modules/documents/approval/http/get_instance_handler.go` | project verdicts; signoff name → snapshot-only |
| `api/openapi/v1/openapi.yaml` | `verdicts` array on ByDocument response |
| generated Go `api.gen.go` + TS `index.d.ts` | regen |
| approval application integration tests | 6 new real-DB tests |
| ports interface (ApprovalRepository) | add `LoadInstanceVerdicts` |

### Test strategy

Real Postgres only for the behavior claims (testdb factory, isolated DB). The snapshot-invariant tests
assert DB-level rejection (proving the constraint, not just app code). Ordering/spanning proven with
multi-stage seed. No fixture-only proof for any acceptance row.

## Execution notes

Executed in TDD order per the plan. Highlights and deviations:

- **Step 0 (tests-first):** 6 real-DB tests written before impl in
  `read_service_instance_verdicts_integration_test.go` — `TestInstanceVerdicts_{Ready,RequestChanges,MultiStageOrdered,None}`
  + `TestInsert{Verdict,Signoff}_EmptySnapshot_Rejected`. Reused the F2d.1 `viewerFixture` harness.
- **Step 1 (migration 0294 + unconditional bind):** as planned; both tables hardened
  (backfill → SET NOT NULL → CHECK <> '') in one idempotent forward-only migration; ledger `'0294'`.
- **Step 2 (D2 fallback removal):** removed snapshot `coalesce` in `LoadStageVerdicts`, both signoff
  loaders, and `loadVerdictByStageActor`; refactored the verdict row-scan into shared `scanVerdicts`.
  Signoff-name else-live branch collapsed to snapshot-only in `mapInstanceResponse` + `buildStageActors`.
- **Step 3–4 (app read + handler):** `LoadInstanceByDocumentForViewWithViewer` → 4-value return; verdict
  projection gated on `viewer != nil`; `Verdicts *[]VerdictRecord` pointer chosen so by-id omits and
  by-document emits present-possibly-empty `[]` (scan returns nil for zero rows, so gating on `viewer`,
  not on `verdicts != nil`, is what distinguishes the two paths).
- **Step 5 (OpenAPI + regen):** `ApprovalVerdictRecordResponse` added; Go + TS regenerated, verified present.
- **Deviation / added work (not in original plan):** D1's `NOT NULL` broke one pre-existing real-DB seed
  — `seedSignoff` in the F2d.1 viewer integration test omitted the snapshot. Fixed the seed
  (`ActorDisplayNameSnapshot: "Signed Actor"`) and re-ran `TestViewerBlock_AlreadySigned` (green, 7.4s).
  This is a required harness consequence of the invariant. All other seed sites are fakes or non-DB unit
  tests (verified by grep), unaffected.
- **Review:** independent `caveman:cavecrew-reviewer` (sonnet) over 7 files → **0 findings**.
- Full test evidence in `evidence.md`.
