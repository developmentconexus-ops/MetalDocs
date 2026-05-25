package application

import (
	"context"
	"errors"
	"sort"
	"strings"

	"metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/authn"
)

const (
	defaultLimit            = 20
	maxLimit                = 100
	searchCapabilityView    = "document.view"
	searchScopeDocument     = "document"
	searchScopeDocumentType = "document_type"
	searchScopeArea         = "area"
)

var ErrTenantRequired = errors.New("search: tenant id required")

type Service struct {
	reader domain.Reader
}

func NewService(reader domain.Reader) *Service {
	return &Service{reader: reader}
}

func (s *Service) SearchDocuments(ctx context.Context, q domain.Query) ([]domain.Document, error) {
	tenantID := strings.TrimSpace(q.TenantID)
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	limit := effectiveLimit(q.Limit)

	q.TenantID = tenantID
	docs, err := s.reader.ListDocuments(ctx, q, limit)
	if err != nil {
		return nil, err
	}

	text := strings.ToLower(strings.TrimSpace(q.Text))
	documentFamily := strings.ToLower(strings.TrimSpace(q.DocumentFamily))
	subject := strings.ToLower(strings.TrimSpace(q.Subject))
	businessUnit := strings.TrimSpace(q.BusinessUnit)
	department := strings.TrimSpace(q.Department)
	classification := strings.ToUpper(strings.TrimSpace(string(q.Classification)))
	tag := strings.ToLower(strings.TrimSpace(q.Tag))

	filtered := make([]domain.Document, 0, len(docs))
	for _, doc := range docs {
		allowed, err := s.canView(ctx, doc)
		if err != nil {
			return nil, err
		}
		if !allowed {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(doc.Title), text) {
			continue
		}
		if documentFamily != "" && strings.ToLower(doc.DocumentFamily) != documentFamily {
			continue
		}
		if subject != "" && strings.ToLower(doc.Subject) != subject {
			continue
		}
		if businessUnit != "" && doc.BusinessUnit != businessUnit {
			continue
		}
		if department != "" && doc.Department != department {
			continue
		}
		if classification != "" && strings.ToUpper(string(doc.Classification)) != classification {
			continue
		}
		if tag != "" && !hasTag(doc.Tags, tag) {
			continue
		}
		if q.ExpiryBefore != nil {
			if doc.ExpiryAt == nil || doc.ExpiryAt.After(q.ExpiryBefore.UTC()) {
				continue
			}
		}
		if q.ExpiryAfter != nil {
			if doc.ExpiryAt == nil || doc.ExpiryAt.Before(q.ExpiryAfter.UTC()) {
				continue
			}
		}
		filtered = append(filtered, doc)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func (s *Service) canView(ctx context.Context, doc domain.Document) (bool, error) {
	if _, hasUser := authn.UserIDFromContext(ctx); !hasUser {
		return false, nil
	}
	if shouldBypassPolicy(ctx) {
		return true, nil
	}

	policies, err := s.policiesForDocument(ctx, doc)
	if err != nil {
		return false, err
	}
	return decidePolicies(ctx, policies), nil
}

func (s *Service) policiesForDocument(ctx context.Context, doc domain.Document) ([]domain.AccessPolicy, error) {
	scopes := []struct {
		scope string
		id    string
	}{
		{scope: searchScopeDocument, id: doc.ID},
		{scope: searchScopeDocumentType, id: doc.DocumentProfile},
		{scope: searchScopeArea, id: areaResourceID(doc.BusinessUnit, doc.Department)},
	}

	var out []domain.AccessPolicy
	for _, scope := range scopes {
		if strings.TrimSpace(scope.id) == "" {
			continue
		}
		items, err := s.reader.ListAccessPolicies(ctx, scope.scope, scope.id)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.Capability == searchCapabilityView {
				out = append(out, item)
			}
		}
	}
	return out, nil
}

func shouldBypassPolicy(ctx context.Context) bool {
	return false
}

func effectiveLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func decidePolicies(ctx context.Context, items []domain.AccessPolicy) bool {
	if len(items) == 0 {
		return true
	}
	// Empty userID is tolerated: only role-keyed policies can still match;
	// user-keyed policies require an authenticated principal and naturally
	// miss when none is present.
	userID, _ := authn.UserIDFromContext(ctx)
	roles := authn.RolesFromContext(ctx)
	rolesSet := map[string]struct{}{}
	for _, role := range roles {
		rolesSet[strings.ToLower(strings.TrimSpace(role))] = struct{}{}
	}

	for _, item := range items {
		if !matchesPolicySubject(item, userID, rolesSet) {
			continue
		}
		if item.Effect == domain.EffectDeny {
			return false
		}
	}
	// No matching deny policy — default allow. Policies are opt-in restrictions.
	return true
}

func matchesPolicySubject(item domain.AccessPolicy, userID string, rolesSet map[string]struct{}) bool {
	switch item.SubjectType {
	case domain.SubjectTypeUser:
		return strings.EqualFold(item.SubjectID, userID)
	case domain.SubjectTypeRole:
		_, ok := rolesSet[strings.ToLower(strings.TrimSpace(item.SubjectID))]
		return ok
	default:
		return false
	}
}

func areaResourceID(businessUnit, department string) string {
	return strings.TrimSpace(businessUnit) + ":" + strings.TrimSpace(department)
}

func hasTag(tags []string, expected string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), expected) {
			return true
		}
	}
	return false
}
