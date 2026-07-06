# M10 Validation Contract (D4)

> **Binding.** Authored 2026-07-06, **before any feature implementation.** The `milestone-validator`
> judges M10 against *this* contract + `milestone.md`. Every clause is executable by a fresh agent
> from a clean checkout. Separation of powers: the validator runs these, writes the verdict, edits
> nothing.

## C0 — Scope fence (forbidden-list)

The aggregate M10 diff **must not** touch any of:
- `api/openapi/**` (contract truth — untouched)
- `**/migrations/**` or any `*.sql` schema file (DB truth — untouched)
- authz capability code: `internal/modules/iam/**/capabilit*`, tripwire arm generators, PDP tiers
- any prior milestone's `evidence.md` / `spec.md` / `qa/*.md` (M0–M9 sealed)
- `.env`, `docs/release/**`, `docs/superpowers/plans/**` (never committed)

Allowed surfaces (exhaustive):
- F-R1: `apps/api/cmd/metaldocs-api/{main.go,metrics_endpoint_test.go}`,
  `internal/platform/config/server.go`, `deploy/**` compose, `ops/DEPLOY.md`
- F-R2: `.github/workflows/governance-check.yml`, `scripts/**` (new sweep script only),
  `wiki/standards/documentation-governance.md`
- F-R3: the 3 named test files under `internal/modules/**`,
  `scripts/check-test-discipline.sh`, M10 evidence/defer ledger docs

Any diff outside the allowed set → **FAIL** (HS-6 scope drift).

## C1 — Per-feature lifecycle conformance

For **each** of F-R1, F-R2, F-R3:
- `spec.md` exists, is consumer-contract-first, and its **approval line is filled and dated before**
  the feature's first code commit (git order check).
- `plan.md` exists.
- `evidence.md` exists with **real command output** (not adjectives), fixture-vs-real labeled.

Missing/empty/out-of-order → FAIL that feature.

## C2 — Deterministic gates (re-run from clean state, all must hold)

| # | Command | Required result |
|---|---------|-----------------|
| C2.1 | `go build ./...` | exit 0 |
| C2.2 | `bash scripts/check-test-discipline.sh` | exit 0, prints `test-discipline: clean` |
| C2.3 | `go test ./apps/api/cmd/metaldocs-api/ -run TestMetricsEndpoint -count=1` | PASS |
| C2.4 | `go run ./scripts/req-trace` | `stale=false`, exit 1, uncovered set **exactly** `{REQ-AUTHN-1, REQ-AUTHN-3, REQ-SEARCH-1, REQ-SEC-3}` |
| C2.5 | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | 0 violations |
| C2.6 | `powershell -File scripts/check-module-boundaries.ps1` | OK / exit 0 |

C2.4 must be **unchanged** from the M9 baseline — M10 introduces no new uncovered MUST and closes none
(REQ-SEARCH-1/REQ-SEC-3 are ratified as defers, not implemented).

## C3 — F-R1 structural + live proof (Dim 9)

- **Structural:** the public `http.Server` handler is composed **without** any `/metrics` route; the
  only registration of `PrometheusHandler()` is on a **separate** `http.Server` bound to `METRICS_ADDR`
  (grep proof: exactly one `PrometheusHandler()` call site, and it is not on the public mux).
- **Regression test** `TestMetricsEndpoint_*` asserts: public composed handler → `GET /metrics` is
  **not 200** (401 or 404); dedicated metrics handler → `GET /metrics` = 200, `Content-Type`
  `text/plain`, body contains the three `metaldocs_http_*` families **and** a `go_`/`process_` line;
  unauthenticated on both. Test green under C2.3.
- **Live drive (labeled real):** API started via `.\scripts\start-api.ps1`; captured `curl` of the
  **public** port `/metrics` = not an unauth scrape (401/404); `curl` of the **metrics** port
  `/metrics` = 200 Prometheus. Both transcripts in `f-r1-*/evidence.md`.
- **Deploy truth:** `deploy/**` compose no longer host-publishes the metrics port on the public
  mapping; `ops/DEPLOY.md` documents the split (public vs metrics listener, scrape target).

## C4 — F-R2 gate proof (Dim 10)

- `.github/workflows/governance-check.yml` gains an ADR-status sweep step that is **blocking**
  (no `continue-on-error: true`, runs on `pull_request`).
- **Negative proof (captured):** with a synthetic ADR whose status block exceeds 3 lines / 400 chars,
  the sweep command **exits non-zero and names the file**. Transcript in `f-r2-*/evidence.md`.
- **Positive proof (captured):** on the clean current tree, the sweep exits 0 with no output.
- `documentation-governance.md` updated — the "optional future extension, not required" language is
  replaced by a statement that CI now enforces the rule (grep proof the old sentence is gone).

## C5 — F-R3 green + ratification (Dim 8)

- C2.2 green (the aggregate proof).
- Each of the 3 repaired test files: (a) still compiles under `go vet` with the integration build tag
  for its package; (b) repair preserves behavior — same assertions, sanctioned primitive substituted
  (evidence shows before/after of the specific line, not a wholesale rewrite).
- The R2 allowlist correction is documented as the **F9.5 rename reconciliation**
  (`templates/repository/` → `templates/infrastructure/`), and `check-test-discipline.sh`'s allowlists
  did not **grow** any entry except this path correction (allowlists may only shrink or be
  path-corrected, never widened).
- **Defer ledger:** REQ-SEARCH-1 and REQ-SEC-3 each recorded with `{finding, why-absent, trigger,
  owner}`; both remain in the C2.4 uncovered set by design.

## C6 — Regression against M0–M9

- No forbidden-list surface (C0) touched.
- The 4 already-CONFIRMED deterministic gates (C2.1, C2.4, C2.5, C2.6) hold — no prior-milestone
  guarantee regressed.
- `git diff --stat main...HEAD` for M10 commits confined to the C0 allowed set + M10 docs.

## C7 — Verdict

`milestone-validator` writes `qa/milestone-qa.md` with a per-clause C1–C6 table and a single
`VERDICT: PASS|FAIL`. PASS requires **every** C-clause satisfied. On FAIL, the validator names the
exact failing clause and the minimum fix feature (HS-4). The validator never edits code and never
flips status.
