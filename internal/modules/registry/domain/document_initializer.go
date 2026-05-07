package domain

import (
	"context"
	"database/sql"
)

// CloneTemplateRequest carries the user-supplied bits of an atomic CD-create
// flow that the documents module needs to materialize an initial document
// revision alongside the new ControlledDocument row.
type CloneTemplateRequest struct {
	TemplateVersionID *string
	Name              string
	FormData          map[string]any
}

// DocumentRef is the minimal handle the registry returns to callers after a
// successful atomic create. The registry stores no document state itself —
// downstream code uses this to redirect to the editor or to enrich responses.
type DocumentRef struct {
	ID          string `json:"id"`
	ContentHash string `json:"contentHash"`
}

// DocumentInitializer is the registry-owned port that the documents module
// implements. It MUST run inside the caller-owned tx so the CD row and its
// initial document rows commit atomically. S3 rendering is intentionally NOT
// part of this contract: storage_key starts empty and the editor renders on
// demand.
type DocumentInitializer interface {
	CloneTemplate(ctx context.Context, tx *sql.Tx, cd *ControlledDocument, req CloneTemplateRequest) (*DocumentRef, error)
}
