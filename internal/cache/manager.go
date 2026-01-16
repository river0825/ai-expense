package cache

import (
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// CacheManager manages caching for domain entities
type CacheManager struct {
	// User cache: key = user_id
	users *LRUCache[string, *domain.User]

	// Category cache: key = category_id
	categories *LRUCache[string, *domain.Category]

	// User categories cache: key = user_id
	userCategories *LRUCache[string, []*domain.Category]

	// Category keyword mappings: key = category_id
	keywords *LRUCache[string, []*domain.CategoryKeyword]

	// Metrics cache: key = date_string (YYYY-MM-DD)
	metrics *LRUCache[string, *domain.DailyMetrics]
}

// DefaultCacheSizes defines default cache sizes for each entity
var DefaultCacheSizes = struct {
	Users          int
	Categories     int
	UserCategories int
	Keywords       int
	Metrics        int
}{
	Users:          1000,   // Cache 1000 users
	Categories:     5000,   // Cache 5000 categories
	UserCategories: 1000,   // Cache categories for 1000 users
	Keywords:       10000,  // Cache keywords for categories
	Metrics:        365,    // Cache 1 year of daily metrics
}

// DefaultTTLs defines default time-to-live for cached items
var DefaultTTLs = struct {
	Users          time.Duration
	Categories     time.Duration
	UserCategories time.Duration
	Keywords       time.Duration
	Metrics        time.Duration
}{
	Users:          1 * time.Hour,       // Users cached for 1 hour
	Categories:     30 * time.Minute,    // Categories cached for 30 min
	UserCategories: 15 * time.Minute,    // User categories cached for 15 min
	Keywords:       1 * time.Hour,       // Keywords cached for 1 hour
	Metrics:        24 * time.Hour,      // Daily metrics cached for 24 hours
}

// NewCacheManager creates a new cache manager with default sizes
func NewCacheManager() *CacheManager {
	return &CacheManager{
		users:          NewLRUCache[string, *domain.User](DefaultCacheSizes.Users),
		categories:     NewLRUCache[string, *domain.Category](DefaultCacheSizes.Categories),
		userCategories: NewLRUCache[string, []*domain.Category](DefaultCacheSizes.UserCategories),
		keywords:       NewLRUCache[string, []*domain.CategoryKeyword](DefaultCacheSizes.Keywords),
		metrics:        NewLRUCache[string, *domain.DailyMetrics](DefaultCacheSizes.Metrics),
	}
}

// User cache operations

// GetUser retrieves a user from cache
func (cm *CacheManager) GetUser(userID string) (*domain.User, bool) {
	return cm.users.Get(userID)
}

// SetUser caches a user
func (cm *CacheManager) SetUser(user *domain.User) {
	cm.users.SetWithTTL(user.UserID, user, DefaultTTLs.Users)
}

// InvalidateUser removes a user from cache
func (cm *CacheManager) InvalidateUser(userID string) {
	cm.users.Delete(userID)
}

// Category cache operations

// GetCategory retrieves a category from cache
func (cm *CacheManager) GetCategory(categoryID string) (*domain.Category, bool) {
	return cm.categories.Get(categoryID)
}

// SetCategory caches a category
func (cm *CacheManager) SetCategory(category *domain.Category) {
	cm.categories.SetWithTTL(category.ID, category, DefaultTTLs.Categories)
	// Invalidate user categories cache when category changes
	cm.userCategories.Delete(category.UserID)
}

// InvalidateCategory removes a category from cache
func (cm *CacheManager) InvalidateCategory(categoryID string, userID string) {
	cm.categories.Delete(categoryID)
	cm.userCategories.Delete(userID)
}

// User categories cache operations

// GetUserCategories retrieves user's categories from cache
func (cm *CacheManager) GetUserCategories(userID string) ([]*domain.Category, bool) {
	return cm.userCategories.Get(userID)
}

// SetUserCategories caches user's categories
func (cm *CacheManager) SetUserCategories(userID string, categories []*domain.Category) {
	cm.userCategories.SetWithTTL(userID, categories, DefaultTTLs.UserCategories)
}

// InvalidateUserCategories removes user's categories from cache
func (cm *CacheManager) InvalidateUserCategories(userID string) {
	cm.userCategories.Delete(userID)
}

// Keywords cache operations

// GetCategoryKeywords retrieves keywords for a category from cache
func (cm *CacheManager) GetCategoryKeywords(categoryID string) ([]*domain.CategoryKeyword, bool) {
	return cm.keywords.Get(categoryID)
}

// SetCategoryKeywords caches keywords for a category
func (cm *CacheManager) SetCategoryKeywords(categoryID string, keywords []*domain.CategoryKeyword) {
	cm.keywords.SetWithTTL(categoryID, keywords, DefaultTTLs.Keywords)
}

// InvalidateCategoryKeywords removes keywords for a category from cache
func (cm *CacheManager) InvalidateCategoryKeywords(categoryID string) {
	cm.keywords.Delete(categoryID)
}

// Metrics cache operations

// GetMetrics retrieves metrics from cache
func (cm *CacheManager) GetMetrics(dateKey string) (*domain.DailyMetrics, bool) {
	return cm.metrics.Get(dateKey)
}

// SetMetrics caches metrics
func (cm *CacheManager) SetMetrics(dateKey string, metrics *domain.DailyMetrics) {
	cm.metrics.SetWithTTL(dateKey, metrics, DefaultTTLs.Metrics)
}

// InvalidateMetrics removes metrics from cache
func (cm *CacheManager) InvalidateMetrics(dateKey string) {
	cm.metrics.Delete(dateKey)
}

// Global cache management

// Clear clears all caches
func (cm *CacheManager) Clear() {
	cm.users.Clear()
	cm.categories.Clear()
	cm.userCategories.Clear()
	cm.keywords.Clear()
	cm.metrics.Clear()
}

// CleanupExpired removes all expired entries from all caches
func (cm *CacheManager) CleanupExpired() int {
	removed := 0
	removed += cm.users.CleanupExpired()
	removed += cm.categories.CleanupExpired()
	removed += cm.userCategories.CleanupExpired()
	removed += cm.keywords.CleanupExpired()
	removed += cm.metrics.CleanupExpired()
	return removed
}

// Stats returns statistics for all caches
func (cm *CacheManager) Stats() map[string]interface{} {
	return map[string]interface{}{
		"users":           cm.users.Stats(),
		"categories":      cm.categories.Stats(),
		"user_categories": cm.userCategories.Stats(),
		"keywords":        cm.keywords.Stats(),
		"metrics":         cm.metrics.Stats(),
		"total_size":      cm.users.Size() + cm.categories.Size() + cm.userCategories.Size() + cm.keywords.Size() + cm.metrics.Size(),
	}
}

// InvalidateUserData clears all data related to a user
func (cm *CacheManager) InvalidateUserData(userID string) {
	cm.InvalidateUser(userID)
	cm.InvalidateUserCategories(userID)
}
