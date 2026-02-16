package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rafiki81/libagentmetrics/agent"
)

func runJSON() error {
	runtime := newScanRuntime()

	agents, err := runtime.scan()
	if err != nil {
		return err
	}

	snapshot := agent.Snapshot{
		Timestamp: time.Now(),
		Agents:    agents,
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
