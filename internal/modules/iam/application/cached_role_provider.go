package application

import (
	"context"
	"sync"
	"time"

	"metaldocs/internal/modules/iam/domain"
)

type cacheEntry struct {
	roles     []domain.Role
	expiresAt time.Time
}

// CachedRoleProvider wraps a RoleProvider with TTL cache and explicit invalidation.
//
// CacheContract (REQ-CACHE-1, RF-3):
//
//	Source of truth: postgres RoleProvider. Cache-aside, per (user,tenant) key.
//	TTL: 30 seconds (default in NewCachedRoleProvider; callers may pass a shorter
//	  value). Staleness bound: max(TTL) — a grant/revoke is visible no later than
//	  30 s even if an explicit invalidation is missed.
//	Invalidation: InvalidateUserTenant, called by:
//	  - AdminService.UpsertUserAndAssignRole (admin_service.go)
//	  - AdminService.ReplaceUserRoles (admin_service.go)
//	  - AreaMembershipService.Grant — insert path (area_membership_service.go)
//	  - AreaMembershipService.Grant — atomic role-change path (area_membership_service.go)
//	  - AreaMembershipService.Revoke (area_membership_service.go)
//	  - PeopleService.Invite — defensive post-create flush (people_service.go)
//	  - PeopleService.PatchAtomic — on TenantRole field change (people_service.go)
//	Failure mode: cache errors are not possible (in-memory map); source errors fall
//	  through to the caller. Tenant isolation is enforced by key: roleCacheKey embeds
//	  both userID and tenantID, so a cache hit can never serve data from a foreign tenant.
//	Eviction: background ticker goroutine (period = TTL) sweeps expired entries.
//	  TTL-based expiry also evicts on the read path before returning a stale hit.
//	Redis: no Redis cache for this provider — all caching is in-process.
type CachedRoleProvider struct {
	base domain.RoleProvider
	ttl  time.Duration
	mu   sync.RWMutex
	// Expired entries are swept by the background goroutine in NewCachedRoleProvider.
	// The cache has no max-size cap, so a very large distinct (user, tenant) working
	// set could grow it between sweeps; acceptable at current scale.
	items map[string]cacheEntry
}

// NewCachedRoleProvider constructs the cache-aside wrapper and starts a
// background goroutine that sweeps expired entries every ttl, stopping when
// ctx is done. ttl <= 0 defaults to 30 seconds. See CacheContract above for
// the invalidation obligations callers must honor.
func NewCachedRoleProvider(ctx context.Context, base domain.RoleProvider, ttl time.Duration) *CachedRoleProvider {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	provider := &CachedRoleProvider{
		base:  base,
		ttl:   ttl,
		items: map[string]cacheEntry{},
	}
	go func() {
		ticker := time.NewTicker(ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				cutoff := now.UTC()
				provider.mu.Lock()
				for key, entry := range provider.items {
					if !cutoff.Before(entry.expiresAt) {
						delete(provider.items, key)
					}
				}
				provider.mu.Unlock()
			}
		}
	}()
	return provider
}

func roleCacheKey(userID, tenantID string) string {
	return userID + "|" + tenantID
}

// RolesByUserID returns the cached role set for (userID, tenantID) when the
// entry is present and unexpired, otherwise fetches from the base provider
// and populates the cache before returning.
func (c *CachedRoleProvider) RolesByUserID(ctx context.Context, userID, tenantID string) ([]domain.Role, error) {
	key := roleCacheKey(userID, tenantID)
	now := time.Now().UTC()

	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	if ok && now.Before(entry.expiresAt) {
		return cloneRoles(entry.roles), nil
	}

	roles, err := c.base.RolesByUserID(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	// All current role-write paths invalidate explicitly: AdminService
	// (upsert/replace), PeopleService.Invite (create), and AreaMembershipService
	// (grant/revoke). If group-membership mutation routes are ever added they must
	// call InvalidateUserTenant too, or stale roles persist until the TTL.
	c.items[key] = cacheEntry{roles: cloneRoles(roles), expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()

	return roles, nil
}

// RolesByUserIDs resolves roles for multiple users with read-through cache.
// Per-user cache hits are served without a DB round trip; misses are fetched in
// a single batch query, then each miss result is stored under its TTL.
func (c *CachedRoleProvider) RolesByUserIDs(ctx context.Context, tenantID string, userIDs []string) (map[string][]domain.Role, error) {
	if len(userIDs) == 0 {
		return map[string][]domain.Role{}, nil
	}
	now := time.Now().UTC()
	out := make(map[string][]domain.Role, len(userIDs))
	var misses []string

	c.mu.RLock()
	for _, uid := range userIDs {
		key := roleCacheKey(uid, tenantID)
		if entry, ok := c.items[key]; ok && now.Before(entry.expiresAt) {
			out[uid] = cloneRoles(entry.roles)
		} else {
			misses = append(misses, uid)
		}
	}
	c.mu.RUnlock()

	if len(misses) == 0 {
		return out, nil
	}

	fetched, err := c.base.RolesByUserIDs(ctx, tenantID, misses)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	for _, uid := range misses {
		roles, active := fetched[uid]
		if !active {
			// User not found / inactive — do not cache negative; simply omit.
			continue
		}
		key := roleCacheKey(uid, tenantID)
		c.items[key] = cacheEntry{roles: cloneRoles(roles), expiresAt: now.Add(c.ttl)}
		out[uid] = cloneRoles(roles)
	}
	c.mu.Unlock()

	return out, nil
}

// UserActiveInTenant delegates to the base provider. Not cached: the answer
// is deactivation-sensitive and must be authoritative on every call.
func (c *CachedRoleProvider) UserActiveInTenant(ctx context.Context, tenantID, userID string) (bool, error) {
	return c.base.UserActiveInTenant(ctx, tenantID, userID)
}

// InvalidateUserTenant evicts the cached role set for (userID, tenantID).
// MUST be called by any admin write that changes a user's roles or tenant
// membership (see CacheContract callers list above); otherwise stale roles
// continue to authorize until the TTL expires.
func (c *CachedRoleProvider) InvalidateUserTenant(userID, tenantID string) {
	c.evict(userID, tenantID)
}

func (c *CachedRoleProvider) evict(userID, tenantID string) {
	c.mu.Lock()
	delete(c.items, roleCacheKey(userID, tenantID))
	c.mu.Unlock()
}

func cloneRoles(in []domain.Role) []domain.Role {
	out := make([]domain.Role, len(in))
	copy(out, in)
	return out
}
