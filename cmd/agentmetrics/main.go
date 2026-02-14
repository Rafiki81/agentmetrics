package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rafaelperezbeato/agentmetrics/internal/agent"
	"github.com/rafaelperezbeato/agentmetrics/internal/config"
	"github.com/rafaelperezbeato/agentmetrics/internal/monitor"
	"github.com/rafaelperezbeato/agentmetrics/internal/tui"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		runTUI()
		return
	}

	switch os.Args[1] {
	case "scan", "s":
		runScan()
	case "watch", "w":
		runWatch()
	case "json":
		runJSON()
	case "export":
		runExport()
	case "alerts":
		runAlerts()
	case "config", "c":
		runConfig()
	case "version", "v", "--version":
		fmt.Printf("agentmetrics v%s\n", version)
	case "help", "h", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func runTUI() {
	cfg := config.Load()
	if err := tui.StartApp(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runScan() {
	cfg := config.Load()
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)

	agents, err := detector.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning agents: %v\n", err)
		os.Exit(1)
	}

	if len(agents) == 0 {
		fmt.Println("No active AI agents detected.")
		fmt.Println("\nSupported agents:")
		for _, a := range registry.Agents {
			fmt.Printf("  - %s (%s)\n", a.Name, a.Description)
		}
		return
	}

	// Collect token metrics (includes cost)
	tokenMon := monitor.NewTokenMonitor()
	tokenMon.Collect(agents)

	// Collect git activity
	gitMon := monitor.NewGitMonitor()
	sessionMon := monitor.NewSessionMonitor()
	for i := range agents {
		gitMon.Collect(&agents[i])
		sessionMon.Collect(&agents[i])
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "AGENT\tSTATUS\tPID\tCPU%%\tMEMORY\tTOKENS\tCOST\tREQS\tMODEL\tBRANCH\tDIRECTORY\n")
	fmt.Fprintf(w, "-----\t------\t---\t----\t------\t------\t----\t----\t-----\t------\t---------\n")

	for _, a := range agents {
		fmt.Fprintf(w, "%s\t%s\t%d\t%.1f%%\t%.1f MB\t%s\t%s\t%d\t%s\t%s\t%s\n",
			a.Info.Name,
			a.Status.String(),
			a.PID,
			a.CPU,
			a.Memory,
			monitor.FormatTokenCount(a.Tokens.TotalTokens),
			monitor.FormatCost(a.Tokens.EstCost),
			a.Tokens.RequestCount,
			a.Tokens.LastModel,
			a.Git.Branch,
			a.WorkDir,
		)
	}
	w.Flush()

	netMon := monitor.NewNetworkMonitor()
	fmt.Println("\nNetwork Connections:")
	for _, a := range agents {
		conns := netMon.GetConnections(a.PID)
		if len(conns) > 0 {
			fmt.Printf("  %s (PID %d):\n", a.Info.Name, a.PID)
			for _, conn := range conns {
				fmt.Printf("    %s\n", monitor.DescribeConnection(conn))
			}
		}
	}
}

func runExport() {
	format := "json"
	path := ""

	if len(os.Args) > 2 {
		format = os.Args[2]
	}
	if len(os.Args) > 3 {
		path = os.Args[3]
	}

	// Do a scan and record to generate some data
	cfg := config.Load()
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)

	agents, err := detector.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning agents: %v\n", err)
		os.Exit(1)
	}

	// Collect all metrics
	tokenMon := monitor.NewTokenMonitor()
	tokenMon.Collect(agents)

	gitMon := monitor.NewGitMonitor()
	sessionMon := monitor.NewSessionMonitor()
	for i := range agents {
		gitMon.Collect(&agents[i])
		sessionMon.Collect(&agents[i])
	}

	history := monitor.NewHistoryStore(cfg.Export.Directory, cfg.Export.MaxHistory)
	history.Record(agents)

	switch format {
	case "json":
		if err := history.ExportJSON(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting JSON: %v\n", err)
			os.Exit(1)
		}
		if path == "" {
			fmt.Printf("Exported to: %s/\n", history.DataDir())
		} else {
			fmt.Printf("Exported to: %s\n", path)
		}
	case "csv":
		if err := history.ExportCSV(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting CSV: %v\n", err)
			os.Exit(1)
		}
		if path == "" {
			fmt.Printf("Exported to: %s/\n", history.DataDir())
		} else {
			fmt.Printf("Exported to: %s\n", path)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (use 'json' or 'csv')\n", format)
		os.Exit(1)
	}
}

func runAlerts() {
	cfg := config.Load()
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)

	agents, err := detector.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning agents: %v\n", err)
		os.Exit(1)
	}

	// Collect metrics needed for alerts
	tokenMon := monitor.NewTokenMonitor()
	tokenMon.Collect(agents)

	sessionMon := monitor.NewSessionMonitor()
	for i := range agents {
		sessionMon.Collect(&agents[i])
	}

	alertMon := monitor.NewAlertMonitor(monitor.AlertThresholds{
		CPUWarning:      cfg.Alerts.CPUWarning,
		CPUCritical:     cfg.Alerts.CPUCritical,
		MemoryWarning:   cfg.Alerts.MemoryWarning,
		MemoryCritical:  cfg.Alerts.MemoryCritical,
		TokenWarning:    cfg.Alerts.TokenWarning,
		TokenCritical:   cfg.Alerts.TokenCritical,
		CostWarning:     cfg.Alerts.CostWarning,
		CostCritical:    cfg.Alerts.CostCritical,
		IdleMinutes:     cfg.Alerts.IdleMinutes,
		CooldownMinutes: cfg.Alerts.CooldownMinutes,
		MaxAlerts:       cfg.Alerts.MaxAlerts,
	})
	for i := range agents {
		alertMon.Check(&agents[i])
	}

	alerts := alertMon.GetAlerts()
	if len(alerts) == 0 {
		fmt.Println("✅ No active alerts.")
		return
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
}

func runWatch() {
	cfg := config.Load()
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)

	fmt.Println("AgentMetrics - Watch mode (Ctrl+C to exit)")
	fmt.Println(strings.Repeat("-", 60))

	for {
		agents, err := detector.Scan()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			time.Sleep(cfg.RefreshInterval.Duration())
			continue
		}

		fmt.Print("\033[H\033[2J")

		fmt.Printf("AgentMetrics - %s\n", time.Now().Format("15:04:05"))
		fmt.Println(strings.Repeat("-", 60))

		if len(agents) == 0 {
			fmt.Println("  No active agents...")
		} else {
			for _, a := range agents {
				statusIcon := "o"
				switch a.Status {
				case agent.StatusRunning:
					statusIcon = "\033[32m*\033[0m"
				case agent.StatusIdle:
					statusIcon = "\033[33m*\033[0m"
				case agent.StatusStopped:
					statusIcon = "\033[31m*\033[0m"
				}

				fmt.Printf("  %s %-20s PID:%-6d CPU:%.1f%%  MEM:%.1fMB\n",
					statusIcon, a.Info.Name, a.PID, a.CPU, a.Memory,
				)
				if a.WorkDir != "" {
					fmt.Printf("    -> %s\n", a.WorkDir)
				}
			}
		}

		fmt.Printf("\n  Next scan in %s...\n", cfg.RefreshInterval.Duration())
		time.Sleep(cfg.RefreshInterval.Duration())
	}
}

func runJSON() {
	cfg := config.Load()
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)

	agents, err := detector.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	snapshot := agent.Snapshot{
		Timestamp: time.Now(),
		Agents:    agents,
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func runConfig() {
	cfg := config.Load()
	cfgPath := config.ConfigPath()

	if len(os.Args) > 2 && os.Args[2] == "edit" {
		// Open config in default editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}
		fmt.Printf("Opening %s with %s...\n", cfgPath, editor)
		cmd := exec.Command(editor, cfgPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 2 && os.Args[2] == "path" {
		fmt.Println(cfgPath)
		return
	}

	if len(os.Args) > 2 && os.Args[2] == "reset" {
		newCfg := config.DefaultConfig()
		if err := newCfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config reset to defaults at:\n  %s\n", cfgPath)
		return
	}

	// Default: show current config
	fmt.Printf("Config: %s\n\n", cfgPath)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(data))
	fmt.Println("\nCommands:")
	fmt.Println("  agentmetrics config edit    Edit config with $EDITOR")
	fmt.Println("  agentmetrics config path    Show config file path")
	fmt.Println("  agentmetrics config reset   Reset to defaults")
}

func printHelp() {
	help := `AgentMetrics v` + version + ` - AI Agent Monitor

USAGE:
  agentmetrics              Launch interactive TUI dashboard
  agentmetrics scan         Quick one-time scan
  agentmetrics watch        Continuous console monitoring
  agentmetrics json         JSON output of current state
  agentmetrics export       Export history (json|csv) [path]
  agentmetrics alerts       View active alerts
  agentmetrics config       View/edit filter configuration
  agentmetrics version      Show version
  agentmetrics help         Show this help

EXPORT:
  agentmetrics export json              Export as JSON to ~/.agentmetrics/history/
  agentmetrics export csv               Export as CSV to ~/.agentmetrics/history/
  agentmetrics export json /tmp/out.json Export to specific path

MONITORED METRICS:
  - CPU / Memory              Process resource usage
  - Tokens (input/output)     Tokens consumed via logs/db
  - Estimated cost ($)        Based on per-model pricing
  - Average latency           API response time
  - Git activity              Branch, commits, diff stats
  - Lines of code (LOC)       +/- changes in the repo
  - Terminal commands          Agent child processes
  - Session duration           Uptime, active/idle time
  - Alerts                    CPU, memory, tokens, cost, idle
  - Network connections       API endpoints
  - File operations           Changes in working directories

CONFIGURATION:
  Config is stored at ~/.agentmetrics/config.json
  All settings are configurable via JSON sections:

  refresh_interval          Scan refresh interval (e.g. "3s")
  detection                 Process detection filters
    ignore_process_patterns Cmdline patterns to ignore
    ignore_paths            Path prefixes to ignore
    skip_system_processes   Ignore system processes
    skip_lsof_for_detection Skip lsof (faster, no iCloud prompts)
    only_exact_process_match Only detect by exact process name
    disabled_agents         Agent IDs to skip
  alerts                    Alert thresholds and behavior
    enabled, cpu/mem/token/cost warning/critical, idle, cooldown, max
  theme                     UI colors (hex values)
    primary, secondary, success, warning, danger, muted, bg, fg, border
  export                    History export settings
    format, directory, max_history
  display                   Dashboard section toggles
    show_tokens, show_cost, show_git, show_terminal, etc.
  keybindings               Keyboard shortcuts
    quit, refresh, export, detail, back, up, down, toggle
  monitor                   Monitor subsystem parameters
    max_log_lines, max_file_ops, max_terminal_commands

SUPPORTED AGENTS:
  - Claude Code         Anthropic's AI agent
  - GitHub Copilot      GitHub's AI programmer
  - OpenAI Codex CLI    OpenAI CLI agent
  - Open Codex          Open-source Codex alternative
  - Aider               AI pair programming in terminal
  - Cody (Sourcegraph)  Sourcegraph's AI assistant
  - Cursor              AI-powered code editor
  - Continue.dev        Open-source AI assistant
  - Codel               Autonomous coding agent
  - MoltBot             AI coding assistant
  - Windsurf (Codeium)  Codeium's AI editor
  - Gemini CLI          Google Gemini CLI agent

TUI SHORTCUTS:
  up/down, j/k    Navigate agents
  Enter           View agent details
  ESC             Back to dashboard
  r               Manual refresh
  e               Export history (JSON)
  Tab             Toggle view
  q               Quit
`
	fmt.Print(help)
}
