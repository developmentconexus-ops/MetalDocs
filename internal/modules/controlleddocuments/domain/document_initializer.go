package domain

import (
	"context"
	"errors"
	"strings"

	"metaldocs/internal/platform/db"
)

var (
	// ErrCloneTemplateNameRequired indicates a clone request was built without
	// the document name required by the downstream documents module.
	ErrCloneTemplateNameRequired = errors.New("clone template request name must not be empty")

	// ErrDictionaryTokenMissing signals a template references a dictionary token
	// that does not exist for the tenant at creation/revision time. Maps to 422
	// (SP-2 D7).
	ErrDictionaryTokenMissing = errors.New("controlled_documents: referenced dictionary token not found")
)

// CloneTemplateRequest carries the user-supplied bits of an atomic CD-create
// flow that the documents module needs to materialize an initial document
// revision alongside the new ControlledDocument row.
type CloneTemplateRequest struct {
	templateVersionID *string
	name              string
	formData          map[string]any
	dictionaryValues  map[string]string
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

// DictionaryValues returns the resolved {placeholderID -> pinned value} map for
// PHDictionary references, resolved off-tx by the caller (SP-2 D1). May be nil.
func (r CloneTemplateRequest) DictionaryValues() map[string]string { return r.dictionaryValues }

// WithDictionaryValues returns a copy of the request carrying resolved dictionary
// placeholder values to pin at creation. Kept off NewCloneTemplateRequest to avoid
// churning its many existing call sites.
func (r CloneTemplateRequest) WithDictionaryValues(v map[string]string) CloneTemplateRequest {
	r.dictionaryValues = v
	return r
}

// DocumentRef is the minimal handle the registry returns to callers after a
// successful atomic create. The registry stores no document state itself —
// downstream code uses this to redirect to the editor or to enrich responses.
type DocumentRef struct {
	ID          string `json:"id"`
	ContentHash string `json:"content_hash"`
	// RevisionID and SessionID carry the seeded document's initial revision and
	// editor session back to internal callers (documents.DuplicateDocument).
	// json:"-" keeps them OUT of the CD atomic-create JSON response so that
	// contract is unchanged.
	RevisionID string `json:"-"`
	SessionID  string `json:"-"`
}

// DocumentInitializer is the controlled-documents-owned port that the documents module
// implements. It MUST run inside the caller-owned tx so the CD row and its
// initial document rows commit atomically. S3 rendering is intentionally NOT
// part of this contract: storage_key starts empty and the editor renders on
// demand.
type DocumentInitializer interface {
	CloneTemplate(ctx context.Context, tx db.Tx, cd *ControlledDocument, req CloneTemplateRequest) (*DocumentRef, error)
	ResolveTemplateStorageKey(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error)
	Exists(ctx context.Context, storageKey string) (bool, error)
	// ResolveTemplateVersionID resolves the effective template version id for the
	// (profile, optional override) OFF-TX, before the caller opens its atomic tx.
	// Callers thread the returned concrete id back into CloneTemplate so the in-tx
	// clone performs no authz-recording taxonomy read (which would deadlock against
	// the audit hash-chain advisory lock held by the atomic tx).
	ResolveTemplateVersionID(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error)
	// ResolveDictionaryValues resolves every PHDictionary placeholder in the given
	// template version's schema to its pinned value (keyed by placeholder ID),
	// OFF-TX before the caller opens its atomic tx (the dictionary read is
	// authz-recording on its own tx — H-PRE-1). Returns ErrDictionaryTokenMissing
	// when a referenced token does not exist (SP-2 D7/D10).
	ResolveDictionaryValues(ctx context.Context, tenantID, templateVersionID string) (map[string]string, error)
}
