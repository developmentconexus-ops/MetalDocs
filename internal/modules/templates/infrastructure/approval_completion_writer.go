package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	approvaldomain "metaldocs/internal/modules/approval/domain"
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
//
// The CAS carries BOTH preconditions the transition depends on — status
// 'draft' AND a committed content_hash. The hash predicate is not a
// duplicated check: under Read Committed a concurrent edit can clear
// content_hash between the caller's fast-path read and this UPDATE, and
// without it the row would flip to under_review and only then trip
// chk_template_version_content_hash_non_draft, aborting the tx with a raw
// 23514 (a 500 leaking the constraint name). Folding the predicate into the
// CAS makes transition and precondition atomic; the IS NOT NULL leg is
// defensive belt-and-braces (content_hash is NOT NULL today, but a NULL would
// silently satisfy the CHECK, which evaluates to NULL rather than false).
func (w *ApprovalCompletionWriter) MarkTemplateVersionUnderReview(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error {
	if err := (&domain.TemplateVersion{Status: domain.VersionStatusDraft}).CanTransition(domain.VersionStatusUnderReview, true); err != nil {
		return fmt.Errorf("MarkTemplateVersionUnderReview: %w", err)
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE templates_template_version
		   SET status       = 'under_review',
		       submitted_at = now(),
		       lock_version = lock_version + 1
		 WHERE id           = $1
		   AND tenant_id    = $2
		   AND status       = 'draft'
		   AND content_hash IS NOT NULL
		   AND content_hash <> ''`,
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
		return w.classifySubmitLockMiss(ctx, tx, tenantID, templateVersionID)
	}
	return nil
}

// classifySubmitLockMiss disambiguates a 0-row submit-lock CAS inside the SAME
// transaction: the CAS now has two preconditions, so "no rows" alone no longer
// identifies which one failed. A still-draft row with no committed hash is the
// content case (approval's ErrTemplateVersionNoContent sentinel, which the
// delivery layer maps to a clean 409); everything else — a non-draft status, an
// absent id, a cross-tenant id — keeps the pre-existing stale/conflict error
// verbatim. The CAS did not raise, so the tx is intact and this re-read is safe.
func (w *ApprovalCompletionWriter) classifySubmitLockMiss(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error {
	var status, contentHash sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT status, content_hash
		  FROM templates_template_version
		 WHERE id        = $1
		   AND tenant_id = $2`,
		templateVersionID, tenantID,
	).Scan(&status, &contentHash)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTemplateVersionNotDraft
	}
	if err != nil {
		return fmt.Errorf("MarkTemplateVersionUnderReview: classify submit-lock miss: %w", err)
	}
	if status.Valid && status.String == string(domain.VersionStatusDraft) &&
		(!contentHash.Valid || contentHash.String == "") {
		return approvaldomain.ErrTemplateVersionNoContent
	}
	return ErrTemplateVersionNotDraft
}

// ErrApproverUserIDRequired is returned (fail-closed, no-fallback) when the
// approval kernel drives a terminal approval without a deciding-actor id.
// approver_id is attribution on a regulated record: an absent value is a
// wiring fault to surface, never a NULL or placeholder to write (F-E4-4).
var ErrApproverUserIDRequired = errors.New("templates: approver user id required for approval completion")

// MarkTemplateVersionApproved transitions templateVersionID from
// under_review to approved and stamps approverUserID onto the row's
// approver_id column (F-E4-4 — approved_at without attribution was the live-QA
// defect). The friendly first-line check mirrors the
// docsdomain.CanTransitionDocumentStatus call site in decision_service.go —
// a static sanity check that the move is a legal system transition; the CAS
// WHERE clause below is the atomic, authoritative enforcement.
//
// reviewer_id is deliberately NOT written here: it means "who reviewed", a
// review-kind stage outcome the approval kernel tracks in its own verdict
// ledger, and the approving actor is not that person. Writing the approver
// into reviewer_id would also silently widen the publish segregation-of-duties
// set (application/lifecycle.go CheckSegregation), which is a behavior change,
// not attribution.
func (w *ApprovalCompletionWriter) MarkTemplateVersionApproved(ctx context.Context, tx db.Tx, tenantID, templateVersionID, approverUserID string) error {
	if err := (&domain.TemplateVersion{Status: domain.VersionStatusUnderReview}).CanTransition(domain.VersionStatusApproved, true); err != nil {
		return fmt.Errorf("MarkTemplateVersionApproved: %w", err)
	}
	if strings.TrimSpace(approverUserID) == "" {
		return ErrApproverUserIDRequired
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE templates_template_version
		   SET status      = 'approved',
		       approved_at = now(),
		       approver_id = $3
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'under_review'`,
		templateVersionID, tenantID, approverUserID,
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
