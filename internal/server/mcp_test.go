package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/mcp"
	"github.com/julienbreux/agy-cost-board/internal/server"
)

func TestMCPEndpoints(t *testing.T) {
	newTestSrv := func(t *testing.T) *server.Server {
		t.Helper()
		provider := bigquery.NewDemoDataProvider()
		engine := attribution.NewEngine(provider, 5*time.Minute)
		return server.NewServer(engine, nil)
	}

	t.Run("POST /mcp handles initialize, tools/list, and tools/call when MCPEnabled is true", func(t *testing.T) {
		srv := newTestSrv(t)
		router := srv.Router()

		// 1. initialize
		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(initBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for POST /mcp initialize, got %d: %s", rec.Code, rec.Body.String())
		}

		var initResp mcp.Response
		if err := json.NewDecoder(rec.Body).Decode(&initResp); err != nil {
			t.Fatalf("failed to decode initialize response: %v", err)
		}
		if initResp.Error != nil {
			t.Fatalf("unexpected RPC error: %+v", initResp.Error)
		}

		// 2. tools/list
		listBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
		reqList := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(listBody))
		reqList.Header.Set("Content-Type", "application/json")
		recList := httptest.NewRecorder()
		router.ServeHTTP(recList, reqList)

		if recList.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for POST /mcp tools/list, got %d", recList.Code)
		}

		// 3. tools/call
		callBody := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_license_status","arguments":{"days":30}}}`
		reqCall := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(callBody))
		reqCall.Header.Set("Content-Type", "application/json")
		recCall := httptest.NewRecorder()
		router.ServeHTTP(recCall, reqCall)

		if recCall.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for POST /mcp tools/call, got %d", recCall.Code)
		}
	})

	t.Run("GET /mcp returns 200 metadata when MCPEnabled is true", func(t *testing.T) {
		srv := newTestSrv(t)
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for GET /mcp, got %d", rec.Code)
		}
	})

	t.Run("POST /mcp and GET /mcp return 404 Not Found when MCPEnabled is false", func(t *testing.T) {
		srv := newTestSrv(t)
		srv.SetMCPEnabled(false)

		reqPost := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		recPost := httptest.NewRecorder()
		srv.Router().ServeHTTP(recPost, reqPost)
		if recPost.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found for POST /mcp when disabled, got %d", recPost.Code)
		}

		reqGet := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		recGet := httptest.NewRecorder()
		srv.Router().ServeHTTP(recGet, reqGet)
		if recGet.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found for GET /mcp when disabled, got %d", recGet.Code)
		}
	})

	t.Run("GET /api/v1/mcp/status and /api/v1/version expose mcp_enabled state", func(t *testing.T) {
		srv := newTestSrv(t)
		srv.SetMCPEnabled(true)

		reqStatus := httptest.NewRequest(http.MethodGet, "/api/v1/mcp/status", nil)
		recStatus := httptest.NewRecorder()
		srv.Router().ServeHTTP(recStatus, reqStatus)
		if recStatus.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for GET /api/v1/mcp/status, got %d", recStatus.Code)
		}

		var status struct {
			MCPEnabled      bool     `json:"mcp_enabled"`
			ProtocolVersion string   `json:"protocol_version"`
			Endpoint        string   `json:"endpoint"`
			Transports      []string `json:"transports"`
		}
		if err := json.NewDecoder(recStatus.Body).Decode(&status); err != nil {
			t.Fatalf("decode /api/v1/mcp/status error: %v", err)
		}
		if !status.MCPEnabled {
			t.Errorf("expected mcp_enabled true, got false")
		}
		if status.Endpoint != "/mcp" {
			t.Errorf("expected endpoint /mcp, got %s", status.Endpoint)
		}

		// Disable MCP and re-verify
		srv.SetMCPEnabled(false)
		recStatusDisabled := httptest.NewRecorder()
		srv.Router().ServeHTTP(recStatusDisabled, httptest.NewRequest(http.MethodGet, "/api/v1/mcp/status", nil))
		if err := json.NewDecoder(recStatusDisabled.Body).Decode(&status); err != nil {
			t.Fatalf("decode /api/v1/mcp/status error: %v", err)
		}
		if status.MCPEnabled {
			t.Errorf("expected mcp_enabled false after SetMCPEnabled(false), got true")
		}
	})
}
