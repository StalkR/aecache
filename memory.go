package aecache

import (
	"context"
	"sync"
	"time"
)

// A memoryCache represents a cache in the process memory.
// It implements Cache interfaces.
type memoryCache struct {
	m       sync.Mutex // protects below
	values  map[string][]byte
	expires map[string]time.Time
}

// newMemoryCache creates a new memoryCache.
func newMemoryCache() *memoryCache {
	return &memoryCache{
		values:  make(map[string][]byte),
		expires: make(map[string]time.Time),
	}
}

// Set sets a key to a value with an expiration.
func (a *memoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if expiration <= 0 {
		return nil
	}
	a.m.Lock()
	defer a.m.Unlock()
	a.values[key] = value
	a.expires[key] = time.Now().Add(expiration)
	return nil
}

// Get gets the value and expiration for a key.
func (a *memoryCache) Get(ctx context.Context, key string) ([]byte, time.Time, error) {
	a.m.Lock()
	defer a.m.Unlock()
	value, okv := a.values[key]
	expires, oke := a.expires[key]
	if !okv || !oke {
		return nil, time.Time{}, ErrCacheMiss
	}
	if expires.Before(time.Now()) {
		delete(a.values, key)
		delete(a.expires, key)
		return nil, time.Time{}, ErrCacheMiss
	}
	return value, expires, nil
}

// Clean deletes expired items.
func (a *memoryCache) Clean(ctx context.Context) error {
	a.m.Lock()
	defer a.m.Unlock()
	for key, expires := range a.expires {
		if expires.Before(time.Now()) {
			delete(a.values, key)
			delete(a.expires, key)
		}
	}
	return nil
}
