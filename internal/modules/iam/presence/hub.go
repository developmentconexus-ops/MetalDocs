package presence

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Hub owns the in-process registry of tenant-scoped presence rooms and
// the single background ticker that polls iam_users.last_seen_at and
// emits join/leave/idle/online deltas to every Conn in the matching
// room. The hub is single-instance only — multi-instance scaling via
// Redis pub/sub is deferred (PR-9 doc § "Out of scope").
type Hub struct {
	repo  Repository
	clock func() time.Time
	log   *slog.Logger
	tick  time.Duration

	// outBuffer is the per-Conn outbound channel capacity. Slow / stuck
	// clients are dropped once their queue fills; we never block the
	// hub on a single misbehaving consumer.
	outBuffer int

	mu    sync.Mutex
	rooms map[string]*room

	nextConnID uint64

	started chan struct{}
	stopped chan struct{}
}

type room struct {
	tenantID string

	mu    sync.Mutex
	conns map[*Conn]struct{}
	prev  map[string]Item // userID -> last broadcast item; nil before first tick
}

// Conn is the hub-side handle on one WebSocket connection. The HTTP
// handler reads outbound events from Out(), forwards them to the
// socket, and calls Close on disconnect.
type Conn struct {
	id       uint64
	tenantID string
	hub      *Hub
	room     *room

	out chan Event

	closeOnce sync.Once
	closed    chan struct{}
}

// ID returns the monotonically-assigned connection id; used for logs.
func (c *Conn) ID() uint64 { return c.id }

// TenantID returns the tenant this connection is joined to. Useful
// for log lines; the hub itself never broadcasts across tenants.
func (c *Conn) TenantID() string { return c.tenantID }

// Out yields the channel the handler pumps to the WS socket. The
// channel is closed exactly once, after the Conn has been removed
// from its room — handlers may safely range over it.
func (c *Conn) Out() <-chan Event { return c.out }

// Closed fires when the Conn is fully torn down.
func (c *Conn) Closed() <-chan struct{} { return c.closed }

// Close removes the Conn from its room and closes the outbound chan.
// Idempotent.
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		c.room.removeConn(c)
		close(c.out)
		close(c.closed)
	})
}

// HubOption is the functional-options shape; reserved for future
// tuning. None defined yet.
type HubOption func(*Hub)

// WithClock overrides the wall clock; used by tests.
func WithClock(fn func() time.Time) HubOption {
	return func(h *Hub) { h.clock = fn }
}

// WithTickInterval overrides the room-tick cadence; used by tests
// that want sub-second ticks.
func WithTickInterval(d time.Duration) HubOption {
	return func(h *Hub) { h.tick = d }
}

// WithOutboundBuffer overrides the per-Conn outbound channel buffer.
func WithOutboundBuffer(n int) HubOption {
	return func(h *Hub) { h.outBuffer = n }
}

// NewHub constructs a hub. Run must be called separately to start the
// ticker goroutine; tests that drive ticks manually can skip Run.
func NewHub(repo Repository, log *slog.Logger, opts ...HubOption) *Hub {
	if repo == nil {
		panic("presence.NewHub: repo is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	h := &Hub{
		repo:      repo,
		clock:     func() time.Time { return time.Now().UTC() },
		log:       log,
		tick:      TickInterval,
		outBuffer: 32,
		rooms:     make(map[string]*room),
		started:   make(chan struct{}),
		stopped:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Run blocks until ctx is cancelled, ticking each room at the
// configured interval. Returns when the goroutine exits so callers
// can wait via a WaitGroup.
func (h *Hub) Run(ctx context.Context) {
	close(h.started)
	defer close(h.stopped)
	t := time.NewTicker(h.tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.tickAll(ctx)
		}
	}
}

// Subscribe registers a new connection for tenantID. The returned Conn
// has not yet received any events — the caller must immediately
// produce and send the initial snapshot (the hub does not double-send
// on Subscribe so the snapshot the caller wrote to the wire is
// guaranteed identical to room.prev after seeding).
//
// seedItems is the snapshot the caller just sent on the wire. The hub
// uses it to initialise room.prev so the first tick produces only
// genuine deltas, not a flood of "join" events.
func (h *Hub) Subscribe(tenantID string, seedItems []Item) *Conn {
	h.mu.Lock()
	r, ok := h.rooms[tenantID]
	if !ok {
		r = &room{
			tenantID: tenantID,
			conns:    make(map[*Conn]struct{}),
		}
		h.rooms[tenantID] = r
	}
	h.mu.Unlock()

	c := &Conn{
		id:       atomic.AddUint64(&h.nextConnID, 1),
		tenantID: tenantID,
		hub:      h,
		room:     r,
		out:      make(chan Event, h.outBuffer),
		closed:   make(chan struct{}),
	}

	r.mu.Lock()
	r.conns[c] = struct{}{}
	// Seed prev once per room; later subscribers do not overwrite it
	// (the hub-driven tick has already advanced room state).
	if r.prev == nil {
		r.prev = itemMap(seedItems)
	}
	r.mu.Unlock()
	return c
}

// tickAll iterates rooms, runs the snapshot query for each, computes
// deltas, and dispatches them. Rooms with no remaining connections
// are dropped here.
func (h *Hub) tickAll(ctx context.Context) {
	h.mu.Lock()
	rooms := make([]*room, 0, len(h.rooms))
	for _, r := range h.rooms {
		rooms = append(rooms, r)
	}
	h.mu.Unlock()

	for _, r := range rooms {
		h.tickRoom(ctx, r)
	}
}

func (h *Hub) tickRoom(ctx context.Context, r *room) {
	r.mu.Lock()
	if len(r.conns) == 0 {
		r.mu.Unlock()
		h.mu.Lock()
		// Re-check under hub lock: a Subscribe could have raced in.
		if existing, ok := h.rooms[r.tenantID]; ok && existing == r && len(existing.conns) == 0 {
			delete(h.rooms, r.tenantID)
		}
		h.mu.Unlock()
		return
	}
	r.mu.Unlock()

	now := h.clock()
	current, err := h.repo.Snapshot(ctx, r.tenantID, now)
	if err != nil {
		h.log.Warn("presence: snapshot failed", "tenant_id", r.tenantID, "err", err)
		return
	}

	events := diff(r.prevSnapshot(), itemMap(current))
	if len(events) == 0 {
		// Still update prev so transient timestamp jitter does not
		// re-emit the same state next tick. itemMap on equal sets
		// returns equal maps; the diff is empty so the comparison
		// stays cheap.
		r.setPrev(itemMap(current))
		return
	}
	r.setPrev(itemMap(current))
	r.broadcast(events, h)
}

// Heartbeat fans out a single {type:"heartbeat"} event to every conn
// in every room. Called by the hub's own heartbeat goroutine (started
// by RunHeartbeat) so each client gets a server-side liveness ping
// even if presence is steady.
func (h *Hub) Heartbeat() {
	h.mu.Lock()
	rooms := make([]*room, 0, len(h.rooms))
	for _, r := range h.rooms {
		rooms = append(rooms, r)
	}
	h.mu.Unlock()
	for _, r := range rooms {
		r.broadcast([]Event{{Type: "heartbeat"}}, h)
	}
}

// RunHeartbeat blocks until ctx is cancelled, emitting Heartbeat()
// every HeartbeatInterval. Intended to run as its own goroutine
// alongside Run.
func (h *Hub) RunHeartbeat(ctx context.Context) {
	t := time.NewTicker(HeartbeatInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.Heartbeat()
		}
	}
}

// removeConn detaches c from its room. Called by Conn.Close.
func (r *room) removeConn(c *Conn) {
	r.mu.Lock()
	delete(r.conns, c)
	r.mu.Unlock()
}

func (r *room) prevSnapshot() map[string]Item {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.prev == nil {
		return map[string]Item{}
	}
	out := make(map[string]Item, len(r.prev))
	for k, v := range r.prev {
		out[k] = v
	}
	return out
}

func (r *room) setPrev(next map[string]Item) {
	r.mu.Lock()
	r.prev = next
	r.mu.Unlock()
}

// broadcast pushes events to every Conn in the room. If a connection's
// outbound buffer is full, the connection is closed — slow consumers
// must not stall presence for every other client.
func (r *room) broadcast(events []Event, h *Hub) {
	r.mu.Lock()
	conns := make([]*Conn, 0, len(r.conns))
	for c := range r.conns {
		conns = append(conns, c)
	}
	r.mu.Unlock()

	for _, c := range conns {
		for _, ev := range events {
			select {
			case c.out <- ev:
			default:
				h.log.Warn("presence: dropping slow conn", "conn_id", c.id, "tenant_id", c.tenantID)
				go c.Close()
			}
		}
	}
}

// itemMap converts a snapshot slice into a userID-keyed map.
func itemMap(items []Item) map[string]Item {
	m := make(map[string]Item, len(items))
	for _, it := range items {
		m[it.UserID] = it
	}
	return m
}

// diff computes the events needed to transition from prev to curr.
// Emits "join" for new users, "leave" for departed users, and
// "online"/"idle" for status transitions on stable users.
// last_seen_at jitter alone does not emit an event.
func diff(prev, curr map[string]Item) []Event {
	out := make([]Event, 0)
	for uid, it := range curr {
		old, existed := prev[uid]
		if !existed {
			out = append(out, Event{
				Type:        "join",
				UserID:      it.UserID,
				DisplayName: it.DisplayName,
				LastSeenAt:  it.LastSeenAt,
				Status:      it.Status,
			})
			continue
		}
		if old.Status != it.Status {
			out = append(out, Event{
				Type:        string(it.Status),
				UserID:      it.UserID,
				DisplayName: it.DisplayName,
				LastSeenAt:  it.LastSeenAt,
				Status:      it.Status,
			})
		}
	}
	for uid, old := range prev {
		if _, ok := curr[uid]; !ok {
			out = append(out, Event{
				Type:        "leave",
				UserID:      old.UserID,
				DisplayName: old.DisplayName,
				LastSeenAt:  old.LastSeenAt,
			})
		}
	}
	return out
}
