package tui

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rafaelperezbeato/agentmetrics/internal/agent"
	"github.com/rafaelperezbeato/agentmetrics/internal/config"
	"github.com/rafaelperezbeato/agentmetrics/internal/monitor"
)

// RenderDashboard renders the main dashboard view
func RenderDashboard(agents []agent.AgentInstance, selected int, alerts []agent.Alert, secEvents []agent.SecurityEvent, localModels []agent.LocalModelInfo, width, height int, s *Styles, disp config.DisplayConfig) string {
	var b strings.Builder

	// Header
	logo := s.Logo.Render("◈ AgentMetrics")
	subtitle := lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(" — AI Agent Monitor")
	timestamp := lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(time.Now().Format("15:04:05"))

	headerLine := lipgloss.JoinHorizontal(lipgloss.Center, logo, subtitle)
	headerPadding := ""
	headerLen := lipgloss.Width(headerLine) + lipgloss.Width(timestamp)
	if width > headerLen {
		headerPadding = strings.Repeat(" ", width-headerLen-2)
	}
	b.WriteString(headerLine + headerPadding + timestamp + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Border).Render(strings.Repeat("─", width)) + "\n")

	if len(agents) == 0 {
		msg := s.Empty.Width(width).Render("\n🔍 Scanning for AI agents...\n\nNo active agents detected.\nStart an AI agent (claude, codex, aider, etc.) to monitor it.\n")
		b.WriteString(msg)
		b.WriteString("\n")
		b.WriteString(renderHelp(width, s))
		return b.String()
	}

	// Summary bar
	running := 0
	totalCPU := 0.0
	totalMem := 0.0
	totalTokens := int64(0)
	totalCost := 0.0
	for _, a := range agents {
		if a.Status == agent.StatusRunning {
			running++
		}
		totalCPU += a.CPU
		totalMem += a.Memory
		totalTokens += a.Tokens.TotalTokens
		totalCost += a.Tokens.EstCost
	}

	summary := fmt.Sprintf(
		" Agents: %s  │  CPU: %s  │  MEM: %s  │  Tokens: %s  │  Cost: %s",
		s.MetricValue.Render(fmt.Sprintf("%d active", running)),
		s.MetricValue.Render(fmt.Sprintf("%.1f%%", totalCPU)),
		s.MetricValue.Render(fmt.Sprintf("%.1f MB", totalMem)),
		s.TokenValue.Render(monitor.FormatTokenCount(totalTokens)),
		s.Cost.Render(monitor.FormatCost(totalCost)),
	)

	// Alert indicator
	alertCount := len(alerts)
	if alertCount > 0 && disp.ShowAlerts {
		critCount := 0
		warnCount := 0
		for _, al := range alerts {
			switch al.Level {
			case agent.AlertCritical:
				critCount++
			case agent.AlertWarning:
				warnCount++
			}
		}
		if critCount > 0 {
			summary += fmt.Sprintf("  │  %s", s.AlertCrit.Render(fmt.Sprintf("🔴 %d alerts", critCount)))
		} else if warnCount > 0 {
			summary += fmt.Sprintf("  │  %s", s.AlertWarn.Render(fmt.Sprintf("⚠ %d warnings", warnCount)))
		}
	}

	// Security indicator
	if len(secEvents) > 0 && disp.ShowSecurity {
		highCount := 0
		for _, evt := range secEvents {
			if evt.Severity == agent.SecSevCritical || evt.Severity == agent.SecSevHigh {
				highCount++
			}
		}
		if highCount > 0 {
			summary += fmt.Sprintf("  │  %s", s.SecurityBanner.Render(fmt.Sprintf("🛡 %d security", highCount)))
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Fg).Render(summary) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Border).Render(strings.Repeat("─", width)) + "\n\n")

	// Agent list
	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	for i, a := range agents {
		card := renderAgentCard(a, cardWidth, i == selected, s, disp)
		b.WriteString(card + "\n")
	}

	// Recent alerts section
	if len(alerts) > 0 && disp.ShowAlerts {
		b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Border).Render(strings.Repeat("─", width)) + "\n")
		b.WriteString(s.Header.Width(width).Render("⚡ Recent Alerts") + "\n")
		maxShow := 3
		start := 0
		if len(alerts) > maxShow {
			start = len(alerts) - maxShow
		}
		for _, al := range alerts[start:] {
			icon := "ℹ"
			style := s.AlertInfo
			switch al.Level {
			case agent.AlertWarning:
				icon = "⚠"
				style = s.AlertWarn
			case agent.AlertCritical:
				icon = "🔴"
				style = s.AlertCrit
			}
			b.WriteString(fmt.Sprintf("  %s %s %s — %s\n",
				lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(al.Timestamp.Format("15:04:05")),
				style.Render(icon),
				style.Render(al.AgentName),
				style.Render(al.Message),
			))
		}
		b.WriteString("\n")
	}

	// Recent security events
	if len(secEvents) > 0 && disp.ShowSecurity {
		b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Border).Render(strings.Repeat("─", width)) + "\n")
		b.WriteString(s.SecurityBanner.Width(width).Render("🛡 Security Events") + "\n")
		maxShow := 5
		start := 0
		if len(secEvents) > maxShow {
			start = len(secEvents) - maxShow
		}
		for _, evt := range secEvents[start:] {
			icon, style := securitySeverityStyle(evt.Severity, s)
			linkedDetail := securityDetailWithLink(evt.Detail, 50)
			b.WriteString(fmt.Sprintf("  %s %s %s %s — %s\n",
				lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(evt.Timestamp.Format("15:04:05")),
				style.Render(icon),
				style.Render(fmt.Sprintf("[%-8s]", evt.Severity)),
				style.Render(evt.Description),
				lipgloss.NewStyle().Foreground(s.Theme.Muted).Italic(true).Render(linkedDetail),
			))
		}
		b.WriteString("\n")
	}

	// Local models section
	if len(localModels) > 0 && disp.ShowLocalModels {
		b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Border).Render(strings.Repeat("─", width)) + "\n")
		b.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(s.Theme.Secondary).
			Width(width).
			Render("🖥 Local Models") + "\n")

		for _, srv := range localModels {
			statusColor := s.Theme.Success
			statusIcon := "●"
			if srv.Status == agent.LocalModelIdle {
				statusColor = s.Theme.Warning
			} else if srv.Status == agent.LocalModelStopped {
				statusColor = s.Theme.Danger
				statusIcon = "○"
			}

			// Server header
			header := fmt.Sprintf("  %s %s  %s",
				lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
				lipgloss.NewStyle().Bold(true).Foreground(s.Theme.Primary).Render(srv.ServerName),
				lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(srv.Endpoint),
			)
			b.WriteString(header + "\n")

			// Active model + stats
			if srv.ActiveModel != "" {
				modelLine := fmt.Sprintf("    Model: %s",
					lipgloss.NewStyle().Bold(true).Foreground(s.Theme.Fg).Render(srv.ActiveModel),
				)
				b.WriteString(modelLine)

				if srv.CPU > 0 || srv.MemoryMB > 0 {
					stats := fmt.Sprintf("  CPU: %s  MEM: %s",
						lipgloss.NewStyle().Foreground(s.Theme.Secondary).Render(fmt.Sprintf("%.1f%%", srv.CPU)),
						lipgloss.NewStyle().Foreground(s.Theme.Secondary).Render(fmt.Sprintf("%.0f MB", srv.MemoryMB)),
					)
					b.WriteString(stats)
				}
				if srv.VRAM_MB > 0 {
					b.WriteString(fmt.Sprintf("  VRAM: %s",
						lipgloss.NewStyle().Foreground(s.Theme.Warning).Render(fmt.Sprintf("%.0f MB", srv.VRAM_MB)),
					))
				}
				b.WriteString("\n")
			}

			// List available models (compact)
			if len(srv.Models) > 0 {
				modelNames := make([]string, 0, len(srv.Models))
				for _, m := range srv.Models {
					name := m.Name
					if m.Running {
						name = lipgloss.NewStyle().Foreground(s.Theme.Success).Render("▶ " + name)
					} else {
						name = lipgloss.NewStyle().Foreground(s.Theme.Muted).Render("  " + name)
					}
					if m.Size != "" {
						name += lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(fmt.Sprintf(" (%s)", m.Size))
					}
					modelNames = append(modelNames, name)
				}
				maxModels := 5
				for i, name := range modelNames {
					if i >= maxModels {
						remaining := len(modelNames) - maxModels
						b.WriteString(fmt.Sprintf("    %s\n",
							lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(fmt.Sprintf("... +%d more", remaining)),
						))
						break
					}
					b.WriteString(fmt.Sprintf("    %s\n", name))
				}
			}
		}
		b.WriteString("\n")
	}

	// Help bar
	b.WriteString(renderHelp(width, s))

	return b.String()
}

// renderAgentCard renders a single agent card
func renderAgentCard(a agent.AgentInstance, width int, selected bool, s *Styles, disp config.DisplayConfig) string {
	style := s.AgentCard.Width(width)
	if selected {
		style = s.AgentCardSelected.Width(width)
	}

	// Line 1: Name and Status
	name := s.AgentName.Render(a.Info.Name)
	status := s.StatusStyle(a.Status.String()).Render("● " + a.Status.String())

	nameWidth := lipgloss.Width(name)
	statusWidth := lipgloss.Width(status)
	gap := width - nameWidth - statusWidth - 4
	if gap < 1 {
		gap = 1
	}
	line1 := name + strings.Repeat(" ", gap) + status

	// Line 2: Metrics
	cpuBar := s.RenderBar(a.CPU, 100, 15)
	memStr := fmt.Sprintf("%.1f MB", a.Memory)

	line2 := fmt.Sprintf(
		"  %s CPU %s %s  │  %s MEM %s  │  PID %s",
		s.MetricLabel.Render(""),
		cpuBar,
		s.MetricValue.Render(fmt.Sprintf("%.1f%%", a.CPU)),
		s.MetricLabel.Render(""),
		s.MetricValue.Render(memStr),
		s.MetricValue.Render(fmt.Sprintf("%d", a.PID)),
	)

	// Line 3: Token metrics + Cost
	line3 := ""
	if disp.ShowTokens {
		if a.Tokens.TotalTokens > 0 || a.Tokens.RequestCount > 0 {
			tokensStr := monitor.FormatTokenCount(a.Tokens.TotalTokens)
			tpsStr := monitor.FormatTokensPerSec(a.Tokens.TokensPerSec)
			reqStr := fmt.Sprintf("%d", a.Tokens.RequestCount)

			line3 = fmt.Sprintf("  %s %s  │  %s %s  │  %s %s",
				s.TokenLabel.Render("◆ Tokens:"),
				s.TokenValue.Render(tokensStr),
				s.TokenLabel.Render("Vel:"),
				s.TokenValue.Render(tpsStr),
				s.TokenLabel.Render("Reqs:"),
				s.TokenValue.Render(reqStr),
			)

			if a.Tokens.EstCost > 0 && disp.ShowCost {
				line3 += fmt.Sprintf("  │  %s %s",
					s.TokenLabel.Render("Cost:"),
					s.Cost.Render(monitor.FormatCost(a.Tokens.EstCost)),
				)
			}

			if a.Tokens.LastModel != "" {
				line3 += fmt.Sprintf("  │  %s",
					s.TokenSource.Render(a.Tokens.LastModel),
				)
			}
		} else {
			line3 = fmt.Sprintf("  %s %s",
				s.TokenLabel.Render("◆ Tokens:"),
				s.TokenSource.Render("no data"),
			)
		}
	}

	// Line 4: Git + Session + LOC (compact)
	line4Parts := []string{}
	if disp.ShowGit && a.Git.Branch != "" {
		branchStr := a.Git.Branch
		if len(branchStr) > 20 {
			branchStr = branchStr[:17] + "..."
		}
		gitStr := fmt.Sprintf("⎇ %s", branchStr)
		if a.Git.Uncommitted > 0 {
			gitStr += fmt.Sprintf(" (+%d)", a.Git.Uncommitted)
		}
		line4Parts = append(line4Parts, s.Git.Render(gitStr))
	}
	if disp.ShowSession && a.Session.Uptime > 0 {
		line4Parts = append(line4Parts, s.Session.Render("⏱ "+monitor.FormatDuration(a.Session.Uptime)))
	}
	if disp.ShowGit && (a.LOC.Added > 0 || a.LOC.Removed > 0) {
		locStr := fmt.Sprintf("+%d/-%d", a.LOC.Added, a.LOC.Removed)
		line4Parts = append(line4Parts, lipgloss.NewStyle().Foreground(lipgloss.Color("#22D3EE")).Render("✎ "+locStr))
	}
	if disp.ShowTerminal && a.Terminal.TotalCommands > 0 {
		line4Parts = append(line4Parts, lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(fmt.Sprintf("⌨ %d cmds", a.Terminal.TotalCommands)))
	}

	line4 := ""
	if len(line4Parts) > 0 {
		line4 = "  " + strings.Join(line4Parts, "  │  ")
	}

	// Line 5: Working directory
	line5 := ""
	if a.WorkDir != "" {
		dir := shortenPath(a.WorkDir)
		line5 = fmt.Sprintf("  %s %s", s.MetricLabel.Render("📂"), lipgloss.NewStyle().Foreground(s.Theme.Secondary).Render(dir))
	}

	content := line1 + "\n" + line2
	if line3 != "" {
		content += "\n" + line3
	}
	if line4 != "" {
		content += "\n" + line4
	}
	if line5 != "" {
		content += "\n" + line5
	}

	return style.Render(content)
}

// RenderDetail renders the agent detail panel
func RenderDetail(a agent.AgentInstance, fileOps []agent.FileOperation, alerts []agent.Alert, secEvents []agent.SecurityEvent, width, height int, s *Styles, disp config.DisplayConfig) string {
	var b strings.Builder

	// Header
	b.WriteString(s.Header.Width(width).Render(fmt.Sprintf("◈ %s — Details", a.Info.Name)))
	b.WriteString("\n")

	// Info section
	infoPanel := s.DetailPanel.Width(width - 4)

	info := fmt.Sprintf(
		"%s %s\n%s %s\n%s %d\n%s %s\n%s %s\n%s %s",
		s.MetricLabel.Render("Status:    "),
		s.StatusStyle(a.Status.String()).Render("● "+a.Status.String()),
		s.MetricLabel.Render("Agent:     "),
		s.MetricValue.Render(a.Info.Name),
		s.MetricLabel.Render("PID:       "),
		a.PID,
		s.MetricLabel.Render("CPU:       "),
		s.MetricValue.Render(fmt.Sprintf("%.1f%%", a.CPU))+" "+s.RenderBar(a.CPU, 100, 20),
		s.MetricLabel.Render("Memory:    "),
		s.MetricValue.Render(fmt.Sprintf("%.1f MB", a.Memory)),
		s.MetricLabel.Render("Directory: "),
		lipgloss.NewStyle().Foreground(s.Theme.Secondary).Render(a.WorkDir),
	)
	b.WriteString(infoPanel.Render(info))
	b.WriteString("\n\n")

	// Token metrics section
	if disp.ShowTokens {
		b.WriteString(s.Header.Width(width).Render("◆ Token Metrics"))
		b.WriteString("\n")
		tokenPanel := s.DetailPanel.Width(width - 4)

		if a.Tokens.TotalTokens > 0 || a.Tokens.RequestCount > 0 {
			sourceLabel := string(a.Tokens.Source)
			if sourceLabel == "" {
				sourceLabel = "unknown"
			}

			tokInfo := fmt.Sprintf(
				"%s %s\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
				s.TokenLabel.Render("Input tokens:   "),
				s.TokenValue.Render(monitor.FormatTokenCount(a.Tokens.InputTokens)),
				s.TokenLabel.Render("Output tokens:  "),
				s.TokenValue.Render(monitor.FormatTokenCount(a.Tokens.OutputTokens)),
				s.TokenLabel.Render("Total tokens:   "),
				s.TokenValue.Render(monitor.FormatTokenCount(a.Tokens.TotalTokens)),
				s.TokenLabel.Render("Speed:          "),
				s.TokenValue.Render(monitor.FormatTokensPerSec(a.Tokens.TokensPerSec)),
				s.TokenLabel.Render("Requests:       "),
				s.TokenValue.Render(fmt.Sprintf("%d", a.Tokens.RequestCount)),
				s.TokenLabel.Render("Model:          "),
				s.TokenValue.Render(a.Tokens.LastModel),
				s.TokenLabel.Render("Estimated cost: "),
				s.Cost.Render(monitor.FormatCost(a.Tokens.EstCost)),
				s.TokenLabel.Render("Avg latency:    "),
				s.TokenValue.Render(formatLatency(a.Tokens.AvgLatencyMs)),
				s.TokenLabel.Render("Data source:    "),
				s.TokenSource.Render(sourceLabel),
			)
			b.WriteString(tokenPanel.Render(tokInfo))
		} else {
			b.WriteString(tokenPanel.Render(s.TokenSource.Render("No token data available for this agent.")))
		}
		b.WriteString("\n\n")
	}

	// Session metrics section
	if disp.ShowSession && a.Session.Uptime > 0 {
		b.WriteString(s.Header.Width(width).Render("⏱ Session"))
		b.WriteString("\n")
		sessPanel := s.DetailPanel.Width(width - 4)
		sessInfo := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
			s.Session.Render("Uptime:       "),
			s.MetricValue.Render(monitor.FormatDuration(a.Session.Uptime)),
			s.Session.Render("Active time:  "),
			s.MetricValue.Render(monitor.FormatDuration(a.Session.ActiveTime)),
			s.Session.Render("Idle time:    "),
			s.MetricValue.Render(monitor.FormatDuration(a.Session.IdleTime)),
			s.Session.Render("Started at:   "),
			lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(a.Session.StartedAt.Format("15:04:05")),
			s.Session.Render("Last active:  "),
			lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(a.Session.LastActiveAt.Format("15:04:05")),
		)
		b.WriteString(sessPanel.Render(sessInfo))
		b.WriteString("\n\n")
	}

	// Git activity section
	if disp.ShowGit && a.Git.Branch != "" {
		b.WriteString(s.Header.Width(width).Render("⎇ Git Activity"))
		b.WriteString("\n")
		gitPanel := s.DetailPanel.Width(width - 4)

		var gitLines []string
		gitLines = append(gitLines, fmt.Sprintf("%s %s",
			s.Git.Render("Branch:       "),
			s.MetricValue.Render(a.Git.Branch),
		))
		gitLines = append(gitLines, fmt.Sprintf("%s %s",
			s.Git.Render("Uncommitted:  "),
			s.MetricValue.Render(fmt.Sprintf("%d changes", a.Git.Uncommitted)),
		))
		if a.LOC.Added > 0 || a.LOC.Removed > 0 {
			gitLines = append(gitLines, fmt.Sprintf("%s %s  %s  (%d files)",
				s.Git.Render("Lines:        "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Render(fmt.Sprintf("+%d", a.LOC.Added)),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(fmt.Sprintf("-%d", a.LOC.Removed)),
				a.LOC.Files,
			))
		}

		// Recent commits
		if len(a.Git.RecentCommits) > 0 {
			gitLines = append(gitLines, "")
			gitLines = append(gitLines, s.Git.Render("Recent commits:"))
			maxCommits := 5
			if len(a.Git.RecentCommits) < maxCommits {
				maxCommits = len(a.Git.RecentCommits)
			}
			for _, c := range a.Git.RecentCommits[:maxCommits] {
				msg := c.Message
				if len(msg) > 50 {
					msg = msg[:47] + "..."
				}
				gitLines = append(gitLines, fmt.Sprintf("  %s %s %s",
					lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(c.Hash[:7]),
					lipgloss.NewStyle().Foreground(s.Theme.Fg).Render(msg),
					lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(c.Time.Format("15:04")),
				))
			}
		}
		b.WriteString(gitPanel.Render(strings.Join(gitLines, "\n")))
		b.WriteString("\n\n")
	}

	// Terminal commands section
	if disp.ShowTerminal && a.Terminal.TotalCommands > 0 {
		b.WriteString(s.Header.Width(width).Render("⌨ Terminal Commands"))
		b.WriteString("\n")
		termPanel := s.DetailPanel.Width(width - 4)

		var termLines []string
		termLines = append(termLines, fmt.Sprintf("%s %s",
			s.TokenLabel.Render("Total commands: "),
			s.MetricValue.Render(fmt.Sprintf("%d", a.Terminal.TotalCommands)),
		))

		if len(a.Terminal.RecentCommands) > 0 {
			termLines = append(termLines, "")
			maxCmds := 10
			start := 0
			if len(a.Terminal.RecentCommands) > maxCmds {
				start = len(a.Terminal.RecentCommands) - maxCmds
			}
			for _, cmd := range a.Terminal.RecentCommands[start:] {
				cmdStr := cmd.Command
				if len(cmdStr) > 60 {
					cmdStr = cmdStr[:57] + "..."
				}
				catStyle := lipgloss.NewStyle().Foreground(s.Theme.Secondary)
				termLines = append(termLines, fmt.Sprintf("  %s %s %s",
					lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(cmd.Timestamp.Format("15:04:05")),
					catStyle.Render(fmt.Sprintf("[%-7s]", cmd.Category)),
					lipgloss.NewStyle().Foreground(s.Theme.Fg).Render(cmdStr),
				))
			}
		}
		b.WriteString(termPanel.Render(strings.Join(termLines, "\n")))
		b.WriteString("\n\n")
	}

	// Command line
	if a.CmdLine != "" {
		b.WriteString(s.MetricLabel.Render("  Command: "))
		cmdDisplay := a.CmdLine
		maxCmd := width - 12
		if len(cmdDisplay) > maxCmd {
			cmdDisplay = cmdDisplay[:maxCmd-3] + "..."
		}
		b.WriteString(lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(cmdDisplay))
		b.WriteString("\n\n")
	}

	// Network connections
	if disp.ShowNetwork && len(a.NetConns) > 0 {
		b.WriteString(s.Header.Width(width).Render("🌐 Network Connections"))
		b.WriteString("\n")
		for _, conn := range a.NetConns {
			desc := monitor.DescribeConnection(conn)
			b.WriteString(fmt.Sprintf("  %s\n", lipgloss.NewStyle().Foreground(s.Theme.Secondary).Render(desc)))
		}
		b.WriteString("\n")
	}

	// File operations
	if disp.ShowFiles && len(fileOps) > 0 {
		b.WriteString(s.Header.Width(width).Render("📄 Recent File Operations"))
		b.WriteString("\n")

		maxOps := 10
		start := 0
		if len(fileOps) > maxOps {
			start = len(fileOps) - maxOps
		}

		for _, op := range fileOps[start:] {
			opStyle := s.FileOpStyle(op.Op)
			timeStr := op.Timestamp.Format("15:04:05")
			filePath := shortenPath(op.Path)

			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				s.MetricLabel.Render(timeStr),
				opStyle.Render(fmt.Sprintf("%-7s", op.Op)),
				lipgloss.NewStyle().Foreground(s.Theme.Fg).Render(filePath),
			))
		}
		b.WriteString("\n")
	}

	// Alerts for this agent
	if disp.ShowAlerts {
		agentAlerts := filterAlertsByAgent(alerts, a.Info.ID)
		if len(agentAlerts) > 0 {
			b.WriteString(s.Header.Width(width).Render("⚡ Alerts"))
			b.WriteString("\n")
			maxAlerts := 5
			start := 0
			if len(agentAlerts) > maxAlerts {
				start = len(agentAlerts) - maxAlerts
			}
			for _, al := range agentAlerts[start:] {
				icon := "ℹ"
				st := s.AlertInfo
				switch al.Level {
				case agent.AlertWarning:
					icon = "⚠"
					st = s.AlertWarn
				case agent.AlertCritical:
					icon = "🔴"
					st = s.AlertCrit
				}
				b.WriteString(fmt.Sprintf("  %s %s %s\n",
					lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(al.Timestamp.Format("15:04:05")),
					st.Render(icon),
					st.Render(al.Message),
				))
			}
			b.WriteString("\n")
		}
	}

	// Security events for this agent
	if disp.ShowSecurity {
		agentSec := filterSecurityByAgent(secEvents, a.Info.ID)
		if len(agentSec) > 0 {
			b.WriteString(s.SecurityBanner.Width(width).Render("🛡 Security Events"))
			b.WriteString("\n")
			maxEvts := 8
			start := 0
			if len(agentSec) > maxEvts {
				start = len(agentSec) - maxEvts
			}
			for _, evt := range agentSec[start:] {
				icon, st := securitySeverityStyle(evt.Severity, s)
				linkedDetail := securityDetailWithLink(evt.Detail, 60)
				blocked := ""
				if evt.Blocked {
					blocked = s.SecurityCritical.Render(" [BLOCKED]")
				}
				b.WriteString(fmt.Sprintf("  %s %s %s %s%s\n",
					lipgloss.NewStyle().Foreground(s.Theme.Muted).Render(evt.Timestamp.Format("15:04:05")),
					st.Render(icon),
					st.Render(evt.Description),
					lipgloss.NewStyle().Foreground(s.Theme.Muted).Italic(true).Render(linkedDetail),
					blocked,
				))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(s.Help.Render("  ESC back  │  r refresh  │  e export  │  q quit"))

	return b.String()
}

// filterAlertsByAgent returns alerts for a specific agent
func filterAlertsByAgent(alerts []agent.Alert, agentID string) []agent.Alert {
	var result []agent.Alert
	for _, a := range alerts {
		if a.AgentID == agentID {
			result = append(result, a)
		}
	}
	return result
}

// formatLatency formats latency in ms for display
func formatLatency(ms int64) string {
	if ms <= 0 {
		return "—"
	}
	if ms >= 1000 {
		return fmt.Sprintf("%.1fs", float64(ms)/1000.0)
	}
	return fmt.Sprintf("%dms", ms)
}

// renderHelp renders the help bar at the bottom
func renderHelp(width int, s *Styles) string {
	help := "  ↑/↓ navigate  │  Enter details  │  e export  │  r refresh  │  q quit"
	return s.Help.Width(width).Render(help)
}

// shortenPath shortens a file path for display
func shortenPath(path string) string {
	home, _ := filepath.Abs("~")
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}

	// Try to abbreviate long paths
	parts := strings.Split(path, "/")
	if len(parts) > 5 {
		return strings.Join(parts[:2], "/") + "/.../" + strings.Join(parts[len(parts)-2:], "/")
	}
	return path
}

// filterSecurityByAgent returns security events for a specific agent
func filterSecurityByAgent(events []agent.SecurityEvent, agentID string) []agent.SecurityEvent {
	var result []agent.SecurityEvent
	for _, e := range events {
		if e.AgentID == agentID {
			result = append(result, e)
		}
	}
	return result
}

// securitySeverityStyle returns the appropriate icon and style for a security severity
func securitySeverityStyle(sev agent.SecuritySeverity, s *Styles) (string, lipgloss.Style) {
	switch sev {
	case agent.SecSevCritical:
		return "🚨", s.SecurityCritical
	case agent.SecSevHigh:
		return "🔴", s.SecurityHigh
	case agent.SecSevMedium:
		return "⚠️", s.SecurityMedium
	case agent.SecSevLow:
		return "ℹ️", s.SecurityLow
	default:
		return "•", lipgloss.NewStyle().Foreground(s.Theme.Muted)
	}
}

// fileHyperlink wraps a file path in an OSC 8 terminal hyperlink.
// Command+click (macOS) or Ctrl+click (Linux) opens the file in the default handler.
// Falls back to plain text if the path doesn't look like an absolute path.
func fileHyperlink(path string, displayText string) string {
	if !strings.HasPrefix(path, "/") {
		return displayText
	}
	// Build file:// URL, encoding spaces and special chars in the path
	fileURL := "file://" + url.PathEscape(path)
	// Replace %2F back to / since PathEscape encodes them
	fileURL = strings.ReplaceAll(fileURL, "%2F", "/")
	// OSC 8 hyperlink: \033]8;;URL\033\\TEXT\033]8;;\033\\
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", fileURL, displayText)
}

// securityDetailWithLink formats a security event detail, making file paths clickable
func securityDetailWithLink(detail string, maxLen int) string {
	display := detail
	if len(display) > maxLen {
		display = display[:maxLen-3] + "..."
	}
	// If the detail looks like a file path, make it a clickable hyperlink
	if strings.HasPrefix(detail, "/") {
		return fileHyperlink(detail, display)
	}
	return display
}
