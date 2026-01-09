package monitor

import (
	"fmt"
	"sync"
	"time"
)

// MetricsCollector collects and exports system metrics for monitoring
type MetricsCollector struct {
	mu sync.RWMutex

	// Signal metrics
	signalsGenerated int64
	signalsAccepted  int64
	signalsRejected  int64

	// Performance metrics
	scanLatency []time.Duration
	aiLatency   []time.Duration
	totalScans  int64

	// Cost metrics
	aiCostTotal  float64
	aiTokensUsed int64

	// ML metrics
	predictionAccuracy float64
	predictionMAPE     float64
	modelRetrains      int64

	// Risk metrics
	dailyLoss         float64
	consecutiveLosses int
	vetoCount         int64

	// System metrics
	uptime        time.Time
	errorsTotal   int64
	apiCallsTotal int64
	apiFailures   int64
}

// Metric represents a single metric value
type Metric struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		uptime:      time.Now(),
		scanLatency: make([]time.Duration, 0, 1000),
		aiLatency:   make([]time.Duration, 0, 1000),
	}
}

// RecordSignalGenerated records a generated signal
func (m *MetricsCollector) RecordSignalGenerated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signalsGenerated++
}

// RecordSignalAccepted records an accepted signal
func (m *MetricsCollector) RecordSignalAccepted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signalsAccepted++
}

// RecordSignalRejected records a rejected signal
func (m *MetricsCollector) RecordSignalRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signalsRejected++
}

// RecordScanLatency records scan latency
func (m *MetricsCollector) RecordScanLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scanLatency = append(m.scanLatency, latency)
	m.totalScans++
}

// RecordAILatency records AI analysis latency
func (m *MetricsCollector) RecordAILatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aiLatency = append(m.aiLatency, latency)
}

// RecordAICost records AI usage cost
func (m *MetricsCollector) RecordAICost(cost float64, tokens int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aiCostTotal += cost
	m.aiTokensUsed += tokens
}

// RecordPredictionMetrics records ML prediction metrics
func (m *MetricsCollector) RecordPredictionMetrics(accuracy, mape float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.predictionAccuracy = accuracy
	m.predictionMAPE = mape
}

// RecordModelRetrain records model retraining event
func (m *MetricsCollector) RecordModelRetrain() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelRetrains++
}

// RecordDailyLoss records daily loss
func (m *MetricsCollector) RecordDailyLoss(loss float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dailyLoss = loss
}

// RecordConsecutiveLoss records consecutive loss
func (m *MetricsCollector) RecordConsecutiveLoss(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consecutiveLosses = count
}

// RecordVeto records veto event
func (m *MetricsCollector) RecordVeto() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.vetoCount++
}

// RecordError records an error
func (m *MetricsCollector) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorsTotal++
}

// RecordAPICall records an API call
func (m *MetricsCollector) RecordAPICall(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.apiCallsTotal++
	if !success {
		m.apiFailures++
	}
}

// GetMetrics returns all current metrics in Prometheus format
func (m *MetricsCollector) GetMetrics() []Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := []Metric{
		// Signal metrics
		{Name: "alpha_signals_generated_total", Value: float64(m.signalsGenerated), Timestamp: time.Now()},
		{Name: "alpha_signals_accepted_total", Value: float64(m.signalsAccepted), Timestamp: time.Now()},
		{Name: "alpha_signals_rejected_total", Value: float64(m.signalsRejected), Timestamp: time.Now()},

		// Performance metrics
		{Name: "alpha_scan_latency_seconds", Value: m.avgDuration(m.scanLatency).Seconds(), Timestamp: time.Now()},
		{Name: "alpha_ai_latency_seconds", Value: m.avgDuration(m.aiLatency).Seconds(), Timestamp: time.Now()},
		{Name: "alpha_scans_total", Value: float64(m.totalScans), Timestamp: time.Now()},

		// Cost metrics
		{Name: "alpha_ai_cost_yuan_total", Value: m.aiCostTotal, Timestamp: time.Now()},
		{Name: "alpha_ai_tokens_used_total", Value: float64(m.aiTokensUsed), Timestamp: time.Now()},

		// ML metrics
		{Name: "alpha_prediction_accuracy", Value: m.predictionAccuracy, Timestamp: time.Now()},
		{Name: "alpha_prediction_mape", Value: m.predictionMAPE, Timestamp: time.Now()},
		{Name: "alpha_model_retrains_total", Value: float64(m.modelRetrains), Timestamp: time.Now()},

		// Risk metrics
		{Name: "alpha_daily_loss_yuan", Value: m.dailyLoss, Timestamp: time.Now()},
		{Name: "alpha_consecutive_losses", Value: float64(m.consecutiveLosses), Timestamp: time.Now()},
		{Name: "alpha_veto_count_total", Value: float64(m.vetoCount), Timestamp: time.Now()},

		// System metrics
		{Name: "alpha_uptime_seconds", Value: time.Since(m.uptime).Seconds(), Timestamp: time.Now()},
		{Name: "alpha_errors_total", Value: float64(m.errorsTotal), Timestamp: time.Now()},
		{Name: "alpha_api_calls_total", Value: float64(m.apiCallsTotal), Timestamp: time.Now()},
		{Name: "alpha_api_failures_total", Value: float64(m.apiFailures), Timestamp: time.Now()},
	}

	// Calculate derived metrics
	if m.signalsGenerated > 0 {
		acceptanceRate := float64(m.signalsAccepted) / float64(m.signalsGenerated) * 100
		metrics = append(metrics, Metric{
			Name:      "alpha_signal_acceptance_rate",
			Value:     acceptanceRate,
			Timestamp: time.Now(),
		})
	}

	if m.apiCallsTotal > 0 {
		successRate := float64(m.apiCallsTotal-m.apiFailures) / float64(m.apiCallsTotal) * 100
		metrics = append(metrics, Metric{
			Name:      "alpha_api_success_rate",
			Value:     successRate,
			Timestamp: time.Now(),
		})
	}

	return metrics
}

// avgDuration calculates average duration
func (m *MetricsCollector) avgDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

// ExportPrometheus exports metrics in Prometheus text format
func (m *MetricsCollector) ExportPrometheus() string {
	metrics := m.GetMetrics()
	output := ""

	for _, metric := range metrics {
		// Format: metric_name{label="value"} value timestamp
		output += metric.Name
		if len(metric.Labels) > 0 {
			output += "{"
			first := true
			for k, v := range metric.Labels {
				if !first {
					output += ","
				}
				output += k + `="` + v + `"`
				first = false
			}
			output += "}"
		}
		output += " " + fmt.Sprintf("%.6f", metric.Value) + " " + fmt.Sprintf("%d", metric.Timestamp.UnixMilli()) + "\n"
	}

	return output
}
