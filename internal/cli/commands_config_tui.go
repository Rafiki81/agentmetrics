package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/Rafiki81/libagentmetrics/config"
	"github.com/rafaelperezbeato/agentmetrics/internal/tui"
)

func runTUI() error {
	cfg := config.Load()
	return tui.StartApp(cfg)
}

func runConfig(args []string) error {
	cfg := config.Load()
	cfgPath := config.ConfigPath()

	if len(args) > 0 && args[0] == "edit" {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}
		fmt.Printf("Opening %s with %s...\n", cfgPath, editor)
		cmd := exec.Command(editor, cfgPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("opening editor: %w", err)
		}
		return nil
	}

	if len(args) > 0 && args[0] == "path" {
		fmt.Println(cfgPath)
		return nil
	}

	if len(args) > 0 && args[0] == "reset" {
		newCfg := config.DefaultConfig()
		if err := newCfg.Save(); err != nil {
			return fmt.Errorf("resetting config: %w", err)
		}
		fmt.Printf("Config reset to defaults at:\n  %s\n", cfgPath)
		return nil
	}

	fmt.Printf("Config: %s\n\n", cfgPath)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(data))
	fmt.Println("\nCommands:")
	fmt.Println("  agentmetrics config edit    Edit config with $EDITOR")
	fmt.Println("  agentmetrics config path    Show config file path")
	fmt.Println("  agentmetrics config reset   Reset to defaults")

	return nil
}
