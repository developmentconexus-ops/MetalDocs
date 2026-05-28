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
	repo.templates["tpl-1"] = &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	status := domain.VersionStatusDraft
	docType := "CONTRACT"
	filter := application.ListFilter{
		TenantID:    "tenant-a",
		DocTypeCode: &docType,
		Status:      &status,
		Limit:       10,
		Offset:      2,
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
	if repo.receivedFilter.Limit != filter.Limit || repo.receivedFilter.Offset != filter.Offset {
		t.Fatalf("expected limit/offset %d/%d, got %d/%d", filter.Limit, filter.Offset, repo.receivedFilter.Limit, repo.receivedFilter.Offset)
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
