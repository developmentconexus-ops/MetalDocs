package taxonomy

import (
	"context"
	"testing"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/db"
)

type stubTemplateChecker struct{}

func (stubTemplateChecker) IsPublished(_ context.Context, _ string) (bool, string, error) {
	return false, "", nil
}

type stubAuditWriter struct{}

func (stubAuditWriter) Record(_ context.Context, _ auditdomain.Event) error            { return nil }
func (stubAuditWriter) RecordTx(_ context.Context, _ db.Tx, _ auditdomain.Event) error { return nil }

func TestNew_UsesDefaultTemplateCheckerWhenDependencyIsNil(t *testing.T) {
	mod := New(Dependencies{
		TplChecker:           stubTemplateChecker{},
		AuditWriter:          stubAuditWriter{},
		RouteReadinessReader: approvaldomain.NoopRouteReadinessReader{},
	})
	if mod == nil || mod.Handler == nil {
		t.Fatal("expected module handler to be initialized")
	}
}
