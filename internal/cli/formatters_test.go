package cli_test

import (
	"strings"
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/cli"
)

func TestFormattersEdgeCases(t *testing.T) {
	t.Run("FormatJSON pretty-prints valid struct", func(t *testing.T) {
		payload := map[string]string{"service": "agy-cost-board", "status": "ok"}
		out, err := cli.FormatJSON(payload)
		if err != nil {
			t.Fatalf("unexpected error formatting JSON: %v", err)
		}
		if !strings.Contains(out, `"service": "agy-cost-board"`) {
			t.Errorf("expected json output to contain service, got: %s", out)
		}
	})

	t.Run("FormatJSON returns error for unmarshalable types", func(t *testing.T) {
		ch := make(chan int)
		_, err := cli.FormatJSON(ch)
		if err == nil {
			t.Errorf("expected error formatting unmarshalable type, got nil")
		}
	})

	t.Run("FormatCSV generates valid comma-separated text", func(t *testing.T) {
		headers := []string{"Name", "Role", "Active"}
		rows := [][]string{
			{"Alice", "Admin", "true"},
			{"Bob", "Developer", "false"},
		}
		out, err := cli.FormatCSV(headers, rows)
		if err != nil {
			t.Fatalf("unexpected error formatting CSV: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines in CSV, got %d", len(lines))
		}
		if lines[0] != "Name,Role,Active" {
			t.Errorf("unexpected header line: %s", lines[0])
		}
	})

	t.Run("FormatCSV works without headers", func(t *testing.T) {
		rows := [][]string{{"1", "2"}}
		out, err := cli.FormatCSV(nil, rows)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "1,2") {
			t.Errorf("expected row in CSV: %s", out)
		}
	})

	t.Run("FormatTable formats aligned ASCII output", func(t *testing.T) {
		headers := []string{"ID", "Description"}
		rows := [][]string{
			{"1", "Item One"},
			{"2", "A Much Longer Description Here"},
		}
		table := cli.FormatTable(headers, rows)
		if !strings.Contains(table, "ID") || !strings.Contains(table, "A Much Longer Description Here") {
			t.Errorf("expected table to contain items, got: %s", table)
		}
	})

	t.Run("FormatTable handles empty input", func(t *testing.T) {
		table := cli.FormatTable(nil, nil)
		if table != "" {
			t.Errorf("expected empty string for empty table, got: %s", table)
		}
	})

	t.Run("FormatTable handles jagged rows", func(t *testing.T) {
		headers := []string{"Col1"}
		rows := [][]string{
			{"Val1", "Extra1"},
			{"Val2"},
		}
		table := cli.FormatTable(headers, rows)
		if !strings.Contains(table, "Extra1") {
			t.Errorf("expected table to contain extra column: %s", table)
		}
	})
}
