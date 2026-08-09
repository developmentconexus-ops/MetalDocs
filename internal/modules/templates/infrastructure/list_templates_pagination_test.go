package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/templates/application"
)

// TestListTemplates_OrderByHasStableTiebreaker pins the T-011 hardening: the
// query orders by t.created_at DESC, t.id DESC so OFFSET-based pages don't
// interleave/duplicate rows when multiple templates share a created_at value.
func TestListTemplates_OrderByHasStableTiebreaker(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := New(db)

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "doc_type_code", "key", "name", "description",
		"latest_version", "lv_id", "lv_revision_number", "lv_status",
		"published_version_id", "pv_version_number", "pv_revision_number", "pv_status",
		"created_by", "system_owned", "created_at", "updated_at", "archived_at",
	}).AddRow(
		"tpl-1", "tenant-1", "finance", "k1", "Template One", "",
		1, "ver-1", 0, "draft",
		nil, nil, nil, nil,
		"user-1", false, time.Now().UTC(), time.Now().UTC(), nil,
	)

	mock.ExpectQuery(`ORDER BY t\.created_at DESC, t\.id DESC`).
		WithArgs("tenant-1", nil, 50, 0, false).
		WillReturnRows(rows)

	out, err := repo.ListTemplates(context.Background(), application.ListFilter{
		TenantID: "tenant-1",
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("got %d templates, want 1", len(out))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListTemplates_ClampsLimitAndOffset pins clampTemplatesLimit's contract
// bounds (mirrors the /templates OpenAPI spec: min 1, max 200, default 50 —
// intentionally wider than the platform-wide pagination.ClampLimit 20/100,
// since /templates is x-pagination-exempt with an explicit permanent-exemption
// note) and the offset<0 -> 0 defensive clamp.
func TestListTemplates_ClampsLimitAndOffset(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		inLimit    int
		inOffset   int
		wantLimit  int
		wantOffset int
	}{
		{"zero limit clamps to contract default 50", 0, 0, 50, 0},
		{"negative limit clamps to contract default 50", -5, 0, 50, 0},
		{"over-max limit clamps to contract max 200", 500, 0, 200, 0},
		{"in-range limit passes through unchanged", 75, 0, 75, 0},
		{"limit at contract max passes through unchanged", 200, 0, 200, 0},
		{"negative offset clamps to zero", 50, -10, 50, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer func() { _ = db.Close() }()

			repo := New(db)

			mock.ExpectQuery(`FROM templates_template t`).
				WithArgs("tenant-1", nil, tc.wantLimit, tc.wantOffset, false).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "tenant_id", "doc_type_code", "key", "name", "description",
					"latest_version", "lv_id", "lv_revision_number", "lv_status",
					"published_version_id", "pv_version_number", "pv_revision_number", "pv_status",
					"created_by", "system_owned", "created_at", "updated_at", "archived_at",
				}))

			_, err = repo.ListTemplates(context.Background(), application.ListFilter{
				TenantID: "tenant-1",
				Limit:    tc.inLimit,
				Offset:   tc.inOffset,
			})
			if err != nil {
				t.Fatalf("ListTemplates: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations (limit/offset not clamped as expected): %v", err)
			}
		})
	}
}

// TestListTemplates_PublishedOnlyEmitsPredicate pins the published=true
// query-param contract: PublishedOnly=true must emit the
// published_version_id IS NOT NULL guard (via the $5 bool arg), restricting
// results to templates that currently have a published version. False
// (default) must pass false so the guard is a no-op and the management view
// (every status) is preserved.
func TestListTemplates_PublishedOnlyEmitsPredicate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		publishedOnly bool
	}{
		{"published true passes true arg", true},
		{"published false passes false arg", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer func() { _ = db.Close() }()

			repo := New(db)

			mock.ExpectQuery(`published_version_id IS NOT NULL`).
				WithArgs("tenant-1", nil, 50, 0, tc.publishedOnly).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "tenant_id", "doc_type_code", "key", "name", "description",
					"latest_version", "lv_id", "lv_revision_number", "lv_status",
					"published_version_id", "pv_version_number", "pv_revision_number", "pv_status",
					"created_by", "system_owned", "created_at", "updated_at", "archived_at",
				}))

			_, err = repo.ListTemplates(context.Background(), application.ListFilter{
				TenantID:      "tenant-1",
				PublishedOnly: tc.publishedOnly,
				Limit:         50,
				Offset:        0,
			})
			if err != nil {
				t.Fatalf("ListTemplates: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestClampTemplatesLimit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   int
		want int
	}{
		{0, templatesDefaultLimit},
		{-1, templatesDefaultLimit},
		{1, 1},
		{200, 200},
		{201, templatesMaxLimit},
		{1000, templatesMaxLimit},
	}
	for _, tc := range cases {
		if got := clampTemplatesLimit(tc.in); got != tc.want {
			t.Errorf("clampTemplatesLimit(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
