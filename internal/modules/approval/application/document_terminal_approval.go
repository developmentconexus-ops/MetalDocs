package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docapp "metaldocs/internal/modules/documents/application"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// instanceStatusWriter is the narrow persistence slice the terminal-approval
// completion needs — a subset of infrastructure.ApprovalRepository, declared
// here so the helper's real dependency surface stays honest (same discipline as
// freezeRepo in freeze.go).
type instanceStatusWriter interface {
	UpdateInstanceStatus(ctx context.Context, tx db.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time) error
}

// documentTerminalApprovalPorts bundles the collaborators every route to
// terminal approval on a DOCUMENT subject needs. releaseRecorder is mandatory
// (ADR 0085: publication is the coordinator's reaction to the approval fact, so
// a terminal approval with no recorder behind it is a wiring defect, never a
// silently unpublishable document); lifecycleEnqueuer is optional, matching the
// pre-existing nil-guarded emit sites.
type documentTerminalApprovalPorts struct {
	repo              instanceStatusWriter
	releaseRecorder   TerminalApprovalReleaseRecorder
	lifecycleEnqueuer docsdomain.LifecycleEventEnqueuer
	// serviceName names the caller in the wiring-defect error so the operator
	// sees which composition-root leg is unwired.
	serviceName string
}

// documentTerminalApprovalInput is the per-call state of "this approval
// instance just reached terminal approval on a document".
type documentTerminalApprovalInput struct {
	TenantID   string
	InstanceID string
	DocumentID string
	// AreaCode is the document's resolved area for the area-grade
	// document.edit assertion. "" fail-closes for non-system actors.
	AreaCode          string
	RevisionVersion   int
	FrozenContentHash string
	FinalApproverID   string
	SubmittedBy       string
	// ApproverCapabilities rides into docapp.ApproverContext. The signoff route
	// carries the deciding actor's capability set; the review-verdict and
	// auto-approve routes carry none, exactly as before this extraction.
	ApproverCapabilities []string
	// FromStatus is the CAS precondition on approval_instances.status. Every
	// current caller passes domain.InstanceInProgress.
	FromStatus domain.InstanceStatus
	Now        time.Time
}

// completeDocumentTerminalApproval is the SINGLE implementation of "an approval
// instance reached terminal approval on a document subject". It runs entirely
// inside the caller's transaction and performs, in order:
//
//  1. approval_instances in_progress → approved (CAS);
//  2. the area-grade document.edit assertion authorizing the documents write;
//  3. the ADR 0085 approval fact + pin + coordinator evaluation
//     (TerminalApprovalReleaseRecorder — the publication trigger);
//  4. documents under_review → approved (OCC on the current status);
//  5. the ADR 0044 document.approved lifecycle event (notifications fanout).
//
// Three routes reach terminal approval — signoff (DecisionService), review
// verdict (ReviewVerdictService) and the ADR 0087 zero-stage auto-approve
// (SubmitService). They share this function rather than three hand-rolled
// copies: F-QA4-14 was exactly the defect of one route silently omitting the
// publication side of the transition, and a third copy would reopen that class.
func completeDocumentTerminalApproval(
	ctx context.Context,
	tx *sql.Tx,
	ports documentTerminalApprovalPorts,
	in documentTerminalApprovalInput,
) error {
	if ports.repo == nil {
		return fmt.Errorf("%s missing required ports: repo", ports.serviceName)
	}
	if ports.releaseRecorder == nil {
		return fmt.Errorf("%s missing required ports: releaseRecorder", ports.serviceName)
	}

	if err := ports.repo.UpdateInstanceStatus(ctx, tx, in.TenantID, in.InstanceID,
		domain.InstanceApproved, in.FromStatus, &in.Now); err != nil {
		return fmt.Errorf("terminal approval: complete instance: %w", err)
	}

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), in.AreaCode); err != nil {
		return err
	}

	// ADR 0085: approval fact + pin + coordinator evaluation, all in THIS
	// transaction and all fail-closed.
	if _, err := ports.releaseRecorder.RecordTerminalApproval(ctx, tx, TerminalApprovalInput{
		TenantID:          in.TenantID,
		DocumentID:        in.DocumentID,
		InstanceID:        in.InstanceID,
		RevisionVersion:   in.RevisionVersion,
		FrozenContentHash: in.FrozenContentHash,
		FinalApproverID:   in.FinalApproverID,
		SubmittedBy:       in.SubmittedBy,
		Approver: docapp.ApproverContext{
			UserID:       in.FinalApproverID,
			Capabilities: in.ApproverCapabilities,
		},
	}); err != nil {
		return err
	}

	// Transition document under_review → approved. Friendly first-line legality
	// check (M4/F4.1) mirrors the DB trigger; the OCC WHERE below remains the
	// atomic CAS enforcement.
	if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusUnderReview, docsdomain.DocStatusApproved); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET status           = 'approved',
		       revision_version = revision_version + 1
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'under_review'`,
		in.DocumentID, in.TenantID,
	)
	if err != nil {
		return fmt.Errorf("terminal approval: approve document: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("terminal approval: approve document rows affected: %w", err)
	}
	if rows == 0 {
		return infrastructure.ErrStaleRevision
	}

	// Additive in-tx domain-event enqueue (ADR 0044; F3.3) — terminal
	// transitions only.
	if ports.lifecycleEnqueuer != nil {
		if err := ports.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, docsdomain.LifecycleEventArgs{
			EventID:      uuid.NewString(),
			TenantID:     in.TenantID,
			EventType:    docsdomain.EventTypeDocumentApproved,
			ResourceType: "approval_instance",
			ResourceID:   in.InstanceID,
			SubmittedBy:  in.SubmittedBy,
			OccurredAt:   in.Now,
		}); err != nil {
			return fmt.Errorf("terminal approval: enqueue lifecycle event: %w", err)
		}
	}
	return nil
}
