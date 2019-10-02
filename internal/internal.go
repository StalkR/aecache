// Package internal holds exported structures for Cloud Datastore.
package internal

import "time"

// A CacheItem represents a cached item in Cloud Datastore.
type CacheItem struct {
	Value   []byte
	Expires time.Time
}
