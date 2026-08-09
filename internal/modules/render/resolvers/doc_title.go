package resolvers

import (
	"context"
	"time"
)

// DocTitleResolver resolves the "doc_title" placeholder to the document's
// title.
type DocTitleResolver struct{}

// Key returns the resolver's registry key, "doc_title".
func (DocTitleResolver) Key() string { return "doc_title" }

// Version returns the resolver's version.
func (DocTitleResolver) Version() int { return 1 }

// Resolve computes the doc_title value for in.
func (DocTitleResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("doc_title", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireRevisionID("doc_title", in.RevisionID); err != nil {
		return ResolvedValue{}, err
	}
	title, err := in.DocumentReader.GetDocumentTitle(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
	}
	inputsHash, err := hashInputs(struct {
		TenantID   TenantID   `json:"tenant_id"`
		RevisionID RevisionID `json:"revision_id"`
	}{in.TenantID, in.RevisionID})
	if err != nil {
		return ResolvedValue{}, err
	}
	return ResolvedValue{
		Value:       title,
		ResolverKey: "doc_title",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
