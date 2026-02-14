package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rafaelperezbeato/agentmetrics/internal/agent"
	"github.com/rafaelperezbeato/agentmetrics/internal/config"
)

// SecurityMonitor analyzes agent activity for unsafe behavior
type SecurityMonitor struct {
	mu        sync.Mutex
	config    config.SecurityConfig
	events    []agent.SecurityEvent
	maxEvents int

	// Track what we've already alerted on (cmd hash -> time)
	seen map[string]time.Time
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(cfg config.SecurityConfig) *SecurityMonitor {
	maxEvents := cfg.MaxEvents
	if maxEvents <= 0 {
		maxEvents = 500
	}
	return &SecurityMonitor{
		config:    cfg,
		events:    make([]agent.SecurityEvent, 0),
		maxEvents: maxEvents,
		seen:      make(map[string]time.Time),
	}
}

// CheckAgent analyzes all activity from an agent for security issues
func (sm *SecurityMonitor) CheckAgent(a *agent.AgentInstance) {
	if !sm.config.Enabled {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Check terminal commands
	sm.checkCommands(a)

	// 2. Check file operations
	sm.checkFileOps(a)

	// 3. Check network connections
	sm.checkNetwork(a)

	// 4. Check for file-based persistence & credential access
	sm.checkFileSecurity(a)

	// Update agent's security events
	a.SecurityEvents = sm.getEventsForAgent(a.Info.ID)
}

// checkCommands scans terminal commands for dangerous patterns
func (sm *SecurityMonitor) checkCommands(a *agent.AgentInstance) {
	for _, cmd := range a.Terminal.RecentCommands {
		cmdLower := strings.ToLower(cmd.Command)

		// Check dangerous commands
		for _, pattern := range sm.config.DangerousCommands {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatDangerousCommand,
					Severity:    agent.SecSevCritical,
					Description: "Dangerous command detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("dangerous_command:%s", pattern),
				})
				break
			}
		}

		// Check privilege escalation
		for _, pattern := range sm.config.EscalationCommands {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatPermEscalation,
					Severity:    agent.SecSevHigh,
					Description: "Privilege escalation attempt",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("escalation:%s", pattern),
				})
				break
			}
		}

		// Check code injection patterns
		for _, pattern := range sm.config.CodeInjectionPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatCodeInjection,
					Severity:    agent.SecSevHigh,
					Description: "Potential code injection",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("code_injection:%s", pattern),
				})
				break
			}
		}

		// Check system modification
		for _, pattern := range sm.config.SystemModifyCommands {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatSystemModify,
					Severity:    agent.SecSevMedium,
					Description: "System modification command",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("system_modify:%s", pattern),
				})
				break
			}
		}

		// Check package installation from unknown sources
		if sm.isPackageInstall(cmdLower) && len(sm.config.AllowedRegistries) > 0 {
			if !sm.isAllowedRegistry(cmdLower) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatPackageInstall,
					Severity:    agent.SecSevMedium,
					Description: "Package install from unverified source",
					Detail:      cmd.Command,
					Rule:        "package_install:unverified",
				})
			}
		}

		// Check reverse shell patterns
		for _, pattern := range sm.config.ReverseShellPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatReverseShell,
					Severity:    agent.SecSevCritical,
					Description: "Reverse shell attempt detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("reverse_shell:%s", pattern),
				})
				break
			}
		}

		// Check obfuscated commands
		for _, pattern := range sm.config.ObfuscationPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatObfuscation,
					Severity:    agent.SecSevHigh,
					Description: "Obfuscated/encoded command detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("obfuscation:%s", pattern),
				})
				break
			}
		}

		// Check container escape patterns
		for _, pattern := range sm.config.ContainerEscapePatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatContainerEscape,
					Severity:    agent.SecSevCritical,
					Description: "Container escape attempt detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("container_escape:%s", pattern),
				})
				break
			}
		}

		// Check environment variable manipulation
		for _, pattern := range sm.config.EnvManipulationPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatEnvManipulation,
					Severity:    agent.SecSevHigh,
					Description: "Environment variable manipulation",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("env_manipulation:%s", pattern),
				})
				break
			}
		}

		// Check credential/keychain access
		for _, pattern := range sm.config.CredentialAccessPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatCredentialAccess,
					Severity:    agent.SecSevCritical,
					Description: "Credential/keychain access detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("credential_access:%s", pattern),
				})
				break
			}
		}

		// Check log/history tampering
		for _, pattern := range sm.config.LogTamperingPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatLogTampering,
					Severity:    agent.SecSevHigh,
					Description: "Log/history tampering detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("log_tampering:%s", pattern),
				})
				break
			}
		}

		// Check remote access commands
		for _, pattern := range sm.config.RemoteAccessPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				// Skip if it's just "ssh-agent" or "ssh-add" (common dev tools)
				if strings.Contains(cmdLower, "ssh-agent") || strings.Contains(cmdLower, "ssh-add") {
					continue
				}
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatRemoteAccess,
					Severity:    agent.SecSevHigh,
					Description: "Remote access command detected",
					Detail:      cmd.Command,
					Rule:        fmt.Sprintf("remote_access:%s", pattern),
				})
				break
			}
		}
	}
}

// checkFileOps analyzes file operations for sensitive file access
func (sm *SecurityMonitor) checkFileOps(a *agent.AgentInstance) {
	// Count deletions for mass deletion detection
	deleteCount := 0
	for _, op := range a.FileOps {
		if op.Op == "DELETE" {
			deleteCount++
		}
	}

	// Check mass deletion
	if sm.config.MassDeletionThreshold > 0 && deleteCount >= sm.config.MassDeletionThreshold {
		key := fmt.Sprintf("%s:mass_delete:%d", a.Info.ID, deleteCount/sm.config.MassDeletionThreshold)
		if _, seen := sm.seen[key]; !seen {
			sm.addEvent(a, agent.SecurityEvent{
				Category:    agent.SecCatMassDeletion,
				Severity:    agent.SecSevHigh,
				Description: fmt.Sprintf("Mass file deletion detected (%d files)", deleteCount),
				Detail:      fmt.Sprintf("%d files deleted in working directory", deleteCount),
				Rule:        fmt.Sprintf("mass_deletion:threshold=%d", sm.config.MassDeletionThreshold),
			})
		}
	}

	// Check sensitive file access
	for _, op := range a.FileOps {
		pathLower := strings.ToLower(op.Path)
		for _, sensitive := range sm.config.SensitiveFiles {
			if strings.Contains(pathLower, strings.ToLower(sensitive)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatSensitiveFile,
					Severity:    agent.SecSevHigh,
					Description: fmt.Sprintf("Sensitive file %s: %s", strings.ToLower(op.Op), sensitive),
					Detail:      op.Path,
					Rule:        fmt.Sprintf("sensitive_file:%s", sensitive),
				})
				break
			}
		}

		// Check if writing secrets patterns to files
		if op.Op == "CREATE" || op.Op == "MODIFY" {
			sm.checkSecretsInFilename(a, op.Path)
		}
	}
}

// checkSecretsInFilename checks if a file being written looks like it could contain secrets
func (sm *SecurityMonitor) checkSecretsInFilename(a *agent.AgentInstance, path string) {
	secretIndicators := []string{
		"api_key", "apikey", "api-key",
		"secret", "password", "token",
		"private_key", "private-key",
		"access_key", "access-key",
	}
	pathLower := strings.ToLower(path)
	for _, indicator := range secretIndicators {
		if strings.Contains(pathLower, indicator) {
			sm.addEvent(a, agent.SecurityEvent{
				Category:    agent.SecCatSecretsExposure,
				Severity:    agent.SecSevMedium,
				Description: "Possible secrets file created/modified",
				Detail:      path,
				Rule:        fmt.Sprintf("secrets_file:%s", indicator),
			})
			return
		}
	}
}

// checkNetwork analyzes network connections for suspicious destinations
func (sm *SecurityMonitor) checkNetwork(a *agent.AgentInstance) {
	for _, conn := range a.NetConns {
		remoteLower := strings.ToLower(conn.RemoteAddr)

		// Check against suspicious hosts
		for _, host := range sm.config.SuspiciousHosts {
			if strings.Contains(remoteLower, strings.ToLower(host)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatSuspiciousNet,
					Severity:    agent.SecSevHigh,
					Description: fmt.Sprintf("Connection to suspicious host: %s", host),
					Detail:      fmt.Sprintf("%s -> %s [%s]", conn.LocalAddr, conn.RemoteAddr, conn.Protocol),
					Rule:        fmt.Sprintf("suspicious_host:%s", host),
				})
				break
			}
		}

		// Check for data exfiltration patterns (outbound on unusual ports)
		if conn.State == "ESTABLISHED" && isUnusualPort(conn.RemoteAddr) {
			sm.addEvent(a, agent.SecurityEvent{
				Category:    agent.SecCatNetworkExfil,
				Severity:    agent.SecSevLow,
				Description: "Connection on unusual port",
				Detail:      fmt.Sprintf("%s -> %s [%s]", conn.LocalAddr, conn.RemoteAddr, conn.Protocol),
				Rule:        "unusual_port",
			})
		}
	}
}

// checkFileSecurity checks file operations for persistence and credential access
func (sm *SecurityMonitor) checkFileSecurity(a *agent.AgentInstance) {
	for _, op := range a.FileOps {
		pathLower := strings.ToLower(op.Path)

		// Check shell persistence (writing to shell config files)
		if op.Op == "MODIFY" || op.Op == "CREATE" {
			for _, pattern := range sm.config.ShellPersistenceFiles {
				if strings.Contains(pathLower, strings.ToLower(pattern)) {
					sm.addEvent(a, agent.SecurityEvent{
						Category:    agent.SecCatShellPersistence,
						Severity:    agent.SecSevMedium,
						Description: fmt.Sprintf("Shell config %s: %s", strings.ToLower(op.Op), pattern),
						Detail:      op.Path,
						Rule:        fmt.Sprintf("shell_persistence:%s", pattern),
					})
					break
				}
			}
		}

		// Check credential file access via file operations
		for _, pattern := range sm.config.CredentialAccessPatterns {
			if strings.Contains(pathLower, strings.ToLower(pattern)) {
				sm.addEvent(a, agent.SecurityEvent{
					Category:    agent.SecCatCredentialAccess,
					Severity:    agent.SecSevCritical,
					Description: fmt.Sprintf("Credential file access: %s", op.Op),
					Detail:      op.Path,
					Rule:        fmt.Sprintf("credential_file:%s", pattern),
				})
				break
			}
		}
	}
}

// addEvent adds a security event, deduplicating by rule+agent
func (sm *SecurityMonitor) addEvent(a *agent.AgentInstance, evt agent.SecurityEvent) {
	// Deduplicate: don't repeat same rule for same agent within 5 minutes
	key := fmt.Sprintf("%s:%s:%s", a.Info.ID, evt.Rule, evt.Detail)
	if last, ok := sm.seen[key]; ok {
		if time.Since(last) < 5*time.Minute {
			return
		}
	}

	evt.Timestamp = time.Now()
	evt.AgentID = a.Info.ID
	evt.AgentName = a.Info.Name
	evt.Blocked = sm.config.BlockDangerousCommands &&
		(evt.Severity == agent.SecSevCritical || evt.Severity == agent.SecSevHigh)

	sm.events = append(sm.events, evt)
	sm.seen[key] = time.Now()

	// Trim events
	if len(sm.events) > sm.maxEvents {
		sm.events = sm.events[len(sm.events)-sm.maxEvents:]
	}
}

// GetEvents returns all security events
func (sm *SecurityMonitor) GetEvents() []agent.SecurityEvent {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	result := make([]agent.SecurityEvent, len(sm.events))
	copy(result, sm.events)
	return result
}

// GetRecentEvents returns events from the last N minutes
func (sm *SecurityMonitor) GetRecentEvents(minutes int) []agent.SecurityEvent {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var result []agent.SecurityEvent
	for _, e := range sm.events {
		if e.Timestamp.After(cutoff) {
			result = append(result, e)
		}
	}
	return result
}

// getEventsForAgent returns events for a specific agent (caller must hold lock)
func (sm *SecurityMonitor) getEventsForAgent(agentID string) []agent.SecurityEvent {
	var result []agent.SecurityEvent
	for _, e := range sm.events {
		if e.AgentID == agentID {
			result = append(result, e)
		}
	}
	return result
}

// EventCounts returns counts by severity
func (sm *SecurityMonitor) EventCounts() (low, medium, high, critical int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, e := range sm.events {
		switch e.Severity {
		case agent.SecSevLow:
			low++
		case agent.SecSevMedium:
			medium++
		case agent.SecSevHigh:
			high++
		case agent.SecSevCritical:
			critical++
		}
	}
	return
}

// isPackageInstall checks if a command is a package installation
func (sm *SecurityMonitor) isPackageInstall(cmdLower string) bool {
	installCmds := []string{
		"npm install", "npm i ",
		"pip install", "pip3 install",
		"go get ", "go install ",
		"cargo install", "cargo add",
		"gem install",
		"brew install",
		"apt install", "apt-get install",
		"yarn add", "pnpm add",
		"composer require",
	}
	for _, ic := range installCmds {
		if strings.Contains(cmdLower, ic) {
			return true
		}
	}
	return false
}

// isAllowedRegistry checks if the install command targets an allowed registry
func (sm *SecurityMonitor) isAllowedRegistry(cmdLower string) bool {
	for _, reg := range sm.config.AllowedRegistries {
		if strings.Contains(cmdLower, strings.ToLower(reg)) {
			return true
		}
	}
	return false
}

// isUnusualPort checks if a remote address uses a non-standard port
func isUnusualPort(addr string) bool {
	commonPorts := map[string]bool{
		"80": true, "443": true, "8080": true, "8443": true,
		"22": true, "53": true, "3000": true, "3001": true,
		"5000": true, "5173": true, "8000": true, "8888": true,
		"9090": true, "9200": true, "27017": true, "5432": true,
		"3306": true, "6379": true, "11211": true,
	}

	// Extract port from addr like "host:port"
	parts := strings.Split(addr, ":")
	if len(parts) < 2 {
		return false
	}
	port := parts[len(parts)-1]
	return !commonPorts[port]
}
