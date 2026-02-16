package cli

import (
	"fmt"

	"github.com/Rafiki81/libagentmetrics/agent"
	"github.com/Rafiki81/libagentmetrics/monitor"
)

func runAlerts() error {
	runtime := newScanRuntime()

	agents, err := runtime.scan()
	if err != nil {
		return err
	}

	collectTokenMetrics(agents)
	collectSessionMetrics(agents)

	alertMon := monitor.NewAlertMonitor(monitor.AlertThresholds{
		CPUWarning:      runtime.cfg.Alerts.CPUWarning,
		CPUCritical:     runtime.cfg.Alerts.CPUCritical,
		MemoryWarning:   runtime.cfg.Alerts.MemoryWarning,
		MemoryCritical:  runtime.cfg.Alerts.MemoryCritical,
		TokenWarning:    runtime.cfg.Alerts.TokenWarning,
		TokenCritical:   runtime.cfg.Alerts.TokenCritical,
		CostWarning:     runtime.cfg.Alerts.CostWarning,
		CostCritical:    runtime.cfg.Alerts.CostCritical,
		IdleMinutes:     runtime.cfg.Alerts.IdleMinutes,
		CooldownMinutes: runtime.cfg.Alerts.CooldownMinutes,
		MaxAlerts:       runtime.cfg.Alerts.MaxAlerts,
	})
	for i := range agents {
		alertMon.Check(&agents[i])
	}

	alerts := alertMon.GetAlerts()
	if len(alerts) == 0 {
		fmt.Println("✅ No active alerts.")
		return nil
	}

	fmt.Printf("⚡ %d active alert(s):\n\n", len(alerts))
	for _, al := range alerts {
		icon := "ℹ"
		switch al.Level {
		case agent.AlertWarning:
			icon = "⚠"
		case agent.AlertCritical:
			icon = "🔴"
		}
		fmt.Printf("  %s %s [%s] %s — %s\n",
			al.Timestamp.Format("15:04:05"),
			icon,
			al.Level,
			al.AgentName,
			al.Message,
		)
	}

	return nil
}
