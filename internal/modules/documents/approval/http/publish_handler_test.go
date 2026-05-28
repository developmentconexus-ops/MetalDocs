package approvalhttp

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeReadServicePublish struct {
	inst *domain.Instance
	err  error
}

func (f *fakeReadServicePublish) LoadInstance(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	return nil, nil
}

func (f *fakeReadServicePublish) LoadActiveInstanceByDocument(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	return f.inst, f.err
}

func (f *fakeReadServicePublish) ListPendingForActor(_ context.Context, _ *sql.DB, _, _, _ string, _, _ int) ([]domain.Instance, error) {
	return nil, nil
}

func (f *fakeReadServicePublish) ListInboxItems(_ context.Context, _ *sql.DB, _, _, _ string, _, _ int) ([]application.InboxView, error) {
	return nil, nil
}

func (f *fakeReadServicePublish) CountPendingForActor(_ context.Context, _ *sql.DB, _, _, _ string) (int, error) {
	return 0, nil
}

func publishTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/documents/{id}/publish", h.PublishHandler)
	mux.HandleFunc("POST /api/v1/documents/{id}/schedule-publish", h.SchedulePublishHandler)
	return mux
}

func TestPublishHandler(t *testing.T) {
	origPublishApproved := publishApproved
	t.Cleanup(func() {
		publishApproved = origPublishApproved
	})

	tests := []struct {
		name       string
		svcErr     error
		wantStatus int
	}{
		{
			name:       "happy path",
			wantStatus: http.StatusOK,
		},
		{
			name:       "authz denied",
			svcErr:     authz.ErrCapDenied{Capability: "doc.publish", AreaCode: "tenant-1", ActorID: "actor-1"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "illegal transition",
			svcErr:     application.ErrInstanceNotApproved,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "stale occ",
			svcErr:     repository.ErrStaleRevision,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "invalid supersede target",
			svcErr:     repository.ErrInvalidScheduledSupersedeTarget,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq application.PublishRequest
			publishApproved = func(_ *Handler, _ context.Context, _ *sql.DB, req application.PublishRequest) (application.PublishResult, error) {
				gotReq = req
				if tt.svcErr != nil {
					return application.PublishResult{}, tt.svcErr
				}
				return application.PublishResult{DocumentID: "doc-1", NewStatus: "published"}, nil
			}

			fakeRead := &fakeReadServicePublish{
				inst: &domain.Instance{ID: "inst-1", DocumentID: "doc-1"},
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/publish", nil)
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "idem-1")
			req.Header.Set("If-Match", "\"v3\"")

			rr := httptest.NewRecorder()
			publishTestMux(&Handler{readSvc: fakeRead}).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if gotReq.TenantID != "tenant-1" || gotReq.InstanceID != "inst-1" || gotReq.PublishedBy != "actor-1" {
				t.Fatalf("unexpected service request: %+v", gotReq)
			}

			if tt.wantStatus == http.StatusOK {
				var out contracts.PublishResponse
				if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if out.DocumentID != "doc-1" || out.NewStatus != "published" {
					t.Fatalf("unexpected response: %+v", out)
				}
			}
		})
	}
}

func TestSchedulePublishHandler(t *testing.T) {
	origSchedulePublish := schedulePublish
	t.Cleanup(func() {
		schedulePublish = origSchedulePublish
	})

	tests := []struct {
		name       string
		svcErr     error
		wantStatus int
	}{
		{
			name:       "happy path",
			wantStatus: http.StatusOK,
		},
		{
			name:       "authz denied",
			svcErr:     authz.ErrCapDenied{Capability: "doc.publish", AreaCode: "tenant-1", ActorID: "actor-1"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "illegal transition",
			svcErr:     application.ErrInstanceNotApproved,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "stale occ",
			svcErr:     repository.ErrStaleRevision,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq application.SchedulePublishRequest
			schedulePublish = func(_ *Handler, _ context.Context, _ *sql.DB, req application.SchedulePublishRequest) (application.SchedulePublishResult, error) {
				gotReq = req
				if tt.svcErr != nil {
					return application.SchedulePublishResult{}, tt.svcErr
				}
				return application.SchedulePublishResult{
					DocumentID:    "doc-1",
					EffectiveDate: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
				}, nil
			}
			fakeRead := &fakeReadServicePublish{
				inst: &domain.Instance{ID: "inst-1", DocumentID: "doc-1", RevisionVersion: 4},
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/schedule-publish", strings.NewReader(`{"effective_from":"2026-05-01T12:00:00Z"}`))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "idem-1")
			req.Header.Set("If-Match", "\"v4\"")

			rr := httptest.NewRecorder()
			publishTestMux(&Handler{readSvc: fakeRead}).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if gotReq.TenantID != "tenant-1" || gotReq.InstanceID != "inst-1" || gotReq.ScheduledBy != "actor-1" {
				t.Fatalf("unexpected service request: %+v", gotReq)
			}
			if gotReq.EffectiveDate.UTC().Format(time.RFC3339) != "2026-05-01T12:00:00Z" {
				t.Fatalf("effective date = %s", gotReq.EffectiveDate.UTC().Format(time.RFC3339))
			}

			if tt.wantStatus == http.StatusOK {
				var out contracts.PublishResponse
				if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if out.DocumentID != "doc-1" || out.NewStatus != "scheduled" || out.EffectiveFrom != "2026-05-01T12:00:00Z" {
					t.Fatalf("unexpected response: %+v", out)
				}
			}
		})
	}
}

func TestSchedulePublishHandler_ResolvesActiveInstanceFromDocument(t *testing.T) {
	origSchedulePublish := schedulePublish
	t.Cleanup(func() {
		schedulePublish = origSchedulePublish
	})

	var gotReq application.SchedulePublishRequest
	schedulePublish = func(_ *Handler, _ context.Context, _ *sql.DB, req application.SchedulePublishRequest) (application.SchedulePublishResult, error) {
		gotReq = req
		return application.SchedulePublishResult{
			DocumentID:    "doc-1",
			EffectiveDate: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		}, nil
	}

	fakeRead := &fakeReadServicePublish{
		inst: &domain.Instance{ID: "inst-99", DocumentID: "doc-1", RevisionVersion: 7},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/schedule-publish", strings.NewReader(`{"effective_from":"2026-05-01T12:00:00Z"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v7\"")

	rr := httptest.NewRecorder()
	publishTestMux(&Handler{readSvc: fakeRead}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotReq.InstanceID != "inst-99" {
		t.Fatalf("instance id = %q, want %q", gotReq.InstanceID, "inst-99")
	}
}

func TestSchedulePublishHandler_PassesExpectedRevisionAndSupersededDocumentID(t *testing.T) {
	origSchedulePublish := schedulePublish
	t.Cleanup(func() {
		schedulePublish = origSchedulePublish
	})

	var gotReq application.SchedulePublishRequest
	schedulePublish = func(_ *Handler, _ context.Context, _ *sql.DB, req application.SchedulePublishRequest) (application.SchedulePublishResult, error) {
		gotReq = req
		return application.SchedulePublishResult{
			DocumentID:    "doc-approved-1",
			EffectiveDate: time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC),
		}, nil
	}

	fakeRead := &fakeReadServicePublish{
		inst: &domain.Instance{ID: "inst-1", DocumentID: "doc-approved-1", RevisionVersion: 2},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-approved-1/schedule-publish", strings.NewReader(`{
		"effective_from":"2026-05-21T14:00:00Z",
		"superseded_document_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "admin", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v2\"")

	rr := httptest.NewRecorder()
	publishTestMux(&Handler{readSvc: fakeRead}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	if gotReq.ExpectedRevision != 2 {
		t.Fatalf("expected revision = %d, want 2", gotReq.ExpectedRevision)
	}
	if gotReq.SupersededDocumentID != "3fa85f64-5717-4562-b3fc-2c963f66afa6" {
		t.Fatalf("superseded document = %q, want 3fa85f64-5717-4562-b3fc-2c963f66afa6", gotReq.SupersededDocumentID)
	}
}

func TestSchedulePublishHandler_ReturnsErrorWhenActiveInstanceLookupFails(t *testing.T) {
	fakeRead := &fakeReadServicePublish{
		err: repository.ErrNoActiveInstance,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-approved-1/schedule-publish", strings.NewReader(`{
		"effective_from":"2026-05-21T14:00:00Z",
		"superseded_document_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "admin", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v2\"")

	rr := httptest.NewRecorder()
	publishTestMux(&Handler{readSvc: fakeRead}).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}
