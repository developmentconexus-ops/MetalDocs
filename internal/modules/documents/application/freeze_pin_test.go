package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	tmpldom "metaldocs/internal/modules/templates/domain"
)

type fakeMaterializeOutboxEnqueuer struct {
	calls      int
	tenantIDs  []string
	revIDs     []string
	hashes     [][]byte
	err        error
}

func (f *fakeMaterializeOutboxEnqueuer) Enqueue(_ context.Context, _ *sql.Tx, tenantID, revisionID string, contentHash []byte) error {
	if f.err != nil {
		return f.err
	}
	f.calls++
	f.tenantIDs = append(f.tenantIDs, tenantID)
	f.revIDs = append(f.revIDs, revisionID)
	f.hashes = append(f.hashes, append([]byte(nil), contentHash...))
	return nil
}

func TestFreezeService_Pin_NoNetworkCall(t *testing.T) {
	resolverKey := "doc_code"
	schema := []tmpldom.Placeholder{
		{ID: "p_user", Name: "user_field", Required: true},
		{ID: "p_comp", Name: "doc_code_field", Computed: true, ResolverKey: &resolverKey},
	}
	existing := []repository.PlaceholderValue{
		{PlaceholderID: "p_user", ValueText: strPtr("user-value"), Source: "user"},
	}
	writer := &fakeFillInWriter{}
	valuesRead := &fakeValuesReader{values: existing}
	reg := resolvers.NewRegistry()
	reg.Register(fixedResolver{key: "doc_code", ver: 3, val: "DOC-001"})
	finalize := &fakeFreezeFinalizer{}
	ctxBuilder := &fakeResolverContextBuilder{input: resolvers.ResolveInput{TenantID: "t", RevisionID: "r"}}
	snapReader := fakeSnapshotReader{snap: v2dom.TemplateSnapshot{
		BodyDocxS3Key:   "templates/body.docx",
		CompositionJSON: []byte(`{}`),
	}}
	fanoutClient := &fakeFanoutClient{}
	materializeOutbox := &fakeMaterializeOutboxEnqueuer{}

	svc := NewFreezeService(fakeSchemaReader{placeholders: schema}, writer, valuesRead, reg, finalize, ctxBuilder, snapReader, &fakeFinalDocxWriter{}, fanoutClient).
		WithMaterializeOutbox(materializeOutbox)

	if err := svc.Pin(context.Background(), nil, "t", "r", ApproverContext{}); err != nil {
		t.Fatalf("Pin error: %v", err)
	}

	if fanoutClient.calls != 0 {
		t.Fatalf("Pin must not call fanout, got %d call(s)", fanoutClient.calls)
	}
	if finalize.calls != 1 {
		t.Fatalf("WriteFreeze should be called once, got %d", finalize.calls)
	}
	if materializeOutbox.calls != 1 {
		t.Fatalf("materialize outbox should be enqueued once, got %d", materializeOutbox.calls)
	}
	if len(materializeOutbox.tenantIDs) == 0 || materializeOutbox.tenantIDs[0] != "t" {
		t.Fatalf("outbox tenantID = %v, want t", materializeOutbox.tenantIDs)
	}
	if len(materializeOutbox.revIDs) == 0 || materializeOutbox.revIDs[0] != "r" {
		t.Fatalf("outbox revisionID = %v, want r", materializeOutbox.revIDs)
	}
	if len(materializeOutbox.hashes[0]) == 0 {
		t.Fatal("outbox content hash should be non-empty")
	}
}

func TestFreezeService_Pin_IdempotentWhenAlreadyFrozen(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	finalize := &fakeFreezeFinalizer{}
	materializeOutbox := &fakeMaterializeOutboxEnqueuer{}

	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		finalize,
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{
			snap:           v2dom.TemplateSnapshot{BodyDocxS3Key: "body.docx"},
			valuesFrozenAt: &frozenAt,
		},
		&fakeFinalDocxWriter{},
		&fakeFanoutClient{},
	).WithMaterializeOutbox(materializeOutbox)

	if err := svc.Pin(context.Background(), nil, "t", "r", ApproverContext{}); err != nil {
		t.Fatalf("Pin() error = %v", err)
	}
	if finalize.calls != 0 {
		t.Fatal("WriteFreeze must not be called when already frozen")
	}
	if materializeOutbox.calls != 0 {
		t.Fatal("outbox must not be enqueued when already frozen")
	}
}

func TestFreezeService_Pin_FailsWithoutMaterializeOutbox(t *testing.T) {
	schema := []tmpldom.Placeholder{{ID: "p_user", Required: true}}
	existing := []repository.PlaceholderValue{
		{PlaceholderID: "p_user", ValueText: strPtr("v"), Source: "user"},
	}
	svc := NewFreezeService(
		fakeSchemaReader{placeholders: schema},
		&fakeFillInWriter{},
		&fakeValuesReader{values: existing},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{snap: v2dom.TemplateSnapshot{BodyDocxS3Key: "body.docx"}},
		&fakeFinalDocxWriter{},
		&fakeFanoutClient{},
	)

	err := svc.Pin(context.Background(), nil, "t", "r", ApproverContext{})
	if err == nil || !containsStr(err.Error(), "materialize outbox enqueuer not configured") {
		t.Fatalf("expected materialize outbox error, got %v", err)
	}
}

func TestFreezeService_Materialize_CallsFanoutAndReturnsResult(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	snapReader := fakeSnapshotReader{
		snap: v2dom.TemplateSnapshot{
			BodyDocxS3Key:   "templates/body.docx",
			CompositionJSON: []byte(`{"blocks":[]}`),
		},
		valuesFrozenAt: &frozenAt,
	}
	schema := []tmpldom.Placeholder{
		{ID: "p_user", Name: "user_field", Required: true},
	}
	valuesRead := &fakeValuesReader{values: []repository.PlaceholderValue{
		{PlaceholderID: "p_user", ValueText: strPtr("user-value"), Source: "user"},
	}}
	fanoutClient := &fakeFanoutClient{resp: fanout.FanoutResponse{
		ContentHash:    "deadbeef00000000000000000000000000000000000000000000000000000000",
		FinalDocxS3Key: "final/r.docx",
	}}

	svc := NewFreezeService(
		fakeSchemaReader{placeholders: schema},
		&fakeFillInWriter{},
		valuesRead,
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		snapReader,
		&fakeFinalDocxWriter{},
		fanoutClient,
	)

	result, err := svc.Materialize(context.Background(), "t", "r")
	if err != nil {
		t.Fatalf("Materialize error: %v", err)
	}
	if fanoutClient.calls != 1 {
		t.Fatalf("expected 1 fanout call, got %d", fanoutClient.calls)
	}
	if result.FinalDocxS3Key != "final/r.docx" {
		t.Errorf("FinalDocxS3Key = %q", result.FinalDocxS3Key)
	}
	if bytesToHex(result.ContentHash) != "deadbeef00000000000000000000000000000000000000000000000000000000" {
		t.Errorf("ContentHash = %s", bytesToHex(result.ContentHash))
	}
}

func TestFreezeService_Materialize_ErrorsWhenNotPinned(t *testing.T) {
	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{snap: v2dom.TemplateSnapshot{BodyDocxS3Key: "body.docx"}},
		&fakeFinalDocxWriter{},
		&fakeFanoutClient{},
	)

	_, err := svc.Materialize(context.Background(), "t", "r")
	if err == nil || !containsStr(err.Error(), "not yet pinned") {
		t.Fatalf("expected not-pinned error, got %v", err)
	}
}

func TestFreezeService_Materialize_ErrorsWithoutFanoutClient(t *testing.T) {
	frozenAt := time.Now().UTC()
	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{valuesFrozenAt: &frozenAt},
		&fakeFinalDocxWriter{},
		nil,
	)

	_, err := svc.Materialize(context.Background(), "t", "r")
	if err == nil || !containsStr(err.Error(), "fanout client not configured") {
		t.Fatalf("expected fanout client error, got %v", err)
	}
}

func TestFreezeService_Pin_OutboxEnqueueError_Returns(t *testing.T) {
	schema := []tmpldom.Placeholder{{ID: "p_user", Required: true}}
	existing := []repository.PlaceholderValue{
		{PlaceholderID: "p_user", ValueText: strPtr("v"), Source: "user"},
	}
	materializeOutbox := &fakeMaterializeOutboxEnqueuer{err: errors.New("db error")}

	svc := NewFreezeService(
		fakeSchemaReader{placeholders: schema},
		&fakeFillInWriter{},
		&fakeValuesReader{values: existing},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{snap: v2dom.TemplateSnapshot{BodyDocxS3Key: "body.docx"}},
		&fakeFinalDocxWriter{},
		&fakeFanoutClient{},
	).WithMaterializeOutbox(materializeOutbox)

	err := svc.Pin(context.Background(), nil, "t", "r", ApproverContext{})
	if err == nil {
		t.Fatal("expected error from outbox enqueue, got nil")
	}
}
