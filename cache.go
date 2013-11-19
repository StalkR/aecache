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

	"appengine"
)

// An Item represents a value to cache with its expiration time.
// It is mostly for internal and exported for memcache/datastore, but can be used.
type Item struct {
	Value   []byte
	Expires time.Time // Zero to never expire.
}

// Cache represents the ability to store, get values and also delete.
type Cache interface {
	Set(c appengine.Context, key string, value []byte, expiration time.Duration) error
	Get(c appengine.Context, key string) ([]byte, error)
	SetItem(c appengine.Context, key string, item Item) error
	GetItem(c appengine.Context, key string) (Item, error)
	Delete(c appengine.Context, key string) error
	Flush(c appengine.Context) error
}

// GCable represents the ability to collect and delete expired items.
type GCable interface {
	GC(c appengine.Context) error
}

// Default layers of cache. They implement Cache interface.
var (
	Appcache  = NewAppcacheLayer()
	Memcache  = MemcacheLayer{}
	Datastore = DatastoreLayer{}
)

// Default is a multi-layer cache combining fastest to slowest caches.
// It implements Cache interface.
var Default = Combined([]Cache{Appcache, Memcache, Datastore})

var (
	// ErrCacheMiss is when key is not found.
	ErrCacheMiss = errors.New("cache: miss")
	// ErrTooBig is when a value/item is too big to fit in the cache.
	ErrTooBig = errors.New("cache: too big")
)

// Set is a wrapper to Default cache Set.
func Set(c appengine.Context, key string, value []byte, expiration time.Duration) error {
	return Default.Set(c, key, value, expiration)
}

// Get is a wrapper to Default cache Get.
func Get(c appengine.Context, key string) ([]byte, error) {
	return Default.Get(c, key)
}

// SetItem is a wrapper to Default cache SetItem.
func SetItem(c appengine.Context, key string, item Item) error {
	return Default.SetItem(c, key, item)
}

// GetItem is a wrapper to Default cache GetItem.
func GetItem(c appengine.Context, key string) (Item, error) {
	return Default.GetItem(c, key)
}

// Delete is a wrapper to Default cache Delete.
func Delete(c appengine.Context, key string) error {
	return Default.Delete(c, key)
}

// Flush is a wrapper to Default cache Flush.
func Flush(c appengine.Context) error {
	return Default.Flush(c)
}

// Helpers to set/get from value to Item. Used in all cache layers.
func set(e Cache, c appengine.Context, key string, value []byte, expiration time.Duration) error {
	item := Item{Value: value}
	if expiration > 0 {
		item.Expires = time.Now().Add(expiration)
	}
	return e.SetItem(c, key, item)
}
func get(e Cache, c appengine.Context, key string) ([]byte, error) {
	item, err := e.GetItem(c, key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}
