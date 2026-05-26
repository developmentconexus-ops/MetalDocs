package jobs

import (
	"context"
	"log"
	"time"

	"metaldocs/internal/modules/documents/repository"
)

func StartSessionSweeper(ctx context.Context, r *repository.Repository, interval time.Duration) (stop func()) {
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
