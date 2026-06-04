package application

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/render/fanout"
)

type ReconstructionRunner interface {
	Reconstruct(ctx context.Context, tenantID, revisionID string) (fanout.ReconstructionEntry, error)
}

type ReconstructionService struct {
	db     *sql.DB
	runner ReconstructionRunner
}

func NewReconstructionService(db *sql.DB, runner ReconstructionRunner) *ReconstructionService {
	return &ReconstructionService{db: db, runner: runner}
}

func (s *ReconstructionService) GetReconstruction(ctx context.Context, tenantID, actorID, docID string) (fanout.ReconstructionEntry, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fanout.ReconstructionEntry{}, fmt.Errorf("reconstruct authz: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return fanout.ReconstructionEntry{}, err
	}

	// document.edit is area-grade: pass the resolved area as-is ("" fail-closes).
	areaCode, _, err := LoadDocumentAreaCode(ctx, tx, tenantID, docID)
	if err != nil {
		return fanout.ReconstructionEntry{}, fmt.Errorf("reconstruct authz: load area: %w", err)
	}
	// ADR 0022 Phase 10 (F2): the redundant doc.reconstruct cap was merged into
	// the canonical CapDocumentEdit — identical grant set, same area-grade check.
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
		return fanout.ReconstructionEntry{}, err
	}

	entry, err := s.runner.Reconstruct(ctx, tenantID, docID)
	if err != nil {
		return fanout.ReconstructionEntry{}, err
	}
	return entry, nil
}
