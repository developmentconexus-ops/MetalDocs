# Engineering remediation program — reconciled synthesis

**Date:** 2026-08-07 · **Status:** proposed, not started · **Evidence:** `docs/superpowers/analysis/inventory/`

Ten discovery lanes produced a sized, `file:line`-cited inventory (~900 lines, 10 reports). Two
independent advisory arms then synthesised it under a brief that ruled status-quo evidence
inadmissible for any structural conclusion and required an inversion test per conclusion
(`inventory/_synthesis-brief.md`, method rule ME-13). Their answers are
`inventory/_synthesis-claude.md` (10 axes) and `inventory/_synthesis-sol.md` (9 axes).

This document reconciles them. Where they diverged, the resolution and its reason are recorded.

## The nine axes

Named for cause, not symptom. A finding belongs to exactly one axis.

| # | Axis | Root cause | Scale |
|---|---|---|---|
| **A1** | The verifier is not one trusted product | Checks accumulated as workflows, scripts, allowlists and prose without one executable manifest, proven negative fixtures, a reproducible toolchain, or authority separate from the author | 20 workflows; 8–13 check-shaped scripts wired to zero workflows; 548/553 integration test functions (342 files, 43,691 LOC) absent from the PR gate; 13 shell/PS guards with no bad-fixture harness; 4 `@latest` CLI families; `only-new-issues: true`; `govulncheck` default-skipped in both paths that reference it |
| **A2** | No ratcheted whole-repo quality baseline | Feature-local delivery added hand-written variants faster than shared primitives and static rules could make the old ones extinct | 42.9k TS LOC with zero general ESLint quality rules; 10 of 15 Go linters off, no `formatters:` block; 4 components over the repo's own 400-line ceiling (max 738); 13 dead `src/` files + 48 unused exports; 53 hand-written clone pairs |
| **A3** | The API contract stops before runtime behaviour | OpenAPI + codegen define routes and shapes, but error emission, validation, identity, pagination and idempotency remain parallel hand-authored dialects | 521 error-emission calls through 6 mechanisms, 12+ local `writeProblem` definitions; 123 OpenAPI constraints with zero request-validator wiring; 11 anonymous request structs; 17 callers of the fail-open actor helper; 4 pagination policies |
| **A4** | Module seams expose implementations, not capabilities | Cross-context collaboration runs through foreign domain types, sentinels, SQL and platform→module imports instead of consumer-owned ports with composition-root adapters | 9/15 domain packages import `database/sql`/`platform/db`; 62 cross-module foreign-sentinel checks; 17+ foreign-table SQL reads; 20 platform→module edges + 11 more; `approval/application` 8,755 LOC / 583 exports |
| **A5** | Persistence correctness is maintained by visual agreement | Transaction lifecycle, SQL text, column order, scans and driver error mapping are repeated by hand instead of represented once in typed machinery | 82 direct `BeginTx` sites / 25 files; 75 constructors take `*sql.DB` vs 6 taking `TxRunner`; 20 Tx/non-Tx twins; 242 scan sites vs 16 helpers; 15 `23505` checks incl. unreachable `lib/pq` branches; 1 confirmed N+1 |
| **A6** | Security properties are configuration-contingent, not fail-closed | Strong controls hold only if an operator picks the right role and configuration, and startup never proves the precondition | 37 FORCE-RLS tables + the audit `REVOKE` both depend on a non-superuser connection; the only compose in the repo passes `${POSTGRES_USER}` to all three services; **zero** boot assertions on `rolsuper`/`rolbypassrls`; no CSP/HSTS/frame/type headers anywhere; KEK absence is a valid no-op; one re-auth timing divergence |
| **A7** | Async operations have no closed feedback loop | The API was instrumented as the observable centre; workers, jobs, propagation, metrics, alerts and deploy artifacts were never designed as one runtime | 2/4 binaries have no health/metrics listener and no compose healthcheck; 0 async spans; tracing in 2/15 modules; 170/215 log calls non-context; 2 parallel metric systems; 0 alert rules; 2 `:latest` images |
| **A8** | ADR 0092 — grant model is dual-source | Tier 1 and tier 2 answer authorization from disjoint relations, so route admission and in-transaction enforcement can disagree by construction | 2 grant sources; 4/15 modules have no tier-2 `authz.Require` and no tripwire coverage; role vocabulary on 7 declaration surfaces |
| **A9** | ADR 0093 — Controlled Information is three peer contexts | One regulated lineage/revision lifecycle and one draft-integrity question were split by artifact kind | 3 modules, ~55k Go LOC; 17+ foreign-table reads; target is 2 aggregates + `NumberSeries` + `TemplateUsePolicy` |

### Divergences resolved

- **Inert controls vs post-merge test topology** — one arm split these; they share one root cause
  (a check that does not fire before handoff is absent) and one fix (the manifest). **Merged into A1.**
- **Frontend as its own axis** — the frontend findings are the same mechanism as the Go lint findings:
  a gate that was never switched on. **Folded into A2.**
- **`SEC-03` sizing** — the claim "production connects as superuser" is not supported. The single
  compose in the repo is self-labelled dev/test (`deploy/compose/docker-compose.yml:10-14`,
  "NEVER copy this flag to a production deployment"). What *is* proven: no deployment artifact in the
  repo provisions a least-privilege role, and no binary asserts its own connection identity at boot
  (`grep` for `rolsuper|rolbypassrls|pg_roles` across `internal/` and `apps/` returns nothing).
  A6 is sized to remove the uncertainty, not to remediate a confirmed production breach.
- **`PERSIST-10`** — "the RLS/BYPASSRLS trap is closed" is true for the integration suite via the
  non-owner `metaldocs_ci` role and false for the application connection. Two different roles.

## Sequence

A1 first is not a preference. Every later axis lands as **mechanism** if its gate fires before
handoff and as **discipline** if it does not, and discipline regresses.

| Order | Axis | Size | Dependency |
|---|---|---:|---|
| 1 | **A1** verification spine | 6–9d | First. Manifest with `fast`/`changed`/`pr`/`full` profiles, negative-fixture standard, pinned tools, evidence bundle, guardian + merge command. Do not block on wiring every slow release check. |
| 2 | **A8** grant unification | 10–15d | After the minimal A1 spine. **Must precede A9 and must not be shelved behind it** — held by both arms across all rounds. |
| 2a | **A6** fail-closed security | 4–7d | Parallel with A8. Bounded and urgent; coordinate only where A8 touches tripwires. |
| 3 | **A3** contract-to-runtime | 10–15d | After A1; overlaps late A8 except on IAM routes. Makes later module moves cheaper. |
| 3a | **A2** quality ratchet | 8–15d then ongoing | Parallel with A3. Land regression-only rules first (3–5d). Do not block A8 on historic lint debt. |
| 4 | **A5** persistence mechanics | 15–25d | Needs A1; benefits from A3's settled conventions. `sqlc` spike (2d) before committing. |
| 5 | **A4** executable seams | 12–20d | After A8 (IAM vocabulary) and A3 (contract patterns). Land graph + SQL-ownership gates before moving seams. |
| 6 | **A9** Controlled Information | 20–30d | After A8; preferably after A3/A4 guards. Migration cost affects this order only — never the target. |
| 7 | **A7** async operational loop | 8–12d | Health/metrics can start right after A1 in parallel; trace propagation follows A3/A5 to avoid rework. |

## Gate topology

The firing-mechanism hierarchy is binding: **unrepresentable → boot-fatal → red build → runtime
assertion → discipline.** Discipline is never the only guard for an enumerable or syntactic property.

| Level | Blocking | Annotating |
|---|---|---|
| Unrepresentable | Generated API/role types; FK/check/unique constraints; one authz evaluator; typed query outputs; one transaction runner API | — |
| Boot-fatal | DB role non-superuser + NOBYPASSRLS; required secrets/KEK by profile; API surface parity; capability registry parity | Optional-dependency degradation where product behaviour permits |
| Red build | Formatting; compile/typecheck; generated drift; boundary guards; migration invariants; OpenAPI lint + breaking; unit tests; affected integration shards; full integration + race before merge for DB/platform/security changes; gitleaks; gosec; govulncheck; High/Critical Grype; negative-fixture suites | Duplication, complexity, coverage trend, knip, medium/low CVEs, performance until budgets exist |
| Runtime assertion | Request-schema validation; authorization denial; optimistic concurrency where declared; idempotency state machine; DB tripwire | SLO warnings before an error budget is ratified |
| Discipline | Only judgments that cannot be mechanised: role-bundle correctness, ADR boundary acceptance, one-time migration review | Architecture trigger review |

A blocking check must be deterministic, locally reproducible, and carry a demonstrated bad fixture.
Otherwise it teaches bypassing.

## The in-loop agent gate

Post-hoc PR review assumes a human author caught by a machine. Here the author **is** a machine whose
failure mode is fluent, internally consistent wrongness — the wrong noun used correctly everywhere,
including inside the guard written to catch it. Comments do not correct that; only making the defect
unwritable does.

1. `verify preflight --changed-from <base>` derives the affected modules, contracts and DB objects
   from the diff graph. The agent does not choose its own lane.
2. Before implementing, the agent records the observable invariants and at least one falsifying
   counterexample, in product/standard nouns rather than current type names.
3. After each coherent patch, `verify changed` runs formatting, generators, typecheck, semantic
   guards, focused unit tests and the smallest affected integration shard — under ~5 minutes.
4. A context-isolated reviewer receives the brief, diff, canonical contract and counterexamples —
   **not** the author's rationale first — and tries to falsify noun choice, ownership, authorization,
   tenant isolation and migration assumptions.
5. `verify pr` emits a content-addressed evidence bundle: source SHA, policy SHA, tool versions,
   commands, exit codes, artifacts. No prose "green" substitutes for it.

Every semantic guard declares the noun it checks, the fixtures that must fail, and its explicit
non-claims. String greps are acceptable only for literal extinction; architectural rules use Go
AST/package graphs, parsed OpenAPI, schema queries or property tests. An agent may add a product test
in its own change but may not make that self-authored test the sole evidence for a security,
boundary, or migration invariant.

## Separation of powers

The repository was made public on 2026-08-07, so branch protection and rulesets are now available at
no cost (both APIs previously returned `403 "Upgrade to GitHub Pro"`; they now return
`404 Branch not protected` and `[]`). Platform enforcement and the guardian design are complementary,
not alternatives:

- **Platform half:** a ruleset on `main` — no direct push, PR required, required status checks.
- **Custody half:** agents work without the `main` push credential; only the operator merges.
- **Guardian half:** the verification manifest, gate runner, protected-path policy and merge command
  live in a separate `metaldocs-guardian` repository to which agent credentials are read-only.
  Product CI calls a pinned guardian commit.
- **Mixed-change ban:** a product change may not modify `.github/workflows/**`, lint configs, guard
  sources, allowlists/baselines, or the guardian pin. Policy changes ride alone, must add or update
  negative fixtures, and run against both the old and new policy.
- **Ratchets:** every allowlist/baseline carries a committed count; growth requires an explicit
  approved trailer, making it a logged decision rather than a silent edit.

This does not claim a solo operator cannot subvert their own process. It prevents the machine author
from silently weakening its own judge inside the same change.

## Do-not-touch

Consolidated from every lane's "actually fine" section. Churn here is a defect.

- Generated `api.gen.go` duplication — generation is the anti-drift mechanism, not duplication.
- The 11-link middleware chain and its `assertSurface` boot checks.
- `approval`'s post-ADR-0089 error registry and `dictionary_reader_adapter.go` — these are the
  **targets** A3 and A4 generalise from, not defects.
- `pagination.ClampLimit`/cursor primitives; `tenant.FromContext`; `templates`' contract-declared
  50/200 pagination exemption.
- Explicit `main.go` composition — legible; introduce no DI container.
- Zero intra-module layer inversions, zero package import cycles, zero frontend import cycles, no
  literal SQL in domain packages, no `util`/`helpers`/`common` packages, `go.mod` with no `replace`.
- Parameter binding and tenant-predicate discipline (83/87 SQL files); the single outbox INSERT site;
  no network calls inside a DB transaction.
- DB constraints/triggers/RLS design including app pre-check + DB backstop; migration gapless and
  no-historical-edit CI checks; the CI-only non-owner RLS test role.
- `tools/cilint` and `scripts/api-lint` with positive/negative fixtures — extend their standard.
- Login/session core (bcrypt 12, constant-time dummy hash, hashed tokens at rest), secure cookie
  defaults, CORS validation, trusted-proxy default deny, typed `platform/config`, gitleaks, gosec,
  the Grype gate, the audit hash chain inside an immutable DB function.
- `slog` JSON as the single logger; the isolated metrics listener; the API's real readiness probe.
- Frontend: query-key discipline (103 `QK.*` vs 1 inline), the two intentional Zustand stores, clean
  `tsc --noEmit`, madge-clean import graph.

## Dropped as noise

Real findings that do not earn a program slot. Reason attached to each.

- **439 bare `fmt.Errorf`** — an upper bound on *candidates*, not broken chains; absent `%w` is often
  a new error being formatted. Enforce forward with AST-aware `wrapcheck`; no count-driven rewrite.
- **110 panic sites** — wiring-time and exhaustive-state panics can be correct. Triage the
  request-reachable ones only.
- **2,005 raw `px` / 82 `rgba()`** — tokens are not intrinsically correct for every dimension. Only
  literals duplicating a declared semantic token are drift.
- **11 missing feature barrels** — a documentation convention; barrels do not cause the 66
  cross-feature imports. Enforce public boundaries directly and amend the rule.
- **Universal `If-Match`** — not a style rule; required only where stale-write risk is observable.
- **21 knip-flagged deps, 30 `t.Skip`, zip-bomb depth, `auth_failure_counters` tenant scope** —
  unverified leads. Confirm before sizing; none is a workstream today.
- **1 generic in 80k LOC, 1 interface-returning constructor, 2 duplicate FE formatters, the confirmed
  N+1** — fix opportunistically in passing.
- **Action SHA pinning** — do it inside A1's one pinning pass, not as a separate workstream.
- **Empty `tests/e2e`** — Playwright E2E exists; delete the placeholder.
- **`DUP-09` dual signoff check** — intended defence in depth, matching the stated doctrine.
- **Lane count discrepancies** (359 vs 342 integration files) — measurement-method difference.

## Provenance

Lane reports: `docs/superpowers/analysis/inventory/*.md`. Synthesis arms:
`_synthesis-claude.md`, `_synthesis-sol.md`. Method: `_synthesis-brief.md`, ME-13 in
`docs/engineering/mechanical-enforcement-register.md`. Prior rulings: ADR 0092 (pending), ADR 0093.
