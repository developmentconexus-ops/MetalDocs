package domain

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// CloneTemplateRequest carries the user-supplied bits of an atomic CD-create
// flow that the documents module needs to materialize an initial document
// revision alongside the new ControlledDocument row.
type CloneTemplateRequest struct {
	TemplateVersionID *string
	Name              string
	FormData          map[string]any
}

var ErrCloneTemplateRequestNameRequired = errors.New("controlled_documents: clone template request name is required")

func NewCloneTemplateRequest(templateVersionID *string, name string, formData map[string]any) (CloneTemplateRequest, error) {
	if strings.TrimSpace(name) == "" {
		return CloneTemplateRequest{}, ErrCloneTemplateRequestNameRequired
	}
	return CloneTemplateRequest{
		TemplateVersionID: templateVersionID,
		Name:              name,
		FormData:          formData,
	}, nil
}

// DocumentRef is the minimal handle the registry returns to callers after a
// successful atomic create. The registry stores no document state itself —
// downstream code uses this to redirect to the editor or to enrich responses.
type DocumentRef struct {
	ID          string `json:"id"`
	ContentHash string `json:"contentHash"`
}

var ErrDocumentRefIDRequired = errors.New("controlled_documents: document ref id is required")

func NewDocumentRef(id, contentHash string) (DocumentRef, error) {
	if strings.TrimSpace(id) == "" {
		return DocumentRef{}, ErrDocumentRefIDRequired
	}
	return DocumentRef{
		ID:          id,
		ContentHash: contentHash,
	}, nil
}

// DocumentInitializer is the controlled-documents-owned port that the documents module
// implements. It MUST run inside the caller-owned tx so the CD row and its
// initial document rows commit atomically. S3 rendering is intentionally NOT
// part of this contract: storage_key starts empty and the editor renders on
// demand.
type DocumentInitializer interface {
	CloneTemplate(ctx context.Context, tx *sql.Tx, cd *ControlledDocument, req CloneTemplateRequest) (*DocumentRef, error)
	ResolveTemplateStorageKey(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error)
	Exists(ctx context.Context, storageKey string) (bool, error)
}
