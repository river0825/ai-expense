package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"
)

// HealthStatus represents the overall system health
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck contains health check information
type HealthCheck struct {
	Status             HealthStatus              `json:"status"`
	Uptime             string                    `json:"uptime"`
	Database           DatabaseHealth            `json:"database"`
	Memory             MemoryHealth              `json:"memory"`
	OperationMetrics   map[string]interface{}    `json:"operation_metrics"`
	Timestamp          time.Time                 `json:"timestamp"`
}

// DatabaseHealth contains database health information
type DatabaseHealth struct {
	Status            string `json:"status"`
	Ping              bool   `json:"ping"`
	OpenConnections   int    `json:"open_connections"`
	InUseConnections  int    `json:"in_use_connections"`
	IdleConnections   int    `json:"idle_connections"`
	WaitCount         int64  `json:"wait_count"`
	WaitDuration      string `json:"wait_duration"`
	MaxIdleClosed     int64  `json:"max_idle_closed"`
	MaxLifetimeClosed int64  `json:"max_lifetime_closed"`
}

// MemoryHealth contains memory health information
type MemoryHealth struct {
	Alloc        string `json:"alloc"`
	TotalAlloc   string `json:"total_alloc"`
	Sys          string `json:"sys"`
	NumGC        uint32 `json:"num_gc"`
	Goroutines   int    `json:"goroutines"`
	LastGCTime   string `json:"last_gc_time"`
}

// HealthChecker checks system health
type HealthChecker struct {
	db        *sql.DB
	collector *MetricsCollector
	startTime time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, collector *MetricsCollector) *HealthChecker {
	return &HealthChecker{
		db:        db,
		collector: collector,
		startTime: time.Now(),
	}
}

// Check performs a health check
func (hc *HealthChecker) Check(ctx context.Context) *HealthCheck {
	check := &HealthCheck{
		Timestamp: time.Now(),
		OperationMetrics: hc.collector.GetSystemStats(),
	}

	// Calculate uptime
	check.Uptime = time.Since(hc.startTime).String()

	// Check database
	check.Database = hc.checkDatabase(ctx)

	// Check memory
	check.Memory = hc.checkMemory()

	// Determine overall status
	check.Status = hc.determineStatus(check)

	return check
}

// checkDatabase checks database health
func (hc *HealthChecker) checkDatabase(ctx context.Context) DatabaseHealth {
	health := DatabaseHealth{Status: "unknown"}

	if hc.db == nil {
		health.Status = "unavailable"
		return health
	}

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := hc.db.PingContext(ctx)
	health.Ping = err == nil
	if err != nil {
		health.Status = "unhealthy"
		return health
	}

	health.Status = "healthy"

	// Get connection stats
	stats := hc.db.Stats()
	health.OpenConnections = stats.OpenConnections
	health.InUseConnections = stats.InUse
	health.IdleConnections = stats.Idle
	health.WaitCount = stats.WaitCount
	health.WaitDuration = stats.WaitDuration.String()
	health.MaxIdleClosed = stats.MaxIdleClosed
	health.MaxLifetimeClosed = stats.MaxLifetimeClosed

	return health
}

// checkMemory checks memory health
func (hc *HealthChecker) checkMemory() MemoryHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health := MemoryHealth{
		Alloc:      formatBytes(m.Alloc),
		TotalAlloc: formatBytes(m.TotalAlloc),
		Sys:        formatBytes(m.Sys),
		NumGC:      m.NumGC,
		Goroutines: runtime.NumGoroutine(),
	}

	if m.LastGC != 0 {
		health.LastGCTime = time.Unix(0, int64(m.LastGC)).String()
	}

	return health
}

// determineStatus determines overall system status
func (hc *HealthChecker) determineStatus(check *HealthCheck) HealthStatus {
	// Database is critical
	if check.Database.Status == "unhealthy" || !check.Database.Ping {
		return HealthStatusUnhealthy
	}

	// Check memory health
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// If memory usage is very high (>500MB)
	if m.Alloc > 500*1024*1024 {
		return HealthStatusDegraded
	}

	// Check error rate from metrics
	sysStats := check.OperationMetrics
	if errorRate, ok := sysStats["error_rate"].(float64); ok {
		if errorRate > 5.0 { // >5% error rate
			return HealthStatusDegraded
		}
	}

	return HealthStatusHealthy
}

// formatBytes formats bytes to human readable format
func formatBytes(bytes uint64) string {
	units := []string{"B", "KB", "MB", "GB"}
	value := float64(bytes)

	for _, unit := range units {
		if value < 1024.0 {
			return formatFloat(value) + unit
		}
		value /= 1024.0
	}

	return formatFloat(value) + "TB"
}

// formatFloat formats float to 2 decimal places
func formatFloat(f float64) string {
	if f < 10 {
		return fmt.Sprintf("%.2f", f)
	} else if f < 100 {
		return fmt.Sprintf("%.1f", f)
	}
	return fmt.Sprintf("%.0f", f)
}
