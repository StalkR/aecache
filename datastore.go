package aecache

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
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
	client, err := datastore.NewClient(ctx, "")
	if err != nil {
		return err
	}
	k := datastore.NameKey("Item", "Cache:"+key, nil)
	_, err = client.Put(ctx, k, &item)
	return err
}

// GetItem gets an item given a key.
func (d Datastore) GetItem(ctx context.Context, key string) (Item, error) {
	item := Item{}
	client, err := datastore.NewClient(ctx, "")
	if err != nil {
		return item, err
	}
	k := datastore.NameKey("Item", "Cache:"+key, nil)
	if err := client.Get(ctx, k, &item); err == datastore.ErrNoSuchEntity {
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
	client, err := datastore.NewClient(ctx, "")
	if err != nil {
		return err
	}
	k := datastore.NameKey("Item", "Cache:"+key, nil)
	return client.Delete(ctx, k)
}

// Flush removes all cache items from the datastore.
func (d Datastore) Flush(ctx context.Context) error {
	client, err := datastore.NewClient(ctx, "")
	if err != nil {
		return err
	}
	q := datastore.NewQuery("Item").KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	return d.batchDelete(ctx, client, keys)
}

// GC deletes expired cache items from the datastore.
func (d Datastore) GC(ctx context.Context) error {
	client, err := datastore.NewClient(ctx, "")
	if err != nil {
		return err
	}
	q := datastore.NewQuery("Item").Filter("Expires >", time.Time{}).Filter("Expires <", time.Now()).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	return d.batchDelete(ctx, client, keys)
}

func (d Datastore) batchDelete(ctx context.Context, client *datastore.Client, keys []*datastore.Key) error {
	// According to error "cannot write more than 500 entities in a single call".
	const batchSize = 500
	for len(keys) > 0 {
		n := batchSize
		if n > len(keys) {
			n = len(keys)
		}
		if err := client.DeleteMulti(ctx, keys[:n]); err != nil {
			return err
		}
		keys = keys[n:]
	}
	return nil
}
