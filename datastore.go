package aecache

import (
	"context"
	"time"

	"google.golang.org/appengine/datastore"
)

// A Datastore represents a cache on top of AppEngine's datastore.
// It implements Cache and GCable interfaces.
type Datastore struct{}

// Set sets a key to value with a given expiration.
func (d Datastore) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return set(ctx, d, key, value, expiration)
}

// Get gets a value given a key.
func (d Datastore) Get(ctx context.Context, key string) ([]byte, error) {
	return get(ctx, d, key)
}

// SetItem sets a key to item.
func (d Datastore) SetItem(ctx context.Context, key string, item Item) error {
	// Datastore can save up to 1MB of []byte in an entity.
	if len(item.Value) >= 1<<20 {
		return ErrTooBig
	}
	k := datastore.NewKey(ctx, "Item", "Cache:"+key, 0, nil)
	_, err := datastore.Put(ctx, k, &item)
	return err
}

// GetItem gets an item given a key.
func (d Datastore) GetItem(ctx context.Context, key string) (Item, error) {
	k := datastore.NewKey(ctx, "Item", "Cache:"+key, 0, nil)
	item := Item{}
	err := datastore.Get(ctx, k, &item)
	if err == datastore.ErrNoSuchEntity {
		return item, ErrCacheMiss
	}
	if err != nil {
		return item, err
	}
	if !item.Expires.IsZero() && item.Expires.Before(time.Now()) {
		return item, ErrCacheMiss
	}
	return item, nil
}

// Delete deletes an item from the cache by key.
func (d Datastore) Delete(ctx context.Context, key string) error {
	k := datastore.NewKey(ctx, "Item", "Cache:"+key, 0, nil)
	return datastore.Delete(ctx, k)
}

// Flush removes all cache items from the datastore.
func (d Datastore) Flush(ctx context.Context) error {
	keys, err := datastore.NewQuery("Item").KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return err
	}
	return datastore.DeleteMulti(ctx, keys)
}

// GC deletes expired cache items from the datastore.
func (d Datastore) GC(ctx context.Context) error {
	q := datastore.NewQuery("Item").Filter("Expires >", time.Time{})
	keys, err := q.Filter("Expires <", time.Now()).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return err
	}
	for len(keys) > 0 {
		// "API error 1 (datastore_v3: BAD_REQUEST): cannot write more than 500 entities in a single call"
		n := 500
		if n > len(keys) {
			n = len(keys)
		}
		if err := datastore.DeleteMulti(ctx, keys[:n]); err != nil {
			return err
		}
		keys = keys[n:]
	}
	return nil
}
