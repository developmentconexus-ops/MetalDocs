// Package tenantdata hosts the TenantDataPort contract — the fan-out seam M7
// F7.3 (export + crypto-shred erasure) uses to reach every tenant-scoped
// table without any module importing another module's repository, SQL, or
// domain internals (invariant 6).
//
// Platform home, not iam/domain: TenantDataPort is implemented by MANY
// modules (documents, controlleddocuments, templates, iam, auth,
// notifications, render, jobs, audit, tokens, taxonomy — one port per module
// that owns >=1 tenant_id table). If the interface lived in
// internal/modules/iam/domain (as the F7.3 spec's illustrative sketch
// suggested), every implementing module would need to import iam/domain to
// satisfy it, and iam already imports several of those modules' published
// ports for other reasons — a straight shot at an import cycle for no
// benefit (iam doesn't own this contract; it merely orchestrates the
// fan-out via the registry in this same package). Hosting the interface in
// internal/platform/tenantdata mirrors internal/platform/idempotency and
// internal/platform/tripwire: a dependency-free seam every module can import
// downward, with no module gaining a new peer/upward dependency. Decision
// recorded here per the task instructions; the orchestrator (iam
// application-layer export/erase services, F7.3 items 5-6) depends on this
// package plus the Port slice built by AllTenantDataPorts, never the other
// way around.
package tenantdata

import (
	"context"
	"database/sql"
	"encoding/json"
)

// TableExport carries one table's tenant-scoped rows, serialized as JSON
// (one json.RawMessage per row, via row_to_json so the shape always matches
// the live column set — no hand-maintained struct drifts from the schema).
type TableExport struct {
	Table string
	Rows  []json.RawMessage
}

// Port is implemented once per module that owns at least one tenant_id (or
// actor_tenant_id) table. Each implementation touches ONLY the tables it
// declares in Tables() — invariant 6 (cross-module access only through a
// published port). The TenantDataPortCoverage integration test
// (tests/integration/tenantdata) is the anti-rot guard: the union of every
// registered port's Tables() must equal the live schema census exactly.
type Port interface {
	// Module returns the owning module's name (must be unique across the
	// registry — enforced by TestAllTenantDataPorts_ModuleNamesUnique).
	Module() string

	// Tables returns every tenant-scoped table this port owns, fully
	// qualified (e.g. "metaldocs.iam_users", "public.documents"). Must be
	// non-empty (enforced by TestAllTenantDataPorts_TablesNonEmpty).
	Tables() []string

	// ExportTenantData returns one TableExport per table in Tables(),
	// including empty ones (an empty Rows slice signals "included, no
	// rows" to the completeness header — never omit a table the port
	// owns). Every SELECT carries an explicit tenant_id (or
	// actor_tenant_id) predicate — never relies on RLS alone.
	ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]TableExport, error)

	// EraseTenantData deletes (or, for audit_events, deliberately does NOT
	// delete) tenantID's rows from every table in Tables(), running inside
	// the caller-supplied tx (the erase orchestrator's single fan-out tx,
	// GUC-seeded per M6 F6.4 idiom). Returns rows-deleted-per-table
	// (fully-qualified table name -> count; 0 for tables intentionally left
	// untouched, e.g. audit_events). Every DELETE carries an explicit
	// tenant_id predicate, and children are deleted before parents within
	// this port's own table set (FK order) — cross-module FK ordering is
	// the orchestrator's job (it calls ports in a fixed sequence), not any
	// one port's.
	EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error)
}

// EarlyEraser is an OPTIONAL extra a Port may implement (type-asserted by the
// orchestrator; most ports do not need it) when one of its own tables holds
// an outbound FK into a table owned by a module ranked LATER in
// tenant_lifecycle_service.go's eraseOrder. The flat, per-module eraseOrder
// can express "module A's rows must be gone before module B's" in only one
// direction; it cannot express a genuine two-way dependency where module A
// also needs module B's rows to still exist for a DIFFERENT one of A's own
// tables (e.g. an actor/user FK back into iam). EraseEarly runs, for every
// port that implements it, before the main per-module eraseOrder fan-out —
// i.e. before ANY ordered port's EraseTenantData — so its deletes are safe
// against a target table another (later-ranked-relative-to-it-doesn't-matter,
// because this runs first) module is about to delete. A table erased here
// must NOT also be erased again in the same port's ordinary EraseTenantData
// (that would double-count rows and delete nothing on the second pass, since
// EraseTable is a plain tenant-scoped DELETE, not idempotent against an
// already-empty table in a way that would be wrong, but would misreport row
// counts) — Tables() still lists it, EraseTenantData must not.
type EarlyEraser interface {
	// EraseEarly deletes tenantID's rows from whichever subset of this
	// port's Tables() must be gone before the main eraseOrder fan-out runs,
	// inside the caller-supplied tx. Returns rows-deleted-per-table, same
	// contract as EraseTenantData.
	EraseEarly(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error)
}
