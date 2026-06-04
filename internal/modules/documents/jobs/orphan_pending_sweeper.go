package jobs

import (
	"context"
	"log"
	"time"

	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
)

func StartOrphanPendingSweeper(ctx context.Context, r *repository.Repository, interval, maxAge time.Duration) (stop func()) {
	// Background root: mark the context so DeleteExpiredPending's authz.BypassSystem
	// is permitted (fail-closed off any HTTP path — ADR 0022 Phase 7, CWE-269).
	ctx = authz.WithBackgroundBypass(ctx)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().UTC().Add(-maxAge)
				n, err := r.DeleteExpiredPending(ctx, cutoff)
				if err != nil {
					log.Printf("orphan_pending_sweeper error: %v", err)
					continue
				}
				if n > 0 {
					log.Printf("orphan_pending_sweeper deleted=%d", n)
				}
			}
		}
	}()
	return cancel
}
