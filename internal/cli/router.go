package cli

import (
	"fmt"
	"os"
)

func Run(args []string, version string) int {
	if len(args) == 0 {
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	switch args[0] {
	case "scan", "s":
		if err := runScan(); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning agents: %v\n", err)
			return 1
		}
	case "watch", "w":
		runWatch()
	case "json":
		if err := runJSON(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	case "export":
		if err := runExport(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	case "alerts":
		if err := runAlerts(); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning agents: %v\n", err)
			return 1
		}
	case "config", "c":
		if err := runConfig(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	case "version", "v", "--version":
		fmt.Printf("agentmetrics v%s\n", version)
	case "help", "h", "--help", "-h":
		printHelp(version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printHelp(version)
		return 1
	}

	return 0
}
