/*
Package aecache implements layers of caching for App Engine.

It provides 2 layers of caching:
 - process memory
 - cloud datastore
And combines them to provide a layered cache, fastest to slowest.

It only caches bytes, on top of which you can use encodings (json, gob, etc).
Faster cache layers are refilled automatically from lower ones when hit.
Values expires strictly after provided expiration.
*/
package aecache

import (
	"context"
	"errors"
	"time"
)

// Errors returned by this package.
var (
	// ErrCacheMiss is when key is not found.
	ErrCacheMiss = errors.New("cache: miss")
	// ErrTooBig is when a value is too big to fit in the cache.
	ErrTooBig = errors.New("cache: too big")
)

// cacher represents the ability to set/get values and clean.
type cacher interface {
	// Set sets a key to a value with an expiration.
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	// Get gets the value and expiration for a key.
	Get(ctx context.Context, key string) ([]byte, time.Time, error)
	// Clean deletes expired items.
	Clean(ctx context.Context) error
}

// defaultCache is the default layered cache (process memory, cloud datastore).
var defaultCache = newCombinedCache(newMemoryCache(), newDatastoreCache())

// Set sets a key to a value with an expiration.
func Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return defaultCache.Set(ctx, key, value, expiration)
}

// Get gets the value and expiration for a key.
func Get(ctx context.Context, key string) ([]byte, time.Time, error) {
	return defaultCache.Get(ctx, key)
}

// Clean deletes expired items.
func Clean(ctx context.Context) error {
	return defaultCache.Clean(ctx)
}
