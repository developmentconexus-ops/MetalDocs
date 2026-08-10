package analyzers

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// table-ownership.json is the single enforcement source for base-table
// ownership. It used to be a Go map literal inside hgcrossmodule.go, which made
// an architectural fact a property of analyzer source: the map recorded
// governance_events as approval-owned because approval writes it, contradicting
// ADR 0044, and nothing could notice. Two foreign writes stayed invisible to the
// A4.0 census as a result (#87/A1 review B4).
//
// Ownership is data now, and two tests in this package hold it honest on every
// `go test ./...`:
//
//	table_ownership_parity_test.go       — each entry agrees with the governed
//	                                       doc in wiki/database/tables/.
//	table_ownership_completeness_test.go — the catalog IS the live base-table
//	                                       universe, in both directions.
//
// The second one is what closes #87/A1 review F1. Parity alone walks
// catalog → docs, and that direction cannot see an omission: 19 of the 56 live
// base tables were simply absent, and because hgViolation treats an unknown
// table as a non-violation, every one of them was unguarded. An omission is not
// a classification.
//
//go:embed table-ownership.json
var tableOwnershipJSON []byte

// Classification of a live base table. Exactly one applies, and the absence of a
// class is an error rather than a default — "unclassified" must not be a
// reachable state, because the analyzer's behaviour for an unknown table
// (silence) is indistinguishable from its behaviour for a compliant one.
const (
	// classModule — one enforcement module owns it; foreign raw SQL violates D1.
	classModule = "module"
	// classPlatform — cross-cutting by architectural design, no module owner,
	// so there is no foreign access to define.
	classPlatform = "platform"
	// classHistorical — in the schema with no live owner. Recorded, not
	// enforced. Empty today; the class exists so a future dead table gets
	// classified rather than dropped from the file.
	classHistorical = "historical"
	// classOutOfScope — deliberately outside H-G enforcement, reason stated.
	classOutOfScope = "out-of-scope"
)

// ownershipEntry is one governed row. WikiOwner is set only where the governed
// doc labels a table with something that is not a module directory (a
// platform-layer owner, or a product surface); recording it here makes the
// divergence explicit and reviewable instead of an unexplained parity failure.
type ownershipEntry struct {
	Table     string `json:"table"`
	Class     string `json:"class"`
	Owner     string `json:"owner"`
	WikiOwner string `json:"wiki_owner"`
	Note      string `json:"note"`
}

// retiredEntry is a table wiki/database/tables still documents that the live
// schema no longer has. Naming them is what lets the completeness test compare
// the doc set against the schema mechanically instead of by reading.
type retiredEntry struct {
	Table string `json:"table"`
	Note  string `json:"note"`
}

type ownershipCatalog struct {
	Doc     string            `json:"_doc"`
	Aliases map[string]string `json:"aliases"`
	Tables  []ownershipEntry  `json:"tables"`
	Retired []retiredEntry    `json:"retired"`
}

// hgOwnerByTable maps every MODULE-OWNED base table to the TOP-LEVEL module that
// owns it (holds its writes). "Top-level" means the first segment under
// internal/modules/: iam/presence ⊂ iam, so an intra-context access across
// sub-packages is NOT cross-module. This is the data ADR-0039 D1 (base table =
// violation) classifies against.
//
// Only classModule entries appear here. A platform/historical/out-of-scope table
// is still in the catalog — so completeness sees it and nothing is omitted — but
// has no owner, so no access to it can be foreign.
var hgOwnerByTable = mustOwnerByTable()

// ownershipEntries is the parsed catalog, kept for the parity and completeness
// tests.
var ownershipEntries = mustOwnershipCatalog().Tables

// ownershipRetired is the retired-table set, likewise.
var ownershipRetired = mustOwnershipCatalog().Retired

// ownershipAliases normalises governed-doc spellings to module directory names.
var ownershipAliases = mustOwnershipCatalog().Aliases

func mustOwnershipCatalog() ownershipCatalog {
	var cat ownershipCatalog
	if err := json.Unmarshal(tableOwnershipJSON, &cat); err != nil {
		// Fail closed and loudly: an unparseable catalog means the analyzer
		// would classify nothing as owned and report zero violations, which is
		// indistinguishable from a clean tree.
		panic(fmt.Sprintf("cilint: table-ownership.json is unreadable: %v", err))
	}
	if len(cat.Tables) == 0 {
		panic("cilint: table-ownership.json declares no tables")
	}
	return cat
}

func mustOwnerByTable() map[string]string {
	out, err := ownerByTable(mustOwnershipCatalog())
	if err != nil {
		panic("cilint: " + err.Error())
	}
	return out
}

// ownerByTable is the fail-closed load. Every rejection below describes a state
// in which the analyzer would keep running while classifying LESS than it
// appears to — a module-owned table with no owner, a class nobody defined, an
// entry shadowed by a duplicate. Each of those makes some table silently
// unguarded, which is indistinguishable from a clean tree in the output. Loading
// has to refuse them rather than degrade.
//
// Split out of mustOwnerByTable so the refusals are testable: a panic wired into
// package-level var initialisation cannot be exercised, and an untested
// fail-closed path is a claim, not a mechanism.
func ownerByTable(cat ownershipCatalog) (map[string]string, error) {
	out := make(map[string]string, len(cat.Tables))
	seen := make(map[string]bool, len(cat.Tables))

	for _, e := range cat.Tables {
		if e.Table == "" {
			return nil, fmt.Errorf("table-ownership.json has an entry with no table name: %+v", e)
		}
		if seen[e.Table] {
			return nil, fmt.Errorf("table-ownership.json declares %q twice", e.Table)
		}
		seen[e.Table] = true

		owner, err := enforcedOwner(e)
		if err != nil {
			return nil, err
		}
		if owner != "" {
			out[e.Table] = owner
		}
	}

	for _, r := range cat.Retired {
		if r.Table == "" || r.Note == "" {
			return nil, fmt.Errorf("table-ownership.json retired entry needs a table and a reason: %+v", r)
		}
		if seen[r.Table] {
			return nil, fmt.Errorf("table-ownership.json lists %q as both live and retired", r.Table)
		}
		seen[r.Table] = true
	}

	return out, nil
}

// enforcedOwner reads one entry's class and returns the module the analyzer
// should enforce for it, or "" when the entry is deliberately outside
// enforcement. Every error path here is a state where the analyzer would keep
// running while guarding less than it appears to.
func enforcedOwner(e ownershipEntry) (string, error) {
	switch e.Class {
	case classModule:
		// Fail closed on unknown ownership INSIDE the governed universe: a
		// table listed as module-owned with no owner would load as "not owned",
		// i.e. silently unguarded — the same failure mode F1 was raised about,
		// one level down.
		if e.Owner == "" {
			return "", fmt.Errorf("table-ownership.json classifies %q as module-owned with no owner", e.Table)
		}
		return e.Owner, nil

	case classPlatform, classHistorical, classOutOfScope:
		// Classified as having no enforcement owner. The reason is mandatory:
		// an unexplained non-module class is indistinguishable from someone
		// silencing a finding.
		if e.Owner != "" {
			return "", fmt.Errorf("table-ownership.json gives %q class %q AND owner %q — pick one", e.Table, e.Class, e.Owner)
		}
		if e.Note == "" {
			return "", fmt.Errorf("table-ownership.json puts %q outside enforcement (class %q) with no reason", e.Table, e.Class)
		}
		return "", nil

	case "":
		return "", fmt.Errorf("table-ownership.json entry %q has no class — omission is not a classification", e.Table)

	default:
		return "", fmt.Errorf("table-ownership.json entry %q has unknown class %q", e.Table, e.Class)
	}
}
