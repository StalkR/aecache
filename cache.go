/*
Package aecache implements layers of caching for AppEngine.

It provides three layers of caching:
 - in-app cache
 - memcache
 - datastore
And combines them to provide a layered cache, fastest to slowest.

It only supports caching slice of bytes, to cache rich data use encoding
like json or gob.
*/
package aecache

import (
	"errors"
	"time"

	"golang.org/x/net/context"
)

// An Item represents a value to cache with its expiration time.
// It is mostly for internal and exported for memcache/datastore, but can be used.
type Item struct {
	Value   []byte
	Expires time.Time // Zero to never expire.
}

// Cache represents the ability to store, get values and also delete.
type Cache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	SetItem(ctx context.Context, key string, item Item) error
	GetItem(ctx context.Context, key string) (Item, error)
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}

// GCable represents the ability to collect and delete expired items.
type GCable interface {
	GC(ctx context.Context) error
}

// Default is a multi-layer cache combining fastest to slowest caches.
// It implements Cache and GCable interface.
var Default = Combined([]Cache{NewAppcache(), Memcache{}, Datastore{}})

// Errors returned by this package.
var (
	// ErrCacheMiss is when key is not found.
	ErrCacheMiss = errors.New("cache: miss")
	// ErrTooBig is when a value/item is too big to fit in the cache.
	ErrTooBig = errors.New("cache: too big")
)

// Set is a wrapper to Default cache Set.
func Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return Default.Set(ctx, key, value, expiration)
}

// Get is a wrapper to Default cache Get.
func Get(ctx context.Context, key string) ([]byte, error) {
	return Default.Get(ctx, key)
}

// SetItem is a wrapper to Default cache SetItem.
func SetItem(ctx context.Context, key string, item Item) error {
	return Default.SetItem(ctx, key, item)
}

// GetItem is a wrapper to Default cache GetItem.
func GetItem(ctx context.Context, key string) (Item, error) {
	return Default.GetItem(ctx, key)
}

// Delete is a wrapper to Default cache Delete.
func Delete(ctx context.Context, key string) error {
	return Default.Delete(ctx, key)
}

// Flush is a wrapper to Default cache Flush.
func Flush(ctx context.Context) error {
	return Default.Flush(ctx)
}

// GC is a wrapper to Default cache GC.
func GC(ctx context.Context) error {
	return Default.GC(ctx)
}

// Helpers to set/get from value to Item. Used in all cache layers.
func set(ctx context.Context, e Cache, key string, value []byte, expiration time.Duration) error {
	item := Item{Value: value}
	if expiration > 0 {
		item.Expires = time.Now().Add(expiration)
	}
	return e.SetItem(ctx, key, item)
}
func get(ctx context.Context, e Cache, key string) ([]byte, error) {
	item, err := e.GetItem(ctx, key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}
