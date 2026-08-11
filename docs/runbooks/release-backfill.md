# Runbook — ADR 0085 Stage C release-backfill (in-flight disposition repair)

**Script:** `go run ./scripts/release-backfill` (`main.go`, `backfill/backfill.go`, `backfill/wire.go`)
**Owner:** ops (operator-run, one-shot per invocation; no schedule, no automation)

**Provenance of this runbook:** derived entirely from the tool's source (`main.go`
doc comment, `backfill/backfill.go`, `backfill/wire.go`) and `wiki/modules/approval.md`
§8.9. Nobody has run this tool while writing it — this is a reconstruction, not a
procedure someone performed and recorded. The tool has existed since commit
`1f2c6376` (ADR 0085 Stage C) with no prior runbook; this one was written on
2026-08-11 after an unrelated change (A5.2's `TxRunner.DoReadOnly` deletion)
mechanically touched one line inside `lookupTenant` and surfaced the gap. The
tool's own behavior did not change as part of that.

## Context

Legacy documents that reached `approved` before the ADR 0085 release-coordinator
pipeline existed are missing the release prerequisites the coordinator needs to
ever publish them (an approval fact it can trust, a materialization dispatch, a
release evaluation). This tool synthesizes those prerequisites for an explicit,
operator-supplied list of document ids — never "all approved" — without
re-approving the document and without trusting any legacy artifact pointer
(`final_docx`/`final_pdf` column presence is never treated as readiness, per
ADR 0085's own framing).

It ships as a `go run` tool, not a fifth binary: no Docker image, no HTTP route,
no long-lived state (same precedent as `scripts/api-lint`, `scripts/req-trace`).

## Two modes — do not conflate them

### Mode 1 — default backfill (no `-repair-only`)

Per document, in one transaction:
1. `RecordApprovalFactTx` — synthesizes the missing approval fact from the
   document's approved `approval_instance`.
2. `FreezeService.RepairMaterialization` — re-pins `frozen_revision_id` from the
   current revision, re-enqueues materialization.
3. `EnqueueReleaseEvaluationTx` — arms the release coordinator.

Preflight (fails closed; a failure aborts only THIS document, the batch
continues) requires: `status = approved`; `controlled_document_id` set;
`current_revision_id` set; BOTH `values_hash` and `values_frozen_at` set
(pinned — a half-pinned document is refused, not repaired); exactly one
`approval_instances` row with `status = approved` carrying a
`frozen_content_hash` (two or more is refused outright, never guessed); that
instance has at least one `approve` signoff, whose actor becomes the recorded
final approver.

### Mode 2 — `-repair-only` (integrity restoration, never history mutation)

For documents whose stored artifacts are invalid (e.g. a blank `frozen.docx` /
`final.pdf` produced before a frozen-body bug fix). Runs ONLY step 2 above; no
approval fact is written and no release evaluation is enqueued — approval
history is untouched. It additionally purges TERMINAL (`dispatched`/`failed`)
staging dispatch rows and TERMINAL (`published`/dead-lettered) delivery
`outbox_events` rows on the document's own dedupe keys, on both the
materialize and PDF legs — this purge is required, not optional, or the
re-dispatch silently swallows (see **Verification** below). It refuses the
WHOLE repair, before writing anything, if any row on either layer is still
in-flight (`pending`/`processing` staging, or an unpublished/not-dead-lettered
delivery event) — it never partially purges.

Preflight requires: `status` is `approved` OR `published` (a closed set — every
other status, including `scheduled`, `superseded`, `obsolete`, is refused);
`current_revision_id` set; BOTH `values_hash` and `values_frozen_at` set; at
least one `approval_instances` row with `status = approved` exists (existence
only — uniqueness is not required here, unlike mode 1).

## When to run

**Not established as a defined trigger or cadence.** Neither the code nor its
doc comments state a schedule — both modes are one-shot operator interventions,
run against document ids the operator has already identified out-of-band.
Which documents qualify (which legacy `approved` documents are missing release
prerequisites; which documents carry invalid frozen artifacts) is not
something this tool or this runbook determines — `-docs` has no "all approved"
mode by design, specifically so that judgement stays with the operator. That
identification step is out of scope for this document.

## Preconditions

- Direct DSN access to the target Postgres database: `-dsn` flag, or the
  `DATABASE_URL` environment variable if `-dsn` is empty (checked in that
  order; the tool exits fatally if both are empty). The tool never reads
  `.env` and never prints the DSN — every error message is passed through a
  `redact()` step first.
- `METALDOCS_JOBS_RIVER_SCHEMA`, if the target environment's River schema is
  not the client's default — passed straight through to
  `riverjobs.NewClientBundle`. Not established what happens if it's needed but
  left unset for a given environment; confirm against
  `internal/platform/jobs/river` if unsure.
- The explicit list of document UUIDs to operate on, already decided (see
  **When to run**).

## Procedure

Dry-run is the default and always the first step. It is guaranteed to write
nothing by construction, not by discipline — every dry-run code path ends in
a deliberate transaction rollback (`errDryRunRollback`), not a conditional
skip of the write calls.

```
go run ./scripts/release-backfill -docs <uuid>[,<uuid>...]
```

Read the printed plan line per document (`outcome=planned ...`) before
applying anything. Then apply:

```
go run ./scripts/release-backfill -docs <uuid>[,<uuid>...] -dry-run=false
```

For the repair-only pass (invalid frozen artifacts only — never touches
approval history):

```
go run ./scripts/release-backfill -docs <uuid> -repair-only -dry-run=false
```

Each document runs in its own transaction. A failure on one document rolls
back only that document and does not abort the rest of the batch. The process
exit code is `1` if ANY document in the batch failed, `0` otherwise — in a
multi-document batch, read the per-document `outcome=` lines, not just the
exit code.

## Verification

The tool prints one report line per document: `doc=<id> outcome=<outcome>
generation=<id> <detail>`. Outcomes: `backfilled`, `already-backfilled`,
`repaired`, `planned` (dry-run), `failed`.

**Why this section exists.** A live run of this tool once printed `repaired`,
its River job completed, and it rendered nothing. The cause was a three-layer
dedupe "swallow stack": the staging outbox tables and the delivery
`outbox_events` table both use `ON CONFLICT DO NOTHING`, so a leftover row on
the same dedupe key makes a re-dispatch report success while doing no work —
repeated on both the materialize and PDF legs. **A process exit code of 0, or
an `outcome=repaired`/`outcome=backfilled` line, is not by itself proof that
materialization happened.**

What each mode actually proves, by construction, differs:

- **`-repair-only`** calls `requireMaterializeEnqueued` before committing: it
  re-reads the materialize staging row inside the SAME transaction, and the
  entire repair rolls back if that row isn't present with `status=pending`.
  So `outcome=repaired` proves the materialize leg of the staging layer really
  got a fresh dispatch row. It does **not** prove the PDF leg or the delivery
  layer (`outbox_events`) — those are written later, by the worker, in a
  separate transaction after this one commits. The code's own comment states
  this limit plainly: the PDF leg is "unverifiable from this tool in
  principle." What repair-only does guarantee is that it purged the stale
  delivery keys before returning `repaired` — not that a downstream publish
  has already happened.
- **Default backfill (no `-repair-only`)** has no equivalent self-check.
  `requireMaterializeEnqueued` is called only from the repair-only path; the
  default path's write sequence ends at `EnqueueReleaseEvaluationTx` with no
  read-back. So `outcome=backfilled` is weaker evidence than `outcome=repaired`
  — it means preflight passed and every write in the transaction committed,
  but it does not itself rule out the same class of dedupe swallow that
  repair-only was specifically hardened against.

Given that asymmetry, treat exit code + outcome line as necessary, not
sufficient, and confirm externally once the worker has had time to process
the pipeline. **Not established:** the exact verification SQL an operator
should run — I have not executed any query against a real database for this
runbook. The tables the code actually touches, to build such a check from,
are: `release_generations` (generation identity, default-backfill only),
`metaldocs.materialize_dispatch_outbox` and `metaldocs.pdf_dispatch_outbox`
(dispatch rows should reach `status='dispatched'`, not stay `pending`), and
`documents` (the frozen/current-revision columns this tool reads and writes).
Where the worker records the resulting artifact facts is outside what this
tool's source shows — see the release coordinator's own docs
(`wiki/modules/approval.md` §8.9 and neighboring sections) before writing that
query.

## Rollback

**Not established.** The tool has no inverse operation, and none is described
in its doc comments or in `wiki/modules/approval.md` §8.9. Do not improvise
one under pressure — escalate instead.

What the transactional design does give you:

- **Before commit:** every failure path (a rejected precondition, a write
  error, or dry-run) rolls back the WHOLE per-document transaction. A
  document that fails or is only planned leaves nothing partial behind.
- **After commit, default backfill mode:** there is no delete/undo path in
  this codebase for a `RecordApprovalFactTx` written in error. Treat a
  committed default-backfill write as a fact that happened, the same way a
  real approval decision would be — it is not something this tool can
  hand-unwind.
- **After commit, repair-only mode:** the purge of terminal staging/delivery
  rows is a `DELETE` with no pre-image recorded anywhere except the ids/keys
  printed on the outcome line — there is no code path to restore a purged
  row. The `values_hash`/`values_frozen_at` `UPDATE` re-assigns
  byte-identical values read back from the same row earlier in the same
  transaction (per the package doc's stated write set), so that part is not
  history-destructive by design — but `frozen_revision_id` DOES change, and
  no prior value is recorded to revert to.
- **Replay safety (not a rollback, but relevant, and NOT the same for both
  modes — corrected 2026-08-11 after review against the code; the original
  text here overstated repair-only's behavior):**
  - **Default backfill** re-running with the same `-docs` list IS a no-op:
    preflight checks for an existing `release_generations` row
    (`pre.existingGenerationID`) and, if found, reports
    `outcome=already-backfilled` without writing anything else.
  - **`-repair-only` re-running is NOT a no-op and does NOT report
    `already-backfilled`.** There is no equivalent short-circuit in
    `runRepairOnly`/`repairPreflight`. Once the prior dispatch/delivery rows
    reach a TERMINAL status, `repairPreflight` classifies them as stale,
    `runRepairOnly` purges them, and `RepairMaterialization` is called again
    — producing another `outcome=repaired` and another re-dispatch, every
    time it's re-run against a document in that state. The dedupe keys are
    not a guard against this: purging the rows that hold those keys is
    exactly what makes the next dispatch possible, not what prevents it.
    The only thing `repairPreflight` refuses outright is re-running while a
    prior dispatch/delivery row is still IN-FLIGHT (non-terminal) — that is
    a safety check against concurrent overlap, not against repeated,
    sequential re-application. An operator who re-runs `-repair-only`
    against the same document more than once should expect to delete and
    re-render delivery artifacts each time, not expect the second run to
    detect "already repaired" and skip.

If a committed write needs to be undone, that requires a manually written,
reviewed remediation against the specific rows this tool wrote — identifiable
from the printed outcome line (generation id; purged dispatch row ids; purged
event ids) — not a procedure this runbook can responsibly hand an operator in
advance.
