package domain

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	// ErrCloneTemplateNameRequired indicates a clone request was built without
	// the document name required by the downstream documents module.
	ErrCloneTemplateNameRequired = errors.New("clone template request name must not be empty")
	// ErrDocumentRefIDRequired indicates a document reference was built without
	// its durable document identifier.
	ErrDocumentRefIDRequired = errors.New("document ref id must not be empty")
)

// CloneTemplateRequest carries the user-supplied bits of an atomic CD-create
// flow that the documents module needs to materialize an initial document
// revision alongside the new ControlledDocument row.
type CloneTemplateRequest struct {
	templateVersionID *string
	name              string
	formData          map[string]any
}

// NewCloneTemplateRequest validates and normalizes the request sent through the
// DocumentInitializer port.
func NewCloneTemplateRequest(templateVersionID *string, name string, formData map[string]any) (CloneTemplateRequest, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return CloneTemplateRequest{}, ErrCloneTemplateNameRequired
	}
	if formData == nil {
		formData = map[string]any{}
	}
	return CloneTemplateRequest{
		templateVersionID: trimOptionalString(templateVersionID),
		name:              trimmedName,
		formData:          formData,
	}, nil
}

func (r CloneTemplateRequest) TemplateVersionID() *string { return r.templateVersionID }

func (r CloneTemplateRequest) Name() string { return r.name }

func (r CloneTemplateRequest) FormData() map[string]any { return r.formData }

// DocumentRef is the minimal handle the registry returns to callers after a
// successful atomic create. The registry stores no document state itself —
// downstream code uses this to redirect to the editor or to enrich responses.
type DocumentRef struct {
	ID          string `json:"id"`
	ContentHash string `json:"contentHash"`
}

// NewDocumentRef validates and normalizes the minimal document handle returned
// by the documents module.
func NewDocumentRef(id, contentHash string) (*DocumentRef, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return nil, ErrDocumentRefIDRequired
	}
	return &DocumentRef{
		ID:          trimmedID,
		ContentHash: strings.TrimSpace(contentHash),
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
