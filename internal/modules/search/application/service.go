package application

import (
	"context"
	"errors"
	"strings"

	"metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/authn"
)

const (
	defaultLimit            = 20
	maxLimit                = 100
	searchCapabilityView    = "document.view"
	searchPolicyEffectDeny  = "deny"
	searchSubjectTypeUser   = "user"
	searchSubjectTypeRole   = "role"
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
	normalized, err := domain.NewQuery(q)
	if err != nil {
		if errors.Is(err, domain.ErrQueryTenantEmpty) {
			return nil, ErrTenantRequired
		}
		return nil, err
	}
	limit := effectiveLimit(q.Limit)

	normalized.Limit = q.Limit
	filtered := make([]domain.Document, 0, limit)
	offset := 0
	for len(filtered) < limit {
		docs, err := s.reader.ListDocuments(ctx, normalized, limit, offset)
		if err != nil {
			return nil, err
		}
		if len(docs) == 0 {
			break
		}
		for _, doc := range docs {
			allowed, err := s.canView(ctx, doc)
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
			filtered = append(filtered, doc)
			if len(filtered) == limit {
				break
			}
		}
		if len(docs) < limit {
			break
		}
		offset += len(docs)
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
	}
	for _, areaID := range areaPolicyResourceIDs(doc) {
		scopes = append(scopes, struct {
			scope string
			id    string
		}{
			scope: searchScopeArea,
			id:    areaID,
		})
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
		rolesSet[strings.ToLower(strings.TrimSpace(string(role)))] = struct{}{}
	}

	for _, item := range items {
		if !matchesPolicySubject(item, userID, rolesSet) {
			continue
		}
		if item.Effect == searchPolicyEffectDeny {
			return false
		}
	}
	// No matching deny policy — default allow. Policies are opt-in restrictions.
	return true
}

func matchesPolicySubject(item domain.AccessPolicy, userID string, rolesSet map[string]struct{}) bool {
	switch item.SubjectType {
	case searchSubjectTypeUser:
		return strings.EqualFold(item.SubjectID, userID)
	case searchSubjectTypeRole:
		_, ok := rolesSet[strings.ToLower(strings.TrimSpace(item.SubjectID))]
		return ok
	default:
		return false
	}
}

func areaResourceID(businessUnit, department string) string {
	normalizedBusinessUnit := strings.TrimSpace(businessUnit)
	normalizedDepartment := strings.TrimSpace(department)
	if normalizedBusinessUnit == "" || normalizedDepartment == "" {
		return ""
	}
	return normalizedBusinessUnit + ":" + normalizedDepartment
}

func areaPolicyResourceIDs(doc domain.Document) []string {
	candidates := []string{doc.BusinessUnit, doc.ProcessArea}
	out := make([]string, 0, len(candidates))
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		id := areaResourceID(candidate, doc.Department)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
