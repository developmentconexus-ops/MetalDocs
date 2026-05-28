package resolvers

import (
	"context"
	"time"
)

type DocCodeResolver struct{}

func (DocCodeResolver) Key() string { return "doc_code" }

func (DocCodeResolver) Version() int { return 1 }

func (DocCodeResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("doc_code", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireControlledDocumentID("doc_code", in.ControlledDocumentID); err != nil {
		return ResolvedValue{}, err
	}
	rec, err := in.RegistryReader.GetControlledDocument(ctx, in.TenantID, in.ControlledDocumentID)
	if err != nil {
		return ResolvedValue{}, err
	}

	inputsHash, err := hashInputs(struct {
		TenantID             TenantID             `json:"tenant_id"`
		ControlledDocumentID ControlledDocumentID `json:"controlled_document_id"`
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
