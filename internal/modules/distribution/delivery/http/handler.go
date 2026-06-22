// Package distributionhttp implements the HTTP delivery layer for the distribution
// module. It satisfies the generated distributionapi.StrictServerInterface using
// the CoverageRepository for data access. Auth (401/403) is produced upstream by
// tier-1 middleware; this handler only sees authenticated, authorized requests.
package distributionhttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	distributionapi "metaldocs/internal/modules/distribution/api"
	distributioninfra "metaldocs/internal/modules/distribution/infrastructure"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// Repository is the minimal surface the handler needs from the infrastructure layer.
type Repository interface {
	Summary(ctx context.Context, tenantID, cdID string) (int, error)
	Recipients(ctx context.Context, tenantID, cdID, cursor string, limit int) (distributioninfra.RecipientsPage, error)
	Coverage(ctx context.Context, tenantID, cdID string) ([]distributioninfra.AreaCoverageRow, error)
}

// Handler implements distributionapi.StrictServerInterface.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler. repo must not be nil.
func NewHandler(repo Repository) *Handler {
	if repo == nil {
		panic("distributionhttp: repo is required")
	}
	return &Handler{repo: repo}
}

// GetDocumentDistribution handles GET /documents/{id}/distribution.
// Returns the distinct obligated-reader count (denominator only; no numerator).
func (h *Handler) GetDocumentDistribution(
	ctx context.Context,
	req distributionapi.GetDocumentDistributionRequestObject,
) (distributionapi.GetDocumentDistributionResponseObject, error) {
	tenantID, ok := extractTenant(ctx)
	if !ok {
		return distributionapi.GetDocumentDistribution404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: distributionapi.NotFoundApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusNotFound, problem.CodeNotFound, "Document not found")),
			),
		}, nil
	}

	cdID := req.Id.String()

	total, err := h.repo.Summary(ctx, tenantID, cdID)
	if err != nil {
		slog.Error("distribution.GetDocumentDistribution", "error", err)
		return distributionapi.GetDocumentDistribution500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: distributionapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	return distributionapi.GetDocumentDistribution200JSONResponse{
		TotalTargets: total,
	}, nil
}

// ListDocumentDistributionRecipients handles GET /documents/{id}/distribution/recipients.
// Returns a paginated list of obligated readers with display names and area info.
func (h *Handler) ListDocumentDistributionRecipients(
	ctx context.Context,
	req distributionapi.ListDocumentDistributionRecipientsRequestObject,
) (distributionapi.ListDocumentDistributionRecipientsResponseObject, error) {
	tenantID, ok := extractTenant(ctx)
	if !ok {
		return distributionapi.ListDocumentDistributionRecipients404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: distributionapi.NotFoundApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusNotFound, problem.CodeNotFound, "Document not found")),
			),
		}, nil
	}

	cdID := req.Id.String()
	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	limit := pagination.DefaultLimit
	if req.Params.Limit != nil && *req.Params.Limit > 0 {
		limit = *req.Params.Limit
	}

	page, err := h.repo.Recipients(ctx, tenantID, cdID, cursor, limit)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			return distributionapi.ListDocumentDistributionRecipients400ApplicationProblemPlusJSONResponse{
				BadRequestApplicationProblemPlusJSONResponse: distributionapi.BadRequestApplicationProblemPlusJSONResponse(
					toProblem(problem.New(http.StatusBadRequest, problem.CodeInvalidCursor, "Invalid cursor")),
				),
			}, nil
		}
		slog.Error("distribution.ListDocumentDistributionRecipients", "error", err)
		return distributionapi.ListDocumentDistributionRecipients500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: distributionapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	items := make([]distributionapi.DistributionRecipient, len(page.Items))
	for i, r := range page.Items {
		src := distributionapi.DistributionRecipientSource(r.Source)
		items[i] = distributionapi.DistributionRecipient{
			UserId:   r.UserID,
			Name:     r.Name,
			AreaCode: r.AreaCode,
			AreaName: r.AreaName,
			Source:   src,
		}
	}

	var nextCursor *string
	if page.HasMore && page.NextCursor != "" {
		nc := page.NextCursor
		nextCursor = &nc
	}

	return distributionapi.ListDocumentDistributionRecipients200JSONResponse{
		Items: items,
		Page: distributionapi.CursorPage{
			HasMore:    page.HasMore,
			NextCursor: nextCursor,
		},
	}, nil
}

// GetDocumentDistributionCoverage handles GET /documents/{id}/distribution/coverage.
// Returns per-area coverage counts ordered by area_name. Empty array for company-scope docs.
func (h *Handler) GetDocumentDistributionCoverage(
	ctx context.Context,
	req distributionapi.GetDocumentDistributionCoverageRequestObject,
) (distributionapi.GetDocumentDistributionCoverageResponseObject, error) {
	tenantID, ok := extractTenant(ctx)
	if !ok {
		return distributionapi.GetDocumentDistributionCoverage404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: distributionapi.NotFoundApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusNotFound, problem.CodeNotFound, "Document not found")),
			),
		}, nil
	}

	cdID := req.Id.String()

	areas, err := h.repo.Coverage(ctx, tenantID, cdID)
	if err != nil {
		slog.Error("distribution.GetDocumentDistributionCoverage", "error", err)
		return distributionapi.GetDocumentDistributionCoverage500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: distributionapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	result := make(distributionapi.GetDocumentDistributionCoverage200JSONResponse, len(areas))
	for i, a := range areas {
		result[i] = distributionapi.DistributionAreaCoverage{
			AreaCode: a.AreaCode,
			AreaName: a.AreaName,
			Total:    a.Total,
		}
	}
	return result, nil
}

// extractTenant reads the authenticated tenant from context. Returns ("", false) when
// auth context is missing — the caller should return 404 (the document is not found
// in the caller's tenant, and we never disclose whether it exists in another).
// 401/403 are handled upstream by tier-1 middleware before the handler is invoked.
func extractTenant(ctx context.Context) (string, bool) {
	if _, ok := authn.UserIDFromContext(ctx); !ok {
		return "", false
	}
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return "", false
	}
	return tenantID, true
}

// toProblem maps a *problem.Problem to the generated distributionapi.Problem shape.
func toProblem(p *problem.Problem) distributionapi.Problem {
	var detail *string
	if p.Detail != "" {
		d := p.Detail
		detail = &d
	}
	var instance *string
	if p.Instance != "" {
		inst := p.Instance
		instance = &inst
	}
	return distributionapi.Problem{
		Status:   p.Status,
		Title:    p.Title,
		Code:     string(p.Code),
		Detail:   detail,
		Instance: instance,
	}
}
