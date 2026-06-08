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
type CachedRoleProvider struct {
	base domain.RoleProvider
	ttl  time.Duration
	mu   sync.RWMutex
	// Expired entries are swept by the background goroutine in NewCachedRoleProvider.
	// The cache has no max-size cap, so a very large distinct (user, tenant) working
	// set could grow it between sweeps; acceptable at current scale.
	items map[string]cacheEntry
}

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
