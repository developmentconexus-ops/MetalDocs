package jobs

import (
	"context"
	"log/slog"
	"time"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
)

// StartSessionSweeper starts a background ticker that periodically expires
// stale editing sessions, and returns a stop func that cancels it.
func StartSessionSweeper(ctx context.Context, r *infrastructure.Repository, interval time.Duration) (stop func()) {
	// Background root: mark the context so ExpireStaleSessions' authz.BypassSystem
	// is permitted (fail-closed off any HTTP path — ADR 0022 Phase 7, CWE-269).
	ctx = authz.WithBackgroundBypass(ctx)
	ctx, cancel := context.WithCancel(ctx) //nolint:gosec // G118: cancel IS the returned stop func; the composition root calls it on shutdown
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n, err := r.ExpireStaleSessions(ctx, time.Now())
				if err != nil {
					slog.Warn("session_sweeper error", "err", err)
					continue
				}
				if n > 0 {
					slog.Info("session_sweeper expired", "expired", n)
				}
			}
		}
	}()
	return cancel
}
