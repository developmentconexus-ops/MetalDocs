package jobs

import (
	"context"
	"log/slog"
	"time"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
)

// StartOrphanPendingSweeper starts a background ticker that periodically
// deletes pending uploads older than maxAge, and returns a stop func that
// cancels it.
func StartOrphanPendingSweeper(ctx context.Context, r *infrastructure.Repository, interval, maxAge time.Duration) (stop func()) {
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
					slog.Warn("orphan_pending_sweeper error", "err", err)
					continue
				}
				if n > 0 {
					slog.Info("orphan_pending_sweeper deleted", "deleted", n)
				}
			}
		}
	}()
	return cancel
}
