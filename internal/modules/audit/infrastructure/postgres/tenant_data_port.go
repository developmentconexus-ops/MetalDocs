package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the audit module's M7 F7.3 TenantDataPort. It owns
// metaldocs.audit_events, metaldocs.audit_export_jobs, and (per the F7.3
// spec interview #2a) public.governance_events — a trigger-adjacent event
// table with no other Go module claiming it; audit is its closest semantic
// owner (event records) and enforces its export/erasure disposition here.
//
// audit_events is the one non-uniform table in this port:
//   - Export is a SKELETON projection only (id, occurred_at, actor_id,
//     action, resource_type, resource_id, trace_id, tenant_id,
//     audit_sequence, prev_hash, row_hash) — never the `payload` column.
//     §0.3 marks payload as the only PII surface; until F7.3's crypto
//     envelope work (Task B, separate task) lands, payload may still be
//     plaintext for pre-existing rows, so the export path here never
//     serializes it — a stricter behavior than "envelope only", chosen
//     because this task does not implement or depend on the crypto seam.
//   - EraseTenantData deletes NOTHING for audit_events and reports 0 for it:
//     rows are append-only (REVOKE UPDATE, DELETE, TRUNCATE grant, §0.3);
//     crypto-shredding the DEK (a different feature) is what makes payload
//     unrecoverable, not a row-level operation this port performs.
//
// audit_export_jobs and governance_events are ordinary tenant tables: full
// row export, full row delete (governance_events before audit_export_jobs
// in erase order is not FK-required — both are independent leaves — order
// is fixed only for determinism).
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the audit TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "audit" }

func (p *TenantDataPort) Tables() []string {
	return []string{
		"metaldocs.audit_events",
		"metaldocs.audit_export_jobs",
		"public.governance_events",
	}
}

// auditEventSkeletonColumns are the non-PII columns of metaldocs.audit_events
// (§0.3) — the only columns this port ever exports for that table.
const auditEventSkeletonColumns = `id, occurred_at, actor_id, action, resource_type, resource_id,
       trace_id, tenant_id, audit_sequence, prev_hash, row_hash`

func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	skeleton, err := p.exportAuditEventSkeleton(ctx, db, tenantID)
	if err != nil {
		return nil, err
	}

	exportJobs, err := tenantdata.ExportTable(ctx, db, "metaldocs.audit_export_jobs", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}

	govEvents, err := tenantdata.ExportTable(ctx, db, "public.governance_events", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}

	return []tenantdata.TableExport{skeleton, exportJobs, govEvents}, nil
}

// exportAuditEventSkeleton runs the PII-excluding projection over
// metaldocs.audit_events. tenant_id on this table is TEXT (§0.1), not uuid.
func (p *TenantDataPort) exportAuditEventSkeleton(ctx context.Context, db *sql.DB, tenantID string) (tenantdata.TableExport, error) {
	query := fmt.Sprintf(
		`SELECT row_to_json(t) FROM (SELECT %s FROM metaldocs.audit_events WHERE tenant_id = $1 ORDER BY audit_sequence) t`,
		auditEventSkeletonColumns,
	)
	rows, err := db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return tenantdata.TableExport{}, fmt.Errorf("tenantdata: export audit_events skeleton: %w", err)
	}
	defer rows.Close()

	out := tenantdata.TableExport{Table: "metaldocs.audit_events", Rows: []json.RawMessage{}}
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return tenantdata.TableExport{}, fmt.Errorf("tenantdata: scan audit_events skeleton row: %w", err)
		}
		out.Rows = append(out.Rows, raw)
	}
	if err := rows.Err(); err != nil {
		return tenantdata.TableExport{}, fmt.Errorf("tenantdata: iterate audit_events skeleton rows: %w", err)
	}
	return out, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := map[string]int64{
		// audit_events: crypto-shred handles the payload (a separate
		// feature); rows are immutable and this port deletes NOTHING here.
		"metaldocs.audit_events": 0,
	}

	n, err := tenantdata.EraseTable(ctx, tx, "public.governance_events", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.governance_events"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.audit_export_jobs", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.audit_export_jobs"] = n

	return counts, nil
}
