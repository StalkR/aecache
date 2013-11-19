package aecache

import (
	"sync"
	"time"

	"appengine"
)

// An Appcache represents a cache in the app process memory.
// It implements Cache and GCable interfaces.
type Appcache struct {
	m     sync.Mutex // To protect items.
	items map[string]Item
}

// NewAppcache creates a new Appcache.
func NewAppcache() Appcache {
	return Appcache{items: make(map[string]Item)}
}

// Set sets a key to value with a given expiration.
func (a Appcache) Set(c appengine.Context, key string, value []byte, expiration time.Duration) error {
	return set(a, c, key, value, expiration)
}

// Get gets a value given a key.
func (a Appcache) Get(c appengine.Context, key string) ([]byte, error) {
	return get(a, c, key)
}

// SetItem sets a key to item.
func (a Appcache) SetItem(c appengine.Context, key string, item Item) error {
	a.m.Lock()
	a.items[key] = item
	a.m.Unlock()
	a.GC(c)
	return nil
}

// GetItem gets an item given a key.
func (a Appcache) GetItem(c appengine.Context, key string) (Item, error) {
	a.m.Lock()
	defer a.m.Unlock()
	item, ok := a.items[key]
	if !ok {
		return Item{}, ErrCacheMiss
	}
	if !item.Expires.IsZero() && item.Expires.Before(time.Now()) {
		delete(a.items, key)
		return Item{}, ErrCacheMiss
	}
	return item, nil
}

// Delete deletes an item from the cache by key.
func (a Appcache) Delete(c appengine.Context, key string) error {
	a.m.Lock()
	defer a.m.Unlock()
	delete(a.items, key)
	return nil
}

// Flush removes all items from the cache.
func (a Appcache) Flush(c appengine.Context) error {
	a.m.Lock()
	defer a.m.Unlock()
	a.items = make(map[string]Item)
	return nil
}

// GC deletes expired items from the cache.
func (a Appcache) GC(c appengine.Context) {
	a.m.Lock()
	defer a.m.Unlock()
	for key, item := range a.items {
		if item.Expires.Before(time.Now()) {
			delete(a.items, key)
		}
	}
}
