package application

import (
	"context"
	"strings"
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
	base  domain.RoleProvider
	ttl   time.Duration
	mu    sync.RWMutex
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
	c.items[key] = cacheEntry{roles: cloneRoles(roles), expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()

	return roles, nil
}

// InvalidateUser invalidates cache entries for a user across all tenants.
func (c *CachedRoleProvider) InvalidateUser(userID string) {
	c.mu.Lock()
	for k := range c.items {
		if strings.HasPrefix(k, userID+"|") {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

func (c *CachedRoleProvider) InvalidateUserTenant(userID, tenantID string) {
	c.evict(userID, tenantID)
}

func (c *CachedRoleProvider) InvalidateAll() {
	c.mu.Lock()
	c.items = map[string]cacheEntry{}
	c.mu.Unlock()
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
