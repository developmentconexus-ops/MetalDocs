package resolvers

import (
	"context"
	"time"
)

type DocTitleResolver struct{}

func (DocTitleResolver) Key() string  { return "doc_title" }
func (DocTitleResolver) Version() int { return 1 }

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
