package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// RoutePreview is the read-only DTO returned by SubmitService.PreviewRoute /
// TemplateSubmitService.PreviewRoute (SLICE 8a, unit 3.2): the resolved
// active route's stages and each stage's ActorSelectors, surfaced to a
// submitter BEFORE they submit so a chosen_actors picker can be built
// client-side for any submit_choice-governed stage. A zero-value RoutePreview
// (RouteID == "", Stages == nil) means no active route resolves for the
// subject yet — informational absence, not an error; submit-time resolution
// (resolveActiveRoute / LoadActiveRouteIDBySubject) remains the sole
// authority on whether a submit will actually succeed.
type RoutePreview struct {
	RouteID   string
	RouteName string
	Stages    []RoutePreviewStage
}

// RoutePreviewStage is one stage of a RoutePreview.
type RoutePreviewStage struct {
	StageOrder int
	Label      string
	Selectors  []domain.ActorSelector
}

// loadRoutePreviewByRouteID loads routeID's stages+selectors via
// repo.LoadRoute and maps them to a RoutePreview. This is the single reusable
// core both SubmitService.PreviewRoute (document) and
// TemplateSubmitService.PreviewRoute (template) delegate to once they have
// each resolved their own subject-specific routeID — so the actual
// route/stage/selector read is defined exactly once.
func loadRoutePreviewByRouteID(ctx context.Context, tx *sql.Tx, repo infrastructure.ApprovalRepository, tenantID, routeID string) (RoutePreview, error) {
	route, err := repo.LoadRoute(ctx, tx, tenantID, routeID)
	if err != nil {
		return RoutePreview{}, fmt.Errorf("preview route: load route: %w", err)
	}
	stages := make([]RoutePreviewStage, 0, len(route.Stages))
	for _, s := range route.Stages {
		stages = append(stages, RoutePreviewStage{
			StageOrder: s.Order,
			Label:      s.Name,
			Selectors:  s.Selectors,
		})
	}
	return RoutePreview{RouteID: route.ID, RouteName: route.Name, Stages: stages}, nil
}

// PreviewRoute resolves the active approval route for documentID exactly the
// way SubmitRevisionForReview would (resolveActiveRoute, ADR 0073) and
// returns its stages+selectors — read-only, no writes, no tx side effects
// beyond the tier-2 authz check. Gated by CapDocumentSubmit (the same
// capability the submit endpoint itself requires; no new capability is
// introduced). Returns a zero-value RoutePreview (not an error) when the
// document has no linked profile or the profile has no active route —
// mirrors resolveActiveRoute's own sentinels, collapsed here to "nothing to
// preview yet" since this is an informational read, not a submit attempt.
func (s *SubmitService) PreviewRoute(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (RoutePreview, error) {
	var preview RoutePreview
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// Mirrors SubmitRevisionForReview's own tier-2 gate: area-grade
		// CapDocumentSubmit, area resolved from the document row (never a
		// request field).
		areaCode, err := resolveSubjectAreaCode(ctx, tx, s.cdRead, tenantID, domain.NewDocumentSubject(documentID))
		if err != nil {
			return fmt.Errorf("preview route: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSubmit), areaCode); err != nil {
			return err
		}

		routeID, err := s.resolveActiveRoute(ctx, tx, tenantID, documentID)
		if err != nil {
			if errors.Is(err, docsdomain.ErrProfileNotConfigured) || errors.Is(err, docsdomain.ErrApprovalRouteMissing) {
				return nil // preview stays the zero value: nothing to preview yet
			}
			return fmt.Errorf("preview route: resolve active route: %w", err)
		}

		p, err := loadRoutePreviewByRouteID(ctx, tx, s.repo, tenantID, routeID)
		if err != nil {
			return err
		}
		preview = p
		return nil
	})
	return preview, err
}

// PreviewRoute resolves the active approval route governing templateID exactly
// the way SubmitTemplateVersionForReview would — the template's doc_type_code
// read through the SAME TemplateVersionReader port, then
// LoadActiveRouteIDBySubject on (template, doc_type_code) (ADR 0086) — and
// returns its stages+selectors. Resolving through the identical port method is
// what keeps preview and submit from drifting onto different routes. Gated by
// CapTemplateSubmit (the same capability the submit endpoint itself requires;
// no new capability is introduced). Returns a zero-value RoutePreview (not an
// error) when no active route resolves for the template's profile yet.
func (s *TemplateSubmitService) PreviewRoute(ctx context.Context, runner db.TxRunner, tenantID, templateID string) (RoutePreview, error) {
	var preview RoutePreview
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// template.submit is ScopeTenant (ADR 0083) — area-blind by design,
		// mirrors SubmitTemplateVersionForReview's own tier-2 gate exactly.
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), "tenant"); err != nil {
			return err
		}

		if s.routeResolver == nil {
			return fmt.Errorf("preview route: route resolver not configured")
		}
		if s.versionReader == nil {
			return fmt.Errorf("preview route: template version reader not configured")
		}
		docTypeCode, err := s.versionReader.LoadTemplateDocTypeCode(ctx, tx, tenantID, templateID)
		if err != nil {
			return fmt.Errorf("preview route: resolve template doc_type_code: %w", err)
		}
		routeID, err := s.routeResolver.LoadActiveRouteIDBySubject(ctx, tx, tenantID, string(domain.SubjectKindTemplate), docTypeCode)
		if err != nil {
			if errors.Is(err, infrastructure.ErrNoActiveApprovalRoute) {
				return nil // preview stays the zero value: nothing to preview yet
			}
			return fmt.Errorf("preview route: resolve active route: %w", err)
		}

		p, err := loadRoutePreviewByRouteID(ctx, tx, s.repo, tenantID, routeID)
		if err != nil {
			return err
		}
		preview = p
		return nil
	})
	return preview, err
}
