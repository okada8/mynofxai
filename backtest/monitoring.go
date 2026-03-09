package backtest

import (
	"sync"
	"time"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	LevelInfo     AlertLevel = "INFO"     // Informational, no immediate action needed
	LevelWarning  AlertLevel = "WARNING"  // Warning, attention required
	LevelCritical AlertLevel = "CRITICAL" // Critical, immediate action required
)

// AlertAction suggests actions to be taken
type AlertAction string

const (
	ActionPauseTrading AlertAction = "PAUSE_TRADING"
	ActionReduceRisk   AlertAction = "REDUCE_RISK"
	ActionCloseAll     AlertAction = "CLOSE_ALL"
	ActionCheckLogs    AlertAction = "CHECK_LOGS"
)

// Alert represents a monitoring alert
type Alert struct {
	Level     AlertLevel    `json:"level"`
	Title     string        `json:"title"`
	Message   string        `json:"message"`
	Metric    string        `json:"metric,omitempty"`
	Value     float64       `json:"value,omitempty"`
	Threshold float64       `json:"threshold,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Actions   []AlertAction `json:"actions,omitempty"`
}

// AlertManager manages alerting logic
type AlertManager struct {
	alerts []Alert
	mu     sync.RWMutex
	// Callback for real-time notification (e.g. WebSocket, Email, Slack)
	NotifyFunc func(Alert)
}

// NewAlertManager creates a new AlertManager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts: make([]Alert, 0),
	}
}

// AddAlert adds a new alert and triggers notification
func (am *AlertManager) AddAlert(level AlertLevel, title, message string, metric string, val, threshold float64, actions ...AlertAction) {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert := Alert{
		Level:     level,
		Title:     title,
		Message:   message,
		Metric:    metric,
		Value:     val,
		Threshold: threshold,
		Timestamp: time.Now(),
		Actions:   actions,
	}

	am.alerts = append(am.alerts, alert)
	
	// Keep history bounded
	if len(am.alerts) > 1000 {
		am.alerts = am.alerts[1:]
	}

	if am.NotifyFunc != nil {
		go am.NotifyFunc(alert)
	}
}

// GetRecentAlerts returns recent alerts
func (am *AlertManager) GetRecentAlerts(limit int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if limit <= 0 || limit > len(am.alerts) {
		limit = len(am.alerts)
	}
	
	// Return copy in reverse order (newest first)
	result := make([]Alert, limit)
	for i := 0; i < limit; i++ {
		result[i] = am.alerts[len(am.alerts)-1-i]
	}
	return result
}

// DashboardMetrics represents real-time dashboard data
type DashboardMetrics struct {
	// Capital
	TotalEquity    float64 `json:"total_equity"`
	DailyPnL       float64 `json:"daily_pnl"`
	DailyPnLPct    float64 `json:"daily_pnl_pct"`
	MaxDrawdown    float64 `json:"max_drawdown"`
	SharpeRatio    float64 `json:"sharpe_ratio"`

	// Risk
	CurrentVaR    float64 `json:"current_var"`
	PositionCount int     `json:"position_count"`
	LeverageUsed  float64 `json:"leverage_used"`

	// Execution
	AvgSlippage   float64 `json:"avg_slippage"`
	OrderFillRate float64 `json:"order_fill_rate"`
	APIErrorRate  float64 `json:"api_error_rate"`

	// AI
	AvgConfidence    float64 `json:"avg_confidence"`
	DecisionAccuracy float64 `json:"decision_accuracy"`
}

// BreakerState represents the state of the circuit breaker
type BreakerState string

const (
	StateClosed   BreakerState = "CLOSED"    // Normal state
	StateOpen     BreakerState = "OPEN"      // Open state, requests rejected
	StateHalfOpen BreakerState = "HALF_OPEN" // Half-open, probing
)

// CircuitBreaker implements the circuit breaker pattern for system stability
type CircuitBreaker struct {
	Name            string
	State           BreakerState
	FailureCount    int
	FailureThreshold int
	LastFailureTime time.Time
	ResetTimeout    time.Duration
	
	// Specific counters
	ApiFailures        int
	LowConfidenceCount int
	
	mu sync.Mutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		Name:             name,
		State:            StateClosed,
		FailureThreshold: threshold,
		ResetTimeout:     timeout,
	}
}

// ReportSuccess reports a successful operation
func (cb *CircuitBreaker) ReportSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.State == StateHalfOpen {
		cb.State = StateClosed
		cb.FailureCount = 0
		cb.ApiFailures = 0
		cb.LowConfidenceCount = 0
	} else if cb.State == StateClosed {
		// Reset counters on success if we were accumulating errors?
		// Usually we reset only after some success streak or time.
		// For simplicity, we can decay failure count or just keep it until threshold.
		// Here we reset API failures on success to count CONSECUTIVE failures.
		cb.ApiFailures = 0
		cb.LowConfidenceCount = 0 // Reset consecutive low confidence
	}
}

// ReportFailure reports a failed operation
func (cb *CircuitBreaker) ReportFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.State == StateOpen {
		return
	}

	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.FailureCount >= cb.FailureThreshold {
		cb.State = StateOpen
	}
}

// ReportAPIFailure specific for API issues
func (cb *CircuitBreaker) ReportAPIFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.ApiFailures++
	if cb.ApiFailures >= 3 { // 3 consecutive API failures
		cb.State = StateOpen
		cb.LastFailureTime = time.Now()
	}
}

// ReportLowConfidence specific for AI issues
func (cb *CircuitBreaker) ReportLowConfidence() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.LowConfidenceCount++
	if cb.LowConfidenceCount >= 5 { // 5 consecutive low confidence decisions
		cb.State = StateOpen
		cb.LastFailureTime = time.Now()
	}
}

// ReportLoss specific for financial triggers
func (cb *CircuitBreaker) ReportLoss(lossPct float64) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	// Single trade loss > 1% -> Trip? Maybe just warning.
	// But requirements say: "Single loss > 1%" -> Warning (AlertManager), not necessarily Breaker.
	// "10 min loss > 3%" -> Breaker.
	
	// Here we implement immediate triggers for critical losses
	if lossPct > 0.03 { // > 3% loss
		cb.State = StateOpen
		cb.LastFailureTime = time.Now()
	}
}

// AllowRequest checks if request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.State == StateClosed {
		return true
	}

	if cb.State == StateOpen {
		if time.Since(cb.LastFailureTime) > cb.ResetTimeout {
			cb.State = StateHalfOpen
			return true // Allow probe
		}
		return false
	}

	// HalfOpen
	return true // Allow probe
}
