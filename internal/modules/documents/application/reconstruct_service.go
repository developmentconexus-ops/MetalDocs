package application

import (
	"context"
	"database/sql"
	"fmt"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/platform/db"
)

// ReconstructionRunner performs the actual document reconstruction against
// the render fanout pipeline.
type ReconstructionRunner interface {
	Reconstruct(ctx context.Context, tenantID, revisionID string) (fanout.ReconstructionEntry, error)
}

// ReconstructionService authorizes and drives document reconstruction: it
// checks document.edit off-tx, then delegates the reconstruction itself to
// the wired ReconstructionRunner.
type ReconstructionService struct {
	txRunner db.TxRunner
	runner   ReconstructionRunner
	cdRead   controlleddocumentsdomain.CDFieldReader
}

// NewReconstructionService constructs a ReconstructionService wired to its
// tx runner, reconstruction runner, and controlled-document field reader
// (used to resolve the area for the document.edit authz check).
func NewReconstructionService(txRunner db.TxRunner, runner ReconstructionRunner, cdRead controlleddocumentsdomain.CDFieldReader) *ReconstructionService {
	return &ReconstructionService{txRunner: txRunner, runner: runner, cdRead: cdRead}
}

// GetReconstruction authorizes actorID for document.edit on docID's area,
// then reconstructs and returns the document entry.
func (s *ReconstructionService) GetReconstruction(ctx context.Context, tenantID, actorID, docID string) (fanout.ReconstructionEntry, error) {
	ctx = authz.WithCapCache(ctx)

	// Authz runs in a short read-only tx; the renderer HTTP round-trip must NOT hold
	// a DB connection open across the network call, so it runs after the tx closes
	// (mirrors the materialize runner, which calls the renderer off-tx too).
	// Read-write tx: authz.Require may audit a system_admin bypass (ADR 0022 F8),
	// which INSERTs — see the Require INVARIANT note in iam/authz/authz.go.
	if err := s.txRunner.Do(ctx, func(tx *sql.Tx) error {
		// document.edit is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := LoadDocumentAreaCode(ctx, tx, s.cdRead, tenantID, docID)
		if err != nil {
			return fmt.Errorf("reconstruct authz: load area: %w", err)
		}
		// ADR 0022 Phase 10 (F2): the redundant doc.reconstruct cap was merged into
		// the canonical CapDocumentEdit — identical grant set, same area-grade check.
		return authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode)
	}); err != nil {
		return fanout.ReconstructionEntry{}, err
	}

	entry, err := s.runner.Reconstruct(ctx, tenantID, docID)
	if err != nil {
		return fanout.ReconstructionEntry{}, err
	}
	return entry, nil
}
