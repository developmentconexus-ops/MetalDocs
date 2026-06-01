package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/tenant"
)

// Type aliases so handlers depend only on application types.
type PendingCommitMeta = repository.PendingCommitMeta
type CommitResult = repository.CommitResult
type RestoreResult = repository.RestoreResult
type RevisionHistoryItem = domain.RevisionHistoryItem

type Repository interface {
	CreateDocument(ctx context.Context, d *domain.Document, initialContentHash string, requiredPlaceholders []templatesdomain.Placeholder) (docID, revID, sessionID string, err error)
	CreateDocumentTx(ctx context.Context, tx *sql.Tx, d *domain.Document, initialContentHash, initialStorageKey string, requiredPlaceholders []templatesdomain.Placeholder) (docID, revID, sessionID string, err error)
	SetRevisionStorageKey(ctx context.Context, revID, storageKey string) error
	GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error)
	UpdateDocumentName(ctx context.Context, tenantID, actorID, docID, name string) error
	UpdateDocumentNameTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, docID, name string) error
	ListDocuments(ctx context.Context, tenantID string) ([]domain.Document, error)
	ListDocumentsForUser(ctx context.Context, tenantID, userID string) ([]domain.Document, error)
	ListDocumentsPaginated(ctx context.Context, tenantID string, opts ListOptions) ([]*domain.Document, error)
	CountDocuments(ctx context.Context, tenantID string, opts ListOptions) (int64, error)
	StatsByStatus(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error)
	StatsByArea(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error)
	UpdateDocumentStatus(ctx context.Context, tenantID, actorID, id string, cur, next domain.DocumentStatus, stampTime bool) error
	MarkArchived(ctx context.Context, tenantID, docID, actorID string) error
	IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error)
	AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, error)
	HeartbeatSession(ctx context.Context, tenantID, sessionID, userID string) error
	ReleaseSession(ctx context.Context, tenantID, sessionID, userID string) error
	ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID string) error
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

type Presigner interface {
	PresignRevisionPUT(ctx context.Context, tenantID, docID, contentHash string) (url, storageKey string, err error)
	PresignObjectGET(ctx context.Context, storageKey string) (url string, err error)
	AdoptTempObject(ctx context.Context, tmpKey, finalKey string) error
	DeleteObject(ctx context.Context, key string) error
	HashObject(ctx context.Context, key string) (string, error)
	SizeObject(ctx context.Context, key string) (int64, error)
	Exists(ctx context.Context, storageKey string) (bool, error)
}

type TemplateReader interface {
	GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error)
}

type FormValidator interface {
	Validate(schemaJSON string, formData json.RawMessage) (valid bool, errs []string, err error)
}

type Audit interface {
	Write(ctx context.Context, tenantID, actorID, action, docID string, meta any)
	WriteTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, action, docID string, meta any) error
}

// ControlledDocumentReader loads a ControlledDocument for validation at create time.
type ControlledDocumentReader interface {
	GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
}

type ControlledDocumentDuplicator interface {
	DuplicateControlledDocument(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (*controlleddocumentsdomain.ControlledDocument, error)
}

type ProfileDefaultTemplateReader interface {
	GetDefaultTemplateVersionID(ctx context.Context, tenantID, profileCode string) (*string, *string, error)
	// returns (*templateVersionID, *templateVersionStatus, error)
}

type Service struct {
	repo                         Repository
	presigner                    Presigner
	tpl                          TemplateReader
	fv                           FormValidator
	audit                        Audit
	controlledDocumentReader     ControlledDocumentReader
	controlledDocumentDuplicator ControlledDocumentDuplicator
	caps                         CapabilityChecker
	profileTemplates             ProfileDefaultTemplateReader
	snapshotSvc                  *SnapshotService
	db                           *sql.DB
}

func (s *Service) WithDB(db *sql.DB) *Service {
	s.db = db
	return s
}

func New(r Repository, p Presigner, t TemplateReader, fv FormValidator, a Audit) *Service {
	return &Service{
		repo:      r,
		presigner: p,
		tpl:       t,
		fv:        fv,
		audit:     a,
	}
}

func NewService(
	r Repository,
	p Presigner,
	t TemplateReader,
	fv FormValidator,
	a Audit,
	controlledDocumentReader ControlledDocumentReader,
	caps CapabilityChecker,
	profileTemplates ProfileDefaultTemplateReader,
) *Service {
	return &Service{
		repo:                     r,
		presigner:                p,
		tpl:                      t,
		fv:                       fv,
		audit:                    a,
		controlledDocumentReader: controlledDocumentReader,
		caps:                     caps,
		profileTemplates:         profileTemplates,
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
	controlledDocumentReader ControlledDocumentReader,
	caps CapabilityChecker,
	profileTemplates ProfileDefaultTemplateReader,
	snap *SnapshotService,
) *Service {
	return &Service{
		repo:                     r,
		presigner:                p,
		tpl:                      t,
		fv:                       fv,
		audit:                    a,
		controlledDocumentReader: controlledDocumentReader,
		caps:                     caps,
		profileTemplates:         profileTemplates,
		snapshotSvc:              snap,
	}
}

func (s *Service) WithControlledDocumentDuplicator(d ControlledDocumentDuplicator) *Service {
	s.controlledDocumentDuplicator = d
	return s
}

var ErrControlledDocumentRequired = errors.New("controlled_document_id is required")
var errControlledDocumentReaderNotConfigured = errors.New("controlled document reader not configured")
var errControlledDocumentDuplicatorNotConfigured = errors.New("controlled document duplicator not configured")
var errCapabilityCheckerNotConfigured = errors.New("capability checker not configured")
var errProfileTemplateReaderNotConfigured = errors.New("profile default template reader not configured")

type CreateDocumentInput struct {
	TenantID             string
	ActorUserID          string
	ControlledDocumentID string
	TemplateVersionID    string
	Name                 string
	FormData             json.RawMessage
}

type CreateDocumentCmd = CreateDocumentInput

type CreateDocumentResult struct {
	DocumentID        string
	InitialRevisionID string
	SessionID         string
}

func buildDocumentForCreate(cmd CreateDocumentInput, cd *controlleddocumentsdomain.ControlledDocument, resolvedTemplateVersionID string) domain.Document {
	return domain.Document{
		TenantID:                cmd.TenantID,
		TemplateVersionID:       resolvedTemplateVersionID,
		Name:                    cmd.Name,
		FormDataJSON:            cmd.FormData,
		CreatedBy:               cmd.ActorUserID,
		ControlledDocumentID:    &cmd.ControlledDocumentID,
		ProfileCodeSnapshot:     &cd.ProfileCode,
		ProcessAreaCodeSnapshot: &cd.ProcessAreaCode,
		Code:                    cd.Code,
	}
}

// CreateDocument is the legacy (non-atomic) create path. It renders via docx-renderer,
// uploads to S3, and audits asynchronously. Only called by DuplicateDocument.
// The atomic flow (cloneIntoTx) sets storage_key in the same tx — no S3 side-effect.
// TODO: when duplicate-document migrates to atomic flow, delete this + repo.CreateDocument + SetRevisionStorageKey.
func (s *Service) CreateDocument(ctx context.Context, cmd CreateDocumentInput) (res *CreateDocumentResult, err error) {
	if strings.TrimSpace(cmd.ControlledDocumentID) == "" {
		return nil, ErrControlledDocumentRequired
	}
	if s.controlledDocumentReader == nil {
		return nil, errControlledDocumentReaderNotConfigured
	}
	if s.caps == nil {
		return nil, errCapabilityCheckerNotConfigured
	}
	if s.profileTemplates == nil {
		return nil, errProfileTemplateReaderNotConfigured
	}

	cd, err := s.controlledDocumentReader.GetByID(ctx, cmd.TenantID, cmd.ControlledDocumentID)
	if err != nil {
		return nil, err
	}
	if !cd.IsActive() {
		return nil, controlleddocumentsdomain.ErrCDNotActive
	}

	if err := s.caps.CanDo(ctx, cmd.ActorUserID, cmd.TenantID, iamdomain.CapDocumentCreate); err != nil {
		return nil, err
	}

	defaultTemplateID, defaultTemplateStatus, err := s.profileTemplates.GetDefaultTemplateVersionID(ctx, cmd.TenantID, cd.ProfileCode)
	if err != nil {
		return nil, err
	}

	var overrideTemplate *controlleddocumentsdomain.TemplateVersionCandidate
	if cd.OverrideTemplateVersionID != nil {
		overrideStatus := "published"
		overrideTemplate = &controlleddocumentsdomain.TemplateVersionCandidate{
			ID:          *cd.OverrideTemplateVersionID,
			ProfileCode: cd.ProfileCode,
			Status:      &overrideStatus,
		}
	}

	var defaultTemplate *controlleddocumentsdomain.TemplateVersionCandidate
	if defaultTemplateID != nil {
		defaultTemplate = &controlleddocumentsdomain.TemplateVersionCandidate{
			ID:          *defaultTemplateID,
			ProfileCode: cd.ProfileCode,
			Status:      defaultTemplateStatus,
		}
	}

	resolution, err := controlleddocumentsdomain.Resolve(controlleddocumentsdomain.TemplateResolutionInput{
		ProfileCode:      cd.ProfileCode,
		OverrideTemplate: overrideTemplate,
		DefaultTemplate:  defaultTemplate,
	})
	if err != nil {
		return nil, err
	}
	resolvedTemplateVersionID := resolution.TemplateVersionID

	docxKey, _, schemaJSON, err := s.tpl.GetPublishedVersion(ctx, cmd.TenantID, resolvedTemplateVersionID)
	if err != nil {
		return nil, fmt.Errorf("template lookup: %w", err)
	}
	if schemaJSON != "" {
		ok, verrs, err := s.fv.Validate(schemaJSON, cmd.FormData)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("form_data_invalid: %v", verrs)
		}
	}

	// Resolve template snapshot pre-INSERT so snapshot columns are written
	// atomically with the documents row (C2/C4: no half-born rows).
	var snap domain.TemplateSnapshot
	var phs []templatesdomain.Placeholder
	if s.snapshotSvc != nil {
		var resolveErr error
		snap, phs, resolveErr = s.snapshotSvc.ResolveTemplate(ctx, cmd.TenantID, resolvedTemplateVersionID)
		if resolveErr != nil {
			return nil, fmt.Errorf("resolve template snapshot: %w", resolveErr)
		}
	}

	// Document content starts as the template's published docx (template
	// passthrough): storage_key points directly at the template docx so the
	// editor opens immediately on first GET. The browser editor is the
	// fill/edit path; there is no server-side render step.
	h := sha256.New()
	h.Write([]byte(docxKey))
	contentHash := fmt.Sprintf("%x", h.Sum(nil))

	doc := buildDocumentForCreate(cmd, cd, resolvedTemplateVersionID)
	doc.TemplateSnapshot = snap
	docID, revID, sessionID, err := s.repo.CreateDocument(ctx, &doc, contentHash, phs)
	if err != nil {
		return nil, err
	}

	finalKey := docxKey // point to template docx directly
	if err := s.repo.SetRevisionStorageKey(ctx, revID, finalKey); err != nil {
		return nil, fmt.Errorf("set revision key: %w", err)
	}

	s.audit.Write(ctx, cmd.TenantID, cmd.ActorUserID, "document.created", docID, map[string]any{"template_version_id": resolvedTemplateVersionID})
	return &CreateDocumentResult{DocumentID: docID, InitialRevisionID: revID, SessionID: sessionID}, nil
}

// cloneIntoTxInput is the internal payload for cloneIntoTx. It mirrors the
// fields buildDocumentForCreate consumes plus the override-template hint the
// controlled-document atomic-create path may carry.
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
}

// cloneIntoTx performs the DB-only portion of CreateDocument inside a
// caller-owned tx so callers (e.g. controlled-document atomic-create) can compose the CD
// insert and the document insert in a single transaction.
//
// Differences from CreateDocument:
//   - No S3 rendering (docx-renderer is not invoked).
//   - Template-passthrough: storage_key is set to the template's published
//     docx key atomically with the insert, so the editor opens immediately
//     on first GET — no lazy materialization, no AdoptTempObject side-effect.
//   - No outbox / audit writes — caller decides which side-effects to run
//     after tx.Commit().
//
// All repo calls thread the tx via CreateDocumentTx.
func (s *Service) cloneIntoTx(ctx context.Context, tx *sql.Tx, in cloneIntoTxInput) (docID string, contentHash string, err error) {
	resolvedTemplateVersionID, err := s.resolveTemplateVersionID(ctx, in.TenantID, in.ProfileCode, in.OverrideTemplateVersionID)
	if err != nil {
		return "", "", err
	}

	docxKey, _, _, err := s.tpl.GetPublishedVersion(ctx, in.TenantID, resolvedTemplateVersionID)
	if err != nil {
		return "", "", fmt.Errorf("template lookup: %w", err)
	}

	// Resolve template snapshot for the same atomicity guarantees as
	// CreateDocument (C2/C4) — still pure DB reads.
	var snap domain.TemplateSnapshot
	var phs []templatesdomain.Placeholder
	if s.snapshotSvc != nil {
		snap, phs, err = s.snapshotSvc.ResolveTemplate(ctx, in.TenantID, resolvedTemplateVersionID)
		if err != nil {
			return "", "", fmt.Errorf("resolve template snapshot: %w", err)
		}
	}

	// content_hash is required by document_revisions but storage_key is empty
	// at this point. Mirror the docx-renderer-not-configured fallback in
	// CreateDocument: sha256(docxKey). Kept stable across replays.
	h := sha256.New()
	h.Write([]byte(docxKey))
	contentHash = fmt.Sprintf("%x", h.Sum(nil))

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

	docID, _, _, err = s.repo.CreateDocumentTx(ctx, tx, &doc, contentHash, docxKey, phs)
	if err != nil {
		return "", "", err
	}
	return docID, contentHash, nil
}

func (s *Service) resolveTemplateVersionID(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error) {
	if s.profileTemplates == nil {
		return "", errProfileTemplateReaderNotConfigured
	}

	defaultTemplateID, defaultTemplateStatus, err := s.profileTemplates.GetDefaultTemplateVersionID(ctx, tenantID, profileCode)
	if err != nil {
		return "", err
	}

	var overrideTemplate *controlleddocumentsdomain.TemplateVersionCandidate
	if templateVersionID != nil && strings.TrimSpace(*templateVersionID) != "" {
		overrideStatus := "published"
		overrideTemplate = &controlleddocumentsdomain.TemplateVersionCandidate{
			ID:          strings.TrimSpace(*templateVersionID),
			ProfileCode: profileCode,
			Status:      &overrideStatus,
		}
	}

	var defaultTemplate *controlleddocumentsdomain.TemplateVersionCandidate
	if defaultTemplateID != nil {
		defaultTemplate = &controlleddocumentsdomain.TemplateVersionCandidate{
			ID:          *defaultTemplateID,
			ProfileCode: profileCode,
			Status:      defaultTemplateStatus,
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

func (s *Service) GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	return s.repo.GetDocument(ctx, tenantID, id)
}

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
	cd, err := s.controlledDocumentDuplicator.DuplicateControlledDocument(ctx, tenantID, *doc.ControlledDocumentID, userID)
	if err != nil {
		return nil, err
	}
	return s.CreateDocument(ctx, CreateDocumentInput{
		TenantID:             tenantID,
		ActorUserID:          userID,
		ControlledDocumentID: cd.ID,
		TemplateVersionID:    doc.TemplateVersionID,
		Name:                 doc.Name,
		FormData:             doc.FormDataJSON,
	})
}

func (s *Service) ListDocuments(ctx context.Context, tenantID string) ([]domain.Document, error) {
	return s.repo.ListDocuments(ctx, tenantID)
}

func (s *Service) ListDocumentsForUser(ctx context.Context, tenantID, userID string) ([]domain.Document, error) {
	return s.repo.ListDocumentsForUser(ctx, tenantID, userID)
}

type DocumentStats struct {
	ByStatus map[string]int64 `json:"byStatus"`
	ByArea   map[string]int64 `json:"byArea"`
}

func (s *Service) ListDocumentsPaginated(ctx context.Context, tenantID, userID string, opts ListOptions) ([]*domain.Document, int64, error) {
	if userID != "" {
		opts.CreatedBy = userID
	}

	items, err := s.repo.ListDocumentsPaginated(ctx, tenantID, opts)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountDocuments(ctx, tenantID, opts)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

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
	if s.db == nil {
		if err := s.repo.UpdateDocumentName(ctx, tenantID, userID, docID, name); err != nil {
			return err
		}
		s.audit.Write(ctx, tenantID, userID, "document.renamed", docID, map[string]any{"name": name})
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rename document: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.UpdateDocumentNameTx(ctx, tx, tenantID, userID, docID, name); err != nil {
		return err
	}
	if err := s.audit.WriteTx(ctx, tx, tenantID, userID, "document.renamed", docID, map[string]any{"name": name}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error) {
	return s.repo.IsDocumentOwner(ctx, tenantID, docID, userID)
}

func (s *Service) ListDocumentComments(ctx context.Context, tenantID, userID, documentID string) ([]domain.Comment, error) {
	return s.repo.ListComments(ctx, tenantID, documentID)
}

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

func (s *Service) UpdateDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int, in domain.CommentUpdateInput) (*domain.Comment, error) {
	if libraryID <= 0 {
		return nil, domain.ErrCommentInvalid
	}
	if in.ContentJSON != nil && len(*in.ContentJSON) == 0 {
		return nil, domain.ErrCommentInvalid
	}
	return s.repo.UpdateComment(ctx, tenantID, documentID, libraryID, userID, in)
}

func (s *Service) DeleteDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int) error {
	if libraryID <= 0 {
		return domain.ErrCommentInvalid
	}
	return s.repo.DeleteComment(ctx, tenantID, documentID, libraryID)
}

type PresignAutosaveCmd struct {
	TenantID, ActorUserID, DocumentID, SessionID, BaseRevisionID, ContentHash string
}

type PresignAutosaveResult struct {
	UploadURL       string
	PendingUploadID string
	ExpiresAt       time.Time
}

func (s *Service) PresignAutosave(ctx context.Context, cmd PresignAutosaveCmd) (*PresignAutosaveResult, error) {
	doc, err := s.repo.GetDocument(ctx, cmd.TenantID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	url, storageKey, err := s.presigner.PresignRevisionPUT(ctx, cmd.TenantID, cmd.DocumentID, cmd.ContentHash)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(15 * time.Minute)
	pendingID, err := s.repo.PresignReserve(ctx, cmd.TenantID, cmd.SessionID, cmd.ActorUserID, cmd.DocumentID, cmd.BaseRevisionID, cmd.ContentHash, storageKey, expiresAt)
	if err != nil {
		return nil, err
	}
	return &PresignAutosaveResult{UploadURL: url, PendingUploadID: pendingID, ExpiresAt: expiresAt}, nil
}

type CommitAutosaveCmd struct {
	TenantID, ActorUserID, DocumentID, SessionID, PendingUploadID string
	FormDataSnapshot                                              json.RawMessage
	PageCount                                                     *int
}

type SyncArtifactMetadataCmd struct {
	TenantID, ActorUserID, DocumentID, SessionID string
	PageCount                                    *int
}

func (s *Service) CommitAutosave(ctx context.Context, cmd CommitAutosaveCmd) (*CommitResult, error) {
	if cmd.PageCount != nil && *cmd.PageCount <= 0 {
		return nil, domain.ErrInvalidPageCount
	}

	doc, err := s.repo.GetDocument(ctx, cmd.TenantID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	meta, err := s.repo.GetPendingForCommit(ctx, cmd.TenantID, cmd.PendingUploadID)
	if err != nil {
		return nil, err
	}

	serverHash, err := s.presigner.HashObject(ctx, meta.StorageKey)
	if err != nil {
		if errors.Is(err, domain.ErrUploadMissing) {
			return nil, domain.ErrUploadMissing
		}
		return nil, fmt.Errorf("hash s3 object: %w", err)
	}
	if serverHash != meta.ExpectedContentHash {
		_ = s.presigner.DeleteObject(ctx, meta.StorageKey)
		return nil, domain.ErrContentHashMismatch
	}

	fileSizeBytes, err := s.presigner.SizeObject(ctx, meta.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("size s3 object: %w", err)
	}

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

func (s *Service) SyncArtifactMetadata(ctx context.Context, cmd SyncArtifactMetadataCmd) (*CommitResult, error) {
	if cmd.PageCount != nil && *cmd.PageCount <= 0 {
		return nil, domain.ErrInvalidPageCount
	}

	doc, err := s.repo.GetDocument(ctx, cmd.TenantID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	if strings.TrimSpace(doc.CurrentRevisionID) == "" {
		return nil, domain.ErrNotFound
	}

	revision, err := s.repo.GetRevision(ctx, cmd.TenantID, cmd.DocumentID, doc.CurrentRevisionID)
	if err != nil {
		return nil, err
	}

	fileSizeBytes, err := s.presigner.SizeObject(ctx, revision.StorageKey)
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

func (s *Service) AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, bool, error) {
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

func (s *Service) HeartbeatSession(ctx context.Context, sessionID, userID string) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.HeartbeatSession(ctx, tenantID, sessionID, userID)
}

func (s *Service) ReleaseSession(ctx context.Context, tenantID, sessionID, userID, docID string) error {
	if err := s.repo.ReleaseSession(ctx, tenantID, sessionID, userID); err != nil {
		return err
	}
	s.audit.Write(ctx, tenantID, userID, "session.released", docID, map[string]any{"session_id": sessionID})
	return nil
}

func (s *Service) ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID, docID string) error {
	if err := s.repo.ForceReleaseSession(ctx, tenantID, adminID, sessionID); err != nil {
		return err
	}
	s.audit.Write(ctx, tenantID, adminID, "session.force_released", docID, map[string]any{"session_id": sessionID})
	return nil
}

func (s *Service) CreateCheckpoint(ctx context.Context, tenantID, docID, actorID, label string) (*domain.Checkpoint, error) {
	cp, err := s.repo.CreateCheckpoint(ctx, tenantID, docID, actorID, label)
	if err != nil {
		return nil, err
	}
	s.audit.Write(ctx, tenantID, actorID, "document.checkpoint_created", docID, map[string]any{"version_num": cp.VersionNum, "label": label})
	return cp, nil
}

func (s *Service) ListCheckpoints(ctx context.Context, tenantID, docID string) ([]domain.Checkpoint, error) {
	return s.repo.ListCheckpoints(ctx, tenantID, docID)
}

func (s *Service) ListRevisionHistory(ctx context.Context, tenantID, docID string) ([]domain.RevisionHistoryItem, error) {
	return s.repo.ListRevisionHistory(ctx, tenantID, docID)
}

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

func (s *Service) Finalize(ctx context.Context, tenantID, docID, actorID string) error {
	if err := s.repo.UpdateDocumentStatus(ctx, tenantID, actorID, docID, domain.DocStatusDraft, domain.DocStatusUnderReview, true); err != nil {
		return err
	}
	s.audit.Write(ctx, tenantID, actorID, "document.finalized", docID, nil)
	return nil
}

func (s *Service) Archive(ctx context.Context, tenantID, docID, actorID string) error {
	if err := s.repo.MarkArchived(ctx, tenantID, docID, actorID); err != nil {
		return err
	}
	s.audit.Write(ctx, tenantID, actorID, "document.archived", docID, nil)
	return nil
}

func (s *Service) SignedRevisionURL(ctx context.Context, tenantID, docID, revID string) (string, error) {
	rev, err := s.repo.GetRevision(ctx, tenantID, docID, revID)
	if err != nil {
		return "", err
	}
	return s.presigner.PresignObjectGET(ctx, rev.StorageKey)
}
