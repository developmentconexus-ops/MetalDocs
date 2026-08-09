package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

func TestArchiveTemplate_Happy(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[template.ID] = template

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	got, err := svc.ArchiveTemplate(context.Background(), application.ArchiveCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-1",
		TemplateID:  template.ID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate returned error: %v", err)
	}
	if got.ArchivedAt == nil {
		t.Fatal("expected ArchivedAt to be set")
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditArchived {
		t.Fatalf("expected one %q audit event, got %v", domain.AuditArchived, repo.audit)
	}
	if repo.audit[0].VersionID != nil {
		t.Fatalf("expected nil version id in archive audit, got %v", repo.audit[0].VersionID)
	}
}

func TestArchiveTemplate_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	archivedAt := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	template := &domain.Template{
		ID:         "tpl-1",
		TenantID:   "tenant-a",
		ArchivedAt: &archivedAt,
	}
	repo.templates[template.ID] = template

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.ArchiveTemplate(context.Background(), application.ArchiveCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-1",
		TemplateID:  template.ID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate returned error: %v", err)
	}
	if got.ArchivedAt == nil || !got.ArchivedAt.Equal(archivedAt) {
		t.Fatalf("expected ArchivedAt to remain %v, got %v", archivedAt, got.ArchivedAt)
	}
	if len(repo.audit) != 0 {
		t.Fatalf("expected no audit events for idempotent archive, got %d", len(repo.audit))
	}
}

func TestArchiveTemplate_UsesTemplateArchiveCapability(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[template.ID] = template

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(db))
	mock.ExpectBegin()
	expectTemplateAuthz(mock, "user-1", "tenant-a", "template.archive")
	mock.ExpectCommit()

	_, err = svc.ArchiveTemplate(authzCtx("tenant-a", "user-1"), application.ArchiveCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-1",
		TemplateID:  template.ID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestPublishTemplateVersionReturnsConcurrentTransitionWhenVersionMoved covers
// the F-T3 remap for the PublishTemplateVersion path: a CAS loss on the version
// row must surface as ErrConcurrentTransition (→ 409) rather than the raw
// ErrStaleLockVersion (→ 412), since publish is a status transition.
func TestPublishTemplateVersionReturnsConcurrentTransitionWhenVersionMoved(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{
		ID:            "tpl-1",
		TenantID:      "tenant-a",
		LatestVersion: 2,
	}
	version := &domain.TemplateVersion{
		ID:             "ver-2",
		TemplateID:     template.ID,
		VersionNumber:  2,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "tenants/tenant-a/templates/tpl-1/versions/2.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
		LockVersion:    0,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	// Simulate concurrent writer ahead of us: the stored lock_version is past 0.
	repo.lockVersions[version.ID] = 1

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	_, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		TemplateID:    template.ID,
		VersionNumber: 2,
	})
	if !errors.Is(err, domain.ErrConcurrentTransition) {
		t.Fatalf("expected ErrConcurrentTransition (409-class), got %v", err)
	}
}

// TestPublishTemplateVersion_NoAutoNextDraft asserts M1·T2: PublishTemplateVersion
// must NOT spawn a next draft row or call CreateVersionTx. The published version
// is committed; no v(n+1) is created. Manual CreateNextVersion is the only path.
func TestPublishTemplateVersion_NoAutoNextDraft(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{
		ID:            "tpl-1",
		TenantID:      "tenant-a",
		LatestVersion: 1,
	}
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	res, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion returned error: %v", err)
	}
	if res.PublishedVersion.Status != domain.VersionStatusPublished {
		t.Fatalf("expected published status, got %q", res.PublishedVersion.Status)
	}
	// Exactly one version row: no draft v2 was created.
	if len(repo.versions) != 1 {
		t.Fatalf("expected 1 version row, got %d", len(repo.versions))
	}
	// Exactly one UpdateTemplateTx call (no second call for LatestVersion bump).
	if got := len(repo.UpdateTemplateTxCalls); got != 1 {
		t.Fatalf("expected exactly 1 UpdateTemplateTx call, got %d (calls: %v)", got, repo.UpdateTemplateTxCalls)
	}
	// LatestVersion must remain 1 (the published version) — no new draft allocated.
	if repo.UpdateTemplateTxCalls[0] != 1 {
		t.Fatalf("UpdateTemplateTx called with LatestVersion=%d, want 1", repo.UpdateTemplateTxCalls[0])
	}
}

// TestPublishTemplateVersion_ObsoletesPreviousAndAudits pins ARC-09: when
// PublishTemplateVersion republishes over an already-published version, the
// obsolete side-effect (ObsoletePreviousPublishedTx) must be accompanied by an
// AuditObsoleted event referencing the superseded version, committed in the
// same tx as the new AuditPublished event.
func TestPublishTemplateVersion_ObsoletesPreviousAndAudits(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{
		ID:                 "tpl-1",
		TenantID:           "tenant-a",
		LatestVersion:      2,
		PublishedVersionID: strPtr("ver-old"),
	}
	oldPublished := &domain.TemplateVersion{
		ID:            "ver-old",
		TemplateID:    template.ID,
		VersionNumber: 1,
		Status:        domain.VersionStatusPublished,
		AuthorID:      "author-0",
	}
	version := &domain.TemplateVersion{
		ID:             "ver-new",
		TemplateID:     template.ID,
		VersionNumber:  2,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/2.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[oldPublished.ID] = oldPublished
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	res, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		TemplateID:    template.ID,
		VersionNumber: 2,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion returned error: %v", err)
	}
	if res.PublishedVersion.Status != domain.VersionStatusPublished {
		t.Fatalf("expected published status, got %q", res.PublishedVersion.Status)
	}
	if oldPublished.ObsoletedAt == nil {
		t.Fatal("expected previously published version to be obsoleted")
	}
	if len(repo.audit) != 2 {
		t.Fatalf("expected 2 audit events (published + obsoleted), got %v", repo.audit)
	}
	var sawPublished, sawObsoleted bool
	for _, a := range repo.audit {
		switch a.Action {
		case domain.AuditPublished:
			sawPublished = true
			if a.VersionID == nil || *a.VersionID != version.ID {
				t.Fatalf("expected published audit versionID %q, got %v", version.ID, a.VersionID)
			}
		case domain.AuditObsoleted:
			sawObsoleted = true
			if a.VersionID == nil || *a.VersionID != oldPublished.ID {
				t.Fatalf("expected obsoleted audit versionID %q, got %v", oldPublished.ID, a.VersionID)
			}
			if got := a.Details["superseded_by_version_id"]; got != version.ID {
				t.Fatalf("expected superseded_by_version_id=%q, got %v", version.ID, got)
			}
		default:
			t.Fatalf("unexpected audit action %q", a.Action)
		}
	}
	if !sawPublished || !sawObsoleted {
		t.Fatalf("expected both published and obsoleted audit events, got %v", repo.audit)
	}
}

// TestPublishTemplateVersion_FirstPublishNoObsoletedAudit pins the negative
// case: a template's FIRST publish (no prior PublishedVersionID) must NOT
// emit an AuditObsoleted event — there is nothing to obsolete.
func TestPublishTemplateVersion_FirstPublishNoObsoletedAudit(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", LatestVersion: 1}
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	_, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion returned error: %v", err)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditPublished {
		t.Fatalf("expected exactly 1 published audit event (no obsoleted), got %v", repo.audit)
	}
}

// TestPublishTemplateVersion_AuditSchemaKeyIsServerDerived pins unit 3.1a S2b:
// the AuditPublished event's schema_key detail is always the server-derived
// key (application.TemplateVersionSchemaKey), never a caller-supplied value —
// client-supplied audit metadata is an integrity smell (no-fallback
// principle: audit truth is derived truth).
func TestPublishTemplateVersion_AuditSchemaKeyIsServerDerived(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", LatestVersion: 1}
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     template.ID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	_, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion returned error: %v", err)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditPublished {
		t.Fatalf("expected exactly 1 published audit event, got %v", repo.audit)
	}
	want := application.TemplateVersionSchemaKey("tenant-a", template.ID, version.VersionNumber)
	if got := repo.audit[0].Details["schema_key"]; got != want {
		t.Fatalf("expected derived schema_key %q, got %v", want, got)
	}
}

// TestPublishTemplateVersion_FromApproved_KernelCompletion pins ADR 0082
// slice S2: PublishTemplateVersion must also accept approved -> published,
// completing a version the kernel signoff flow already brought to approved
// (the kernel flow ends at approved; legacy Approve, which used to own
// approved->published, is retired in a later slice). Before S2 this hard-
// required draft and rejected an approved version with
// ErrInvalidStateTransition.
func TestPublishTemplateVersion_FromApproved_KernelCompletion(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{
		ID:            "tpl-1",
		TenantID:      "tenant-a",
		LatestVersion: 1,
	}
	// Kernel signoff completion stamped the TRUE approval time; publish must
	// preserve it, not clobber it with the publish timestamp (reviewer
	// finding, unit 3.1a S2).
	kernelApprovedAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  1,
		Status:         domain.VersionStatusApproved,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
		ApprovedAt:     &kernelApprovedAt,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	res, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "publisher-1", // != author-1
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion returned error: %v", err)
	}
	if res.PublishedVersion.Status != domain.VersionStatusPublished {
		t.Fatalf("expected published status, got %q", res.PublishedVersion.Status)
	}
	if res.PublishedVersion.ApprovedAt == nil || !res.PublishedVersion.ApprovedAt.Equal(kernelApprovedAt) {
		t.Fatalf("expected kernel-stamped ApprovedAt %v preserved, got %v", kernelApprovedAt, res.PublishedVersion.ApprovedAt)
	}
	if res.PublishedVersion.PublishedAt == nil || res.PublishedVersion.PublishedAt.Equal(kernelApprovedAt) {
		t.Fatalf("expected PublishedAt stamped at publish time, distinct from ApprovedAt; got %v", res.PublishedVersion.PublishedAt)
	}
}

// TestPublishTemplateVersion_NoRoles_CapabilityAlone pins ADR 0022: an actor
// holding only the template.publish capability, with no roles in context at
// all, must be able to publish a draft — authz is capability-based, never
// role-based. This regresses if a role check is ever reintroduced on this
// path.
func TestPublishTemplateVersion_NoRoles_CapabilityAlone(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", LatestVersion: 1}
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     template.ID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	res, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "publisher-1", // holds no roles; capability is the sole gate
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion returned error: %v", err)
	}
	if res.PublishedVersion.Status != domain.VersionStatusPublished {
		t.Fatalf("expected published status, got %q", res.PublishedVersion.Status)
	}
}

// TestPublishTemplateVersion_SegregationViolation_SelfPublish pins the
// surviving identity-SoD gate: an actor cannot publish a version they
// authored, regardless of role/capability. This is unaffected by the S2
// role-gate removal — CheckSegregation is identity-based (user IDs), not
// role-based, and runs before the (now-deleted) role-binding check.
func TestPublishTemplateVersion_SegregationViolation_SelfPublish(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", LatestVersion: 1}
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     template.ID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "author-1", // == AuthorID: self-publish
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if !errors.Is(err, domain.ErrISOSegregationViolation) {
		t.Fatalf("expected ErrISOSegregationViolation, got %v", err)
	}
}
