//go:build integration
// +build integration

package application

// TestPublishApproved_DoesNotAutoCreateNextVersion guards the invariant that
// PublishApproved transitions the document status to "published" and bumps
// revision_version, but inserts NO new documents row. This mirrors the template
// guard TestPublishTemplateVersion_NoAutoNextDraft (ADR 0052) on the document
// side. The test MUST be a real-DB integration test because PublishApproved
// issues raw tx.ExecContext — only a live database can observe the row count.

import (
	"context"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

func TestPublishApproved_DoesNotAutoCreateNextVersion(t *testing.T) {
	ctx := context.Background()

	// -----------------------------------------------------------------------
	// 1. Open isolated per-test database (ADR 0034 testdb harness).
	// -----------------------------------------------------------------------
	database, dbName := testdb.Open(t)

	// Set search_path and widen the pool (H-PRE-1: off-tx reads must not deadlock).
	if _, err := database.ExecContext(ctx,
		`ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`,
	); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	database.SetMaxIdleConns(0)
	database.SetMaxIdleConns(4)
	database.SetMaxOpenConns(4)

	// -----------------------------------------------------------------------
	// 2. Seed prerequisites.
	//    NewDocument auto-wires: tenant → taxonomy → controlled_doc → document.
	//    NewApprovalInstance auto-wires: route (via profileForCD).
	// -----------------------------------------------------------------------
	tnt := testdb.NewTenant(t, database)
	user := testdb.NewUser(t, database, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(user.ID),
		testdb.WithStatus("approved"),
	)

	inst := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithStatus("approved"),
	)

	// -----------------------------------------------------------------------
	// 3. Capture pre-publish document count in the lineage.
	// -----------------------------------------------------------------------
	var prePubCount int
	if err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.documents WHERE controlled_document_id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ControlledDocumentID, tnt.ID,
	).Scan(&prePubCount); err != nil {
		t.Fatalf("pre-publish count query: %v", err)
	}
	if prePubCount != 1 {
		t.Fatalf("expected 1 document in lineage before publish, got %d", prePubCount)
	}

	// -----------------------------------------------------------------------
	// 4. Build the services and call PublishApproved.
	// -----------------------------------------------------------------------
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	emitter := NewSQLEmitter()
	clock := RealClock{}
	cdRead := controlleddocumentsdomain.NoopCDFieldReader{}

	svcs := NewServices(repo, emitter, clock, cdRead)
	runner := db.NewTxRunner(database)

	res, err := svcs.Publish.PublishApproved(ctx, runner, PublishRequest{
		TenantID:   tnt.ID,
		InstanceID: inst.ID,
		PublishedBy: user.ID,
	})
	if err != nil {
		t.Fatalf("PublishApproved: %v", err)
	}

	// -----------------------------------------------------------------------
	// 5. Assert status transition + revision_version bump.
	// -----------------------------------------------------------------------
	if res.DocumentID == "" {
		t.Fatal("PublishResult.DocumentID must not be empty")
	}
	if res.NewStatus != "published" {
		t.Fatalf("PublishResult.NewStatus = %q; want \"published\"", res.NewStatus)
	}

	var gotStatus string
	var gotRevVersion int
	if err := database.QueryRowContext(ctx,
		`SELECT status, revision_version FROM public.documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ID, tnt.ID,
	).Scan(&gotStatus, &gotRevVersion); err != nil {
		t.Fatalf("query document row after publish: %v", err)
	}
	if gotStatus != "published" {
		t.Fatalf("documents.status = %q after PublishApproved; want \"published\"", gotStatus)
	}
	expectedRevVersion := doc.RevisionVersion + 1
	if gotRevVersion != expectedRevVersion {
		t.Fatalf("documents.revision_version = %d after PublishApproved; want %d (pre=%d + 1)",
			gotRevVersion, expectedRevVersion, doc.RevisionVersion)
	}

	// -----------------------------------------------------------------------
	// 6. Core invariant: NO new documents row was inserted into the lineage.
	// -----------------------------------------------------------------------
	var postPubCount int
	if err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.documents WHERE controlled_document_id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ControlledDocumentID, tnt.ID,
	).Scan(&postPubCount); err != nil {
		t.Fatalf("post-publish count query: %v", err)
	}
	if postPubCount != prePubCount {
		t.Fatalf("PublishApproved inserted a new documents row: lineage count went %d → %d; expected no change",
			prePubCount, postPubCount)
	}
}
