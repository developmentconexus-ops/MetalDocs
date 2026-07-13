package infrastructure_test

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// fakeDisplayNamePort is a test double for iamdomain.UserDisplayNameReader.
type fakeDisplayNamePort struct {
	value string
	err   error
	// recorded args
	calledTenantID string
	calledUserID   string
}

func (f *fakeDisplayNamePort) DisplayName(_ context.Context, tenantID, userID string) (string, error) {
	f.calledTenantID = tenantID
	f.calledUserID = userID
	return f.value, f.err
}

func (f *fakeDisplayNamePort) DisplayNames(_ context.Context, _ string, _ []string) (map[string]string, error) {
	return nil, nil
}

var _ iamdomain.UserDisplayNameReader = (*fakeDisplayNamePort)(nil)
