// Package presence implements the tenant-scoped WebSocket presence stream
// + HTTP snapshot fallback that the IAM Admin Center consumes (PR-9 of the
// IAM Admin Center rebuild). Presence is derived from iam_users.last_seen_at,
// which is bumped on authenticated requests (debounced) and on WS heartbeats.
//
// Status classification:
//
//	online — last_seen_at >= now - 60s
//	idle   — last_seen_at in [now-5m, now-60s)
//	off    — last_seen_at <  now-5m (excluded from snapshot)
package presence

import "time"

// IdleAfter is the duration since last_seen_at after which a user is
// classified "idle" rather than "online".
const IdleAfter = 60 * time.Second

// OfflineAfter is the duration since last_seen_at after which a user
// disappears from the presence snapshot entirely.
const OfflineAfter = 5 * time.Minute

// BumpDebounce is the minimum interval between two consecutive
// iam_users.last_seen_at updates for the same user. The middleware
// short-circuits cheaper updates inside this window to avoid hammering
// the row on every request.
const BumpDebounce = 60 * time.Second

// HeartbeatInterval is how often the server emits {"type":"heartbeat"}
// on each WS connection. Clients are expected to send their own ping
// at the same cadence; the server idles a connection out after
// ClientIdleTimeout without traffic.
const HeartbeatInterval = 30 * time.Second

// ClientIdleTimeout is the maximum time a WS connection may be silent
// before the server closes it.
const ClientIdleTimeout = 90 * time.Second

// TickInterval is the cadence at which each tenant room polls the
// presence query and diffs against the previous snapshot to emit
// join/leave/online/idle deltas. 15s is sufficient for tenant-admin
// observability — sub-second presence is out of scope.
const TickInterval = 15 * time.Second

// Status is the presence status reported on the wire.
type Status string

const (
	StatusOnline Status = "online"
	StatusIdle   Status = "idle"
)

// Item is one entry in a presence snapshot.
type Item struct {
	UserID      string    `json:"userId"`
	DisplayName string    `json:"displayName"`
	LastSeenAt  time.Time `json:"lastSeenAt"`
	Status      Status    `json:"status"`
}

// Event is the envelope emitted on the WebSocket. Fields outside the
// type's payload remain zero-valued and are omitted by the JSON encoder.
type Event struct {
	Type        string    `json:"type"`
	UserID      string    `json:"userId,omitempty"`
	DisplayName string    `json:"displayName,omitempty"`
	LastSeenAt  time.Time `json:"lastSeenAt,omitempty"`
	Status      Status    `json:"status,omitempty"`
	Presence    []Item    `json:"presence,omitempty"`
}

// ClassifyStatus returns StatusOnline / StatusIdle for the supplied
// last_seen_at relative to now. Callers must filter out users older
// than OfflineAfter before calling this — Classify does not encode a
// third state.
func ClassifyStatus(lastSeen, now time.Time) Status {
	if now.Sub(lastSeen) < IdleAfter {
		return StatusOnline
	}
	return StatusIdle
}
