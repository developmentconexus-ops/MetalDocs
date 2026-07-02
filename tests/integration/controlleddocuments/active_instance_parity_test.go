//go:build integration
// +build integration

package controlleddocuments_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	docrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/tests/integration/testdb"
)

// F2.2 / B2-B3-B4 parity gate (D6: parity-before-delete). Proves the ported
// GetActiveInstance (documents-owned ActiveInstanceReader port + 1:1 view map)
// returns the EXACT result the deleted inline SQL produced, across every state
// branch: active-only, published-only, under_review (+ derived in-progress
// approval instance), and none. The raw baseline below is the verbatim SQL that
// lived in GetActiveInstance at HEAD before the port replaced it.
//
// Relocated from internal/modules/controlleddocuments/infrastructure (TST-07):
// that package cannot import tests/integration/testdb because testdb/factory.go
// itself imports controlleddocuments/infrastructure (for NewCDFieldReaderPG),
// which is an import cycle under `-tags integration`. This test only needs the
// package's exported surface (NewPostgresControlledDocumentRepository), so it
// lives here as a black-box _test package alongside the other module-scoped
// testdb-driven integration suites.

// rawGetActiveInstance replays the pre-port inline reads (the documents FULL
// OUTER JOIN projection + the derived in_progress approval_instances lookup)
// directly against the test pool. This is the parity baseline.
func rawGetActiveInstance(t *testing.T, db *sql.DB, tenantID, controlledDocumentID string) *controlleddocumentsdomain.ActiveDocumentInstance {
	t.Helper()
	ctx := context.Background()
	var (
		docID          sql.NullString
		contentHash    sql.NullString
		revisionVer    sql.NullInt64
		approvalState  sql.NullString
		publishedDocID sql.NullString
	)
	err := db.QueryRowContext(ctx, `
SELECT active.id,
       COALESCE(active.content_hash_at_submit,
                (SELECT r.content_hash FROM document_revisions r
                  WHERE r.document_id = active.id
                  ORDER BY r.created_at DESC LIMIT 1)),
       active.revision_version,
       active.status,
       pub.id::text
  FROM (SELECT id, content_hash_at_submit, revision_version, status
          FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status IN ('draft','under_review','approved','rejected','scheduled')
         LIMIT 1) active
  FULL OUTER JOIN
       (SELECT id FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status = 'published'
         ORDER BY revision_number DESC
         LIMIT 1) pub ON TRUE`,
		tenantID, controlledDocumentID,
	).Scan(&docID, &contentHash, &revisionVer, &approvalState, &publishedDocID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		t.Fatalf("raw active projection: %v", err)
	}
	if !docID.Valid && !publishedDocID.Valid {
		return nil
	}

	inst := &controlleddocumentsdomain.ActiveDocumentInstance{}
	if docID.Valid {
		v := docID.String
		inst.DocumentID = &v
	}
	if contentHash.Valid {
		v := contentHash.String
		inst.ContentHash = &v
	}
	if revisionVer.Valid {
		v := int(revisionVer.Int64)
		inst.RevisionVersion = &v
	}
	if approvalState.Valid {
		v := approvalState.String
		inst.ApprovalState = &v
	}
	if publishedDocID.Valid {
		v := publishedDocID.String
		inst.PublishedDocumentID = &v
	}

	if inst.DocumentID != nil && inst.ApprovalState != nil && *inst.ApprovalState == "under_review" {
		var approvalInstanceID sql.NullString
		err := db.QueryRowContext(ctx, `
SELECT id::text
  FROM approval_instances
 WHERE document_id = $1::uuid
   AND tenant_id = $2::uuid
   AND status = 'in_progress'
 ORDER BY submitted_at DESC
 LIMIT 1`,
			*inst.DocumentID, tenantID,
		).Scan(&approvalInstanceID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("raw approval instance: %v", err)
		}
		if approvalInstanceID.Valid {
			v := approvalInstanceID.String
			inst.ApprovalInstanceID = &v
		}
	}
	return inst
}

func ptrStr(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func ptrInt(p *int) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *p)
}

func assertActiveInstanceParity(t *testing.T, label string, want, got *controlleddocumentsdomain.ActiveDocumentInstance) {
	t.Helper()
	if (want == nil) != (got == nil) {
		t.Fatalf("%s: nil mismatch: raw=%v port=%v", label, want == nil, got == nil)
	}
	if want == nil {
		return
	}
	if ptrStr(want.DocumentID) != ptrStr(got.DocumentID) {
		t.Fatalf("%s: DocumentID raw=%s port=%s", label, ptrStr(want.DocumentID), ptrStr(got.DocumentID))
	}
	if ptrStr(want.ContentHash) != ptrStr(got.ContentHash) {
		t.Fatalf("%s: ContentHash raw=%s port=%s", label, ptrStr(want.ContentHash), ptrStr(got.ContentHash))
	}
	if ptrInt(want.RevisionVersion) != ptrInt(got.RevisionVersion) {
		t.Fatalf("%s: RevisionVersion raw=%s port=%s", label, ptrInt(want.RevisionVersion), ptrInt(got.RevisionVersion))
	}
	if ptrStr(want.ApprovalState) != ptrStr(got.ApprovalState) {
		t.Fatalf("%s: ApprovalState raw=%s port=%s", label, ptrStr(want.ApprovalState), ptrStr(got.ApprovalState))
	}
	if ptrStr(want.PublishedDocumentID) != ptrStr(got.PublishedDocumentID) {
		t.Fatalf("%s: PublishedDocumentID raw=%s port=%s", label, ptrStr(want.PublishedDocumentID), ptrStr(got.PublishedDocumentID))
	}
	if ptrStr(want.ApprovalInstanceID) != ptrStr(got.ApprovalInstanceID) {
		t.Fatalf("%s: ApprovalInstanceID raw=%s port=%s", label, ptrStr(want.ApprovalInstanceID), ptrStr(got.ApprovalInstanceID))
	}
}

func TestActiveInstanceReader_ParityWithRawGetActiveInstance(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	repo := cdinfra.NewPostgresControlledDocumentRepository(db, docrepo.NewActiveInstanceReaderPG(db))

	t.Run("active_draft_only", func(t *testing.T) {
		doc := testdb.NewDocument(t, db, testdb.WithStatus("draft"))
		want := rawGetActiveInstance(t, db, doc.TenantID, doc.ControlledDocumentID)
		got, err := repo.GetActiveInstance(ctx, doc.TenantID, doc.ControlledDocumentID)
		if err != nil {
			t.Fatalf("port GetActiveInstance: %v", err)
		}
		assertActiveInstanceParity(t, "active_draft_only", want, got)
		if got == nil || got.DocumentID == nil || got.PublishedDocumentID != nil {
			t.Fatalf("active_draft_only: expected active doc, no published; got %+v", got)
		}
	})

	t.Run("published_only", func(t *testing.T) {
		doc := testdb.NewDocument(t, db, testdb.WithStatus("published"))
		want := rawGetActiveInstance(t, db, doc.TenantID, doc.ControlledDocumentID)
		got, err := repo.GetActiveInstance(ctx, doc.TenantID, doc.ControlledDocumentID)
		if err != nil {
			t.Fatalf("port GetActiveInstance: %v", err)
		}
		assertActiveInstanceParity(t, "published_only", want, got)
		if got == nil || got.PublishedDocumentID == nil || got.DocumentID != nil {
			t.Fatalf("published_only: expected published doc, no active; got %+v", got)
		}
	})

	t.Run("under_review_with_in_progress_approval", func(t *testing.T) {
		cd := testdb.NewControlledDoc(t, db)
		doc := testdb.NewDocument(t, db, testdb.WithControlledDoc(cd), testdb.WithStatus("under_review"))
		ai := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("in_progress"))
		want := rawGetActiveInstance(t, db, doc.TenantID, doc.ControlledDocumentID)
		got, err := repo.GetActiveInstance(ctx, doc.TenantID, doc.ControlledDocumentID)
		if err != nil {
			t.Fatalf("port GetActiveInstance: %v", err)
		}
		assertActiveInstanceParity(t, "under_review_with_in_progress_approval", want, got)
		if got == nil || got.ApprovalInstanceID == nil || *got.ApprovalInstanceID != ai.ID {
			t.Fatalf("under_review: expected approval instance %s; got %+v", ai.ID, got)
		}
	})

	t.Run("none", func(t *testing.T) {
		cd := testdb.NewControlledDoc(t, db)
		want := rawGetActiveInstance(t, db, cd.TenantID, cd.ID)
		got, err := repo.GetActiveInstance(ctx, cd.TenantID, cd.ID)
		if err != nil {
			t.Fatalf("port GetActiveInstance: %v", err)
		}
		assertActiveInstanceParity(t, "none", want, got)
		if want != nil || got != nil {
			t.Fatalf("none: expected nil,nil; raw=%+v port=%+v", want, got)
		}
	})
}
