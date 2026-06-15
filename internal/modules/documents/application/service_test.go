package application_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

// newPermissiveMockDB returns a *sql.DB that satisfies Begin/Commit/Exec calls
// without any SQL assertions. Used by tests that exercise code paths touching
// s.db but whose repo/audit fakes absorb all real work.
func newPermissiveMockDB(t *testing.T) *sql.DB {
	t.Helper()
	anyMatcher := sqlmock.QueryMatcherFunc(func(_, _ string) error { return nil })
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(anyMatcher))
	if err != nil {
		t.Fatalf("newPermissiveMockDB: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	t.Cleanup(func() { _ = mockDB.Close() })
	for i := 0; i < 10; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	for i := 0; i < 50; i++ {
		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	return mockDB
}

type fakeRepo struct {
	createDocErr    error
	createDocIDs    [3]string
	updateStatusErr error
	acquireSess      *domain.Session
	acquireErr       error
	pendingMeta      *application.PendingCommitMeta
	pendingErr       error
	commitResult     *application.CommitResult
	commitErr        error
	restoreResult    *application.RestoreResult
	restoreErr       error
	ownerReturn      bool
	ownerErr         error
	checkpointResult *domain.Checkpoint
	checkpointErr    error

	docReturn   *domain.Document
	listReturn  []domain.Document
	checkpoints []domain.Checkpoint

	statusCalls int
	statusCur   domain.DocumentStatus
	statusNext  domain.DocumentStatus
	statusStamp bool

	revisionReturn *domain.Revision
	renameErr      error
	renameName     string
	renameDocID    string
	renameTenantID string

	lastOpts            application.ListOptions
	listPaginatedReturn []*domain.Document
	listPaginatedErr    error
	countReturn         int64
	countErr            error
	statsByStatusReturn map[string]int64
	statsByStatusErr    error
	statsByAreaReturn   map[string]int64
	statsByAreaErr      error
	commitTenantID      string
	commitFileSizeBytes int64
	commitPageCount     *int
	commitPageSource    *string
	syncTenantID        string
	syncFileSizeBytes   int64
	syncPageCount       *int
	syncPageSource      *string
	revisionHistory     []domain.RevisionHistoryItem
	revisionHistoryErr  error
}

var _ application.Repository = (*fakeRepo)(nil)

func (f *fakeRepo) CreateDocumentTx(_ context.Context, _ db.Tx, _ *domain.Document, _, _ string, _ []templatesdomain.Placeholder) (string, string, string, error) {
	if f.createDocErr != nil {
		return "", "", "", f.createDocErr
	}
	return f.createDocIDs[0], f.createDocIDs[1], f.createDocIDs[2], nil
}

func (f *fakeRepo) GetDocument(_ context.Context, _, _ string) (*domain.Document, error) {
	if f.docReturn == nil {
		return nil, errors.New("document not configured")
	}
	return f.docReturn, nil
}

func (f *fakeRepo) UpdateDocumentName(_ context.Context, tenantID, actorID, docID, name string) error {
	f.renameTenantID = tenantID
	f.renameDocID = docID
	f.renameName = name
	return f.renameErr
}

func (f *fakeRepo) UpdateDocumentNameTx(_ context.Context, _ db.Tx, tenantID, actorID, docID, name string) error {
	f.renameTenantID = tenantID
	f.renameDocID = docID
	f.renameName = name
	return f.renameErr
}

func (f *fakeRepo) ListDocuments(_ context.Context, _ string) ([]domain.Document, error) {
	return f.listReturn, nil
}

func (f *fakeRepo) ListDocumentsForUser(_ context.Context, _, _ string) ([]domain.Document, error) {
	return f.listReturn, nil
}

func (f *fakeRepo) ListDocumentsPaginated(_ context.Context, _ string, opts application.ListOptions) ([]*domain.Document, int64, bool, error) {
	f.lastOpts = opts
	if f.listPaginatedErr != nil {
		return nil, 0, false, f.listPaginatedErr
	}
	// Single-snapshot query yields the grand total alongside the page rows.
	return f.listPaginatedReturn, f.countReturn, false, nil
}

func (f *fakeRepo) CountDocuments(_ context.Context, _ string, opts application.ListOptions) (int64, error) {
	f.lastOpts = opts
	if f.countErr != nil {
		return 0, f.countErr
	}
	return f.countReturn, nil
}

func (f *fakeRepo) StatsByStatus(_ context.Context, _ string, opts application.ListOptions) (map[string]int64, error) {
	f.lastOpts = opts
	if f.statsByStatusErr != nil {
		return nil, f.statsByStatusErr
	}
	return f.statsByStatusReturn, nil
}

func (f *fakeRepo) StatsByArea(_ context.Context, _ string, opts application.ListOptions) (map[string]int64, error) {
	f.lastOpts = opts
	if f.statsByAreaErr != nil {
		return nil, f.statsByAreaErr
	}
	return f.statsByAreaReturn, nil
}

func (f *fakeRepo) UpdateDocumentStatus(_ context.Context, _, _, _ string, cur, next domain.DocumentStatus, stampTime bool) error {
	f.statusCalls++
	f.statusCur = cur
	f.statusNext = next
	f.statusStamp = stampTime
	return f.updateStatusErr
}

func (f *fakeRepo) AcquireSession(_ context.Context, _, _, _ string) (*domain.Session, error) {
	return f.acquireSess, f.acquireErr
}

func (f *fakeRepo) HeartbeatSession(_ context.Context, _, _, _ string) error { return nil }

func (f *fakeRepo) ReleaseSession(_ context.Context, _, _, _ string) error { return nil }

func (f *fakeRepo) ForceReleaseSession(_ context.Context, _, _, _ string) error { return nil }

func (f *fakeRepo) ForceReleaseSessionTx(_ context.Context, _ db.Tx, _, _, _ string) error {
	return nil
}

func (f *fakeRepo) ExpireStaleSessions(_ context.Context, _ time.Time) (int, error) { return 0, nil }

func (f *fakeRepo) PresignReserve(_ context.Context, _, _, _, _, _, _, _ string, _ time.Time) (string, error) {
	return "pending_1", nil
}

func (f *fakeRepo) GetPendingForCommit(_ context.Context, _, _ string) (*application.PendingCommitMeta, error) {
	if f.pendingErr != nil {
		return nil, f.pendingErr
	}
	return f.pendingMeta, nil
}

func (f *fakeRepo) CommitUpload(_ context.Context, tenantID, _, _, _, _, _ string, _ []byte, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*application.CommitResult, error) {
	f.commitTenantID = tenantID
	f.commitFileSizeBytes = fileSizeBytes
	f.commitPageCount = pageCount
	f.commitPageSource = pageCountSource
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	if f.commitResult == nil {
		f.commitResult = &application.CommitResult{RevisionID: "rev_1", RevisionNum: 1}
	}
	return f.commitResult, nil
}

func (f *fakeRepo) SyncCurrentRevisionArtifactMetadata(_ context.Context, tenantID, _, _, _ string, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*application.CommitResult, error) {
	f.syncTenantID = tenantID
	f.syncFileSizeBytes = fileSizeBytes
	f.syncPageCount = pageCount
	f.syncPageSource = pageCountSource
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	if f.commitResult == nil {
		f.commitResult = &application.CommitResult{RevisionID: "rev_1"}
	}
	return f.commitResult, nil
}

func (f *fakeRepo) CreateCheckpoint(_ context.Context, _, _, _, _ string) (*domain.Checkpoint, error) {
	if f.checkpointErr != nil {
		return nil, f.checkpointErr
	}
	return f.checkpointResult, nil
}

func (f *fakeRepo) ListCheckpoints(_ context.Context, _, _ string) ([]domain.Checkpoint, error) {
	return f.checkpoints, nil
}

func (f *fakeRepo) RestoreCheckpoint(_ context.Context, _, _, _ string, _ int) (*application.RestoreResult, error) {
	if f.restoreErr != nil {
		return nil, f.restoreErr
	}
	return f.restoreResult, nil
}

func (f *fakeRepo) IsDocumentOwner(_ context.Context, _, _, _ string) (bool, error) {
	if f.ownerErr != nil {
		return false, f.ownerErr
	}
	return f.ownerReturn, nil
}

func (f *fakeRepo) GetRevision(_ context.Context, _, _, _ string) (*domain.Revision, error) {
	if f.revisionReturn == nil {
		return nil, errors.New("revision not configured")
	}
	return f.revisionReturn, nil
}

func (f *fakeRepo) ListRevisionHistory(_ context.Context, _, _ string) ([]domain.RevisionHistoryItem, error) {
	if f.revisionHistoryErr != nil {
		return nil, f.revisionHistoryErr
	}
	return f.revisionHistory, nil
}

func (f *fakeRepo) DeleteExpiredPending(_ context.Context, _ time.Time) (int, error) { return 0, nil }

func (f *fakeRepo) CreateComment(_ context.Context, _, _, _ string, _ domain.CommentCreateInput) (*domain.Comment, error) {
	return &domain.Comment{}, nil
}

func (f *fakeRepo) ListComments(_ context.Context, _, _ string) ([]domain.Comment, error) {
	return nil, nil
}

func (f *fakeRepo) UpdateComment(_ context.Context, _, _ string, _ int, _ string, _ domain.CommentUpdateInput) (*domain.Comment, error) {
	return &domain.Comment{}, nil
}

func (f *fakeRepo) DeleteComment(_ context.Context, _, _ string, _ int) error {
	return nil
}

func (f *fakeRepo) GetFinalizePrereqs(_ context.Context, _, _ string) (*domain.FinalizePrereqs, error) {
	return nil, domain.ErrDocumentNotDraft
}

func (f *fakeRepo) MarkArchived(_ context.Context, _, _, _ string) error { return nil }

func (f *fakeRepo) MarkArchivedTx(_ context.Context, _ db.Tx, _, _, _ string) error { return nil }

type fakePresigner struct {
	hashReturn  string
	hashErr     error
	sizeReturn  int64
	sizeErr     error
	adoptErr    error
	deleteCalls int
	deleteErr   error
	exists      bool
	existsErr   error
}

func (f *fakePresigner) PresignRevisionPUT(_ context.Context, _, _, _ string) (string, string, error) {
	return "https://example/upload", "documents/doc_1/revisions/rev_1.docx", nil
}

func (f *fakePresigner) HashObject(_ context.Context, _ string) (string, error) {
	if f.hashErr != nil {
		return "", f.hashErr
	}
	return f.hashReturn, nil
}

func (f *fakePresigner) SizeObject(_ context.Context, _ string) (int64, error) {
	if f.sizeErr != nil {
		return 0, f.sizeErr
	}
	return f.sizeReturn, nil
}

func (f *fakePresigner) AdoptTempObject(_ context.Context, _, _ string) error {
	return f.adoptErr
}

func (f *fakePresigner) DeleteObject(_ context.Context, _ string) error {
	f.deleteCalls++
	return f.deleteErr
}

func (f *fakePresigner) PresignObjectGET(_ context.Context, storageKey string) (string, error) {
	return "https://example/get/" + storageKey, nil
}

func (f *fakePresigner) Exists(_ context.Context, _ string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.exists, nil
}

type fakeTplReader struct {
	err error
}

func (f fakeTplReader) GetPublishedVersion(_ context.Context, _, _ string) (string, string, string, error) {
	if f.err != nil {
		return "", "", "", f.err
	}
	return "tpl/docx/key.docx", "tpl/schema/key.json", `{"type":"object"}`, nil
}

type fakeFormVal struct {
	valid bool
	errs  []string
	err   error
}

func (f fakeFormVal) Validate(_ string, _ json.RawMessage) (bool, []string, error) {
	if f.err != nil {
		return false, nil, f.err
	}
	if !f.valid {
		return false, f.errs, nil
	}
	return true, nil, nil
}

type noopAudit struct {
	calls      int
	lastAction string
}

func (n *noopAudit) Write(_ context.Context, _, _, action, _ string, _ any) {
	n.calls++
	n.lastAction = action
}

func (n *noopAudit) WriteTx(_ context.Context, _ db.Tx, _, _, action, _ string, _ any) error {
	n.calls++
	n.lastAction = action
	return nil
}

func TestAcquireSession_Readonly_WhenTaken(t *testing.T) {
	repo := &fakeRepo{
		acquireSess: &domain.Session{ID: "sess_taken", DocumentID: "doc_1", UserID: "other"},
		acquireErr:  domain.ErrSessionTaken,
	}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})

	sess, readonly, err := svc.AcquireSession(context.Background(), "tenant_1", "doc_1", "user_1")
	if err != nil {
		t.Fatalf("AcquireSession() error = %v", err)
	}
	if !readonly {
		t.Fatalf("expected readonly session")
	}
	if sess == nil || sess.ID != "sess_taken" {
		t.Fatalf("unexpected session: %+v", sess)
	}
}

func TestAcquireSession_Success_RecordsAudit(t *testing.T) {
	repo := &fakeRepo{acquireSess: &domain.Session{ID: "sess_1", DocumentID: "doc_1", UserID: "user_1"}}
	audit := &noopAudit{}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, audit)

	sess, readonly, err := svc.AcquireSession(context.Background(), "tenant_1", "doc_1", "user_1")
	if err != nil {
		t.Fatalf("AcquireSession() error = %v", err)
	}
	if readonly {
		t.Fatalf("expected writable session")
	}
	if sess == nil || sess.ID != "sess_1" {
		t.Fatalf("unexpected session: %+v", sess)
	}
	if audit.calls == 0 || audit.lastAction != "session.acquired" {
		t.Fatalf("expected audit record for acquire, got calls=%d action=%q", audit.calls, audit.lastAction)
	}
}

func TestCreateCheckpoint_OK(t *testing.T) {
	repo := &fakeRepo{checkpointResult: &domain.Checkpoint{ID: "cp_1", DocumentID: "doc_1", VersionNum: 3}}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})

	cp, err := svc.CreateCheckpoint(context.Background(), "tenant_1", "doc_1", "user_1", "Milestone")
	if err != nil {
		t.Fatalf("CreateCheckpoint() error = %v", err)
	}
	if cp == nil || cp.ID != "cp_1" {
		t.Fatalf("unexpected checkpoint: %+v", cp)
	}
}

func TestRenameDocument_OK(t *testing.T) {
	repo := &fakeRepo{docReturn: &domain.Document{ID: "doc_1", TenantID: "tenant_1", Status: domain.DocStatusDraft}}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{}).
		WithRunner(db.NewTxRunner(newPermissiveMockDB(t)))

	err := svc.RenameDocument(context.Background(), "tenant_1", "user_1", "doc_1", "  New Name  ")
	if err != nil {
		t.Fatalf("RenameDocument() error = %v", err)
	}
	if repo.renameTenantID != "tenant_1" || repo.renameDocID != "doc_1" || repo.renameName != "New Name" {
		t.Fatalf("unexpected rename args: tenant=%q doc=%q name=%q", repo.renameTenantID, repo.renameDocID, repo.renameName)
	}
}

func TestRenameDocument_InvalidName(t *testing.T) {
	repo := &fakeRepo{}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})

	err := svc.RenameDocument(context.Background(), "tenant_1", "user_1", "doc_1", "   ")
	if !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestCommitAutosave_ForwardsArtifactMetadata(t *testing.T) {
	repo := &fakeRepo{
		docReturn: &domain.Document{ID: "doc_1", Status: domain.DocStatusDraft},
		pendingMeta: &application.PendingCommitMeta{
			ExpectedContentHash: "h1",
			StorageKey:          "tenants/t/documents/d/revisions/h1.docx",
		},
		commitResult: &application.CommitResult{RevisionID: "rev_2", RevisionNum: 2},
	}
	presigner := &fakePresigner{hashReturn: "h1", sizeReturn: 1304}
	svc := application.New(repo, presigner, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})
	pageCount := 3

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:         "tenant_1",
		ActorUserID:      "user_1",
		DocumentID:       "doc_1",
		SessionID:        "sess_1",
		PendingUploadID:  "pending_1",
		FormDataSnapshot: []byte(`{"ok":true}`),
		PageCount:        &pageCount,
	})
	if err != nil {
		t.Fatalf("CommitAutosave: %v", err)
	}
	if repo.commitFileSizeBytes != 1304 {
		t.Fatalf("file size = %d, want 1304", repo.commitFileSizeBytes)
	}
	if repo.commitPageCount == nil || *repo.commitPageCount != 3 {
		t.Fatalf("page count not forwarded: %#v", repo.commitPageCount)
	}
	if repo.commitPageSource == nil || *repo.commitPageSource != "eigenpal_client" {
		t.Fatalf("page count source = %#v, want eigenpal_client", repo.commitPageSource)
	}
}

func TestCommitAutosave_InvalidPageCount(t *testing.T) {
	repo := &fakeRepo{}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})
	invalid := 0

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:   "tenant_1",
		DocumentID: "doc_1",
		PageCount:  &invalid,
	})
	if !errors.Is(err, domain.ErrInvalidPageCount) {
		t.Fatalf("expected ErrInvalidPageCount, got %v", err)
	}
}

func TestCommitAutosave_LogsCleanupFailureOnHashMismatch(t *testing.T) {
	repo := &fakeRepo{
		docReturn: &domain.Document{ID: "doc_1", Status: domain.DocStatusDraft},
		pendingMeta: &application.PendingCommitMeta{
			ExpectedContentHash: "h1",
			StorageKey:          "tenants/t/documents/d/revisions/h1.docx",
		},
	}
	presigner := &fakePresigner{
		hashReturn: "WRONG",          // forces hash mismatch path
		deleteErr:  errors.New("s3 down"), // forces delete to fail → WARN fires
	}

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	_, err := application.New(repo, presigner, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{}).
		CommitAutosave(context.Background(), application.CommitAutosaveCmd{
			TenantID:         "tenant_1",
			ActorUserID:      "user_1",
			DocumentID:       "doc_1",
			SessionID:        "sess_1",
			PendingUploadID:  "pending_1",
			FormDataSnapshot: []byte(`{"ok":true}`),
		})

	// G1: behavior unchanged — still returns ErrContentHashMismatch
	if !errors.Is(err, domain.ErrContentHashMismatch) {
		t.Fatalf("expected ErrContentHashMismatch, got %v", err)
	}
	// G2: delete was still attempted
	if presigner.deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", presigner.deleteCalls)
	}
	// G3: WARN with context was emitted
	logged := buf.String()
	if logged == "" {
		t.Fatal("expected WARN log output, got empty buffer")
	}
	// Assert every attribute the impl promises to log — key names and values — so a
	// regression dropping document_id or err from the WarnContext call fails the test.
	for _, want := range []string{
		"level=WARN",
		"storage_key", "tenants/t/documents/d/revisions/h1.docx",
		"document_id", "doc_1",
		"err", "s3 down",
	} {
		if !bytes.Contains([]byte(logged), []byte(want)) {
			t.Fatalf("WARN log missing %q; got: %s", want, logged)
		}
	}
}

func TestSyncArtifactMetadata_ForwardsArtifactMetadata(t *testing.T) {
	repo := &fakeRepo{
		docReturn: &domain.Document{
			ID:                "doc_1",
			Status:            domain.DocStatusDraft,
			CurrentRevisionID: "rev_1",
		},
		revisionReturn: &domain.Revision{
			ID:         "rev_1",
			DocumentID: "doc_1",
			StorageKey: "tenants/t/documents/d/revisions/rev_1.docx",
		},
		commitResult: &application.CommitResult{RevisionID: "rev_1"},
	}
	presigner := &fakePresigner{sizeReturn: 1304}
	svc := application.New(repo, presigner, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})
	pageCount := 3

	_, err := svc.SyncArtifactMetadata(context.Background(), application.SyncArtifactMetadataCmd{
		TenantID:    "tenant_1",
		ActorUserID: "user_1",
		DocumentID:  "doc_1",
		SessionID:   "sess_1",
		PageCount:   &pageCount,
	})
	if err != nil {
		t.Fatalf("SyncArtifactMetadata: %v", err)
	}
	if repo.syncFileSizeBytes != 1304 {
		t.Fatalf("file size = %d, want 1304", repo.syncFileSizeBytes)
	}
	if repo.syncPageCount == nil || *repo.syncPageCount != 3 {
		t.Fatalf("page count not forwarded: %#v", repo.syncPageCount)
	}
	if repo.syncPageSource == nil || *repo.syncPageSource != "eigenpal_client" {
		t.Fatalf("page count source = %#v, want eigenpal_client", repo.syncPageSource)
	}
}

func TestSyncArtifactMetadata_InvalidPageCount(t *testing.T) {
	repo := &fakeRepo{}
	svc := application.New(repo, &fakePresigner{}, fakeTplReader{}, fakeFormVal{valid: true}, &noopAudit{})
	invalid := 0

	_, err := svc.SyncArtifactMetadata(context.Background(), application.SyncArtifactMetadataCmd{
		TenantID:   "tenant_1",
		DocumentID: "doc_1",
		PageCount:  &invalid,
	})
	if !errors.Is(err, domain.ErrInvalidPageCount) {
		t.Fatalf("expected ErrInvalidPageCount, got %v", err)
	}
}
