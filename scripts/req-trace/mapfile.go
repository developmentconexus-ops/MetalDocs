package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MapEntry is one hand-maintained manual-evidence entry from
// wiki/architecture/req-trace-map.yaml. Only kind "commit" is accepted here —
// this is the anti-gaming boundary from validation-contract.md §2.3 / F9.2
// HARD RULES: doc-annotation evidence auto-derives from the REQ line itself
// (ExtractReqs), never from this map, so bulk-"satisfied" gaming via the map
// is structurally impossible.
type MapEntry struct {
	Req  string `yaml:"req"`
	Kind string `yaml:"kind"`
	Ref  string `yaml:"ref"`
	Note string `yaml:"note"`
}

type mapFile struct {
	Entries []MapEntry `yaml:"entries"`
}

// ParseMap reads and validates the manual evidence map. A missing file is
// treated as zero entries (the map is optional — a fresh repo with no manual
// entries is still a valid, if sparser, gate run). Every present entry must
// have kind "commit" and a non-empty note naming its source; anything else is
// a hard parse error so a gaming attempt fails loudly rather than silently
// being dropped.
func ParseMap(path string) ([]MapEntry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read map %s: %w", path, err)
	}

	var mf mapFile
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("parse map %s: %w", path, err)
	}

	for i, e := range mf.Entries {
		if e.Req == "" {
			return nil, fmt.Errorf("map %s entry %d: missing req", path, i)
		}
		if e.Kind != "commit" {
			return nil, fmt.Errorf("map %s entry %d (%s): kind %q not accepted — only kind: commit is a valid manual-map evidence kind (anti-gaming, contract §2.3)", path, i, e.Req, e.Kind)
		}
		if e.Ref == "" {
			return nil, fmt.Errorf("map %s entry %d (%s): missing ref (commit hash)", path, i, e.Req)
		}
		if e.Note == "" {
			return nil, fmt.Errorf("map %s entry %d (%s): missing note — every map entry must name its evidence source", path, i, e.Req)
		}
	}

	return mf.Entries, nil
}
