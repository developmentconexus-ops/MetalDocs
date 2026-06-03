package presence

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// fakeRepo is a hand-rolled Repository fake. Tests drive state via
// SetSnapshot / RecordBump; reads/writes are mutex-guarded so the hub
// goroutine can interleave with assertions safely.
type fakeRepo struct {
	mu        sync.Mutex
	snapshots map[string][]Item
	bumps     map[string]time.Time
	tenants   []string
	failNext  error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		snapshots: make(map[string][]Item),
		bumps:     make(map[string]time.Time),
	}
}

func (r *fakeRepo) SetSnapshot(tenantID string, items []Item) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshots[tenantID] = append([]Item(nil), items...)
}

func (r *fakeRepo) Bumps() map[string]time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]time.Time, len(r.bumps))
	for k, v := range r.bumps {
		out[k] = v
	}
	return out
}

func (r *fakeRepo) BumpLastSeen(_ context.Context, userID string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failNext != nil {
		err := r.failNext
		r.failNext = nil
		return err
	}
	r.bumps[userID] = at
	return nil
}

func (r *fakeRepo) Snapshot(_ context.Context, tenantID string, now time.Time) ([]Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := append([]Item(nil), r.snapshots[tenantID]...)
	cutoff := now.Add(-OfflineAfter)
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if it.LastSeenAt.Before(cutoff) {
			continue
		}
		it.Status = ClassifyStatus(it.LastSeenAt, now)
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeenAt.After(out[j].LastSeenAt) })
	return out, nil
}

func (r *fakeRepo) TenantsWithRecentPresence(_ context.Context, since time.Time) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = since
	return append([]string(nil), r.tenants...), nil
}

// --- model classification ---------------------------------------------------

func TestClassifyStatus(t *testing.T) {
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		last time.Time
		want Status
	}{
		{"fresh", now.Add(-5 * time.Second), StatusOnline},
		{"edge_online", now.Add(-59 * time.Second), StatusOnline},
		{"edge_idle", now.Add(-IdleAfter), StatusIdle},
		{"deep_idle", now.Add(-3 * time.Minute), StatusIdle},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ClassifyStatus(c.last, now); got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

// --- hub diff ---------------------------------------------------------------

func TestDiff_Join_Leave_StatusTransitions(t *testing.T) {
	now := time.Now().UTC()
	prev := map[string]Item{
		"alice": {UserID: "alice", DisplayName: "Alice", LastSeenAt: now.Add(-10 * time.Second), Status: StatusOnline},
		"bob":   {UserID: "bob", DisplayName: "Bob", LastSeenAt: now.Add(-2 * time.Minute), Status: StatusIdle},
	}
	curr := map[string]Item{
		"alice": {UserID: "alice", DisplayName: "Alice", LastSeenAt: now.Add(-90 * time.Second), Status: StatusIdle},
		// bob departed
		"carol": {UserID: "carol", DisplayName: "Carol", LastSeenAt: now, Status: StatusOnline},
	}
	got := diff(prev, curr)

	types := make(map[string]string)
	for _, ev := range got {
		types[ev.UserID] = ev.Type
	}
	if types["alice"] != "idle" {
		t.Fatalf("alice: got %q want idle", types["alice"])
	}
	if types["bob"] != "leave" {
		t.Fatalf("bob: got %q want leave", types["bob"])
	}
	if types["carol"] != "join" {
		t.Fatalf("carol: got %q want join", types["carol"])
	}
	if len(got) != 3 {
		t.Fatalf("got %d events want 3", len(got))
	}
}

// --- snapshot HTTP fallback -------------------------------------------------

func TestHandler_Snapshot_TenantScoped(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.SetSnapshot("tenant-a", []Item{{UserID: "a1", DisplayName: "Alice", LastSeenAt: now}})
	repo.SetSnapshot("tenant-b", []Item{{UserID: "b1", DisplayName: "Bob", LastSeenAt: now}})

	hub := NewHub(repo, nil)
	h := NewHandler(hub, repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/presence/snapshot", nil).
		WithContext(tenant.WithTenantID(context.Background(), "tenant-a"))
	rec := httptest.NewRecorder()
	h.handleSnapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200; body=%s", rec.Code, rec.Body.String())
	}
	var got map[string][]Item
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if n := len(got["items"]); n != 1 || got["items"][0].UserID != "a1" {
		t.Fatalf("tenant scoping: got %+v want only tenant-a's user", got)
	}
}

func TestHandler_Snapshot_MissingTenantReturns500(t *testing.T) {
	repo := newFakeRepo()
	hub := NewHub(repo, nil)
	h := NewHandler(hub, repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/presence/snapshot", nil)
	rec := httptest.NewRecorder()
	h.handleSnapshot(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want 500", rec.Code)
	}
}

// --- middleware debounce ----------------------------------------------------

func TestBumpMiddleware_DebouncesPerUser(t *testing.T) {
	repo := newFakeRepo()
	mw := NewBumpMiddleware(repo, nil)

	clock := time.Now().UTC()
	mw.WithClock(func() time.Time { return clock })
	mw.WithDebounce(60 * time.Second)

	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called++ })
	h := mw.Wrap(next)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil).
			WithContext(iamdomain.WithAuthContext(context.Background(), "alice", nil))
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	// Bump runs on goroutine; wait briefly for it to land.
	if !waitForBump(repo, "alice", time.Second) {
		t.Fatalf("expected bump for alice within 1s")
	}
	if called != 5 {
		t.Fatalf("downstream called %d want 5", called)
	}

	// Advance past debounce; next call should bump again.
	mw.lastBump["alice"] = clock.Add(-2 * time.Minute)
	clock = clock.Add(time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil).
		WithContext(iamdomain.WithAuthContext(context.Background(), "alice", nil))
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !waitForBumpAt(repo, "alice", clock, time.Second) {
		t.Fatalf("expected updated bump at %v", clock)
	}
}

func TestBumpMiddleware_NoUserNoBump(t *testing.T) {
	repo := newFakeRepo()
	mw := NewBumpMiddleware(repo, nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := mw.Wrap(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(50 * time.Millisecond)
	if got := repo.Bumps(); len(got) != 0 {
		t.Fatalf("anonymous request bumped: %+v", got)
	}
}

// --- WS stream (integration via httptest.Server) ----------------------------

func TestStream_TenantIsolation(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.SetSnapshot("tenant-a", []Item{{UserID: "a1", DisplayName: "Alice", LastSeenAt: now}})
	repo.SetSnapshot("tenant-b", []Item{{UserID: "b1", DisplayName: "Bob", LastSeenAt: now}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub(repo, nil, WithTickInterval(50*time.Millisecond))
	go hub.Run(ctx)

	h := NewHandler(hub, repo, nil).
		WithAcceptOptions(&websocket.AcceptOptions{InsecureSkipVerify: true})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force a known tenant on the request so we exercise tenant.WithTenantID.
		tid := r.URL.Query().Get("tenant")
		if tid == "" {
			http.Error(w, "tenant required", http.StatusBadRequest)
			return
		}
		ctx := tenant.WithTenantID(r.Context(), tid)
		h.handleStream(w, r.WithContext(ctx))
	}))
	defer srv.Close()

	// Connect as tenant-a.
	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	dialCtx, dialCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dialCancel()
	c, _, err := websocket.Dial(dialCtx, wsURL+"?"+url.Values{"tenant": []string{"tenant-a"}}.Encode(), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	// Initial snapshot should contain only tenant-a's user.
	ev := readEvent(t, ctx, c)
	if ev.Type != "snapshot" {
		t.Fatalf("first frame: got %q want snapshot", ev.Type)
	}
	if len(ev.Presence) != 1 || ev.Presence[0].UserID != "a1" {
		t.Fatalf("snapshot leaked tenant data: %+v", ev.Presence)
	}

	// Mutate tenant-b — tenant-a connection MUST NOT receive any event.
	repo.SetSnapshot("tenant-b", []Item{
		{UserID: "b1", DisplayName: "Bob", LastSeenAt: now},
		{UserID: "b2", DisplayName: "Bob2", LastSeenAt: now},
	})

	// Wait long enough for several ticks. Any non-heartbeat frame is a leak.
	deadline := time.After(300 * time.Millisecond)
LOOP:
	for {
		select {
		case <-deadline:
			break LOOP
		default:
			readCtx, readCancel := context.WithTimeout(ctx, 80*time.Millisecond)
			ev, err := readEventNonFatal(readCtx, c)
			readCancel()
			if err != nil {
				continue
			}
			if ev.Type != "heartbeat" {
				if ev.UserID == "b1" || ev.UserID == "b2" {
					t.Fatalf("tenant-a connection received cross-tenant event: %+v", ev)
				}
			}
		}
	}
}

func TestStream_SnapshotMatchesHTTP(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.SetSnapshot("tenant-a", []Item{
		{UserID: "u1", DisplayName: "U1", LastSeenAt: now.Add(-5 * time.Second)},
		{UserID: "u2", DisplayName: "U2", LastSeenAt: now.Add(-2 * time.Minute)},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hub := NewHub(repo, nil, WithTickInterval(time.Hour))
	h := NewHandler(hub, repo, nil).
		WithAcceptOptions(&websocket.AcceptOptions{InsecureSkipVerify: true})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := tenant.WithTenantID(r.Context(), "tenant-a")
		h.handleStream(w, r.WithContext(ctx))
	}))
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	dialCtx, dialCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dialCancel()
	c, _, err := websocket.Dial(dialCtx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	ev := readEvent(t, ctx, c)
	if ev.Type != "snapshot" {
		t.Fatalf("got %q want snapshot", ev.Type)
	}

	// HTTP snapshot at the same instant must contain the same userIDs.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/presence/snapshot", nil).
		WithContext(tenant.WithTenantID(context.Background(), "tenant-a"))
	rec := httptest.NewRecorder()
	h.handleSnapshot(rec, req)
	var got map[string][]Item
	_ = json.Unmarshal(rec.Body.Bytes(), &got)

	if len(got["items"]) != len(ev.Presence) {
		t.Fatalf("size mismatch: ws=%d http=%d", len(ev.Presence), len(got["items"]))
	}
	wsIDs := userIDs(ev.Presence)
	httpIDs := userIDs(got["items"])
	if !sameSet(wsIDs, httpIDs) {
		t.Fatalf("id sets differ: ws=%v http=%v", wsIDs, httpIDs)
	}
}

// --- helpers ----------------------------------------------------------------

func readEvent(t *testing.T, ctx context.Context, c *websocket.Conn) Event {
	t.Helper()
	rctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, data, err := c.Read(rctx)
	if err != nil {
		t.Fatalf("ws read: %v", err)
	}
	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("decode ws frame: %v: %s", err, string(data))
	}
	return ev
}

func readEventNonFatal(ctx context.Context, c *websocket.Conn) (Event, error) {
	_, data, err := c.Read(ctx)
	if err != nil {
		return Event{}, err
	}
	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		return Event{}, err
	}
	return ev, nil
}

func userIDs(items []Item) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.UserID)
	}
	sort.Strings(out)
	return out
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func waitForBump(repo *fakeRepo, userID string, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if _, ok := repo.Bumps()[userID]; ok {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

func waitForBumpAt(repo *fakeRepo, userID string, at time.Time, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if got, ok := repo.Bumps()[userID]; ok && got.Equal(at) {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// ensure errors.Is on context cancel doesn't trip a vet warning when imported.
var _ = errors.Is
