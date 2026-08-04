//go:build integration

package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	controlleddocumentsapi "metaldocs/internal/modules/controlleddocuments/api"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// seedPreviewActor provisions a tenant, taxonomy, and actor user. When role != ""
// it also inserts a user_process_areas grant of that role in the seeded area
// (effective_from=now, NULL effective_to). The role is wired through the curated
// role_capabilities mapping — "author" carries controlled_documents.create — so
// the capability is granted via real fixture seeding, never a forced DB grant.
// Returns (tenantID, actorUserID, profileCode, areaCode).
func seedPreviewActor(t *testing.T, db *sql.DB, role string) (string, string, string, string) {
	t.Helper()
	ten := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID))
	user := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))

	if role != "" {
		testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`INSERT INTO metaldocs.user_process_areas
				    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
				 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
				user.ID, ten.ID, tax.ProcessAreaCode, role,
			)
			return err
		})
	}
	return ten.ID, user.ID, tax.ProfileCode, tax.ProcessAreaCode
}

func previewRequest(t *testing.T, tenantID, actorUserID string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/preview-code", nil)
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	// TxRunner.seedTxIdentityFromContext reads the actor via
	// platformtenant.ActorFromContext — seed both the tenant and actor GUC
	// sources so the SUT-owned tx has metaldocs.actor_id set (else P0001 → 500).
	ctx = tenant.WithActorID(ctx, actorUserID)
	ctx = iamdomain.WithAuthContext(ctx, actorUserID, []iamdomain.Role{"author"})
	return req.WithContext(ctx)
}

// TestPreviewCode_NonMember_Returns403 is the regression guard for the tier-2
// 500→403 bug: an authenticated actor WITHOUT controlled_documents.create in the
// target area must get 403 FORBIDDEN_CAPABILITY problem+json (RFC 9457), driven
// through the real authz.Require denial — not a synthesized sentinel.
func TestPreviewCode_NonMember_Returns403(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, profile, area := seedPreviewActor(t, db, "") // no grant

	handler := newTestHandler(db)
	rec := httptest.NewRecorder()
	handler.PreviewControlledDocumentCode(rec, previewRequest(t, tenantID, actorID),
		controlleddocumentsapi.PreviewControlledDocumentCodeParams{ProfileCode: profile, AreaCode: area})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q, want application/problem+json", ct)
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != problem.CodeForbiddenCapability.String() {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeForbiddenCapability)
	}
}

// TestPreviewCode_Member_Returns200 is the positive half: an actor WITH the
// capability in the target area gets 200 and a formatted code.
func TestPreviewCode_Member_Returns200(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, profile, area := seedPreviewActor(t, db, "author") // grant create

	handler := newTestHandler(db)
	rec := httptest.NewRecorder()
	handler.PreviewControlledDocumentCode(rec, previewRequest(t, tenantID, actorID),
		controlleddocumentsapi.PreviewControlledDocumentCodeParams{ProfileCode: profile, AreaCode: area})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var resp controlleddocumentsapi.PreviewCodeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if resp.Code == "" {
		t.Fatalf("expected a formatted preview code, got empty; body=%s", rec.Body.String())
	}
}
