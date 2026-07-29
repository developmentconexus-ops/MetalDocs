package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// ErrDocumentHasNoCurrentRevision is returned (fail-closed) when a document
// reaches terminal approval without a current_revision_id. The release
// generation identity (ADR 0085) is keyed on the revision, and the
// materialization pipeline itself fails closed on a document with no current
// revision (documents/infrastructure/snapshot_repository.go
// ReadCurrentRevisionBodyKey), so a missing revision here is an impossible
// state, not a legitimate "not yet" case — it must never be substituted for.
var ErrDocumentHasNoCurrentRevision = errors.New("approval: document has no current revision; cannot key a release generation")

// ErrReleaseGenerationNotFound is returned when a coordinator evaluation is
// delivered for a generation identity that has no row. Non-retryable: the
// generation was never created (or was erased with its tenant).
var ErrReleaseGenerationNotFound = errors.New("approval: release generation not found")

// ErrReleaseGenerationMismatch is returned by the generation-guarded artifact
// write when the artifact belongs to a DIFFERENT generation than the one the
// document currently carries. ADR 0085: "the final-artifact write carries the
// expected generation identity and refuses to overwrite a different
// generation's artifact."
var ErrReleaseGenerationMismatch = errors.New("approval: artifact write refused; generation is not the document's current freeze head")

// ReleaseEvaluationEnqueuer enqueues a coordinator evaluation for one release
// generation inside the caller's transaction (transactional outbox). Both fact
// producers enqueue through it in the SAME transaction as the fact write, so a
// committed fact always has an evaluation behind it.
//
// RunAt is the ADR 0085 timer: zero means "evaluate now"; a future instant arms
// the effective-date timer. Firing just invokes the same idempotent evaluation.
type ReleaseEvaluationEnqueuer interface {
	EnqueueReleaseEvaluationTx(ctx context.Context, tx db.Tx, key domain.ReleaseGenerationKey, runAt time.Time) error
}

// ApprovalFactInput carries everything the terminal-approval transaction knows
// about the generation it is closing.
type ApprovalFactInput struct {
	TenantID           string
	DocumentID         string
	ApprovalInstanceID string
	// RevisionVersion is the instance's revision_version at terminal approval —
	// identity, not an OCC token.
	RevisionVersion int
	// FrozenContentHash is the instance's pinned freeze hash. Mandatory: the
	// generation identity is meaningless without it.
	FrozenContentHash string
	// FinalApproverID / SubmittedBy are the CAUSAL identities ADR 0085 requires
	// the release governance event to carry in its payload (the event's own
	// actor is the system principal).
	FinalApproverID string
	SubmittedBy     string
}

// releaseGenerationRow is the in-tx snapshot of one public.release_generations
// row.
type releaseGenerationRow struct {
	ID                 string
	GenerationSeq      int64
	TenantID           string
	DocumentID         string
	ApprovalInstanceID string
	RevisionID         string
	RevisionVersion    int
	FrozenContentHash  string
	ApprovalFactAt     sql.NullTime
	FinalApproverID    sql.NullString
	SubmittedBy        sql.NullString
	FinalDocxS3Key     sql.NullString
	FinalPDFS3Key      sql.NullString
	ArtifactFactAt     sql.NullTime
	HoldReason         sql.NullString
	ReleasedAt         sql.NullTime
}

// Key rebuilds the ADR 0085 generation identity from a loaded row.
func (r releaseGenerationRow) Key() domain.ReleaseGenerationKey {
	return domain.ReleaseGenerationKey{
		TenantID:           r.TenantID,
		SubjectKind:        domain.SubjectKindDocument,
		DocumentID:         r.DocumentID,
		ApprovalInstanceID: r.ApprovalInstanceID,
		RevisionID:         r.RevisionID,
		RevisionVersion:    r.RevisionVersion,
		FrozenContentHash:  r.FrozenContentHash,
	}
}

const releaseGenerationColumns = `id::text, generation_seq, tenant_id::text, document_id::text,
       approval_instance_id::text, revision_id::text, revision_version, frozen_content_hash,
       approval_fact_at, final_approver_id, submitted_by,
       final_docx_s3_key, final_pdf_s3_key, artifact_fact_at,
       hold_reason, released_at`

func scanReleaseGeneration(row interface{ Scan(...any) error }) (releaseGenerationRow, error) {
	var g releaseGenerationRow
	err := row.Scan(
		&g.ID, &g.GenerationSeq, &g.TenantID, &g.DocumentID,
		&g.ApprovalInstanceID, &g.RevisionID, &g.RevisionVersion, &g.FrozenContentHash,
		&g.ApprovalFactAt, &g.FinalApproverID, &g.SubmittedBy,
		&g.FinalDocxS3Key, &g.FinalPDFS3Key, &g.ArtifactFactAt,
		&g.HoldReason, &g.ReleasedAt,
	)
	return g, err
}

// RecordApprovalFactTx creates (or idempotently re-reads) the release
// generation for a terminal approval and stamps its approval fact — all inside
// the caller's terminal-approval transaction. It returns the generation row id
// so the caller can thread it into the materialization pipeline (the artifact
// fact keys on the same generation).
//
// Fail-closed by construction (ADR 0085): there is no best-effort branch. Any
// error propagates and rolls the terminal-approval transaction back, so a
// committed approval always has a committed approval fact.
//
// Idempotent: ON CONFLICT on the generation identity re-stamps nothing (the
// original approval_fact_at wins) and returns the original row id, so a
// replayed signoff or a fast-forward that re-enters this path is a no-op.
func RecordApprovalFactTx(ctx context.Context, tx db.Tx, in ApprovalFactInput) (string, domain.ReleaseGenerationKey, error) {
	if in.TenantID == "" || in.DocumentID == "" || in.ApprovalInstanceID == "" {
		return "", domain.ReleaseGenerationKey{}, fmt.Errorf("recordApprovalFact: tenant/document/instance required")
	}
	if in.FrozenContentHash == "" {
		return "", domain.ReleaseGenerationKey{}, fmt.Errorf("recordApprovalFact: frozen content hash required")
	}

	// The generation identity is keyed on the document's CURRENT revision — the
	// authored body that was frozen and that materialization will render.
	var revisionID sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT current_revision_id::text
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		in.DocumentID, in.TenantID,
	).Scan(&revisionID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ReleaseGenerationKey{}, fmt.Errorf("recordApprovalFact: document %s not found", in.DocumentID)
	}
	if err != nil {
		return "", domain.ReleaseGenerationKey{}, fmt.Errorf("recordApprovalFact: load current revision: %w", err)
	}
	if !revisionID.Valid || revisionID.String == "" {
		return "", domain.ReleaseGenerationKey{}, ErrDocumentHasNoCurrentRevision
	}

	var id string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO release_generations
		    (tenant_id, subject_kind, document_id, approval_instance_id,
		     revision_id, revision_version, frozen_content_hash,
		     approval_fact_at, final_approver_id, submitted_by)
		VALUES ($1::uuid, 'document', $2::uuid, $3::uuid,
		        $4::uuid, $5, $6, now(), NULLIF($7, ''), NULLIF($8, ''))
		ON CONFLICT ON CONSTRAINT ux_release_generations_identity DO UPDATE
		   SET approval_fact_at   = COALESCE(release_generations.approval_fact_at, EXCLUDED.approval_fact_at),
		       final_approver_id  = COALESCE(release_generations.final_approver_id, EXCLUDED.final_approver_id),
		       submitted_by       = COALESCE(release_generations.submitted_by, EXCLUDED.submitted_by),
		       updated_at         = now()
		RETURNING id::text`,
		in.TenantID, in.DocumentID, in.ApprovalInstanceID,
		revisionID.String, in.RevisionVersion, in.FrozenContentHash,
		in.FinalApproverID, in.SubmittedBy,
	).Scan(&id)
	if err != nil {
		return "", domain.ReleaseGenerationKey{}, fmt.Errorf("recordApprovalFact: upsert release generation: %w", err)
	}
	return id, domain.ReleaseGenerationKey{
		TenantID:           in.TenantID,
		SubjectKind:        domain.SubjectKindDocument,
		DocumentID:         in.DocumentID,
		ApprovalInstanceID: in.ApprovalInstanceID,
		RevisionID:         revisionID.String,
		RevisionVersion:    in.RevisionVersion,
		FrozenContentHash:  in.FrozenContentHash,
	}, nil
}

// LoadReleaseGenerationByIDTx reads one generation row by its id, without
// locking. Used by the artifact-fact producer, which locks via the row it then
// updates.
func LoadReleaseGenerationByIDTx(ctx context.Context, tx db.Tx, tenantID, generationID string) (releaseGenerationRow, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT `+releaseGenerationColumns+`
		  FROM release_generations
		 WHERE id = $1::uuid AND tenant_id = $2::uuid
		 FOR UPDATE`,
		generationID, tenantID,
	)
	g, err := scanReleaseGeneration(row)
	if errors.Is(err, sql.ErrNoRows) {
		return releaseGenerationRow{}, ErrReleaseGenerationNotFound
	}
	if err != nil {
		return releaseGenerationRow{}, fmt.Errorf("load release generation: %w", err)
	}
	return g, nil
}

// loadReleaseGenerationByKeyTx locks and reads the generation matching the ADR
// 0085 identity carried in a coordinator job payload.
func loadReleaseGenerationByKeyTx(ctx context.Context, tx db.Tx, key domain.ReleaseGenerationKey) (releaseGenerationRow, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT `+releaseGenerationColumns+`
		  FROM release_generations
		 WHERE tenant_id            = $1::uuid
		   AND subject_kind         = $2
		   AND document_id          = $3::uuid
		   AND approval_instance_id = $4::uuid
		   AND revision_id          = $5::uuid
		   AND revision_version     = $6
		   AND frozen_content_hash  = $7
		 FOR UPDATE`,
		key.TenantID, string(key.SubjectKind), key.DocumentID,
		key.ApprovalInstanceID, key.RevisionID, key.RevisionVersion, key.FrozenContentHash,
	)
	g, err := scanReleaseGeneration(row)
	if errors.Is(err, sql.ErrNoRows) {
		return releaseGenerationRow{}, ErrReleaseGenerationNotFound
	}
	if err != nil {
		return releaseGenerationRow{}, fmt.Errorf("load release generation by key: %w", err)
	}
	return g, nil
}

// isFreezeHeadTx implements the ADR 0085 predicate conjunct "G is still D's
// freeze head (no newer generation superseded it)". generation_seq gives the
// total order; a newer generation exists exactly when the document was
// returned to draft and re-approved.
func isFreezeHeadTx(ctx context.Context, tx db.Tx, g releaseGenerationRow) (bool, error) {
	var newer bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM release_generations
		     WHERE tenant_id      = $1::uuid
		       AND document_id    = $2::uuid
		       AND generation_seq > $3
		)`,
		g.TenantID, g.DocumentID, g.GenerationSeq,
	).Scan(&newer)
	if err != nil {
		return false, fmt.Errorf("release generation head check: %w", err)
	}
	return !newer, nil
}

// ArtifactKind distinguishes the two halves of the artifact-ready(G) set.
type ArtifactKind string

// ArtifactKind values.
const (
	ArtifactFinalDocx ArtifactKind = "final_docx"
	ArtifactFinalPDF  ArtifactKind = "final_pdf"
)

// ArtifactFactInput is one half of the artifact set, recorded in the same
// transaction as the artifact's own persistence write.
type ArtifactFactInput struct {
	TenantID            string
	ReleaseGenerationID string
	Kind                ArtifactKind
	S3Key               string
}

// ArtifactFactResult reports whether this write COMPLETED the artifact set
// (and therefore stamped artifact_fact_at and enqueued a coordinator
// evaluation).
type ArtifactFactResult struct {
	// GenerationKey is the identity the caller may log / re-evaluate on.
	GenerationKey domain.ReleaseGenerationKey
	// Complete is true when the full DOCX+PDF set now exists for G.
	Complete bool
	// AlreadyComplete is true when artifact_fact_at was already stamped before
	// this call (idempotent replay).
	AlreadyComplete bool
}

// RecordArtifactFactTx records one artifact pointer against its generation and,
// when that completes the full DOCX+PDF set, stamps artifact_fact_at and
// enqueues a coordinator evaluation — all in the caller's artifact-persistence
// transaction (ADR 0085: "emitting the artifact fact is part of the
// artifact-persistence transaction").
//
// Generation-guarded: the write is refused (ErrReleaseGenerationMismatch) when
// the generation is no longer the document's freeze head, so a late artifact
// from a superseded generation can never overwrite the current one's facts.
//
// Idempotent: re-recording the same pointer is a no-op; artifact_fact_at is
// stamped once (COALESCE) and the evaluation is only enqueued on the
// transition to complete.
func RecordArtifactFactTx(ctx context.Context, tx db.Tx, enqueuer ReleaseEvaluationEnqueuer, in ArtifactFactInput) (ArtifactFactResult, error) {
	if in.TenantID == "" || in.ReleaseGenerationID == "" {
		return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: tenant/generation required")
	}
	if in.S3Key == "" {
		return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: %s s3 key must not be empty", in.Kind)
	}

	g, err := LoadReleaseGenerationByIDTx(ctx, tx, in.TenantID, in.ReleaseGenerationID)
	if err != nil {
		return ArtifactFactResult{}, err
	}
	head, err := isFreezeHeadTx(ctx, tx, g)
	if err != nil {
		return ArtifactFactResult{}, err
	}
	if !head {
		return ArtifactFactResult{}, fmt.Errorf("%w (generation %s, document %s)",
			ErrReleaseGenerationMismatch, g.ID, g.DocumentID)
	}

	alreadyComplete := g.ArtifactFactAt.Valid

	var column string
	switch in.Kind {
	case ArtifactFinalDocx:
		column = "final_docx_s3_key"
	case ArtifactFinalPDF:
		column = "final_pdf_s3_key"
	default:
		return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: unknown artifact kind %q", in.Kind)
	}

	// Write this half. The generation row is already locked FOR UPDATE, so the
	// read-back below cannot race another artifact writer.
	//
	// column comes from the closed set matched immediately above — never from
	// caller input — so the concatenation is not an injection surface.
	if _, err := tx.ExecContext(ctx, `
		UPDATE release_generations
		   SET `+column+` = $1, updated_at = now()
		 WHERE id = $2::uuid`,
		in.S3Key, g.ID,
	); err != nil {
		return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: write %s pointer: %w", in.Kind, err)
	}

	// Stamp artifact_fact_at iff the FULL set now exists. COALESCE keeps the
	// first stamp: a replay never moves the fact's timestamp.
	var artifactFactAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		UPDATE release_generations
		   SET artifact_fact_at = COALESCE(artifact_fact_at, now()),
		       updated_at       = now()
		 WHERE id = $1::uuid
		   AND final_docx_s3_key IS NOT NULL
		   AND final_pdf_s3_key  IS NOT NULL
		RETURNING artifact_fact_at`,
		g.ID,
	).Scan(&artifactFactAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Set still incomplete — the other half has not landed yet. Not an
		// error: the coordinator holds on `materializing` until it does.
		return ArtifactFactResult{GenerationKey: g.Key()}, nil
	}
	if err != nil {
		return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: stamp artifact fact: %w", err)
	}

	result := ArtifactFactResult{
		GenerationKey:   g.Key(),
		Complete:        artifactFactAt.Valid,
		AlreadyComplete: alreadyComplete,
	}

	// Enqueue the coordinator evaluation only on the transition to complete —
	// in this same transaction, so a committed artifact fact always has an
	// evaluation behind it.
	if result.Complete && !alreadyComplete {
		if enqueuer == nil {
			return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: release evaluation enqueuer not configured")
		}
		if err := enqueuer.EnqueueReleaseEvaluationTx(ctx, tx, g.Key(), time.Time{}); err != nil {
			return ArtifactFactResult{}, fmt.Errorf("recordArtifactFact: enqueue evaluation: %w", err)
		}
	}
	return result, nil
}
