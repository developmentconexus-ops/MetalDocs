package application_test

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/audit/application"
	"metaldocs/internal/modules/audit/domain"
)

type captureReader struct {
	query domain.ListEventsQuery
}

func (r *captureReader) ListEvents(_ context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {
	r.query = query
	return nil, nil
}

func TestNewServicePanicsWithoutReader(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil reader")
		}
	}()

	_ = application.NewService(nil)
}

func TestListEventsFailsWhenServiceReaderMissing(t *testing.T) {
	t.Parallel()

	service := &application.Service{}

	_, err := service.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-a"})

	if !errors.Is(err, application.ErrReaderRequired) {
		t.Fatalf("expected ErrReaderRequired, got %v", err)
	}
}

func TestListEventsRequiresTenantID(t *testing.T) {
	t.Parallel()

	service := application.NewService(&captureReader{})

	_, err := service.ListEvents(t.Context(), domain.ListEventsQuery{})

	if !errors.Is(err, application.ErrTenantRequired) {
		t.Fatalf("expected ErrTenantRequired, got %v", err)
	}
}

func TestListEventsPreservesTenantIDDuringNormalization(t *testing.T) {
	t.Parallel()

	reader := &captureReader{}
	service := application.NewService(reader)

	_, err := service.ListEvents(t.Context(), domain.ListEventsQuery{
		ResourceType: " document ",
		ResourceID:   " doc-1 ",
		TenantID:     " tenant-a ",
	})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}

	if reader.query.TenantID != "tenant-a" {
		t.Fatalf("TenantID = %q, want tenant-a", reader.query.TenantID)
	}
	if reader.query.ResourceType != "document" {
		t.Fatalf("ResourceType = %q, want document", reader.query.ResourceType)
	}
	if reader.query.ResourceID != "doc-1" {
		t.Fatalf("ResourceID = %q, want doc-1", reader.query.ResourceID)
	}
	if reader.query.Limit != 50 {
		t.Fatalf("Limit = %d, want 50", reader.query.Limit)
	}
}
