package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

// ErrTemplateVersionNotUnderReview is returned (fail-closed, no-fallback)
// when a completion-port CAS UPDATE affects zero rows — the version was not
// in the expected under_review status for this tenant+id (already terminal,
// concurrently transitioned, or the id/tenant pair does not exist).
var ErrTemplateVersionNotUnderReview = errors.New("templates: version not under_review for approval completion")

// ErrTemplateVersionNotDraft is returned (fail-closed) when the submit-lock
// CAS UPDATE affects zero rows — the version was not in draft status for this
// tenant+id when kernel submit tried to lock it into under_review (already
// submitted, concurrently transitioned, or the id/tenant pair does not exist).
var ErrTemplateVersionNotDraft = errors.New("templates: version not draft for approval submit-lock")

// ApprovalCompletionWriter is the templates-side adapter for the approval
// module's TemplateCompletionWriter port (M3 P3.S2b-3b-iii-b). It is the
// ONLY surface through which approval's DecisionService writes a terminal
// approval-kernel outcome back onto templates_template_version — approval
// never imports templates infrastructure or writes the table directly.
//
// The two transitions this writes (under_review -> approved,
// under_review -> draft) are exactly the target statuses templates' own
// TemplateVersion.CanTransition already models; the approval kernel is now
// the sole actor driving this specific transition for a template instance
// governed by an approval route.
type ApprovalCompletionWriter struct{}

// NewApprovalCompletionWriter constructs the adapter. It is stateless — the
// tx is supplied per call by the caller (approval's TxRunner-owned tx), so
// the version transition commits atomically with the signoff write.
func NewApprovalCompletionWriter() *ApprovalCompletionWriter {
	return &ApprovalCompletionWriter{}
}

// MarkTemplateVersionUnderReview transitions templateVersionID from draft to
// under_review as part of the approval kernel's submit tx (M3
// P3.S2b-3b-iii-b, hub Option (a) resolution). This is the submit-lock that
// closes the concurrent-edit hole: the templates module's own edit/upload
// gates (autosave.go, PresignTemplateUpload) permit writes ONLY while
// status='draft', so flipping to under_review here makes those endpoints
// reject edits for the duration of the approval. It commits atomically with
// InsertInstance/InsertStageInstances in the same tx, so a version can never
// be "submitted to the kernel" yet still author-editable.
//
// The tripwire on templates_template_version accepts any of
// {template.create,edit,submit,approve,publish}; the caller
// (TemplateSubmitService) has already asserted template.submit earlier in
// this same tx via authz.Require, so the GUC accumulation authorizes this
// UPDATE — no additional assertion is needed here.
func (w *ApprovalCompletionWriter) MarkTemplateVersionUnderReview(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error {
	if err := (&domain.TemplateVersion{Status: domain.VersionStatusDraft}).CanTransition(domain.VersionStatusUnderReview, true); err != nil {
		return fmt.Errorf("MarkTemplateVersionUnderReview: %w", err)
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE templates_template_version
		   SET status       = 'under_review',
		       submitted_at = now(),
		       lock_version = lock_version + 1
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'draft'`,
		templateVersionID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionUnderReview: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionUnderReview: rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTemplateVersionNotDraft
	}
	return nil
}

// MarkTemplateVersionApproved transitions templateVersionID from
// under_review to approved. The friendly first-line check mirrors the
// docsdomain.CanTransitionDocumentStatus call site in decision_service.go —
// a static sanity check that the move is a legal system transition; the CAS
// WHERE clause below is the atomic, authoritative enforcement.
func (w *ApprovalCompletionWriter) MarkTemplateVersionApproved(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error {
	if err := (&domain.TemplateVersion{Status: domain.VersionStatusUnderReview}).CanTransition(domain.VersionStatusApproved, true); err != nil {
		return fmt.Errorf("MarkTemplateVersionApproved: %w", err)
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE templates_template_version
		   SET status      = 'approved',
		       approved_at = now()
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'under_review'`,
		templateVersionID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionApproved: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionApproved: rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTemplateVersionNotUnderReview
	}
	return nil
}

// MarkTemplateVersionRejected transitions templateVersionID from
// under_review back to draft so its author can edit and resubmit — the same
// under_review -> draft arc a document rejection drives.
func (w *ApprovalCompletionWriter) MarkTemplateVersionRejected(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error {
	if err := (&domain.TemplateVersion{Status: domain.VersionStatusUnderReview}).CanTransition(domain.VersionStatusDraft, true); err != nil {
		return fmt.Errorf("MarkTemplateVersionRejected: %w", err)
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE templates_template_version
		   SET status = 'draft'
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'under_review'`,
		templateVersionID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionRejected: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionRejected: rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTemplateVersionNotUnderReview
	}
	return nil
}
