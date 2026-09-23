package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/julienbreux/agy-ge-board/internal/cli"
)

func TestFormatters(t *testing.T) {
	headers := []string{"USER", "MODEL", "TOKENS", "COST"}
	rows := [][]string{
		{"alex@example.com", "gemini-1.5-pro", "150000", "$15.50"},
		{"sophia@example.com", "gemini-1.5-flash", "95000", "$3.20"},
	}

	t.Run("FormatJSON serializes slice or struct cleanly", func(t *testing.T) {
		out, err := cli.FormatJSON(rows)
		if err != nil {
			t.Fatalf("unexpected error formatting JSON: %v", err)
		}
		if !strings.Contains(out, "alex@example.com") {
			t.Errorf("JSON output missing alex@example.com: %s", out)
		}
		if !strings.Contains(out, "gemini-1.5-pro") {
			t.Errorf("JSON output missing gemini-1.5-pro: %s", out)
		}
	})

	t.Run("FormatCSV generates valid CSV text with headers", func(t *testing.T) {
		out, err := cli.FormatCSV(headers, rows)
		if err != nil {
			t.Fatalf("unexpected error formatting CSV: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines (1 header + 2 rows), got %d", len(lines))
		}
		if lines[0] != "USER,MODEL,TOKENS,COST" {
			t.Errorf("unexpected header line: %s", lines[0])
		}
		if !strings.Contains(lines[1], "alex@example.com") {
			t.Errorf("first row missing user email: %s", lines[1])
		}
	})

	t.Run("FormatTable creates aligned terminal table", func(t *testing.T) {
		out := cli.FormatTable(headers, rows)
		if !strings.Contains(out, "USER") || !strings.Contains(out, "MODEL") {
			t.Errorf("table output missing headers: %s", out)
		}
		if !strings.Contains(out, "alex@example.com") {
			t.Errorf("table output missing data: %s", out)
		}

		empty := cli.FormatTable(nil, nil)
		if empty != "" {
			t.Errorf("expected empty string for nil table, got: %s", empty)
		}
	})
}

func TestCLICommandsWithDemo(t *testing.T) {
	t.Run("cost command executes with --demo and returns table", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"cost", "--demo", "--format=table"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "USER") || !strings.Contains(out, "alex.turner@example.com") {
			t.Errorf("expected table with demo users, got: %s", out)
		}
	})

	t.Run("cost command executes with --format=json", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"cost", "--demo", "--format=json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, `"user_id": "alex.turner@example.com"`) {
			t.Errorf("expected JSON with user_id, got: %s", out)
		}
	})

	t.Run("cost command executes with --format=csv", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"cost", "--demo", "--format=csv"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "USER,MODEL,USAGE_DATE") {
			t.Errorf("expected CSV header, got: %s", out)
		}
	})

	t.Run("license command executes with --demo", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"license", "--demo"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "SEAT QUOTA") || !strings.Contains(out, "DORMANT") {
			t.Errorf("expected license quota and dormant stats, got: %s", out)
		}
	})

	t.Run("license command executes with --format=json", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"license", "--demo", "--format=json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, `"seat_quota"`) {
			t.Errorf("expected JSON with seat_quota, got: %s", out)
		}
	})

	t.Run("license command executes with --format=csv", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"license", "--demo", "--format=csv"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "SEAT_QUOTA,ASSIGNED_SEATS") {
			t.Errorf("expected CSV header, got: %s", out)
		}
	})

	t.Run("user command executes with --demo", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"user", "alex.turner@example.com", "--demo"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "alex.turner@example.com") || !strings.Contains(out, "TOTAL TOKENS") {
			t.Errorf("expected user summary output, got: %s", out)
		}
	})

	t.Run("user command executes with --format=json", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"user", "alex.turner@example.com", "--demo", "--format=json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, `"user_id": "alex.turner@example.com"`) {
			t.Errorf("expected JSON with user_id, got: %s", out)
		}
	})

	t.Run("doctor command executes in demo mode", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"doctor", "--demo"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "OK") && !strings.Contains(out, "DOCTOR") {
			t.Errorf("expected doctor check report, got: %s", out)
		}
	})

	t.Run("doctor command executes without project id", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"doctor", "--project="})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "[WARN]") {
			t.Errorf("expected doctor check warning for missing project, got: %s", out)
		}
	})

	t.Run("doctor command executes with project id and flags", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{
			"doctor",
			"--project=my-test-prj",
			"--telemetry-table=my-test-prj.ds.telemetry",
			"--billing-table=my-test-prj.billing.export",
		})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "my-test-prj") {
			t.Errorf("expected doctor output to report configured project, got: %s", out)
		}
	})

	t.Run("serve command registers flags properly", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		rootCmd := cli.NewRootCommand()
		serveCmd, _, err := rootCmd.Find([]string{"serve"})
		if err != nil {
			t.Fatalf("failed to find serve command: %v", err)
		}
		if serveCmd == nil || serveCmd.Name() != "serve" {
			t.Fatalf("expected serve command to be registered")
		}
		portFlag := serveCmd.Flag("port")
		if portFlag == nil {
			t.Fatalf("expected port flag on serve command")
		}
		if portFlag.DefValue != "9090" {
			t.Errorf("expected default port 9090 from env, got %s", portFlag.DefValue)
		}
	})

	t.Run("tui command registers days flag", func(t *testing.T) {
		rootCmd := cli.NewRootCommand()
		tuiCmd, _, err := rootCmd.Find([]string{"tui"})
		if err != nil {
			t.Fatalf("failed to find tui command: %v", err)
		}
		if tuiCmd == nil || tuiCmd.Name() != "tui" {
			t.Fatalf("expected tui command to be registered")
		}
		if tuiCmd.Flag("days") == nil {
			t.Errorf("expected days flag on tui command")
		}
	})

	t.Run("doctor command reports missing billing or telemetry table", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{
			"doctor",
			"--project=valid-project",
		})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Missing telemetry table") {
			t.Errorf("expected doctor to report missing tables, got: %s", out)
		}
	})
}
