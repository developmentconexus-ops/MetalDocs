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
	docs, err := s.reader.ListDocuments(ctx, normalized, limit)
	if err != nil {
		return nil, err
	}
	// TODO(search): push filtering into the DB reader so tenant-sized result sets
	// do not require a full in-memory scan before authz and limit trimming.

	text := strings.ToLower(strings.TrimSpace(normalized.Text))
	documentType := strings.ToLower(strings.TrimSpace(normalized.DocumentType))
	documentProfile := strings.ToLower(strings.TrimSpace(normalized.DocumentProfile))
	documentFamily := strings.ToLower(strings.TrimSpace(normalized.DocumentFamily))
	processArea := strings.ToLower(strings.TrimSpace(normalized.ProcessArea))
	subject := strings.ToLower(strings.TrimSpace(normalized.Subject))
	ownerID := strings.TrimSpace(normalized.OwnerID)
	businessUnit := strings.TrimSpace(normalized.BusinessUnit)
	department := strings.TrimSpace(normalized.Department)
	classification := strings.ToUpper(strings.TrimSpace(string(normalized.Classification)))
	status := strings.ToUpper(strings.TrimSpace(string(normalized.Status)))
	tag := strings.ToLower(strings.TrimSpace(normalized.Tag))

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
		if documentType != "" && strings.ToLower(doc.DocumentType) != documentType {
			continue
		}
		if documentProfile != "" && strings.ToLower(doc.DocumentProfile) != documentProfile {
			continue
		}
		if documentFamily != "" && strings.ToLower(doc.DocumentFamily) != documentFamily {
			continue
		}
		if processArea != "" && strings.ToLower(doc.ProcessArea) != processArea {
			continue
		}
		if subject != "" && strings.ToLower(doc.Subject) != subject {
			continue
		}
		if ownerID != "" && doc.OwnerID != ownerID {
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
		if status != "" && strings.ToUpper(string(doc.Status)) != status {
			continue
		}
		if tag != "" && !hasTag(doc.Tags, tag) {
			continue
		}
		if normalized.ExpiryBefore != nil {
			if doc.ExpiryAt == nil || doc.ExpiryAt.After(normalized.ExpiryBefore.UTC()) {
				continue
			}
		}
		if normalized.ExpiryAfter != nil {
			if doc.ExpiryAt == nil || doc.ExpiryAt.Before(normalized.ExpiryAfter.UTC()) {
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
