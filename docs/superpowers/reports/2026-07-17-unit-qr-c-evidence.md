# Unit QR-C — PDF dispatch event-contract fix (evidence)

Branch: `unit/qr-c-pdf-event-contract` (base `0480c2f4`)
Chip: QR-C (HARNESS-CORE + docs/HARNESS-PROFILE.md + REVIEW-STANDARD)
Status at write: implementation + scoped ladder GREEN. Hub adjudications received
(HUB-MARKER-d9099bff): migration `0309` **CONFIRMED final**; producer-(b)
disposition **OPTION B — EXTERMINATE, ratified**; `main.go` delete-only scope-lock
extension granted. Dual gate pending final SHA.

---

## Defects (from QA-2 re-drive, commit `0480c2f4`)

### F-QA2-2 (primary) — PDF events never carried `final_docx_s3_key`
`docgen_v2_pdf` events shipped with an empty `FinalDocxS3Key`, so every PDF event
dead-lettered at the consumer with `pdf job runner: missing final_docx_s3_key`
(`internal/platform/worker/pdf_job_runner.go:113-118`, fail-closed check). The
producer path had **no column** to carry the renderer-produced key from the
enqueuing tx onto the dispatched event — the field was structurally always `""`.

### F-QA2-1 (secondary) — nil PDF runner silently marked events published
`internal/platform/worker/service.go` handled `EventTypePDFConvert` with a nil
`pdfRunner` by leaving `handleErr == nil`, so the outbox row was marked
**published** and the PDF was silently never generated. The bootstrap comment at
`worker.go` falsely claimed "the outbox retries later" — it did not.

---

## Fix shape

### F-QA2-1 — fail loud on unconfigured runner
- `service.go`: added `errPDFRunnerNotConfigured` / `errMaterializeRunnerNotConfigured`
  sentinels; both `EventTypePDFConvert` and `EventTypeMaterializeFanout` now set
  `handleErr` to the sentinel when their runner is nil, routing the event to the
  retry / dead-letter path instead of a false "published".
- `bootstrap/worker.go`: rewrote the false comment to state FAIL LOUD (F-QA2-1).
- Regression: `service_test.go` `TestWorkerService_NilPDFRunner_FailsLoud` +
  `TestWorkerService_NilMaterializeRunner_FailsLoud` (assert published==0,
  failures==1, LastError names the unconfigured runner). Confirmed RED first
  (published==1 pre-fix), GREEN post-fix.

### F-QA2-2 — thread `final_docx_s3_key` end-to-end (new event-contract snapshot)
The staging outbox row is the event-contract snapshot for its dispatch (same
reason `content_hash bytea NOT NULL` already lives there, baseline `0001:1363`).
Threading, source → wire:

1. **Migration `0309`** (expand-only): `ALTER TABLE metaldocs.pdf_dispatch_outbox
   ADD COLUMN final_docx_s3_key text;` — nullable, no backfill. Only the pdf
   table (materialize has no docx key at its enqueue time). NOT NULL tightening
   is a deliberate follow-up once the two already-dead-lettered rows are purged
   by retention.
2. **`StagingOutboxRepository.EnqueuePDF`** (new, pdf-only): INSERT persists the
   key as the 4th column; **fails closed** on empty key (`final_docx_s3_key must
   not be empty`); panics if invoked on a non-pdf table. Generic `Enqueue`
   (materialize) is untouched.
3. **`dispatchjobs.Enqueuer.EnqueuePDFTx`** gains the key param, threads it into
   `dispatchFields.FinalDocxS3Key`, and calls `EnqueuePDF`. Extracted shared
   `insertRiverJob` so the materialize path (`enqueueTx`) and pdf path share the
   `*sql.Tx` assertion + River insert.
4. **`dispatchFields` / `buildPDFEvent`** (`args.go`, `workers.go`): field added;
   `PDFConvertPayload.FinalDocxS3Key` is populated from it.
5. **Producer (a) — `MaterializeJobRunner`** (`materialize_job_runner.go`, the
   **sole** reachable/honest pdf producer): threads `result.FinalDocxS3Key`
   (held in the same tx as `WriteFinalDocxInTx`) into `EnqueuePDFTx`.

New invariant recorded: **a PDF dispatch of a never-materialized docx cannot be
requested** — enforced at the write boundary (`EnqueuePDF` rejects an empty key,
fail-closed). The consumer's own fail-closed check
(`pdf_job_runner.go:113-118`) was **never weakened** and the worker still never
guesses the key — this fix supplies the key from the producer, it does not relax
the guard.

---

## Stop-on-contradiction finding → OPTION B (exterminate), hub-ratified

Producer (b) — the old synchronous in-tx pdf-dispatch block on `DecisionService`
— was **structurally dead code**. Proof against base `0480c2f4`'s
`decision_service.go` (hub independently re-verified; hub line-anchors in
parens):

- The document-approve branch hard-requires the async-freeze seam:
  `if s.pinInvoker == nil { return ...err("pinInvoker not configured") }`
  (~:616 / hub ~:594) **precedes** `shouldDispatchPDF = true` (:651 / hub :629),
  in the same branch.
- The producer-(b) dispatch guard required the negation:
  `if shouldDispatchPDF && s.pdfDispatch != nil && s.pinInvoker == nil` (:787 /
  hub :765). `shouldDispatchPDF == true` implies `pinInvoker != nil`, so the
  `pinInvoker == nil` conjunct is a **contradiction** => the block (:766) is
  unreachable.

Lineage: dead since **ADR 0015 async-freeze** made `MaterializeJobRunner` the
post-Pin pdf producer; the synchronous block + `pdfDispatch` seam are a
pre-0015 remnant that was never removed (it even sat in the `Ready()`
required-ports assertion yet was never invoked on any reachable path).

Ruling grounds: **stop-on-contradiction + legacy-fallback-extermination**
(hard break, no defensive threading of a dead seam) + **global-max**
(`MaterializeJobRunner` becomes the sole honest producer). Hub ratified OPTION B.

**Deleted** (no dead seam left behind):
- `decision_service.go`: the dead dispatch block, `pdfDispatch` field,
  `WithPDFOutbox`, `pdfDispatchEnqueuer` interface, the `Ready()` required-ports
  entry, and the orphaned locals `shouldDispatchPDF` / `pdfTenantID` /
  `pdfRevisionID`. The interim `finalDocxReader` port + `WithFinalDocxReader`
  (option-A scaffolding) were also dropped.
- `documents/infrastructure/snapshot_repository.go`: fully reverted — the
  `ApprovalFinalDocxReader` adapter + `ReadFinalDocxS3Key`'s `q ...DBTX` variadic
  + the `platform/db` import (all option-A scaffolding) are gone.
- `apps/api/cmd/metaldocs-api/main.go` (delete-only per scope lock): the Decision
  `.WithPDFOutbox(...)` and `.WithFinalDocxReader(...)` calls. `pdfDispatchEnqueuer`
  the local var stays — still fed to `freezeService.WithMaterializeOutbox(...)`.
- Tests: `TestRecordSignoff_PinInvoker_PDFOutboxNotEnqueued` + the
  `fakePDFOutboxEnqueuer` fake (guarded the deleted block); the `WithPDFOutbox`
  wiring-test call sites and the `pdfDispatch` entry in the `Ready()` port list.

**Kept:** the shared `*dispatchjobs.Enqueuer` (live via the freeze materialize
path) and `EnqueuePDFTx` itself (live producer (a) calls it).

Out-of-boundary (hub registered as ROADMAP defer, NOT touched): the pre-existing
raw `UPDATE documents` at `decision_service.go:609/:685` (legacy approval→documents
raw-SQL, re-home candidate).

---

## Test-discipline note

Two coverage tiers, both real:

- **Repo-layer contract (sqlmock, sibling pattern)** —
  `pdf_outbox_repository_test.go`: `EnqueuePDF_PersistsFinalDocxKey` asserts the
  INSERT names `final_docx_s3_key` and binds the key as `$4`;
  `EnqueuePDF_EmptyKeyFailsClosed` asserts the guard rejects before any INSERT.
  Follows the established sibling pattern (every existing
  `pdf_outbox_repository_test.go` case is sqlmock).
- **Real-DB round-trip (testdb factory, `//go:build integration`)** —
  `dispatchjobs/dispatch_integration_test.go` threads a non-empty renderer key
  through `EnqueuePDFTx` and asserts it (a) persists on
  `pdf_dispatch_outbox.final_docx_s3_key` (the migration-0309 column) and (b)
  round-trips onto the published `docgen_v2_pdf` payload end-to-end. This is the
  real-DB proof of the exact producer→wire path QR-C fixes.

Signature-change caller sweep: `EnqueuePDFTx` gaining the required
`finalDocxS3Key` param orphaned four callers in
`dispatch_integration_test.go` and the `noopPDFEnqueuer` fake in
`materialize_job_runner_integration_test.go`. Both are `//go:build integration`,
so untagged `go test` never compiled them and the first ladder masked the break;
both dual-gate arms caught it under `-tags integration`. Repaired in commit
`503ad2a1`; `go vet -tags integration` on both packages is clean. **Lesson: a
scoped `go test` without `-tags integration` does not compile integration files —
always vet the integration lane after any signature change on a producer seam.**

---

## Ladder (run from `C:/Users/leandro.theodoro/Documents/MetalDocs`)

- `go build ./...` — clean.
- `go vet ./internal/modules/render/fanout/... ./internal/platform/worker/...
  ./internal/modules/approval/application/... ./internal/modules/documents/infrastructure/...`
  — clean.
- `go vet -tags integration ./internal/modules/render/fanout/dispatchjobs/...
  ./internal/platform/worker/...` — clean (integration lane compiles; DB-run
  deferred to a live drive with DATABASE_URL, per the integration-file header).
- `go test`:
  - `internal/platform/worker` — ok (incl. both nil-runner fail-loud + producer-a key threading)
  - `internal/modules/render/fanout/...` (+ dispatchjobs, retention) — ok (incl. EnqueuePDF contract + empty-key fail-closed + buildPDFEvent payload threading)
  - `internal/modules/approval/...` (application/domain/http/infrastructure/jobs) — ok
  - `internal/modules/documents/infrastructure` — ok
  - `internal/platform/bootstrap` — ok
  - post-extermination re-run: full approval/* + worker + fanout/* + documents-infra + bootstrap all ok.

Full `./...` not run (touched set is scoped; DB/platform touch limited to the
expand-only migration + one new pdf-only repo method — per selective-ladder policy).

---

## Diff surface (post-extermination)

Producers/threading (F-QA2-2): `staging_outbox.go` (new pdf-only `EnqueuePDF`),
`dispatchjobs/{enqueuer,args,workers}.go`, `materialize_job_runner.go` (producer a).
Fail-loud (F-QA2-1): `service.go`, `bootstrap/worker.go`. Extermination (Option B):
`decision_service.go` (dead block + seam removed), `main.go` (delete-only),
`snapshot_repository.go` (fully reverted). Migration:
`db/migrations/0309_pdf_dispatch_outbox_final_docx_key.sql`. Tests: worker
`service_test.go` + `materialize_job_runner_test.go`, dispatchjobs
`{enqueuer,workers}_test.go`, fanout `pdf_outbox_repository_test.go`, approval
`decision_service_{freeze,wiring}_test.go` (obsolete-guard deletions).

---

## Gates status

1. **Migration `0309`** — hub CONFIRMED final (next-free; highest committed `0308`).
   Lock scope = number 0309 only, expand-only; releases at CLOSED.
2. **Producer-(b) disposition** — hub ratified **OPTION B (exterminate)**. Done.
3. **Working-context confirm** — chip operates directly in the main repo dir
   (`C:\Users\leandro.theodoro\Documents\MetalDocs`) on branch
   `unit/qr-c-pdf-event-contract`; hub holds main-branch commits until merge.
4. Commit on branch (NEVER push) — done. Round-1 gate arms (both REJECT) caught
   the integration-lane compile break at `661bec56`; repaired at `503ad2a1`.
   Round-2: cold Opus **ACCEPT** on `8a3ba0ae`; codex **REJECT** on two grounds —
   (a) `main.go` was not strictly delete-only (a 5-line explanatory comment was
   added beside the `.WithPDFOutbox` removal) and (b) its `-tags integration` vet
   "gate" exited 1. (a) is repaired at `f4e76878` (comment stripped; the sole
   base-diff is the `.WithPDFOutbox(pdfDispatchEnqueuer)` deletion). (b) was a
   **codex-sandbox artifact**, not a code defect: the exact failure was
   `go: creating work dir: mkdir …\Temp\go-build…: Access is denied` — the
   read-only sandbox could not create Go's temp build dir, so compilation never
   started. In the real environment `go vet -tags integration
   ./internal/modules/render/fanout/dispatchjobs/... ./internal/platform/worker/...`
   exits 0 (recorded in the ladder above; the cold Opus arm independently
   confirmed both integration files carry the 6-param signature). Round-3 re-runs
   both arms on `f4e76878`, with the codex arm given a writable `GOTMPDIR`/`GOCACHE`
   so its vet gate can actually compile.
5. **Dual gate** on the final fixed SHA `60ec85f0` (round 3): cold Opus **ACCEPT**
   + GPT-5.6 Sol (via codex) **ACCEPT** → **P6-DUAL-GATE: AGREEMENT**. Codex's
   `-tags integration` vet again hit its sandbox temp-dir denial (env limitation,
   not a code defect; not treated as REJECT by that arm); authoritative real-env
   vet EXIT=0, and the cold arm confirmed both integration files carry the 6-param
   signature. Evidence pack: `.mnfs/unit-qr-c/_chip-qr-c/EVIDENCE.md`.
