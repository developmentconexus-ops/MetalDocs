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
		profile = flag.String("profile", ProfileFast, "which profile to run: "+strings.Join(profileOrder, ", "))
		list    = flag.Bool("list", false, "print the registry grouped by profile and exit")
		audit   = flag.Bool("audit", false, "report checks with no CI job, and exit non-zero if any exist")
		only    = flag.String("only", "", "comma-separated check IDs to run, ignoring the profile")
		base    = flag.String("base", "origin/main", "base ref for --profile=changed")
		jobs    = flag.Int("j", defaultParallelism(), "how many checks to run concurrently")
		verbose = flag.Bool("v", false, "stream output of passing checks too")
	)
	flag.Parse()

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
		os.Exit(printAudit())
	}

	selected, err := selectChecks(*profile, *only, *base)
	if err != nil {
		fatalf("%v", err)
	}
	if len(selected) == 0 {
		fmt.Println("verify: nothing to run")
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
	results := run(selected, *jobs, *verbose)
	os.Exit(report(results, *profile))
}

// ---------------------------------------------------------------- selection

func selectChecks(profile, only, base string) ([]Check, error) {
	if only != "" {
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
			var unknown []string
			for id := range want {
				unknown = append(unknown, id)
			}
			sort.Strings(unknown)
			return nil, fmt.Errorf("unknown check ID(s): %s", strings.Join(unknown, ", "))
		}
		return out, nil
	}

	if profile == ProfileChanged {
		changed, err := changedFiles(base)
		if err != nil {
			return nil, fmt.Errorf("--profile=changed needs a diff against %s: %w", base, err)
		}
		var out []Check
		for _, c := range checks {
			if !hasProfile(c, ProfilePR) {
				continue
			}
			if matchesPaths(c, changed) {
				out = append(out, c)
			}
		}
		return out, nil
	}

	if !validProfile(profile) {
		return nil, fmt.Errorf("unknown profile %q; valid: %s", profile, strings.Join(profileOrder, ", "))
	}
	var out []Check
	for _, c := range checks {
		if hasProfile(c, profile) {
			out = append(out, c)
		}
	}
	return out, nil
}

func hasProfile(c Check, p string) bool {
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

func changedFiles(base string) ([]string, error) {
	mb, err := exec.Command("git", "merge-base", base, "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("git merge-base %s HEAD: %w", base, err)
	}
	out, err := exec.Command("git", "diff", "--name-only", strings.TrimSpace(string(mb))).Output()
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

	if want := readTrimmed(".nvmrc"); want != "" {
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
	out, err := exec.Command(bin, args...).Output()
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

func readTrimmed(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// ---------------------------------------------------------------- execution

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
	out, err := exec.Command("git", "rev-parse", "--is-shallow-repository").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func run(selected []Check, parallelism int, verbose bool) []result {
	if parallelism < 1 {
		parallelism = 1
	}
	results := make([]result, len(selected))
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	var printMu sync.Mutex

	for i, c := range selected {
		wg.Add(1)
		go func(i int, c Check) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if reason := missingInfra(c); reason != "" {
				results[i] = result{check: c, status: statusSkip, reason: reason}
				printMu.Lock()
				fmt.Printf("  SKIP  %-24s %s\n", c.ID, reason)
				printMu.Unlock()
				return
			}

			start := time.Now()
			cmd := exec.CommandContext(context.Background(), c.Argv[0], c.Argv[1:]...)
			cmd.Dir = c.Dir
			out, err := cmd.CombinedOutput()
			d := time.Since(start)

			status := statusPass
			if err != nil {
				status = statusFail
			}
			results[i] = result{check: c, status: status, output: string(out), duration: d}

			printMu.Lock()
			fmt.Printf("  %-5s %-24s %6.1fs\n", status, c.ID, d.Seconds())
			if verbose && status == statusPass && len(out) > 0 {
				fmt.Println(indent(string(out)))
			}
			printMu.Unlock()
		}(i, c)
	}
	wg.Wait()
	return results
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
			fmt.Printf("  $ %s\n", strings.Join(r.check.Argv, " "))
			if r.check.CIJob != "" {
				fmt.Printf("  CI job: %s\n", r.check.CIJob)
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

// printAudit reports registry entries with no corresponding CI job. Such a
// check runs on developer machines and not in CI, which makes it advice
// rather than a control — the same failure mode as an unwired script.
func printAudit() int {
	var gaps []Check
	for _, c := range checks {
		if c.CIJob == "" {
			gaps = append(gaps, c)
		}
	}
	fmt.Printf("verify --audit: %d checks, %d with no CI job\n", len(checks), len(gaps))
	if len(gaps) == 0 {
		return 0
	}
	fmt.Println("\nThese run locally but nothing enforces them on a PR:")
	for _, c := range gaps {
		fmt.Printf("  - %-24s %s\n", c.ID, c.Desc)
	}
	return 1
}

// ------------------------------------------------------------------- helpers

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
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
