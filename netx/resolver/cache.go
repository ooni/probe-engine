package resolver

import (
	"context"
	"sync"
)

// CacheResolver is a resolver that caches successful replies.
type CacheResolver struct {
	Resolver
	mu    sync.Mutex
	cache map[string][]string
}

// LookupHost implements Resolver.LookupHost
func (r *CacheResolver) LookupHost(
	ctx context.Context, hostname string) ([]string, error) {
	r.mu.Lock()
	if r.cache == nil {
		r.cache = make(map[string][]string)
	}
	entry := r.cache[hostname]
	r.mu.Unlock()
	if entry != nil {
		return entry, nil
	}
	entry, err := r.Resolver.LookupHost(ctx, hostname)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.cache[hostname] = entry
	r.mu.Unlock()
	return entry, nil
}
