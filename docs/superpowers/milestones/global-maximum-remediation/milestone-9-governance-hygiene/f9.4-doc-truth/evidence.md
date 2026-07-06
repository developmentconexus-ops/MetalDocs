# Feature F9.4 — Evidence — doc-truth

> **Milestone:** 9 · **Feature:** `f9.4-doc-truth` · **Pass:** INITIAL PASS — Tasks 1-4 only.
> Task 5 (wiki-curator over the enumerated list) and Task 6 (final verification) are **deferred**
> to the FINAL pass, which runs after F9.5 lands (structure-hygiene renames `documents/repository/`
> and `templates/repository/` out of existence and settles the approval nested-vs-promoted decision).
> Running the curator now would fix file:line anchors that F9.5 is about to move again — spec.md
> and plan.md both call this out explicitly ("final verification pass runs after F9.5").
>
> Contract: `../validation-contract.md` §4 (binding).

## Task 1 — runtime extraction (read-only)

### 1.1 Module inventory

```
$ ls internal/modules/
audit auth controlleddocuments distribution documents iam jobs notifications
render search security taxonomy templates tokens
```

14 directories. **No `docs` module. `tokens` exists.** CLAUDE.md previously listed `docs` and omitted
`tokens` — confirmed defect (matches contract §0.1 exactly).

`internal/modules/documents/approval/` is a full-layer subtree (`api application domain http
infrastructure jobs repository`) nested inside `documents` rather than being its own top-level
module — the "hidden 15th module" the contract flags at §0.2. F9.5's mini-gate decides
promote-vs-ADR-exception; this pass only adds a placeholder footnote pointing at that decision.

### 1.2 Middleware chain — `apps/api/cmd/metaldocs-api/chain.go:25-40`

```go
func apiChain(recovery, otel, httpObs, cors, origin, preAuthLoginLimit, authn, iamAuthz, presence, rateLimit, methodNotAllowed func(http.Handler) http.Handler) []chainLink {
	return []chainLink{
		{"panic_recovery", recovery},
		{"otel", otel},
		{"http_obs", httpObs},
		{"cors", cors},
		{"origin_protection", origin},
		{"pre_auth_login_rate_limit", preAuthLoginLimit},
		{"authn", authn},
		{"iam_authz", iamAuthz},
		{"presence_bump", presence},
		{"rate_limit", rateLimit},
		{"method_not_allowed", methodNotAllowed},
	}
}
```

Chain order (outermost first): `panic_recovery → otel → http_obs → cors → origin_protection →
pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit → method_not_allowed`.

**No idempotency link anywhere in this chain.** CLAUDE.md's old wording ("…tier-1 authz→idempotency
→handler") is false on two counts: (a) idempotency is not a chain link at all, (b) the old hand-list
also silently dropped `origin_protection` and `presence_bump`. This matches the spec's own interview
record (§ Interview record row 2) verbatim — no contradiction found.

### 1.3 Janitor / periodic-job set

`internal/modules/jobs/maintenance/periodic.go:29-60` (`PeriodicJobs()`):

```go
func PeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		// stuck-instance-watchdog   — river.PeriodicInterval(5*time.Minute)
		// idempotency-janitor       — river.PeriodicInterval(15*time.Minute)
		// audit-integrity-validator — river.PeriodicInterval(time.Hour)
		// document-review-surfacer  — river.PeriodicInterval(time.Hour)   (M6 F6.2 addition)
	}
}
```
(IDs read verbatim from source: `stuck-instance-watchdog`, `idempotency-janitor`,
`audit-integrity-validator`, `document-review-surfacer`; all `Queue: "maintenance"`.)

`apps/api/cmd/metaldocs-api/main.go:645-672`:
- `metaldocs-api` builds a River `ClientBundle` with
  `PeriodicJobs: append(maintenance.PeriodicJobs(), retention.PeriodicJob())` — i.e. api's River
  client is configured with the SAME 4 maintenance periodic jobs plus the M5 F5.4 staging-outbox
  retention/purge periodic job.
- Comment at `main.go:654-663` (verbatim): "PeriodicJobs is defined here too (not just in
  metaldocs-jobs) because River only enqueues periodic jobs from the elected leader's own
  Config.PeriodicJobs; metaldocs-api joins the same leader election but does NOT subscribe the
  'maintenance' queue and has nil Workers here, so it enqueues-when-leader but never executes these
  jobs (ADR 0067 dual-define, jobs-only execute topology)."
- So: **both binaries join the same River leader election**; whichever wins enqueues; only
  `metaldocs-jobs` subscribes `maintenance` and holds the actual `river.Worker[Args]`
  implementations, so only `metaldocs-jobs` ever executes these jobs.

**Lease-reaper / lease scheduler: retired.** `db/migrations/0273_drop_job_leases.sql` drops the 3
lease SQL functions + `metaldocs.job_leases`; M5 `f5.2-janitors-on-river/evidence.md` census: `grep
-rn "acquire_lease|heartbeat_lease|release_lease|job_leases|pg_try_advisory_lock" --include=*.go
internal apps | grep -v _test.go` → 1 unrelated hit (an outbox TODO comment, not `job_leases`); 0
lease-scheduler refs. `internal/modules/jobs/scheduler/` package deleted (T4, commit `b067f3a1`).
CLAUDE.md's old "4 leader-elected janitors — …, lease-reaper" wording named a component (the
lease-reaper) that no longer exists and undercounted the actual periodic-job set (which is 4
janitors + 1 retention purge, not "4 leader-elected janitors" as a closed set).

**Watchdog is alert-only**, not auto-cancel — `wiki/decisions/0068-stuck-instance-watchdog-alert-only.md`:
"The stuck-instance watchdog is alert-only. Every instance `in_progress` past the 7-day [threshold]…
[no auto-cancel]." Confirmed no contradiction with spec's interview record.

**No spec-vs-runtime contradiction found on this item** — runtime matches what spec.md's interview
record already asserted (row 3). Note for the record: the extracted set is 4 periodic jobs (3
original janitors + document-review-surfacer added in M6, + retention purge appended separately in
api's Config), not literally "3 janitors" nor "4 leader-elected janitors" as CLAUDE.md's retired
wording implied — corrected wording below names the mechanism (River periodic + leader election)
rather than hand-counting a job list that keeps growing (M6 already added one job after the M5
evidence was written).

### 1.4 Idempotency execution sites

`grep idempotency_keys` under `internal/` (files, not lines) — key hits:
- `internal/modules/documents/approval/infrastructure/postgres_route_admin_idemp_store.go`
- `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go`
- `internal/modules/documents/approval/application/signoff_idemp.go`
- `internal/platform/idempotency/postgres_store.go` (shared platform store)
- `internal/modules/jobs/idempotency_janitor/job.go` (the janitor that sweeps `idempotency_keys`,
  distinct from the enforcement sites above)

Idempotency enforcement is per-handler/per-service (approval signoff, approval route-admin, and any
consumer of `internal/platform/idempotency`), never a step in the shared `chain.go` middleware list.

## Task 2 — CLAUDE.md corrections

| # | Old text | New text | Runtime evidence |
|---|----------|----------|-------------------|
| 1 | "4 binaries: `metaldocs-api` (sync + authz, stateless; also hosts the 4 leader-elected janitors — stuck-instance-watchdog, idempotency-janitor, audit-integrity-validator, lease-reaper), `metaldocs-worker` …, `metaldocs-jobs` (async schedules via River — scheduled-publish + notifications fanout), `docx-renderer`…" | "4 binaries: `metaldocs-api` (sync + authz, stateless; also joins River leader election so it can enqueue the periodic maintenance jobs — `stuck-instance-watchdog`, `idempotency-janitor`, `audit-integrity-validator`, `document-review-surfacer`, plus the outbox-retention purge — though only `metaldocs-jobs` subscribes the `maintenance` queue and actually executes them, per ADR 0067's dual-define pattern; `apps/api/cmd/metaldocs-api/main.go:645-672`), `metaldocs-worker` (async outbox consumers), `metaldocs-jobs` (hosts + executes the River periodic jobs from `internal/modules/jobs/maintenance/periodic.go`, plus scheduled-publish + notifications fanout), `docx-renderer` (internal only). The stuck-instance watchdog is alert-only, not auto-canceling (ADR 0068); the old Postgres-lease scheduler and its `lease-reaper` are retired (M5)." | `periodic.go:29-60`; `main.go:645-672` (comment verbatim quoted above); `db/migrations/0273_drop_job_leases.sql`; `f5.2-janitors-on-river/evidence.md` census; `wiki/decisions/0068-...md` |
| 2 | "**14 bounded-context modules** under `internal/modules/`: audit · auth · controlleddocuments · distribution · docs · documents · iam · jobs · notifications · render · search · security · taxonomy · templates." | "**14 bounded-context modules** under `internal/modules/`: audit · auth · controlleddocuments · distribution · documents · iam · jobs · notifications · render · search · security · taxonomy · templates · tokens. … (`documents/approval` is a nested exception inside `documents` rather than its own top-level module — ADR: approval nested exception, F9.5.)" | `ls internal/modules/` (Task 1.1 above) — exact match, count stays 14 |
| 3 | "**Fixed request lifecycle.** Middleware chain (panic→trace→obs→cors→rate-limit→authn→tier-1 authz→idempotency→handler) is inherited; new routes don't reinvent auth/validation/errors." | "**Fixed request lifecycle.** Middleware chain is `panic_recovery → otel → http_obs → cors → origin_protection → pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit → method_not_allowed` (`apps/api/cmd/metaldocs-api/chain.go:25`); new routes don't reinvent auth/validation/errors. Idempotency is **not** a chain link — it is enforced per-handler/per-service where needed (e.g. `internal/modules/documents/approval/application/signoff_idemp.go`, `internal/platform/idempotency/postgres_store.go`)." | `chain.go:25-40` (Task 1.2 above) |

No spec-vs-runtime contradictions surfaced on any of the three corrected claims — runtime matches
what spec.md's own interview record (B1.5) already asserted. All three corrections applied to
`CLAUDE.md` as edits (not a restructure — no new sections added).

## Task 3 — skill reference fix

`.claude/skills/developing-new-work/references/invariant-checklist.md`:
- Line ~56 (old): "**Fixed request lifecycle**: panic→trace→obs→cors→rate-limit→authn→tier-1
  authz→idempotency→handler is inherited; new routes don't re-wire it.
  `apps/api/cmd/metaldocs-api/chain.go:25`."
- New: "**Fixed request lifecycle**: `panic_recovery → otel → http_obs → cors → origin_protection →
  pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit →
  method_not_allowed` is inherited; new routes don't re-wire it.
  `apps/api/cmd/metaldocs-api/chain.go:25`. Idempotency is **not** a chain link — it is enforced
  per-handler/per-service where needed (e.g.
  `internal/modules/documents/approval/application/signoff_idemp.go`)."
- `Last verified` bumped: `2026-06-28` → `2026-07-06`.

## Task 4 — mission-touched wiki docs (enumeration)

Method: union of (a) `git log --name-only 1eed326e^..HEAD -- wiki/` (mission range; `1eed326e` =
`docs(mission): scaffold global-maximum-remediation program` — the first mission commit) and
(b) every `wiki/*.md` path literal grepped out of
`docs/superpowers/milestones/global-maximum-remediation/*/f*/evidence.md`. Union taken because (a)
alone misses docs that evidence files *reference* as context without a wiki-side commit in range,
and (b) alone misses docs actually edited but not narrated by path in prose.

Commands run:
```
git log --name-only --pretty=format:"" 1eed326e^..HEAD -- wiki/ | sort -u
grep -rohE 'wiki/[A-Za-z0-9/_.-]+\.md' docs/superpowers/milestones/global-maximum-remediation --include=evidence.md | sort -u
```

**43 unique wiki-file paths** (curator pass pending final pass, i.e. Task 5 — NOT run this pass):

| # | Path | Source |
|---|------|--------|
| 1 | `wiki/README.md` | evidence |
| 2 | `wiki/architecture/api-contract.md` | git log |
| 3 | `wiki/architecture/backend-blueprint.md` | evidence |
| 4 | `wiki/architecture/backend-target-architecture.md` | evidence |
| 5 | `wiki/architecture/req-trace-map.yaml` | git log |
| 6 | `wiki/architecture/req-traceability.md` | git log |
| 7 | `wiki/architecture/tenant-context.md` | git log |
| 8 | `wiki/backend/_artifacts/stage1/module-iam.md` | git log |
| 9 | `wiki/backend/legacy-register.md` | evidence |
| 10 | `wiki/backend/repo-topology.md` | evidence |
| 11 | `wiki/backend/stage2-evaluation.md` | evidence |
| 12 | `wiki/concepts/approval-routes.md` | evidence |
| 13 | `wiki/concepts/authz-tiers.md` | git log |
| 14 | `wiki/database/tables/approval_signoffs.md` | git log |
| 15 | `wiki/database/tables/documents.md` | git log |
| 16 | `wiki/decisions/0013-template-revision-labels.md` | git log |
| 17 | `wiki/decisions/0015-async-freeze-pin-materialize.md` | git log |
| 18 | `wiki/decisions/0018-approval-route-lifecycle.md` | git log |
| 19 | `wiki/decisions/0022-authz-capability-coherence.md` | git log |
| 20 | `wiki/decisions/0022-execution-history.md` | git log |
| 21 | `wiki/decisions/0027-rls-adoption-sequencing.md` | git log |
| 22 | `wiki/decisions/0052-template-manual-versioning.md` | git log |
| 23 | `wiki/decisions/0065-version-references-are-nested-value-objects.md` | git log |
| 24 | `wiki/decisions/0066-optimistic-concurrency-transport-split.md` | git log |
| 25 | `wiki/decisions/0067-async-job-infrastructure-consolidated-onto-river.md` | git log |
| 26 | `wiki/decisions/0068-stuck-instance-watchdog-alert-only.md` | git log |
| 27 | `wiki/decisions/0069-document-periodic-review-and-reason-for-change.md` | git log |
| 28 | `wiki/decisions/0070-tenant-lifecycle-onboarding-export-crypto-shred-erasure.md` | git log |
| 29 | `wiki/decisions/0071-distributed-rate-limiting-shared-store.md` | git log |
| 30 | `wiki/decisions/index.md` | git log |
| 31 | `wiki/modules/approval-tech-debt.md` | evidence |
| 32 | `wiki/modules/approval.md` | git log |
| 33 | `wiki/modules/documents.md` | git log |
| 34 | `wiki/modules/iam.md` | git log |
| 35 | `wiki/modules/jobs.md` | git log |
| 36 | `wiki/modules/templates.md` | git log |
| 37 | `wiki/quality/index.md` | git log |
| 38 | `wiki/quality/integration-test-harness.md` | git log |
| 39 | `wiki/quality/legacy-test-policy.md` | git log |
| 40 | `wiki/quality/test-discipline.md` | git log |
| 41 | `wiki/runbooks/backup-restore.md` | git log |
| 42 | `wiki/runbooks/index.md` | git log |
| 43 | `wiki/standards/documentation-governance.md` | git log |

This list is the Task-5 input for the FINAL pass. **Not curated this pass** — per the sequencing
note at the top of spec.md/plan.md, running wiki-curator now would fix file:line anchors pointing
into `internal/modules/documents/repository/` and `internal/modules/templates/repository/`, which
F9.5 is about to rename out of existence. Curating now would be wasted/wrong work.

## Deferred to FINAL pass (post-F9.5)

- **Task 5** — wiki-curator dispatch over the 43-doc list above (stamps + anchors + README index).
- **Task 6** — re-run Task 1 extraction after F9.5 lands; confirm CLAUDE.md module inventory and any
  layout wording still exact against the post-rename tree (F9.5 may promote `documents/approval` to
  its own module, changing the count to 15, or ratify the ADR exception, replacing the
  "(ADR: approval nested exception, F9.5)" placeholder in CLAUDE.md with the real ADR number).

## Deviations / contradictions flagged

**None.** All three runtime facts extracted in Task 1 matched what spec.md's interview record (B1.5)
already asserted from its own prior verification — no spec-vs-runtime conflict requiring an HS-6
stop. The one wording nuance worth flagging for the record (not a contradiction, a precision note):
the retired "4 leader-elected janitors" framing undercounted even before this pass — M6 added a 4th
periodic job (`document-review-surfacer`) after the M5 evidence that introduced the "4 janitors"
framing was written, and a 5th periodic job (outbox retention/purge) is appended separately in api's
`main.go`. The corrected CLAUDE.md wording avoids hand-counting the job list at all and instead names
the mechanism (River periodic jobs + leader election, dual-define per ADR 0067) so it does not go
stale again the next time a periodic job is added.

## Diff scope (this pass)

- `CLAUDE.md` (System Facts section only — module list, lifecycle line, binaries/janitor sentence).
- `.claude/skills/developing-new-work/references/invariant-checklist.md` (lifecycle line + stamp).
- `docs/superpowers/milestones/global-maximum-remediation/milestone-9-governance-hygiene/f9.4-doc-truth/evidence.md` (this file).

No `wiki/**` edits this pass (Task 5 deferred). No `internal/modules/` edits. No commits made.
