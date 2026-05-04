//go:build integration

package documents_integration

import (
	"context"
	"sync"
	"testing"

	"metaldocs/internal/testdb"
)

func TestCreateDocument_ConcurrentSubmitsSerialised(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()

	cdID := testdb.InsertControlledDocument(t, db, testdb.DevTenantID)

	const N = 10
	var wg sync.WaitGroup
	errs := make(chan error, N)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, _, _, err := testdb.CreateDocumentForCD(ctx, db, testdb.DevTenantID, cdID)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent create failed: %v", err)
		}
	}

	rows, err := db.QueryContext(ctx,
		`SELECT revision_number FROM metaldocs.documents
		  WHERE controlled_document_id=$1 ORDER BY revision_number`, cdID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []int
	for rows.Next() {
		var n int
		_ = rows.Scan(&n)
		got = append(got, n)
	}
	if len(got) != N {
		t.Fatalf("expected %d rows, got %d (%v)", N, len(got), got)
	}
	for i, n := range got {
		if n != i+1 {
			t.Fatalf("revision_number[%d] = %d, want %d", i, n, i+1)
		}
	}
}
