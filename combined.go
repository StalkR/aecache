package aecache

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// A combinedCache represents the combination of multiple caches.
// It implements Cache interface.
type combinedCache []Cache

// newMemoryCache creates a new combinedCache.
func newCombinedCache(caches ...Cache) combinedCache {
	return combinedCache(caches)
}

// Set sets a key to a value with an expiration.
// It updates all caches from fastest to slowest.
func (a combinedCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	for _, e := range a {
		if err := e.Set(ctx, key, value, expiration); err != nil {
			return err
		}
	}
	return nil
}

// Get gets the value and expiration for a key.
// It looks through all the cache layers, from fastest to slowest.
// When found, a layer refreshes its parent caches.
func (a combinedCache) Get(ctx context.Context, key string) ([]byte, time.Time, error) {
	if len(a) == 0 {
		return nil, time.Time{}, ErrCacheMiss
	}
	value, expires, err := a[0].Get(ctx, key)
	if err == nil {
		return value, expires, nil
	}
	if err != ErrCacheMiss {
		return nil, time.Time{}, err
	}
	value, expires, err = combinedCache(a[1:]).Get(ctx, key)
	if err == nil {
		if err := a[0].Set(ctx, key, value, expires.Sub(time.Now())); err != nil {
			return nil, time.Time{}, err
		}
	}
	return value, expires, err
}

// Clean deletes expired items.
func (a combinedCache) Clean(ctx context.Context) error {
	var errors []string
	for _, e := range a {
		if err := e.Clean(ctx); err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("cache: %v error(s)\n%v", len(errors), strings.Join(errors, "\n"))
	}
	return nil
}
