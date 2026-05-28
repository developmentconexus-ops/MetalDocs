package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
)

type ErrCapDenied struct {
	Capability string
	AreaCode   string
	ActorID    string
}

func (e ErrCapDenied) Error() string {
	return fmt.Sprintf("authz: capability %q denied for actor %q in area %q", e.Capability, e.ActorID, e.AreaCode)
}

type capCacheKey struct{}

type capCache struct {
	mu              sync.Mutex
	granted         map[string]bool
	assertedByTxPtr map[*sql.Tx]*assertedCache
}

type assertedCache struct {
	items []map[string]string
	set   map[string]struct{}
}

func WithCapCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, capCacheKey{}, &capCache{
		granted:         make(map[string]bool),
		assertedByTxPtr: make(map[*sql.Tx]*assertedCache),
	})
}

// Require enforces a tier-2 area-scoped authz check inside the caller's transaction.
// See wiki/decisions/0007-two-tier-authz.md for the boundary between tier-1
// (CapabilityService.CanDo, HTTP middleware) and tier-2 (this function).
//
// Callers MUST set the metaldocs.actor_id and metaldocs.tenant_id GUCs on the
// transaction before invoking. Use MustActorID/MustTenantID helpers if reading
// them from elsewhere.
//
// Pass areaCode = "tenant" to skip the area filter (degenerates to a tier-1
// equivalent inside the transaction).
func Require(ctx context.Context, tx *sql.Tx, capability, areaCode string) error {
	actorID, err := MustActorID(ctx, tx)
	if err != nil {
		return err
	}
	tenantID, err := MustTenantID(ctx, tx)
	if err != nil {
		return err
	}
	if cacheGranted(ctx, tx, actorID, tenantID, capability, areaCode) {
		return appendAssertedCap(ctx, tx, capability, areaCode)
	}

	// system_admin bypass — check before capability query
	var isAdmin bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
)`, actorID, tenantID).Scan(&isAdmin); err != nil {
		return fmt.Errorf("authz: system_admin check: %w", err)
	}
	if isAdmin {
		storeGranted(ctx, tx, actorID, tenantID, capability, areaCode)
		return appendAssertedCap(ctx, tx, capability, areaCode)
	}

	var granted bool
	err = tx.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)`, capability, areaCode, actorID, tenantID).Scan(&granted)
	if err != nil {
		return err
	}

	if !granted {
		return ErrCapDenied{Capability: capability, AreaCode: areaCode, ActorID: actorID}
	}

	storeGranted(ctx, tx, actorID, tenantID, capability, areaCode)
	return appendAssertedCap(ctx, tx, capability, areaCode)
}

func BypassSystem(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)")
	return err
}

func cacheGranted(ctx context.Context, tx *sql.Tx, actorID, tenantID, capability, areaCode string) bool {
	cache, ok := ctx.Value(capCacheKey{}).(*capCache)
	if !ok || cache == nil {
		return false
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	return cache.granted[grantCacheKey(tx, actorID, tenantID, capability, areaCode)]
}

func storeGranted(ctx context.Context, tx *sql.Tx, actorID, tenantID, capability, areaCode string) {
	cache, ok := ctx.Value(capCacheKey{}).(*capCache)
	if !ok || cache == nil {
		return
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.granted[grantCacheKey(tx, actorID, tenantID, capability, areaCode)] = true
}

func cacheKey(capability, areaCode string) string {
	return capability + "\x00" + areaCode
}

func grantCacheKey(tx *sql.Tx, actorID, tenantID, capability, areaCode string) string {
	return fmt.Sprintf("%p\x00%s\x00%s\x00%s", tx, actorID, tenantID, cacheKey(capability, areaCode))
}

func appendAssertedCap(ctx context.Context, tx *sql.Tx, capability, areaCode string) error {
	cache, _ := ctx.Value(capCacheKey{}).(*capCache)
	var asserted *assertedCache
	if cache != nil {
		cache.mu.Lock()
		asserted = cache.assertedByTxPtr[tx]
		cache.mu.Unlock()
	}

	if asserted == nil {
		loaded, err := loadAssertedCaps(ctx, tx)
		if err != nil {
			return err
		}
		asserted = loaded
		if cache != nil {
			cache.mu.Lock()
			cache.assertedByTxPtr[tx] = asserted
			cache.mu.Unlock()
		}
	}

	compoundKey := cacheKey(capability, areaCode)
	if _, exists := asserted.set[compoundKey]; exists {
		return nil
	}
	// Safe to mutate asserted without cache.mu: database/sql guarantees *sql.Tx
	// is not safe for concurrent use, so only one goroutine owns this tx at a time.
	asserted.items = append(asserted.items, map[string]string{"cap": capability, "area": areaCode})
	asserted.set[compoundKey] = struct{}{}

	encoded, err := json.Marshal(asserted.items)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "SELECT set_config('metaldocs.asserted_caps', $1, true)", string(encoded))
	return err
}

func loadAssertedCaps(ctx context.Context, tx *sql.Tx) (*assertedCache, error) {
	var raw sql.NullString
	if err := tx.QueryRowContext(ctx, "SELECT current_setting('metaldocs.asserted_caps', true)").Scan(&raw); err != nil {
		return nil, err
	}

	items := make([]map[string]string, 0, 4)
	if raw.Valid && raw.String != "" {
		if err := json.Unmarshal([]byte(raw.String), &items); err != nil {
			return nil, err
		}
	}
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[cacheKey(item["cap"], item["area"])] = struct{}{}
	}
	return &assertedCache{items: items, set: set}, nil
}
