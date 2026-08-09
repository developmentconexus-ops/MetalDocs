package main

// The check registry. This is the single source of truth for what
// "verified" means in this repo: CI workflows call `go run ./tools/verify
// --profile=X` and nothing else, so a check exists here or it does not exist.
//
// Adding a check here without wiring the corresponding workflow to a profile
// that contains it produces a check that runs on laptops and not in CI —
// which is the same lie as an unwired script. `--audit` reports that case.

// Profile names. Ordered cheapest to most complete; each is a superset of the
// one before it EXCEPT `changed`, which is `pr` filtered by the diff.
const (
	ProfileFast    = "fast"    // inner loop, deterministic, target < 5 min
	ProfileChanged = "changed" // `pr` restricted to checks whose paths changed
	ProfilePR      = "pr"      // everything a PR must pass pre-merge
	ProfileFull    = "full"    // `pr` + integration suites (needs Postgres)
)

var profileOrder = []string{ProfileFast, ProfileChanged, ProfilePR, ProfileFull}

// Infra requirements. A check declaring any of these is skipped (loudly, with
// its reason) when the requirement is absent, and is never silently dropped.
const (
	needsPostgres = "postgres"    // METALDOCS_DATABASE_URL pointing at a live DB
	needsDocker   = "docker"      // a working docker daemon
	needsNetwork  = "network"     // fetches a tool or package at run time
	needsGitDepth = "git-history" // needs real history, not a shallow clone
)

// Check is one verifiable claim.
type Check struct {
	ID   string
	Desc string

	// Profiles this check belongs to. `changed` is derived, never listed.
	Profiles []string

	// Argv. Index 0 is looked up on PATH. No shell interpolation: a check is
	// an argv, so quoting bugs cannot silently change what ran.
	Argv []string

	// Dir is relative to repo root; empty means repo root.
	Dir string

	// Needs lists infra requirements. Empty means it runs on a bare laptop
	// with Go, Node, pnpm and git.
	Needs []string

	// After lists check IDs that must run, and pass, before this one starts.
	// This is an ORDERING edge, not an infra requirement — deliberately a
	// different field from Needs so the two classes of dependency can never
	// be confused by a reader or a future PR: Needs asks "is Postgres up",
	// After asks "did check X already succeed". Only declare an edge when a
	// later check consumes an earlier one's output (docx-v2-test reads the
	// dist/meta.json docx-v2-build produces) — most checks are independent
	// and must stay that way so -j keeps meaning what it means. A selection
	// that includes a check without its After predecessor is refused, not
	// run — see validateSelectionOrdering in main.go. A predecessor and its
	// dependent must always be selected together (--only=a,b, or a profile
	// that carries both).
	After []string

	// Paths are prefix/glob patterns; the `changed` profile runs this check
	// only when a changed file matches. Empty means always run under
	// `changed` — use that for checks whose scope is the whole repo.
	Paths []string

	// CIJob names the workflow job this check corresponds to, as
	// "<workflow file>:<job id>". Used by --audit to prove the mapping.
	// Empty means the check is local-only and no CI job runs it, which
	// --audit reports as a gap.
	CIJob string
}

// checks is the registry. Keep it sorted by ID.
var checks = []Check{
	// ---- Go: build and vet ------------------------------------------------
	{
		ID:       "go-build",
		Desc:     "go build ./... — the whole module compiles",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "build", "./..."},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:   "gofmt",
		Desc: "every tracked Go file is gofmt-clean",
		// Nothing enforced this before A1, and 96 files had drifted by the
		// time anyone swept. golangci-lint's enabled set has no formatter and
		// go vet does not look at layout, so the convention was held up
		// entirely by habit.
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-gofmt.sh"},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:       "go-vet",
		Desc:     "go vet ./...",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "vet", "./..."},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:   "go-vet-integration",
		Desc: "go vet -tags integration ./... — integration-tagged files are not compiled by an untagged build, so a seam signature change can break them invisibly",
		// Deliberately in `fast`: this is cheap and it is the exact gap that
		// has bitten this repo before (bit QR-C).
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "vet", "-tags", "integration", "./..."},
		CIJob:    "ci.yml:verify",
	},

	// ---- Go: lint ---------------------------------------------------------
	{
		ID:       "cilint",
		Desc:     "custom Go analyzers (hgcrossmodule, nosqltxindomain, platformboundary, txownership, legacyvocab) against the recorded baseline",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/cilint", "./..."},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:   "staticcheck",
		Desc: "staticcheck 2025.1.1 — pinned to the version CI uses",
		// 2024.1 fails to compile under Go 1.25 (its vendored x/tools hits
		// "invalid array length"); 2025.1.1 is the first release with Go 1.25
		// support and the same check set. Moved here from invariants.yml when
		// the staticcheck-action step was replaced by this check.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "honnef.co/go/tools/cmd/staticcheck@2025.1.1", "./..."},
		Needs:    []string{needsNetwork},
		CIJob:    "ci.yml:verify",
	},

	// ---- Contract ---------------------------------------------------------
	{
		ID:       "problem-codes-fresh",
		Desc:     "generated problem-code artifacts match the registry",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./cmd/problem-codes-dump", "-check"},
		// cmd/problem-codes-dump/ is the check's own tool: without it here, a
		// PR that only weakens the tool (and touches no internal/ or spec
		// file) would select zero checks over the change that most needs
		// catching. Same class of gap as required-gate-selftest's missing
		// scripts/required-gate.jq (whole-branch review C2).
		Paths: []string{"internal/", "api/openapi/", "wiki/references/problem-codes.md", "cmd/problem-codes-dump/"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "api-lint-base-path-v1",
		Desc:     "PATH-BASE-PREFIX on the v1 spec",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint", "-only", "PATH-BASE-PREFIX", "api/openapi/v1/openapi.yaml"},
		// scripts/api-lint/ is the check's own tool: api-lint-strict only lints
		// api/openapi/v1/openapi.yaml, so this check is the SOLE gate on
		// internal-e2e.yaml's base-prefix rule (see the sibling entry below) —
		// without scripts/api-lint/ here, a PR neutering PATH-BASE-PREFIX
		// inside the tool while touching no api/openapi/ file selects zero
		// checks over the spec that most needs catching (whole-branch review
		// C2 class).
		Paths: []string{"api/openapi/", "scripts/api-lint/"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "api-lint-base-path-e2e",
		Desc:     "PATH-BASE-PREFIX on the internal-e2e spec",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint", "-only", "PATH-BASE-PREFIX", "api/openapi/internal-e2e.yaml"},
		// scripts/api-lint/ is the check's own tool — see the comment on
		// api-lint-base-path-v1 above; this is the sole gate on
		// internal-e2e.yaml's base-prefix rule, so the gap was worse here.
		Paths: []string{"api/openapi/", "scripts/api-lint/"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "api-lint-strict",
		Desc:     "full API design-system lint, strict",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint/", "-strict", "api/openapi/v1/openapi.yaml", "."},
		Paths:    []string{"api/openapi/", "scripts/api-lint/"},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:       "api-lint-selftest",
		Desc:     "the api-lint tool's own tests",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "test", "./scripts/api-lint/...", "-count=1"},
		Paths:    []string{"scripts/api-lint/"},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:       "contract-sync",
		Desc:     "spec/generated/runtime contract sync across modules",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-contract-sync-all.ps1"},
		// scripts/check-contract-sync-all.ps1 is the check's own definition;
		// without it a PR editing only the script (weakening what it checks)
		// selects zero checks (whole-branch review C2).
		Paths: []string{"api/openapi/", "internal/", "scripts/check-contract-sync-all.ps1"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "codegen-drift-backend",
		Desc:     "go generate ./... produces no diff in generated Go artifacts (api.gen.go, httpsurface_gen.go, httpsurface_e2e_gen.go)",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-codegen-drift-backend.sh"},
		// scripts/check-codegen-drift-backend.sh is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"api/openapi/", "internal/", "apps/", "cmd/", "scripts/check-codegen-drift-backend.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "codegen-drift-frontend",
		Desc:     "pnpm run gen:api produces no diff in frontend/apps/web/src/lib/api-types/",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-codegen-drift-frontend.sh"},
		// scripts/check-codegen-drift-frontend.sh is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"api/openapi/", "frontend/", "scripts/check-codegen-drift-frontend.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "openapi-lint-v1",
		Desc:     "redocly lint on the v1 spec",
		Profiles: []string{ProfilePR, ProfileFull},
		// Pinned to 2.46.0, latest at pin time (2026-08-08) — same reasoning
		// as ci.yml's oasdiff pin: an unpinned lint tool can change what it
		// accepts with no diff in this repo to review.
		Argv:  []string{"npx", "--yes", "@redocly/cli@2.46.0", "lint", "api/openapi/v1/openapi.yaml"},
		Needs: []string{needsNetwork},
		// redocly.yaml at repo root configures this lint's rule set (rules:
		// struct: error, etc.) — without it here, a PR that only loosens
		// redocly.yaml (e.g. turning a rule off) selects zero checks
		// (whole-branch review C3 class: a repo-root config file absent from
		// every check's Paths).
		Paths: []string{"api/openapi/", "redocly.yaml"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "openapi-lint-e2e",
		Desc: "redocly lint on the internal-e2e spec",
		// Task 12: the e2e scaffolding document did not exist until now — a
		// gate authored before its subject would pass vacuously. Same command
		// as openapi-lint-v1, second document; the file stays excluded from
		// the public bundle and frontend codegen (codegen-drift-frontend),
		// only its own contract hygiene is gated here.
		Profiles: []string{ProfilePR, ProfileFull},
		// Pinned to 2.46.0 — same reasoning as openapi-lint-v1 above.
		Argv:  []string{"npx", "--yes", "@redocly/cli@2.46.0", "lint", "api/openapi/internal-e2e.yaml"},
		Needs: []string{needsNetwork},
		// redocly.yaml at repo root configures this lint's rule set (rules:
		// struct: error, etc.) — without it here, a PR that only loosens
		// redocly.yaml (e.g. turning a rule off) selects zero checks
		// (whole-branch review C3 class: a repo-root config file absent from
		// every check's Paths).
		Paths: []string{"api/openapi/", "redocly.yaml"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "oasdiff-breaking",
		Desc: "oasdiff breaking-change gate: PR head spec vs base-branch spec, --fail-on ERR",
		// The base-branch spec this diffs against is materialized to
		// /tmp/openapi.base.yaml by a workflow prerequisite step
		// (ci.yml:lint-contract "Materialize base-branch spec"), which needs
		// the PR's base SHA — a fact this registry cannot express as an argv
		// (no shell interpolation, see Check.Argv), so that materialization
		// stays a workflow step rather than becoming part of this check.
		// Running `--only=oasdiff-breaking` on a laptop without first
		// producing that file fails because the file does not exist, not
		// because of a real breaking change — expected, not a defect.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"oasdiff", "breaking", "/tmp/openapi.base.yaml", "api/openapi/v1/openapi.yaml", "--fail-on", "ERR"},
		Paths:    []string{"api/openapi/v1/"},
		CIJob:    "ci.yml:verify",
	},

	// ---- Architecture invariants -----------------------------------------
	{
		ID:       "module-boundaries",
		Desc:     "cross-module access goes through published interfaces",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-module-boundaries.ps1"},
		// scripts/check-module-boundaries.ps1 is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"internal/", "scripts/check-module-boundaries.ps1"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "test-discipline",
		Desc:     "new tests use the canonical framework for their class",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-test-discipline.sh"},
		// scripts/check-test-discipline.sh is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"internal/", "tests/", "apps/", "scripts/check-test-discipline.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "test-discipline-selftest",
		Desc:     "check-test-discipline.sh reads code and ignores Go line comments",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-test-discipline-selftest.sh"},
		Paths:    []string{"scripts/check-test-discipline.sh", "scripts/check-test-discipline-selftest.sh", "scripts/testdata/test-discipline/"},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:   "testdb-bypass-guard",
		Desc: "no _test.go file bypasses testdb.Open via raw DATABASE_URL/METALDOCS_DATABASE_URL + sql.Open (ADR 0034) — the class of defect fixed at least five times, most recently 1a0663f5",
		// Same self-invocation shape as verify-audit below: the check's own
		// logic lives in tools/verify itself (testdbbypass.go), tested by
		// go test ./tools/verify/... like any other file in this package,
		// and its Argv shells back into `go run ./tools/verify
		// --testdb-bypass-guard` as a subprocess rather than a separate
		// script or tool directory.
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/verify", "--testdb-bypass-guard"},
		// tools/verify/ is the check's own definition (whole-branch review
		// C2 class): without it here, a PR that only weakens
		// testdbbypass.go's detection while touching no internal/, apps/,
		// or tests/ file selects zero checks over the change that most
		// needs catching.
		Paths: []string{"internal/", "apps/", "tests/", "tools/verify/"},
		CIJob: "ci.yml:verify",
	},

	// ---- Governance -------------------------------------------------------
	{
		ID:       "adr-status",
		Desc:     "no ADR status block exceeds its line/char budget",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-adr-status.sh"},
		// scripts/check-adr-status.sh is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"wiki/decisions/", "scripts/check-adr-status.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "wiki-tally",
		Desc:     "every module doc's severity tally matches its tech-debt register",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/wiki-tally-check.ps1", "-All"},
		// scripts/wiki-tally-check.ps1 is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"wiki/modules/", "scripts/wiki-tally-check.ps1"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "db-dictionary",
		Desc:     "every baseline table has a wiki dictionary page",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-db-dictionary-coverage.ps1"},
		// scripts/check-db-dictionary-coverage.ps1 is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"db/baseline/", "wiki/database/tables/", "scripts/check-db-dictionary-coverage.ps1"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "migration-gapless",
		Desc:     "db/migrations is a gapless sequence and no historical migration was edited after merge",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-migration-gapless.sh"},
		Needs:    []string{needsGitDepth},
		// scripts/check-migration-gapless.sh is the check's own definition —
		// named explicitly in the whole-branch review's C2 finding.
		Paths: []string{"db/migrations/", "scripts/check-migration-gapless.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "governance-diff-rules",
		Desc: "cross-cutting diff rules: contract changes ship with an OpenAPI update, domain changes ship with tests, ops changes ship with a runbook update",
		// No Paths: check-governance.ps1 reads git diff itself and its rules
		// span internal/modules/, api/openapi/, tests/, scripts/, deploy/ and
		// docs/runbooks/ — a path list narrow enough to be honest here would
		// just be "the whole repo except docs prose and frontend", which is
		// not a claim worth making.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-governance.ps1"},
		Needs:    []string{needsGitDepth},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:       "invariant-coverage-map",
		Desc:     "every invariant row in e2e/COVERAGE.md has ≥1 mapped spec ID",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-invariant-coverage-map.sh"},
		// scripts/check-invariant-coverage-map.sh is the check's own
		// definition — named explicitly in the whole-branch review's C2
		// finding.
		Paths: []string{"frontend/apps/web/e2e/COVERAGE.md", "scripts/check-invariant-coverage-map.sh"},
		CIJob: "ci.yml:verify",
	},

	// ---- Security -----------------------------------------------------------
	// Both entries below are TRANSITIONAL local maxima under this repository's
	// "labelled or it's a defect" rule (CLAUDE.md, "Global Maximum, Not Local
	// Maximum"). Measured in docs/superpowers/reports/2026-08-08-gosec-govulncheck-measurement.md
	// (Task 4 of the CI restructure); Task 9 consumes that measurement rather
	// than re-measuring.
	{
		ID:   "gosec",
		Desc: "no gosec rule fires on the Go tree",
		// 64 findings across 11 rules (top: G304=23, G201=10, G104=8), a mix of
		// real issues and false positives on the sample triaged — not a clean
		// tree, so `full`-only/advisory, not blocking.
		//
		// Global-maximum structure: gosec blocking in `pr` + `full`, every
		// finding fixed or carrying a gosec-native `#nosec Gxxx -- reason`
		// suppression (`//nolint:gosec` is golangci-lint syntax; standalone
		// gosec does not read it — the two existing comments on this pattern,
		// tools/verify/main.go:335 and tools/verify/audit.go:76, show up as
		// live findings under a naive registration for exactly this reason).
		// Promoting milestone: "gosec backlog triage" — unscheduled on
		// docs/superpowers/ROADMAP.md as of 2026-08-08.
		//
		// -exclude-dir=.claude is load-bearing, not cosmetic: an unexcluded
		// scan walks into .claude/worktrees/<sibling>/, a sibling git worktree
		// with its own go.mod, inflating the true 176 import directories to
		// 333 (157, 47%, from the sibling worktree alone) — measured on this
		// machine both before and after the flag. `-no-fail` from the
		// measurement's own invocation is deliberately NOT carried over: this
		// registration must fail the check like any other, not just record it.
		Profiles: []string{ProfileFull},
		Argv:     []string{"go", "run", "github.com/securego/gosec/v2/cmd/gosec@latest", "-quiet", "-exclude-dir=.claude", "./..."},
		Needs:    []string{needsNetwork},
		CIJob:    "nightly.yml:security-scan",
	},
	{
		ID:   "govulncheck",
		Desc: "no known-vulnerable symbol is reachable from any binary",
		// 19 total findings, but only 2 are called/reachable (GO-2026-5970 in
		// golang.org/x/text, GO-2026-5856 in stdlib crypto/tls) — both
		// call-graph-verified, not a naive dependency match. Those 2 are NOT
		// yet remediated as of this registration, so this stays `full`-only
		// (advisory), never `pr`-blocking, until they are.
		//
		// Global-maximum structure: govulncheck blocking in `pr` + `full` with
		// zero called vulnerabilities outstanding. Promoting milestone:
		// "called-CVE remediation" — unscheduled on docs/superpowers/ROADMAP.md
		// as of 2026-08-08; its two entry criteria are already measured: bump
		// golang.org/x/text to v0.39.0, bump the Go toolchain to go1.26.5.
		Profiles: []string{ProfileFull},
		Argv:     []string{"go", "run", "golang.org/x/vuln/cmd/govulncheck@latest", "./..."},
		Needs:    []string{needsNetwork},
		CIJob:    "nightly.yml:security-scan",
	},

	// ---- Frontend ---------------------------------------------------------
	{
		ID:       "fe-eslint",
		Desc:     "eslint across the workspace, including the eigenpal import boundary",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "lint"},
		// package.json (root) is where the "lint" script this check invokes
		// is defined — without it here, a PR rewriting "lint" to a no-op
		// while touching no frontend/, packages/, apps/, or config file
		// disarms eslint and selects nothing that would notice (whole-branch
		// review C2 class).
		Paths: []string{"frontend/", "packages/", "apps/", "eslint.config.mjs", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "css-token-discipline",
		Desc:     "no new raw hex colors in module.css",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-css-token-discipline.sh"},
		// scripts/check-css-token-discipline.sh is the check's own
		// definition — named explicitly in the whole-branch review's C2
		// finding.
		Paths: []string{"frontend/", "scripts/check-css-token-discipline.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "eigenpal-selector-pin",
		Desc:     "eigenpal version and selector counts are pinned together (ADR 0046, second half)",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-eigenpal-selector-pin.sh"},
		// scripts/check-eigenpal-selector-pin.sh is the check's own
		// definition — named explicitly in the whole-branch review's C2
		// finding.
		Paths: []string{"frontend/", "packages/eigenpal-adapter/", "scripts/check-eigenpal-selector-pin.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "fe-typecheck",
		Desc:     "tsc over @metaldocs/web",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "--filter", "@metaldocs/web", "run", "typecheck"},
		// Root toolchain/dependency files, not just frontend/ and packages/:
		// a lockfile-only or Node-version-only PR (Dependabot's normal shape)
		// otherwise selects zero frontend checks while the pathless Go
		// checks still run, so `required` reports success over an unrun
		// frontend suite (whole-branch review C3).
		Paths: []string{"frontend/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "fe-test",
		Desc:     "vitest over @metaldocs/web",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "--filter", "@metaldocs/web", "run", "test"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"frontend/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "docx-v2-typecheck",
		Desc:     "tsc over the docx-v2 workspace",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "typecheck:docx-v2"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"apps/docx-renderer/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		// I8: docx-renderer.yml:node is where this check actually runs today,
		// but it is a new-topology workflow that is NOT inside ci.yml:required's
		// needs: closure and will not gate a merge after the Phase 4 ruleset
		// swap (required_status_checks becomes exactly [{"context": "required"}]).
		// ci.yml:verify's own --profile=changed selection also runs this check
		// (Paths above match), so that is the honest CIJob for audit purposes.
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "docx-v2-build",
		Desc:     "the docx-v2 workspace builds; produces the dist/meta.json that docx-v2-test's bundle guard reads",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "build:docx-v2"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"apps/docx-renderer/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		// I8: same CIJob reasoning as docx-v2-typecheck above.
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "docx-v2-test",
		Desc: "docx-v2 unit tests",
		// D-14 (closed by R4): depends on docx-v2-build having already run —
		// bundle-guard.test.ts reads dist/meta.json. This used to be enforced
		// only by docx-renderer.yml:node splitting into two `verify`
		// invocations, so the now-required ci.yml:verify job — which runs
		// both checks in one `--profile=changed` invocation — raced them.
		// The After edge below is the real fix: an ordering primitive in the
		// registry itself, honoured by run() regardless of how many
		// invocations a workflow happens to split across.
		//
		// D-14 UPDATE (final review, Critical 2): the npm scripts this argv
		// calls used to be `pnpm -r run <script>`, which is recursive over
		// EVERY workspace including frontend/apps/web — so this check and
		// fe-test ran the same 154-file vitest suite concurrently in the same
		// tree, with docx-v2-build's `pnpm -r run build` writing
		// frontend/apps/web/dist underneath both. That race, not the
		// build-before-test ordering, was the reproducible source of
		// flaky/contradictory results between this check and fe-test. The npm
		// scripts now filter to `./packages/**` + `./apps/**` only, which does
		// not include frontend/apps/web — so docx-v2-test and fe-test no
		// longer touch the same files at all.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "test:docx-v2"},
		After:    []string{"docx-v2-build"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"apps/docx-renderer/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		// I8: same CIJob reasoning as docx-v2-typecheck above.
		CIJob: "ci.yml:verify",
	},

	// ---- Tests ------------------------------------------------------------
	{
		ID:       "go-test-unit",
		Desc:     "go test ./... (no integration tag)",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "test", "-count=1", "-timeout", "600s", "./..."},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:   "go-test-integration",
		Desc: "the full integration suite with -race",
		// A1 item 4: this is why `full` exists. It is push-only in CI today,
		// which makes it a post-mortem rather than a gate.
		Profiles: []string{ProfileFull},
		Argv:     []string{"go", "test", "-tags", "integration", "-count=1", "-race", "-timeout", "900s", "./tests/...", "./internal/...", "./apps/..."},
		Needs:    []string{needsPostgres},
		// Declared honestly wide (R3, ci.yml:test-integration --changed). The
		// Argv above only names tests/, internal/, apps/, but anything that can
		// change what those packages build against or run against can break
		// the suite without touching a line inside them:
		//   - go.mod, go.sum: a dependency bump changes every package's build,
		//     which is everything ./tests/... ./internal/... ./apps/... compile.
		//   - db/: migrations, baseline schema, grants, prerequisites, dev
		//     seeds and reference data are what the suite's real Postgres is
		//     bootstrapped from (internal/platform/migrate,
		//     internal/platform/config read db/... at runtime) — a schema-only
		//     edit here can break the suite with zero Go diff.
		//   - internal/, apps/, tests/: the three roots the Argv actually runs.
		// When in doubt this unit's brief says include the path, so this list
		// is deliberately roots, not the narrower subset that happened to be
		// touched by any one past incident.
		Paths: []string{"go.mod", "go.sum", "db/", "internal/", "apps/", "tests/"},
		CIJob: "ci.yml:test-integration",
	},

	// ---- Traceability -----------------------------------------------------
	{
		ID:       "req-trace-selftest",
		Desc:     "the req-trace tool's own tests",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "test", "./scripts/req-trace/...", "-count=1"},
		Paths:    []string{"scripts/req-trace/"},
		CIJob:    "ci.yml:verify",
	},
	{
		ID:       "req-trace",
		Desc:     "every MUST requirement cites live test evidence",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/req-trace"},
		Needs:    []string{needsGitDepth},
		// scripts/req-trace/ is the check's own tool (whole-branch review C2
		// class): without it here, a PR that only weakens the tool (e.g. what
		// counts as "live test evidence") while touching no
		// wiki/architecture/, internal/, or apps/ file selects zero checks
		// over the change that most needs catching. req-trace-selftest also
		// covers this directory, but that is unit-test coverage of the tool's
		// internals, not a substitute for this check itself running against
		// a diff that touches the tool.
		Paths: []string{"wiki/architecture/", "internal/", "apps/", "scripts/req-trace/"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "required-gate-selftest",
		Desc:     "the ci.yml `required` aggregator accepts and rejects the right needs-result sets",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-required-gate.sh"},
		// scripts/required-gate.jq is the expression this check exists to
		// pin down (audit rule A5 binds it to ci.yml:required's needs: list,
		// but A5 only compares the key array, not the predicate) — a PR
		// editing only the .jq (e.g. loosening `== "success"` to
		// `!= "failure"`) must select this check too, or it is the one PR
		// that can weaken the gate expression while selecting zero gate
		// checks (whole-branch review C2).
		Paths: []string{".github/workflows/ci.yml", "scripts/check-required-gate.sh", "scripts/required-gate.jq", "scripts/testdata/required-gate/"},
		CIJob: "ci.yml:verify",
	},

	// ---- Anti-drift ---------------------------------------------------------
	{
		ID:   "verify-audit",
		Desc: "registry vs CI workflow YAML cross-check (--audit) reports 0 findings",
		// R4: --audit was this branch's central anti-drift claim and was wired
		// into no workflow — a detector nothing runs detects nothing. This
		// entry is what runs it. It has to be registered like any other check
		// (Argv shells back out to `verify --audit` as a subprocess) rather
		// than special-cased, or it would repeat the exact defect it exists to
		// close: a control that only fires when a human remembers to type it.
		//
		// No Paths: the audit's correctness depends on the whole registry and
		// every workflow file, not a scoped subset — same reasoning as
		// governance-diff-rules above.
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/verify", "--audit"},
		CIJob:    "ci.yml:verify",
	},
}
