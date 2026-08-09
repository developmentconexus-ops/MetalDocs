package domain

import (
	"crypto/sha256"
	"errors"
)

// ErrSnapshotTemplateNotFound is returned when the template version a
// document's snapshot should be resolved from cannot be found.
var ErrSnapshotTemplateNotFound = errors.New("snapshot_template_not_found")

// TemplateSnapshot holds the template artifacts copied onto a document at create time.
type TemplateSnapshot struct {
	PlaceholderSchemaJSON []byte
	CompositionJSON       []byte
	BodyDocxBytes         []byte
	BodyDocxS3Key         string
}

// RevisionRef identifies one document_revisions row together with everything
// the freeze pipeline needs to render and verify it: where the body lives and
// what its authored bytes are supposed to hash to.
//
// ID is document_revisions.id — NOT the documents.id that the freeze pipeline's
// `revisionID` parameter has always carried (a long-standing misnomer, see
// FreezeService.RepairMaterialization's note). The frozen_revision_id pin is
// this ID.
//
// ContentHash is document_revisions.content_hash: lowercase sha256 hex of the
// revision body bytes, established at upload time by
// objectstore.VerifiedStore.Confirm. An empty ID means the document has no such
// revision, and every caller must fail closed rather than substitute another
// body (no-fallback principle).
type RevisionRef struct {
	ID          string
	StorageKey  string
	ContentHash string
}

// SnapshotHashes holds the sha256 digests of each snapshot field.
type SnapshotHashes struct {
	PlaceholderSchemaHash []byte
	CompositionHash       []byte
	BodyDocxHash          []byte
}

// Hashes computes deterministic sha256 digests for the snapshot fields.
func (s TemplateSnapshot) Hashes() SnapshotHashes {
	ph := sha256.Sum256(s.PlaceholderSchemaJSON)
	ch := sha256.Sum256(s.CompositionJSON)
	bh := sha256.Sum256(s.BodyDocxBytes)
	return SnapshotHashes{ph[:], ch[:], bh[:]}
}
