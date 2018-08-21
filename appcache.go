package aecache

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

// An Appcache represents a cache in the app process memory.
// It implements Cache and GCable interfaces.
type Appcache struct {
	sync.Mutex // To protect items.
	items      map[string]Item
}

// NewAppcache creates a new Appcache.
func NewAppcache() *Appcache {
	return &Appcache{items: make(map[string]Item)}
}

// Set sets a key to value with a given expiration.
func (a *Appcache) Set(c context.Context, key string, value []byte, expiration time.Duration) error {
	return set(a, c, key, value, expiration)
}

// Get gets a value given a key.
func (a *Appcache) Get(c context.Context, key string) ([]byte, error) {
	return get(a, c, key)
}

// SetItem sets a key to item.
func (a *Appcache) SetItem(c context.Context, key string, item Item) error {
	a.Lock()
	a.items[key] = item
	a.Unlock()
	a.GC(c)
	return nil
}

// GetItem gets an item given a key.
func (a *Appcache) GetItem(c context.Context, key string) (Item, error) {
	a.Lock()
	defer a.Unlock()
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
func (a *Appcache) Delete(c context.Context, key string) error {
	a.Lock()
	defer a.Unlock()
	delete(a.items, key)
	return nil
}

// Flush removes all items from the cache.
func (a *Appcache) Flush(c context.Context) error {
	a.Lock()
	defer a.Unlock()
	a.items = make(map[string]Item)
	return nil
}

// GC deletes expired items from the cache.
func (a *Appcache) GC(c context.Context) {
	a.Lock()
	defer a.Unlock()
	for key, item := range a.items {
		if item.Expires.Before(time.Now()) {
			delete(a.items, key)
		}
	}
}
