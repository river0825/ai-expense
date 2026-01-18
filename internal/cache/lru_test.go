package cache

import (
	"sync"
	"testing"
	"time"
)

// TestLRUCacheBasicOperations tests basic set/get operations
func TestLRUCacheBasicOperations(t *testing.T) {
	cache := NewLRUCache[string, string](3)

	t.Run("SetAndGet", func(t *testing.T) {
		cache.Set("key1", "value1")
		val, found := cache.Get("key1")
		if !found {
			t.Fatal("expected key to be found")
		}
		if val != "value1" {
			t.Errorf("expected 'value1', got '%s'", val)
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		_, found := cache.Get("nonexistent")
		if found {
			t.Fatal("expected key not to be found")
		}
	})

	t.Run("UpdateExisting", func(t *testing.T) {
		cache.Set("key1", "updated")
		val, _ := cache.Get("key1")
		if val != "updated" {
			t.Errorf("expected 'updated', got '%s'", val)
		}
	})
}

// TestLRUCacheEviction tests LRU eviction when cache is full
func TestLRUCacheEviction(t *testing.T) {
	cache := NewLRUCache[string, string](3)

	t.Run("EvictLeastRecentlyUsed", func(t *testing.T) {
		cache.Set("a", "1")
		cache.Set("b", "2")
		cache.Set("c", "3")
		stats := cache.Stats()
		if stats["size"].(int) != 3 {
			t.Errorf("expected size 3, got %d", stats["size"])
		}

		// Add 4th item, should evict 'a'
		cache.Set("d", "4")

		_, found := cache.Get("a")
		if found {
			t.Fatal("expected 'a' to be evicted")
		}

		_, found = cache.Get("d")
		if !found {
			t.Fatal("expected 'd' to be in cache")
		}

		stats = cache.Stats()
		if stats["evicted"].(int64) != 1 {
			t.Errorf("expected 1 eviction, got %d", stats["evicted"])
		}
	})

	t.Run("AccessMovesToFront", func(t *testing.T) {
		cache.Clear()
		cache.Set("x", "1")
		cache.Set("y", "2")
		cache.Set("z", "3")

		// Access 'x' to move it to front
		cache.Get("x")

		// Add new item, should evict 'y' (least recently used)
		cache.Set("w", "4")

		_, foundX := cache.Get("x")
		_, foundY := cache.Get("y")
		_, foundW := cache.Get("w")

		if !foundX {
			t.Fatal("expected 'x' to still be in cache")
		}
		if foundY {
			t.Fatal("expected 'y' to be evicted")
		}
		if !foundW {
			t.Fatal("expected 'w' to be in cache")
		}
	})
}

// TestLRUCacheTTL tests time-to-live functionality
func TestLRUCacheTTL(t *testing.T) {
	cache := NewLRUCache[string, string](10)

	t.Run("TTLExpiry", func(t *testing.T) {
		cache.SetWithTTL("tempkey", "tempval", 100*time.Millisecond)

		// Should exist immediately
		_, found := cache.Get("tempkey")
		if !found {
			t.Fatal("expected key to exist")
		}

		// Wait for expiry
		time.Sleep(150 * time.Millisecond)

		_, found = cache.Get("tempkey")
		if found {
			t.Fatal("expected key to have expired")
		}
	})

	t.Run("NoTTL", func(t *testing.T) {
		cache.Clear()
		cache.Set("permanent", "value")

		// Wait and check still exists
		time.Sleep(50 * time.Millisecond)
		_, found := cache.Get("permanent")
		if !found {
			t.Fatal("expected permanent key to still exist")
		}
	})

	t.Run("CleanupExpired", func(t *testing.T) {
		cache.Clear()
		cache.SetWithTTL("exp1", "val1", 50*time.Millisecond)
		cache.SetWithTTL("exp2", "val2", 50*time.Millisecond)
		cache.Set("permanent", "val3")

		time.Sleep(100 * time.Millisecond)

		removed := cache.CleanupExpired()
		if removed != 2 {
			t.Errorf("expected 2 expired entries removed, got %d", removed)
		}

		_, found := cache.Get("permanent")
		if !found {
			t.Fatal("expected permanent key to still exist")
		}
	})
}

// TestLRUCacheDelete tests delete functionality
func TestLRUCacheDelete(t *testing.T) {
	cache := NewLRUCache[string, string](5)
	cache.Set("key1", "val1")
	cache.Set("key2", "val2")

	t.Run("DeleteExisting", func(t *testing.T) {
		deleted := cache.Delete("key1")
		if !deleted {
			t.Fatal("expected delete to return true")
		}

		_, found := cache.Get("key1")
		if found {
			t.Fatal("expected key to be deleted")
		}
	})

	t.Run("DeleteNonExistent", func(t *testing.T) {
		deleted := cache.Delete("nonexistent")
		if deleted {
			t.Fatal("expected delete to return false for non-existent key")
		}
	})
}

// TestLRUCacheClear tests clear functionality
func TestLRUCacheClear(t *testing.T) {
	cache := NewLRUCache[string, string](5)
	cache.Set("a", "1")
	cache.Set("b", "2")
	cache.Set("c", "3")

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}

	_, found := cache.Get("a")
	if found {
		t.Fatal("expected cache to be empty")
	}
}

// TestLRUCacheStats tests statistics tracking
func TestLRUCacheStats(t *testing.T) {
	cache := NewLRUCache[string, string](5)

	t.Run("HitAndMissTracking", func(t *testing.T) {
		cache.Set("key1", "val1")

		// 3 hits
		cache.Get("key1")
		cache.Get("key1")
		cache.Get("key1")

		// 2 misses
		cache.Get("key2")
		cache.Get("key3")

		stats := cache.Stats()
		if stats["hits"].(int64) != 3 {
			t.Errorf("expected 3 hits, got %d", stats["hits"])
		}
		if stats["misses"].(int64) != 2 {
			t.Errorf("expected 2 misses, got %d", stats["misses"])
		}

		hitRate := cache.HitRate()
		expected := 3.0 / 5.0
		if hitRate < expected-0.001 || hitRate > expected+0.001 {
			t.Errorf("expected hit rate %.2f, got %.2f", expected, hitRate)
		}
	})

	t.Run("SizeTracking", func(t *testing.T) {
		cache.Clear()
		cache.Set("a", "1")
		cache.Set("b", "2")

		if cache.Size() != 2 {
			t.Errorf("expected size 2, got %d", cache.Size())
		}

		stats := cache.Stats()
		if stats["max_size"] != 5 {
			t.Errorf("expected max_size 5, got %d", stats["max_size"])
		}
	})
}

// TestLRUCacheConcurrency tests thread-safety
func TestLRUCacheConcurrency(t *testing.T) {
	cache := NewLRUCache[string, int](1000)
	var wg sync.WaitGroup
	numGoroutines := 100
	opsPerGoroutine := 100

	t.Run("ConcurrentSetGet", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsPerGoroutine; j++ {
					key := "key" + string(rune(id%10))
					cache.Set(key, j)
					cache.Get(key)
				}
			}(i)
		}
		wg.Wait()

		if cache.Size() > 1000 {
			t.Errorf("expected size <= 1000, got %d", cache.Size())
		}
	})

	t.Run("ConcurrentDelete", func(t *testing.T) {
		cache.Clear()
		for i := 0; i < 100; i++ {
			cache.Set("key"+string(rune(i)), i)
		}

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					cache.Delete("key" + string(rune((id*10+j)%100)))
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestLRUCacheGenericTypes tests with different types
func TestLRUCacheGenericTypes(t *testing.T) {
	t.Run("IntegerCache", func(t *testing.T) {
		intCache := NewLRUCache[string, int](5)
		intCache.Set("num", 42)
		val, found := intCache.Get("num")
		if !found || val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	t.Run("SliceCache", func(t *testing.T) {
		sliceCache := NewLRUCache[string, []string](5)
		items := []string{"a", "b", "c"}
		sliceCache.Set("list", items)
		val, found := sliceCache.Get("list")
		if !found || len(val) != 3 || val[0] != "a" {
			t.Errorf("expected ['a', 'b', 'c'], got %v", val)
		}
	})

	t.Run("StructCache", func(t *testing.T) {
		type TestData struct {
			Name  string
			Value int
		}
		structCache := NewLRUCache[string, *TestData](5)
		data := &TestData{Name: "test", Value: 100}
		structCache.Set("data", data)
		val, found := structCache.Get("data")
		if !found || val.Value != 100 {
			t.Errorf("expected value 100, got %d", val.Value)
		}
	})
}

// TestLRUCacheEdgeCases tests edge cases
func TestLRUCacheEdgeCases(t *testing.T) {
	t.Run("CacheSizeOne", func(t *testing.T) {
		cache := NewLRUCache[string, string](1)
		cache.Set("a", "1")
		cache.Set("b", "2")

		_, foundA := cache.Get("a")
		_, foundB := cache.Get("b")

		if foundA {
			t.Fatal("expected 'a' to be evicted")
		}
		if !foundB {
			t.Fatal("expected 'b' to be in cache")
		}
	})

	t.Run("EmptyCache", func(t *testing.T) {
		cache := NewLRUCache[string, string](10)
		if cache.Size() != 0 {
			t.Errorf("expected empty cache size 0, got %d", cache.Size())
		}
		if cache.HitRate() != 0 {
			t.Errorf("expected hit rate 0, got %f", cache.HitRate())
		}
	})

	t.Run("UpdateDoesNotEvict", func(t *testing.T) {
		cache := NewLRUCache[string, string](2)
		cache.Set("a", "1")
		cache.Set("b", "2")

		// Update existing key - should not cause eviction
		cache.Set("a", "1_updated")

		_, foundA := cache.Get("a")
		_, foundB := cache.Get("b")

		if !foundA || !foundB {
			t.Fatal("updating existing key should not cause eviction")
		}
	})
}
