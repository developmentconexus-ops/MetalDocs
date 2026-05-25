package resolvers

import (
	"context"
	"fmt"
	"time"
)

type DocCodeResolver struct{}

func (DocCodeResolver) Key() string { return "doc_code" }

func (DocCodeResolver) Version() int { return 1 }

func (r DocCodeResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if in.TenantID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key())
	}
	rec, err := in.RegistryReader.GetControlledDocument(ctx, in.TenantID, in.ControlledDocumentID)
	if err != nil {
		return ResolvedValue{}, err
	}

	inputsHash, err := hashInputs(struct {
		TenantID             string `json:"tenant_id"`
		ControlledDocumentID string `json:"controlled_document_id"`
	}{
		TenantID:             in.TenantID,
		ControlledDocumentID: in.ControlledDocumentID,
	})
	if err != nil {
		return ResolvedValue{}, err
	}

	return ResolvedValue{
		Value:       rec.DocCode,
		ResolverKey: "doc_code",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
