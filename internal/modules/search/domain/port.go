package domain

import "context"

type Reader interface {
	ListDocuments(ctx context.Context, tenantID string) ([]Document, error)
	ListAccessPolicies(ctx context.Context, resourceScope, resourceID string) ([]AccessPolicy, error)
}
