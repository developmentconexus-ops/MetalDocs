package approvalhttp

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// supersedeIntRows returns a single int64 row.
type supersedeIntRows struct {
	value int64
	done  bool
}

func (r *supersedeIntRows) Columns() []string { return []string{"revision_version"} }
func (r *supersedeIntRows) Close() error      { return nil }
func (r *supersedeIntRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

type supersedeTestStmt struct{ value int64 }

func (s *supersedeTestStmt) Close() error                                 { return nil }
func (s *supersedeTestStmt) NumInput() int                                { return -1 }
func (s *supersedeTestStmt) Exec(_ []driver.Value) (driver.Result, error) { return nil, nil }
func (s *supersedeTestStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return &supersedeIntRows{value: s.value}, nil
}

type supersedeTestConn struct{ revVersion int64 }

func (c *supersedeTestConn) Prepare(_ string) (driver.Stmt, error) {
	return &supersedeTestStmt{value: c.revVersion}, nil
}
func (c *supersedeTestConn) Close() error              { return nil }
func (c *supersedeTestConn) Begin() (driver.Tx, error) { return c, nil }
func (c *supersedeTestConn) Commit() error             { return nil }
func (c *supersedeTestConn) Rollback() error           { return nil }

type supersedeTestDriver struct{ conn *supersedeTestConn }

func (d *supersedeTestDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

var supersedeDBCounter int

func newSupersedeTestDB(t *testing.T, revVersion int64) *sql.DB {
	t.Helper()
	supersedeDBCounter++
	name := fmt.Sprintf("supersede_test_%d", supersedeDBCounter)
	sql.Register(name, &supersedeTestDriver{conn: &supersedeTestConn{revVersion: revVersion}})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open supersede test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func supersedeTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/documents/{id}/supersede", h.SupersedeHandler)
	return mux
}

func TestSupersedeHandler(t *testing.T) {
	origPublishSuperseding := publishSuperseding
	t.Cleanup(func() {
		publishSuperseding = origPublishSuperseding
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
			svcErr:     authz.ErrCapDenied{Capability: "doc.supersede", AreaCode: "tenant-1", ActorID: "actor-1"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "illegal transition",
			svcErr:     errors.New("approval: illegal transition"),
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
			var gotReq application.SupersedeRequest
			publishSuperseding = func(_ *Handler, _ context.Context, _ *sql.DB, req application.SupersedeRequest) (application.SupersedeResult, error) {
				gotReq = req
				if tt.svcErr != nil {
					return application.SupersedeResult{}, tt.svcErr
				}
				return application.SupersedeResult{
					NewDocumentStatus:   "published",
					PriorDocumentStatus: "superseded",
				}, nil
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-2/supersede", strings.NewReader(`{"superseded_document_id":"11111111-1111-1111-1111-111111111111"}`))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "idem-1")
			req.Header.Set("If-Match", "\"v5\"")

			rr := httptest.NewRecorder()
			supersedeTestMux(&Handler{db: newSupersedeTestDB(t, 5)}).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if gotReq.TenantID != "tenant-1" || gotReq.NewDocumentID != "doc-2" || gotReq.PriorDocumentID != "11111111-1111-1111-1111-111111111111" || gotReq.SupersededBy != "actor-1" {
				t.Fatalf("unexpected service request: %+v", gotReq)
			}
			if gotReq.NewRevisionVersion != 5 || gotReq.PriorRevisionVersion != 0 {
				t.Fatalf("unexpected revision mapping: %+v", gotReq)
			}

			if tt.wantStatus == http.StatusOK {
				var out contracts.SupersedeResponse
				if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if out.DocumentID != "doc-2" || out.SupersededID != "11111111-1111-1111-1111-111111111111" {
					t.Fatalf("unexpected response: %+v", out)
				}
			}
		})
	}
}
