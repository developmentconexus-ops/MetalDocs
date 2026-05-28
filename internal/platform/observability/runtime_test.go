package observability

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStaticRuntimeStatusProviderReadyDegradesOnDependencyFailure(t *testing.T) {
	provider := NewStaticRuntimeStatusProvider("memory", "disk", true, DependencyCheck{
		Name: "gotenberg",
		Check: func(context.Context) (DependencyCheckResult, error) {
			return DependencyCheckResult{}, errors.New("downstream unavailable")
		},
	})

	code, payload := provider.Ready(context.Background())
	if code != 503 {
		t.Fatalf("code = %d, want 503", code)
	}
	if payload["status"] != "degraded" {
		t.Fatalf("status = %v, want degraded", payload["status"])
	}
}

func TestPostgresRuntimeStatusProviderReadyUsesSharedDeadlineAcrossChecks(t *testing.T) {
	provider := NewPostgresRuntimeStatusProvider(nil, "postgres", "minio", true,
		DependencyCheck{
			Name: "first",
			Check: func(ctx context.Context) (DependencyCheckResult, error) {
				<-ctx.Done()
				return DependencyCheckResult{}, ctx.Err()
			},
		},
		DependencyCheck{
			Name: "second",
			Check: func(ctx context.Context) (DependencyCheckResult, error) {
				<-ctx.Done()
				return DependencyCheckResult{}, ctx.Err()
			},
		},
	)

	start := time.Now()
	code, payload := provider.Ready(context.Background())
	elapsed := time.Since(start)

	if code != 503 {
		t.Fatalf("code = %d, want 503", code)
	}
	if payload["status"] != "degraded" {
		t.Fatalf("status = %v, want degraded", payload["status"])
	}
	if elapsed >= 3500*time.Millisecond {
		t.Fatalf("Ready took %v, want shared readiness budget under 3.5s", elapsed)
	}
}

func TestPostgresRuntimeStatusProviderRuntimeMetricsOmitsFailedSections(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM metaldocs.auth_identities").
		WillReturnError(errors.New("auth query failed"))
	mock.ExpectQuery("FROM metaldocs.auth_sessions").
		WillReturnRows(sqlmock.NewRows([]string{"active", "expired", "revoked"}).AddRow(7, 2, 1))
	mock.ExpectQuery("FROM metaldocs.outbox_events").
		WillReturnError(errors.New("outbox query failed"))

	provider := NewPostgresRuntimeStatusProvider(db, "postgres", "minio", true)
	metrics := provider.RuntimeMetrics(context.Background())

	if metrics["repositoryStatus"] != "degraded" {
		t.Fatalf("repositoryStatus = %v, want degraded", metrics["repositoryStatus"])
	}
	authSection, ok := metrics["auth"].(map[string]any)
	if !ok {
		t.Fatalf("auth section missing")
	}
	if _, exists := authSection["users"]; exists {
		t.Fatalf("users metrics should be omitted on query failure")
	}
	if _, exists := metrics["worker"]; exists {
		t.Fatalf("worker metrics should be omitted on query failure")
	}
	sessions, ok := authSection["sessions"].(map[string]any)
	if !ok {
		t.Fatalf("sessions metrics missing")
	}
	if sessions["active"] != 7 {
		t.Fatalf("sessions active = %v, want 7", sessions["active"])
	}
	errs, ok := metrics["errors"].(map[string]string)
	if !ok {
		t.Fatalf("errors map missing or wrong type: %T", metrics["errors"])
	}
	if !strings.Contains(errs["auth"], "auth query failed") {
		t.Fatalf("auth error = %q", errs["auth"])
	}
	if !strings.Contains(errs["worker"], "outbox query failed") {
		t.Fatalf("worker error = %q", errs["worker"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
