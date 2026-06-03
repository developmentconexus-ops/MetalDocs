package presence

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// Handler owns the HTTP surface for PR-9 presence:
//
//	GET /api/v1/iam/presence/snapshot — JSON {items: []Item}
//	GET /api/v1/iam/presence/stream   — WS upgrade; tenant-scoped
//
// Both are CapUserView, both tenant-scoped via tenant.FromContext
// (which is populated by the auth middleware). The handler itself
// performs no authz beyond requiring a tenant — the middleware stack
// (newPermissionResolver in apps/api/cmd/metaldocs-api/permissions.go)
// has already enforced the capability.
type Handler struct {
	hub  *Hub
	repo Repository
	log  *slog.Logger

	// acceptOpts is consulted on every WS upgrade. Defaults are safe:
	// OriginPatterns is empty (origin check enforced by the upstream
	// originProtection middleware). InsecureSkipVerify is never set.
	acceptOpts *websocket.AcceptOptions
}

// NewHandler constructs the presence HTTP handler. repo is consulted
// for the snapshot endpoint and for the initial WS payload; hub
// drives the long-lived WS connection.
func NewHandler(hub *Hub, repo Repository, log *slog.Logger) *Handler {
	if hub == nil {
		panic("presence.NewHandler: hub is nil")
	}
	if repo == nil {
		panic("presence.NewHandler: repo is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &Handler{hub: hub, repo: repo, log: log}
}

// WithAcceptOptions overrides the websocket.AcceptOptions used on
// upgrade. The default is zero-value; production callers should rely
// on the upstream originProtection middleware rather than embedding
// origin patterns here.
func (h *Handler) WithAcceptOptions(opts *websocket.AcceptOptions) *Handler {
	h.acceptOpts = opts
	return h
}

// RegisterRoutes mounts the snapshot and stream endpoints on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/presence/snapshot", h.handleSnapshot)
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/presence/stream", h.handleStream)
}

func (h *Handler) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "tenant context missing"))
		return
	}
	items, err := h.repo.Snapshot(r.Context(), tenantID, time.Now().UTC())
	if err != nil {
		h.log.Warn("presence: snapshot failed", "tenant_id", tenantID, "err", err)
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "snapshot failed"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "tenant context missing"))
		return
	}

	opts := h.acceptOpts
	if opts == nil {
		opts = &websocket.AcceptOptions{}
	}
	c, err := websocket.Accept(w, r, opts)
	if err != nil {
		h.log.Warn("presence: ws accept failed", "tenant_id", tenantID, "err", err)
		return
	}
	// Detach from the request context so server shutdown can drive the
	// WS lifetime explicitly rather than being torn down mid-write.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer c.Close(websocket.StatusNormalClosure, "")

	// Initial snapshot — produced and sent BEFORE Subscribe seeds the
	// hub room so the wire-side snapshot is guaranteed to match what
	// the room treats as the "starting state" for diff computation.
	items, err := h.repo.Snapshot(ctx, tenantID, time.Now().UTC())
	if err != nil {
		h.log.Warn("presence: ws initial snapshot failed", "tenant_id", tenantID, "err", err)
		c.Close(websocket.StatusInternalError, "snapshot failed")
		return
	}
	initial := Event{Type: "snapshot", Presence: items}
	if err := writeJSON(ctx, c, initial); err != nil {
		return
	}

	conn := h.hub.Subscribe(tenantID, items)
	defer conn.Close()

	// Reader: discards client frames (presence stream is one-way for
	// now) but enables ping-induced read deadlines via SetReadLimit +
	// the inactivity timer below. Errors close the connection.
	go func() {
		defer cancel()
		c.SetReadLimit(1024)
		for {
			if _, _, rerr := c.Read(ctx); rerr != nil {
				return
			}
		}
	}()

	idleTimer := time.NewTimer(ClientIdleTimeout)
	defer idleTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-conn.Out():
			if !ok {
				return
			}
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(ClientIdleTimeout)
			if err := writeJSON(ctx, c, ev); err != nil {
				return
			}
		case <-idleTimer.C:
			c.Close(websocket.StatusPolicyViolation, "client idle")
			return
		}
	}
}

func writeJSON(ctx context.Context, c *websocket.Conn, ev Event) error {
	wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	w, err := c.Writer(wctx, websocket.MessageText)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return err
	}
	if err := json.NewEncoder(w).Encode(ev); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
