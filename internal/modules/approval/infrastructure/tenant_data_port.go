package infrastructure

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the documents/approval module's M7 F7.3 TenantDataPort.
// It owns public.approval_instances, public.approval_routes,
// public.approval_route_stage_selectors, public.approval_signoffs,
// public.approval_review_verdicts, and public.approval_delegations
// (§3.2: crypto-shred PII in payload + delete rows; this task deletes rows —
// the crypto-shred half is Task B's concern, a separate feature).
//
// approval_signoffs.tenant column is actor_tenant_id (uuid), not tenant_id
// (§0.1) — its export/erase predicate uses that column name instead.
//
// approval_route_stage_selectors (unit 3.2 / M4, migration 0303) carries its
// own tenant_id (NOT NULL). Its FK to approval_route_stages is ON DELETE
// CASCADE and route_stages cascade from approval_routes, so a routes-delete
// would sweep it — but a tenant table MUST be registered explicitly (export
// would otherwise silently omit it; the tenantdata coverage guard flags any
// unregistered tenant_id table). Erased explicitly before approval_routes.
//
// Erase order (children before parents, no ON DELETE CASCADE between the three
// aggregate roots — plain FKs, so explicit order matters):
//  1. approval_signoffs (FK -> approval_instances, approval_stage_instances)
//  2. approval_instances (FK -> approval_routes; approval_stage_instances
//     cascades from approval_instances automatically)
//  3. approval_route_stage_selectors (FK -> approval_route_stages, itself a
//     child of approval_routes — erased before the routes delete)
//  4. approval_routes (blocked by enforce_route_immutable only while a
//     referencing approval_instances row exists — already gone by step 4)
//  +. approval_review_verdicts (actor_tenant_id, no FK) and
//     approval_delegations (tenant_id, no FK, independent) — unconstrained order
//
// Cross-module FK note (orchestrator concern): approval_signoffs and
// approval_instances both FK to metaldocs.iam_users(tenant_id, user_id) —
// the iam port's EraseTenantData (which deletes iam_users) MUST run AFTER
// this port in the orchestrator's fan-out sequence, or iam's delete would
// fail with a dangling-reference violation in the other direction (actually
// the opposite: deleting iam_users first would violate THIS port's rows
// referencing it — so documents/approval erase must run BEFORE iam erase).
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the documents/approval TenantDataPort backed
// by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "documents-approval" }

func (p *TenantDataPort) Tables() []string {
	return []string{
		"public.approval_instances",
		"public.approval_routes",
		"public.approval_route_stage_selectors",
		"public.approval_signoffs",
		"public.approval_review_verdicts",
		"public.approval_delegations",
	}
}

func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	instances, err := tenantdata.ExportTable(ctx, db, "public.approval_instances", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	routes, err := tenantdata.ExportTable(ctx, db, "public.approval_routes", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	selectors, err := tenantdata.ExportTable(ctx, db, "public.approval_route_stage_selectors", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	signoffs, err := tenantdata.ExportTable(ctx, db, "public.approval_signoffs", "actor_tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	verdicts, err := tenantdata.ExportTable(ctx, db, "public.approval_review_verdicts", "actor_tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	delegations, err := tenantdata.ExportTable(ctx, db, "public.approval_delegations", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return []tenantdata.TableExport{instances, routes, selectors, signoffs, verdicts, delegations}, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	n, err := tenantdata.EraseTable(ctx, tx, "public.approval_signoffs", "actor_tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.approval_signoffs"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "public.approval_review_verdicts", "actor_tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.approval_review_verdicts"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "public.approval_instances", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.approval_instances"] = n

	// Child of approval_route_stages (child of approval_routes) — erase before
	// the routes delete. Own tenant_id predicate.
	n, err = tenantdata.EraseTable(ctx, tx, "public.approval_route_stage_selectors", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.approval_route_stage_selectors"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "public.approval_delegations", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.approval_delegations"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "public.approval_routes", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.approval_routes"] = n

	return counts, nil
}
