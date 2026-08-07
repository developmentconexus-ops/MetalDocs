# Synthesis — ARM=claude

## 1. Axes

| # | Axis (named for cause) | Root cause | Member findings | Scale | Blocks |
|---|---|---|---|---|---|
| A1 | **Inert control plane** | Controls were counted as done when written, never validated to fire — the ME-13 error class applied to tooling itself | CICD §3 (13 unwired check scripts incl. the complete backup-restore gate), SEC-05 (govulncheck default-skip in both wired paths), GOI-01/02/03 (no formatters stanza, `only-new-issues: true`, 15 linters), FE-01 (0 ESLint quality rules), FE-02 (token gate currently red), FE-06 (gate checks hex only), TEST-02 (13 guard scripts, 0 negative tests), PERSIST-09 (baseline-equivalence unwired), CICD §4 + OBS-12/14 (4 CLI tools `@latest`, minio `:latest`, `-mod=mod` in vendored builds), release-readiness dispatch-only | ~30 controls that exist and do not fire, or fire non-reproducibly | Every other axis: a fix that lands without a firing gate regresses; per the brief's second constraint, each of these axes must be treated as if the control were absent |
| A2 | **Post-merge verification topology** | Integration-suite cost pushed 99% of it past the merge; no single local verify entry point ever existed | TEST-01/CICD §6 (548/553 integration funcs, all `-race` integration, post-merge only), CICD §5 (no `make verify`; green-local ≠ green-CI for most of 20 workflows), FE-09 (vitest cannot produce a summary), TEST-05 (backend e2e is `.gitkeep`), TEST-06 (30 untriaged `t.Skip`), TEST-03 (allowlist doc drift), TEST-04/07 (distribution 1 test file, iam ratio 0.44) | 43,691 LOC / 342 files of integration tests giving first signal after merge | The in-loop agent gate (§4): an AI author's cheapest feedback loop is structurally blind to a third of the correctness surface |
| A3 | **Fail-open runtime privilege** | The DB schema assumes a least-privilege role that no deployed environment actually connects as; crypto fail-open treated as first-class | SEC-03 (app connects as Docker superuser, nothing asserts otherwise), SEC-04 (audit REVOKE + FORCE RLS both no-ops vs superuser), SEC-08 (KEK unset → silent no-op crypto); riders fixed in the same package: SEC-01 (reauth timing oracle), SEC-07 (no session cap), SEC-09 (zip-bomb depth, verify-then-fix), PERSIST-07 (auth_failure_counters tenant question) | 37 RLS tables + the whole Part 11 audit-tamper-evidence chain inert in the only compose stack that exists | Trusting *any* DB-enforced invariant; every "DB enforces invariants" claim in CLAUDE.md is conditional on this axis |
| A4 | **Missing request-boundary kernel** | No platform-owned handler kit existed, so 13 modules each hand-built error emission, actor extraction, validation, pagination, idempotency, and concurrency headers | HTTP-01 (6 error mechanisms, 521 sites, 12+ local defs), HTTP-02/12 (dead `ActorFromContext`, per-module wrappers), DUP-01 (dual `UserIDFromContext`, 17 fail-open sites), HTTP-03 (4 named + 14 inline error mappers), HTTP-04 (11 inline request structs), HTTP-05 (123 spec constraints, 0 runtime validation), HTTP-06 (2 pagination envelope shapes), HTTP-07 (If-Match 1/13 modules), HTTP-08 (idempotency ×4 wirings), HTTP-11 (ClampLimit at 2 layers), DUP-05/10 (drifted limit defaults), SEC-06 (no security headers anywhere — a one-middleware fix once the chain is the delivery spine), OBS-09/10 (`writeProblem` logs nothing; `Instance` never set — solved for free by a single error writer) | ~700 call sites across 13 modules re-solving 7 boundary concerns | Any change to problem+json framing, correlation, or validation policy is a 12+-place edit; the stale-write hazard class (no If-Match) stays open module by module |
| A5 | **Persistence without a mechanical spine** | No schema-checked query layer and no enforced tx chokepoint, so every repository hand-rolls lifecycle, scanning, and pg error mapping | PERSIST-01/02 (82 `BeginTx` sites/25 files bypass TxRunner; 2 modules built second tx ports), PERSIST-03 (dead `DoReadOnly` — delete, G1 already ruled), PERSIST-04 (23505 ×15 sites, dead `lib/pq` branches + the lib/pq dep itself), PERSIST-05 (242 hand scans / 16 helpers, 0 sqlc), PERSIST-06 (bespoke filter builders), DUP-04 (20 Tx/non-Tx hand-copied pairs incl. duplicated `authz.Require` lines), DUP-06/07/08 (53 hand-written clone pairs, concentrated in approval/iam/templates/documents repos), PERSIST-08 (confirmed N+1) | 87 files of raw SQL; a column rename is a grep-and-pray across every one | A7 (moving tables between merged modules is manual archaeology without a generated query layer); any TxRunner-level policy change (retry, metric, isolation) is a 25-file edit |
| A6 | **Ungoverned module seams** | No ruled ownership pattern for cross-module seams — the correct shape (consumer-declared interface + composition-root adapter) exists exactly once | LAYERING-03/04 (producer-owned types/funcs/interfaces at every sampled seam vs LAYERING-05's lone counter-example), LAYERING-06 (domain→domain enum), LAYERING-07 (62 cross-module `errors.Is` on undeclared contracts), LAYERING-08/09 (documents↔approval foreign-table SQL, 16 sites), DUP-02 (~30 sentinel declarations, 15 collided names), LAYERING-01/02 (9/15 domain packages import `database/sql`/`platform/db` — needs a ruling, not 9 bug fixes), LAYERING-12/13 (platform→module 20 edges incl. tenantdata registry), DUP-03 (12 hand-written `tenant_data_port.go` skeletons), GOI-07 (approval/application: 583 exported symbols, 30 files), the 7 known module cycles | 6 module pairs coupled through producer internals; check-module-boundaries.ps1 sees none of it | Independent module evolution; the "published Go interface" rule (REQ-TOP-1) is currently prose |
| A7 | **ADR 0093 unimplemented** | The documents/controlleddocuments/templates split is ruled one bounded context (template = version-scoped role); the code still ships three modules | ADR 0093 itself; LAYERING-10 (approval reads `controlled_documents` directly); dissolves on landing: the documents↔controlleddocuments seams inside LAYERING-03, the `ErrApprovalRouteMissing` ×3 inside DUP-02, 2 of the 7 module cycles | 3 modules → 1 context; largest single structural move in the program | The document domain's conceptual integrity; every new document feature currently picks one of three homes by convention |
| A8 | **ADR 0092 unimplemented (grant model)** | AuthZ grants are dual-source (tier-1 roles+groups vs tier-2 user_process_areas — ADR 0007 ratified an unfinished migration); the unification spec is written and waiting | ADR 0092 spec; SEC-02 (audit/distribution/search/security have zero tier-2 `authz.Require` and zero tripwire arms — one tier where the model demands two) | 4/15 modules single-tier; two grant sources that can disagree | The capability model's single-source claim; both prior advisory rounds ruled this must not be shelved behind the artifact axis |
| A9 | **Async half of operations never built** | Observability was built request-path-first; the worker/jobs fleet and everything downstream of /metrics got nothing | OBS-01/02 (worker+jobs: no listener, no healthcheck), OBS-03/04 (2/15 modules traced, 0 async spans), OBS-05/06 (dual metrics systems, hardcoded-zero outbox gauges), OBS-07 (trace = string grep), OBS-08 (no slog context bridge, 79% context-less log calls), OBS-11 (no scheduler log, no River UI), OBS-15 (0 alert rules in-repo), OBS-13 (3 near-identical Dockerfiles) | Half the running fleet is blind; incident response = grep two stdout streams for a string | Any SLA on outbox lag or job latency; post-v1 operability |
| A10 | **Frontend enforcement vacuum (code half)** | With no gate ever on (A1's FE findings), structure rules stayed prose and dead code accumulated | FE-03 (4 god components to 738 LOC), FE-04 (11/14 features missing barrels), FE-05 (66 cross-feature imports on a 19-pair hand allowlist), FE-08 (tsconfig strict-only), FE-10 (13 dead `src/` files incl. an abandoned `canvas/` subtree, 48 dead exports), FE-07 (duplicate formatters), HTTP-10 (generated API client used by 4 files; typed surface bypassed) | ~43k TS LOC with structure rules enforced nowhere | Frontend maintainability claims; the wiki's frontend rules currently describe an aspiration |

**Axis paragraphs (one each):**

**A1.** This is the program's own ME-13 mirror: the repo *has* a control for nearly everything the lanes found, and roughly thirty of those controls are inert — unwired, defaulted off, scoped to new lines only, red-and-ignored, or non-reproducible because the tool it runs floats on `@latest`. Treating them as absent (the brief's binding rule) is what turns "we have govulncheck/backup-verification/lint" into the honest "we don't." Fixing this axis first converts every later axis's fix from discipline into mechanism.

**A2.** Distinct from A1: these controls fire, but after the merge. 548 of 553 integration test functions, all integration `-race` coverage, and most perf signal arrive once the change is already on main. For a repo whose authors are agents, the PR check *is* the review; today it is blind to a third of the correctness surface. The companion defect is local↔CI parity: there is no manifest of the ~25 checks CI runs and no single command that runs them, so "green locally" is unverifiable in principle.

**A3.** The single sharpest finding in the program. Postgres GRANT/REVOKE and FORCE RLS are no-ops against a superuser, and the only deployable compose stack connects as one, unverified. Two independently well-built controls — 37-table tenant RLS and the audit-chain REVOKE — are simultaneously defeated by one wiring fact. For an ISO 13485 / Part 11 product this is not hardening backlog; it is the difference between having tamper evidence and describing it. Cheap to fix (dedicated `metaldocs_app` role + a boot-fatal `rolsuper/rolbypassrls` assertion), and the boot assertion makes regression unrepresentable-in-deployment.

**A4.** Every delivery-layer concern that a framework normally owns — writing an error, naming the actor, validating a body, paginating, idempotency, optimistic concurrency, security headers — was solved per-module by whichever agent got there first. The fix is one platform handler kit: a single problem writer (which is also where OBS-09/10's log-on-error and `Instance` correlation land for free), one actor helper (killing the fail-open `UserIDFromContext` twin), `openapi3filter.ValidateRequest` as middleware (turning 123 spec constraints from documentation into enforcement), and the approval module's post-ADR-0089 error registry promoted from local exemplar to platform default.

**A5.** The persistence layer is 100% artisanal: hand SQL, hand scans, hand tx lifecycle, hand pg-error decoding, with the shared TxRunner bypassed at 82 sites and two modules having built their own second tx port. The global-maximum structure here is a generated query layer (sqlc) plus a cilint-enforced TxRunner chokepoint: scan/SELECT mismatch becomes a compile error, the 20 Tx/non-Tx twins collapse to one generated function each, and the dead `lib/pq` branches go with the dependency. Adopt incrementally, repository by repository — a big-bang rewrite is neither needed nor wanted.

**A6.** The repo knows the correct seam shape — `dictionary_reader_adapter.go` proves it — and applies it once out of six-plus seams. Everything else is producer-owned types, producer-owned interfaces, raw `errors.Is` on foreign sentinels, and SQL straight into another module's tables, which no existing check can see (a boundary checker cannot read a SQL string, but a lint that diffs SQL table references against the `tenant_data_port` ownership registry can — the registry already exists). This axis is a ruling (consumer-owned ports; is `*sql.Tx` in domain ports a sanctioned convention or not) plus mechanical enforcement, then migration seam by seam.

**A7.** Ruled under the ME-13-corrected method in the prior round; what remains is execution. Sequenced after A4/A5 deliberately: with a request kernel and a generated query layer, merging three modules is mostly moving declarations and re-pointing generated queries; without them it is hand-editing hundreds of hand-written SQL strings and three sets of hand-rolled delivery plumbing. Migration cost is admissible for sequencing, and it says: spine first, merge second.

**A8.** The spec exists, both advisory arms twice ruled it must not wait behind the artifact axis, and SEC-02 shows the cost of the current state: four modules run one authz tier where the architecture demands two. Slotting it early (after A3, so authz tests run against an honest DB role) also means A4 and A7 land on the unified grant model instead of migrating twice.

**A9.** The worker and jobs binaries — the half of the fleet that executes the regulated async work — expose no health, no metrics, no spans, no dashboard, and the repo contains zero alert rules. The api binary's observability is real; the async half was simply never built. Minimum professional bar: a health/metrics listener in both binaries, compose healthchecks, W3C context through the outbox (the `TraceID` column exists; it carries the wrong thing), a context-aware `slog.Handler`, River UI, and a starter Prometheus/Alertmanager rule set — all free OSS.

**A10.** Once A1 turns the frontend gates on, this axis is the cleanup they will demand plus the structure rules that need cheap mechanical checks (400-line ceiling, barrel presence, allowlist ratchet). The knip-flagged `canvas/` subtree reads like an abandoned parallel implementation and needs a human call before deletion.

## 2. Noise — real, but no program slot

- GOI-10 (generics unused) — not a defect; revisit only if A5's generated layer leaves `[]any` helpers standing.
- GOI-08 (1 constructor returns interface) — one-line fix in passing during A5; not a slot.
- GOI-04 (110 panic sites) — majority are wiring-time guards; a half-day triage inside A2's test work, not an axis.
- GOI-06 (capitalized error strings) — dissolves when A1 removes `only-new-issues` and runs full-repo lint.
- GOI-09 (7 unsupervised goroutines) — inspect during A9's worker pass; likely benign sweepers.
- PERSIST-08 (one confirmed N+1) — fix inline this week; a program slot for one UPDATE loop is overhead.
- FE-07 (2 duplicate formatters) — inline fix when touching those files.
- FE-11 (21 knip-flagged deps) — verify before believing (react-pdf/react-icons are suspect false positives); prune opportunistically.
- HTTP-09 (`withAdminCtx`) — module-local and legible; absorbed by A4 if the kernel makes it redundant, otherwise leave.
- TEST-03 (allowlist doc drift) — 20-minute doc fix + confirm operator approval of `c1b37817`; fold into A2.
- OBS-12/14 pinning specifics — folded into A1's one pinning pass; listed here so nobody makes them a workstream.
- LAYERING-11 (iam fan-in) — informational; fan-in on the authz core is expected, not actionable.
- DUP-09 (dual `ErrActorAlreadySigned`) — intended defense-in-depth per the stated invariant; at most a rename for the identifier collision.
- CICD lane's 359 vs testing lane's 342 integration-file counts — measurement-method difference, not a finding.

## 3. Do-not-touch (consolidated from the lanes' "actually fine" sections)

- `api.gen.go` clones across 15 modules — correct codegen output; never count them as duplication.
- Middleware chain (`chain.go`) and its test — structure is sound; A4 adds below it, not into it.
- Auth login/session core — bcrypt 12, constant-time, HMAC-signed hashed tokens, atomic mutation+audit; deliberate, competent, leave it (A3's SEC-01 rider touches only the *reauth* path).
- `internal/platform/config/*` — typed, validated, per-domain config; the shape to preserve.
- `tools/cilint` + `scripts/api-lint` — the two self-tested guard layers; they are the model A1 holds the shell scripts to.
- Outbox: single INSERT site, network calls outside tx — the invariant holds; don't churn it.
- Tenant predicate discipline (83/87 SQL files) and parameter binding ($N everywhere) — strong; A5 must not regress it.
- Migration gapless/no-historical-edit CI check; audit hash chain design (IMMUTABLE SQL function — undermined only by A3, not by itself).
- secret-scan (gitleaks, full history, triaged allowlist); Grype CVE gate; workflow caching; `api-contract.yml`'s 7-job design.
- Approval module's error registry (`approval/http/errors.go`) and `dictionary_reader_adapter.go` — these are the *targets* A4 and A6 generalize from.
- `pagination` platform helper; templates' documented 50/200 deviation (the reference for how to record divergence).
- Metrics-listener isolation (infra port); readiness probe depth; single slog JSON logger.
- Frontend: query-key discipline (103 QK sites), Zustand two-store rule, clean `tsc --noEmit`, madge-clean import graph, sanctioned `fetch()` tier-2/3 sites, `role-vocabulary.ts`.
- `main.go` hand-wiring (1312 lines, no DI framework) — legible, traceable; do not introduce a DI container.
- No util/helpers/common packages; go.mod zero replaces — keep both properties.

## 4. Target methodology

### 4.1 Blocks vs annotates

**Blocking (a wrong failure here is rare and cheap to see):** build + vet + full-repo golangci-lint with formatters (after one triage pass baselines the backlog); codegen drift (both directions); api-lint `-strict`; cilint; migration gapless; secret scan; openapi-breaking; unit tests; module-scoped integration tests (see 4.2); Grype on lockfile change; govulncheck (annotate for a 2-week triage window, then flip to block — flipping day one on an untriaged backlog teaches bypass); boot-fatal DB-role assertion (A3); token-discipline and ESLint boundary+quality rules; frontend typecheck+vitest.

**Annotating only (verdict requires judgment or the tool false-positives):** jscpd duplication ratio; knip dead code/exports (dynamic-import false positives); coverage deltas; k6 perf until thresholds are written into the scripts and validated (today it is unverified whether perf can fail at all — say so in the job name); axe-diff (currently `|| true`; keep soft but surface the artifact in the PR, don't swallow it silently).

Rule: every blocking check must have a negative fixture proving it fails on bad input (extend the existing test-framework hard gate to shell/PS guards — this is TEST-02's fix and the only durable answer to "does the guard fire").

### 4.2 Where each gate fires, per axis (highest achievable level, and why not higher)

| Axis | Highest level | Mechanism | Why not higher |
|---|---|---|---|
| A1 | red build | pinned tools + negative-fixture rule + `make verify` manifest | can't make "the gate is wired" unrepresentable; a CI meta-check that every `check-*` script is referenced by a workflow gets close |
| A2 | red build (pre-merge) | affected-package integration tests on PR (path→module map, tags compiled); full suite stays post-merge + nightly | full suite pre-merge = 20-min PRs; selective-per-touched-module is the proven policy (integration-ladder memory) promoted into CI |
| A3 | boot-fatal | startup asserts `NOT rolsuper AND NOT rolbypassrls` and KEK-present-when-crypto-required, refuse to serve | unrepresentable would need infra-as-code the repo doesn't have; boot-fatal is equivalent in practice |
| A4 | red build | one platform writer/kit + cilint ban on local `writeProblem`/`WriteError`/inline decode; spec validation is runtime assertion by construction (openapi3filter middleware — present by default, omission is the exception) | Go cannot forbid writing to a ResponseWriter; the ban-lint is the ceiling |
| A5 | unrepresentable (per adopted query) | sqlc: scan/column mismatch = compile error; TxRunner chokepoint via cilint ban on `BeginTx` outside `platform/db` | hand-SQL stragglers remain red-build until migrated; that's the honest transitional state |
| A6 | red build | import-graph cycle check (go-arch-lint or a cilint analyzer — the current script checks layer only); SQL-table-ownership lint diffing query table refs against the tenantdata registry; consumer-owned-port rule as a lint on cross-module type refs | boundaries between packages in one Go module cannot be compiler-enforced; lint is the ceiling |
| A7 | unrepresentable (after merge) | one module = intra-package access is simply legal; the seam ceases to exist | — |
| A8 | DB runtime assertion | unified grant table + tripwire arms generated from the registry (the GMR-M2 pattern, already proven) | grants are data; the DB is the right last line |
| A9 | runtime assertion | healthchecks gate compose; alert rules fire on lag/dead-letter; span propagation testable in integration | observability can't be a build property |
| A10 | red build | ESLint quality rules + LOC-ceiling and barrel checks + shrink-only allowlist ratchet (count may only decrease, checked in CI) | — |

### 4.3 The in-loop agent gate

The author is a machine whose failure mode is fluent, internally consistent wrongness — the wrong noun used everywhere including inside the guard meant to catch it. Post-hoc review doesn't catch that; two mechanisms do:

1. **Single-source vocabularies, compiled.** Wherever a noun matters (capabilities, problem codes, table ownership, queue names, allowlists), it must live in exactly one generated-from registry so a wrong noun is a compile/lint failure, not a review catch. The repo already proved this pattern three times (tripwire arms from Go registry, problem-code freshness, req-trace). A4/A5/A6 each extend it (error writer registry, sqlc schema, table-ownership lint). An agent cannot consistently misuse a noun the compiler owns.
2. **`make verify` as the authoring loop, not the CI mirror only.** One manifest-driven entry point (`just`/Make) that runs the exact pre-merge set, with a `verify-fast` tier (fmt, vet, scoped lint, affected unit tests) cheap enough to run *during* authoring — wired as a Claude Code hook on edit/stop so the gate fires mid-loop, not at PR time. The manifest is the parity contract: CI runs `make verify`, nothing else, so local-green and CI-green are the same claim. This is the single highest-leverage methodology artifact in the program.
3. **Negative-fixture discipline for guards** (4.1): an agent that writes a guard must also write the input that trips it, in the same PR — the only structural defense against a guard that shares the author's blind spot.

### 4.4 Separation of powers

Facts: solo operator, agent authors, private repo, free plan — branch protection returns 403 and required-checks configuration is unverifiable from the repo. Honest ceiling: **hard prevention is unattainable at zero spend on a private free-tier repo.** What is achievable, mechanically:

- **Gate-changes-ride-alone check** (new blocking workflow): fail any PR that modifies `.github/workflows/**`, `.golangci.yml`, `scripts/api-lint/**` allowlists, `.gitleaks.toml`, `eslint.config.mjs`, or `axe-baseline.json` *and* anything else in the same PR. One commit can no longer loosen the judge and satisfy it simultaneously. (The check can be deleted by a PR — but only by a PR that visibly touches only gate files, which is the observable event the operator reviews.)
- **Shrink-only ratchets:** every allowlist/baseline carries a committed count; CI fails if the count grows without a `RATCHET-GROW:` trailer in the commit message naming the operator's approval. Growth becomes a logged, greppable decision instead of a silent edit.
- **Post-merge tamper journal:** a scheduled workflow hashes the gate-config set on main and appends to a journal file; any delta opens an issue. Detection, not prevention — stated as such.
- **Role split for agents:** the operator's standing model (implementer agents never merge; the hub/operator merges) is the human half; the three mechanisms above are the machine half. If the repo ever goes public or Pro, flip on real branch protection + required checks and the meta-workflow becomes defense-in-depth.

### 4.5 Named free tooling

| Tool | Replaces | Run cost |
|---|---|---|
| sqlc | 242 hand scans, Tx/non-Tx twins, 23505 decoding | codegen step in `make verify`; incremental adoption |
| kin-openapi `openapi3filter` (already vendored) | 123 hand-reimplemented spec constraints | one middleware, µs per request |
| golangci-lint v2 full-repo + formatters + {unused, ineffassign, unparam, wrapcheck, noctx, goconst, nolintlint, misspell} | the diff-scoped 15-linter config | one-time backlog triage, then CI as today |
| go-arch-lint (or one new cilint analyzer) | check-module-boundaries.ps1's cycle blindness | seconds |
| just (or Make) + verify manifest | the absent local↔CI parity | free |
| pinact/frizbee | tag-pinned actions → SHA-pinned; plus version-pin the 4 `@latest` CLIs | one pass + renovate-style bumps via dependabot (free) |
| riverqueue/riverui | hand-querying `river_job` | one container |
| Prometheus + Alertmanager + Grafana OSS | the zero in-repo alert rules; converge OBS-05's dual metrics onto the Prom registry (JSON endpoint becomes a thin adapter or dies) | 3 containers, compose-profile-gated |
| typescript-eslint recommended-type-checked + react-hooks | FE-01's empty ruleset | lint time |
| knip, jscpd | dead-code and clone annotation | annotate-only jobs |
| govulncheck (wired, default ON) | the theater version | ~1 min/PR |

## 5. Sequence

| Order | Axis | Size | Depends on | Parallel with |
|---|---|---|---|---|
| 1 | **A1 control-plane activation** | 3–5d | — | A3 |
| 1′ | **A3 runtime privilege** | 2–3d | — | A1 |
| 2 | **A2 pre-merge topology + verify manifest** | 4–6d | A1 (pinning, manifest) | A9 |
| 3 | **A8 grant unification** | 4–6d | A3 (honest role for authz tests) | A4 start |
| 4 | **A4 request-boundary kernel** | 6–9d | A1 (lints to hold it) | A5 start, A10 |
| 5 | **A5 persistence spine** | 8–12d | A1; incremental thereafter | A4 tail, A9 |
| 6 | **A7 ADR 0093 merge** | 10–15d | A4 + A5 substantially landed | — |
| 7 | **A6 seam discipline (final pass)** | 5–8d | A7 (2 cycles + 2 seams dissolve first) | A10 tail |
| — | **A9 async observability** | 4–6d | A1 only | anything |
| — | **A10 frontend cleanup** | 4–6d | A1 (gates on) | any backend axis |

**Highest-leverage first axis: A1.** Every later axis lands as mechanism instead of discipline only if the gates fire, reproducibly, before the fix arrives; it is also the prerequisite for the in-loop agent gate, which is the half of this program that outlives the refactor. A3 runs in the same window because it is small, independent, and the regulated-domain exposure is live. A8 before A4/A7 honors the standing must-not-be-shelved ruling and prevents migrating authz twice. A7 after A4/A5 is a pure sequencing-cost argument (admissible for sequencing): the spine makes the merge mechanical. Total: roughly 50–70 working days serialized; ~35–45 with the marked parallelism.

## 6. Inversion tests (one line per structural conclusion)

- **A3 (least-privilege role):** if the app already connected as a locked-down role, would I argue for superuser? No — Part 11/ISO 13485 tamper evidence requires the DB to be able to refuse its own client; the conclusion rests on the standard, not the current wiring.
- **A4 (one platform handler kit):** if 13 modules today shared one kit, would I argue for per-module plumbing? No — per-module divergence has a named failure mode (wire-contract drift, inconsistent correlation) with no compensating benefit; mature frameworks (any of the last 20 years) own this layer.
- **A5 (generated query layer + tx chokepoint):** if the repo were sqlc-everywhere, would I argue for hand SQL? No — hand SQL's failure mode (scan/SELECT mismatch surfacing at runtime) is the exact class a regulated system pays most for; mature Go shops generate or check queries.
- **A6 (consumer-owned ports):** if every seam were consumer-owned today, would I argue for producer-owned? No — DIP's failure mode (producer change ripples through N consumers with no insulating seam) is independent of which shape the code currently has.
- **A7 (one document context):** ruled in the prior round under status-quo-inadmissible evidence — the surviving argument is the domain one (a template is a document lifecycle role, not a peer aggregate); implementation order here is sequencing only.
- **A8 (single grant source):** if grants were unified today, would I argue for dual-source? No — two sources of one authorization truth is a named failure mode (silent disagreement = silent privilege drift) regardless of which source the code currently reads.
- **A2 (pre-merge integration signal):** if the full suite ran pre-merge today, would I argue to move it post-merge? Only the cost argument survives, and it argues for *selective* pre-merge, not zero — which is the conclusion.
- **A9 (instrument the async fleet):** if worker/jobs were instrumented, would I argue to remove it? No — "every long-running service exposes liveness and work-queue depth" is a decades-old ops baseline.
- **A1/A10 (gates must fire):** a control that does not fire is absent by the brief's own binding rule; the inversion is vacuous — no state of the code argues for inert controls.

## 7. Where I disagree with a lane

- **http-delivery HTTP-10 (18 hand-written frontend DTO files):** the frontend lane inspected the same surface and found most are aliases of `components['schemas']` or legitimate runtime models. HTTP-10 is mis-sized; the real finding is narrower — the `openapi-fetch` client has only 4 consumers. A10 carries the narrow version.
- **persistence PERSIST-03 (DoReadOnly dead, framed as gap):** memory and the lane's own note confirm G1 deliberately collapsed DoReadOnly→Do with an api-lint guard. Not a gap — a deletion task. Moved to A5 as cleanup, severity floor.
- **security SEC-02 sizing:** "no tier-2 in 4 modules" is real but tier-1 remains boot-fatal and generated; it is one-tier-instead-of-two, not unprotected. It belongs inside A8's scope rather than as standalone hazard remediation — treating it separately would migrate those modules onto a grant model A8 is about to replace.
- **observability OBS-05 (dual metrics = drift):** partially disagree with the implicit remedy weight — the JSON endpoint may be a consumed product surface (frontend dashboard). The fix is convergence (Prometheus as truth, JSON as adapter), not deletion; the drift finding stands, the "two systems forever" framing overstates it.
- **cicd §1 blocking labels:** the lane itself flags this, and I underline it — every "blocking" classification in the program is inferred from exit-code shape; with branch protection unverifiable (403), *nothing* in this repo is provably required for merge. The methodology in §4.4 is the answer; until it lands, all blocking claims are provisional.
- **duplication DUP-03 remedy direction:** the lane sized 12 hand-written tenant_data_ports as duplication; the cheaper global-maximum is not deduplication into a helper but declaration — modules declare table lists as data, one platform implementation ranges over them, and the same registry then powers A6's SQL-ownership lint. One structure serves two axes; I place it in A6, not A5.
