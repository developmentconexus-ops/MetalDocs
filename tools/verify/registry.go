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
	ProfileRelease = "release" // `full` minus the checks whose subject is a PR diff
)

var profileOrder = []string{ProfileFast, ProfileChanged, ProfilePR, ProfileFull, ProfileRelease}

// releaseExcluded lists the checks `release` drops from `full`.
//
// `release` runs on a tag, where there is no pull request and therefore no
// base to diff against. The three checks below do not merely happen to use git
// — their subject IS the difference between a branch and its base, so on a tag
// they can only fail for the wrong reason (a missing base spec, an empty
// diff), and a gate that fails for the wrong reason teaches people to ignore
// it. Everything else in `full` runs, including the integration suite and both
// security scanners: a release is the one moment where "everything, no path
// scoping, no exceptions" is the correct cost.
//
// Membership is expressed here rather than as a Profiles entry on 40-odd
// checks so that a NEW check is in `release` by default. Opting out has to be
// a deliberate line in this map, which is reviewable; forgetting to opt IN is
// silent, and silence is how coverage rots.
var releaseExcluded = map[string]string{
	"oasdiff-breaking":      "diffs the PR head spec against a base-branch spec materialized by a PR-only workflow step; on a tag the file does not exist",
	"governance-diff-rules": "rules about what a PR's diff must contain (contract change ships a spec update, etc.); a tag has no diff to rule on",
	"migration-gapless":     "\"no historical migration edited after merge\" is a property of a diff against origin/main",
}

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
	// later check consumes an earlier one's output (docx-test reads the
	// dist/meta.json docx-build produces) — most checks are independent
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

	// Fixture is the mechanical proof that this check can fail: bad input
	// fed to the check's own command, which must exit non-zero. See
	// fixtures.go. Nil is allowed only with a FixtureWaiver.
	Fixture *Fixture

	// FixtureWaiver records, in one sentence, why this blocking check carries
	// no negative fixture, and who owns closing that. It is not a bypass —
	// it is the audit trail for a known hole. Audit rule A7 requires every
	// blocking check to carry exactly one of Fixture or FixtureWaiver, so a
	// hole cannot exist without being named.
	FixtureWaiver string
}

// checks is the registry. Keep it sorted by ID.
var checks = []Check{
	// ---- Go: build and vet ------------------------------------------------
	// go-build (go build ./... — the whole module compiles) deleted Task 12:
	// go test ./... compiles every package including those with no test
	// files, and go vet ./... compiles them again. Neither links binaries,
	// so go build ./... verified nothing the other two lacked. See the
	// deletion ledger (docs/superpowers/reports/2026-08-08-workflow-deletion-ledger.md).
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
		Fixture: &Fixture{
			Dir:  "gofmt",
			Want: "internal/bad/unformatted.go",
		},
	},
	{
		ID:            "go-vet",
		FixtureWaiver: "third-party tool (cmd/vet), not a guard this repo authored; its own failure modes are Go's to prove, and a fixture here would test the Go toolchain.",
		Desc:          "go vet ./...",
		Profiles:      []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:          []string{"go", "vet", "./..."},
		CIJob:         "ci.yml:verify",
	},
	{
		ID:            "go-vet-integration",
		FixtureWaiver: "same third-party tool as go-vet under a build tag; what this entry adds is the tag, which the registry declares rather than implements.",
		Desc:          "go vet -tags integration ./... — integration-tagged files are not compiled by an untagged build, so a seam signature change can break them invisibly",
		// Deliberately in `fast`: this is cheap and it is the exact gap that
		// has bitten this repo before (bit QR-C).
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "vet", "-tags", "integration", "./..."},
		CIJob:    "ci.yml:verify",
	},

	// ---- Go: lint ---------------------------------------------------------
	{
		ID:            "golangci-lint",
		FixtureWaiver: "third-party linter aggregator (pinned); a fixture here would test golangci-lint's own analyzers, not this repo. The repo-authored part is .golangci.yml's enabled set, which every run exercises against the whole tree.",
		Desc:          "golangci-lint over apps/api, internal and tools (.golangci.yml scope)",
		// A1.1. This ran for months as a bare golangci-lint-action step in
		// ci.yml:lint-go — outside the registry, so `verify --audit` could not
		// see it, `verify --profile=pr` did not run it, and a laptop run and a
		// CI run disagreed about what "verified" means. That is a second,
		// parallel definition of the gate, which is the exact thing A1 exists
		// to remove. The command now lives here; the workflow installs the
		// pinned binary and calls the verifier, the same shape ci.yml already
		// uses for oasdiff.
		//
		// Pinned at v2.11.4 — the newest v2.11.x, which is what the Action's
		// `version: v2.11` resolved to, so this pin changes no behaviour today
		// and stops the resolution from moving tomorrow.
		//
		// No Paths: .golangci.yml at the repo root configures the whole run, and
		// the scope below spans three trees. A path filter here would let a
		// change loosen the config while selecting nothing that notices.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"golangci-lint", "run", "--timeout=5m", "./apps/api/...", "./internal/...", "./tools/..."},
		CIJob:    "ci.yml:lint-go",
	},
	{
		ID: "arch-lint",
		// The analyzer list is the one tools/cilint/internal/analyzers.RunAll
		// actually invokes. It said five for as long as there were ten
		// (pass13-guards.md §3) — a registry that describes the wrong product
		// is the same class of untruth as a check that does not run.
		Desc:     "custom Go analyzers (hgcrossmodule, nosqltxindomain, platformboundary, txownership, legacyvocab, outboxpair, postcommitaudit, deliveryauditsink, nodualmode, noresponsemap) against the recorded baseline",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/cilint", "./..."},
		CIJob:    "ci.yml:verify",
		Fixture: &Fixture{
			// A4.0: approval mutating documents' base table. The sandbox has
			// no tools/cilint/baseline.json, and cilint fails on any finding
			// when no baseline is configured — so the fixture proves the
			// analyzer, not the ratchet.
			Dir:  "arch-lint",
			Want: `writes "documents"'s base table "documents"`,
		},
	},
	// staticcheck (pinned honnef.co/go/tools/cmd/staticcheck) deleted Task 12
	// / spec §4.6: golangci-lint becomes the single Go lint umbrella and
	// .golangci.yml already enables staticcheck, so this standalone entry was
	// running the same analyzer a second time at whole-tree scope instead of
	// golangci-lint's diff-scoped run — redundant, not broken. See the
	// deletion ledger.

	// ---- Contract ---------------------------------------------------------
	{
		ID:       "problem-codes-drift",
		Desc:     "generated problem-code artifacts match the registry",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./cmd/problem-codes-dump", "-check"},
		// cmd/problem-codes-dump/ is the check's own tool: without it here, a
		// PR that only weakens the tool (and touches no internal/ or spec
		// file) would select zero checks over the change that most needs
		// catching. Same class of gap as required-gate-selftest's missing
		// scripts/required-gate.jq (whole-branch review C2).
		Paths:   []string{"internal/", "api/openapi/", "wiki/references/problem-codes.md", "cmd/problem-codes-dump/"},
		CIJob:   "ci.yml:verify",
		Fixture: &Fixture{Dir: "problem-codes-drift", Want: "problem-codes-dump"},
	},
	// api-lint-base-path-v1 (PATH-BASE-PREFIX on the v1 spec, standalone
	// -only run) deleted Task 12 (six-control table, spec §4.5): -only is a
	// filter, not a mode (scripts/api-lint/main.go:21,64-67), so
	// PATH-BASE-PREFIX already runs inside api-lint below on the same
	// v1 spec file. api-lint-e2e-base-path survives unchanged — api-lint
	// never touches internal-e2e.yaml, so that file's base-prefix rule still
	// needs its own gate.
	{
		ID:       "api-lint-e2e-base-path",
		Desc:     "PATH-BASE-PREFIX on the internal-e2e spec",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint", "-only", "PATH-BASE-PREFIX", "api/openapi/internal-e2e.yaml"},
		// scripts/api-lint/ is the check's own tool — see the comment on
		// api-lint-base-path-v1 above; this is the sole gate on
		// internal-e2e.yaml's base-prefix rule, so the gap was worse here.
		Paths: []string{"api/openapi/", "scripts/api-lint/"},
		CIJob: "ci.yml:verify",
		Fixture: &Fixture{
			Dir:          "api-lint-e2e-base-path",
			ArgvOverride: []string{"go", "run", "./scripts/api-lint", "-only", "PATH-BASE-PREFIX", "{{fixture}}/bad-spec.yaml"},
			Want:         "PATH-BASE-PREFIX",
		},
	},
	{
		ID:       "api-lint",
		Desc:     "full API design-system lint, strict",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./scripts/api-lint/", "-strict", "api/openapi/v1/openapi.yaml", "."},
		Paths:    []string{"api/openapi/", "scripts/api-lint/"},
		CIJob:    "ci.yml:verify",
		Fixture: &Fixture{
			Dir: "api-lint",
			// modulesRoot stays the real repo root: -strict's code rules read
			// scripts/api-lint/tripwire-allowlist.txt relative to it and treat
			// its absence as a hard error, which would fail the fixture for the
			// wrong reason. Only the spec argument is the bad input.
			ArgvOverride: []string{"go", "run", "./scripts/api-lint/", "-strict", "{{fixture}}/bad-spec.yaml", "."},
			Want:         "PATH-BASE-PREFIX",
		},
	},
	{
		ID:            "api-lint-selftest",
		FixtureWaiver: "this check IS a negative-fixture suite — scripts/api-lint/testdata plus exit_code_test.go already assert non-zero exit on bad specs. Fixturing it would fixture a fixture harness.",
		Desc:          "the api-lint tool's own tests",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "test", "./scripts/api-lint/...", "-count=1"},
		Paths:         []string{"scripts/api-lint/"},
		CIJob:         "ci.yml:verify",
	},
	{
		ID:            "contract-sync",
		FixtureWaiver: "TRANSITIONAL, no fixture yet. The check drives pwsh scripts that regenerate and compare per-module contract artifacts, which needs a sandbox carrying the generator toolchain. Owner: #87/A1 remainder — recorded in the Phase 1 handoff as a known A1.2 gap, not deferred to another axis.",
		Desc:          "spec/generated/runtime contract sync across modules",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pwsh", "-NoProfile", "-File", "./scripts/check-contract-sync-all.ps1"},
		// scripts/check-contract-sync-all.ps1 is the check's own definition;
		// without it a PR editing only the script (weakening what it checks)
		// selects zero checks (whole-branch review C2).
		Paths: []string{"api/openapi/", "internal/", "scripts/check-contract-sync-all.ps1"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "codegen-drift-backend",
		FixtureWaiver: "TRANSITIONAL, no fixture yet. Running it in a sandbox means running oapi-codegen and go generate there; the honest fixture is a recorded drift diff, which needs generator inputs the sandbox does not have. Owner: #87/A1 remainder.",
		Desc:          "go generate ./... produces no diff in generated Go artifacts (api.gen.go, httpsurface_gen.go, httpsurface_e2e_gen.go)",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"bash", "scripts/check-codegen-drift-backend.sh"},
		// scripts/check-codegen-drift-backend.sh is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"api/openapi/", "internal/", "apps/", "cmd/", "scripts/check-codegen-drift-backend.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "codegen-drift-frontend",
		FixtureWaiver: "TRANSITIONAL, no fixture yet. Same generator-toolchain problem as codegen-drift-backend, via pnpm run gen:api. Owner: #87/A1 remainder.",
		Desc:          "pnpm run gen:api produces no diff in frontend/apps/web/src/lib/api-types/",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"bash", "scripts/check-codegen-drift-frontend.sh"},
		// scripts/check-codegen-drift-frontend.sh is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"api/openapi/", "frontend/", "scripts/check-codegen-drift-frontend.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "openapi-lint-v1",
		FixtureWaiver: "third-party tool (@redocly/cli, pinned); a fixture here would test Redocly's rule engine, not this repo.",
		Desc:          "redocly lint on the v1 spec",
		Profiles:      []string{ProfilePR, ProfileFull},
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
		ID:            "openapi-lint-e2e",
		FixtureWaiver: "third-party tool (@redocly/cli, pinned); same as openapi-lint-v1.",
		Desc:          "redocly lint on the internal-e2e spec",
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
		ID:            "oasdiff-breaking",
		FixtureWaiver: "third-party tool (oasdiff); it also needs a base-branch spec materialized by a workflow step, so a sandbox run would fail on the missing file rather than on a breaking change.",
		Desc:          "oasdiff breaking-change gate: PR head spec vs base-branch spec, --fail-on ERR",
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
		ID:       "module-imports",
		Desc:     "cross-module access goes through published interfaces",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-module-boundaries.ps1"},
		// scripts/check-module-boundaries.ps1 is the check's own definition
		// (whole-branch review C2 class).
		Paths:   []string{"internal/", "scripts/check-module-boundaries.ps1"},
		CIJob:   "ci.yml:verify",
		Fixture: &Fixture{Dir: "module-imports", Want: "[module-imports] FAIL"},
	},
	{
		ID:       "test-conventions",
		Desc:     "new tests use the canonical framework for their class",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-test-discipline.sh"},
		Fixture:  &Fixture{Dir: "test-conventions", Want: "code_violation_integration_test.go"},
		// scripts/check-test-discipline.sh is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"internal/", "tests/", "apps/", "scripts/check-test-discipline.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "test-conventions-selftest",
		FixtureWaiver: "this check IS the negative-fixture harness for test-conventions (scripts/check-test-discipline-selftest.sh builds a throwaway repo and asserts finding counts).",
		Desc:          "check-test-discipline.sh reads code and ignores Go line comments",
		Profiles:      []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:          []string{"bash", "scripts/check-test-discipline-selftest.sh"},
		Paths:         []string{"scripts/check-test-discipline.sh", "scripts/check-test-discipline-selftest.sh", "scripts/testdata/test-discipline/"},
		CIJob:         "ci.yml:verify",
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
		// Deliberately no Paths. The check's own scan (trackedTestFiles in
		// testdbbypass.go) is `git ls-files "*_test.go"` — repo-wide, not
		// scoped to internal/, apps/, or tests/. A `_test.go` file under
		// cmd/ or scripts/ (both tracked, both outside those trees) can
		// bypass the factory exactly like one under internal/ can, so any
		// Paths list here would have to enumerate every directory that can
		// ever hold a `_test.go` file — a claim that silently rots as the
		// repo grows new ones. matchesPaths' documented default (no declared
		// Paths -> repo-scoped, always runs) is the fail-closed answer:
		// narrowing this check to a path set is a claim that nothing outside
		// it can break the check, and the repo-wide scan disproves that
		// claim on its face.
		CIJob: "ci.yml:verify",
		Fixture: &Fixture{
			Dir:  "testdb-bypass-guard",
			Want: "internal/fixture/bypass_test.go",
		},
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
		Fixture: &Fixture{
			Dir:  "adr-status",
			Want: "ADR status-field budget exceeded",
		},
	},
	{
		ID:       "wiki-debt-tally",
		Desc:     "every module doc's severity tally matches its tech-debt register",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/wiki-tally-check.ps1", "-All"},
		// scripts/wiki-tally-check.ps1 is the check's own definition
		// (whole-branch review C2 class).
		Paths:   []string{"wiki/modules/", "scripts/wiki-tally-check.ps1"},
		CIJob:   "ci.yml:verify",
		Fixture: &Fixture{Dir: "wiki-debt-tally", Want: "SWEEP FAIL (1/1): fixture"},
	},
	{
		ID:       "db-docs-coverage",
		Desc:     "every baseline table has a wiki dictionary page",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-db-dictionary-coverage.ps1"},
		// scripts/check-db-dictionary-coverage.ps1 is the check's own
		// definition (whole-branch review C2 class).
		Paths:   []string{"db/baseline/", "wiki/database/tables/", "scripts/check-db-dictionary-coverage.ps1"},
		CIJob:   "ci.yml:verify",
		Fixture: &Fixture{Dir: "db-docs-coverage", Want: "Missing dictionary pages"},
	},
	{
		ID:       "migration-gapless",
		Desc:     "db/migrations is a gapless sequence and no historical migration was edited after merge",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-migration-gapless.sh"},
		Needs:    []string{needsGitDepth},
		// scripts/check-migration-gapless.sh is the check's own definition —
		// named explicitly in the whole-branch review's C2 finding.
		Paths:   []string{"db/migrations/", "scripts/check-migration-gapless.sh"},
		CIJob:   "ci.yml:verify",
		Fixture: &Fixture{Dir: "migration-gapless", Want: "Gap: migration 0002 missing"},
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
		Fixture:  &Fixture{Dir: "governance-diff-rules", Want: "API contract change detected"},
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
		Fixture: &Fixture{
			Dir:  "invariant-coverage-map",
			Want: "Unmapped invariants found",
		},
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
		// The 64-finding backlog measured in
		// docs/superpowers/reports/2026-08-08-gosec-govulncheck-measurement.md
		// (Task 4) was triaged to zero in Task 9 (D9): every finding is either
		// a genuine fix (raw-string-comment corruption bugs, defer-based
		// sql.Rows.Close ordering, %w error wrapping) or a `// #nosec Gxxx --
		// reason` suppression on the exact flagged line, with the reason
		// stating the concrete fact that makes the code safe (a fixed literal
		// path/table/argv, a bounds-checked conversion, a computed
		// placeholder index never a value, an env var NAME not a credential).
		// `-nosec-require-rules -nosec-require-justification` enforce this
		// shape going forward: gosec itself now rejects a bare `#nosec` or one
		// missing a rule ID, so the triage can't silently rot back into an
		// unscoped blanket suppression. The two pre-existing `//nolint:gosec`
		// comments (tools/verify/main.go, tools/verify/audit.go,
		// tools/verify/testdbbypass.go) were converted to gosec-native
		// `// #nosec` syntax in the same pass — `//nolint:gosec` is
		// golangci-lint syntax; standalone gosec never read it, so those
		// suppressions were dead and the underlying findings were live under
		// a naive registration.
		//
		// PROMOTED to `pr` (A1.4, 2026-08-09). The blocker was never the
		// findings — it was that this check's CIJob was nightly.yml:
		// security-scan, which no needs: edge connects to ci.yml's required
		// gate. "Runs in some workflow" is not closure: a nightly failure
		// blocks nothing and merges anyway. Repointing the CIJob at
		// ci.yml:verify (already inside the required closure, already running
		// the `changed` profile) fixes that without adding a status context,
		// so it needs no branch-ruleset change — the earlier note tying
		// promotion to the ruleset swap conflated "new required job" with
		// "existing required job runs one more check".
		//
		// Promotion was gated on a live run, not on the prior triage note:
		// pinned gosec v2.28.0 reported 1 issue (G705 XSS-taint at
		// internal/platform/idempotency/middleware.go:186) whose suppression
		// was written as //nolint:gosec — golangci-lint syntax that standalone
		// gosec has never read, so the suppression was dead and the finding
		// live. Converted to a gosec-native `// #nosec G705 -- reason` in the
		// same commit; re-run is clean. Comment syntax only, no behaviour
		// change.
		//
		// -exclude-dir=.claude is load-bearing, not cosmetic: an unexcluded
		// scan walks into .claude/worktrees/<sibling>/, a sibling git worktree
		// with its own go.mod, inflating the true 176 import directories to
		// 333 (157, 47%, from the sibling worktree alone) — measured on this
		// machine both before and after the flag. `-no-fail` from the
		// measurement's own invocation is deliberately NOT carried over: this
		// registration must fail the check like any other, not just record it.
		//
		// Pinned at v2.28.0 (latest release as of 2026-08-09), not @latest: a
		// scanner that silently gains rules is a gate whose meaning changes with
		// no diff here to review — a new rule class turns a green branch red for
		// a reason nobody chose, and a withdrawn rule stops guarding just as
		// quietly. Bump deliberately. Audit rule A9 keeps it that way.
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "run", "github.com/securego/gosec/v2/cmd/gosec@v2.28.0", "-quiet", "-exclude-dir=.claude", "-nosec-require-rules", "-nosec-require-justification", "./..."},
		Needs:         []string{needsNetwork},
		FixtureWaiver: "third-party scanner (gosec, pinned); a fixture here would test gosec's own rule engine, not this repo. The repo-authored part is the #nosec justification shape, which -nosec-require-rules -nosec-require-justification enforce at every run.",
		// No Paths, deliberately: a security scan scoped by path is a security
		// scan that can be dodged by touching a file outside the list.
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "govulncheck",
		Desc: "no known-vulnerable symbol is reachable from any binary",
		// 19 total findings, but only 2 were called/reachable (GO-2026-5970 in
		// golang.org/x/text, GO-2026-5856 in stdlib crypto/tls) — both
		// call-graph-verified, not a naive dependency match. As of 2026-08-08
		// both are remediated (golang.org/x/text bumped to v0.39.0, Go
		// toolchain bumped to go1.26.5 — both versions confirmed as the fix
		// from govulncheck's own "Fixed in" output, not assumed): a fresh
		// govulncheck run reports 0 called vulnerabilities. This entry's
		// remediation criteria are met.
		//
		// PROMOTED to `pr` (A1.4, 2026-08-09), same reasoning as gosec above:
		// the CIJob moved from nightly.yml:security-scan (outside ci.yml's
		// required closure) to ci.yml:verify (inside it). Verified by a live
		// run at the pinned version before promoting: 0 called vulnerabilities,
		// 105s.
		//
		// Pinned at v1.6.0 (latest release as of 2026-08-09), not @latest — same
		// reasoning as gosec above. Note the pin fixes the *analyzer*, not the
		// vulnerability database: govulncheck queries vuln.go.dev at run time, so
		// newly disclosed CVEs still surface. That is the intended split — data
		// moves, tool does not.
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "run", "golang.org/x/vuln/cmd/govulncheck@v1.6.0", "./..."},
		Needs:         []string{needsNetwork},
		FixtureWaiver: "third-party scanner (govulncheck, pinned); its input is the live vulnerability database, so a synthetic bad fixture would assert against data this repo does not control.",
		// No Paths — same reasoning as gosec above.
		CIJob: "ci.yml:verify",
	},

	// ---- Frontend ---------------------------------------------------------
	{
		ID:            "eslint",
		FixtureWaiver: "third-party tool; the repo-authored part is the eigenpal import boundary rule, whose fixture belongs with the eslint config — deferred, no owner assigned outside A1.",
		Desc:          "eslint across the workspace, including the eigenpal import boundary",
		Profiles:      []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "run", "lint"},
		// package.json (root) is where the "lint" script this check invokes
		// is defined — without it here, a PR rewriting "lint" to a no-op
		// while touching no frontend/, packages/, apps/, or config file
		// disarms eslint and selects nothing that would notice (whole-branch
		// review C2 class).
		Paths: []string{"frontend/", "packages/", "apps/", "eslint.config.mjs", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "css-tokens",
		Desc:     "no new raw hex colors in module.css",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-css-token-discipline.sh"},
		Fixture: &Fixture{
			Dir:  "css-tokens",
			Want: "RAW-HEX",
		},
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
		Fixture:  &Fixture{Dir: "eigenpal-selector-pin"},
		// scripts/check-eigenpal-selector-pin.sh is the check's own
		// definition — named explicitly in the whole-branch review's C2
		// finding.
		Paths: []string{"frontend/", "packages/eigenpal-adapter/", "scripts/check-eigenpal-selector-pin.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "fe-typecheck",
		FixtureWaiver: "third-party tool (tsc); a fixture would prove TypeScript rejects bad types.",
		Desc:          "tsc over @metaldocs/web",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "--filter", "@metaldocs/web", "run", "typecheck"},
		// Root toolchain/dependency files, not just frontend/ and packages/:
		// a lockfile-only or Node-version-only PR (Dependabot's normal shape)
		// otherwise selects zero frontend checks while the pathless Go
		// checks still run, so `required` reports success over an unrun
		// frontend suite (whole-branch review C3).
		Paths: []string{"frontend/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "fe-test",
		FixtureWaiver: "a test suite (vitest); same as go-test-unit.",
		Desc:          "vitest over @metaldocs/web",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "--filter", "@metaldocs/web", "run", "test"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"frontend/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "docx-typecheck",
		FixtureWaiver: "third-party tool (tsc); same as fe-typecheck.",
		Desc:          "tsc over the docx-v2 workspace",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "run", "typecheck:docx-v2"},
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
		ID:            "docx-build",
		FixtureWaiver: "a build step, not a guard — it produces dist/meta.json for docx-test. It fails when the workspace does not build; there is no rule to feed bad input to.",
		Desc:          "the docx-v2 workspace builds; produces the dist/meta.json that docx-test's bundle guard reads",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "run", "build:docx-v2"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"apps/docx-renderer/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		// I8: same CIJob reasoning as docx-typecheck above.
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "docx-test",
		FixtureWaiver: "a test suite (vitest over docx-v2); same as go-test-unit.",
		Desc:          "docx-v2 unit tests",
		// D-14 (closed by R4): depends on docx-build having already run —
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
		// tree, with docx-build's `pnpm -r run build` writing
		// frontend/apps/web/dist underneath both. That race, not the
		// build-before-test ordering, was the reproducible source of
		// flaky/contradictory results between this check and fe-test. The npm
		// scripts now filter to `./packages/**` + `./apps/**` only, which does
		// not include frontend/apps/web — so docx-test and fe-test no
		// longer touch the same files at all.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"pnpm", "run", "test:docx-v2"},
		After:    []string{"docx-build"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"apps/docx-renderer/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		// I8: same CIJob reasoning as docx-typecheck above.
		CIJob: "ci.yml:verify",
	},

	// ---- Tests ------------------------------------------------------------
	{
		ID:            "go-test-unit",
		FixtureWaiver: "a test suite, not a guard: it fails when a test fails, which is the property, and every test in it is its own fixture.",
		Desc:          "go test ./... (no integration tag)",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "test", "-count=1", "-timeout", "600s", "./..."},
		CIJob:         "ci.yml:verify",
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
		ID:            "req-trace-selftest",
		FixtureWaiver: "same shape as api-lint-selftest: a Go test suite over scripts/req-trace, with its own testdata.",
		Desc:          "the req-trace tool's own tests",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "test", "./scripts/req-trace/...", "-count=1"},
		Paths:         []string{"scripts/req-trace/"},
		CIJob:         "ci.yml:verify",
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
		Fixture: &Fixture{
			Dir:          "req-trace",
			ArgvOverride: []string{"go", "run", "./scripts/req-trace", "-repo", "{{fixture}}"},
			Want:         "UNCOVERED MUST REQ(s) (1):",
		},
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
		Fixture: &Fixture{
			Dir:          "required-gate-selftest",
			CopyFromRepo: []string{"scripts/required-gate.jq"},
			Want:         "pass-mislabelled.json",
		},
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
		// verify-audit's fixture is a workflow that names an unknown check ID
		// (rule A1). Rule A7 itself is proven by the guard-fixtures entry
		// below being the thing A7 demands.
		Fixture: &Fixture{
			Dir:          "verify-audit",
			CopyFromRepo: []string{"scripts/required-gate.jq"},
			Want:         "no-such-check-id",
		},
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "guard-fixtures",
		Desc: "every blocking guard is fed bad input and must exit non-zero (--guard-fixtures)",
		// A1.2. Same argument as verify-audit directly above, applied one level
		// up: a negative-fixture spine that only runs when someone remembers to
		// type it proves nothing on the day a guard silently stops guarding.
		// Registered like any other check so CI runs it, so --audit sees it, and
		// so the required gate can depend on it.
		//
		// No Paths: a guard can be defanged from far outside its own directory
		// (a script it calls, a config it reads, a registry Argv edit), so
		// scoping this by path would reintroduce the reachability hole A1.4
		// exists to close.
		//
		// Not in ProfileFast: each fixture is a real subprocess against a real
		// sandbox. Parallel it is fast enough for PR and full, too slow for the
		// inner loop.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/verify", "--guard-fixtures"},
		// The harness cannot have a negative fixture of its own without infinite
		// regress. Its two failure modes were exercised by hand before it
		// landed, and both are covered structurally: a check whose command exits
		// 0 is reported as "the guard does not guard", and a check that fails for
		// the wrong reason is caught by Want. Audit rule A7 then makes the
		// coverage itself blocking — a new ProfilePR check with neither Fixture
		// nor FixtureWaiver fails verify-audit.
		FixtureWaiver: "this check IS the negative-fixture harness; fixturing it would be circular (owner: #87/A1)",
		CIJob:         "ci.yml:verify",
	},
}
