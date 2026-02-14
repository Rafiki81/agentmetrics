package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/rafaelperezbeato/agentmetrics/internal/agent"
)

// AlertThresholds defines configurable alert thresholds
type AlertThresholds struct {
	CPUWarning      float64 // CPU% threshold for warning (default 80)
	CPUCritical     float64 // CPU% threshold for critical (default 95)
	MemoryWarning   float64 // Memory MB threshold for warning (default 500)
	MemoryCritical  float64 // Memory MB threshold for critical (default 1000)
	TokenWarning    int64   // Token count threshold for warning (default 500000)
	TokenCritical   int64   // Token count threshold for critical (default 2000000)
	CostWarning     float64 // Cost USD threshold for warning (default 1.0)
	CostCritical    float64 // Cost USD threshold for critical (default 5.0)
	IdleMinutes     int     // Minutes idle before alert (default 10)
	CooldownMinutes int     // Minutes between repeated alerts (default 5)
	MaxAlerts       int     // Maximum stored alerts (default 100)
}

// DefaultThresholds returns default alert thresholds
func DefaultThresholds() AlertThresholds {
	return AlertThresholds{
		CPUWarning:      80,
		CPUCritical:     95,
		MemoryWarning:   500,
		MemoryCritical:  1000,
		TokenWarning:    500000,
		TokenCritical:   2000000,
		CostWarning:     1.0,
		CostCritical:    5.0,
		IdleMinutes:     10,
		CooldownMinutes: 5,
		MaxAlerts:       100,
	}
}

// AlertMonitor checks agents against thresholds and generates alerts
type AlertMonitor struct {
	mu         sync.Mutex
	thresholds AlertThresholds
	alerts     []agent.Alert
	maxAlerts  int
	// Track what we've already alerted on to avoid spam
	alerted map[string]time.Time // key: "agentID:alertType" -> last alert time
}

// NewAlertMonitor creates a new alert monitor
func NewAlertMonitor(thresholds AlertThresholds) *AlertMonitor {
	maxAlerts := thresholds.MaxAlerts
	if maxAlerts <= 0 {
		maxAlerts = 100
	}
	return &AlertMonitor{
		thresholds: thresholds,
		alerts:     make([]agent.Alert, 0),
		maxAlerts:  maxAlerts,
		alerted:    make(map[string]time.Time),
	}
}

// Check evaluates an agent against thresholds and generates alerts
func (am *AlertMonitor) Check(a *agent.AgentInstance) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// CPU alerts
	if a.CPU >= am.thresholds.CPUCritical {
		am.addAlert(a, agent.AlertCritical, fmt.Sprintf("Critical CPU: %.1f%%", a.CPU), "cpu")
	} else if a.CPU >= am.thresholds.CPUWarning {
		am.addAlert(a, agent.AlertWarning, fmt.Sprintf("High CPU: %.1f%%", a.CPU), "cpu")
	}

	// Memory alerts
	if a.Memory >= am.thresholds.MemoryCritical {
		am.addAlert(a, agent.AlertCritical, fmt.Sprintf("Critical memory: %.1f MB", a.Memory), "mem")
	} else if a.Memory >= am.thresholds.MemoryWarning {
		am.addAlert(a, agent.AlertWarning, fmt.Sprintf("High memory: %.1f MB", a.Memory), "mem")
	}

	// Token alerts
	if a.Tokens.TotalTokens >= am.thresholds.TokenCritical {
		am.addAlert(a, agent.AlertCritical,
			fmt.Sprintf("Critical tokens: %s", FormatTokenCount(a.Tokens.TotalTokens)), "tokens")
	} else if a.Tokens.TotalTokens >= am.thresholds.TokenWarning {
		am.addAlert(a, agent.AlertWarning,
			fmt.Sprintf("High tokens: %s", FormatTokenCount(a.Tokens.TotalTokens)), "tokens")
	}

	// Cost alerts
	if a.Tokens.EstCost >= am.thresholds.CostCritical {
		am.addAlert(a, agent.AlertCritical,
			fmt.Sprintf("Critical cost: %s", FormatCost(a.Tokens.EstCost)), "cost")
	} else if a.Tokens.EstCost >= am.thresholds.CostWarning {
		am.addAlert(a, agent.AlertWarning,
			fmt.Sprintf("High cost: %s", FormatCost(a.Tokens.EstCost)), "cost")
	}

	// Idle alert
	if am.thresholds.IdleMinutes > 0 && !a.Session.LastActiveAt.IsZero() {
		idleDur := time.Since(a.Session.LastActiveAt).Minutes()
		if idleDur >= float64(am.thresholds.IdleMinutes) {
			am.addAlert(a, agent.AlertInfo,
				fmt.Sprintf("Agent idle for %.0f min", idleDur), "idle")
		}
	}
}

// addAlert adds an alert if we haven't alerted on this recently (cooldown from config)
func (am *AlertMonitor) addAlert(a *agent.AgentInstance, level agent.AlertLevel, msg, alertType string) {
	cooldown := time.Duration(am.thresholds.CooldownMinutes) * time.Minute
	if cooldown <= 0 {
		cooldown = 5 * time.Minute
	}
	key := a.Info.ID + ":" + alertType
	if last, ok := am.alerted[key]; ok {
		if time.Since(last) < cooldown {
			return // Cooldown period
		}
	}

	alert := agent.Alert{
		Timestamp: time.Now(),
		Level:     level,
		AgentID:   a.Info.ID,
		AgentName: a.Info.Name,
		Message:   msg,
	}

	am.alerts = append(am.alerts, alert)
	am.alerted[key] = time.Now()

	// Trim alerts
	if len(am.alerts) > am.maxAlerts {
		am.alerts = am.alerts[len(am.alerts)-am.maxAlerts:]
	}
}

// GetAlerts returns all alerts
func (am *AlertMonitor) GetAlerts() []agent.Alert {
	am.mu.Lock()
	defer am.mu.Unlock()
	result := make([]agent.Alert, len(am.alerts))
	copy(result, am.alerts)
	return result
}

// GetRecentAlerts returns alerts from the last N minutes
func (am *AlertMonitor) GetRecentAlerts(minutes int) []agent.Alert {
	am.mu.Lock()
	defer am.mu.Unlock()

	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var result []agent.Alert
	for _, a := range am.alerts {
		if a.Timestamp.After(cutoff) {
			result = append(result, a)
		}
	}
	return result
}

// AlertCount returns counts by level
func (am *AlertMonitor) AlertCount() (info, warning, critical int) {
	am.mu.Lock()
	defer am.mu.Unlock()
	for _, a := range am.alerts {
		switch a.Level {
		case agent.AlertInfo:
			info++
		case agent.AlertWarning:
			warning++
		case agent.AlertCritical:
			critical++
		}
	}
	return
}
