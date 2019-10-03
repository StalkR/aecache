package aecache

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/StalkR/aecache/internal"
)

// A datastoreCache represents a cache on top of Cloud Datastore.
type datastoreCache struct {
	m         sync.Mutex // protects below
	connected bool
	client    *datastore.Client
}

// newDatastoreCache creates a new datastoreCache.
func newDatastoreCache() *datastoreCache {
	return &datastoreCache{}
}

// connect connects a client to the datastore.
// It detects the project ID from credentials.
func (a *datastoreCache) connect(ctx context.Context) error {
	a.m.Lock()
	defer a.m.Unlock()
	if a.connected {
		return nil
	}
	client, err := datastore.NewClient(ctx, datastore.DetectProjectID)
	if err != nil {
		return err
	}
	a.client = client
	a.connected = true
	return nil
}

// Set sets a key to a value with an expiration.
func (a *datastoreCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if expiration <= 0 {
		return nil
	}
	// Per https://godoc.org/cloud.google.com/go/datastore
	// - []byte (up to 1 megabyte in length)
	if len(value) >= 1<<20 {
		return ErrTooBig
	}
	if err := a.connect(ctx); err != nil {
		return err
	}
	k := datastore.NameKey("CacheItem", key, nil)
	item := internal.CacheItem{
		Value:   value,
		Expires: time.Now().Add(expiration),
	}
	if _, err := a.client.Put(ctx, k, &item); err != nil {
		return err
	}
	return nil
}

// Get gets the value and expiration for a key.
func (a *datastoreCache) Get(ctx context.Context, key string) ([]byte, time.Time, error) {
	if err := a.connect(ctx); err != nil {
		return nil, time.Time{}, err
	}
	k := datastore.NameKey("CacheItem", key, nil)
	item := internal.CacheItem{}
	err := a.client.Get(ctx, k, &item)
	if err == datastore.ErrNoSuchEntity {
		return nil, time.Time{}, ErrCacheMiss
	}
	if err != nil {
		return nil, time.Time{}, err
	}
	if item.Expires.Before(time.Now()) {
		if err := a.client.Delete(ctx, k); err != nil {
			return nil, time.Time{}, err
		}
		return nil, time.Time{}, ErrCacheMiss
	}
	return item.Value, item.Expires, nil
}

// Clean deletes expired items.
func (a *datastoreCache) Clean(ctx context.Context) error {
	if err := a.connect(ctx); err != nil {
		return err
	}
	q := datastore.NewQuery("CacheItem").Filter("Expires <", time.Now()).KeysOnly()
	keys, err := a.client.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	// Batch deletes, per error "cannot write more than 500 entities in a single call".
	const batchSize = 500
	for len(keys) > 0 {
		n := batchSize
		if n > len(keys) {
			n = len(keys)
		}
		if err := a.client.DeleteMulti(ctx, keys[:n]); err != nil {
			return err
		}
		keys = keys[n:]
	}
	return nil
}
