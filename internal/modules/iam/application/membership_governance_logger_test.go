package application

import (
	"context"
	"testing"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/requesttrace"
)

// recordingAuditWriter captures RecordTx calls so the mapping can be asserted.
// Record (post-commit) must never be called by the in-tx logger.
type recordingAuditWriter struct {
	txEvents   []auditdomain.Event
	plainCalls int
}

func (w *recordingAuditWriter) Record(_ context.Context, _ auditdomain.Event) error {
	w.plainCalls++
	return nil
}

func (w *recordingAuditWriter) RecordTx(_ context.Context, _ db.Tx, event auditdomain.Event) error {
	w.txEvents = append(w.txEvents, event)
	return nil
}

// IsErased satisfies auditdomain.Writer's embedded ErasureChecker. This
// fixture never simulates an erased tenant.
func (w *recordingAuditWriter) IsErased(context.Context, string) (bool, error) {
	return false, nil
}

func grantedByPtr(s string) *string { return &s }

func TestAuditMembershipLogger_LogTx_MapsGrant(t *testing.T) {
	writer := &recordingAuditWriter{}
	logger := NewAuditMembershipLogger(writer)

	ctx := requesttrace.WithTraceID(context.Background(), "trace-abc")
	membership := domain.UserProcessArea{
		UserID:    "u1",
		TenantID:  "t1",
		AreaCode:  "A1",
		Role:      domain.RoleEditor,
		GrantedBy: grantedByPtr("admin"),
	}

	if err := logger.LogTx(ctx, noopMembershipTx{}, "role.grant", membership); err != nil {
		t.Fatalf("LogTx: %v", err)
	}
	if writer.plainCalls != 0 {
		t.Fatalf("post-commit Record must not be called, got %d", writer.plainCalls)
	}
	if len(writer.txEvents) != 1 {
		t.Fatalf("expected 1 in-tx event, got %d", len(writer.txEvents))
	}
	ev := writer.txEvents[0]
	if ev.Action != "iam.area_membership.granted" {
		t.Fatalf("Action = %q, want iam.area_membership.granted", ev.Action)
	}
	if ev.ResourceType != "area_membership" {
		t.Fatalf("ResourceType = %q, want area_membership", ev.ResourceType)
	}
	if ev.ResourceID != "u1" {
		t.Fatalf("ResourceID = %q, want u1", ev.ResourceID)
	}
	if ev.ActorID != "admin" {
		t.Fatalf("ActorID = %q, want admin (from GrantedBy)", ev.ActorID)
	}
	if ev.TenantID != "t1" {
		t.Fatalf("TenantID = %q, want t1", ev.TenantID)
	}
	if ev.TraceID != "trace-abc" {
		t.Fatalf("TraceID = %q, want trace-abc (recovered from ctx)", ev.TraceID)
	}
	if ev.PayloadJSON != `{"area_code":"A1","role":"editor"}` {
		t.Fatalf("PayloadJSON = %q, want area_code+role", ev.PayloadJSON)
	}
	if ev.ID == "" {
		t.Fatal("ID must be set")
	}
}

func TestAuditMembershipLogger_LogTx_MapsRevoke(t *testing.T) {
	writer := &recordingAuditWriter{}
	logger := NewAuditMembershipLogger(writer)

	membership := domain.UserProcessArea{
		UserID:    "u9",
		TenantID:  "t1",
		AreaCode:  "A2",
		Role:      domain.RoleApprover,
		GrantedBy: grantedByPtr("revoker"),
	}

	if err := logger.LogTx(context.Background(), noopMembershipTx{}, "role.revoke", membership); err != nil {
		t.Fatalf("LogTx: %v", err)
	}
	if len(writer.txEvents) != 1 {
		t.Fatalf("expected 1 in-tx event, got %d", len(writer.txEvents))
	}
	if got := writer.txEvents[0].Action; got != "iam.area_membership.revoked" {
		t.Fatalf("Action = %q, want iam.area_membership.revoked", got)
	}
}

func TestAuditMembershipLogger_LogTx_NilTxErrors(t *testing.T) {
	writer := &recordingAuditWriter{}
	logger := NewAuditMembershipLogger(writer)

	err := logger.LogTx(context.Background(), nil, "role.grant", domain.UserProcessArea{})
	if err == nil {
		t.Fatal("expected error for nil tx, got nil")
	}
	if len(writer.txEvents) != 0 {
		t.Fatalf("no event must be written on nil tx, got %d", len(writer.txEvents))
	}
}

func TestNewAuditMembershipLogger_NilWriterPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil writer")
		}
	}()
	NewAuditMembershipLogger(nil)
}
