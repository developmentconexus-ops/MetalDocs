# Feature F6.4 — Surfacer contract-conformance + marker consumer

> **Milestone:** 6 · **Feature:** `f6.4-surfacer-contract-and-consumer` · **Kind:** HS-4 fix-feature (validator FAIL, `qa/milestone-qa.md` 2026-07-05)
> **Contract:** binding `../validation-contract.md` §4 (surfacer) + §6(d) (review-due filter). This feature makes the shipped code **conform** to §4 — no contract re-open, no HS-7 erratum.
> **Approved:** 2026-07-05 / Leandro — design locked via operator "prove the global maximum" interview (two decisions below), answered in-session; operator directed conform + run-integration-myself.

## Why this feature exists

The M6 milestone-validator returned **FAIL** with two blocking findings on the F6.2 River surfacer
(the headline eQMS deliverable) + one non-blocking. This fix-feature closes all three and executes
the authored-not-executed integration suite. Full evidence trail in the validator verdict
`../qa/milestone-qa.md` (C3 findings 1–3, C7).

## The two locked design decisions (global-maximum, evidence-backed)

### D1 — Surfacer conforms to §4.2/§4.3 (per-tenant seed), NOT an all-tenant sweep

**Consumer contract (the invariant is the consumer):** every tenant-scoped async write in the
jobs/worker binaries seeds `metaldocs.tenant_id` (via `authz.SeedTxTenant`/`SeedTxIdentity`) so
FORCE-RLS backstops the SQL predicate — M3 invariant (ADR 0027, `validation-contract.md` §2.3),
machine-checked by the `ASYNC-TENANT-SEED` lint over the 33 FORCE-RLS tables (`documents` included).

**Shipped defect:** `document_review_surfacer/job.go` runs GUC-unset under `BypassSystem`;
`ReviewSurfaceWriterPG.MarkSurfaced` is a single all-tenant `UPDATE` on the NULL-GUC "all tenants"
RLS branch. It falsely cites `stuck_instance_watchdog` as precedent — the watchdog does the opposite
(`SeedTxTenant(inst.TenantID)` at `job.go:174`). It escaped the lint via a **port-indirection blind
spot** (the UPDATE lives in a documents-module port, outside the lint's `asyncHandlerRoots` scan). §4.2
was the human backstop for that blind spot; it was silently ignored and the isolation test
(`job_integration_test.go:130`) asserts the **inverse** of §4.3 ("sweep must cover both tenants").

**Why conform (not ratify the sweep):** the codebase does sanction some GUC-unset system sweeps
(idempotency-janitor TTL DELETE, audit-integrity scan) — so the sweep is not *inherently* unsafe. But
conform is the **global maximum** for an eQMS/Part-11 system: (1) it **minimizes the unseeded-sweep
exception surface** — adding `documents` (controlled-records crown jewel) to that set enlarges the
blast radius of any future `WHERE`-clause bug and forces every future author into the subtle
"is-my-write-attributed?" judgment; (2) **defense-in-depth** — per-tenant seed means a buggy predicate
can only touch the seeded tenant; (3) **zero governance debt** — no operator erratum, no allowlist
entry, no exception a future auditor must re-litigate; (4) it makes the §4.3 isolation test **writable**
(seed A → assert B untouched); (5) cost is negligible (yearly cadence). Uniform with the attributed-write
precedent (watchdog) the code already claims to follow.

**Target shape:** the surfacer tick does a cross-tenant **read** (bypass, like
`watchdog.listStuckInstances`) to enumerate the tenants that have due docs, then **per tenant**:
`SeedTxTenant(tenantID)` in a tx → tenant-scoped `MarkSurfaced` → commit. No unseeded write survives.
Read-port `ListDueForReview` stays tenant-scoped per §4.1 (jobs never issues raw `documents` SQL).

### D2 — The review-due filter READS the marker (worklist), making surfacing non-inert

**Consumer contract:** `GET /documents?review_due=true` returns the documents the surfacer has
**flagged for the current review cycle** (`review_surfaced_at IS NOT NULL AND review_surfaced_at >=
review_due_at`, still published + effective) — an audit-anchored eQMS worklist. `mark-reviewed`
advancing `review_due_at` auto-expels the doc (`review_surfaced_at < review_due_at`).

**Shipped defect:** `review_surfaced_at` (migration 0276) is **write-only** — the filter
(`repository.go:485`) recomputes `review_due_at <= now()` and never reads it; no DTO/FE reads it;
`mark-reviewed` doesn't touch it. The hourly job + column are inert. Spec interview #7 promised the
filter reads surfaced docs; the code diverged.

**Why worklist (not DTO-badge, not remove):** (1) **Part-11 value** — the marker is an attributable,
timestamped record that the system flagged the doc (§11.10); pure recompute can't prove *when* a doc
entered the review queue; (2) **one truth** — a recompute filter + a surfaced marker are two
definitions of "due" that diverge at the ≤1h margin (a split-brain, finding-3 class); worklist
collapses them to one predicate; (3) **minimal delta** — activates already-built idempotency
semantics + the auto-expel behavior; (4) honors the spec without re-opening it. Remove was rejected:
it contradicts `milestone.md` F6.2 (requires scheduled surfacing + integration proof) → would re-open
the milestone, and discards the audit anchor.

**Also:** expose `review_surfaced_at` on the document response DTO (read-side already surfaces the four
review fields from T6 — consistency), so the FE can show "flagged for review on <date>".

### D3 (finding 3) — one source of truth for the "due core" predicate

Extract the shared **due-core** (`status='published' AND review_due_at IS NOT NULL AND review_due_at
<= $now AND effective_from <= $now AND (effective_to IS NULL OR effective_to > $now)`) to a single
SQL fragment/const consumed by `MarkSurfaced` eligibility + `ListDueForReview`. The filter's
surfaced-state predicate (`review_surfaced_at >= review_due_at`) derives from the same core.

## Non-goals (mandatory)

- **No new capability, no new route, no new migration family.** `review_surfaced_at`/0276 stays (D2
  gives it a consumer). No schema change beyond possibly a supporting index for the surfaced filter.
- **No notification/escalation** on overdue review — still bounded-deferred (contract §8 / M8).
- **No change to F6.3** (reason-for-change) — it passed validation.
- **No re-open of contract §4** — this feature conforms *to* it. (If conform proved infeasible we'd
  stop at HS-7; it is feasible.)
- **No change to the mark-reviewed authz/OCC path** beyond the marker interplay needed for D2.
- **No lint-scope expansion** (closing the ASYNC-TENANT-SEED port blind spot is a real follow-up but
  out of this fix's boundary — recorded as a bounded defer with a trigger).

## Validation Gate (acceptance + named tests + proof commands)

| # | Acceptance | Proof |
|---|-----------|-------|
| 1 | Surfacer seeds per-tenant; **no unseeded tenant-scoped write** remains | code review + `SeedTxTenant` present before every `MarkSurfaced`; census: surfacer tick has no GUC-unset write path |
| 2 | Cross-tenant **isolation** proof (matches §4.3) | rewritten `TestIntegration_Surfacer_CrossTenant_*`: seed A + B due docs, run **under tenant-A seed only** → A surfaced, **B untouched**; asserts the contract, not its inverse |
| 3 | Idempotent (twice → once) still holds under per-tenant model | `TestIntegration_Surfacer_Idempotent_SecondRunNoOp` green on DB |
| 4 | `review_due=true` returns the **surfaced** set; `mark-reviewed` expels a doc | integration: surface a doc → filter includes it → mark-reviewed → filter excludes it |
| 5 | `review_surfaced_at` exposed on document DTO; pin test | `Test…ReviewFieldsWireContract` extended; oasdiff/api-lint green |
| 6 | Due-core predicate single-sourced | one const/fragment; grep shows the three sites reference it |
| 7 | **Integration suite executed on real Postgres** (surfacer isolation/idempotency, F6.2 CHECK/OCC/tripwire-negative, F6.3 reason-persist/reason-on-audit) | `go test -tags integration -run <targeted>` real output attached to evidence (loaded `.env` via PowerShell, secret never printed) |
| 8 | Build/vet/registry/tripwire/pins still green | `go build ./...`, `go vet ./...`, deterministic gates |

## Interview record (operator, 2026-07-05)

| Q | Operator answer | Resolution |
|---|-----------------|------------|
| Finding 1: conform vs ratify the sweep | "Explain why global maximum — proofs/facts/evidence/references" | Delivered the evidence proof (M3 invariant + lint blind-spot + watchdog precedent + exception-surface argument) → **conform** (D1) |
| Finding 2: marker consumer | "Same — prove global maximum" | Delivered proof (Part-11 audit anchor + one-truth + minimal-delta) → **filter reads marker / worklist** (D2) |
| Q3: run integration suite (no DB creds) | "You run it yourself — you have all the .env, read the wiki about it" | Load `.env` → `$env:METALDOCS_DATABASE_URL` via PowerShell (sanctioned, never printed); run targeted M6 integration tests (gate #7) |
