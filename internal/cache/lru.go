package cache

import (
	"container/list"
	"sync"
	"time"
)

// Item represents a cached item with value and optional expiry
type Item[V any] struct {
	Value     V
	ExpiresAt time.Time
}

// LRUCache is a thread-safe in-memory LRU cache with optional TTL support
type LRUCache[K comparable, V any] struct {
	maxSize  int
	items    map[K]*list.Element
	lruList  *list.List
	mu       sync.RWMutex
	evicted  int64
	hits     int64
	misses   int64
}

// cacheEntry holds the actual cache entry
type cacheEntry[K comparable, V any] struct {
	key   K
	value V
	expiry time.Time
}

// NewLRUCache creates a new LRU cache with specified max size
func NewLRUCache[K comparable, V any](maxSize int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		maxSize: maxSize,
		items:   make(map[K]*list.Element),
		lruList: list.New(),
	}
}

// Get retrieves a value from cache, returns (value, found, expired)
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		c.misses++
		var zero V
		return zero, false
	}

	entry := elem.Value.(*cacheEntry[K, V])

	// Check if entry has expired
	if !entry.expiry.IsZero() && time.Now().After(entry.expiry) {
		c.lruList.Remove(elem)
		delete(c.items, key)
		c.misses++
		var zero V
		return zero, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(elem)
	c.hits++
	return entry.value, true
}

// Set adds or updates a value in cache
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, 0)
}

// SetWithTTL adds or updates a value in cache with TTL (0 = no expiry)
func (c *LRUCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, update it
	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry[K, V])
		entry.value = value
		entry.expiry = time.Time{}
		if ttl > 0 {
			entry.expiry = time.Now().Add(ttl)
		}
		c.lruList.MoveToFront(elem)
		return
	}

	// Create new entry
	expiry := time.Time{}
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}

	entry := &cacheEntry[K, V]{
		key:    key,
		value:  value,
		expiry: expiry,
	}

	elem := c.lruList.PushFront(entry)
	c.items[key] = elem

	// Evict oldest item if cache is full
	if len(c.items) > c.maxSize {
		c.evictOldest()
	}
}

// evictOldest removes the least recently used item
func (c *LRUCache[K, V]) evictOldest() {
	elem := c.lruList.Back()
	if elem == nil {
		return
	}

	c.lruList.Remove(elem)
	entry := elem.Value.(*cacheEntry[K, V])
	delete(c.items, entry.key)
	c.evicted++
}

// Delete removes a key from cache
func (c *LRUCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return false
	}

	c.lruList.Remove(elem)
	delete(c.items, key)
	return true
}

// Clear removes all items from cache
func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*list.Element)
	c.lruList = list.New()
	c.evicted = 0
	c.hits = 0
	c.misses = 0
}

// Size returns current number of items in cache
func (c *LRUCache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// HitRate returns the cache hit rate (hits / total accesses)
func (c *LRUCache[K, V]) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}

// Stats returns cache statistics
func (c *LRUCache[K, V]) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"size":       len(c.items),
		"max_size":   c.maxSize,
		"hits":       c.hits,
		"misses":     c.misses,
		"evicted":    c.evicted,
		"hit_rate":   hitRate,
		"total":      total,
	}
}

// CleanupExpired removes all expired entries from cache
func (c *LRUCache[K, V]) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Collect expired keys
	var expiredKeys []K
	for key, elem := range c.items {
		entry := elem.Value.(*cacheEntry[K, V])
		if !entry.expiry.IsZero() && now.After(entry.expiry) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired entries
	for _, key := range expiredKeys {
		if elem, exists := c.items[key]; exists {
			c.lruList.Remove(elem)
			delete(c.items, key)
			removed++
		}
	}

	return removed
}
