package domain

import "context"

type Reader interface {
	ListDocuments(ctx context.Context, query Query, limit int) ([]Document, error)
	ListAccessPolicies(ctx context.Context, resourceScope, resourceID string) ([]AccessPolicy, error)
}
