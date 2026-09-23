package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julienbreux/agy-ge-board/internal/cli"
)

func TestSetupCommand(t *testing.T) {
	t.Run("runs diagnostic checks in demo mode and outputs table", func(t *testing.T) {
		cli.ResetFlags()
		cmd := cli.NewRootCommand()
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"setup", "--demo"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error executing setup: %v", err)
		}

		out := stdout.String()
		if !strings.Contains(out, "Google Cloud ADC & Project") {
			t.Errorf("expected ADC check in table output, got: %s", out)
		}
		if !strings.Contains(out, "BigQuery Telemetry Sink") {
			t.Errorf("expected Telemetry Sink check in output, got: %s", out)
		}
		if !strings.Contains(out, "PASSED: 5") {
			t.Errorf("expected 5 passed checks in output summary, got: %s", out)
		}
	})

	t.Run("runs diagnostic checks and outputs JSON", func(t *testing.T) {
		cli.ResetFlags()
		cmd := cli.NewRootCommand()
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"setup", "--demo", "--format", "json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error executing setup: %v", err)
		}

		out := stdout.String()
		if !strings.Contains(out, `"overall_status": "OK"`) {
			t.Errorf("expected overall_status OK in JSON, got: %s", out)
		}
		if !strings.Contains(out, `"checks":`) {
			t.Errorf("expected checks array in JSON, got: %s", out)
		}
	})

	t.Run("runs with --dry-run and prints provisioning plan", func(t *testing.T) {
		cli.ResetFlags()
		cmd := cli.NewRootCommand()
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"setup", "--demo", "--dry-run", "--project", "test-dryrun-proj"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := stdout.String()
		if !strings.Contains(out, "[DRY-RUN] bq mk --dataset") {
			t.Errorf("expected dry-run bq command, got: %s", out)
		}
		if !strings.Contains(out, "[DRY-RUN] gcloud logging sinks create") {
			t.Errorf("expected dry-run sink create command, got: %s", out)
		}
	})

	t.Run("persists configuration when --save flag is provided", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgFile := filepath.Join(tmpDir, ".agy-ge-board.yaml")

		cli.ResetFlags()
		cmd := cli.NewRootCommand()
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"setup",
			"--demo",
			"--save",
			"--config", cfgFile,
			"--project", "saved-project",
			"--telemetry-table", "saved-project.telemetry.inference_logs",
			"--billing-table", "saved-project.billing.gcp_billing_export_v1_000",
			"--seat-quota", "42",
		})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			t.Fatalf("expected config file %s to be created", cfgFile)
		}

		content, err := os.ReadFile(cfgFile)
		if err != nil {
			t.Fatalf("failed reading config file: %v", err)
		}
		if !strings.Contains(string(content), "saved-project") {
			t.Errorf("expected saved-project in file content, got: %s", string(content))
		}
		if !strings.Contains(string(content), "seat_quota: 42") {
			t.Errorf("expected seat_quota 42 in file content, got: %s", string(content))
		}
	})

	t.Run("subcommands inherit settings from config file when flags omitted", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgFile := filepath.Join(tmpDir, ".agy-ge-board.yaml")

		// Pre-populate configuration file
		configData := `project_id: config-project
telemetry_table: config-project.telemetry.inference_logs
billing_table: config-project.billing.gcp_billing_export_v1_000
seat_quota: 33
`
		if err := os.WriteFile(cfgFile, []byte(configData), 0644); err != nil {
			t.Fatalf("failed creating config: %v", err)
		}

		cli.ResetFlags()
		cmd := cli.NewRootCommand()
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		// Run cost command with --config and --demo (so it uses synthetic provider without hitting GCP)
		cmd.SetArgs([]string{"cost", "--demo", "--config", cfgFile})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error executing cost with config: %v", err)
		}

		// Verify license command also inherits seat quota
		cli.ResetFlags()
		cmdLicense := cli.NewRootCommand()
		var lstdout, lstderr bytes.Buffer
		cmdLicense.SetOut(&lstdout)
		cmdLicense.SetErr(&lstderr)
		cmdLicense.SetArgs([]string{"license", "--demo", "--config", cfgFile})

		err = cmdLicense.Execute()
		if err != nil {
			t.Fatalf("unexpected error executing license with config: %v", err)
		}

		out := lstdout.String()
		if !strings.Contains(out, "33") {
			t.Errorf("expected seat quota 33 inherited from config in license output, got: %s", out)
		}
	})
}
