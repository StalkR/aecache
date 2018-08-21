package aecache

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/memcache"
)

// A Memcache represents a cache on top of AppEngine's memcache.
// It implements Cache interface.
type Memcache struct{}

// Set sets a key to value with a given expiration.
func (m Memcache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return set(ctx, m, key, value, expiration)
}

// Get gets a value given a key.
func (m Memcache) Get(ctx context.Context, key string) ([]byte, error) {
	return get(ctx, m, key)
}

// SetItem sets a key to item.
func (m Memcache) SetItem(ctx context.Context, key string, item Item) error {
	var expiration time.Duration
	if !item.Expires.IsZero() {
		if item.Expires.Before(time.Now()) {
			return nil
		}
		expiration = item.Expires.Sub(time.Now())
	}
	return memcache.Gob.Set(ctx, &memcache.Item{
		Key:        key,
		Object:     item,
		Expiration: expiration,
	})
}

// GetItem gets an item given a key.
func (m Memcache) GetItem(ctx context.Context, key string) (Item, error) {
	var item Item
	_, err := memcache.Gob.Get(ctx, key, &item)
	if err == memcache.ErrCacheMiss {
		return Item{}, ErrCacheMiss
	}
	if err != nil {
		return Item{}, err
	}
	if !item.Expires.IsZero() && item.Expires.Before(time.Now()) {
		return Item{}, ErrCacheMiss
	}
	return item, nil
}

// Delete deletes an item from the cache by key.
func (m Memcache) Delete(ctx context.Context, key string) error {
	return memcache.Delete(ctx, key)
}

// Flush removes all items from memcache, even items not added by this package.
func (m Memcache) Flush(ctx context.Context) error {
	return memcache.Flush(ctx)
}
