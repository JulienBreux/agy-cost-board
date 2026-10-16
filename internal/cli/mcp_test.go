package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/cli"
	"github.com/julienbreux/agy-cost-board/internal/mcp"
)

func TestMCPCommand(t *testing.T) {
	t.Run("mcp command processes stdio JSON-RPC requests in demo mode", func(t *testing.T) {
		cli.ResetFlags()
		input := strings.Join([]string{
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`,
			`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
			`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_overview_metrics","arguments":{"days":7}}}`,
		}, "\n") + "\n"

		inBuf := bytes.NewBufferString(input)
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		rootCmd := cli.NewRootCommand()
		rootCmd.SetIn(inBuf)
		rootCmd.SetOut(outBuf)
		rootCmd.SetErr(errBuf)
		rootCmd.SetArgs([]string{"mcp", "--demo"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("expected mcp command to succeed, got error: %v (stderr: %s)", err, errBuf.String())
		}

		lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 JSON-RPC response lines, got %d: %s", len(lines), outBuf.String())
		}

		var resp1 mcp.Response
		if err := json.Unmarshal([]byte(lines[0]), &resp1); err != nil {
			t.Fatalf("failed to unmarshal line 1: %v", err)
		}
		if resp1.Error != nil {
			t.Errorf("unexpected error on initialize: %+v", resp1.Error)
		}

		var resp3 mcp.Response
		if err := json.Unmarshal([]byte(lines[2]), &resp3); err != nil {
			t.Fatalf("failed to unmarshal line 3: %v", err)
		}
		if resp3.Error != nil {
			t.Errorf("unexpected error on tools/call: %+v", resp3.Error)
		}
	})

	t.Run("mcp command returns error when --mcp-enabled=false", func(t *testing.T) {
		cli.ResetFlags()
		inBuf := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n")
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		rootCmd := cli.NewRootCommand()
		rootCmd.SetIn(inBuf)
		rootCmd.SetOut(outBuf)
		rootCmd.SetErr(errBuf)
		rootCmd.SetArgs([]string{"mcp", "--demo", "--mcp-enabled=false"})

		err := rootCmd.Execute()
		if err == nil {
			t.Fatalf("expected error when --mcp-enabled=false, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "disabled") {
			t.Errorf("expected disabled error message, got: %v", err)
		}
	})

	t.Run("mcp command returns error when _MCP_ENABLED=false", func(t *testing.T) {
		cli.ResetFlags()
		t.Setenv("_MCP_ENABLED", "false")

		inBuf := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n")
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		rootCmd := cli.NewRootCommand()
		rootCmd.SetIn(inBuf)
		rootCmd.SetOut(outBuf)
		rootCmd.SetErr(errBuf)
		rootCmd.SetArgs([]string{"mcp", "--demo"})

		err := rootCmd.Execute()
		if err == nil {
			t.Fatalf("expected error when _MCP_ENABLED=false, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "disabled") {
			t.Errorf("expected disabled error message, got: %v", err)
		}
	})

	t.Run("serve command configures MCP endpoint and prints status in startup banner", func(t *testing.T) {
		cli.ResetFlags()
		ctx, cancel := t.Context(), func() {}
		_ = ctx
		_ = cancel

		outBuf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(outBuf)
		rootCmd.SetErr(outBuf)
		// Use an invalid port so ListenAndServe returns immediately after printing banner and wiring server
		rootCmd.SetArgs([]string{"serve", "--demo", "--port=65536", "--mcp-enabled=false"})

		_ = rootCmd.Execute()
		if !strings.Contains(outBuf.String(), "MCP Server  : disabled (_MCP_ENABLED=false)") {
			t.Errorf("expected disabled MCP Server in serve banner, got: %s", outBuf.String())
		}
	})
}
