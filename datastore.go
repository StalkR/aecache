package aecache

import (
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// A Datastore represents a cache on top of AppEngine's datastore.
// It implements Cache and GCable interfaces.
type Datastore struct{}

// Set sets a key to value with a given expiration.
func (d Datastore) Set(c appengine.Context, key string, value []byte, expiration time.Duration) error {
	return set(d, c, key, value, expiration)
}

// Get gets a value given a key.
func (d Datastore) Get(c appengine.Context, key string) ([]byte, error) {
	return get(d, c, key)
}

// SetItem sets a key to item.
func (d Datastore) SetItem(c appengine.Context, key string, item Item) error {
	// Datastore can save up to 1MB of []byte in an entity.
	if len(item.Value) >= 1<<20 {
		return ErrTooBig
	}
	k := datastore.NewKey(c, "Item", "Cache:"+key, 0, nil)
	_, err := datastore.Put(c, k, &item)
	return err
}

// GetItem gets an item given a key.
func (d Datastore) GetItem(c appengine.Context, key string) (Item, error) {
	k := datastore.NewKey(c, "Item", "Cache:"+key, 0, nil)
	item := Item{}
	err := datastore.Get(c, k, &item)
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
func (d Datastore) Delete(c appengine.Context, key string) error {
	k := datastore.NewKey(c, "Item", "Cache:"+key, 0, nil)
	return datastore.Delete(c, k)
}

// Flush removes all cache items from the datastore.
func (d Datastore) Flush(c appengine.Context) error {
	keys, err := datastore.NewQuery("Item").KeysOnly().GetAll(c, nil)
	if err != nil {
		return err
	}
	return datastore.DeleteMulti(c, keys)
}

// GC deletes expired cache items from the datastore.
func (d Datastore) GC(c appengine.Context) error {
	q := datastore.NewQuery("Item").Filter("Expires >", time.Time{})
	keys, err := q.Filter("Expires <", time.Now()).KeysOnly().GetAll(c, nil)
	if err != nil {
		return err
	}
	return datastore.DeleteMulti(c, keys)
}
