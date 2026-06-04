package jobs

import (
	"context"
	"log"
	"time"

	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
)

func StartSessionSweeper(ctx context.Context, r *repository.Repository, interval time.Duration) (stop func()) {
	// Background root: mark the context so ExpireStaleSessions' authz.BypassSystem
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
				n, err := r.ExpireStaleSessions(ctx, time.Now())
				if err != nil {
					log.Printf("session_sweeper error: %v", err)
					continue
				}
				if n > 0 {
					log.Printf("session_sweeper expired=%d", n)
				}
			}
		}
	}()
	return cancel
}
