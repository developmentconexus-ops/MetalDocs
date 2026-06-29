package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"metaldocs/internal/modules/tokens/application"
	"metaldocs/internal/modules/tokens/domain"
)

// runnerThatFails proves the reserved-name guard fires off-tx: if Create opens
// a transaction the test fails immediately.
type runnerThatFails struct{}

func (runnerThatFails) Do(_ context.Context, _ func(*sql.Tx) error) error {
	return errors.New("tx must not open")
}
func (runnerThatFails) DoReadOnly(_ context.Context, _ func(*sql.Tx) error) error {
	return errors.New("tx must not open")
}

// staticReserved is a simple in-memory ReservedNames implementation for tests.
type staticReserved map[string]struct{}

func (s staticReserved) IsReserved(name string) bool { _, ok := s[name]; return ok }

func TestCreate_ReservedName_RejectedOffTx(t *testing.T) {
	noopRequire := func(_ context.Context, _ *sql.Tx, _, _ string) error { return nil }
	noopSeed := func(_ context.Context, _ *sql.Tx, _, _ string) error { return nil }

	svc := application.NewServiceForTest(runnerThatFails{}, &fakeRepo{}, &fakeAudit{}, noopRequire, noopSeed).
		WithReservedNames(staticReserved{"author": {}, "approval_date": {}})

	_, err := svc.Create(context.Background(), application.CreateCommand{
		TenantID: "11111111-1111-1111-1111-111111111111",
		ActorID:  "22222222-2222-2222-2222-222222222222",
		Name:     "author", Value: "X", Label: "L",
	})
	if !errors.Is(err, domain.ErrReservedName) {
		t.Fatalf("want ErrReservedName, got %v", err)
	}
}
