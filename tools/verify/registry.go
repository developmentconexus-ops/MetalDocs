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
	"oasdiff-breaking":                   "diffs the PR head spec against a base-branch spec materialized by a PR-only workflow step; on a tag the file does not exist",
	"governance-diff-rules":              "rules about what a PR's diff must contain (contract change ships a spec update, etc.); a tag has no diff to rule on",
	"migration-gapless":                  "\"no historical migration edited after merge\" is a property of a diff against origin/main",
	"eslint-suppression-baseline-growth": "compares eslint-suppressions.json against the merge base with origin/main; release.yml's checkout has no \"fetch base ref\" step (there is no PR base on a tag) so origin/main is not guaranteed to resolve, and even when it does, \"did this PR's diff grow the baseline\" is not a question a tag build can ask",
	"dead-code-baseline-growth":          "compares dead-code-baseline.json against the merge base with origin/main; same tag-has-no-PR-base reasoning as eslint-suppression-baseline-growth immediately above",
}

// The PR and full integration checks intentionally share one computed package
// universe. Only the race flag and the profile/CI owner differ; roots and tags
// must not drift into two silently different suites.
var integrationPartition = &Partition{
	Roots: []string{"./tests/...", "./internal/...", "./apps/..."},
	Tags:  "integration",
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

	// Partition declares that this check's subject is a set of Go test
	// packages that CI may split across runners. When set, Argv must spend
	// packagesPlaceholder exactly once, and the package list that replaces it
	// is resolved from `go list` at run time — never from a hand-kept list.
	// See partition.go for why that distinction is the whole point.
	Partition *Partition

	// Stage names a subject the verifier materialises before running Argv,
	// exposed to Argv as a placeholder. The only mode is stageTrackedTree
	// ({{tracked}} — the tracked tree at HEAD, from `git archive`), and it
	// exists because a check whose subject is "the working directory" has a
	// different subject on every machine (#87/A1 review F2). Empty means the
	// check runs against the repo as it stands, which is right for everything
	// that reads source files rather than cataloguing them.
	Stage string

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

	// FixtureWaiver records why this blocking check carries no negative
	// fixture. Audit rule A7 requires every blocking check to carry exactly
	// one of Fixture or FixtureWaiver, so a hole cannot exist unnamed.
	//
	// The waiver is a CLASSIFICATION, not prose. It used to be a free string,
	// and a free string is an escape hatch: three repo-authored guards carried
	// "TRANSITIONAL, no fixture yet" and A7 counted them as covered, which is
	// exactly the coverage claim the rule exists to prevent (#87/A1 review B3).
	// Kind is a closed enum of the three reasons a fixture cannot exist —
	// none of which is "not yet" — so an unfixtured repo-authored guard is now
	// unrepresentable rather than merely discouraged.
	FixtureWaiver *Waiver
}

// WaiverKind enumerates the only admissible reasons a blocking check has no
// negative fixture. Adding a fourth kind is a deliberate, reviewable act; a
// waiver that fits none of these means the check needs a fixture.
type WaiverKind string

const (
	// WaiverThirdParty — the failing behaviour under test belongs to a pinned
	// third-party tool. A fixture would assert that someone else's product
	// works, which this repo neither owns nor can fix.
	WaiverThirdParty WaiverKind = "third-party-tool"

	// WaiverTestSuite — the check IS a test suite or a negative-fixture
	// harness. Its bad-input proof is its own test cases; fixturing it would
	// fixture a fixture.
	WaiverTestSuite WaiverKind = "test-suite"

	// WaiverBuildStep — the check is a build prerequisite producing an
	// artifact, not a rule. It fails when the build fails; there is no bad
	// input to feed a rule that does not exist.
	WaiverBuildStep WaiverKind = "build-prerequisite"
)

// waiverKinds is the closed set A7 validates against.
var waiverKinds = map[WaiverKind]bool{
	WaiverThirdParty: true,
	WaiverTestSuite:  true,
	WaiverBuildStep:  true,
}

// Waiver is a classified absence of a fixture. Why is still required — the
// kind says which rule admits the waiver, Why says why THIS check is an
// instance of it — but Why can no longer smuggle in a fourth kind.
type Waiver struct {
	Kind WaiverKind
	Why  string
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
			Want: []string{"internal/bad/unformatted.go"},
		},
	},
	{
		ID:            "go-vet",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party tool (cmd/vet), not a guard this repo authored; its own failure modes are Go's to prove, and a fixture here would test the Go toolchain."},
		Desc:          "go vet ./...",
		Profiles:      []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:          []string{"go", "vet", "./..."},
		CIJob:         "ci.yml:verify",
	},
	{
		ID:            "go-vet-integration",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "same third-party tool as go-vet under a build tag; what this entry adds is the tag, which the registry declares rather than implements."},
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
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party linter aggregator (pinned); a fixture here would test golangci-lint's own analyzers, not this repo. The repo-authored part is .golangci.yml's enabled set, which every run exercises against the whole tree."},
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
		Desc:     "custom Go analyzers (hgcrossmodule, nosqltxindomain, platformboundary, txownership, legacyvocab, outboxpair, postcommitaudit, deliveryauditsink, nodualmode, noresponsemap, problemwriter, actorextraction) against the recorded baseline",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"go", "run", "./tools/cilint", "./..."},
		CIJob:    "ci.yml:verify",
		Fixture: &Fixture{
			// A4.0: approval mutating documents' base table. The sandbox has
			// no tools/cilint/baseline.json, and cilint fails on any finding
			// when no baseline is configured — so the fixture proves the
			// analyzer, not the ratchet.
			Dir: "arch-lint",
			Want: []string{
				`writes "documents"'s base table "documents"`,
				// B5: the sandbox also contains the exact path of a live READ
				// exemption (security/.../postgres/repository.go x audit_events)
				// which reads that table legally and then writes it. The write
				// must be reported. If a future edit made exemptions
				// verb-agnostic again, this line stops appearing and the fixture
				// fails — which is the only way "a read exemption is not a write
				// permit" is a fact rather than a comment.
				`writes "audit"'s base table "audit_events"`,
				// A3.1: the same sandbox carries a thirteenth local
				// writeProblem. Both halves of the ban are pinned, because
				// either one alone leaves the clone buildable — a serializer
				// that skips the media type is still a serializer, and a
				// hand-written envelope is still an envelope even if the
				// function never names *problem.Problem.
				`takes both an http.ResponseWriter and a *problem.Problem`,
				`naming it here means a second error envelope`,
				// A3.2 gate C3: the same sandbox carries a second clone written
				// against import ALIASES (h "net/http", p ".../problem"). The
				// function name is pinned because the two unqualified Want
				// lines above are already satisfied by the unaliased clone —
				// without naming it, an analyzer that went back to matching the
				// identifiers "http" and "problem" would still pass this
				// fixture. That file names no media type, so the only rule that
				// can report it is the signature rule resolving by import path.
				`function writeAliasedProblem takes both an http.ResponseWriter and a *problem.Problem`,
				// A3.3: the same sandbox carries the three actor-extraction
				// fixtures. All three are pinned separately because each one
				// alone leaves the fail-open shape reachable.
				//
				// (a) the low-level accessor called under its normal import.
				`the low-level identity-storage accessor that returns "" for a missing actor (A3.3)`,
				// (b) the SAME call under an import alias. Pinned by file name
				// because the message is identical to (a)'s, so without naming
				// the file an analyzer that went back to matching the
				// identifier "iamdomain" would still satisfy the line above.
				// Only a rule resolving by import PATH reports this file.
				`aliased_lowlevel_actor.go`,
				// (c) both discard forms of the CANONICAL accessor. Rule 1
				// without Rule 2 is one underscore wide: banning the fail-open
				// accessor accomplishes nothing if the fail-closed one can be
				// called and its answer thrown away. The bool-returning and
				// error-returning siblings are pinned separately so removing
				// either from the analyzer's table fails the fixture.
				`discards the presence bool of authn.UserIDFromContext (A3.3)`,
				`discards the error of authn.RequireUserID (A3.3)`,
				// A3.3 enforcement round (#108 review 4902506890). Each line
				// below is a shape the first implementation let through, and
				// each is pinned by FILE because the messages are shared with
				// the fixtures above — without the file name an analyzer that
				// dropped the widening would still satisfy the message lines.
				//
				// E1: the accessor referenced without being called. A CallExpr-
				// only rule reads `extract := iamdomain.UserIDFromContext` as an
				// assignment and reports nothing, and the call one line later is
				// then invisible to it.
				`indirect_lowlevel_actor.go`,
				// E2: dot imports of both protected paths. A dot import deletes
				// the qualifier that Rules 1 and 2 resolve, so the ban has to be
				// on the ImportSpec. Both paths are pinned: banning only the
				// low-level one leaves the canonical one dot-importable, and the
				// discarded underscore comes straight back.
				`dot-imports metaldocs/internal/modules/iam/domain`,
				`dot-imports metaldocs/internal/platform/authn`,
				// E3: the discard written as a declaration. `var x, _ = f(ctx)`
				// is a ValueSpec, not an AssignStmt — one binding, two node
				// types, and a matcher written against the assignment form alone
				// has a spelling that turns it off.
				`declaration_discard.go`,
				// E4: suppression has to be a real directive with a real reason.
				// A bare directive is the invariant switched off with nothing for
				// a reviewer to disagree with; the same characters inside a
				// string literal are data that a raw-line scan cannot tell from
				// code. Both files still carry a violation and both must still be
				// reported. (That a WELL-FORMED directive does suppress is the
				// complementary half, proven in
				// tools/cilint/internal/analyzers/actorextraction_test.go — an
				// absence is not assertable through Want.)
				`bare_suppression.go`,
				`string_literal_suppression.go`,
				// #108 review round 1, finding 3: the fresh gap AFTER the E1 alias
				// fix, on the OTHER accessor. E1 taught Rule 1 to follow a reference
				// to the low-level accessor through a local alias; this file proves
				// Rule 2 has the identical hole on the canonical, fail-closed
				// accessor: `extract := authn.RequireUserID` binds it to a local, and
				// `actor, _ := extract(ctx)` discards its error with call.Fun as a
				// bare *ast.Ident that isSelector cannot resolve. Pinned by FILE name
				// because the message text is shared with ignored_presence.go's
				// direct-call fixture — without the file name, an analyzer that
				// regressed to matching only `authn.RequireUserID(ctx)` verbatim would
				// still satisfy the message lines above.
				`indirect_presence_discard.go`,
			},
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
		Fixture: &Fixture{Dir: "problem-codes-drift", Want: []string{"problem-codes-dump"}},
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
			Want:         []string{"PATH-BASE-PREFIX"},
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
			Want:         []string{"PATH-BASE-PREFIX"},
		},
	},
	{
		ID:            "api-lint-selftest",
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "this check IS a negative-fixture suite — scripts/api-lint/testdata plus exit_code_test.go already assert non-zero exit on bad specs. Fixturing it would fixture a fixture harness."},
		Desc:          "the api-lint tool's own tests",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "test", "./scripts/api-lint/...", "-count=1"},
		Paths:         []string{"scripts/api-lint/"},
		CIJob:         "ci.yml:verify",
	},
	{
		ID: "contract-sync",
		Fixture: &Fixture{
			// The earlier waiver claimed this check "regenerates artifacts" and
			// so needed a generator toolchain in the sandbox. Reading the script
			// says otherwise: it compares surfaces that already exist — spec
			// path keys, generated operation ids, the FE wrapper, the wiki
			// status line. Nothing is regenerated, so nothing blocks a fixture.
			//
			// The tree carries all four gated modules with their surfaces
			// aligned, then removes one operation id from documents' generated
			// package — the exact shape of "the spec grew a route and the
			// generated boundary did not". Three modules stay clean, so the
			// fixture also proves the guard localises the drift rather than
			// failing wholesale.
			Dir:          "contract-sync",
			CopyFromRepo: []string{"scripts/check-module-contract-sync.ps1"},
			Want: []string{
				"[DRIFT] generated backend package presence",
				"drifted module(s): documents",
			},
		},
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
		ID: "codegen-drift-backend",
		Fixture: &Fixture{
			// The property is "running the generator changes a committed
			// generated file", and that is generator-agnostic: the sandbox is a
			// tiny module whose //go:generate directive rewrites api.gen.go with
			// one more operation id than the committed copy. oapi-codegen is not
			// needed to prove the guard notices — that was the mistake in the
			// waiver this replaces, which treated the repo's particular
			// generator as part of the property.
			Dir:  "codegen-drift-backend",
			Want: []string{"Run 'go generate ./...' and commit"},
		},
		Desc:     "go generate ./... produces no diff in generated Go artifacts (api.gen.go, httpsurface_gen.go, httpsurface_e2e_gen.go)",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-codegen-drift-backend.sh"},
		// scripts/check-codegen-drift-backend.sh is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"api/openapi/", "internal/", "apps/", "cmd/", "scripts/check-codegen-drift-backend.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID: "codegen-drift-frontend",
		Fixture: &Fixture{
			// Same shape as codegen-drift-backend: the sandbox provides a
			// frontend/apps/web with a real `gen:api` script that rewrites
			// src/lib/api-types/index.d.ts, and a committed copy that predates
			// it. openapi-typescript is not part of the property either.
			Dir:  "codegen-drift-frontend",
			Want: []string{"Run 'pnpm run gen:api' in frontend/apps/web and commit"},
		},
		Desc:     "pnpm run gen:api produces no diff in frontend/apps/web/src/lib/api-types/",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-codegen-drift-frontend.sh"},
		// scripts/check-codegen-drift-frontend.sh is the check's own
		// definition (whole-branch review C2 class).
		Paths: []string{"api/openapi/", "frontend/", "scripts/check-codegen-drift-frontend.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "docker-build",
		FixtureWaiver: &Waiver{Kind: WaiverBuildStep, Why: "a production image build; it fails when the affected Dockerfile cannot build, and there is no rule to feed bad input to."},
		Desc:          "affected production Dockerfiles build without pushing images",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"bash", "scripts/check-docker-build.sh"},
		Needs:         []string{needsDocker, needsNetwork},
		// The script owns artifact selection from the diff; this broad subject
		// list makes changes to the selector, Docker inputs, or the verifier
		// itself select the check rather than allowing its own definition to
		// change without exercising it.
		Paths: []string{
			"deploy/docker/", "frontend/apps/web/", "apps/docx-renderer/", "packages/",
			"apps/api/", "apps/worker/", "apps/jobs/", "internal/", "db/",
			"go.mod", "go.sum", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc",
			".dockerignore", "scripts/check-docker-build.sh", "tools/verify/",
		},
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "dockerfile-go-version-drift",
		Desc: "every Dockerfile's golang builder stage is >= go.mod's go directive",
		// Trunk-health blocker (2026-08-11): deploy/docker/*.Dockerfile hardcoded
		// `FROM golang:1.25-alpine` while go.mod's `go` directive had moved to
		// 1.26.5 (GO-2026-5856), and nothing forced the two to agree — no CI job
		// builds container images, so `go mod download` was the first thing to
		// notice, inside the image build itself. Same shape as
		// problem-codes-drift/codegen-drift-*: a value restated instead of
		// derived (docs/engineering/defect-class-catalog.md Class 2,
		// Hand-Synced Enumerations). Derivation is not available here — a
		// Dockerfile FROM line cannot read go.mod — so this check is the
		// prevention rung instead (rung 3: regenerate-and-diff's sibling for a
		// value that can only be compared, not generated).
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-dockerfile-go-version.sh"},
		// scripts/check-dockerfile-go-version.sh is the check's own definition
		// (whole-branch review C2 class); go.mod is its source of truth; the
		// Dockerfiles are its subject, discovered by `git ls-files` rather than
		// hand-listed. No Paths, deliberately: a hand-maintained prefix list is
		// a selection hole waiting to happen — a stale Dockerfile added at a
		// path the list didn't anticipate would never select this check under
		// `changed`, even though the check's own git-ls-files discovery would
		// have caught it once it ran (found live, 2026-08-11 review on #114:
		// frontend/apps/web/Dockerfile.scratch-proof matched none of the old
		// Paths prefixes and the check silently skipped). Same reasoning as
		// governance-diff-rules below, which reads git diff itself for the
		// same reason. Repo-scoped is the safe direction (see matchesPaths).
		CIJob: "ci.yml:verify",
		Fixture: &Fixture{
			Dir: "dockerfile-go-version-drift",
			// One property per line, and every parser shape the check claims to
			// handle gets its own. Exit-code-only would be vacuous here: the
			// first fixture alone already fails the run, so three fixtures
			// covering lowercase/second-stage/--platform could be added, the
			// parser fix reverted, and the harness would still report "ok"
			// (measured 2026-08-11 on #114 — the pre-fix guard passed with all
			// four fixtures present). Line numbers are deliberately excluded
			// from the match: they pin the fixture's layout, not the rule.
			Want: []string{
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/worker.Dockerfile pins golang:1.25 but go.mod requires go >= 1.26.5",
				// lowercase `from` — Dockerfile instructions are not case-sensitive.
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/lowercase.Dockerfile pins golang:1.20 but go.mod requires go >= 1.26.5",
				// a SECOND builder stage, indented, below a compliant first one:
				// the defect the old head-1 parser could not see at all.
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/multistage.Dockerfile pins golang:1.24 but go.mod requires go >= 1.26.5",
				// `FROM --platform=<p> golang:...` — the flag sits between the
				// instruction and the image reference.
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/platform.Dockerfile pins golang:1.22 but go.mod requires go >= 1.26.5",
				// A non-numeric tag must be REPORTED, not merely fatal. Under
				// `set -euo pipefail` the version-extracting pipeline exits 1
				// when it matches nothing, which aborted the script before its
				// own error message could run -- a silent non-zero exit with no
				// output, which in CI is a red check nobody can diagnose. This
				// fixture sorts early (`latest` < `lowercase`), so a regression
				// that restores the abort drops every Want below it too.
				"could not parse a numeric golang version from: FROM golang:latest",
				// A continued FROM (backslash, default escape char): the
				// golang: reference sits on the physical line AFTER `FROM`,
				// invisible to a same-line match (CodeRabbit review on #114,
				// scripts/check-dockerfile-go-version.sh:113) -- this must be
				// refused, not silently skipped.
				"this check cannot parse a continued FROM instruction (line ends with a line-continuation character)",
				// A `# escape=` parser directive can swap the continuation
				// character to a backtick, same bypass shape via a different
				// escape char -- must also be refused, not silently skipped.
				"found a '# escape=' parser directive",
				// A digest-pinned golang stage (`FROM golang@sha256:<hex>`):
				// normal supply-chain practice, and the OLD `grep -qiE
				// 'golang:'` gate never saw it -- no `golang:` substring in a
				// digest reference -- so it was skipped with a bare
				// `continue`, no diagnostic, `checked` never incremented
				// (independent review on #114). The image-reference/tag split
				// now recognises the stage first and only then finds no
				// numeric tag to compare, so this must surface as the
				// existing loud "could not parse" failure, not silence.
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/digest.Dockerfile:1: could not parse a numeric golang version from: FROM golang@sha256:8728bc8be765db56dd0dd650b1b31a0396b03cd4e46689dc0c3e2bc4de3ad587 AS builder",
				// A legitimately-spelled lowercase filename, `legacy.dockerfile`:
				// the OLD case-sensitive `git ls-files` pathspec (`*Dockerfile*`,
				// capital D) never returned it at all -- not excluded, not
				// skipped, simply absent from `dockerfiles`, so a stale golang:1.19
				// pin here was completely unexamined and the run still exited 0
				// (independent review on #114). Discovery is now basename-scoped
				// AND case-insensitive (`:(glob,icase)**/*.Dockerfile`,
				// `:(glob,icase)**/Dockerfile`); this proves the fix actually
				// reaches a lowercase-named file -- see
				// scripts/check-dockerfile-go-version.sh's header for the full
				// account of why basename-scoping (not just case-insensitivity)
				// was the actual fix.
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/legacy.dockerfile pins golang:1.19 but go.mod requires go >= 1.26.5 (line 1)",
				// FROM $VAR / FROM ${VAR}: a build variable standing in for the
				// WHOLE image reference, not just its tag. The OLD repo-component
				// match saw no literal `golang` anywhere on the line -- the token
				// IS the variable reference -- so it fell through the golang-stage
				// test and was skipped with a bare `continue`, no diagnostic,
				// `checked` never incremented (independent review on #114, second
				// round): `ARG BASE_IMAGE=golang:1.10` + `FROM $BASE_IMAGE` reported
				// "all OK" and exited 0. Contrast with a variable TAG
				// (`FROM golang:${GO_VERSION}`), which this check already catches
				// correctly and is deliberately absent from this fixture tree --
				// adding one that passed would prove nothing, since nothing here
				// claims that shape is a defect.
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/argvar.Dockerfile:2: FROM's image reference is a build variable standing in for the WHOLE image ($BASE_IMAGE), not just a tag",
				// The gitlink itself (see the Gitlinks field above): a path
				// `git ls-files` returns that is not a readable regular file on
				// disk used to hit a bare `[[ -f "$df" ]] || continue` with no
				// diagnostic at all, before the per-file loop ever opened the
				// path (independent review on #114, second round).
				"DOCKERFILE-GO-VERSION-DRIFT: deploy/docker/phantom.Dockerfile: git tracks this path but it is not a readable regular file in this checkout",
			},
			// The scope exclusion is a rule, and this is its firing mechanism.
			// The fixture tree carries vendor/go.opentelemetry.io/otel/
			// dependencies.Dockerfile pinning golang:1.19 -- drifted on purpose,
			// so it WOULD be reported if discovery stopped excluding vendor/.
			// Want cannot express this: dropping the exclusion adds a line to
			// the output, and every existing Want still matches, so the harness
			// would report ok on a check that had started failing PRs over
			// upstream code nobody here can edit (`go mod vendor` overwrites it).
			NotWant: []string{
				"vendor/go.opentelemetry.io/otel/dependencies.Dockerfile",
			},
			// A discovered path that is git-tracked but not a readable
			// regular file (a submodule gitlink; a dangling symlink was not
			// constructible on the Windows checkout this fixture was authored
			// on -- core.symlinks=false there never materializes a real
			// symlink node, so it cannot prove this defect on every checkout
			// this harness runs on; a gitlink hits the identical `[[ -f ]]`
			// failure platform-independently, see the Gitlinks field
			// doc-comment in tools/verify/fixtures.go) used to vanish with no
			// diagnostic at the top of the per-file loop, before any content
			// was read (independent review on #114, second round).
			Gitlinks: []string{"deploy/docker/phantom.Dockerfile"},
		},
	},
	{
		ID:            "openapi-lint-v1",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party tool (@redocly/cli, pinned); a fixture here would test Redocly's rule engine, not this repo."},
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
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party tool (@redocly/cli, pinned); same as openapi-lint-v1."},
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
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party tool (oasdiff); it also needs a base-branch spec materialized by a workflow step, so a sandbox run would fail on the missing file rather than on a breaking change."},
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
		//
		// --err-ignore points at a repo-tracked, diff-reviewable allowlist
		// (api/openapi/v1/oasdiff-err-ignore.txt) — the honest waiver
		// channel for a disclosed, intentional breaking change, same audit
		// shape as scripts/check-governance-waivers.txt: each line is the
		// tool's own exact output text for one reviewed finding, not a
		// blanket suppression. An undisclosed or newly-introduced breaking
		// change still fails this gate, because its text won't match any
		// line in the file.
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"oasdiff", "breaking", "/tmp/openapi.base.yaml", "api/openapi/v1/openapi.yaml", "--err-ignore", "api/openapi/v1/oasdiff-err-ignore.txt", "--fail-on", "ERR"},
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
		Fixture: &Fixture{Dir: "module-imports", Want: []string{"[module-imports] FAIL"}},
	},
	{
		ID:       "test-conventions",
		Desc:     "new tests use the canonical framework for their class",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-test-discipline.sh"},
		Fixture:  &Fixture{Dir: "test-conventions", Want: []string{"code_violation_integration_test.go"}},
		// scripts/check-test-discipline.sh is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"internal/", "tests/", "apps/", "scripts/check-test-discipline.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "test-conventions-selftest",
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "this check IS the negative-fixture harness for test-conventions (scripts/check-test-discipline-selftest.sh builds a throwaway repo and asserts finding counts)."},
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
			Want: []string{"internal/fixture/bypass_test.go"},
		},
	},
	{
		ID:       "idempotency-identity-scope-guard",
		Desc:     "no actorFromCtx-shaped closure (func(context.Context)(string,string,error)) outside internal/platform/idempotency/identity.go calls tenant.FromContext directly (#90/A3.5 adoption guard)",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-idempotency-identity-scope.sh"},
		// Deliberately no Paths, same reasoning as testdb-bypass-guard above:
		// the script's own scan is `git ls-files '*.go'` repo-wide, not scoped
		// to internal/. A hand-rolled actorFromCtx-shaped resolver could be
		// added under apps/ or tests/ just as easily as under internal/, and
		// the whole point of this guard is that it does not trust a caller to
		// know where the next violation will land.
		CIJob: "ci.yml:verify",
		Fixture: &Fixture{
			Dir: "idempotency-identity-scope-guard",
			Want: []string{
				"internal/modules/fixture/bad_handler.go",
				// Locks the PR #122 review fix (chatgpt-codex-connector,
				// script line 149): an ungrouped `import "path"` line used to
				// resolve to alias "import" instead of "tenant", making the
				// scanner blind to this exact shape. See the fixture file's
				// own header for the reproduction.
				"internal/modules/fixture/bad_handler_ungrouped_import.go",
			},
			// NotWant proves the identity.go exclusion itself is guarded, not
			// merely conventional: the fixture tree plants, at exactly
			// internal/platform/idempotency/identity.go, a copy of the real
			// function shape (actorFromCtx signature, calls tenant.FromContext)
			// that WOULD fire this guard everywhere else in the tree. If the
			// script's exclusion of that one path is ever deleted, the check's
			// output starts naming this path — NotWant catches that the moment
			// it happens, rather than relying on nobody removing the `grep -v`.
			NotWant: []string{"internal/platform/idempotency/identity.go"},
		},
	},

	// ---- Governance -------------------------------------------------------
	{
		ID:       "adr-status",
		Desc:     "no ADR status block exceeds its line/char budget",
		Profiles: []string{ProfileFast, ProfileFull},
		Argv:     []string{"bash", "scripts/check-adr-status.sh"},
		// scripts/check-adr-status.sh is the check's own definition
		// (whole-branch review C2 class).
		Paths: []string{"wiki/decisions/", "scripts/check-adr-status.sh"},
		CIJob: "nightly.yml:governance-hygiene",
		Fixture: &Fixture{
			Dir:  "adr-status",
			Want: []string{"ADR status-field budget exceeded"},
		},
	},
	{
		ID:       "wiki-debt-tally",
		Desc:     "every module doc's severity tally matches its tech-debt register",
		Profiles: []string{ProfileFast, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/wiki-tally-check.ps1", "-All"},
		// scripts/wiki-tally-check.ps1 is the check's own definition
		// (whole-branch review C2 class).
		Paths:   []string{"wiki/modules/", "scripts/wiki-tally-check.ps1"},
		CIJob:   "nightly.yml:governance-hygiene",
		Fixture: &Fixture{Dir: "wiki-debt-tally", Want: []string{"SWEEP FAIL (1/1): fixture"}},
	},
	{
		ID:       "db-docs-coverage",
		Desc:     "every baseline table has a wiki dictionary page",
		Profiles: []string{ProfileFast, ProfileFull},
		Argv:     []string{"pwsh", "-NoProfile", "-File", "./scripts/check-db-dictionary-coverage.ps1"},
		// scripts/check-db-dictionary-coverage.ps1 is the check's own
		// definition (whole-branch review C2 class).
		Paths:   []string{"db/baseline/", "wiki/database/tables/", "scripts/check-db-dictionary-coverage.ps1"},
		CIJob:   "nightly.yml:governance-hygiene",
		Fixture: &Fixture{Dir: "db-docs-coverage", Want: []string{"Missing dictionary pages"}},
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
		// The script is accumulate-then-exit (fail=0 up front, every
		// violation printed and flagged, one `exit 1` at the bottom -- see
		// the script's own header comment), so a single execution proves
		// ALL of its properties at once, each pinned by its own Want/
		// NotWant entry below. Each Want is satisfiable by exactly one
		// fixture file, so none of them can rot silently behind another.
		//
		// Tree (layered, base + head + head2 + base2 — see Fixture.Dir doc
		// comment for the head2 and base2 mechanisms):
		//   base:  0001 v1                                            0005 v1
		//   head:  0001 v2 (edited)   0003 v1 (new)   0004 v1 (new)
		//   head2:                                     0004 v2 (edited again)
		//   base2:                                                    0005 v2 (edited by origin/main, post-fork)
		// head2 is a second commit on the BRANCH (child of head); base2 is
		// a sibling commit on origin/main itself (child of base, not of
		// head/head2) -- see advanceOriginMain's doc comment for why that
		// distinction needs git plumbing, not a second `git commit`.
		// 0002 never exists in any layer -- that absence is itself the
		// gap-detection fixture; 0003 fills out the range above the gap so
		// the sequence check has something to range over without also
		// becoming a historical-edit or false-positive candidate.
		//
		// Property 1 -- gapless sequence (Want "Gap: migration 0002
		// missing"): 0001, 0003, 0004, 0005 exist, 0002 does not -- a real
		// gap, must always fire.
		//
		// Property 2 -- historical edit, true positive (Want
		// "db/migrations/0001_first.sql"): 0001 exists on origin/main
		// (base) and was edited afterward -- a real violation of "an
		// already-applied migration must never change" -- must still fire.
		// This is the proof the base-existence precondition does not gut
		// the guard.
		//
		// Property 3 -- historical edit, false positive #1, branch-only
		// (NotWant "db/migrations/0004_fourth.sql"): 0004 does NOT exist
		// on origin/main; it was added and then edited AGAIN, both within
		// the same branch (head -> head2) -- exactly PR #113's shape.
		// Pre-fix, `git log --diff-filter=M` over a range starting at
		// origin/$BASE's tip reports 0004 as Modified too (head2's diff
		// against head), even though it never touched origin/main.
		// Post-fix it must never be named -- this is the proof that false
		// positive is gone.
		//
		// Property 4 -- historical edit, false positive #2, base-only
		// (NotWant "db/migrations/0005_fifth.sql"): 0005 DOES exist on
		// origin/main at the merge base, but only origin/main's own
		// post-fork commit (base2) edits it -- this branch (head/head2)
		// never touches it. Pre-fix, the SYMMETRIC `origin/$BASE...HEAD`
		// range walks base2 too (reachable from origin/main, not from
		// HEAD), and since 0005 genuinely exists at the merge base the
		// base-existence check does not save it -- flagged as a violation
		// on a branch that made no such edit. This is the CodeRabbit
		// finding from PR #123's own review: the identical defect as
		// Property 3, the other lineage. Post-fix, the ASYMMETRIC
		// `$MERGE_BASE..HEAD` range never reaches base2 at all, since it
		// only walks commits reachable from HEAD.
		Fixture: &Fixture{
			Dir: "migration-gapless",
			Want: []string{
				"Gap: migration 0002 missing",
				"db/migrations/0001_first.sql",
			},
			NotWant: []string{
				"db/migrations/0004_fourth.sql",
				"db/migrations/0005_fifth.sql",
			},
		},
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
		Fixture:  &Fixture{Dir: "governance-diff-rules", Want: []string{"API contract change detected"}},
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
			Want: []string{"Unmapped invariants found"},
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
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party scanner (gosec, pinned); a fixture here would test gosec's own rule engine, not this repo. The repo-authored part is the #nosec justification shape, which -nosec-require-rules -nosec-require-justification enforce at every run."},
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
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party scanner (govulncheck, pinned); its input is the live vulnerability database, so a synthetic bad fixture would assert against data this repo does not control."},
		// No Paths — same reasoning as gosec above.
		CIJob: "ci.yml:verify",
	},
	{
		ID:   "secret-scan",
		Desc: "no secret is committed anywhere in this repo's git history",
		// MOVED OUT OF YAML (#87/A1 review B1). This gate used to exist only as
		// a `docker run ghcr.io/gitleaks/gitleaks ...` step inside
		// ci.yml:security, which made ci.yml the second definition of a
		// blocking property: `verify --audit` could not see it, a local
		// `verify --profile=pr` did not run it, and a change to what the gate
		// accepts was reviewable only as workflow YAML. The property now lives
		// here, like every other blocking property, and CI calls the verifier.
		//
		// Run from source rather than from the container the YAML step used:
		// `go run <module>@<version>` is the same shape gosec and govulncheck
		// already use, it needs no docker daemon (so a Windows dev machine can
		// run the same command CI runs), and a module version is as immutable
		// as an image digest. The version is the exact one the retired YAML
		// step pinned — v8.24.3 — so this move changes where the gate is
		// defined, not what it accepts.
		//
		// --config .gitleaks.toml is load-bearing: the repo-authored allowlist
		// is what makes the difference between this scan and a stock gitleaks
		// run, and it is the part a fixture must hold honest (see Fixture).
		Profiles: []string{ProfilePR, ProfileFull},
		Argv: []string{
			"go", "run", "github.com/zricethezav/gitleaks/v8@v8.24.3",
			"detect", "--source", ".", "--config", ".gitleaks.toml",
			"--redact", "-v", "--exit-code", "1",
		},
		// needsGitDepth is not decoration: `detect` walks git history, so on a
		// shallow clone it scans a truncated history and reports "no leaks
		// found" over commits it never read — a green that means nothing.
		Needs: []string{needsNetwork, needsGitDepth},
		// No Paths, deliberately: a secret scan scoped by path is a secret
		// scan that can be dodged by committing the secret somewhere else.
		Fixture: &Fixture{
			Dir: "secret-scan",
			// The repo's own .gitleaks.toml, byte-identical — the fixture
			// proves THIS config still detects, not that some other config
			// would. The fixture's planted credential sits at a path the
			// allowlist does not cover, so an allowlist edit that widens far
			// enough to disarm the scan turns this fixture red.
			CopyFromRepo: []string{".gitleaks.toml"},
			Want:         []string{"aws-access-token", "Finding:"},
		},
		CIJob: "ci.yml:security",
	},
	{
		ID:   "vuln-scan",
		Desc: "no dependency carries a known high-or-critical vulnerability",
		// MOVED OUT OF YAML (#87/A1 review B1), same reason as secret-scan
		// above: this was `uses: anchore/scan-action` with `fail-build: true`,
		// a blocking gate defined in workflow YAML and invisible to the
		// registry. ci.yml:security now calls the verifier and keeps only the
		// SARIF upload, which is reporting, not the gate.
		//
		// Container rather than `go run`, unlike secret-scan: building grype
		// from source costs 9m11s cold on this machine (measured), against a
		// pulled image that runs the scan in a fraction of that. The image is
		// digest-pinned, not tagged (A9) — a tag is a movable pointer, and a
		// scanner that silently changes version is a gate whose meaning
		// changes with no diff here to review. The digest is anchore/grype
		// v0.116.1.
		//
		// --fail-on high reproduces the retired step's severity-cutoff: high.
		// The second -o writes SARIF for ci.yml:security's upload step; the
		// file is gitignored because it is a run artifact, not a source.
		//
		// THE SUBJECT IS THE TRACKED TREE, NOT THE WORKING DIRECTORY (#87/A1
		// review F2). `dir:/src` used to be the repo as it sits on disk, which
		// gave the gate a different subject on every machine: CI runs a bare
		// checkout, a developer's tree also has node_modules, build outputs,
		// .gocache/.gomodcache/.pnpm-store and sibling worktrees under .claude/.
		// Those are not this repo's dependencies, and grype does not know that —
		// a stale metaldocs-api.exe reported x/crypto v0.31.0 and stdlib
		// go1.26.1 as HIGH while go.mod pins x/crypto v0.53.0, and cached build
		// objects from a sibling branch reported four more. The cure was a
		// fifteen-line --exclude list, which is a workaround with a maintenance
		// tail: every new gitignored directory needs another line, and the first
		// one anybody forgets makes local and CI mean different things again.
		//
		// Stage: stageTrackedTree makes the subject a pure function of the
		// commit. The verifier extracts `git archive HEAD` into a scratch
		// directory and hands grype THAT — no untracked file, no ignored file,
		// no cache, no build output, on any machine. Every --exclude line is
		// gone because nothing is left to exclude, and CI's scan scope is
		// unchanged: a bare checkout of HEAD is exactly what is now scanned
		// everywhere. See stage.go for the mechanism and stage_test.go for the
		// proof (a gitignored artifact cannot enter the scan; a tracked file
		// always does).
		//
		// Two mounts, and the split is load-bearing: /src is the staged tree
		// (read subject) and /out is the real repo root, so the SARIF lands in
		// the workspace where ci.yml:security's upload step looks for it rather
		// than inside a directory that is deleted the moment the check ends.
		//
		// The deliberate trade-off: an UNCOMMITTED dependency bump is not
		// scanned, because it is not in HEAD. That is the same answer CI gives.
		//
		// The named volume caches the vulnerability database across runs. With
		// --rm and no volume, every invocation re-downloads it — the same cost
		// the retired step avoided with the Action's cache-db: true.
		Profiles: []string{ProfilePR, ProfileFull},
		Stage:    stageTrackedTree,
		Argv: []string{
			"docker", "run", "--rm",
			"-v", trackedTreePlaceholder + ":/src",
			"-v", repoRootPlaceholder + ":/out",
			"-v", "metaldocs-grype-db:/root/.cache/grype",
			"anchore/grype@sha256:1e71065c0a4cff3e6bd3b8add525ffac4343eb4971694eb90a31cf6d4d3e85db",
			"dir:/src", "--fail-on", "high",
			"-o", "table", "-o", "sarif=/out/.grype.sarif",
		},
		Needs:         []string{needsDocker, needsNetwork},
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party scanner (grype, digest-pinned); its verdict is a join of this repo's dependency manifests with an externally maintained vulnerability database, so a synthetic bad fixture would assert against data this repo does not control — the same reason govulncheck is waived."},
		// No Paths — same reasoning as gosec and secret-scan above.
		CIJob: "ci.yml:security",
	},

	// ---- Frontend ---------------------------------------------------------
	{
		ID:            "eslint",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "eslint, a pinned third-party engine; a fixture here would assert that eslint reports a rule violation. The one repo-authored rule in that config, the eigenpal import boundary, has its own registry check (eigenpal-selector-pin) and its own fixture."},
		Desc:          "eslint across the workspace, including the eigenpal import boundary",
		Profiles:      []string{ProfileFast, ProfilePR, ProfileFull},
		// A2.1 review round 3 (Finding 1, CRITICAL): this used to be
		// []string{"pnpm", "run", "lint"} — indirection through package.json's
		// "lint" script. Nothing pinned or content-checked what that script
		// actually ran, so a one-line, easily-overlooked package.json diff
		// appending "--suppress-all" made ESLint itself silently absorb any
		// new violation into eslint-suppressions.json on disk and exit 0 —
		// defeating this check's entire purpose while it kept reporting PASS.
		// Reproduced live in the cold review of this PR.
		//
		// The fix is to stop trusting package.json's script body at all: this
		// Argv is the exact, reviewed invocation, run directly via `pnpm
		// exec` (still resolving the pinned local eslint from
		// node_modules/.bin the same way `pnpm run` would) rather than
		// through a script name whose body lives in a file this check does
		// not otherwise inspect. There is no longer a "lint" script value to
		// tamper with — the flags are Go source in this registry, reviewed
		// like any other code change. This is the "prefer unrepresentable
		// over guarded" fix (CLAUDE.md): the attack is no longer expressible
		// through package.json, not merely harder to land unnoticed.
		//
		// package.json's "lint"/"lint:prune" scripts still exist for
		// developer convenience (documented in
		// scripts/check-eslint-suppression-expiry.sh) and MUST stay
		// byte-identical in substance to the flags below; they are no longer
		// what CI runs. Residual gap: package.json's devDependencies (and
		// pnpm-lock.yaml) still govern which eslint binary `pnpm exec`
		// resolves, so a supply-chain swap there (e.g. a pnpm override
		// redirecting the "eslint" package to a shim) is still a trust
		// boundary this check does not close — but that is the same
		// pinned-third-party-engine boundary already named in this check's
		// FixtureWaiver above, not a gap this fix introduces.
		Argv: []string{"pnpm", "exec", "eslint", ".", "--suppressions-location", "eslint-suppressions.json", "--pass-on-unpruned-suppressions"},
		// package.json (root) stays in Paths: it still pins the eslint
		// version/plugins this check runs (devDependencies) even though its
		// "lint" script body is no longer what gets executed. Without it
		// here, a PR bumping/swapping an eslint devDependency while touching
		// no frontend/, packages/, apps/, or config file would still be
		// invisible to `changed` (whole-branch review C2 class).
		//
		// eslint-suppressions.json is this Argv's own --suppressions-location
		// input (round 3 finding, both reviewers): without it here, a PR that
		// edits ONLY the suppression baseline (e.g. lowers a count by hand, or
		// via `--prune-suppressions`) selects eslint-suppression-expiry and
		// eslint-suppression-baseline-growth under `changed` but not this
		// check — so ESLint itself never re-runs against the new baseline,
		// which is exactly the scenario the baseline exists to govern.
		// eslint-suppressions.expiry.json is NOT added here: it is read only
		// by check-eslint-suppression-expiry.sh, not by this Argv, so it has
		// no bearing on whether eslint itself needs to re-run.
		Paths: []string{"frontend/", "packages/", "apps/", "eslint.config.mjs", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc", "eslint-suppressions.json"},
		CIJob: "ci.yml:verify",
	},
	{
		// A2.1 (issue #91): eslint's own "lint" script now runs with
		// --pass-on-unpruned-suppressions (see package.json), so eslint
		// itself never fails just because a baselined finding got fixed. That
		// means the suppressions file can otherwise sit forever with no
		// forcing function to revisit it — this check is that forcing
		// function, repo-authored (not third-party), hence a real Fixture
		// below rather than a waiver.
		ID:       "eslint-suppression-expiry",
		Desc:     "every eslint-suppressions.json baseline has a live (non-expired) eslint-suppressions.expiry.json entry (A2.1)",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-eslint-suppression-expiry.sh"},
		Fixture: &Fixture{
			Dir:  "eslint-suppression-expiry",
			Want: []string{"EXPIRED"},
		},
		// Same Paths as "eslint" above: whenever frontend/packages/apps code
		// or the eslint config that decides which rules are ratcheted can
		// change, the suppression baseline and its expiry dates are back in
		// scope too. Plus the check's own inputs (C2 class).
		Paths: []string{"frontend/", "packages/", "apps/", "eslint.config.mjs", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc", "eslint-suppressions.json", "eslint-suppressions.expiry.json", "scripts/check-eslint-suppression-expiry.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		// A2.1 review round 2 (R1): eslint-suppression-expiry above closes
		// the "does the baseline ever get revisited" half of the ratchet,
		// but nothing stopped the baseline from GROWING in the meantime —
		// `eslint . --suppressions-location eslint-suppressions.json
		// --suppress-all` silently absorbs a brand-new violation into the
		// file and exits 0, and a subsequent plain `pnpm run lint` then
		// passes clean (proven live in the cold review of this PR, and
		// reproduced by scripts/check-eslint-suppression-baseline-growth.sh's
		// own guard fixture below). A baseline that can grow with no gate
		// noticing is a baseline, not a ratchet — this check is the
		// missing monotonicity half.
		//
		// Comparison point is the merge base with origin/main, computed
		// live by `git merge-base` inside the script — never a second
		// checked-in copy of eslint-suppressions.json. A duplicate baseline
		// file is exactly the hand-synced-enumeration defect class this
		// repo keeps hitting (see the dynamically-derived rule list in
		// check-eslint-suppression-expiry.sh's own comment). Shrinking or
		// disappearing (file, rule) entries always pass — burn-down must
		// never be blocked, only growth is gated.
		//
		// needsGitDepth + --require-infra (ci.yml:verify sets it) means a
		// shallow clone FAILS this check rather than silently skipping it;
		// the script itself also fails closed if origin/main cannot be
		// resolved or `git merge-base` errors, for the same reason. See
		// the script's own header for the full edge-case inventory
		// (first-file-introduction, renames, waiver escape).
		ID:       "eslint-suppression-baseline-growth",
		Desc:     "eslint-suppressions.json never grows a (file, rule) count relative to the merge base with origin/main (A2.1 review round 2, R1)",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-eslint-suppression-baseline-growth.sh"},
		Needs:    []string{needsGitDepth},
		Fixture: &Fixture{
			Dir:  "eslint-suppression-baseline-growth",
			Want: []string{"GREW"},
		},
		// eslint-suppressions.json is the check's actual subject; the
		// script and the waiver file are its own inputs (C2 class, same
		// reasoning as every other script-named-in-Paths entry above).
		Paths: []string{"eslint-suppressions.json", "scripts/check-eslint-suppression-baseline-growth.sh", "scripts/check-governance-waivers.txt"},
		CIJob: "ci.yml:verify",
	},
	{
		// A2.2 (issue #91): knip has no built-in baseline/suppression
		// mechanism the way ESLint 10's --suppressions-location does (see
		// the "eslint" check above and its two ratchets), so this script IS
		// the baseline: it runs knip scoped to exactly `files` (dead files)
		// and `exports` (unused exports) — A2.2's owned slice, not clones,
		// not component size, not unused dependencies/types — and compares
		// against dead-code-baseline.json, failing only on entries knip
		// reports now that the baseline does not already record.
		//
		// knip.json's own "ignore" (design-source/**, scripts/perf/**,
		// tools/perfbench/**) keeps design mockups and standalone
		// k6/bench scripts — neither reachable from any package's entry
		// graph — out of the count; everything else, including every
		// workspace's src/ (the issue's 29-and-growing dead-file count and
		// its abandoned documents/canvas/ subtree) stays in scope.
		ID:       "knip-dead-code",
		Desc:     "knip reports no dead src/ file or unused export beyond dead-code-baseline.json (A2.2)",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-dead-code.sh"},
		Fixture: &Fixture{
			Dir:          "knip-dead-code",
			ArgvOverride: []string{"bash", "scripts/check-dead-code.sh", "--baseline", "{{fixture}}/baseline.json", "--current-json", "{{fixture}}/current.json"},
			Want:         []string{"NEW DEAD FILE", "NEW UNUSED EXPORT"},
		},
		// dead-code-baseline.json is this check's own baseline input
		// (A2.1 review's bot-caught bug: "a check whose baseline could
		// change without the check re-running" — same reasoning as
		// eslint-suppressions.json in the "eslint" check's Paths above).
		// knip.json and the check script are the check's own definition.
		Paths: []string{"frontend/", "packages/", "apps/", "knip.json", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc", "dead-code-baseline.json", "scripts/check-dead-code.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		// A2.2 review-precedent guard (issue #91): knip-dead-code above
		// proves the working tree never has MORE debt than
		// dead-code-baseline.json. It proves nothing about the baseline
		// FILE itself — nothing stops a PR from hand-editing new entries
		// straight into the "accepted" list, which knip-dead-code would
		// then treat as clean. This is the same monotonicity gap A2.1
		// closed for eslint-suppressions.json with
		// eslint-suppression-baseline-growth (see that check's comment
		// above); this check is its structural twin for
		// dead-code-baseline.json. Comparison point is `git merge-base`
		// computed live, never a second checked-in copy of the baseline.
		ID:       "dead-code-baseline-growth",
		Desc:     "dead-code-baseline.json never grows a dead file or (file, export) pair relative to the merge base with origin/main (A2.2)",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-dead-code-baseline-growth.sh"},
		Needs:    []string{needsGitDepth},
		Fixture: &Fixture{
			Dir:  "dead-code-baseline-growth",
			Want: []string{"GREW"},
		},
		// dead-code-baseline.json is the check's actual subject; the
		// script and the waiver file are its own inputs (C2 class, same
		// reasoning as eslint-suppression-baseline-growth's Paths above).
		Paths: []string{"dead-code-baseline.json", "scripts/check-dead-code-baseline-growth.sh", "scripts/check-governance-waivers.txt"},
		CIJob: "ci.yml:verify",
	},
	{
		// A2.2 (issue #91): knip-dead-code and dead-code-baseline-growth
		// above prove the tree never has more debt than the baseline and
		// that the baseline itself never grows unreviewed, but neither has
		// a forcing function to make anyone ever revisit it once written --
		// same gap eslint-suppression-expiry closes for
		// eslint-suppressions.json (see that check's comment above). This is
		// its structural twin for dead-code-baseline.json: keyed by the
		// whole baseline rather than per-rule, since knip's baseline has no
		// natural per-rule grouping the way ESLint's does.
		ID:       "dead-code-baseline-expiry",
		Desc:     "dead-code-baseline.json has a live (non-expired) dead-code-baseline.expiry.json entry when non-empty (A2.2)",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-dead-code-baseline-expiry.sh"},
		Fixture: &Fixture{
			Dir: "dead-code-baseline-expiry",
			// A bare "EXPIRED" substring would match any of the script's
			// five distinct failure branches (missing expiry file,
			// malformed expiry JSON, missing "expires" field, malformed
			// date format, invalid calendar date, or an actually-past
			// date) — it would not prove THIS fixture (a non-empty
			// baseline plus a fixed, always-past expires: "2020-01-01")
			// fails for the reason under test rather than by accident of
			// one of the other four. Pinned to the exact stable message
			// text emitted by the real script for this fixture's input
			// (verified live: `bash scripts/check-dead-code-baseline-expiry.sh`
			// against the fixture's two files prints this line, minus the
			// "(today: ...)" suffix which is excluded here because it is
			// not stable across days — review finding on PR #119).
			Want: []string{"EXPIRED — baseline expired on 2020-01-01"},
		},
		// Same reasoning as eslint-suppression-expiry's Paths: whenever
		// frontend/packages/apps code, knip's config, or the baseline
		// itself can change, the expiry gate is back in scope too, plus the
		// check's own inputs (C2 class).
		Paths: []string{"frontend/", "packages/", "apps/", "knip.json", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".nvmrc", "dead-code-baseline.json", "dead-code-baseline.expiry.json", "scripts/check-dead-code-baseline-expiry.sh"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:       "css-tokens",
		Desc:     "no new raw hex colors in module.css",
		Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
		Argv:     []string{"bash", "scripts/check-css-token-discipline.sh"},
		Fixture: &Fixture{
			Dir:  "css-tokens",
			Want: []string{"RAW-HEX"},
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
		ID:       "fe-boundary-allowlist",
		Desc:     "the cross-feature ALLOWLIST is shrink-only and references existing feature directories",
		Profiles: []string{ProfilePR, ProfileFull},
		Argv:     []string{"node", "scripts/check-fe-boundary-allowlist.mjs"},
		Needs:    []string{needsGitDepth},
		Fixture: &Fixture{
			Dir:          "fe-boundary-allowlist",
			CopyFromRepo: []string{"scripts/check-fe-boundary-allowlist.mjs"},
			Want:         []string{"frontend boundary ALLOWLIST grew"},
		},
		Paths: []string{"frontend/apps/web/src/features/", "eslint.config.mjs", "scripts/check-fe-boundary-allowlist.mjs"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "fe-build",
		FixtureWaiver: &Waiver{Kind: WaiverBuildStep, Why: "a production frontend build; it fails when TypeScript or Vite cannot produce the deployable bundle, and there is no rule to feed bad input to."},
		Desc:          "production build of @metaldocs/web",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "--filter", "@metaldocs/web", "run", "build"},
		Paths:         []string{"frontend/apps/web/", "packages/", "api/openapi/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob:         "ci.yml:verify",
	},
	{
		ID:            "fe-typecheck",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party tool (tsc); a fixture would prove TypeScript rejects bad types."},
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
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "a test suite (vitest); same as go-test-unit."},
		Desc:          "vitest over @metaldocs/web",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"pnpm", "--filter", "@metaldocs/web", "run", "test"},
		// Same C3 fix as fe-typecheck above, same reason.
		Paths: []string{"frontend/", "packages/", "pnpm-lock.yaml", "package.json", "pnpm-workspace.yaml", ".nvmrc"},
		CIJob: "ci.yml:verify",
	},
	{
		ID:            "docx-typecheck",
		FixtureWaiver: &Waiver{Kind: WaiverThirdParty, Why: "third-party tool (tsc); same as fe-typecheck."},
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
		FixtureWaiver: &Waiver{Kind: WaiverBuildStep, Why: "a build step, not a guard — it produces dist/meta.json for docx-test. It fails when the workspace does not build; there is no rule to feed bad input to."},
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
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "a test suite (vitest over docx-v2); same as go-test-unit."},
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
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "a test suite, not a guard: it fails when a test fails, which is the property, and every test in it is its own fixture."},
		Desc:          "go test ./... (no integration tag)",
		Profiles:      []string{ProfilePR, ProfileFull},
		Argv:          []string{"go", "test", "-count=1", "-timeout", "600s", "./..."},
		CIJob:         "ci.yml:verify",
	},
	{
		ID:            "go-test-integration",
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "a test suite, not a guard: it fails when a test fails, which is the property, and every test in it is its own fixture."},
		Desc:          "the integration suite without -race (PR fast path)",
		Profiles:      []string{ProfilePR},
		// The package list is not written here. It is `go list` over the roots
		// the Partition below names, resolved on the commit under test, so a
		// new test package joins this suite by existing rather than by being
		// added to a second list. That is also what makes ci.yml's four-shard
		// matrix safe: the shards partition a set neither this file nor that
		// one enumerates. See partition.go.
		Argv:      []string{"go", "test", "-tags", "integration", "-count=1", "-timeout", "900s", packagesPlaceholder},
		Partition: integrationPartition,
		Needs:     []string{needsPostgres},
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
		//   - tools/verify/: this check's subject is now COMPUTED — which
		//     packages run comes from partition.go's `go list`, and how they
		//     are split comes from shardOf plus testweights.json. A PR that
		//     edits the partitioner changes what the suite executes, so a
		//     `--changed` selection that skipped the suite on such a PR would
		//     be a change to the runner that the runner never ran. Found for
		//     real: the PR that introduced sharding reported four green shards
		//     that each selected zero checks.
		// When in doubt this unit's brief says include the path, so this list
		// is deliberately roots, not the narrower subset that happened to be
		// touched by any one past incident.
		Paths: []string{"go.mod", "go.sum", "db/", "internal/", "apps/", "tests/", "tools/verify/"},
		CIJob: "ci.yml:test-integration",
	},
	{
		ID:            "go-test-integration-race",
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "a test suite, not a guard: it fails when a test fails, which is the property, and every test in it is its own fixture."},
		Desc:          "the full integration suite with -race (nightly and release)",
		Profiles:      []string{ProfileFull},
		Argv:          []string{"go", "test", "-tags", "integration", "-count=1", "-race", "-timeout", "900s", packagesPlaceholder},
		Partition:     integrationPartition,
		Needs:         []string{needsPostgres},
		Paths:         []string{"go.mod", "go.sum", "db/", "internal/", "apps/", "tests/", "tools/verify/"},
		CIJob:         "nightly.yml:integration-race",
	},

	// ---- Traceability -----------------------------------------------------
	{
		ID:            "req-trace-selftest",
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "same shape as api-lint-selftest: a Go test suite over scripts/req-trace, with its own testdata."},
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
			Want:         []string{"UNCOVERED MUST REQ(s) (1):"},
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
			Want:         []string{"pass-mislabelled.json"},
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
			Want:         []string{"no-such-check-id"},
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
		FixtureWaiver: &Waiver{Kind: WaiverTestSuite, Why: "this check IS the negative-fixture harness: it fails when any guard exits 0 on bad input, and its own bad-input proof is the 22 fixtures it runs. Fixturing it would be circular."},
		CIJob:         "ci.yml:verify",
	},
}
