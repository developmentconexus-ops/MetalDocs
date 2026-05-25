package domain

import "context"

type Reader interface {
	ListDocuments(ctx context.Context, tenantID string, limit int) ([]Document, error)
	ListAccessPolicies(ctx context.Context, resourceScope, resourceID string) ([]AccessPolicy, error)
}
