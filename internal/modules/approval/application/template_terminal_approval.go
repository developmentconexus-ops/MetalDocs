package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"metaldocs/internal/modules/approval/domain"
)

// templateTerminalApprovalPorts bundles the collaborators every route to
// terminal approval on a TEMPLATE subject needs. Both are mandatory: a template
// terminal approval with no completion writer behind it would flip the approval
// instance while leaving templates_template_version stuck in under_review —
// exactly the split-brain the port exists to prevent.
type templateTerminalApprovalPorts struct {
	repo               instanceStatusWriter
	templateCompletion TemplateCompletionWriter
	// serviceName names the caller in the wiring-defect error so the operator
	// sees which composition-root leg is unwired.
	serviceName string
}

// templateTerminalApprovalInput is the per-call state of "this approval
// instance just reached terminal approval on a template version".
type templateTerminalApprovalInput struct {
	TenantID   string
	InstanceID string
	// TemplateVersionID is the instance's subject key (the artifact under
	// approval), not the owning template.
	TemplateVersionID string
	// ApproverID is stamped onto the version's approver_id (F-E4-4: approved_at
	// never lands without attribution). On the ADR 0087 auto-approve route it is
	// the submitter — on a zero-stage route the submission IS the approving act.
	ApproverID string
	// FromStatus is the CAS precondition on approval_instances.status.
	FromStatus domain.InstanceStatus
	Now        time.Time
}

// completeTemplateTerminalApproval is the SINGLE path a template-subject
// approval instance takes to terminal approval, whatever drove it there: the
// last signoff (DecisionService) or an ADR 0087 zero-stage route
// (TemplateSubmitService's auto-approve). It exists for the same reason
// completeDocumentTerminalApproval does — the F-QA4-14 defect class, where one
// route to terminal approval silently omits the completion side effect and the
// subject is left stranded mid-lifecycle.
//
// Order is deliberate: the instance CAS runs first, so a lost CAS (someone else
// already completed this instance) aborts before the version transition is
// attempted. Templates never freeze and never pin, and there is no documents
// transition or release-coordinator leg — a template version's publication is
// the templates module's own concern.
func completeTemplateTerminalApproval(
	ctx context.Context,
	tx *sql.Tx,
	ports templateTerminalApprovalPorts,
	in templateTerminalApprovalInput,
) error {
	if ports.repo == nil || ports.templateCompletion == nil {
		return fmt.Errorf("%s missing required ports: repo=%v templateCompletion=%v",
			ports.serviceName, ports.repo != nil, ports.templateCompletion != nil)
	}

	if err := ports.repo.UpdateInstanceStatus(ctx, tx, in.TenantID, in.InstanceID,
		domain.InstanceApproved, in.FromStatus, &in.Now); err != nil {
		return fmt.Errorf("%s: complete instance: %w", ports.serviceName, err)
	}

	// Drives templates_template_version under_review -> approved atomically in
	// this same tx (M3 P3.S2b-3b-iii-b).
	if err := ports.templateCompletion.MarkTemplateVersionApproved(ctx, tx, in.TenantID, in.TemplateVersionID, in.ApproverID); err != nil {
		return fmt.Errorf("%s: mark template version approved: %w", ports.serviceName, err)
	}
	return nil
}
