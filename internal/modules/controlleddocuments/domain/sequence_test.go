//go:build integration

package domain_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"metaldocs/internal/modules/controlleddocuments/infrastructure"
	"metaldocs/internal/platform/tenant"
)

func TestSequenceAllocatorNextAndIncrement_Concurrent(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("DATABASE_URL or METALDOCS_DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		t.Skipf("cannot connect to database: %v", err)
	}

	tenantID := tenant.DevTenantID
	profileCode := "seqtest" + strings.ToLower(time.Now().UTC().Format("150405"))
	areaCode := "rh"

	// The taxonomy tables (document_profiles, document_process_areas) and the
	// cd_sequence_counters table each carry the capability tripwire
	// (trg_require_cap_asserted): document_profiles/process_areas require
	// taxonomy.manage, cd_sequence_counters requires controlled_documents.create.
	// The service layer normally asserts these tx-locally via authz.Require; this
	// raw-SQL setup must do the same or every INSERT trips P0001. Seed all three
	// rows in ONE tx that asserts both caps. This also replaces the old
	// allocator.EnsureCounter call, which ran on the pool (autocommit) and so
	// could never carry a tx-local assertion.
	setupTx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin setup tx: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(),
		`SELECT set_config('metaldocs.asserted_caps',
			'[{"cap":"taxonomy.manage"},{"cap":"controlled_documents.create"}]', true)`,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("assert setup caps: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO metaldocs.document_profiles
			(code, tenant_id, family_code, name, description, review_interval_days, editable_by_role, alias)
		VALUES
			($1, $2, 'quality', 'Seq Test', 'integration sequence test', 30, 'admin', $1)`,
		profileCode, tenantID,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("insert profile: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO metaldocs.document_process_areas
			(code, tenant_id, name, description)
		VALUES
			($1, $2, 'Resources & HR', 'integration test area')
		ON CONFLICT (tenant_id, code) DO NOTHING`,
		areaCode, tenantID,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("insert process area: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO public.cd_sequence_counters
			(tenant_id, profile_code, process_area_code, next_seq)
		VALUES
			($1, $2, $3, 1)
		ON CONFLICT (tenant_id, profile_code, process_area_code) DO NOTHING`,
		tenantID, profileCode, areaCode,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("seed counter: %v", err)
	}
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("commit setup tx: %v", err)
	}

	allocator := infrastructure.NewPostgresSequenceAllocator(db)

	const workers = 50
	results := make([]int, 0, workers)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			// Each allocation runs in its own committed tx — the production shape.
			// The FOR UPDATE lock is held until Commit, serializing the 50
			// concurrent workers into a gap-free 1..50 (the allocator now rejects
			// a nil tx, so autocommit is no longer an option).
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Errorf("begin tx: %v", err)
				return
			}
			// NextAndIncrement writes cd_sequence_counters (ensure-INSERT + locked
			// UPDATE), both gated by the controlled_documents.create tripwire — the
			// service asserts it tx-locally, so the test must too.
			if _, err := tx.ExecContext(context.Background(),
				`SELECT set_config('metaldocs.asserted_caps',
					'[{"cap":"controlled_documents.create"}]', true)`,
			); err != nil {
				_ = tx.Rollback()
				t.Errorf("assert worker cap: %v", err)
				return
			}
			next, err := allocator.NextAndIncrement(context.Background(), tx, tenantID, profileCode, areaCode)
			if err != nil {
				_ = tx.Rollback()
				t.Errorf("next and increment: %v", err)
				return
			}
			if err := tx.Commit(); err != nil {
				t.Errorf("commit: %v", err)
				return
			}
			mu.Lock()
			results = append(results, next)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(results) != workers {
		t.Fatalf("expected %d results, got %d", workers, len(results))
	}

	seen := map[int]bool{}
	for _, v := range results {
		if seen[v] {
			t.Fatalf("duplicate sequence value found: %d", v)
		}
		seen[v] = true
	}
	for i := 1; i <= workers; i++ {
		if !seen[i] {
			t.Fatalf("missing sequence value %d in %v", i, results)
		}
	}
}

