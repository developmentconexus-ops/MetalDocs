package application

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/platform/db"
)

type ReconstructionRunner interface {
	Reconstruct(ctx context.Context, tenantID, revisionID string) (fanout.ReconstructionEntry, error)
}

type ReconstructionService struct {
	txRunner db.TxRunner
	runner   ReconstructionRunner
}

func NewReconstructionService(txRunner db.TxRunner, runner ReconstructionRunner) *ReconstructionService {
	return &ReconstructionService{txRunner: txRunner, runner: runner}
}

func (s *ReconstructionService) GetReconstruction(ctx context.Context, tenantID, actorID, docID string) (fanout.ReconstructionEntry, error) {
	ctx = authz.WithCapCache(ctx)
	var entry fanout.ReconstructionEntry
	if err := s.txRunner.DoReadOnly(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
			return err
		}

		// document.edit is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := LoadDocumentAreaCode(ctx, tx, tenantID, docID)
		if err != nil {
			return fmt.Errorf("reconstruct authz: load area: %w", err)
		}
		// ADR 0022 Phase 10 (F2): the redundant doc.reconstruct cap was merged into
		// the canonical CapDocumentEdit — identical grant set, same area-grade check.
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return err
		}

		var reconstructErr error
		entry, reconstructErr = s.runner.Reconstruct(ctx, tenantID, docID)
		if reconstructErr != nil {
			return reconstructErr
		}
		return nil
	}); err != nil {
		return fanout.ReconstructionEntry{}, err
	}
	return entry, nil
}
