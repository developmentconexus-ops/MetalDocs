//go:build integration
// +build integration

package repository_test

// SEC-06 (P1, wiki/reviews/grade-a-simplification-report-2026-07-01.md §SEC,
// T-003 in wiki/modules/documents/_artifacts/05-industry.md) claimed `documents`
// table writes (Create / UpdateName / UpdateStatus / MarkArchived / Unarchive)
// had only the tier-1 route->capability gate, with no in-tx tier-2
// authz.Require — asymmetric vs approval-instance writes.
//
// Investigation (2026-07-01) found this finding stale: every one of the five
// named write paths already carries an in-tx authz.Require(CapDocumentCreate|
// CapDocumentEdit, docArea) call (repository.go CreateDocumentTx, UpdateDocumentNameTx,
// UpdateDocumentStatus, MarkArchivedTx, Unarchive), added in commits 2f44a29be,
// 88ab8f8dc, 6cc95955c (2026-05-11..2026-06-11) — all AFTER the 2026-05-10
// 05-industry.md snapshot the finding cites. Register drift, not a live defect.
//
// These tests are the missing verification: they prove authz.Require's actual
// grant/deny behavior on each write path (deny with no grant -> ErrCapDenied,
// allow with an area-scoped grant -> no ErrCapDenied), closing the gap that the
// existing repository_create_integration_test.go left (it only satisfies the DB
// tripwire via testdb.SetCapsOnTx, which bypasses authz.Require's own
// grant-check logic rather than exercising it).

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

// grantAreaRole gives userID the given role (which must carry document.create /
// document.edit per db/reference-data/0001_product_reference_data.sql — 'author'
// carries both) scoped to areaCode via metaldocs.user_process_areas, mirroring
// the real grant path authz.Require reads (upa.role = rc.role, effective_to IS
// NULL). Uses the membership.manage tripwire cap, like
// get_active_instance_authz_integration_test.go's viewer-grant helper.
func grantAreaRole(t *testing.T, db *sql.DB, tenantID, userID, areaCode, role string) {
	t.Helper()
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.user_process_areas
			    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
			userID, tenantID, areaCode, role,
		)
		return err
	})
}

// newAreaScopedDocument seeds a tenant + taxonomy + a document actually carrying
// a non-empty process_area_code_snapshot (NewDocument leaves that column NULL,
// which would make every area-grade check below degenerate to the "" area — not
// a meaningful test of area-scoped grants). Built directly through
// repo.CreateDocumentTx exactly like repository_create_integration_test.go,
// using the system_admin bypass so document creation itself isn't gated on the
// authz behavior under test.
func newAreaScopedDocument(t *testing.T, db *sql.DB) (repo *repository.Repository, tenantID, areaCode, docID, ownerID string) {
	t.Helper()
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(owner.ID),
	)

	r := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
		Name:                    "SEC-06 Fixture Doc",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               owner.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "SEC06-" + tax.ProcessAreaCode,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// owner is system_admin: authz.Require bypasses tier-2 area check, so this
	// fixture-build does not itself depend on the behavior under test.
	if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, owner.ID); err != nil {
		t.Fatalf("seed tx identity: %v", err)
	}
	newDocID, _, _, err := r.CreateDocumentTx(ctx, tx, doc, "hash-sec06", "", nil)
	if err != nil {
		t.Fatalf("CreateDocumentTx fixture: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit fixture tx: %v", err)
	}

	return r, tnt.ID, tax.ProcessAreaCode, newDocID, owner.ID
}

// --- UpdateDocumentNameTx --------------------------------------------------

func TestUpdateDocumentNameTx_NoGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, _, docID, _ := newAreaScopedDocument(t, db)

	stranger := testdb.NewUser(t, db, testdb.WithTenant(tenantID))

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	err = repo.UpdateDocumentNameTx(ctx, tx, tenantID, stranger.ID, docID, "Renamed by stranger")
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("UpdateDocumentNameTx with no grant = %v, want authz.ErrCapDenied", err)
	}
	if denied.Capability != string(iamdomain.CapDocumentEdit) {
		t.Fatalf("ErrCapDenied.Capability = %q, want %q", denied.Capability, iamdomain.CapDocumentEdit)
	}
}

func TestUpdateDocumentNameTx_WithGrant_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, areaCode, docID, _ := newAreaScopedDocument(t, db)

	author := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	grantAreaRole(t, db, tenantID, author.ID, areaCode, "author")

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.UpdateDocumentNameTx(ctx, tx, tenantID, author.ID, docID, "Renamed by author"); err != nil {
		var denied authz.ErrCapDenied
		if errors.As(err, &denied) {
			t.Fatalf("UpdateDocumentNameTx with document.edit grant = %v, want no authz.ErrCapDenied", err)
		}
		t.Fatalf("UpdateDocumentNameTx: %v", err)
	}
}

// --- UpdateDocumentStatus ---------------------------------------------------

func TestUpdateDocumentStatus_NoGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, _, docID, _ := newAreaScopedDocument(t, db)

	stranger := testdb.NewUser(t, db, testdb.WithTenant(tenantID))

	err := repo.UpdateDocumentStatus(context.Background(), tenantID, stranger.ID, docID, domain.DocStatusDraft, domain.DocStatusDraft, false)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("UpdateDocumentStatus with no grant = %v, want authz.ErrCapDenied", err)
	}
	if denied.Capability != string(iamdomain.CapDocumentEdit) {
		t.Fatalf("ErrCapDenied.Capability = %q, want %q", denied.Capability, iamdomain.CapDocumentEdit)
	}
}

func TestUpdateDocumentStatus_WithGrant_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, areaCode, docID, _ := newAreaScopedDocument(t, db)

	author := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	grantAreaRole(t, db, tenantID, author.ID, areaCode, "author")

	// cur == next keeps this a status-preserving no-op WHERE match, isolating the
	// assertion to the authz gate rather than the lifecycle-transition CHECK.
	err := repo.UpdateDocumentStatus(context.Background(), tenantID, author.ID, docID, domain.DocStatusDraft, domain.DocStatusDraft, false)
	if err != nil {
		var denied authz.ErrCapDenied
		if errors.As(err, &denied) {
			t.Fatalf("UpdateDocumentStatus with document.edit grant = %v, want no authz.ErrCapDenied", err)
		}
		t.Fatalf("UpdateDocumentStatus: %v", err)
	}
}

// --- MarkArchivedTx ----------------------------------------------------------

func TestMarkArchivedTx_NoGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, _, docID, _ := newAreaScopedDocument(t, db)

	stranger := testdb.NewUser(t, db, testdb.WithTenant(tenantID))

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	err = repo.MarkArchivedTx(ctx, tx, tenantID, docID, stranger.ID)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("MarkArchivedTx with no grant = %v, want authz.ErrCapDenied", err)
	}
	if denied.Capability != string(iamdomain.CapDocumentEdit) {
		t.Fatalf("ErrCapDenied.Capability = %q, want %q", denied.Capability, iamdomain.CapDocumentEdit)
	}
}

func TestMarkArchivedTx_WithGrant_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, areaCode, docID, _ := newAreaScopedDocument(t, db)

	author := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	grantAreaRole(t, db, tenantID, author.ID, areaCode, "author")

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.MarkArchivedTx(ctx, tx, tenantID, docID, author.ID); err != nil {
		var denied authz.ErrCapDenied
		if errors.As(err, &denied) {
			t.Fatalf("MarkArchivedTx with document.edit grant = %v, want no authz.ErrCapDenied", err)
		}
		t.Fatalf("MarkArchivedTx: %v", err)
	}
}

// --- Unarchive ----------------------------------------------------------------
// Unarchive has no live caller (dead code as of 2026-07-01 — grep found zero
// production/HTTP callers) but SEC-06 names it explicitly, and its tier-2 check
// is real code that must stay correct if/when it is wired up.

func TestUnarchive_NoGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, areaCode, docID, ownerID := newAreaScopedDocument(t, db)

	// Archive first (as the system_admin owner) so Unarchive has a real
	// archived_at to clear; otherwise the not-found/no-op branch could mask the
	// authz assertion.
	ctx := context.Background()
	archiveTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin archive tx: %v", err)
	}
	if err := authz.SeedTxIdentity(ctx, archiveTx, tenantID, ownerID); err != nil {
		t.Fatalf("seed archive tx identity: %v", err)
	}
	if err := repo.MarkArchivedTx(ctx, archiveTx, tenantID, docID, ownerID); err != nil {
		t.Fatalf("MarkArchivedTx setup: %v", err)
	}
	if err := archiveTx.Commit(); err != nil {
		t.Fatalf("commit archive tx: %v", err)
	}
	_ = areaCode

	stranger := testdb.NewUser(t, db, testdb.WithTenant(tenantID))

	err = repo.Unarchive(ctx, tenantID, docID, stranger.ID)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("Unarchive with no grant = %v, want authz.ErrCapDenied", err)
	}
	if denied.Capability != string(iamdomain.CapDocumentEdit) {
		t.Fatalf("ErrCapDenied.Capability = %q, want %q", denied.Capability, iamdomain.CapDocumentEdit)
	}
}

func TestUnarchive_WithGrant_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	repo, tenantID, areaCode, docID, ownerID := newAreaScopedDocument(t, db)

	ctx := context.Background()
	archiveTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin archive tx: %v", err)
	}
	if err := authz.SeedTxIdentity(ctx, archiveTx, tenantID, ownerID); err != nil {
		t.Fatalf("seed archive tx identity: %v", err)
	}
	if err := repo.MarkArchivedTx(ctx, archiveTx, tenantID, docID, ownerID); err != nil {
		t.Fatalf("MarkArchivedTx setup: %v", err)
	}
	if err := archiveTx.Commit(); err != nil {
		t.Fatalf("commit archive tx: %v", err)
	}

	author := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	grantAreaRole(t, db, tenantID, author.ID, areaCode, "author")

	if err := repo.Unarchive(ctx, tenantID, docID, author.ID); err != nil {
		var denied authz.ErrCapDenied
		if errors.As(err, &denied) {
			t.Fatalf("Unarchive with document.edit grant = %v, want no authz.ErrCapDenied", err)
		}
		t.Fatalf("Unarchive: %v", err)
	}
}

// --- CreateDocumentTx ---------------------------------------------------------
// CreateDocumentTx is covered for authz.Require's real grant/deny behavior here;
// repository_create_integration_test.go's coverage uses testdb.SetCapsOnTx,
// which satisfies the DB tripwire directly and does not exercise
// authz.Require's own grant lookup.

func TestCreateDocumentTx_NoGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(owner.ID),
	)

	stranger := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID))

	r := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
		Name:                    "Stranger Create Attempt",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               stranger.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "SEC06-CREATE-DENY",
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, stranger.ID); err != nil {
		t.Fatalf("seed tx identity: %v", err)
	}

	_, _, _, err = r.CreateDocumentTx(ctx, tx, doc, "hash", "", nil)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("CreateDocumentTx with no grant = %v, want authz.ErrCapDenied", err)
	}
	if denied.Capability != string(iamdomain.CapDocumentCreate) {
		t.Fatalf("ErrCapDenied.Capability = %q, want %q", denied.Capability, iamdomain.CapDocumentCreate)
	}
}

func TestCreateDocumentTx_WithGrant_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(owner.ID),
	)

	author := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID))
	grantAreaRole(t, db, tnt.ID, author.ID, tax.ProcessAreaCode, "author")

	r := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
		Name:                    "Author Create",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               author.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "SEC06-CREATE-ALLOW",
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, author.ID); err != nil {
		t.Fatalf("seed tx identity: %v", err)
	}

	if _, _, _, err := r.CreateDocumentTx(ctx, tx, doc, "hash", "", nil); err != nil {
		var denied authz.ErrCapDenied
		if errors.As(err, &denied) {
			t.Fatalf("CreateDocumentTx with document.create+edit grant = %v, want no authz.ErrCapDenied", err)
		}
		t.Fatalf("CreateDocumentTx: %v", err)
	}
}
