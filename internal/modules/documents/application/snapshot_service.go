package application

import (
	"context"
	"fmt"

	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
)

// SnapshotTemplateReader loads a template's artifact data for snapshotting.
type SnapshotTemplateReader interface {
	LoadForSnapshot(ctx context.Context, tenantID, templateID string) (domain.TemplateSnapshot, error)
}

// SnapshotService resolves template artifacts for use at document creation time.
type SnapshotService struct {
	templates SnapshotTemplateReader
}

// NewSnapshotService constructs a SnapshotService.
func NewSnapshotService(t SnapshotTemplateReader) *SnapshotService {
	return &SnapshotService{templates: t}
}

// ResolveTemplate loads the template snapshot and required-placeholder list
// without writing to the DB. Used by Service.Create so the snapshot is
// written atomically with the documents row inside CreateDocument.
func (s *SnapshotService) ResolveTemplate(ctx context.Context, tenantID, templateID string) (domain.TemplateSnapshot, []templatesdomain.Placeholder, error) {
	snap, err := s.templates.LoadForSnapshot(ctx, tenantID, templateID)
	if err != nil {
		return domain.TemplateSnapshot{}, nil, err
	}
	phs, err := parseRequiredPlaceholders(snap.PlaceholderSchemaJSON)
	if err != nil {
		return domain.TemplateSnapshot{}, nil, fmt.Errorf("parse placeholder schema: %w", err)
	}
	return snap, phs, nil
}

// parseRequiredPlaceholders extracts placeholders with Required=true from
// the placeholder schema JSON blob. Returns empty slice on empty/nil input.
func parseRequiredPlaceholders(schemaJSON []byte) ([]templatesdomain.Placeholder, error) {
	all, err := parsePlaceholderSchema(schemaJSON)
	if err != nil {
		return nil, err
	}
	var out []templatesdomain.Placeholder
	for _, p := range all {
		if p.Required {
			out = append(out, p)
		}
	}
	return out, nil
}
