package pokecache

import (
	"sync"
	"time"
)

// Cache represents a cache with a map, a mutex, and cache expiration.
type Cache struct {
	data        map[string]CacheEntry
	mutex       sync.RWMutex
	expiration  time.Duration
	cleanupDone chan struct{}
}

type CacheEntry struct {
	createdAt time.Time
	val       []byte
}

// NewCache creates a new cache with a configurable expiration interval.
func NewCache(expiration time.Duration) *Cache {
	cache := &Cache{
		data:        make(map[string]CacheEntry),
		expiration:  expiration,
		cleanupDone: make(chan struct{}),
	}

	// Start a background goroutine to periodically clean up expired entries.
	go cache.CleanupExpiredEntries()
	return cache
}

// Add adds a new entry to the cache with the specified key and value.
func (c *Cache) Add(key string, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = CacheEntry{
		createdAt: time.Now(),
		val:       value,
	}
}

// Get gets an entry from the cache with the specified key.
// It returns a []byte and bool indicating if the entry was found.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, found := c.data[key]
	if !found {
		return nil, false
	}

	//Check if the entry has expired
	if time.Since(entry.createdAt) > c.expiration {
		delete(c.data, key)
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
}

func (c *Cache) CleanupExpiredEntries() {
	ticker := time.NewTicker(c.expiration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			for key, entry := range c.data {
				if time.Since(entry.createdAt) > c.expiration {
					delete(c.data, key)
				}
			}
			c.mutex.Unlock()
		case <-c.cleanupDone:
			return
		}
	}
}

// Close stops the background cleanup goroutine and releases resources.
func (c *Cache) Close() {
	close(c.cleanupDone)
}
