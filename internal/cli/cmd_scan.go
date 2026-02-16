package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Rafiki81/libagentmetrics/monitor"
)

func runScan() error {
	runtime := newScanRuntime()

	agents, err := runtime.scan()
	if err != nil {
		return err
	}

	if len(agents) == 0 {
		fmt.Println("No active AI agents detected.")
		fmt.Println("\nSupported agents:")
		for _, a := range runtime.registry.Agents {
			fmt.Printf("  - %s (%s)\n", a.Name, a.Description)
		}
		return nil
	}

	collectTokenMetrics(agents)
	collectGitAndSessionMetrics(agents)

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

	return nil
}
