package cli

import (
	"fmt"

	"github.com/Rafiki81/libagentmetrics/monitor"
)

func runExport(args []string) error {
	format := "json"
	path := ""

	if len(args) > 0 {
		format = args[0]
	}
	if len(args) > 1 {
		path = args[1]
	}

	runtime := newScanRuntime()

	agents, err := runtime.scan()
	if err != nil {
		return err
	}

	collectTokenMetrics(agents)
	collectGitAndSessionMetrics(agents)

	history := monitor.NewHistoryStore(runtime.cfg.Export.Directory, runtime.cfg.Export.MaxHistory)
	history.Record(agents)

	switch format {
	case "json":
		if err := history.ExportJSON(path); err != nil {
			return fmt.Errorf("exporting JSON: %w", err)
		}
	case "csv":
		if err := history.ExportCSV(path); err != nil {
			return fmt.Errorf("exporting CSV: %w", err)
		}
	default:
		return fmt.Errorf("unknown format: %s (use 'json' or 'csv')", format)
	}

	if path == "" {
		fmt.Printf("Exported to: %s/\n", history.DataDir())
	} else {
		fmt.Printf("Exported to: %s\n", path)
	}

	return nil
}
