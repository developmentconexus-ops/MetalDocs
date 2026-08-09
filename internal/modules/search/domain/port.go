package domain

import "context"

// Reader is the read port Service uses to fetch tenant-scoped, actor-visible
// documents.
type Reader interface {
	// ListDocuments returns the tenant documents matching query that are visible
	// to query.ActorUserID. Per-document visibility is enforced in the query
	// itself (data layer), not post-fetch.
	ListDocuments(ctx context.Context, query Query, limit, offset int) ([]Document, error)
}
