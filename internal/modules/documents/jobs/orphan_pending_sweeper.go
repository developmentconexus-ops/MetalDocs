package jobs

import (
	"context"
	"log"
	"time"

	"metaldocs/internal/modules/documents/repository"
)

func StartOrphanPendingSweeper(ctx context.Context, r *repository.Repository, interval, maxAge time.Duration) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-maxAge)
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
