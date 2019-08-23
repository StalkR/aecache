package aecache

import (
	"context"
	"time"
)

// Combined represents the combination of multiple caches.
// It implements Cache interface.
type Combined []Cache

// Set sets a key to value with a given expiration, one layer after another.
// If the error is not important, consider using a goroutine for it.
func (d Combined) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return set(ctx, d, key, value, expiration)
}

// Get gets a value given a key by looking through the cache layers.
// When found, a layer refreshes its parent cache.
func (d Combined) Get(ctx context.Context, key string) ([]byte, error) {
	return get(ctx, d, key)
}

// SetItem sets a key to item, one layer after another.
// If the error is not used, best use a goroutine for it.
func (d Combined) SetItem(ctx context.Context, key string, item Item) error {
	for _, e := range d {
		if err := e.SetItem(ctx, key, item); err != nil {
			return err
		}
	}
	return nil
}

// GetItem gets an item given a key by looking recursively at the layers.
// When found, a layer refreshes its parent cache.
func (d Combined) GetItem(ctx context.Context, key string) (Item, error) {
	if len(d) == 0 {
		return Item{}, ErrCacheMiss
	}
	item, err := d[0].GetItem(ctx, key)
	if err == nil {
		return item, nil
	}
	if err != ErrCacheMiss {
		return Item{}, err
	}
	item, err = Combined(d[1:]).GetItem(ctx, key)
	if err == nil {
		go d[0].SetItem(ctx, key, item)
	}
	return item, err
}

// Delete deletes an item from the cache by key.
func (d Combined) Delete(ctx context.Context, key string) error {
	for _, e := range d {
		if err := e.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Flush removes all items from the cache.
func (d Combined) Flush(ctx context.Context) error {
	for _, e := range d {
		if err := e.Flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GC deletes expired items from the cache layers that support it.
func (d Combined) GC(ctx context.Context) error {
	for _, e := range d {
		g, ok := e.(GCable)
		if !ok {
			continue
		}
		if err := g.GC(ctx); err != nil {
			return err
		}
	}
	return nil
}
