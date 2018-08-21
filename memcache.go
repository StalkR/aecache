package aecache

import (
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
)

// A Memcache represents a cache on top of AppEngine's memcache.
// It implements Cache interface.
type Memcache struct{}

// Set sets a key to value with a given expiration.
func (m Memcache) Set(c appengine.Context, key string, value []byte, expiration time.Duration) error {
	return set(m, c, key, value, expiration)
}

// Get gets a value given a key.
func (m Memcache) Get(c appengine.Context, key string) ([]byte, error) {
	return get(m, c, key)
}

// SetItem sets a key to item.
func (m Memcache) SetItem(c appengine.Context, key string, item Item) error {
	var expiration time.Duration
	if !item.Expires.IsZero() {
		if item.Expires.Before(time.Now()) {
			return nil
		}
		expiration = item.Expires.Sub(time.Now())
	}
	return memcache.Gob.Set(c, &memcache.Item{
		Key:        key,
		Object:     item,
		Expiration: expiration,
	})
}

// GetItem gets an item given a key.
func (m Memcache) GetItem(c appengine.Context, key string) (Item, error) {
	var item Item
	_, err := memcache.Gob.Get(c, key, &item)
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
func (m Memcache) Delete(c appengine.Context, key string) error {
	return memcache.Delete(c, key)
}

// Flush removes all items from memcache, even items not added by this package.
func (m Memcache) Flush(c appengine.Context) error {
	return memcache.Flush(c)
}
