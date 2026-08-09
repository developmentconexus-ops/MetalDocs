package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// reqLiteralRe matches any REQ ID literal, anywhere in a line (comments,
// string literals, t.Fatalf messages — all legitimate per contract §2.1
// evidence kind (a): "REQ-… literal in any *_test.go under internal/ or apps/").
var reqLiteralRe = regexp.MustCompile(`REQ-[A-Z]+-[0-9]+`)

// ScanTestCitations walks the given top-level dirs (relative to root) and
// returns, for every REQ ID literal found in a *_test.go file, the sorted,
// de-duplicated list of relative file paths that cite it. Non-test .go files
// are not evidence sources — the contract's evidence kind (a) is scoped to
// *_test.go specifically, so a REQ ID mentioned only in production code
// comments does not count as covered.
func ScanTestCitations(root string, dirs []string) (map[string][]string, error) {
	result := map[string]map[string]bool{}

	for _, d := range dirs {
		start := filepath.Join(root, d)
		if _, err := os.Stat(start); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, rerr := os.ReadFile(path) // #nosec G122 G304 -- path is produced by filepath.Walk over root-joined, caller-supplied dirs (fixed CLI config, not external input); this is a local CI tool walking its own repo checkout, not a multi-tenant service, so symlink TOCTOU is not an exploitable trust-boundary crossing here.
			if rerr != nil {
				return fmt.Errorf("read %s: %w", path, rerr)
			}
			ids := reqLiteralRe.FindAllString(string(content), -1)
			if len(ids) == 0 {
				return nil
			}
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				rel = path
			}
			rel = filepath.ToSlash(rel)
			for _, id := range ids {
				if result[id] == nil {
					result[id] = map[string]bool{}
				}
				result[id][rel] = true
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", start, err)
		}
	}

	out := make(map[string][]string, len(result))
	for id, files := range result {
		list := make([]string, 0, len(files))
		for f := range files {
			list = append(list, f)
		}
		sort.Strings(list)
		out[id] = list
	}
	return out, nil
}
