package aecache

import (
	"time"

	"appengine"
)

// Combined represents the combination of multiple caches.
// It implements Cache interface.
type Combined []Cache

// Set sets a key to value with a given expiration, one layer after another.
// If the error is not important, consider using a goroutine for it.
func (d Combined) Set(c appengine.Context, key string, value []byte, expiration time.Duration) error {
	return set(d, c, key, value, expiration)
}

// GetItem gets a value given a key by looking through the cache layers.
// When found, a layer refreshes its parent cache.
func (d Combined) Get(c appengine.Context, key string) ([]byte, error) {
	return get(d, c, key)
}

// SetItem sets a key to item, one layer after another.
// If the error is not used, best use a goroutine for it.
func (d Combined) SetItem(c appengine.Context, key string, item Item) error {
	for _, e := range d {
		if err := e.SetItem(c, key, item); err != nil {
			return err
		}
	}
	return nil
}

// GetItem gets an item given a key by looking recursively at the layers.
// When found, a layer refreshes its parent cache.
func (d Combined) GetItem(c appengine.Context, key string) (Item, error) {
	if len(d) == 0 {
		return Item{}, ErrCacheMiss
	}
	item, err := d[0].GetItem(c, key)
	if err == nil {
		return item, nil
	}
	if err != ErrCacheMiss {
		return Item{}, err
	}
	item, err = Combined(d[1:]).GetItem(c, key)
	if err == nil {
		go d[0].SetItem(c, key, item)
	}
	return item, err
}

// Delete deletes an item from the cache by key.
func (d Combined) Delete(c appengine.Context, key string) error {
	for _, e := range d {
		if err := e.Delete(c, key); err != nil {
			return err
		}
	}
	return nil
}

// Flush removes all items from the cache.
func (d Combined) Flush(c appengine.Context) error {
	for _, e := range d {
		if err := e.Flush(c); err != nil {
			return err
		}
	}
	return nil
}

// GC deletes expired items from the cache layers that support it.
func (d Combined) GC(c appengine.Context) error {
	for _, e := range d {
		g, ok := e.(GCable)
		if !ok {
			continue
		}
		if err := g.GC(c); err != nil {
			return err
		}
	}
	return nil
}
