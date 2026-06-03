package presence

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// BumpMiddleware updates iam_users.last_seen_at for the authenticated
// user on the request's behalf. Updates are coalesced per-user via an
// in-memory map: at most one UPDATE per user per BumpDebounce window
// (60s). This keeps presence reasonably fresh without hammering the
// table on every request.
//
// The middleware must sit AFTER the auth middleware in the stack so
// iamdomain.UserIDFromContext is populated. It is a no-op for requests
// without a user.
type BumpMiddleware struct {
	repo Repository
	log  *slog.Logger
	now  func() time.Time

	mu        sync.Mutex
	lastBump  map[string]time.Time
	debounce  time.Duration
	updateCtx func(parent context.Context) (context.Context, context.CancelFunc)
}

// NewBumpMiddleware constructs the middleware.
func NewBumpMiddleware(repo Repository, log *slog.Logger) *BumpMiddleware {
	if repo == nil {
		panic("presence.NewBumpMiddleware: repo is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &BumpMiddleware{
		repo:     repo,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
		lastBump: make(map[string]time.Time),
		debounce: BumpDebounce,
		updateCtx: func(parent context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(parent, 2*time.Second)
		},
	}
}

// WithClock overrides the wall-clock; used by tests.
func (m *BumpMiddleware) WithClock(fn func() time.Time) *BumpMiddleware {
	m.now = fn
	return m
}

// WithDebounce overrides the per-user debounce window; tests use this
// to verify the gating without sleeping the test runner.
func (m *BumpMiddleware) WithDebounce(d time.Duration) *BumpMiddleware {
	m.debounce = d
	return m
}

// Wrap returns next wrapped with the bump behaviour.
func (m *BumpMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := iamdomain.UserIDFromContext(r.Context())
		if userID != "" {
			m.maybeBump(userID)
		}
		next.ServeHTTP(w, r)
	})
}

// maybeBump consults the in-memory debounce map and, if the user is
// outside the window, fires a background UPDATE. The update runs on
// its own goroutine + bounded context so the HTTP request never waits
// on the database.
func (m *BumpMiddleware) maybeBump(userID string) {
	now := m.now()
	m.mu.Lock()
	last, seen := m.lastBump[userID]
	if seen && now.Sub(last) < m.debounce {
		m.mu.Unlock()
		return
	}
	m.lastBump[userID] = now
	m.mu.Unlock()

	go func(uid string, ts time.Time) {
		ctx, cancel := m.updateCtx(context.Background())
		defer cancel()
		if err := m.repo.BumpLastSeen(ctx, uid, ts); err != nil {
			m.log.Warn("presence: bump last_seen_at failed", "user_id", uid, "err", err)
		}
	}(userID, now)
}

// BumpNow forces a synchronous bump for userID; used by the WS
// handler on heartbeat so the user is reflected as online to the
// next room tick without waiting for the next HTTP request.
func (m *BumpMiddleware) BumpNow(ctx context.Context, userID string) error {
	now := m.now()
	m.mu.Lock()
	m.lastBump[userID] = now
	m.mu.Unlock()
	return m.repo.BumpLastSeen(ctx, userID, now)
}
