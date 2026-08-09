package application

import (
	"context"
	"errors"
	"testing"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	tmpldom "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

type fakeMaterializeOutboxEnqueuer struct {
	calls         int
	tenantIDs     []string
	revIDs        []string
	hashes        [][]byte
	generationIDs []string
	err           error
}

func (f *fakeMaterializeOutboxEnqueuer) EnqueueMaterializeTx(_ context.Context, _ db.Tx, tenantID, revisionID string, contentHash []byte, releaseGenerationID string) error {
	if f.err != nil {
		return f.err
	}
	f.calls++
	f.tenantIDs = append(f.tenantIDs, tenantID)
	f.revIDs = append(f.revIDs, revisionID)
	f.hashes = append(f.hashes, append([]byte(nil), contentHash...))
	f.generationIDs = append(f.generationIDs, releaseGenerationID)
	return nil
}

func TestFreezeService_Pin_NoNetworkCall(t *testing.T) {
	resolverKey := "doc_code"
	schema := []tmpldom.Placeholder{
		{ID: "p_user", Name: "user_field", Required: true},
		{ID: "p_comp", Name: "doc_code_field", Computed: true, ResolverKey: &resolverKey},
	}
	existing := []infrastructure.PlaceholderValue{
		{PlaceholderID: "p_user", ValueText: strPtr("user-value"), Source: "user"},
	}
	writer := &fakeFillInWriter{}
	valuesRead := &fakeValuesReader{values: existing}
	reg := resolvers.NewRegistry()
	reg.Register(fixedResolver{key: "doc_code", ver: 3, val: "DOC-001"})
	finalize := &fakeFreezeFinalizer{}
	ctxBuilder := &fakeResolverContextBuilder{input: resolvers.ResolveInput{TenantID: "t", RevisionID: "r"}}
	snapReader := fakeSnapshotReader{
		snap: v2dom.TemplateSnapshot{
			BodyDocxS3Key:   "templates/body.docx",
			CompositionJSON: []byte(`{}`),
		},
		currentRef: v2dom.RevisionRef{
			ID:         "11111111-1111-4111-8111-111111111111",
			StorageKey: "documents/d/revisions/head.docx",
		},
	}
	fanoutClient := &fakeFanoutClient{}
	materializeOutbox := &fakeMaterializeOutboxEnqueuer{}

	svc := NewFreezeService(fakeSchemaReader{placeholders: schema}, writer, valuesRead, reg, finalize, ctxBuilder, snapReader, fanoutClient).
		WithMaterializeOutbox(materializeOutbox)

	if err := svc.Pin(context.Background(), fakeTx{}, "t", "r", ApproverContext{}, ""); err != nil {
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
	// Migration 0313: the freeze write must pin WHICH document_revisions row
	// was frozen, and it must be the revision id from the seam — never the
	// documents.id ("r") the pipeline's revisionID parameter carries.
	if finalize.pinnedRevisionID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("WriteFreeze pinned %q, want the current revision id", finalize.pinnedRevisionID)
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
		&fakeFanoutClient{},
	).WithMaterializeOutbox(materializeOutbox)

	if err := svc.Pin(context.Background(), fakeTx{}, "t", "r", ApproverContext{}, ""); err != nil {
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
	existing := []infrastructure.PlaceholderValue{
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
		&fakeFanoutClient{},
	)

	err := svc.Pin(context.Background(), fakeTx{}, "t", "r", ApproverContext{}, "")
	if err == nil || !containsStr(err.Error(), "materialize outbox enqueuer not configured") {
		t.Fatalf("expected materialize outbox error, got %v", err)
	}
}

func TestFreezeService_Materialize_CallsFanoutAndReturnsResult(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	pinnedBody := []byte("PK\x03\x04 pinned revision body")
	snapReader := fakeSnapshotReader{
		snap: v2dom.TemplateSnapshot{
			BodyDocxS3Key:   "templates/body.docx",
			CompositionJSON: []byte(`{"blocks":[]}`),
		},
		valuesFrozenAt: &frozenAt,
		frozenRef: v2dom.RevisionRef{
			ID:          "22222222-2222-4222-8222-222222222222",
			StorageKey:  "documents/d/revisions/abc.docx",
			ContentHash: hashOf(pinnedBody),
		},
	}
	bodies := &fakeRevisionBodyReader{bodies: map[string][]byte{
		"documents/d/revisions/abc.docx": pinnedBody,
	}}
	schema := []tmpldom.Placeholder{
		{ID: "p_user", Name: "user_field", Required: true},
	}
	valuesRead := &fakeValuesReader{values: []infrastructure.PlaceholderValue{
		{PlaceholderID: "p_user", ValueText: strPtr("user-value"), Source: "user"},
	}}
	fanoutClient := &fakeFanoutClient{resp: fanout.Response{
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
		fanoutClient,
	).WithRevisionBodyReader(bodies)

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
	// F-QA3-1: the frozen body is the editor revision, never the template
	// snapshot. Freezing the snapshot is what produced blank signed artifacts.
	if fanoutClient.req.BodyDocxS3Key != "documents/d/revisions/abc.docx" {
		t.Errorf("BodyDocxS3Key = %q, want the pinned editor revision", fanoutClient.req.BodyDocxS3Key)
	}
	if len(bodies.keys) != 1 || bodies.keys[0] != "documents/d/revisions/abc.docx" {
		t.Errorf("verification read keys = %v, want the pinned key exactly once", bodies.keys)
	}
}

// Migration 0313 lineage: Materialize must render the revision the PIN names,
// even when the document head has moved on since the freeze. Rendering the head
// is exactly the drift the pin exists to prevent.
func TestFreezeService_Materialize_ReadsPinNotCurrentRevision(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	pinnedBody := []byte("the approved body")
	headBody := []byte("edits made after the freeze")

	snapReader := fakeSnapshotReader{
		snap:           v2dom.TemplateSnapshot{CompositionJSON: []byte(`{}`)},
		valuesFrozenAt: &frozenAt,
		currentRef: v2dom.RevisionRef{
			ID:          "33333333-3333-4333-8333-333333333333",
			StorageKey:  "documents/d/revisions/head.docx",
			ContentHash: hashOf(headBody),
		},
		frozenRef: v2dom.RevisionRef{
			ID:          "22222222-2222-4222-8222-222222222222",
			StorageKey:  "documents/d/revisions/pinned.docx",
			ContentHash: hashOf(pinnedBody),
		},
	}
	bodies := &fakeRevisionBodyReader{bodies: map[string][]byte{
		"documents/d/revisions/pinned.docx": pinnedBody,
		"documents/d/revisions/head.docx":   headBody,
	}}
	fanoutClient := &fakeFanoutClient{resp: fanout.Response{
		ContentHash:    "deadbeef00000000000000000000000000000000000000000000000000000000",
		FinalDocxS3Key: "final/r.docx",
	}}

	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		snapReader,
		fanoutClient,
	).WithRevisionBodyReader(bodies)

	if _, err := svc.Materialize(context.Background(), "t", "r"); err != nil {
		t.Fatalf("Materialize error: %v", err)
	}
	if fanoutClient.req.BodyDocxS3Key != "documents/d/revisions/pinned.docx" {
		t.Fatalf("BodyDocxS3Key = %q, want the PINNED revision, not the head",
			fanoutClient.req.BodyDocxS3Key)
	}
}

// Fail-closed lineage: if the bytes under the pinned revision's key do not hash
// to that revision's recorded content_hash, the job fails with a distinct error
// and NOTHING is rendered — a signature must never land on substituted bytes.
func TestFreezeService_Materialize_FailsClosedOnHashMismatch(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	storedBody := []byte("bytes that are NOT what the revision recorded")

	snapReader := fakeSnapshotReader{
		snap:           v2dom.TemplateSnapshot{CompositionJSON: []byte(`{}`)},
		valuesFrozenAt: &frozenAt,
		frozenRef: v2dom.RevisionRef{
			ID:          "22222222-2222-4222-8222-222222222222",
			StorageKey:  "documents/d/revisions/pinned.docx",
			ContentHash: hashOf([]byte("the bytes that were approved")),
		},
	}
	bodies := &fakeRevisionBodyReader{bodies: map[string][]byte{
		"documents/d/revisions/pinned.docx": storedBody,
	}}
	fanoutClient := &fakeFanoutClient{}

	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		snapReader,
		fanoutClient,
	).WithRevisionBodyReader(bodies)

	_, err := svc.Materialize(context.Background(), "t", "r")
	if !errors.Is(err, ErrFrozenBodyHashMismatch) {
		t.Fatalf("expected ErrFrozenBodyHashMismatch, got %v", err)
	}
	if fanoutClient.calls != 0 {
		t.Fatalf("fanout must not run on a mismatch, got %d call(s)", fanoutClient.calls)
	}
}

// A pinned revision with no recorded content_hash is unverifiable, and
// unverifiable is a failure — not a reason to skip the check (no-fallback).
func TestFreezeService_Materialize_FailsClosedWithoutRecordedHash(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	snapReader := fakeSnapshotReader{
		snap:           v2dom.TemplateSnapshot{CompositionJSON: []byte(`{}`)},
		valuesFrozenAt: &frozenAt,
		frozenRef: v2dom.RevisionRef{
			ID:         "22222222-2222-4222-8222-222222222222",
			StorageKey: "documents/d/revisions/pinned.docx",
		},
	}
	fanoutClient := &fakeFanoutClient{}

	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		snapReader,
		fanoutClient,
	).WithRevisionBodyReader(&fakeRevisionBodyReader{})

	_, err := svc.Materialize(context.Background(), "t", "r")
	if !errors.Is(err, ErrFrozenBodyHashMismatch) {
		t.Fatalf("expected ErrFrozenBodyHashMismatch, got %v", err)
	}
	if fanoutClient.calls != 0 {
		t.Fatalf("fanout must not run without a recorded hash, got %d call(s)", fanoutClient.calls)
	}
}

// Without a body reader there is no way to verify lineage at all; Materialize
// refuses rather than rendering unverified bytes.
func TestFreezeService_Materialize_FailsWithoutBodyReader(t *testing.T) {
	frozenAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	fanoutClient := &fakeFanoutClient{}

	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{
			snap:           v2dom.TemplateSnapshot{CompositionJSON: []byte(`{}`)},
			valuesFrozenAt: &frozenAt,
			frozenRef: v2dom.RevisionRef{
				ID:          "22222222-2222-4222-8222-222222222222",
				StorageKey:  "documents/d/revisions/pinned.docx",
				ContentHash: hashOf([]byte("body")),
			},
		},
		fanoutClient,
	)

	_, err := svc.Materialize(context.Background(), "t", "r")
	if err == nil || !containsStr(err.Error(), "revision body reader not configured") {
		t.Fatalf("expected body-reader error, got %v", err)
	}
	if fanoutClient.calls != 0 {
		t.Fatalf("fanout must not run unverified, got %d call(s)", fanoutClient.calls)
	}
}

// F-QA3-1 fail-closed guard: a document with no PINNED revision must NOT be
// materialized from the template snapshot (nor from the current head) — a
// signed artifact that omits the reviewed content is worse than a failed job.
func TestFreezeService_Materialize_ErrorsWithoutPinnedRevisionBody(t *testing.T) {
	frozenAt := time.Now().UTC()
	fanoutClient := &fakeFanoutClient{}
	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{
			snap:           v2dom.TemplateSnapshot{BodyDocxS3Key: "templates/body.docx"},
			valuesFrozenAt: &frozenAt,
		},
		fanoutClient,
	)

	_, err := svc.Materialize(context.Background(), "t", "r")
	if err == nil || !containsStr(err.Error(), "no pinned frozen revision body") {
		t.Fatalf("expected no-pinned-revision error, got %v", err)
	}
	if fanoutClient.calls != 0 {
		t.Fatalf("fanout must not be called without an editor body, got %d calls", fanoutClient.calls)
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
		nil,
	)

	_, err := svc.Materialize(context.Background(), "t", "r")
	if err == nil || !containsStr(err.Error(), "fanout client not configured") {
		t.Fatalf("expected fanout client error, got %v", err)
	}
}

func TestFreezeService_Pin_OutboxEnqueueError_Returns(t *testing.T) {
	schema := []tmpldom.Placeholder{{ID: "p_user", Required: true}}
	existing := []infrastructure.PlaceholderValue{
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
		&fakeFanoutClient{},
	).WithMaterializeOutbox(materializeOutbox)

	err := svc.Pin(context.Background(), fakeTx{}, "t", "r", ApproverContext{}, "")
	if err == nil {
		t.Fatal("expected error from outbox enqueue, got nil")
	}
}

func TestFreezeService_Pin_RequiresTx(t *testing.T) {
	svc := NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		&fakeFreezeFinalizer{},
		&fakeResolverContextBuilder{},
		fakeSnapshotReader{},
		&fakeFanoutClient{},
	).WithMaterializeOutbox(&fakeMaterializeOutboxEnqueuer{})
	err := svc.Pin(context.Background(), nil, "t", "r", ApproverContext{}, "")
	if err == nil || !containsStr(err.Error(), "tx required") {
		t.Fatalf("expected tx-required error, got %v", err)
	}
}
