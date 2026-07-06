# M10 Terminal-Remediation — Milestone-Validator Verdict (C1–C7)

> **Program:** global-maximum-remediation · **Milestone:** milestone-10-terminal-remediation (HS-5)
> **Judged by:** independent `milestone-validator` subagent (fresh session, separation of powers).
> **Date:** 2026-07-06 · **Aggregate diff:** `157a81bb^..0f6333cf` (3 commits, HEAD).
> Binding contract: `validation-contract.md` (D4) + `milestone.md` + `references/milestone-end-validation.md`.
> The validator judges and writes this file only — it edited no source, flipped no status.

## Inputs loaded (all present, all readable)

- `milestone.md`, `validation-contract.md` (D4) — read in full.
- F-R1 / F-R2 / F-R3 `spec.md` + `plan.md` + `evidence.md` (9 files) — read in full.
- Program `README.md` (status table, HS ledger, terminal linkage) + `mission.md §8/§9` provenance.
- Aggregate milestone diff `git diff 157a81bb^..0f6333cf` (name-status + stat + per-file).

No input missing → judgment proceeds (not blind).

---

## C0 — Scope fence (forbidden-list)

**PASS.** `git diff --name-status 157a81bb^..0f6333cf` — every path is inside the D4 allowed set:

- F-R1: `apps/api/cmd/metaldocs-api/{main.go,main_test.go,metrics_endpoint_test.go}`,
  `internal/platform/config/{server.go,server_test.go}`, `deploy/compose/docker-compose.yml`,
  `ops/DEPLOY.md`. ✓
- F-R2: `.github/workflows/governance-check.yml`, `scripts/check-adr-status.sh` (new),
  `wiki/standards/documentation-governance.md`. ✓
- F-R3: `internal/modules/controlleddocuments/domain/sequence_test.go`,
  `internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go`,
  `scripts/check-test-discipline.sh`. ✓
- M10 docs: the milestone's own `milestone.md`, `validation-contract.md`, 9 feature docs, and the two
  **program-level** mission artifacts `qa/mission-validation.md` + `qa/terminal-reaudit-2026-07-06.md`
  (HS-5 provenance — mission-level `qa/`, NOT any prior *milestone's* evidence). ✓

**No** touch to `api/openapi/**`, `**/migrations/**` / `*.sql`, authz capability code
(`iam/**/capabilit*`, tripwire generators, PDP tiers), any M0–M9 `evidence.md`/`spec.md`/`qa/*.md`,
`.env`, `docs/release/**`, or `docs/superpowers/plans/**`. Scope fence holds; no HS-6 drift.

---

## C1 — Per-feature lifecycle conformance

**PASS** for all three features.

| Feature | spec.md | plan.md | evidence.md | Approval line | Consumer-contract-first | Non-goals |
|---------|---------|---------|-------------|---------------|------------------------|-----------|
| F-R1 | ✓ | ✓ execution-shaped (6 files, TDD order, test strategy) | ✓ real cmd output, live drive labeled REAL | APPROVED 2026-07-06 (operator "Go"), dated | ✓ 4-row consumer table (scraper / public client / operator / composition root) | ✓ (no metric-family change, no auth on metrics, no k8s) |
| F-R2 | ✓ | ✓ (3 files, negative→positive→wire→doc order) | ✓ negative + positive proof captured | APPROVED 2026-07-06, dated | ✓ 3-row consumer table (CI / ADR author / governance doc) | ✓ (no rule change, no ADR reformat, no new workflow file) |
| F-R3 | ✓ | ✓ (3 test files + ledger, per-rule fix table) | ✓ gate output + vet + defer ledger | APPROVED 2026-07-06, dated | ✓ 4-row consumer table (gate / maintainer / req-trace / program record) | ✓ (no rule rewrite, no allowlist widening, no drive-by, no REQ build) |

- **Approval provenance:** each spec's approval line is filled and dated 2026-07-06, tracing to the
  operator's up-front "Go" on the M10 remediation recorded in `milestone.md` §HS-5 *before* any
  feature began. The three bounded fixes are mechanical/contract-declared, not guessed — consumer
  contracts are read from the real consumer sites (Prometheus scrape topology, `governance-check.yml`,
  the `check-test-discipline.sh` gate). Fail-closed honored.
- **Commit bundling note (not a defect):** each feature's spec/plan/evidence landed in the *same*
  commit as its code rather than a prior commit. The approval predates the code (operator Go →
  milestone.md → implementation); the bundling is a workflow artifact, and C1 binds on *artifact
  presence + filled dated approval + honored contract*, all of which hold. No guessed contract.
- **`evidence.md` acceptance ↔ `spec.md` Validation Gate:** row-for-row match for each feature
  (verified against F-R1 §Validation Gate 1–6, F-R2 1–4, F-R3 1–5).
- **Deviations carry rationale:** F-R1 live drive run in-container (SASL note); F-R3 worker-goroutine
  `t.Fatalf` faithfulness note — both written in evidence.

---

## C2 — Deterministic gates, re-run from clean state

Every gate re-run by the validator itself (not trusted from transcripts). Actual command + output:

| # | Command | Required | Actual | Result |
|---|---------|----------|--------|--------|
| C2.1 | `go build ./...` | exit 0 | `BUILD_EXIT=0` | **PASS** |
| C2.2 | `bash scripts/check-test-discipline.sh` | exit 0, `test-discipline: clean` | `test-discipline: clean (124 integration test files checked)` `GATE_EXIT=0` | **PASS** |
| C2.3a | `go test ./apps/api/cmd/metaldocs-api/ -run TestMetricsEndpoint -count=1` | PASS | `--- PASS: TestMetricsEndpoint_DedicatedListener_REQF83` `ok ... 7.365s` (log shows public `/metrics`→401) | **PASS** |
| C2.3b | `go test ./internal/platform/config/ -run MetricsAddr -v` | PASS | `--- PASS: TestLoadServerConfig_MetricsAddr` (8 subtests incl. invalid-addr) | **PASS** |
| C2.4 | `go run ./scripts/req-trace` | exit 1, stale=false, uncovered exactly `{REQ-AUTHN-1, REQ-AUTHN-3, REQ-SEARCH-1, REQ-SEC-3}` | `4 MUST uncovered` = that exact set; `stale=false`; `exit status 1` | **PASS** |
| C2.5 | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | 0 violations | `0 violation(s)` `APILINT_EXIT=0` | **PASS** |
| C2.6 | `pwsh -File scripts/check-module-boundaries.ps1` | OK / exit 0 | `[module-boundaries] OK` `MB_EXIT=0` | **PASS** |
| C2.7 | `go vet -tags integration` on F-R3's 3 packages (controlleddocuments/domain, jobs/stuck_instance_watchdog, templates/infrastructure) | exit 0 (compiles, no testdb cycle) | `VET_EXIT=0` | **PASS** |

C2.4 is **unchanged from the M9 baseline** — M10 introduces no new uncovered MUST and closes none
(REQ-SEARCH-1/REQ-SEC-3 remain in the set by design as ratified defers). No flaky/environment-coupled
green: all gates run from clean checkout state.

---

## C3 — F-R1 structural + live proof (Dim 9)

**PASS (structural + regression + live).**

- **Single call site (grep-invariant):** `grep -rn 'PrometheusHandler()' apps/api/cmd/metaldocs-api/`
  → exactly one production call site, `main.go:887 metricsMux.Handle("GET /metrics", httpObs.PrometheusHandler())`
  (the second hit is the mirroring line in `metrics_endpoint_test.go`). It is on `metricsMux`, served
  by a **separate** `http.Server{Addr: serverCfg.MetricsAddr, Handler: Recovery(metricsMux)}`
  (`main.go:886-892`). The public `server.Handler = handler` (the chained API handler, `main.go:855`)
  has **no** `/metrics` route — the old top-level rootMux `/metrics` dispatch is removed, not dangling.
- **Lifecycle:** both listeners start (`main.go:900-908`); `shutdownServer` takes both servers + both
  error channels, selects on either bind error as fail-fast, drains both within the shutdown budget
  (`main.go:910, 928-964`). Config `MetricsAddr` fail-fast validated at boot, mirroring `APP_PORT`
  (`internal/platform/config/server.go:19-60`).
- **Regression test** `TestMetricsEndpoint_DedicatedListener_REQF83` green under C2.3a; its log line
  confirms public `GET /metrics` → `status=401` (fail-closed chain — left the public surface).
- **Live drive (labeled REAL, accepted as recorded runtime evidence per task instruction):** rebuilt
  `api` image, container `health=healthy`; probed from inside the container (exactly how an
  infra-network Prometheus reaches `:9090`, which is not host-published): public `:8081/metrics` →
  **401 Unauthorized**; metrics `:9090/metrics` → **200** `Content-Type: text/plain; version=0.0.4`
  Prometheus exposition with `metaldocs_http_requests_total`, `go_goroutines`, etc. The
  `route="/metrics" 1` series in the scrape is correctly explained as the count of the **public-port
  401 probe** (a real request through the instrumented chain), **not** a self-scrape of `:9090` — the
  metrics mux is not on the `httpObs.Wrap` chain, so scraping it adds no sample. Reasoning verified
  against source (metricsMux = `Recovery(mux)` only). Honest fixture-vs-real labeling.
- **Deploy truth:** `docker-compose.yml` sets `METRICS_ADDR: ":9090"` and host-publishes **only**
  `${APP_PORT}:8081` (no `ports:` for 9090). `ops/DEPLOY.md §132-151` documents the two-listener table,
  scrape target `http://api:9090/metrics`, and the do-not-publish warning.

---

## C4 — F-R2 gate proof (Dim 10)

**PASS (blocking + negative + positive + doc-truth).**

- **Blocking wiring:** `.github/workflows/governance-check.yml` — step `ADR status-field budget gate`
  runs `bash scripts/check-adr-status.sh` in the `check` job; `grep 'continue-on-error'` → the only
  match is a *comment* asserting blocking-ness; no `continue-on-error:` key anywhere. Triggered
  `on: pull_request: branches: [main]`. A non-zero exit fails the job.
- **Negative proof (validator re-ran it):** built a synthetic ADR with a 4-line / 713-char
  `> **Status:**` block in a `mktemp -d` dir → `bash scripts/check-adr-status.sh <tmp>` →
  `::error::ADR status-field budget exceeded ...` `9999-synthetic.md: 4 lines, 713 chars`, `exit=1`.
  Gate fires and names the offender. (Note: the script parses the `> **Status:**` *blockquote* form
  — the real ADR convention — not a `## Status` heading; verified against `wiki/decisions/0001*`.)
- **Positive proof:** `bash scripts/check-adr-status.sh` on real `wiki/decisions/` → `adr-status: clean`,
  exit 0. No false positive on the clean tree.
- **Single source (no split-brain):** one script consumed by both CI and the manual sweep; the
  governance doc's manual command points at the same script — local and CI cannot diverge.
- **Doc-truth:** the old "Wiring this sweep into CI ... is an optional future extension, not required"
  sentence is gone; replaced by a `**CI-enforced (F-R2)**` paragraph
  (`documentation-governance.md:33-36`) naming the blocking step + single-source script. (Remaining
  "optional" hits at lines 16/17/97 refer to an optional date-clause and README stubs — unrelated.)

---

## C5 — F-R3 green + ratification + quality-bar re-measure (Dim 8)

**PASS (root-cause fix, not suppression).**

- **Aggregate green:** C2.2 `test-discipline: clean`, exit 0. All 4 violations resolved at root:
  - **R2** — allowlist path corrected `templates/repository/…` → `templates/infrastructure/…`
    (F9.5-rename reconciliation). Full non-comment diff of the allowlist = exactly this one line pair
    (old removed, corrected added). Allowlist did **not** widen — it is path-corrected, per the D4
    "shrink or path-correct only" rule.
  - **R4** — `job_integration_test.go:186` now `FROM metaldocs.documents` (schema-qualified).
  - **R1** — `sequence_test.go:60,121` both inline `set_config('metaldocs.asserted_caps',…)` sites
    replaced by the sanctioned `testdb.SetCapsOnTx(t, tx, …)`; no inline `asserted_caps` write remains.
    Same tripwire assertion, sanctioned primitive — behavior-preserving (happy path byte-equivalent;
    the `t.Fatalf`-on-unreachable-failure faithfulness delta is documented and assertion-strength
    preserving).
- **Compile proof:** C2.7 `go vet -tags integration` on all 3 packages exit 0 — the `domain_test`
  external package importing `testdb` (which imports `controlleddocuments/domain`) does **not** cycle.
- **Defer ratification (honest defers):** the F-R3 evidence ledger records REQ-SEARCH-1 (search
  reindex procedure) and REQ-SEC-3 (OWASP ASVS) each with `{finding, why-absent, trigger, owner}`;
  both are absent **product features**, not hygiene; both remain in the C2.4 uncovered set by design
  (verified: the set is exactly the 4 expected). Triggers are external events, not time-decay.
- **Quality-bar re-measure:** the Dim-8 bar ("gate GREEN at HEAD from clean state") is met by
  independent re-run, and the mission-introduced stale-allowlist root cause (M9 F9.5 rename) is fixed
  at source — not symptom-patched by whitelisting the failing file.

---

## C6 — Regression against M0–M9 + forbidden-list

**PASS.** Forbidden-list swept, no hit:

- Suite-green-as-pass: **not present** — each feature's acceptance is mapped to a specific re-run
  command + output above; no bare "all green."
- Fixture-as-real: **not present** — F-R1 live drive explicitly labeled REAL (container runtime), and
  the `route="/metrics"` counter is honestly attributed to the public-port probe, not a self-scrape.
- Guessed contract: **not present** — consumer contracts read from real consumer sites (fail-closed).
- Split-brain: **not present** — ADR sweep single-sourced in one script (CI + manual); metrics config
  single-sourced in `ServerConfig`.
- Self-judged close: **not present** — this validator judged only; wrote this file only; flipped no
  status.
- Scope drift: **not present** — C0 fence holds; no REQ built, no rule rewritten.
- Symptom-patch: **not present** — R2 fixed at the rename root, not by whitelisting.
- **Regression M0–M9:** the 4 already-CONFIRMED deterministic gates hold — C2.1 `go build` exit 0,
  C2.4 `req-trace` unchanged 4-set/stale=false, C2.5 `api-lint` 0 violations, C2.6 module-boundaries
  OK. No prior-milestone guarantee regressed; aggregate diff confined to C0 allowed set + M10 docs.

**Senior-review of the aggregate diff (C3-class judgment):** three cleanly-separated fixes, no
duplicated logic, no dead code (old rootMux `/metrics` dispatch removed), no contract defined two
ways, no feature breaking another. A staff engineer would approve this diff. Every fix installs or
repairs a **gate/boundary** (process topology / blocking CI gate / green gate + corrected allowlist),
consistent with the mission thesis — none is a mere convention.

**Retrospective ("could it be built better"):** the F-R1 metrics listener is intentionally
unauthenticated (isolation by network, per spec non-goal) — a future post-v1 hardening could add mTLS
/ ServiceMonitor for k8s, already named as a defer; not an unsoundness. No materially better
construction is warranted within M10's half-session appetite. Nothing here FAILs the milestone.

---

## C7 — Verdict

Every C-clause (C0–C6) satisfied with real commands and real outputs; both dimensions
(code-wise senior-clean; function-wise end-to-end proven, incl. live runtime for F-R1) pass. The
three DEBT dimensions are closed **structurally**: Dim 9 by process topology, Dim 10 by a blocking CI
gate, Dim 8 by a green gate + root-cause allowlist reconciliation with ratified bounded defers.

No fix feature required.

**VERDICT: PASS**

> Separation of powers: this verdict flips no status. On operator HS-1 approval, the main session may
> flip M10 status and re-dispatch `mission-validator` for the second terminal-acceptance run against
> `mission.md §8`. M10 does not itself declare the mission done.
