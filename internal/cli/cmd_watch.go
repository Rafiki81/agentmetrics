package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Rafiki81/libagentmetrics/agent"
)

func runWatch() {
	runtime := newScanRuntime()

	fmt.Println("AgentMetrics - Watch mode (Ctrl+C to exit)")
	fmt.Println(strings.Repeat("-", 60))

	for {
		agents, err := runtime.scan()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			time.Sleep(runtime.cfg.RefreshInterval.Duration())
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

		fmt.Printf("\n  Next scan in %s...\n", runtime.cfg.RefreshInterval.Duration())
		time.Sleep(runtime.cfg.RefreshInterval.Duration())
	}
}
