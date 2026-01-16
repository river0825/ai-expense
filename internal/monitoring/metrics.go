package monitoring

import (
	"sync"
	"sync/atomic"
	"time"
)

// OperationMetrics tracks metrics for a specific operation
type OperationMetrics struct {
	Name           string
	Count          int64
	TotalDuration  int64 // nanoseconds
	MinDuration    int64
	MaxDuration    int64
	ErrorCount     int64
	LastRecordedAt time.Time
	mu             sync.RWMutex
}

// RecordOperation records an operation execution
func (om *OperationMetrics) RecordOperation(duration time.Duration, err error) {
	durationNs := duration.Nanoseconds()

	atomic.AddInt64(&om.Count, 1)
	atomic.AddInt64(&om.TotalDuration, durationNs)

	if err != nil {
		atomic.AddInt64(&om.ErrorCount, 1)
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	if om.MinDuration == 0 || durationNs < om.MinDuration {
		om.MinDuration = durationNs
	}
	if durationNs > om.MaxDuration {
		om.MaxDuration = durationNs
	}
	om.LastRecordedAt = time.Now()
}

// GetStats returns current statistics
func (om *OperationMetrics) GetStats() map[string]interface{} {
	om.mu.RLock()
	defer om.mu.RUnlock()

	count := atomic.LoadInt64(&om.Count)
	totalDur := atomic.LoadInt64(&om.TotalDuration)
	errCount := atomic.LoadInt64(&om.ErrorCount)

	var avgDur int64
	if count > 0 {
		avgDur = totalDur / count
	}

	errorRate := 0.0
	if count > 0 {
		errorRate = float64(errCount) / float64(count) * 100
	}

	return map[string]interface{}{
		"name":             om.Name,
		"count":            count,
		"avg_duration_ms":  float64(avgDur) / 1e6,
		"min_duration_ms":  float64(om.MinDuration) / 1e6,
		"max_duration_ms":  float64(om.MaxDuration) / 1e6,
		"total_duration":   time.Duration(totalDur).String(),
		"error_count":      errCount,
		"error_rate":       errorRate,
		"last_recorded":    om.LastRecordedAt,
	}
}

// Reset clears all metrics
func (om *OperationMetrics) Reset() {
	om.mu.Lock()
	defer om.mu.Unlock()

	atomic.StoreInt64(&om.Count, 0)
	atomic.StoreInt64(&om.TotalDuration, 0)
	atomic.StoreInt64(&om.ErrorCount, 0)
	om.MinDuration = 0
	om.MaxDuration = 0
}

// MetricsCollector collects metrics for all operations
type MetricsCollector struct {
	operations map[string]*OperationMetrics
	mu         sync.RWMutex
	startTime  time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		operations: make(map[string]*OperationMetrics),
		startTime:  time.Now(),
	}
}

// GetOrCreateMetric gets or creates an operation metric
func (mc *MetricsCollector) GetOrCreateMetric(name string) *OperationMetrics {
	mc.mu.RLock()
	if metric, exists := mc.operations[name]; exists {
		mc.mu.RUnlock()
		return metric
	}
	mc.mu.RUnlock()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Double-check after acquiring write lock
	if metric, exists := mc.operations[name]; exists {
		return metric
	}

	metric := &OperationMetrics{Name: name}
	mc.operations[name] = metric
	return metric
}

// RecordOperation records an operation
func (mc *MetricsCollector) RecordOperation(name string, duration time.Duration, err error) {
	metric := mc.GetOrCreateMetric(name)
	metric.RecordOperation(duration, err)
}

// GetMetrics returns all collected metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := make(map[string]interface{})
	for name, operation := range mc.operations {
		metrics[name] = operation.GetStats()
	}

	return metrics
}

// GetMetricByName returns metrics for a specific operation
func (mc *MetricsCollector) GetMetricByName(name string) map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if metric, exists := mc.operations[name]; exists {
		return metric.GetStats()
	}
	return nil
}

// GetSystemStats returns overall system statistics
func (mc *MetricsCollector) GetSystemStats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalOps int64
	var totalErrors int64
	var avgLatency int64
	var maxLatency int64

	for _, metric := range mc.operations {
		count := atomic.LoadInt64(&metric.Count)
		errCount := atomic.LoadInt64(&metric.ErrorCount)
		totalDur := atomic.LoadInt64(&metric.TotalDuration)

		totalOps += count
		totalErrors += errCount
		if totalDur > 0 && count > 0 {
			avgLatency += totalDur / count
		}
		if metric.MaxDuration > maxLatency {
			maxLatency = metric.MaxDuration
		}
	}

	var avgErrorRate float64
	if len(mc.operations) > 0 && totalOps > 0 {
		avgErrorRate = float64(totalErrors) / float64(totalOps) * 100
	}

	operationCount := len(mc.operations)
	if operationCount > 0 {
		avgLatency /= int64(operationCount)
	}

	uptime := time.Since(mc.startTime)

	return map[string]interface{}{
		"uptime":              uptime.String(),
		"total_operations":    totalOps,
		"total_errors":        totalErrors,
		"error_rate":          avgErrorRate,
		"avg_latency_ms":      float64(avgLatency) / 1e6,
		"max_latency_ms":      float64(maxLatency) / 1e6,
		"monitored_operations": operationCount,
	}
}

// Reset clears all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, metric := range mc.operations {
		metric.Reset()
	}
	mc.startTime = time.Now()
}

// OperationTimer is a helper for timing operations
type OperationTimer struct {
	collector *MetricsCollector
	name      string
	startTime time.Time
}

// NewOperationTimer creates a new operation timer
func NewOperationTimer(collector *MetricsCollector, name string) *OperationTimer {
	return &OperationTimer{
		collector: collector,
		name:      name,
		startTime: time.Now(),
	}
}

// End records the operation end time
func (ot *OperationTimer) End(err error) {
	duration := time.Since(ot.startTime)
	ot.collector.RecordOperation(ot.name, duration, err)
}

// EndOK records successful operation end
func (ot *OperationTimer) EndOK() {
	ot.End(nil)
}

// ContextOperations defines common operation names for monitoring
const (
	OpAutoSignup             = "auto_signup"
	OpCreateExpense          = "create_expense"
	OpGetExpenses            = "get_expenses"
	OpGetExpensesByDateRange = "get_expenses_by_date_range"
	OpParseConversation      = "parse_conversation"
	OpSuggestCategory        = "suggest_category"
	OpGetMetrics             = "get_metrics"
	OpUserExists             = "user_exists"
	OpCategoryLookup         = "category_lookup"
	OpWebhookLineMessage     = "webhook_line_message"
	OpWebhookTelegramMessage = "webhook_telegram_message"
	OpWebhookSlackMessage    = "webhook_slack_message"
	OpWebhookTeamsMessage    = "webhook_teams_message"
	OpWebhookDiscordMessage  = "webhook_discord_message"
	OpWebhookWhatsAppMessage = "webhook_whatsapp_message"
	OpCacheGet               = "cache_get"
	OpCacheSet               = "cache_set"
	OpDBQuery                = "db_query"
	OpDBInsert               = "db_insert"
	OpDBUpdate               = "db_update"
)
