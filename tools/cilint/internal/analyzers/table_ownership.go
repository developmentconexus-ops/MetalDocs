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
// Ownership is data now, and the parity test in this package holds it against
// the governed documentation (wiki/database/tables/<table>.md) on every
// `go test ./...`, so this file cannot drift away from the docs silently.
//
//go:embed table-ownership.json
var tableOwnershipJSON []byte

// ownershipEntry is one governed row. WikiOwner is set only where the governed
// doc labels a table with something that is not a module directory (a
// platform-layer owner); recording it here makes the divergence explicit and
// reviewable instead of an unexplained parity failure.
type ownershipEntry struct {
	Table     string `json:"table"`
	Owner     string `json:"owner"`
	WikiOwner string `json:"wiki_owner"`
	Note      string `json:"note"`
}

type ownershipCatalog struct {
	Doc     string            `json:"_doc"`
	Aliases map[string]string `json:"aliases"`
	Tables  []ownershipEntry  `json:"tables"`
}

// hgOwnerByTable maps every owned base table to the TOP-LEVEL module that owns
// it (holds its writes). "Top-level" means the first segment under
// internal/modules/: iam/presence ⊂ iam, so an intra-context access across
// sub-packages is NOT cross-module. This is the data ADR-0039 D1 (base table =
// violation) classifies against.
var hgOwnerByTable = mustOwnerByTable()

// ownershipEntries is the parsed catalog, kept for the parity test.
var ownershipEntries = mustOwnershipCatalog().Tables

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
	cat := mustOwnershipCatalog()
	out := make(map[string]string, len(cat.Tables))
	for _, e := range cat.Tables {
		if e.Table == "" || e.Owner == "" {
			panic(fmt.Sprintf("cilint: table-ownership.json has an entry with an empty table or owner: %+v", e))
		}
		if prev, dup := out[e.Table]; dup {
			panic(fmt.Sprintf("cilint: table-ownership.json declares %q twice (%q, %q)", e.Table, prev, e.Owner))
		}
		out[e.Table] = e.Owner
	}
	return out
}
