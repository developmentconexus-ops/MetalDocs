package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

// TestPublishTemplateVersion_RoleBinding is a table-driven verification that
// PublishTemplateVersion enforces the version's PendingApproverRole binding
// (Tier 2 authz) in addition to the existing template.publish capability gate
// (Tier 1). Covers the residual half of templates-tech-debt T-004.
func TestPublishTemplateVersion_RoleBinding(t *testing.T) {
	t.Parallel()

	type setup struct {
		actorRoles          []string
		pendingApproverRole string
		expectTx            bool // role check short-circuits before tx when binding fails
	}
	type expect struct {
		err         error
		statusAfter domain.VersionStatus
		auditAction domain.AuditAction
		auditCount  int
	}

	cases := []struct {
		name  string
		setup setup
		want  expect
	}{
		{
			name: "happy_path_capability_and_role_match",
			setup: setup{
				actorRoles:          []string{"approver"},
				pendingApproverRole: "approver",
				expectTx:            true,
			},
			want: expect{
				err:         nil,
				statusAfter: domain.VersionStatusPublished,
				auditAction: domain.AuditPublished,
				auditCount:  1,
			},
		},
		{
			name: "forbidden_role_when_actor_lacks_binding",
			setup: setup{
				actorRoles:          []string{"author", "system_admin"},
				pendingApproverRole: "approver",
				expectTx:            false,
			},
			want: expect{
				err:         domain.ErrForbiddenRole,
				statusAfter: domain.VersionStatusDraft,
				auditAction: domain.AuditPublishForbiddenRole,
				auditCount:  1,
			},
		},
		{
			name: "no_binding_publish_succeeds_with_capability_alone",
			setup: setup{
				actorRoles:          []string{"author"},
				pendingApproverRole: "",
				expectTx:            true,
			},
			want: expect{
				err:         nil,
				statusAfter: domain.VersionStatusPublished,
				auditAction: domain.AuditPublished,
				auditCount:  1,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			repo := newFakeRepo()
			template := &domain.Template{
				ID:            "tpl-1",
				TenantID:      "tenant-a",
				LatestVersion: 1,
			}
			version := &domain.TemplateVersion{
				ID:                  "ver-1",
				TemplateID:          template.ID,
				VersionNumber:       1,
				Status:              domain.VersionStatusDraft,
				DocxStorageKey:      "templates/tpl-1/versions/1.docx",
				ContentHash:         "hash_ok",
				AuthorID:            "author-1",
				PendingApproverRole: tc.setup.pendingApproverRole,
			}
			repo.templates[template.ID] = template
			repo.versions[version.ID] = version

			svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithDB(db)

			if tc.setup.expectTx {
				mock.ExpectBegin()
				expectTemplateAuthz(mock, "approver-1", "tenant-a", "template.publish")
				mock.ExpectCommit()
			}

			_, gotErr := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
				TenantID:      "tenant-a",
				ActorUserID:   "approver-1",
				ActorRoles:    tc.setup.actorRoles,
				TemplateID:    template.ID,
				VersionNumber: 1,
				DocxKey:       "templates/tpl-1/versions/1.docx",
				SchemaKey:     "templates/tpl-1/versions/1.schema.json",
			})

			if tc.want.err != nil {
				if !errors.Is(gotErr, tc.want.err) {
					t.Fatalf("err: want %v, got %v", tc.want.err, gotErr)
				}
			} else if gotErr != nil {
				t.Fatalf("unexpected error: %v", gotErr)
			}

			stored := repo.versions[version.ID]
			if stored.Status != tc.want.statusAfter {
				t.Fatalf("status after: want %q, got %q", tc.want.statusAfter, stored.Status)
			}

			if len(repo.audit) != tc.want.auditCount {
				t.Fatalf("audit count: want %d, got %d (%+v)", tc.want.auditCount, len(repo.audit), repo.audit)
			}
			if tc.want.auditCount > 0 && repo.audit[0].Action != tc.want.auditAction {
				t.Fatalf("audit action: want %q, got %q", tc.want.auditAction, repo.audit[0].Action)
			}

			if tc.setup.expectTx {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("sqlmock expectations: %v", err)
				}
			}
		})
	}
}
