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
		CIJob:    "test-smoke.yml:unit",
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
		CIJob:    "invariants.yml:staticcheck",
	},
	{
		ID:       "go-vet",
		Desc:     "go vet ./...",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "vet", "./..."},
		CIJob:    "invariants.yml:staticcheck",
	},
	{
		ID:   "go-vet-integration",
		Desc: "go vet -tags integration ./... — integration-tagged files are not compiled by an untagged build, so a seam signature change can break them invisibly",
		// Deliberately in `fast`: this is cheap and it is the exact gap that
		// has bitten this repo before (bit QR-C).
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "vet", "-tags", "integration", "./..."},
		CIJob:    "invariants.yml:staticcheck",
	},

	// ---- Go: lint ---------------------------------------------------------
	{
		ID:       "cilint",
		Desc:     "custom Go analyzers (hgcrossmodule, nosqltxindomain, platformboundary, txownership, legacyvocab) against the recorded baseline",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/cilint", "./..."},
		CIJob:    "invariants.yml:cilint",
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
		CIJob:    "invariants.yml:staticcheck",
	},

	// ---- Contract ---------------------------------------------------------
	{
		ID:       "problem-codes-fresh",
		Desc:     "generated problem-code artifacts match the registry",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./cmd/problem-codes-dump", "-check"},
		Paths:    []string{"internal/", "api/openapi/", "wiki/references/problem-codes.md"},
		CIJob:    "api-contract.yml:problem-codes-freshness",
	},
	{
		ID:       "api-lint-base-path-v1",
		Desc:     "PATH-BASE-PREFIX on the v1 spec",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint", "-only", "PATH-BASE-PREFIX", "api/openapi/v1/openapi.yaml"},
		Paths:    []string{"api/openapi/"},
		CIJob:    "api-contract.yml:spec-base-path-gate",
	},
	{
		ID:       "api-lint-base-path-e2e",
		Desc:     "PATH-BASE-PREFIX on the internal-e2e spec",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint", "-only", "PATH-BASE-PREFIX", "api/openapi/internal-e2e.yaml"},
		Paths:    []string{"api/openapi/"},
		CIJob:    "api-contract.yml:spec-base-path-gate",
	},
	{
		ID:       "api-lint-strict",
		Desc:     "full API design-system lint, strict",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint/", "-strict", "api/openapi/v1/openapi.yaml", "."},
		Paths:    []string{"api/openapi/", "scripts/api-lint/"},
		CIJob:    "api-contract.yml:api-design-system-lint",
	},
	{
		ID:       "api-lint-selftest",
		Desc:     "the api-lint tool's own tests",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "test", "./scripts/api-lint/...", "-count=1"},
		Paths:    []string{"scripts/api-lint/"},
		CIJob:    "api-contract.yml:api-design-system-lint",
	},
	{
		ID:       "contract-sync",
		Desc:     "spec/generated/runtime contract sync across modules",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-contract-sync-all.ps1"},
		Paths:    []string{"api/openapi/", "internal/"},
		CIJob:    "api-contract.yml:contract-sync",
	},

	// ---- Architecture invariants -----------------------------------------
	{
		ID:       "module-boundaries",
		Desc:     "cross-module access goes through published interfaces",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-module-boundaries.ps1"},
		Paths:    []string{"internal/"},
		CIJob:    "module-boundaries.yml:conformance",
	},
	{
		ID:       "test-discipline",
		Desc:     "new tests use the canonical framework for their class",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-test-discipline.sh"},
		Paths:    []string{"internal/", "tests/", "apps/"},
		CIJob:    "module-boundaries.yml:conformance",
	},
	{
		ID:       "test-discipline-selftest",
		Desc:     "check-test-discipline.sh reads code and ignores Go line comments",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-test-discipline-selftest.sh"},
		Paths:    []string{"scripts/check-test-discipline.sh", "scripts/check-test-discipline-selftest.sh", "scripts/testdata/test-discipline/"},
		CIJob:    "module-boundaries.yml:conformance",
	},

	// ---- Governance -------------------------------------------------------
	{
		ID:       "adr-status",
		Desc:     "no ADR status block exceeds its line/char budget",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-adr-status.sh"},
		Paths:    []string{"wiki/decisions/"},
		CIJob:    "governance-check.yml:check",
	},
	{
		ID:       "wiki-tally",
		Desc:     "every module doc's severity tally matches its tech-debt register",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/wiki-tally-check.ps1", "-All"},
		Paths:    []string{"wiki/modules/"},
		CIJob:    "governance-check.yml:wiki-tally",
	},
	{
		ID:       "db-dictionary",
		Desc:     "every baseline table has a wiki dictionary page",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-db-dictionary-coverage.ps1"},
		Paths:    []string{"db/baseline/", "wiki/database/tables/"},
		CIJob:    "governance-check.yml:db-dictionary-coverage",
	},

	// ---- Frontend ---------------------------------------------------------
	{
		ID:       "fe-eslint",
		Desc:     "eslint across the workspace, including the eigenpal import boundary",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "lint"},
		Paths:    []string{"frontend/", "packages/", "apps/", "eslint.config.mjs"},
		CIJob:    "lint.yml:eslint",
	},
	{
		ID:       "css-token-discipline",
		Desc:     "no new raw hex colors in module.css",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-css-token-discipline.sh"},
		Paths:    []string{"frontend/"},
		CIJob:    "lint.yml:css-token-discipline",
	},
	{
		ID:       "eigenpal-selector-pin",
		Desc:     "eigenpal version and selector counts are pinned together (ADR 0046, second half)",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-eigenpal-selector-pin.sh"},
		Paths:    []string{"frontend/", "packages/eigenpal-adapter/"},
		CIJob:    "lint.yml:eigenpal-selector-pin",
	},
	{
		ID:       "fe-typecheck",
		Desc:     "tsc over @metaldocs/web",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "--filter", "@metaldocs/web", "run", "typecheck"},
		Paths:    []string{"frontend/", "packages/"},
		CIJob:    "fe-ci.yml:web-typecheck-test",
	},
	{
		ID:       "fe-test",
		Desc:     "vitest over @metaldocs/web",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "--filter", "@metaldocs/web", "run", "test"},
		Paths:    []string{"frontend/", "packages/"},
		CIJob:    "fe-ci.yml:web-typecheck-test",
	},
	{
		ID:       "docx-v2-typecheck",
		Desc:     "tsc over the docx-v2 workspace",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "typecheck:docx-v2"},
		Paths:    []string{"apps/docx-renderer/", "packages/"},
		CIJob:    "docx-renderer.yml:node",
	},
	{
		ID:       "docx-v2-build",
		Desc:     "the docx-v2 workspace builds; produces the dist/meta.json that docx-v2-test's bundle guard reads",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "build:docx-v2"},
		Paths:    []string{"apps/docx-renderer/", "packages/"},
		CIJob:    "docx-renderer.yml:node",
	},
	{
		ID:   "docx-v2-test",
		Desc: "docx-v2 unit tests",
		// Depends on docx-v2-build having already run: bundle-guard.test.ts
		// reads dist/meta.json. The registry has no dependency edges, so CI
		// enforces the order by using two verify invocations (see docx-renderer.yml:node)
		// and a local `--profile=pr` can still race. Recorded as D-14.
		//
		// D-14 UPDATE (final review, Critical 2): the npm scripts this argv
		// calls used to be `pnpm -r run <script>`, which is recursive over
		// EVERY workspace including frontend/apps/web — so this check and
		// fe-test ran the same 154-file vitest suite concurrently in the same
		// tree, with docx-v2-build's `pnpm -r run build` writing
		// frontend/apps/web/dist underneath both. That race, not the
		// build-before-test ordering below, was the reproducible source of
		// flaky/contradictory results between this check and fe-test. The npm
		// scripts now filter to `./packages/**` + `./apps/**` only, which does
		// not include frontend/apps/web — so docx-v2-test and fe-test no
		// longer touch the same files at all, and D-14 is left covering only
		// the narrower build-before-test ordering within the docx workspaces
		// themselves.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "test:docx-v2"},
		Paths:    []string{"apps/docx-renderer/", "packages/"},
		CIJob:    "docx-renderer.yml:node",
	},

	// ---- Tests ------------------------------------------------------------
	{
		ID:       "go-test-unit",
		Desc:     "go test ./... (no integration tag)",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "test", "-count=1", "-timeout", "600s", "./..."},
		CIJob:    "test-smoke.yml:unit",
	},
	{
		ID:   "go-test-integration",
		Desc: "the full integration suite with -race",
		// A1 item 4: this is why `full` exists. It is push-only in CI today,
		// which makes it a post-mortem rather than a gate.
		Profiles: []string{ProfileFull},
		Argv:     []string{"go", "test", "-tags", "integration", "-count=1", "-race", "-timeout", "900s", "./tests/...", "./internal/...", "./apps/..."},
		Needs:    []string{needsPostgres},
		CIJob:    "test-full.yml:full",
	},

	// ---- Traceability -----------------------------------------------------
	{
		ID:       "req-trace-selftest",
		Desc:     "the req-trace tool's own tests",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "test", "./scripts/req-trace/...", "-count=1"},
		Paths:    []string{"scripts/req-trace/"},
		CIJob:    "req-traceability.yml:gate",
	},
	{
		ID:       "req-trace",
		Desc:     "every MUST requirement cites live test evidence",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/req-trace"},
		Needs:    []string{needsGitDepth},
		Paths:    []string{"wiki/architecture/", "internal/", "apps/"},
		CIJob:    "req-traceability.yml:gate",
	},
	{
		ID:       "required-gate-selftest",
		Desc:     "the ci.yml `required` aggregator accepts and rejects the right needs-result sets",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-required-gate.sh"},
		Paths:    []string{".github/workflows/ci.yml", "scripts/check-required-gate.sh", "scripts/testdata/required-gate/"},
		CIJob:    "ci.yml:governance",
	},
}
