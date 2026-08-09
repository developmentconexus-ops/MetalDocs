package resolvers

import (
	"context"
	"errors"
	"time"
)

// AuthorResolver resolves the "author" placeholder to the revision author's
// display name, falling back to their user ID when no display name is set.
type AuthorResolver struct{}

// Key returns the resolver's registry key, "author".
func (AuthorResolver) Key() string { return "author" }

// Version returns the resolver's version.
func (AuthorResolver) Version() int { return 1 }

// Resolve computes the author value for in.
func (AuthorResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	if err := requireTenantID("author", in.TenantID); err != nil {
		return ResolvedValue{}, err
	}
	if err := requireRevisionID("author", in.RevisionID); err != nil {
		return ResolvedValue{}, err
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
		TenantID   TenantID   `json:"tenant_id"`
		RevisionID RevisionID `json:"revision_id"`
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
