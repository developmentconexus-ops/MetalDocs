package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// ApprovalVersionReader is the templates-side adapter for the approval
// module's TemplateVersionReader port (M3 P3.S2b-3b-ii). It is the ONLY
// surface through which approval's TemplateSubmitService reads
// templates_template_version status — approval never imports this package's
// sibling types or queries the table directly, preserving the module
// boundary (approval -> templates ONLY via an approval-owned injected
// interface). Unlike TemplateVersionReader above (used by documents/
// controlleddocuments, off-tx via the shared *sql.DB pool), this adapter
// runs its read on the CALLER's transaction (db.Tx) so the draft-status
// check participates in approval's submit tx exactly like every other
// in-tx read on that path (mirrors SubmitDefaultsResolver.
// LoadControlledDocumentID's tx-scoped shape).
type ApprovalVersionReader struct{}

// NewApprovalVersionReader constructs the adapter. It is stateless — the tx
// is supplied per call by the caller (approval's TxRunner-owned tx).
func NewApprovalVersionReader() *ApprovalVersionReader {
	return &ApprovalVersionReader{}
}

// LoadTemplateVersionStatus returns templates_template_version.status for
// templateVersionID, scoped to tenantID, read inside tx. ok is false when no
// row matches (absent id, or a cross-tenant id — the WHERE clause itself
// enforces tenant scoping, no reliance on RLS alone).
func (r *ApprovalVersionReader) LoadTemplateVersionStatus(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (string, bool, error) {
	var status sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT status
		  FROM templates_template_version
		 WHERE id = $1
		   AND tenant_id = $2`,
		templateVersionID, tenantID,
	).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !status.Valid {
		return "", false, nil
	}
	return status.String, true, nil
}

// LoadTemplateVersionContentHash returns templates_template_version.content_hash
// for templateVersionID, scoped to tenantID, read inside tx (M3
// P3.S2b-3b-iii-b). Used by approval's DecisionService in place of the
// document-only freeze pin when the signoff's parent instance is
// subject_kind='template'. ok is false when no row matches, or the row has
// no content_hash yet (empty string — never populated).
func (r *ApprovalVersionReader) LoadTemplateVersionContentHash(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (string, bool, error) {
	var hash sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT content_hash
		  FROM templates_template_version
		 WHERE id = $1
		   AND tenant_id = $2`,
		templateVersionID, tenantID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !hash.Valid || hash.String == "" {
		return "", false, nil
	}
	return hash.String, true, nil
}

// LoadTemplateInboxMeta (Unit 4.2) batch-resolves {templateId, name} for each
// of versionIDs, scoped to tenantID, read inside tx. One set-based query
// (tv.id = ANY($1)) joining templates_template_version -> templates_template
// — the ONLY surface through which approval's worklist read-model resolves
// template display data; approval never queries these tables directly. A
// version id with no matching row (absent, or cross-tenant) is simply absent
// from the returned map.
func (r *ApprovalVersionReader) LoadTemplateInboxMeta(ctx context.Context, tx db.Tx, tenantID string, versionIDs []string) (map[string]domain.TemplateInboxMeta, error) {
	out := make(map[string]domain.TemplateInboxMeta)
	if len(versionIDs) == 0 {
		return out, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT tv.id, t.id, t.name
		  FROM templates_template_version tv
		  JOIN templates_template t
		    ON t.id = tv.template_id
		   AND t.tenant_id = tv.tenant_id
		 WHERE tv.tenant_id = $1
		   AND tv.id = ANY($2)`,
		tenantID, pq.Array(versionIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var versionID, templateID, name string
		if err := rows.Scan(&versionID, &templateID, &name); err != nil {
			return nil, err
		}
		out[versionID] = domain.TemplateInboxMeta{TemplateID: templateID, Title: name}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ErrTemplateDocTypeCodeMissing is returned by LoadTemplateDocTypeCode when the
// template does not exist for the tenant, or exists with an empty
// doc_type_code. Both are defects post-ADR-0086 (migration 0315's
// templates_template_doc_type_code_required_check makes the empty case
// unreachable for author-created templates), and both must fail closed: the
// doc_type_code is the template ROUTE's subject key, so any substitute value
// would silently route the submission through another profile's route.
var ErrTemplateDocTypeCodeMissing = errors.New("templates: template has no doc_type_code")

// LoadTemplateDocTypeCode returns templates_template.doc_type_code for
// templateID, scoped to tenantID, read inside tx (ADR 0086). It is the route
// governance selector for a template submit/preview: ROUTE.subject_key =
// doc_type_code.
func (r *ApprovalVersionReader) LoadTemplateDocTypeCode(ctx context.Context, tx db.Tx, tenantID, templateID string) (string, error) {
	var code sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT doc_type_code
		  FROM templates_template
		 WHERE id = $1
		   AND tenant_id = $2`,
		templateID, tenantID,
	).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTemplateDocTypeCodeMissing
	}
	if err != nil {
		return "", err
	}
	if !code.Valid || code.String == "" {
		return "", ErrTemplateDocTypeCodeMissing
	}
	return code.String, nil
}
