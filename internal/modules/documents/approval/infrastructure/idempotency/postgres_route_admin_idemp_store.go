package idempotency

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/platform/idempotency"
)

const (
	routeAdminCreateRouteTemplate     = "POST /api/v1/approval/routes"
	routeAdminUpdateRouteTemplate     = "PUT /api/v1/approval/routes/{id}"
	routeAdminDeactivateRouteTemplate = "DELETE /api/v1/approval/routes/{id}"
)

// routeAdminReplayEnvelope is the JSON shape persisted in
// idempotency_keys.response_body for a completed route-admin slot.
type routeAdminReplayEnvelope struct {
	RouteID    string `json:"route_id"`
	NewVersion *int   `json:"new_version,omitempty"`
}

type routeAdminReplayHandle struct {
	store  *idempotency.Store
	handle *idempotency.ReplayHandle
}

func (h *routeAdminReplayHandle) Complete(routeID string, newVersion *int) error {
	if h == nil || h.store == nil || h.handle == nil {
		return errors.New("idempotency store not configured")
	}
	body, err := json.Marshal(routeAdminReplayEnvelope{RouteID: routeID, NewVersion: newVersion})
	if err != nil {
		return fmt.Errorf("route_admin idempotency: marshal replay response: %w", err)
	}
	return h.store.CompleteReplay(h.handle, 200, body)
}

func (h *routeAdminReplayHandle) Fail(cause error) error {
	if h == nil || h.store == nil || h.handle == nil {
		return nil
	}
	return h.store.FailReplay(h.handle, cause)
}

// PostgresRouteAdminIdempStore implements application.RouteAdminIdempStore.
type PostgresRouteAdminIdempStore struct {
	create     *idempotency.Store
	update     *idempotency.Store
	deactivate *idempotency.Store
}

// NewPostgresRouteAdminIdempStore constructs a PostgresRouteAdminIdempStore
// backed by db. Panics if db is nil — this store must always be wired at the
// composition root, unlike NewPostgresSignoffIdempStore's nil-tolerant path.
func NewPostgresRouteAdminIdempStore(db *sql.DB) *PostgresRouteAdminIdempStore {
	if db == nil {
		panic("postgres_route_admin_idemp_store: db is nil")
	}
	return &PostgresRouteAdminIdempStore{
		create:     idempotency.New(db, routeAdminCreateRouteTemplate),
		update:     idempotency.New(db, routeAdminUpdateRouteTemplate),
		deactivate: idempotency.New(db, routeAdminDeactivateRouteTemplate),
	}
}

// BeginCreateReplay claims or replays an idempotency slot for the create-route
// endpoint. See beginReplayRaw for the (handle, replay, err) return contract.
func (s *PostgresRouteAdminIdempStore) BeginCreateReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.RouteAdminReplayCommitter, *application.RouteAdminReplay, error) {
	return s.beginReplay(ctx, s.create, tenantID, actorID, idempKey, payloadHash)
}

// BeginUpdateReplay claims or replays an idempotency slot for the update-route
// endpoint. See beginReplayRaw for the (handle, replay, err) return contract.
func (s *PostgresRouteAdminIdempStore) BeginUpdateReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.RouteAdminReplayCommitter, *application.RouteAdminReplay, error) {
	return s.beginReplay(ctx, s.update, tenantID, actorID, idempKey, payloadHash)
}

// BeginDeactivateReplay claims or replays an idempotency slot for the
// deactivate-route endpoint. See beginReplayRaw for the (handle, replay, err) return contract.
func (s *PostgresRouteAdminIdempStore) BeginDeactivateReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.RouteAdminReplayCommitter, *application.RouteAdminReplay, error) {
	return s.beginReplay(ctx, s.deactivate, tenantID, actorID, idempKey, payloadHash)
}

func (s *PostgresRouteAdminIdempStore) beginReplay(ctx context.Context, store *idempotency.Store, tenantID, actorID, idempKey, payloadHash string) (application.RouteAdminReplayCommitter, *application.RouteAdminReplay, error) {
	handle, replay, err := beginReplayRaw(ctx, store, tenantID, actorID, idempKey, payloadHash)
	if err != nil || replay == nil {
		if handle == nil {
			return nil, nil, err
		}
		return &routeAdminReplayHandle{store: store, handle: handle}, nil, nil
	}
	var envelope routeAdminReplayEnvelope
	if err := json.Unmarshal(replay.Body, &envelope); err != nil {
		return nil, nil, err
	}
	return nil, &application.RouteAdminReplay{RouteID: envelope.RouteID, NewVersion: envelope.NewVersion}, nil
}
