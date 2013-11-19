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
		c.Infof("cache: combined %T hit", d[0])
		return item, nil
	}
	if err != ErrCacheMiss {
		c.Errorf("cache: combined %T get: %v", d[0], err)
		return Item{}, err
	}
	c.Infof("cache: combined %T miss", d[0])
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

/*

// Get gets a value given a key by looking through the cache layers.
// When found, a layer refreshes its parent cache.
func (x L3) Get(c appengine.Context, key string) ([]byte, error) {
	// First layer: in-app memory cache.
	item, err := inapp.Get(key)
	if err == nil {
		c.Infof("cache: in-app hit")
		return item.Value, nil
	}
	// Second layer: memcache.
	item, err = memcache.Get(c, key)
	if err == nil {
		c.Infof("cache: memcache hit")
		go inapp.Set(key, item)
		return item.Value, nil
	}
	if err != ErrCacheMiss {
		c.Errorf("cache: memcache get: %v", err)
		return nil, err
	}
	// Third layer: datastore.
	item, err = datastore.Get(c, key)
	if err == nil {
		c.Infof("cache: datastore hit")
		go inapp.Set(key, item)
		go memcache.Set(key, item)
		return item.Value, nil
	}
	if err != ErrCacheMiss {
		c.Errorf("cache: datastore get: %v", err)
		return nil, err
	}
	c.Infof("cache: miss")
	return nil, ErrCacheMiss
}

// Set sets a key to value with a given expiration.
func (x L3) Set(c appengine.Context, key string, value []byte, expiration time.Duration) error {
	// First layer: in-app memory cache.
	inapp.SetValue(key, value, expiration)
	// Second layer: memcache.
	if err := memcache.SetValue(c, key, value, expiration); err != nil {
		c.Errorf("cache: memcache set: %v", err)
		return err
	}
	// Third layer: datastore.
	if err := datastore.SetValue(c, key, value, expiration); err != nil {
		c.Errorf("cache: datastore set: %v", err)
		return err
	}
	return nil
}
*/
