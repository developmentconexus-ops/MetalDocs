# Pass 13 — Architecture Guard Matrix

**Date:** 2026-08-09
**Baseline:** main @ 418070bf
**Status:** reproduced-current (every claim below verified by direct file read / grep / registry read against this baseline in the `arch-audit` worktree)
**Scope:** every mechanical guard named in the audit brief, cross-checked against `tools/verify/registry.go`, `.github/workflows/{ci,smoke,nightly,docx-renderer,release}.yml`, and each guard's own source + fixtures.

---

## 0. CI workflow map (ground truth for "runs on PR?")

| Workflow | Trigger | Jobs | Blocks PR merge? |
|---|---|---|---|
| `ci.yml` | `pull_request` → `main` only, no path filter (deliberate — a path filter would leave a required status permanently pending) | `verify` (`tools/verify --require-infra --profile=changed`), `test-integration` (needs `verify`; `--only=go-test-integration --changed`), `security` (gitleaks + grype/anchore), `lint-go` (bare `golangci-lint-action`, NOT a `tools/verify` check), `required` (`always()`, jq-validates all 4 deps == `"success"`) | **Yes** — `required` is the branch-protection gate |
| `nightly.yml` | `schedule 05:00` + `workflow_dispatch` | `perf` (KNOWN RED, no DB secret), `e2e` (Playwright+axe), `axe` (baseline hygiene), `stress` (`-race -count=1`, N=500), `security-scan` (`tools/verify --require-infra --only=gosec,govulncheck`) | **No** — file's own header says "Nothing here blocks a PR." |
| `smoke.yml` | `workflow_dispatch` only | `ops/smoke/healthz.sh` against operator-supplied URL | No — manual, no staging env yet |
| `docx-renderer.yml` | `pull_request` + `push:main`, no path filter | `tools/verify --only=docx-typecheck,docx-build,docx-test` | Redundant belt duplicating `ci.yml:verify`'s coverage of the same 3 checks (documented as such) |
| `release.yml` | `push tags:v*` | SBOM generation | N/A — unrelated to architecture guards |

**Key finding: `--profile=full` and `--profile=pr` are never invoked wholesale by any workflow.** Only `--profile=changed` (in `ci.yml:verify`) and explicit `--only=X` selections (`test-integration`, `nightly:security-scan`, `docx-renderer.yml`) run in CI. Checks whose only `Profiles` entry is `full` (`gosec`, `govulncheck`, `go-test-integration` with race) only run via the targeted `--only=` selections listed above, never via a full-profile sweep — meaning **gosec and govulncheck cannot currently block a PR merge** (`CIJob: nightly.yml:security-scan`, outside `ci.yml:required`'s dependency closure). This is self-documented in `tools/verify`'s own `--audit` mechanism (rule A6, below) as a known, transitional gap.

---

## 1. `scripts/check-module-boundaries.ps1` (registry ID `module-imports`)

- **Intends to prove:** REQ-TOP-1 — cross-module Go imports touch only a module's published surface (`domain`/`application`/`api` layers, or an explicit `$publishedPackages` allowlist: `authz`, `fanout`, `fanout/dispatchjobs`, `resolvers`).
- **Actually proves:** that no `.go` file under `internal/modules/**` (excluding `_test.go`) contains a raw string literal `"metaldocs/internal/modules/<other>/<layer>"` where `<layer>` is anything other than `domain`/`application`/`api`/an allowlisted package — a text-regex scan over raw file content, not an AST import-block parse.
- **Blind spots:**
  - Regex-over-raw-text, not AST: a match inside a comment or unrelated string literal could false-positive/negative (unverified in practice, but structurally possible).
  - Scope is `internal/modules/**` only — code in `apps/api/`, `cmd/`, `internal/platform/` importing a module's forbidden layer (e.g. composition-root wiring reaching into a repository package) is entirely invisible to this guard.
  - `$publishedPackages` can be widened by any PR with no ADR-citation requirement enforced mechanically.
- **Wired into `tools/verify`:** yes — `registry.go`, ID `module-imports`, Profiles `fast,pr,full`.
- **Runs on PR:** yes, via `ci.yml:verify --profile=changed` (declared `Paths` include `internal/` and the script itself).
- **Positive fixture:** none found.
- **Negative fixture:** none found — no Pester test, no `_test.go`, no `scripts/testdata/module-boundaries/` fixture directory anywhere in the tree.
- **Bypassable:** yes — add a new entry to `$publishedPackages` (no gate on justification), or route the import through `apps/api`/`cmd`/`internal/platform` instead of another module.
- **Manual allowlist:** `$debtAllowList`, currently **empty**; `$publishedPackages` (4 entries) functions as a permanent structural allowlist.
- **Single source of truth:** yes for REQ-TOP-1 specifically (Go-import-path level). A related but distinct invariant — raw-SQL cross-module base-table reads — is enforced separately by `tools/cilint`'s `HGCrossModule` analyzer (§7); the two do not overlap in what they catch and are complementary, not duplicative.

---

## 2. `scripts/api-lint` (Go tool, `scripts/api-lint/*.go`, ~6,572 LOC / 24 files)

- **Intends to prove:** OpenAPI-spec ↔ Go-handler ↔ wire-shape ↔ tenant-isolation-adjacent contract consistency (route base-path prefixing, response envelope shape, authz-drift, pagination shape, casing, nullable/required parity, tenant-seed chokepoints, tripwire-arm parity/drift — see catalog below).
- **Actually proves:** exactly what each of its ~15 named rules checks (`PATH-BASE-PREFIX`, `ENVELOPE-DRIFT`, `AUTHZ-DRIFT`, `PAGINATION-DRIFT`, `CASING-DRIFT`, `SHAPE-NULLABLE-NOT-REQUIRED`, `WIRE-NULLABLE-OMITEMPTY`, `PROBLEM-DUMP-IMPORT`, `SEED-CHOKEPOINT`(+`-ALLOWLIST-STALE`), `ASYNC-TENANT-SEED`(+`-ALLOWLIST-STALE`), `SOLE-RLS-ASYNC-READ`(+`-ALLOWLIST-STALE`), `TRIPWIRE-ARM-PARITY`, `TRIPWIRE-ARM-DRIFT`). Explicit design statement in `main.go`: "EVERY rule this linter emits is BLOCKING… a NEW rule is blocking by construction" — no reported-only tier exists.
- **Blind spots:** rule-specific (see §6 for tripwire detail); the `-only` flag is a **post-hoc filter on findings from all rules**, not a mode selector — running `-only X` still executes every rule internally.
- **Wired into `tools/verify`:** yes, as 4 distinct checks — `api-lint-e2e-base-path`, `api-lint` (`-strict`, full run), `api-lint-selftest` (`go test ./scripts/api-lint/...`). All `ProfilePR+Full`, `CIJob: ci.yml:verify`.
- **Runs on PR:** yes.
- **Positive/negative fixtures:** extensive — every rule file has a matching `_test.go`, plus `testdata/repo_good/` and `testdata/repo_missing/` fixture trees and `e2e_test.go`/`exit_code_test.go`. **This is the most test-disciplined guard mechanism in the repo.**
- **Bypassable:** the 3 `*-ALLOWLIST-STALE` rules are the best-engineered waiver mechanism found in the whole audit — a hand-coded Go allowlist (`seed_chokepoint_rule.go`, `async_tenant_seed_rule.go`, `sole_rls_read_rule.go`) paired with a companion rule that **fails the build if the allowlist references a file/symbol that no longer exists**, i.e. the waiver cannot silently rot. Contrast with `check-governance-waivers.txt` (§10) and `tools/cilint/baseline.json` (§3), neither of which self-audits its own staleness.
- **Manual allowlist:** yes (the 3 `*-ALLOWLIST-STALE`-guarded lists above), but self-auditing as described.
- **Single source of truth:** yes — `api/openapi/v1/openapi.yaml` is the declared route truth; this tool is the sole mechanical enforcer of spec↔code parity for the rules it owns.

---

## 3. `tools/cilint` (custom Go AST analyzers + baseline ratchet)

- **Intends to prove (per registry.go's `Desc` field):** "hgcrossmodule, nosqltxindomain, platformboundary, txownership, legacyvocab" — 5 analyzers.
- **Actually proves — 10 analyzers, not 5.** `tools/cilint/internal/analyzers/analyzers.go`'s `RunAll()` invokes: `TxOwnership`, `LegacyVocab`, `OutboxPair`, `PlatformBoundary`, `PostCommitAudit`, `DeliveryAuditSink`, `NoSQLTxInDomain`, `NoDualMode`, `NoResponseMap`, `HGCrossModule`. **`registry.go`'s own `Desc` field is stale/incomplete** — it omits `OutboxPair`, `PostCommitAudit`, `DeliveryAuditSink`, `NoDualMode`, `NoResponseMap` entirely. This is itself a doc-vs-code drift inside the guard tooling.
- **Baseline-and-ratchet mechanism** (`tools/cilint/baseline.go`, applies by default to `tools/cilint/baseline.json` since the registered check passes no `--baseline` override): findings keyed by `(analyzer, file, message)` — deliberately **not** by line number, to avoid churn. Compared against the recorded `count`: new finding → fail ("NEW VIOLATION"); count increased → fail; count **decreased** → **also fails** ("STALE BASELINE", forcing an explicit `--update-baseline` commit) — so debt is monotonically forced to shrink, never silently drifts upward unnoticed. Current `baseline.json`: 23 entries, all `hgcrossmodule` findings, each carrying a `reason` but `owner: "unclassified — no owner identified as of 2026-08-07"`.
- **Bypass surface on the baseline itself:** nothing in code enforces that `baseline.json` entries only originate from `--update-baseline` — any contributor can hand-edit the JSON in a PR to add a permanent suppression. Unlike the diff-scoped waiver channel (§10), a baseline.json entry, once merged, suppresses that exact finding for **every future PR forever**, not just the PR that added it. This is a materially weaker control than it first appears.
- **Per-analyzer detail:**

  | Analyzer | Property | Mechanic | Known bypass | Fixture? |
  |---|---|---|---|---|
  | `HGCrossModule` | DB read-ownership | regex `FROM`/`JOIN` + table literal vs hand-synced `hgOwnerByTable` (~40 tables → 10 modules) | **only catches reads** — `UPDATE table SET`/`INSERT INTO table` never match `FROM`/`JOIN` and are entirely invisible; also defeated by `fmt.Sprintf("... FROM %s", tbl)` parameterization; exemption matching is by file-path suffix so a rename silently drops an exemption (documented as having already happened once, X7) | yes, `hgcrossmodule_test.go`, comprehensive |
  | `PlatformBoundary` | REQ-TOP-2 platform domain-freedom | AST import scan of `internal/platform/**`, flags `internal/modules/**` imports unless the *whole file's package dir* is in `platformBoundaryAllowed` (4 entries: `bootstrap`, `authn`, `docgenv2`, `objectstore`) | whole-package exemption, not per-line — any new import inside an allowed package is invisible | no `_test.go` found |
  | `NoSQLTxInDomain` | domain persistence purity | flags `sql.Tx`/`DB`/`Rows`/`Row`/`Result` identifier references under `internal/modules/<m>/domain/` | matches literal identifier `sql` only — `import dbsql "database/sql"` then `dbsql.Tx` defeats it; a driver-native type like `pgx.Tx` is invisible | yes, `nosqltxindomain_test.go` |
  | `TxOwnership` | tx-boundary discipline | flags `BeginTx`/`Begin`/`Commit`/`Rollback` method calls outside `allowedTxPackages` | matches method name only, any receiver — false-flags an unrelated `Commit()` method; `//cilint:allow-tx` needs no justification text | no `_test.go` found |
  | `LegacyVocab` | retired-vocabulary guard | regex over `.go`/`.ts`/`.tsx` for `finalized`/`document.finalize`/`document.archive` | no inline suppression by design; `//cilint:allow-legacy` explicitly retired and proven-dead by a negative test | yes, in `analyzers_test.go` |
  | `OutboxPair`, `PostCommitAudit`, `DeliveryAuditSink`, `NoDualMode`, `NoResponseMap` | outbox pairing / post-commit audit ordering / no-optional-dependency-branching / no-ad-hoc-response-map | AST-level, real live-wired invariant guards | fixtures exist for `OutboxPair`, `PostCommitAudit`, `DeliveryAuditSink` (`analyzers_test.go`); `NoDualMode`/`NoResponseMap` have their own `_test.go` files | mixed |

- **Wired into `tools/verify`:** yes — `arch-lint` check, `go run ./tools/cilint ./...`, Profiles `fast,pr,full`, `CIJob: ci.yml:verify`.
- **Runs on PR:** yes.
- **Bypassable:** several concrete, documented mechanisms above (identifier-alias evasion, method-name-only matching, whole-file exemption, write-blind SQL ownership, hand-editable baseline).
- **Manual allowlist:** yes, at three layers simultaneously for `hgcrossmodule` alone — `hgExempt` (8 permanent ADR-0039 D3(d)-(f) entries), `hgPendingRemediation` (now empty debt ledger), and `baseline.json` (23 entries) function as three stacked suppression mechanisms for one rule.
- **Single source of truth:** `hgOwnerByTable` is explicitly labelled in-source `// TRANSITIONAL — hand-synced enumeration, the repo's known meta-defect`, citing a named deletion milestone ("M3-final: cross-module SQL closure") — this is the correctly-labelled-transitional pattern CLAUDE.md's Global Maximum rule asks for.

---

## 4. `tools/verify` — the master registry

- **Files:** `main.go` (869 lines — CLI/execution engine), `registry.go` (678 lines — **the single source of truth check table**, self-documented as such), `audit.go` (413 lines — the `--audit` self-check), `testdbbypass.go` (278 lines).
- **Full registered-check count: 40 distinct IDs.** Representative slice (full detail available on request; every check has ID/Profiles/Argv/CIJob):
  `gofmt`, `go-vet`, `go-vet-integration`, `arch-lint`, `problem-codes-drift`, `api-lint-e2e-base-path`, `api-lint`, `api-lint-selftest`, `contract-sync`, `codegen-drift-backend`, `codegen-drift-frontend`, `openapi-lint-v1`, `openapi-lint-e2e`, `oasdiff-breaking`, `module-imports`, `test-conventions`(+`-selftest`), `testdb-bypass-guard`, `adr-status`, `wiki-debt-tally`, `db-docs-coverage`, `migration-gapless`, `governance-diff-rules`, `invariant-coverage-map`, `gosec`, `govulncheck`, `eslint`, `css-tokens`, `eigenpal-selector-pin`, `fe-typecheck`, `fe-test`, `docx-typecheck`, `docx-build`, `docx-test`, `go-test-unit`, `go-test-integration`, `req-trace`(+`-selftest`), `required-gate-selftest`, `verify-audit`.
  Only `gosec`, `govulncheck`, `go-test-integration` are `ProfileFull`-only (never run by a whole-profile sweep in CI — only via targeted `--only=`).
- **`--audit` self-check (`audit.go`)** enforces 6 anti-drift rules by regex-parsing workflow YAML against `--only=`/`--profile=` flag text:
  - A1: a workflow runs an `--only=` ID not in the registry → fail.
  - A2: a check's `CIJob` names a nonexistent workflow:job → fail.
  - A3: the claimed job exists but its flags don't actually include that check ID → fail.
  - A4: a check has no `CIJob` at all (local-only, unenforced on PR) → flagged.
  - A5: `ci.yml:required`'s `needs:` list must exactly match `scripts/required-gate.jq`'s job-key array.
  - A6: every `ProfilePR` check's `CIJob` must sit inside `ci.yml:required`'s transitive `needs:` closure — this is precisely why `gosec`/`govulncheck` (CIJob `nightly.yml:security-scan`) are flagged as a known, commented-on gap: they cannot block a PR today.
  - `verify-audit` is itself a registered check (self-referential, runs `--audit`) on every profile including `fast`/`changed` — so audit-rot is gated on every PR, not just when a human remembers to run it manually.
- **`--changed` scoping (`main.go`):** outside a `pull_request`/`pull_request_target` GitHub event, `--changed` is a **no-op** (fail-closed — an empty diff on `push` would otherwise silently select zero checks). A check with no declared `Paths` **always matches** under `--changed` (documented fail-closed default). Path matching is plain `strings.HasPrefix`, not a glob engine, despite the `Paths` field's doc-comment calling entries "prefix/glob patterns" — a minor internal doc/code mismatch.
- **Ordering DAG (`After` edges):** `validateOrdering`/`detectOrderingCycle` reject unknown or cyclic `After` edges across the whole registry on every run; `validateSelectionOrdering` refuses to silently run a `--only` selection missing a declared `After` predecessor — proven-by-reproduction fix for a prior stale-artifact false-pass bug (`docx-test` reading stale `dist/meta.json`).
- **Fixtures:** `main_test.go`, `audit_test.go`, `testdbbypass_test.go` all exist and exercise the engine itself.
- **Bypassable:** the engine's own selection/ordering logic is well-guarded; the weak point is that individual registered checks can themselves be weak (see §§1-3, 5-12) — `tools/verify` faithfully runs what's registered, it doesn't strengthen a weak check.
- **Single source of truth:** yes, explicitly, by design and by the `--audit` self-check closing the loop between registry claims and actual workflow YAML.

---

## 5. Contract parity (OpenAPI ↔ generated ↔ frontend types)

| Check | Mechanic | Blind spot |
|---|---|---|
| `codegen-drift-backend` | `go generate ./...` then `git diff --exit-code` on 3 glob patterns (`**/api.gen.go`, `**/httpsurface_gen.go`, `**/httpsurface_e2e_gen.go`) | only those 3 patterns — any other generated artifact `go generate` touches is unchecked |
| `codegen-drift-frontend` | `pnpm run gen:api` then `git diff --exit-code -- src/lib/api-types/` | narrow to that directory |
| `oasdiff-breaking` | `oasdiff breaking base.yaml current.yaml --fail-on ERR`, base materialized from PR base SHA by a prerequisite step | requires network-adjacent `oasdiff` binary fetch (`--require-infra`) |
| `openapi-lint-v1`/`-e2e` | Redocly structural lint, pinned `@redocly/cli@2.46.0`, `redocly.yaml` enables only `operation-summary: off`, `security-defined: off`, `struct: error` | **most default Redocly rules are off** — repo comment: "133 errors at time of introduction… silenced pending a dedicated cleanup ticket" — an unlabelled-scope local maximum |
| `contract-sync` (`check-contract-sync-all.ps1`) | iterates **only 4 of 15 modules**: `templates`, `documents`, `controlleddocuments`, `taxonomy`. `approval` explicitly excluded (comment: "ownership entangled in M9 F9.5 approval-promotion decision") | **10 of 15 modules (incl. approval) have zero mechanical spec↔runtime↔wiki sync check.** Per-module check is a **substring-presence scanner** (`content.IndexOf(pattern) >= 0`), not structural parsing — proves textual presence anywhere in the file, including in a comment or dead code, not that a route is actually wired/executed. `Test-WikiStatusContradiction` is a prose-matching heuristic, trivially defeated by rewording a wiki claim without changing its truth. |

---

## 6. Authz guards (tripwire)

- **Source of truth:** `internal/platform/tripwire/arms.go` — `TripwireArms []Arm{Table, Op, Caps, When*}` is the single Go source for the DB-trigger `enforce_capability_asserted()`'s required-capability arms (ADR 0083 subject-discrimination). `RenderMigration()` deterministically regenerates the SQL from `TripwireArms`.
- **`TRIPWIRE-ARM-PARITY`:** (i) every `Cap` must be a real, registered capability (`iamdomain.IsValidCapability`); (ii) `RenderMigration()` must byte-equal the committed golden file `internal/platform/tripwire/golden/0301_...sql` — missing golden file is a hard error under `-strict` (CI), silently skipped otherwise.
- **`TRIPWIRE-ARM-DRIFT`:** for every Go function that both calls `authz.Require(ctx, tx, cap, area)` AND a literal-SQL `UPDATE`/`INSERT INTO`/`DELETE FROM` against a gated table, requires at least one asserted capability to be a member of that table's arm (match-one, mirroring the DB trigger's own CASE semantics). **Explicitly function-local only** — cross-layer drift (assert-in-service, write-in-repository, different functions) is out of scope by contract, deferred.
- **Blind spots:** only catches `Exec`/`ExecContext` calls with a **literal** SQL string — dynamic/query-builder SQL is invisible; function-local scope misses caller/callee splits; match-one semantics let an over-broad or merely-adjacent capability satisfy the check without asserting the semantically intended one.
- **Wired into `tools/verify`:** yes, inside the `api-lint` check (§2), `ProfilePR+Full`, BLOCKING per api-lint's design.
- **Live-drivable backstop (the true runtime enforcement):** `tests/integration/templates/tripwire_caps_test.go`, `tests/integration/documents/tripwire_documents_test.go`, `tests/integration/iam/tenants_tripwire_test.go` — actual DB-level tests proving the Postgres trigger itself rejects unauthorized writes. The api-lint rules are static pre-checks that Go source and generated SQL agree with each other; the DB trigger is the actual last line of defense (per CLAUDE.md's "two-tier PDP + DB tripwire last line").
- **Fixtures:** the arm-parity/drift rules have dedicated `_test.go` coverage inside `scripts/api-lint`.
- **Manual allowlist:** `tripwire-allowlist.txt` (pairing exceptions, format documented in file header — not independently re-verified against current content this pass).

---

## 7. DB ownership guards

**The only mechanical SQL-ownership guard in the repo is `HGCrossModule`** (§3 table). No separate `tools/`/`scripts/` SQL-ownership scanner exists outside `tools/cilint`.

- **Materially significant blind spot, restated for emphasis:** it is a **read**-ownership guard only. Cross-module **writes** via `UPDATE table SET ...` / `INSERT INTO table ...` are entirely invisible — neither keyword matches the `FROM`/`JOIN` regex. A module can freely mutate another module's base table via raw SQL today with zero mechanical detection, which is a strictly worse violation of REQ-TOP-1/DB-ownership than the read case the guard actually catches.
- Waiver stacking: `hgExempt` (permanent, 8 entries) + `hgPendingRemediation` (structural, currently empty) + `baseline.json` (23 entries, hand-editable, not scoped to `--update-baseline` origin) = three layers of suppression for one rule.

---

## 8. Migration guards

- **`migration-gapless` (`scripts/check-migration-gapless.sh`, `ProfilePR+Full`, needs git-history):** (1) if `db/migrations/*.sql` non-empty, verifies numeric filename prefixes form a gapless `MIN..MAX` run; (2) via `git log`, verifies no migration file already merged to `origin/main` was modified afterward.
- **Current state:** post the 2026-08-09 baseline fold (0257–0315 squashed into `db/baseline/`), `db/migrations/` holds only a README — the gapless half is currently vacuous (nothing to check); only the no-post-merge-edit half is live.
- **No migration-checksum / schema-drift tool** found inside `tools/verify`'s registry; `db/baseline/` drift is covered (if at all) by `check-db-bootstrap.ps1`, which is **not** a registered `tools/verify` check and therefore not confirmed gated on PR by this audit's method.

---

## 9. `.golangci.yml`

- `version: "2"`, `linters.default: none`, 13 enabled: `errcheck, govet, staticcheck, nilerr, errorlint, gosec, gocritic, revive, gocyclo, gocognit, sqlclosecheck, rowserrcheck, bodyclose, exhaustive, contextcheck`.
- **No `formatters:` block** — gofmt enforcement lives entirely in the separate `gofmt` registry check, not golangci-lint.
- Exclusions: `_test\.go$` exempt from `gosec,gocyclo,gocognit`; `.*\.gen\.go$` exempt from **all 13 linters**; `exclusions.generated: strict` additionally auto-exempts anything golangci-lint's own header-comment heuristic recognizes as generated.
- **CI invocation (`ci.yml:lint-go`)** scopes to `./apps/api/... ./internal/... ./tools/...` — **excludes `./scripts/...` and `./cmd/...`**, both real Go trees (`scripts/api-lint`, `scripts/req-trace`, `cmd/problem-codes-dump`, `cmd/gen-tripwire`). Those directories get only `go vet`/`go test`, never gosec/gocritic/revive/etc.
- This job is **whole-tree blocking**, not diff-scoped, and is a **bare GitHub Action step, not a `tools/verify` registry check** — invisible to the `--audit` A1-A6 self-checks.

---

## 10. Governance waiver channel (PR #99, commit `7adcd675`)

- **Files:** `scripts/check-governance-waivers.txt` (data) + `scripts/check-governance.ps1`'s `Test-Waived` function (mechanism).
- **Mechanic:** `Test-Waived($ruleId)` runs `git diff -w "$BaseRef...HEAD" -- scripts/check-governance-waivers.txt`, keeps only **added** lines, checks for a `<ruleId> |` prefix match. **Validity is diff-based, not content-based** — a waiver line only counts for the PR whose diff *adds* it; once merged, the same line sitting in the file does nothing for a later PR (this is the "auditable" property the channel is named for).
- **Governed rules:** `api-contract-openapi` (handler changed, spec didn't) and `domain-tests` (`internal/modules/` changed, nothing under `tests/`) — both labelled TRANSITIONAL with a named deletion trigger ("generated-boundary expansion", ROADMAP 4.8).
- **Current entries:** 2, both `2026-08-09`, both citing PR #99.
- **Bypassable:** yes, trivially at the PR-author level — no mechanical validation of the justification text; any contributor can add a plausible one-line waiver in their own PR. The control is human PR review (the diff makes the claim *visible and contestable*, not *unbypassable*). No CODEOWNERS restriction on the waiver file was found in this pass (open question, not fully verified).
- **No dedicated test/fixture** found for `check-governance.ps1`.

---

## 11. CSS token gate

- **`css-tokens` (`scripts/check-css-token-discipline.sh`):** `git ls-files 'src/**/*.module.css'`, greps `#[0-9a-fA-F]{3,8}`, reports every match **unless the entire file** is in a hardcoded 26-entry `ALLOWLIST`.
- **Blind spots:** whole-file exemption, not whole-line/whole-match — any of the 26 grandfathered files can have a brand-new raw hex color added anywhere with zero detection, contrary to the stated "no NEW raw hex" intent. Scope is `*.module.css` only — plain `.css`, inline `style={{color:'#fff'}}`, and CSS-in-JS are all out of scope. Pure grep, no CSS/AST parsing — comment-embedded hex strings false-positive; indirected custom-property values false-negative.
- **No dedicated test fixture** found.

---

## 12. Frontend import-boundary ESLint rules

- **File:** root `eslint.config.mjs` — the entire frontend ESLint surface (no separate config under `frontend/`). Explicitly narrow by design (top-of-file comment: "enforces ONE invariant... NOT a full lint regime"). `@typescript-eslint` and `eslint-plugin-react-hooks` are registered but **all their rules are OFF** (kept alive only so pre-existing `eslint-disable` comments don't error on an unknown rule name) — **the `eslint` registry check enforces no general code-quality rules at all**, only two import boundaries.
- **Boundary 1 — Eigenpal ACL (ADR 0046):** `no-restricted-imports` blocks `@eigenpal[/**]` everywhere except `packages/eigenpal-adapter/**` and `packages/editor-ui/**`.
- **Boundary 2 — Feature boundary (F1.4, GMR M1):** one `no-restricted-imports` block per feature dir (13 `FEATURE_NAMES`), forbidding cross-feature imports except a hand-enumerated, comment-labelled **shrink-only** `ALLOWLIST` of 19 pre-existing `(from,to)` edges. Uses regex patterns (not glob groups) specifically to handle relative-path depth (`../../`) without false-matching `documents` against `controlled-documents`.
- **Design care:** flat-config blocks **replace** (don't merge) `no-restricted-imports` per matching file, so each per-feature block explicitly re-includes the eigenpal patterns to avoid silently dropping that guard under `features/**` — a documented, deliberately handled flat-config footgun.
- **Blind spots:** the "shrink-only" `ALLOWLIST` discipline is a comment-level convention only — no test asserts the list only shrinks over commits; dynamic `import()` matching against modern ESLint versions wasn't independently re-verified this pass.
- **No dedicated test file** found for the ESLint config itself; the closest self-check is `eslint` passing clean on the whole tree, gated on every PR.

---

## Guard blind-spot matrix — the 8 named semantic properties

| # | Property | Firing guard today? | Detail | Gap owner (rulebook axis) |
|---|---|---|---|---|
| 1 | Module acyclicity | **NO** | `check-module-boundaries.ps1` only proves layer/published-package import-visibility rules — it has no notion of the module *graph* and cannot detect a reciprocal A→B, B→A edge pair. Rulebook §5 confirms 7 known module-level cycles exist today, undetected by any guard. | **#87/A1** (verifier product) — the fix is a new `tools/verify` check that computes the module import graph and runs SCC/cycle detection; this is tooling capability, not a new architectural boundary decision. |
| 2 | SQL/data ownership | **PARTIAL** | `HGCrossModule` catches cross-module **reads** (`FROM`/`JOIN`) but is structurally blind to cross-module **writes** (`UPDATE`/`INSERT INTO`) — a strictly worse violation, currently undetected. | **#93/A4** (new architecture checks) — extending the existing analyzer to also match `UPDATE <table> SET` / `INSERT INTO <table>` is a genuine new semantic check, not a mechanical registry fix, since it requires building the ownership-violation-on-write case from scratch (new fixtures, new baseline entries expected). |
| 3 | Foreign sentinel independence | **NO** | No guard anywhere checks whether one module's error handling depends on another module's exported sentinel error values/types (`errors.Is`/`errors.As` against a foreign package's sentinel). Confirmed via targeted grep across `tools/cilint` and `scripts/api-lint` — no analyzer references sentinel/error-identity cross-module coupling. | **#93/A4** — this requires a brand-new analyzer (no existing one is close in shape); it is a new architecture check per the rulebook's V-ARCH-5. |
| 4 | Platform domain-freedom | **YES, but overstated** | `PlatformBoundary` analyzer fires and is wired into CI, but exempts 4 whole packages (`bootstrap`, `authn`, `docgenv2`, `objectstore`) — target state is an empty allowlist. **`wiki/architecture/backend-blueprint.md` claims this invariant is "MET + CI-locked," which is false** given the live exemptions (cross-referenced in pass14). The guard exists and is real, but the wiki's characterization of its completeness does not match code truth. | Guard itself: **#87/A1** territory only if allowlist-shrinking tooling is wanted (e.g. a stale-allowlist rule like api-lint's `*-ALLOWLIST-STALE` pattern, §2, applied here); the wiki-accuracy gap is **not** a tooling gap, it's pass14 material. |
| 5 | Domain persistence purity | **YES, bypassable** | `NoSQLTxInDomain` fires and is wired into CI, but matches only the literal identifier `sql` — a renamed import alias (`dbsql "database/sql"`) defeats it entirely, as does any non-stdlib driver type (`pgx.Tx`) referenced directly in a domain package. | **#87/A1** — this is a hardening fix to an existing, correctly-scoped analyzer (switch from identifier-name matching to import-path-resolved type matching), not a new architectural decision. |
| 6 | Consumer-owned port shape | **NO** | No guard checks that Go interfaces are defined by the consuming module (idiomatic Go / hexagonal-port discipline) rather than exported by the producing module and imported wholesale. Not found in `tools/cilint`, `scripts/api-lint`, or `check-module-boundaries.ps1`. | **#93/A4** — a new analyzer would need to walk interface declarations and their usage sites across module boundaries; this is new semantic territory, not a registry wiring fix. |
| 7 | Error-writer uniqueness | **NO** | No guard verifies that exactly one code path per module (or per request) writes the RFC 9457 `problem+json` response — i.e., that error-writing isn't duplicated/forked across handler and service layers. `errcheck.exclude-functions` in `.golangci.yml` exempts `problem.Write` from unchecked-error linting, which is adjacent but does not check writer *uniqueness*. Not the same shape as `PostCommitAudit`/`DeliveryAuditSink` (those guard ordering/single-sink for audit writes and outbox delivery, not the generic problem+json response path). | **#93/A4** — closest existing analyzer shape (`PostCommitAudit`) could plausibly be adapted, but the invariant itself ("exactly one error-writer per request") has not been formalized as a rule anywhere, so this is new architecture-check work, not a tooling wiring gap. |
| 8 | Runtime request validation | **NO firing guard; partially structural** | Contract-first + `oapi-codegen` strict-server generation makes request-shape validation structurally likely for the 4 modules inside `contract-sync`'s gate, but there is **no mechanical check that every mounted route is actually served through the generated `ServerInterface` wrapper** rather than a hand-written `mux.HandleFunc` bypassing generated validation. `check-contract-sync-all.ps1`'s `Test-WikiStatusContradiction` heuristic gestures at this (string-matches `HandlerWithOptions`/`ServerInterfaceWrapper` presence) but only for the 4 gated modules, and only as prose-matching, not AST-verified wiring. | **#87/A1** for the 4 already-gated modules (extend the existing substring check to real AST/route-table verification); **#93/A4** for the 10 ungated modules (approval + 9 others), since building any check there starts from zero. |

---

## Cross-cutting findings worth flagging to the audit owner

1. **`registry.go`'s `arch-lint` description undercounts cilint's analyzers 5-for-10** (§3) — a doc-vs-code drift inside the guard tooling itself, not just in the wiki.
2. **`gosec`/`govulncheck` cannot block a PR today** — `ProfileFull`-only, `CIJob: nightly.yml:security-scan`, outside `ci.yml:required`'s closure. Self-documented as transitional pending an unscheduled "ruleset swap" milestone. This is the single highest-severity live gap found: a real gosec/govulncheck finding on a PR branch is currently invisible to the merge gate.
3. **Three governance/waiver mechanisms of very different rigor** exist side by side: `check-governance-waivers.txt` (diff-scoped, free-text, PR-review-dependent), `tools/cilint/baseline.json` (count-ratchet, but hand-editable with no origin enforcement — the weakest of the three despite looking the most mechanical), and `scripts/api-lint`'s `*-ALLOWLIST-STALE` companion rules (self-auditing, the strongest). An implementer reaching for "the waiver pattern" should be pointed at the third, not copy the second.
4. **Zero-fixture checks:** `module-imports`, `governance-diff-rules`, `css-tokens`, and (per spot-check) `platformboundary.go`/`txownership.go`/`legacyvocab.go` inside cilint have no dedicated positive/negative test. Only `api-lint`'s sub-rules, `test-conventions`, and `required-gate` are proven-by-test at the guard-mechanism level, not just at the target-code level.
5. **DB ownership is write-blind** (§7) — the single most concrete, fixable gap in the whole matrix: extending `HGCrossModule`'s regex to also match `UPDATE`/`INSERT INTO` is a bounded, well-scoped piece of work against an existing, well-tested analyzer.
