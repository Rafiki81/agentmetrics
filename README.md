# ◈ AgentMetrics

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![CI](https://github.com/Rafiki81/agentmetrics/actions/workflows/ci.yml/badge.svg)](https://github.com/Rafiki81/agentmetrics/actions/workflows/ci.yml)
[![Release](https://github.com/Rafiki81/agentmetrics/actions/workflows/release.yml/badge.svg)](https://github.com/Rafiki81/agentmetrics/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux-lightgrey?logo=apple)](https://github.com/Rafiki81/agentmetrics)
[![Status](https://img.shields.io/badge/Status-Work%20In%20Progress-orange?style=flat)](https://github.com/Rafiki81/agentmetrics)
[![Testing](https://img.shields.io/badge/Testing-In%20Progress-blueviolet?style=flat)](https://github.com/Rafiki81/agentmetrics)

> ⚠️ **Work In Progress** — This project is under active development and currently being tested. Features may change, and bugs are expected. Contributions and feedback are welcome!

**Real-time terminal dashboard for monitoring local AI coding agents.**

AgentMetrics automatically detects and monitors AI coding agents running on your machine — Claude Code, GitHub Copilot, Codex CLI, Aider, Cursor, Cline, and more — providing live metrics on CPU, memory, tokens, cost, git activity, and session timing, all in a beautiful TUI.

> 📸 *Screenshot coming soon*

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| **Auto-Detection** | Automatically discovers 12+ AI agents via process scanning |
| **Token Monitoring** | Tracks input/output tokens, throughput, model used, and request count |
| **Cost Estimation** | Real-time cost estimates based on model pricing (OpenAI, Anthropic, etc.) |
| **CPU & Memory** | Live process-level resource usage with visual bars |
| **Git Activity** | Branch, uncommitted changes, recent commits, lines added/removed |
| **Session Tracking** | Uptime, active time, idle time, start time |
| **Terminal Commands** | Captures commands executed by child processes |
| **Network Connections** | Monitors active API connections (remote addr, port, state) |
| **File Operations** | Tracks file reads/writes in the working directory |
| **Alert System** | Configurable thresholds for CPU, memory, tokens, cost, and idle time |
| **Security Monitoring** | Detects dangerous commands, sensitive file access, privilege escalation, code injection, and suspicious network activity |
| **Local Model Monitoring** | Auto-detects and monitors Ollama, LM Studio, llama.cpp, vLLM, LocalAI, text-generation-webui, GPT4All |
| **Clickable File Paths** | Cmd+click on file paths in security events to open them directly (OSC 8 terminal hyperlinks) |
| **History & Export** | Export metrics to JSON or CSV; historical session data |
| **Tokyo Night Theme** | Beautiful dark terminal UI with the Tokyo Night color palette |

## 🚀 Supported Agents

| Agent | Detection Method |
|-------|-----------------|
| Claude Code | `claude` process |
| GitHub Copilot | VS Code extensions + `copilot` processes |
| OpenAI Codex CLI | `codex` process |
| Aider | `aider` process / Python |
| Cursor | `cursor` / `Cursor` process |
| Cline (VS Code) | VS Code + cline extension |
| Continue (VS Code) | VS Code + continue extension |
| Tabnine | `tabnine` process |
| Amazon CodeWhisperer | VS Code/JetBrains plugin |
| Sourcegraph Cody | `cody` process / VS Code extension |
| JetBrains AI | JetBrains IDE process |
| Replit AI | `replit` process |

## 📦 Installation

### From GitHub Releases (recommended)

Download the latest binary from [**Releases**](https://github.com/Rafiki81/agentmetrics/releases):

```bash
# macOS (Apple Silicon)
curl -Lo agentmetrics.tar.gz https://github.com/Rafiki81/agentmetrics/releases/latest/download/agentmetrics_darwin_arm64.tar.gz
tar xzf agentmetrics.tar.gz
sudo mv agentmetrics /usr/local/bin/

# macOS (Intel)
curl -Lo agentmetrics.tar.gz https://github.com/Rafiki81/agentmetrics/releases/latest/download/agentmetrics_darwin_amd64.tar.gz
tar xzf agentmetrics.tar.gz
sudo mv agentmetrics /usr/local/bin/

# Linux (amd64)
curl -Lo agentmetrics.tar.gz https://github.com/Rafiki81/agentmetrics/releases/latest/download/agentmetrics_linux_amd64.tar.gz
tar xzf agentmetrics.tar.gz
sudo mv agentmetrics /usr/local/bin/
```

### With `go install`

```bash
go install github.com/rafaelperezbeato/agentmetrics/cmd/agentmetrics@latest
```

### From source

```bash
# Clone the repository
git clone https://github.com/rafaelperezbeato/agentmetrics.git
cd agentmetrics

# Build
make build

# Or install to $GOPATH/bin
make install
```

### Prerequisites

- **Go 1.24+** (only for building from source)
- **macOS** or **Linux** (uses `ps`, `lsof`, `pgrep` for process inspection)

## 🎯 Usage

### Interactive TUI Dashboard

```bash
# Launch the full dashboard
agentmetrics

# Or
agentmetrics tui
```

**Keyboard shortcuts:**

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate between agents |
| `Enter` | Open detailed view for selected agent |
| `ESC` | Go back to main dashboard |
| `e` | Export current metrics |
| `r` | Force refresh |
| `q` | Quit |

### CLI Commands

```bash
# Quick scan — list detected agents
agentmetrics scan

# JSON output (great for scripting)
agentmetrics json

# Watch mode — auto-refresh every N seconds
agentmetrics watch        # default 5s
agentmetrics watch 10     # every 10s

# Export metrics to file
agentmetrics export                  # JSON to ~/.agentmetrics/history/
agentmetrics export --format csv     # CSV format
agentmetrics export --output ./out   # Custom output directory

# View active alerts
agentmetrics alerts

# Manage configuration
agentmetrics config show             # Show current config
agentmetrics config path             # Show config file path
agentmetrics config reset            # Reset to defaults

# Version
agentmetrics version

# Help
agentmetrics help
```

## ⚙️ Configuration

Configuration is stored in `~/.agentmetrics/config.json`. Created automatically on first run with all defaults.

```bash
# View current config
agentmetrics config show

# Edit with your $EDITOR
agentmetrics config edit

# Reset to defaults
agentmetrics config reset

# Show config file path
agentmetrics config path
```

### Full Config Example

```json
{
  "refresh_interval": "3s",
  "detection": {
    "ignore_process_patterns": [
      "CursorUIViewService", "com.apple.", "/System/Library/",
      "/usr/libexec/", "/usr/sbin/", "WindowServer",
      "loginwindow", "launchd", "kernel_task"
    ],
    "ignore_paths": ["/Library/", "/System/", "/private/", "/usr/"],
    "skip_system_processes": true,
    "skip_lsof_for_detection": false,
    "only_exact_process_match": false,
    "disabled_agents": []
  },
  "alerts": {
    "enabled": true,
    "cpu_warning": 80,
    "cpu_critical": 95,
    "memory_warning_mb": 500,
    "memory_critical_mb": 1000,
    "token_warning": 500000,
    "token_critical": 2000000,
    "cost_warning_usd": 1.0,
    "cost_critical_usd": 5.0,
    "idle_minutes": 10,
    "cooldown_minutes": 5,
    "max_alerts": 100
  },
  "theme": {
    "primary": "#7C3AED",
    "secondary": "#06B6D4",
    "success": "#10B981",
    "warning": "#F59E0B",
    "danger": "#EF4444",
    "muted": "#6B7280",
    "background": "#1A1B26",
    "background_alt": "#24283B",
    "foreground": "#C0CAF5",
    "border": "#3B4261"
  },
  "export": {
    "format": "json",
    "directory": "",
    "max_history": 10000
  },
  "display": {
    "show_tokens": true,
    "show_cost": true,
    "show_git": true,
    "show_terminal": true,
    "show_network": true,
    "show_files": true,
    "show_session": true,
    "show_alerts": true,
    "show_security": true,
    "show_local_models": true
  },
  "keybindings": {
    "quit": "q",
    "refresh": "r",
    "export": "e",
    "detail": "enter",
    "back": "esc",
    "up": "up",
    "down": "down",
    "toggle": "tab"
  },
  "monitor": {
    "max_log_lines": 50,
    "max_file_ops": 200,
    "max_terminal_commands": 50,
    "watch_dirs": []
  },
  "security": {
    "enabled": true,
    "block_dangerous_commands": false,
    "dangerous_commands": ["rm -rf /", "rm -rf ~", "mkfs.", "dd if=", ":(){:|:&};:", "chmod 777", "..."],
    "sensitive_files": [".env", ".ssh/", "id_rsa", ".aws/credentials", ".kube/config", "..."],
    "suspicious_hosts": ["pastebin.com", "ngrok.io", "interact.sh", "..."],
    "escalation_commands": ["sudo ", "su root", "pkexec ", "chmod u+s", "..."],
    "code_injection_patterns": ["eval(", "exec(", "| bash", "$(curl ", "..."],
    "system_modify_commands": ["crontab", "launchctl", "iptables", "visudo", "..."],
    "reverse_shell_patterns": ["bash -i >& /dev/tcp/", "nc -e /bin/", "socat exec:", "mkfifo /tmp/", "..."],
    "obfuscation_patterns": ["base64 --decode", "base64 -d", "xxd -r", "printf '\\x", "..."],
    "container_escape_patterns": ["docker run --privileged", "docker.sock", "--cap-add=ALL", "nsenter ", "..."],
    "env_manipulation_patterns": ["export LD_PRELOAD=", "export DYLD_INSERT_LIBRARIES=", "export PATH=", "..."],
    "credential_access_patterns": ["security find-generic-password", "security dump-keychain", "Login Data", "..."],
    "log_tampering_patterns": ["history -c", "> ~/.zsh_history", "unset HISTFILE", "shred ", "..."],
    "remote_access_patterns": ["ssh ", "scp ", "rsync ", "ssh -L", "ssh -R", "..."],
    "shell_persistence_files": [".bashrc", ".zshrc", ".profile", "LaunchAgents/", "..."],
    "mass_deletion_threshold": 10,
    "max_events": 500
  },
  "local_models": {
    "enabled": true,
    "endpoints": []
  }
}
```

### Config Sections

| Section | Description |
|---------|-------------|
| `refresh_interval` | Dashboard refresh interval (e.g. `"3s"`, `"500ms"`, `"1m"`) |
| `detection` | Process scanning filters — ignore patterns, paths, system processes |
| `alerts` | Alert thresholds (CPU, memory, tokens, cost, idle) + cooldown + max |
| `theme` | Full UI color scheme via hex values (Tokyo Night by default) |
| `export` | History export format (`json`/`csv`), directory, max records |
| `display` | Toggle which dashboard sections appear (tokens, git, session, etc.) |
| `keybindings` | Customize all keyboard shortcuts |
| `monitor` | Subsystem limits (max file ops, terminal commands, log lines) |
| `security` | Security monitoring rules — dangerous commands, sensitive files, network, escalation |
| `local_models` | Local model server monitoring — auto-detect + custom endpoints |

### Alert Thresholds

| Threshold | Default | Description |
|-----------|---------|-------------|
| `cpu_warning` | 80% | CPU usage warning level |
| `cpu_critical` | 95% | CPU usage critical level |
| `memory_warning_mb` | 500 MB | Memory warning level |
| `memory_critical_mb` | 1000 MB | Memory critical level |
| `token_warning` | 500K | Token count warning level |
| `token_critical` | 2M | Token count critical level |
| `cost_warning_usd` | $1.00 | Cost warning level |
| `cost_critical_usd` | $5.00 | Cost critical level |
| `idle_minutes` | 10 min | Alert after N minutes idle |
| `cooldown_minutes` | 5 min | Cooldown between repeated alerts |
| `max_alerts` | 100 | Maximum alerts stored in memory |

### Security Rules

The `security` section provides real-time detection of unsafe agent behavior:

| Category | What it detects | Severity |
|----------|----------------|----------|
| **Dangerous commands** | `rm -rf /`, `mkfs.`, fork bombs, `dd if=`, `chmod 777` | 🚨 CRITICAL |
| **Reverse shell** | `bash -i >& /dev/tcp/`, `nc -e`, `socat`, `mkfifo` | 🚨 CRITICAL |
| **Container escape** | `docker --privileged`, `-v /:/host`, `nsenter`, `--cap-add=ALL` | 🚨 CRITICAL |
| **Credential access** | macOS Keychain (`security find-*`), browser passwords, `pass show` | 🚨 CRITICAL |
| **Privilege escalation** | `sudo`, `su root`, `pkexec`, `chmod u+s` | 🔴 HIGH |
| **Code injection** | `eval()`, `curl \| bash`, `$(wget ...)`, `exec()` | 🔴 HIGH |
| **Obfuscated commands** | `base64 --decode \| bash`, `xxd -r`, hex-encoded payloads | 🔴 HIGH |
| **Env manipulation** | Modify `PATH`, `LD_PRELOAD`, `DYLD_INSERT_LIBRARIES`, disable TLS validation | 🔴 HIGH |
| **Log tampering** | `history -c`, `> ~/.zsh_history`, `unset HISTFILE`, `shred` | 🔴 HIGH |
| **Remote access** | `ssh`, `scp`, `rsync` outbound, port forwarding (`ssh -L/-R/-D`) | 🔴 HIGH |
| **Sensitive file access** | `.env`, `.ssh/`, `.aws/credentials`, `id_rsa` | 🔴 HIGH |
| **Mass file deletion** | More than N files deleted in one scan cycle | 🔴 HIGH |
| **Suspicious network** | Connections to `pastebin.com`, `ngrok.io`, `interact.sh` | 🔴 HIGH |
| **System modification** | `crontab`, `launchctl`, `iptables`, `visudo` | ⚠️ MEDIUM |
| **Shell persistence** | Modification of `.bashrc`, `.zshrc`, `.profile`, LaunchAgents | ⚠️ MEDIUM |
| **Secrets exposure** | Files with `api_key`, `password`, `token` in name | ⚠️ MEDIUM |
| **Package install** | Installs from unverified registries (when allow-list set) | ⚠️ MEDIUM |
| **Unusual ports** | Outbound connections on non-standard ports | ℹ️ LOW |

All rules are fully configurable. Set `block_dangerous_commands: true` to flag events as blocked.

Security events in the TUI include **clickable file paths** — hold `Cmd` and click on any file path to open it directly in Finder (via OSC 8 terminal hyperlinks).

### 🖥️ Local Model Monitoring

AgentMetrics auto-detects and monitors local AI model servers running on your machine:

| Server | Default Port | Detection Method |
|--------|-------------|-----------------|
| **Ollama** | 11434 | Native API (`/api/tags` + `/api/ps`) — full model info, VRAM usage |
| **LM Studio** | 1234 | OpenAI-compatible (`/v1/models`) |
| **llama.cpp** | 8080 | OpenAI-compatible (`/v1/models`) |
| **vLLM** | 8000 | OpenAI-compatible (`/v1/models`) |
| **LocalAI** | 8080 | OpenAI-compatible (`/v1/models`) |
| **text-generation-webui** | 5000 | OpenAI-compatible (`/v1/models`) |
| **GPT4All** | 4891 | OpenAI-compatible (`/v1/models`) |

The dashboard displays:
- **Server status** — running/stopped indicator per server
- **Active model** — currently loaded model name
- **Resources** — CPU, memory (MB), VRAM (MB) usage
- **Model list** — available models with sizes and quantization levels

You can also add **custom endpoints** in the config:

```json
{
  "local_models": {
    "enabled": true,
    "endpoints": [
      { "name": "My Server", "id": "custom1", "url": "http://localhost:9090" }
    ]
  }
}
```

## 🏗️ Architecture

```
agentmetrics/
├── cmd/agentmetrics/
│   └── main.go              # CLI entry point & subcommands
├── internal/
│   ├── agent/
│   │   ├── types.go         # Core data types (AgentInstance, TokenInfo, etc.)
│   │   ├── registry.go      # Agent definitions & detection patterns
│   │   └── detector.go      # Process scanner & agent discovery
│   ├── monitor/
│   │   ├── tokens.go        # Token usage collection (sqlite, logs)
│   │   ├── cost.go          # Model pricing & cost estimation
│   │   ├── process.go       # CPU & memory monitoring
│   │   ├── network.go       # Network connection monitoring
│   │   ├── filesystem.go    # File operation tracking
│   │   ├── git.go           # Git activity monitoring
│   │   ├── terminal.go      # Terminal command capture
│   │   ├── session.go       # Session timing & uptime
│   │   ├── alerts.go        # Alert system with thresholds
│   │   ├── security.go      # Security monitoring & threat detection
│   │   ├── localmodels.go   # Local model server auto-detection & monitoring
│   │   └── history.go       # History storage & export
│   ├── tui/
│   │   ├── app.go           # Bubble Tea model (Init/Update/View)
│   │   ├── dashboard.go     # Dashboard & detail view rendering
│   │   └── styles.go        # Tokyo Night color palette & styles
│   └── config/
│       └── config.go        # Configuration management
├── .github/
│   └── workflows/
│       ├── ci.yml           # CI: build + test on push/PR
│       └── release.yml      # Release: GoReleaser on version tags
├── .goreleaser.yml          # GoReleaser config (cross-compile + publish)
├── Makefile
├── go.mod
└── go.sum
```

### How It Works

1. **Detection** — Scans running processes (`ps aux`) and matches against known agent signatures (process names, command patterns)
2. **Enrichment** — For each detected agent, collects:
   - CPU/Memory via process stats
   - Token data from SQLite databases (e.g., Claude's `~/.claude/` DB) or log files
   - Git status from the agent's working directory
   - Network connections via `lsof`
   - Session timing from process start time
3. **Cost Estimation** — Maps detected models to pricing tables and calculates running cost
4. **Alerts** — Evaluates metrics against configurable thresholds
5. **Security** — Analyzes commands, file ops, and network for unsafe behavior
6. **Local Models** — Probes known local model servers (Ollama, LM Studio, llama.cpp, vLLM, LocalAI, text-generation-webui, GPT4All) via HTTP APIs to collect status, loaded models, and resource usage
7. **Rendering** — Displays everything in a real-time Bubble Tea TUI with Tokyo Night styling. Security events include clickable file paths (OSC 8 hyperlinks) for quick navigation

### Token Data Sources

AgentMetrics reads token data from multiple sources depending on the agent:

| Agent | Source | Method |
|-------|--------|--------|
| Claude Code | `~/.claude/` SQLite DB | Direct query |
| GitHub Copilot | VS Code telemetry logs | Log parsing |
| Others | Process environment / logs | Heuristics |

## 🛠️ Development

```bash
# Format code
make fmt

# Run tests
make test

# Run linter (requires golangci-lint)
make lint

# Download dependencies
make deps

# Test release locally (requires goreleaser)
make release-dry

# Create a new release (bumps version, tags, pushes — GitHub Actions builds binaries)
make release V=0.2.0

# See all available commands
make help
```

### CI/CD

The project uses **GitHub Actions** for continuous integration and automated releases:

| Workflow | Trigger | What it does |
|----------|---------|-------------|
| **CI** | Push/PR to `main` | Build, test, vet |
| **Release** | Push tag `v*` | Run tests → GoReleaser builds binaries → GitHub Release |

**Creating a release:**

```bash
# Automated (updates version in code, commits, tags, pushes)
make release V=0.2.0

# Manual
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GoReleaser generates binaries for:
- macOS (Apple Silicon + Intel)
- Linux (amd64 + arm64)

Binaries are published as `.tar.gz` archives in [GitHub Releases](https://github.com/Rafiki81/agentmetrics/releases) with SHA256 checksums.

## 📊 Monitored Metrics

| Category | Metrics |
|----------|---------|
| **Process** | CPU %, Memory (MB), PID, Status |
| **Tokens** | Input, Output, Total, Tokens/sec, Request count |
| **Cost** | Estimated cost (USD) based on model pricing |
| **Model** | Detected model name, average latency |
| **Git** | Branch, uncommitted changes, recent commits, LOC +/- |
| **Session** | Uptime, active time, idle time, start time |
| **Terminal** | Commands executed by child processes |
| **Network** | Active connections (remote address, port, protocol) |
| **Files** | Recent file operations (read/write/create) |
| **Alerts** | CPU, memory, token, cost, and idle alerts |
| **Security** | Dangerous commands, privilege escalation, sensitive files, suspicious network |
| **Local Models** | Server status, active model, CPU/MEM/VRAM usage, available models with sizes |

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

<p align="center">
  Built with <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a> & <a href="https://github.com/charmbracelet/lipgloss">Lip Gloss</a>
</p>
