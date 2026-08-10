// Command verify is the single entry point for "is this tree good".
//
// A1 item 5: CI calls this and nothing else, so "green locally" and "green in
// CI" are the same claim. Before it existed, every workflow spelled its own
// commands, a developer had no way to run the same set, and "it passes
// locally" was not evidence of anything — this repo has shipped red CI on
// exactly that basis, with a local Go and Node newer than CI's.
//
// Usage:
//
//	go run ./tools/verify --profile=fast     # inner loop
//	go run ./tools/verify --profile=changed  # only what the diff touches
//	go run ./tools/verify --profile=pr       # what a PR must pass
//	go run ./tools/verify --profile=full     # + integration suites
//	go run ./tools/verify --list             # what each profile contains
//	go run ./tools/verify --audit            # registry vs CI job mapping
//
// Design rules, each of which exists because its absence caused a real defect
// here:
//
//   - Checks are argv, never shell strings. A quoting bug cannot silently
//     change what ran.
//   - A check that cannot run is reported as SKIP with its reason and is
//     counted in the summary. It is never silently dropped — a check that
//     disappears when its precondition is missing is an inert control.
//   - Toolchain versions are preflighted against go.mod and .nvmrc. A run on
//     the wrong Go or Node is not a verification of what CI will do, so it is
//     called out rather than left to be discovered as a mystery CI failure.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type result struct {
	check    Check
	status   string // "PASS", "FAIL", "SKIP"
	reason   string // why skipped
	output   string
	duration time.Duration
}

const (
	statusPass = "PASS"
	statusFail = "FAIL"
	statusSkip = "SKIP"
)

func main() {
	var (
		profile           = flag.String("profile", ProfileFast, "which profile to run: "+strings.Join(profileOrder, ", "))
		list              = flag.Bool("list", false, "print the registry grouped by profile and exit")
		audit             = flag.Bool("audit", false, "report checks with no CI job, and exit non-zero if any exist")
		testdbBypassGuard = flag.Bool("testdb-bypass-guard", false, "report _test.go files that bypass testdb.Open via raw DATABASE_URL/METALDOCS_DATABASE_URL + sql.Open, and exit non-zero if any exist")
		guardFixtures     = flag.Bool("guard-fixtures", false, "feed each guard its negative fixture and require a non-zero exit; --only narrows to specific check IDs")
		only              = flag.String("only", "", "comma-separated check IDs to run, ignoring the profile")
		ciJob             = flag.String("ci-job", "", "narrow the selection to checks whose declared CIJob is this workflow job (e.g. ci.yml:verify); a CI job passes its own name so ownership is decided by the registry, not by the workflow")
		changed           = flag.Bool("changed", false, "narrow whatever selection --only/--profile made to checks whose declared Paths the diff against --base touches; --profile=changed implies this")
		base              = flag.String("base", "origin/main", "base ref for --changed / --profile=changed")
		jobs              = flag.Int("j", defaultParallelism(), "how many checks to run concurrently")
		verbose           = flag.Bool("v", false, "stream output of passing checks too")
		// In CI a SKIP is indistinguishable from a PASS at the exit code, so a
		// job can report green over zero executed tests. This flag makes
		// "cannot run here" fatal, and CI always passes it.
		requireInfra = flag.Bool("require-infra", false, "treat missing infra as a failure instead of a skip; CI always sets this")
	)
	flag.Parse()

	if err := validateOrdering(checks); err != nil {
		fatalf("invalid check registry: %v", err)
	}

	root, err := repoRoot()
	if err != nil {
		fatalf("cannot locate repo root: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		fatalf("cannot enter repo root %s: %v", root, err)
	}

	switch {
	case *list:
		printList()
		return
	case *audit:
		os.Exit(printAudit(filepath.Join(".github", "workflows")))
	case *testdbBypassGuard:
		os.Exit(printTestdbBypassGuard())
	case *guardFixtures:
		os.Exit(printGuardFixtures(root, splitIDs(*only)))
	}

	selected, scoped, err := selectChecks(*profile, *only, *base, *changed, *ciJob)
	if err != nil {
		fatalf("%v", err)
	}
	if len(selected) == 0 {
		fmt.Println(emptySelectionMessage(scoped, *base))
		return
	}

	warns := preflight()
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "verify: PREFLIGHT %s\n", w)
	}
	if len(warns) > 0 {
		fmt.Fprintln(os.Stderr, "verify: the run below may not match CI. Fix the above before trusting a green result.")
		fmt.Fprintln(os.Stderr)
	}

	fmt.Printf("verify: profile=%s checks=%d parallelism=%d\n\n", *profile, len(selected), *jobs)
	results := run(selected, *jobs, *verbose, *requireInfra)
	os.Exit(report(results, *profile))
}

// ---------------------------------------------------------------- selection

// selectChecks resolves --only/--profile to a set of checks, then applies
// --changed diff-scoping on top if either --changed was passed explicitly or
// the profile is `changed` (which has always meant "scoped", since before
// this flag existed). This is what lets a workflow job say "my set of
// checks, further narrowed to what the diff touches" —
// `--only=go-test-integration --changed` — instead of --changed being a
// whole selection mode of its own that can't be intersected with anything.
//
// The returned bool tells the caller whether scoping was actually applied,
// so an empty result can be explained: "0 checks, diff-scoped" is a fact
// worth printing, "0 checks" alone is not (see emptySelectionMessage).
func selectChecks(profile, only, base string, changedFlag bool, ciJob string) (selected []Check, scoped bool, err error) {
	scoped = changedFlag
	switch {
	case only != "":
		selected, err = selectByIDs(only)
	case profile == ProfileChanged:
		selected = selectByProfile(ProfilePR)
		scoped = true
	default:
		if !validProfile(profile) {
			return nil, false, fmt.Errorf("unknown profile %q; valid: %s", profile, strings.Join(profileOrder, ", "))
		}
		selected = selectByProfile(profile)
	}
	if err != nil {
		return nil, false, err
	}
	if ciJob != "" {
		selected, err = scopeToCIJob(selected, ciJob)
		if err != nil {
			return nil, false, err
		}
	}
	if scoped {
		selected, err = scopeToChanged(selected, base)
		if err != nil {
			return nil, false, err
		}
	}
	if err := validateSelectionOrdering(selected); err != nil {
		return nil, false, err
	}
	return selected, scoped, nil
}

// validateSelectionOrdering refuses a selection that contains a check
// without its After predecessor. This is the selection-boundary half of the
// excluded-predecessor fix (run()'s wait loop is the other half — see its
// doc comment).
//
// The prior ruling — run the dependent anyway, warn on stderr, do not
// refuse — was proven fail-open by reproduction: `apps/docx-renderer/dist/meta.json`
// already existed on disk from an earlier, unrelated build, and
// `--only=docx-test` alone ran bundle-guard.test.ts anyway, which read
// that stale file and reported PASS without ever re-validating current
// source. verify has no state that survives past its own process, so it
// cannot tell "the predecessor already ran and passed elsewhere" apart from
// "the predecessor never ran" — refusing is the only answer that isn't a
// guess. A caller that wants both checks must include both in the same
// selection, e.g. `--only=docx-build,docx-test`, which is what
// docx-renderer.yml now does (see that file).
func validateSelectionOrdering(selected []Check) error {
	inSelection := make(map[string]bool, len(selected))
	for _, c := range selected {
		inSelection[c.ID] = true
	}
	var missing []string
	for _, c := range selected {
		for _, dep := range c.After {
			if !inSelection[dep] {
				missing = append(missing, fmt.Sprintf(
					"%s declares After %s, but %s is not in this selection", c.ID, dep, dep))
			}
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf("selection drops an ordering predecessor:\n  %s\ninclude the missing check(s) in --only (or pick a profile/--changed selection that keeps both sides of the edge)",
		strings.Join(missing, "\n  "))
}

// emptySelectionMessage explains a zero-check run. A silent "0 checks, exit
// 0" is indistinguishable from a broken selector — the exact shape of the
// bug this unit exists to close (a PR touching only files no check declares
// used to skip every job and let `required` report green over nothing). A
// diff-scoped zero and an ordinary zero are different facts and read
// differently on purpose.
func emptySelectionMessage(scoped bool, base string) string {
	if scoped {
		return fmt.Sprintf("verify: 0 checks selected — --changed found no diff against %s that touches any declared Paths. A check with no declared Paths always matches, so this means the diff is entirely outside every path-scoped check and there is no repo-scoped check in the selection either. Confirm this is expected.", base)
	}
	return "verify: nothing to run"
}

// selectByIDs resolves an explicit --only list. An unknown ID is an error
// rather than a silent no-op: a workflow step that names a check that no
// longer exists must go red, not quietly verify nothing.
func selectByIDs(only string) ([]Check, error) {
	want := map[string]bool{}
	for _, id := range strings.Split(only, ",") {
		want[strings.TrimSpace(id)] = true
	}
	var out []Check
	for _, c := range checks {
		if want[c.ID] {
			out = append(out, c)
			delete(want, c.ID)
		}
	}
	if len(want) > 0 {
		unknown := make([]string, 0, len(want))
		for id := range want {
			unknown = append(unknown, id)
		}
		sort.Strings(unknown)
		return nil, fmt.Errorf("unknown check ID(s): %s", strings.Join(unknown, ", "))
	}
	return out, nil
}

// scopeToCIJob narrows `selected` to the checks that declare this CI job as
// their owner. It is what lets a profile stay the honest answer to "what
// blocks a PR" while CI still splits that set across several jobs.
//
// Without it, a profile and a job were the same thing: ci.yml:verify runs
// `--profile=changed`, so it ran EVERY `pr` check, including ones whose
// declared CIJob is another job with another job's prerequisites installed.
// That was not hypothetical — registering golangci-lint (A1.1) put a check
// owned by ci.yml:lint-go into the verify job's selection, where the binary
// does not exist, and the job failed with "golangci-lint is not on PATH".
// The two escapes from that are both worse: drop the check out of `pr` (and
// lie about what blocks a PR, which is what A1.1 exists to stop), or install
// every job's prerequisites in every job (and pay for the duplicated run).
//
// An unknown job name is an error, not an empty selection: a workflow step
// naming a job no check owns must go red, exactly as --only does for an
// unknown ID. Ownership is checked against the whole registry, not the
// current selection, so a legitimately empty *scoped* result (the diff
// touched none of this job's checks) still reads as zero checks rather than
// as a typo.
func scopeToCIJob(selected []Check, ciJob string) ([]Check, error) {
	known := false
	for _, c := range checks {
		if c.CIJob == ciJob {
			known = true
			break
		}
	}
	if !known {
		return nil, fmt.Errorf("--ci-job=%q: no registry check declares that CIJob", ciJob)
	}
	var out []Check
	for _, c := range selected {
		if c.CIJob == ciJob {
			out = append(out, c)
		}
	}
	return out, nil
}

// scopeToChanged narrows `selected` to the checks whose declared Paths
// intersect the diff against base — except outside a pull request, where it
// returns `selected` unchanged.
//
// That exception exists because the diff itself lies outside a PR: on a
// push to main, --base defaults to origin/main and HEAD already IS
// origin/main, so the diff is empty. Scoping to an empty diff means
// selecting zero checks and exiting 0 — green over nothing, which is the
// defect this whole unit exists to eliminate (the YAML path map had the same
// hole, just via `if:` conditions and `skipped` counting as green for
// `required`). Falling back to the full set is the only fail-closed answer:
// there is no diff outside a PR that could safely narrow anything.
func scopeToChanged(selected []Check, base string) ([]Check, error) {
	if runningAsNonPRCI() {
		return selected, nil
	}
	changed, err := changedFiles(base)
	if err != nil {
		return nil, fmt.Errorf("--changed needs a diff against %s: %w", base, err)
	}
	return filterByChanged(selected, changed), nil
}

// filterByChanged is the pure intersection: which of `selected` does the
// diff actually touch. Split out from scopeToChanged so the composition
// logic is testable without a git repository — matchesPaths already
// guarantees a pathless check always matches, so this filter never has to
// special-case one.
func filterByChanged(selected []Check, changed []string) []Check {
	var out []Check
	for _, c := range selected {
		if matchesPaths(c, changed) {
			out = append(out, c)
		}
	}
	return out
}

// runningAsNonPRCI reports whether this is a GitHub Actions run triggered by
// something other than a pull request — a push to main, a schedule, a
// manual dispatch. GITHUB_EVENT_NAME is GitHub's documented signal for the
// triggering event
// (https://docs.github.com/actions/learn-github-actions/variables#default-environment-variables)
// and is unset outside Actions, so an empty value means "not GitHub Actions
// at all" (a laptop) rather than "not a pull request" — the fallback must
// not fire there, or `--profile=changed` on a developer's machine would stop
// being diff-scoped and start being the full set, silently changing
// behaviour this flag has always had.
func runningAsNonPRCI() bool {
	ev := os.Getenv("GITHUB_EVENT_NAME")
	if ev == "" {
		return false
	}
	return ev != "pull_request" && ev != "pull_request_target"
}

func selectByProfile(profile string) []Check {
	var out []Check
	for _, c := range checks {
		if hasProfile(c, profile) {
			out = append(out, c)
		}
	}
	return out
}

func hasProfile(c Check, p string) bool {
	// `release` is defined by exclusion, not membership — see releaseExcluded
	// in registry.go for why. Anything `full` runs, `release` runs too, unless
	// it is named there.
	if p == ProfileRelease {
		if _, excluded := releaseExcluded[c.ID]; excluded {
			return false
		}
		return hasProfile(c, ProfileFull)
	}
	for _, cp := range c.Profiles {
		if cp == p {
			return true
		}
	}
	return false
}

func validProfile(p string) bool {
	for _, known := range profileOrder {
		if known == p {
			return true
		}
	}
	return false
}

// matchesPaths reports whether any changed file falls under one of the
// check's declared path prefixes. A check with no declared paths is
// repo-scoped and always runs — the safe direction. Narrowing a check to a
// path set is a claim that nothing outside it can break the check, so the
// default must be "run it".
func matchesPaths(c Check, changed []string) bool {
	if len(c.Paths) == 0 {
		return true
	}
	for _, f := range changed {
		f = filepath.ToSlash(f)
		for _, p := range c.Paths {
			if strings.HasPrefix(f, p) {
				return true
			}
		}
	}
	return false
}

// refPattern is the only shape --base is allowed to take. It is the single
// value in this program that reaches a subprocess from outside the
// compile-time registry, so it is validated at the boundary rather than
// trusted. This is what lets command() below state its invariant as a fact.
var refPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,254}$`)

func changedFiles(base string) ([]string, error) {
	if !refPattern.MatchString(base) {
		return nil, fmt.Errorf("--base %q is not a valid git ref name", base)
	}
	mb, err := capture("git", "merge-base", base, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("git merge-base %s HEAD: %w", base, err)
	}
	out, err := capture("git", "diff", "--name-only", strings.TrimSpace(string(mb)))
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	var files []string
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// ---------------------------------------------------------------- preflight

// preflight compares the running toolchain against the versions CI pins. It
// returns human-readable warnings; it does not abort. Aborting would make
// verify unusable on a machine that is merely ahead, but staying silent is
// how "passes locally, fails in CI" became normal here: a local Go 1.26 and
// CI's 1.25 disagree about gofmt output, and a local Node 26 and CI's 22
// disagree about cross-realm ArrayBuffer identity. Both produced real
// confusion in this repo.
func preflight() []string {
	var warns []string

	if want := goDirective(); want != "" {
		if got := toolVersion("go", `go(\d+\.\d+(?:\.\d+)?)`, "version"); got != "" {
			if !sameMinor(got, want) {
				warns = append(warns, fmt.Sprintf(
					"Go %s is running but go.mod says %s. gofmt and vet output differ across minor versions; a green run here does not prove a green run in CI.", got, want))
			}
		}
	}

	if want := nvmrcVersion(); want != "" {
		if got := toolVersion("node", `v?(\d+\.\d+\.\d+)`, "--version"); got != "" {
			if !sameMinor(got, want) {
				warns = append(warns, fmt.Sprintf(
					"Node %s is running but .nvmrc says %s. Runtime identity checks (crypto, ArrayBuffer) differ across majors.", got, want))
			}
		}
	}

	return warns
}

func goDirective() string {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	m := regexp.MustCompile(`(?m)^go (\d+\.\d+(?:\.\d+)?)`).FindSubmatch(b)
	if m == nil {
		return ""
	}
	return string(m[1])
}

func toolVersion(bin, pattern string, args ...string) string {
	out, err := capture(append([]string{bin}, args...)...)
	if err != nil {
		return ""
	}
	m := regexp.MustCompile(pattern).FindSubmatch(out)
	if m == nil {
		return ""
	}
	return string(m[1])
}

func sameMinor(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	if len(as) < 2 || len(bs) < 2 {
		return a == b
	}
	return as[0] == bs[0] && as[1] == bs[1]
}

func nvmrcVersion() string {
	b, err := os.ReadFile(".nvmrc")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// ---------------------------------------------------------------- execution

// command is the ONLY place in this program that constructs a subprocess.
//
// Every argv that reaches it is one of exactly two things: a string literal at
// the call site, or a value already validated against refPattern. There is no
// path from unvalidated input to here — which is the whole reason this is one
// function instead of six scattered exec sites. gosec cannot see that
// invariant across call sites, so it is asserted once, here, next to the code
// that makes it true, and it is checkable by reading this file top to bottom.
//
// Adding an exec call anywhere else in this package defeats that argument.
// Route it through here instead.
//
// The one non-literal element an Argv may carry is repoRootPlaceholder,
// expanded here by expandArgv. It exists because a container check must hand
// `docker -v` an ABSOLUTE host path, and there is no shell here to say $PWD.
// Expanding it here (rather than letting registry entries call os.Getwd
// themselves) keeps the invariant above true: the substitution set is closed,
// it is one function, and its value comes from the process's own working
// directory, never from a check's input.
func command(ctx context.Context, dir string, argv []string) *exec.Cmd {
	argv = expandArgv(argv)
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) // #nosec G204 -- argv is compile-time literals, refPattern-validated, or an expandArgv token; see the invariant above.
	cmd.Dir = dir
	return cmd
}

// repoRootPlaceholder expands to the absolute path of the directory the
// verifier is running in — the repo root, since every Check's paths are
// repo-relative.
//
// Deliberately not named with any of gosec G101's credential words (token,
// secret, key, pw). The first spelling of this constant was `repoToken`, and
// G101 fired on it at HIGH/LOW-confidence in both gosec and golangci-lint —
// a false positive, but one whose only other cure is a #nosec suppression.
// Renaming removes the finding instead of silencing it.
const repoRootPlaceholder = "{{repo}}"

// expandArgv substitutes the closed token set into a copy of argv. A token
// that cannot be resolved is left verbatim rather than silently dropped: the
// command then fails loudly with the token visible in its own error, which is
// what a reader needs to diagnose it.
func expandArgv(argv []string) []string {
	root, err := os.Getwd()
	if err != nil {
		return argv
	}
	out := make([]string, len(argv))
	for i, a := range argv {
		out[i] = strings.ReplaceAll(a, repoRootPlaceholder, root)
	}
	return out
}

func capture(argv ...string) ([]byte, error) {
	return command(context.Background(), "", argv).Output()
}

// missingInfra returns the reason a check cannot run here, or "".
func missingInfra(c Check) string {
	for _, need := range c.Needs {
		switch need {
		case needsPostgres:
			if os.Getenv("METALDOCS_DATABASE_URL") == "" {
				return "METALDOCS_DATABASE_URL is unset; this check needs a live Postgres"
			}
		case needsDocker:
			if _, err := exec.LookPath("docker"); err != nil {
				return "docker is not on PATH"
			}
		case needsGitDepth:
			if isShallow() {
				return "this is a shallow clone; the check needs real history"
			}
		case needsNetwork:
			// Not probed. A network check that fails offline fails loudly with
			// its own error, which is more informative than a guess here.
		}
	}
	if _, err := exec.LookPath(c.Argv[0]); err != nil {
		return fmt.Sprintf("%s is not on PATH", c.Argv[0])
	}
	return ""
}

func isShallow() bool {
	out, err := capture("git", "rev-parse", "--is-shallow-repository")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// validateOrdering rejects a registry whose After edges cannot be honoured:
// an edge naming an unknown ID, or a cycle. Both are startup failures rather
// than a runtime deadlock or a silent serialization — run() below assumes
// the graph it is handed is already a DAG over known IDs, and this is the
// one place that assumption is made true. Checked against the WHOLE
// registry, not just a selection, because a cycle that only shows up under
// some future --only combination is still a defect in the registry today.
func validateOrdering(all []Check) error {
	known := map[string]bool{}
	byID := map[string]Check{}
	for _, c := range all {
		known[c.ID] = true
		byID[c.ID] = c
	}
	if err := checkKnownAfterEdges(all, known); err != nil {
		return err
	}
	return detectOrderingCycle(all, byID)
}

// checkKnownAfterEdges rejects an After edge naming an unregistered check
// ID. Split out of validateOrdering so the "is every edge valid" question is
// separate from "is the edge set acyclic" — cycle detection assumes it is
// already walking known IDs, and this is what makes that assumption true.
func checkKnownAfterEdges(all []Check, known map[string]bool) error {
	for _, c := range all {
		for _, dep := range c.After {
			if !known[dep] {
				return fmt.Errorf("check %q declares After %q, which is not a registered check ID", c.ID, dep)
			}
		}
	}
	return nil
}

// formatOrderingCycle renders the cycle found by detectOrderingCycle's DFS:
// the portion of the walk stack from the repeated node's first occurrence
// through the edge that closed the loop.
func formatOrderingCycle(stack []string, dep string) error {
	start := 0
	for i, s := range stack {
		if s == dep {
			start = i
			break
		}
	}
	return fmt.Errorf("ordering cycle in After edges: %s -> %s",
		strings.Join(stack[start:], " -> "), dep)
}

// detectOrderingCycle runs a colored DFS over the After graph and rejects a
// cycle. Checked against the WHOLE registry, not just a selection, because a
// cycle that only shows up under some future --only combination is still a
// defect in the registry today. Iteration order over the ID set is sorted
// first so a registry with more than one cycle always reports the same one.
func detectOrderingCycle(all []Check, byID map[string]Check) error {
	const (
		white = iota
		gray
		black
	)
	color := map[string]int{}
	var stack []string
	var visit func(id string) error
	visit = func(id string) error {
		color[id] = gray
		stack = append(stack, id)
		for _, dep := range byID[id].After {
			switch color[dep] {
			case gray:
				return formatOrderingCycle(stack, dep)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
		return nil
	}

	ids := make([]string, 0, len(all))
	for _, c := range all {
		ids = append(ids, c.ID)
	}
	sort.Strings(ids) // deterministic error on a multi-cycle registry
	for _, id := range ids {
		if color[id] == white {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// runState holds everything the per-check goroutines share: the shared
// results/done slices, the concurrency semaphore, and the run-wide flags.
// Splitting run() into this struct plus methods separates the three things
// it used to do in one function body — waiting on ordering, scheduling
// under the parallelism limit, and reporting each outcome as it lands —
// into one method per concern, each independently readable and testable in
// isolation from the goroutine plumbing.
type runState struct {
	indexByID    map[string]int
	results      []result
	done         []chan struct{}
	sem          chan struct{}
	printMu      sync.Mutex
	verbose      bool
	requireInfra bool
}

func newRunState(selected []Check, parallelism int, verbose, requireInfra bool) *runState {
	indexByID := make(map[string]int, len(selected))
	for i, c := range selected {
		indexByID[c.ID] = i
	}
	// done[i] closes once results[i] is final. A dependent waits on its
	// predecessor's done channel, not on a semaphore slot, so waiting costs
	// nothing but a blocked goroutine — the concurrency limit only governs
	// checks that are actually running, which is what keeps independent
	// checks concurrent while ordered ones still serialize relative to each
	// other.
	done := make([]chan struct{}, len(selected))
	for i := range done {
		done[i] = make(chan struct{})
	}
	return &runState{
		indexByID:    indexByID,
		results:      make([]result, len(selected)),
		done:         done,
		sem:          make(chan struct{}, parallelism),
		verbose:      verbose,
		requireInfra: requireInfra,
	}
}

func (st *runState) printLine(format string, args ...any) {
	st.printMu.Lock()
	fmt.Printf(format, args...)
	st.printMu.Unlock()
}

// awaitPredecessors blocks until every check c.After depends on has a final
// result, recording c's own result and returning false the moment one of
// them makes running c impossible. It returns true only when every
// predecessor passed and c is clear to run.
func (st *runState) awaitPredecessors(i int, c Check) bool {
	for _, dep := range c.After {
		di, ok := st.indexByID[dep]
		if !ok {
			// Excluded from this selection. main() refuses this case
			// before run() is ever called (validateSelectionOrdering,
			// called from selectChecks) — this branch is the same rule
			// enforced a second time, in the execution kernel itself, so
			// a future caller that assembles `selected` some other way
			// than selectChecks still cannot slip an incomplete
			// selection past run(). FAIL, not SKIP: report() only turns
			// FAILs into a non-zero exit, and a SKIP here would be
			// exactly the "green over nothing" shape this whole unit
			// exists to close.
			reason := fmt.Sprintf("refusing: predecessor %s is not part of this selection", dep)
			st.results[i] = result{check: c, status: statusFail, reason: reason}
			st.printLine("  %s  %-24s %s\n", statusFail, c.ID, reason)
			return false
		}
		<-st.done[di]
		if st.results[di].status != statusPass {
			pred := st.results[di]
			reason := fmt.Sprintf("not run: predecessor %s did not pass (%s)", pred.check.ID, pred.status)
			st.results[i] = result{check: c, status: statusSkip, reason: reason}
			st.printLine("  %s  %-24s %s\n", "SKIP", c.ID, reason)
			return false
		}
	}
	return true
}

// skipForMissingInfra records and reports a SKIP (or, under --require-infra,
// a FAIL) when c cannot run in this environment. Returns true when it did so
// — the caller must not proceed to execute c.
func (st *runState) skipForMissingInfra(i int, c Check) bool {
	reason := missingInfra(c)
	if reason == "" {
		return false
	}
	// One decision site, two verdicts. Deriving this in report() instead
	// would mean two places agree on what a skip means.
	status, label := statusSkip, "SKIP"
	if st.requireInfra {
		status, label = statusFail, "FAIL"
	}
	st.results[i] = result{check: c, status: status, reason: reason}
	st.printLine("  %s  %-24s %s\n", label, c.ID, reason)
	return true
}

// execute runs c's Argv and records the PASS/FAIL result.
func (st *runState) execute(i int, c Check) {
	start := time.Now()

	argv := c.Argv
	if c.Stage != "" {
		root, err := repoRoot()
		if err == nil {
			var dir string
			var cleanup func()
			dir, cleanup, err = stage(context.Background(), c.Stage, root)
			if err == nil {
				defer cleanup()
				argv = withStageDir(argv, dir)
			}
		}
		if err != nil {
			// A staging failure is a FAIL, never a silent fall-through to the
			// working directory: falling back would run the check against a
			// different subject than the one it declares, which is the whole
			// defect staging exists to remove.
			st.results[i] = result{
				check:    c,
				status:   statusFail,
				output:   fmt.Sprintf("could not stage %s: %v", c.Stage, err),
				duration: time.Since(start),
			}
			st.printMu.Lock()
			fmt.Printf("  %-5s %-24s %6.1fs\n", statusFail, c.ID, time.Since(start).Seconds())
			st.printMu.Unlock()
			return
		}
	}

	out, err := command(context.Background(), c.Dir, argv).CombinedOutput()
	d := time.Since(start)

	status := statusPass
	if err != nil {
		status = statusFail
	}
	st.results[i] = result{check: c, status: status, output: string(out), duration: d}

	st.printMu.Lock()
	fmt.Printf("  %-5s %-24s %6.1fs\n", status, c.ID, d.Seconds())
	if st.verbose && status == statusPass && len(out) > 0 {
		fmt.Println(indent(string(out)))
	}
	st.printMu.Unlock()
}

// runOne is the full lifecycle of one check's goroutine: wait for
// predecessors, acquire a scheduling slot, skip if infra is missing,
// otherwise execute.
func (st *runState) runOne(i int, c Check) {
	if !st.awaitPredecessors(i, c) {
		return
	}

	st.sem <- struct{}{}
	defer func() { <-st.sem }()

	if st.skipForMissingInfra(i, c) {
		return
	}

	st.execute(i, c)
}

func run(selected []Check, parallelism int, verbose bool, requireInfra bool) []result {
	if parallelism < 1 {
		parallelism = 1
	}
	st := newRunState(selected, parallelism, verbose, requireInfra)

	var wg sync.WaitGroup
	for i, c := range selected {
		wg.Add(1)
		go func(i int, c Check) {
			defer wg.Done()
			defer close(st.done[i])
			st.runOne(i, c)
		}(i, c)
	}
	wg.Wait()
	return st.results
}

// ---------------------------------------------------------------- reporting

func report(results []result, profile string) int {
	var failed, skipped []result
	pass := 0
	for _, r := range results {
		switch r.status {
		case statusFail:
			failed = append(failed, r)
		case statusSkip:
			skipped = append(skipped, r)
		default:
			pass++
		}
	}

	if len(failed) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 72))
		for _, r := range failed {
			fmt.Printf("\nFAIL %s — %s\n", r.check.ID, r.check.Desc)
			// Expanded, not raw: a reader copying this line into a shell must
			// get the command that actually ran, tokens resolved.
			fmt.Printf("  $ %s\n", strings.Join(expandArgv(r.check.Argv), " "))
			if r.check.CIJob != "" {
				fmt.Printf("  CI job: %s\n", r.check.CIJob)
			}
			if r.reason != "" {
				fmt.Printf("  not runnable here: %s\n", r.reason)
			}
			fmt.Println(indent(r.output))
		}
		fmt.Println(strings.Repeat("=", 72))
	}

	fmt.Printf("\nverify(%s): %d passed, %d failed, %d skipped\n", profile, pass, len(failed), len(skipped))

	// Skips are printed again at the end on purpose. A skip is a hole in the
	// claim being made, and a hole that scrolled off the top of the output is
	// a hole nobody sees.
	if len(skipped) > 0 {
		fmt.Println("\nNot verified by this run:")
		for _, r := range skipped {
			fmt.Printf("  - %-24s %s\n", r.check.ID, r.reason)
		}
	}

	if len(failed) > 0 {
		return 1
	}
	return 0
}

// printGuardFixtures runs the negative-fixture spine and returns the process
// exit code. The cancel() lives here rather than inline in main's switch so it
// is not stranded by an os.Exit in the same function (gocritic exitAfterDefer).
func printGuardFixtures(root string, only []string) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runGuardFixtures(ctx, root, only); err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		return 1
	}
	return 0
}

func printList() {
	for _, p := range profileOrder {
		if p == ProfileChanged {
			fmt.Printf("\n%s: the `pr` set, filtered to checks whose declared paths appear in the diff\n", p)
			continue
		}
		fmt.Printf("\n%s:\n", p)
		for _, c := range checks {
			if !hasProfile(c, p) {
				continue
			}
			needs := ""
			if len(c.Needs) > 0 {
				needs = "  [needs " + strings.Join(c.Needs, ",") + "]"
			}
			fmt.Printf("  %-24s %s%s\n", c.ID, c.Desc, needs)
		}
	}
	fmt.Println()
}

// ------------------------------------------------------------------- helpers

func repoRoot() (string, error) {
	out, err := capture("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func defaultParallelism() int {
	n := runtime.NumCPU() / 2
	if n < 1 {
		return 1
	}
	if n > 6 {
		return 6
	}
	return n
}

func indent(s string) string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "    " + l
	}
	return strings.Join(lines, "\n")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "verify: "+format+"\n", args...)
	os.Exit(2)
}
