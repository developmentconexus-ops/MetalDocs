package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/platform/db"
)

// ArtifactFactWriter adapts the approval module's release-generation facts to
// the render workers' consumer-owned port
// (internal/platform/worker.ArtifactFactRecorder). It is the ONLY seam through
// which the render pipeline touches release state — the workers never write
// public.release_generations themselves.
type ArtifactFactWriter struct {
	enqueuer ReleaseEvaluationEnqueuer
}

// NewArtifactFactWriter builds the adapter. The enqueuer is mandatory: an
// artifact fact with no evaluation behind it would leave the document stuck on
// the `materializing` hold forever.
func NewArtifactFactWriter(enqueuer ReleaseEvaluationEnqueuer) *ArtifactFactWriter {
	return &ArtifactFactWriter{enqueuer: enqueuer}
}

// RecordArtifactFactTx satisfies worker.ArtifactFactRecorder.
func (w *ArtifactFactWriter) RecordArtifactFactTx(ctx context.Context, tx db.Tx, tenantID, releaseGenerationID, kind, s3Key string) error {
	if w.enqueuer == nil {
		return fmt.Errorf("artifact fact writer: release evaluation enqueuer not configured")
	}
	_, err := RecordArtifactFactTx(ctx, tx, w.enqueuer, ArtifactFactInput{
		TenantID:            tenantID,
		ReleaseGenerationID: releaseGenerationID,
		Kind:                ArtifactKind(kind),
		S3Key:               s3Key,
	})
	return err
}

// derefString reads an optional string field. Used for the instance's
// frozen_content_hash, which is *string on the domain model; the terminal
// path treats an empty result as the impossible-state failure it is rather
// than substituting anything.
func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// TerminalApprovalInput is everything the shared terminal-approval path needs.
// Both routes to terminal approval — signoff (DecisionService) and review
// verdict (ReviewVerdictService) — go through it, which is what makes the two
// paths structurally identical rather than merely similar (F-QA4-14).
type TerminalApprovalInput struct {
	TenantID          string
	DocumentID        string
	InstanceID        string
	RevisionVersion   int
	FrozenContentHash string
	FinalApproverID   string
	SubmittedBy       string
	Approver          docapp.ApproverContext
}

// TerminalApprovalReleaseRecorder is the ONE seam both terminal-approval
// routes cross to produce the ADR 0085 approval fact. It exists as an
// interface (rather than the two ports it composes) for two reasons:
//
//   - the composition root wires ONE collaborator, so the signoff and
//     review-verdict routes cannot be wired asymmetrically — the wiring
//     asymmetry is exactly what F-QA4-14 was;
//   - it keeps the release substrate's SQL out of the decision services'
//     unit tests, which drive hand-rolled fake drivers.
//
// It is mandatory on the document terminal-approval path: publication is the
// coordinator's reaction to the approval fact, so a terminal approval with no
// recorder behind it is a wiring defect, surfaced as an error, never a
// silently unpublishable document.
type TerminalApprovalReleaseRecorder interface {
	RecordTerminalApproval(ctx context.Context, tx db.Tx, in TerminalApprovalInput) (string, error)
}

// ReleaseFactRecorder is the production TerminalApprovalReleaseRecorder: it
// composes the async-freeze seam (ADR 0015 PinInvoker) with the coordinator
// evaluation port (ADR 0085) and runs both on the caller's transaction.
type ReleaseFactRecorder struct {
	pinInvoker PinInvoker
	enqueuer   ReleaseEvaluationEnqueuer
}

// NewReleaseFactRecorder builds the production recorder. Both collaborators
// are mandatory; a nil one fails the terminal approval closed rather than
// degrading it.
func NewReleaseFactRecorder(pinInvoker PinInvoker, enqueuer ReleaseEvaluationEnqueuer) *ReleaseFactRecorder {
	return &ReleaseFactRecorder{pinInvoker: pinInvoker, enqueuer: enqueuer}
}

// RecordTerminalApproval satisfies TerminalApprovalReleaseRecorder.
func (r *ReleaseFactRecorder) RecordTerminalApproval(ctx context.Context, tx db.Tx, in TerminalApprovalInput) (string, error) {
	if r == nil {
		return "", fmt.Errorf("terminal approval: release fact recorder not configured")
	}
	return recordTerminalApprovalRelease(ctx, tx, r.pinInvoker, r.enqueuer, in)
}

// recordTerminalApprovalRelease is the single implementation of "an approval
// instance reached terminal approval" (ADR 0085 §"Approval fact"):
//
//  1. open the release generation and stamp the approval fact;
//  2. pin the document (async-freeze boundary, ADR 0015), threading the
//     generation so the artifacts it produces key back to it;
//  3. fast-forward any artifacts that already exist for this content;
//  4. enqueue the coordinator evaluation.
//
// All of it runs in the CALLER's terminal-approval transaction, with no
// best-effort branch anywhere: any failure rolls the approval back. A
// committed terminal approval therefore always has a committed approval fact
// and a queued evaluation — the property the whole coordinator rests on.
func recordTerminalApprovalRelease(
	ctx context.Context,
	tx db.Tx,
	pinInvoker PinInvoker,
	enqueuer ReleaseEvaluationEnqueuer,
	in TerminalApprovalInput,
) (string, error) {
	if pinInvoker == nil {
		return "", fmt.Errorf("terminal approval: pinInvoker not configured")
	}
	if enqueuer == nil {
		return "", fmt.Errorf("terminal approval: release evaluation enqueuer not configured")
	}
	if in.FrozenContentHash == "" {
		// The generation identity is the frozen hash. An instance reaching
		// terminal approval without a pin is an impossible state (the freeze
		// boundary sets it); fail closed rather than key a generation on "".
		return "", ErrContentHashMismatch
	}

	generationID, generationKey, err := RecordApprovalFactTx(ctx, tx, ApprovalFactInput{
		TenantID:           in.TenantID,
		DocumentID:         in.DocumentID,
		ApprovalInstanceID: in.InstanceID,
		RevisionVersion:    in.RevisionVersion,
		FrozenContentHash:  in.FrozenContentHash,
		FinalApproverID:    in.FinalApproverID,
		SubmittedBy:        in.SubmittedBy,
	})
	if err != nil {
		return "", fmt.Errorf("terminal approval: record approval fact: %w", err)
	}

	// Pin's second parameter is named revisionID but carries the DOCUMENT id
	// throughout this chain (documents.id is the aggregate the freeze columns
	// live on) — preserved verbatim here, not "corrected", because the whole
	// pin/fanout/worker chain agrees on it.
	if err := pinInvoker.Pin(ctx, tx, in.TenantID, in.DocumentID, in.Approver, generationID); err != nil {
		return "", fmt.Errorf("terminal approval: pin: %w", err)
	}

	// Fast-forward: when the document was already materialized for this exact
	// content (Pin is idempotent and returns early on an already-frozen
	// revision), no render job will fire and no artifact fact would ever
	// arrive. Recording the existing pointers here keeps the generation from
	// parking on `materializing` forever.
	if err := fastForwardExistingArtifacts(ctx, tx, enqueuer, in.TenantID, in.DocumentID, generationID); err != nil {
		return "", err
	}

	if err := enqueuer.EnqueueReleaseEvaluationTx(ctx, tx, generationKey, time.Time{}); err != nil {
		return "", fmt.Errorf("terminal approval: enqueue release evaluation: %w", err)
	}
	return generationID, nil
}

// fastForwardExistingArtifacts records artifact facts for final artifacts that
// already exist on the document.
func fastForwardExistingArtifacts(ctx context.Context, tx db.Tx, enqueuer ReleaseEvaluationEnqueuer, tenantID, documentID, generationID string) error {
	var docxKey, pdfKey sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT final_docx_s3_key, final_pdf_s3_key
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		documentID, tenantID,
	).Scan(&docxKey, &pdfKey)
	if err != nil {
		return fmt.Errorf("terminal approval: read existing artifacts: %w", err)
	}
	for _, a := range []struct {
		kind ArtifactKind
		key  sql.NullString
	}{
		{ArtifactFinalDocx, docxKey},
		{ArtifactFinalPDF, pdfKey},
	} {
		if !a.key.Valid || a.key.String == "" {
			continue
		}
		if _, err := RecordArtifactFactTx(ctx, tx, enqueuer, ArtifactFactInput{
			TenantID:            tenantID,
			ReleaseGenerationID: generationID,
			Kind:                a.kind,
			S3Key:               a.key.String,
		}); err != nil {
			return fmt.Errorf("terminal approval: fast-forward %s fact: %w", a.kind, err)
		}
	}
	return nil
}
