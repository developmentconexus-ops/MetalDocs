# M4 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M4 (versioning kernel correctness)
> **Authored:** 2026-07-04, **before any implementation** (mission D4). Committed before the first code change.
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7). The load-bearing clauses are the **§1 full transition table**, the
> **§1.4 app↔DB-trigger parity requirement**, the **§1.5 `rejected` full-removal scope**, the **§2
> publish-race single-winner + terminal-state table**, and the **§3 idiom decision**.
>
> **Operator decisions locked (2026-07-04, interview record in each feature `spec.md`):**
> 1. **`rejected`** — remove it entirely, inside M4 (full-stack: enum + CHECK + DB trigger arcs + OpenAPI
>    enum + FE + reader + tests). It is a dead status: no app service ever enters it (reject writes `draft`).
> 2. **`scheduled`** — keep it, one-way (`scheduled → published` is its cutover exit; `scheduled → draft`
>    stays a DB-legal arc for parity but no app path exercises it — see §1.3). No feature rework.
> 3. **`concurrency idiom`** — operator delegated the choice to a full engineering analysis (§3). Original
>    decision was "unify on `If-Match`, migrate templates." **SUPERSEDED 2026-07-04 (HS-7 erratum §3.7):**
>    verification proved the "templates is the minority straggler" premise false — templates has zero
>    If-Match and its body idiom is self-consistent. Corrected binding decision: **ADR-record the split as
>    intentional (ADR 0066)**, name `If-Match` as the target, defer full unification to its own change (not
>    M4). Re-open headlined for ratification at the HS-1 gate.

---

## 0. Runtime-truth basis (the facts this contract is built on)

All claims traced to source at authoring time (2026-07-04). Runtime truth beats docs (CLAUDE.md).

### 0.1 The nine status **values** vs the transition **graph**

- `internal/modules/documents/domain/model.go:8-20` declares **9** `DocumentStatus` values: `draft`,
  `under_review`, `approved`, `rejected`, `scheduled`, `published`, `superseded`, `obsolete`, `archived`.
- **Two of the nine are NOT lifecycle-transition nodes:**
  - **`rejected`** — DB-tolerated (the trigger below has `under_review→rejected` and `rejected→draft`
    arcs) but **the application never enters it**: the reject path (`approval/application/decision_service.go:461`)
    and the cancel path (`approval/application/cancel_service.go:151`) both write `status='draft'`, not
    `'rejected'`. `wiki/bugs/audit-2026-05-03.md` A6 records this as a deliberate bypass (fix `2977ef96`).
    **Operator decision (1): remove it** (§1.5).
  - **`archived`** — **not a status transition at all**. `MarkArchived`/`MarkArchivedTx`
    (`repository/repository.go:1616,1841`) set `archived_at = now()` **without changing status**
    (comment `repository.go:1614`: "sets archived_at … without changing its status"; pinned by
    `TestMarkArchived_StampsTimestampWithoutStatusChange`; an unarchive path sets `archived_at = NULL`,
    `repository.go:1675`). No code writes `status='archived'` (grep census = 0 non-test writes). ADR 0010
    (`wiki/decisions/0010-soft-archive-via-timestamp.md`) governs: the DB trigger defines **no** `→archived`
    arc; archive is a soft timestamp flag orthogonal to the lifecycle. **`archived` is retained** (ADR 0010,
    not in scope for removal) but is **outside the state machine** — every `→archived` / `archived→` edge is
    **illegal** in the transition function, matching the DB trigger.

### 0.2 The authoritative graph is the DB trigger — the current app "FSM" is dead

- **`enforce_document_transition()`** (`db/baseline/0001_current_schema.sql:549-585`), fired by
  `trg_documents_legal_transition BEFORE UPDATE ON documents` (baseline:3858), is the **authoritative,
  last-line** transition graph. Its arcs (verbatim):
  ```
  under_review → draft        [special case: requires GUC metaldocs.cancel_in_progress set]
  draft        → under_review
  under_review → approved | rejected
  rejected     → draft
  approved     → published | scheduled | draft
  scheduled    → published | draft
  published    → superseded | obsolete
  superseded   → obsolete
  ```
  A `status` UPDATE where `OLD.status = NEW.status` is a no-op (the trigger's
  `IF OLD.status IS DISTINCT FROM NEW.status` short-circuits — self-edges are neither transitions nor
  errors). Any other `OLD→NEW` raises `illegal status transition` (SQLSTATE `check_violation`, `23514`).
- **The app-domain "state machine" is dead code.** `documents/domain/state.go` `CanTransitionDocument`
  covers only `draft→{under_review,archived}` and `under_review→archived` and is **called 0×** in
  production (only `model_test.go`; `docs/superpowers/specs/2026-06-26-verified-storage-kernel-design.md:329`
  flags it "Dead-FSM cleanup … called 0×"). It is also **wrong** vs the DB trigger (it lists
  `→archived` arcs the trigger forbids and omits every post-approval arc). It is not a friendly first
  line — it is a decorative liability.
- **Real app-side enforcement today** is the per-service OCC `UPDATE … WHERE status = '<cur>' AND
  revision_version = $N` in each approval service — the scattered guards the review named. Each is a
  partial, hand-maintained echo of the DB trigger. **This is the split-brain M4 closes.**

### 0.3 The scattered lifecycle guards (F4.1 collapses these; census target after = 0)

| Transition | Site (file:line) | Current guard |
|---|---|---|
| draft → under_review | `approval/application/submit_service.go:186` | `WHERE … status='draft'` |
| under_review → approved | `approval/application/decision_service.go:405` | `WHERE … status='under_review'` |
| under_review → draft (reject) | `approval/application/decision_service.go:461` | `WHERE … status='under_review'` |
| under_review → draft (cancel) | `approval/application/cancel_service.go:151,155` | `WHERE … status='under_review'` + `metaldocs.cancel_in_progress` GUC |
| approved → published (manual) | `approval/application/publish_service.go:76,102` | `if instance.Status != InstanceApproved`; `WHERE … status='approved'` |
| approved → scheduled | `approval/application/publish_service.go:235,279` | `if instance.Status != InstanceApproved`; `WHERE … status='approved'` |
| approved → published (supersede-new) | `approval/application/supersede_service.go:90` | `WHERE … status='approved'` |
| scheduled → published (cutover) | `approval/application/scheduler_service.go:32,206` | `WHERE … status='scheduled'`; `if state.Status != DocStatusScheduled` |
| published → superseded (supersede-prior) | `approval/application/supersede_service.go:107` | `WHERE … status='published'` |
| published/superseded → obsolete | `approval/application/obsolete_service.go:81,96` | `if priorStatus not in (published,superseded)`; `WHERE … status=$priorStatus` |

Note `UpdateDocumentStatus` (`repository/repository.go:613`) is a **generic** cur→next status writer
used by legacy/edge paths; its `stampTime→archived` branch (`repository.go:616`) is a vestige of the
pre-ADR-0010 hard-archive and writes no reachable transition (the trigger rejects `→archived`).

### 0.4 The two publish paths (F4.2 target) — independent, both OCC on `revision_version`

- **Manual publish** (`approval/application/publish_service.go:96-115`, api binary, user-driven):
  `UPDATE documents SET status='published', revision_version=revision_version+1 WHERE id=$1 AND
  tenant_id=$2 AND status='approved' AND revision_version=$3`. Zero rows → `ErrStaleRevision`.
- **Scheduled cutover** (`approval/application/scheduler_service.go:25-33,119-150`, jobs binary, River
  `ScheduledPublishWorker`): loads the row `FOR UPDATE` (scheduler_service.go:184-191), validates
  `status='scheduled'` + `schedule_generation` + `effective_from` (scheduledJobMatchesState:205-219),
  then `UPDATE … SET status='published', revision_version=revision_version+1 WHERE id=$1 AND tenant_id=$2
  AND status='scheduled' AND revision_version=$3`. Zero rows → `errScheduledPublishNoOp` (mapped to a
  successful no-op; River retry-safe).
- **No shared choke point.** Each path owns its UPDATE. **Runtime looks safe by construction** — the two
  UPDATEs match on **mutually exclusive** `status` predicates (`approved` vs `scheduled`), so a given row
  can satisfy at most one; the OCC `revision_version` predicate closes the same-status double-fire. F4.2
  **proves** this rather than asserting it, and installs a single choke point only if the proof fails.

### 0.5 The two concurrency idioms (F4.3 target)

- **Documents:** `If-Match: "vN"` request header → parsed by `parseIfMatch`
  (`approval/http/handler.go:145-164`) to an int, threaded to the OCC `WHERE revision_version=$N`.
  Applied on publish/signoff/obsolete/cancel handlers. Column: `documents.revision_version`
  (int, monotonic — trigger `trg_documents_revision_version_monotonic`, baseline:3865).
- **Templates:** body `expected_lock_version` (JSON, required) → `routes_schema.go:32-57` → OCC
  `WHERE lock_version=$N`. Column: `templates_template_version.lock_version` (int).
- Both are compare-and-swap optimistic concurrency over one integer version column. Only the **transport**
  differs (HTTP header vs request body). Analysis + decision in §3.

### 0.6 DB invariants that STAY the last line (do not move to the app)

- `enforce_document_transition` (transition legality) — **stays**; F4.1 mirrors it, does not replace it.
- `enforce_revision_version_monotonic` (baseline:617, trigger :3865) — revision_version cannot decrease.
- `documents_status_check` (migration `0265`) — the status value-set CHECK.
- The `metaldocs.cancel_in_progress` GUC gate on `under_review→draft` — a DB-level authorization of the
  cancel/reject rollback; **stays**.
- Write-tripwire `enforce_capability_asserted` (M2) and FORCE-RLS tenant policies (M3) — **stay**; no
  approval-service refactor may regress them.

---

## 1. F4.1 — unified exhaustive state machine + `rejected` removal

### 1.1 The single transition function (binding shape)

A **single** exhaustive transition function in `documents/domain` (replacing the dead
`CanTransitionDocument`), mirroring the templates pattern (`templates/domain/version.go`
`TemplateVersion.CanTransition`). Binding properties:

1. **Total over the status set** — it returns a defined legal/illegal answer for **every** ordered pair
   `(cur, next)` of `DocumentStatus` values (post-removal set: 8 values incl. `archived`). It never
   panics, never has an implicit "default true".
2. **Legal iff the pair is in the §1.2 table** — the legal set is **exactly** the DB trigger's arc set
   (post-`rejected`-removal). No arc the trigger lacks; no arc the trigger has is omitted (§1.4 parity).
3. **Illegal-returns-are-typed** — an illegal transition yields `ErrInvalidStateTransition` (mirror
   templates), not a bare bool where the caller must guess.
4. **Self-edges** (`cur == next`) are **not transitions** — the function reports them as not-a-transition
   (illegal for the purpose of "may I move"), consistent with the trigger treating `OLD=NEW` as a no-op.
   (Callers never route a no-op status write through the transition check.)

### 1.2 ★ The full transition table (authoritative — the compared-against contract)

Rows = current status; columns = next status. `L` = legal transition, `·` = illegal. Post-`rejected`-removal
(`rejected` row/column deleted with the status). `archived` is present as a value but is **outside** the
graph — every edge into/out of it is illegal (soft-archive is not a status transition, §0.1). This table
**equals** the DB trigger arc set (§1.4).

| from ↓ \ to → | draft | under_review | approved | scheduled | published | superseded | obsolete | archived |
|---|---|---|---|---|---|---|---|---|
| **draft**        | ·  | **L** | ·     | ·     | ·     | ·     | ·     | · |
| **under_review** | **L**¹ | · | **L** | ·     | ·     | ·     | ·     | · |
| **approved**     | **L**² | · | ·     | **L** | **L** | ·     | ·     | · |
| **scheduled**    | **L**² | · | ·     | ·     | **L** | ·     | ·     | · |
| **published**    | ·  | ·     | ·     | ·     | ·     | **L** | **L** | · |
| **superseded**   | ·  | ·     | ·     | ·     | ·     | ·     | **L** | · |
| **obsolete**     | ·  | ·     | ·     | ·     | ·     | ·     | ·     | · |  (terminal) |
| **archived**     | ·  | ·     | ·     | ·     | ·     | ·     | ·     | · |  (not a node) |

Footnotes (DB-trigger conditions the function documents; the DB remains the enforcer):
- ¹ `under_review → draft` is **GUC-gated at the DB** (`metaldocs.cancel_in_progress`). It is the shared
  arc for **reject** (`decision_service`) and **cancel** (`cancel_service`). The app function reports it
  legal; the DB trigger enforces the GUC as the authorization. **F4.1 implementation MUST verify** (not
  assume) that the reject path sets that GUC exactly as cancel does — if reject reaches the trigger
  without the GUC, that is a pre-existing latent bug to surface (HS-2 boundary call), not something F4.1
  invents around.
- ² `approved → draft` and `scheduled → draft` are **DB-legal arcs with no current app caller** (§0.2).
  They are included as **legal** so the app function is a faithful mirror of the authoritative trigger
  (§1.4). They are not new features; no route is added to exercise them in M4.

**Terminal states:** `obsolete` (no outgoing), `archived` (not a node). `published`/`superseded` are
**not** terminal (they flow to `superseded`/`obsolete`).

### 1.3 Routing the approval services through the function (the wiring — not just authoring)

The review's defect is that the current FSM is **dead** (called 0×). Re-authoring an unused exhaustive
function would be a **symptom-patch (C6 FAIL)**. Binding requirement:

- **Every approval service in §0.3 calls the unified function as its friendly first line** before its OCC
  UPDATE — the function is the single app-side authority for "is this move legal", replacing each site's
  ad-hoc `if status != X` reasoning. The OCC `WHERE status='<cur>' AND revision_version=$N` **stays** (it
  is the atomic CAS that also carries optimistic concurrency); what changes is that the **legality
  decision** is centralized, not re-encoded per site.
- **After F4.1, a grep for lifecycle status-equality guards (`if …Status != …`, `status != '<lifecycle>'`)
  in the approval services returns 0** outside the unified function (or each residual is explicitly
  allowlisted with a reason — e.g. an instance-status guard `instance.Status != InstanceApproved`, which
  is the **approval-instance** state, a *different* state machine from the **document** status and out of
  F4.1 scope; such entries are named, not silently kept).
- The dead `CanTransitionDocument` + its `model_test.go` are **deleted or replaced** by the unified
  function + its coverage test (no two document-status FSMs remain — that would be a fresh split-brain).

### 1.4 ★ App↔DB-trigger parity (the anti-split-brain clause — binding)

The unified function's legal set MUST equal `enforce_document_transition`'s arc set, and this MUST be
**pinned by a test**, so the friendly-first-line and the last-line can never silently drift:

- A test enumerates the function's legal pairs and asserts they equal the DB trigger's arcs (the §1.2
  table). Preferred: an **integration** test that, for each legal pair, drives the real `UPDATE` and
  asserts the trigger permits it, and for a representative illegal pair asserts the trigger raises
  `check_violation` — proving the two agree on real Postgres. A **unit** table-equality test (function
  legal set == the §1.2 constant) is the minimum; the integration cross-check is the target (targeted
  `-run`, testdb factory).
- When `rejected` is removed (§1.5), **both** the function and the DB trigger lose the `under_review→rejected`
  and `rejected→draft` arcs **in the same change**, keeping parity. The parity test is the gate that
  proves they moved together.

### 1.5 ★ `rejected` full removal (operator-authorized scope, full-stack — binding)

`rejected` is removed end-to-end. Contract-first for the wire layer; DB-migration for the schema/trigger;
each layer has a proof:

| Layer | Change | Proof |
|---|---|---|
| Domain enum | Delete `DocStatusRejected` (`model.go:12`) | `go build ./...` green; no references remain |
| Reader semantics | Remove `DocStatusRejected` from `activeInstanceStatuses` (`repository/active_instance_reader.go:39`) | The active-instance set is `{draft,under_review,approved,scheduled}`; `active_instance_parity_test.go` updated + green |
| DB CHECK | Migration (next number) tightens `documents_status_check` to the 8-value set (drop `'rejected'`), **with a row pre-check** (fail loudly if any live `status='rejected'` row exists — pattern of migration `0265:75-93`) | Migration applies clean on a DB with zero rejected rows; pre-check aborts if any exist |
| DB trigger | Same migration rewrites `enforce_document_transition` to drop the `under_review→rejected` and `rejected→draft` arcs | Trigger raises `check_violation` on `under_review→rejected` after migration (negative proof) |
| OpenAPI contract | Remove `rejected` from the document-status enum in `api/openapi/**`; **regenerate** BE (`api.gen.go` ×3) + FE (`api-types/index.d.ts`) — **zero hand-edits to generated files** | `oapi-codegen` + `openapi-typescript` regen clean; openapi lint green; the enum no longer lists `rejected` |
| Frontend | Remove `rejected` handling from `parseDocumentStatus.ts`, `documentStatusPresentation.ts`, `StatusPill.tsx`, `documentWorkflow.ts`, `libraryStatus.ts`, `approvalWorkflow.ts` and any test fixtures | `tsc --noEmit` green; `vitest` for documents/approval green; no `rejected` branch remains |
| Tests | Delete/retarget the raw-SQL `under_review→rejected→draft` drive (`application/service_review_roundtrip_integration_test.go:220-232`) — it exercised a now-removed arc | Suite compiles; the removed-arc drive is gone or asserts the trigger now rejects it |

**HS-2 guard:** if removal surfaces a **live** dependence on `rejected` (a production row, an external
consumer, a required audit state) that makes removal unsafe, **stop and surface it** — do not force the
migration. The pre-check is the tripwire.

### 1.6 F4.1 exit criteria (all required)

Unified exhaustive transition function exists in `documents/domain` (typed error, total over the status
set, mirrors templates) · legal set == §1.2 table · **every §0.3 approval service routes through it**;
scattered document-status lifecycle guards census = 0 (or allowlisted with reason) · dead
`CanTransitionDocument` removed (no second document-status FSM) · **app↔DB-trigger parity test green**
(§1.4) · coverage test proves all pairs handled (every `L` allowed, representative `·` rejected) ·
`rejected` removed full-stack per §1.5 with each layer's proof (incl. contract regen with zero
hand-edits, DB migration with row pre-check + negative trigger proof) · `go build ./...` green ·
targeted go tests + vitest green · openapi lint green · M0–M3 gates not regressed.

---

## 2. F4.2 — scheduled-vs-manual publish race (proof, or single choke point)

### 2.1 The race (binding shape)

A **real concurrent integration test** (testdb factory, real Postgres, **NOT sqlmock**), `//go:build
integration`, targeted `-run`. It races the **manual publish** path (§0.4) and the **scheduled cutover**
path (§0.4) against **one** document revision, and exercises **both interleavings**:

- **Interleaving A:** manual publish begins first (grabs the row), scheduled cutover second.
- **Interleaving B:** scheduled cutover begins first (`FOR UPDATE`), manual publish second.

Determinism: the two goroutines are synchronized (e.g. a barrier / shared start signal, or ordered
statement injection) so each interleaving is actually exercised, not left to chance. Both goroutines
target the same `(document_id, revision_version)`.

### 2.2 ★ Expected outcome (the single-winner + terminal-state contract — binding)

For **each** interleaving, the test asserts:

| Assertion | Expected |
|---|---|
| Winners | **Exactly one** of the two publish attempts succeeds (its UPDATE affects 1 row). |
| Loser | The other observes **0 rows affected** and returns its no-op/stale sentinel — manual → `ErrStaleRevision`; scheduled → `errScheduledPublishNoOp` (a **successful no-op**, not an error surfaced to River). |
| Terminal status | The document ends `published` (exactly once). No double-publish, no lost transition, no `revision_version` skipped or decremented. |
| revision_version | Bumped **exactly once** (by the winner). The monotonic trigger (§0.6) never fires. |
| Side effects | Exactly one `document_published` governance event / outbox row for the win (no duplicate publish fanout). |

The precondition that makes the two mutually exclusive is the **status predicate divergence** (manual
requires `approved`, scheduled requires `scheduled`) plus the OCC `revision_version` CAS. The test proves
the mechanism holds under true concurrency.

### 2.3 If the proof fails → single choke point (HS-2-bounded)

If any interleaving shows two winners, a lost transition, or a wrong terminal state, F4.2 adds a **single
`PublishRevision` choke point** both paths route through (one method owning the load-lock-validate-update
sequence), and the race test is re-run against it until green. Introducing the choke point is **in
scope**; anything larger (moving publish enforcement into/out of the DB, a cross-module contract change)
is **HS-2** — stop and surface. **Expected, per §0.4:** the proof passes as-is and no choke point is
needed; the choke point is the contingency, and whichever outcome occurs is recorded in `evidence.md`.

### 2.4 F4.2 exit criteria

Real concurrent integration test (testdb, not sqlmock) green · both interleavings deterministically
exercised · exactly-one-winner + correct terminal state asserted per §2.2 · loser returns the correct
sentinel · single publish side-effect · if the proof failed, a single `PublishRevision` choke point was
added and the test passes against it (recorded) · targeted `-run` only, no full suite · `go build ./...`
green.

---

## 3. F4.3 — concurrency-idiom unification (full analysis + decision)

Operator delegated the choice to a rigorous analysis. Here it is; the **decision is binding** for the
milestone (HS-7 applies to it).

### 3.1 What we have

- **Documents:** optimistic concurrency via the HTTP `If-Match: "vN"` header (RFC 7232 conditional
  request), parsed to `revision_version`. This is the HTTP-native precondition mechanism — the same
  header a caller would use with an `ETag` returned on GET.
- **Templates:** optimistic concurrency via a required JSON body field `expected_lock_version`, mapped to
  `lock_version`. This is an RPC-style in-payload precondition.
- Both are CAS over one integer column; only the transport differs.

### 3.2 What we want to achieve

One idiom for one mechanism across the two kernel modules, so a client (and the generated FE client) uses
a **single** optimistic-concurrency convention, and the contract stops teaching two ways to do the same
thing (the review's dimension-6 minor DEBT). Pre-v1 is the cheap window to cut over (mission thesis).

### 3.3 What a fresh professional implementation would do

Industry convention for HTTP optimistic concurrency is **`ETag` + `If-Match`** (RFC 7232; Google AIP-154
"resource freshness"; Zalando REST guidelines "optimistic locking via ETag/If-Match"; Microsoft/Stripe
patterns). A greenfield professional API exposes the version as an `ETag` on GET and requires `If-Match`
on mutating requests. The body-field approach (`expected_lock_version`) is a workable but non-idiomatic
RPC-ism: it couples the precondition into every request schema, is invisible to HTTP caches/proxies, and
duplicates what a standard header already expresses. Documents already sit on the idiomatic side.

### 3.4 Decision (binding) — ⚠ SUPERSEDED 2026-07-04 by the §3.7 erratum (HS-7)

> **HS-7 erratum, 2026-07-04:** the decision below rested on a premise that F4.3 verification proved
> **false**. It is superseded by **§3.7**. Kept verbatim for audit. **Do not implement §3.4/§3.5 as
> written** — the binding F4.3 decision is now §3.7 (ADR-record the split; ADR 0066).

~~**Unify on the `If-Match` header idiom; migrate templates (the minority) to it; contract-first; record an
ADR.**~~ Rationale (as originally written): it moves both modules to the HTTP-native, standards-endorsed
mechanism (global maximum, not the local maximum of "document whichever split we happen to have");
documents need no change; the migration is bounded to templates' mutating write endpoints; pre-v1 cutover
cost is low and there are no external consumers yet. This **kills** the DEBT (mission intent: "unify …
migrate the minority"), rather than merely annotating it.

**Why superseded:** see §3.7. The "templates is the minority straggler" framing was false runtime truth.

### 3.5 Scope of the templates migration (binding)

- **Contract-first:** in `api/openapi/**`, templates' mutating endpoints that today take
  `expected_lock_version` in the body take an `If-Match` header precondition instead; **regenerate** BE +
  FE (zero hand-edits to generated files).
- **Handler:** templates parse `If-Match` (reuse the documents `parseIfMatch` shape or a shared helper),
  map to `lock_version`. The **column name `lock_version` may stay** (it is internal; renaming it is
  out of scope — the *wire idiom* is what unifies, not the column).
- **The `"vN"` value format:** the header value convention (documents use `"vN"`) is applied uniformly;
  if templates' `lock_version` is a bare integer, the shared helper handles the agreed format — the ADR
  records the exact on-the-wire value grammar so both modules match.
- **Consumers:** the FE template-edit consumers move from sending a body field to sending the header;
  `tsc --noEmit` + `vitest` for templates green.
- **ADR:** `wiki/decisions/` ADR "Optimistic concurrency is transported via `If-Match`, uniformly"
  records the decision, the RFC/AIP/Zalando basis, the templates cutover, and the pre-v1 atomic-cutover
  exception. Cited by the F4.3 commits.

**Fallback (recorded, not expected):** if the templates migration surfaces a hard blocker that makes the
header cutover unsafe within M4 (e.g. a non-regenerable consumer, or a semantic mismatch in what the
version guards), F4.3 falls back to the acceptance-permitted **"ADR-record the split"** — a written ADR
justifying two idioms as intentional — and the migration becomes a tracked follow-on. This fallback is
**HS-7-gated**: taking it requires updating this §3 decision with operator approval, not silently.

### 3.6 F4.3 exit criteria — ⚠ SUPERSEDED 2026-07-04 by §3.7

> Superseded together with §3.4/§3.5. The binding F4.3 exit criteria are now **§3.7**. Kept for audit.

~~One optimistic-concurrency wire idiom (`If-Match`) across documents + templates~~ — templates migrated
contract-first (openapi + regen, zero hand-edits), handler + FE consumers updated, `lock_version` CAS
semantics preserved · ADR landed + cited · `tsc --noEmit` + targeted vitest green · openapi lint green ·
`go build ./...` green · **OR** the §3.5 fallback ADR-split with operator-approved §3 update.

---

### 3.7 ★ CORRECTED DECISION (binding, 2026-07-04 — HS-7 erratum) — ADR-record the split

**What changed.** §3.4/§3.5 above decided "unify on `If-Match`, migrate templates the minority." F4.3
verification (2026-07-04) proved the premise **false**, so per HS-7 (§0 header: "fix the code to the
contract, or re-open this contract with operator approval — never silently edit the contract to match the
code") this §3 is **re-opened with a loud, dated, auditable erratum** and the decision is switched to the
§3.5-permitted fallback. The re-open is surfaced as a **headline HS-7 item at the M4 HS-1 operator gate**
for ratification (nothing is pushed or merged before that gate).

**The false premise (corrected runtime truth):**

| Original claim (§3.4/§3.5, §0.5) | Verified truth (F4.3, 2026-07-04) |
|---|---|
| templates is "the minority" already near the If-Match convention | templates has **zero** If-Match usage anywhere (BE + FE); its `expected_lock_version` body field is its **only** OCC write (`UpdateSchemas`), self-consistent with its own `lock_version` column + `stale_lock_version` error |
| an existing system-wide If-Match standard makes templates non-conformant ("DEC-01") | the cited decision is **CON-01** in `wiki/modules/documents.md` — a **documents-module-internal** decision (submit-canonical-over-finalize), **not** a system-wide OCC ADR. No cross-module If-Match standard exists |
| migrating templates "finishes convergence" (bounded cleanup) | migrating templates **creates a new cross-module OCC standard** and imports a documents-local decision into templates — a first-class architectural change, not a cleanup |

**What still holds:** §3.3's analysis — that `If-Match` is the idiomatic HTTP OCC transport (RFC 7232 /
AIP-154 / Zalando) and the right **long-term target** — is unchanged. Only the "templates is a straggler
to migrate now, inside M4" framing was wrong.

**Corrected decision (binding):** **ADR-record the two-idiom split as intentional** (ADR 0066), naming
`If-Match` as the target transport and deferring full unification to its own deliberate cross-module
change (candidate M9 governance-hygiene or standalone), **not** M4. Rationale: M4's objective is
**versioning kernel correctness**; a cosmetic cross-module transport refactor of a correct, self-consistent
module (templates BE + FE + contract + tests) is scope creep + regression risk M4 has no reason to take.
This is the global-maximum call — record the target and unify deliberately — over the local maximum of
smuggling a risky refactor into a correctness milestone.

**Corrected F4.3 exit criteria (binding, replaces §3.6):** ADR 0066 landed under `wiki/decisions/`
recording the intentional split + If-Match target + deferred-unification charter · contract §3 re-opened
with this dated erratum (not silent) · f4.3 `spec.md`/`plan.md` updated to the corrected decision ·
`evidence.md` records the false-premise correction and the HS-7 disposition · **no templates code/contract
change in M4** (documents + templates both unchanged at the wire) · HS-7 re-open headlined at the HS-1
operator gate. No BE/FE/openapi build steps apply (documentation/decision-only feature).

---

## 4. DB-as-last-line invariants preserved (binding — cross-feature)

- **The DB trigger `enforce_document_transition` stays the authoritative enforcer.** F4.1's app function
  is the *friendly first line* mirroring it; F4.1 does **not** delete the trigger or move transition
  legality into the app only. The `rejected` arc removal (§1.5) edits the trigger **and** the app function
  together (parity, §1.4).
- **Monotonic `revision_version` trigger, status-value CHECK, cancel GUC gate, M2 write-tripwire, M3
  FORCE-RLS** — all stay; no approval-service refactor may weaken or duplicate-as-authoritative any of
  them.
- **No new business logic in triggers.** F4.1 only edits the existing transition trigger's arc list
  (removing `rejected`); it adds no generic logic there.
- **Contract-first for every wire change** (§1.5 OpenAPI enum, §3.5 templates header): `openapi.yaml` +
  regen only; zero hand-edits to generated files.
- **ADR 0013** (revision numbering) semantics **unchanged**.

## 5. Cross-feature constraints (bind all three features)

- Runtime truth beats docs; the DB trigger is the source of truth for the transition graph.
- **Subagent-driven implementation** (sonnet implement/review, haiku mechanical, never fable, ≤15
  concurrent); main session orchestrates/reviews/commits; the `milestone-validator` judges and writes
  `qa/milestone-qa.md`; main session flips status only on PASS.
- **testdb factory** for the F4.2 race + F4.1 parity integration tests (real concurrent, not sqlmock);
  **targeted `-run`** only — the full integration suite is NOT run locally (20+ min box); bounded defers
  recorded in `evidence.md` with triggers.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored (never
  force-add).**
- **HS-7:** implementation compared section-by-section to this contract; drift → fix code to contract or
  re-open contract with operator approval.
- Module boundary: approval-service refactors stay within `documents`; cross-module access via published
  interfaces only (no reach into another module's repository/SQL).

## 6. Bounded defers (recorded, with triggers)

| Defer | Rationale | Trigger / owner |
|---|---|---|
| `archived` status-value cleanup (it is a soft-archive flag never written as `status`, yet occupies the enum/CHECK) | ADR 0010 retains it deliberately; removing it is a separate contract+FE change the operator did not authorize for M4 (unlike `rejected`) | M9 governance-hygiene (dead-code/doc-truth), if desired |
| `approved→draft` / `scheduled→draft` unused DB-legal arcs | Included as legal for app↔DB parity (§1.2 note ²); no app caller exercises them. Pruning them from the trigger is a lifecycle-narrowing decision beyond M4's "unify, don't redesign" scope | Revisit if a future feature needs unapprove/unschedule, or M9 hygiene |
| F4.2 race integration run if the local box cannot run `-tags integration` | 20-min box constraint (mission §10) | Run on CI / capable box before program close-out; drive authored regardless (M1–M3 precedent) |
| Templates `lock_version` **column** rename to match documents' `revision_version` | Out of scope — §3 unifies the *wire idiom*, not internal column names; renaming is churn with no external benefit | M9 structure-hygiene, if desired |
