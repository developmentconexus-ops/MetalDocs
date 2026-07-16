//go:build integration
// +build integration

// Integration coverage for the QA-1 F13 pdf_status derivation
// (internal/modules/documents/application/view_service.go GetViewURL,
// ~lines 82-97): 'ready' iff documents.final_pdf_s3_key is present (checked
// FIRST), else consults the PDFOutboxStateReader port — the production
// implementation being internal/modules/render/fanout's
// PDFPipelineStateReader, which reports "failed" iff metaldocs.outbox_events
// carries a dead-lettered row for this document (docx_materialize or
// docgen_v2_pdf, tenant-scoped via payload->>'tenant_id'), else "" (pending).
//
// This drives the REAL production read path end-to-end: application.ViewService
// wired with the real db.TxRunner and the real fanout.PDFPipelineStateReader
// against a real testdb, only the S3 presigner faked (no object storage in
// this suite, mirroring view_service_test.go's fakeViewPresigner). ViewService
// runs GetViewURL inside a Do (read-write) tx and calls authz.Require for
// document.view; the actor is seeded as tenant system_admin so the bypass
// branch (ADR 0022 F8) is taken, matching the DB-integration idiom used by
// tripwire_documents_test.go and concurrent_revision_test.go (testdb factory,
// real capability seeding, no ad-hoc bootstrapping).
package documents_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/render/fanout"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// fakePDFViewPresigner is a trivial local fake mirroring
// internal/modules/documents/application/view_service_test.go's
// fakeViewPresigner — no real object storage is exercised by this suite,
// only the pdf_status derivation.
type fakePDFViewPresigner struct {
	url string
	err error
}

func (f *fakePDFViewPresigner) AssertedPresignGet(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.url, nil
}

// pdfPipelineStatusCtx builds the ctx carrying tenant+actor identity the real
// TxRunner chokepoint (internal/platform/db/runner.go seedTxIdentityFromContext)
// auto-seeds on every Do — the same shape view_service_test.go's viewAuthzCtx
// builds for the sqlmock-driven unit tests.
func pdfPipelineStatusCtx(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	return platformtenant.WithActorID(ctx, actorID)
}

// TestPDFPipelineStatus_ViewService_RealOutboxReader drives
// application.ViewService.GetViewURL wired with the REAL
// fanout.PDFPipelineStateReader against a real testdb, proving the three
// pdf_status outcomes the F13 fix depends on: pending (no signal yet), failed
// (a dead-lettered outbox row for this document), and ready short-circuiting
// ahead of a stale dead-lettered row once final_pdf_s3_key is set.
func TestPDFPipelineStatus_ViewService_RealOutboxReader(t *testing.T) {
	db, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	outbox := fanout.NewPDFPipelineStateReader(db)

	t.Run("pending_when_no_pdf_key_and_no_dead_letter", func(t *testing.T) {
		doc := testdb.NewDocument(t, db,
			testdb.WithTenant(tnt.ID),
			testdb.WithOwner(actor.ID),
			testdb.WithStatus("approved"),
		)

		presigner := &fakePDFViewPresigner{}
		svc := application.NewViewService(platformdb.NewTxRunner(db), presigner, outbox)

		result, err := svc.GetViewURL(pdfPipelineStatusCtx(tnt.ID, actor.ID), tnt.ID, actor.ID, doc.ID)
		if err != nil {
			t.Fatalf("GetViewURL: %v", err)
		}
		if result.PDFStatus != "pending" {
			t.Fatalf("PDFStatus = %q, want %q", result.PDFStatus, "pending")
		}
		if result.SignedURL != "" {
			t.Fatalf("SignedURL = %q, want empty (no pdf key yet)", result.SignedURL)
		}
	})

	t.Run("failed_when_dead_lettered_outbox_row_exists", func(t *testing.T) {
		doc := testdb.NewDocument(t, db,
			testdb.WithTenant(tnt.ID),
			testdb.WithOwner(actor.ID),
			testdb.WithStatus("approved"),
		)

		insertDeadLetteredOutboxEvent(t, db, doc.ID, tnt.ID, string(messaging.EventTypeMaterializeFanout))

		presigner := &fakePDFViewPresigner{}
		svc := application.NewViewService(platformdb.NewTxRunner(db), presigner, outbox)

		result, err := svc.GetViewURL(pdfPipelineStatusCtx(tnt.ID, actor.ID), tnt.ID, actor.ID, doc.ID)
		if err != nil {
			t.Fatalf("GetViewURL: %v", err)
		}
		if result.PDFStatus != "failed" {
			t.Fatalf("PDFStatus = %q, want %q", result.PDFStatus, "failed")
		}
		if result.SignedURL != "" {
			t.Fatalf("SignedURL = %q, want empty (no pdf key present)", result.SignedURL)
		}
	})

	t.Run("ready_wins_over_stale_dead_letter_once_pdf_key_is_set", func(t *testing.T) {
		doc := testdb.NewDocument(t, db,
			testdb.WithTenant(tnt.ID),
			testdb.WithOwner(actor.ID),
			testdb.WithStatus("approved"),
		)

		// Same dead-lettered row shape as the "failed" case above — proves ready
		// wins even with a live/stale dead-lettered outbox row still present,
		// because view_service.go checks final_pdf_s3_key FIRST.
		insertDeadLetteredOutboxEvent(t, db, doc.ID, tnt.ID, string(messaging.EventTypeMaterializeFanout))

		pdfKey := "tenants/" + tnt.ID + "/docs/" + doc.ID + "/final.pdf"
		testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(context.Background(),
				`SELECT set_config('metaldocs.tenant_id', $1, true)`, tnt.ID,
			); err != nil {
				return err
			}
			_, err := tx.ExecContext(context.Background(),
				`UPDATE public.documents SET final_pdf_s3_key = $1, updated_at = now() WHERE id = $2::uuid`,
				pdfKey, doc.ID,
			)
			return err
		})

		presigner := &fakePDFViewPresigner{url: "https://signed.example/" + doc.ID + ".pdf"}
		svc := application.NewViewService(platformdb.NewTxRunner(db), presigner, outbox)

		result, err := svc.GetViewURL(pdfPipelineStatusCtx(tnt.ID, actor.ID), tnt.ID, actor.ID, doc.ID)
		if err != nil {
			t.Fatalf("GetViewURL: %v", err)
		}
		if result.PDFStatus != "ready" {
			t.Fatalf("PDFStatus = %q, want %q (final_pdf_s3_key must short-circuit the stale dead-lettered row)", result.PDFStatus, "ready")
		}
		if result.SignedURL != presigner.url {
			t.Fatalf("SignedURL = %q, want %q", result.SignedURL, presigner.url)
		}
	})
}

// insertDeadLetteredOutboxEvent inserts a metaldocs.outbox_events row shaped
// exactly as fanout.PDFPipelineStateReader.ReadState requires to report
// "failed": aggregate_id=docID, event_type=eventType (docx_materialize or
// docgen_v2_pdf), payload->>'tenant_id'=tenantID, dead_lettered_at set. No
// trg_require_cap_asserted tripwire guards metaldocs.outbox_events, so this is
// a plain insert (no capability/tenant GUC seeding needed) — matching how the
// real outbox producers/consumers write this table.
func insertDeadLetteredOutboxEvent(t *testing.T, db *sql.DB, docID, tenantID, eventType string) {
	t.Helper()
	eventID := uuid.NewString()
	idempotencyKey := "idem-" + eventID
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO metaldocs.outbox_events
		   (event_id, event_type, aggregate_type, aggregate_id, occurred_at, version,
		    idempotency_key, producer, trace_id, payload,
		    attempt_count, last_error, last_attempt_at, dead_lettered_at)
		 VALUES ($1, $2, 'document', $3, now(), 1,
		         $4, 'test-seed', 'trace-'||$1, $5::jsonb,
		         5, 'ErrCapabilityNotAsserted: document.edit not asserted', now(), now())`,
		eventID, eventType, docID, idempotencyKey,
		`{"tenant_id":"`+tenantID+`"}`,
	)
	if err != nil {
		t.Fatalf("insertDeadLetteredOutboxEvent: %v", err)
	}
}
