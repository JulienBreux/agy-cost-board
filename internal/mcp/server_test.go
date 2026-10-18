package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/mcp"
)

func newTestMCPServer() *mcp.Server {
	engine := attribution.NewEngine(bigquery.NewDemoDataProviderWithQuota(15), 5*time.Minute)
	return mcp.NewServer(engine, "1.2.3-test")
}

func TestMCPStdioLifecycleAndTools(t *testing.T) {
	srv := newTestMCPServer()

	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"antigravity","version":"2.0"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_overview_metrics","arguments":{"days":30}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_license_status","arguments":{"days":30}}}`,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_users_costs","arguments":{"days":14,"model":"gemini-4.0-pro"}}}`,
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_user_dashboard","arguments":{"user_id":"alex.turner@google.com","days":30,"monthly_budget":200}}}`,
	}

	input := strings.Join(requests, "\n") + "\n"
	var out bytes.Buffer

	err := srv.ServeStdio(context.Background(), strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("ServeStdio returned unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	// 8 messages sent, 1 notification ("notifications/initialized") produces no response -> 7 responses expected
	if len(lines) != 7 {
		t.Fatalf("expected 7 JSON-RPC responses, got %d:\n%s", len(lines), out.String())
	}

	// 1. Verify initialize response
	var initResp mcp.Response
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed to unmarshal init response: %v", err)
	}
	if initResp.Error != nil {
		t.Fatalf("unexpected error in initialize: %+v", initResp.Error)
	}

	// 2. Verify ping response
	var pingResp mcp.Response
	if err := json.Unmarshal([]byte(lines[1]), &pingResp); err != nil {
		t.Fatalf("failed to unmarshal ping response: %v", err)
	}
	if pingResp.Error != nil {
		t.Fatalf("unexpected error in ping: %+v", pingResp.Error)
	}

	// 3. Verify tools/list response contains all 4 tools
	var listResp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			Tools []mcp.Tool `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &listResp); err != nil {
		t.Fatalf("failed to unmarshal tools/list response: %v", err)
	}
	if len(listResp.Result.Tools) != 4 {
		t.Fatalf("expected 4 tools in tools/list, got %d", len(listResp.Result.Tools))
	}

	// 4. Verify tool calls (lines 3..6) succeed and return non-empty text JSON content
	for idx := 3; idx <= 6; idx++ {
		var callResp struct {
			JSONRPC string         `json:"jsonrpc"`
			Result  mcp.ToolResult `json:"result"`
			Error   *mcp.RPCError  `json:"error"`
		}
		if err := json.Unmarshal([]byte(lines[idx]), &callResp); err != nil {
			t.Fatalf("line %d: failed to unmarshal tool call response: %v", idx, err)
		}
		if callResp.Error != nil {
			t.Fatalf("line %d: unexpected RPC error: %+v", idx, callResp.Error)
		}
		if callResp.Result.IsError {
			t.Fatalf("line %d: tool call returned isError=true: %+v", idx, callResp.Result.Content)
		}
		if len(callResp.Result.Content) == 0 || callResp.Result.Content[0].Text == "" {
			t.Fatalf("line %d: expected non-empty text content in tool result", idx)
		}
	}

	// Verify FinOps Phase 1 & Phase 2 fields in get_overview_metrics (line 3), get_license_status (line 4), and get_user_dashboard (line 6)
	if !strings.Contains(lines[3], "total_wtu") || !strings.Contains(lines[3], "weekly_pool") || !strings.Contains(lines[3], "scenarios") {
		t.Errorf("expected get_overview_metrics MCP output to include Phase 1 & Phase 2 FinOps fields (weekly_pool, scenarios)")
	}
	if !strings.Contains(lines[4], "standard_seats") || !strings.Contains(lines[4], "plus_seats") || !strings.Contains(lines[4], "weekly_pool") {
		t.Errorf("expected get_license_status MCP output to include standard_seats, plus_seats, and weekly_pool")
	}
	if !strings.Contains(lines[5], "fair_share_payg_cost") || !strings.Contains(lines[5], "when_used_payg_cost") {
		t.Errorf("expected get_users_costs MCP output to include fair_share_payg_cost and when_used_payg_cost")
	}
	if !strings.Contains(lines[6], "top_effort") || !strings.Contains(lines[6], "total_wtu") || !strings.Contains(lines[6], "cache_hit_rate") || !strings.Contains(lines[6], "included_cost") {
		t.Errorf("expected get_user_dashboard MCP output to include Phase 1 FinOps fields")
	}
	if !strings.Contains(listResp.Result.Tools[0].Description, "WTU") {
		t.Errorf("expected get_overview_metrics tool description to mention WTU, got: %s", listResp.Result.Tools[0].Description)
	}
}

func TestMCPErrorHandling(t *testing.T) {
	srv := newTestMCPServer()

	t.Run("unknown method returns JSON-RPC method not found error", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":99,"method":"unknown/method"}` + "\n"
		var out bytes.Buffer
		if err := srv.ServeStdio(context.Background(), strings.NewReader(req), &out); err != nil {
			t.Fatalf("unexpected stdio error: %v", err)
		}
		var resp mcp.Response
		if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if resp.Error == nil || resp.Error.Code != -32601 {
			t.Fatalf("expected -32601 Method not found error, got %+v", resp.Error)
		}
	})

	t.Run("invalid JSON returns parse error", func(t *testing.T) {
		req := `{broken json` + "\n"
		var out bytes.Buffer
		if err := srv.ServeStdio(context.Background(), strings.NewReader(req), &out); err != nil {
			t.Fatalf("unexpected stdio error: %v", err)
		}
		var resp mcp.Response
		if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if resp.Error == nil || resp.Error.Code != -32700 {
			t.Fatalf("expected -32700 Parse error, got %+v", resp.Error)
		}
	})

	t.Run("missing user_id in get_user_dashboard returns tool error", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_user_dashboard","arguments":{}}}` + "\n"
		var out bytes.Buffer
		_ = srv.ServeStdio(context.Background(), strings.NewReader(req), &out)
		var callResp struct {
			Result mcp.ToolResult `json:"result"`
		}
		if err := json.Unmarshal(out.Bytes(), &callResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !callResp.Result.IsError {
			t.Fatalf("expected isError=true for missing user_id")
		}
	})

	t.Run("unknown user_id in get_user_dashboard returns tool error", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"get_user_dashboard","arguments":{"user_id":"nobody@example.com"}}}` + "\n"
		var out bytes.Buffer
		_ = srv.ServeStdio(context.Background(), strings.NewReader(req), &out)
		var callResp struct {
			Result mcp.ToolResult `json:"result"`
		}
		if err := json.Unmarshal(out.Bytes(), &callResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !callResp.Result.IsError {
			t.Fatalf("expected isError=true for unknown user_id")
		}
	})

	t.Run("unknown tool name returns tool error", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"nonexistent_tool","arguments":{}}}` + "\n"
		var out bytes.Buffer
		_ = srv.ServeStdio(context.Background(), strings.NewReader(req), &out)
		var callResp struct {
			Result mcp.ToolResult `json:"result"`
		}
		if err := json.Unmarshal(out.Bytes(), &callResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !callResp.Result.IsError {
			t.Fatalf("expected isError=true for unknown tool")
		}
	})
}

func TestMCPHTTPHandler(t *testing.T) {
	srv := newTestMCPServer()

	t.Run("POST /mcp with tools/list returns 200 OK with tools", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "get_overview_metrics") {
			t.Errorf("expected tools/list response to contain get_overview_metrics, got %s", rec.Body.String())
		}
	})

	t.Run("POST /mcp with notification returns 202 Accepted", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("expected 202 Accepted for notification, got %d", rec.Code)
		}
	})

	t.Run("GET /mcp returns server metadata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on GET /mcp info, got %d", rec.Code)
		}
	})
}
