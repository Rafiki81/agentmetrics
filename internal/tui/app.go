package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rafaelperezbeato/agentmetrics/internal/agent"
	"github.com/rafaelperezbeato/agentmetrics/internal/config"
	"github.com/rafaelperezbeato/agentmetrics/internal/monitor"
)

// View represents current UI view
type View int

const (
	ViewDashboard View = iota
	ViewDetail
)

// Model is the main Bubble Tea model
type Model struct {
	// Data
	agents        []agent.AgentInstance
	detector      *agent.Detector
	fileMon       *monitor.FileWatcher
	netMon        *monitor.NetworkMonitor
	procMon       *monitor.ProcessMonitor
	tokenMon      *monitor.TokenMonitor
	gitMon        *monitor.GitMonitor
	termMon       *monitor.TerminalMonitor
	sessionMon    *monitor.SessionMonitor
	alertMon      *monitor.AlertMonitor
	secMon        *monitor.SecurityMonitor
	localModelMon *monitor.LocalModelMonitor
	history       *monitor.HistoryStore
	config        *config.Config
	styles        *Styles
	alerts        []agent.Alert
	secEvents     []agent.SecurityEvent
	localModels   []agent.LocalModelInfo

	// UI state
	currentView View
	selected    int
	width       int
	height      int

	// Timing
	lastRefresh time.Time
	err         error
}

// tickMsg triggers periodic refresh
type tickMsg time.Time

// agentScanMsg carries scan results
type agentScanMsg struct {
	agents []agent.AgentInstance
	err    error
}

// NewModel creates the initial model
func NewModel(cfg *config.Config) Model {
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)
	fileMon := monitor.NewFileWatcher(cfg.Monitor.MaxFileOps)
	netMon := monitor.NewNetworkMonitor()
	procMon := monitor.NewProcessMonitor(nil)
	tokenMon := monitor.NewTokenMonitor()
	gitMon := monitor.NewGitMonitor()
	termMon := monitor.NewTerminalMonitor(cfg.Monitor.MaxTermCommands)
	sessionMon := monitor.NewSessionMonitor()

	// Build alert thresholds from config
	thresholds := monitor.AlertThresholds{
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
	}
	alertMon := monitor.NewAlertMonitor(thresholds)

	// Security monitor from config
	secMon := monitor.NewSecurityMonitor(cfg.Security)

	// Local model monitor
	localModelMon := monitor.NewLocalModelMonitor(cfg.LocalModels)

	// History store from config
	histDir := cfg.Export.Directory
	history := monitor.NewHistoryStore(histDir, cfg.Export.MaxHistory)

	// Build styles from theme config
	styles := NewStyles(cfg.Theme)

	return Model{
		detector:      detector,
		fileMon:       fileMon,
		netMon:        netMon,
		procMon:       procMon,
		tokenMon:      tokenMon,
		gitMon:        gitMon,
		termMon:       termMon,
		sessionMon:    sessionMon,
		alertMon:      alertMon,
		secMon:        secMon,
		localModelMon: localModelMon,
		history:       history,
		config:        cfg,
		styles:        styles,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.scanAgents(),
		m.tick(),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(
			m.scanAgents(),
			m.tick(),
		)

	case agentScanMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.agents = msg.agents
			m.lastRefresh = time.Now()

			// Update file watcher with agent working dirs
			for _, a := range m.agents {
				if a.WorkDir != "" {
					m.fileMon.AddDir(a.WorkDir)
				}
			}

			// Update network info
			for i, a := range m.agents {
				conns := m.netMon.GetConnections(a.PID)
				m.agents[i].NetConns = conns
			}

			// Update file ops
			for i, a := range m.agents {
				if a.WorkDir != "" {
					ops := m.fileMon.GetOperationsForDir(a.WorkDir)
					m.agents[i].FileOps = ops
				}
			}

			// Collect token metrics (includes cost + latency)
			m.tokenMon.Collect(m.agents)

			// Collect git activity + LOC
			for i := range m.agents {
				m.gitMon.Collect(&m.agents[i])
			}

			// Collect terminal commands
			for i := range m.agents {
				m.termMon.Collect(&m.agents[i])
			}

			// Update session metrics
			for i := range m.agents {
				m.sessionMon.Collect(&m.agents[i])
			}

			// Check alerts
			if m.config.Alerts.Enabled {
				for i := range m.agents {
					m.alertMon.Check(&m.agents[i])
				}
				m.alerts = m.alertMon.GetRecentAlerts(30)
			}

			// Security analysis (after terminal + file + network data is collected)
			if m.config.Security.Enabled {
				for i := range m.agents {
					m.secMon.CheckAgent(&m.agents[i])
				}
				m.secEvents = m.secMon.GetRecentEvents(60)
			}

			// Record history
			m.history.Record(m.agents)

			// Collect local model server info
			if m.config.LocalModels.Enabled {
				m.localModels = m.localModelMon.Collect()
			}
		}
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case ViewDetail:
		if m.selected >= 0 && m.selected < len(m.agents) {
			a := m.agents[m.selected]
			return RenderDetail(a, a.FileOps, m.alerts, m.secEvents, m.width, m.height, m.styles, m.config.Display)
		}
		m.currentView = ViewDashboard
		return RenderDashboard(m.agents, m.selected, m.alerts, m.secEvents, m.localModels, m.width, m.height, m.styles, m.config.Display)
	default:
		return RenderDashboard(m.agents, m.selected, m.alerts, m.secEvents, m.localModels, m.width, m.height, m.styles, m.config.Display)
	}
}

// handleKey processes keyboard input
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	kb := m.config.Keybindings

	switch {
	case key == kb.Quit || key == "ctrl+c":
		m.fileMon.Stop()
		return m, tea.Quit

	case key == kb.Up || key == "k":
		if m.currentView == ViewDashboard && m.selected > 0 {
			m.selected--
		}

	case key == kb.Down || key == "j":
		if m.currentView == ViewDashboard && m.selected < len(m.agents)-1 {
			m.selected++
		}

	case key == kb.Detail:
		if m.currentView == ViewDashboard && len(m.agents) > 0 {
			m.currentView = ViewDetail
		}

	case key == kb.Back:
		if m.currentView == ViewDetail {
			m.currentView = ViewDashboard
		}

	case key == kb.Refresh:
		return m, m.scanAgents()

	case key == kb.Export:
		// Export current history
		if err := m.history.ExportJSON(""); err == nil {
			// Silently exported
		}
		return m, nil

	case key == kb.Toggle:
		if m.currentView == ViewDashboard {
			m.currentView = ViewDetail
		} else {
			m.currentView = ViewDashboard
		}
	}

	return m, nil
}

// scanAgents performs an async agent scan
func (m Model) scanAgents() tea.Cmd {
	return func() tea.Msg {
		agents, err := m.detector.Scan()
		return agentScanMsg{agents: agents, err: err}
	}
}

// tick returns a command that sends a tick after the refresh interval
func (m Model) tick() tea.Cmd {
	interval := m.config.RefreshInterval.Duration()
	if interval <= 0 {
		interval = 3 * time.Second
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// StartApp starts the TUI application
func StartApp(cfg *config.Config) error {
	model := NewModel(cfg)

	// Start file watcher
	model.fileMon.Start(1 * time.Second)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
