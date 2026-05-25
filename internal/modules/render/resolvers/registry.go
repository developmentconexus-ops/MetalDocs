package resolvers

import "sync"

type Registry struct {
	mu    sync.RWMutex
	items map[string]ComputedResolver
}

func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]ComputedResolver),
	}
}

func (r *Registry) Register(cr ComputedResolver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[cr.Key()] = cr
}

func (r *Registry) Get(key string) (ComputedResolver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cr, ok := r.items[key]
	return cr, ok
}

func (r *Registry) Known() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]int, len(r.items))
	for key, cr := range r.items {
		out[key] = cr.Version()
	}
	return out
}
