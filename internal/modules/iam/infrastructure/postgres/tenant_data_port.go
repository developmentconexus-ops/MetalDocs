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
// Erase order (children before parents, within iam's own tables):
//  1. iam_group_members (FK -> iam_groups)
//  2. iam_user_roles (FK -> iam_users)
//  3. user_process_areas (FK -> iam_users via tenant_id+granted_by/revoked_by)
//  4. tenant_lifecycle_jobs (FK -> tenants; independent of the rest)
//  5. iam_groups (FK -> tenants)
//  6. iam_users
//  7. tenant_plans (independent, PK = tenant_id)
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

func (p *TenantDataPort) Module() string { return "iam" }

func (p *TenantDataPort) Tables() []string {
	tables := []string{
		"metaldocs.iam_users",
		"metaldocs.iam_user_roles",
		"metaldocs.iam_groups",
		"metaldocs.iam_group_members",
		"metaldocs.tenant_plans",
		"public.user_process_areas",
	}
	if p.hasTenantLifecycleJobs {
		tables = append(tables, "metaldocs.tenant_lifecycle_jobs")
	}
	return tables
}

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

	// public.user_process_areas: hard delete (contract §3.2) — retention is
	// impossible, its FKs point at iam_users and document_process_areas rows
	// erasure removes. Requires the erasure-context GUCs (see type doc).
	n, err = tenantdata.EraseTable(ctx, tx, "public.user_process_areas", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.user_process_areas"] = n

	if p.hasTenantLifecycleJobs {
		n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.tenant_lifecycle_jobs", "tenant_id", tenantID)
		if err != nil {
			return nil, err
		}
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
