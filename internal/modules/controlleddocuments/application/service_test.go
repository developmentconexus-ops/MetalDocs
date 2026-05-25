package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

func TestNewControlledDocumentService_PanicsOnNilRequiredDependencies(t *testing.T) {
	validRepo := newFakeControlledDocumentRepository()
	validSeq := &fakeSequenceAllocator{next: 1}
	validTpl := &fakeTemplateVersionChecker{}
	validProfiles := &fakeProfileReader{}
	validAreas := &fakeAreaReader{}
	validLogger := &fakeGovernanceLogger{}
	assertPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatalf("%s: expected panic", name)
			}
		}()
		fn()
	}

	assertPanic("nil repository", func() {
		NewControlledDocumentService(nil, nil, validSeq, validTpl, validProfiles, validAreas, validLogger, nil)
	})
	assertPanic("nil sequence allocator", func() {
		NewControlledDocumentService(nil, validRepo, nil, validTpl, validProfiles, validAreas, validLogger, nil)
	})
	assertPanic("nil template checker", func() {
		NewControlledDocumentService(nil, validRepo, validSeq, nil, validProfiles, validAreas, validLogger, nil)
	})
	assertPanic("nil profile reader", func() {
		NewControlledDocumentService(nil, validRepo, validSeq, validTpl, nil, validAreas, validLogger, nil)
	})
	assertPanic("nil area reader", func() {
		NewControlledDocumentService(nil, validRepo, validSeq, validTpl, validProfiles, nil, validLogger, nil)
	})
	assertPanic("nil governance logger", func() {
		NewControlledDocumentService(nil, validRepo, validSeq, validTpl, validProfiles, validAreas, nil, nil)
	})
}

func TestCreate_AutoCode(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	seq := &fakeSequenceAllocator{next: 1}
	svc := NewControlledDocumentService(nil, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, newInvariantReadyDocumentInitializer())
	svc.now = func() time.Time { return time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC) }

	res, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           "Welding Procedure",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		VisibilityScope: "restricted",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cd := res.ControlledDocument
	if cd.Code != "PO-QUALITY-001" {
		t.Fatalf("expected PO-QUALITY-001, got %q", cd.Code)
	}
	if cd.SequenceNum == nil || *cd.SequenceNum != 1 {
		t.Fatalf("expected sequence 1, got %+v", cd.SequenceNum)
	}
	if len(logger.events) != 0 {
		t.Fatalf("expected zero governance events, got %+v", logger.events)
	}
	if cd.Visibility.Scope != controlleddocumentsdomain.VisibilityScopeRestricted {
		t.Fatalf("visibility scope = %q, want restricted", cd.Visibility.Scope)
	}
	if len(cd.Visibility.AreaCodes) != 1 || cd.Visibility.AreaCodes[0] != "quality" {
		t.Fatalf("area grants = %+v, want [quality]", cd.Visibility.AreaCodes)
	}
}

func TestCreate_ManualCode(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	svc := NewControlledDocumentService(nil, repo, &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, newInvariantReadyDocumentInitializer())

	res, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:         "tenant-a",
		ProfileCode:      "po",
		ProcessAreaCode:  "quality",
		Title:            "Legacy Document",
		OwnerUserID:      "owner-1",
		ActorUserID:      "actor-1",
		ManualCode:       stringPtr("PO-LEG-47"),
		ManualCodeReason: stringPtr("Legacy migration from spreadsheet"),
		VisibilityScope:  "company",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cd := res.ControlledDocument
	if cd.Code != "PO-LEG-47" {
		t.Fatalf("unexpected code: %q", cd.Code)
	}
	if cd.SequenceNum != nil {
		t.Fatalf("expected nil sequence for manual code, got %+v", cd.SequenceNum)
	}
	if len(logger.events) != 1 || logger.events[0].EventType != "numbering.override" {
		t.Fatalf("expected numbering.override event, got %+v", logger.events)
	}
	if cd.Visibility.Scope != controlleddocumentsdomain.VisibilityScopeCompany {
		t.Fatalf("visibility scope = %q, want company", cd.Visibility.Scope)
	}
}

func TestCreate_ManualCode_MissingReason(t *testing.T) {
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())
	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           "Legacy Document",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		ManualCode:      stringPtr("PO-LEG-47"),
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrManualCodeReasonRequired) {
		t.Fatalf("expected ErrManualCodeReasonRequired, got %v", err)
	}
}

func TestCreate_ManualCode_ShortReason(t *testing.T) {
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())
	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:         "tenant-a",
		ProfileCode:      "po",
		ProcessAreaCode:  "quality",
		Title:            "Legacy Document",
		OwnerUserID:      "owner-1",
		ActorUserID:      "actor-1",
		ManualCode:       stringPtr("PO-LEG-47"),
		ManualCodeReason: stringPtr("too short"),
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrManualCodeReasonRequired) {
		t.Fatalf("expected ErrManualCodeReasonRequired, got %v", err)
	}
}

func TestCreate_UsesControlledDocumentConstructorValidation(t *testing.T) {
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())
	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           " ",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		VisibilityScope: "company",
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrCDTitleRequired) {
		t.Fatalf("expected ErrCDTitleRequired, got %v", err)
	}
}

func TestCreate_DuplicateCode(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	repo.codeExists = true
	svc := NewControlledDocumentService(nil, repo, &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:         "tenant-a",
		ProfileCode:      "po",
		ProcessAreaCode:  "quality",
		Title:            "Welding Procedure",
		OwnerUserID:      "owner-1",
		ActorUserID:      "actor-1",
		ManualCode:       stringPtr("PO-01"),
		ManualCodeReason: stringPtr("Manual override due to migration"),
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrCDCodeTaken) {
		t.Fatalf("expected ErrCDCodeTaken, got %v", err)
	}
}

func TestCreate_OverrideTemplate_GovernanceEvent(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	checker := &fakeTemplateVersionChecker{byID: map[string]templateVersionState{
		"tpl-ovr-1": {status: stringPtr("published"), profileCode: "po"},
	}}
	svc := NewControlledDocumentService(nil, repo, &fakeSequenceAllocator{next: 1}, checker, &fakeProfileReader{}, &fakeAreaReader{}, logger, newInvariantReadyDocumentInitializer())

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:                  "tenant-a",
		ProfileCode:               "po",
		ProcessAreaCode:           "quality",
		Title:                     "Welding Procedure",
		OwnerUserID:               "owner-1",
		ActorUserID:               "actor-1",
		OverrideTemplateVersionID: stringPtr("tpl-ovr-1"),
		OverrideTemplateReason:    stringPtr("Emergency temporary override for legal form"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logger.events) != 1 || logger.events[0].EventType != "template.override" {
		t.Fatalf("expected template.override event, got %+v", logger.events)
	}
}

func TestCreate_OverrideTemplate_MissingReason(t *testing.T) {
	checker := &fakeTemplateVersionChecker{byID: map[string]templateVersionState{
		"tpl-ovr-1": {status: stringPtr("published"), profileCode: "po"},
	}}
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, checker, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:                  "tenant-a",
		ProfileCode:               "po",
		ProcessAreaCode:           "quality",
		Title:                     "Welding Procedure",
		OwnerUserID:               "owner-1",
		ActorUserID:               "actor-1",
		OverrideTemplateVersionID: stringPtr("tpl-ovr-1"),
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrOverrideReasonRequired) {
		t.Fatalf("expected ErrOverrideReasonRequired, got %v", err)
	}
}

func TestCreate_ProfileArchived(t *testing.T) {
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	profiles := &fakeProfileReader{item: &taxonomydomain.DocumentProfile{Code: "po", TenantID: "tenant-a", ArchivedAt: &archivedAt}}
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, profiles, &fakeAreaReader{}, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           "Welding Procedure",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
	})
	if !errors.Is(err, taxonomydomain.ErrProfileArchived) {
		t.Fatalf("expected ErrProfileArchived, got %v", err)
	}
}

func TestCreate_AreaArchived(t *testing.T) {
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	areas := &fakeAreaReader{item: &taxonomydomain.ProcessArea{Code: "quality", TenantID: "tenant-a", ArchivedAt: &archivedAt}}
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, areas, &fakeGovernanceLogger{}, newInvariantReadyDocumentInitializer())

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           "Welding Procedure",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
	})
	if !errors.Is(err, taxonomydomain.ErrAreaArchived) {
		t.Fatalf("expected ErrAreaArchived, got %v", err)
	}
}

type fakeControlledDocumentRepository struct {
	codeExists     bool
	created        *controlleddocumentsdomain.ControlledDocument
	lastListFilter controlleddocumentsdomain.CDFilter
}

func newFakeControlledDocumentRepository() *fakeControlledDocumentRepository {
	return &fakeControlledDocumentRepository{}
}

func (f *fakeControlledDocumentRepository) GetByID(_ context.Context, _, _ string) (*controlleddocumentsdomain.ControlledDocument, error) {
	if f.created == nil {
		return nil, controlleddocumentsdomain.ErrCDNotFound
	}
	copy := *f.created
	return &copy, nil
}

func (f *fakeControlledDocumentRepository) GetByCode(_ context.Context, _, _, _ string) (*controlleddocumentsdomain.ControlledDocument, error) {
	return nil, controlleddocumentsdomain.ErrCDNotFound
}

func (f *fakeControlledDocumentRepository) CodeExists(_ context.Context, _, _, _ string) (bool, error) {
	return f.codeExists, nil
}

func (f *fakeControlledDocumentRepository) List(_ context.Context, _ string, filter controlleddocumentsdomain.CDFilter) ([]controlleddocumentsdomain.ControlledDocument, error) {
	f.lastListFilter = filter
	return nil, nil
}
func (f *fakeControlledDocumentRepository) CanRead(_ context.Context, _, _, _ string) (bool, error) {
	return true, nil
}

func (f *fakeControlledDocumentRepository) Create(_ context.Context, doc *controlleddocumentsdomain.ControlledDocument) error {
	copy := *doc
	f.created = &copy
	return nil
}

func (f *fakeControlledDocumentRepository) CreateTx(_ context.Context, _ *sql.Tx, doc *controlleddocumentsdomain.ControlledDocument) error {
	copy := *doc
	f.created = &copy
	return nil
}

func (f *fakeControlledDocumentRepository) UpdateStatus(_ context.Context, _, _ string, _ controlleddocumentsdomain.CDStatus, _ time.Time) error {
	return nil
}

func (f *fakeControlledDocumentRepository) UpdateStatusTx(_ context.Context, _ *sql.Tx, _, _ string, _ controlleddocumentsdomain.CDStatus, _ time.Time) error {
	return nil
}

type fakeSequenceAllocator struct {
	next int
}

func (f *fakeSequenceAllocator) NextAndIncrement(_ context.Context, _ controlleddocumentsdomain.DBExecutor, _, _, _ string) (int, error) {
	v := f.next
	f.next++
	return v, nil
}

func (f *fakeSequenceAllocator) Peek(_ context.Context, _, _, _ string) (int, error) {
	return f.next, nil
}

func (f *fakeSequenceAllocator) EnsureCounter(_ context.Context, _, _, _ string) error {
	return nil
}

type templateVersionState struct {
	status      *string
	profileCode string
}

type fakeTemplateVersionChecker struct {
	byID map[string]templateVersionState
}

func (f *fakeTemplateVersionChecker) GetTemplateVersionState(_ context.Context, _ string, templateVersionID string) (*string, string, error) {
	if f.byID == nil {
		return nil, "", nil
	}
	item, ok := f.byID[templateVersionID]
	if !ok {
		return nil, "", nil
	}
	return item.status, item.profileCode, nil
}

type fakeProfileReader struct {
	item *taxonomydomain.DocumentProfile
}

func (f *fakeProfileReader) GetByCode(_ context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	if f.item == nil {
		return &taxonomydomain.DocumentProfile{Code: code, TenantID: tenantID}, nil
	}
	copy := *f.item
	return &copy, nil
}

type fakeAreaReader struct {
	item *taxonomydomain.ProcessArea
}

func (f *fakeAreaReader) GetByCode(_ context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error) {
	if f.item == nil {
		return &taxonomydomain.ProcessArea{Code: code, TenantID: tenantID}, nil
	}
	copy := *f.item
	return &copy, nil
}

type fakeGovernanceLogger struct {
	events []taxonomydomain.GovernanceEvent
}

func (f *fakeGovernanceLogger) Log(_ context.Context, e taxonomydomain.GovernanceEvent) error {
	f.events = append(f.events, e)
	return nil
}

func stringPtr(v string) *string { return &v }

func newInvariantReadyDocumentInitializer() *fakeDocumentInitializer {
	return &fakeDocumentInitializer{exists: true}
}

type fakeDocumentInitializer struct {
	called bool
	ref    *controlleddocumentsdomain.DocumentRef
	err    error
	gotReq controlleddocumentsdomain.CloneTemplateRequest
	gotCD  *controlleddocumentsdomain.ControlledDocument

	storageKey                        string
	exists                            bool
	existsErr                         error
	gotResolveTemplateStorageKeyCalls int
	gotExistsCalls                    int
	gotExistsStorageKey               string
}

func (f *fakeDocumentInitializer) CloneTemplate(_ context.Context, _ *sql.Tx, cd *controlleddocumentsdomain.ControlledDocument, req controlleddocumentsdomain.CloneTemplateRequest) (*controlleddocumentsdomain.DocumentRef, error) {
	f.called = true
	f.gotReq = req
	f.gotCD = cd
	if f.err != nil {
		return nil, f.err
	}
	return f.ref, nil
}

func (f *fakeDocumentInitializer) ResolveTemplateStorageKey(_ context.Context, _, _ string, _ *string) (string, error) {
	f.gotResolveTemplateStorageKeyCalls++
	if f.storageKey == "" {
		return "system/templates/blank.docx", nil
	}
	return f.storageKey, nil
}

func (f *fakeDocumentInitializer) Exists(_ context.Context, storageKey string) (bool, error) {
	f.gotExistsCalls++
	f.gotExistsStorageKey = storageKey
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.exists, nil
}

func TestControlledDocumentService_Create_AtomicWithDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	expectRegistryCreateAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectCommit()

	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	seq := &fakeSequenceAllocator{next: 3}
	docInit := &fakeDocumentInitializer{
		ref:    &controlleddocumentsdomain.DocumentRef{ID: "doc-xyz", ContentHash: "hash-1"},
		exists: true,
	}

	svc := NewControlledDocumentService(db, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, docInit)

	res, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "dc",
		ProcessAreaCode: "rh",
		Title:           "HR Policy",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		DocumentName:    "HR Policy v1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ControlledDocument.Code != "DC-RH-003" {
		t.Fatalf("expected DC-RH-003, got %q", res.ControlledDocument.Code)
	}
	if res.DocumentRef == nil || res.DocumentRef.ID != "doc-xyz" {
		t.Fatalf("expected DocumentRef.ID=doc-xyz, got %+v", res.DocumentRef)
	}
	if !docInit.called {
		t.Fatalf("expected docInit.CloneTemplate to be called")
	}
	if docInit.gotReq.Name != "HR Policy v1" {
		t.Fatalf("expected DocumentName forwarded, got %q", docInit.gotReq.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestControlledDocumentService_Create_InitializerError_RollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	expectRegistryCreateAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectRollback()

	repo := newFakeControlledDocumentRepository()
	docInit := &fakeDocumentInitializer{err: errors.New("clone failed"), exists: true}
	svc := NewControlledDocumentService(db, repo, &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

	_, err = svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "dc",
		ProcessAreaCode: "rh",
		Title:           "HR Policy",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		DocumentName:    "HR Policy v1",
	})
	if err == nil {
		t.Fatalf("expected error from initializer, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations (expected rollback): %v", err)
	}
}

func TestControlledDocumentService_Create_AtomicTemplateArtifactMissing_FailsClosedBeforePersisting(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	seq := &fakeSequenceAllocator{next: 4}
	docInit := &fakeDocumentInitializer{
		ref:        &controlleddocumentsdomain.DocumentRef{ID: "doc-xyz", ContentHash: "hash-1"},
		storageKey: "system/templates/blank.docx",
		exists:     false,
	}
	svc := NewControlledDocumentService(nil, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, docInit)

	templateVersionID := "00000000-0000-0000-0000-000000000102"
	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:          "tenant-a",
		ProfileCode:       "dc",
		ProcessAreaCode:   "rh",
		Title:             "HR Policy",
		OwnerUserID:       "owner-1",
		ActorUserID:       "actor-1",
		DocumentName:      "HR Policy v1",
		TemplateVersionID: &templateVersionID,
	})
	if !errors.Is(err, ErrTemplateArtifactMissing) {
		t.Fatalf("expected ErrTemplateArtifactMissing, got %v", err)
	}
	if docInit.gotResolveTemplateStorageKeyCalls != 1 {
		t.Fatalf("expected template artifact resolver to be called once, got %d", docInit.gotResolveTemplateStorageKeyCalls)
	}
	if docInit.gotExistsCalls != 1 {
		t.Fatalf("expected template artifact checker to be called once, got %d", docInit.gotExistsCalls)
	}
	if docInit.gotExistsStorageKey != "system/templates/blank.docx" {
		t.Fatalf("storage key = %q, want system/templates/blank.docx", docInit.gotExistsStorageKey)
	}
	if repo.created != nil {
		t.Fatalf("expected controlled document not to persist when template artifact is missing")
	}
	if seq.next != 4 {
		t.Fatalf("expected sequence allocation to remain unused, got next=%d", seq.next)
	}
	if docInit.called {
		t.Fatalf("expected document initializer not to run when template artifact is missing")
	}
}

func TestControlledDocumentService_Create_TemplateArtifactInvariantUnconfigured_FailsClosed(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	seq := &fakeSequenceAllocator{next: 5}
	svc := NewControlledDocumentService(nil, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	templateVersionID := "00000000-0000-0000-0000-000000000102"
	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:          "tenant-a",
		ProfileCode:       "dc",
		ProcessAreaCode:   "rh",
		Title:             "HR Policy",
		OwnerUserID:       "owner-1",
		ActorUserID:       "actor-1",
		DocumentName:      "HR Policy v1",
		TemplateVersionID: &templateVersionID,
	})
	if err == nil {
		t.Fatalf("expected invariant configuration error, got nil")
	}
	if !errors.Is(err, ErrTemplateArtifactInvariantUnconfigured) {
		t.Fatalf("expected ErrTemplateArtifactInvariantUnconfigured, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("expected controlled document not to persist when invariant is unconfigured")
	}
	if seq.next != 5 {
		t.Fatalf("expected sequence allocation to remain unused, got next=%d", seq.next)
	}
}

func TestControlledDocumentService_Create_ManualCodeReasonWinsBeforeArtifactCheck(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	seq := &fakeSequenceAllocator{next: 6}
	docInit := &fakeDocumentInitializer{storageKey: "system/templates/blank.docx", exists: false}
	svc := NewControlledDocumentService(nil, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "dc",
		ProcessAreaCode: "rh",
		Title:           "HR Policy",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		ManualCode:      stringPtr("HR-001"),
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrManualCodeReasonRequired) {
		t.Fatalf("expected ErrManualCodeReasonRequired, got %v", err)
	}
	if docInit.gotResolveTemplateStorageKeyCalls != 0 {
		t.Fatalf("expected no template artifact resolution for invalid manual-code request, got %d", docInit.gotResolveTemplateStorageKeyCalls)
	}
	if docInit.gotExistsCalls != 0 {
		t.Fatalf("expected no template artifact existence checks for invalid manual-code request, got %d", docInit.gotExistsCalls)
	}
	if seq.next != 6 {
		t.Fatalf("expected sequence allocation to remain unused, got next=%d", seq.next)
	}
}

func TestControlledDocumentService_Create_AutoTemplateArtifactMissing_AfterAuthzBeforeSequence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	expectRegistryCreateAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectRollback()

	repo := newFakeControlledDocumentRepository()
	seq := &fakeSequenceAllocator{next: 9}
	docInit := &fakeDocumentInitializer{storageKey: "system/templates/blank.docx", exists: false}
	svc := NewControlledDocumentService(db, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

	_, err = svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "dc",
		ProcessAreaCode: "rh",
		Title:           "HR Policy",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		DocumentName:    "HR Policy v1",
	})
	if !errors.Is(err, ErrTemplateArtifactMissing) {
		t.Fatalf("expected ErrTemplateArtifactMissing, got %v", err)
	}
	if docInit.gotResolveTemplateStorageKeyCalls != 1 {
		t.Fatalf("expected template artifact resolution after authz, got %d calls", docInit.gotResolveTemplateStorageKeyCalls)
	}
	if seq.next != 9 {
		t.Fatalf("expected sequence allocation to remain unused, got next=%d", seq.next)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestControlledDocumentService_Create_OverrideTemplateValidationWinsBeforeArtifactCheck(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	seq := &fakeSequenceAllocator{next: 10}
	docInit := &fakeDocumentInitializer{storageKey: "system/templates/blank.docx", exists: false}
	checker := &fakeTemplateVersionChecker{byID: map[string]templateVersionState{
		"tpl-ovr-1": {status: stringPtr("published"), profileCode: "qa"},
	}}
	svc := NewControlledDocumentService(nil, repo, seq, checker, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:                  "tenant-a",
		ProfileCode:               "dc",
		ProcessAreaCode:           "rh",
		Title:                     "HR Policy",
		OwnerUserID:               "owner-1",
		ActorUserID:               "actor-1",
		DocumentName:              "HR Policy v1",
		OverrideTemplateVersionID: stringPtr("tpl-ovr-1"),
		OverrideTemplateReason:    stringPtr("Emergency temporary override for legal form"),
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrTemplateProfileMismatch) {
		t.Fatalf("expected ErrTemplateProfileMismatch, got %v", err)
	}
	if docInit.gotResolveTemplateStorageKeyCalls != 0 {
		t.Fatalf("expected no template artifact resolution for invalid override request, got %d", docInit.gotResolveTemplateStorageKeyCalls)
	}
	if docInit.gotExistsCalls != 0 {
		t.Fatalf("expected no template artifact existence checks for invalid override request, got %d", docInit.gotExistsCalls)
	}
	if seq.next != 10 {
		t.Fatalf("expected sequence allocation to remain unused, got next=%d", seq.next)
	}
}

func TestControlledDocumentService_PreviewCode_ReturnsFormatted(t *testing.T) {
	svc := NewControlledDocumentService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 7}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)
	code, err := svc.PreviewCode(context.Background(), "tenant-a", "DC", "RH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "DC-RH-007" {
		t.Fatalf("expected DC-RH-007, got %q", code)
	}
}

func TestPeekSeq_ProfileTenantValidationError(t *testing.T) {
	svc := NewControlledDocumentService(
		nil,
		newFakeControlledDocumentRepository(),
		&fakeSequenceAllocator{next: 7},
		&fakeTemplateVersionChecker{},
		&fakeProfileReader{item: nil},
		&fakeAreaReader{},
		&fakeGovernanceLogger{},
		nil,
	)
	profileErr := taxonomydomain.ErrProfileNotFound
	svc.profiles = profileReaderFunc(func(context.Context, string, string) (*taxonomydomain.DocumentProfile, error) {
		return nil, profileErr
	})

	_, err := svc.PeekSeq(context.Background(), "tenant-a", "DC", "RH")
	if !errors.Is(err, profileErr) {
		t.Fatalf("expected profile error, got %v", err)
	}
}

func TestList_MissingActorContextReturnsErrActorMissing(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	svc := NewControlledDocumentService(nil, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	_, err := svc.List(context.Background(), "tenant-a", controlleddocumentsdomain.CDFilter{})
	if !errors.Is(err, ErrActorMissing) {
		t.Fatalf("List: expected ErrActorMissing, got %v", err)
	}
	if repo.lastListFilter.ActorUserID != nil {
		t.Fatalf("ActorUserID must not be set when actor missing, got %v", *repo.lastListFilter.ActorUserID)
	}
}

func TestList_AppliesActorFilterWhenContextPresent(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	svc := NewControlledDocumentService(nil, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	ctx := iamdomain.WithAuthContext(context.Background(), "actor-1", []iamdomain.Role{"system_admin"})
	_, err := svc.List(ctx, "tenant-a", controlleddocumentsdomain.CDFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.lastListFilter.ActorUserID == nil || *repo.lastListFilter.ActorUserID != "actor-1" {
		t.Fatalf("ActorUserID = %v, want pointer to \"actor-1\"", repo.lastListFilter.ActorUserID)
	}
}

func TestGet_MissingActorContextReturnsErrActorMissing(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	svc := NewControlledDocumentService(nil, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	_, err := svc.Get(context.Background(), "tenant-a", "cd-1")
	if !errors.Is(err, ErrActorMissing) {
		t.Fatalf("Get: expected ErrActorMissing, got %v", err)
	}
}

func TestCreateRevision_MissingActorContextReturnsErrActorMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := newFakeControlledDocumentRepository()
	repo.created = &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd-1",
		TenantID:        "tenant-a",
		ProfileCode:     "pop",
		ProcessAreaCode: "general",
		Code:            "POP-GENERAL-014",
		Status:          controlleddocumentsdomain.CDStatusActive,
		OwnerUserID:     "owner-1",
	}
	svc := NewControlledDocumentService(db, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, &fakeDocumentInitializer{})

	_, err = svc.CreateRevision(context.Background(), CreateRevisionCmd{
		TenantID:          "tenant-a",
		CDID:              "cd-1",
		Name:              "E2E Approval Flow",
		TemplateVersionID: stringPtr("template-1"),
	})
	if !errors.Is(err, ErrActorMissing) {
		t.Fatalf("CreateRevision: expected ErrActorMissing, got %v", err)
	}
}

func TestCreateRevision_SetsAuthzContextBeforeClone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	expectRegistryCreateAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectCommit()

	repo := newFakeControlledDocumentRepository()
	repo.created = &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd-1",
		TenantID:        "tenant-a",
		ProfileCode:     "pop",
		ProcessAreaCode: "general",
		Code:            "POP-GENERAL-014",
		Status:          controlleddocumentsdomain.CDStatusActive,
		OwnerUserID:     "owner-1",
	}
	docInit := &fakeDocumentInitializer{
		ref: &controlleddocumentsdomain.DocumentRef{ID: "doc-rev-2", ContentHash: "hash-2"},
	}
	svc := NewControlledDocumentService(db, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

	ctx := iamdomain.WithAuthContext(context.Background(), "actor-1", []iamdomain.Role{"system_admin"})
	ref, err := svc.CreateRevision(ctx, CreateRevisionCmd{
		TenantID:          "tenant-a",
		CDID:              "cd-1",
		Name:              "E2E Approval Flow 2026-05-19 REV01",
		TemplateVersionID: stringPtr("template-1"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref == nil || ref.ID != "doc-rev-2" {
		t.Fatalf("expected doc-rev-2, got %+v", ref)
	}
	if !docInit.called {
		t.Fatalf("expected CloneTemplate to be called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestCreateRevision_MapsActiveSiblingUniqueViolationToDomainConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	expectRegistryCreateAuthz(mock, "actor-1", "tenant-a")

	repo := newFakeControlledDocumentRepository()
	repo.created = &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd-1",
		TenantID:        "tenant-a",
		ProfileCode:     "pop",
		ProcessAreaCode: "general",
		Code:            "POP-GENERAL-014",
		Status:          controlleddocumentsdomain.CDStatusActive,
		OwnerUserID:     "owner-1",
	}
	docInit := &fakeDocumentInitializer{
		err: fmt.Errorf("insert document: %w", &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "ux_documents_cd_active",
		}),
	}
	svc := NewControlledDocumentService(db, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

	ctx := iamdomain.WithAuthContext(context.Background(), "actor-1", []iamdomain.Role{"system_admin"})
	_, err = svc.CreateRevision(ctx, CreateRevisionCmd{
		TenantID: "tenant-a",
		CDID:     "cd-1",
		Name:     "E2E Approval Flow 2026-05-19 REV01",
	})
	if !errors.Is(err, controlleddocumentsdomain.ErrActiveRevisionExists) {
		t.Fatalf("expected ErrActiveRevisionExists, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func expectRegistryCreateAuthz(mock sqlmock.Sqlmock, actorID, tenantID string) {
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.tenant_id', $1, true)")).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.actor_id', $1, true)")).
		WithArgs(actorID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(actorID))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(tenantID))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles
   WHERE user_id   = $1
     AND tenant_id = $2::uuid
     AND role_code = 'system_admin'
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(nil))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

type profileReaderFunc func(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)

func (f profileReaderFunc) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	return f(ctx, tenantID, code)
}
