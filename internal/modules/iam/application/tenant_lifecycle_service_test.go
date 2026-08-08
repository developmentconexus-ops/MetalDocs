package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/tenantdata"
)

// fakeTenantDataPort is a minimal tenantdata.Port test double for exercising
// buildExportArtifact/orderedPorts without a live DB (M7 F7.3 Task E's named
// unit-test deliverable: "artifact header/census construction... fakes").
type fakeTenantDataPort struct {
	module string
	tables []string
}

func (p fakeTenantDataPort) Module() string   { return p.module }
func (p fakeTenantDataPort) Tables() []string { return p.tables }
func (p fakeTenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	return nil, nil
}
func (p fakeTenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	return nil, nil
}

var _ tenantdata.Port = fakeTenantDataPort{}

func TestBuildExportArtifact_HeaderAndCensus(t *testing.T) {
	ports := []tenantdata.Port{
		fakeTenantDataPort{module: "iam", tables: []string{"metaldocs.iam_users"}},
		fakeTenantDataPort{module: "documents", tables: []string{"public.documents"}},
	}
	exports := map[string][]tenantdata.TableExport{
		"iam": {
			{Table: "metaldocs.iam_users", Rows: []json.RawMessage{[]byte(`{"id":"u1"}`)}},
		},
		"documents": {
			{Table: "public.documents", Rows: nil}, // included, zero rows — must still appear in census
		},
	}
	blobs := []objectstore.ObjectInfo{
		{Key: "tenants/t1/exports/foo.pdf", SizeBytes: 42, LastModified: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
	}

	artifact := buildExportArtifact(ports, exports, "t1", "job1", blobs)

	if artifact.Header.TenantID != "t1" || artifact.Header.JobID != "job1" {
		t.Fatalf("header = %+v, want tenant_id=t1 job_id=job1", artifact.Header)
	}
	if len(artifact.Header.Census) != 2 {
		t.Fatalf("census length = %d, want 2 (one entry per exported table, including zero-row tables)", len(artifact.Header.Census))
	}
	wantCensus := map[string]string{
		"metaldocs.iam_users": "included",
		"public.documents":    "included",
	}
	for _, c := range artifact.Header.Census {
		if want, ok := wantCensus[c.Table]; !ok || c.Status != want {
			t.Errorf("census entry %+v not expected or wrong status", c)
		}
	}

	if len(artifact.Tables) != 2 {
		t.Fatalf("tables length = %d, want 2", len(artifact.Tables))
	}
	var sawDocuments bool
	for _, tbl := range artifact.Tables {
		if tbl.Table == "public.documents" {
			sawDocuments = true
			if len(tbl.Rows) != 0 {
				t.Errorf("public.documents rows = %v, want empty (zero rows is a valid included table)", tbl.Rows)
			}
		}
	}
	if !sawDocuments {
		t.Error("expected public.documents in Tables even though it has zero rows")
	}

	if len(artifact.BlobManifest) != 1 || artifact.BlobManifest[0].Key != "tenants/t1/exports/foo.pdf" {
		t.Fatalf("blob manifest = %+v, want one entry for the given blob", artifact.BlobManifest)
	}
	if artifact.BlobManifest[0].SizeBytes != 42 {
		t.Errorf("blob manifest size = %d, want 42", artifact.BlobManifest[0].SizeBytes)
	}
}

func TestBuildExportArtifact_NoBlobsYieldsEmptyManifestNotNilPanic(t *testing.T) {
	artifact := buildExportArtifact(nil, map[string][]tenantdata.TableExport{}, "t1", "job1", nil)
	if len(artifact.BlobManifest) != 0 {
		t.Errorf("blob manifest = %v, want empty", artifact.BlobManifest)
	}
	if len(artifact.Header.Census) != 0 {
		t.Errorf("census = %v, want empty when no exports given", artifact.Header.Census)
	}
}

func TestOrderedPorts_FollowsEraseOrder(t *testing.T) {
	// Deliberately registered out of order and including one unlisted module
	// (should sort to the end, safety-net behavior).
	ports := []tenantdata.Port{
		fakeTenantDataPort{module: "iam"},
		fakeTenantDataPort{module: "unknown-future-module"},
		fakeTenantDataPort{module: "documents-approval"},
		fakeTenantDataPort{module: "audit"},
		fakeTenantDataPort{module: "documents"},
	}

	got := orderedPorts(ports)

	var gotOrder []string
	for _, p := range got {
		gotOrder = append(gotOrder, p.Module())
	}

	// documents-approval and documents must precede audit and iam;
	// iam must be strictly last among the known modules; the unknown
	// module must land after everything else (append-to-end safety net).
	indexOf := func(name string) int {
		for i, m := range gotOrder {
			if m == name {
				return i
			}
		}
		return -1
	}

	if indexOf("documents-approval") >= indexOf("documents") {
		t.Errorf("documents-approval must come before documents; got order %v", gotOrder)
	}
	if indexOf("documents") >= indexOf("audit") {
		t.Errorf("documents must come before audit; got order %v", gotOrder)
	}
	if indexOf("audit") >= indexOf("iam") {
		t.Errorf("audit must come before iam (iam is last); got order %v", gotOrder)
	}
	if indexOf("unknown-future-module") != len(gotOrder)-1 {
		t.Errorf("unregistered module must sort last as a safety net; got order %v", gotOrder)
	}
}

// fakeTenantLookup, fakeJobInserter, fakeJobStore, fakeEnqueuer, fakeRunner
// are minimal test doubles for exercising TenantLifecycleService.RequestExport/
// RequestErase's own-input validation and not-fully-wired guard without a
// live DB (M7 F7.3 Task E's "enqueue error mapping... fakes" deliverable).
// The tx-body behavior (authz.Require, tenant lookup, insert, audit) is
// covered by the integration suite (tests/integration), which has a real
// *sql.Tx to exercise authz.SeedTxIdentity/Require against — these unit
// tests target only the pure-Go guard clauses in request() that run before
// any tx is opened.

func TestRequestExport_BlankTenantIDIsValidationError(t *testing.T) {
	svc := NewTenantLifecycleService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	_, err := svc.RequestExport(context.Background(), "", "actor1", "tenantA")
	if !errors.Is(err, iamdomain.ErrTenantOnboardValidation) {
		t.Fatalf("RequestExport(blank tenantID) = %v, want iamdomain.ErrTenantOnboardValidation", err)
	}
}

func TestRequestErase_BlankActorIsValidationError(t *testing.T) {
	svc := NewTenantLifecycleService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	_, err := svc.RequestErase(context.Background(), "tenant1", "", "tenantA")
	if !errors.Is(err, iamdomain.ErrTenantOnboardValidation) {
		t.Fatalf("RequestErase(blank actor) = %v, want iamdomain.ErrTenantOnboardValidation", err)
	}
}

func TestRequestExport_NotFullyWiredReturnsError(t *testing.T) {
	// runner/tenants/jobs all nil (service constructed with no persistence
	// wiring, e.g. a boot path without SQLDB) must fail cleanly rather than
	// nil-pointer-panic when a request reaches it.
	svc := NewTenantLifecycleService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	_, err := svc.RequestExport(context.Background(), "tenant1", "actor1", "tenantA")
	if err == nil {
		t.Fatal("RequestExport() with no wiring = nil error, want a wiring error")
	}
	if errors.Is(err, iamdomain.ErrTenantOnboardValidation) {
		t.Fatalf("RequestExport() with no wiring should not be a validation error, got %v", err)
	}
}

func TestRunJob_BlankJobIDIsError(t *testing.T) {
	svc := NewTenantLifecycleService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err := svc.RunJob(context.Background(), ""); err == nil {
		t.Fatal("RunJob(\"\") = nil, want error")
	}
}

func TestRunJob_NoStoreWiredIsError(t *testing.T) {
	svc := NewTenantLifecycleService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err := svc.RunJob(context.Background(), "job1"); err == nil {
		t.Fatal("RunJob() with no store wired = nil, want error")
	}
}

// fakeLifecycleJobStore lets TestRunJob_AlreadyReadyIsNoOp exercise RunJob's
// idempotency short-circuit without a live DB.
type fakeLifecycleJobStore struct {
	job          iamdomain.TenantLifecycleJob
	loadErr      error
	markRunCount int
}

func (f *fakeLifecycleJobStore) LoadLifecycleJob(ctx context.Context, jobID string) (iamdomain.TenantLifecycleJob, error) {
	return f.job, f.loadErr
}
func (f *fakeLifecycleJobStore) MarkLifecycleJobRunning(ctx context.Context, jobID string) error {
	f.markRunCount++
	return nil
}
func (f *fakeLifecycleJobStore) MarkLifecycleJobReady(ctx context.Context, jobID, objectKey string) error {
	return nil
}
func (f *fakeLifecycleJobStore) MarkLifecycleJobFailed(ctx context.Context, jobID, errMsg string) error {
	return nil
}

func TestRunJob_AlreadyReadyIsNoOp(t *testing.T) {
	store := &fakeLifecycleJobStore{job: iamdomain.TenantLifecycleJob{ID: "job1", Status: "ready"}}
	svc := NewTenantLifecycleService(nil, nil, store, nil, nil, nil, nil, nil, nil, nil)
	if err := svc.RunJob(context.Background(), "job1"); err != nil {
		t.Fatalf("RunJob() on an already-ready job = %v, want nil (idempotent no-op)", err)
	}
	if store.markRunCount != 0 {
		t.Errorf("RunJob() on an already-ready job called MarkLifecycleJobRunning %d times, want 0", store.markRunCount)
	}
}
