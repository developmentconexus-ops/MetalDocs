# Feature F5.6 — Evidence — fanout queue-wiring fix (+ two connection-correctness follow-ups)

> **Milestone:** 5 · **Feature:** `f5.6-fanout-queue-wiring-fix` · **Closed:** 2026-07-04
> **Origin:** HS-6 scope-surface — the defect was found by M5's own required live QA drive
> (`milestone.md` validation-definition item 6), operator-directed fix-in-place (AskUserQuestion
> 2026-07-04: "Open bounded fix feature F5.6 now"). Two further connection-correctness defects were
> then surfaced by re-driving the fix live, each handled under the same HS-6 discipline (stop,
> surface via AskUserQuestion, fix) rather than silent-patched.
> **Contract:** `spec.md`. Validation Gate proved below.
> Engine: `superpowers:subagent-driven-development` — fresh subagent per fix task; main session
> reviewed diffs, committed, and performed the live re-drives itself (they need the running system).

## What was implemented

Three commits, each closing one defect the previous one's live re-drive exposed. The root cause was
never the fanout worker's logic (F5.5 already proved that correct) — it was purely the enqueue-side
queue routing plus two `database/sql` connection-usage bugs inside the reader-fanout path.

- **T1 `4095c8c4` — queue routing (the specced fix).**
  `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go` —
  `EnqueueLifecycleEventTx` passed `nil` `InsertOpts`, so River defaulted the job to queue
  `"default"`, which `metaldocs-jobs` never subscribes (it subscribes only `"temporal"` and
  `"maintenance"`). Jobs sat in `river_job` forever, never dequeued → fanout silently never ran for
  any lifecycle event. Fixed to `&river.InsertOpts{Queue: "temporal"}`, matching the M5 convention
  (`scheduled_publish_job.go:88`, `dispatchjobs/enqueuer.go:96`). `Client` field narrowed to an
  unexported `riverInserter` interface for unit-testability (concrete `*river.Client[*sql.Tx]`
  still satisfies it structurally — production wiring at `metaldocs-api/main.go:525` unchanged).
  Unit test asserts `InsertOpts.Queue == "temporal"` against a recording fake.

- **T2 `5433c004` — open-cursor exec (found by the first live re-drive).** With the queue fixed, the
  job dequeued but failed 5× with `driver: bad connection`. Root cause: `fanoutToReaders` iterated an
  open `tx.QueryContext` result set while calling `tx.ExecContext` (the per-recipient INSERT) inside
  the `rows.Next()` loop — a `*sql.Tx` has one underlying connection, and interleaving an Exec with a
  still-open Query result set on it surfaces as `driver: bad connection`. First fix: buffer all reader
  ids into a `[]string`, drain/close `rows`, then loop the inserts. Regression unit test
  (`go-sqlmock`, already a dep) proved 3 readers each inserted.

- **T3 `47f47b05` — global-maximum rewrite (operator-directed).** Buffer-then-insert *avoids* the bug
  but keeps an N+1 round-trip loop. Per the Global-Maximum rule (judge the foundation, don't lock in a
  local max), `fanoutToReaders` was rewritten as a single set-based `INSERT INTO ... SELECT ... FROM
  metaldocs.v_cd_obligated_readers` — the fan-out is computed in Postgres, one statement, no Go-side
  cursor and no per-recipient round-trips. Structurally *eliminates* the open-cursor bug class (only
  one statement exists) rather than working around it. Same idempotency
  (`ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING`),
  same tenant-seeded tx. `fanoutToAuthor`/`insertRow` (single-recipient path, no cursor) untouched.
  Unit test updated to expect the single INSERT...SELECT exec.

## Verification

| Check | Command / action | Result |
|-------|------------------|--------|
| Build (all 4 binaries) | `go build ./...` | clean, after each of T1/T2/T3 |
| Vet | `go vet ./...` | clean |
| Unit — queue routing | `go test ./internal/modules/documents/approval/jobs/...` | PASS (`InsertOpts.Queue=="temporal"` vs fake) |
| Unit — fanout worker | `go test ./internal/modules/notifications/infrastructure/...` | PASS (4 tests incl. multi-reader regression, adapted to set-based exec) |
| Runnable | `.\scripts\start-api.ps1 -Build` → all 3 host binaries rebuilt + relaunched, API `listening :8081`, `migrations done` | PASS |
| **Live drive (item 6)** | black-box HTTP: login → create company-scope CD → autosave commit → submit → signoff approve → **direct publish** → poll `GET /api/v1/notifications` as admin | **PASS** — see below |
| **River job ground truth** | `docker exec metaldocs-postgres psql ... 'SELECT id,queue,state,attempt,finalized_at FROM public.river_job WHERE kind=''notification_fanout'''` | job **23** (post-fix drive): `queue=temporal, state=completed, attempt=1` | 

### Live drive proof (real system, real Postgres, HTTP-only)

Published document `bfbed0eb-d0f9-44b8-b461-98550b4d9239` (controlled doc `ec200048-…`, code
`PO-RH-004`). Publish → `200 {"new_status":"published"}`. Admin (company-scope obligated reader)
notification materialized on the first poll (~14s after publish):

```json
{"items":[{"event_type":"document.published","recipient_user_id":"admin",
  "resource_type":"document","resource_id":"bfbed0eb-d0f9-44b8-b461-98550b4d9239",
  "title":"Novo documento controlado publicado para leitura",
  "message":"Um documento controlado que você deve ler foi publicado.",
  "status":"PENDING","read_at":null,"id":"26b43afe-86ea-430f-be0a-f89fd51a28a5",
  "created_at":"2026-07-04T19:10:44.000902-03:00"}],
 "page":{"has_more":false,"next_cursor":null}}
```

River job `id=23` (this drive's fanout job): `queue=temporal`, `state=completed`, `attempt=1` — proves
the fix path dequeues on the subscribed queue and the set-based insert commits on the first attempt
(no `driver: bad connection`). Capabilities honored (not roles): create/submit/publish ran as `admin`
(has `controlled_documents.create` + `document.publish`), signoff as `approver` (has
`document.signoff`).

## Acceptance vs spec Validation Gate

| Criterion (spec.md) | Met? | Evidence |
|----------------------|------|----------|
| 1. Enqueuer passes `Queue:"temporal"`, unit-asserted | **yes** | `4095c8c4` diff + `InsertOpts.Queue` fake assertion |
| 2. `go build ./...` green, all 4 binaries | **yes** | clean after every commit; `start-api.ps1 -Build` rebuilt all |
| 3. Live re-drive shows a materialized notification row for a driven publish within a bounded poll | **yes** | admin row above, ~14s; job 23 `completed attempt 1` |
| 4. No regression to F5.2–F5.5 / existing `"temporal"` behavior | **yes** | build+vet clean; scheduled-publish (Step A) unaffected — same queue, one more producer; F5.5 race test untouched and still valid against set-based insert |
| 5. Section-by-section match; no HS-7 | **yes** | first documented contract for this defect; no prior contract to diverge from |

## Review disposition

- Every fix diff read in full by the main session before commit (`git show` on `4095c8c4`,
  `5433c004`, `47f47b05`). T3 set-based rewrite reviewed for arg/placeholder correctness ($1 reused
  in projection + WHERE, $2 WHERE-only, $3–$8 positional) and idempotency preservation.
- **Two HS-6 stop-and-surface events** handled per discipline, not silent-patched: (a) the original
  queue-mismatch (→ F5.6 opened via AskUserQuestion); (b) the `driver: bad connection` defect (→
  surfaced via AskUserQuestion, operator chose "wait for retry then diagnose"; retry confirmed
  non-transient, root-caused to open-cursor exec). The global-max rewrite (T3) was likewise
  operator-directed after surfacing that buffer-then-insert was a local maximum.
- No subagent-dispatch anomaly across the three fix dispatches + one live-drive dispatch.

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| Pre-fix orphan River job `id=6` (`queue=default`, from the first pre-F5.6 drive) lingers in `river_job` | Local dev-tenant scratch row on an unsubscribed queue (inert). Does not affect correctness of shipped code. (Job `id=12`, stuck 6× on `driver: bad connection`, **self-healed on its next retry** under the new binary — `state=completed, attempt=7, finalized 22:12:39` — extra proof the T3 fix is general, not drive-specific.) | Purely cosmetic dev-DB cleanup; no production or code impact. Owner: none required. |
| `-race` execution of F5.5's race test | Env-capped (`CGO_ENABLED=0`, no cgo on this box) — unchanged from F5.5 | Re-run with `-race` if a cgo-capable CI runs the integration suite; not blocking. |

## Autosave-commit 500 (surfaced in passing by the live drive — NOT part of F5.6)

The live drive hit a `500` on `POST /documents/{id}/autosave/commit` when `form_data_snapshot` was
omitted (spec marks it optional, but a NOT-NULL on `documents.form_data_json` fires `SQLSTATE 23502`,
surfaced as a bare `INTERNAL_ERROR`). Worked around by sending `form_data_snapshot: {}`. This is a
pre-existing documents-module contract/impl mismatch, **out of F5.6's boundary** (notifications
async wiring) — recorded here for traceability, to be triaged separately (candidate its own fix
feature); it did not block the fanout proof.
