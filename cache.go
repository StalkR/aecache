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

// Cache represents the ability to set/get values but also clean.
type Cache interface {
	// Set sets a key to a value with an expiration.
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	// Get gets the value and expiration for a key.
	Get(ctx context.Context, key string) ([]byte, time.Time, error)
	// Clean deletes expired items.
	Clean(ctx context.Context) error
}

// New creates a new layered cache (process memory, cloud datastore).
// ProjectID is necessary for Cloud Datastore, see datastore.NewClient.
// https://godoc.org/cloud.google.com/go/datastore#NewClient
func New(projectID string) Cache {
	return newCombinedCache(newMemoryCache(), newDatastoreCache(projectID))
}
