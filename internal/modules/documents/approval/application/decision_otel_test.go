package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	platformdb "metaldocs/internal/platform/db"
)

// stubApprovalRepo implements repository.ApprovalRepository; all methods return
// zero/nil or a sentinel error so the tx function short-circuits on first call.
type stubApprovalRepo struct{}

func (stubApprovalRepo) InsertInstance(_ context.Context, _ platformdb.Tx, _ domain.Instance) error {
	return nil
}
func (stubApprovalRepo) InsertStageInstances(_ context.Context, _ platformdb.Tx, _ []domain.StageInstance) error {
	return nil
}
func (stubApprovalRepo) InsertSignoff(_ context.Context, _ platformdb.Tx, _ domain.Signoff) (repository.SignoffInsertResult, error) {
	return repository.SignoffInsertResult{}, nil
}
func (stubApprovalRepo) LoadSignoffByActor(_ context.Context, _ platformdb.Tx, _, _, _ string) (*domain.Signoff, error) {
	return nil, nil
}
func (stubApprovalRepo) LoadInstance(_ context.Context, _ platformdb.Tx, _, _ string) (*domain.Instance, error) {
	return nil, errors.New("stub repo")
}
func (stubApprovalRepo) LoadInstancesByIDs(_ context.Context, _ platformdb.Tx, _ string, _ []string) ([]domain.Instance, error) {
	return nil, nil
}
func (stubApprovalRepo) LoadActiveInstanceByDocument(_ context.Context, _ platformdb.Tx, _, _ string) (*domain.Instance, error) {
	return nil, nil
}
func (stubApprovalRepo) ValidateScheduledSupersedeTarget(_ context.Context, _ platformdb.Tx, _, _, _ string) error {
	return nil
}
func (stubApprovalRepo) LoadCurrentPublishedHeadForDocument(_ context.Context, _ platformdb.Tx, _, _ string) (string, error) {
	return "", nil
}
func (stubApprovalRepo) LoadCurrentPublishedHead(_ context.Context, _ platformdb.Tx, _, _ string) (string, error) {
	return "", nil
}
func (stubApprovalRepo) GetDocumentRevisionVersion(_ context.Context, _ platformdb.Tx, _, _ string) (int, error) {
	return 0, nil
}
func (stubApprovalRepo) ListRoutes(_ context.Context, _ string) ([]repository.Route, error) {
	return nil, nil
}
func (stubApprovalRepo) ListRoutesTx(_ context.Context, _ platformdb.Tx, _ string) ([]repository.Route, error) {
	return nil, nil
}
func (stubApprovalRepo) MarkSuperseded(_ context.Context, _ platformdb.Tx, _, _ string) error {
	return nil
}
func (stubApprovalRepo) UpdateStageStatus(_ context.Context, _ platformdb.Tx, _, _ string, _, _ domain.StageStatus) error {
	return nil
}
func (stubApprovalRepo) UpdateInstanceStatus(_ context.Context, _ platformdb.Tx, _, _ string, _ domain.InstanceStatus, _ domain.InstanceStatus, _ *time.Time) error {
	return nil
}
func (stubApprovalRepo) LoadPriorSignoffs(_ context.Context, _ platformdb.Tx, _, _, _ string) ([]domain.Signoff, error) {
	return nil, nil
}
func (stubApprovalRepo) LoadStageSignoffs(_ context.Context, _ platformdb.Tx, _, _ string) ([]domain.Signoff, error) {
	return nil, nil
}
func (stubApprovalRepo) HasUnresolvedComments(_ context.Context, _ platformdb.Tx, _, _ string) (bool, error) {
	return false, nil
}
func (stubApprovalRepo) LoadActiveDocumentContentHash(_ context.Context, _ platformdb.Tx, _, _ string) (string, error) {
	return "", nil
}
func (stubApprovalRepo) ResolveEligibleActors(_ context.Context, _ platformdb.Tx, _, _, _ string) ([]string, error) {
	return nil, nil
}
func (stubApprovalRepo) LoadActorDisplayName(_ context.Context, _, _ string) (string, error) {
	return "", errors.New("stub repo")
}
func (stubApprovalRepo) LoadRoute(_ context.Context, _ platformdb.Tx, _, _ string) (domain.Route, error) {
	return domain.Route{}, nil
}

// stubDecisionTxRunner calls fn(ctx) so the flow enters the tx func and hits the repo.
type stubDecisionTxRunner struct{}

func (stubDecisionTxRunner) Do(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return fn(nil)
}
func (stubDecisionTxRunner) DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return fn(nil)
}

// stubEventEmitter satisfies application.EventEmitter.
type stubEventEmitter struct{}

func (stubEventEmitter) Emit(_ context.Context, _ platformdb.Tx, _ application.GovernanceEvent) error {
	return nil
}

// stubClock satisfies application.Clock.
type stubClock struct{}

func (stubClock) Now() time.Time { return time.Time{} }

func TestRecordSignoff_EmitsSignoffRecordSpan(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	svc := application.NewDecisionService(
		stubApprovalRepo{},
		stubEventEmitter{},
		stubClock{},
	)

	req := application.SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      "inst-1",
		StageInstanceID: "stage-1",
		ActorUserID:     "user-1",
		Decision:        domain.Decision("approve"),
	}

	_, _ = svc.RecordSignoff(context.Background(), stubDecisionTxRunner{}, req)

	spans := sr.Ended()
	for _, s := range spans {
		if s.Name() == "signoff.record" {
			for _, a := range s.Attributes() {
				if a.Key == attribute.Key("signoff.verdict") {
					if a.Value.AsString() != string(req.Decision) {
						t.Errorf("signoff.verdict: got %q, want %q", a.Value.AsString(), string(req.Decision))
					}
				}
			}
			return
		}
	}
	names := make([]string, len(spans))
	for i, s := range spans {
		names[i] = s.Name()
	}
	t.Fatalf("want span named signoff.record, got %v", names)
}
