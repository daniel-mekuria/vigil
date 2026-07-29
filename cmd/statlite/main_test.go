package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pvrlabs/statlite/internal/version"
)

func TestEntrypointRejectsRetiredTargetTypesClearly(t *testing.T) {
	for _, retiredType := range []string{"statlite-health", "statlite", "host"} {
		t.Run(retiredType, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "statlite.yaml")
			config := `server:
  listen: "127.0.0.1:9090"
storage:
  sqlite_path: "./statlite.sqlite"
polling:
  interval: "30s"
targets:
  - name: "obsolete-self"
    type: "` + retiredType + `"
    url: "http://127.0.0.1:9090/healthz"
`
			if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
				t.Fatalf("write config: %v", err)
			}

			cmd := exec.Command("go", "run", ".", "--config", configPath)
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("go run succeeded for retired type %q; output=%s", retiredType, output)
			}
			if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() == 0 {
				t.Fatalf("go run error = %v, want clean non-zero exit; output=%s", err, output)
			}
			message := string(output)
			for _, want := range []string{
				`targets[0].type`,
				`unsupported type`,
				`"` + retiredType + `"`,
				`spring`,
				`statlite-metrics`,
			} {
				if !strings.Contains(message, want) {
					t.Fatalf("startup output = %q, missing %q", message, want)
				}
			}
			for _, forbidden := range []string{"panic:", "goroutine", "runtime error", "stack trace"} {
				if strings.Contains(strings.ToLower(message), forbidden) {
					t.Fatalf("startup output = %q, contains unexpected crash text %q", message, forbidden)
				}
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	var out bytes.Buffer
	printVersion(&out)

	got := out.String()
	want := "statlite " + version.Version + "\n"
	if got != want {
		t.Fatalf("printVersion() = %q, want %q", got, want)
	}
	if !strings.HasPrefix(got, "statlite v") {
		t.Fatalf("printVersion() = %q, want leading statlite v", got)
	}
}

func TestStartupMessage(t *testing.T) {
	got := startupMessage("0.0.0.0:9090", 3)
	want := "StatLite starting: version=" + version.Version + " listen=0.0.0.0:9090 targets=3"
	if got != want {
		t.Fatalf("startupMessage() = %q, want %q", got, want)
	}
}

func TestPrintHelp(t *testing.T) {
	var out bytes.Buffer
	printHelp(&out)

	got := out.String()
	for _, want := range []string{
		"StatLite - tiny self-hosted metrics dashboard for small servers.",
		"Spring Boot Actuator",
		"Usage:",
		"statlite [--config path]",
		"--version",
		"--help",
		"Docs: README.md, docs/configuration.md",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printHelp() missing %q\n%s", want, got)
		}
	}
}
