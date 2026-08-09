package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/tenant"
)

// Type aliases so handlers depend only on application types.

// PendingCommitMeta describes a reserved-but-not-yet-committed autosave
// upload, as returned by Repository.GetPendingForCommit.
type PendingCommitMeta = infrastructure.PendingCommitMeta

// CommitResult is the outcome of committing (or syncing metadata for) an
// autosave upload: the resulting revision identity and whether the call was
// an idempotent replay.
type CommitResult = infrastructure.CommitResult

// RestoreResult is the outcome of restoring a checkpoint: the new revision
// created from it, and whether the restore was an idempotent replay.
type RestoreResult = infrastructure.RestoreResult

// RevisionHistoryItem is one entry in a document's revision history listing.
type RevisionHistoryItem = domain.RevisionHistoryItem

// Repository is the persistence port Service uses for all document,
// revision, session, checkpoint, and comment reads/writes.
type Repository interface {
	CreateDocumentTx(ctx context.Context, tx db.Tx, d *domain.Document, initialContentHash, initialStorageKey string, requiredPlaceholders []templatesdomain.Placeholder) (docID, revID, sessionID string, err error)
	SeedDictionaryValuesTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, values map[string]string) error
	GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error)
	UpdateDocumentName(ctx context.Context, tenantID, actorID, docID, name string) error
	UpdateDocumentNameTx(ctx context.Context, tx db.Tx, tenantID, actorID, docID, name string) error
	ListDocuments(ctx context.Context, tenantID string) ([]domain.Document, error)
	ListDocumentsForUser(ctx context.Context, tenantID, userID string) ([]domain.Document, error)
	ListDocumentsPaginated(ctx context.Context, tenantID string, opts ListOptions) ([]*domain.Document, int64, bool, error)
	CountDocuments(ctx context.Context, tenantID string, opts ListOptions) (int64, error)
	StatsByStatus(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error)
	StatsByArea(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error)
	UpdateDocumentStatus(ctx context.Context, tenantID, actorID, id string, cur, next domain.DocumentStatus, stampTime bool) error
	MarkArchived(ctx context.Context, tenantID, docID, actorID string) error
	MarkArchivedTx(ctx context.Context, tx db.Tx, tenantID, docID, actorID string) error
	IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error)
	AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, error)
	HeartbeatSession(ctx context.Context, tenantID, sessionID, userID string) error
	ReleaseSession(ctx context.Context, tenantID, sessionID, userID string) error
	ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID string) error
	ForceReleaseSessionTx(ctx context.Context, tx db.Tx, tenantID, adminID, sessionID string) error
	ExpireStaleSessions(ctx context.Context, now time.Time) (int, error)
	PresignReserve(ctx context.Context, tenantID, sessionID, userID, docID, baseRev, contentHash, storageKey string, expiresAt time.Time) (string, error)
	GetPendingForCommit(ctx context.Context, tenantID, pendingID string) (*PendingCommitMeta, error)
	CommitUpload(ctx context.Context, tenantID, sessionID, userID, docID, pendingID, serverComputedHash string, formDataSnapshot []byte, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*CommitResult, error)
	SyncCurrentRevisionArtifactMetadata(ctx context.Context, tenantID, sessionID, userID, docID string, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*CommitResult, error)
	CreateCheckpoint(ctx context.Context, tenantID, docID, actorUserID, label string) (*domain.Checkpoint, error)
	ListCheckpoints(ctx context.Context, tenantID, docID string) ([]domain.Checkpoint, error)
	ListRevisionHistory(ctx context.Context, tenantID, docID string) ([]domain.RevisionHistoryItem, error)
	RestoreCheckpoint(ctx context.Context, tenantID, docID, actorUserID string, versionNum int) (*RestoreResult, error)
	GetRevision(ctx context.Context, tenantID, docID, revID string) (*domain.Revision, error)
	DeleteExpiredPending(ctx context.Context, olderThan time.Time) (int, error)
	CreateComment(ctx context.Context, tenantID, documentID, authorID string, in domain.CommentCreateInput) (*domain.Comment, error)
	ListComments(ctx context.Context, tenantID, documentID string) ([]domain.Comment, error)
	UpdateComment(ctx context.Context, tenantID, documentID string, libraryID int, userID string, in domain.CommentUpdateInput) (*domain.Comment, error)
	DeleteComment(ctx context.Context, tenantID, documentID string, libraryID int) error
}

const (
	revisionUploadTTL = 15 * time.Minute
	objectDownloadTTL = 15 * time.Minute
)

// Presigner presigns and manages revision/document objects in the object
// store: upload/download URLs, upload confirmation, existence, size, and
// deletion.
type Presigner interface {
	PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error)
	Confirm(ctx context.Context, tenantID, key, expectedHash string) (objectstore.VerifiedPointer, error)
	Exists(ctx context.Context, key string) (bool, error)
	Size(ctx context.Context, key string) (int64, error)
	Delete(ctx context.Context, key string) error
}

// TemplateReader reads a published template version's docx/schema artifacts,
// used to seed a new document at creation time.
type TemplateReader interface {
	GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error)
}

// FormValidator validates a document's form-data JSON against a template's
// JSON schema.
type FormValidator interface {
	Validate(schemaJSON string, formData json.RawMessage) (valid bool, errs []string, err error)
}

// Audit records document lifecycle events, either fire-and-forget (Write) or
// transactionally alongside the state change it describes (WriteTx).
type Audit interface {
	Write(ctx context.Context, tenantID, actorID, action, docID string, meta any)
	WriteTx(ctx context.Context, tx db.Tx, tenantID, actorID, action, docID string, meta any) error
}

// ControlledDocumentDuplicator duplicates a controlled document's identity so
// Service.DuplicateDocument can seed a new document from an existing one.
type ControlledDocumentDuplicator interface {
	DuplicateControlledDocument(ctx context.Context, in DuplicateControlledDocumentInput) (*CreateDocumentResult, error)
}

// DuplicateControlledDocumentInput carries the source document's content down to
// the controlled-document duplicator so the SINGLE seeded document is correct.
type DuplicateControlledDocumentInput struct {
	TenantID             string
	ControlledDocumentID string
	ActorUserID          string
	DocumentName         string
	FormData             json.RawMessage
}

// ProfileDefaultTemplateReader resolves a taxonomy profile's default template
// version, used when a document is created without an explicit template
// override.
type ProfileDefaultTemplateReader interface {
	GetDefaultTemplateVersionID(ctx context.Context, tenantID, profileCode string) (*string, *string, error)
	// returns (*templateVersionID, *templateVersionStatus, error)
}

// DictionaryValueReader resolves a tenant dictionary token name to its current
// value. The composition root backs it with the tokens module's published
// DictionaryReader; documents imports NO tokens types (SP-2 §11, invariant #6).
type DictionaryValueReader interface {
	Lookup(ctx context.Context, tenantID, name string) (value string, found bool, err error)
}

// Service is the documents module's application service: document CRUD,
// autosave/commit, sessions, checkpoints, comments, and exports all flow
// through it, enforcing the document.edit/document.view capabilities and the
// draft/under_review write-eligibility rules along the way.
type Service struct {
	repo                         Repository
	presigner                    Presigner
	tpl                          TemplateReader
	fv                           FormValidator
	audit                        Audit
	controlledDocumentDuplicator ControlledDocumentDuplicator
	profileTemplates             ProfileDefaultTemplateReader
	snapshotSvc                  *SnapshotService
	runner                       db.TxRunner
	eligibility                  domain.ApproverEligibilityReader
	dictReader                   DictionaryValueReader
}

// WithRunner attaches the db.TxRunner Service uses for authz-enforcing
// transactions (e.g. RequireDocumentView, ListCheckpoints).
func (s *Service) WithRunner(runner db.TxRunner) *Service {
	s.runner = runner
	return s
}

// WithEligibility wires the ApproverEligibilityReader used by
// mayWriteWorkingContent. Typically called with repo, which already implements
// the interface. Must be called before any autosave or session-acquire path is
// exercised under under_review status.
func (s *Service) WithEligibility(r domain.ApproverEligibilityReader) *Service {
	s.eligibility = r
	return s
}

// WithDictionaryReader injects the tenant-dictionary read port used to pin
// dictionary placeholder values at document creation (SP-2 D1). Nil disables
// dictionary resolution (feature-off safe).
func (s *Service) WithDictionaryReader(r DictionaryValueReader) *Service {
	s.dictReader = r
	return s
}

// ResolveDictionaryValues resolves every PHDictionary placeholder in the given
// template version's schema to its pinned value, keyed by placeholder ID. Runs
// OFF-TX (the dictionary read is authz-recording on its own tx — H-PRE-1). Returns
// domain.ErrDictionaryTokenMissing if a referenced token does not exist (D7);
// authz/infra errors propagate unchanged for the caller to map (403/5xx).
func (s *Service) ResolveDictionaryValues(ctx context.Context, tenantID, templateVersionID string) (map[string]string, error) {
	if s.dictReader == nil || s.snapshotSvc == nil {
		return nil, nil
	}
	phs, err := s.snapshotSvc.ResolveAllPlaceholders(ctx, tenantID, templateVersionID)
	if err != nil {
		return nil, fmt.Errorf("resolve dictionary values: load schema: %w", err)
	}
	out := make(map[string]string)
	for _, p := range phs {
		if p.Type != templatesdomain.PHDictionary {
			continue
		}
		val, found, err := s.dictReader.Lookup(ctx, tenantID, p.Name)
		if err != nil {
			return nil, err // authz/infra — caller maps to 403/5xx
		}
		if !found {
			return nil, fmt.Errorf("%w: %q", domain.ErrDictionaryTokenMissing, p.Name)
		}
		out[p.ID] = val
	}
	return out, nil
}

// mayWriteWorkingContent decides whether actorID may write the working-content
// revision chain of doc. Calls IsDocumentOwner / IsEligibleApprover as plain
// reads — these MUST be called outside any lock-holding or write transaction
// (advisory-lock-deadlock-constraint).
func (s *Service) mayWriteWorkingContent(ctx context.Context, tenantID, docID, actorID string, doc *domain.Document) (bool, error) {
	switch doc.Status {
	case domain.DocStatusDraft:
		owner, err := s.repo.IsDocumentOwner(ctx, tenantID, docID, actorID)
		if err != nil {
			return false, err
		}
		return domain.CanWriteWorkingContent(string(doc.Status), owner, false), nil
	case domain.DocStatusUnderReview:
		if s.eligibility == nil {
			return false, nil
		}
		eligible, err := s.eligibility.IsEligibleApprover(ctx, tenantID, docID, actorID)
		if err != nil {
			return false, err
		}
		return domain.CanWriteWorkingContent(string(doc.Status), false, eligible), nil
	default:
		return false, nil
	}
}

// New constructs a Service wired only with a repo, presigner, template
// reader, form validator, and audit sink (no profile-template resolver or
// snapshot service).
//
// Deprecated: use NewService.
func New(r Repository, p Presigner, t TemplateReader, fv FormValidator, a Audit) *Service {
	return &Service{
		repo:      r,
		presigner: p,
		tpl:       t,
		fv:        fv,
		audit:     a,
	}
}

// NewService constructs a Service wired with a repo, presigner, template
// reader, form validator, audit sink, and profile default-template reader.
func NewService(
	r Repository,
	p Presigner,
	t TemplateReader,
	fv FormValidator,
	a Audit,
	profileTemplates ProfileDefaultTemplateReader,
) *Service {
	return &Service{
		repo:             r,
		presigner:        p,
		tpl:              t,
		fv:               fv,
		audit:            a,
		profileTemplates: profileTemplates,
	}
}

// NewServiceWithSnapshot is like NewService but also wires a SnapshotService
// that will copy template artifacts onto the document at create time.
func NewServiceWithSnapshot(
	r Repository,
	p Presigner,
	t TemplateReader,
	fv FormValidator,
	a Audit,
	profileTemplates ProfileDefaultTemplateReader,
	snap *SnapshotService,
) *Service {
	return &Service{
		repo:             r,
		presigner:        p,
		tpl:              t,
		fv:               fv,
		audit:            a,
		profileTemplates: profileTemplates,
		snapshotSvc:      snap,
	}
}

// WithControlledDocumentDuplicator attaches the ControlledDocumentDuplicator
// used by DuplicateDocument.
func (s *Service) WithControlledDocumentDuplicator(d ControlledDocumentDuplicator) *Service {
	s.controlledDocumentDuplicator = d
	return s
}

// ErrControlledDocumentRequired is returned by DuplicateDocument when the
// source document has no controlled_document_id to duplicate from.
var ErrControlledDocumentRequired = errors.New("controlled_document_id is required")
var errControlledDocumentDuplicatorNotConfigured = errors.New("controlled document duplicator not configured")
var errProfileTemplateReaderNotConfigured = errors.New("profile default template reader not configured")

// CreateDocumentResult identifies the document, initial revision, and editor
// session created by a document-creation flow.
type CreateDocumentResult struct {
	DocumentID        string
	InitialRevisionID string
	SessionID         string
}

// cloneIntoTxInput is the input payload for cloneIntoTx. It carries the
// controlled-document context (tenant, CD identity, profile/area/code, owner)
// together with the document name, initial form-data, and an optional
// override-template hint needed to seed the initial document row inside the
// caller's tx.
type cloneIntoTxInput struct {
	TenantID                  string
	ControlledDocumentID      string
	ProfileCode               string
	ProcessAreaCode           string
	Code                      string
	OverrideTemplateVersionID *string
	OwnerUserID               string
	Name                      string
	FormData                  json.RawMessage
	DictionaryValues          map[string]string
}

// cloneIntoTx seeds a document + initial revision + editor session + snapshot
// columns + required placeholders inside a caller-owned tx. It is the sole
// document-seeding path available to callers such as the controlled-document
// atomic-create flow.
//
//   - Template-passthrough storage: storage_key is set to the template's
//     published docx key atomically with the insert, so the editor opens
//     immediately on first GET (no lazy materialization, no AdoptTempObject).
//   - No S3 rendering: the docx-renderer is not invoked.
//   - No audit/outbox writes: the caller runs post-commit side-effects.
//
// All repo calls thread the tx via CreateDocumentTx.
func (s *Service) cloneIntoTx(ctx context.Context, tx db.Tx, in cloneIntoTxInput) (docID string, revID string, sessionID string, contentHash string, err error) {
	resolvedTemplateVersionID, err := s.resolveTemplateVersionID(ctx, in.TenantID, in.ProfileCode, in.OverrideTemplateVersionID)
	if err != nil {
		return "", "", "", "", err
	}

	docxKey, _, _, err := s.tpl.GetPublishedVersion(ctx, in.TenantID, resolvedTemplateVersionID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("template lookup: %w", err)
	}

	// Resolve template snapshot atomically with the document/revision insert
	// below — pure DB reads, no S3 side-effects.
	var snap domain.TemplateSnapshot
	var phs []templatesdomain.Placeholder
	if s.snapshotSvc != nil {
		snap, phs, err = s.snapshotSvc.ResolveTemplate(ctx, in.TenantID, resolvedTemplateVersionID)
		if err != nil {
			return "", "", "", "", fmt.Errorf("resolve template snapshot: %w", err)
		}
	}

	// F-D6: the creation revision has no rendered docx bytes yet, so there is no
	// real content hash to record. Use the empty-string "not-yet-materialized"
	// sentinel — the exact pattern the canonical sibling spawnNextDraft uses
	// (lifecycle.go:469-475 / NewTemplateVersionDraft leaves ContentHash empty so
	// the publish gate forces a real edit before publish). We must NOT fabricate a
	// hash over the storage-key string: a synthetic sha256(docxKey) is not
	// content-addressable, so it silently occupies a real-looking slot in the
	// UNIQUE(document_id, content_hash) dedup space (RestoreCheckpoint, repository.go
	// ~1481), masquerading as a real content hash even though no bytes exist. (The
	// submit path is unaffected either way: content_hash_at_submit is computed
	// independently in submit_service.go, not read from this column.) NULL is not an option:
	// document_revisions.content_hash is text NOT NULL (baseline ~1955) and NULLs
	// are distinct under the UNIQUE index, which would break dedup; "" stays a
	// single well-defined sentinel and readers already COALESCE it (repo ~1792).
	contentHash = ""

	formDataJSON := in.FormData
	if len(formDataJSON) == 0 {
		formDataJSON = json.RawMessage(`{}`)
	}

	doc := domain.Document{
		TenantID:                in.TenantID,
		TemplateVersionID:       resolvedTemplateVersionID,
		Name:                    in.Name,
		FormDataJSON:            formDataJSON,
		CreatedBy:               in.OwnerUserID,
		ControlledDocumentID:    &in.ControlledDocumentID,
		ProfileCodeSnapshot:     &in.ProfileCode,
		ProcessAreaCodeSnapshot: &in.ProcessAreaCode,
		Code:                    in.Code,
		TemplateSnapshot:        snap,
	}

	docID, revID, sessionID, err = s.repo.CreateDocumentTx(ctx, tx, &doc, contentHash, docxKey, phs)
	if err != nil {
		return "", "", "", "", err
	}
	if len(in.DictionaryValues) > 0 {
		if err := s.repo.SeedDictionaryValuesTx(ctx, tx, in.TenantID, revID, in.DictionaryValues); err != nil {
			return "", "", "", "", fmt.Errorf("seed dictionary values: %w", err)
		}
	}
	return docID, revID, sessionID, contentHash, nil
}

func (s *Service) resolveTemplateVersionID(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error) {
	if s.profileTemplates == nil {
		return "", errProfileTemplateReaderNotConfigured
	}

	var overrideTemplate *controlleddocumentsdomain.TemplateVersionCandidate
	if templateVersionID != nil && strings.TrimSpace(*templateVersionID) != "" {
		overrideStatus := string(templatesdomain.VersionStatusPublished)
		overrideTemplate = &controlleddocumentsdomain.TemplateVersionCandidate{
			ID:          strings.TrimSpace(*templateVersionID),
			ProfileCode: profileCode,
			Status:      &overrideStatus,
		}
	}

	// Only fetch the profile default when no concrete override was supplied. A
	// concrete override always wins in Resolve, so fetching the default is wasted
	// work — and on the atomic CD-create path that fetch issues an authz-recording
	// taxonomy read (GetByCode) on a fresh connection, which deadlocks against the
	// audit hash-chain advisory lock held by the caller's tx. Skipping it keeps the
	// in-tx clone free of off-tx authz reads.
	var defaultTemplate *controlleddocumentsdomain.TemplateVersionCandidate
	if overrideTemplate == nil {
		defaultTemplateID, defaultTemplateStatus, err := s.profileTemplates.GetDefaultTemplateVersionID(ctx, tenantID, profileCode)
		if err != nil {
			return "", err
		}
		if defaultTemplateID != nil {
			defaultTemplate = &controlleddocumentsdomain.TemplateVersionCandidate{
				ID:          *defaultTemplateID,
				ProfileCode: profileCode,
				Status:      defaultTemplateStatus,
			}
		}
	}

	resolution, err := controlleddocumentsdomain.Resolve(controlleddocumentsdomain.TemplateResolutionInput{
		ProfileCode:      profileCode,
		OverrideTemplate: overrideTemplate,
		DefaultTemplate:  defaultTemplate,
	})
	if err != nil {
		return "", err
	}
	return resolution.TemplateVersionID, nil
}

func (s *Service) resolveTemplateStorageKey(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error) {
	resolvedTemplateVersionID, err := s.resolveTemplateVersionID(ctx, tenantID, profileCode, templateVersionID)
	if err != nil {
		return "", err
	}
	docxKey, _, _, err := s.tpl.GetPublishedVersion(ctx, tenantID, resolvedTemplateVersionID)
	if err != nil {
		return "", fmt.Errorf("template lookup: %w", err)
	}
	return docxKey, nil
}

func (s *Service) templateArtifactExists(ctx context.Context, storageKey string) (bool, error) {
	if s.presigner == nil {
		return false, errors.New("document presigner not configured")
	}
	return s.presigner.Exists(ctx, storageKey)
}

// GetDocument returns the document row for id.
func (s *Service) GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	return s.repo.GetDocument(ctx, tenantID, id)
}

// RequireDocumentView asserts the actor holds CapDocumentView (tenant-grade) for
// the given document. document.view is tenant-grade — the "tenant" sentinel keeps
// the area filter intentionally OFF (ADR 0022 Phase 8, same as ViewService).
// RW tx is mandatory: the F8 bypass audit may INSERT (ADR 0022 Phase 11).
//
// docID is part of the published contract but currently unused: the capability is
// tenant-grade, so the check is identical for every document in the tenant. It is
// retained in the signature so the gate can become document-scoped (area-grade)
// without a contract break if document.view is ever reclassified.
func (s *Service) RequireDocumentView(ctx context.Context, tenantID, actorID, docID string) error {
	_ = docID // tenant-grade cap: no per-document filtering today (see doc comment).
	ctx = authz.WithCapCache(ctx)
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		return authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant")
	})
}

// DuplicateDocument duplicates docID's controlled document via the wired
// ControlledDocumentDuplicator, seeding a new document from the source's
// name and form data. Fails with ErrControlledDocumentRequired if the source
// document has no controlled_document_id.
func (s *Service) DuplicateDocument(ctx context.Context, tenantID, userID, docID string) (*CreateDocumentResult, error) {
	if s.controlledDocumentDuplicator == nil {
		return nil, errControlledDocumentDuplicatorNotConfigured
	}
	doc, err := s.repo.GetDocument(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc.ControlledDocumentID == nil || strings.TrimSpace(*doc.ControlledDocumentID) == "" {
		return nil, ErrControlledDocumentRequired
	}
	return s.controlledDocumentDuplicator.DuplicateControlledDocument(ctx, DuplicateControlledDocumentInput{
		TenantID:             tenantID,
		ControlledDocumentID: *doc.ControlledDocumentID,
		ActorUserID:          userID,
		DocumentName:         doc.Name,
		FormData:             doc.FormDataJSON,
	})
}

// DocumentStats reports document counts broken down by status and by area.
type DocumentStats struct {
	ByStatus map[string]int64 `json:"by_status"`
	ByArea   map[string]int64 `json:"by_area"`
}

// ListDocumentsPaginated returns a page of documents matching opts (scoped to
// userID when non-empty) plus the total row count and a has-more flag, taken
// from a single consistent snapshot query.
func (s *Service) ListDocumentsPaginated(ctx context.Context, tenantID, userID string, opts ListOptions) (items []*domain.Document, total int64, hasMore bool, err error) {
	if userID != "" {
		opts.CreatedBy = userID
	}

	// Single-snapshot query returns the page rows and the grand total together,
	// so total is always consistent with the page (no separate-count TOCTOU race).
	items, total, hasMore, err = s.repo.ListDocumentsPaginated(ctx, tenantID, opts)
	if err != nil {
		return nil, 0, false, err
	}

	return items, total, hasMore, nil
}

// DocumentStats returns document counts by status and by area for opts
// (scoped to userID when non-empty).
func (s *Service) DocumentStats(ctx context.Context, tenantID, userID string, opts ListOptions) (*DocumentStats, error) {
	if userID != "" {
		opts.CreatedBy = userID
	}

	byStatus, err := s.repo.StatsByStatus(ctx, tenantID, opts)
	if err != nil {
		return nil, err
	}

	byArea, err := s.repo.StatsByArea(ctx, tenantID, opts)
	if err != nil {
		return nil, err
	}

	return &DocumentStats{
		ByStatus: byStatus,
		ByArea:   byArea,
	}, nil
}

// RenameDocument renames a draft document, rejecting empty/over-length names
// and non-draft documents.
func (s *Service) RenameDocument(ctx context.Context, tenantID, userID, docID, newName string) error {
	name := strings.TrimSpace(newName)
	if name == "" || len(name) > 255 {
		return domain.ErrInvalidName
	}
	doc, err := s.repo.GetDocument(ctx, tenantID, docID)
	if err != nil {
		return err
	}
	if doc.Status != domain.DocStatusDraft {
		return domain.ErrInvalidStateTransition
	}
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.repo.UpdateDocumentNameTx(ctx, tx, tenantID, userID, docID, name); err != nil {
			return err
		}
		if err := s.audit.WriteTx(ctx, tx, tenantID, userID, "document.renamed", docID, map[string]any{"name": name}); err != nil {
			return err
		}
		return nil
	})
}

// IsDocumentOwner reports whether userID owns (created) docID.
func (s *Service) IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error) {
	return s.repo.IsDocumentOwner(ctx, tenantID, docID, userID)
}

// ListDocumentComments returns all comments on documentID.
func (s *Service) ListDocumentComments(ctx context.Context, tenantID, documentID string) ([]domain.Comment, error) {
	return s.repo.ListComments(ctx, tenantID, documentID)
}

// AddDocumentComment validates and creates a new comment on documentID,
// stamping it with the trimmed authorDisplay name.
func (s *Service) AddDocumentComment(ctx context.Context, tenantID, userID, authorDisplay, documentID string, in domain.CommentCreateInput) (*domain.Comment, error) {
	if in.LibraryCommentID <= 0 {
		return nil, domain.ErrCommentInvalid
	}
	if len(in.ContentJSON) == 0 {
		return nil, domain.ErrCommentInvalid
	}
	trimmedAuthor := strings.TrimSpace(authorDisplay)
	if trimmedAuthor == "" || len(trimmedAuthor) > 255 {
		return nil, domain.ErrCommentInvalid
	}
	in.AuthorDisplay = trimmedAuthor
	return s.repo.CreateComment(ctx, tenantID, documentID, userID, in)
}

// UpdateDocumentComment validates and applies an update to the comment
// identified by libraryID on documentID.
func (s *Service) UpdateDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int, in domain.CommentUpdateInput) (*domain.Comment, error) {
	if libraryID <= 0 {
		return nil, domain.ErrCommentInvalid
	}
	if in.ContentJSON != nil && len(*in.ContentJSON) == 0 {
		return nil, domain.ErrCommentInvalid
	}
	return s.repo.UpdateComment(ctx, tenantID, documentID, libraryID, userID, in)
}

// DeleteDocumentComment deletes the comment identified by libraryID on
// documentID.
func (s *Service) DeleteDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int) error {
	if libraryID <= 0 {
		return domain.ErrCommentInvalid
	}
	return s.repo.DeleteComment(ctx, tenantID, documentID, libraryID)
}

// PresignAutosaveCmd is the input to Service.PresignAutosave.
type PresignAutosaveCmd struct {
	TenantID, ActorUserID, DocumentID, SessionID, BaseRevisionID, ContentHash string
}

// PresignAutosaveResult carries the presigned upload URL and reservation
// returned by Service.PresignAutosave.
type PresignAutosaveResult struct {
	UploadURL       string
	PendingUploadID string
	ExpiresAt       time.Time
}

// PresignAutosave presigns an upload URL and reserves a pending-commit slot
// for an autosave write, after confirming cmd.ActorUserID may write the
// document's working content.
func (s *Service) PresignAutosave(ctx context.Context, cmd PresignAutosaveCmd) (*PresignAutosaveResult, error) {
	doc, err := s.repo.GetDocument(ctx, cmd.TenantID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	mayWrite, err := s.mayWriteWorkingContent(ctx, cmd.TenantID, cmd.DocumentID, cmd.ActorUserID, doc)
	if err != nil {
		return nil, err
	}
	if !mayWrite {
		return nil, domain.ErrInvalidStateTransition
	}
	storageKey := documentRevisionKey(cmd.TenantID, cmd.DocumentID, cmd.ContentHash)
	url, err := s.presigner.PresignPut(ctx, cmd.TenantID, storageKey, revisionUploadTTL)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(revisionUploadTTL)
	pendingID, err := s.repo.PresignReserve(ctx, cmd.TenantID, cmd.SessionID, cmd.ActorUserID, cmd.DocumentID, cmd.BaseRevisionID, cmd.ContentHash, storageKey, expiresAt)
	if err != nil {
		return nil, err
	}
	return &PresignAutosaveResult{UploadURL: url, PendingUploadID: pendingID, ExpiresAt: expiresAt}, nil
}

// CommitAutosaveCmd is the input to Service.CommitAutosave.
type CommitAutosaveCmd struct {
	TenantID, ActorUserID, DocumentID, SessionID, PendingUploadID string
	FormDataSnapshot                                              json.RawMessage
	PageCount                                                     *int
}

// SyncArtifactMetadataCmd is the input to Service.SyncArtifactMetadata.
type SyncArtifactMetadataCmd struct {
	TenantID, ActorUserID, DocumentID, SessionID string
	PageCount                                    *int
}

// CommitAutosave confirms a previously presigned autosave upload against its
// reservation and commits it as a new revision, after re-confirming
// cmd.ActorUserID may write the document's working content.
func (s *Service) CommitAutosave(ctx context.Context, cmd CommitAutosaveCmd) (*CommitResult, error) {
	if cmd.PageCount != nil && *cmd.PageCount <= 0 {
		return nil, domain.ErrInvalidPageCount
	}

	doc, err := s.repo.GetDocument(ctx, cmd.TenantID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	mayWrite, err := s.mayWriteWorkingContent(ctx, cmd.TenantID, cmd.DocumentID, cmd.ActorUserID, doc)
	if err != nil {
		return nil, err
	}
	if !mayWrite {
		return nil, domain.ErrInvalidStateTransition
	}
	meta, err := s.repo.GetPendingForCommit(ctx, cmd.TenantID, cmd.PendingUploadID)
	if err != nil {
		return nil, err
	}

	vp, err := s.presigner.Confirm(ctx, cmd.TenantID, meta.StorageKey, meta.ExpectedContentHash)
	if err != nil {
		switch {
		case errors.Is(err, objectstore.ErrObjectMissing):
			return nil, domain.ErrUploadMissing
		case errors.Is(err, objectstore.ErrHashMismatch):
			return nil, domain.ErrContentHashMismatch
		case errors.Is(err, objectstore.ErrObjectTooLarge):
			return nil, domain.ErrUploadTooLarge
		default:
			return nil, fmt.Errorf("confirm s3 object: %w", err)
		}
	}
	serverHash := vp.ContentHash
	fileSizeBytes := vp.SizeBytes

	var pageCountSource *string
	if cmd.PageCount != nil {
		source := "eigenpal_client"
		pageCountSource = &source
	}
	res, err := s.repo.CommitUpload(
		ctx,
		cmd.TenantID,
		cmd.SessionID,
		cmd.ActorUserID,
		cmd.DocumentID,
		cmd.PendingUploadID,
		serverHash,
		cmd.FormDataSnapshot,
		fileSizeBytes,
		cmd.PageCount,
		pageCountSource,
	)
	if err != nil {
		return nil, err
	}
	if !res.AlreadyConsumed {
		s.audit.Write(ctx, cmd.TenantID, cmd.ActorUserID, "document.autosaved", cmd.DocumentID, map[string]any{"revision_id": res.RevisionID, "revision_num": res.RevisionNum})
	}
	return res, nil
}

// SyncArtifactMetadata re-syncs the current revision's page count/size
// metadata (e.g. after a client-side re-render) without creating a new
// revision.
func (s *Service) SyncArtifactMetadata(ctx context.Context, cmd SyncArtifactMetadataCmd) (*CommitResult, error) {
	if cmd.PageCount != nil && *cmd.PageCount <= 0 {
		return nil, domain.ErrInvalidPageCount
	}

	doc, err := s.repo.GetDocument(ctx, cmd.TenantID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	mayWrite, err := s.mayWriteWorkingContent(ctx, cmd.TenantID, cmd.DocumentID, cmd.ActorUserID, doc)
	if err != nil {
		return nil, err
	}
	if !mayWrite {
		return nil, domain.ErrInvalidStateTransition
	}
	if strings.TrimSpace(doc.CurrentRevisionID) == "" {
		return nil, domain.ErrNotFound
	}

	revision, err := s.repo.GetRevision(ctx, cmd.TenantID, cmd.DocumentID, doc.CurrentRevisionID)
	if err != nil {
		return nil, err
	}

	fileSizeBytes, err := s.presigner.Size(ctx, revision.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("size current revision object: %w", err)
	}

	var pageCountSource *string
	if cmd.PageCount != nil {
		source := "eigenpal_client"
		pageCountSource = &source
	}

	return s.repo.SyncCurrentRevisionArtifactMetadata(
		ctx,
		cmd.TenantID,
		cmd.SessionID,
		cmd.ActorUserID,
		cmd.DocumentID,
		fileSizeBytes,
		cmd.PageCount,
		pageCountSource,
	)
}

// AcquireSession acquires (or reports as already-taken) an editor session on
// docID for userID, after confirming userID may write the document's working
// content.
func (s *Service) AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, bool, error) {
	doc, err := s.repo.GetDocument(ctx, tenantID, docID)
	if err != nil {
		return nil, false, err
	}
	mayWrite, err := s.mayWriteWorkingContent(ctx, tenantID, docID, userID, doc)
	if err != nil {
		return nil, false, err
	}
	if !mayWrite {
		return nil, false, domain.ErrInvalidStateTransition
	}
	sess, err := s.repo.AcquireSession(ctx, tenantID, docID, userID)
	if errors.Is(err, domain.ErrSessionTaken) {
		return sess, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	s.audit.Write(ctx, tenantID, userID, "session.acquired", docID, map[string]any{"session_id": sess.ID})
	return sess, false, nil
}

// HeartbeatSession refreshes sessionID's liveness for userID, resolving the
// tenant from ctx.
func (s *Service) HeartbeatSession(ctx context.Context, sessionID, userID string) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.HeartbeatSession(ctx, tenantID, sessionID, userID)
}

// ReleaseSession releases sessionID and records a session.released audit
// entry.
func (s *Service) ReleaseSession(ctx context.Context, tenantID, sessionID, userID, docID string) error {
	if err := s.repo.ReleaseSession(ctx, tenantID, sessionID, userID); err != nil {
		return err
	}
	s.audit.Write(ctx, tenantID, userID, "session.released", docID, map[string]any{"session_id": sessionID})
	return nil
}

// ForceReleaseSession forcibly releases sessionID as adminID and records a
// session.force_released audit entry in the same transaction.
func (s *Service) ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID, docID string) error {
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.repo.ForceReleaseSessionTx(ctx, tx, tenantID, adminID, sessionID); err != nil {
			return err
		}
		if err := s.audit.WriteTx(ctx, tx, tenantID, adminID, "session.force_released", docID, map[string]any{"session_id": sessionID}); err != nil {
			return fmt.Errorf("documents: audit force-release session: %w", err)
		}
		return nil
	})
}

// CreateCheckpoint creates a labeled checkpoint of docID's current revision
// and records a document.checkpoint_created audit entry.
func (s *Service) CreateCheckpoint(ctx context.Context, tenantID, docID, actorID, label string) (*domain.Checkpoint, error) {
	cp, err := s.repo.CreateCheckpoint(ctx, tenantID, docID, actorID, label)
	if err != nil {
		return nil, err
	}
	s.audit.Write(ctx, tenantID, actorID, "document.checkpoint_created", docID, map[string]any{"version_num": cp.VersionNum, "label": label})
	return cp, nil
}

// ListCheckpoints lists docID's checkpoints, after asserting actorID holds
// document.view (tenant-grade).
func (s *Service) ListCheckpoints(ctx context.Context, tenantID, actorID, docID string) ([]domain.Checkpoint, error) {
	// Gate: document.view is tenant-grade (ADR 0022 Phase 8); mirrors RequireDocumentView.
	// RW tx: F8 bypass audit may INSERT.
	ctx = authz.WithCapCache(ctx)
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		return authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant")
	}); err != nil {
		return nil, err
	}
	return s.repo.ListCheckpoints(ctx, tenantID, docID)
}

// ListRevisionHistory lists docID's revision history.
func (s *Service) ListRevisionHistory(ctx context.Context, tenantID, docID string) ([]domain.RevisionHistoryItem, error) {
	return s.repo.ListRevisionHistory(ctx, tenantID, docID)
}

// RestoreCheckpoint restores docID to the checkpoint at versionNum as a new
// revision, and records a document.checkpoint_restored audit entry.
func (s *Service) RestoreCheckpoint(ctx context.Context, tenantID, docID, actorID string, versionNum int) (*RestoreResult, error) {
	res, err := s.repo.RestoreCheckpoint(ctx, tenantID, docID, actorID, versionNum)
	if err != nil {
		return nil, err
	}
	s.audit.Write(ctx, tenantID, actorID, "document.checkpoint_restored", docID, map[string]any{
		"version_num":        versionNum,
		"new_revision_id":    res.NewRevisionID,
		"new_revision_num":   res.NewRevisionNum,
		"source_revision_id": res.CheckpointRevID,
		"idempotent":         res.Idempotent,
	})
	return res, nil
}

// Archive marks docID archived and records a document.archived audit entry
// in the same transaction.
func (s *Service) Archive(ctx context.Context, tenantID, docID, actorID string) error {
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.repo.MarkArchivedTx(ctx, tx, tenantID, docID, actorID); err != nil {
			return err
		}
		if err := s.audit.WriteTx(ctx, tx, tenantID, actorID, "document.archived", docID, nil); err != nil {
			return fmt.Errorf("documents: audit archive: %w", err)
		}
		return nil
	})
}

// SignedRevisionURL presigns a short-lived GET URL for revID's stored object.
func (s *Service) SignedRevisionURL(ctx context.Context, tenantID, docID, revID string) (string, error) {
	rev, err := s.repo.GetRevision(ctx, tenantID, docID, revID)
	if err != nil {
		return "", err
	}
	return s.presigner.PresignGet(ctx, rev.StorageKey, objectDownloadTTL)
}
