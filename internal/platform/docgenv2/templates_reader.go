package docgenv2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
)

// TemplatesTemplateReader implements documents/application.TemplateReader
// for templates authored via the templates module. Schema is stored in the
// database (not S3), so schemaKey is always "" and schemaJSON is always "".
type TemplatesTemplateReader struct {
	db *sql.DB
}

func NewTemplatesTemplateReader(db *sql.DB) *TemplatesTemplateReader {
	return &TemplatesTemplateReader{db: db}
}

func (r *TemplatesTemplateReader) GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error) {
	if r.db == nil {
		return "", "", "", errors.New("templates template reader: db is nil")
	}
	err = r.db.QueryRowContext(ctx, `
		SELECT tv.docx_storage_key
		FROM templates_template_version tv
		JOIN templates_template tpl ON tpl.id = tv.template_id
		WHERE tv.id = $1
		  AND (tpl.tenant_id = $2::uuid OR tpl.tenant_id = $3::uuid)
		  AND tv.status = 'published'`,
		templateVersionID, tenantID, systemTemplateTenantID,
	).Scan(&docxKey)
	if err != nil {
		return "", "", "", fmt.Errorf("templates reader: %w", err)
	}
	return docxKey, "", "", nil
}

// legacyTemplateReadTotal counts GetPublishedVersion calls that were served by the
// legacy fallback (public.templates/template_versions) after the canonical
// templates_template_version lookup reported not-found. ARC-01: the canonical
// family is now the primary read path (2026-07-01); this counter is the residual
// legacy-usage signal that gates DB-01 (dropping the legacy tables/reader once it
// stays at zero across a full run window). Process-wide atomic counter, matching
// the platform's existing lightweight-metrics idiom (see
// internal/platform/observability/http.go routeMetrics, internal/modules/iam/
// authz/bypass_audit.go) rather than a Prometheus client — this process has no
// metrics-exporter plumbing today. Not yet wired into the /api/v1/metrics
// endpoint; LegacyTemplateReadCount exposes it for tests and future wiring.
var legacyTemplateReadTotal uint64

// LegacyTemplateReadCount returns the number of FanoutTemplateReader calls served
// by the legacy fallback reader since process start. Exported for tests and for a
// future /api/v1/metrics provider (see legacyTemplateReadTotal).
func LegacyTemplateReadCount() uint64 {
	return atomic.LoadUint64(&legacyTemplateReadTotal)
}

// FanoutTemplateReader reads the canonical templates_template_version family
// first; only when the canonical reader reports domain not-found (sql.ErrNoRows)
// does it fall back to the legacy public.templates/template_versions reader.
// ARC-01 (2026-07-01): flipped from legacy-first to canonical-first now that the
// canonical family is the live write path for template publication. The legacy
// fallback is retained, instrumented, until DB-01 proves zero residual legacy
// hits and removes the legacy reader/tables.
//
// Error semantics: a non-not-found error from the canonical reader (e.g. a real
// DB error) surfaces immediately and does NOT fall back — only "no such
// published version" triggers the legacy read, so callers cannot have a live DB
// outage silently masked as "check the legacy table instead."
type FanoutTemplateReader struct {
	primary   *TemplatesTemplateReader // canonical: templates_template_version
	secondary *TemplateReader          // legacy: template_versions (fallback only)
}

func NewFanoutTemplateReader(primary *TemplatesTemplateReader, secondary *TemplateReader) *FanoutTemplateReader {
	return &FanoutTemplateReader{primary: primary, secondary: secondary}
}

func (f *FanoutTemplateReader) GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error) {
	docxKey, schemaKey, schemaJSON, err = f.primary.GetPublishedVersion(ctx, tenantID, templateVersionID)
	if err == nil {
		return docxKey, schemaKey, schemaJSON, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", "", fmt.Errorf("templates reader primary: %w", err)
	}

	docxKey, schemaKey, schemaJSON, err = f.secondary.GetPublishedVersion(ctx, tenantID, templateVersionID)
	if err != nil {
		return "", "", "", fmt.Errorf("templates reader secondary: %w", err)
	}

	atomic.AddUint64(&legacyTemplateReadTotal, 1)
	slog.Warn("docgenv2: served published template version from legacy fallback reader",
		"tenant_id", tenantID,
		"template_version_id", templateVersionID,
		"legacy_template_read_total", atomic.LoadUint64(&legacyTemplateReadTotal),
	)
	return docxKey, schemaKey, schemaJSON, nil
}
