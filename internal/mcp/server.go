package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
)

// ProtocolVersion is the Model Context Protocol specification version supported by this server.
const ProtocolVersion = "2025-03-26"

// Request represents a JSON-RPC 2.0 request or notification message.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response message.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool describes an MCP tool definition and its JSON Schema input contract.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ContentBlock represents an item inside a ToolResult content array.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolResult represents the result returned from tools/call.
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// Server implements a dual-transport (Stdio & Streamable HTTP) MCP server for cost and quota tools.
type Server struct {
	engine  *attribution.Engine
	version string
}

// NewServer creates a new MCP Server backed by the provided attribution Engine.
func NewServer(engine *attribution.Engine, version string) *Server {
	if version == "" {
		version = "dev"
	}
	return &Server{
		engine:  engine,
		version: version,
	}
}

// Tools returns the list of cost and quota MCP tools exposed by agy-cost-board.
func (s *Server) Tools() []Tool {
	return []Tool{
		{
			Name:        "get_overview_metrics",
			Description: "Retrieve project-wide Antigravity AI billed spend (Included seat allowance vs. PAYG overage), total token volumes, Weighted Token Units (WTU), overall Cache Hit Rate %, active users count, Gemini Enterprise seat quota utilization, per-model cost breakdown, token/cost share categories, thinking effort stats, and client breakdown.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"days": map[string]any{
						"type":        "integer",
						"description": "Lookback telemetry window in days (default: 30).",
						"default":     30,
					},
				},
			},
		},
		{
			Name:        "get_license_status",
			Description: "Retrieve Gemini Enterprise license seat quota, assigned vs. active/dormant seats, utilization rate, estimated monthly savings, and dormant users list.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"days": map[string]any{
						"type":        "integer",
						"description": "Lookback window in days to evaluate seat activity (default: 30).",
						"default":     30,
					},
				},
			},
		},
		{
			Name:        "get_users_costs",
			Description: "List proportional per-user AI cost attributions (including Included vs. PAYG cost, WTU, and Cache Hit Rate %) across the project, with optional lookback days and model filters.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"days": map[string]any{
						"type":        "integer",
						"description": "Lookback telemetry window in days (default: 30).",
						"default":     30,
					},
					"model": map[string]any{
						"type":        "string",
						"description": "Optional model identifier filter (e.g., gemini-2.5-pro, gemini-2.5-flash).",
					},
				},
			},
		},
		{
			Name:        "get_user_dashboard",
			Description: "Retrieve a developer's personal consumption driving dashboard, including total spend (Included vs. PAYG), WTU, Cache Hit Rate %, daily/weekly burn rate, monthly budget quota consumed percentage, projected month-end spend, model breakdown, thinking effort breakdown, client stats, and optimization tips.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"user_id"},
				"properties": map[string]any{
					"user_id": map[string]any{
						"type":        "string",
						"description": "Developer user email or identifier (e.g., alex.turner@google.com).",
					},
					"days": map[string]any{
						"type":        "integer",
						"description": "Lookback telemetry window in days (default: 30).",
						"default":     30,
					},
					"monthly_budget": map[string]any{
						"type":        "number",
						"description": "Target monthly budget threshold for quota calculations (default: 150.0).",
						"default":     150.0,
					},
				},
			},
		},
	}
}

// ServeStdio processes line-delimited JSON-RPC 2.0 messages from r and writes responses to w.
func (s *Server) ServeStdio(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	// Allow up to 4MB per JSON-RPC message line
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	enc := json.NewEncoder(w)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		resp, hasResponse := s.HandleMessage(ctx, []byte(line))
		if !hasResponse {
			continue
		}
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("encoding stdio response: %w", err)
		}
	}
	return scanner.Err()
}

// ServeHTTP handles MCP Streamable HTTP requests (`POST /mcp` for JSON-RPC messages and `GET /mcp` for discovery metadata).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":            "agy-cost-board-mcp",
			"version":         s.version,
			"protocolVersion": ProtocolVersion,
			"transport":       "streamable-http",
			"toolsCount":      len(s.Tools()),
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4*1024*1024))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{
			JSONRPC: "2.0",
			ID:      json.RawMessage("null"),
			Error:   &RPCError{Code: -32700, Message: "Failed to read request body"},
		})
		return
	}

	resp, hasResponse := s.HandleMessage(r.Context(), body)
	if !hasResponse {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleMessage processes a single raw JSON-RPC 2.0 payload and returns a Response when appropriate.
// Notifications (requests without an ID) return hasResponse=false.
func (s *Server) HandleMessage(ctx context.Context, raw []byte) (Response, bool) {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return Response{
			JSONRPC: "2.0",
			ID:      json.RawMessage("null"),
			Error:   &RPCError{Code: -32700, Message: "Parse error: " + err.Error()},
		}, true
	}

	// JSON-RPC notifications have no "id" field and must not receive a response
	isNotification := len(req.ID) == 0 || string(req.ID) == "null" && strings.HasPrefix(req.Method, "notifications/")
	if isNotification && (strings.HasPrefix(req.Method, "notifications/") || req.Method == "initialized") {
		return Response{}, false
	}

	switch req.Method {
	case "initialize":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": ProtocolVersion,
				"capabilities": map[string]any{
					"tools": map[string]any{
						"listChanged": false,
					},
				},
				"serverInfo": map[string]any{
					"name":    "agy-cost-board",
					"version": s.version,
				},
			},
		}, true

	case "ping":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{},
		}, true

	case "tools/list":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.Tools(),
			},
		}, true

	case "tools/call":
		res := s.handleToolCall(ctx, req.Params)
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  res,
		}, true

	default:
		if isNotification {
			return Response{}, false
		}
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found: " + req.Method,
			},
		}, true
	}
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (s *Server) handleToolCall(ctx context.Context, rawParams json.RawMessage) ToolResult {
	var params toolCallParams
	if len(rawParams) > 0 {
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return errorToolResult(fmt.Sprintf("invalid tool call params: %v", err))
		}
	}

	args := params.Arguments
	if args == nil {
		args = map[string]any{}
	}

	switch params.Name {
	case "get_overview_metrics":
		days := intArg(args, "days", 30)
		metrics, err := s.engine.GetOverviewMetrics(ctx, days)
		if err != nil {
			return errorToolResult(fmt.Sprintf("failed to get overview metrics: %v", err))
		}
		return jsonToolResult(metrics)

	case "get_license_status":
		days := intArg(args, "days", 30)
		gov, err := s.engine.GetLicenseGovernance(ctx, days)
		if err != nil {
			return errorToolResult(fmt.Sprintf("failed to get license governance: %v", err))
		}
		return jsonToolResult(gov)

	case "get_users_costs":
		days := intArg(args, "days", 30)
		model := stringArg(args, "model", "")
		costs, err := s.engine.GetAttributedCosts(ctx, days, model)
		if err != nil {
			return errorToolResult(fmt.Sprintf("failed to get attributed user costs: %v", err))
		}
		return jsonToolResult(costs)

	case "get_user_dashboard":
		userID := strings.TrimSpace(stringArg(args, "user_id", ""))
		if userID == "" {
			return errorToolResult("missing required argument: user_id")
		}
		days := intArg(args, "days", 30)
		budget := floatArg(args, "monthly_budget", 150.0)
		driving, err := s.engine.GetUserConsumptionDriving(ctx, userID, days, budget)
		if err != nil {
			return errorToolResult(fmt.Sprintf("failed to get user dashboard for %q: %v", userID, err))
		}
		summary, _ := s.engine.GetUserSummary(ctx, userID, days)
		return jsonToolResult(map[string]any{
			"dashboard": driving,
			"summary":   summary,
		})

	default:
		return errorToolResult("unknown tool: " + params.Name)
	}
}

func jsonToolResult(data any) ToolResult {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errorToolResult(fmt.Sprintf("failed to serialize tool output: %v", err))
	}
	return ToolResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: string(b),
			},
		},
	}
}

func errorToolResult(msg string) ToolResult {
	return ToolResult{
		IsError: true,
		Content: []ContentBlock{
			{
				Type: "text",
				Text: msg,
			},
		},
	}
}

func intArg(args map[string]any, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		if int(n) > 0 {
			return int(n)
		}
	case int:
		if n > 0 {
			return n
		}
	}
	return defaultVal
}

func floatArg(args map[string]any, key string, defaultVal float64) float64 {
	v, ok := args[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		if n > 0 {
			return n
		}
	case int:
		if n > 0 {
			return float64(n)
		}
	}
	return defaultVal
}

func stringArg(args map[string]any, key string, defaultVal string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return defaultVal
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultVal
}
