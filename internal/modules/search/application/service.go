package application

import (
	"context"
	"errors"
	"strings"

	"metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/pagination"
)

// ErrTenantRequired is returned when a search is attempted without a
// resolved tenant id.
var ErrTenantRequired = errors.New("search: tenant id required")

// Service orchestrates tenant- and actor-scoped document search.
type Service struct {
	reader domain.Reader
}

// NewService builds a Service backed by reader.
func NewService(reader domain.Reader) *Service {
	return &Service{reader: reader}
}

// SearchDocuments returns the tenant documents matching q that the authenticated
// caller is allowed to see. Per-document visibility is enforced by the reader at
// the data layer against the unified model (role/area capabilities +
// controlled_document_{area,user}_grants). An unauthenticated caller sees
// nothing.
func (s *Service) SearchDocuments(ctx context.Context, q domain.Query) ([]domain.Document, error) {
	actorUserID, hasUser := authn.UserIDFromContext(ctx)
	if actorUserID = strings.TrimSpace(actorUserID); !hasUser || actorUserID == "" {
		// No authenticated principal → no visibility into any document.
		if q.TenantID == "" {
			return nil, ErrTenantRequired
		}
		return []domain.Document{}, nil
	}

	q.ActorUserID = actorUserID
	normalized, err := domain.NewQuery(q)
	if err != nil {
		if errors.Is(err, domain.ErrQueryTenantEmpty) {
			return nil, ErrTenantRequired
		}
		return nil, err
	}
	normalized.Limit = q.Limit

	// Search offset is not yet wired end-to-end: domain.Query has no Offset
	// field and the OpenAPI /search/documents contract declares no offset
	// param, so this always requests the first page (CON-11/APP-03 finding —
	// reported, not fixed here; a contract change would be needed to expose
	// paging beyond the first `limit` results).
	limit := pagination.ClampLimit(q.Limit)
	docs, err := s.reader.ListDocuments(ctx, normalized, limit, 0)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		return []domain.Document{}, nil
	}
	return docs, nil
}
