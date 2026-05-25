package resolvers

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type AuthorResolver struct{}

func (AuthorResolver) Key() string { return "author" }

func (AuthorResolver) Version() int { return 1 }

func (r AuthorResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if in.TenantID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key())
	}
	if in.RevisionID == "" {
		return ResolvedValue{}, fmt.Errorf("%s: RevisionID is required", r.Key())
	}
	if in.RevisionReader == nil {
		return ResolvedValue{}, errors.New("author resolver: revision reader is nil")
	}
	author, err := in.RevisionReader.GetAuthor(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
	}
	authorValue := author.DisplayName
	if authorValue == "" {
		authorValue = author.UserID
	}

	inputsHash, err := hashInputs(struct {
		TenantID   string `json:"tenant_id"`
		RevisionID string `json:"revision_id"`
	}{
		TenantID:   in.TenantID,
		RevisionID: in.RevisionID,
	})
	if err != nil {
		return ResolvedValue{}, err
	}

	return ResolvedValue{
		Value:       authorValue,
		ResolverKey: "author",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
