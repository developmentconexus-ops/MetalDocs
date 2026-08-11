package postgres

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the iam module's M7 F7.3 TenantDataPort implementation.
// It owns the identity/membership tables plus the tenant plan row and the
// process-area grant table (public.user_process_areas) — the census assigns
// user_process_areas to iam because its repository
// (iam/infrastructure/postgres/user_area_repository.go) lives here, not
// under documents or taxonomy. It also owns metaldocs.tenant_lifecycle_jobs
// per the F7.3 task assignment (the export/erase job-tracking table
// introduced by the parallel Task A/F7.2-F7.3 migration work) — included
// only when the table exists (existence probed once at construction, since
// Tables() carries no context/db parameters to check live) so this port
// still builds and runs correctly against a database snapshot from before
// that migration lands.
//
// Erase order — TWO PHASES, not one (see EraseEarly and the
// tenantdata.EarlyEraser doc for why a single flat order cannot express
// this): user_process_areas and capability_bindings both hold an outbound
// FK into metaldocs.document_process_areas, owned by the "taxonomy" module,
// which tenant_lifecycle_service.go's eraseOrder ranks BEFORE "iam" (iam
// runs last overall because iam_users is the FK target of other modules'
// actor/user columns — see that file's eraseOrder comment). Left alone,
// taxonomy's port would delete document_process_areas rows while these two
// iam-owned tables still reference them, raising a foreign-key violation —
// live-reproduced against PR #113 (fk_capability_bindings_scope_area,
// 2026-08-11 cold review) for any tenant holding an area-scoped
// capability_bindings row; user_process_areas has carried the identical,
// previously-latent defect since ADR 0037 (no existing test seeded both an
// area grant and ran a real erase together, so it was uncovered, not
// passing). CORRECTION: an earlier version of this comment claimed
// registry.go's AllTenantDataPorts declaration order governs execution
// order — it does not. AllTenantDataPorts is registration only; execution
// order is exclusively decided by orderedPorts()/eraseOrder in
// tenant_lifecycle_service.go, where taxonomy (rank 6) runs BEFORE iam
// (rank 11, last) — the reverse of what that comment assumed.
//
// Phase EARLY (EraseEarly, called by the orchestrator before ANY ordered
// port runs — see tenantdata.EarlyEraser):
//  1. user_process_areas (FK -> iam_users via tenant_id+granted_by/revoked_by,
//     which still exist at this point — the late phase below hasn't run yet
//     — and FK -> document_process_areas, which taxonomy hasn't deleted yet
//     either, since taxonomy is an ordinary ordered port and this phase runs
//     before the ordered fan-out starts)
//  2. capability_bindings (ADR 0092 D1 / A8.1) — FK -> iam_users
//     (subject_user_id), iam_groups (subject_group_id), and
//     document_process_areas (scope_ref), all still present for the same
//     reason. metaldocs.roles (the FK target of role_code) is a global,
//     non-tenant-scoped catalog, never erased.
//
// Phase LATE (EraseTenantData, called in the normal eraseOrder fan-out,
// children before parents within iam's remaining tables):
//  1. iam_group_members (FK -> iam_groups)
//  2. iam_user_roles (FK -> iam_users)
//  3. tenant_lifecycle_jobs — NOT deleted: it is the lifecycle ledger (the
//     running erase job's own row lives here); rows are retained with
//     requested_by scrubbed to 'erased' (see EraseTenantData)
//  4. iam_groups (FK -> tenants)
//  5. iam_users
//  6. tenant_plans (independent, PK = tenant_id)
//
// public.user_process_areas is hard-DELETEd per contract §3.2 — an
// UPDATE-revoke alternative was rejected in review because retained rows FK
// (tenant_id, granted_by/revoked_by) -> iam_users and (tenant_id, area_code)
// -> document_process_areas, both of which erasure deletes, so retention is
// physically impossible. The table's trg_user_process_areas_no_delete
// trigger rejects DELETE in normal operation; the erasure tx runs with the
// tenant-erasure context GUC that the amended reject function honors
// (migration owned by the erase-orchestrator task), and with
// metaldocs.bypass_authz set (trg_require_cap_asserted fires on DELETE here
// with no DELETE CASE branch — fail-closed P0001 otherwise).
type TenantDataPort struct {
	db                     *sql.DB
	hasTenantLifecycleJobs bool
}

// NewTenantDataPort constructs the iam TenantDataPort backed by db, probing
// once whether metaldocs.tenant_lifecycle_jobs exists yet.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	p := &TenantDataPort{db: db}
	if db != nil {
		var regclass sql.NullString
		if err := db.QueryRow(`SELECT to_regclass('metaldocs.tenant_lifecycle_jobs')::text`).Scan(&regclass); err == nil {
			p.hasTenantLifecycleJobs = regclass.Valid && regclass.String != ""
		}
	}
	return p
}

var _ tenantdata.Port = (*TenantDataPort)(nil)
var _ tenantdata.EarlyEraser = (*TenantDataPort)(nil)

// Module returns the owning module name, "iam".
func (p *TenantDataPort) Module() string { return "iam" }

// Tables returns the tenant-scoped tables this port owns (see the type doc
// for the full erase-order rationale), including
// metaldocs.tenant_lifecycle_jobs only when that table exists.
func (p *TenantDataPort) Tables() []string {
	tables := []string{
		"metaldocs.iam_users",
		"metaldocs.iam_user_roles",
		"metaldocs.iam_groups",
		"metaldocs.iam_group_members",
		"metaldocs.tenant_plans",
		"public.user_process_areas",
		"metaldocs.capability_bindings",
	}
	if p.hasTenantLifecycleJobs {
		tables = append(tables, "metaldocs.tenant_lifecycle_jobs")
	}
	return tables
}

// ExportTenantData exports every row scoped to tenantID from each table
// returned by Tables, for inclusion in a tenant data export artifact.
func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	var out []tenantdata.TableExport
	for _, table := range p.Tables() {
		exp, err := tenantdata.ExportTable(ctx, db, table, "tenant_id", tenantID)
		if err != nil {
			return nil, err
		}
		out = append(out, exp)
	}
	return out, nil
}

// EraseEarly deletes user_process_areas and capability_bindings — see the
// type doc's Phase EARLY note for why these two specific tables cannot wait
// for iam's normal (LATE) turn in the eraseOrder fan-out. Called by the
// orchestrator (tenant_lifecycle_service.go) before any ordered port runs,
// including taxonomy's, whose document_process_areas delete both tables'
// FKs would otherwise race.
func (p *TenantDataPort) EraseEarly(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	// public.user_process_areas: hard delete (contract §3.2) — retention is
	// impossible, its FKs point at iam_users and document_process_areas rows
	// erasure removes. Requires the erasure-context GUCs (see type doc).
	n, err := tenantdata.EraseTable(ctx, tx, "public.user_process_areas", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.user_process_areas"] = n

	// metaldocs.capability_bindings (ADR 0092 D1 / A8.1): FKs to iam_users,
	// iam_groups, and document_process_areas — role_code's FK target,
	// metaldocs.roles, is a global catalog table with no tenant_id and is
	// never erased, only referenced.
	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.capability_bindings", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.capability_bindings"] = n

	return counts, nil
}

// EraseTenantData deletes (or, for tenant_lifecycle_jobs, scrubs) every
// remaining row scoped to tenantID within tx, in the child-before-parent
// order documented on the type (Phase LATE). user_process_areas and
// capability_bindings are NOT deleted here — see EraseEarly — deleting them
// again here would be a silent no-op (already gone) that would misreport
// their row counts as zero instead of omitting them, so they are left out
// of this method's counts entirely rather than double-counted incorrectly.
func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.iam_group_members", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.iam_group_members"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.iam_user_roles", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.iam_user_roles"] = n

	if p.hasTenantLifecycleJobs {
		// tenant_lifecycle_jobs is the lifecycle LEDGER — deleting it would
		// destroy the running erase job's own compliance record mid-erasure
		// (the RUNNING row is in this very table; deleting it made
		// MarkLifecycleJobReady a silent 0-row no-op). Rows are retained like
		// the audit skeleton — its only FK targets the tombstoned tenants row
		// — and the one PII-bearing column (requested_by, a TEXT user id
		// whose iam_users row this erasure deletes) is scrubbed instead. The
		// table's tripwire trigger is BEFORE INSERT only, so this UPDATE
		// needs no arm.
		res, err := tx.ExecContext(ctx,
			`UPDATE metaldocs.tenant_lifecycle_jobs SET requested_by = 'erased' WHERE tenant_id = $1::uuid`,
			tenantID)
		if err != nil {
			return nil, err
		}
		n, _ = res.RowsAffected()
		counts["metaldocs.tenant_lifecycle_jobs"] = n
	}

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.iam_groups", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.iam_groups"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.iam_users", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.iam_users"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.tenant_plans", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.tenant_plans"] = n

	return counts, nil
}
