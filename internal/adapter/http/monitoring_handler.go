package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/monitoring"
)

// MonitoringHandler handles monitoring endpoints
type MonitoringHandler struct {
	healthChecker *monitoring.HealthChecker
	collector     *monitoring.MetricsCollector
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(healthChecker *monitoring.HealthChecker, collector *monitoring.MetricsCollector) *MonitoringHandler {
	return &MonitoringHandler{
		healthChecker: healthChecker,
		collector:     collector,
	}
}

// Health returns system health status
func (mh *MonitoringHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := mh.healthChecker.Check(ctx)

	w.Header().Set("Content-Type", "application/json")

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if health.Status == monitoring.HealthStatusDegraded {
		statusCode = http.StatusOK // Still OK but indicates degradation
	} else if health.Status == monitoring.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

// Metrics returns all collected metrics
func (mh *MonitoringHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	metrics := mh.collector.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// SystemMetrics returns system-wide metrics summary
func (mh *MonitoringHandler) SystemMetrics(w http.ResponseWriter, r *http.Request) {
	stats := mh.collector.GetSystemStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// OperationMetrics returns metrics for a specific operation
func (mh *MonitoringHandler) OperationMetrics(w http.ResponseWriter, r *http.Request) {
	operationName := r.URL.Query().Get("operation")
	if operationName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "operation parameter required"})
		return
	}

	metrics := mh.collector.GetMetricByName(operationName)
	if metrics == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "operation not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// ReadinessProbe returns readiness status
func (mh *MonitoringHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	health := mh.healthChecker.Check(ctx)

	statusCode := http.StatusOK
	if health.Status != monitoring.HealthStatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready":  health.Status == monitoring.HealthStatusHealthy,
		"status": health.Status,
	})
}

// LivenessProbe returns liveness status
func (mh *MonitoringHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"alive": true})
}

// ResetMetrics resets all metrics (admin only)
func (mh *MonitoringHandler) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "POST method required"})
		return
	}

	mh.collector.Reset()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "metrics reset"})
}
