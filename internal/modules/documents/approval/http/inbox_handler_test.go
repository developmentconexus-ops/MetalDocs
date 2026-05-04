package approvalhttp

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

type fakeReadServiceInbox struct {
	views []application.InboxView
	total int
	err   error

	called      bool
	gotTenantID string
	gotActorID  string
	gotAreaCode string
	gotLimit    int
	gotOffset   int
}

func (f *fakeReadServiceInbox) LoadInstance(_ context.Context, _ *sql.DB, _, _, _ string) (*domain.Instance, error) {
	return nil, nil
}

func (f *fakeReadServiceInbox) LoadActiveInstanceByDocument(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	return nil, nil
}

func (f *fakeReadServiceInbox) ListPendingForActor(_ context.Context, _ *sql.DB, _, _, _ string, _, _ int) ([]domain.Instance, error) {
	return nil, nil
}

func (f *fakeReadServiceInbox) ListInboxItems(_ context.Context, _ *sql.DB, tenantID, actorID, areaCode string, limit, offset int) ([]application.InboxView, error) {
	f.called = true
	f.gotTenantID = tenantID
	f.gotActorID = actorID
	f.gotAreaCode = areaCode
	f.gotLimit = limit
	f.gotOffset = offset
	if f.err != nil {
		return nil, f.err
	}
	return f.views, nil
}

func (f *fakeReadServiceInbox) CountPendingForActor(_ context.Context, _ *sql.DB, _, _, _ string) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.total, nil
}

func inboxTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v2/approval/inbox", h.InboxHandler)
	return mux
}

func TestInboxHandler_HappyEmptyList(t *testing.T) {
	fakeSvc := &fakeReadServiceInbox{}
	h := &Handler{readSvc: fakeSvc}
	mux := inboxTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/approval/inbox", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !fakeSvc.called {
		t.Fatalf("expected read service call")
	}
	if fakeSvc.gotLimit != 25 || fakeSvc.gotOffset != 0 {
		t.Fatalf("unexpected defaults limit=%d offset=%d", fakeSvc.gotLimit, fakeSvc.gotOffset)
	}
}

func TestInboxHandler_ValidLimitParam(t *testing.T) {
	fakeSvc := &fakeReadServiceInbox{}
	h := &Handler{readSvc: fakeSvc}
	mux := inboxTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/approval/inbox?area_code=finance&limit=40&offset=10", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakeSvc.gotAreaCode != "finance" {
		t.Fatalf("area_code = %q, want %q", fakeSvc.gotAreaCode, "finance")
	}
	if fakeSvc.gotLimit != 40 || fakeSvc.gotOffset != 10 {
		t.Fatalf("limit/offset = %d/%d, want %d/%d", fakeSvc.gotLimit, fakeSvc.gotOffset, 40, 10)
	}
}

func TestInboxHandler_InvalidLimitTooLarge(t *testing.T) {
	fakeSvc := &fakeReadServiceInbox{}
	h := &Handler{readSvc: fakeSvc}
	mux := inboxTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/approval/inbox?limit=101", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if fakeSvc.called {
		t.Fatalf("service should not be called on invalid limit")
	}
}

func TestInboxHandler_PopulatesTitleQuorumAndTotal(t *testing.T) {
	fakeSvc := &fakeReadServiceInbox{
		views: []application.InboxView{{
			InstanceID:     "inst-1",
			DocumentID:     "doc-1",
			DocumentTitle:  "Doc One",
			AreaCode:       "finance",
			SubmittedBy:    "user-1",
			SubmittedAt:    time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
			StageLabel:     "Stage 1",
			QuorumProgress: "1/2",
		}},
		total: 7,
	}
	h := &Handler{readSvc: fakeSvc}
	mux := inboxTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/approval/inbox?limit=1", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rr.Code, rr.Body.String())
	}

	var resp contracts.InboxResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Total != 7 {
		t.Errorf("Total = %d, want 7", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(resp.Items))
	}
	got := resp.Items[0]
	if got.DocumentTitle != "Doc One" {
		t.Errorf("DocumentTitle = %q, want %q", got.DocumentTitle, "Doc One")
	}
	if got.QuorumProgress != "1/2" {
		t.Errorf("QuorumProgress = %q, want %q", got.QuorumProgress, "1/2")
	}
	if got.InstanceID != "inst-1" || got.DocumentID != "doc-1" {
		t.Errorf("ID mapping wrong: %+v", got)
	}
	if got.StageLabel != "Stage 1" || got.AreaCode != "finance" {
		t.Errorf("Stage/Area mapping wrong: %+v", got)
	}
	if got.SubmittedBy != "user-1" {
		t.Errorf("SubmittedBy = %q, want %q", got.SubmittedBy, "user-1")
	}
	if got.SubmittedAt != "2026-05-01T12:00:00Z" {
		t.Errorf("SubmittedAt = %q, want RFC3339 UTC", got.SubmittedAt)
	}
}
