package cli

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdoutStderr(t *testing.T, fn func()) (string, string) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stderr pipe: %v", err)
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	fn()

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr

	stdoutBytes, err := io.ReadAll(stdoutReader)
	if err != nil {
		t.Fatalf("reading stdout: %v", err)
	}
	stderrBytes, err := io.ReadAll(stderrReader)
	if err != nil {
		t.Fatalf("reading stderr: %v", err)
	}

	_ = stdoutReader.Close()
	_ = stderrReader.Close()

	return string(stdoutBytes), string(stderrBytes)
}

func TestRunVersionCommand(t *testing.T) {
	stdout, stderr := captureStdoutStderr(t, func() {
		exitCode := Run([]string{"version"}, "0.9.1")
		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d", exitCode)
		}
	})

	if !strings.Contains(stdout, "agentmetrics v0.9.1") {
		t.Fatalf("expected version output, got: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got: %q", stderr)
	}
}

func TestRunVersionAliases(t *testing.T) {
	cases := []string{"v", "--version"}

	for _, cmd := range cases {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			stdout, stderr := captureStdoutStderr(t, func() {
				exitCode := Run([]string{cmd}, "0.9.1")
				if exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d", exitCode)
				}
			})

			if !strings.Contains(stdout, "agentmetrics v0.9.1") {
				t.Fatalf("expected version output, got: %q", stdout)
			}
			if stderr != "" {
				t.Fatalf("expected empty stderr, got: %q", stderr)
			}
		})
	}
}

func TestRunHelpCommand(t *testing.T) {
	stdout, stderr := captureStdoutStderr(t, func() {
		exitCode := Run([]string{"help"}, "0.9.1")
		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d", exitCode)
		}
	})

	if !strings.Contains(stdout, "USAGE:") {
		t.Fatalf("expected help text with USAGE, got: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got: %q", stderr)
	}
}

func TestRunHelpAliases(t *testing.T) {
	cases := []string{"h", "-h", "--help"}

	for _, cmd := range cases {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			stdout, stderr := captureStdoutStderr(t, func() {
				exitCode := Run([]string{cmd}, "0.9.1")
				if exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d", exitCode)
				}
			})

			if !strings.Contains(stdout, "USAGE:") {
				t.Fatalf("expected help text with USAGE, got: %q", stdout)
			}
			if stderr != "" {
				t.Fatalf("expected empty stderr, got: %q", stderr)
			}
		})
	}
}

func TestRunConfigAliasesPath(t *testing.T) {
	cases := []string{"config", "c"}

	for _, cmd := range cases {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			stdout, stderr := captureStdoutStderr(t, func() {
				exitCode := Run([]string{cmd, "path"}, "0.9.1")
				if exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d", exitCode)
				}
			})

			if strings.TrimSpace(stdout) == "" {
				t.Fatalf("expected non-empty config path output")
			}
			if !strings.Contains(stdout, "config.json") {
				t.Fatalf("expected config path output, got: %q", stdout)
			}
			if stderr != "" {
				t.Fatalf("expected empty stderr, got: %q", stderr)
			}
		})
	}
}

func TestRunUnknownCommand(t *testing.T) {
	stdout, stderr := captureStdoutStderr(t, func() {
		exitCode := Run([]string{"unknown-cmd"}, "0.9.1")
		if exitCode != 1 {
			t.Fatalf("expected exit code 1, got %d", exitCode)
		}
	})

	if !strings.Contains(stderr, "Unknown command: unknown-cmd") {
		t.Fatalf("expected unknown command error, got: %q", stderr)
	}
	if !strings.Contains(stdout, "USAGE:") {
		t.Fatalf("expected help output for unknown command, got: %q", stdout)
	}
}
