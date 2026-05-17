package application

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	registrydomain "metaldocs/internal/modules/registry/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

func TestCreate_AutoCode(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	seq := &fakeSequenceAllocator{next: 1}
	svc := NewRegistryService(nil, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, nil)
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
	if cd.Visibility.Scope != registrydomain.VisibilityScopeRestricted {
		t.Fatalf("visibility scope = %q, want restricted", cd.Visibility.Scope)
	}
	if len(cd.Visibility.AreaCodes) != 1 || cd.Visibility.AreaCodes[0] != "quality" {
		t.Fatalf("area grants = %+v, want [quality]", cd.Visibility.AreaCodes)
	}
}

func TestCreate_ManualCode(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	svc := NewRegistryService(nil, repo, &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, nil)

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
	if cd.Visibility.Scope != registrydomain.VisibilityScopeCompany {
		t.Fatalf("visibility scope = %q, want company", cd.Visibility.Scope)
	}
}

func TestCreate_ManualCode_MissingReason(t *testing.T) {
	svc := NewRegistryService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)
	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           "Legacy Document",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		ManualCode:      stringPtr("PO-LEG-47"),
	})
	if !errors.Is(err, registrydomain.ErrManualCodeReasonRequired) {
		t.Fatalf("expected ErrManualCodeReasonRequired, got %v", err)
	}
}

func TestCreate_ManualCode_ShortReason(t *testing.T) {
	svc := NewRegistryService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)
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
	if !errors.Is(err, registrydomain.ErrManualCodeReasonRequired) {
		t.Fatalf("expected ErrManualCodeReasonRequired, got %v", err)
	}
}

func TestCreate_DuplicateCode(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	repo.codeExists = true
	svc := NewRegistryService(nil, repo, &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

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
	if !errors.Is(err, registrydomain.ErrCDCodeTaken) {
		t.Fatalf("expected ErrCDCodeTaken, got %v", err)
	}
}

func TestCreate_OverrideTemplate_GovernanceEvent(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	logger := &fakeGovernanceLogger{}
	checker := &fakeTemplateVersionChecker{byID: map[string]templateVersionState{
		"tpl-ovr-1": {status: stringPtr("published"), profileCode: "po"},
	}}
	svc := NewRegistryService(nil, repo, &fakeSequenceAllocator{next: 1}, checker, &fakeProfileReader{}, &fakeAreaReader{}, logger, nil)

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
	svc := NewRegistryService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, checker, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	_, err := svc.Create(context.Background(), CreateControlledDocumentCmd{
		TenantID:                  "tenant-a",
		ProfileCode:               "po",
		ProcessAreaCode:           "quality",
		Title:                     "Welding Procedure",
		OwnerUserID:               "owner-1",
		ActorUserID:               "actor-1",
		OverrideTemplateVersionID: stringPtr("tpl-ovr-1"),
	})
	if !errors.Is(err, registrydomain.ErrOverrideReasonRequired) {
		t.Fatalf("expected ErrOverrideReasonRequired, got %v", err)
	}
}

func TestCreate_ProfileArchived(t *testing.T) {
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	profiles := &fakeProfileReader{item: &taxonomydomain.DocumentProfile{Code: "po", TenantID: "tenant-a", ArchivedAt: &archivedAt}}
	svc := NewRegistryService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, profiles, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

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
	svc := NewRegistryService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, areas, &fakeGovernanceLogger{}, nil)

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
	created        *registrydomain.ControlledDocument
	lastListFilter registrydomain.CDFilter
}

func newFakeControlledDocumentRepository() *fakeControlledDocumentRepository {
	return &fakeControlledDocumentRepository{}
}

func (f *fakeControlledDocumentRepository) GetByID(_ context.Context, _, _ string) (*registrydomain.ControlledDocument, error) {
	if f.created == nil {
		return nil, registrydomain.ErrCDNotFound
	}
	copy := *f.created
	return &copy, nil
}

func (f *fakeControlledDocumentRepository) GetByCode(_ context.Context, _, _, _ string) (*registrydomain.ControlledDocument, error) {
	return nil, registrydomain.ErrCDNotFound
}

func (f *fakeControlledDocumentRepository) CodeExists(_ context.Context, _, _, _ string) (bool, error) {
	return f.codeExists, nil
}

func (f *fakeControlledDocumentRepository) List(_ context.Context, _ string, filter registrydomain.CDFilter) ([]registrydomain.ControlledDocument, error) {
	f.lastListFilter = filter
	return nil, nil
}
func (f *fakeControlledDocumentRepository) CanRead(_ context.Context, _, _, _ string) (bool, error) {
	return true, nil
}

func (f *fakeControlledDocumentRepository) Create(_ context.Context, doc *registrydomain.ControlledDocument) error {
	copy := *doc
	f.created = &copy
	return nil
}

func (f *fakeControlledDocumentRepository) CreateTx(_ context.Context, _ *sql.Tx, doc *registrydomain.ControlledDocument) error {
	copy := *doc
	f.created = &copy
	return nil
}

func (f *fakeControlledDocumentRepository) UpdateStatus(_ context.Context, _, _ string, _ registrydomain.CDStatus, _ time.Time) error {
	return nil
}

func (f *fakeControlledDocumentRepository) UpdateStatusTx(_ context.Context, _ *sql.Tx, _, _ string, _ registrydomain.CDStatus, _ time.Time) error {
	return nil
}

type fakeSequenceAllocator struct {
	next int
}

func (f *fakeSequenceAllocator) NextAndIncrement(_ context.Context, _ registrydomain.DBExecutor, _, _, _ string) (int, error) {
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

func (f *fakeTemplateVersionChecker) GetTemplateVersionState(_ context.Context, templateVersionID string) (*string, string, error) {
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

type fakeDocumentInitializer struct {
	called bool
	ref    *registrydomain.DocumentRef
	err    error
	gotReq registrydomain.CloneTemplateRequest
	gotCD  *registrydomain.ControlledDocument
}

func (f *fakeDocumentInitializer) CloneTemplate(_ context.Context, _ *sql.Tx, cd *registrydomain.ControlledDocument, req registrydomain.CloneTemplateRequest) (*registrydomain.DocumentRef, error) {
	f.called = true
	f.gotReq = req
	f.gotCD = cd
	if f.err != nil {
		return nil, f.err
	}
	return f.ref, nil
}

func TestRegistryService_Create_AtomicWithDocument(t *testing.T) {
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
	docInit := &fakeDocumentInitializer{ref: &registrydomain.DocumentRef{ID: "doc-xyz", ContentHash: "hash-1"}}

	svc := NewRegistryService(db, repo, seq, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, logger, docInit)

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

func TestRegistryService_Create_InitializerError_RollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	expectRegistryCreateAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectRollback()

	repo := newFakeControlledDocumentRepository()
	docInit := &fakeDocumentInitializer{err: errors.New("clone failed")}
	svc := NewRegistryService(db, repo, &fakeSequenceAllocator{next: 1}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, docInit)

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

func TestRegistryService_PreviewCode_ReturnsFormatted(t *testing.T) {
	svc := NewRegistryService(nil, newFakeControlledDocumentRepository(), &fakeSequenceAllocator{next: 7}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)
	code, err := svc.PreviewCode(context.Background(), "tenant-a", "DC", "RH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "DC-RH-007" {
		t.Fatalf("expected DC-RH-007, got %q", code)
	}
}

func TestList_DoesNotInjectEmptyActorFilter(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	svc := NewRegistryService(nil, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	_, err := svc.List(context.Background(), "tenant-a", registrydomain.CDFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.lastListFilter.ActorUserID != nil {
		t.Fatalf("ActorUserID = %v, want nil", *repo.lastListFilter.ActorUserID)
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
