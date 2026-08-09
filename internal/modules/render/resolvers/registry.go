package resolvers

import "sync"

// Registry is a concurrency-safe lookup of ComputedResolver implementations
// keyed by their Key().
type Registry struct {
	mu    sync.RWMutex
	items map[string]ComputedResolver
}

// NewRegistry builds an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]ComputedResolver),
	}
}

// Register adds cr to the registry under cr.Key(), replacing any existing
// resolver registered under the same key.
func (r *Registry) Register(cr ComputedResolver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[cr.Key()] = cr
}

// Get looks up the resolver registered under key.
func (r *Registry) Get(key string) (ComputedResolver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cr, ok := r.items[key]
	return cr, ok
}

// Known returns every registered resolver key mapped to its Version().
func (r *Registry) Known() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]int, len(r.items))
	for key, cr := range r.items {
		out[key] = cr.Version()
	}
	return out
}
