package wiring

import (
	"context"
	"sort"
	"strings"

	authdomain "metaldocs/internal/modules/auth/domain"
	docapp "metaldocs/internal/modules/documents/application"
)

// authUserLister is the narrow function-typed seam the documents-side adapter
// consumes. *auth.Service.ListUsers satisfies it directly — pass the method
// value. Keeping the seam this thin (a single function, not a multi-method
// interface) avoids hauling the entire auth.Service surface into the documents
// composition path and keeps the unit test fake trivial.
type authUserLister interface {
	ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)
}

// DocumentsIAMUserOptions adapts an auth user-lister to
// documents.application.IAMUserOptionsReader. It is the production adapter for
// the consumer-defined port at internal/modules/documents/application/iam_user_options.go.
//
// Semantics (binding — see f3.1 spec.md):
//   - Filters auth.ManagedUser.IsActive == true (deactivated users dropped).
//   - Maps to UserOption{UserID, DisplayName}.
//   - Sorts by strings.ToLower(DisplayName) ASC; tie-break by UserID ASC.
//   - Returns a non-nil empty slice when no users qualify.
//   - Propagates the underlying error unchanged on failure.
type DocumentsIAMUserOptions struct {
	lister authUserLister
}

// NewDocumentsIAMUserOptions returns the production
// documents.application.IAMUserOptionsReader. Panics if lister is nil — the
// composition root is the only caller and must wire a real source.
func NewDocumentsIAMUserOptions(lister authUserLister) *DocumentsIAMUserOptions {
	if lister == nil {
		panic("wiring: NewDocumentsIAMUserOptions: lister is nil")
	}
	return &DocumentsIAMUserOptions{lister: lister}
}

// ListUserOptions implements docapp.IAMUserOptionsReader.
func (a *DocumentsIAMUserOptions) ListUserOptions(ctx context.Context, tenantID string) ([]docapp.UserOption, error) {
	users, err := a.lister.ListUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]docapp.UserOption, 0, len(users))
	for _, u := range users {
		if !u.IsActive {
			continue
		}
		out = append(out, docapp.UserOption{
			UserID:      u.UserID,
			DisplayName: u.DisplayName,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		li := strings.ToLower(out[i].DisplayName)
		lj := strings.ToLower(out[j].DisplayName)
		if li != lj {
			return li < lj
		}
		return out[i].UserID < out[j].UserID
	})
	return out, nil
}
