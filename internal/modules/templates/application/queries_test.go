package application_test

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

func TestGetTemplate_Happy(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.GetTemplate(context.Background(), "tenant-a", tpl.ID)
	if err != nil {
		t.Fatalf("GetTemplate returned error: %v", err)
	}
	if got.ID != tpl.ID || got.TenantID != tpl.TenantID {
		t.Fatalf("expected template %+v, got %+v", tpl, got)
	}
}

func TestGetTemplate_CrossTenant(t *testing.T) {
	repo := newFakeRepo()
	repo.ignoreTenantOnGetTemplate = true
	repo.templates["tpl-1"] = &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.GetTemplate(context.Background(), "tenant-b", "tpl-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetVersion_Happy(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	ver := &domain.TemplateVersion{ID: "ver-2", TemplateID: tpl.ID, VersionNumber: 2}
	repo.templates[tpl.ID] = tpl
	repo.versions[ver.ID] = ver

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.GetVersion(context.Background(), "tenant-a", tpl.ID, 2)
	if err != nil {
		t.Fatalf("GetVersion returned error: %v", err)
	}
	if got.ID != ver.ID || got.TemplateID != ver.TemplateID || got.VersionNumber != ver.VersionNumber {
		t.Fatalf("expected version %+v, got %+v", ver, got)
	}
}

func TestGetVersion_CrossTenant(t *testing.T) {
	repo := newFakeRepo()
	repo.ignoreTenantOnGetTemplate = true
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[tpl.ID] = tpl
	repo.versions["ver-1"] = &domain.TemplateVersion{ID: "ver-1", TemplateID: tpl.ID, VersionNumber: 1}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.GetVersion(context.Background(), "tenant-b", tpl.ID, 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListTemplates_Happy(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates["tpl-2"] = &domain.Template{ID: "tpl-2", TenantID: "tenant-b"}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.ListTemplates(context.Background(), application.ListFilter{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("ListTemplates returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "tpl-1" {
		t.Fatalf("expected one template tpl-1, got %+v", got)
	}
}

func TestListTemplates_SelectionFilterPassThrough(t *testing.T) {
	repo := newFakeRepo()
	publishedVerID := "ver-1"
	repo.templates["tpl-1"] = &domain.Template{ID: "tpl-1", TenantID: "tenant-a", PublishedVersionID: &publishedVerID}
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	status := domain.VersionStatusDraft
	docType := "CONTRACT"
	filter := application.ListFilter{
		TenantID:      "tenant-a",
		DocTypeCode:   &docType,
		Status:        &status,
		PublishedOnly: true,
		Limit:         10,
		Offset:        2,
	}

	_, err := svc.ListTemplates(context.Background(), filter)
	if err != nil {
		t.Fatalf("ListTemplates returned error: %v", err)
	}
	if repo.receivedFilter.TenantID != filter.TenantID {
		t.Fatalf("expected TenantID %q, got %q", filter.TenantID, repo.receivedFilter.TenantID)
	}
	if repo.receivedFilter.DocTypeCode == nil || *repo.receivedFilter.DocTypeCode != docType {
		t.Fatalf("expected DocTypeCode %q, got %v", docType, repo.receivedFilter.DocTypeCode)
	}
	if repo.receivedFilter.PublishedOnly != true {
		t.Fatalf("expected PublishedOnly true, got %v", repo.receivedFilter.PublishedOnly)
	}
	if repo.receivedFilter.Limit != filter.Limit || repo.receivedFilter.Offset != filter.Offset {
		t.Fatalf("expected limit/offset %d/%d, got %d/%d", filter.Limit, filter.Offset, repo.receivedFilter.Limit, repo.receivedFilter.Offset)
	}
}

// TestListTemplates_PublishedOnlyFiltersUnpublished proves the published=true
// contract at the application layer: PublishedOnly=true must exclude
// templates with no published version (PublishedVersionID == nil) and keep
// only those that have one. PublishedOnly=false (default) must return every
// status — the management view.
func TestListTemplates_PublishedOnlyFiltersUnpublished(t *testing.T) {
	repo := newFakeRepo()
	publishedVerID := "ver-1"
	repo.templates["tpl-published"] = &domain.Template{
		ID: "tpl-published", TenantID: "tenant-a", PublishedVersionID: &publishedVerID,
	}
	repo.templates["tpl-draft"] = &domain.Template{
		ID: "tpl-draft", TenantID: "tenant-a", PublishedVersionID: nil,
	}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.ListTemplates(context.Background(), application.ListFilter{
		TenantID:      "tenant-a",
		PublishedOnly: true,
	})
	if err != nil {
		t.Fatalf("ListTemplates returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "tpl-published" {
		t.Fatalf("expected only tpl-published, got %+v", got)
	}

	gotAll, err := svc.ListTemplates(context.Background(), application.ListFilter{
		TenantID: "tenant-a",
	})
	if err != nil {
		t.Fatalf("ListTemplates (management view) returned error: %v", err)
	}
	if len(gotAll) != 2 {
		t.Fatalf("expected both templates in management view, got %+v", gotAll)
	}
}

func TestListAudit_Happy(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[tpl.ID] = tpl
	repo.audit = append(repo.audit,
		&domain.AuditEvent{TenantID: "tenant-a", TemplateID: tpl.ID, Action: domain.AuditCreated},
		&domain.AuditEvent{TenantID: "tenant-a", TemplateID: "tpl-other", Action: domain.AuditCreated},
	)

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.ListAudit(context.Background(), "tenant-a", tpl.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListAudit returned error: %v", err)
	}
	if len(got) != 1 || got[0].TemplateID != tpl.ID {
		t.Fatalf("expected one audit event for %q, got %+v", tpl.ID, got)
	}
}

// TestGetDocxURL_EmptyStorageKey_ErrUploadMissing is DELETED (ADR 0088): it
// pinned the `DocxStorageKey == ""` branch, which was dead code
// (docx_storage_key is NOT NULL and every writer sets it from templateDocxKey)
// and is now doubly unreachable, since creation materializes the object at
// that very key. TestGetDocxURL_ObjectMissing_ErrUploadMissing below covers the
// sentinel's real, remaining meaning: the object itself is gone.

func TestGetDocxURL_ObjectMissing_ErrUploadMissing(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	ver := &domain.TemplateVersion{ID: "ver-1", TemplateID: tpl.ID, VersionNumber: 1, DocxStorageKey: "tenant-a/tpl-1/v1.docx"}
	repo.templates[tpl.ID] = tpl
	repo.versions[ver.ID] = ver

	notExists := false
	presign := &fakePresigner{existsResult: &notExists}
	svc := application.New(repo, presign, fakeClock{}, &fakeUUID{})

	url, err := svc.GetDocxURL(context.Background(), application.GetDocxURLCmd{
		TenantID: "tenant-a", TemplateID: tpl.ID, VersionNumber: 1,
	})
	if !errors.Is(err, domain.ErrUploadMissing) {
		t.Fatalf("expected ErrUploadMissing, got %v", err)
	}
	if url != "" {
		t.Fatalf("expected empty URL, got %q", url)
	}
}

func TestGetDocxURL_ObjectPresent_ReturnsURL(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	key := "tenant-a/tpl-1/v1.docx"
	ver := &domain.TemplateVersion{ID: "ver-1", TemplateID: tpl.ID, VersionNumber: 1, DocxStorageKey: key}
	repo.templates[tpl.ID] = tpl
	repo.versions[ver.ID] = ver

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	url, err := svc.GetDocxURL(context.Background(), application.GetDocxURLCmd{
		TenantID: "tenant-a", TemplateID: tpl.ID, VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("GetDocxURL returned error: %v", err)
	}
	if want := "https://example/get/" + key; url != want {
		t.Fatalf("expected URL %q, got %q", want, url)
	}
}

func TestGetDocxURL_ExistsError_FailsClosed(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	ver := &domain.TemplateVersion{ID: "ver-1", TemplateID: tpl.ID, VersionNumber: 1, DocxStorageKey: "tenant-a/tpl-1/v1.docx"}
	repo.templates[tpl.ID] = tpl
	repo.versions[ver.ID] = ver

	storeErr := errors.New("store unreachable")
	presign := &fakePresigner{existsErr: storeErr}
	svc := application.New(repo, presign, fakeClock{}, &fakeUUID{})

	url, err := svc.GetDocxURL(context.Background(), application.GetDocxURLCmd{
		TenantID: "tenant-a", TemplateID: tpl.ID, VersionNumber: 1,
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected store error to propagate, got %v", err)
	}
	if url != "" {
		t.Fatalf("expected empty URL, got %q", url)
	}
	if len(presign.getCalls) != 0 {
		t.Fatalf("expected PresignGet not to be called, got calls %v", presign.getCalls)
	}
}

func TestGetDocxURL_CrossTenant(t *testing.T) {
	repo := newFakeRepo()
	repo.ignoreTenantOnGetTemplate = true
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[tpl.ID] = tpl
	repo.versions["ver-1"] = &domain.TemplateVersion{ID: "ver-1", TemplateID: tpl.ID, VersionNumber: 1, DocxStorageKey: "tenant-a/tpl-1/v1.docx"}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.GetDocxURL(context.Background(), application.GetDocxURLCmd{
		TenantID: "tenant-b", TemplateID: tpl.ID, VersionNumber: 1,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListAudit_CrossTenant(t *testing.T) {
	repo := newFakeRepo()
	repo.ignoreTenantOnGetTemplate = true
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[tpl.ID] = tpl
	repo.audit = append(repo.audit, &domain.AuditEvent{TenantID: "tenant-a", TemplateID: tpl.ID, Action: domain.AuditCreated})

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.ListAudit(context.Background(), "tenant-b", tpl.ID, 10, 0)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
