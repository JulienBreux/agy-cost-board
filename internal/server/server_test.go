package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/julienbreux/agy-cost-board/internal/server"
	"github.com/julienbreux/agy-cost-board/internal/setup"
)

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()
	provider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(provider, 5*time.Minute)

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body><h1>AGY GE Board</h1></body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('loaded');"),
		},
	}

	srv := server.NewServer(engine, mockFS)
	return srv.Router()
}

func TestAPIEndpoints(t *testing.T) {
	router := setupTestServer(t)

	t.Run("GET /healthz returns status 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var res map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if res["status"] != "ok" {
			t.Errorf("expected status: ok, got %v", res["status"])
		}
	})

	t.Run("GET /api/v1/metrics/overview returns high-level metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/overview?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var metrics domain.OverviewMetrics
		if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
			t.Fatalf("failed to decode overview metrics: %v", err)
		}
		if metrics.TotalBilledCost <= 0 {
			t.Errorf("expected positive total billed cost, got %f", metrics.TotalBilledCost)
		}
		if metrics.TotalTokens <= 0 {
			t.Errorf("expected positive total tokens, got %d", metrics.TotalTokens)
		}
		if len(metrics.DailyTrends) == 0 {
			t.Errorf("expected non-empty daily spend trend")
		}
	})

	t.Run("GET /api/v1/costs/users returns list of attributed user costs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/costs/users?days=14", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var costs []domain.AllocatedUserCost
		if err := json.NewDecoder(rec.Body).Decode(&costs); err != nil {
			t.Fatalf("failed to decode costs: %v", err)
		}
		if len(costs) == 0 {
			t.Errorf("expected non-empty costs slice")
		}
		if costs[0].UserID == "" || costs[0].AllocatedCost < 0 {
			t.Errorf("invalid cost record: %+v", costs[0])
		}
	})

	t.Run("GET /api/v1/costs/users with model filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/costs/users?days=14&model=gemini-1.5-pro", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var costs []domain.AllocatedUserCost
		if err := json.NewDecoder(rec.Body).Decode(&costs); err != nil {
			t.Fatalf("failed to decode filtered costs: %v", err)
		}
		for _, c := range costs {
			if c.Model != "gemini-1.5-pro" {
				t.Errorf("expected model gemini-1.5-pro, got %s", c.Model)
			}
		}
	})

	t.Run("GET /api/v1/licenses/status returns license governance metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/licenses/status?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var gov domain.LicenseGovernanceSummary
		if err := json.NewDecoder(rec.Body).Decode(&gov); err != nil {
			t.Fatalf("failed to decode governance: %v", err)
		}
		if gov.SeatQuota <= 0 {
			t.Errorf("expected positive seat quota, got %d", gov.SeatQuota)
		}
		if gov.DormantSeats < 0 {
			t.Errorf("unexpected dormant seats: %d", gov.DormantSeats)
		}
	})

	t.Run("GET /api/v1/users/{id} returns user breakdown summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/alex.turner@example.com?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var user domain.UserCostSummary
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode user summary: %v", err)
		}
		if user.UserID != "alex.turner@example.com" {
			t.Errorf("expected alex.turner@example.com, got %s", user.UserID)
		}
		if user.TotalCost <= 0 {
			t.Errorf("expected positive total cost, got %f", user.TotalCost)
		}
	})

	t.Run("GET /api/v1/users/unknown@example.com returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/unknown.nobody@example.com?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", rec.Code)
		}
	})

	t.Run("Static asset serving and SPA fallback to index.html", func(t *testing.T) {
		// Exact static asset
		req1 := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)
		if rec1.Code != http.StatusOK {
			t.Errorf("expected 200 for /assets/app.js, got %d", rec1.Code)
		}

		// Root path
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusOK {
			t.Errorf("expected 200 for /, got %d", rec2.Code)
		}

		// SPA route (e.g. /licenses) falls back to index.html
		req3 := httptest.NewRequest(http.MethodGet, "/licenses", nil)
		rec3 := httptest.NewRecorder()
		router.ServeHTTP(rec3, req3)
		if rec3.Code != http.StatusOK {
			t.Errorf("expected 200 for SPA fallback /licenses, got %d", rec3.Code)
		}
	})

	t.Run("GET /api/v1/setup/status returns diagnostic report", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var report domain.DiagnosticReport
		if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}

		if len(report.Checks) != 5 {
			t.Errorf("expected 5 checks in report, got %d", len(report.Checks))
		}
		if report.OverallStatus != domain.StatusOK {
			t.Errorf("expected overall OK in demo setup, got %s", report.OverallStatus)
		}
	})

	t.Run("query parsing uses default for invalid or negative days", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/overview?days=invalid", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 with default days on invalid query, got %d", rec.Code)
		}

		reqNeg := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/overview?days=-10", nil)
		recNeg := httptest.NewRecorder()
		router.ServeHTTP(recNeg, reqNeg)
		if recNeg.Code != http.StatusOK {
			t.Errorf("expected 200 with default days on negative query, got %d", recNeg.Code)
		}
	})
}

type mockSetupRunner struct {
	invoked bool
}

func (m *mockSetupRunner) RunAll(ctx context.Context, cfg setup.Config) *domain.DiagnosticReport {
	m.invoked = true
	return &domain.DiagnosticReport{
		ProjectID:     cfg.ProjectID,
		OverallStatus: domain.StatusWarning,
		Checks: []domain.CheckResult{
			{ID: "test-check", Status: domain.StatusWarning},
		},
	}
}

func TestServerWithCustomRunner(t *testing.T) {
	provider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(provider, 5*time.Minute)
	runner := &mockSetupRunner{}

	srv := server.NewServerWithSetup(engine, nil, setup.Config{ProjectID: "custom-test-proj"}, runner)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !runner.invoked {
		t.Errorf("expected custom runner to be invoked")
	}
	var report domain.DiagnosticReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if report.OverallStatus != domain.StatusWarning {
		t.Errorf("expected status WARNING from custom runner, got %s", report.OverallStatus)
	}
	if report.ProjectID != "custom-test-proj" {
		t.Errorf("expected project custom-test-proj, got %s", report.ProjectID)
	}
}

func TestCurrentUserEndpoint(t *testing.T) {
	router := setupTestServer(t)

	t.Run("GET /api/v1/me extracts identity from X-Goog-Authenticated-User-Email", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("X-Goog-Authenticated-User-Email", "accounts.google.com:sarah.connor@example.com")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var user domain.CurrentUserIdentity
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("decode err: %v", err)
		}
		if user.UserID != "sarah.connor@example.com" {
			t.Errorf("expected user_id sarah.connor@example.com, got %s", user.UserID)
		}
		if !user.Authenticated {
			t.Errorf("expected authenticated true")
		}
		if user.Source != "iap" {
			t.Errorf("expected source iap, got %s", user.Source)
		}
	})

	t.Run("GET /api/v1/me falls back to demo user when unauthenticated in demo mode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var user domain.CurrentUserIdentity
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("decode err: %v", err)
		}
		if user.UserID == "" {
			t.Errorf("expected demo user_id to be populated")
		}
		if user.Source != "demo" {
			t.Errorf("expected source demo, got %s", user.Source)
		}
	})
}

func TestUserActivityEndpoint(t *testing.T) {
	router := setupTestServer(t)

	t.Run("GET /api/v1/users/{id}/activity returns recent inference events", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/alex.turner@example.com/activity?days=30&limit=5", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}
		var activity []domain.UserActivityLog
		if err := json.NewDecoder(rec.Body).Decode(&activity); err != nil {
			t.Fatalf("decode err: %v", err)
		}
		if len(activity) == 0 {
			t.Errorf("expected non-empty activity for alex.turner@example.com")
		}
		if len(activity) > 5 {
			t.Errorf("expected at most 5 records, got %d", len(activity))
		}
		first := activity[0]
		if first.UserID != "alex.turner@example.com" {
			t.Errorf("expected user_id alex.turner@example.com, got %s", first.UserID)
		}
		if first.TotalTokens <= 0 {
			t.Errorf("expected total_tokens > 0, got %d", first.TotalTokens)
		}
	})
}

func TestUserDashboardEndpoint(t *testing.T) {
	router := setupTestServer(t)

	t.Run("GET /api/v1/users/{id}/dashboard returns driving metrics and recommendations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/alex.turner@example.com/dashboard?days=30&budget=250", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}
		var driving domain.UserConsumptionDriving
		if err := json.NewDecoder(rec.Body).Decode(&driving); err != nil {
			t.Fatalf("decode err: %v", err)
		}
		if driving.UserID != "alex.turner@example.com" {
			t.Errorf("expected user_id alex.turner@example.com, got %s", driving.UserID)
		}
		if driving.MonthlyBudget != 250.0 {
			t.Errorf("expected budget 250.0, got %f", driving.MonthlyBudget)
		}
		if len(driving.DailyTrends) == 0 {
			t.Errorf("expected daily trends")
		}
	})
}

