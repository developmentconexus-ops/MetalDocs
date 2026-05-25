package resolvers

import (
	"context"
	"fmt"
	"time"
)

type DocTitleResolver struct{}

func (DocTitleResolver) Key() string  { return "doc_title" }
func (DocTitleResolver) Version() int { return 1 }

func (r DocTitleResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if in.TenantID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key())
	}
	if in.RevisionID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: RevisionID is required", r.Key())
	}
	title, err := in.DocumentReader.GetDocumentTitle(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
	}
	inputsHash, err := hashInputs(struct {
		TenantID   string `json:"tenant_id"`
		RevisionID string `json:"revision_id"`
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
