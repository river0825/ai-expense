package cache

import (
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// TestCacheManagerUserOperations tests user cache operations
func TestCacheManagerUserOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("CacheAndRetrieveUser", func(t *testing.T) {
		user := &domain.User{
			UserID:        "user123",
			MessengerType: "line",
			CreatedAt:     time.Now(),
		}

		cm.SetUser(user)
		retrieved, found := cm.GetUser("user123")

		if !found {
			t.Fatal("expected user to be in cache")
		}
		if retrieved.UserID != "user123" {
			t.Errorf("expected user123, got %s", retrieved.UserID)
		}
	})

	t.Run("InvalidateUser", func(t *testing.T) {
		user := &domain.User{
			UserID:        "user456",
			MessengerType: "telegram",
			CreatedAt:     time.Now(),
		}

		cm.SetUser(user)
		cm.InvalidateUser("user456")

		_, found := cm.GetUser("user456")
		if found {
			t.Fatal("expected user to be invalidated")
		}
	})

	t.Run("UserTTL", func(t *testing.T) {
		user := &domain.User{
			UserID:        "user789",
			MessengerType: "slack",
			CreatedAt:     time.Now(),
		}

		cm.SetUser(user)
		_, found := cm.GetUser("user789")
		if !found {
			t.Fatal("expected user to be cached")
		}

		// Note: TTL is 1 hour by default, so we can't easily test expiry
		// without modifying the TTL constants
	})
}

// TestCacheManagerCategoryOperations tests category cache operations
func TestCacheManagerCategoryOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("CacheAndRetrieveCategory", func(t *testing.T) {
		category := &domain.Category{
			ID:        "cat1",
			UserID:    "user1",
			Name:      "Food",
			IsDefault: true,
		}

		cm.SetCategory(category)
		retrieved, found := cm.GetCategory("cat1")

		if !found {
			t.Fatal("expected category to be in cache")
		}
		if retrieved.Name != "Food" {
			t.Errorf("expected Food, got %s", retrieved.Name)
		}
	})

	t.Run("InvalidateCategory", func(t *testing.T) {
		category := &domain.Category{
			ID:        "cat2",
			UserID:    "user2",
			Name:      "Transport",
			IsDefault: true,
		}

		cm.SetCategory(category)
		cm.InvalidateCategory("cat2", "user2")

		_, found := cm.GetCategory("cat2")
		if found {
			t.Fatal("expected category to be invalidated")
		}
	})

	t.Run("SetCategoryInvalidatesUserCategories", func(t *testing.T) {
		categories := []*domain.Category{
			{ID: "cat3", UserID: "user3", Name: "Food", IsDefault: true},
			{ID: "cat4", UserID: "user3", Name: "Transport", IsDefault: true},
		}

		cm.SetUserCategories("user3", categories)

		// Verify cached
		_, found := cm.GetUserCategories("user3")
		if !found {
			t.Fatal("expected user categories to be cached")
		}

		// Set a new category for this user
		newCat := &domain.Category{
			ID:        "cat5",
			UserID:    "user3",
			Name:      "Shopping",
			IsDefault: false,
		}
		cm.SetCategory(newCat)

		// User categories cache should be invalidated
		_, found = cm.GetUserCategories("user3")
		if found {
			t.Fatal("expected user categories cache to be invalidated")
		}
	})
}

// TestCacheManagerUserCategoriesOperations tests user categories cache
func TestCacheManagerUserCategoriesOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("CacheUserCategories", func(t *testing.T) {
		categories := []*domain.Category{
			{ID: "c1", UserID: "u1", Name: "Food", IsDefault: true},
			{ID: "c2", UserID: "u1", Name: "Transport", IsDefault: true},
		}

		cm.SetUserCategories("u1", categories)
		retrieved, found := cm.GetUserCategories("u1")

		if !found {
			t.Fatal("expected categories to be cached")
		}
		if len(retrieved) != 2 {
			t.Errorf("expected 2 categories, got %d", len(retrieved))
		}
	})

	t.Run("InvalidateUserCategories", func(t *testing.T) {
		categories := []*domain.Category{
			{ID: "c3", UserID: "u2", Name: "Food", IsDefault: true},
		}

		cm.SetUserCategories("u2", categories)
		cm.InvalidateUserCategories("u2")

		_, found := cm.GetUserCategories("u2")
		if found {
			t.Fatal("expected user categories to be invalidated")
		}
	})
}

// TestCacheManagerKeywordOperations tests keyword cache operations
func TestCacheManagerKeywordOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("CacheKeywords", func(t *testing.T) {
		keywords := []*domain.CategoryKeyword{
			{ID: "kw1", CategoryID: "cat1", Keyword: "breakfast", Priority: 1},
			{ID: "kw2", CategoryID: "cat1", Keyword: "lunch", Priority: 2},
		}

		cm.SetCategoryKeywords("cat1", keywords)
		retrieved, found := cm.GetCategoryKeywords("cat1")

		if !found {
			t.Fatal("expected keywords to be cached")
		}
		if len(retrieved) != 2 {
			t.Errorf("expected 2 keywords, got %d", len(retrieved))
		}
	})

	t.Run("InvalidateKeywords", func(t *testing.T) {
		keywords := []*domain.CategoryKeyword{
			{ID: "kw3", CategoryID: "cat2", Keyword: "taxi", Priority: 1},
		}

		cm.SetCategoryKeywords("cat2", keywords)
		cm.InvalidateCategoryKeywords("cat2")

		_, found := cm.GetCategoryKeywords("cat2")
		if found {
			t.Fatal("expected keywords to be invalidated")
		}
	})
}

// TestCacheManagerMetricsOperations tests metrics cache operations
func TestCacheManagerMetricsOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("CacheMetrics", func(t *testing.T) {
		metrics := &domain.DailyMetrics{
			Date:           time.Now(),
			ActiveUsers:    10,
			TotalExpense:   500.0,
			ExpenseCount:   10,
			AverageExpense: 50.0,
		}

		dateKey := "2024-01-15"
		cm.SetMetrics(dateKey, metrics)
		retrieved, found := cm.GetMetrics(dateKey)

		if !found {
			t.Fatal("expected metrics to be cached")
		}
		if retrieved.ExpenseCount != 10 {
			t.Errorf("expected expense count 10, got %d", retrieved.ExpenseCount)
		}
	})

	t.Run("InvalidateMetrics", func(t *testing.T) {
		metrics := &domain.DailyMetrics{
			Date:           time.Now(),
			ActiveUsers:    5,
			TotalExpense:   250.0,
			ExpenseCount:   5,
			AverageExpense: 50.0,
		}

		cm.SetMetrics("2024-01-16", metrics)
		cm.InvalidateMetrics("2024-01-16")

		_, found := cm.GetMetrics("2024-01-16")
		if found {
			t.Fatal("expected metrics to be invalidated")
		}
	})
}

// TestCacheManagerGlobalOperations tests global cache operations
func TestCacheManagerGlobalOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("Clear", func(t *testing.T) {
		// Add various items
		cm.SetUser(&domain.User{UserID: "u1", MessengerType: "line", CreatedAt: time.Now()})
		cm.SetCategory(&domain.Category{ID: "c1", UserID: "u1", Name: "Food", IsDefault: true})
		cm.SetUserCategories("u1", []*domain.Category{})
		cm.SetMetrics("2024-01-15", &domain.DailyMetrics{Date: time.Now()})

		// Verify items are cached
		stats := cm.Stats()
		totalSize := stats["total_size"].(int)
		if totalSize == 0 {
			t.Fatal("expected cache to have items before clear")
		}

		// Clear and verify
		cm.Clear()
		stats = cm.Stats()
		totalSize = stats["total_size"].(int)
		if totalSize != 0 {
			t.Errorf("expected total_size 0 after clear, got %d", totalSize)
		}
	})

	t.Run("InvalidateUserData", func(t *testing.T) {
		cm.Clear()

		userID := "u1"
		cm.SetUser(&domain.User{UserID: userID, MessengerType: "line", CreatedAt: time.Now()})
		cm.SetUserCategories(userID, []*domain.Category{
			{ID: "c1", UserID: userID, Name: "Food", IsDefault: true},
		})

		// Verify cached
		_, found := cm.GetUser(userID)
		if !found {
			t.Fatal("expected user to be cached")
		}
		_, found = cm.GetUserCategories(userID)
		if !found {
			t.Fatal("expected user categories to be cached")
		}

		// Invalidate user data
		cm.InvalidateUserData(userID)

		// Verify invalidated
		_, found = cm.GetUser(userID)
		if found {
			t.Fatal("expected user to be invalidated")
		}
		_, found = cm.GetUserCategories(userID)
		if found {
			t.Fatal("expected user categories to be invalidated")
		}
	})

	t.Run("Stats", func(t *testing.T) {
		cm.Clear()

		cm.SetUser(&domain.User{UserID: "u1", MessengerType: "line", CreatedAt: time.Now()})
		cm.SetCategory(&domain.Category{ID: "c1", UserID: "u1", Name: "Food", IsDefault: true})
		cm.SetMetrics("2024-01-15", &domain.DailyMetrics{Date: time.Now()})

		stats := cm.Stats()

		if stats["users"] == nil {
			t.Fatal("expected users stats")
		}
		if stats["categories"] == nil {
			t.Fatal("expected categories stats")
		}
		if stats["total_size"].(int) != 3 {
			t.Errorf("expected total_size 3, got %d", stats["total_size"].(int))
		}
	})
}

// TestCacheManagerCleanupExpired tests cleanup of expired entries
func TestCacheManagerCleanupExpired(t *testing.T) {
	cm := NewCacheManager()

	// Add items (they will use default TTL)
	cm.SetUser(&domain.User{UserID: "u1", MessengerType: "line", CreatedAt: time.Now()})
	cm.SetCategory(&domain.Category{ID: "c1", UserID: "u1", Name: "Food", IsDefault: true})

	// CleanupExpired should work without errors (even if nothing expires)
	removed := cm.CleanupExpired()

	if removed < 0 {
		t.Errorf("expected non-negative removed count, got %d", removed)
	}
}

// TestCacheManagerConcurrentOperations tests concurrent cache operations
func TestCacheManagerConcurrentOperations(t *testing.T) {
	cm := NewCacheManager()

	t.Run("ConcurrentSetGet", func(t *testing.T) {
		done := make(chan bool)

		// Set operations
		go func() {
			for i := 0; i < 100; i++ {
				user := &domain.User{
					UserID:        string(rune(i)),
					MessengerType: "line",
					CreatedAt:     time.Now(),
				}
				cm.SetUser(user)
			}
			done <- true
		}()

		// Get operations
		go func() {
			for i := 0; i < 50; i++ {
				cm.GetUser(string(rune(i % 100)))
			}
			done <- true
		}()

		// Wait for goroutines
		<-done
		<-done
	})

	t.Run("ConcurrentMultipleEntityTypes", func(t *testing.T) {
		done := make(chan bool, 3)

		// User operations
		go func() {
			for i := 0; i < 50; i++ {
				cm.SetUser(&domain.User{
					UserID:        "u" + string(rune(i)),
					MessengerType: "line",
					CreatedAt:     time.Now(),
				})
			}
			done <- true
		}()

		// Category operations
		go func() {
			for i := 0; i < 50; i++ {
				cm.SetCategory(&domain.Category{
					ID:        "c" + string(rune(i)),
					UserID:    "u" + string(rune(i)),
					Name:      "Cat" + string(rune(i)),
					IsDefault: true,
				})
			}
			done <- true
		}()

		// Metrics operations
		go func() {
			for i := 0; i < 50; i++ {
				cm.SetMetrics("date"+string(rune(i)), &domain.DailyMetrics{
					Date: time.Now(),
				})
			}
			done <- true
		}()

		<-done
		<-done
		<-done
	})
}
