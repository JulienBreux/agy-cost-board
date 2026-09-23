package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienbreux/agy-ge-board/internal/attribution"
	"github.com/julienbreux/agy-ge-board/internal/bigquery"
	"github.com/julienbreux/agy-ge-board/internal/cli"
	"github.com/julienbreux/agy-ge-board/internal/domain"
	"github.com/julienbreux/agy-ge-board/internal/server"
	"github.com/julienbreux/agy-ge-board/web"
)

func TestEndToEndCLIFlows(t *testing.T) {
	t.Run("E2E CLI cost command with table output", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"cost", "--demo", "--format=table", "--days=14"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("CLI command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "USER") || !strings.Contains(out, "MODEL") || !strings.Contains(out, "ALLOCATED COST") {
			t.Errorf("expected table header in output, got: %s", out)
		}
		if !strings.Contains(out, "alex.turner@example.com") {
			t.Errorf("expected demo user in output, got: %s", out)
		}
	})

	t.Run("E2E CLI license command with JSON output", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"license", "--demo", "--format=json", "--days=30"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("CLI command failed: %v", err)
		}

		var summary domain.LicenseGovernanceSummary
		if err := json.Unmarshal(buf.Bytes(), &summary); err != nil {
			t.Fatalf("failed to parse JSON license output: %v, raw: %s", err, buf.String())
		}

		if summary.SeatQuota <= 0 {
			t.Errorf("expected positive seat quota, got %d", summary.SeatQuota)
		}
		if summary.EstimatedMonthlySavings <= 0 && summary.DormantSeats > 0 {
			t.Errorf("expected savings when dormant seats exist: %f", summary.EstimatedMonthlySavings)
		}
	})

	t.Run("E2E CLI user command breakdown", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := cli.NewRootCommand()
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"user", "sophia.chen@example.com", "--demo", "--format=json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("CLI command failed: %v", err)
		}

		var user domain.UserCostSummary
		if err := json.Unmarshal(buf.Bytes(), &user); err != nil {
			t.Fatalf("failed to parse JSON user summary: %v, raw: %s", err, buf.String())
		}

		if user.UserID != "sophia.chen@example.com" {
			t.Errorf("expected sophia.chen@example.com, got %s", user.UserID)
		}
		if user.TotalCost <= 0 {
			t.Errorf("expected positive total cost, got %f", user.TotalCost)
		}
		if len(user.ModelBreakdown) == 0 {
			t.Errorf("expected non-empty model breakdown")
		}
	})
}

func TestEndToEndHTTPServerFlows(t *testing.T) {
	// Setup real Engine with Demo provider and Embedded Web FS
	provider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(provider, 5*time.Minute)
	staticFS, err := web.FS()
	if err != nil {
		t.Fatalf("failed to load embedded web FS: %v", err)
	}

	srv := server.NewServer(engine, staticFS)
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	client := ts.Client()

	t.Run("GET /healthz probe", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/healthz")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), `"status":"ok"`) {
			t.Errorf("unexpected healthz body: %s", string(body))
		}
	})

	t.Run("GET /api/v1/metrics/overview end-to-end", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/api/v1/metrics/overview?days=30")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		var metrics domain.OverviewMetrics
		if err := json.NewDecoder(res.Body).Decode(&metrics); err != nil {
			t.Fatalf("failed to decode overview metrics: %v", err)
		}

		if metrics.TotalBilledCost <= 0 {
			t.Errorf("expected positive billed cost, got %f", metrics.TotalBilledCost)
		}
		if metrics.TotalTokens <= 0 {
			t.Errorf("expected positive tokens, got %d", metrics.TotalTokens)
		}
		if len(metrics.DailyTrends) == 0 {
			t.Errorf("expected daily trends")
		}
	})

	t.Run("GET /api/v1/costs/users end-to-end", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/api/v1/costs/users?days=14")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		var costs []domain.AllocatedUserCost
		if err := json.NewDecoder(res.Body).Decode(&costs); err != nil {
			t.Fatalf("failed to decode costs: %v", err)
		}

		if len(costs) == 0 {
			t.Fatalf("expected non-empty cost records")
		}

		var totalCost float64
		for _, c := range costs {
			totalCost += c.AllocatedCost
		}
		if totalCost <= 0 {
			t.Errorf("expected positive total allocated cost, got: %f", totalCost)
		}
	})

	t.Run("GET / embedded SPA index.html", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "AGY & Gemini Enterprise") {
			t.Errorf("expected title in SPA index.html, got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, `<div id="root"></div>`) {
			t.Errorf("expected root div in SPA index.html")
		}
	})

	t.Run("GET /licenses SPA client-side fallback routing", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/licenses")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for SPA route, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), `<div id="root"></div>`) {
			t.Errorf("expected SPA fallback HTML, got: %s", string(body))
		}
	})
}
