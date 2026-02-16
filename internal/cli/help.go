package cli

import "fmt"

func printHelp(version string) {
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
