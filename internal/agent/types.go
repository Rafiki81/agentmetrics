package agent

import "time"

// Status represents the current state of an agent
type Status int

const (
	StatusUnknown Status = iota
	StatusRunning        // Agent process is active
	StatusIdle           // Agent detected but not actively working
	StatusStopped        // Agent was running but has stopped
)

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "RUNNING"
	case StatusIdle:
		return "IDLE"
	case StatusStopped:
		return "STOPPED"
	default:
		return "UNKNOWN"
	}
}

func (s Status) Color() string {
	switch s {
	case StatusRunning:
		return "#00FF00" // green
	case StatusIdle:
		return "#FFFF00" // yellow
	case StatusStopped:
		return "#FF0000" // red
	default:
		return "#888888" // gray
	}
}

// AgentInfo holds metadata about a known agent type
type AgentInfo struct {
	Name           string   // Display name (e.g., "Claude Code")
	ID             string   // Internal identifier
	ProcessNames   []string // Process names to look for
	LogPaths       []string // Known log file paths (supports ~ expansion)
	Ports          []int    // Known ports the agent may use
	Description    string   // Short description
	DetectPatterns []string // Patterns in cmdline to detect the agent
}

// TokenSource indicates how token data was obtained
type TokenSource string

const (
	TokenSourceNone      TokenSource = ""          // No token data available
	TokenSourceLog       TokenSource = "log"       // Parsed from agent log files
	TokenSourceDB        TokenSource = "db"        // Parsed from agent database
	TokenSourceNetwork   TokenSource = "network"   // Estimated from network traffic
	TokenSourceEstimated TokenSource = "estimated" // Heuristic estimation
)

// TokenMetrics holds token usage data for an agent
type TokenMetrics struct {
	InputTokens   int64       `json:"input_tokens"`    // Tokens sent to the model
	OutputTokens  int64       `json:"output_tokens"`   // Tokens received from the model
	TotalTokens   int64       `json:"total_tokens"`    // InputTokens + OutputTokens
	TokensPerSec  float64     `json:"tokens_per_sec"`  // Current throughput rate
	RequestCount  int         `json:"request_count"`   // Number of API requests observed
	LastModel     string      `json:"last_model"`      // Last model used (e.g., "gpt-4o")
	Source        TokenSource `json:"source"`          // How the data was obtained
	LastRequestAt time.Time   `json:"last_request_at"` // Time of last observed request
	EstCost       float64     `json:"est_cost"`        // Estimated cost in USD
	AvgLatencyMs  int64       `json:"avg_latency_ms"`  // Average API latency in ms
}

// GitActivity holds git-related metrics for an agent's working directory
type GitActivity struct {
	Branch        string      `json:"branch"`
	RecentCommits []GitCommit `json:"recent_commits"`
	Uncommitted   int         `json:"uncommitted"` // Number of uncommitted changes
	LinesAdded    int         `json:"lines_added"`
	LinesRemoved  int         `json:"lines_removed"`
	FilesChanged  int         `json:"files_changed"`
}

// GitCommit represents a single git commit
type GitCommit struct {
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Author  string    `json:"author"`
}

// TerminalActivity holds terminal command tracking for an agent
type TerminalActivity struct {
	RecentCommands []TerminalCommand `json:"recent_commands"`
	TotalCommands  int               `json:"total_commands"`
}

// TerminalCommand represents a detected terminal command
type TerminalCommand struct {
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Category  string    `json:"category"` // "build", "test", "install", "run", "git", "other"
}

// SessionMetrics holds session timing data
type SessionMetrics struct {
	StartedAt    time.Time     `json:"started_at"`
	Uptime       time.Duration `json:"uptime"`
	ActiveTime   time.Duration `json:"active_time"` // Time with CPU > threshold
	IdleTime     time.Duration `json:"idle_time"`   // Time with CPU <= threshold
	LastActiveAt time.Time     `json:"last_active_at"`
}

// LOCMetrics holds lines-of-code metrics
type LOCMetrics struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Net     int `json:"net"` // Added - Removed
	Files   int `json:"files_modified"`
}

// AlertLevel represents severity of an alert
type AlertLevel string

const (
	AlertInfo     AlertLevel = "INFO"
	AlertWarning  AlertLevel = "WARNING"
	AlertCritical AlertLevel = "CRITICAL"
	AlertSecurity AlertLevel = "SECURITY"
)

// SecurityCategory categorizes the type of security event
type SecurityCategory string

const (
	SecCatDangerousCommand SecurityCategory = "dangerous_command"  // rm -rf, sudo, chmod 777
	SecCatSensitiveFile    SecurityCategory = "sensitive_file"     // .env, .ssh/, credentials
	SecCatNetworkExfil     SecurityCategory = "network_exfil"      // curl piped to shell, wget suspicious
	SecCatPackageInstall   SecurityCategory = "package_install"    // installing unknown packages
	SecCatPermEscalation   SecurityCategory = "perm_escalation"    // sudo, su, chmod
	SecCatSecretsExposure  SecurityCategory = "secrets_exposure"   // writing API keys to files
	SecCatMassDeletion     SecurityCategory = "mass_deletion"      // bulk file deletions
	SecCatSystemModify     SecurityCategory = "system_modify"      // crontab, launchctl, systemctl
	SecCatCodeInjection    SecurityCategory = "code_injection"     // eval, exec of remote code
	SecCatSuspiciousNet    SecurityCategory = "suspicious_network" // connections to unknown hosts
	SecCatReverseShell     SecurityCategory = "reverse_shell"      // reverse shell attempts
	SecCatObfuscation      SecurityCategory = "obfuscation"        // base64/hex encoded commands
	SecCatContainerEscape  SecurityCategory = "container_escape"   // docker --privileged, nsenter
	SecCatEnvManipulation  SecurityCategory = "env_manipulation"   // PATH, LD_PRELOAD hijacking
	SecCatCredentialAccess SecurityCategory = "credential_access"  // keychain, browser passwords
	SecCatLogTampering     SecurityCategory = "log_tampering"      // history -c, shred logs
	SecCatRemoteAccess     SecurityCategory = "remote_access"      // ssh, scp, rsync outbound
	SecCatShellPersistence SecurityCategory = "shell_persistence"  // .bashrc/.zshrc modification
)

// SecuritySeverity indicates how dangerous the event is
type SecuritySeverity string

const (
	SecSevLow      SecuritySeverity = "LOW"      // Informational, potentially risky
	SecSevMedium   SecuritySeverity = "MEDIUM"   // Likely risky, needs review
	SecSevHigh     SecuritySeverity = "HIGH"     // Dangerous, immediate attention
	SecSevCritical SecuritySeverity = "CRITICAL" // Very dangerous, possible attack or data loss
)

// SecurityEvent represents a detected security-relevant action by an agent
type SecurityEvent struct {
	Timestamp   time.Time        `json:"timestamp"`
	AgentID     string           `json:"agent_id"`
	AgentName   string           `json:"agent_name"`
	Category    SecurityCategory `json:"category"`
	Severity    SecuritySeverity `json:"severity"`
	Description string           `json:"description"`
	Detail      string           `json:"detail"`  // The actual command/file/connection
	Blocked     bool             `json:"blocked"` // Whether the action was blocked
	Rule        string           `json:"rule"`    // Which rule triggered this
}

// Alert represents a triggered alert
type Alert struct {
	Timestamp time.Time  `json:"timestamp"`
	Level     AlertLevel `json:"level"`
	AgentID   string     `json:"agent_id"`
	AgentName string     `json:"agent_name"`
	Message   string     `json:"message"`
}

// AgentInstance represents a running or detected agent instance
type AgentInstance struct {
	Info           AgentInfo
	PID            int
	Status         Status
	StartTime      time.Time
	LastSeen       time.Time
	CPU            float64          // CPU usage percentage
	Memory         float64          // Memory in MB
	CmdLine        string           // Full command line
	WorkDir        string           // Working directory
	LogLines       []string         // Recent log lines
	FileOps        []FileOperation  // Recent file operations
	NetConns       []NetConnection  // Active network connections
	Tokens         TokenMetrics     // Token usage metrics
	Git            GitActivity      // Git activity metrics
	Terminal       TerminalActivity // Terminal command tracking
	Session        SessionMetrics   // Session timing
	LOC            LOCMetrics       // Lines of code metrics
	SecurityEvents []SecurityEvent  // Security events detected
}

// FileOperation represents a file change detected
type FileOperation struct {
	Timestamp time.Time
	Path      string
	Op        string // "CREATE", "MODIFY", "DELETE", "RENAME"
}

// NetConnection represents a network connection
type NetConnection struct {
	LocalAddr  string
	RemoteAddr string
	State      string
	Protocol   string
}

// Snapshot is a point-in-time capture of all agent activity
type Snapshot struct {
	Timestamp time.Time
	Agents    []AgentInstance
	Alerts    []Alert `json:"alerts,omitempty"`
}
